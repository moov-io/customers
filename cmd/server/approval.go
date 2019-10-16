// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/moov-io/base/admin"
	moovhttp "github.com/moov-io/base/http"
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	ofacMatchThreshold float32 = func() float32 {
		if v := os.Getenv("OFAC_MATCH_THRESHOLD"); v != "" {
			f, err := strconv.ParseFloat(v, 32)
			if err == nil && f > 0.00 {
				return float32(f)
			}
		}
		return 0.99 // default, 99%
	}()

	errNoAddressId = errors.New("no Address ID found")
)

// addApprovalRoutes contains "back office" admin endpoints used to validate (or reject) a Customer
// TODO(adam): We need to hide these behind an admin level auth, but we'll write them for now.
// What about a header like x-admin-id ??
func addApprovalRoutes(logger log.Logger, svc *admin.Server, repo customerRepository, ofac *ofacSearcher) {
	svc.AddHandler("/customers/{customerID}/status", updateCustomerStatus(logger, repo, ofac))
	svc.AddHandler("/customers/{customerID}/addresses/{addressId}", updateCustomerAddress(logger, repo))
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
func validCustomerStatusTransition(existing *client.Customer, futureStatus CustomerStatus, repo customerRepository, ofac *ofacSearcher, requestID string) error {
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
			return fmt.Errorf("customer=%s is missing fist/last name", existing.ID)
		}
		if existing.BirthDate.IsZero() {
			return fmt.Errorf("customer=%s is missing date of birth", existing.ID)
		}
		if !containsValidPrimaryAddress(existing.Addresses) {
			return fmt.Errorf("customer=%s is missing a valid primary Address", existing.ID)
		}
	case CustomerStatusOFAC:
		searchResult, err := repo.getLatestCustomerOFACSearch(existing.ID)
		if err != nil {
			return fmt.Errorf("validCustomerStatusTransition: error getting OFAC search: %v", err)
		}
		if searchResult == nil {
			if err := ofac.storeCustomerOFACSearch(existing, ""); err != nil {
				return fmt.Errorf("validCustomerStatusTransition: problem with OFAC search: %v", err)
			}
			searchResult, err = repo.getLatestCustomerOFACSearch(existing.ID)
			if err != nil || searchResult == nil {
				return fmt.Errorf("validCustomerStatusTransition: inner lookup searchResult=%#v: %v", searchResult, err)
			}
		}
		if searchResult.match > ofacMatchThreshold {
			return fmt.Errorf("validCustomerStatusTransition: customer=%s has positive OFAC match (%.2f) with SDN=%s", existing.ID, searchResult.match, searchResult.entityId)
		}
		return nil
	case CustomerStatusCIP: // TODO(adam): need to impl lookup
		// What can we do to validate an SSN?
		// https://www.ssa.gov/employer/randomization.html (not much)
		return fmt.Errorf("customers=%s %s to CIP transition needs to lookup encrypted SSN", existing.ID, existing.Status)
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

func updateCustomerStatus(logger log.Logger, repo customerRepository, ofac *ofacSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)

		if r.Method != "PUT" {
			moovhttp.Problem(w, fmt.Errorf("unsupported HTTP verb %s", r.Method))
			return
		}

		customerID, requestID := getCustomerID(w, r), moovhttp.GetRequestID(r)
		if customerID == "" {
			return
		}

		var req updateCustomerStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		cust, err := repo.getCustomer(customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := validCustomerStatusTransition(cust, req.Status, repo, ofac, requestID); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// Update Customer's status in the database
		if err := repo.updateCustomerStatus(customerID, req.Status, req.Comment); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		respondWithCustomer(logger, w, customerID, requestID, repo)
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
// If we keep address info around does it have GDPR implications?
// PUT /customers/{customerID}/addresses/{addressId} only accept {"type": "Primary/Secondary", "validated": true/false}

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

		customerID, addressId := getCustomerID(w, r), getAddressId(w, r)
		if customerID == "" || addressId == "" {
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

		requestID := moovhttp.GetRequestID(r)
		logger.Log("approval", fmt.Sprintf("updating address=%s for customer=%s", addressId, customerID), "requestID", requestID)

		if err := repo.updateCustomerAddress(customerID, addressId, req.Type, req.Validated); err != nil {
			logger.Log("approval", fmt.Sprintf("error updating customer=%s address=%s: %v", customerID, addressId, err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		respondWithCustomer(logger, w, customerID, requestID, repo)
	}
}
