// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package fed

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/base/log"
)

func TestFED(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ping" {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`PONG`))
	}))

	client := NewClient(log.NewNopLogger(), svc.URL, false)
	if err := client.Ping(); err != nil {
		t.Fatal(err)
	}
	svc.Close()

	// test LookupInstitution
	svc = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fed/ach/search" {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"achParticipants": [{"routingNumber": "121042882"}]}`)) // partial fed.AchDictionary response
	}))

	client = NewClient(log.NewNopLogger(), svc.URL, false)
	if details, err := client.LookupInstitution("121042882"); err != nil {
		t.Fatal(err)
	} else {
		if details.RoutingNumber != "121042882" {
			t.Errorf("unexpected ACH details: %#v", details)
		}
	}
	svc.Close()
}

func TestFED__NewClient(t *testing.T) {
	client := NewClient(log.NewNopLogger(), "", false)
	if client == nil {
		t.Error("expected FED client")
	}
}
