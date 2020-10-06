// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/documents/storage"
	"github.com/moov-io/customers/pkg/route"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
)

const maxImageSize int64 = 20 * 1024 * 1024 // 20MB

var (
	allowedContentTypes = map[string]bool{
		"image/jpeg":    true,
		"image/svg+xml": true,
		"image/gif":     true,
		"image/png":     true,
	}

	errMissingFile     = errors.New("expected multipart upload with key of 'file'")
	errUnsupportedType = fmt.Errorf("file type must be one of %s", listAllowedContentTypes())
)

func listAllowedContentTypes() string {
	types := make([]string, len(allowedContentTypes))
	idx := 0
	for t := range allowedContentTypes {
		types[idx] = t
		idx++
	}
	return strings.Join(types, ",")
}

func RegisterRoutes(logger log.Logger, r *mux.Router, repo Repository, bucketFunc storage.BucketFunc) {
	logger = logger.WithKeyValue("package", "configuration")

	r.Methods("GET").Path("/configuration/customers").HandlerFunc(getOrganizationConfig(logger, repo))
	r.Methods("PUT").Path("/configuration/customers").HandlerFunc(updateOrganizationConfig(logger, repo))
	r.Methods("GET").Path("/configuration/logo").HandlerFunc(getOrganizationLogo(logger, repo, bucketFunc))
	r.Methods("PUT").Path("/configuration/logo").HandlerFunc(uploadOrganizationLogo(logger, repo, bucketFunc))
}

func getOrganizationConfig(logger log.Logger, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}
		cfg, err := repo.Get(organization)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cfg)
	}
}

func updateOrganizationConfig(logger log.Logger, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		var body client.OrganizationConfiguration
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		cfg, err := repo.Update(organization, &body)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cfg)
	}
}

func getOrganizationLogo(logger log.Logger, repo Repository, bucketFactory storage.BucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organization := route.GetOrganization(w, r)
		if organization == "" {
			logger.Log("get organization logo called with no organization header")
			return
		}

		logger = logger.WithKeyValue("organization", organization)

		bucket, err := bucketFactory()
		if err != nil {
			logger.LogError("problem retrieving logo image", err)
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		rdr, err := bucket.NewReader(r.Context(), makeDocumentKey(organization), nil)
		if err != nil {
			msg := "error retrieving logo file"
			if gcerrors.Code(err) == gcerrors.NotFound {
				msg = msg + " - file not found"
				logger.Log("configuration", msg, "error", err, "organization", organization, "requestID", requestID)
				http.NotFound(w, r)
				return
			}

			logger.LogError(msg, err)
			moovhttp.Problem(w, fmt.Errorf(msg))
			return
		}
		defer rdr.Close()

		fBytes, err := ioutil.ReadAll(rdr)
		if err != nil || fBytes == nil {
			logger.LogError("problem reading logo file", err)
			moovhttp.Problem(w, fmt.Errorf("problem reading logo file - error=%v", err))
			return
		}

		w.Header().Set("Content-Type", rdr.ContentType())
		w.WriteHeader(http.StatusOK)
		w.Write(fBytes)
	}
}

func uploadOrganizationLogo(logger log.Logger, repo Repository, bucketFactory storage.BucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organization := route.GetOrganization(w, r)
		if organization == "" {
			logger.Log("upload logo called with no organization header")
			return
		}

		file, _, err := r.FormFile("file")
		if file == nil || err != nil {
			logger.LogError(errMissingFile.Error(), err)
			moovhttp.Problem(w, errMissingFile)
			return
		}
		defer file.Close()

		// Detect the content type by reading the first 512 bytes of r.Body (read into file as we expect a multipart request)
		buf := make([]byte, 512)
		if _, err := file.Read(buf); err != nil && err != io.EOF {
			logger.LogError("problem reading file", err)
			moovhttp.Problem(w, err)
			return
		}

		contentType := http.DetectContentType(buf)
		if !allowedContentTypes[contentType] {
			logger.WithKeyValue("contentType", contentType).
				Log("unsupported content type for logo image file")
			moovhttp.Problem(w, errUnsupportedType)
			return
		}

		bucket, err := bucketFactory()
		if err != nil {
			logger.LogError("problem uploading logo image", err)
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		ctx, cancelFn := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancelFn()

		writer, err := bucket.NewWriter(ctx, makeDocumentKey(organization), &blob.WriterOptions{
			ContentDisposition: "inline",
			ContentType:        contentType,
		})
		if err != nil {
			logger.LogError("problem uploading logo image", err)
			moovhttp.Problem(w, err)
			return
		}
		defer writer.Close()

		written, err := io.Copy(writer, io.LimitReader(io.MultiReader(bytes.NewReader(buf), file), maxImageSize))
		if err != nil || written == 0 {
			logger.LogError("problem uploading logo image", err)
			moovhttp.Problem(w, fmt.Errorf("problem writing file - wrote %d bytes with error=%v", written, err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func makeDocumentKey(organization string) string {
	return path.Join("organizations", organization, "logo")
}
