package accounts

import (
	"github.com/moov-io/customers/pkg/client"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CreateAccount client.CreateAccount

func (c CreateAccount) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.HolderName, validation.Required, validation.Length(2, 60)),
		validation.Field(&c.AccountNumber, validation.Required, validation.Length(2, 20)),
		validation.Field(&c.RoutingNumber, validation.Required, validation.Length(2, 10)),
		// TODO(adam): validate type is an acceptable value
		validation.Field(&c.Type, validation.Required, validation.Length(2, 12)),
	)
}
