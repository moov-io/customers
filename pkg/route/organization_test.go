// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoute__GetNamespace(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("x-organization", "foo")

	if org := GetOrganization(w, req); org != "foo" {
		t.Errorf("unexpected ns: %v", org)
	}
}

func TestRoute__GetNamespaceMissing(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)

	if org := GetOrganization(w, req); org != "" {
		t.Errorf("unexpected ns: %v", org)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}
