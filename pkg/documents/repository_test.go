// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package documents

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/base"
	"github.com/moov-io/base/database"
	"github.com/moov-io/base/log"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/customers"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/moov-io/customers/pkg/watchman"
	"github.com/stretchr/testify/require"
)

func TestDocumentRepository(t *testing.T) {
	tests := []struct {
		dbName string
		db     *sql.DB
	}{
		{
			dbName: "sqlite",
			db:     database.CreateTestSqliteDB(t).DB,
		},
		{
			dbName: "mysql",
			db:     database.CreateTestMySQLDB(t).DB,
		},
	}

	for _, tc := range tests {
		defer tc.db.Close()

		t.Run(tc.dbName, func(t *testing.T) {
			logger := log.NewNopLogger()
			documentRepo := NewDocumentRepo(logger, tc.db)
			customerRepo := customers.NewCustomerRepo(logger, tc.db)
			organization := "test"

			// check empty docs
			docs, err := documentRepo.getCustomerDocuments(base.ID(), organization)
			require.NoError(t, err)
			require.Empty(t, docs)

			// create test customer with organization
			router := mux.NewRouter()
			ssnStorage := customers.NewSSNStorage(secrets.TestStringKeeper(t), customers.NewCustomerSSNRepository(logger, tc.db))
			ofacSearcher := customers.NewOFACSearcher(customerRepo, &watchman.TestWatchmanClient{})
			customers.AddCustomerRoutes(log.NewNopLogger(), router, customerRepo, ssnStorage, ofacSearcher)
			body := `{"firstName": "jane", "lastName": "doe", "email": "jane@example.com", "birthDate": "1991-04-01", "ssn": "123456789", "type": "individual"}`
			req := httptest.NewRequest("POST", "/customers", strings.NewReader(body))

			req.Header.Add("X-Organization", organization)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			require.Equal(t, http.StatusOK, res.Code)

			var cust client.Customer
			if err := json.NewDecoder(res.Body).Decode(&cust); err != nil {
				t.Fatal(err)
			}

			// Write a Document and read it back
			doc := &client.Document{
				DocumentID:  base.ID(),
				Type:        "DriversLicense",
				ContentType: "image/png",
			}
			if err := documentRepo.writeCustomerDocument(cust.CustomerID, doc); err != nil {
				t.Fatal(err)
			}
			docs, err = documentRepo.getCustomerDocuments(cust.CustomerID, organization)
			require.NoError(t, err)
			require.Len(t, docs, 1)

			require.Equal(t, doc.DocumentID, docs[0].DocumentID)
			require.Equal(t, "image/png", docs[0].ContentType)

			// make sure we read the document
			exists, err := documentRepo.exists(cust.CustomerID, doc.DocumentID, organization)
			require.Equal(t, true, exists)
			require.NoError(t, err)
		})
	}
}

func TestDocumentsRepository__Delete(t *testing.T) {
	db := database.CreateTestSqliteDB(t)
	repo := &sqlDocumentRepository{db.DB, log.NewNopLogger()}

	type document struct {
		*client.Document
		deleted bool
	}

	customerID := base.ID()
	docs := make([]*document, 10)
	for i := 0; i < len(docs); i++ {
		doc := &client.Document{
			DocumentID:  base.ID(),
			Type:        "DriversLicense",
			ContentType: "image/png",
		}
		err := repo.writeCustomerDocument(customerID, doc)
		require.NoError(t, err)
		docs[i] = &document{
			Document: doc,
		}
	}

	// mark documents to be deleted
	indexesToDelete := []int{1, 2, 5, 8}
	for _, idx := range indexesToDelete {
		require.Less(t, idx, len(docs))
		docs[idx].deleted = true
		require.NoError(t,
			repo.deleteCustomerDocument(customerID, docs[idx].DocumentID),
		)
	}

	deletedDocIDs := make(map[string]bool)
	// query all documents that have been marked as deleted
	query := ` select document_id from documents where deleted_at is not null `
	stmt, err := repo.db.Prepare(query)
	require.NoError(t, err)

	rows, err := stmt.Query()
	require.NoError(t, err)

	for rows.Next() {
		var ID *string
		require.NoError(t, rows.Scan(&ID))
		deletedDocIDs[*ID] = true
	}

	for _, doc := range docs {
		_, ok := deletedDocIDs[doc.DocumentID]
		require.Equal(t, doc.deleted, ok)
	}

	// make sure we find the document as deleted
	exists, err := repo.exists(customerID, docs[0].DocumentID, "")
	require.Equal(t, false, exists)
	require.Equal(t, err, sql.ErrNoRows)
}
