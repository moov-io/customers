# Customers
**Home** | **[Configuration](configuration.md)** | **[Running](running.md)** | **[Client](https://github.com/moov-io/customers/blob/master/pkg/client/README.md)**

---

## Purpose
This project focuses on verifying the identity of people who are legally able to hold and transfer currency in the United States. It provides services related to:
 - [Know Your Customer](https://en.wikipedia.org/wiki/Know_your_customer) (KYC)
 - [Customer Identification Program](https://en.wikipedia.org/wiki/Customer_Identification_Program) (CIP)
 - [Office of Foreign Asset Control](https://www.treasury.gov/about/organizational-structure/offices/Pages/Office-of-Foreign-Assets-Control.aspx) (OFAC) checks
 - Verification workflows to comply with US federal law and ensure authentic transfers

The goal of this project is to provide objective, detailed due diligence on individuals and companies in the financial sector  — in a modernized and extensible way.

**Dependencies**
1. [Fed](./fed.md)
1. [PayGate](./paygate.md)
1. [Watchman](./watchman.md)

<!--
**Extending Customers**

1. [Local Development](./local-dev.md)
1. [High Availability](./ha.md)
-->

## Models
This project contains several models used in the HTTP endpoints. These are generated from the OpenAPI specification in the [/pkg/client](./pkg/client/) Go package. The primary models used in this project are:

### Customer
`Customer` represents an individual or business (sole proprietorship or corporation).
For creating a `Customer`, see the [API documentation](https://moov-io.github.io/customers/api/#post-/customers).

#### Customer Status and Approval

Approval is represented by the `status` field of a `Customer` and can have the following values: `Deceased`, `Rejected`, `ReceiveOnly`, `Verified`, `Frozen`, `Unknown` (default)
Approvals can only be done manually, but we are aiming for automated approval. In order for a `Customer` to be approved into:
 - `ReceiveOnly` requires an [OFAC search](https://github.com/moov-io/watchman) that results in a value below the specified threshold.
    - This status is used to receive funds.
 - `Verified` requires a valid Social Security Number (SSN) and an OFAC check.
    - This status is used to receive or send funds.

### Account
`Account` represents a demand-deposit account at a financial institution. The account number is encrypted.
For creating an `Account`, see the [API documentation](https://moov-io.github.io/customers/api/#post-/customers/{customerID}/accounts).

#### Account Validation
In order to use an account for ACH transactions, it will need to be validated. This ensures access and authorization to the financial instrument. This project supports the following strategies that can be used for account validation:

* `micro-deposits` - Two deposits of less than $0.25 (and an optional withdraw) transferred to the customer's bank account
* `instant` - Vendors like Plaid and MX provide the ability to verify a customer's bank account instantly using their online banking credentials

See more information on [how account validation strategies work](./account-validation.md).

### Document
`Document` represents a customer's document uploaded to persistent storage. All documents are encrypted.
For uploading a `Document`, see the [API documentation](https://moov-io.github.io/customers/api/#post-/customers/{customerID}/documents).


## Database Migrations

Migrations allow the application's database schema to evolve over time.  When an application starts, it automatically checks for database migrations and runs them if needed to keep the database schema up to date. Information about the current schema version (the latest applied migration) is stored in the `schema_migrations` table.

### Creating a Migration

Migrations are stored as files in the [/migrations](./migrations) directory. Contents of each file are executed by the database driver. The migration files should consist of valid SQL queries. The file names must adhere to the format: `{version}_{title}.up.sql`

- `verision` of the migration should be represented as an integer with 3 digits (with leading zeros: e.g., 007). The migrations are applied in ascending order based on the version numbers. You can find examples of different migrations in [./migrations](./migrations).
- `title` should describe what the migration is doing, e.g., `create_accounts_table`, `add_name_to_accounts`.

### Embedding Migrations

We use [pkger](https://github.com/markbates/pkger) to embed migration files into our application. Please [install it](https://github.com/markbates/pkger#installation) before you proceed.

Running `make embed-migrations` will generate a `cmd/server/pkged.go` file with the encoded contents from the `/migrations` directory which will be included into application build. Make sure to commit the generated file to the git repository.

## Getting Help

 channel | info
 ------- | -------
[Documentation](https://moov-io.github.io/customers) | Project documentation for our community.
[GitHub Issues](https://github.com/moov-io/customers/issues) | Public tracker of issues with our community. Please open a GitHub Issue if you're able to reproduce problems or to request features.
Twitter [@moov](https://twitter.com/moov)	| You can follow Moov's Twitter feed to get updates on our projects. You can also tweet us to ask questions or share comments.
Slack [#moov-io](https://slack.moov.io/) | Join the slack channel to discuss with other contributors about the development of Moov's open source projects.

---
**[Next - Configuration](configuration.md)**
