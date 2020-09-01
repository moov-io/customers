// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"

// 	moovhttp "github.com/moov-io/base/http"
// 	"github.com/moov-io/customers/pkadmin"
// 	"github.com/moov-io/customers/cmd/server/route"

// 	"github.com/go-kit/kit/log"
// )

// func updateAccountStatus(logger log.Logger, repo Repository) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		w = route.Responder(logger, w, r)
// 		w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 		if r.Method != "PUT" {
// 			moovhttp.Problem(w, fmt.Errorf("unsupported HTTP verb %s", r.Method))
// 			return
// 		}

// 		accountID := getAccountID(w, r)
// 		if accountID == "" {
// 			return
// 		}

// 		var req admin.UpdateAccountStatus
// 		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 			moovhttp.Problem(w, err)
// 			return
// 		}
// 		switch req.Status {
// 		case admin.NONE, admin.VALIDATED:
// 			// do nothing
// 		default:
// 			moovhttp.Problem(w, fmt.Errorf("invalid status: %s", req.Status))
// 			return
// 		}

// 		if err := repo.updateAccountStatus(accountID, req.Status); err != nil {
// 			moovhttp.Problem(w, err)
// 			return
// 		}

// 		w.WriteHeader(http.StatusOK)
// 	}
// }
