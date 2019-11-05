// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers"
	client "github.com/moov-io/customers/client"
	ofac "github.com/moov-io/ofac/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type ofacSearchResult struct {
	entityId string
	sdnName  string
	sdnType  string
	match    float32
}

type ofacSearcher struct {
	repo       customerRepository
	ofacClient OFACClient
}

// storeCustomerOFACSearch performs OFAC searches against the Customer's name and nickname if populated.
// The higher matching search result is stored in s.customerRepository for use later (in approvals)
func (s *ofacSearcher) storeCustomerOFACSearch(cust *client.Customer, requestID string) error {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	sdn, err := s.ofacClient.Search(ctx, formatCustomerName(cust), requestID)
	if err != nil {
		return fmt.Errorf("ofacSearcher.storeCustomerOFACSearch: name search for customer=%s: %v", cust.ID, err)
	}
	var nickSDN *ofac.Sdn
	if cust.NickName != "" {
		nickSDN, err = s.ofacClient.Search(ctx, cust.NickName, requestID)
		if err != nil {
			return fmt.Errorf("ofacSearcher.storeCustomerOFACSearch: nickname search for customer=%s: %v", cust.ID, err)
		}
	}
	// Save the higher matching SDN (from name search or nick name)
	switch {
	case nickSDN != nil && nickSDN.Match > sdn.Match:
		err = s.repo.saveCustomerOFACSearch(cust.ID, ofacSearchResult{
			entityId: nickSDN.EntityID,
			sdnName:  nickSDN.SdnName,
			sdnType:  nickSDN.SdnType,
			match:    nickSDN.Match,
		})
	case sdn != nil:
		err = s.repo.saveCustomerOFACSearch(cust.ID, ofacSearchResult{
			entityId: sdn.EntityID,
			sdnName:  sdn.SdnName,
			sdnType:  sdn.SdnType,
			match:    sdn.Match,
		})
	}
	if err != nil {
		return fmt.Errorf("ofacSearcher.storeCustomerOFACSearch: saveCustomerOFACSearch customer=%s: %v", cust.ID, err)
	}
	return nil
}

func addOFACRoutes(logger log.Logger, r *mux.Router, repo customerRepository, ofac *ofacSearcher) {
	r.Methods("PUT").Path("/customers/{customerID}/refresh/ofac").HandlerFunc(refreshOFACSearch(logger, repo, ofac))
}

func refreshOFACSearch(logger log.Logger, repo customerRepository, ofac *ofacSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		requestID, userID := moovhttp.GetRequestID(r), moovhttp.GetUserID(r)
		customerID := getCustomerID(w, r)
		if customerID == "" {
			return
		}

		cust, err := repo.getCustomer(customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		logger.Log("ofac", fmt.Sprintf("running live OFAC search for customer=%s", customerID), "requestID", requestID, "userID", userID)

		if err := ofac.storeCustomerOFACSearch(cust, requestID); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		result, err := repo.getLatestCustomerOFACSearch(customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if result.match > ofacMatchThreshold {
			err = fmt.Errorf("customer=%s matched against OFAC entity=%s with a score of %.2f - rejecting customer", cust.ID, result.entityId, result.match)
			logger.Log("ofac", err.Error(), "requestID", requestID, "userID", userID)

			cust.Status = customers.Rejected.String()
			if err := repo.updateCustomerStatus(cust.ID, customers.Rejected, "manual OFAC refresh"); err != nil {
				moovhttp.Problem(w, err)
				return
			}

			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cust)
	}
}
