// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/documents/storage"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestRouterGet(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	req := httptest.NewRequest("GET", "/configuration/customers", nil)
	req.Header.Set("X-Organization", "moov")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusOK)

	var response client.OrganizationConfiguration
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.LegalEntity == "" {
		t.Errorf("LegalEntity=%q", response.LegalEntity)
	}
	if response.PrimaryAccount == "" {
		t.Errorf("PrimaryAccount=%q", response.PrimaryAccount)
	}
}

func TestRouterGetErr(t *testing.T) {
	repo := &mockRepository{
		err: errors.New("bad error"),
	}

	req := httptest.NewRequest("GET", "/configuration/customers", nil)
	req.Header.Set("X-Organization", "moov")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRouterGetMissing(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	req := httptest.NewRequest("GET", "/configuration/customers", nil)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRouterUpdate(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(&client.OrganizationConfiguration{
		LegalEntity:    base.ID(),
		PrimaryAccount: base.ID(),
	})

	req := httptest.NewRequest("PUT", "/configuration/customers", &body)
	req.Header.Set("X-Organization", "moov")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusOK)

	var response client.OrganizationConfiguration
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.LegalEntity == "" {
		t.Errorf("LegalEntity=%q", response.LegalEntity)
	}
	if response.PrimaryAccount == "" {
		t.Errorf("PrimaryAccount=%q", response.PrimaryAccount)
	}
}

func TestRouterUpdateErr(t *testing.T) {
	repo := &mockRepository{
		err: errors.New("bad error"),
	}

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(&client.OrganizationConfiguration{
		LegalEntity:    base.ID(),
		PrimaryAccount: base.ID(),
	})

	req := httptest.NewRequest("PUT", "/configuration/customers", &body)
	req.Header.Set("X-Organization", "moov")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRouterUpdateMissing(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(&client.OrganizationConfiguration{
		LegalEntity:    base.ID(),
		PrimaryAccount: base.ID(),
	})

	req := httptest.NewRequest("PUT", "/configuration/customers", &body)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRouterGetOrganizationLogo_missingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/configuration/logo", nil)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, &mockRepository{}, nil)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, "missing X-Organization header", response["error"])
}

func TestRouterGetOrganizationLogo_noLogoUploaded(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/configuration/logo", nil)
	req.Header.Add("x-organization", "orgID")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, &mockRepository{}, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouterUploadAndGetLogo(t *testing.T) {
	w := httptest.NewRecorder()
	req := multipartRequest(t, "file", "image.png")
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "moov")

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, &mockRepository{}, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusNoContent, w.Result().StatusCode)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/configuration/logo", nil)
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "moov")

	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Result().StatusCode)
	require.Equal(t, "image/png", w.Header().Get("Content-Type"))
}

func TestRouterUploadLogo_missingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, &mockRepository{}, nil)
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/configuration/logo", nil))
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, "missing X-Organization header", response["error"])
}

func TestRouterUploadLogo_missingFile(t *testing.T) {
	w := httptest.NewRecorder()
	req := multipartRequest(t, "foo", "image.png")
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "moov")

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, &mockRepository{}, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, errMissingFile.Error(), response["error"])
}

func TestRouterUploadLogo_unsupportedFileType(t *testing.T) {
	w := httptest.NewRecorder()
	req := multipartRequest(t, "file", "bogus.txt")
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "moov")

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, &mockRepository{}, storage.NewTestBucket(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, errUnsupportedType.Error(), response["error"])
}

func multipartRequest(t *testing.T, fieldName string, fileName string) *http.Request {
	fd, err := os.Open(filepath.Join("testdata", fileName))
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(fieldName, fd.Name())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(part, fd); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PUT", "/configuration/logo", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}
