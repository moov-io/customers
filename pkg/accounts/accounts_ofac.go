package accounts

import (
	"context"
	"fmt"
	"time"

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/watchman"
	"github.com/pkg/errors"
)

type AccountOfacSearcher struct {
	Repo           Repository
	WatchmanClient watchman.Client
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
	if sdn == nil {
		return nil
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
