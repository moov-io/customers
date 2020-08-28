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
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/moov-io/customers/cmd/server/accounts/validator/microdeposits"
	"github.com/moov-io/customers/cmd/server/paygate"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/secrets"
	payclient "github.com/moov-io/paygate/pkg/client"
	"github.com/stretchr/testify/require"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func TestRouter__ValidateAccounts(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	repo := setupTestAccountRepository(t)
	keeper := secrets.TestStringKeeper(t)

	paygateClient := &paygate.MockClient{
		Micro: &payclient.MicroDeposits{
			Amounts: []string{"USD 0.03", "USD 0.07"},
			Status:  payclient.PROCESSED,
		},
	}

	handler := mux.NewRouter()
	RegisterRoutes(log.NewNopLogger(), handler, repo, testFedClient, paygateClient, keeper, keeper)

	// create account
	acct, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          client.CHECKING,
	})
	if err != nil {
		t.Fatal(err)
	}

	// make our validation request
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(client.UpdateValidation{
		Strategy:      "micro-deposits",
		MicroDeposits: []string{"USD 0.03", "USD 0.07"},
	}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/accounts/%s/validate", customerID, acct.AccountID), &body)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

type accountValidationRequest struct {
	Strategy string `json:"strategy"`
	Vendor   string `json:"vendor"`
}

func TestRouter__InitAccountValidation(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	repo := setupTestAccountRepository(t)
	// keeper := secrets.TestStringKeeper(t)

	strategies := map[validator.StrategyKey]validator.Strategy{
		validator.StrategyKey{"test", "moov"}: validator.TestStrategy(),
	}

	// create account
	acc, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          client.CHECKING,
	})
	require.NoError(t, err)

	t.Run("Test micro-deposits strategy", func(t *testing.T) {
		paygateClient := &paygate.MockClient{
			// this is how we make call to initiate micro deposit successful
			Err: nil,
		}

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

		// // create account
		// acct, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
		// 	AccountNumber: "123",
		// 	RoutingNumber: "987654320",
		// 	Type:          client.CHECKING,
		// })
		// if err != nil {
		// 	t.Fatal(err)
		// }

		// // make our validation request
		// var body bytes.Buffer
		// if err := json.NewEncoder(&body).Encode(client.UpdateValidation{
		// 	Strategy:      "micro-deposits",
		// 	MicroDeposits: []string{"USD 0.03", "USD 0.07"},
		// }); err != nil {
		// 	t.Fatal(err)
		// }
		// req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/accounts/%s/validate", customerID, acct.AccountID), &body)

		// w := httptest.NewRecorder()
		// handler.ServeHTTP(w, req)
		// w.Flush()

		// if w.Code != http.StatusOK {
		// 	t.Errorf("bogus HTTP status: %d", w.Code)
		// }
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
			ValidationID   string            `json:"validationID"`
			Status         string            `json:"status"`
			Strategy       string            `json:"strategy"`
			Vendor         string            `json:"vendor"`
			VendorResponse map[string]string `json:"vendor_response"`
		}

		want := &validationResponse{
			ValidationID: "1234",
			Status:       "pending",
			Strategy:     "test",
			Vendor:       "moov",
			VendorResponse: map[string]string{
				"test": "ok",
			},
		}

		got := &validationResponse{}
		json.NewDecoder(w.Body).Decode(got)
		if diff := cmp.Diff(got, want); len(diff) != 0 {
			t.Errorf(diff)
		}
	})

	// fmt.Println(strategies)

	// require.NoError(t, err)
	// fmt.Println(acc)

	// handler := initAccountValidation(log.NewNopLogger(), repo, strategies)

	// w := httptest.NewRecorder()
	// req, err := http.NewRequest("POST", "/", nil)
	// require.NoError(t, err)

	// // mux.SetURLVars(req, map[string]string{"foo": "bar"})
	// handler(w, req)

	// type validationResponse struct {
	// 	ValidationID   string            `json:"validationID"`
	// 	Status         string            `json:"status"`
	// 	Strategy       string            `json:"strategy"`
	// 	Vendor         string            `json:"vendor"`
	// 	VendorResponse map[string]string `json:"vendor_response"`
	// }

	// want := &validationResponse{
	// 	ValidationID: "1234",
	// 	Status:       "pending",
	// 	Strategy:     "test",
	// 	Vendor:       "moov",
	// 	VendorResponse: map[string]string{
	// 		"test": "ok",
	// 	},
	// }

	// fmt.Println(w.Body.String())

	// got := &validationResponse{}
	// json.NewDecoder(w.Body).Decode(got)
	// if diff := cmp.Diff(got, want); len(diff) != 0 {
	// 	t.Errorf(diff)
	// }

}
