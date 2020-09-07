// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package mx

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/mxenabled/atrium-go"
)

type mxStrategy struct {
	client *atrium.APIClient
}

type StrategyOptions struct {
	ClientID string
	APIKey   string
}

func NewStrategy(options StrategyOptions) validator.Strategy {
	client := atrium.AtriumClient(options.APIKey, options.ClientID)

	return &mxStrategy{
		client: client,
	}
}

func (s *mxStrategy) InitAccountValidation(userID, accountID, customerID string) (*validator.VendorResponse, error) {
	ctx := context.Background()

	body := atrium.UserCreateRequestBody{
		User: &atrium.User{
			Identifier: customerID,
		},
	}

	response, _, err := s.client.Users.CreateUser(ctx, body)
	if err != nil {
		return nil, err
	}

	widgetBody := atrium.ConnectWidgetRequestBody{}
	widgetResponse, _, err := s.client.ConnectWidget.GetConnectWidget(ctx, response.User.GUID, widgetBody)
	if err != nil {
		return nil, err
	}

	return &validator.VendorResponse{
		"connect_widget_url": widgetResponse.User.ConnectWidgetURL,
	}, nil
}

type completeAccountValidationRequest struct {
	MemberGUID string `json:"member_guid" mapstructure:"member_guid"`
	UserGUID   string `json:"user_guid" mapstructure:"user_guid"`
}

func (s *mxStrategy) CompleteAccountValidation(userID, customerID string, account *client.Account, accountNumber string, request *validator.VendorRequest) (*validator.VendorResponse, error) {
	input := &completeAccountValidationRequest{}
	if err := mapstructure.Decode(request, input); err != nil {
		return nil, fmt.Errorf("unable to parse request params: %v", err)
	}

	ctx := context.Background()
	response, _, err := s.client.Verification.ListAccountNumbers(ctx, input.MemberGUID, input.UserGUID)

	if err != nil {
		return nil, err
	}

	// look for account number and routing number in MX results
	for _, acc := range response.AccountNumbers {
		if acc.AccountNumber == accountNumber && acc.RoutingNumber == account.RoutingNumber {
			return &validator.VendorResponse{
				"result": "validated",
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to validate account=%s", account.AccountID)
}
