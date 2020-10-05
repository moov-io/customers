// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package documents

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moov-io/base"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"

	"github.com/moov-io/customers/internal/database"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/documents/storage"
	"github.com/moov-io/customers/pkg/secrets"

	"github.com/moov-io/base/log"
	"github.com/gorilla/mux"
)

type testDocumentRepository struct {
	documents []*client.Document
	err       error
	docExists bool
	written   *client.Document
}

func (r *testDocumentRepository) exists(customerID string, documentID string, organization string) (bool, error) {
	if r.docExists {
		return true, nil
	}
	return false, r.err
}

func (r *testDocumentRepository) getCustomerDocuments(customerID string, organization string) ([]*client.Document, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.documents, nil
}

func (r *testDocumentRepository) writeCustomerDocument(customerID string, doc *client.Document) error {
	r.written = doc
	return r.err
}

func (r *testDocumentRepository) deleteCustomerDocument(customerID string, documentID string) error {
	r.written = nil
	return r.err
}

func TestDocuments__getDocumentID(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("x-request-id", "test")

	if id := getDocumentID(w, req); id != "" {
		t.Errorf("unexpected id: %v", id)
	}
}

func TestDocuments__getCustomerDocuments(t *testing.T) {
	repo := &testDocumentRepository{}
	repo.err = errors.New("bad error")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo/documents", nil)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("x-organization", "test")

	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, repo, secrets.TestKeeper(t), storage.TestBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}

	// reset error and try again
	repo.err = nil
	repo.documents = []*client.Document{
		{
			DocumentID: base.ID(),
			Type:       "DriversLicense",
		},
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d\n%s", w.Code, w.Body.String())
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

func TestDocumentsUploadAndRetrieval(t *testing.T) {
	repo := &testDocumentRepository{docExists: true}

	w := httptest.NewRecorder()
	req := multipartRequest(t)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")

	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, repo, secrets.TestKeeper(t), storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var doc client.Document
	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	if doc.DocumentID == "" {
		t.Fatal("failed to read document")
	}
	if doc.ContentType != "image/jpeg" {
		t.Errorf("unknown content type: %s", doc.ContentType)
	}

	// Test the HTTP retrieval route
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/customers/foo/documents/%s", doc.DocumentID), nil)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)
}

func TestDocumentsUpload_fileTooLarge(t *testing.T) {
	req := multipartFileOfSize(t, "file", int64(maxDocumentSize)+512)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")

	w := httptest.NewRecorder()
	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, &testDocumentRepository{}, secrets.TestKeeper(t), storage.TestBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)
	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Contains(t, string(b), "file exceeds maximum size of 20MB")
}

// bufio Peek(n) returns an EOF error if the file is smaller than n bytes
func TestDocumentsUpload_fileSmallerThanPeek(t *testing.T) {
	req := multipartFileOfSize(t, "file", 256)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")

	w := httptest.NewRecorder()
	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, &testDocumentRepository{}, secrets.TestKeeper(t), storage.TestBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)
}

func TestDocumentsUpload_formTooLarge(t *testing.T) {
	req := multipartFileOfSize(t, "file", int64(maxFormSize)+512)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")

	w := httptest.NewRecorder()
	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, &testDocumentRepository{}, secrets.TestKeeper(t), storage.TestBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)
	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Contains(t, string(b), "request body exceeds maximum size of 25MB")
}

func TestDocumentsUpload_missingFileKey(t *testing.T) {
	req := multipartFileOfSize(t, "bogusKey", 10)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")

	w := httptest.NewRecorder()
	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, &testDocumentRepository{}, secrets.TestKeeper(t), storage.TestBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)
	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Contains(t, string(b), "expected multipart upload with key of 'file'")
}

func TestDocumentsUpload_repoErr(t *testing.T) {
	repo := &testDocumentRepository{err: errors.New("real bad error")}
	req := multipartFileOfSize(t, "file", 10)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")

	w := httptest.NewRecorder()
	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, repo, secrets.TestKeeper(t), storage.TestBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)
	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Contains(t, string(b), "real bad error")
}

func TestDocumentsUpload_keeperErr(t *testing.T) {
	keeper := secrets.TestKeeper(t)
	keeper.Close()
	req := multipartFileOfSize(t, "file", 10)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")

	w := httptest.NewRecorder()
	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, &testDocumentRepository{}, keeper, storage.TestBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)
	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Contains(t, string(b), "Keeper has been closed")
}

func TestDocumentsUpload_BucketErr(t *testing.T) {
	keeper := secrets.TestKeeper(t)
	bucketFunc := func() (*blob.Bucket, error) {
		return nil, errors.New("bucket error")
	}
	req := multipartFileOfSize(t, "file", 10)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")

	w := httptest.NewRecorder()
	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, &testDocumentRepository{}, keeper, bucketFunc)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)
	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Contains(t, string(b), "bucket error")
}

func multipartRequest(t *testing.T) *http.Request {
	fd, err := os.Open(filepath.Join("testdata", "colorado.jpg"))
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

func multipartFileOfSize(t *testing.T, fileKey string, size int64) *http.Request {
	var body bytes.Buffer
	mp := multipart.NewWriter(&body)
	part, err := mp.CreateFormFile(fileKey, "exceedinglyLargeFile")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(make([]byte, size)); err != nil {
		t.Fatal(err)
	}
	if err := mp.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/customers/abc/documents?type=driverslicense", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", mp.FormDataContentType())

	return req
}

func TestDocuments__retrieveError(t *testing.T) {
	repo := &testDocumentRepository{
		err: errors.New("bad error"),
	}

	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, repo, secrets.TestKeeper(t), storage.TestBucket)

	customerID, documentID := base.ID(), base.ID()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/customers/%s/documents/%s", customerID, documentID), nil)
	req.Header.Set("X-Organization", base.ID())
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestDocumentsRetrieve_bucketErr(t *testing.T) {
	repo := &testDocumentRepository{
		docExists: true,
	}
	keeper := secrets.TestKeeper(t)
	bucketFunc := func() (*blob.Bucket, error) { return nil, errors.New("bucket error") }
	customerID, documentID := base.ID(), base.ID()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/customers/%s/documents/%s", customerID, documentID), nil)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "test")
	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, repo, keeper, bucketFunc)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)
	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Contains(t, string(b), "bucket error")
}

func TestDocuments__delete(t *testing.T) {
	db := database.CreateTestSqliteDB(t)
	repo := &sqlDocumentRepository{db.DB, log.NewNopLogger()}

	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, repo, secrets.TestKeeper(t), storage.TestBucket)

	customerID := base.ID()
	// create document
	doc := &client.Document{
		DocumentID:  base.ID(),
		Type:        "DriversLicense",
		ContentType: "image/png",
		UploadedAt:  time.Now(),
	}
	err := repo.writeCustomerDocument(customerID, doc)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/customers/%s/documents/%s", customerID, doc.DocumentID), nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)

	var count int
	row := db.DB.QueryRow("SELECT COUNT(*) FROM documents where deleted_at is null")
	require.NoError(t, row.Scan(&count))
	require.Equal(t, 0, count)
}

func TestDocuments__uploadCustomerDocument(t *testing.T) {
	repo := &testDocumentRepository{}

	u, err := url.Parse("/customers/foo/documents?type=other")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := multipartRequest(t)
	req.Header.Set("x-request-id", "test")
	req.URL = u // replace query params with invalid values

	router := mux.NewRouter()
	AddDocumentRoutes(log.NewNopLogger(), router, repo, secrets.TestKeeper(t), storage.TestBucket)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestDocuments__makeDocumentKey(t *testing.T) {
	key := makeDocumentKey("a", "b")

	if key != "customers/a/documents/b" {
		t.Errorf("got %q", key)
	}
}

func TestDocuments__sizeLimit(t *testing.T) {
	lim := sizeLimit(100 << 20)

	require.Equal(t, "100MB", lim.String())
}
