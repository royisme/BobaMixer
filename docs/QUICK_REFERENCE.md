# Quick Reference Guide

## Common Commands

### Development Setup
```bash
make dev              # Complete setup (deps + hooks)
make deps             # Download dependencies only
make hooks            # Install git hooks only
```

### Building
```bash
make build            # Build for current platform
make build-all        # Build for all platforms
make install          # Install to GOPATH/bin
```

### Testing
```bash
make test             # Run all tests
make test-coverage    # Run tests with coverage
make coverage         # Generate HTML coverage report
```

### Code Quality
```bash
make fmt              # Format all Go code
make vet              # Run go vet
make lint             # Run golangci-lint (full)
make lint-fast        # Run golangci-lint (fast mode)
make check            # Run all checks (fmt + vet + lint + test)
```

### CI/CD
```bash
make ci               # Run all CI checks locally
```

### Utilities
```bash
make clean            # Remove build artifacts
make tidy             # Tidy go.mod
make run              # Run the application
make help             # Show all available commands
```

## Git Hooks

### Pre-commit Hook
Automatically runs on every commit:
- Formats staged Go files with `gofmt`
- Runs `go vet` on changed packages
- Blocks commits with unformatted code

### Bypass Hook (Not Recommended)
```bash
git commit --no-verify
```

### Reinstall Hook
```bash
make hooks
```

## CI Pipeline

### Jobs
1. **Lint & Format** (~1 min)
   - Code formatting check
   - go vet analysis
   - golangci-lint checks

2. **Test** (~1-2 min)
   - Unit tests with race detection
   - Coverage reporting

3. **Build** (~2 min)
   - Multi-platform builds
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

### Artifacts
- Coverage reports (7 days retention)
- Build binaries (7 days retention)

## Workflow Examples

### Starting a New Feature
```bash
# 1. Set up development environment
make dev

# 2. Create feature branch
git checkout -b feature/my-feature

# 3. Make changes
# ... edit files ...

# 4. Run checks
make check

# 5. Commit (pre-commit hook runs automatically)
git add .
git commit -m "Add my feature"

# 6. Push
git push origin feature/my-feature
```

### Before Pushing
```bash
# Run all checks locally (same as CI)
make ci
```

### Quick Iteration
```bash
# Make changes
vim internal/cli/root.go

# Fast check
make lint-fast

# Run tests
make test

# Commit
git commit -am "Fix issue"
```

## Troubleshooting

### golangci-lint not found
```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or use go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Pre-commit hook not running
```bash
# Reinstall hooks
make hooks

# Verify
git config core.hooksPath
# Should output: .githooks
```

### Tests failing
```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestName ./internal/package
```

### Build failing
```bash
# Clean and rebuild
make clean
make build

# Check for issues
make vet
```

## File Locations

- **Makefile**: `/Makefile`
- **Pre-commit hook**: `/.githooks/pre-commit`
- **golangci-lint config**: `/.golangci.yml`
- **CI workflow**: `/.github/workflows/ci.yml`
- **Contributing guide**: `/CONTRIBUTING.md`

## Quick Tips

1. **Always run `make check` before pushing**
2. **Use `make ci` to simulate CI locally**
3. **The pre-commit hook will auto-format your code**
4. **Use `make help` to see all available commands**
5. **Build artifacts are in `dist/` (ignored by git)**

## Performance

| Command | Time |
|---------|------|
| `make lint-fast` | ~30s |
| `make test` | ~1-2s |
| `make build` | ~2-3s |
| `make check` | ~30-40s |
| `make ci` | ~40-50s |

## Resources

- Contributing Guide: [CONTRIBUTING.md](CONTRIBUTING.md)
- Makefile Help: `make help`
- CI Configuration: [.github/workflows/ci.yml](.github/workflows/ci.yml)
