BINARY ?= bin/gocensus
GOLANGCI_LINT ?= golangci-lint
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null)
ifeq ($(GIT_VERSION),)
GIT_VERSION := dev
endif
VERSION ?= $(GIT_VERSION)
LDFLAGS ?= -s -w -X main.version=$(VERSION)

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/gocensus

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
