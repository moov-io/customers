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

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
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
	logger = logger.WithKeyValue("package", "documents")

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
			logger.WithKeyValue("customerID", customerID).LogErrorF("failed to %v", err)
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

		customerID := route.GetCustomerID(w, r)
		if customerID == "" {
			return
		}

		logger = logger.WithKeyValue("customerID", customerID)

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
				logger.LogError("max form size exceeded", err)
				moovhttp.Problem(w, fmt.Errorf("request body exceeds maximum size of %s", maxFormSize))
				return
			}
			logger.LogError("error reading form file", err)
			moovhttp.Problem(w, fmt.Errorf("expected multipart upload with key of 'file' error=%v", err))
			return
		}
		defer file.Close()

		if fileHeader.Size > int64(maxDocumentSize) {
			logger.LogErrorF("failed to peek: %v", err)
			moovhttp.Problem(w, fmt.Errorf("file exceeds maximum size of %s", maxDocumentSize))
			return
		}

		fileReader := bufio.NewReader(file)
		sniff, err := fileReader.Peek(512)
		if err != nil && err != io.EOF {
			logger.LogError("peek failed", err)
			moovhttp.Problem(w, err)
			return
		}
		contentType := http.DetectContentType(sniff)

		// Grab our cloud bucket before writing into our database
		bucket, err := bucketFactory()
		if err != nil {
			logger.LogError("failed to create bucket", err)
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
			logger.LogError("failed to write customer document", err)
			moovhttp.Problem(w, err)
			return
		}
		logger.Log(fmt.Sprintf("uploading document=%s (content-type: %s)", doc.DocumentID, contentType))

		// Write our document from the request body
		ctx, cancelFn := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancelFn()

		fBytes := make([]byte, fileHeader.Size)
		_, err = fileReader.Read(fBytes)
		if err != nil {
			logger.LogError("read failed", err)
			moovhttp.Problem(w, err)
			return
		}
		encryptedDoc, err := keeper.Encrypt(ctx, fBytes)
		if err != nil {
			logger.LogError("failed to encrypt document", err)
			moovhttp.Problem(w, fmt.Errorf("file upload error - %v", err))
			return
		}

		documentKey := makeDocumentKey(customerID, doc.DocumentID)
		logger.Log(fmt.Sprintf("writing %s", documentKey))

		err = bucket.WriteAll(ctx, documentKey, encryptedDoc, &blob.WriterOptions{
			ContentDisposition: "inline",
			ContentType:        contentType,
		})
		if err != nil {
			logger.LogError("problem uploading document", err)
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

		customerID, documentID := route.GetCustomerID(w, r), getDocumentID(w, r)
		if customerID == "" || documentID == "" {
			return
		}
		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		logger = logger.WithKeyValue("customerID", customerID).
			WithKeyValue("documentID", documentID)

		// reject the request if the document is deleted
		if exists, err := repo.exists(customerID, documentID, organization); !exists || err != nil {
			if err != nil {
				logger.LogError("failed to check document existance", err)
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
			logger.LogError("failed reading document from storage bucket", err)
			moovhttp.Problem(w, err)
			return
		}

		doc, err := keeper.Decrypt(ctx, encryptedDoc)
		if err != nil {
			logger.LogError("failed to decrypt document", err)
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

		customerID, documentID := route.GetCustomerID(w, r), getDocumentID(w, r)
		if customerID == "" || documentID == "" {
			return
		}

		logger = logger.WithKeyValue("customerID", customerID).
			WithKeyValue("documentID", documentID)

		err := repo.deleteCustomerDocument(customerID, documentID)
		if err != nil {
			logger.LogError("failed to delete document", err)
			moovhttp.Problem(w, fmt.Errorf("failed to %v", err))
			return
		}

		logger.Log("successfully deleted document")

		w.WriteHeader(http.StatusNoContent)
	}
}

func makeDocumentKey(customerID, documentID string) string {
	return path.Join("customers", customerID, "documents", documentID)
}
