# Changelog

All notable changes to BobaMixer will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-11-13

### Added
- **Phase 0-4 Complete**: Core infrastructure and adapters
  - CLI commands: `ls`, `use`, `stats`, `edit`, `doctor`, `budget`, `hooks`, `action`, `report`
  - Configuration system with profiles, routes, pricing, and secrets
  - SQLite database for usage tracking
  - HTTP adapter for API providers (Anthropic, OpenRouter, etc.)
  - Tool adapter for CLI tools
  - TUI dashboard with usage statistics and trends

- **Phase 5**: Routing DSL and Suggestion Engine
  - `boba route test` command for offline route evaluation
  - Epsilon-greedy exploration (3% default rate)
  - Rule-based routing with context matching
  - Suggestion engine with confidence scores
  - Support for intent, text patterns, context size, project types, branches, time of day

- **Budget & Cost Tracking**
  - Budget alerts and monitoring
  - Daily and hard cap limits
  - Cost trend analysis (7/30 days)
  - Profile-based spending breakdown

- **Git Integration**
  - Hook management (`boba hooks install/remove/track`)
  - Project-level configuration discovery
  - Branch-aware routing

- **Documentation**
  - Comprehensive README
  - Quick Reference Guide
  - Adapter Development Guide
  - Routing Cookbook
  - Operations Guide (backup/cleanup)
  - FAQ

- **Release Infrastructure**
  - GoReleaser configuration
  - Multi-platform builds (macOS/Linux, amd64/arm64)
  - Automated changelog generation
  - Semantic versioning

### Security
- Strict file permissions for secrets.yaml (0600)
- #nosec annotations for justified security exceptions
- No API keys or request content stored in logs or database

### Fixed
- All golangci-lint errors resolved
- Package documentation added to all packages
- Error handling improved across codebase
- File permissions hardened

[0.1.0]: https://github.com/royisme/BobaMixer/releases/tag/v0.1.0
