// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	"gocloud.dev/secrets"
	"gocloud.dev/secrets/gcpkms"
	"gocloud.dev/secrets/localsecrets"
	"gocloud.dev/secrets/vault"
)

type secretFunc func(path string) (*secrets.Keeper, error)

var (
	getSecretKeeper secretFunc = func(path string) (*secrets.Keeper, error) {
		ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancelFn()

		return openSecretKeeper(ctx, path, os.Getenv("CLOUD_PROVIDER"))
	}
)

// openSecretKeeper ...
func openSecretKeeper(ctx context.Context, path, cloudProvider string) (*secrets.Keeper, error) {
	switch strings.ToLower(cloudProvider) {
	case "", "local":
		return openLocal()
	case "gcp":
		return openGCPKMS()
	case "vault":
		return openVault(path)
	}
	return nil, fmt.Errorf("unknown secrets cloudProvider=%s", cloudProvider)
}

// openLocal ...
//
// 'base64key://'
// The URL hostname must be a base64-encoded key, of length 32 bytes when decoded.
func openLocal() (*secrets.Keeper, error) {
	var key [32]byte
	if v := os.Getenv("SECRETS_LOCAL_BASE64_KEY"); v != "" {
		k, err := localsecrets.Base64Key(v)
		if err != nil {
			return nil, fmt.Errorf("problem reading SECRETS_LOCAL_BASE64_KEY: %v", err)
		}
		key = k
	} else {
		k, err := localsecrets.NewRandomKey()
		if err != nil {
			return nil, err
		}
		key = k
	}
	return localsecrets.NewKeeper(key), nil
}

// openGCPKMS ...
//
// The environmental variable SECRETS_GCP_KEY_RESOURCE_ID is required and has the following form:
//  'projects/MYPROJECT/locations/MYLOCATION/keyRings/MYKEYRING/cryptoKeys/MYKEY'
//
// See https://cloud.google.com/kms/docs/object-hierarchy#key for more information
//
// gcpkms://projects/[PROJECT_ID]/locations/[LOCATION]/keyRings/[KEY_RING]/cryptoKeys/[KEY]
func openGCPKMS() (*secrets.Keeper, error) {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	client, done, err := gcpkms.Dial(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer done()

	return gcpkms.OpenKeeper(client, os.Getenv("SECRETS_GCP_KEY_RESOURCE_ID"), nil), nil
}

// openVault ...
//
// vault://mykey
func openVault(path string) (*secrets.Keeper, error) {
	defaultVaultConfig := vaultapi.DefaultConfig()
	cfg := &vault.Config{
		Token:     os.Getenv("VAULT_TOKEN"),
		APIConfig: *defaultVaultConfig,
	}
	if v := os.Getenv("VAULT_SERVER_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("VAULT_SERVER_URL"); v != "" {
		cfg.APIConfig.Address = v
	}

	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	api, err := vault.Dial(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return vault.OpenKeeper(api, path, nil), nil
}
