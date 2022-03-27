build:
	@CGO_ENABLED=0 go build -a -o bin/cmd ./cmd
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -o bin/cmd.exe cmd/main.go

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
	@go get -u ./...
.PHONY: mod-upgrade