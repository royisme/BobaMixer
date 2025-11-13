# Contributing to BobaMixer

Thank you for your interest in contributing to BobaMixer! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.25.4 or later
- Git
- Make (optional but recommended)

### Initial Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/royisme/BobaMixer.git
   cd BobaMixer
   ```

2. **Install dependencies**
   ```bash
   make deps
   # or
   go mod download
   ```

3. **Install Git hooks**
   ```bash
   make hooks
   # or manually:
   chmod +x .githooks/pre-commit
   git config core.hooksPath .githooks
   ```

4. **Verify setup**
   ```bash
   make test
   make build
   ```

## Development Workflow

### Quick Start

The fastest way to get started:
```bash
make dev    # Sets up everything (dependencies + git hooks)
make run    # Run the application
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Install to GOPATH/bin
make install
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Generate HTML coverage report
make coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Run fast linting (for quick checks)
make lint-fast

# Run all checks (format, vet, lint, test)
make check
```

### Pre-commit Hook

The pre-commit hook automatically:
- Formats all staged Go files with `gofmt`
- Runs `go vet` on changed packages
- Blocks commits with unformatted code

If you need to bypass the hook (not recommended):
```bash
git commit --no-verify
```

## Code Style

### Formatting

- All Go code must be formatted with `gofmt`
- The pre-commit hook enforces this automatically
- Run `make fmt` before committing

### Linting

- We use `golangci-lint` with a custom configuration
- Configuration is in `.golangci.yml`
- Run `make lint` to check for issues
- All linting issues must be resolved before merging

### Best Practices

1. **Error Handling**: Always check and handle errors
   ```go
   if err != nil {
       return fmt.Errorf("operation failed: %w", err)
   }
   ```

2. **Testing**: Write tests for new functionality
   ```bash
   make test
   ```

3. **Documentation**: Add comments for exported functions
   ```go
   // ProcessData processes the input data and returns the result.
   func ProcessData(input string) (string, error) {
       // ...
   }
   ```

4. **Commit Messages**: Use clear, descriptive commit messages
   ```
   Add user authentication feature
   
   - Implement login/logout functionality
   - Add session management
   - Update UI for auth flow
   ```

## CI/CD

### Continuous Integration

Our CI pipeline runs on every push and pull request:

1. **Lint & Format** - Fast checks that fail early
   - Code formatting verification
   - `go vet` analysis
   - `golangci-lint` checks

2. **Test** - Comprehensive testing
   - Unit tests with race detection
   - Coverage reporting

3. **Build** - Multi-platform builds
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

### Running CI Locally

Before pushing, run all CI checks locally:
```bash
make ci
```

This runs the same checks as the CI pipeline.

## Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes**
   - Write code
   - Add tests
   - Update documentation

3. **Run checks locally**
   ```bash
   make check
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "Add my feature"
   ```
   The pre-commit hook will run automatically.

5. **Push and create PR**
   ```bash
   git push origin feature/my-feature
   ```

6. **Wait for CI**
   - All CI checks must pass
   - Address any review comments

## Common Issues

### golangci-lint not found

Install golangci-lint:
```bash
# macOS/Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or use go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Pre-commit hook not running

Ensure the hook is installed:
```bash
make hooks
```

Verify configuration:
```bash
git config core.hooksPath
# Should output: .githooks
```

### Tests failing

Run tests with verbose output:
```bash
go test -v ./...
```

## Makefile Commands

Run `make help` to see all available commands:

```bash
make help
```

Available commands:
- `build` - Build the binary
- `test` - Run tests
- `lint` - Run linters
- `fmt` - Format code
- `check` - Run all checks
- `clean` - Clean build artifacts
- `help` - Show help message

## Questions?

If you have any questions or need help, please:
- Open an issue
- Check existing documentation
- Ask in pull request comments
