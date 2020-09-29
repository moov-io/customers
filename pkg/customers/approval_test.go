// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

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
	searcher := createTestOFACSearcher(repo, nil)
	ssnRepo := &testCustomerSSNRepository{}

	svc := admin.NewServer(":10001")
	defer svc.Shutdown()
	AddApprovalRoutes(log.NewNopLogger(), svc, repo, ssnRepo, searcher)
	go svc.Listen()

	body := strings.NewReader(`{"status": "ReceiveOnly", "comment": "test comment"}`)
	req, err := http.NewRequest("PUT", "http://localhost:10001/customers/foo/status", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("x-organization", "test")
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
