package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/kelseyhightower/envconfig"
	"github.com/moov-io/customers/internal/database"
)

var flagLogFormat = flag.String("log.format", "", "Format for log lines (Options: json, plain")

type Config struct {
	RootPassword string `split_words:"true" default:"secret"`
	User         string `default:"moov"`
	Password     string `default:"secret"`
	Address      string `default:"tcp(localhost:3306)"`
	Database     string `default:"paygate_test"`
}

func main() {
	var config Config
	err := envconfig.Process("mysql", &config)
	if err != nil {
		panic(err)
	}

	err = runCmd(os.Args[1], &config)
	if err != nil {
		panic(err)
	}
}

func runCmd(cmd string, config *Config) error {
	switch cmd {
	case "setup":
		err := dropDB(config)
		if err != nil {
			return err
		}

		err = createDB(config)
		if err != nil {
			return err
		}

		err = migrateDB(config)
		if err != nil {
			return err
		}
	case "create":
		err := createDB(config)
		if err != nil {
			return err
		}
	case "drop":
		err := dropDB(config)
		if err != nil {
			return err
		}
	case "migrate":
		err := migrateDB(config)
		if err != nil {
			return err
		}
	}

	return nil
}

func dropDB(config *Config) error {
	dsn := fmt.Sprintf("root:%s@%s/", config.RootPassword, config.Address)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DROP DATABASE IF EXISTS " + config.Database)
	if err != nil {
		return err
	}

	fmt.Printf("Database %s was deleted\n", config.Database)
	return nil
}

func createDB(config *Config) error {
	dsn := fmt.Sprintf("root:%s@%s/", config.RootPassword, config.Address)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE " + config.Database)
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", config.User, config.Password))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s . * TO '%s'@'%%';", config.Database, config.User))
	if err != nil {
		return err
	}

	fmt.Printf("Database %s was created\n", config.Database)
	return nil
}

// TODO: having migration inside database.New (inside Connect method) makes it
// ambiguous we should extract migtation method into separate public method
// that we can call from here.
func migrateDB(config *Config) error {
	var logger log.Logger

	// migrate database
	if strings.ToLower(*flagLogFormat) == "json" {
		logger = log.NewJSONLogger(os.Stderr)
	} else {
		logger = log.NewLogfmtLogger(os.Stderr)
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	dbConf := &database.MySQLConfig{
		User:     config.User,
		Password: config.Password,
		Address:  config.Address,
		Database: config.Database,
	}

	db, err := database.NewMySQL(logger, dbConf)
	if err != nil {
		return err
	}
	db.Close()

	fmt.Printf("Database %s was migrated\n", config.Database)

	return nil
}
