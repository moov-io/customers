package customers

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/usstates"
)

// customerRequest holds the information for creating a Customer from the HTTP api
//
// TODO(adam): What GDPR implications does this information have? IIRC if any EU citizen uses
// this software we have to fully comply.
type customerRequest struct {
	FirstName  string              `json:"firstName"`
	MiddleName string              `json:"middleName"`
	LastName   string              `json:"lastName"`
	NickName   string              `json:"nickName"`
	Suffix     string              `json:"suffix"`
	Type       client.CustomerType `json:"type"`
	BirthDate  *time.Time          `json:"birthDate"`
	Email      string              `json:"email"`
	SSN        string              `json:"SSN"`
	Phones     []phone             `json:"phones"`
	Addresses  []address           `json:"addresses"`
	Metadata   map[string]string   `json:"metadata"`
}

type phone struct {
	Number string `json:"number"`
	Type   string `json:"type"`
}

type address struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"` // TODO(adam): validate against US postal codes
	Country    string `json:"country"`
}

func (add address) validate() error {
	if !usstates.Valid(add.State) {
		return fmt.Errorf("create customer: invalid state=%s", add.State)
	}
	return nil
}

func (req customerRequest) validate() error {
	if req.FirstName == "" || req.LastName == "" {
		return errors.New("create customer: empty name field(s)")
	}
	if err := validateCustomerType(req.Type); err != nil {
		return fmt.Errorf("create customer: %v", err)
	}
	if err := validateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("create customer: %v", err)
	}
	for i := range req.Addresses {
		if err := req.Addresses[i].validate(); err != nil {
			return fmt.Errorf("address=%v validation failed: %v", req.Addresses[i], err)
		}
	}
	return nil
}

func validateCustomerType(t client.CustomerType) error {
	norm := func(t client.CustomerType) string {
		return strings.ToLower(string(t))
	}
	switch norm(t) {
	case norm(client.INDIVIDUAL), norm(client.BUSINESS):
		return nil
	}
	return fmt.Errorf("unknown type: %s", t)
}

func validateMetadata(meta map[string]string) error {
	// both are arbitrary limits, open an issue if this needs bumped
	if len(meta) > 100 {
		return errors.New("metadata is limited to 100 entries")
	}
	for k, v := range meta {
		if length := utf8.RuneCountInString(v); length > 1000 {
			return fmt.Errorf("metadata key %s value is too long at %d", k, length)
		}
	}
	return nil
}

func (req customerRequest) asCustomer(storage *ssnStorage) (*client.Customer, *SSN, error) {
	customer := &client.Customer{
		CustomerID: base.ID(),
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		NickName:   req.NickName,
		Suffix:     req.Suffix,
		Type:       req.Type,
		BirthDate:  req.BirthDate,
		Email:      req.Email,
		Status:     client.UNKNOWN,
		Metadata:   req.Metadata,
	}
	for i := range req.Phones {
		customer.Phones = append(customer.Phones, client.Phone{
			Number: req.Phones[i].Number,
			Type:   req.Phones[i].Type,
		})
	}
	for i := range req.Addresses {
		customer.Addresses = append(customer.Addresses, client.CustomerAddress{
			AddressID:  base.ID(),
			Address1:   req.Addresses[i].Address1,
			Address2:   req.Addresses[i].Address2,
			City:       req.Addresses[i].City,
			State:      req.Addresses[i].State,
			PostalCode: req.Addresses[i].PostalCode,
			Country:    req.Addresses[i].Country,
		})
	}
	if req.SSN != "" {
		ssn, err := storage.encryptRaw(customer.CustomerID, req.SSN)
		return customer, ssn, err
	}
	return customer, nil, nil
}

// formatCustomerName returns a Customer's name joined as one string. It accounts for
// first, middle, last and suffix. Each field is whitespace trimmed.
func formatCustomerName(c *client.Customer) string {
	if c == nil {
		return ""
	}
	out := strings.TrimSpace(c.FirstName)
	if c.MiddleName != "" {
		out += " " + strings.TrimSpace(c.MiddleName)
	}
	out = strings.TrimSpace(out + " " + strings.TrimSpace(c.LastName))
	if c.Suffix != "" {
		out += " " + c.Suffix
	}
	return strings.TrimSpace(out)
}
