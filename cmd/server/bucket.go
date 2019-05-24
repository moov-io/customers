// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/moov-io/customers/internal/blobstore"

	"gocloud.dev/blob"
)

type bucketFunc func() (*blob.Bucket, error)

func getBucket() (*blob.Bucket, error) {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	cloudProvider, bucketName := os.Getenv("CLOUD_PROVIDER"), os.Getenv("BUCKET_NAME")
	if cloudProvider == "" || bucketName == "" {
		return nil, fmt.Errorf("storage: missing CLOUD_PROVIDER=%s and/or BUCKET_NAME=%s", cloudProvider, bucketName)
	}
	return blobstore.OpenBucket(ctx, cloudProvider, bucketName)
}
