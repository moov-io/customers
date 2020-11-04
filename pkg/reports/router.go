package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/base/log"

	"github.com/moov-io/customers/pkg/accounts"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/customers"
	"github.com/moov-io/customers/pkg/route"
)

func AddRoutes(
	logger log.Logger,
	r *mux.Router,
	customerRepo customers.CustomerRepository,
	accountRepo accounts.Repository,
) {
	logger = logger.Set("package", "reports")

	r.Methods("GET").Path("/reports/accounts").HandlerFunc(getCustomerAccounts(logger, customerRepo, accountRepo))
}

type CustomerAccount struct {
	Customer *client.Customer `json:"customer"`
	Account  *client.Account  `json:"account"`
}

func getCustomerAccounts(
	logger log.Logger,
	customerRepo customers.CustomerRepository,
	accountRepo accounts.Repository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		accountIDsInput := r.URL.Query().Get("accountIDs")
		accountIDsInput = strings.TrimSpace(accountIDsInput)
		accountIDs := strings.Split(accountIDsInput, ",")
		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		limit := 25
		if len(accountIDs) > limit {
			moovhttp.Problem(w, fmt.Errorf("exceeded limit of %d accountIDs, found %d", limit, len(accountIDs)))
			return
		}

		allAccounts, err := accountRepo.GetCustomerAccountsByIDs(accountIDs)
		if err != nil {
			logger.LogErrorf("error getting customers' accounts: %v", err)
			moovhttp.Problem(w, err)
			return
		}

		results := make([]*client.ReportAccountResponse, 0)
		for _, acc := range allAccounts {
			cust, err := customerRepo.GetCustomer(acc.CustomerID, organization)
			if err != nil {
				logger.LogErrorf("error getting customer: %v", err)
				moovhttp.Problem(w, err)
				return
			}
			results = append(results, &client.ReportAccountResponse{
				Customer: *cust,
				Account:  *acc,
			})
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(results); err != nil {
			moovhttp.Problem(w, err)
			return
		}
	}
}
