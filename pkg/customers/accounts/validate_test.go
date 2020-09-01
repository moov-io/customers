// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/moov-io/base"
// 	"github.com/moov-io/customers/cmd/server/paygate"
// 	"github.com/moov-io/customers/pkg/client"
// 	"github.com/moov-io/customers/pkg/secrets"
// 	payclient "github.com/moov-io/paygate/pkg/client"

// 	"github.com/go-kit/kit/log"
// 	"github.com/gorilla/mux"
// )

// func TestRouter__ValidateAccounts(t *testing.T) {
// 	customerID, userID := base.ID(), base.ID()
// 	repo := setupTestAccountRepository(t)
// 	keeper := secrets.TestStringKeeper(t)

// 	paygateClient := &paygate.MockClient{
// 		Micro: &payclient.MicroDeposits{
// 			Amounts: []string{"USD 0.03", "USD 0.07"},
// 			Status:  payclient.PROCESSED,
// 		},
// 	}

// 	handler := mux.NewRouter()
// 	RegisterRoutes(log.NewNopLogger(), handler, repo, testFedClient, paygateClient, keeper, keeper)

// 	// create account
// 	acct, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
// 		AccountNumber: "123",
// 		RoutingNumber: "987654320",
// 		Type:          client.CHECKING,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// make our validation request
// 	var body bytes.Buffer
// 	if err := json.NewEncoder(&body).Encode(client.UpdateValidation{
// 		Strategy:      "micro-deposits",
// 		MicroDeposits: []string{"USD 0.03", "USD 0.07"},
// 	}); err != nil {
// 		t.Fatal(err)
// 	}
// 	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/accounts/%s/validate", customerID, acct.AccountID), &body)

// 	w := httptest.NewRecorder()
// 	handler.ServeHTTP(w, req)
// 	w.Flush()

// 	if w.Code != http.StatusOK {
// 		t.Errorf("bogus HTTP status: %d", w.Code)
// 	}
// }
