build:
	@CGO_ENABLED=0 go build -buildvcs=false -a -o bin/app ./cmd
.PHONY: build

test:
	CGO_ENABLED=0 go test -v ./...
.PHONY: test

fmt:
	goimports -w -local github.com/alessio-perugini/f1calendarbot .
.PHONY: format

lint:
	golangci-lint run ./...
.PHONY: lint

mod-upgrade:
	@go get -u ./...
.PHONY: mod-upgrade
