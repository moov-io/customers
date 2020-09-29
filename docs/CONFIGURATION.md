# Customers
**[Purpose](README.md)** | **Configuration** | **[Running](RUNNING.md)** | **[Client](../pkg/client/README.md)**

---

## Configuration

The following environmental variables can be set to configure behavior in Accounts.

| Environmental Variable | Description | Default |
|-----|-----|-----|
| `HTTPS_CERT_FILE` | Filepath containing a certificate (or intermediate chain) to be served by the HTTP server. Requires all traffic be over secure HTTP. | Empty |
| `HTTPS_KEY_FILE`  | Filepath of a private key matching the leaf certificate from `HTTPS_CERT_FILE`. | Empty |
| `DATABASE_TYPE` | Which database option to use (Options: `sqlite`, `mysql`) | Default: `sqlite` |

#### Fed

The Moov [Fed](https://github.com/moov-io/fed) service is used for routing number lookup and verification.

| Environmental Variable | Description | Default |
|-----|-----|-----|
| `FED_ENDPOINT` | HTTP address for Moov Fed interaction to lookup ABA routing numbers. | `http://fed.apps.svc.cluster.local:8080` |
| `FED_DEBUG_CALLS` | Print debugging information with all Fed API calls. | `false` |

#### PayGate

The Moov [PayGate](https://github.com/moov-io/paygate) service is used to initiate micro-deposits for account validation.

| Environmental Variable | Description | Default |
|-----|-----|-----|
| `PAYGATE_ENDPOINT` | HTTP address for Moov PayGate interactions. | `http://paygate.apps.svc.cluster.local:8080` |
| `PAYGATE_DEBUG_CALLS` | Print debugging information with all PayGate API calls. | `false` |

#### Watchman

The Moov [Watchman](https://github.com/moov-io/watchman) service is used for OFAC and other sanctions list searching and compliance.

| Environmental Variable | Description | Default |
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

- `DOCUMENTS_BUCKET`: The name of the bucket to use. Must be created outside of Customers if using a cloud provider. Make sure proper access and encryption controls are setup on this bucket to prevent exposure or unauthorized access. Example: `./storage/` (For `file` type backends or `moov-customers-storage` for GCP/GCS)
- `DOCUMENTS_PROVIDER`: Provider name which determines which of the following environmental variables are used to initialize Customer's persistence.

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
   - Generate this key by running `./cmd/genkey/` and copying the `base64key://...` value

##### Google Cloud Storage

- `SECRETS_GCP_KEY_RESOURCE_ID`: A Google Cloud resource ID used to interact with their Key Management Service (KMS). This value has the form `projects/MYPROJECT/locations/MYLOCATION/keyRings/MYKEYRING/cryptoKeys/MYKEY` and [their documentation has more details](https://cloud.google.com/kms/docs/object-hierarchy#key).

##### Vault storage

- `VAULT_SERVER_TOKEN`: A Vault generated value used to authenticate. See [the Hashicorp Vault documentation](https://www.vaultproject.io/docs/concepts/tokens.html) for more details.
- `VAULT_SERVER_URL`: A URL for accessing the vault instance. In production environments this should be an HTTPS (TLS) secured connection.

---
**[Next - Running](RUNNING.md)**
