.PHONY: build test test-unit test-integration lint security fmt clean help

BINARY_NAME=k8t
GO=go
GOFLAGS=-v

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS=-ldflags "-s -w \
	-X main.Version=$(VERSION) \
	-X main.GitCommit=$(COMMIT) \
	-X main.BuildDate=$(BUILD_DATE)"

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the k8t binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/k8t

install: ## Install k8t to $GOPATH/bin
	$(GO) install $(GOFLAGS) $(LDFLAGS) ./cmd/k8t

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	$(GO) test -v -race -coverprofile=coverage.out ./tests/unit/... ./pkg/...

test-integration: ## Run integration tests (requires kind cluster)
	$(GO) test -v -race ./tests/integration/...

test-contract: ## Run contract tests (RBAC validation)
	$(GO) test -v ./tests/contract/...

lint: ## Run linters (golangci-lint)
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

security: ## Run security scanners (gosec + govulncheck)
	@which gosec > /dev/null || (echo "gosec not found. Install: go install github.com/securego/gosec/v2/cmd/gosec@latest" && exit 1)
	@which govulncheck > /dev/null || (echo "govulncheck not found. Install: go install golang.org/x/vuln/cmd/govulncheck@latest" && exit 1)
	gosec -quiet ./...
	govulncheck ./...

fmt: ## Format Go code
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || true

vet: ## Run go vet
	$(GO) vet ./...

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out

coverage: test-unit ## Generate coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

benchmark: ## Run performance benchmarks
	$(GO) test -bench=. -benchmem ./tests/unit/...

mod-tidy: ## Tidy go.mod dependencies
	$(GO) mod tidy

mod-verify: ## Verify go.mod dependencies
	$(GO) mod verify

release: ## Build release binaries (requires goreleaser)
	@which goreleaser > /dev/null || (echo "goreleaser not found. Install: https://goreleaser.com/install/" && exit 1)
	goreleaser release --snapshot --clean

ci: fmt vet lint security test ## Run all CI checks

all: clean fmt vet lint test build ## Clean, format, lint, test, and build
