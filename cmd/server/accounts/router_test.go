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
	"strings"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/moov-io/customers/cmd/server/accounts/validator/testvalidator"
	"github.com/moov-io/customers/cmd/server/fed"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/stretchr/testify/require"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func TestAccountRoutes(t *testing.T) {
	customerID := base.ID()

	handler := setupRouter(t)

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

	// validate account
	httpInitAccountValidation(t, handler, customerID, account.AccountID)
	httpCompleteAccountValidation(t, handler, customerID, account.AccountID)

	// delete and expect no accounts
	httpDeleteAccount(t, handler, customerID, account.AccountID)
	accounts = httpReadAccounts(t, handler, customerID)
	if len(accounts) != 0 {
		t.Errorf("got accounts: %v", accounts)
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

	handler := setupRouter(t)

	// create an account
	account := httpCreateAccount(t, handler, customerID)
	if account == nil {
		t.Fatal("missing account")
	}

	httpDecryptAccountNumber(t, handler, customerID, account.AccountID)
}

func TestRoutes__EmptyAccounts(t *testing.T) {
	customerID := base.ID()

	handler := setupRouter(t)

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

func setupRouter(t *testing.T) *mux.Router {
	handler := mux.NewRouter()

	testFedClient := &fed.MockClient{}
	accounts := setupTestAccountRepository(t)
	validations := &validator.MockRepository{}
	keeper := secrets.TestStringKeeper(t)
	validationStrategies := map[validator.StrategyKey]validator.Strategy{
		{Strategy: "test", Vendor: "moov"}: testvalidator.NewStrategy(),
	}

	RegisterRoutes(log.NewNopLogger(), handler, accounts, validations, testFedClient, keeper, keeper, validationStrategies)

	return handler
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

func httpInitAccountValidation(t *testing.T, handler *mux.Router, customerID, accountID string) {
	params := &client.InitAccountValidationRequest{
		Strategy: "test",
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(params); err != nil {
		t.Fatal(err)
	}
	body := bytes.NewReader(buf.Bytes())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/accounts/%s/validate", customerID, accountID), body)
	handler.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status %d: %v", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "initiated") {
		t.Errorf("expected successful response: %v", w.Body.String())
	}
}

func httpCompleteAccountValidation(t *testing.T, handler *mux.Router, customerID, accountID string) {
	params := &client.CompleteAccountValidationRequest{
		Strategy: "test",
		VendorRequest: validator.VendorRequest{
			"result": "success",
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(params); err != nil {
		t.Fatal(err)
	}
	body := bytes.NewReader(buf.Bytes())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/accounts/%s/validate", customerID, accountID), body)
	handler.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status %d: %v", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "validated") {
		t.Errorf("expected successful response: %v", w.Body.String())
	}
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

func httpDecryptAccountNumber(t *testing.T, handler *mux.Router, customerID, accountID string) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/accounts/%s/decrypt", customerID, accountID), nil)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status %d: %v", w.Code, w.Body.String())
	}

	// check if response body contains accountNumber.
	// Example:
	//   {"accountNumber":"XueflKMjfidC2Ifommst9iSK+xF/sn2x+pK/K"}
	require.Contains(t, w.Body.String(), `"accountNumber":`)
}
