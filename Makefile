BINARY ?= bin/gocensus
GOLANGCI_LINT ?= golangci-lint

.PHONY: build
build:
	go build -o $(BINARY) ./cmd/gocensus

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run ./...

.PHONY: fmt
fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

.PHONY: check
check: fmt lint test build

