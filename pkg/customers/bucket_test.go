// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

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

func TestBucket__openBucket(t *testing.T) {
	// test the invalid cases
	bucket, err := openBucket(context.TODO(), "", "", nil)
	if bucket != nil || err == nil {
		t.Errorf("expected error: bucket=%v error=%v", bucket, err)
	}
	bucket, err = openBucket(context.TODO(), "", "other", nil)
	if bucket != nil || err == nil {
		t.Errorf("expected error: bucket=%v error=%v", bucket, err)
	}
}

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

	// error case
	if _, err := getBucket("", "", nil)(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBucketAWS(t *testing.T) {
	bucket, err := openBucket(context.TODO(), "", "aws", nil)
	if err == nil || bucket != nil {
		t.Errorf("expected error bucket=%v", bucket)
	}
}

func TestBucketGCP(t *testing.T) {
	bucket, err := openBucket(context.TODO(), "", "gcp", nil)
	if err == nil || bucket != nil {
		t.Errorf("expected error bucket=%v", bucket)
	}
}
