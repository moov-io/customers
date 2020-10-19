package internal

import (
	"database/sql"

	"github.com/moov-io/base/log"

	"github.com/moov-io/customers/pkg/secrets"
	"github.com/moov-io/customers/pkg/secrets/hash"
)

type account struct {
	id                     string
	encryptedAccountNumber string
	hashedAccountNumber    string
	sha256AccountNumber    string
}

// RehashStoredAccountNumber generates SHA256 hash with salt for rows that need it
func RehashStoredAccountNumber(logger log.Logger, db *sql.DB, appSalt string, keeper *secrets.StringKeeper) error {

	err := findAccountsInBatches(logger, db, func(acc account) error {
		accountNumber, err := keeper.DecryptString(acc.encryptedAccountNumber)
		if err != nil {
			return err
		}

		sha256Hash, err := hash.SHA256Hash(appSalt, accountNumber)
		if err != nil {
			return err
		}

		if err := updateAccountSHA256Hash(acc.id, sha256Hash, db); err != nil {
			return err
		}
		return nil
	})

	return err
}

// findAccountsInBatches will select all accounts with empty
// sha256_account_number in 100 records batches
// NOTE: if we don't set sha256_account_number of account in updateFunc it will
// loop forever
func findAccountsInBatches(logger log.Logger, db *sql.DB, updateFunc func(acc account) error) error {
	// query 100 rows that should be rehashed
	query := `
		select
			account_id, encrypted_account_number
		from
			accounts
		where
			sha256_account_number = ''
		limit
			100;
	`

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for {
		rows, err := stmt.Query()
		if err != nil {
			return err
		}

		var accounts []account

		for rows.Next() {
			var acc account
			if err := rows.Scan(&acc.id, &acc.encryptedAccountNumber); err != nil {
				rows.Close()
				return err
			}

			accounts = append(accounts, acc)
		}
		rows.Close()

		if len(accounts) == 0 {
			return nil
		}

		for _, acc := range accounts {
			if err := updateFunc(acc); err != nil {
				logger.LogErrorF("Failed to update account (%s): %v", acc.id, err)
			}
		}
	}
}

func updateAccountSHA256Hash(accountID string, hash string, db *sql.DB) error {
	query := `
		update
			accounts
		set
			hashed_account_number = ?,
			sha256_account_number = ?
		where
			account_id = ?;
	`
	_, err := db.Exec(query, hash, hash, accountID)
	return err
}
