// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package fed

import (
	"github.com/moov-io/customers/client"
)

type MockClient struct {
	Err     error
	Details *client.InstitutionDetails
}

func (c *MockClient) Ping() error {
	return c.Err
}

func (c *MockClient) LookupInstitution(routingNumber string) (*client.InstitutionDetails, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.Details, nil
}
