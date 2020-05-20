// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/admin"
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/internal/testclient"

	"github.com/go-kit/kit/log"
)

func TestAdmin__updateAccountStatus(t *testing.T) {
	customerID := base.ID()
	accountID := base.ID()

	repo := &mockRepository{
		Accounts: []*client.Account{
			{
				AccountID: accountID,
			},
		},
	}

	svc, c := testclient.Admin(t)
	RegisterAdminRoutes(log.NewNopLogger(), svc, repo)

	req := admin.UpdateAccountStatus{
		Status: admin.VALIDATED,
	}
	resp, err := c.CustomersApi.UpdateAccountStatus(context.TODO(), customerID, accountID, req)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK || err != nil {
		t.Errorf("bogus HTTP status: %d", resp.StatusCode)
		t.Fatal(err)
	}

	// retry, but expect an error
	repo.Err = errors.New("bad error")

	resp, _ = c.CustomersApi.UpdateAccountStatus(context.TODO(), customerID, accountID, req)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", resp.StatusCode)
	}
}
