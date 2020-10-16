package customers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/base/log"

	"github.com/moov-io/customers/pkg/route"
)

var (
	ErrAddressTypeDuplicate = errors.New("customer already has an address with type 'primary'")
)

func AddCustomerAddressRoutes(logger log.Logger, r *mux.Router, repo CustomerRepository) {
	logger = logger.WithKeyValue("package", "customers")

	r.Methods("POST").Path("/customers/{customerID}/addresses").HandlerFunc(createCustomerAddress(logger, repo))
	r.Methods("PUT").Path("/customers/{customerID}/addresses/{addressID}").HandlerFunc(updateCustomerAddress(logger, repo))
	r.Methods("DELETE").Path("/customers/{customerID}/addresses/{addressID}").HandlerFunc(deleteCustomerAddress(logger, repo))
}

func createCustomerAddress(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customerID, requestID := route.GetCustomerID(w, r), moovhttp.GetRequestID(r)
		if customerID == "" {
			return
		}

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		var reqAddr address
		if err := json.NewDecoder(r.Body).Decode(&reqAddr); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		cust, err := repo.GetCustomer(customerID, organization)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if cust == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// todo: vince-10/12/2020: we need to perform this conversion for validation til we develop a clean separation layer between client and model structs
		var addrs []address
		for _, addr := range cust.Addresses {
			addrs = append(addrs, address{
				Type:       strings.ToLower(addr.Type),
				Address1:   addr.Address1,
				Address2:   addr.Address2,
				City:       addr.City,
				State:      addr.State,
				PostalCode: addr.PostalCode,
				Country:    addr.Country,
			})
		}
		if err := validateAddresses(append(addrs, reqAddr)); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if err := repo.addCustomerAddress(customerID, reqAddr); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		logger.Log(fmt.Sprintf("added address for customer=%s", customerID))
		respondWithCustomer(logger, w, customerID, organization, requestID, repo)
	}
}

func updateCustomerAddress(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, addressId := route.GetCustomerID(w, r), getAddressId(w, r)
		if customerID == "" || addressId == "" {
			return
		}

		requestID := moovhttp.GetRequestID(r)
		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		var req updateCustomerAddressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if err := req.validate(); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if req.Type == "primary" {
			cust, err := repo.GetCustomer(customerID, organization)
			if err != nil {
				moovhttp.Problem(w, err)
				return
			}

			for _, addr := range cust.Addresses {
				if addr.Type == "primary" && addr.AddressID != addressId {
					moovhttp.Problem(w, ErrAddressTypeDuplicate)
					return
				}
			}
		}

		if err := repo.updateCustomerAddress(customerID, addressId, req); err != nil {
			logger.LogErrorF("error updating customer's address: customer=%s address=%s: %v", customerID, addressId, err)
			moovhttp.Problem(w, err)
			return
		}

		logger.Log(fmt.Sprintf("updating address=%s for customer=%s", addressId, customerID))

		respondWithCustomer(logger, w, customerID, organization, requestID, repo)
	}
}

func deleteCustomerAddress(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, addressId := route.GetCustomerID(w, r), getAddressId(w, r)
		if customerID == "" || addressId == "" {
			return
		}

		err := repo.deleteCustomerAddress(customerID, addressId)
		if err != nil {
			logger.LogErrorF("error deleting customer's address: customer=%s address=%s: %v", customerID, addressId, err)
			moovhttp.Problem(w, err)
			return
		}

		logger.Log(fmt.Sprintf("successfully deleted address=%s for customer=%s", addressId, customerID))

		w.WriteHeader(http.StatusNoContent)
	}
}

func getAddressId(w http.ResponseWriter, r *http.Request) string {
	varName := "addressID"
	v, ok := mux.Vars(r)[varName]
	if !ok || v == "" {
		moovhttp.Problem(w, fmt.Errorf("path variable %s not found in url", varName))
		return ""
	}
	return v
}

type updateCustomerAddressRequest struct {
	address   `json:",inline"`
	Validated bool `json:"validated"`
}
