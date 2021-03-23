BINARY_NAME := github-exporter
DOCKER_REPO ?= infinityworks/github-exporter
ARCH ?= darwin
GOBUILD_VERSION_ARGS := ""
TAG := $(shell cat VERSION)


.PHONY: install test build docker

all: build

install:
	@go mod download

test:
	@go test -v -race ./...

build: *.go
	@go build -v -o build/bin/$(ARCH)/$(BINARY_NAME) $(GOBUILD_VERSION_ARGS)

docker:
	docker build -t ${DOCKER_REPO}:$(TAG) .
