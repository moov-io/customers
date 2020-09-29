// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package documents

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/moov-io/base"
	moovhttp "github.com/moov-io/base/http"

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/documents/storage"
	"github.com/moov-io/customers/pkg/route"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"
)

var (
	errNoDocumentID = errors.New("no Document ID found")

	maxDocumentSize int64 = 20 * 1024 * 1024 // 20MB
)

func AddDocumentRoutes(logger log.Logger, r *mux.Router, repo DocumentRepository, bucketFactory storage.BucketFunc) {
	r.Methods("GET").Path("/customers/{customerID}/documents").HandlerFunc(getCustomerDocuments(logger, repo))
	r.Methods("POST").Path("/customers/{customerID}/documents").HandlerFunc(uploadCustomerDocument(logger, repo, bucketFactory))
	r.Methods("GET").Path("/customers/{customerID}/documents/{documentId}").HandlerFunc(retrieveRawDocument(logger, repo, bucketFactory))
	r.Methods("DELETE").Path("/customers/{customerID}/documents/{documentId}").HandlerFunc(deleteCustomerDocument(logger, repo))
}

func getDocumentID(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["documentId"]
	if !ok || v == "" {
		moovhttp.Problem(w, errNoDocumentID)
		return ""
	}
	return v
}

func getCustomerDocuments(logger log.Logger, repo DocumentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, organization := route.GetCustomerID(w, r), route.GetOrganization(w, r)
		if customerID == "" || organization == "" {
			return
		}

		docs, err := repo.getCustomerDocuments(customerID, organization)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(docs)
	}
}

func readDocumentType(v string) (string, error) {
	orig := v
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "driverslicense", "passport":
		return v, nil
	case "utilitybill", "bankstatement":
		return v, nil
	}
	return "", fmt.Errorf("unknown Document type: %s", orig)
}

func uploadCustomerDocument(logger log.Logger, repo DocumentRepository, bucketFactory storage.BucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		// TODO(adam): should we store x-organization along with the Document?

		documentType, err := readDocumentType(r.URL.Query().Get("type"))
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		file, _, err := r.FormFile("file")
		if file == nil || err != nil {
			moovhttp.Problem(w, fmt.Errorf("expected multipart upload with key of 'file' error=%v", err))
			return
		}
		defer file.Close()

		// Detect the content type by reading the first 512 bytes of r.Body (read into file as we expect a multipart request)
		buf := make([]byte, 512)
		if _, err := file.Read(buf); err != nil && err != io.EOF {
			moovhttp.Problem(w, err)
			return
		}
		contentType := http.DetectContentType(buf)
		doc := &client.Document{
			DocumentID:  base.ID(),
			Type:        documentType,
			ContentType: contentType,
			UploadedAt:  time.Now(),
		}

		// Grab our cloud bucket before writing into our database
		bucket, err := bucketFactory()
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		customerID, requestID := route.GetCustomerID(w, r), moovhttp.GetRequestID(r)
		if customerID == "" {
			return
		}
		if err := repo.writeCustomerDocument(customerID, doc); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		logger.Log("documents", fmt.Sprintf("uploading document=%s (content-type: %s) for customer=%s", doc.DocumentID, contentType, customerID), "requestID", requestID)

		// Write our document from the request body
		ctx, cancelFn := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancelFn()

		documentKey := makeDocumentKey(customerID, doc.DocumentID)
		logger.Log("documents", fmt.Sprintf("writing %s", documentKey), "requestID", requestID)

		writer, err := bucket.NewWriter(ctx, documentKey, &blob.WriterOptions{
			ContentDisposition: "inline",
			ContentType:        contentType,
		})
		if err != nil {
			logger.Log("documents", fmt.Sprintf("problem uploading document=%s: %v", doc.DocumentID, err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		defer writer.Close()

		// Concat our mime buffer back with the multipart file
		n, err := io.Copy(writer, io.LimitReader(io.MultiReader(bytes.NewReader(buf), file), maxDocumentSize))
		if err != nil || n == 0 {
			moovhttp.Problem(w, fmt.Errorf("documents: wrote %d bytes: %v", n, err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(doc)
	}
}

func retrieveRawDocument(logger log.Logger, repo DocumentRepository, bucketFactory storage.BucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, documentId := route.GetCustomerID(w, r), getDocumentID(w, r)
		if customerID == "" || documentId == "" {
			return
		}
		requestID := moovhttp.GetRequestID(r)

		bucket, err := bucketFactory()
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancelFn()

		documentKey := makeDocumentKey(customerID, documentId)
		signedURL, err := bucket.SignedURL(ctx, documentKey, &blob.SignedURLOptions{
			Expiry: 15 * time.Minute,
		})
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if signedURL == "" {
			moovhttp.Problem(w, fmt.Errorf("document=%s not found", documentId))
			return
		}

		logger.Log("documents", fmt.Sprintf("redirecting for document=%s", documentKey), "requestID", requestID)
		http.Redirect(w, r, signedURL, http.StatusFound)
	}
}

func deleteCustomerDocument(logger log.Logger, repo DocumentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		requestID := moovhttp.GetRequestID(r)

		customerID, documentID := route.GetCustomerID(w, r), getDocumentID(w, r)
		if customerID == "" || documentID == "" {
			return
		}

		err := repo.deleteCustomerDocument(customerID, documentID)
		if err != nil {
			moovhttp.Problem(w, fmt.Errorf("deleting document: %v", err))
			logger.Log("documents", fmt.Sprintf("error deleting document=%s for customer=%s: %v", documentID, customerID, err), "requestID", requestID)
			return
		}

		logger.Log("documents", fmt.Sprintf("successfully deleted document=%s for customer=%s", documentID, customerID), "requestID", requestID)

		w.WriteHeader(http.StatusNoContent)
	}
}

func makeDocumentKey(customerID, documentId string) string {
	return fmt.Sprintf("customer-%s-document-%s", customerID, documentId)
}

type DocumentRepository interface {
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
func (r *sqlDocumentRepository) getCustomerDocuments(customerID string, organization string) ([]*client.Document, error) {
	query := `select document_id, documents.type, content_type, uploaded_at from documents join customers on customers.
	organization = ? where documents.customer_id = ? and documents.deleted_at is null`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getCustomerDocuments: prepare %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(organization, customerID)
	if err != nil {
		return nil, fmt.Errorf("getCustomerDocuments: query %v", err)
	}
	defer rows.Close()

	var docs []*client.Document
	for rows.Next() {
		var doc client.Document
		if err := rows.Scan(&doc.DocumentID, &doc.Type, &doc.ContentType, &doc.UploadedAt); err != nil {
			return nil, fmt.Errorf("getCustomerDocuments: scan: %v", err)
		}
		docs = append(docs, &doc)
	}
	return docs, nil
}

func (r *sqlDocumentRepository) writeCustomerDocument(customerID string, doc *client.Document) error {
	query := `insert into documents (document_id, customer_id, type, content_type, uploaded_at) values (?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("writeCustomerDocument: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(doc.DocumentID, customerID, doc.Type, doc.ContentType, doc.UploadedAt); err != nil {
		return fmt.Errorf("writeCustomerDocument: exec: %v", err)
	}
	return nil
}

func (r *sqlDocumentRepository) deleteCustomerDocument(customerID string, documentID string) error {
	query := `update documents set deleted_at = ? where customer_id = ? and document_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), customerID, documentID)
	return err
}
