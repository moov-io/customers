// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package paygate

import (
	"github.com/moov-io/paygate/pkg/client"
)

type MockClient struct {
	Micro *client.MicroDeposits
	Err   error
}

func (c *MockClient) Ping() error {
	return c.Err
}

func (c *MockClient) GetMicroDeposits(accountID, userID string) (*client.MicroDeposits, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.Micro, nil
}

func (c *MockClient) InitiateMicroDeposits(userID string, destination client.Destination) error {
	return c.Err
}
