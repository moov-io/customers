// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/pkg/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestRouterGet(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.NamespaceConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	req := httptest.NewRequest("GET", "/configuration/customers", nil)
	req.Header.Set("X-Organization", "moov")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusOK)

	var response client.NamespaceConfiguration
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
	RegisterRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRouterGetMissing(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.NamespaceConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	req := httptest.NewRequest("GET", "/configuration/customers", nil)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRouterUpdate(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.NamespaceConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(&client.NamespaceConfiguration{
		LegalEntity:    base.ID(),
		PrimaryAccount: base.ID(),
	})

	req := httptest.NewRequest("PUT", "/configuration/customers", &body)
	req.Header.Set("X-Organization", "moov")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusOK)

	var response client.NamespaceConfiguration
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
	json.NewEncoder(&body).Encode(&client.NamespaceConfiguration{
		LegalEntity:    base.ID(),
		PrimaryAccount: base.ID(),
	})

	req := httptest.NewRequest("PUT", "/configuration/customers", &body)
	req.Header.Set("X-Organization", "moov")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRouterUpdateMissing(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.NamespaceConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(&client.NamespaceConfiguration{
		LegalEntity:    base.ID(),
		PrimaryAccount: base.ID(),
	})

	req := httptest.NewRequest("PUT", "/configuration/customers", &body)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, w.Code, http.StatusBadRequest)
}
