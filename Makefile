.PHONY: install test build

install:
	@go mod download

test:
	@go test -v -race ./...

build:
	@go build -o github-exporter

goimports:
	@goimports -w -local "github.com/githubexporter/github-exporter/" .