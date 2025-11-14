export GOTOOLCHAIN=auto

.PHONY: help build test lint fmt vet clean install dev hooks run coverage cover

# Variables
BINARY_NAME=boba
GO=go
GOFLAGS=-v
LDFLAGS=-s -w
VERSION?=$(shell git describe --tags --always --dirty --abbrev=8 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS=-ldflags "-X github.com/royisme/bobamixer/internal/version.Version=$(VERSION) -X github.com/royisme/bobamixer/internal/version.Commit=$(COMMIT) -X github.com/royisme/bobamixer/internal/version.Date=$(DATE)"
BUILD_DIR=dist
COVERAGE_FILE=coverage.out
COVER_MIN_TOTAL?=60
CORE_MIN_COVER?=80
CORE_PACKAGES?=internal/store/config internal/domain/stats internal/domain/suggestions
COVERAGE_PACKAGES?=./internal/store/config ./internal/domain/stats ./internal/domain/suggestions
TOOLS_BIN?=$(CURDIR)/bin
GOLANGCI_LINT_VERSION?=v1.60.1
GOLANGCI_LINT=$(TOOLS_BIN)/golangci-lint

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -trimpath $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/boba

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -trimpath $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/boba
	GOOS=linux GOARCH=arm64 $(GO) build -trimpath $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/boba
	GOOS=darwin GOARCH=amd64 $(GO) build -trimpath $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/boba
	GOOS=darwin GOARCH=arm64 $(GO) build -trimpath $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/boba
	GOOS=windows GOARCH=amd64 $(GO) build -trimpath $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/boba

install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install -trimpath $(BUILD_FLAGS) ./cmd/boba

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Coverage report:"
	$(GO) tool cover -func=$(COVERAGE_FILE)

coverage: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Open coverage.html in your browser to view the report"

cover: ## Run coverage with quality gates
	@echo "Running coverage with enforcement..."
	$(GO) test -covermode=atomic -coverprofile=$(COVERAGE_FILE) $(COVERAGE_PACKAGES)
	@$(GO) tool cover -func=$(COVERAGE_FILE)
	@$(GO) run ./scripts/covcheck -profile $(COVERAGE_FILE) -module github.com/royisme/bobamixer -min-total $(COVER_MIN_TOTAL) -min-core $(CORE_MIN_COVER) $(foreach pkg,$(CORE_PACKAGES),-core $(pkg))

fmt: ## Format Go code
	@echo "Formatting code..."
	gofmt -w -s .

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

lint: $(GOLANGCI_LINT) ## Run golangci-lint
	@echo "Running golangci-lint ($(GOLANGCI_LINT_VERSION))..."
	$(GOLANGCI_LINT) run --timeout=5m

lint-fast: $(GOLANGCI_LINT) ## Run golangci-lint in fast mode
	@echo "Running golangci-lint (fast mode, $(GOLANGCI_LINT_VERSION))..."
	$(GOLANGCI_LINT) run --fast --timeout=3m

$(GOLANGCI_LINT):
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@mkdir -p $(TOOLS_BIN)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
	       sh -s -- -b $(TOOLS_BIN) $(GOLANGCI_LINT_VERSION)

check: fmt vet lint-fast test ## Run all checks (format, vet, lint, test)
	@echo "All checks passed!"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod verify

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	$(GO) mod tidy

hooks: ## Install git hooks
	@echo "Installing git hooks..."
	@chmod +x .githooks/pre-commit
	@git config core.hooksPath .githooks
	@echo "Git hooks installed successfully!"
	@echo "Pre-commit hook will now run automatically on each commit."

run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	$(GO) run ./cmd/boba

dev: hooks ## Setup development environment
	@echo "Setting up development environment..."
	@$(MAKE) deps
	@$(MAKE) hooks
	@echo "Development environment ready!"
	@echo "Run 'make help' to see available commands."

ci: ## Run CI checks locally
	@echo "Running CI checks..."
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) lint
	@$(MAKE) test-coverage
	@$(MAKE) build
	@echo "All CI checks passed!"

# Version management
version: ## Show current version
	@echo "Current version: $(shell git describe --tags --abbrev=0 2>/dev/null || echo 'v1.0.0-dev')"
	@./dist/boba version 2>/dev/null || echo "Build the binary first with 'make build'"

bump: ## Show version bump information (same as 'boba bump --dry-run')
	@echo "Analyzing changes since last tag..."
	@./dist/boba bump --dry-run || echo "Build the binary first with 'make build'"

bump-patch: ## Bump patch version and commit changes
	@./dist/boba bump patch || echo "Build the binary first with 'make build'"

bump-minor: ## Bump minor version and commit changes
	@./dist/boba bump minor || echo "Build the binary first with 'make build'"

bump-major: ## Bump major version and commit changes
	@./dist/boba bump major || echo "Build the binary first with 'make build'"

bump-auto: ## Auto-detect version bump type based on conventional commits
	@./dist/boba bump auto || echo "Build the binary first with 'make build'"

# Release targets
release: ## Auto-detect version and create release tag
	@./dist/boba release --auto || echo "Build the binary first with 'make build'"

release-auto: ## Auto-detect version and create release tag (alias for release)
	@$(MAKE) release

tag: ## Create and push a new version tag
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "Tag $(VERSION) pushed successfully!"

release-patch: ## Create patch release (vX.Y.Z+1)
	@$(MAKE) bump-patch
	@$(MAKE) release

release-minor: ## Create minor release (vX.Y+1.0)
	@$(MAKE) bump-minor
	@$(MAKE) release

release-major: ## Create major release (vX+1.0.0)
	@$(MAKE) bump-major
	@$(MAKE) release
