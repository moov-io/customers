// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
)

type bucketFunc func() (*blob.Bucket, error)

func getBucket(bucketName, cloudProvider string, fileblobSigner *fileblob.URLSignerHMAC) bucketFunc {
	return func() (*blob.Bucket, error) {
		ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancelFn()

		if bucketName == "" || cloudProvider == "" {
			return nil, fmt.Errorf("storage: missing BUCKET_NAME=%s and/or CLOUD_PROVIDER=%s", bucketName, cloudProvider)
		}
		return openBucket(ctx, bucketName, cloudProvider, fileblobSigner)
	}
}

// openBucket returns a Go Cloud Development Kit (Go CDK) Bucket object which can be used to read and write arbitrary
// blobs from a cloud provider blob store. Checkout https://gocloud.dev/ref/blob/ for more details
func openBucket(ctx context.Context, bucketName, cloudProvider string, fileblobSigner *fileblob.URLSignerHMAC) (*blob.Bucket, error) {
	switch strings.ToLower(cloudProvider) {
	case "aws":
		return awsBucket(ctx, bucketName)
	case "file":
		return fileBucket(ctx, bucketName, fileblobSigner)
	case "gcp":
		return gcpBucket(ctx, bucketName)
	default:
		return nil, fmt.Errorf("invalid cloud provider: %s", cloudProvider)
	}
}

func awsBucket(ctx context.Context, bucketName string) (*blob.Bucket, error) {
	c := &aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewEnvCredentials(), // reads AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	}
	s := session.Must(session.NewSession(c))
	return s3blob.OpenBucket(ctx, s, bucketName, nil)
}

func fileblobSigner(baseURL, secret string) (*fileblob.URLSignerHMAC, error) {
	if u, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL %s error=%s", baseURL, err)
	} else {
		return fileblob.NewURLSignerHMAC(u, []byte(secret)), nil
	}
}

func fileBucket(ctx context.Context, bucketName string, signer *fileblob.URLSignerHMAC) (*blob.Bucket, error) {
	if err := os.Mkdir(bucketName, 0777); strings.Contains(bucketName, "..") || (err != nil && !os.IsExist(err)) {
		return nil, fmt.Errorf("problem creating %s error=%v", bucketName, err)
	}
	return fileblob.OpenBucket(bucketName, &fileblob.Options{
		URLSigner: signer,
	})
}

func gcpBucket(ctx context.Context, bucketName string) (*blob.Bucket, error) {
	// DefaultCredentials assumes a user has logged in with gcloud
	// https://cloud.google.com/docs/authentication/getting-started
	creds, err := gcp.DefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	c, err := gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
	if err != nil {
		return nil, err
	}
	return gcsblob.OpenBucket(ctx, c, bucketName, nil)
}
