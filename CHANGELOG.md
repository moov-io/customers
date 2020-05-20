## v0.4.0 (Unreleased)

ADDITIONS

- customers: validate state abbreviations
- accounts: add endpoints (from PayGate) with encrypted account numbers
- accounts: include endpoint for transit encryption of an account number
- accounts: add endpoint for updating status
- accounts: validate micro-deposits with paygate HTTP calls

IMPROVEMENTS

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
