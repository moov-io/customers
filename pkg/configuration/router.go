// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"encoding/json"
	"net/http"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/cmd/server/route"
	"github.com/moov-io/customers/pkg/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func RegisterRoutes(logger log.Logger, r *mux.Router, repo Repository) {
	r.Methods("GET").Path("/configuration/customers").HandlerFunc(getNamespaceConfig(logger, repo))
	r.Methods("PUT").Path("/configuration/customers").HandlerFunc(updateNamespaceConfig(logger, repo))
}

func getNamespaceConfig(logger log.Logger, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		namespace := route.GetNamespace(w, r)
		if namespace == "" {
			return
		}
		cfg, err := repo.Get(namespace)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cfg)
	}
}

func updateNamespaceConfig(logger log.Logger, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		namespace := route.GetNamespace(w, r)
		if namespace == "" {
			return
		}

		var body client.NamespaceConfiguration
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		cfg, err := repo.Update(namespace, &body)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cfg)
	}
}
