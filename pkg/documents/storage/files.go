// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/pkg/route"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"gocloud.dev/blob/fileblob"
)

func AddFileblobRoutes(logger log.Logger, r *mux.Router, signer *fileblob.URLSignerHMAC, bucketFactory BucketFunc) {
	if v := os.Getenv("FILEBLOB_BASE_URL"); v != "" {
		u, err := url.Parse(v)
		if u != nil && err == nil {
			r.Methods("GET").Path(u.Path).HandlerFunc(proxyLocalFile(logger, signer, bucketFactory))
		}
	} else {
		r.Methods("GET").Path("/files").HandlerFunc(proxyLocalFile(logger, signer, bucketFactory))
	}
}

func proxyLocalFile(logger log.Logger, signer *fileblob.URLSignerHMAC, bucketFactory BucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		ctx, cancelFn := context.WithTimeout(context.TODO(), 30*time.Second)
		defer cancelFn()

		key, err := signer.KeyFromURL(ctx, r.URL)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		bucket, err := bucketFactory()
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		// Grab the blob.Reader for proxying to endpoint
		rdr, err := bucket.NewReader(ctx, key, nil)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		defer rdr.Close()

		logger.Log("files", fmt.Sprintf("proxying document=%s contentType=%s", key, rdr.ContentType()), "requestID", moovhttp.GetRequestID(r))

		w.Header().Set("Content-Disposition", "inline")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", rdr.Size()))
		w.Header().Set("Content-Type", rdr.ContentType())
		w.WriteHeader(http.StatusOK)

		if n, err := io.Copy(w, rdr); err != nil || n == 0 {
			moovhttp.Problem(w, fmt.Errorf("proxyLocalFile: n=%d error=%v", n, err))
			return
		}
	}
}
