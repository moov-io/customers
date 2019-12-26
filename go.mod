module github.com/moov-io/customers

go 1.12

require (
	github.com/antihax/optional v1.0.0
	github.com/aws/aws-sdk-go v1.19.45
	github.com/go-kit/kit v0.9.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/vault/api v1.0.1
	github.com/lopezator/migrator v0.2.0
	github.com/mattn/go-sqlite3/v2/v2 v2.0.2
	github.com/moov-io/base v0.11.0-rc1.0.20191121181647-cd3e7a9609db
	github.com/moov-io/watchman v0.13.0
	github.com/ory/dockertest/v3 v3.5.2
	github.com/prometheus/client_golang v1.2.1
	gocloud.dev v0.17.0
	gocloud.dev/secrets/vault v0.15.0
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
)

replace go4.org v0.0.0-20190430205326-94abd6928b1d => go4.org v0.0.0-20190313082347-94abd6928b1d
