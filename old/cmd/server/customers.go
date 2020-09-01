// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func addCustomerRoutes(logger log.Logger, r *mux.Router, repo customerRepository, customerSSNStorage *ssnStorage, ofac *ofacSearcher) {
	r.Methods("GET").Path("/customers").HandlerFunc(searchCustomers(logger, repo))
	r.Methods("GET").Path("/customers/{customerID}").HandlerFunc(getCustomer(logger, repo))
	r.Methods("POST").Path("/customers").HandlerFunc(createCustomer(logger, repo, customerSSNStorage, ofac))
	r.Methods("PUT").Path("/customers/{customerID}/metadata").HandlerFunc(replaceCustomerMetadata(logger, repo))
	r.Methods("POST").Path("/customers/{customerID}/address").HandlerFunc(addCustomerAddress(logger, repo))
}
