SRC := $(wildcard cmd/omega-server/*.go pkg/**/*.go internal/**/*.go assets/*)
GOOSE_TAGS := no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb
GOOSE_CMD := go run -tags "$(GOOSE_TAGS)" github.com/pressly/goose/v3/cmd/goose@latest

run: build
	./bin/omega-server
.PHONY: run

## build: build the application
build: bin/omega-server
.PHONY: build

bin/omega-server: $(SRC) scripts/build
	./scripts/build omega-server

## run/live: run the application with reloading on file changes
run/live: .air.toml
	go run github.com/air-verse/air@latest
.PHONY: run/live

## audit: run quality control checks
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)"
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	@command -v bodyclose >/dev/null 2>&1 || go install github.com/timakin/bodyclose@latest
	go vet -vettool="$$(which bodyclose)" ./...
	go run github.com/client9/misspell/cmd/misspell@latest -locale="US" -error -source="text" **/*
.PHONY: audit

## test: run all tests
test:
	go test -v -race -buildvcs ./...
.PHONY: test

## test/cover: run all tests and display coverage
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out
.PHONY: test/cover

## upgradeable: list direct dependencies that have upgrades available
upgradeable:
	@go run github.com/oligot/go-mod-upgrade@latest
.PHONY: upgradeable

## tidy: tidy modfiles and format .go files
tidy: fmt deps
.PHONY: tidy

## fmt: format .go files
fmt:
	go fmt ./...
	find . -name \*.go -not -path .git -not -path bin -exec go run golang.org/x/tools/cmd/goimports@latest -w {} \;
.PHONY: fmt

## deps: install and verify dependencies
deps:
	go get -u ./...
	go mod tidy -v
	go mod verify
	go mod download
	go mod vendor
.PHONY: deps

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help

## migrations/create name=$1: create a new database migration
migrations/create:
	$(GOOSE_CMD) -s create $(name) sql
.PHONY: migrations/create

## migrations/up: apply all migrations
migrations/up:
	$(GOOSE_CMD) up
.PHONY: migrations/up

## migrations/up-by-one migrate up a single migration from the current version
migrations/up-by-one:
	$(GOOSE_CMD) up-to $(version)
.PHONY: migrations/up-by-one

## migrations/up-to version=$1: migrate up to a specific version
migrations/up-to:
	$(GOOSE_CMD) up-to $(version)
.PHONY: migrations/up-to

## migrations/down: roll back a single migration from the current version
migrations/down:
	$(GOOSE_CMD) down
.PHONY: migrations/down

## migrations/down-to: roll back to a specific version
migrations/down-to:
	$(GOOSE_CMD) down-to $(version)
.PHONY: migrations/down-to

## migrations/redo: re-run the last migration
migrations/redo:
	$(GOOSE_CMD) redo
.PHONY: migrations/redo

# migrations/reset: roll back all migrations
migrations/reset:
	$(GOOSE_CMD) reset
.PHONY: migrations/reset

## migrations/status: prints the status of all migrations
migrations/status:
	$(GOOSE_CMD) status
.PHONY: migrations/status

## migrations/version: print the current in-use migration version
migrations/version:
	$(GOOSE_CMD) version
.PHONY: migrations/version

## migrations/fix: apply sequential ordering to migrations
migrations/fix:
	$(GOOSE_CMD) fix
.PHONY: migrations/fix

## migrations/validate: check migration files without running them
migrations/validate:
	$(GOOSE_CMD) validate
.PHONY: migrations/validate

## postgres/run: run postgres in a container
postgres/run:
	podman compose up postgres -d
.PHONY: postgres/run
