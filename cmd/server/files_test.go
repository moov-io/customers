// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moov-io/base"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"
)

func TestFiles__proxyLocalFile(t *testing.T) {
	// first, upload a file
	bucket, err := testBucket()
	if err != nil {
		t.Fatal(err)
	}

	customerID, documentID := base.ID(), base.ID()
	documentKey := makeDocumentKey(customerID, documentID)

	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	writer, err := bucket.NewWriter(ctx, documentKey, &blob.WriterOptions{
		ContentDisposition: "inline",
		ContentType:        "image/jpg",
	})
	if err != nil {
		t.Fatal(err)
	}
	fd, err := os.Open(filepath.Join("..", "..", "testdata", "colorado.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if n, err := io.Copy(writer, fd); err != nil || n == 0 {
		t.Fatalf("n=%d error=%v", n, err)
	}
	writer.Close()

	// Grab our SignedURL
	signedURL, err := bucket.SignedURL(ctx, documentKey, &blob.SignedURLOptions{
		Expiry: 15 * time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(signedURL)
	if u == nil || err != nil {
		t.Fatalf("u=%s err=%v", u, err)
	}
	u.Scheme, u.Host = "", "" // blank out 'http://' and 'localhost' in below assumption
	signer, err := fileblobSigner("http://localhost", "secret")
	if err != nil {
		t.Fatal(err)
	}

	// Make our request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/files"+u.String(), nil)
	router := mux.NewRouter()
	addFileblobRoutes(log.NewNopLogger(), router, signer, func() (*blob.Bucket, error) { return bucket, nil })
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
	if v := w.Header().Get("Content-Length"); v != "212203" {
		t.Errorf("Content-Length: %s", v)
	}
	if v := w.Header().Get("Content-Type"); v != "image/jpg" {
		t.Errorf("Content-Type: %s", v)
	}
}
