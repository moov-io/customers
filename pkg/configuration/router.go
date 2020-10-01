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

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	maxImageSize int64 = 20 * 1024 * 1024 // 20MB

	// serves two purposes:
	//   1. defines content/file types supported in the API
	//   2. maps a supported content MIME type to it's common file extension
	contentTypeFileExtensions = map[string]string{
		"image/jpeg":    ".jpg",
		"image/svg+xml": ".svg",
		"image/gif":     ".gif",
		"image/png":     ".png",
	}

	errMissingFile     = errors.New("expected multipart upload with key of 'file'")
	errUnsupportedType = fmt.Errorf("image MIME type must be one of %s", strings.Join(supportedContentTypes(), ","))
)

// returns a list of supported MIME types from contentTypeFileExtensions
func supportedContentTypes() []string {
	var types []string
	for key := range contentTypeFileExtensions {
		types = append(types, key)
	}
	return types
}

func RegisterRoutes(logger log.Logger, r *mux.Router, repo Repository, bucketFactory storage.BucketFunc) {
	r.Methods("GET").Path("/configuration/customers").HandlerFunc(getOrganizationConfig(logger, repo))
	r.Methods("PUT").Path("/configuration/customers").HandlerFunc(updateOrganizationConfig(logger, repo))
	r.Methods("GET").Path("/configuration/logo").HandlerFunc(getOrganizationLogo(logger, repo, bucketFactory))
	r.Methods("PUT").Path("/configuration/logo").HandlerFunc(uploadOrganizationLogo(logger, repo, bucketFactory))
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
		respondWithJSON(w, http.StatusOK, cfg)
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
		respondWithJSON(w, http.StatusOK, cfg)
	}
}

func getOrganizationLogo(logger log.Logger, repo Repository, bucketFactory storage.BucketFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := moovhttp.GetRequestID(r)
		organization := route.GetOrganization(w, r)
		if organization == "" {
			logger.Log("configuration", "get organization logo called with no organization header", "requestID", requestID)
			return
		}

		cfg, err := repo.Get(organization)
		if err != nil {
			logger.Log("configuration", "error retrieving organization configuration", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		if cfg.LogoFile == "" {
			logger.Log("configuration", "no logo uploaded for organization", "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf("no logo uploaded for organization %s", organization))
			return
		}

		bucket, err := bucketFactory()
		if err != nil {
			logger.Log("configuration", "problem retrieving logo image", "error", err, "organization", organization, "logoFile", cfg.LogoFile, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		rdr, err := bucket.NewReader(r.Context(), makeDocumentKey(organization, cfg.LogoFile), nil)
		if err != nil {
			msg := "error retrieving logo file"
			if gcerrors.Code(err) == gcerrors.NotFound {
				msg = msg + " - file not found"
			}
			logger.Log("configuration", msg, "error", err, "organization", organization, "logoFile", cfg.LogoFile, "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf(msg))
			return
		}
		defer rdr.Close()

		fBytes, err := ioutil.ReadAll(rdr)
		if err != nil || fBytes == nil {
			logger.Log("configuration", "problem reading logo file", "error", err, "organization", organization, "logoFile", cfg.LogoFile, "requestID", requestID)
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
		requestID := moovhttp.GetRequestID(r)
		organization := route.GetOrganization(w, r)
		if organization == "" {
			logger.Log("configuration", "upload organization logo called with no organization header", "requestID", requestID)
			return
		}

		file, _, err := r.FormFile("file")
		if file == nil || err != nil {
			logger.Log("configuration", errMissingFile, "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, errMissingFile)
			return
		}
		defer file.Close()

		// Detect the content type by reading the first 512 bytes of r.Body (read into file as we expect a multipart request)
		buf := make([]byte, 512)
		if _, err := file.Read(buf); err != nil && err != io.EOF {
			logger.Log("configuration", "problem reading file", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		contentType := http.DetectContentType(buf)

		ext, supported := contentTypeFileExtensions[contentType]
		if !supported {
			logger.Log("configuration", "unsupported content type for logo image file", "error", err, "contentType", contentType, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, errUnsupportedType)
			return
		}

		originalConfig, err := repo.Get(organization)
		if err != nil {
			logger.Log("configuration", "error retrieving organization configuration", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		replaceExisting := (originalConfig.LogoFile != "") // keep track so we can return appropriate HTTP status code
		originalConfig.LogoFile = fmt.Sprintf("organization-%s-logo%s", organization, ext)

		updatedCfg, err := repo.Update(organization, originalConfig)
		if err != nil {
			logger.Log("configuration", "problem updating configuration", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		bucket, err := bucketFactory()
		if err != nil {
			logger.Log("configuration", "problem uploading logo image", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		// Write our document from the request body
		ctx, cancelFn := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancelFn()

		writer, err := bucket.NewWriter(ctx, makeDocumentKey(organization, updatedCfg.LogoFile), &blob.WriterOptions{
			ContentDisposition: "inline",
			ContentType:        contentType,
		})
		if err != nil {
			logger.Log("configuration", "problem uploading logo image", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		defer writer.Close()

		written, err := io.Copy(writer, io.LimitReader(io.MultiReader(bytes.NewReader(buf), file), maxImageSize))
		if err != nil || written == 0 {
			logger.Log("configuration", "problem uploading logo image", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf("problem writing file - wrote %d bytes with error=%v", written, err))
			return
		}

		status := http.StatusCreated // respond with 201 if the resource didn't previously exist and was successfully created
		if replaceExisting {
			status = http.StatusOK // status ok if the the resource already existed and was successfully modified
		}
		respondWithJSON(w, status, updatedCfg)
	}
}

func respondWithJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func makeDocumentKey(organization string, docID string) string {
	return path.Join("organiations", organization, docID)
}
