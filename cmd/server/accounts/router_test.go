// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	watchman "github.com/moov-io/watchman/client"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/cmd/server/fed"
	"github.com/moov-io/customers/cmd/server/paygate"
	"github.com/moov-io/customers/internal/testclient"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/secrets"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	testFedClient     = &fed.MockClient{}
	testPayGateClient = &paygate.MockClient{}
)

type testWatchmanClient struct {
	// error to be returned instead of field from above
	err error
}

func (c *testWatchmanClient) Ping() error {
	return c.err
}

func (c *testWatchmanClient) Search(_ context.Context, name string, _ string) (*watchman.OfacSdn, error) {
	if c.err != nil {
		return nil, c.err
	}
	return &watchman.OfacSdn{
		EntityID: "123",
		SdnName:  "dummy_sdn",
		SdnType:  "dummy_sdn_type",
		Programs: nil,
		Title:    "dummy_title",
		Remarks:  "dummy_remarks",
		Match:    10.2,
	}, nil
}

func createTestOFACSearcher(repo Repository, watchmanClient WatchmanClient) *AccountOfacSearcher {
	if repo == nil {
		repo = &testAccountRepository{}
	}
	if watchmanClient == nil {
		watchmanClient = &testWatchmanClient{}
	}
	return &AccountOfacSearcher{Repo: repo, WatchmanClient: watchmanClient}
}

func TestAccountRoutes(t *testing.T) {
	customerID := base.ID()
	repo := setupTestAccountRepository(t)
	keeper := secrets.TestStringKeeper(t)

	handler := mux.NewRouter()
	RegisterRoutes(
		log.NewNopLogger(), handler, repo, testFedClient, testPayGateClient, keeper, keeper, createTestOFACSearcher(repo, nil))

	// first read, expect no accounts
	accounts := httpReadAccounts(t, handler, customerID)
	if len(accounts) != 0 {
		t.Errorf("got accounts: %v", accounts)
	}

	// create an account
	account := httpCreateAccount(t, handler, customerID)
	if account.MaskedAccountNumber != "*8749" {
		t.Errorf("masked account number: %q", account.MaskedAccountNumber)
	}

	// re-read, find account
	accounts = httpReadAccounts(t, handler, customerID)
	if len(accounts) != 1 || accounts[0].AccountID != account.AccountID {
		t.Errorf("got accounts: %v", accounts)
	}

	// delete and expect no accounts
	httpDeleteAccount(t, handler, customerID, account.AccountID)
	accounts = httpReadAccounts(t, handler, customerID)
	if len(accounts) != 0 {
		t.Errorf("got accounts: %v", accounts)
	}
}

func TestCreateAccountAndCheckAccountOfacSearch(t *testing.T) {
	customerID := base.ID()
	repo := setupTestAccountRepository(t)
	keeper := secrets.TestStringKeeper(t)

	handler := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), handler, repo, testFedClient, testPayGateClient, keeper, keeper, createTestOFACSearcher(repo, nil))

	// first read, expect no accounts
	accounts := httpReadAccounts(t, handler, customerID)
	if len(accounts) != 0 {
		t.Errorf("got accounts: %v", accounts)
	}

	// create an account
	account := httpCreateAccount(t, handler, customerID)
	if account.MaskedAccountNumber != "*8749" {
		t.Errorf("masked account number: %q", account.MaskedAccountNumber)
	}

	// Check for ofac search after creating
	accountOfacSearch := httpReadAccountOfacSearch(t, handler, customerID, account.AccountID)
	if accountOfacSearch == nil {
		t.Errorf("got account ofac search: %v", accountOfacSearch)
	}
}

func TestAccountCreationRequest(t *testing.T) {
	req := &createAccountRequest{}
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}

	req.HolderName = "John Doe"
	req.AccountNumber = "12345"
	req.RoutingNumber = "987654320"
	req.Type = client.SAVINGS

	if err := req.validate(); err != nil {
		t.Error(err)
	}
}

func TestRoutes__DecryptAccountNumber(t *testing.T) {
	customerID := base.ID()
	repo := setupTestAccountRepository(t)
	keeper := secrets.TestStringKeeper(t)

	handler := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), handler, repo, testFedClient, testPayGateClient, keeper, keeper, createTestOFACSearcher(repo, nil))

	client := testclient.New(t, handler)

	// create an account
	account := httpCreateAccount(t, handler, customerID)
	if account == nil {
		t.Fatal("missing account")
	}

	transit, resp, err := client.CustomersApi.DecryptAccountNumber(context.TODO(), customerID, account.AccountID, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		t.Error(err)
	}
	if transit.AccountNumber == "" {
		t.Error("missing transit AccountNumber")
	}

	decrypted, err := keeper.DecryptString(transit.AccountNumber)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != "18749" {
		t.Errorf("decrypted=%q", decrypted)
	}
}

func TestRoutes__EmptyAccounts(t *testing.T) {
	customerID := base.ID()
	repo := setupTestAccountRepository(t)
	keeper := secrets.TestStringKeeper(t)

	handler := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), handler, repo, testFedClient, testPayGateClient, keeper, keeper, createTestOFACSearcher(repo, nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/customers/%s/accounts", customerID), nil)
	handler.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	if body := w.Body.String(); body != "[]\n" {
		t.Errorf("unexpected response body: %q", body)
	}
}

func httpReadAccounts(t *testing.T, handler *mux.Router, customerID string) []*client.Account {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/customers/%s/accounts", customerID), nil)
	handler.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var wrapper []*client.Account
	if err := json.NewDecoder(w.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}
	return wrapper
}

func httpCreateAccount(t *testing.T, handler *mux.Router, customerID string) *client.Account {
	params := &createAccountRequest{
		HolderName:    "John Doe",
		AccountNumber: "18749",
		RoutingNumber: "987654320",
		Type:          client.SAVINGS,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(params); err != nil {
		t.Fatal(err)
	}
	body := bytes.NewReader(buf.Bytes())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/accounts", customerID), body)
	handler.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status %d: %v", w.Code, w.Body.String())
	}

	var account client.Account
	if err := json.NewDecoder(w.Body).Decode(&account); err != nil {
		t.Fatal(err)
	}
	return &account
}

func httpDeleteAccount(t *testing.T, handler *mux.Router, customerID, accountID string) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/customers/%s/accounts/%s", customerID, accountID), nil)
	handler.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status %d: %v", w.Code, w.Body.String())
	}
}

func httpReadAccountOfacSearch(t *testing.T, handler *mux.Router, customerID, accountID string) *client.OfacSearch {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/customers/%s/accounts/%s/ofac", customerID, accountID), nil)
	handler.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status %d: %v", w.Code, w.Body.String())
	}

	var ofacSearch client.OfacSearch
	if err := json.NewDecoder(w.Body).Decode(&ofacSearch); err != nil {
		t.Fatal(err)
	}
	return &ofacSearch
}
