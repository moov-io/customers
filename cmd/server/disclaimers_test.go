// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/base"
	client "github.com/moov-io/customers/client"
	"github.com/moov-io/customers/internal/database"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type testDisclaimerRepository struct {
	disclaimers []*client.Disclaimer
	err         error
}

func (r *testDisclaimerRepository) getCustomerDisclaimers(customerID string) ([]*client.Disclaimer, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.disclaimers, nil
}

func (r *testDisclaimerRepository) acceptDisclaimer(customerID, disclaimerID string) error {
	return r.err
}

func TestDisclaimers__getDisclaimerID(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("x-request-id", "test")

	if id := getDisclaimerID(w, req); id != "" {
		t.Errorf("unexpected ID: %s", id)
	}
}

func TestDisclaimers__getCustomerDisclaimers(t *testing.T) {
	repo := &testDisclaimerRepository{
		disclaimers: []*client.Disclaimer{
			{
				ID:   base.ID(),
				Text: "terms and conditions",
			},
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/adam/disclaimers", nil)
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	addDisclaimerRoute(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	var resp []*client.Disclaimer
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp) != 1 {
		t.Errorf("disclaimers: %#v", resp)
	}

	// set an error and verify
	repo.err = errors.New("bad error")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestDisclaimers__acceptDisclaimer(t *testing.T) {
	disclaimerID := base.ID()
	repo := &testDisclaimerRepository{
		disclaimers: []*client.Disclaimer{
			{
				ID:   disclaimerID,
				Text: "terms and conditions",
			},
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/adam/disclaimers/%s", disclaimerID), nil)
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	addDisclaimerRoute(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	// set an error and verify
	repo.err = errors.New("bad error")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestDisclaimersRepository(t *testing.T) {
	customerID := base.ID()

	check := func(t *testing.T, repo *sqlDisclaimerRepository) {
		defer repo.close()

		disclaimers, err := repo.getCustomerDisclaimers(customerID)
		if err != nil {
			t.Fatal(err)
		}
		if len(disclaimers) != 0 {
			t.Errorf("unexpected disclaimers: %#v", disclaimers)
		}

		// write a Disclaimer and verify
		disc, err := repo.insertDisclaimer("terms and conditions")
		if err != nil {
			t.Fatal(err)
		}
		disclaimers, err = repo.getCustomerDisclaimers(customerID)
		if err != nil {
			t.Fatal(err)
		}
		if len(disclaimers) != 1 {
			t.Errorf("unexpected disclaimers: %#v", disclaimers)
		}

		// Accept the disclaimer
		if err := repo.acceptDisclaimer(customerID, disc.ID); err != nil {
			t.Fatal(err)
		}

		// Verify a different disclaimer ID is rejected
		if err := repo.acceptDisclaimer(customerID, base.ID()); err == nil {
			t.Error("expected error")
		}
	}

	// SQLite tests
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, &sqlDisclaimerRepository{sqliteDB.DB, log.NewNopLogger()})

	// MySQL tests
	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, &sqlDisclaimerRepository{mysqlDB.DB, log.NewNopLogger()})
}
