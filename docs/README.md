## Customers

**Purpose** | **[Running](../README.md#running-locally)** | **[Configuration](./config.md)**

### Purpose

The Customers project focuses on solving authentic identification of humans who are legally able to hold and transfer currency within the US. Primarily this project solves [Know Your Customer](https://en.wikipedia.org/wiki/Know_your_customer) (KYC), [Customer Identification Program](https://en.wikipedia.org/wiki/Customer_Identification_Program) (CIP), [Office of Foreign Asset Control](https://www.treasury.gov/about/organizational-structure/offices/Pages/Office-of-Foreign-Assets-Control.aspx) (OFAC) checks and verification workflows to comply with US federal law and ensure authentic transfers. Also, Customers has an objective to be a service for detailed due diligence on individuals and companies for Financial Institutions and services in a modernized and extensible way.

### Table of Contents

**Running Customers**

1. [Configuration](../README.md#configuration)
1. [API Endpoints](https://moov-io.github.io/customers/api/)
1. [Admin Endpoints](https://moov-io.github.io/customers/admin/)

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

## Getting Help

 channel | info
 ------- | -------
 [Project Documentation](https://github.com/moov-io/customers/tree/master/docs/) | Our project documentation available online.
 [Hosted Documentation](https://docs.moov.io/customers/) | Hosted documentation for enterprise solutions.
 Google Group [moov-users](https://groups.google.com/forum/#!forum/moov-users)| The Moov users Google group is for contributors other people contributing to the Moov project. You can join them without a google account by sending an email to [moov-users+subscribe@googlegroups.com](mailto:moov-users+subscribe@googlegroups.com). After receiving the join-request message, you can simply reply to that to confirm the subscription.
Twitter [@moov_io](https://twitter.com/moov_io)	| You can follow Moov.IO's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io/customers) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](https://slack.moov.io/) | Join our slack channel (`#customers`) to have an interactive discussion about the development of the project.
