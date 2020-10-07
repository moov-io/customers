// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package documents

import (
	"bufio"
	"context"
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

type sizeLimit uint64

func (l sizeLimit) String() string {
	return fmt.Sprintf("%dMB", l>>20)
}

const (
	maxDocumentSize sizeLimit = 20 << 20                    // 20MB
	maxFormSize     sizeLimit = maxDocumentSize + (5 << 20) // restricts request body size to allow for the document plus a small buffer
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
		r.Body = http.MaxBytesReader(w, r.Body, int64(maxFormSize))
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			if strings.Contains(err.Error(), "request body too large") {
				logger.Log("documents", "max form size exceeded", "customerID", customerID, "requestID", requestID)
				moovhttp.Problem(w, fmt.Errorf("request body exceeds maximum size of %s", maxFormSize))
				return
			}
			logger.Log("documents", "error reading form file", "error", err, "customerID", customerID, "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf("expected multipart upload with key of 'file' error=%v", err))
			return
		}
		defer file.Close()

		if fileHeader.Size > int64(maxDocumentSize) {
			logger.Log("documents", "max file size exceeded", "customerID", customerID, "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf("file exceeds maximum size of %s", maxDocumentSize))
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
			moovhttp.Problem(w, fmt.Errorf("file upload error - %v", err))
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
