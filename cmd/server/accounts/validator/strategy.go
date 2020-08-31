package validator

import "github.com/moov-io/customers/client"

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
