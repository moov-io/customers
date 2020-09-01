package customers

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/pkg/client"
)

type customerRepository interface {
	getCustomer(customerID string) (*client.Customer, error)
	createCustomer(c *client.Customer) error
	updateCustomerStatus(customerID string, status client.CustomerStatus, comment string) error

	searchCustomers(params searchParams) ([]*client.Customer, error)

	getCustomerMetadata(customerID string) (map[string]string, error)
	replaceCustomerMetadata(customerID string, metadata map[string]string) error

	addCustomerAddress(customerID string, address address) error
	updateCustomerAddress(customerID, addressID string, _type string, validated bool) error

	getLatestCustomerOFACSearch(customerID string) (*ofacSearchResult, error)
	saveCustomerOFACSearch(customerID string, result ofacSearchResult) error
}

type sqlCustomerRepository struct {
	db *sql.DB
}

func (r *sqlCustomerRepository) close() error {
	return r.db.Close()
}

func (r *sqlCustomerRepository) createCustomer(c *client.Customer) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Insert customer record
	query := `insert into customers (customer_id, first_name, middle_name, last_name, nick_name, suffix, type, birth_date, status, email, created_at, last_modified)
values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	_, err = stmt.Exec(c.CustomerID, c.FirstName, c.MiddleName, c.LastName, c.NickName, c.Suffix, c.Type, c.BirthDate, c.Status, c.Email, now, now)
	if err != nil {
		return fmt.Errorf("createCustomer: insert into customers err=%v | rollback=%v", err, tx.Rollback())
	}

	// Insert customer phone numbers
	query = `replace into customers_phones (customer_id, number, valid, type) values (?, ?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("createCustomer: insert into customers_phones err=%v | rollback=%v", err, tx.Rollback())
	}
	for i := range c.Phones {
		_, err := stmt.Exec(c.CustomerID, c.Phones[i].Number, c.Phones[i].Valid, c.Phones[i].Type)
		if err != nil {
			stmt.Close()
			return fmt.Errorf("createCustomer: customers_phones exec err=%v | rollback=%v", err, tx.Rollback())
		}
	}
	stmt.Close()

	// Insert customer addresses
	query = `replace into customers_addresses(address_id, customer_id, type, address1, address2, city, state, postal_code, country, validated) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("createCustomer: insert into customers_addresses err=%v | rollback=%v", err, tx.Rollback())
	}
	for i := range c.Addresses {
		_, err := stmt.Exec(c.Addresses[i].AddressID, c.CustomerID, c.Addresses[i].Type, c.Addresses[i].Address1, c.Addresses[i].Address2, c.Addresses[i].City, c.Addresses[i].State, c.Addresses[i].PostalCode, c.Addresses[i].Country, c.Addresses[i].Validated)
		if err != nil {
			stmt.Close()
			return fmt.Errorf("createCustomer: customers_addresses exec err=%v | rollback=%v", err, tx.Rollback())
		}
	}
	stmt.Close()

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("createCustomer: tx.Commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) getCustomer(customerID string) (*client.Customer, error) {
	query := `select first_name, middle_name, last_name, nick_name, suffix, type, birth_date, status, email, created_at, last_modified from customers where customer_id = ? and deleted_at is null limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(customerID)

	var cust client.Customer
	cust.CustomerID = customerID
	err = row.Scan(&cust.FirstName, &cust.MiddleName, &cust.LastName, &cust.NickName, &cust.Suffix, &cust.Type, &cust.BirthDate, &cust.Status, &cust.Email, &cust.CreatedAt, &cust.LastModified)
	stmt.Close()
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return nil, fmt.Errorf("getCustomer: %v", err)
	}
	if cust.FirstName == "" {
		return nil, nil // not found
	}

	phones, err := r.readPhones(customerID)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: readPhones: %v", err)
	}
	cust.Phones = phones

	addresses, err := r.readAddresses(customerID)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: readAddresses: %v", err)
	}
	cust.Addresses = addresses

	metadata, err := r.getCustomerMetadata(customerID)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: getCustomerMetadata: %v", err)
	}
	cust.Metadata = metadata

	return &cust, nil
}

func (r *sqlCustomerRepository) readPhones(customerID string) ([]client.Phone, error) {
	query := `select number, valid, type from customers_phones where customer_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: prepare customers_phones: err=%v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerID)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: query customers_phones: err=%v", err)
	}
	defer rows.Close()

	var phones []client.Phone
	for rows.Next() {
		var p client.Phone
		if err := rows.Scan(&p.Number, &p.Valid, &p.Type); err != nil {
			return nil, fmt.Errorf("getCustomer: scan customers_phones: err=%v", err)
		}
		phones = append(phones, p)
	}
	return phones, rows.Err()
}

func (r *sqlCustomerRepository) readAddresses(customerID string) ([]client.CustomerAddress, error) {
	query := `select address_id, type, address1, address2, city, state, postal_code, country, validated from customers_addresses where customer_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("readAddresses: prepare customers_addresses: err=%v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerID)
	if err != nil {
		return nil, fmt.Errorf("readAddresses: query customers_addresses: err=%v", err)
	}
	defer rows.Close()

	var adds []client.CustomerAddress
	for rows.Next() {
		var a client.CustomerAddress
		if err := rows.Scan(&a.AddressID, &a.Type, &a.Address1, &a.Address2, &a.City, &a.State, &a.PostalCode, &a.Country, &a.Validated); err != nil {
			return nil, fmt.Errorf("readAddresses: scan customers_addresses: err=%v", err)
		}
		adds = append(adds, a)
	}
	return adds, rows.Err()
}

func (r *sqlCustomerRepository) updateCustomerStatus(customerID string, status client.CustomerStatus, comment string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("updateCustomerStatus: tx begin: %v", err)
	}

	// update 'customers' table
	query := `update customers set status = ? where customer_id = ?;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("updateCustomerStatus: update customers prepare: %v", err)
	}
	if _, err := stmt.Exec(status, customerID); err != nil {
		stmt.Close()
		return fmt.Errorf("updateCustomerStatus: update customers exec: %v", err)
	}
	stmt.Close()

	// update 'customer_status_updates' table
	query = `insert into customer_status_updates (customer_id, future_status, comment, changed_at) values (?, ?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("updateCustomerStatus: insert status prepare: %v", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(customerID, status, comment, time.Now()); err != nil {
		return fmt.Errorf("updateCustomerStatus: insert status exec: %v", err)
	}
	return tx.Commit()
}

func (r *sqlCustomerRepository) getCustomerMetadata(customerID string) (map[string]string, error) {
	out := make(map[string]string)

	query := `select meta_key, meta_value from customer_metadata where customer_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return out, fmt.Errorf("getCustomerMetadata: prepare: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerID)
	if err != nil {
		return out, fmt.Errorf("getCustomerMetadata: query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		key, value := "", ""
		if err := rows.Scan(&key, &value); err != nil {
			return out, fmt.Errorf("getCustomerMetadata: scan: %v", err)
		}
		out[key] = value
	}
	return out, nil
}

func (r *sqlCustomerRepository) replaceCustomerMetadata(customerID string, metadata map[string]string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("replaceCustomerMetadata: tx begin: %v", err)
	}

	// Delete each existing k/v pair
	query := `delete from customer_metadata where customer_id = ?;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("replaceCustomerMetadata: delete prepare: %v", err)
	}
	if _, err := stmt.Exec(customerID); err != nil {
		stmt.Close()
		return fmt.Errorf("replaceCustomerMetadata: delete exec: %v", err)
	}
	stmt.Close()

	// Insert each k/v pair
	query = `insert into customer_metadata (customer_id, meta_key, meta_value) values (?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("replaceCustomerMetadata: insert prepare: %v", err)
	}
	defer stmt.Close()
	for k, v := range metadata {
		if _, err := stmt.Exec(customerID, k, v); err != nil {
			return fmt.Errorf("replaceCustomerMetadata: insert %s: %v", k, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("replaceCustomerMetadata: commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) addCustomerAddress(customerID string, req address) error {
	query := `insert into customers_addresses (address_id, customer_id, type, address1, address2, city, state, postal_code, country, validated) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("addCustomerAddress: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(base.ID(), customerID, "Secondary", req.Address1, req.Address2, req.City, req.State, req.PostalCode, req.Country, false); err != nil {
		return fmt.Errorf("addCustomerAddress: exec: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updateCustomerAddress(customerID, addressID string, _type string, validated bool) error {
	query := `update customers_addresses set type = ?, validated = ? where customer_id = ? and address_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("updateCustomerAddress: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(_type, validated, customerID, addressID); err != nil {
		return fmt.Errorf("updateCustomerAddress: exec: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) getLatestCustomerOFACSearch(customerID string) (*ofacSearchResult, error) {
	query := `select entity_id, sdn_name, sdn_type, match, created_at from customer_ofac_searches where customer_id = ? order by created_at desc limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getLatestCustomerOFACSearch: prepare: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(customerID)
	var res ofacSearchResult
	if err := row.Scan(&res.EntityID, &res.SDNName, &res.SDNType, &res.Match, &res.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // nothing found
		}
		return nil, fmt.Errorf("getLatestCustomerOFACSearch: scan: %v", err)
	}
	return &res, nil
}

func (r *sqlCustomerRepository) saveCustomerOFACSearch(customerID string, result ofacSearchResult) error {
	query := `insert into customer_ofac_searches (customer_id, entity_id, sdn_name, sdn_type, match, created_at) values (?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: prepare: %v", err)
	}
	defer stmt.Close()

	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}

	if _, err := stmt.Exec(customerID, result.EntityID, result.SDNName, result.SDNType, result.Match, result.CreatedAt); err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: exec: %v", err)
	}
	return nil
}
