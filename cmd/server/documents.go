// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

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
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"
)

var (
	errNoDocumentId = errors.New("no Document ID found")

	maxDocumentSize int64 = 20 * 1024 * 1024 // 20MB
)

func addDocumentRoutes(logger log.Logger, r *mux.Router, repo documentRepository, bucketFactory bucketFunc) {
	r.Methods("GET").Path("/customers/{customerId}/documents").HandlerFunc(getCustomerDocuments(logger, repo))
	r.Methods("POST").Path("/customers/{customerId}/documents").HandlerFunc(uploadCustomerDocument(logger, repo, bucketFactory))
	r.Methods("GET").Path("/customers/{customerId}/documents/{documentId}").HandlerFunc(retrieveRawDocument(logger, repo, bucketFactory))
}

func getDocumentId(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["documentId"]
	if !ok || v == "" {
		moovhttp.Problem(w, errNoDocumentId)
		return ""
	}
	return v
}

func getCustomerDocuments(logger log.Logger, repo documentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)

		customerId := getCustomerId(w, r)
		if customerId == "" {
			return
		}

		docs, err := repo.getCustomerDocuments(customerId)
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

func uploadCustomerDocument(logger log.Logger, repo documentRepository, bucketFactory bucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)

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
			Id:          base.ID(),
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

		customerId, requestId := getCustomerId(w, r), moovhttp.GetRequestId(r)
		if customerId == "" {
			return
		}
		if err := repo.writeCustomerDocument(customerId, doc); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if logger != nil {
			logger.Log("documents", fmt.Sprintf("uploading document=%s (content-type: %s) for customer=%s", doc.Id, contentType, customerId), "requestId", requestId)
		}

		// Write our document from the request body
		ctx, cancelFn := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancelFn()

		documentKey := makeDocumentKey(customerId, doc.Id)
		if logger != nil {
			logger.Log("documents", fmt.Sprintf("writing %s", documentKey), "requestId", requestId)
		}

		writer, err := bucket.NewWriter(ctx, documentKey, &blob.WriterOptions{
			ContentDisposition: "inline",
			ContentType:        contentType,
		})
		if err != nil {
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

func retrieveRawDocument(logger log.Logger, repo documentRepository, bucketFactory bucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)

		customerId, documentId := getCustomerId(w, r), getDocumentId(w, r)
		if customerId == "" || documentId == "" {
			return
		}
		requestId := moovhttp.GetRequestId(r)

		bucket, err := bucketFactory()
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancelFn()

		documentKey := makeDocumentKey(customerId, documentId)
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

		if logger != nil {
			logger.Log("documents", fmt.Sprintf("redirecting for document=%s", documentKey), "requestId", requestId)
		}
		http.Redirect(w, r, signedURL, http.StatusFound)
	}
}

func makeDocumentKey(customerId, documentId string) string {
	return fmt.Sprintf("customer-%s-document-%s", customerId, documentId)
}

type documentRepository interface {
	getCustomerDocuments(customerId string) ([]*client.Document, error)
	writeCustomerDocument(customerId string, doc *client.Document) error
}

type sqliteDocumentRepository struct {
	db *sql.DB
}

func (r *sqliteDocumentRepository) close() error {
	return r.db.Close()
}

func (r *sqliteDocumentRepository) getCustomerDocuments(customerId string) ([]*client.Document, error) {
	query := `select document_id, type, content_type, uploaded_at from documents where customer_id = ?`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getCustomerDocuments: prepare %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerId)
	if err != nil {
		return nil, fmt.Errorf("getCustomerDocuments: query %v", err)
	}
	defer rows.Close()

	var docs []*client.Document
	for rows.Next() {
		var doc client.Document
		if err := rows.Scan(&doc.Id, &doc.Type, &doc.ContentType, &doc.UploadedAt); err != nil {
			return nil, fmt.Errorf("getCustomerDocuments: scan: %v", err)
		}
		docs = append(docs, &doc)
	}
	return docs, nil
}

func (r *sqliteDocumentRepository) writeCustomerDocument(customerId string, doc *client.Document) error {
	query := `insert into documents (document_id, customer_id, type, content_type, uploaded_at) values (?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("writeCustomerDocument: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(doc.Id, customerId, doc.Type, doc.ContentType, doc.UploadedAt); err != nil {
		return fmt.Errorf("writeCustomerDocument: exec: %v", err)
	}
	return nil
}
