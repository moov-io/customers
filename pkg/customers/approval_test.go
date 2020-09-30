// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/base"
	"github.com/stretchr/testify/require"

	"github.com/moov-io/customers/pkg/client"

	"github.com/go-kit/kit/log"
)

func TestCustomers__updateCustomerStatus(t *testing.T) {
	repo := &testCustomerRepository{
		customer: &client.Customer{
			CustomerID: base.ID(),
			Status:     "none",
		},
	}
	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t), createTestOFACSearcher(repo, nil))

	updateStatusRequest := client.UpdateCustomerStatus{
		Status:  "ReceiveOnly",
		Comment: "test comment",
	}
	payload, err := json.Marshal(&updateStatusRequest)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/customers/_id_/status", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)

	var customer client.Customer
	err = json.NewDecoder(res.Body).Decode(&customer)
	require.NoError(t, err)
	require.NotNil(t, customer)
	require.Equal(t, updateStatusRequest.Status, repo.updatedStatus)
}

func TestCustomers__containsValidPrimaryAddress(t *testing.T) {
	if containsValidPrimaryAddress(nil) {
		t.Error("no addresses, so can't be found")
	}
	addresses := []client.CustomerAddress{
		{
			Type:      "Primary",
			Validated: false,
		},
	}
	if containsValidPrimaryAddress(addresses) {
		t.Error("Address isn't validated")
	}
	addresses[0].Validated = true
	if !containsValidPrimaryAddress(addresses) {
		t.Error("Address should be Primary and Validated")
	}
	addresses[0].Type = "Secondary"
	if containsValidPrimaryAddress(addresses) {
		t.Error("Address is Secondary")
	}
}
