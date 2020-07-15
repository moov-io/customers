// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/base64"
	"fmt"

	"gocloud.dev/secrets"
	"gocloud.dev/secrets/localsecrets"
)

func main() {
	uri := ensureValidKeyURI()

	fmt.Println("Generated random key")
	fmt.Println(uri)
}

func ensureValidKeyURI() string {
	for i := 0; i < 10; i++ {
		if uri := createKeyURI(); uri != "" {
			fmt.Println(i)
			return uri
		}
	}
	return ""
}

func createKeyURI() string {
	// For some reason this creates invalid keys sometimes..
	key, err := localsecrets.NewRandomKey()
	if err != nil {
		panic(fmt.Sprintf("ERROR creating random key: %v", err))
	}

	uri := fmt.Sprintf("%s://%s", localsecrets.Scheme, base64.StdEncoding.EncodeToString(key[:]))

	// Initialize a Keeper to verify it's a valid key
	keeper, err := secrets.OpenKeeper(context.TODO(), uri)
	if err != nil {
		return ""
	}
	defer keeper.Close()

	return uri
}
