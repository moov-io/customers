# Customers
**[Home](README.md)** | **[Configuration](configuration.md)** | **Running** | **[Client](https://github.com/moov-io/customers/blob/master/pkg/client/README.md)**

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

Once the systems start you can access the following endpoints on `localhost`:
1. [Customer's API Endpoints](https://moov-io.github.io/customers/api/) (Port: `8097`)
1. [Customer's Admin Endpoints](https://moov-io.github.io/customers/admin/) (Port: `9097`)
1. [Watchman's API Endpoints](https://moov-io.github.io/watchman/api) (Port: `8084`)


### Deployment

You can download [our docker image `moov/customers`](https://hub.docker.com/r/moov/customers/) from Docker Hub or use this repository. No configuration is required to serve on `:8087` and metrics at `:9097/metrics` in Prometheus format. We also have docker images for [OpenShift](https://quay.io/repository/moov/customers?tab=tags).

---
**[Next - Client](https://github.com/moov-io/customers/blob/master/pkg/client/README.md)**
