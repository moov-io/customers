// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"testing"

	"github.com/moov-io/base"
	client "github.com/moov-io/customers/client"
	ofac "github.com/moov-io/ofac/client"
)

func createTestOFACSearcher(repo customerRepository, ofacClient OFACClient) *ofacSearcher {
	if repo == nil {
		repo = &testCustomerRepository{}
	}
	if ofacClient == nil {
		ofacClient = &testOFACClient{}
	}
	return &ofacSearcher{repo: repo, ofacClient: ofacClient}
}

func TestOFACSearcher__storeCustomerOFACSearch(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	ofacClient := &testOFACClient{}
	searcher := createTestOFACSearcher(repo, ofacClient)

	ofacClient.sdn = &ofac.Sdn{
		EntityID: "1241421",
		SdnName:  "Jane Doe",
		Match:    0.99,
	}
	customerID := base.ID()
	if err := searcher.storeCustomerOFACSearch(&client.Customer{ID: customerID}, "requestID"); err != nil {
		t.Fatal(err)
	}
	res, err := repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res.entityId != "1241421" {
		t.Errorf("ofacSearchResult: %#v", res)
	}

	// retry but with NickName set (test coverage)
	customerID = base.ID()
	if err := searcher.storeCustomerOFACSearch(&client.Customer{ID: customerID, NickName: "John Doe"}, "requestID"); err != nil {
		t.Fatal(err)
	}
}
