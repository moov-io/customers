FROM golang:1.16-buster as builder
WORKDIR /go/src/github.com/moov-io/customers
RUN apt-get update && apt-get install -y make gcc g++ time
COPY . .
RUN make build

FROM debian:10
LABEL maintainer="Moov <support@moov.io>"

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /go/src/github.com/moov-io/customers/bin/server /bin/server

VOLUME "/data"
ENV SQLITE_DB_PATH /data/customers.db

# USER moov
EXPOSE 8080
EXPOSE 9090
ENTRYPOINT ["/bin/server"]
