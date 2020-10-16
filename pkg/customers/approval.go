// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"encoding/json"
	"fmt"
	"net/http"

	moovhttp "github.com/moov-io/base/http"

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/route"

	"github.com/moov-io/base/log"
)

func updateCustomerStatus(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID := route.GetCustomerID(w, r)
		if customerID == "" {
			return
		}

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		var req client.UpdateCustomerStatus
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		cust, err := repo.GetCustomer(customerID, organization)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if cust == nil {
			moovhttp.Problem(w, fmt.Errorf("customerID=%s not found", customerID))
			return
		}

		if err := repo.updateCustomerStatus(customerID, req.Status, req.Comment); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		requestID := moovhttp.GetRequestID(r)
		respondWithCustomer(logger, w, customerID, organization, requestID, repo)
	}
}
