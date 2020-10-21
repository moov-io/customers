package internal

import (
	"fmt"
	"testing"

	"github.com/moov-io/base/log"
	"github.com/moov-io/customers/internal/database"
	"github.com/moov-io/customers/pkg/accounts"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/stretchr/testify/require"
)

func TestRehashAccountNumber(t *testing.T) {
	db := database.CreateTestMySQLDB(t)
	defer db.Close()

	logger := log.NewNopLogger()
	keeper := secrets.TestStringKeeper(t)

	// create test records
	var createdAccountIDs []string
	repo := accounts.NewRepo(logger, db.DB)

	for i := 0; i < 5; i++ {
		req := &accounts.CreateAccountRequest{
			HolderName:    "John Doe",
			AccountNumber: fmt.Sprintf("12345678%d", i),
			RoutingNumber: "987654320",
			Type:          client.ACCOUNTTYPE_CHECKING,
		}
		req.Disfigure(keeper, "app salt")

		err := req.Validate()
		require.NoError(t, err)

		acc, err := repo.CreateCustomerAccount("1", "1", req)
		require.NoError(t, err)
		createdAccountIDs = append(createdAccountIDs, acc.AccountID)
	}

	// rehash account numbers
	err := RehashStoredAccountNumber(logger, db.DB, "app salt", keeper)
	require.NoError(t, err)

	// test account numbers were re-hashed
	for _, accountID := range createdAccountIDs {
		row := db.DB.QueryRow(`select sha256_account_number from accounts where account_id = ?;`, accountID)
		require.NoError(t, row.Err())

		var acc account
		err := row.Scan(&acc.sha256AccountNumber)
		require.NoError(t, err)
		require.NotEmpty(t, acc.sha256AccountNumber)
	}
}
