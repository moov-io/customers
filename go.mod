module github.com/moov-io/customers

go 1.12

require (
	github.com/antihax/optional v1.0.0
	github.com/aws/aws-sdk-go v1.33.6
	github.com/go-kit/kit v0.10.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/vault/api v1.0.4
	github.com/lopezator/migrator v0.3.0
	github.com/mattn/go-sqlite3 v2.0.6+incompatible
	github.com/moov-io/ach v1.4.1
	github.com/moov-io/base v0.11.0
	github.com/moov-io/fed v0.5.1
	github.com/moov-io/paygate v0.8.0
	github.com/moov-io/watchman v0.14.2
	github.com/ory/dockertest/v3 v3.6.0
	github.com/prometheus/client_golang v1.7.1
	github.com/rickar/cal v1.0.5 // indirect
	gocloud.dev v0.20.0
	gocloud.dev/secrets/hashivault v0.19.0
	golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace go4.org v0.0.0-20190430205326-94abd6928b1d => go4.org v0.0.0-20190313082347-94abd6928b1d
