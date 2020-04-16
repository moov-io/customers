// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func TestRoutes(t *testing.T) {
	customerID := base.ID()
	repo := setupTestAccountRepository(t)

	handler := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), handler, repo)

	// first read, expect no accounts
	accounts := httpReadAccounts(t, handler, customerID)
	if len(accounts) != 0 {
		t.Errorf("got accounts: %v", accounts)
	}

	// create an account
	account := httpCreateAccount(t, handler, customerID)

	// re-read, find account
	accounts = httpReadAccounts(t, handler, customerID)
	if len(accounts) != 1 || accounts[0].Id != account.Id {
		t.Errorf("got accounts: %v", accounts)
	}

	// delete and expect no accounts
	httpDeleteAccount(t, handler, customerID, account.Id)
	accounts = httpReadAccounts(t, handler, customerID)
	if len(accounts) != 0 {
		t.Errorf("got accounts: %v", accounts)
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
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          "Checking",
		HolderType:    "individual",
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