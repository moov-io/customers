// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/documents/storage"
	"github.com/moov-io/customers/pkg/route"
	"gocloud.dev/blob"

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
	r.Methods("GET").Path("/configuration/logo").HandlerFunc(getOrganizationLogo(logger, repo))
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

func getOrganizationLogo(logger log.Logger, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
			msg := "expected multipart upload with key of 'file'"
			logger.Log("configuration", msg, "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf("%s error=%v", msg, err))
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
			moovhttp.Problem(w, fmt.Errorf("image MIME type must be one of %s", strings.Join(supportedContentTypes(), ",")))
			return
		}

		orgCfg, err := repo.Get(organization)
		if err != nil {
			logger.Log("configuration", "error retrieving organization configuration", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		replaceExisting := (orgCfg.LogoFile != "") // keep track so we can return appropriate HTTP status code
		orgCfg.LogoFile = fmt.Sprintf("organization-%s-logo%s", organization, ext)

		bucket, err := bucketFactory()
		if err != nil {
			logger.Log("configuration", "problem uploading logo image", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		defer bucket.Close()

		writer, err := bucket.NewWriter(r.Context(), orgCfg.LogoFile, &blob.WriterOptions{
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

		updatedCfg, err := repo.Update(organization, orgCfg)
		if err != nil {
			logger.Log("configuration", "problem updating configuration", "error", err, "organization", organization, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		status := http.StatusCreated // respond with 201 if the resource didn't previously exist and was successfully created
		if replaceExisting {
			status = http.StatusOK // status ok if the the resource already existed and was successfully modified
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(updatedCfg)
	}
}
