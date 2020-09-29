// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-kit/kit/log"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
)

type BucketFunc func() (*blob.Bucket, error)

func GetBucket(logger log.Logger, bucketName, cloudProvider string, FileblobSigner *fileblob.URLSignerHMAC) BucketFunc {
	return func() (*blob.Bucket, error) {
		ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancelFn()

		if bucketName == "" {
			return nil, errors.New("missing document bucket")
		}
		if cloudProvider == "" {
			return nil, errors.New("missing documents cloud provider")
		}

		return openBucket(ctx, logger, bucketName, cloudProvider, FileblobSigner)
	}
}

// openBucket returns a Go Cloud Development Kit (Go CDK) Bucket object which can be used to read and write arbitrary
// blobs from a cloud provider blob store. Checkout https://gocloud.dev/ref/blob/ for more details
func openBucket(ctx context.Context, logger log.Logger, bucketName, cloudProvider string, FileblobSigner *fileblob.URLSignerHMAC) (*blob.Bucket, error) {
	switch strings.ToLower(cloudProvider) {
	case "aws":
		return awsBucket(ctx, logger, bucketName)
	case "file":
		return fileBucket(ctx, logger, bucketName, FileblobSigner)
	case "gcp":
		return gcpBucket(ctx, logger, bucketName)
	default:
		return nil, fmt.Errorf("invalid cloud provider: %s", cloudProvider)
	}
}

func awsBucket(ctx context.Context, logger log.Logger, bucketName string) (*blob.Bucket, error) {
	c := &aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewEnvCredentials(), // reads AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	}
	s := session.Must(session.NewSession(c))

	bucket, err := s3blob.OpenBucket(ctx, s, bucketName, nil)
	if err != nil {
		logger.Log("storage", fmt.Sprintf("ERROR creating %s gcp bucket: %v", bucketName, err))
	} else {
		logger.Log("storage", fmt.Sprintf("created %s aws bucket: %T", bucketName, bucket))
	}
	return bucket, err
}

func FileblobSigner(baseURL, secret string) (*fileblob.URLSignerHMAC, error) {
	if u, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL %s error=%s", baseURL, err)
	} else {
		return fileblob.NewURLSignerHMAC(u, []byte(secret)), nil
	}
}

func fileBucket(ctx context.Context, logger log.Logger, bucketName string, signer *fileblob.URLSignerHMAC) (*blob.Bucket, error) {
	if err := os.Mkdir(bucketName, 0777); strings.Contains(bucketName, "..") || (err != nil && !os.IsExist(err)) {
		return nil, fmt.Errorf("problem creating %s error=%v", bucketName, err)
	}
	logger.Log("storage", fmt.Sprintf("created %s for file bucket", bucketName))
	return fileblob.OpenBucket(bucketName, &fileblob.Options{
		URLSigner: signer,
	})
}

func gcpBucket(ctx context.Context, logger log.Logger, bucketName string) (*blob.Bucket, error) {
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

	bucket, err := gcsblob.OpenBucket(ctx, c, bucketName, nil)
	if err != nil {
		logger.Log("storage", fmt.Sprintf("ERROR creating %s gcp bucket: %v", bucketName, err))
	} else {
		logger.Log("storage", fmt.Sprintf("created %s gcp bucket: %v", bucketName, err))
	}
	return bucket, err
}
