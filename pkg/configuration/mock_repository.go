// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"github.com/moov-io/customers/pkg/client"
)

type mockRepository struct {
	cfg *client.NamespaceConfiguration
	err error
}

func (r *mockRepository) Get(namespace string) (*client.NamespaceConfiguration, error) {
	if r.err != nil {
		var cfg client.NamespaceConfiguration
		return &cfg, r.err
	}
	return r.cfg, nil
}

func (r *mockRepository) Update(namespace string, cfg *client.NamespaceConfiguration) (*client.NamespaceConfiguration, error) {
	if r.err != nil {
		var cfg client.NamespaceConfiguration
		return &cfg, r.err
	}
	return r.cfg, nil
}
