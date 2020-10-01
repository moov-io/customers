// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestRouterGet(t *testing.T) {
	org := "moov"
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
			LogoFile:       fmt.Sprintf("%s-logo.jpg", org),
		},
	}

	req := httptest.NewRequest("GET", "/configuration/customers", nil)
	req.Header.Set("X-Organization", org)
	w := httptest.NewRecorder()

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
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
	if response.LogoFile == "" {
		t.Errorf("LogoFile=%q", response.LogoFile)
	}
}

func TestRouterGetErr(t *testing.T) {
	repo := &mockRepository{
		err: errors.New("bad error"),
	}

	req := httptest.NewRequest("GET", "/configuration/customers", nil)
	req.Header.Set("X-Organization", "moov")
	w := httptest.NewRecorder()

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
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

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRouterUpdate(t *testing.T) {
	org := "moov"
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
			LogoFile:       fmt.Sprintf("%s-logo.jpg", org),
		},
	}

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(&client.OrganizationConfiguration{
		LegalEntity:    base.ID(),
		PrimaryAccount: base.ID(),
	})

	req := httptest.NewRequest("PUT", "/configuration/customers", &body)
	req.Header.Set("X-Organization", org)
	w := httptest.NewRecorder()

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
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
	if response.LogoFile == "" {
		t.Errorf("LogoFile=%q", response.LogoFile)
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

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
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

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
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
	require.Equal(t, response["error"], "missing X-Organization header")
}

func TestRouterGetOrganizationLogo_repoError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/configuration/logo", nil)
	req.Header.Add("x-organization", "orgID")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, &mockRepository{err: errors.New("real bad error")}, nil)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, response["error"], "real bad error")
}

func TestRouterGetOrganizationLogo_noLogoUploaded(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/configuration/logo", nil)
	req.Header.Add("x-organization", "orgID")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, &mockRepository{
		cfg: &client.OrganizationConfiguration{},
	}, nil)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, response["error"], "no logo uploaded for organization orgID")
}

func TestRouterGetOrganizationLogo_missingFile(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/configuration/logo", nil)
	req.Header.Add("x-organization", "orgID")
	w := httptest.NewRecorder()

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, &mockRepository{
		cfg: &client.OrganizationConfiguration{LogoFile: "some-file.png"},
	}, bucketFunc)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Contains(t, response["error"], "file not found")
}

func TestRouterUploadRetrieveLogo(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	w := httptest.NewRecorder()
	req := multipartRequest(t, "file", "moov.jpg")
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "moov")

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusCreated, w.Result().StatusCode)
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
	if response.LogoFile == "" {
		t.Errorf("LogoFile=%q", response.LogoFile)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/configuration/logo", nil)
	req.Header.Add("x-organization", "moov")
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Result().StatusCode, "failed to retrieve logo file")
}

func TestRouterUpdateLogo(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
			LogoFile:       "organization-moov-logo.jpg",
		},
	}

	w := httptest.NewRecorder()
	req := multipartRequest(t, "file", "moov.jpg")
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "moov")

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestRouterUploadLogo_missingHeader(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	w := httptest.NewRecorder()
	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/configuration/logo", nil))
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, response["error"], "missing X-Organization header")
}

func TestRouterUploadLogo_missingFile(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	w := httptest.NewRecorder()
	req := multipartRequest(t, "foo", "moov.jpg")
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "moov")

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, response["error"], errMissingFile.Error())
}

func TestRouterUploadLogo_unsupportedFileType(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.OrganizationConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	w := httptest.NewRecorder()
	req := multipartRequest(t, "file", "bogus.txt")
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "moov")

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, response["error"], errUnsupportedType.Error())
}

func TestRouterUploadLogo_repoError(t *testing.T) {
	repo := &mockRepository{
		err: errors.New("scary DB error"),
	}

	w := httptest.NewRecorder()
	req := multipartRequest(t, "file", "moov.jpg")
	req.Header.Set("x-request-id", "test")
	req.Header.Set("X-organization", "moov")

	tempDir, bucketFunc := storage.NewTestBucket()
	defer os.RemoveAll(tempDir)

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo, bucketFunc)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
	response := make(map[string]string)
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	require.Contains(t, response, "error")
	require.Equal(t, response["error"], "scary DB error")
}

func TestRouterGetLogo(t *testing.T) {

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
