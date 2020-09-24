// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"testing"

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/watchman"
	watchmanClient "github.com/moov-io/watchman/client"
)

func TestStoreAccountOFACSearchSuccess(t *testing.T) {
	mockRepository := mockRepository{}
	watchmanClient := watchman.NewTestWatchmanClient(&watchmanClient.OfacSdn{
		EntityID: "123",
		SdnName:  "dummy_sdn",
		SdnType:  "dummy_sdn_type",
		Programs: nil,
		Title:    "dummy_title",
		Remarks:  "dummy_remarks",
		Match:    10.2,
	}, nil)

	accountOfacSearcher := createTestOFACSearcher(&mockRepository, watchmanClient)
	requestID := "someRequestId"
	account := &client.Account{
		AccountID:  "123",
		HolderName: "John Doe",
	}
	err := accountOfacSearcher.StoreAccountOFACSearch(account, requestID)
	if err != nil {
		t.Errorf("got err: %v", err)
	}
}

func TestStoreAccountOFACSearchErrorNilAccountHolderName(t *testing.T) {
	mockRepository := mockRepository{}
	accountOfacSearcher := createTestOFACSearcher(&mockRepository, nil)
	requestID := "someRequestId"
	account := &client.Account{
		AccountID:  "123",
		HolderName: "",
	}
	err := accountOfacSearcher.StoreAccountOFACSearch(account, requestID)
	if err == nil {
		t.Errorf("got resp: %v", err)
	}
}

func TestStoreAccountOFACSearchErrorNilAccount(t *testing.T) {
	mockRepository := mockRepository{}
	accountOfacSearcher := createTestOFACSearcher(&mockRepository, nil)
	requestID := "someRequestId"
	err := accountOfacSearcher.StoreAccountOFACSearch(nil, requestID)
	if err == nil {
		t.Errorf("got resp: %v", err)
	}
}
