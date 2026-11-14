build:
	@CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/app ./cmd
.PHONY: build

test:
	go test -race ./...
.PHONY: test

fmt:
	golangci-lint fmt
.PHONY: fmt

lint:
	golangci-lint run ./...
.PHONY: lint

tidy-check:
	go mod tidy -diff
.PHONY: tidy-check

mod-upgrade:
	@go get -u ./...
.PHONY: mod-upgrade
