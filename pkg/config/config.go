package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/moov-io/base/database"
)

type Config struct {
	Database *database.DatabaseConfig
}

func New() *Config {
	return &Config{
		Database: &database.DatabaseConfig{},
	}
}

func (c *Config) Load() error {
	dbType := strings.ToLower(os.Getenv("DATABASE_TYPE"))
	switch dbType {
	case "sqlite", "":
		path := os.Getenv("SQLITE_DB_PATH")
		if path == "" || strings.Contains(path, "..") {
			// set default if empty or trying to escape
			// don't filepath.ABS to avoid full-fs reads
			path = "customers.db"
		}

		c.Database.SqlLite = &database.SqlLiteConfig{
			Path: path,
		}

	case "mysql":
		c.Database.MySql = &database.MySqlConfig{
			Address:  os.Getenv("MYSQL_ADDRESS"),
			User:     os.Getenv("MYSQL_USER"),
			Password: os.Getenv("MYSQL_PASSWORD"),
		}
		c.Database.DatabaseName = os.Getenv("MYSQL_DATABASE")

	default:
		return fmt.Errorf("unknown database type: %q", dbType)
	}

	return nil
}
