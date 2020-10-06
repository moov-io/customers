// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/base"
	"github.com/moov-io/base/log"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/watchman"
	watchmanClient "github.com/moov-io/watchman/client"
)

func createTestOFACSearcher(repo CustomerRepository, client watchman.Client) *OFACSearcher {
	if repo == nil {
		repo = &testCustomerRepository{}
	}
	if client == nil {
		client = &watchman.TestWatchmanClient{}
	}
	return &OFACSearcher{repo: repo, watchmanClient: client}
}

func TestOFACSearcher__nil(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	ofacClient := &watchman.TestWatchmanClient{}
	searcher := createTestOFACSearcher(repo, ofacClient)

	if err := searcher.storeCustomerOFACSearch(nil, "requestID"); err == nil {
		t.Error("expected error")
	}
}

func TestOFACSearcher__storeCustomerOFACSearch(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	ofacClient := watchman.NewTestWatchmanClient(&watchmanClient.OfacSdn{
		EntityID: "1241421",
		SdnName:  "Jane Doe",
		Match:    0.99,
	}, nil)
	searcher := createTestOFACSearcher(repo, ofacClient)

	customerID := base.ID()
	if err := searcher.storeCustomerOFACSearch(&client.Customer{CustomerID: customerID}, "requestID"); err != nil {
		t.Fatal(err)
	}
	res, err := repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res.EntityID != "1241421" {
		t.Errorf("ofac search: %#v", res)
	}
	if res.CreatedAt.IsZero() {
		t.Errorf("res.CreatedAt=%v", res.CreatedAt)
	}

	// retry but with NickName set (test coverage)
	customerID = base.ID()
	if err := searcher.storeCustomerOFACSearch(&client.Customer{CustomerID: customerID, NickName: "John Doe"}, "requestID"); err != nil {
		t.Fatal(err)
	}
}

func TestOFACApproval__getLatest(t *testing.T) {
	logger := log.NewNopLogger()
	router := mux.NewRouter()

	customerID := base.ID()

	repo := &testCustomerRepository{
		customer: &client.Customer{
			CustomerID: customerID,
		},
		savedSearchResult: &client.OfacSearch{
			EntityID: "142",
			Match:    1.0,
		},
	}

	AddOFACRoutes(logger, router, repo, nil)

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
			CustomerID: customerID,
		},
		savedSearchResult: &client.OfacSearch{
			EntityID: "142",
			Match:    1.0,
		},
	}
	testWatchmanClient := watchman.NewTestWatchmanClient(nil, nil)
	ofac := &OFACSearcher{
		repo:           repo,
		watchmanClient: testWatchmanClient,
	}

	AddOFACRoutes(logger, router, repo, ofac)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/refresh/ofac", customerID), nil)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
	var result client.OfacSearch
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.EntityID != "142" {
		t.Errorf("result=%#v", result)
	}

	repo.savedSearchResult.Match = 0.90 // match isn't high enough to block

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
			CustomerID: customerID,
		},
		savedSearchResult: &client.OfacSearch{
			EntityID: "142",
			Match:    0.88,
		},
	}
	testWatchmanClient := watchman.NewTestWatchmanClient(nil, errors.New("bad error"))
	ofac := &OFACSearcher{
		repo:           repo,
		watchmanClient: testWatchmanClient,
	}

	AddOFACRoutes(logger, router, repo, ofac)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/refresh/ofac", customerID), nil)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d - %s", w.Code, w.Body.String())
	}
}
