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

	"github.com/moov-io/base"
	client "github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/identity/pkg/logging"
	"github.com/moov-io/identity/pkg/stime"
	tmwt "github.com/moov-io/tumbler/pkg/middleware/middlewaretest"
	watchman "github.com/moov-io/watchman/client"

	"github.com/gorilla/mux"
)

func createTestOFACSearcher(t *testing.T, repo customerRepository, watchmanClient WatchmanClient) *ofacSearcher {
	t.Helper()
	if repo == nil {
		repo = &testCustomerRepository{}
	}
	if watchmanClient == nil {
		watchmanClient = &testWatchmanClient{}
	}
	return &ofacSearcher{repo: repo, watchmanClient: watchmanClient}
}

func TestOFACSearcher__nil(t *testing.T) {
	repo := createTestCustomerRepository(t)

	ofacClient := &testWatchmanClient{}
	searcher := createTestOFACSearcher(t, repo, ofacClient)

	if err := searcher.storeCustomerOFACSearch(nil); err == nil {
		t.Error("expected error")
	}
}

func TestOFACSearcher__storeCustomerOFACSearch(t *testing.T) {
	repo := createTestCustomerRepository(t)

	ofacClient := &testWatchmanClient{}
	searcher := createTestOFACSearcher(t, repo, ofacClient)

	ofacClient.sdn = &watchman.OfacSdn{
		EntityID: "1241421",
		SdnName:  "Jane Doe",
		Match:    0.99,
	}
	customerID := base.ID()
	if err := searcher.storeCustomerOFACSearch(&client.Customer{CustomerID: customerID}); err != nil {
		t.Fatal(err)
	}
	res, err := repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res.EntityID != "1241421" {
		t.Errorf("ofacSearchResult: %#v", res)
	}
	if res.CreatedAt.IsZero() {
		t.Errorf("res.CreatedAt=%v", res.CreatedAt)
	}

	// retry but with NickName set (test coverage)
	customerID = base.ID()
	if err := searcher.storeCustomerOFACSearch(&client.Customer{CustomerID: customerID, NickName: "John Doe"}); err != nil {
		t.Fatal(err)
	}
}

func TestOFACApproval__getLatest(t *testing.T) {
	logger := logging.NewNopLogger()
	router := mux.NewRouter()

	session := tmwt.NewRandomClaims()
	times := stime.NewStaticTimeService()

	testMiddleware := tmwt.NewTestMiddleware(times, session)
	router.Use(testMiddleware.Handler)

	customerID := base.ID()

	repo := &testCustomerRepository{
		customer: &client.Customer{
			CustomerID: customerID,
		},
		savedOFACSearchResult: &ofacSearchResult{
			EntityID: "142",
			Match:    1.0,
		},
	}

	NewCustomersController(logger, nil, repo, nil).AppendRoutes(router)

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

	if w.Code != http.StatusInternalServerError {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestOFACApproval__refresh(t *testing.T) {
	logger := logging.NewNopLogger()
	router := mux.NewRouter()

	session := tmwt.NewRandomClaims()
	times := stime.NewStaticTimeService()

	testMiddleware := tmwt.NewTestMiddleware(times, session)
	router.Use(testMiddleware.Handler)

	customerID := base.ID()

	repo := &testCustomerRepository{
		customer: &client.Customer{
			CustomerID: customerID,
		},
		savedOFACSearchResult: &ofacSearchResult{
			EntityID: "142",
			Match:    1.0,
		},
	}
	testWatchmanClient := &testWatchmanClient{}
	ofac := &ofacSearcher{
		repo:           repo,
		watchmanClient: testWatchmanClient,
	}

	NewCustomersController(logger, ofac, repo, nil).AppendRoutes(router)

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
	if result.EntityID != "142" {
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
	logger := logging.NewNopLogger()
	router := mux.NewRouter()

	session := tmwt.NewRandomClaims()
	times := stime.NewStaticTimeService()

	testMiddleware := tmwt.NewTestMiddleware(times, session)
	router.Use(testMiddleware.Handler)

	customerID := base.ID()

	repo := &testCustomerRepository{
		customer: &client.Customer{
			CustomerID: customerID,
		},
		savedOFACSearchResult: &ofacSearchResult{
			EntityID: "142",
			Match:    0.88,
		},
	}
	testWatchmanClient := &testWatchmanClient{err: errors.New("bad error")}
	ofac := &ofacSearcher{
		repo:           repo,
		watchmanClient: testWatchmanClient,
	}

	NewCustomersController(logger, ofac, repo, nil).AppendRoutes(router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/refresh/ofac", customerID), nil)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusInternalServerError {
		t.Errorf("bogus HTTP status: %d - %s", w.Code, w.Body.String())
	}
}
