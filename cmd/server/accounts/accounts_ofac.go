package accounts

import (
	"context"
	"fmt"
	"github.com/moov-io/customers/pkg/client"
	watchman "github.com/moov-io/watchman/client"
	"github.com/pkg/errors"
	"time"
)

type AccountOfacSearcher struct {
	Repo           Repository
	WatchmanClient WatchmanClient
}

type WatchmanClient interface {
	Ping() error

	Search(ctx context.Context, name string, requestID string) (*watchman.OfacSdn, error)
}

// StoreAccountOFACSearch performs OFAC searches against the Account's HolderName and nickname if populated.
// The search result is stored in s.Repo for use later (in approvals)
func (s *AccountOfacSearcher) StoreAccountOFACSearch(account *client.Account, requestID string) error {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	if account == nil {
		return errors.New("nil account")
	}

	if account.HolderName == "" {
		return errors.New("no account HolderName to perform check with")
	}

	sdn, err := s.WatchmanClient.Search(ctx, account.HolderName, requestID)
	if err != nil {
		return fmt.Errorf("AccountOfacSearcher.StoreAccountOFACSearch: name search for account=%s: %v", account.AccountID, err)
	}
	err = s.Repo.saveAccountOFACSearch(account.AccountID, &client.OfacSearch{
		EntityID:  sdn.EntityID,
		SdnName:   sdn.SdnName,
		SdnType:   sdn.SdnType,
		Match:     sdn.Match,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("AccountOfacSearcher.StoreAccountOFACSearch: saveAccountOFACSearch account=%s: %v", account.AccountID, err)
	}

	return nil
}
