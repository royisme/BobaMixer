.PHONY: help build test lint fmt vet clean install dev hooks run coverage

# Variables
BINARY_NAME=boba
GO=go
GOFLAGS=-v
LDFLAGS=-s -w
BUILD_DIR=dist
COVERAGE_FILE=coverage.out

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
	$(GO) build $(GOFLAGS) -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/boba

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/boba
	GOOS=linux GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/boba
	GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/boba
	GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/boba
	GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/boba

install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install -trimpath -ldflags="$(LDFLAGS)" ./cmd/boba

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

fmt: ## Format Go code
	@echo "Formatting code..."
	gofmt -w -s .

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

lint-fast: ## Run golangci-lint in fast mode
	@echo "Running golangci-lint (fast mode)..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fast --timeout=3m; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

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
