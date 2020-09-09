package validator

import (
	"fmt"
	"time"

	"github.com/moov-io/base"
)

type MockRepository struct {
	Validations []*Validation
	Err         error
}

func (r *MockRepository) CreateValidation(validation *Validation) error {
	if r.Err != nil {
		return r.Err
	}

	validation.ValidationID = base.ID()
	now := time.Now()
	validation.CreatedAt = now
	validation.UpdatedAt = now

	if validation.Status == "" {
		validation.Status = StatusInit
	}

	r.Validations = append(r.Validations, validation)
	fmt.Println(r.Validations)

	return nil
}

func (r *MockRepository) GetValidation(accountID, validationID string) (*Validation, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	if len(r.Validations) == 0 {
		return nil, fmt.Errorf("Validation: %s was not found", validationID)
	}

	return r.Validations[0], nil
}

func (r *MockRepository) UpdateValidation(validation *Validation) error {
	if r.Err != nil {
		return r.Err
	}

	return nil
}

func (r *MockRepository) Close() error {
	return nil
}
