package validator

import (
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/internal/database"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	repo := createTestRepository(t)

	accountID := base.ID()

	// no validations yet
	_, err := repo.GetValidation(accountID, "hello")
	require.Error(t, err)

	validation := &Validation{
		AccountID: accountID,
		Status:    StatusInit,
		Strategy:  "instant",
		Vendor:    "mx",
	}

	err = repo.CreateValidation(validation)
	require.NoError(t, err)
	require.NotEmpty(t, validation.ValidationID)

	validation, err = repo.GetValidation(accountID, validation.ValidationID)
	require.NoError(t, err)
	require.Equal(t, "instant", validation.Strategy)
	require.Equal(t, "mx", validation.Vendor)
	require.NotEmpty(t, validation.CreatedAt)
	require.Equal(t, validation.CreatedAt, validation.UpdatedAt)

	validation.Strategy = "micro-deposits"
	validation.Vendor = "moov"
	validation.Status = StatusCompleted

	err = repo.UpdateValidation(validation)
	require.NoError(t, err)
	require.NotEqual(t, validation.CreatedAt, validation.UpdatedAt)

}

func createTestRepository(t *testing.T) Repository {
	t.Helper()

	db := database.CreateTestMySQLDB(t)

	return &sqlRepository{db.DB}
}
