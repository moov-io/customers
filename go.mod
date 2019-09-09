module github.com/moov-io/customers

go 1.12

require (
	github.com/antihax/optional v0.0.0-20180407024304-ca021399b1a6
	github.com/aws/aws-sdk-go v1.23.18
	github.com/go-kit/kit v0.9.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/vault/api v1.2.2
	github.com/lopezator/migrator v0.2.0
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/moov-io/base v0.10.0
	github.com/moov-io/ofac v0.10.0
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/prometheus/client_golang v1.1.0
	gocloud.dev v0.17.0
	gocloud.dev/secrets/vault v0.15.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gotest.tools v2.3.0 // indirect
)

replace go4.org v0.0.0-20190430205326-94abd6928b1d => go4.org v0.0.0-20190313082347-94abd6928b1d
