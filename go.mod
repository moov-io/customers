module github.com/moov-io/customers

go 1.12

require (
	github.com/antihax/optional v1.0.0
	github.com/aws/aws-sdk-go v1.29.1
	github.com/go-kit/kit v0.10.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/vault/api v1.0.2
	github.com/lopezator/migrator v0.2.0
	github.com/mattn/go-sqlite3 v1.13.0
	github.com/moov-io/ach v1.3.1
	github.com/moov-io/base v0.11.0
	github.com/moov-io/fed v0.5.0
	github.com/moov-io/watchman v0.13.2
	github.com/ory/dockertest/v3 v3.5.4
	github.com/prometheus/client_golang v1.5.1
	gocloud.dev v0.19.0
	gocloud.dev/secrets/hashivault v0.19.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
)

replace go4.org v0.0.0-20190430205326-94abd6928b1d => go4.org v0.0.0-20190313082347-94abd6928b1d
