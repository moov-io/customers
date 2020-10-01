// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/moov-io/base/admin"
	"github.com/moov-io/customers/pkg/validator"
	"github.com/stretchr/testify/require"
)

func TestMain__setupSigner(t *testing.T) {
	logger := log.NewNopLogger()

	signer := setupSigner(logger, "", "file")
	if signer == nil {
		t.Fatal("expected non-nil Signer")
	}

	signer = setupSigner(logger, "", "other")
	if signer != nil {
		t.Fatal("expected nil Signer")
	}
}

func TestMain__setupValidationStrategies(t *testing.T) {
	logger := log.NewNopLogger()
	adminServer := admin.NewServer(*adminAddr)

	strategies, err := setupValidationStrategies(logger, adminServer)
	require.NoError(t, err)

	if os.Getenv("PLAID_CLIENT_ID") != "" {
		_, found := strategies[validator.StrategyKey{Strategy: "instant", Vendor: "plaid"}]
		require.True(t, found)
	}

	if os.Getenv("ATRIUM_CLIENT_ID") != "" {
		_, found := strategies[validator.StrategyKey{Strategy: "instant", Vendor: "mx"}]
		require.True(t, found)
	}

	// microdeposits / moov for now should always be there
	_, found := strategies[validator.StrategyKey{Strategy: "micro-deposits", Vendor: "moov"}]
	require.True(t, found)

}
