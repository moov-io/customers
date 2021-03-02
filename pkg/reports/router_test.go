package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/base"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"

	"github.com/moov-io/base/database"
	"github.com/moov-io/customers/pkg/accounts"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/customers"
)

func TestGetAccounts(t *testing.T) {
	router := mux.NewRouter()
	logger := log.NewNopLogger()
	db := database.CreateTestSQLiteDB(t).DB
	customerRepo := customers.NewCustomerRepo(logger, db)
	accountRepo := accounts.NewRepo(logger, db)
	organization := "organization"
	AddRoutes(log.NewNopLogger(), router, customerRepo, accountRepo)

	// Check request without query parameters
	req := httptest.NewRequest("GET", "/reports/accounts", nil)
	req.Header.Set("x-organization", organization)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)

	var got []*client.ReportAccountResponse
	err := json.Unmarshal(res.Body.Bytes(), &got)
	require.NoError(t, err)
	require.Empty(t, got)

	// create sample customers and their accounts
	var accountIDs []string
	for i := 0; i < 3; i++ {
		cust := &client.Customer{
			CustomerID: base.ID(),
			FirstName:  fmt.Sprintf("%d", i),
		}
		err := customerRepo.CreateCustomer(cust, organization)
		require.NoError(t, err)

		account, err := accountRepo.CreateCustomerAccount(
			cust.CustomerID,
			"test-user-id",
			&accounts.CreateAccountRequest{},
		)
		require.NoError(t, err)

		accountIDs = append(accountIDs, account.AccountID)
	}

	q := req.URL.Query()
	q.Add("accountIDs", strings.Join(accountIDs, ","))
	req.URL.RawQuery = q.Encode()
	req = httptest.NewRequest("GET", req.URL.String(), nil)
	req.Header.Set("x-organization", organization)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)

	err = json.Unmarshal(res.Body.Bytes(), &got)
	require.NoError(t, err)

	require.Len(t, got, len(accountIDs))
	var gotAccountIDs []string
	for _, e := range got {
		require.NotNil(t, e.Account)
		require.NotNil(t, e.Customer)
		gotAccountIDs = append(gotAccountIDs, e.Account.AccountID)
	}

	require.ElementsMatch(t, accountIDs, gotAccountIDs)
}
