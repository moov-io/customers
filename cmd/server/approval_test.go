// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/base/admin"
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func TestCustomers__updateCustomerStatus(t *testing.T) {
	repo := &testCustomerRepository{
		customer: &client.Customer{
			CustomerID: base.ID(),
			Status:     "none",
		},
	}
	searcher := createTestOFACSearcher(repo, nil)
	ssnRepo := &testCustomerSSNRepository{}

	svc := admin.NewServer(":10001")
	defer svc.Shutdown()
	addApprovalRoutes(log.NewNopLogger(), svc, repo, ssnRepo, searcher)
	go svc.Listen()

	body := strings.NewReader(`{"status": "ReceiveOnly", "comment": "test comment"}`)
	req, err := http.NewRequest("PUT", "http://localhost:10001/customers/foo/status", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status: %d: %v", resp.StatusCode, string(respBody))
	}

	var customer client.Customer
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&customer); err != nil {
		t.Fatal(err)
	}
	if customer.CustomerID == "" {
		t.Errorf("missing customer JSON: %#v", customer)
	}
	if repo.updatedStatus != client.RECEIVE_ONLY {
		t.Errorf("unexpected status: %s", repo.updatedStatus)
	}
}

func TestCustomers__getAddressId(t *testing.T) {
	var addr string

	router := mux.NewRouter()
	router.Methods("GET").Path("/addresses/{addressId}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addr = getAddressId(w, r)
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/addresses/foobar", nil)
	router.ServeHTTP(w, req)
	w.Flush()

	if addr != "foobar" {
		t.Errorf("addr=%s", addr)
	}
	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestCustomers__updateCustomerStatusFailure(t *testing.T) {
	repo := &testCustomerRepository{}
	searcher := createTestOFACSearcher(repo, nil)
	ssnRepo := &testCustomerSSNRepository{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo/status", nil)
	updateCustomerStatus(log.NewNopLogger(), repo, ssnRepo, searcher)(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	// try the proper HTTP verb
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/customers/foo/status", nil)
	updateCustomerStatus(log.NewNopLogger(), repo, ssnRepo, searcher)(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
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

func TestCustomers__updateCustomerAddress(t *testing.T) {
	repo := &testCustomerRepository{
		customer: &client.Customer{
			CustomerID: base.ID(),
		},
	}
	searcher := createTestOFACSearcher(repo, nil)
	ssnRepo := &testCustomerSSNRepository{}

	svc := admin.NewServer(":10002")
	defer svc.Shutdown()
	addApprovalRoutes(log.NewNopLogger(), svc, repo, ssnRepo, searcher)
	go svc.Listen()

	body := strings.NewReader(`{"type": "primary", "validated": true}`)
	req, err := http.NewRequest("PUT", "http://localhost:10002/customers/foo/addresses/bar", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status: %d: %v", resp.StatusCode, string(respBody))
	}

	// quick updateCustomerAddressRequest.validate() call
	request := &updateCustomerAddressRequest{}
	if err := request.validate(); err == nil {
		t.Errorf("expected error")
	}
}

func TestCustomers__updateCustomerAddressFailure(t *testing.T) {
	repo := &testCustomerRepository{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo/address/bar", nil)
	updateCustomerAddress(log.NewNopLogger(), repo)(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	// try the proper HTTP verb
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/customers/foo/address/bar", nil)
	updateCustomerAddress(log.NewNopLogger(), repo)(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestCustomerRepository__updateCustomerAddress(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Phones: []phone{
			{
				Number: "123.456.7890",
				Type:   "Checking",
			},
		},
		Addresses: []address{
			{
				Address1:   "123 1st st",
				City:       "fake city",
				State:      "CA",
				PostalCode: "90210",
				Country:    "US",
			},
		},
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.createCustomer(cust); err != nil {
		t.Fatal(err)
	}

	// update the Address
	if err := repo.updateCustomerAddress(cust.CustomerID, cust.Addresses[0].AddressID, "Primary", true); err != nil {
		t.Error(err)
	}

	cust, err := repo.getCustomer(cust.CustomerID)
	if err != nil {
		t.Error(err)
	}
	if len(cust.Addresses) != 1 {
		t.Errorf("got %d Addresses", len(cust.Addresses))
	}
	if !strings.EqualFold(cust.Addresses[0].Type, "primary") {
		t.Errorf("cust.Addresses[0].Type=%s", cust.Addresses[0].Type)
	}
	if !cust.Addresses[0].Validated {
		t.Errorf("cust.Addresses[0].Validated=%v", cust.Addresses[0].Validated)
	}
}
