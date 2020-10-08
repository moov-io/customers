# Customers
**[Purpose](README.md)** | **[Configuration](#configuration)** | **[Persistence](#persistent-storage)** | **[Running](RUNNING.md)** | **[Client](../pkg/client/README.md)**

---

## Configuration

The following environment variables can be set to configure behavior in Accounts.

| Environment Variable | Description | Default |
|-----|-----|-----|
| `HTTPS_CERT_FILE` | Filepath containing a certificate (or intermediate chain) to be served by the HTTP server. Requires all traffic be over secure HTTP. | Empty |
| `HTTPS_KEY_FILE`  | Filepath of a private key matching the leaf certificate from `HTTPS_CERT_FILE`. | Empty |
| `DATABASE_TYPE` | Which database option to use (Options: `sqlite`, `mysql`) | Default: `sqlite` |

#### Fed

The Moov [Fed](https://github.com/moov-io/fed) service is used for routing number lookup and verification.

| Environment Variable | Description | Default |
|-----|-----|-----|
| `FED_ENDPOINT` | HTTP address for Moov Fed interaction to lookup ABA routing numbers. | `http://fed.apps.svc.cluster.local:8080` |
| `FED_DEBUG_CALLS` | Print debugging information with all Fed API calls. | `false` |

#### PayGate

The Moov [PayGate](https://github.com/moov-io/paygate) service is used to initiate micro-deposits for account validation.

| Environment Variable | Description | Default |
|-----|-----|-----|
| `PAYGATE_ENDPOINT` | HTTP address for Moov PayGate interactions. | `http://paygate.apps.svc.cluster.local:8080` |
| `PAYGATE_DEBUG_CALLS` | Print debugging information with all PayGate API calls. | `false` |

#### Watchman

The Moov [Watchman](https://github.com/moov-io/watchman) service is used for OFAC and other sanctions list searching and compliance.

| Environment Variable | Description | Default |
|-----|-----|-----|
| `OFAC_MATCH_THRESHOLD` | Percent match against OFAC data that's required for paygate to block a transaction. | `99%` |
| `WATCHMAN_ENDPOINT` | HTTP address for [OFAC](https://github.com/moov-io/watchman) interaction, defaults to Kubernetes inside clusters and local dev otherwise. | Kubernetes DNS |
| `WATCHMAN_DEBUG_CALLS` | Print debugging information with all Watchman API calls. | `false` |

#### Account Numbers

Customers has an endpoint which encrypts an account number for transit to another service. This encryption is currently done with a symmetric key to the other service.

- `TRANSIT_LOCAL_BASE64_KEY`: A URI used to temporarily encrypt account numbers for transit over the network. This value needs to look like `base64key://value` where `value` is a base64 encoded 32 byte random key. Callers of endpoints that respond with encrypted values need this same key to decrypt.
  - Generate this key by running `./cmd/genkey/` and copying the `base64key://...` value

#### Account Verification

##### Plaid

Following parameters should be set via environment to configure the account validation strategy with Plaid:

* `PLAID_CLIENT_ID`: Client ID
* `PLAID_SECRET`: API secret (depends on the environent)
* `PLAID_ENVIRONMENT`: Plaid environment (e.g., sandbox, development or production | default `sandbox`)
* `PLAID_CLIENT_NAME`: The app name that should be displayed in Link

[Here](https://plaid.com/docs/#api-keys-and-access) you can find more information on how to get them.

##### MX

Following parameters should be set via environment to configure the account validation strategy with MX:

* `ATRIUM_CLIENT_ID`: Client ID
* `ATRIUM_API_KEY`: API Key

[Here](https://atrium.mx.com/docs#authentication-and-security) you can find more information on how to get them.


#### Database

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

#### Persistent Storage

The following environment variables control which backend service is initialized for persistent storage. These all follow a similar [blob storage](https://gocloud.dev/howto/blob/) API provided by a library that Google [built and maintains](https://github.com/google/go-cloud).

- `DOCUMENTS_STORAGE_PROVIDER`: Determines which backend service is used for document persistence. Defaults to [local filesystem storage](#local-filesystem-storage).
- `DOCUMENTS_BUCKET_NAME`: The name of the bucket to use for document storage (`/customers/{customerID}/documents/*` endpoints). Must be created outside of Customers if using a cloud provider. Make sure proper access and encryption controls are setup on this bucket to prevent exposure or unauthorized access. Example: `./storage/` (For `file` type backends or `moov-customers-storage` for GCP/GCS). Defaults to `./storage`.

##### AWS S3 Storage (`aws`)

For more information see the [Go Cloud Development Kit docs for s3blob](https://pkg.go.dev/gocloud.dev/blob/s3blob). The following environment variables are used to configure AWS S3 storage:

- `AWS_REGION`: Amazon region name of where the bucket exists.
- `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`: Standard AWS access credentials used by applications.

##### Google Cloud Storage (`gcp`)

For more information see the [Go Cloud Development Kit docs for gcsblob](https://pkg.go.dev/gocloud.dev/blob/gcsblob). Google's auth uses the standard [service account authorization](https://cloud.google.com/docs/authentication/getting-started) when deploying services. The following environment variables are used to configure GCP storage:

- `GOOGLE_APPLICATION_CREDENTIALS`: A filepath to the GCP service account json file.

##### Local Filesystem Storage (`file`)

For more information see the [Go Cloud Development Kit docs for fileblob](https://pkg.go.dev/gocloud.dev/blob/fileblob). This is the default if no other provider is specified. The following environment variables are used to configure local storage:

- `FILEBLOB_BASE_URL`: A filepath for storage on local disk. (Default: `./storage/`)
- `FILEBLOB_HMAC_SECRET`: HMAC secret value used to sign URLs. You *MUST* change this for production usage! (Default: `secret`)

#### Secrets (key management) Providers

The following environment variables control which backend service is utilized for secret key management. These all follow a similar [key management](https://gocloud.dev/howto/secrets/) API provided by a library that Google [built and maintains](https://github.com/google/go-cloud).

- `DOCUMENTS_SECRET_PROVIDER`: Determines which environment variables are used to initialize persistant document storage. Defaults to `local` (see [local filesystem](#local-filesystem)).
- `SSN_SECRET_PROVIDER`: Determines which environment variables are used to initialize SSN storage persistence. Defaults to `local` (see [local filesystem](#local-filesystem)).
  - `SSN_SECRET_KEY`: Holds the documents encryption/decryption key **if** the documents secret provider is `local`.

##### Local Filesystem (`local`)

The local secrets keeper (see [GoCloud Dev Kit - Secrets](https://gocloud.dev/howto/secrets/#local)) uses a 32 byte, base64 encoded encryption/decryption key. This value must be in the form `base64key://value` where `value` is the base64 encoded 32 byte random key.

This repository provides a script for generating properly formatted local keys (see ./cmd/genkey). New keys can be generated by running `go run ./cmd/genkey`

- `TRANSIT_LOCAL_BASE64_KEY`: The secret key to encrypt account numbers for storage in the database.
- `DOCUMENTS_SECRET_KEY`: The encryption/decryption key used for document storage and retrieval **if** the documents secret provider is `local`.
- `SSN_SECRET_KEY`: The encryption/decryption key used for customer SSN storage and retrieval **if** the SSN secret provider is `local`.

##### Google Cloud Storage (`gcp`)

This secrets provider uses the [Google Cloud KMS](https://cloud.google.com/kms/docs/object-hierarchy#key). Secret Keys are identified by a GCP Resource ID in the form `projects/project-id/locations/location/keyRings/keyring/cryptoKeys/key` and [their documentation has more details]().

- `SECRETS_GCP_KEY_RESOURCE_ID`: A Google Cloud resource ID used to interact with their Key Management Service (KMS). 

##### HashiCorp Vault Storage (`vault`)

- `VAULT_SERVER_TOKEN`: A Vault generated value used to authenticate. See [the HashiCorp Vault documentation](https://www.vaultproject.io/docs/concepts/tokens.html) for more details.
- `VAULT_SERVER_URL`: A URL for accessing the vault instance. In production environments this should be an HTTPS (TLS) secured connection.

---
**[Next - Running](RUNNING.md)**
