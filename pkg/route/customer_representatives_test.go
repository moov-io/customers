// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package route

import (
	"net/http/httptest"
	"testing"
)

func TestCustomers__GetRepresentativeID(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)

	if id := GetRepresentativeID(w, req); id != "" {
		t.Errorf("unexpected id: %v", id)
	}
}
