package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/go-kit/kit/log"
	"gocloud.dev/blob"
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
