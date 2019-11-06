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

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/moov-io/base"
	client "github.com/moov-io/customers/client"
	ofac "github.com/moov-io/ofac/client"
)

func createTestOFACSearcher(repo customerRepository, ofacClient OFACClient) *ofacSearcher {
	if repo == nil {
		repo = &testCustomerRepository{}
	}
	if ofacClient == nil {
		ofacClient = &testOFACClient{}
	}
	return &ofacSearcher{repo: repo, ofacClient: ofacClient}
}

func TestOFACSearcher__nil(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	ofacClient := &testOFACClient{}
	searcher := createTestOFACSearcher(repo, ofacClient)

	if err := searcher.storeCustomerOFACSearch(nil, "requestID"); err == nil {
		t.Error("expected error")
	}
}

func TestOFACSearcher__storeCustomerOFACSearch(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	ofacClient := &testOFACClient{}
	searcher := createTestOFACSearcher(repo, ofacClient)

	ofacClient.sdn = &ofac.Sdn{
		EntityID: "1241421",
		SdnName:  "Jane Doe",
		Match:    0.99,
	}
	customerID := base.ID()
	if err := searcher.storeCustomerOFACSearch(&client.Customer{ID: customerID}, "requestID"); err != nil {
		t.Fatal(err)
	}
	res, err := repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res.EntityId != "1241421" {
		t.Errorf("ofacSearchResult: %#v", res)
	}
	if res.CreatedAt.IsZero() {
		t.Errorf("res.CreatedAt=%v", res.CreatedAt)
	}

	// retry but with NickName set (test coverage)
	customerID = base.ID()
	if err := searcher.storeCustomerOFACSearch(&client.Customer{ID: customerID, NickName: "John Doe"}, "requestID"); err != nil {
		t.Fatal(err)
	}
}

func TestOFACApproval__getLatest(t *testing.T) {
	logger := log.NewNopLogger()
	router := mux.NewRouter()

	customerID := base.ID()

	repo := &testCustomerRepository{
		customer: &client.Customer{
			ID: customerID,
		},
		savedOFACSearchResult: &ofacSearchResult{
			EntityId: "142",
			Match:    1.0,
		},
	}

	addOFACRoutes(logger, router, repo, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/customers/%s/ofac", customerID), nil)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	// error case
	repo.err = errors.New("bad error")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestOFACApproval__refresh(t *testing.T) {
	logger := log.NewNopLogger()
	router := mux.NewRouter()

	customerID := base.ID()

	repo := &testCustomerRepository{
		customer: &client.Customer{
			ID: customerID,
		},
		savedOFACSearchResult: &ofacSearchResult{
			EntityId: "142",
			Match:    1.0,
		},
	}
	testOFACClient := &testOFACClient{}
	ofac := &ofacSearcher{
		repo:       repo,
		ofacClient: testOFACClient,
	}

	addOFACRoutes(logger, router, repo, ofac)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/refresh/ofac", customerID), nil)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
	var result ofacSearchResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.EntityId != "142" {
		t.Errorf("result=%#v", result)
	}

	repo.savedOFACSearchResult.Match = 0.90 // match isn't high enough to block

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestOFACApproval__refreshErr(t *testing.T) {
	logger := log.NewNopLogger()
	router := mux.NewRouter()

	customerID := base.ID()

	repo := &testCustomerRepository{
		customer: &client.Customer{
			ID: customerID,
		},
		savedOFACSearchResult: &ofacSearchResult{
			EntityId: "142",
			Match:    0.88,
		},
	}
	testOFACClient := &testOFACClient{err: errors.New("bad error")}
	ofac := &ofacSearcher{
		repo:       repo,
		ofacClient: testOFACClient,
	}

	addOFACRoutes(logger, router, repo, ofac)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/refresh/ofac", customerID), nil)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d - %s", w.Code, w.Body.String())
	}
}
