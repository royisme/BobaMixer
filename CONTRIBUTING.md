# Contributing to BobaMixer

Thank you for your interest in contributing to BobaMixer! This document provides guidelines for contributing to the project.

## 📋 Table of Contents

- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Commit Message Format](#commit-message-format)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## 🛠️ Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/yourusername/BobaMixer.git
   cd BobaMixer
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Set up development tools**
   ```bash
   make dev
   ```

4. **Run tests**
   ```bash
   make test
   ```

## 📝 Code Style

### Go Code Style

We use `golangci-lint` for code formatting and linting:

```bash
make fmt
make lint
```

### Documentation

- Write clear, concise documentation
- Use markdown for all documentation
- Update docs when adding new features

## 💬 Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix  
- **docs**: Documentation changes
- **style**: Code formatting (no functional changes)
- **refactor**: Code refactoring
- **test**: Test additions/modifications
- **chore**: Build process, dependency updates
- **perf**: Performance improvements

### Examples

#### New Feature
```
feat(routing): add intelligent epsilon-greedy exploration

Implement epsilon-greedy algorithm for intelligent routing decisions
to balance exploration vs exploitation when selecting AI providers.

- Add epsilon parameter with default 0.1
- Implement exploration strategy with random selection
- Add metrics tracking for exploration effectiveness

Closes #123
```

#### Bug Fix
```
fix(config): resolve profile loading issue in Windows

Fixed bug where profiles.yaml couldn't be loaded on Windows systems
due to incorrect path handling. Use filepath.Join() for cross-platform
compatibility.

Fixes #456
```

#### Breaking Changes
```
feat(api): redesign configuration schema v2

BREAKING CHANGE: Configuration format has changed from v1 to v2.
Old configurations are no longer supported. Migration guide is
provided in docs/configuration/migration.md.

- New profiles.yaml structure
- Enhanced routing rules
- Better validation

Migration script available: `boba migrate --from-v1`
```

## 🔄 Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/intelligent-routing
   ```

2. **Make your changes**
   - Follow the commit message format
   - Add tests for new functionality
   - Update documentation

3. **Run tests and checks**
   ```bash
   make check
   ```

4. **Create Pull Request**
   - Use descriptive title
   - Link related issues
   - Include screenshots if applicable
   - Request review from maintainers

5. **Address feedback**
   - Respond to reviewer comments
   - Make requested changes
   - Push updates to your branch

## 🚀 Release Process

We use automated releases based on conventional commits:

### Automated Versioning

- **feat** → Minor version bump (X.Y.0)
- **fix** → Patch version bump (X.Y.Z)
- **BREAKING CHANGE** → Major version bump (X.0.0)

### Creating a Release

1. **Merge changes to main**
   ```bash
   git checkout main
   git pull origin main
   git merge feature/your-feature
   git push origin main
   ```

2. **Create release tag** (optional, goreleaser can auto-create)
   ```bash
   # Goreleaser will create this automatically
   # Or manually: git tag v1.2.3
   # git push origin v1.2.3
   ```

3. **Release artifacts are automatically created**
   - GitHub Release with changelog
   - Binary builds for all platforms
   - Homebrew formula update
   - Documentation deployment

### Release Checklist

Before creating a release, ensure:

- [ ] All tests pass
- [ ] Documentation is updated
- [ ] CHANGELOG.md is current
- [ ] Version number is appropriate
- [ ] Breaking changes are documented
- [ ] Security review is complete (if needed)

## 📊 Tracking Changes

Every significant change should be documented:

### Features
- What problem does this solve?
- How does it work?
- Examples and usage

### Bug Fixes  
- What was the issue?
- Root cause analysis
- How it was fixed

### Breaking Changes
- What changed and why
- Migration guide
- Deprecation timeline

## 🐛 Reporting Issues

When reporting issues, please include:

1. **Environment information**
   - BobaMixer version
   - Operating system
   - Go version

2. **Steps to reproduce**
   - Minimal reproduction case
   - Expected vs actual behavior

3. **Additional context**
   - Configuration files
   - Error messages
   - Logs

## 💬 Getting Help

- Create an issue for bugs or feature requests
- Start a discussion for questions
- Check existing issues and documentation

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.