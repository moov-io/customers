package config

import (
	"os"
	"testing"

	"github.com/moov-io/base/database"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("When DATABASE_TYPE is not set", func(t *testing.T) {
		os.Setenv("DATABASE_TYPE", "")

		conf := New()
		conf.Load()

		require.Equal(t, conf.Database.SqlLite.Path, "customers.db")
	})

	t.Run("When DATABASE_TYPE is set to mysql", func(t *testing.T) {
		os.Setenv("DATABASE_TYPE", "mysql")
		os.Setenv("MYSQL_USER", "user")
		os.Setenv("MYSQL_PASSWORD", "password")
		os.Setenv("MYSQL_ADDRESS", "tcp(localhost:1234)")
		os.Setenv("MYSQL_DATABASE", "test")

		conf := New()
		conf.Load()

		require.Equal(t, "test", conf.Database.DatabaseName)
		want := &database.MySqlConfig{
			Address:  "tcp(localhost:1234)",
			User:     "user",
			Password: "password",
		}
		require.Equal(t, want, conf.Database.MySql)
	})
}
