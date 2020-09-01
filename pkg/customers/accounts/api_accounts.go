package accounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/moov-io/ach"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/fed"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/moov-io/customers/pkg/secrets/hash"
	"github.com/moov-io/customers/pkg/secrets/mask"
	"github.com/moov-io/identity/pkg/logging"
	tmw "github.com/moov-io/tumbler/pkg/middleware"

	"github.com/gorilla/mux"
)

type Controller interface {
	AppendRoutes(router *mux.Router) *mux.Router
}

func NewAccountsController(
	logger logging.Logger,
	repo Repository,
	keeper *secrets.StringKeeper,
	fedClient fed.Client,
) Controller {
	return &accountsController{
		logger:    logger,
		repo:      repo,
		keeper:    keeper,
		fedClient: fedClient,
	}
}

type accountsController struct {
	logger    logging.Logger
	repo      Repository
	keeper    *secrets.StringKeeper
	fedClient fed.Client
}

func (c accountsController) AppendRoutes(router *mux.Router) *mux.Router {
	router.
		Name("customers.getCustomerAccounts").
		Methods("GET").
		Path("/customers/{customerID}/accounts").
		HandlerFunc(c.getCustomerAccountsHandler)

	router.
		Name("customers.createCustomerAccount").
		Methods("POST").
		Path("/customers/{ID}/accounts").
		HandlerFunc(c.createCustomerAccountHandler)

	router.
		Name("customers.deleteCustomerAccount").
		Methods("DELETE").
		Path("/customers/{customerID}/accounts/{accountID}").
		HandlerFunc(c.deleteCustomerAccountHandler)

	router.
		Name("customers.decryptAccountNumber").
		Methods("POST").
		Path("/customers/{customerID}/accounts/{accountID}/decrypt").
		HandlerFunc(c.decryptAccountNumberHandler)

	router.
		Name("customers.validateAccount").
		Methods("POST").
		Path("/customers/{customerID}/accounts/{accountID}/validate").
		HandlerFunc(c.validateAccountHandler)

	return router
}

func (c *accountsController) getCustomerAccountsHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		customerID := getCustomerID(w, r)
		if customerID == "" {
			return
		}
		accounts, err := c.repo.getCustomerAccounts(customerID)
		if err != nil {
			errorResponse(w, err)
			return
		}
		accounts = decorateInstitutionDetails(accounts, c.fedClient)
		jsonResponse(w, accounts)
	})
}

func decorateInstitutionDetails(accounts []*client.Account, client fed.Client) []*client.Account {
	for i := range accounts {
		if details, _ := client.LookupInstitution(accounts[i].RoutingNumber); details != nil {
			accounts[i].Institution = *details
		}
	}
	return accounts
}

type createAccountRequest struct {
	HolderName    string             `json:"holderName"`
	AccountNumber string             `json:"accountNumber"`
	RoutingNumber string             `json:"routingNumber"`
	Type          client.AccountType `json:"type"`

	// fields we compute from the inbound AccountNumber
	encryptedAccountNumber string
	hashedAccountNumber    string
	maskedAccountNumber    string
}

func (req *createAccountRequest) validate() error {
	if req.HolderName == "" {
		return errors.New("missing HolderName")
	}
	if req.AccountNumber == "" {
		return errors.New("missing AccountNumber")
	}
	if err := ach.CheckRoutingNumber(req.RoutingNumber); err != nil {
		return err
	}

	at := func(t1, t2 client.AccountType) bool {
		return strings.EqualFold(string(t1), string(t2))
	}
	if !at(req.Type, client.CHECKING) && !at(req.Type, client.SAVINGS) {
		return fmt.Errorf("invalid account type: %s", req.Type)
	}

	return nil
}

func (req *createAccountRequest) disfigure(keeper *secrets.StringKeeper) error {
	if enc, err := keeper.EncryptString(req.AccountNumber); err != nil {
		return fmt.Errorf("problem encrypting account number: %v", err)
	} else {
		req.encryptedAccountNumber = enc
	}
	if v, err := hash.AccountNumber(req.AccountNumber); err != nil {
		return fmt.Errorf("problem hashing account number: %v", err)
	} else {
		req.hashedAccountNumber = v
	}
	req.maskedAccountNumber = mask.AccountNumber(req.AccountNumber)
	return nil
}

func getAccountID(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["accountID"]
	if !ok || v == "" {
		errorResponse(w, errors.New("missing accountID"))
		return ""
	}
	return v
}

func getCustomerID(w http.ResponseWriter, r *http.Request) string {
	customerID := mux.Vars(r)["customerID"]
	if customerID == "" {
		errorResponse(w, errors.New("missing customerID"))
		return ""
	}
	return customerID
}

func (c *accountsController) createCustomerAccountHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		var request createAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			errorResponse(w, err)
			return
		}
		if err := request.validate(); err != nil {
			c.logger.Log(fmt.Sprintf("problem validating account: %v", err))
			errorResponse(w, err)
			return
		}
		if err := request.disfigure(c.keeper); err != nil {
			c.logger.Log(fmt.Sprintf("problem disfiguring account: %v", err))
			errorResponse(w, err)
			return
		}

		if _, err := c.fedClient.LookupInstitution(request.RoutingNumber); err != nil {
			c.logger.Log(fmt.Sprintf("problem looking up routing number=%q: %v", request.RoutingNumber, err))
			errorResponse(w, err)
			return
		}

		customerID, userID := getCustomerID(w, r), moovhttp.GetUserID(r)
		account, err := c.repo.createCustomerAccount(customerID, userID, &request)
		if err != nil {
			c.logger.Log(fmt.Sprintf("problem saving account: %v", err))
			errorResponse(w, err)
			return
		}

		jsonResponse(w, account)
	})
}

func (c *accountsController) deleteCustomerAccountHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		params := mux.Vars(r)
		_ = params["customerID"]
		_ = params["accountID"]

		// @TODO do stuff
		result, err := c.stub()
		if err != nil {
			errorResponse(w, err)
			return
		}

		jsonResponse(w, result)
	})
}

func (c *accountsController) decryptAccountNumberHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		params := mux.Vars(r)
		_ = params["customerID"]
		_ = params["accountID"]

		// @TODO do stuff
		result, err := c.stub()
		if err != nil {
			errorResponse(w, err)
			return
		}

		jsonResponse(w, result)
	})
}

func (c *accountsController) validateAccountHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		params := mux.Vars(r)
		_ = params["customerID"]
		_ = params["accountID"]

		// @TODO do stuff
		result, err := c.stub()
		if err != nil {
			errorResponse(w, err)
			return
		}

		jsonResponse(w, result)
	})
}

// Delete once all the stubs have been removed.
func (c *accountsController) stub() (string, error) {
	return "", errors.New("not implemented yet")
}
