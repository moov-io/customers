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

	"github.com/moov-io/customers/pkg/customers"

	"github.com/google/go-cmp/cmp"
	"github.com/moov-io/base"
	payclient "github.com/moov-io/paygate/pkg/client"
	"github.com/stretchr/testify/require"

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/paygate"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/moov-io/customers/pkg/validator"
	"github.com/moov-io/customers/pkg/validator/microdeposits"
	"github.com/moov-io/customers/pkg/validator/testvalidator"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
)

func TestRouter__AccountValidation(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	organization := "moov"
	accounts := setupTestAccountRepository(t)
	validations := &validator.MockRepository{}

	// create account
	acc, err := accounts.CreateCustomerAccount(customerID, userID, &CreateAccountRequest{
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          client.ACCOUNTTYPE_CHECKING,
	})
	require.NoError(t, err)

	t.Run("Test when validation was not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Organization", organization)
		req = mux.SetURLVars(req, map[string]string{
			"customerID":   customerID,
			"accountID":    acc.AccountID,
			"validationID": "xxx",
		})

		handler := getAccountValidation(log.NewNopLogger(), accounts, validations)
		handler(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), fmt.Sprintf("validation: %s was not found", "xxx"))
	})

	t.Run("Test when validation was not found", func(t *testing.T) {
		validation := &validator.Validation{
			AccountID: acc.AccountID,
			Strategy:  "test",
			Vendor:    "moov",
			Status:    validator.StatusInit,
		}
		err = validations.CreateValidation(validation)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Organization", organization)
		req = mux.SetURLVars(req, map[string]string{
			"customerID":   customerID,
			"accountID":    acc.AccountID,
			"validationID": validation.ValidationID,
		})

		handler := getAccountValidation(log.NewNopLogger(), accounts, validations)
		handler(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var response client.AccountValidationResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}

		require.Equal(t, validation.ValidationID, response.ValidationID)
		require.Equal(t, validation.Status, response.Status)
	})
}

func TestRouter__InitAccountValidation(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	organization := "moov"
	accounts := setupTestAccountRepository(t)
	validations := &validator.MockRepository{}

	strategies := map[validator.StrategyKey]validator.Strategy{
		{Strategy: "test", Vendor: "moov"}: testvalidator.NewStrategy(),
	}

	// create account
	acc, err := accounts.CreateCustomerAccount(customerID, userID, &CreateAccountRequest{
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          client.ACCOUNTTYPE_CHECKING,
	})
	require.NoError(t, err)

	t.Run("Test when account is validated already", func(t *testing.T) {
		acc, err := accounts.CreateCustomerAccount(customerID, userID, &CreateAccountRequest{
			AccountNumber: "1234",
			RoutingNumber: "987654321",
			Type:          client.ACCOUNTTYPE_CHECKING,
		})
		require.NoError(t, err)

		err = accounts.updateAccountStatus(acc.AccountID, client.ACCOUNTSTATUS_VALIDATED)
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
		req.Header.Set("X-Organization", organization)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := initAccountValidation(log.NewNopLogger(), accounts, validations, strategies)
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
		req.Header.Set("X-Organization", organization)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := initAccountValidation(log.NewNopLogger(), accounts, validations, strategies)
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
		req.Header.Set("X-Organization", organization)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := initAccountValidation(log.NewNopLogger(), accounts, validations, strategies)
		handler(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// get created validation from validation mock repo
		require.Len(t, validations.Validations, 1)
		validation := validations.Validations[0]

		want := &client.InitAccountValidationResponse{
			ValidationID: validation.ValidationID,
			Strategy:     "test",
			Vendor:       "moov",
			Status:       validator.StatusInit,
			CreatedAt:    validation.CreatedAt,
			UpdatedAt:    validation.UpdatedAt,
			VendorResponse: validator.VendorResponse{
				"result": "initiated",
			},
		}

		got := &client.InitAccountValidationResponse{}
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
			{Strategy: "micro-deposits", Vendor: "moov"}: microdeposits.NewStrategy(paygateClient),
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
		req.Header.Set("X-Organization", organization)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"customerID": customerID, "accountID": acc.AccountID})

		handler := initAccountValidation(log.NewNopLogger(), accounts, validations, strategies)
		handler(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRouter__CompleteAccountValidation(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	organization := "moov"
	repo := setupTestAccountRepository(t)
	keeper := secrets.TestStringKeeper(t)
	validations := &validator.MockRepository{}

	strategies := map[validator.StrategyKey]validator.Strategy{
		{Strategy: "test", Vendor: "moov"}: testvalidator.NewStrategy(),
	}

	acc, err := repo.CreateCustomerAccount(customerID, userID, &CreateAccountRequest{
		AccountNumber: "123456",
		RoutingNumber: "987654323",
		Type:          client.ACCOUNTTYPE_CHECKING,
	})
	require.NoError(t, err)

	t.Run("Test when account is validated already", func(t *testing.T) {
		acc, err := repo.CreateCustomerAccount(customerID, userID, &CreateAccountRequest{
			AccountNumber: "123",
			RoutingNumber: "987654320",
			Type:          client.ACCOUNTTYPE_CHECKING,
		})
		require.NoError(t, err)

		err = repo.updateAccountStatus(acc.AccountID, client.ACCOUNTSTATUS_VALIDATED)
		require.NoError(t, err)

		// build request for test strategy
		params := &client.CompleteAccountValidationRequest{
			VendorRequest: validator.VendorRequest{
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
		req.Header.Set("X-Organization", organization)
		req = mux.SetURLVars(req, map[string]string{
			"customerID":   customerID,
			"accountID":    acc.AccountID,
			"validationID": "xxx",
		})

		handler := completeAccountValidation(log.NewNopLogger(), repo, validations, keeper, strategies)
		handler(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), fmt.Sprintf("expected accountID=%s status to be 'none'", acc.AccountID))
	})

	t.Run("Test when validation was not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Organization", organization)
		req = mux.SetURLVars(req, map[string]string{
			"customerID":   customerID,
			"accountID":    acc.AccountID,
			"validationID": "xxx",
		})

		handler := completeAccountValidation(log.NewNopLogger(), repo, validations, keeper, strategies)
		handler(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)

		require.Contains(t, w.Body.String(), fmt.Sprintf("validation: %s was not found", "xxx"))
	})

	// Test validation status should be init
	// Test validation status after complete should be complete
	t.Run("Test when validation status is not init", func(t *testing.T) {
		validation := &validator.Validation{
			AccountID: acc.AccountID,
			Strategy:  "test",
			Vendor:    "moov",
			Status:    validator.StatusCompleted,
		}
		err = validations.CreateValidation(validation)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Organization", organization)
		req = mux.SetURLVars(req, map[string]string{
			"customerID":   customerID,
			"accountID":    acc.AccountID,
			"validationID": validation.ValidationID,
		})

		handler := completeAccountValidation(log.NewNopLogger(), repo, validations, keeper, strategies)
		handler(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), fmt.Sprintf("validation: %s status to be 'init'", validation.ValidationID))
	})

	t.Run("Test 'test' strategy", func(t *testing.T) {
		encrypted, err := keeper.EncryptString("1234")
		require.NoError(t, err)

		acc, err := repo.CreateCustomerAccount(customerID, userID, &CreateAccountRequest{
			AccountNumber:          "1234",
			RoutingNumber:          "987654321",
			Type:                   client.ACCOUNTTYPE_CHECKING,
			encryptedAccountNumber: encrypted,
		})
		require.NoError(t, err)

		customerRepo := customers.NewCustomerRepo(log.NewNopLogger(), repo.db.DB)
		cust := &client.Customer{
			CustomerID: customerID,
			FirstName:  "mary",
			LastName:   "doe",
			Type:       client.CUSTOMERTYPE_INDIVIDUAL,
		}
		custErr := customerRepo.CreateCustomer(cust, organization)
		if custErr != nil {
			t.Fatal(custErr)
		}

		validation := &validator.Validation{
			AccountID: acc.AccountID,
			Strategy:  "test",
			Vendor:    "moov",
			Status:    validator.StatusInit,
		}
		err = validations.CreateValidation(validation)
		require.NoError(t, err)

		// build request for test strategy
		params := &client.CompleteAccountValidationRequest{
			VendorRequest: validator.VendorRequest{
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
		req.Header.Set("X-Organization", organization)
		req = mux.SetURLVars(req, map[string]string{
			"customerID":   customerID,
			"accountID":    acc.AccountID,
			"validationID": validation.ValidationID,
		})

		handler := completeAccountValidation(log.NewNopLogger(), repo, validations, keeper, strategies)
		handler(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		want := &client.CompleteAccountValidationResponse{
			ValidationID: validation.ValidationID,
			Strategy:     "test",
			Vendor:       "moov",
			Status:       validator.StatusCompleted,
			CreatedAt:    validation.CreatedAt,
			UpdatedAt:    validation.UpdatedAt,
			VendorResponse: validator.VendorResponse{
				"result": "validated",
			},
		}

		got := &client.CompleteAccountValidationResponse{}
		json.NewDecoder(w.Body).Decode(got)
		if diff := cmp.Diff(got, want); len(diff) != 0 {
			t.Errorf(diff)
		}
	})

	// this test now does not add any value as microdeposits strategy
	// has own tests. The only value I see right now is that it shows how
	// request for microdeposits strategy may look like...
	t.Run("Test micro-deposits strategy", func(t *testing.T) {
		// Redefine customerID and userID to use new unique values
		customerID, userID := base.ID(), base.ID()
		encrypted, err := keeper.EncryptString("12345")
		require.NoError(t, err)

		acc, err := repo.CreateCustomerAccount(customerID, userID, &CreateAccountRequest{
			AccountNumber:          "12345",
			RoutingNumber:          "987654322",
			Type:                   client.ACCOUNTTYPE_CHECKING,
			encryptedAccountNumber: encrypted,
		})
		require.NoError(t, err)

		validation := &validator.Validation{
			AccountID: acc.AccountID,
			Strategy:  "micro-deposits",
			Vendor:    "moov",
			Status:    validator.StatusInit,
		}
		err = validations.CreateValidation(validation)
		require.NoError(t, err)

		paygateClient := &paygate.MockClient{
			Micro: &payclient.MicroDeposits{
				MicroDepositID: base.ID(),
				Amounts: []payclient.Amount{
					{Currency: "USD", Value: 3},
					{Currency: "USD", Value: 7},
				},
				Status: payclient.PROCESSED,
			},
		}

		strategies := map[validator.StrategyKey]validator.Strategy{
			{Strategy: "micro-deposits", Vendor: "moov"}: microdeposits.NewStrategy(paygateClient),
		}

		customerRepo := customers.NewCustomerRepo(log.NewNopLogger(), repo.db.DB)
		cust := &client.Customer{
			CustomerID: customerID,
			FirstName:  "john",
			LastName:   "doe",
			Type:       client.CUSTOMERTYPE_INDIVIDUAL,
		}
		custErr := customerRepo.CreateCustomer(cust, organization)
		if custErr != nil {
			t.Fatal(custErr)
		}

		// build request with strategy params
		params := &client.CompleteAccountValidationRequest{
			VendorRequest: validator.VendorRequest{
				"micro-deposits": []payclient.Amount{
					{Currency: "USD", Value: 3},
					{Currency: "USD", Value: 7},
				},
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
		req.Header.Set("X-Organization", organization)
		req = mux.SetURLVars(req, map[string]string{
			"customerID":   customerID,
			"accountID":    acc.AccountID,
			"validationID": validation.ValidationID,
		})

		handler := completeAccountValidation(log.NewNopLogger(), repo, validations, keeper, strategies)
		handler(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// check if account status was set to validated
		updatedAccount, err := repo.getCustomerAccount(customerID, acc.AccountID)
		require.NoError(t, err)
		require.Equal(t, client.ACCOUNTSTATUS_VALIDATED, updatedAccount.Status)

		// check if validation status was set to completed
		validation, err = validations.GetValidation(validation.AccountID, validation.ValidationID)
		require.NoError(t, err)
		require.Equal(t, validator.StatusCompleted, validation.Status)
	})
}
