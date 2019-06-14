// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"testing"

	"gocloud.dev/secrets"
)

var (
	testSecretKeeper secretFunc = func(_ string) (*secrets.Keeper, error) {
		return openLocal()
	}
)

func TestSecrets(t *testing.T) {
	// We assume CLOUD_PROVIDER is unset
	keeper, err := getSecretKeeper("foo")
	if err != nil {
		t.Fatal(err)
	}

	encrypted, err := keeper.Encrypt(context.Background(), []byte("hello, world"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := keeper.Decrypt(context.Background(), encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if v := string(out); v != "hello, world" {
		t.Errorf("got %q", v)
	}
}
