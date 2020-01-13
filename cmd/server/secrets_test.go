// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"gocloud.dev/secrets"
)

var (
	testSecretKey    = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("1"), 32))
	testSecretKeeper = func(base64Key string) secretFunc {
		return func(path string) (*secrets.Keeper, error) {
			return openLocal(base64Key)
		}
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

func TestSecrets__openLocal(t *testing.T) {
	if _, err := openLocal("invalid key"); err == nil {
		t.Error("expected error")
	} else {
		if !strings.Contains(err.Error(), "SECRETS_LOCAL_BASE64_KEY") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}
