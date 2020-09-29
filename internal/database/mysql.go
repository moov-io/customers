// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/moov-io/base/docker"

	"github.com/go-kit/kit/log"
	kitprom "github.com/go-kit/kit/metrics/prometheus"
	gomysql "github.com/go-sql-driver/mysql"
	"github.com/lopezator/migrator"
	"github.com/ory/dockertest/v3"
	stdprom "github.com/prometheus/client_golang/prometheus"
)

var (
	mysqlConnections = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "mysql_connections",
		Help: "How many MySQL connections and what status they're in.",
	}, []string{"state"})

	// mySQLErrDuplicateKey is the error code for duplicate entries
	// https://dev.mysql.com/doc/refman/8.0/en/server-error-reference.html#error_er_dup_entry
	mySQLErrDuplicateKey uint16 = 1062

	// ToDo: Consideration: For _type _status integer (1,2,3,4,5,6...) or string ("01", "1", "ABC")

	mysqlMigrations = migrator.Migrations(
		execsql(
			"create_customers",
			`create table if not exists customers(customer_id varchar(40), first_name varchar(40), middle_name varchar(40), last_name varchar(40), nick_name varchar(40), suffix varchar(3), birth_date datetime, status integer, email varchar(120), created_at datetime, last_modified datetime, deleted_at datetime, PRIMARY KEY (customer_id));`),

		execsql(
			"create_customers_phones",
			`create table if not exists customers_phones (customer_id VARCHAR(40), number VARCHAR(20), valid BOOLEAN, type integer, unique (customer_id, number));`,
		),
		execsql(
			"create_customers_addresses",
			`create table if not exists customers_addresses(address_id varchar(40) primary key, customer_id varchar(40), type SMALLINT, address1 varchar(120), address2 varchar(120), city varchar(50), state varchar(2), postal_code varchar(9), country varchar(3), validated BOOLEAN, unique (customer_id, address1));`,
		),
		execsql(
			"create_customer_metadata",
			`create table if not exists customer_metadata(customer_id varchar(40), meta_key varchar(40), meta_value varchar(512), unique(meta_key, meta_value));`,
		),
		execsql(
			"customer_status_updates",
			`create table if not exists customer_status_updates(customer_id varchar(40), future_status integer, comment varchar(512), changed_at datetime);`,
		),
		execsql(
			"create_customer_ofac_searches",
			`create table if not exists customer_ofac_searches(customer_id varchar(40), entity_id varchar(40), sdn_name varchar(40), sdn_type integer, percentage_match double precision (5,2), created_at datetime);`,
		),
		execsql(
			"create_customer_ssn",
			`create table if not exists customer_ssn(customer_id varchar(40) primary key, ssn BLOB, ssn_masked varchar(9), created_at datetime);`,
		),
		execsql(
			"create_documents",
			`create table if not exists documents(document_id varchar(40) primary key, customer_id varchar(40), type varchar(120), content_type varchar(40), uploaded_at datetime);`,
		),
		execsql(
			"create_disclaimers",
			`create table if not exists disclaimers(disclaimer_id varchar(40) primary key, text text, document_id varchar(40), created_at datetime, deleted_at datetime);`,
		),
		execsql(
			"create_disclaimer_acceptances",
			`create table if not exists disclaimer_acceptances(disclaimer_id varchar(40), customer_id varchar(40), accepted_at datetime, unique(disclaimer_id, customer_id));`,
		),
		execsql(
			"create_accounts",
			`create table if not exists accounts(account_id varchar(40) primary key, customer_id varchar(40), user_id varchar(40), encrypted_account_number varchar(40), hashed_account_number varchar(40), masked_account_number varchar(15), routing_number varchar(10), status varchar(12), type varchar(12), created_at datetime, deleted_at datetime);`,
		),
		execsql(
			"alter_customers_status",
			`alter table customers modify status varchar(20);`,
		),
		execsql(
			"expand_accounts_encrypted_account_number",
			`alter table accounts modify encrypted_account_number varchar(100);`,
		),
		execsql(
			"add_customer_type",
			`alter table customers add column type varchar(25)`,
		),
		execsql(
			"update_customer_type",
			`update customers set type = 'individual' where type is null;`,
		),
		execsql(
			"unique_accounts_per_customer",
			`alter table accounts add constraint unique accounts_unique_to_customer (customer_id, hashed_account_number, routing_number);`,
		),
		execsql(
			"add_holder_name_to_accounts",
			`alter table accounts add column holder_name varchar(60) default '';`,
		),
		execsql(
			"create_validations",
			`create table if not exists validations(
				validation_id  varchar(40) primary key,
				account_id varchar(40),
				status varchar(20),
				strategy varchar(20),
				vendor varchar(20),
				created_at datetime,
				updated_at datetime,
				INDEX idx_validations_validation_account_ids (validation_id, account_id)
			);`,
		),
		execsql(
			"create_account_ofac_searches",
			`create table if not exists account_ofac_searches(account_ofac_search_id varchar(40) primary key, account_id varchar(40), entity_id varchar(40), sdn_name varchar(40), sdn_type integer, percentage_match double precision (5,2), created_at datetime);`,
		),
		execsql(
			"add_namespace__to__customers",
			"alter table customers add column namespace varchar(40) not null;",
		),
		execsql(
			"create_namespace_configuration",
			`create table if not exists namespace_configuration(namespace varchar(40) primary key not null, legal_entity varchar(40) not null, primary_account varchar(40) not null);`,
		),
		execsql(
			"add_deleted_at__to__documents",
			"alter table documents add column deleted_at datetime;",
		),
		execsql(
			"add_deleted_at__to__customers_addresses",
			`alter table customers_addresses add column deleted_at datetime;`,
		),
		execsql(
			"rename_customers_namespace_to_organization",
			`alter table customers rename column namespace to organization;`,
		),
		execsql(
			"rename_namespace_configuration__to__organization_configuration",
			`alter table namespace_configuration rename to organization_configuration;`,
		),
		execsql(
			"rename_organization_configuration__namespace_to_organization",
			`alter table organization_configuration rename column namespace to organization;`,
		),
	)
)

type discardLogger struct{}

func (l discardLogger) Print(v ...interface{}) {}

func init() {
	gomysql.SetLogger(discardLogger{})
}

type mysql struct {
	dsn    string
	logger log.Logger

	connections *kitprom.Gauge
}

func (my *mysql) Connect() (*sql.DB, error) {
	db, err := sql.Open("mysql", my.dsn)
	if err != nil {
		return nil, err
	}

	// Check out DB is up and working
	if err := db.Ping(); err != nil {
		return nil, err
	}

	migratorLogger := migrator.WithLogger(migrator.LoggerFunc(func(msg string, args ...interface{}) {
		my.logger.Log("mysql", msg)
	}))

	// Migrate our database
	if m, err := migrator.New(migratorLogger, mysqlMigrations); err != nil {
		return nil, err
	} else {
		if err := m.Migrate(db); err != nil {
			return nil, err
		}
	}

	// Setup metrics after the database is setup
	go func() {
		t := time.NewTicker(1 * time.Minute)
		for range t.C {
			stats := db.Stats()
			my.connections.With("state", "idle").Set(float64(stats.Idle))
			my.connections.With("state", "inuse").Set(float64(stats.InUse))
			my.connections.With("state", "open").Set(float64(stats.OpenConnections))
		}
	}()

	return db, nil
}

func mysqlConnection(logger log.Logger, user, pass string, address string, database string) *mysql {
	timeout := "30s"
	if v := os.Getenv("MYSQL_TIMEOUT"); v != "" {
		timeout = v
	}
	params := fmt.Sprintf("timeout=%s&charset=utf8mb4&parseTime=true&sql_mode=ALLOW_INVALID_DATES", timeout)
	dsn := fmt.Sprintf("%s:%s@%s/%s?%s", user, pass, address, database, params)
	return &mysql{
		dsn:         dsn,
		logger:      logger,
		connections: mysqlConnections,
	}
}

// TestMySQLDB is a wrapper around sql.DB for MySQL connections designed for tests to provide
// a clean database for each testcase.  Callers should cleanup with Close() when finished.
type TestMySQLDB struct {
	DB *sql.DB

	container *dockertest.Resource
}

func (r *TestMySQLDB) Close() error {
	r.container.Close()
	return r.DB.Close()
}

// CreateTestMySQLDB returns a TestMySQLDB which can be used in tests
// as a clean mysql database. All migrations are ran on the db before.
//
// Callers should call close on the returned *TestMySQLDB.
func CreateTestMySQLDB(t *testing.T) *TestMySQLDB {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}
	if !docker.Enabled() {
		t.Skip("Docker not enabled")
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "8",
		Env: []string{
			"MYSQL_USER=moov",
			"MYSQL_PASSWORD=secret",
			"MYSQL_ROOT_PASSWORD=secret",
			"MYSQL_DATABASE=paygate",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = pool.Retry(func() error {
		db, err := sql.Open("mysql", fmt.Sprintf("moov:secret@tcp(localhost:%s)/paygate", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping()
	})
	if err != nil {
		resource.Close()
		t.Fatal(err)
	}

	logger := log.NewNopLogger()
	address := fmt.Sprintf("tcp(localhost:%s)", resource.GetPort("3306/tcp"))

	db, err := mysqlConnection(logger, "moov", "secret", address, "paygate").Connect()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		pool.Purge(resource)
	})
	return &TestMySQLDB{db, resource}
}

// MySQLUniqueViolation returns true when the provided error matches the MySQL code
// for duplicate entries (violating a unique table constraint).
func MySQLUniqueViolation(err error) bool {
	match := strings.Contains(err.Error(), fmt.Sprintf("Error %d: Duplicate entry", mySQLErrDuplicateKey))
	if e, ok := err.(*gomysql.MySQLError); ok {
		return match || e.Number == mySQLErrDuplicateKey
	}
	return match
}
