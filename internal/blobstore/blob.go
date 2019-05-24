// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package blobstore

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
)

// OpenBucket returns a Go Cloud Development Kit (Go CDK) Bucket object which can be used to read and write arbitrary
// blobs from a cloud provider blob store. Checkout https://gocloud.dev/ref/blob/ for more details
func OpenBucket(ctx context.Context, cloudProvider, bucketName string) (*blob.Bucket, error) {
	switch strings.ToLower(cloudProvider) {
	case "aws":
		return awsBucket(ctx, bucketName)
	case "file":
		os.Mkdir(bucketName, 0777) // TODO(adam): reject paths with '..'
		return fileblob.OpenBucket(bucketName, nil)
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
