package customers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	moovhttp "github.com/moov-io/base/http"
	client "github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/identity/pkg/logging"
	tmw "github.com/moov-io/tumbler/pkg/middleware"

	"github.com/gorilla/mux"
)

type Controller interface {
	AppendRoutes(router *mux.Router) *mux.Router
}

func NewCustomersController(
	logger logging.Logger,
	ofac *ofacSearcher,
	repo customerRepository,
	ssnStorage *ssnStorage,
) Controller {
	return &customersController{
		logger:     logger,
		ofac:       ofac,
		repo:       repo,
		ssnStorage: ssnStorage,
	}
}

type customersController struct {
	logger logging.Logger

	ofac       *ofacSearcher
	repo       customerRepository
	ssnStorage *ssnStorage
}

func (c customersController) AppendRoutes(router *mux.Router) *mux.Router {
	router.
		Name("customers.searchCustomers").
		Methods("GET").
		Path("/customers").
		HandlerFunc(c.searchCustomersHandler)

	router.
		Name("customers.createCustomer").
		Methods("POST").
		Path("/customers").
		HandlerFunc(c.createCustomerHandler)

	router.
		Name("customers.retrieveCustomer").
		Methods("GET").
		Path("/customers/{customerID}").
		HandlerFunc(c.retrieveCustomerHandler)

	router.
		Name("customers.addCustomerAddress").
		Methods("POST").
		Path("/customers/{customerID}/address").
		HandlerFunc(c.addCustomerAddressHandler)

	router.
		Name("customers.updateCustomerMetadata").
		Methods("PUT").
		Path("/customers/{customerID}/metadata").
		HandlerFunc(c.updateCustomerMetadataHandler)

	router.
		Name("customers.updateCustomerStatus").
		Methods("PUT").
		Path("/customers/{customerID}/status").
		HandlerFunc(c.updateCustomerStatusHandler)

	router.
		Name("customers.getCustomerDisclaimers").
		Methods("POST").
		Path("/customers/{customerID}/disclaimers").
		HandlerFunc(c.getCustomerDisclaimersHandler)

	router.
		Name("customers.acceptDisclaimer").
		Methods("POST").
		Path("/customers/{customerID}/disclaimers/{disclaimerID}").
		HandlerFunc(c.acceptDisclaimerHandler)

	router.
		Name("customers.getCustomerDocuments").
		Methods("GET").
		Path("/customers/{customerID}/documents").
		HandlerFunc(c.getCustomerDocumentsHandler)

	router.
		Name("customers.uploadCustomerDocument").
		Methods("POST").
		Path("/customers/{customerID}/documents").
		HandlerFunc(c.uploadCustomerDocumentHandler)

	router.
		Name("customers.getCustomerDocument").
		Methods("GET").
		Path("/customers/{customerID}/documents/{documentID}").
		HandlerFunc(c.getCustomerDocumentHandler)

	router.
		Name("customers.getLatestOFACSearch").
		Methods("GET").
		Path("/customers/{customerID}/ofac").
		HandlerFunc(c.getLatestOFACSearchHandler)

	router.
		Name("customers.refreshOFACSearch").
		Methods("PUT").
		Path("/customers/{customerID}/refresh/ofac").
		HandlerFunc(c.refreshOFACSearchHandler)

	return router
}

// func getCustomer(logger log.Logger, repo customerRepository) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		w = route.Responder(logger, w, r)
//
// 		customerID := route.GetCustomerID(w, r)
// 		if customerID == "" {
// 			return
// 		}
//
// 		respondWithCustomer(c.logger, w, customerID, c.repo)
// 	}
// }

func (c *customersController) createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		var req customerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, err)
			return
		}
		if err := req.validate(); err != nil {
			c.logger.Log(fmt.Sprintf("error validating new customer error=%v", err))
			errorResponse(w, err)
			return
		}

		cust, ssn, err := req.asCustomer(c.ssnStorage)
		if err != nil {
			c.logger.Log(fmt.Sprintf("problem transforming request into Customer=%s: %v", cust.CustomerID, err))
			errorResponse(w, err)
			return
		}
		if ssn != nil {
			err := c.ssnStorage.repo.saveCustomerSSN(ssn)
			if err != nil {
				c.logger.Log(fmt.Sprintf("problem saving SSN for Customer=%s: %v", cust.CustomerID, err))
				errorResponse(w, fmt.Errorf("saveCustomerSSN: %v", err))
				return
			}
		}
		if err := c.repo.createCustomer(cust); err != nil {
			errorResponse(w, err)
			return
		}
		if err := c.repo.replaceCustomerMetadata(cust.CustomerID, cust.Metadata); err != nil {
			c.logger.Log(fmt.Sprintf("updating metadata for customer=%s failed: %v", cust.CustomerID, err))
			errorResponse(w, err)
			return
		}

		// Perform an OFAC search with the Customer information
		if err := c.ofac.storeCustomerOFACSearch(cust); err != nil {
			c.logger.Log(fmt.Sprintf("error with OFAC search for customer=%s: %v", cust.CustomerID, err))
		}

		c.logger.Log(fmt.Sprintf("created customer=%s", cust.CustomerID))

		cust, err = c.repo.getCustomer(cust.CustomerID)
		if err != nil {
			errorResponse(w, err)
			return
		}

		jsonResponse(w, cust)
	})
}

func (c *customersController) retrieveCustomerHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		customerID := mux.Vars(r)["customerID"]
		if customerID == "" {
			errorResponse(w, errors.New("missing customerID"))
			return
		}
		respondWithCustomer(c.logger, w, customerID, c.repo)
	})
}

func (c *customersController) addCustomerAddressHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		customerID := mux.Vars(r)["customerID"]
		if customerID == "" {
			errorResponse(w, errors.New("missing customerID"))
			return
		}

		var req address
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, err)
			return
		}

		if err := c.repo.addCustomerAddress(customerID, req); err != nil {
			errorResponse(w, err)
			return
		}

		c.logger.Log(fmt.Sprintf("added address for customer=%s", customerID))

		respondWithCustomer(c.logger, w, customerID, c.repo)

		// customerID := mux.Vars(r)["customerID"]
		// if customerID == "" {
		// 	return
		// }

		// var req updateCustomerAddressRequest
		// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 	moovhttp.Problem(w, err)
		// 	return
		// }
		// if err := req.validate(); err != nil {
		// 	moovhttp.Problem(w, err)
		// 	return
		// }

		// logger.Log("approval", fmt.Sprintf("updating address=%s for customer=%s", addressId, customerID))

		// if err := repo.updateCustomerAddress(customerID, addressId, req.Type, req.Validated); err != nil {
		// 	logger.Log("approval", fmt.Sprintf("error updating customer=%s address=%s: %v", customerID, addressId, err))
		// 	moovhttp.Problem(w, err)
		// 	return
		// }
		// respondWithCustomer(logger, w, customerID, repo)
	})
}

func respondWithCustomer(logger logging.Logger, w http.ResponseWriter, customerID string, repo customerRepository) {
	cust, err := repo.getCustomer(customerID)
	if err != nil {
		logger.Log(fmt.Sprintf("getCustomer: lookup: %v", err))
		moovhttp.Problem(w, err)
		return
	}
	if cust == nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		jsonResponse(w, cust)
	}
}

type replaceMetadataRequest struct {
	Metadata map[string]string `json:"metadata"`
}

func (c *customersController) updateCustomerMetadataHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		var req replaceMetadataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, err)
			return
		}
		if err := validateMetadata(req.Metadata); err != nil {
			errorResponse(w, err)
			return
		}
		customerID := mux.Vars(r)["customerID"]
		if customerID == "" {
			return
		}
		if err := c.repo.replaceCustomerMetadata(customerID, req.Metadata); err != nil {
			errorResponse(w, err)
			return
		}
		respondWithCustomer(c.logger, w, customerID, c.repo)
	})
}

func (c *customersController) updateCustomerStatusHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		customerID := mux.Vars(r)["customerID"]
		if customerID == "" {
			errorResponse(w, errors.New("missing customerID"))
			return
		}

		var req updateCustomerStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		cust, err := c.repo.getCustomer(customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if cust == nil {
			moovhttp.Problem(w, fmt.Errorf("customerID=%s not found", customerID))
			return
		}

		// Update Customer's status in the database
		if err := c.repo.updateCustomerStatus(customerID, req.Status, req.Comment); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		respondWithCustomer(c.logger, w, customerID, c.repo)
	})
}

func (c *customersController) getCustomerDisclaimersHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		params := mux.Vars(r)
		_ = params["customerID"]

		// @TODO do stuff
		result, err := c.stub()
		if err != nil {
			errorResponse(w, err)
			return
		}

		jsonResponse(w, result)
	})
}

func (c *customersController) acceptDisclaimerHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		params := mux.Vars(r)
		_ = params["customerID"]
		_ = params["disclaimerID"]

		// @TODO do stuff
		result, err := c.stub()
		if err != nil {
			errorResponse(w, err)
			return
		}

		jsonResponse(w, result)
	})
}

func (c *customersController) getCustomerDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		params := mux.Vars(r)
		_ = params["customerID"]

		// @TODO do stuff
		result, err := c.stub()
		if err != nil {
			errorResponse(w, err)
			return
		}

		jsonResponse(w, result)
	})
}

func (c *customersController) uploadCustomerDocumentHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		params := mux.Vars(r)
		_ = params["customerID"]

		// @TODO do stuff
		result, err := c.stub()
		if err != nil {
			errorResponse(w, err)
			return
		}

		jsonResponse(w, result)
	})
}

func (c *customersController) getCustomerDocumentHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		params := mux.Vars(r)
		_ = params["customerID"]
		_ = params["documentID"]

		// @TODO do stuff
		result, err := c.stub()
		if err != nil {
			errorResponse(w, err)
			return
		}

		jsonResponse(w, result)
	})
}

func (c *customersController) getLatestOFACSearchHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		customerID := mux.Vars(r)["customerID"]
		if customerID == "" {
			errorResponse(w, errors.New("missing customerID"))
			return
		}
		result, err := c.repo.getLatestCustomerOFACSearch(customerID)
		if err != nil {
			errorResponse(w, err)
			return
		}
		jsonResponse(w, result)
	})
}

func (c *customersController) refreshOFACSearchHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		customerID := mux.Vars(r)["customerID"]
		if customerID == "" {
			errorResponse(w, errors.New("missing customerID"))
			return
		}

		cust, err := c.repo.getCustomer(customerID)
		if err != nil {
			errorResponse(w, err)
			return
		}

		c.logger.Log(fmt.Sprintf("running live OFAC search for customer=%s", customerID))

		if err := c.ofac.storeCustomerOFACSearch(cust); err != nil {
			c.logger.Log(fmt.Sprintf("error refreshing ofac search: %v", err))
			errorResponse(w, err)
			return
		}
		result, err := c.repo.getLatestCustomerOFACSearch(customerID)
		if err != nil {
			c.logger.Log(fmt.Sprintf("error getting latest ofac search: %v", err))
			errorResponse(w, err)
			return
		}
		if result.Match > ofacMatchThreshold {
			err = fmt.Errorf("customer=%s matched against OFAC entity=%s with a score of %.2f - rejecting customer", cust.CustomerID, result.EntityID, result.Match)
			c.logger.Log(err.Error())

			if err := c.repo.updateCustomerStatus(cust.CustomerID, client.REJECTED, "manual OFAC refresh"); err != nil {
				c.logger.Log(fmt.Sprintf("error updating customer=%s error=%v", cust.CustomerID, err))
				errorResponse(w, err)
				return
			}
		}
		jsonResponse(w, result)
	})
}

// Delete once all the stubs have been removed.
func (c *customersController) stub() (string, error) {
	return "", errors.New("not implemented yet")
}
