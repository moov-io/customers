// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moov-io/base"
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type testDocumentRepository struct {
	documents []*client.Document
	err       error

	written *client.Document
}

func (r *testDocumentRepository) getCustomerDocuments(customerId string) ([]*client.Document, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.documents, nil
}

func (r *testDocumentRepository) writeCustomerDocument(customerId string, doc *client.Document) error {
	r.written = doc
	return r.err
}

func TestDocuments__getDocumentId(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)

	if id := getDocumentId(w, req); id != "" {
		t.Errorf("unexpected id: %v", id)
	}
}

func TestDocuments__getCustomerDocuments(t *testing.T) {
	repo := &testDocumentRepository{}
	repo.err = errors.New("bad error")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo/documents", nil)

	router := mux.NewRouter()
	addDocumentRoutes(log.NewNopLogger(), router, repo, testBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}

	// reset error and try again
	repo.err = nil
	repo.documents = []*client.Document{
		{
			Id:   base.ID(),
			Type: "DriversLicense",
		},
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestDocuments__readDocumentType(t *testing.T) {
	if v, err := readDocumentType(" "); err == nil {
		t.Errorf("expected error v=%q", v)
	}
	if v, err := readDocumentType("other"); err == nil {
		t.Errorf("expected error v=%q", v)
	}
	if v, err := readDocumentType("DriversLicense"); err != nil {
		t.Errorf("expected no error v=%q: %v", v, err)
	}
	if v, err := readDocumentType("PASSPORT"); err != nil {
		t.Errorf("expected no error v=%q: %v", v, err)
	}
	if v, err := readDocumentType("utilitybill"); err != nil {
		t.Errorf("expected no error v=%q: %v", v, err)
	}
	if v, err := readDocumentType("BankSTATEMENT"); err != nil {
		t.Errorf("expected no error v=%q: %v", v, err)
	}
}

func multipartRequest(t *testing.T) *http.Request {
	fd, err := os.Open(filepath.Join("..", "..", "testdata", "colorado.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", fd.Name())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(part, fd); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Not *actually* a drivers license photo...
	req, err := http.NewRequest("POST", "/customers/foo/documents?type=DriversLicense", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestDocumentsUploadAndRetrieval(t *testing.T) {
	repo := &testDocumentRepository{}

	w := httptest.NewRecorder()
	req := multipartRequest(t)

	router := mux.NewRouter()
	addDocumentRoutes(log.NewNopLogger(), router, repo, testBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var doc client.Document
	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	if doc.Id == "" {
		t.Fatal("failed to read document")
	}
	if doc.ContentType != "image/jpeg" {
		t.Errorf("unknown content type: %s", doc.ContentType)
	}

	// Test the HTTP retrieval route
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/customers/foo/documents/%s", doc.Id), nil)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusFound {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
	if loc := w.Header().Get("Location"); !strings.Contains(loc, makeDocumentKey("foo", doc.Id)) {
		t.Errorf("unexpected SignedURL: %s", loc)
	}
}

func TestDocuments__uploadCustomerDocument(t *testing.T) {
	repo := &testDocumentRepository{}

	u, err := url.Parse("/customers/foo/documents?type=other")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := multipartRequest(t)
	req.URL = u // replace query params with invalid values

	router := mux.NewRouter()
	addDocumentRoutes(log.NewNopLogger(), router, repo, testBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestDocuments__makeDocumentKey(t *testing.T) {
	if v := makeDocumentKey("a", "b"); v != "customer-a-document-b" {
		t.Errorf("got %q", v)
	}
}

func TestDocumentRepository(t *testing.T) {
	db, err := createTestSqliteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()

	customerId := base.ID()
	repo := &sqliteDocumentRepository{db.db}
	defer repo.close()

	docs, err := repo.getCustomerDocuments(customerId)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Errorf("got %d unexpected documents: %#v", len(docs), docs)
	}

	// Write a Document and read it back
	doc := &client.Document{
		Id:          base.ID(),
		Type:        "DriversLicense",
		ContentType: "image/png",
	}
	if err := repo.writeCustomerDocument(customerId, doc); err != nil {
		t.Fatal(err)
	}
	docs, err = repo.getCustomerDocuments(customerId)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Errorf("got %d unexpected documents: %#v", len(docs), docs)
	}
	if docs[0].Id != doc.Id {
		t.Errorf("docs[0].Id=%s doc.Id=%s", docs[0].Id, doc.Id)
	}
}
