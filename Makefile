VERSION ?= "latest"

build:
	CGO_ENABLED=0 go build -o bin/cmd ./cmd
.PHONY: build

test:
	CGO_ENABLED=0 go test -v ./...
.PHONY: test

fmt:
	goimports -w .
.PHONY: format

lint:
	golangci-lint run ./...
.PHONY: lint