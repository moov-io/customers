PLATFORM=$(shell uname -s | tr '[:upper:]' '[:lower:]')
VERSION := $(shell grep -Eo '(v[0-9]+[\.][0-9]+[\.][0-9]+(-[a-zA-Z0-9]*)?)' version.go)

USERID := $(shell id -u $$USER)
GROUPID:= $(shell id -g $$USER)

.PHONY: build build-server build-examples docker release check install

build: build-server

build-server:
	pkger -include /migrations
	CGO_ENABLED=1 go build -o ./bin/server github.com/moov-io/customers/cmd/server

.PHONY: check
check:
ifeq ($(OS),Windows_NT)
	@echo "Skipping checks on Windows, currently unsupported."
else
	@wget -O lint-project.sh https://raw.githubusercontent.com/moov-io/infra/master/go/lint-project.sh
	@chmod +x ./lint-project.sh
	GOCYCLO_LIMIT=27 ./lint-project.sh
endif

.PHONY: admin
admin:
	@rm -rf ./pkg/admin
	docker run --rm \
		-u $(USERID):$(GROUPID) \
		-v ${PWD}:/local openapitools/openapi-generator-cli:v4.3.1 batch -- /local/.openapi-generator/admin-generator-config.yml
	rm -f ./pkg/admin/go.mod ./pkg/admin/go.sum
	gofmt -w ./pkg/admin/
	go build github.com/moov-io/customers/pkg/admin

.PHONY: client
client:
	@rm -rf ./pkg/client
	docker run --rm \
		-u $(USERID):$(GROUPID) \
		-v ${PWD}:/local openapitools/openapi-generator-cli:v4.3.1 batch -- /local/.openapi-generator/client-generator-config.yml
	rm -f ./pkg/client/go.mod ./pkg/client/go.sum
	gofmt -w ./pkg/client/
	go build github.com/moov-io/customers/pkg/client

.PHONY: clean
clean:
	@rm -rf ./bin/ cover.out coverage.txt openapi-generator-cli-*.jar misspell* staticcheck* lint-project.sh

dist: clean
	pkger -include /migrations
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=1 GOOS=windows go build -o bin/customers.exe github.com/moov-io/customers/cmd/server
else
	CGO_ENABLED=1 GOOS=$(PLATFORM) go build -o bin/customers-$(PLATFORM)-amd64 github.com/moov-io/customers/cmd/server
endif

docker: clean
# Docker image
	docker build --pull -t moov/customers:$(VERSION) -f Dockerfile .
	docker tag moov/customers:$(VERSION) moov/customers:latest
# OpenShift Docker image
#	docker build --pull -t quay.io/moov/customers:$(VERSION) -f Dockerfile-openshift --build-arg VERSION=$(VERSION) .
#	docker tag quay.io/moov/customers:$(VERSION) quay.io/moov/customers:latest

clean-integration:
	docker-compose kill
	docker-compose rm -v -f

test-integration: clean-integration
	docker-compose up -d
	sleep 10
	curl -v http://localhost:9097/live

release: docker AUTHORS
	go vet ./...
	go test -coverprofile=cover-$(VERSION).out ./...
	git tag -f $(VERSION)

release-push:
	docker push moov/customers:$(VERSION)
	docker push moov/customers:latest

# quay-push:
# 	docker push quay.io/moov/customers:$(VERSION)
# 	docker push quay.io/moov/customers:latest

.PHONY: cover-test cover-web
cover-test:
	go test -coverprofile=cover.out ./...
cover-web:
	go tool cover -html=cover.out

# From https://github.com/genuinetools/img
.PHONY: AUTHORS
AUTHORS:
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@echo "$(shell git log --format='\n%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@

install:
	go install github.com/markbates/pkger/cmd/pkger
