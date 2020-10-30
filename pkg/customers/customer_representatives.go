// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"

	"github.com/moov-io/base"
	moovhttp "github.com/moov-io/base/http"

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/model"
	"github.com/moov-io/customers/pkg/route"

	"github.com/moov-io/base/log"
)

func AddCustomerRepresentativeRoutes(logger log.Logger, r *mux.Router, repo CustomerRepository, customerSSNStorage *ssnStorage) {
	logger = logger.Set("package", "customers")

	r.Methods("GET").Path("/customers/{customerID}/representatives").HandlerFunc(searchCustomerRepresentatives(logger, repo))
	r.Methods("GET").Path("/customers/{customerID}/representatives/{representativeID}").HandlerFunc(getCustomerRepresentative(logger, repo))
	r.Methods("PUT").Path("/customers/{customerID}/representatives/{representativeID}").HandlerFunc(updateCustomerRepresentative(logger, repo, customerSSNStorage))
	r.Methods("DELETE").Path("/customers/{customerID}/representatives/{representativeID}").HandlerFunc(deleteCustomerRepresentative(logger, repo))
	r.Methods("POST").Path("/customers/{customerID}/representatives").HandlerFunc(createCustomerRepresentative(logger, repo, customerSSNStorage))
}

func getCustomerRepresentative(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		representativeID, requestID := route.GetRepresentativeID(w, r), moovhttp.GetRequestID(r)
		if representativeID == "" {
			return
		}

		respondWithCustomerRepresentative(logger, w, representativeID, requestID, repo)
	}
}

func deleteCustomerRepresentative(logger log.Logger, repo CustomerRepository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		representativeID := route.GetRepresentativeID(w, r)
		if representativeID == "" {
			return
		}

		err := repo.deleteCustomerRepresentative(representativeID)
		if err != nil {
			moovhttp.Problem(w, fmt.Errorf("deleting customer representative: %v", err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func respondWithCustomerRepresentative(logger log.Logger, w http.ResponseWriter, representativeID, requestID string, repo CustomerRepository) {
	cust, err := repo.GetCustomerRepresentative(representativeID)
	if err != nil {
		logger.LogErrorf("getCustomerRepresentative: lookup: %v", err)
		moovhttp.Problem(w, err)
		return
	}
	if cust == nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cust)
	}
}

// customerRepresentativeRequest holds the information for creating a Customer Representative from the HTTP api
type customerRepresentativeRequest struct {
	RepresentativeID string         `json:"-"`
	CustomerID       string         `json:"customerID"`
	FirstName        string         `json:"firstName"`
	LastName         string         `json:"lastName"`
	JobTitle         string         `json:"jobTitle"`
	BirthDate        model.YYYYMMDD `json:"birthDate"`
	SSN              string         `json:"SSN"`
	Phones           []phone        `json:"phones"`
	Addresses        []address      `json:"addresses"`
}

func createCustomerRepresentative(logger log.Logger, repo CustomerRepository, customerSSNStorage *ssnStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		var req customerRepresentativeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := req.validate(); err != nil {
			logger.LogErrorf("error validating new customer representative: %v", err)
			moovhttp.Problem(w, err)
			return
		}

		representative, ssn, err := req.asCustomerRepresentative(customerSSNStorage)
		if err != nil {
			logger.LogErrorf("problem transforming request into Customer Representative=%s: %v", representative.RepresentativeID, err)
			moovhttp.Problem(w, err)
			return
		}
		if ssn != nil {
			err := customerSSNStorage.repo.saveSSN(ssn)
			if err != nil {
				logger.LogErrorf("problem saving SSN for Customer Representative=%s: %v", representative.RepresentativeID, err)
				moovhttp.Problem(w, fmt.Errorf("saveSSN: %v", err))
				return
			}
		}
		if err := repo.CreateCustomerRepresentative(representative); err != nil {
			logger.LogErrorf("createCustomer: %v", err)
			moovhttp.Problem(w, err)
			return
		}

		logger.Logf("created customer representative=%s", representative.RepresentativeID)

		representative, err = repo.GetCustomerRepresentative(representative.RepresentativeID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(representative)
	}
}

func updateCustomerRepresentative(logger log.Logger, repo CustomerRepository, customerSSNStorage *ssnStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		var req customerRepresentativeRequest
		req.RepresentativeID = route.GetRepresentativeID(w, r)
		if req.RepresentativeID == "" {
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := req.validate(); err != nil {
			logger.LogErrorf("error validating customer payload: %v", err)
			moovhttp.Problem(w, err)
			return
		}

		representative, ssn, err := req.asCustomerRepresentative(customerSSNStorage)
		if err != nil {
			logger.LogErrorf("transforming request into Customer Representative=%s: %v", representative.RepresentativeID, err)
			moovhttp.Problem(w, err)
			return
		}
		if ssn != nil {
			err := customerSSNStorage.repo.saveSSN(ssn)
			if err != nil {
				logger.LogErrorf("error saving SSN for Customer Representative=%s: %v", representative.RepresentativeID, err)
				moovhttp.Problem(w, fmt.Errorf("saving customer's SSN: %v", err))
				return
			}
		}
		if err := repo.updateCustomerRepresentative(representative); err != nil {
			logger.LogErrorf("error updating customer representative: %v", err)
			moovhttp.Problem(w, fmt.Errorf("updating customer representative: %v", err))
			return
		}

		logger.Logf("updated customer representative=%s", representative.RepresentativeID)
		representative, err = repo.GetCustomerRepresentative(representative.RepresentativeID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(representative)
	}
}

func (req customerRepresentativeRequest) validate() error {
	if req.FirstName == "" || req.LastName == "" {
		return errors.New("invalid customer representative fields: empty name field(s)")
	}
	if err := validateAddresses(req.Addresses); err != nil {
		return fmt.Errorf("invalid customer representative addresses: %v", err)
	}
	if err := validatePhones(req.Phones); err != nil {
		return fmt.Errorf("invalid customer representative phone: %v", err)
	}

	return nil
}

func (r *sqlCustomerRepository) GetCustomerRepresentative(representativeID string) (*client.CustomerRepresentative, error) {
	reps, err := r.searchCustomerRepresentatives(RepresentativeSearchParams{
		Count:             1,
		RepresentativeIDs: []string{representativeID},
	})
	if err != nil {
		return nil, fmt.Errorf("getting customer representative: %v", err)
	}

	if len(reps) == 0 {
		return nil, errors.New("customer representative not found")
	}

	return reps[0], nil
}

func (r *sqlCustomerRepository) CreateCustomerRepresentative(c *client.CustomerRepresentative) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Insert customer record
	query := `insert into customer_representatives (representative_id, customer_id, first_name, last_name, birth_date, created_at, last_modified)
values (?, ?, ?, ?, ?, ?, ?);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var birthDate *string
	if c.BirthDate != "" {
		birthDate = &c.BirthDate
	}

	now := time.Now()
	_, err = stmt.Exec(c.RepresentativeID, c.CustomerID, c.FirstName, c.LastName, birthDate, now, now)
	if err != nil {
		return fmt.Errorf("CreateCustomerRepresentative: insert into customer_representatives err=%v | rollback=%v", err, tx.Rollback())
	}

	err = r.updatePhonesByOwnerID(tx, c.RepresentativeID, client.OWNERTYPE_REPRESENTATIVE, c.Phones)
	if err != nil {
		return fmt.Errorf("updating customer representative's phones: %v", err)
	}

	err = r.updateAddressesByOwnerID(tx, c.RepresentativeID, client.OWNERTYPE_REPRESENTATIVE, c.Addresses)
	if err != nil {
		return fmt.Errorf("updating customer representative's addresses: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("CreateCustomerRepresentative: tx.Commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updateCustomerRepresentative(c *client.CustomerRepresentative) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `update customer_representatives set first_name = ?, last_name = ?, birth_date = ?, last_modified = ? where representative_id = ? and customer_id = ? and deleted_at is null;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	res, err := stmt.Exec(c.FirstName, c.LastName, c.BirthDate, now, c.RepresentativeID, c.CustomerID)
	if err != nil {
		return fmt.Errorf("updating customer representative: %v", err)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %v", err)
	}

	if numRows == 0 {
		return fmt.Errorf("no records to update with customer representative id=%s", c.RepresentativeID)
	}

	err = r.updatePhonesByOwnerID(tx, c.RepresentativeID, client.OWNERTYPE_REPRESENTATIVE, c.Phones)
	if err != nil {
		return fmt.Errorf("updating customer representative's phones: %v", err)
	}

	err = r.updateAddressesByOwnerID(tx, c.RepresentativeID, client.OWNERTYPE_REPRESENTATIVE, c.Addresses)
	if err != nil {
		return fmt.Errorf("updating customer representative's addresses: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("updateCustomerRepresentative: tx.Commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) deleteCustomerRepresentative(representativeID string) error {
	query := `update customer_representatives set deleted_at = ? where representative_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), representativeID)
	return err
}

func (req customerRepresentativeRequest) asCustomerRepresentative(storage *ssnStorage) (*client.CustomerRepresentative, *SSN, error) {
	if req.RepresentativeID == "" {
		req.RepresentativeID = base.ID()
	}

	representative := &client.CustomerRepresentative{
		RepresentativeID: req.RepresentativeID,
		CustomerID:       req.CustomerID,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		BirthDate:        string(req.BirthDate),
	}

	for i := range req.Phones {
		representative.Phones = append(representative.Phones, client.Phone{
			Number:    req.Phones[i].Number,
			Type:      req.Phones[i].Type,
			OwnerType: req.Phones[i].OwnerType,
		})
	}
	for i := range req.Addresses {
		representative.Addresses = append(representative.Addresses, client.Address{
			AddressID:  base.ID(),
			Address1:   req.Addresses[i].Address1,
			Address2:   req.Addresses[i].Address2,
			City:       req.Addresses[i].City,
			State:      req.Addresses[i].State,
			PostalCode: req.Addresses[i].PostalCode,
			Country:    req.Addresses[i].Country,
			Type:       req.Addresses[i].Type,
			OwnerType:  req.Addresses[i].OwnerType,
		})
	}
	if req.SSN != "" {
		ssn, err := storage.encryptRaw(representative.RepresentativeID, client.OWNERTYPE_REPRESENTATIVE, req.SSN)
		return representative, ssn, err
	}
	return representative, nil, nil
}
