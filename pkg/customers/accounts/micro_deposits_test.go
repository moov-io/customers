// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

// import (
// 	"errors"
// 	"strings"
// 	"testing"

// 	"github.com/moov-io/base"
// 	"github.com/moov-io/customers/cmd/server/paygate"
// 	"github.com/moov-io/paygate/pkg/client"
// )

// func TestAccounts__handleMicroDepositValidation(t *testing.T) {
// 	repo := &mockRepository{}
// 	paygateClient := &paygate.MockClient{}

// 	accountID := base.ID()
// 	customerID := base.ID()
// 	userID := base.ID()

// 	var requestAmounts []string

// 	// initiate micro-deposits
// 	err := handleMicroDepositValidation(repo, paygateClient, accountID, customerID, userID, requestAmounts)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// validate amounts
// 	requestAmounts = []string{"USD 0.04", "USD 0.09"}
// 	paygateClient.Micro = &client.MicroDeposits{
// 		MicroDepositID: base.ID(),
// 		Amounts:        requestAmounts,
// 		Status:         client.PROCESSED,
// 	}
// 	err = handleMicroDepositValidation(repo, paygateClient, accountID, customerID, userID, requestAmounts)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestAccounts__RejectPendingMicroDeposits(t *testing.T) {
// 	repo := &mockRepository{}

// 	accountID, customerID := base.ID(), base.ID()
// 	userID := base.ID()

// 	requestAmounts := []string{"USD 0.04", "USD 0.09"}
// 	paygateClient := &paygate.MockClient{
// 		Micro: &client.MicroDeposits{
// 			MicroDepositID: base.ID(),
// 			Amounts:        requestAmounts,
// 			Status:         client.PENDING,
// 		},
// 	}

// 	// initiate micro-deposits
// 	err := handleMicroDepositValidation(repo, paygateClient, accountID, customerID, userID, requestAmounts)
// 	if err == nil {
// 		t.Fatal("expected error")
// 	}
// 	if !strings.Contains(err.Error(), "is in status: pending") {
// 		t.Fatal(err)
// 	}
// }

// func TestAccounts__handleMicroDepositValidationErr(t *testing.T) {
// 	repo := &mockRepository{}
// 	paygateClient := &paygate.MockClient{
// 		Err: errors.New("bad error"),
// 	}

// 	if err := handleMicroDepositValidation(repo, paygateClient, "", "", "", nil); err == nil {
// 		t.Error("expected error")
// 	}
// }

// func TestAccounts__validateAmounts(t *testing.T) {
// 	micro := &client.MicroDeposits{
// 		Amounts: []string{"USD 0.04", "USD 0.09"},
// 	}
// 	requestAmounts := []string{"USD 0.04", "USD 0.09"}

// 	if err := validateAmounts(micro, requestAmounts); err != nil {
// 		t.Error(err)
// 	}

// 	requestAmounts = []string{"USD 0.04"}
// 	if err := validateAmounts(micro, requestAmounts); err != nil {
// 		if !strings.Contains(err.Error(), "invalid number of micro-deposits") {
// 			t.Error(err)
// 		}
// 	}

// 	requestAmounts = []string{"USD 0.04", "USD 0.07"}
// 	if err := validateAmounts(micro, requestAmounts); err != nil {
// 		if !strings.Contains(err.Error(), "incorrect micro-deposit") {
// 			t.Error(err)
// 		}
// 	}
// }

// func TestAccounts__validateAmountsErr(t *testing.T) {
// 	if err := validateAmounts(nil, nil); err == nil {
// 		t.Error("expected error")
// 	}

// 	micro := &client.MicroDeposits{
// 		Amounts: []string{"USD 0.04"},
// 	}
// 	if err := validateAmounts(micro, nil); err == nil {
// 		t.Error("expected error")
// 	}
// }
