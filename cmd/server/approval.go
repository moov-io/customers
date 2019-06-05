// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/moov-io/base/admin"
	moovhttp "github.com/moov-io/base/http"
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	errNoAddressId = errors.New("no Address ID found")
)

// addApprovalRoutes contains "back office" endpoints used to validate (or reject) a Customer
// TODO(adam): We need to hide these behind an admin level auth, but we'll write them for now 'x-admin-id' ??
func addApprovalRoutes(logger log.Logger, svc *admin.Server, repo customerRepository) {
	svc.AddHandler("/customers/{customerId}/status", updateCustomerStatus(logger, repo))
	svc.AddHandler("/customers/{customerId}/addresses/{addressId}", updateCustomerAddress(logger, repo))
}

type updateCustomerStatusRequest struct {
	Comment string         `json:"comment,omitempty"`
	Status  CustomerStatus `json:"status"`
}

// validCustomerStatusTransition determines if a future CustomerStatus is valid for a given
// Customer. There are several rules which apply to a CustomerStatus, such as:
//  - Deceased, Rejected statuses can never be changed
//  - KYC is only valid if the Customer has first, last, address, and date of birth
//  - OFAC can only be after an OFAC search has been performed (and search info recorded)
//  - CIP can only be if the SSN has been set
func validCustomerStatusTransition(existing *client.Customer, futureStatus CustomerStatus) error {
	eql := func(s string, status CustomerStatus) bool {
		return strings.EqualFold(s, string(status))
	}
	// Check Deceased and Rejected
	if eql(existing.Status, CustomerStatusDeceased) || eql(existing.Status, CustomerStatusRejected) {
		return fmt.Errorf("customer status '%s' cannot be changed", existing.Status)
	}
	switch futureStatus {
	case CustomerStatusKYC:
		if existing.FirstName == "" || existing.LastName == "" {
			return fmt.Errorf("customer=%s is missing fist/last name", existing.Id)
		}
		if existing.BirthDate.IsZero() {
			return fmt.Errorf("customer=%s is missing date of birth", existing.Id)
		}
		if !containsValidPrimaryAddress(existing.Addresses) {
			return fmt.Errorf("customer=%s is missing a valid primary Address", existing.Id)
		}
	case CustomerStatusOFAC: // TODO(adam): need to impl lookup
		// I think we should perform the OFAC search when requested and store the highest match EntityId, name and match % in a new database table.
		// Then it's a valid transition only when a record exists and is below the threshold.
		return fmt.Errorf("customers=%s %s to OFAC transition needs to lookup OFAC search results", existing.Id, existing.Status)
	case CustomerStatusCIP: // TODO(adam): need to impl lookup
		// What can we do to validate an SSN?
		// https://www.ssa.gov/employer/randomization.html (not much)
		return fmt.Errorf("customers=%s %s to CIP transition needs to lookup encrypted SSN", existing.Id, existing.Status)
	}
	return nil
}

func containsValidPrimaryAddress(addrs []client.Address) bool {
	for i := range addrs {
		if strings.EqualFold(addrs[i].Type, "primary") && addrs[i].Validated {
			return true
		}
	}
	return false
}

func updateCustomerStatus(logger log.Logger, repo customerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)

		if r.Method != "PUT" {
			moovhttp.Problem(w, fmt.Errorf("unsupported HTTP verb %s", r.Method))
			return
		}

		customerId, requestId := getCustomerId(w, r), moovhttp.GetRequestId(r)
		if customerId == "" {
			return
		}

		var req updateCustomerStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		cust, err := repo.getCustomer(customerId)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := validCustomerStatusTransition(cust, req.Status); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// Update Customer's status in the database
		if err := repo.updateCustomerStatus(customerId, req.Status, req.Comment); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		respondWithCustomer(logger, w, customerId, requestId, repo)
	}
}

func getAddressId(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["addressId"]
	if !ok || v == "" {
		moovhttp.Problem(w, errNoAddressId)
		return ""
	}
	return v
}

// TODO(adam): Should Addresses have a 'Type: Previous'? I don't think we ever want to delete an address, but it can be marked as old.
// PUT /customers/{customerId}/addresses/{addressId} only accept {"type": "Primary/Secondary", "validated": true/false}

type updateCustomerAddressRequest struct {
	Type      string `json:"type"`
	Validated bool   `json:"validated"`
}

func (req *updateCustomerAddressRequest) validate() error {
	switch strings.ToLower(req.Type) {
	case "primary", "secondary":
		return nil
	default:
		return fmt.Errorf("updateCustomerAddressRequest: unknown type: %s", req.Type)
	}
}

func updateCustomerAddress(logger log.Logger, repo customerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)

		if r.Method != "PUT" {
			moovhttp.Problem(w, fmt.Errorf("unsupported HTTP verb %s", r.Method))
			return
		}

		customerId, addressId := getCustomerId(w, r), getAddressId(w, r)
		if customerId == "" || addressId == "" {
			return
		}

		var req updateCustomerAddressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := req.validate(); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		requestId := moovhttp.GetRequestId(r)
		logger.Log("approval", fmt.Sprintf("updating address=%s for customer=%s", addressId, customerId), "requestId", requestId)

		if err := repo.updateCustomerAddress(customerId, addressId, req.Type, req.Validated); err != nil {
			logger.Log("approval", fmt.Sprintf("error updating customer=%s address=%s: %v", customerId, addressId, err), "requestId", requestId)
			moovhttp.Problem(w, err)
			return
		}
		respondWithCustomer(logger, w, customerId, requestId, repo)
	}
}
