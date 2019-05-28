// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"gocloud.dev/blob"
)

var (
	testBucket bucketFunc = func() (*blob.Bucket, error) {
		signer, err := fileblobSigner("http://localhost:8087", "secret")
		if err != nil {
			panic(fmt.Sprintf("testBucket: %v", err))
		}

		ctx, cancelFn := context.WithTimeout(context.TODO(), 1*time.Second)
		defer cancelFn()

		dir, _ := ioutil.TempDir("", "testBucket")
		return fileBucket(ctx, dir, signer)
	}
)

func TestBucket__getBucket(t *testing.T) {
	dir, _ := ioutil.TempDir("", "customers-getBucket")

	signer, err := fileblobSigner("http://localhost:8087", "secret")
	if err != nil {
		t.Fatal(err)
	}

	bucket, err := getBucket(dir, "file", signer)()
	if err != nil {
		t.Fatal(err)
	}
	if bucket == nil {
		t.Fatalf("nil blob.Bucket: %#v", bucket)
	}
}
