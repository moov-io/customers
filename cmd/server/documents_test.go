// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/base"
	client "github.com/moov-io/customers/client"

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

func TestDocuments__getCustomerDocuments(t *testing.T) {
	repo := &testDocumentRepository{}
	repo.err = errors.New("bad error")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo/documents", nil)

	router := mux.NewRouter()
	addDocumentRoutes(nil, router, repo, testBucket)
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

func TestDocuments__uploadCustomerDocument(t *testing.T) {
	// TODO(adam):
}

func TestDocuments__retrieveRawDocument(t *testing.T) {
	// TODO(adam):
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
