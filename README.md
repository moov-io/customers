moov-io/customers
===

[![GoDoc](https://godoc.org/github.com/moov-io/customers?status.svg)](https://godoc.org/github.com/moov-io/customers)
[![Build Status](https://travis-ci.com/moov-io/customers.svg?branch=master)](https://travis-ci.com/moov-io/customers)
[![Coverage Status](https://codecov.io/gh/moov-io/customers/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/customers)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/customers)](https://goreportcard.com/report/github.com/moov-io/customers)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/customers/master/LICENSE)

The Customers project focuses on solving authentic identification of humans who are legally able to hold and transfer currency within the US. Primarily this project solves [Know Your Customer](https://en.wikipedia.org/wiki/Know_your_customer) (KYC), [Customer Identification Program](https://en.wikipedia.org/wiki/Customer_Identification_Program) (CIP), [Office of Foreign Asset Control](https://www.treasury.gov/about/organizational-structure/offices/Pages/Office-of-Foreign-Assets-Control.aspx) (OFAC) checks and verification workflows to comply with US federal law and ensure authentic transfers. Also, Customers has an objective to be a service for detailed due diligence on individuals and companies for Financial Institutions and services in a modernized and extensible way.

Docs: [Project](https://github.com/moov-io/customers/tree/master/docs/) | [API Endpoints](https://moov-io.github.io/customers/) | [Admin API Endpoints](https://moov-io.github.io/customers/admin/)

## Project Status

Moov Customers is under active development, so please star the project if you are interested in its progress. We are developing an extensible HTTP API for interactions along with an OpenAPI specification file for generating clients for integration projects.

### Running Locally

Customers has a [Docker Compose](https://docs.docker.com/compose/gettingstarted/) setup which you can run locally. This uses the latest releases of Customers and Watchman.

```
$ docker-compose up
Creating customers_watchman_1 ... done
Creating customers_customers_1 ... done
...
customers_1  | ts=2020-03-06T22:56:24.2184402Z caller=main.go:50 startup="Starting moov-io/customers server version v0.4.0-rc1"
customers_1  | ts=2020-03-06T22:56:24.393462Z caller=watchman.go:102 watchman="using http://watchman:8084 for Watchman address"
customers_1  | ts=2020-03-06T22:56:24.3951132Z caller=main.go:171 startup="binding to :8087 for HTTP server"
```

Once the systems start you can access Customers via `http://localhost:8087` and Watchman's [web interface or api](http://localhost:8084).

### Deployment

You can download [our docker image `moov/customers`](https://hub.docker.com/r/moov/customers/) from Docker Hub or use this repository. No configuration is required to serve on `:8087` and metrics at `:9097/metrics` in Prometheus format.

### Configuration

The following environmental variables can be set to configure behavior in Accounts.

| Environmental Variable | Description | Default |
|-----|-----|-----|
| `HTTPS_CERT_FILE` | Filepath containing a certificate (or intermediate chain) to be served by the HTTP server. Requires all traffic be over secure HTTP. | Empty |
| `HTTPS_KEY_FILE`  | Filepath of a private key matching the leaf certificate from `HTTPS_CERT_FILE`. | Empty |
| `OFAC_ENDPOINT` | HTTP address for [OFAC](https://github.com/moov-io/ofac) interaction, defaults to Kubernetes inside clusters and local dev otherwise. | Kubernetes DNS |
| `OFAC_MATCH_THRESHOLD` | Percent match against OFAC data that's required for paygate to block a transaction. | `99%` |
| `DATABASE_TYPE` | Which database option to use (Options: `sqlite`, `mysql`) | Default: `sqlite` |

#### Fed

The Moov [Fed](https://github.com/moov-io/fed) service is used for routing number lookup and verification.

| Environmental Variable | Description | Default |
|-----|-----|-----|
| `FED_ENDPOINT` | HTTP address for Moov Fed interaction to lookup ABA routing numbers. | `http://fed.apps.svc.cluster.local:8080` |

#### Account Numbers

Customers has an endpoint which encrypts an account number for transit to another service. This encryption is currently done with a symmetric key to the other service.

- `TRANSIT_LOCAL_BASE64_KEY`: A URI used to encrypt account numbers for transit. This value needs to look like `base64key://value` where `value` is a base64 encoded 32 byte random key.

#### Storage

Based on `DATABASE_TYPE` the following environment variables will be read to configure connections for a specific database.

##### MySQL

- `MYSQL_ADDRESS`: TCP address for connecting to the mysql server. (example: `tcp(hostname:3306)`)
- `MYSQL_DATABASE`: Name of database to connect into.
- `MYSQL_PASSWORD`: Password of user account for authentication.
- `MYSQL_USER`: Username used for authentication,

Refer to the mysql driver documentation for [connection parameters](https://github.com/go-sql-driver/mysql#dsn-data-source-name).

- `MYSQL_TIMEOUT`: Timeout parameter specified on (DSN) data source name. (Default: `30s`)

##### SQLite

- `SQLITE_DB_PATH`: Local filepath location for the customers SQLite database. (Default: `customers.db`)

Refer to the sqlite driver documentation for [connection parameters](https://github.com/mattn/go-sqlite3#connection-string).

#### Document Storage

The following environment variables control which backend service is initialized for Document persistence. These all follow a similar ["blob storage"](https://gocloud.dev/ref/blob/) API provided by a library that Google [build and maintains](https://github.com/google/go-cloud).

- `BUCKET_NAME`: The name of the bucket to use. Must be created outside of Customers if using a cloud provider. Make sure proper access and encryption controls are setup on this bucket to prevent exposure or unauthorized access. Example: `./storage/` (For `file` type backends)
- `CLOUD_PROVIDER`: Provider name which determines which of the following environmental variables are used to initialize Customer's persistence.

##### AWS S3 Storage

For more information see the [Go Cloud Development Kit docs for s3blob](https://godoc.org/gocloud.dev/blob/s3blob). Use `CLOUD_PROVIDER=aws` to read the following environmental variables:

- `AWS_REGION`: Amazon region name of where the bucket exists.
- `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`: Standard AWS access credentials used by applications.

##### Google Cloud Storage

For more information see the [Go Cloud Development Kit docs for gcsblob](https://godoc.org/gocloud.dev/blob/gcsblob). Google's auth uses the standard [service account authorization](https://cloud.google.com/docs/authentication/getting-started) when deploying services. Use `CLOUD_PROVIDER=gcp` to read the following environmental variables:

- `GOOGLE_APPLICATION_CREDENTIALS`: A filepath to the GCP service account json file.

##### Local filesystem Storage

For more information see the [Go Cloud Development Kit docs for fileblob](https://godoc.org/gocloud.dev/blob/fileblob). This is the default if no other provider is specified. Use `CLOUD_PROVIDER=file` to read the following environmental variables:

- `FILEBLOB_BASE_URL`: A filepath for storage on local disk. (Default: `./storage/`)
- `FILEBLOB_HMAC_SECRET`: HMAC secret value used to sign URLs. You *MUST* change this for production usage! (Default: `secret`)

#### Social Security Number (SSN) Storage

- `CLOUD_PROVIDER`: Provider name which determines which of the following environmental variables are used to initialize Customer's persistence.

##### Local storage

- `SECRETS_LOCAL_BASE64_KEY`: A URI used to encrypt account numbers for storage in the database. This value needs to look like `base64key://value` where `value` is a base64 encoded 32 byte random key.

##### Google Cloud Storage

- `SECRETS_GCP_KEY_RESOURCE_ID`: A Google Cloud resource ID used to interact with their Key Management Service (KMS). This value has the form `projects/MYPROJECT/locations/MYLOCATION/keyRings/MYKEYRING/cryptoKeys/MYKEY` and [their documentation has more details](https://cloud.google.com/kms/docs/object-hierarchy#key).

##### Vault storage

- `VAULT_SERVER_TOKEN`: A Vault generated value used to authenticate. See [the Hashicorp Vault documentation](https://www.vaultproject.io/docs/concepts/tokens.html) for more details.
- `VAULT_SERVER_URL`: A URL for accessing the vault instance. In production environments this should be an HTTPS (TLS) secured connection.

## Customer Approval

Currently approval of Customers is represented by the [`status` field of a `Customer`](https://api.moov.io/#operation/getCustomer) and can have the following values: `Deceased`, `Rejected`, `None` (Default), `ReviewRequired`, `KYC`, `OFAC`, and `CIP`. These values can only be changed via the "admin" endpoints exposed in Customers. Admin endpoints are served from Customer's admin port (`9097`). Approvals (updates to a Customer status) can only be done manually, but we are aiming for automated approval. In order for a Customer to be approved into OFAC or higher there must be an [OFAC search](https://github.com/moov-io/ofac#moov-ioofac) performed without positive matches and CIP requires a valid Social Security Number (SSN).

## Getting Help

 channel | info
 ------- | -------
 [Project Documentation](https://github.com/moov-io/customers/tree/master/docs/) | Our project documentation available online.
 [Hosted Documentation](https://docs.moov.io/customers/) | Hosted documentation for enterprise solutions.
 Google Group [moov-users](https://groups.google.com/forum/#!forum/moov-users)| The Moov users Google group is for contributors other people contributing to the Moov project. You can join them without a google account by sending an email to [moov-users+subscribe@googlegroups.com](mailto:moov-users+subscribe@googlegroups.com). After receiving the join-request message, you can simply reply to that to confirm the subscription.
Twitter [@moov_io](https://twitter.com/moov_io)	| You can follow Moov.IO's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io/customers) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](https://slack.moov.io/) | Join our slack channel (`#customers`) to have an interactive discussion about the development of the project.

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/ach/blob/master/CODE_OF_CONDUCT.md) to get started!

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) and uses Go 1.14 or higher. See [Golang's install instructions](https://golang.org/doc/install) for help setting up Go. You can download the source code and we offer [tagged and released versions](https://github.com/moov-io/customers/releases/latest) as well. We highly recommend you use a tagged release for production.

## License

Apache License 2.0 See [LICENSE](LICENSE) for details.
