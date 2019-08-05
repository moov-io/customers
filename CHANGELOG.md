## v0.2.0 (Unreleased)

ADDITIONS

- cmd/server: bind HTTP server with TLS if HTTPS_* variables are defined

BUILD

- docs: update docs.moov.io links after design refresh
- build: push moov/customers:latest on 'make release-push'

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
