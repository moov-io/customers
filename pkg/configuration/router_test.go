// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/moov-io/base"
	"github.com/moov-io/customers/pkg/client"
)

func TestRouterGet(t *testing.T) {
	repo := &mockRepository{
		cfg: &client.NamespaceConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
		},
	}

	req := httptest.NewRequest("GET", "/configuration/customers", nil)
	req.Header.Set("X-Namespace", "moov")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %v", w.Code)
	}

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
	req.Header.Set("X-Namespace", "moov")
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %v", w.Code)
	}

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
