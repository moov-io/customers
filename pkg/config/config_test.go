package config

import (
	"os"
	"testing"

	"github.com/moov-io/base/database"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("When DATABASE_TYPE is not set", func(t *testing.T) {
		setenv(t, "DATABASE_TYPE", "")

		conf := New()
		err := conf.Load()

		require.NoError(t, err)

		require.Equal(t, conf.Database.SqlLite.Path, "customers.db")
	})

	t.Run("When DATABASE_TYPE is set to mysql", func(t *testing.T) {
		setenv(t, "DATABASE_TYPE", "mysql")
		setenv(t, "MYSQL_USER", "user")
		setenv(t, "MYSQL_PASSWORD", "password")
		setenv(t, "MYSQL_ADDRESS", "tcp(localhost:1234)")
		setenv(t, "MYSQL_DATABASE", "test")

		conf := New()
		err := conf.Load()

		require.NoError(t, err)
		require.Equal(t, "test", conf.Database.DatabaseName)
		want := &database.MySqlConfig{
			Address:  "tcp(localhost:1234)",
			User:     "user",
			Password: "password",
		}
		require.Equal(t, want, conf.Database.MySql)
	})
}

// setenv restores env variables after test
func setenv(t *testing.T, key, val string) {
	if oldVal, found := os.LookupEnv(key); found {
		t.Cleanup(func() {
			os.Setenv(key, oldVal)
		})
	} else {
		t.Cleanup(func() {
			os.Unsetenv(key)
		})
	}

	os.Setenv(key, val)
}
