# Release Guide

This guide covers BobaMixer's automated release workflow and version management system.

## Quick Release

### Automatic Release (Recommended)
```bash
# Automatic release with version detection from conventional commits
make release-auto
# or
./dist/boba-maint release --auto
```

### Manual Version Bumps
```bash
# Patch release (0.1.0 -> 0.1.1)
make release-patch

# Minor release (0.1.0 -> 0.2.0)  
make release-minor

# Major release (0.1.0 -> 1.0.0)
make release-major
```

## Version Management

### Version Bump Commands
```bash
# Preview version bump without applying
./dist/boba-maint bump patch --dry-run
./dist/boba-maint bump minor --dry-run
./dist/boba-maint bump major --dry-run
./dist/boba-maint bump auto --dry-run  # Auto-detect from commits

# Apply version bump
./dist/boba-maint bump patch
./dist/boba-maint bump minor
./dist/boba-maint bump major
./dist/boba-maint bump auto  # Smart detection based on commits
```

### Automatic Version Detection
The `auto` command analyzes recent commits following conventional commits:

- `feat:` commits → minor version bump
- `fix:` commits → patch version bump  
- `BREAKING CHANGE:` → major version bump
- `perf!` → major version bump

### Conventional Commits Format
```
feat: add new AI provider support
fix: resolve routing logic error
docs: update installation guide
refactor: optimize token counting
BREAKING CHANGE: change configuration format
```

## Release Process

### Prerequisites
1. Clean working directory (`git status` clean)
2. Up-to-date main branch (`git pull origin main`)
3. Built binary (`make build`)

### Automated Release Flow
1. **Version Detection**: Analyze commits for version bump type
2. **Version Update**: Update version files and create git tag
3. **Build & Test**: Run tests and build binaries
4. **GitHub Release**: Create GitHub release with changelog
5. ** Goreleaser**: Build and distribute binaries
   > `make release-auto` (and the `release-patch/minor/major` targets) now perform these steps and push the new tag to `origin`, so GitHub Actions starts automatically without a separate `git push`.

### Manual Release Steps
```bash
# 1. Build current version
make build

# 2. Determine version bump type
./dist/boba-maint bump auto --dry-run

# 3. Apply version bump  
./dist/boba-maint bump auto

# 4. Create and push release commit/tag
./dist/boba-maint release --part patch   # or --auto, --part minor, etc.
```

## Development Scripts

### Development Helper
```bash
# Interactive development menu
./scripts/dev.sh
```
Features:
- Quick build and test
- Version bumping
- Release preparation
- Common development tasks

### Release Automation Script  
```bash
# Interactive release management
./scripts/release.sh
```
Features:
- Guided version bumping
- Automated release process
- Safety checks and validation
- Rollback capabilities

## Makefile Targets

```bash
# Version Management
make bump-patch    # Bump patch version
make bump-minor    # Bump minor version  
make bump-major    # Bump major version
make bump-auto     # Auto-detect and bump

# Release Management
make release-patch # Create patch release
make release-minor # Create minor release
make release-major # Create major release
make release-auto  # Auto-release with version detection

# Development
make build         # Build binary
make test          # Run tests
make check         # Run all checks
```

## Configuration

### Goreleaser Configuration
- `.goreleaser.yml` - Multi-platform build configuration
- Supports Linux, macOS, Windows (AMD64, ARM64)
- Creates GitHub releases with assets
- Generates Homebrew formula updates

### Version Information
- Build version injected at compile time
- `./dist/boba version` shows version details
- Includes commit SHA, build date, and builder info

## Changelog Generation

Releases automatically include:
- Conventional commits since last release
- Version comparison and summary
- Links to relevant commits and issues

View with: `./dist/boba changelog [from-version] [to-version]`

## Release Safety

### Pre-release Checks
- Working directory must be clean
- Tests must pass
- Version must be valid semver
- No conflicting tags exist

### Rollback Process
```bash
# Delete problematic tag
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3

# Delete GitHub release (via web interface)
# Restore version files from git if needed
git checkout HEAD~1 -- internal/version/version.go
```

## Best Practices

1. **Use Conventional Commits**: Follow commit message format for auto-versioning
2. **Test Before Release**: Always run `make test` before releasing
3. **Use --dry-run**: Preview version changes before applying
4. **Regular Releases**: Use `make release-auto` for frequent small releases
5. **Monitor CI**: Check GitHub Actions for any build issues

## Troubleshooting

### Common Issues

**"Working directory not clean"**
```bash
git status
git add .
git commit -m "chore: prepare for release"
```

**"No conventional commits found"**
```bash
# Check recent commits
git log --oneline -10
# Manual version bump
./dist/boba-maint bump patch
```

**"Tag already exists"**
```bash
# Delete local tag
git tag -d v1.2.3
# Delete remote tag
git push origin :refs/tags/v1.2.3
# Retry release
```

### Getting Help
- Check `./dist/boba --help` for command options
- Review GitHub Actions logs for CI failures
- Verify goreleaser configuration syntax
