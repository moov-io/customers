// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/moov-io/base"
	"github.com/moov-io/base/admin"
	client "github.com/moov-io/customers/client"
	ofac "github.com/moov-io/ofac/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func TestCustomers__updateCustomerStatus(t *testing.T) {
	repo := &testCustomerRepository{
		customer: &client.Customer{
			Id: base.ID(),
		},
	}
	searcher := createTestOFACSearcher(repo, nil)

	svc := admin.NewServer(":10001")
	defer svc.Shutdown()
	addApprovalRoutes(log.NewNopLogger(), svc, repo, searcher)
	go svc.Listen()

	body := strings.NewReader(`{"status": "ReviewRequired", "comment": "test comment"}`)
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
	if customer.Id == "" {
		t.Errorf("missing customer JSON: %#v", customer)
	}
	if repo.updatedStatus != CustomerStatusReviewRequired {
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

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo/status", nil)
	updateCustomerStatus(log.NewNopLogger(), repo, searcher)(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	// try the proper HTTP verb
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/customers/foo/status", nil)
	updateCustomerStatus(log.NewNopLogger(), repo, searcher)(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestCustomers__containsValidPrimaryAddress(t *testing.T) {
	if containsValidPrimaryAddress(nil) {
		t.Error("no addresses, so can't be found")
	}
	addresses := []client.Address{
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

func TestCustomers__validCustomerStatusTransition(t *testing.T) {
	cust := &client.Customer{
		Id:     base.ID(),
		Status: CustomerStatusNone,
	}
	repo := &testCustomerRepository{}
	searcher := createTestOFACSearcher(repo, nil)

	if err := validCustomerStatusTransition(cust, CustomerStatusDeceased, repo, searcher, "requestId"); err != nil {
		t.Errorf("expected no error: %v", err)
	}

	// block Deceased and Rejected customers
	cust.Status = CustomerStatusDeceased
	if err := validCustomerStatusTransition(cust, CustomerStatusKYC, repo, searcher, "requestId"); err == nil {
		t.Error("expected error")
	}
	cust.Status = CustomerStatusRejected
	if err := validCustomerStatusTransition(cust, CustomerStatusKYC, repo, searcher, "requestId"); err == nil {
		t.Error("expected error")
	}

	// normal KYC approval (rejected due to missing info)
	cust.FirstName, cust.LastName = "Jane", "Doe"
	cust.Status = CustomerStatusReviewRequired
	if err := validCustomerStatusTransition(cust, CustomerStatusKYC, repo, searcher, "requestId"); err == nil {
		t.Error("expected error")
	}
	cust.BirthDate = time.Now()
	if err := validCustomerStatusTransition(cust, CustomerStatusKYC, repo, searcher, "requestId"); err == nil {
		t.Error("expected error")
	}
	cust.Addresses = append(cust.Addresses, client.Address{
		Type:     "primary",
		Address1: "123 1st st",
	})

	// CIP transistions are WIP // TODO(adam):
	cust.Status = CustomerStatusReviewRequired
	if err := validCustomerStatusTransition(cust, CustomerStatusCIP, repo, searcher, "requestId"); err == nil {
		t.Error("CIP transition is WIP")
	}
}

func TestCustomers__validCustomerStatusTransitionError(t *testing.T) {
	cust := &client.Customer{
		Id:     base.ID(),
		Status: CustomerStatusReviewRequired,
	}
	repo := &testCustomerRepository{}
	ofacClient := &testOFACClient{}
	searcher := createTestOFACSearcher(repo, ofacClient)

	repo.err = errors.New("bad error")
	if err := validCustomerStatusTransition(cust, CustomerStatusOFAC, repo, searcher, ""); err == nil {
		t.Error("expected error, but got none")
	}
	repo.err = nil

	ofacClient.err = errors.New("bad error")
	if err := validCustomerStatusTransition(cust, CustomerStatusOFAC, repo, searcher, ""); err == nil {
		t.Error("expected error, but got none")
	}
}

func TestCustomers__validCustomerStatusTransitionOFAC(t *testing.T) {
	cust := &client.Customer{
		Id:     base.ID(),
		Status: CustomerStatusReviewRequired,
	}
	repo := &testCustomerRepository{}
	searcher := createTestOFACSearcher(repo, nil)

	repo.ofacSearchResult = &ofacSearchResult{
		sdnName: "Jane Doe",
		match:   0.10,
	}
	if err := validCustomerStatusTransition(cust, CustomerStatusOFAC, repo, searcher, "requestId"); err != nil {
		t.Errorf("unexpected error in OFAC transition: %v", err)
	}

	// OFAC transition with positive match
	repo.ofacSearchResult.match = 0.99
	if err := validCustomerStatusTransition(cust, CustomerStatusOFAC, repo, searcher, "requestId"); err != nil {
		if !strings.Contains(err.Error(), "positive OFAC match") {
			t.Errorf("unexpected error in OFAC transition: %v", err)
		}
	}

	// OFAC transition with no stored result
	repo.ofacSearchResult = nil
	if c, ok := searcher.ofacClient.(*testOFACClient); ok {
		c.sdn = &ofac.Sdn{
			EntityID: "12124",
		}
	}
	if err := validCustomerStatusTransition(cust, CustomerStatusOFAC, repo, searcher, "requestId"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if repo.savedOFACSearchResult.entityId != "12124" {
		t.Errorf("unexpected saved OFAC result: %#v", repo.savedOFACSearchResult)
	}
}

func TestCustomers__updateCustomerAddress(t *testing.T) {
	repo := &testCustomerRepository{
		customer: &client.Customer{
			Id: base.ID(),
		},
	}
	searcher := createTestOFACSearcher(repo, nil)

	svc := admin.NewServer(":10002")
	defer svc.Shutdown()
	addApprovalRoutes(log.NewNopLogger(), svc, repo, searcher)
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

	req := customerRequest{
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
	}
	cust, err := repo.createCustomer(req)
	if err != nil {
		t.Fatal(err)
	}

	// update the Address
	if err := repo.updateCustomerAddress(cust.Id, cust.Addresses[0].Id, "Primary", true); err != nil {
		t.Error(err)
	}

	cust, err = repo.getCustomer(cust.Id)
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
