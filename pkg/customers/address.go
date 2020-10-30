package customers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/moov-io/customers/pkg/client"
	"net/http"

	"github.com/gorilla/mux"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/base/log"

	"github.com/moov-io/customers/pkg/route"
)

var (
	ErrAddressTypeDuplicate = errors.New("customer already has an address with type 'primary'")
)

func AddCustomerAddressRoutes(logger log.Logger, r *mux.Router, repo CustomerRepository) {
	logger = logger.Set("package", "customers")

	r.Methods("POST").Path("/customers/{customerID}/addresses").HandlerFunc(createAddress(logger, client.OWNERTYPE_CUSTOMER, repo))
	r.Methods("PUT").Path("/customers/{customerID}/addresses/{addressID}").HandlerFunc(updateAddress(logger, client.OWNERTYPE_CUSTOMER, repo))
	r.Methods("DELETE").Path("/customers/{customerID}/addresses/{addressID}").HandlerFunc(deleteAddress(logger, client.OWNERTYPE_CUSTOMER, repo))

	r.Methods("POST").Path("/customers/{customerID}/representatives/{representativeID}/addresses").HandlerFunc(createAddress(logger, client.OWNERTYPE_REPRESENTATIVE, repo))
	r.Methods("PUT").Path("/customers/{customerID}/representatives/{representativeID}/addresses/{addressID}").HandlerFunc(updateAddress(logger, client.OWNERTYPE_REPRESENTATIVE, repo))
	r.Methods("DELETE").Path("/customers/{customerID}/representatives/{representativeID}/addresses/{addressID}").HandlerFunc(deleteAddress(logger, client.OWNERTYPE_REPRESENTATIVE, repo))
}

func createAddress(logger log.Logger, ownerType client.OwnerType, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customerID, requestID := route.GetCustomerID(w, r), moovhttp.GetRequestID(r)
		if customerID == "" {
			return
		}

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		var ownerID string
		var addresses []client.Address
		if ownerType == client.OWNERTYPE_CUSTOMER {
			ownerID = customerID
			cust, err := repo.GetCustomer(customerID, organization)
			if err != nil {
				moovhttp.Problem(w, err)
				return
			}
			if cust == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			addresses = cust.Addresses
		} else {
			representativeID := route.GetRepresentativeID(w, r)
			if representativeID == "" {
				return
			}
			rep, err := repo.GetCustomerRepresentative(representativeID)
			if err != nil {
				moovhttp.Problem(w, err)
				return
			}
			if rep == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			addresses = rep.Addresses
		}

		//ownerID, owner, ownerType, err := getOwnerIDAndType(w, customerID, representativeID, organization, repo)

		var reqAddr address
		if err := json.NewDecoder(r.Body).Decode(&reqAddr); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// todo: vince-10/12/2020: we need to perform this conversion for validation til we develop a clean separation layer between client and model structs
		var addrs []address
		for _, addr := range addresses {
			addrs = append(addrs, address{
				Type:       addr.Type,
				OwnerType:  addr.OwnerType,
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

		if err := repo.addAddress(ownerID, ownerType, reqAddr); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		logger.Logf("added address for %s=%s", string(ownerType), ownerID)
		respondWithCustomer(logger, w, customerID, organization, requestID, repo)
	}
}

func updateAddress(logger log.Logger, ownerType client.OwnerType, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, addressId := route.GetCustomerID(w, r), getAddressId(w, r)
		if customerID == "" || addressId == "" {
			return
		}

		ownerID := customerID
		var representativeID string
		if ownerType == client.OWNERTYPE_REPRESENTATIVE {
			representativeID = route.GetRepresentativeID(w, r)
			ownerID = representativeID
		}

		requestID := moovhttp.GetRequestID(r)
		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		var req updateAddressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if err := req.validate(); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if req.Type == "primary" {
			var addresses []client.Address
			if ownerType == client.OWNERTYPE_CUSTOMER {
				cust, err := repo.GetCustomer(customerID, organization)
				if err != nil {
					moovhttp.Problem(w, err)
					return
				}
				addresses = cust.Addresses
			} else {
				rep, err := repo.GetCustomerRepresentative(representativeID)
				if err != nil {
					moovhttp.Problem(w, err)
					return
				}
				addresses = rep.Addresses

			}

			for _, addr := range addresses {
				if addr.Type == "primary" && addr.AddressID != addressId {
					moovhttp.Problem(w, ErrAddressTypeDuplicate)
					return
				}
			}
		}

		if err := repo.updateAddress(ownerID, addressId, ownerType, req); err != nil {
			logger.LogErrorf("error updating %s's address: %s=%s address=%s: %v", string(ownerType), string(ownerType), ownerID, addressId, err)
			moovhttp.Problem(w, err)
			return
		}

		logger.Logf("updating address=%s for %s=%s", addressId, string(ownerType), customerID)

		respondWithCustomer(logger, w, customerID, organization, requestID, repo)
	}
}

func deleteAddress(logger log.Logger, ownerType client.OwnerType, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, addressId := route.GetCustomerID(w, r), getAddressId(w, r)
		if customerID == "" {
			return
		}

		if addressId == "" {
			return
		}

		ownerID := customerID
		if ownerType == client.OWNERTYPE_REPRESENTATIVE {
			representativeID := route.GetRepresentativeID(w, r)
			ownerID = representativeID
		}

		err := repo.deleteAddress(ownerID, ownerType, addressId)
		if err != nil {
			logger.LogErrorf("error deleting %s's address: %s=%s address=%s: %v", string(ownerType), string(ownerType), customerID, addressId, err)
			moovhttp.Problem(w, err)
			return
		}

		logger.Logf("successfully deleted address=%s for %s=%s", addressId, string(ownerType), customerID)

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

type updateAddressRequest struct {
	address   `json:",inline"`
	Validated bool `json:"validated"`
}
