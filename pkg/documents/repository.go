// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package documents

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/moov-io/base/log"
	"github.com/moov-io/customers/pkg/client"
)

type DocumentRepository interface {
	exists(customerID string, documentID string, organization string) (bool, error)
	getCustomerDocuments(customerID string, organization string) ([]*client.Document, error)

	writeCustomerDocument(customerID string, doc *client.Document) error
	deleteCustomerDocument(customerID string, documentID string) error
}

type sqlDocumentRepository struct {
	db     *sql.DB
	logger log.Logger
}

func NewDocumentRepo(logger log.Logger, db *sql.DB) DocumentRepository {
	return &sqlDocumentRepository{
		db:     db,
		logger: logger,
	}
}

func (r *sqlDocumentRepository) exists(customerID string, documentID string, organization string) (bool, error) {
	query := `select documents.document_id from documents
inner join customers on customers.organization = ?
where documents.customer_id = ? and documents.document_id = ? and documents.deleted_at is null
limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false, fmt.Errorf("prepare exists: %v", err)
	}
	defer stmt.Close()

	var docID string
	if err := stmt.QueryRow(organization, customerID, documentID).Scan(&docID); err != nil {
		return false, err
	}
	return documentID == docID, nil
}

func (r *sqlDocumentRepository) getCustomerDocuments(customerID string, organization string) ([]*client.Document, error) {
	query := `select document_id, documents.type, content_type, uploaded_at from documents
inner join customers on customers.customer_id = documents.customer_id
where customers.organization = ? and documents.customer_id = ? and documents.deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("prepare listing documents: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(organization, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed querying customer documents: %v", err)
	}
	defer rows.Close()

	docs := make([]*client.Document, 0)
	for rows.Next() {
		var doc client.Document
		if err := rows.Scan(&doc.DocumentID, &doc.Type, &doc.ContentType, &doc.UploadedAt); err != nil {
			return nil, fmt.Errorf("scan customer documents: %v", err)
		}
		docs = append(docs, &doc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *sqlDocumentRepository) writeCustomerDocument(customerID string, doc *client.Document) error {
	query := `insert into documents (document_id, customer_id, type, content_type, uploaded_at) values (?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("prepare write: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(doc.DocumentID, customerID, doc.Type, doc.ContentType, doc.UploadedAt); err != nil {
		return fmt.Errorf("write customer document: %v", err)
	}
	return nil
}

func (r *sqlDocumentRepository) deleteCustomerDocument(customerID string, documentID string) error {
	query := `update documents set deleted_at = ? where customer_id = ? and document_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("prepare delete: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), customerID, documentID)
	return err
}
