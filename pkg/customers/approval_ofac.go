// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"context"
	"errors"
	"fmt"
	"time"

	client "github.com/moov-io/customers/pkg/client"
	watchman "github.com/moov-io/watchman/client"
)

type ofacSearchResult struct {
	EntityID  string    `json:"entityID"`
	SDNName   string    `json:"sdnName"`
	SDNType   string    `json:"sdnType"`
	Match     float32   `json:"match"`
	CreatedAt time.Time `json:"createdAt"`
}

type ofacSearcher struct {
	repo           customerRepository
	watchmanClient WatchmanClient
}

// storeCustomerOFACSearch performs OFAC searches against the Customer's name and nickname if populated.
// The higher matching search result is stored in s.customerRepository for use later (in approvals)
func (s *ofacSearcher) storeCustomerOFACSearch(cust *client.Customer) error {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	if cust == nil {
		return errors.New("nil Customer")
	}

	sdn, err := s.watchmanClient.Search(ctx, formatCustomerName(cust), "")
	if err != nil {
		return fmt.Errorf("ofacSearcher.storeCustomerOFACSearch: name search for customer=%s: %v", cust.CustomerID, err)
	}
	var nickSDN *watchman.OfacSdn
	if cust.NickName != "" {
		nickSDN, err = s.watchmanClient.Search(ctx, cust.NickName, "")
		if err != nil {
			return fmt.Errorf("ofacSearcher.storeCustomerOFACSearch: nickname search for customer=%s: %v", cust.CustomerID, err)
		}
	}
	// Save the higher matching SDN (from name search or nick name)
	switch {
	case nickSDN != nil && nickSDN.Match > sdn.Match:
		err = s.repo.saveCustomerOFACSearch(cust.CustomerID, ofacSearchResult{
			EntityID:  nickSDN.EntityID,
			SDNName:   nickSDN.SdnName,
			SDNType:   nickSDN.SdnType,
			Match:     nickSDN.Match,
			CreatedAt: time.Now(),
		})
	case sdn != nil:
		err = s.repo.saveCustomerOFACSearch(cust.CustomerID, ofacSearchResult{
			EntityID:  sdn.EntityID,
			SDNName:   sdn.SdnName,
			SDNType:   sdn.SdnType,
			Match:     sdn.Match,
			CreatedAt: time.Now(),
		})
	}
	if err != nil {
		return fmt.Errorf("ofacSearcher.storeCustomerOFACSearch: saveCustomerOFACSearch customer=%s: %v", cust.CustomerID, err)
	}
	return nil
}
