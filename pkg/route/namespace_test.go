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

	if ns := GetNamespace(w, req); ns != "foo" {
		t.Errorf("unexpected ns: %v", ns)
	}
}

func TestRoute__GetNamespaceMissing(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)

	if ns := GetNamespace(w, req); ns != "" {
		t.Errorf("unexpected ns: %v", ns)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}
