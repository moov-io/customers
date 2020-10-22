package internal

import (
	"fmt"
	"testing"

	"github.com/moov-io/base/database"
	"github.com/moov-io/base/log"
	"github.com/moov-io/customers/pkg/accounts"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/stretchr/testify/require"
)

func TestRehashAccountNumber(t *testing.T) {
	db := database.CreateTestSqliteDB(t)
	defer db.Close()

	logger := log.NewNopLogger()
	keeper := secrets.TestStringKeeper(t)

	// create test records
	var createdAccountIDs []string
	repo := accounts.NewRepo(logger, db.DB)

	const accountsCount = 5

	for i := 0; i < accountsCount; i++ {
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

	// clear sha256_account_number column as at time we run re-hash
	// migration it is NULL
	query := `update accounts set sha256_account_number = NULL;`
	_, err := db.DB.Exec(query)
	require.NoError(t, err)

	// rehash account numbers
	updatedRecordsCount, err := RehashStoredAccountNumber(logger, db.DB, "app salt", keeper)
	require.NoError(t, err)
	require.Equal(t, accountsCount, updatedRecordsCount)

	// test account numbers were re-hashed
	for _, accountID := range createdAccountIDs {
		row := db.DB.QueryRow(`select hashed_account_number, sha256_account_number from accounts where account_id = ?;`, accountID)
		require.NoError(t, row.Err())

		var sha256AccountNumber, hashedAccountNumber string
		err := row.Scan(&hashedAccountNumber, &sha256AccountNumber)
		require.NoError(t, err)
		require.NotEmpty(t, sha256AccountNumber)
		require.Empty(t, hashedAccountNumber)
	}
}
