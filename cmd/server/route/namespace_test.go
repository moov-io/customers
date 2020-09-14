// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package route

import (
	"net/http/httptest"
	"testing"
)

func TestRoute__GetNamespace(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("x-namespace", "foo")

	if ns := GetNamespace(w, req); ns != "foo" {
		t.Errorf("unexpected ns: %v", ns)
	}
}
