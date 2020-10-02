package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
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
		return fileBucket(ctx, log.NewNopLogger(), dir, signer)
	}
)

// NewTestBucket sets up and returns a new BucketFunc
func NewTestBucket(t *testing.T) BucketFunc {
	signer, err := FileblobSigner("http://localhost:8087", "secret")
	if err != nil {
		panic(err)
	}
	tempDir, _ := ioutil.TempDir("", "testBucket")
	if err := os.Mkdir(tempDir, 0777); strings.Contains(tempDir, "..") || (err != nil && !os.IsExist(err)) {
		os.RemoveAll(tempDir)
		panic(err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return func() (*blob.Bucket, error) {
		return fileblob.OpenBucket(tempDir, &fileblob.Options{
			URLSigner: signer,
		})
	}
}
