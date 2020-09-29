// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	kitprom "github.com/go-kit/kit/metrics/prometheus"
	"github.com/lopezator/migrator"
	"github.com/mattn/go-sqlite3"
	stdprom "github.com/prometheus/client_golang/prometheus"
)

var (
	sqliteConnections = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "sqlite_connections",
		Help: "How many sqlite connections and what status they're in.",
	}, []string{"state"})

	sqliteVersionLogOnce sync.Once

	sqliteMigrations = migrator.Migrations(
		execsql(
			"create_customers",
			`create table if not exists customers(customer_id primary key, first_name, middle_name, last_name, nick_name, suffix, birth_date datetime, status, email, created_at datetime, last_modified datetime, deleted_at datetime);`,
		),
		execsql(
			"create_customers_phones",
			`create table if not exists customers_phones(customer_id, number, valid, type, unique (customer_id, number) on conflict abort);`,
		),
		execsql(
			"create_customers_addresses",
			`create table if not exists customers_addresses(address_id primary key, customer_id, type, address1, address2, city, state, postal_code, country, validated, 
			deleted_at datetime, unique (customer_id, address1) on conflict abort);`,
		),
		execsql(
			"create_customer_metadata",
			`create table if not exists customer_metadata(customer_id, meta_key, meta_value, unique(meta_key, meta_value));`,
		),
		execsql(
			"customer_status_updates",
			`create table if not exists customer_status_updates(customer_id, future_status, comment, changed_at datetime);`,
		),
		execsql(
			"create_customer_ofac_searches",
			`create table if not exists customer_ofac_searches(customer_id, entity_id, sdn_name, sdn_type, percentage_match, created_at datetime);`,
		),
		execsql(
			"create_customer_ssn",
			`create table if not exists customer_ssn(customer_id primary key, ssn, ssn_masked, created_at datetime);`,
		),
		execsql(
			"create_documents",
			`create table if not exists documents(document_id primary key, customer_id, type, content_type, uploaded_at datetime, deleted_at datetime);`,
		),
		execsql(
			"create_disclaimers",
			`create table if not exists disclaimers(disclaimer_id primary key, text, document_id, created_at datetime, deleted_at datetime);`,
		),
		execsql(
			"create_disclaimer_acceptances",
			`create table if not exists disclaimer_acceptances(disclaimer_id, customer_id, accepted_at datetime, unique(disclaimer_id, customer_id) on conflict ignore);`,
		),
		execsql(
			"create_accounts",
			`create table if not exists accounts(account_id primary key, customer_id, user_id, encrypted_account_number, hashed_account_number, masked_account_number, routing_number, status, type, created_at datetime, deleted_at datetime);`,
		),
		execsql(
			"add_customer_type",
			`alter table customers add column type; update customers set type = 'individual' where type is null;`,
		),
		execsql(
			"unique_accounts_per_customer",
			`
PRAGMA foreign_keys=off;
BEGIN TRANSACTION;

ALTER TABLE accounts RENAME TO accounts_old;

CREATE TABLE accounts
(
  account_id primary key,
  customer_id,
  user_id,
  encrypted_account_number,
  hashed_account_number,
  masked_account_number,
  routing_number,
  status,
  type,
  created_at datetime,
  deleted_at datetime,
  CONSTRAINT accounts_unique_to_customer UNIQUE (customer_id, hashed_account_number, routing_number)
);

INSERT INTO accounts SELECT * FROM accounts_old;

COMMIT;

PRAGMA foreign_keys=on;`,
		),
		execsql(
			"add_holder_name_to_accounts",
			`alter table accounts add column holder_name default '';`,
		),
		execsql(
			"create_validations",
			`create table if not exists validations(
				validation_id primary key,
				account_id,
				status,
				strategy,
				vendor,
				created_at datetime,
				updated_at datetime
			);`,
		),
		execsql(
			"create validations index",
			`
			create unique index
				idx_validations_validation_account_ids
			on
				validations (validation_id, account_id);
			`,
		),
		execsql(
			"create_account_ofac_searches",
			`create table if not exists account_ofac_searches(account_ofac_search_id varchar(40) primary key, account_id varchar(40), entity_id varchar(40), sdn_name varchar(40), sdn_type integer, percentage_match double precision (5,2), created_at datetime);`,
		),
		execsql(
			"add_namespace__to__customers",
			"alter table customers add column namespace varchar(40) not null default 'unknown';",
		),
		execsql(
			"create_organization_configuration",
			`create table organization_configuration(organization varchar(40) primary key not null, legal_entity varchar(40) not null, primary_account varchar(40) not null);`,
		),
		execsql(
			"rename_customers_namespace_to_organization",
			`alter table customers rename column namespace to organization;`,
		),
	)
)

type sqlite struct {
	path string

	connections *kitprom.Gauge
	logger      log.Logger

	err error
}

func (s *sqlite) Connect() (*sql.DB, error) {
	if s.err != nil {
		return nil, fmt.Errorf("sqlite had error %v", s.err)
	}

	sqliteVersionLogOnce.Do(func() {
		if v, _, _ := sqlite3.Version(); v != "" {
			s.logger.Log("main", fmt.Sprintf("sqlite version %s", v))
		}
	})

	db, err := sql.Open("sqlite3", s.path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return db, err
	}

	migratorLogger := migrator.WithLogger(migrator.LoggerFunc(func(msg string, args ...interface{}) {
		s.logger.Log("sqlite", msg)
	}))

	// Migrate our database
	if m, err := migrator.New(migratorLogger, sqliteMigrations); err != nil {
		return db, err
	} else {
		if err := m.Migrate(db); err != nil {
			return db, err
		}
	}

	// Spin up metrics only after everything works
	go func() {
		t := time.NewTicker(1 * time.Second)
		for range t.C {
			stats := db.Stats()
			s.connections.With("state", "idle").Set(float64(stats.Idle))
			s.connections.With("state", "inuse").Set(float64(stats.InUse))
			s.connections.With("state", "open").Set(float64(stats.OpenConnections))
		}
	}()

	return db, err
}

func sqliteConnection(logger log.Logger, path string) *sqlite {
	return &sqlite{
		path:        path,
		logger:      logger,
		connections: sqliteConnections,
	}
}

func getSqlitePath() string {
	path := os.Getenv("SQLITE_DB_PATH")
	if path == "" || strings.Contains(path, "..") {
		// set default if empty or trying to escape
		// don't filepath.ABS to avoid full-fs reads
		path = "customers.db"
	}
	return path
}

// TestSQLiteDB is a wrapper around sql.DB for SQLite connections designed for tests to provide
// a clean database for each testcase.  Callers should cleanup with Close() when finished.
type TestSQLiteDB struct {
	DB *sql.DB

	dir string // temp dir created for sqlite files
}

func (r *TestSQLiteDB) Close() error {
	if err := r.DB.Close(); err != nil {
		return err
	}
	return os.RemoveAll(r.dir)
}

// CreateTestSqliteDB returns a TestSQLiteDB which can be used in tests
// as a clean sqlite database. All migrations are ran on the db before.
//
// Callers should call close on the returned *TestSQLiteDB.
func CreateTestSqliteDB(t *testing.T) *TestSQLiteDB {
	dir, err := ioutil.TempDir("", "customers-sqlite")
	if err != nil {
		t.Fatalf("sqlite test: %v", err)
	}

	db, err := sqliteConnection(log.NewNopLogger(), filepath.Join(dir, "customers.db")).Connect()
	if err != nil {
		t.Fatalf("sqlite test: %v", err)
	}
	return &TestSQLiteDB{db, dir}
}

// SqliteUniqueViolation returns true when the provided error matches the SQLite error
// for duplicate entries (violating a unique table constraint).
func SqliteUniqueViolation(err error) bool {
	match := strings.Contains(err.Error(), "UNIQUE constraint failed")
	if e, ok := err.(sqlite3.Error); ok {
		return match || e.Code == sqlite3.ErrConstraint
	}
	return match
}
