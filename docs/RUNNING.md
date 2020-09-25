# Customers
**[Purpose](README.md)** | **[Configuration](CONFIGURATION.md)** | **Running** | **[Client](../pkg/client/README.md)**

--- 

## Running

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


### Customer Approval

Currently approval of Customers is represented by the [`status` field of a `Customer`](https://api.moov.io/#operation/getCustomer) and can have the following values: `Deceased`, `Rejected`, `Unknown`, (Default) `ReceiveOnly`, `Verified`. These values can only be changed via the "admin" endpoints exposed in Customers. Admin endpoints are served from Customer's admin port (`9097`). Approvals (updates to a Customer status) can only be done manually, but we are aiming for automated approval. In order for a Customer to be approved into `ReceiveOnly` there needs to be an [OFAC search](https://github.com/moov-io/watchman) performed without positive matches and or `Verified`  requires a valid Social Security Number (SSN) in addition to an OFAC search.

### Deployment

You can download [our docker image `moov/customers`](https://hub.docker.com/r/moov/customers/) from Docker Hub or use this repository. No configuration is required to serve on `:8087` and metrics at `:9097/metrics` in Prometheus format. We also have docker images for [OpenShift](https://quay.io/repository/moov/customers?tab=tags).

---
**[Next - Client](../pkg/client/README.md)**
