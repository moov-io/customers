// Copyright 2020 The Moov Authors
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
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/route"

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
func addApprovalRoutes(logger log.Logger, svc *admin.Server, repo customerRepository, customerSSNRepo customerSSNRepository, ofac *ofacSearcher) {
	svc.AddHandler("/customers/{customerID}/status", updateCustomerStatus(logger, repo, customerSSNRepo, ofac))
	svc.AddHandler("/customers/{customerID}/addresses/{addressId}", updateCustomerAddress(logger, repo))
}

type updateCustomerStatusRequest struct {
	Comment string                `json:"comment,omitempty"`
	Status  client.CustomerStatus `json:"status"`
}

func containsValidPrimaryAddress(addrs []client.CustomerAddress) bool {
	for i := range addrs {
		if strings.EqualFold(addrs[i].Type, "primary") && addrs[i].Validated {
			return true
		}
	}
	return false
}

func updateCustomerStatus(logger log.Logger, repo customerRepository, customerSSNRepo customerSSNRepository, ofac *ofacSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		if r.Method != "PUT" {
			moovhttp.Problem(w, fmt.Errorf("unsupported HTTP verb %s", r.Method))
			return
		}

		customerID, requestID := route.GetCustomerID(w, r), moovhttp.GetRequestID(r)
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
		if cust == nil {
			moovhttp.Problem(w, fmt.Errorf("customerID=%s not found", customerID))
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
		w = route.Responder(logger, w, r)

		if r.Method != "PUT" {
			moovhttp.Problem(w, fmt.Errorf("unsupported HTTP verb %s", r.Method))
			return
		}

		customerID, addressId := route.GetCustomerID(w, r), getAddressId(w, r)
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
