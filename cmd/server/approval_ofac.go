// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"time"

	client "github.com/moov-io/customers/client"
	ofac "github.com/moov-io/ofac/client"
)

type ofacSearchResult struct {
	entityId string
	sdnName  string
	sdnType  string
	match    float32
}

type ofacSearcher struct {
	repo       customerRepository
	ofacClient OFACClient
}

// storeCustomerOFACSearch performs OFAC searches against the Customer's name and nickname if populated.
// The higher matching search result is stored in s.customerRepository for use later (in approvals)
func (s *ofacSearcher) storeCustomerOFACSearch(cust *client.Customer, requestId string) error {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	sdn, err := s.ofacClient.Search(ctx, formatCustomerName(cust), requestId)
	if err != nil {
		return fmt.Errorf("ofacSearcher.storeCustomerOFACSearch: name search for customer=%s: %v", cust.ID, err)
	}
	var nickSDN *ofac.Sdn
	if cust.NickName != "" {
		nickSDN, err = s.ofacClient.Search(ctx, cust.NickName, requestId)
		if err != nil {
			return fmt.Errorf("ofacSearcher.storeCustomerOFACSearch: nickname search for customer=%s: %v", cust.ID, err)
		}
	}
	// Save the higher matching SDN (from name search or nick name)
	switch {
	case nickSDN != nil && nickSDN.Match > sdn.Match:
		err = s.repo.saveCustomerOFACSearch(cust.ID, ofacSearchResult{
			entityId: nickSDN.EntityID,
			sdnName:  nickSDN.SdnName,
			sdnType:  nickSDN.SdnType,
			match:    nickSDN.Match,
		})
	case sdn != nil:
		err = s.repo.saveCustomerOFACSearch(cust.ID, ofacSearchResult{
			entityId: sdn.EntityID,
			sdnName:  sdn.SdnName,
			sdnType:  sdn.SdnType,
			match:    sdn.Match,
		})
	}
	if err != nil {
		return fmt.Errorf("ofacSearcher.storeCustomerOFACSearch: saveCustomerOFACSearch customer=%s: %v", cust.ID, err)
	}
	return nil
}
