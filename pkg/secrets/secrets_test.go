// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package secrets

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSecrets(t *testing.T) {
	ctx := context.Background()
	keeper, err := OpenSecretKeeper(ctx, "foo", "", "")
	if err != nil {
		t.Fatal(err)
	}

	encrypted, err := keeper.Encrypt(ctx, []byte("hello, world"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := keeper.Decrypt(ctx, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if v := string(out); v != "hello, world" {
		t.Errorf("got %q", v)
	}
}

func TestSecrets__OpenLocal(t *testing.T) {
	if _, err := OpenLocal("invalid key"); err == nil {
		t.Error("expected error")
	} else {
		if !strings.Contains(err.Error(), "illegal base64 data") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	keeper, err := testSecretKeeper(testSecretKey)("test-path")
	if err != nil {
		t.Fatal(err)
	}
	enc, err := keeper.Encrypt(context.Background(), []byte("hello, world"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := keeper.Decrypt(context.Background(), enc)
	if err != nil {
		t.Fatal(err)
	}
	if v := string(out); v != "hello, world" {
		t.Errorf("got %q", v)
	}
}

func TestSecrets__OpenLocalURL(t *testing.T) {
	if _, err := testSecretKeeper("base64key://invalid-key")("string-keeper"); err == nil {
		t.Error("expected error")
	}

	keeper, err := testSecretKeeper(testSecretKeyURL)("string-keeper")
	if err != nil {
		t.Error(err)
	}
	str := NewStringKeeper(keeper, 1*time.Second)

	encrypted, err := str.EncryptString("123")
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := str.DecryptString(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != "123" {
		t.Errorf("decrypted=%s", decrypted)
	}
}

func TestStringKeeper__cycle(t *testing.T) {
	keeper, err := testSecretKeeper(testSecretKey)("string-keeper")
	if err != nil {
		t.Fatal(err)
	}

	str := NewStringKeeper(keeper, 1*time.Second)
	defer str.Close()

	encrypted, err := str.EncryptString("123")
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := str.DecryptString(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != "123" {
		t.Errorf("decrypted=%s", decrypted)
	}
}

func TestStringKeeper__nil(t *testing.T) {
	keeper := TestStringKeeper(t)
	keeper.Close()

	keeper = nil

	if _, err := keeper.EncryptString(""); err == nil {
		t.Error("expected error")
	}
	if _, err := keeper.DecryptString(""); err == nil {
		t.Error("expected error")
	}
}

func TestSecrets__TestStringKeeper(t *testing.T) {
	keeper := TestStringKeeper(t)
	if keeper == nil {
		t.Fatal("nil StringKeeper")
	}
	keeper.Close()
}

func TestOpenSecretKeeper(t *testing.T) {
	ctx := context.Background()

	// Just call these and make sure they don't panic.
	//
	// The result depends on env variables, which in TravisCI is different than local.
	require.NotPanics(t, func() { OpenSecretKeeper(ctx, "", "gcp", "") })
	require.NotPanics(t, func() { OpenSecretKeeper(ctx, "", "vault", "") })
	require.NotPanics(t, func() { OpenSecretKeeper(ctx, "", "", "superSecretKey") })
	require.NotPanics(t, func() { OpenSecretKeeper(ctx, "", "local", "superSecretKey") })
	_, err := OpenSecretKeeper(ctx, "", "foo", "")
	require.Error(t, err)
}
