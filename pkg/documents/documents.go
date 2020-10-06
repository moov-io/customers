// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package documents

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
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
	"gocloud.dev/secrets"
)

var (
	errNoDocumentID = errors.New("no Document ID found")
)

const (
	maxDocumentSize int64 = 20 << 20                    // 20MB
	maxFormSize     int64 = maxDocumentSize + (5 << 20) // restricts request body size to allow for the document plus a small buffer
)

func AddDocumentRoutes(logger log.Logger, r *mux.Router, repo DocumentRepository, keeper *secrets.Keeper, bucketFactory storage.BucketFunc) {
	r.Methods("GET").Path("/customers/{customerID}/documents").HandlerFunc(getCustomerDocuments(logger, repo))
	r.Methods("POST").Path("/customers/{customerID}/documents").HandlerFunc(uploadCustomerDocument(logger, repo, keeper, bucketFactory))
	r.Methods("GET").Path("/customers/{customerID}/documents/{documentID}").HandlerFunc(retrieveRawDocument(logger, repo, keeper, bucketFactory))
	r.Methods("DELETE").Path("/customers/{customerID}/documents/{documentID}").HandlerFunc(deleteCustomerDocument(logger, repo))
}

func getDocumentID(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["documentID"]
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
			logger.Log("documents", fmt.Sprintf("failed to %v", err), "customerID", customerID)
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

func uploadCustomerDocument(logger log.Logger, repo DocumentRepository, keeper *secrets.Keeper, bucketFactory storage.BucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		requestID := moovhttp.GetRequestID(r)
		customerID := route.GetCustomerID(w, r)
		if customerID == "" {
			return
		}
		documentType, err := readDocumentType(r.URL.Query().Get("type"))
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// if r.Body is larger than maxFormSize an error will be returned when the body is read
		r.Body = http.MaxBytesReader(w, r.Body, maxFormSize)
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			if strings.Contains(err.Error(), "request body too large") {
				logger.Log("documents", "max form size exceeded", "customerID", customerID, "requestID", requestID)
				moovhttp.Problem(w, fmt.Errorf("request body exceeds maximum size of %dMB", maxFormSize/(1<<20)))
				return
			}
			logger.Log("documents", "error reading form file", "error", err, "customerID", customerID, "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf("expected multipart upload with key of 'file' error=%v", err))
			return
		}
		defer file.Close()

		if fileHeader.Size > maxDocumentSize {
			logger.Log("documents", "max file size exceeded", "customerID", customerID, "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf("file exceeds maximum size of %dMB", maxDocumentSize/(1<<20)))
			return
		}

		fileReader := bufio.NewReader(file)
		sniff, err := fileReader.Peek(512)
		if err != nil && err != io.EOF {
			logger.Log("documents", "peek failed", "error", err, "customerID", customerID, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		contentType := http.DetectContentType(sniff)

		// Grab our cloud bucket before writing into our database
		bucket, err := bucketFactory()
		if err != nil {
			logger.Log("documents", fmt.Sprintf("failed to create bucket: %v", err), "customerID", customerID, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		doc := &client.Document{
			DocumentID:  base.ID(),
			Type:        documentType,
			ContentType: contentType,
			UploadedAt:  time.Now(),
		}
		if err := repo.writeCustomerDocument(customerID, doc); err != nil {
			logger.Log("documents", fmt.Sprintf("failed to %v", err), "customerID", customerID, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		logger.Log("documents", fmt.Sprintf("uploading document=%s (content-type: %s) for customer=%s", doc.DocumentID, contentType, customerID), "requestID", requestID)

		// Write our document from the request body
		ctx, cancelFn := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancelFn()

		fBytes := make([]byte, fileHeader.Size)
		_, err = fileReader.Read(fBytes)
		if err != nil {
			logger.Log("documents", "read failed", "error", err, "customerID", customerID, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		encryptedDoc, err := keeper.Encrypt(ctx, fBytes)
		if err != nil {
			logger.Log("documents", "failed to encrypt document", "error", err, "customer", customerID, "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf("file upload error - err=%v", err))
			return
		}

		documentKey := makeDocumentKey(customerID, doc.DocumentID)
		logger.Log("documents", fmt.Sprintf("writing %s", documentKey), "requestID", requestID)

		err = bucket.WriteAll(ctx, documentKey, encryptedDoc, &blob.WriterOptions{
			ContentDisposition: "inline",
			ContentType:        contentType,
		})
		if err != nil {
			logger.Log("documents", fmt.Sprintf("problem uploading document=%s: %v", doc.DocumentID, err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(doc)
	}
}

func retrieveRawDocument(logger log.Logger, repo DocumentRepository, keeper *secrets.Keeper, bucketFactory storage.BucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		requestID := moovhttp.GetRequestID(r)
		customerID, documentID := route.GetCustomerID(w, r), getDocumentID(w, r)
		if customerID == "" || documentID == "" {
			return
		}
		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		// reject the request if the document is deleted
		if exists, err := repo.exists(customerID, documentID, organization); !exists || err != nil {
			if err != nil {
				logger.Log("documents", fmt.Sprintf("failed to %v", err), "customerID", customerID, "documentID", documentID, "requestID", requestID)
			}
			http.NotFound(w, r)
			return
		}

		bucket, err := bucketFactory()
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancelFn()

		documentKey := makeDocumentKey(customerID, documentID)
		rdr, err := bucket.NewReader(ctx, documentKey, nil)
		if err != nil {
			moovhttp.Problem(w, fmt.Errorf("read documentID=%s: %v", documentKey, err))
			return
		}
		defer rdr.Close()

		encryptedDoc, err := ioutil.ReadAll(rdr)
		if err != nil {
			logger.Log("documents", "failed reading document from storage bucket", "error", err, "customerID", customerID, "documentID", documentID, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		doc, err := keeper.Decrypt(ctx, encryptedDoc)
		if err != nil {
			logger.Log("documents", "failed to decrypt document", "error", err, "customerID", customerID, "documentID", documentID, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", rdr.ContentType())
		w.WriteHeader(http.StatusOK)
		if n, err := w.Write(doc); n == 0 || err != nil {
			moovhttp.Problem(w, fmt.Errorf("failed writing documentID=%s (bytes read: %d): %v", documentKey, n, err))
			return
		}
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
			moovhttp.Problem(w, fmt.Errorf("failed to %v", err))
			logger.Log("documents", fmt.Sprintf("deleting document=%s for customer=%s: %v", documentID, customerID, err), "requestID", requestID)
			return
		}

		logger.Log("documents", fmt.Sprintf("successfully deleted document=%s for customer=%s", documentID, customerID), "requestID", requestID)

		w.WriteHeader(http.StatusNoContent)
	}
}

func makeDocumentKey(customerID, documentID string) string {
	return path.Join("customers", customerID, "documents", documentID)
}

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
