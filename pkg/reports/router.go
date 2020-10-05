package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	moovhttp "github.com/moov-io/base/http"

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
		requestID := moovhttp.GetRequestID(r)

		accountIDsInput := r.URL.Query().Get("accountIDs")
		accountIDsInput = strings.TrimSpace(accountIDsInput)
		accountIDs := strings.Split(accountIDsInput, ",")

		limit := 25
		if len(accountIDs) > limit {
			moovhttp.Problem(w, fmt.Errorf("exceeded limit of %d accountIDs, found %d", limit, len(accountIDs)))
			return
		}

		allAccounts, err := accountRepo.GetCustomerAccountsByIDs(accountIDs)
		if err != nil {
			logger.Log("customers", "error getting customers' accounts", "error", err, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		results := make([]*client.ReportAccountResponse, 0)
		for _, acc := range allAccounts {
			cust, err := customerRepo.GetCustomer(acc.CustomerID)
			if err != nil {
				logger.Log("customers", "error getting customer", "error", err, "requestID", requestID)
				moovhttp.Problem(w, err)
				return
			}
			results = append(results, &client.ReportAccountResponse{
				Customer: *cust,
				Account:  *acc,
			})
		}

		err = json.NewEncoder(w).Encode(results)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}
