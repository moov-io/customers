## v0.5.0 (Unreleased)

Customers v0.5.0 comes with several new features including Customer searching/filtering, additional Instant Account Validation strategies, and a reformed model for separating models.

**BREAKING CHANGES**

We now require the `X-Organization` HTTP header (can be changed with `ORGANIZATION_HEADER`) on requests. This is to enforce isolation of models for multi-tenant installs. The value can be a free-form string so a UUID, random string, or other identifier can be supplied.

- Accounts require HolderName (legal name on the financial account)
- ./pkg/client and ./pkg/admin package move

ADDITIONS

- accounts: `instant` validation can be performed with Plaid or MX.
- api,client: expose institution details on accounts
- accounts: return InstitutionDetails next to routing number
- cmd/server: add an endpoint for searching customers
- customers: added search filter params `status`, `type` and pagination with `skip` and `count`
- accounts: perform OFAC search of HolderName

IMPROVEMENTS

- accounts: return updated model after updating status
- customers: accept yyyy-mm-dd formatted birthDates
- customers: using base http GetSkipAndCount() to get skip and count from request
- customers: perform the OFAC search inline of creation flow
- accounts: reject duplicate accounts for a customer
- cmd/server: read OFAC_ENDPOINT or WATCHMAN_ENDPOINT
- customers: send back an array of search results, not null
- database/mysql: fix migration for customer type
- docs: reference ./cmd/genkey/ for SECRETS_LOCAL_BASE64_KEY
- fed: support debugging API calls
- paygate: support debugging API calls
- watchman: support debugging API calls
- api,client: mark SSN as optional on CreateCustomer

BUG FIXEs

- search: always return an allocated array for JSON marshal
- accounts: return an empty array if no accounts are found
- cmd/server: return 404 if customer isn't found
- customers: render nil birthDate as null instead of time.Zero
- customers: save customer type and marshal it back
- customers: validate customer type on creation request
- database/mysql: change customers.status to string
- database/mysql: expand encrypted_account_number

BUILD

- build: upgrade github.com/moov-io/watchman to v0.15.0
- build: mount sqlite path as volume
- chore(deps): update module aws/aws-sdk-go to v1.34.9
- chore(deps): update golang docker tag to v1.15

## v0.4.1 (Released 2020-07-16)

IMPROVEMENTS

- cmd/keygen: add script for generating high-quality gocloud.dev key URIs
- docs: clarify TRANSIT_LOCAL_BASE64_KEY is used for temporary encryption

BUG FIXES

- cmd/server: fix create customer query for mysql

## v0.4.0 (Released 2020-07-09)

ADDITIONS

- build: add OpenShift docker image
- customers: validate state abbreviations
- accounts: add endpoints (from PayGate) with encrypted account numbers
- accounts: include endpoint for transit encryption of an account number
- accounts: add endpoint for updating status
- accounts: validate micro-deposits with paygate HTTP calls

IMPROVEMENTS

- accounts: micro-deposits weren't found if there's no MicroDepositID
- api: use shared models from other OpenAPI specifications
- api,client: use short api summaries
- cmd/server: upgrade Watchman to v0.14.0 (was called OFAC)
- cmd/server: lookup individual and entiy SDNs from Watchman
- cmd/server: add version handler to admin HTTP server
- pkg/secrets: move to /pkg/ for external usage
- secrets/mask: leave last 4 digits

BUG FIXES

- pkg/secrets: read "base64key://" keys in local keeper

BUILD

- build: upgrade github.com/moov-io/paygate to v0.8.0
- build: switch to github Actions instead of TravisCI
- build: update Copyright headers for 2020
- build: run sonatype-nexus-community/nancy in CI
- build: test docker-compose setup in CI
- build: run infra Go lint script
- build: run CI in Windows
- chore(deps): update module aws/aws-sdk-go to v1.31.0
- disclaimers: remove `omitempty` from text field on admin create body

## v0.3.0 (Released 2019-11-12)

ADDITIONS

- cmd/server: add routes to create and accept Disclaimers
- cmd/server: add endpoint for manual OFAC refresh
- cmd/server: add endpoint for getting the latest OFAC search result

IMPROVEMENTS

- ofac: bump minimum threshold to 99% matches
- cmd/server: allow email, phones, and addresses to be optional on a Customer

BUILD

- build: download CI tools rather than install
- build: upgrade openapi-generator to 4.2.0

## v0.2.0 (Released 2019-08-20)

BREAKING CHANGE

In our OpenAPI we've renamed fields generated as `Id` to `ID`, which is more in-line with Go's style conventions.

ADDITIONS

- cmd/server: bind HTTP server with TLS if HTTPS_* variables are defined

BUILD

- docs: update docs.moov.io links after design refresh
- build: push moov/customers:latest on 'make release-push'
- build: upgrade openapi-generator to 4.1.0
- cmd/server: upgrade github.com/moov-io/base to v0.10.0

## v0.1.1 (Released 2019-06-19)

BUG FIXES

- Only read `VAULT_SERVER_TOKEN` not `VAULT_TOKEN`.

## v0.1.0 (Released 2019-06-19)

ADDITIONS

- cmd/server: initial storage and HTTP routes for documents
- cmd/server: initial retrieval and proxy of uploaded documents
- cmd/server: support an arbitrary map[string]string on customers
- cmd/server: whitelist only certain CustomerStatus transistions
- cmd/server: ensure a Customer address is validated and Primary
- cmd/server: add routes for adding and approving an Address, updating Customer status
- cmd/server: search Customers in OFAC when they're created
- cmd/server: lookup the latest OFAC search result for Customers on status transistion
- cmd/server: add persistence for storing encrypted SSN's
- cmd/server: save SSN when creating a Customer

BUG FIXES

- cmd/server: include Customer metadata in getCustomer repository method

## v0.0.0 (Released 2019-05-16)

- Initial release
