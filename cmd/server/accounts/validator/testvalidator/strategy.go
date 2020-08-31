package testvalidator

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
)

type testStrategy struct{}

func NewStrategy() validator.Strategy {
	return &testStrategy{}
}

func (t *testStrategy) InitAccountValidation(userID, accountID, customerID string) (*validator.VendorResponse, error) {
	return &validator.VendorResponse{
		"result": "success",
	}, nil
}

type completeAccountValidationRequest struct {
	Result string
}

func (t *testStrategy) CompleteAccountValidation(userID, accountID, customerID string, request *validator.VendorRequest) (*validator.VendorResponse, error) {
	input := &completeAccountValidationRequest{}
	if err := mapstructure.Decode(request, input); err != nil {
		return nil, fmt.Errorf("unable to parse request params: %v", err)
	}

	if input.Result != "success" {
		return nil, errors.New("account validation failed (test strategy)")
	}

	return &validator.VendorResponse{
		"result": "success",
	}, nil
}
