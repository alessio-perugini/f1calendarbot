APP_VERSION ?= "latest"

build:
	@CGO_ENABLED=0 go build -a -ldflags "-X main.version=${APP_VERSION}" -o bin/cmd ./cmd
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags "-X main.version=${APP_VERSION}" -o bin/cmd.exe cmd/main.go

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

mod-upgrade:
	@go get -u ./... && make mod-tidy
.PHONY: mod-upgrade