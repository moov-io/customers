package validator

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/moov-io/base"
)

type Validation struct {
	ValidationID string
	AccountID    string
	Status       string
	Strategy     string
	Vendor       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

const (
	StatusInit      = "init"
	StatusCompleted = "completed"
)

type Repository interface {
	CreateValidation(*Validation) error
	GetValidation(accountID, validationID string) (*Validation, error)
	UpdateValidation(*Validation) error
}

func NewRepo(db *sql.DB) Repository {
	return &sqlRepository{
		db: db,
	}
}

type sqlRepository struct {
	db *sql.DB
}

func (r *sqlRepository) CreateValidation(validation *Validation) error {
	validation.ValidationID = base.ID()
	now := time.Now()
	validation.CreatedAt = now
	validation.UpdatedAt = now

	query := `
		insert into validations (
			validation_id,
			account_id,
			status,
			strategy,
			vendor,
			created_at,
			updated_at
		) values (?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		validation.ValidationID,
		validation.AccountID,
		validation.Status,
		validation.Strategy,
		validation.Vendor,
		validation.CreatedAt,
		validation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("problem creating validation=%s: %v", validation.ValidationID, err)
	}
	return nil
}

func (r *sqlRepository) GetValidation(accountID, validationID string) (*Validation, error) {
	query := `
		select
			validation_id,
			account_id,
			status,
			strategy,
			vendor,
			created_at,
			updated_at
		from
			validations
		where
			validation_id = ? and
			account_id = ?
		limit 1;
		`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	validation := Validation{}
	row := stmt.QueryRow(validationID, accountID)
	if err := row.Scan(
		&validation.ValidationID,
		&validation.AccountID,
		&validation.Status,
		&validation.Strategy,
		&validation.Vendor,
		&validation.CreatedAt,
		&validation.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &validation, nil
}

func (r *sqlRepository) UpdateValidation(validation *Validation) error {
	now := time.Now()
	query := `
		update
			validations
		set
			account_id = ?,
			status = ?,
			strategy = ?,
			vendor = ?,
			updated_at = ?
		where
			validation_id = ?;
	`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		validation.AccountID,
		validation.Status,
		validation.Strategy,
		validation.Vendor,
		time.Now(),
		validation.ValidationID,
	)
	if err != nil {
		return fmt.Errorf("problem updating validation=%s: %v", validation.ValidationID, err)
	}

	validation.UpdatedAt = now
	return nil
}

func (r *sqlRepository) Close() error {
	return r.db.Close()
}
