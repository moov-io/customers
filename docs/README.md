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

`Customer` represents an individual or business (Sole-Proprietorships or Corporation). The data required for each will change with version v0.5.0 of Customers. See the [API documentation](https://moov-io.github.io/customers/#post-/customers) for creating a Customer.

### Account

`Account` represents a demand-deposit account at a financial institution. The account number is encrypted. See the [API documentation](https://moov-io.github.io/customers/#post-/customers/{customerID}/accounts) for creating an Account.

#### Account Validation

In order to use an account for ACH transactions, it will need to be validated. This is to ensure access and authorization to the financial instrument. Customers supports following strategies that can be used for account validation:

* micro-deposits - two deposits of less than $0.50 (and an optional withdraw) are transferred to customer's bank account and then customer providing deposits amounts as verification
* instant - some vendors like Plaid, MX, Yodelee provide the ability to verify customer's bank account instantly using their online banking credentials

See more information on [how account validation strategies work](./account-validation.md).

---
**[Next - Configuration](CONFIGURATION.md)**
