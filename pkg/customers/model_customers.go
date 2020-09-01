package customers

import (
	"fmt"

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/usstates"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CreateCustomer client.CreateCustomer

func (c CreateCustomer) Validate() error {
	// switch strings.ToLower(string(c.Type)) {
	// // case strings.ToLower(
	// }
	return validation.ValidateStruct(&c,
		validation.Field(&c.FirstName, validation.Required, validation.Length(2, 255)),
		validation.Field(&c.LastName, validation.Required, validation.Length(2, 255)),
		// TODO(adam): is there a builtin validator?
		validation.Field(&c.Email, validation.Required, validation.Length(3, 255)),
	)
}

type CreateCustomerAddress client.CreateCustomerAddress

func (c CreateCustomerAddress) Validate() error {
	if !usstates.Valid(c.State) {
		return fmt.Errorf("invalid state=%s", c.State)
	}
	return validation.ValidateStruct(&c,
		validation.Field(&c.Address1, validation.Required, validation.Length(2, 120)),
		validation.Field(&c.City, validation.Required, validation.Length(2, 50)),
		// TODO(adam): validate with usstates package
		validation.Field(&c.State, validation.Required, validation.Length(2, 2)),
		validation.Field(&c.PostalCode, validation.Required, validation.Length(2, 10)),
		// TODO(adam): validate?
		validation.Field(&c.Country, validation.Required, validation.Length(2, 3)),
	)
}

type CustomerMetadata client.CustomerMetadata

func (c CustomerMetadata) Validate() error {
	return nil
}

type UpdateCustomerStatus client.UpdateCustomerStatus

func (c UpdateCustomerStatus) Validate() error {
	return validation.ValidateStruct(&c,
		// TODO(adam): validate status is an acceptable value
		validation.Field(&c.Status, validation.Required, validation.Length(2, 20)),
	)
}
