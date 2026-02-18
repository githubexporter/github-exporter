.PHONY: install test build test-coverage coverage-report coverage

install:
	@go mod download

build:
	@go build -o main

test:
	@go test -v -race ./...

test-coverage:
	@go test -v -race -coverprofile=coverage.out -coverpkg=./... ./...

coverage-report:
	@go tool cover -html=coverage.out

coverage: test-coverage coverage-report