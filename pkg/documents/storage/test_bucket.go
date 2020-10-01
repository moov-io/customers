package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
)

var (
	TestBucket BucketFunc = func() (*blob.Bucket, error) {
		signer, err := FileblobSigner("http://localhost:8087", "secret")
		if err != nil {
			panic(fmt.Sprintf("testBucket: %v", err))
		}

		ctx, cancelFn := context.WithTimeout(context.TODO(), 1*time.Second)
		defer cancelFn()

		dir, _ := ioutil.TempDir("", "testBucket")
		return fileBucket(ctx, dir, signer)
	}
)

// NewTestBucket returns a new BucketFunc, along with the underlying "temp" directory.
// It's the caller's responsibility to remove the temp dir when it's no longer needed.
func NewTestBucket() (string, func() (*blob.Bucket, error)) {
	signer, err := FileblobSigner("http://localhost:8087", "secret")
	if err != nil {
		panic(err)
	}
	tempDir, _ := ioutil.TempDir("", "testBucket")
	if err := os.Mkdir(tempDir, 0777); strings.Contains(tempDir, "..") || (err != nil && !os.IsExist(err)) {
		os.RemoveAll(tempDir)
		panic(err)
	}

	bucketFunc := func() (*blob.Bucket, error) {
		return fileblob.OpenBucket(tempDir, &fileblob.Options{
			URLSigner: signer,
		})
	}
	return tempDir, bucketFunc
}
