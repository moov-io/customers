// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"encoding/json"
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
)

// AddApprovalRoutes contains "back office" admin endpoints used to validate (or reject) a Customer
// TODO(adam): We need to hide these behind an admin level auth, but we'll write them for now.
// What about a header like x-admin-id ??
func AddApprovalRoutes(logger log.Logger, svc *admin.Server, repo CustomerRepository, customerSSNRepo CustomerSSNRepository, ofac *OFACSearcher) {
	svc.AddHandler("/customers/{customerID}/status", updateCustomerStatus(logger, repo, customerSSNRepo, ofac))
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

func updateCustomerStatus(logger log.Logger, repo CustomerRepository, customerSSNRepo CustomerSSNRepository, ofac *OFACSearcher) http.HandlerFunc {
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
