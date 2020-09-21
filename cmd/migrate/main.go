package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/moov-io/customers/internal/database"
)

var flagLogFormat = flag.String("log.format", "", "Format for log lines (Options: json, plain")

func main() {
	var logger log.Logger

	if strings.ToLower(*flagLogFormat) == "json" {
		logger = log.NewJSONLogger(os.Stderr)
	} else {
		logger = log.NewLogfmtLogger(os.Stderr)
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	// create database
	db, err := sql.Open("mysql", "root:secret@tcp(localhost:3306)/")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db_name := os.Getenv("MYSQL_DATABASE")

	_, err = db.Exec("DROP DATABASE IF EXISTS " + db_name)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE DATABASE " + db_name)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE USER IF NOT EXISTS 'moov'@'%' IDENTIFIED BY 'secret'")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("GRANT ALL PRIVILEGES ON * . * TO 'moov'@'%';")
	if err != nil {
		panic(err)
	}

	db, err = database.New(logger, "mysql")

	if err != nil {
		panic(err)
		// logger.Log("main", err)
		// os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Log("main", err)
		}
	}()

	fmt.Println("Done!")
}
