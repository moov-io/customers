moov-io/customers
===

[![GoDoc](https://godoc.org/github.com/moov-io/customers?status.svg)](https://godoc.org/github.com/moov-io/customers)
[![Build Status](https://travis-ci.com/moov-io/customers.svg?branch=master)](https://travis-ci.com/moov-io/customers)
[![Coverage Status](https://codecov.io/gh/moov-io/customers/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/customers)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/customers)](https://goreportcard.com/report/github.com/moov-io/customers)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/customers/master/LICENSE)

The Customers project focuses on solving authentic identification of humans who are legally able to hold and transfer currency within the US. Primarily this project solves [Know Your Customer](https://en.wikipedia.org/wiki/Know_your_customer) (KYC), [Customer Identification Program](https://en.wikipedia.org/wiki/Customer_Identification_Program) (CIP), [Office of Foreign Asset Control](https://www.treasury.gov/about/organizational-structure/offices/Pages/Office-of-Foreign-Assets-Control.aspx) (OFAC) checks and verification workflows to comply with US federal law and ensure authentic transfers. Also, Customers has an objective to be a service for detailed due diligence on individuals and companies for Financial Institutions and services in a modernized and extensible way.

Docs: [docs.moov.io](https://docs.moov.io/en/latest/) | [api docs](https://api.moov.io/apps/customers/)

## Project Status

Moov Customers is under active development, so please star the project if you are interested in its progress. We are developing an extensible HTTP API for interactions along with an OpenAPI specification file for generating clients for integration projects.

## Getting Started

TODO

### Configuration

| Environmental Variable | Description | Default |
|-----|-----|-----|
| TODO | Description | ` ` |

#### Document Storage

BUCKET_NAME
CLOUD_PROVIDER

AWS_REGION
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY

GCP service account
GOOGLE_APPLICATION_CREDENTIALS
https://cloud.google.com/docs/authentication/getting-started

FILEBLOB_BASE_URL
FILEBLOB_HMAC_SECRET

## Getting Help

 channel | info
 ------- | -------
 [Project Documentation](https://docs.moov.io/en/latest/) | Our project documentation available online.
 Google Group [moov-users](https://groups.google.com/forum/#!forum/moov-users)| The Moov users Google group is for contributors other people contributing to the Moov project. You can join them without a google account by sending an email to [moov-users+subscribe@googlegroups.com](mailto:moov-users+subscribe@googlegroups.com). After receiving the join-request message, you can simply reply to that to confirm the subscription.
Twitter [@moov_io](https://twitter.com/moov_io)	| You can follow Moov.IO's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io) | If you are able to reproduce an problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](http://moov-io.slack.com/) | Join our slack channel (`#customers`) to have an interactive discussion about the development of the project. [Request an invite to the slack channel](https://join.slack.com/t/moov-io/shared_invite/enQtNDE5NzIwNTYxODEwLTRkYTcyZDI5ZTlkZWRjMzlhMWVhMGZlOTZiOTk4MmM3MmRhZDY4OTJiMDVjOTE2MGEyNWYzYzY1MGMyMThiZjg)

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/ach/blob/master/CODE_OF_CONDUCT.md) to get started!

Note: This project uses Go Modules, which requires Go 1.11 or higher, but we ship the vendor directory in our repository.

## License

Apache License 2.0 See [LICENSE](LICENSE) for details.
