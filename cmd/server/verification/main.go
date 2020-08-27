package verification

import (
	"net/http"

	"github.com/moov-io/customers/cmd/server/accounts"
	"github.com/moov-io/customers/pkg/secrets"
)

type AccountVerifier interface {
	InitiateAccountVerification() http.HandlerFunc
	CompleteAccountVerification(repo accounts.Repository, keeper *secrets.StringKeeper) http.HandlerFunc
}
