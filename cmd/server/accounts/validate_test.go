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

	"github.com/google/go-cmp/cmp"
	"github.com/moov-io/base"
	"github.com/moov-io/customers/admin"
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/moov-io/customers/cmd/server/accounts/validator/microdeposits"
	"github.com/moov-io/customers/cmd/server/accounts/validator/testvalidator"
	"github.com/moov-io/customers/cmd/server/paygate"
	payclient "github.com/moov-io/paygate/pkg/client"
	"github.com/stretchr/testify/require"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func TestRouter__InitAccountValidation(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	repo := setupTestAccountRepository(t)
	// keeper := secrets.TestStringKeeper(t)

	strategies := map[validator.StrategyKey]validator.Strategy{
		validator.StrategyKey{"test", "moov"}: testvalidator.NewStrategy(),
	}

	// create account
	acc, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          client.CHECKING,
	})
	require.NoError(t, err)

	t.Run("Test when account is validated already", func(t *testing.T) {
		acc, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
			AccountNumber: "1234",
			RoutingNumber: "987654321",
			Type:          client.CHECKING,
		})
		require.NoError(t, err)

		err = repo.updateAccountStatus(acc.AccountID, admin.VALIDATED)
		require.NoError(t, err)

		params := map[string]string{
			"strategy": "test",
		}

		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(params)
		require.NoError(t, err)
		body := bytes.NewReader(buf.Bytes())

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := initAccountValidation(log.NewNopLogger(), repo, strategies)
		handler(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), fmt.Sprintf("expected accountID=%s status to be 'none'", acc.AccountID))
	})

	t.Run("Test unknown strategy is requested", func(t *testing.T) {
		params := map[string]string{
			"strategy": "unknown",
		}

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(params)
		require.NoError(t, err)
		body := bytes.NewReader(buf.Bytes())

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := initAccountValidation(log.NewNopLogger(), repo, strategies)
		handler(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "strategy unknown for vendor moov was not found")
	})

	t.Run("Test 'test' strategy", func(t *testing.T) {
		params := map[string]string{
			"strategy": "test",
		}

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(params)
		require.NoError(t, err)
		body := bytes.NewReader(buf.Bytes())

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := initAccountValidation(log.NewNopLogger(), repo, strategies)
		handler(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		type validationResponse struct {
			VendorResponse map[string]string `json:"vendor_response"`
		}

		want := &validationResponse{
			VendorResponse: map[string]string{
				"result": "initiated",
			},
		}

		got := &validationResponse{}
		json.NewDecoder(w.Body).Decode(got)
		if diff := cmp.Diff(got, want); len(diff) != 0 {
			t.Errorf(diff)
		}
	})

	// this test now does not add any value as microdeposits strategy
	// has own tests
	t.Run("Test micro-deposits strategy", func(t *testing.T) {
		paygateClient := &paygate.MockClient{}

		strategies := map[validator.StrategyKey]validator.Strategy{
			validator.StrategyKey{"micro-deposits", "moov"}: microdeposits.NewStrategy(paygateClient),
		}

		// build request with strategy params
		params := map[string]string{
			"strategy": "micro-deposits",
		}

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(params)
		require.NoError(t, err)
		body := bytes.NewReader(buf.Bytes())

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := initAccountValidation(log.NewNopLogger(), repo, strategies)
		handler(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRouter__CompleteAccountValidation(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	repo := setupTestAccountRepository(t)

	strategies := map[validator.StrategyKey]validator.Strategy{
		validator.StrategyKey{"test", "moov"}: testvalidator.NewStrategy(),
	}

	t.Run("Test when account is validated already", func(t *testing.T) {
		acc, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
			AccountNumber: "123",
			RoutingNumber: "987654320",
			Type:          client.CHECKING,
		})
		require.NoError(t, err)

		err = repo.updateAccountStatus(acc.AccountID, admin.VALIDATED)
		require.NoError(t, err)

		// build request for test strategy
		params := map[string]interface{}{
			"strategy": "test",
			"vendor_request": map[string]string{
				"result": "success",
			},
		}

		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(params)
		require.NoError(t, err)
		body := bytes.NewReader(buf.Bytes())

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := completeAccountValidation(log.NewNopLogger(), repo, strategies)
		handler(w, req)

		// require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), fmt.Sprintf("expected accountID=%s status to be 'none'", acc.AccountID))
	})

	t.Run("Test unknown strategy is requested", func(t *testing.T) {
		acc, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
			AccountNumber: "1236",
			RoutingNumber: "987654323",
			Type:          client.CHECKING,
		})
		require.NoError(t, err)

		params := map[string]string{
			"strategy": "unknown",
		}

		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(params)
		require.NoError(t, err)
		body := bytes.NewReader(buf.Bytes())

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := completeAccountValidation(log.NewNopLogger(), repo, strategies)
		handler(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "strategy unknown for vendor moov was not found")
	})

	t.Run("Test 'test' strategy", func(t *testing.T) {
		acc, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
			AccountNumber: "1234",
			RoutingNumber: "987654321",
			Type:          client.CHECKING,
		})
		require.NoError(t, err)

		// build request for test strategy
		params := map[string]interface{}{
			"strategy": "test",
			"vendor_request": map[string]string{
				"result": "success",
			},
		}

		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(params)
		require.NoError(t, err)
		body := bytes.NewReader(buf.Bytes())

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := completeAccountValidation(log.NewNopLogger(), repo, strategies)
		handler(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		type validationResponse struct {
			VendorResponse map[string]string `json:"vendor_response"`
		}

		want := &validationResponse{
			VendorResponse: map[string]string{
				"result": "validated",
			},
		}

		got := &validationResponse{}
		json.NewDecoder(w.Body).Decode(got)
		if diff := cmp.Diff(got, want); len(diff) != 0 {
			t.Errorf(diff)
		}
	})

	// this test now does not add any value as microdeposits strategy
	// has own tests. The only value I see right now is that it shows how
	// request for microdeposits strategy may look like...
	t.Run("Test micro-deposits strategy", func(t *testing.T) {
		acc, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
			AccountNumber: "12345",
			RoutingNumber: "987654322",
			Type:          client.CHECKING,
		})
		require.NoError(t, err)

		paygateClient := &paygate.MockClient{
			Micro: &payclient.MicroDeposits{
				MicroDepositID: base.ID(),
				Amounts:        []string{"USD 0.03", "USD 0.07"},
				Status:         payclient.PROCESSED,
			},
		}

		strategies := map[validator.StrategyKey]validator.Strategy{
			validator.StrategyKey{"micro-deposits", "moov"}: microdeposits.NewStrategy(paygateClient),
		}

		// build request with strategy params
		params := map[string]interface{}{
			"strategy": "micro-deposits",
			"vendor_request": map[string][]string{
				"micro-deposits": []string{"USD 0.03", "USD 0.07"},
			},
		}

		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(params)
		require.NoError(t, err)
		body := bytes.NewReader(buf.Bytes())

		w := httptest.NewRecorder()
		req, err := http.NewRequest("PUT", "/", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := completeAccountValidation(log.NewNopLogger(), repo, strategies)
		handler(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		updatedAccount, err := repo.getCustomerAccount(customerID, acc.AccountID)
		require.NoError(t, err)
		require.Equal(t, client.VALIDATED, updatedAccount.Status)
	})
}
