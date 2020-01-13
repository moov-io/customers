// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"testing"

	"github.com/go-kit/kit/log"
)

func TestSetupStorageBucket(t *testing.T) {
	logger := log.NewNopLogger()

	signer := setupStorageBucket(logger, "", "file")
	if signer == nil {
		t.Fatal("expected non-nil Signer")
	}

	signer = setupStorageBucket(logger, "", "other")
	if signer != nil {
		t.Fatal("expected nil Signer")
	}
}
