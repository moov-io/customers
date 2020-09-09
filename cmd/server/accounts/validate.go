// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/cmd/server/paygate"
	"github.com/moov-io/customers/cmd/server/route"
	"github.com/moov-io/customers/pkg/client"

	"github.com/go-kit/kit/log"
)

func validateAccount(logger log.Logger, repo Repository, paygateClient paygate.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, accountID := route.GetCustomerID(w, r), getAccountID(w, r)
		if customerID == "" || accountID == "" {
			return
		}

		// Lookup the account and verify it needs to be validated
		account, err := repo.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if !strings.EqualFold(string(account.Status), string(client.NONE)) {
			moovhttp.Problem(w, fmt.Errorf("unexpected accountID=%s status=%s", accountID, account.Status))
			return
		}

		var req client.UpdateValidation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, fmt.Errorf("unable to read UpdateValidation: %v", err))
			return
		}

		switch req.Strategy {
		case "micro-deposits":
			userID := moovhttp.GetUserID(r)
			if err := handleMicroDepositValidation(repo, paygateClient, accountID, customerID, userID, req.MicroDeposits); err != nil {
				moovhttp.Problem(w, err)
				return
			}

		default:
			moovhttp.Problem(w, fmt.Errorf("unknown strategy %s", req.Strategy))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
