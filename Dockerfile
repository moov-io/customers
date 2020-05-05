FROM golang:1.14-buster as builder
WORKDIR /go/src/github.com/moov-io/customers
RUN apt-get update && apt-get install -y make gcc g++ time
COPY . .
RUN go mod download
RUN make build

FROM debian:10
RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /go/src/github.com/moov-io/customers/bin/server /bin/server
# USER moov # TODO(adam): non-root users

EXPOSE 8080
EXPOSE 9090
ENTRYPOINT ["/bin/server"]
