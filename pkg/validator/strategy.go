// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package validator

import "github.com/moov-io/customers/pkg/client"

type StrategyKey struct {
	Strategy string
	Vendor   string
}

type Strategy interface {
	InitAccountValidation(userID, accountID, customerID string) (*VendorResponse, error)
	CompleteAccountValidation(userID, customerID string, account *client.Account, accountNumber string, request *VendorRequest) (*VendorResponse, error)
}

type VendorRequest map[string]interface{}
type VendorResponse map[string]interface{}
