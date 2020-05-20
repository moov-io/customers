// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"github.com/go-kit/kit/log"
	"github.com/moov-io/base/admin"
)

func RegisterAdminRoutes(logger log.Logger, svc *admin.Server, repo Repository) {
	svc.AddHandler("/customers/{customerID}/accounts/{accountID}/status", updateAccountStatus(logger, repo))
}
