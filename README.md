moov-io/customers
===

[![GoDoc](https://godoc.org/github.com/moov-io/customers?status.svg)](https://godoc.org/github.com/moov-io/customers)
[![Build Status](https://github.com/moov-io/customers/workflows/Go/badge.svg)](https://github.com/moov-io/customers/actions)
[![Coverage Status](https://codecov.io/gh/moov-io/customers/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/customers)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/customers)](https://goreportcard.com/report/github.com/moov-io/customers)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/customers/master/LICENSE)

This project focuses on verifying the identity of people who are legally able to hold and transfer currency in the United States. It provides services related to:
 - [Know Your Customer](https://en.wikipedia.org/wiki/Know_your_customer) (KYC)
 - [Customer Identification Program](https://en.wikipedia.org/wiki/Customer_Identification_Program) (CIP)
 - [Office of Foreign Asset Control](https://www.treasury.gov/about/organizational-structure/offices/Pages/Office-of-Foreign-Assets-Control.aspx) (OFAC) checks
 - Verification workflows to comply with US federal law and ensure authentic transfers

The goal of this project is to provide objective, detailed due diligence on individuals and companies in the financial sector — in a modernized and extensible way.

If you believe you have identified a security vulnerability please responsibly report the issue as via email to security@moov.io. Please do not post it to a public issue tracker.

[FFIEC Bank Secrecy Act - Customer Identification Program](https://www.fdic.gov/regulations/examinations/bsa/ffiec_cip.pdf)

Docs: [docs](https://moov-io.github.io/customers/) | [API Endpoints](https://moov-io.github.io/customers/api/) | [Admin API Endpoints](https://moov-io.github.io/customers/admin/)

## Project Status

Moov Customers has been deprecated in favor of a more direct service for uploading ACH files. We’ve also realized that KYC/KYB, data modeling, workflows, and risk analysis are difficult to open-source. We are interested in hearing how you have solved these problems internally and want to develop something in the open.

## Getting Started

Customers has a [Docker Compose](https://docs.docker.com/compose/gettingstarted/) setup which you can run locally. This uses the latest releases of Customers and Watchman.

```
$ docker-compose up
Creating customers_watchman_1 ... done
Creating customers_customers_1 ... done
...
customers_1  | ts=2020-03-06T22:56:24.2184402Z caller=main.go:50 startup="Starting moov-io/customers server version v0.4.0-rc1"
customers_1  | ts=2020-03-06T22:56:24.393462Z caller=watchman.go:102 watchman="using http://watchman:8084 for Watchman address"
customers_1  | ts=2020-03-06T22:56:24.3951132Z caller=main.go:171 startup="binding to :8087 for HTTP server"
```

Once the systems start you can access Customers via `http://localhost:8087` and Watchman's [web interface or api](http://localhost:8084):

1. [API Endpoints](https://moov-io.github.io/customers/api/)
1. [Admin Endpoints](https://moov-io.github.io/customers/admin/)

Read through the [project docs](docs/README.md) over here to get an understanding of the purpose of this project and how to run it.

## Getting Help

 channel | info
 ------- | -------
[Documentation](https://moov-io.github.io/customers) | Project documentation for our community.
[GitHub Issues](https://github.com/moov-io/customers/issues) | Public tracker of issues with our community. Please open a GitHub Issue if you're able to reproduce problems or to request features.
Twitter [@moov](https://twitter.com/moov)	| You can follow Moov's Twitter feed to get updates on our projects. You can also tweet us to ask questions or share comments.
Slack [#moov-io](https://slack.moov.io/) | Join the slack channel to discuss with other contributors about the development of Moov's open source projects.

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/ach/blob/master/CODE_OF_CONDUCT.md) to get started! Checkout our [issues for first time contributors](https://github.com/moov-io/customers/contribute) for something to help out with.

### Test Coverage

Improving test coverage is a good candidate for new contributors while also allowing the project to move more quickly by reducing regressions issues that might not be caught before a release is pushed out to our users. One great way to improve coverage is by adding edge cases and different inputs to functions.

## License

Apache License 2.0 See [LICENSE](LICENSE) for details.
