// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package secrets

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"gocloud.dev/secrets"
)

type secretFunc func(path string) (*secrets.Keeper, error)

var (
	testSecretKey    = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("1"), 32))
	testSecretKeyURL = fmt.Sprintf("base64key://%s", testSecretKey)
	testSecretKeeper = func(base64Key string) secretFunc {
		return func(path string) (*secrets.Keeper, error) {
			return OpenLocal(base64Key)
		}
	}
)

func TestStringKeeper(t *testing.T) *StringKeeper {
	t.Helper()
	keeper, err := testSecretKeeper(testSecretKey)("string-keeper")
	if err != nil {
		t.Fatal(err)
	}
	return NewStringKeeper(keeper, 1*time.Second)
}

func TestKeeper(t *testing.T) *secrets.Keeper {
	keeper, err := testSecretKeeper(testSecretKey)("test-keeper")
	if err != nil {
		t.Fatal(err)
	}
	return keeper
}
