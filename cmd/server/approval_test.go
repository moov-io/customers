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
	"github.com/moov-io/customers"
	client "github.com/moov-io/customers/client"
	watchman "github.com/moov-io/watchman/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func TestCustomers__updateCustomerStatus(t *testing.T) {
	repo := &testCustomerRepository{
		customer: &client.Customer{
			ID:     base.ID(),
			Status: "none",
		},
	}
	searcher := createTestOFACSearcher(repo, nil)
	ssnRepo := &testCustomerSSNRepository{}

	svc := admin.NewServer(":10001")
	defer svc.Shutdown()
	addApprovalRoutes(log.NewNopLogger(), svc, repo, ssnRepo, searcher)
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
	if customer.ID == "" {
		t.Errorf("missing customer JSON: %#v", customer)
	}
	if repo.updatedStatus != customers.ReviewRequired {
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
		ID:     base.ID(),
		Status: customers.None.String(),
	}
	repo := &testCustomerRepository{}
	searcher := createTestOFACSearcher(repo, nil)

	ssn := &SSN{customerID: cust.ID, encrypted: []byte("secret")}

	if err := validCustomerStatusTransition(cust, ssn, customers.Deceased, repo, searcher, "requestID"); err != nil {
		t.Errorf("expected no error: %v", err)
	}

	// block Deceased and Rejected customers
	cust.Status = customers.Deceased.String()
	if err := validCustomerStatusTransition(cust, ssn, customers.KYC, repo, searcher, "requestID"); err == nil {
		t.Error("expected error")
	}
	cust.Status = customers.Rejected.String()
	if err := validCustomerStatusTransition(cust, ssn, customers.KYC, repo, searcher, "requestID"); err == nil {
		t.Error("expected error")
	}

	// normal KYC approval (rejected due to missing info)
	cust.FirstName, cust.LastName = "Jane", "Doe"
	cust.Status = customers.ReviewRequired.String()
	if err := validCustomerStatusTransition(cust, ssn, customers.KYC, repo, searcher, "requestID"); err == nil {
		t.Error("expected error")
	}
	cust.BirthDate = time.Now()
	if err := validCustomerStatusTransition(cust, ssn, customers.KYC, repo, searcher, "requestID"); err == nil {
		t.Error("expected error")
	}
	cust.Addresses = append(cust.Addresses, client.Address{
		Type:     "primary",
		Address1: "123 1st st",
	})

	// CIP transistions are WIP // TODO(adam):
	cust.Status = customers.ReviewRequired.String()
	if err := validCustomerStatusTransition(cust, nil, customers.CIP, repo, searcher, "requestID"); err != nil {
		if !strings.Contains(err.Error(), "is missing SSN") {
			t.Errorf("CIP: unexpected error: %v", err)
		}
	} else {
		t.Error("CIP transition is WIP")
	}
	if err := validCustomerStatusTransition(cust, ssn, customers.CIP, repo, searcher, "requestID"); err == nil {
		t.Error("CIP transition is WIP")
	}
}

func TestCustomers__validCustomerStatusTransitionError(t *testing.T) {
	cust := &client.Customer{
		ID:     base.ID(),
		Status: customers.ReviewRequired.String(),
	}
	repo := &testCustomerRepository{}
	client := &testWatchmanClient{}
	searcher := createTestOFACSearcher(repo, client)

	ssn := &SSN{customerID: cust.ID, encrypted: []byte("secret")}

	repo.err = errors.New("bad error")
	if err := validCustomerStatusTransition(cust, ssn, customers.OFAC, repo, searcher, ""); err == nil {
		t.Error("expected error, but got none")
	}
	repo.err = nil

	client.err = errors.New("bad error")
	if err := validCustomerStatusTransition(cust, ssn, customers.OFAC, repo, searcher, ""); err == nil {
		t.Error("expected error, but got none")
	}
}

func TestCustomers__validCustomerStatusTransitionOFAC(t *testing.T) {
	cust := &client.Customer{
		ID:     base.ID(),
		Status: customers.ReviewRequired.String(),
	}
	repo := &testCustomerRepository{}
	searcher := createTestOFACSearcher(repo, nil)

	ssn := &SSN{customerID: cust.ID, encrypted: []byte("secret")}

	repo.ofacSearchResult = &ofacSearchResult{
		SDNName: "Jane Doe",
		Match:   0.10,
	}
	if err := validCustomerStatusTransition(cust, ssn, customers.OFAC, repo, searcher, "requestID"); err != nil {
		t.Errorf("unexpected error in OFAC transition: %v", err)
	}

	// OFAC transition with positive match
	repo.ofacSearchResult.Match = 0.99
	if err := validCustomerStatusTransition(cust, ssn, customers.OFAC, repo, searcher, "requestID"); err != nil {
		if !strings.Contains(err.Error(), "positive OFAC match") {
			t.Errorf("unexpected error in OFAC transition: %v", err)
		}
	}

	// OFAC transition with no stored result
	repo.ofacSearchResult = nil
	if c, ok := searcher.watchmanClient.(*testWatchmanClient); ok {
		c.sdn = &watchman.OfacSdn{
			EntityID: "12124",
		}
	}
	if err := validCustomerStatusTransition(cust, ssn, customers.OFAC, repo, searcher, "requestID"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if repo.savedOFACSearchResult.EntityId != "12124" {
		t.Errorf("unexpected saved OFAC result: %#v", repo.savedOFACSearchResult)
	}
}

func TestCustomers__updateCustomerAddress(t *testing.T) {
	repo := &testCustomerRepository{
		customer: &client.Customer{
			ID: base.ID(),
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
	}).asCustomer(testCustomerSSNStorage)
	if err := repo.createCustomer(cust); err != nil {
		t.Fatal(err)
	}

	// update the Address
	if err := repo.updateCustomerAddress(cust.ID, cust.Addresses[0].ID, "Primary", true); err != nil {
		t.Error(err)
	}

	cust, err := repo.getCustomer(cust.ID)
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
