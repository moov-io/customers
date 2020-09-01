// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package fed

import (
	"errors"
	"testing"

	"github.com/moov-io/customers/pkg/client"
)

func TestCache(t *testing.T) {
	mock := &MockClient{}
	cache := NewCacheClient(mock, 2)

	// Lookup, but find nothing
	details, err := cache.LookupInstitution("987654320")
	if err != nil {
		t.Fatal(err)
	}
	if details != nil {
		t.Errorf("unexpected InstitutionDetails: %#v", details)
	}

	// set details and cache
	mock.Details = &client.InstitutionDetails{
		RoutingNumber: "987654320",
	}
	details, err = cache.LookupInstitution("987654320")
	if err != nil {
		t.Fatal(err)
	}
	if details == nil {
		t.Error("expected details")
	}

	// read details again (from cache)
	details, err = cache.LookupInstitution("987654320")
	if err != nil {
		t.Fatal(err)
	}
	if details == nil {
		t.Error("expected details")
	}
}

func TestCachePing(t *testing.T) {
	mock := &MockClient{}
	cache := NewCacheClient(mock, 2)

	// Healthy
	if err := cache.Ping(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Healthy
	mock.Err = errors.New("bad error")
	if err := cache.Ping(); err == nil {
		t.Error("expected error")
	}
}
