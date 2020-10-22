# Customers
**Purpose** | **[Configuration](CONFIGURATION.md)** | **[Running](RUNNING.md)** | **[Client](../pkg/client/README.md)**

---

## Purpose

The Customers project focuses on solving authentic identification of humans who are legally able to hold and transfer currency within the US. Primarily this project solves [Know Your Customer](https://en.wikipedia.org/wiki/Know_your_customer) (KYC), [Customer Identification Program](https://en.wikipedia.org/wiki/Customer_Identification_Program) (CIP), [Office of Foreign Asset Control](https://www.treasury.gov/about/organizational-structure/offices/Pages/Office-of-Foreign-Assets-Control.aspx) (OFAC) checks and verification workflows to comply with US federal law and ensure authentic transfers. Also, Customers has an objective to be a service for detailed due diligence on individuals and companies for Financial Institutions and services in a modernized and extensible way.

**Dependencies**

1. [Fed](./fed.md)
1. [PayGate](./paygate.md)
1. [Waatchman](./waatchman.md)

<!--
**Extending Customers**

1. [Local Development](./local-dev.md)
1. [High Availability](./ha.md)
-->

## Models

Moov Customers has several models which are used throughout the HTTP endpoints. These are generated from the OpenAPI specification in the `github.com/moov-io/customers/pkg/client` Go package.

### Customer

`Customer` represents an individual or business (Sole-Proprietorships or Corporation). The data required for each will change with version v0.5.0 of Customers. See the [API documentation](https://moov-io.github.io/customers/api/#post-/customers) for creating a Customer.

### Account

`Account` represents a demand-deposit account at a financial institution. The account number is encrypted. See the [API documentation](https://moov-io.github.io/customers/api/#post-/customers/{customerID}/accounts) for creating an Account.

#### Account Validation

In order to use an account for ACH transactions, it will need to be validated. This is to ensure access and authorization to the financial instrument. Customers supports following strategies that can be used for account validation:

* micro-deposits - two deposits of less than $0.50 (and an optional withdraw) are transferred to customer's bank account and then customer providing deposits amounts as verification
* instant - some vendors like Plaid, MX, Yodelee provide the ability to verify customer's bank account instantly using their online banking credentials

See more information on [how account validation strategies work](./account-validation.md).

## Database Migrations

Migrations allow us to evolve application database schema over time.  When
appication starts it automatically checks database migrations and run them if
needed to keep the database schema up to date. Information about the current
schema version (the version of latest applied migration) is stored in the
`schema_migrations` table.

### Creating a Migration

Migrations are stored as files in the [/migrations](./migrations) directory.
Content of each file is passed to a database driver for execution. Migration
file should consist of valid SQL queries. 

Migration file name have to follow the format: `{version}_{title}.up.sql`

- `verision` of the migration should be represented as integer with 3 digits (with
leading zeros: e.g., 007). All migrations are applied upward in order of
increasing version number. You can find examples of different migrations in
[./migrations](./migrations).
- `title` should describe action of the migration, e.g.,
  `create_accounts_table`, `add_name_to_accounts`.


### Embedding Migrations

We use [pkger](https://github.com/markbates/pkger) to embed migration files
into our application. Please, [install
it](https://github.com/markbates/pkger#installation) before you proceed.

Running `make embed-migrations` will generate `cmd/server/pkged.go` file with
encoded content of `/migrations` directory that will be included into
application build. Please, commit generated file to the git repository.

## Getting Help

 channel | info
 ------- | -------
[Project Documentation](https://docs.moov.io/) | Our project documentation available online.
Twitter [@moov_io](https://twitter.com/moov_io)	| You can follow Moov.IO's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io/customers/issues) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](https://slack.moov.io/) | Join our slack channel to have an interactive discussion about the development of the project.

---
**[Next - Configuration](CONFIGURATION.md)**
