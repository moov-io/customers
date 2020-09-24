// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"io/ioutil"
	"testing"
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

func TestBucket__GetBucket(t *testing.T) {
	dir, _ := ioutil.TempDir("", "customers-getBucket")

	signer, err := FileblobSigner("http://localhost:8087", "secret")
	if err != nil {
		t.Fatal(err)
	}

	bucket, err := GetBucket(dir, "file", signer)()
	if err != nil {
		t.Fatal(err)
	}
	if bucket == nil {
		t.Fatalf("nil blob.Bucket: %#v", bucket)
	}

	// error case
	if _, err := GetBucket("", "", nil)(); err == nil {
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
