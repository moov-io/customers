// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"github.com/moov-io/customers/pkg/client"
)

type mockRepository struct {
	cfg *client.OrganizationConfiguration
	err error
}

func (r *mockRepository) Get(namespace string) (*client.OrganizationConfiguration, error) {
	if r.err != nil {
		var cfg client.OrganizationConfiguration
		return &cfg, r.err
	}
	return r.cfg, nil
}

func (r *mockRepository) Update(namespace string, cfg *client.OrganizationConfiguration) (*client.OrganizationConfiguration, error) {
	if r.err != nil {
		var cfg client.OrganizationConfiguration
		return &cfg, r.err
	}
	return r.cfg, nil
}
