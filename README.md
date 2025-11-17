# BobaMixer 🧋

> **Intelligent Router & Cost Optimizer for AI Workflows**

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/royisme/BobaMixer)](https://github.com/royisme/BobaMixer/releases)
[![golangci-lint](https://img.shields.io/badge/lint-passing-brightgreen)](https://golangci-lint.run/)

[**📚 Documentation**](https://royisme.github.io/BobaMixer/) | [**🚀 Quick Start**](#quick-start) | [**中文文档**](README.zh.md)


---

## Feature Overview

**Core (Control Plane + boba run)**
- Manage Providers / Tools / Bindings as first-class objects
- Run local AI CLI tools with auto-injected credentials and endpoints via `boba run`
- Optional local proxy to consolidate requests

**Advanced (legacy/optional)**
- Routing / Profiles
- Budget & Pricing controls
- Usage Stats & Git hooks

---

## Why BobaMixer?

In daily AI development, have you encountered these pain points?

- 🔑 **API Key Chaos** - Multiple AI service credentials scattered everywhere, switching providers requires config file changes
- 💸 **Runaway Costs** - API bills skyrocket without warning, no real-time monitoring or budget control
- 🎯 **Routing Decisions** - Which model for which task? How to balance cost vs quality?
- 📊 **Missing Usage Data** - Cannot track token consumption or cost distribution, lack of optimization insights
- 🔄 **High Switching Cost** - Moving from Claude to OpenAI requires code changes, no flexible orchestration

**BobaMixer was born to solve these problems** — It's your AI workflow control plane, letting you orchestrate AI models like microservices.

---

## Core Capabilities

### 1. Unified Control Plane

No more hardcoded API keys and endpoints in code - everything is configuration-driven:

```bash
# View all available AI providers
$ boba providers

Provider              Kind        Endpoint                      Status
claude-anthropic      anthropic   https://api.anthropic.com    ✓ Ready
claude-zai            anthropic   https://api.z.ai/api/...      ✓ Ready
openai-official       openai      https://api.openai.com        ✓ Ready
gemini-official       gemini      https://generativelanguage... ✓ Ready

# Bind local CLI tool to provider
$ boba bind claude claude-zai

# Auto-inject config at runtime
$ boba run claude "Write a function to calculate fibonacci"
```

**Core Value**: Decoupled configuration from code, configure once, apply globally.

### 2. Local HTTP Proxy

Start an intelligent proxy locally to intercept all AI API calls:

```bash
# Start proxy server (127.0.0.1:7777)
$ boba proxy serve &

# All requests through proxy are automatically logged
# Supports both OpenAI and Anthropic API formats
```

**Technical Highlights**:
- **Zero-intrusion integration** - Just modify `ANTHROPIC_BASE_URL` env var
- **Automatic token parsing** - Extract precise input/output tokens from responses
- **Real-time cost calculation** - Calculate per-request cost based on latest pricing
- **Thread-safe** - Concurrent request support with `sync.RWMutex`

## Advanced Capabilities (Legacy / Optional)

> The following modules are advanced/legacy features. They are not part of the core Control Plane + `boba run` path, but remain available for power users.

### [Advanced] Intelligent Routing Engine

Automatically select optimal model based on task characteristics:

```yaml
# ~/.boba/routes.yaml
rules:
  - id: "large-context"
    if: "ctx_chars > 50000"
    use: "claude-anthropic"     # Long context → Claude
    explain: "Large context requires Claude's 200K window"

  - id: "code-review"
    if: "text.matches('review|audit|refactor')"
    use: "openai-gpt4"           # Code review → GPT-4
    fallback: "claude-anthropic"

  - id: "budget-conscious"
    if: "time_of_day == 'night' && budget.remaining < 5.0"
    use: "gemini-flash"          # Night + low budget → Cheap model
```

Test routing decisions:

```bash
$ boba route test "Please review this PR and check for security issues"

=== Routing Decision ===
Profile: openai-gpt4
Rule ID: code-review
Explanation: Code review tasks use GPT-4 for best results
Fallback: claude-anthropic
```

**Core Algorithm**: Epsilon-Greedy exploration + Rule engine, auto-balancing between cost optimization and quality exploration.

### [Advanced] Budget Management & Alerts

Multi-level budget control to prevent cost overruns:

```bash
# View current budget status
$ boba budget --status

Budget Scope: project (my-chatbot)
========================================
Today:  $2.34 of $10.00 (23.4%)
Period: $45.67 of $300.00 (15.2%)
Days Remaining: 23

# Set budget limits
$ boba budget --daily 10.00 --cap 300.00

# Auto-switch to cheaper provider when over budget
$ boba action --auto
```

**Technical Implementation**:
- Pre-request budget check (`checkBudgetBeforeRequest`)
- Conservative token estimation (1000 input, 500 output)
- HTTP 429 response when budget exceeded
- Graceful degradation - allows pass-through without budget config

### [Advanced] Usage Analytics & Cost Tracking

Precise token-level tracking with multi-dimensional analysis:

```bash
# View today's stats
$ boba stats --today

Today's Usage
=============
Tokens:   45,678
Cost:     $1.23
Sessions: 12

# 7-day trend analysis
$ boba stats --7d --by-profile

Last 7 Days Usage
=================
Total Tokens:   312,456
Total Cost:     $8.76
Avg Daily Cost: $1.25

By Profile:
-----------
- openai-gpt4: tokens=180K cost=$6.20 sessions=45 avg_latency=1200ms usage=57.6% cost=70.8%
- claude-sonnet: tokens=90K cost=$1.80 sessions=23 avg_latency=980ms usage=28.8% cost=20.5%
- gemini-flash: tokens=42K cost=$0.76 sessions=18 avg_latency=650ms usage=13.5% cost=8.7%

# Export report
$ boba report --format json --output monthly-report.json
```

**Data Schema**:
- `sessions` table - Records session metadata (project, branch, profile, latency)
- `usage_records` table - Precise token & cost records, 3 estimation levels (exact/mapped/heuristic)
- SQLite storage - Local, no external database dependency

### 6. Real-time Pricing Updates

Auto-fetch latest model pricing from OpenRouter API:

```bash
# Configure pricing refresh strategy
# ~/.boba/pricing.yaml
refresh:
  interval_hours: 24
  on_startup: false

# Manually verify pricing data
$ boba doctor --pricing

Pricing Validation
==================
✓ OpenRouter API accessible
✓ Cache fresh (updated 2 hours ago)
✓ 1,247 models loaded
✓ Fallback to vendor JSON available
```

**Loading Strategy** (Multi-layer Fallback):
1. OpenRouter API (15s timeout)
2. Local cache (24h TTL)
3. Vendor JSON (embedded data)
4. pricing.yaml (user-defined)
5. profiles.yaml cost_per_1k (final fallback)

---

## Technical Architecture

### Modular Design

```
BobaMixer
├── cmd/boba              # CLI entry point
├── internal/cli          # Command implementations
├── internal/domain       # Core domain logic
│   ├── budget           # Budget tracking
│   ├── pricing          # Pricing mgmt (OpenRouter)
│   ├── routing          # Routing engine
│   ├── stats            # Statistical analysis
│   └── suggestions      # Optimization suggestions
├── internal/proxy        # HTTP proxy server
├── internal/store        # Data storage
│   ├── config           # Config loading
│   └── sqlite           # SQLite operations
└── internal/ui           # TUI Dashboard (Bubble Tea)
```

### Key Tech Stack

- **Language**: Go 1.25+ (Type-safe, concurrency-friendly, single-binary deployment)
- **TUI**: Bubble Tea (Modern terminal UI framework)
- **Storage**: SQLite (Zero-config, local, SQL analytics support)
- **Linting**: golangci-lint (Strict code quality standards)
- **API Integration**: OpenRouter Models API (1000+ model pricing)

### Go Best Practices

Project strictly follows Go language standards:

- ✅ **golangci-lint verified** - 0 issues
- ✅ **Documentation** - All exported types/functions have doc comments
- ✅ **Error handling** - Complete error wrapping & graceful degradation
- ✅ **Concurrency safety** - `sync.RWMutex` protects shared state
- ✅ **Security** - All exceptions marked with `#nosec` after audit

---

## Quick Start

### Installation

```bash
# Using Go
go install github.com/royisme/bobamixer/cmd/boba@latest

# Or using Homebrew
brew tap royisme/tap
brew install bobamixer
```

### First Time Setup - Interactive Onboarding 🎯

BobaMixer will automatically guide you through all configurations, **no manual YAML editing required**:

```bash
# 1. Launch BobaMixer (first run triggers onboarding wizard)
$ boba

# Onboarding wizard will automatically:
# ✓ Detect local CLI tools (claude/codex/gemini)
# ✓ Let you select Provider
# ✓ Guide you to input API Key (secure input, auto-save)
# ✓ Create all config files
# ✓ Verify configuration

# 2. Ready to use after setup
$ boba run claude --version
```

### Alternative: CLI Setup (for power users)

If you prefer command-line setup:

```bash
# 1. Initialize config directory
$ boba init

# 2. Configure API Key (secure input, no YAML editing needed)
$ boba secrets set claude-anthropic-official
Enter API key: ********
✓ API key saved to ~/.boba/secrets.yaml (permissions: 0600)

# 3. Bind tool to Provider
$ boba bind claude claude-anthropic-official

# 4. Verify configuration
$ boba doctor

# 5. Run
$ boba run claude --version
```

### Environment Variables (optional)

You can also use environment variables (suitable for CI/CD or temporary use):

```bash
# BobaMixer prioritizes environment variables
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."

# Then run directly
$ boba run claude --version
```

### Launch TUI Dashboard

```bash
$ boba
```

**Interactive controls**:
- `↑/↓` Select tool
- `B` Switch Provider binding
- `X` Toggle Proxy
- `V` Stats view
- `R` Run tool
- `q` Quit

---

## Use Cases

### Case 1: Team Collaboration - Unified API Management

**Problem**: Team members manage API keys separately, easy to leak and hard to audit.

**Solution**:
```bash
# 1. Create .boba-project.yaml in project root
$ cat .boba-project.yaml
project:
  name: "my-chatbot"
  type: ["backend", "ai"]
  preferred_profiles: ["claude-anthropic", "openai-gpt4"]

budget:
  daily_usd: 20.0
  hard_cap: 600.0

# 2. Each team member configures ~/.boba/secrets.yaml

# 3. Project-level budget auto-applies
$ cd my-chatbot
$ boba budget --status  # Auto-detects project budget
```

### Case 2: Cost Optimization - Auto Downgrade

**Problem**: Development uses expensive models, costs skyrocket during testing.

**Solution**:
```yaml
# routes.yaml - Auto-select model based on branch
rules:
  - id: "production"
    if: "branch == 'main'"
    use: "claude-opus"

  - id: "development"
    if: "branch.matches('dev|feature')"
    use: "claude-haiku"  # 80% cheaper

  - id: "test"
    if: "project_type contains 'test'"
    use: "gemini-flash"  # Cheapest
```

### Case 3: Multi-Model Comparison - A/B Testing

**Problem**: Want to evaluate different models on real workloads.

**Solution**:
```bash
# Enable exploration mode (3% random routing)
$ boba init --explore-rate 0.03

# After 7 days, view analysis
$ boba stats --7d --by-profile

By Profile:
- openai-gpt4: avg_latency=1200ms cost=$6.20 usage=70%
- claude-sonnet: avg_latency=980ms cost=$1.80 usage=27%
- gemini-flash: avg_latency=650ms cost=$0.76 usage=3% (explore)

# View optimization suggestions
$ boba action

💡 Suggestion: Switch to claude-sonnet for 40% cost reduction
   Impact: -$30/month, <5% quality difference
   Command: boba use claude-sonnet
```

---

## Advanced Features

### Git Hooks Integration

Auto-track AI calls during commits:

```bash
# Install hooks
$ boba hooks install

# Auto-record AI usage on each commit
$ git commit -m "feat: add authentication"
[BobaMixer] Tracked: 3 AI calls, 12K tokens, $0.34
```

### Suggestion Engine

Generate optimization suggestions based on historical data:

```bash
$ boba action

💡 High-priority suggestions:
  1. [COST] Switch 'openai-gpt4' to 'claude-sonnet' for code tasks
     → Save $45/month (current: $120/mo → projected: $75/mo)

  2. [PERF] Enable caching for repetitive queries
     → Reduce latency by 60% (avg: 1200ms → 480ms)

  3. [BUDGET] Daily spending on track to exceed monthly cap
     → Action needed: Reduce usage or increase cap

# Auto-apply high-priority suggestions
$ boba action --auto
```

---

## Command Reference

```bash
# Control Plane
boba providers                           # List all providers
boba tools                               # List local CLI tools
boba bind <tool> <provider>              # Create binding
boba run <tool> [args...]                # Run tool

# HTTP Proxy
boba proxy serve                         # Start proxy
boba proxy status                        # Check status

# Usage & Statistics
boba stats [--today|--7d|--30d]         # View statistics
boba report --format json --out file     # Export report

# Budget Management
boba budget --status                     # View budget
boba budget --daily 10 --cap 300        # Set limits

# Routing
boba route test "Your prompt here"       # Test routing
boba route test @prompt.txt              # Test from file

# Optimization
boba action                              # View suggestions
boba action --auto                       # Auto-apply

# Configuration
boba init                                # Initialize config
boba edit <profiles|routes|pricing|secrets>
boba doctor                              # Health check

# Advanced
boba hooks install                       # Install Git hooks
boba completions install --shell bash    # Shell completion
```

---

## Config File Structure

```
~/.boba/
├── providers.yaml      # AI service provider configs
├── tools.yaml          # Local CLI tools
├── bindings.yaml       # Tool ↔ Provider bindings
├── secrets.yaml        # API keys (permissions: 0600)
├── routes.yaml         # Routing rules
├── pricing.yaml        # Pricing configuration
├── settings.yaml       # UI preferences
├── usage.db            # SQLite database
└── logs/               # Structured logs
```

---

## Developer Guide

### Build

```bash
# Clone repository
git clone https://github.com/royisme/BobaMixer.git
cd BobaMixer

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Lint check
make lint
```

### Requirements

- Go 1.25+ (set `GOTOOLCHAIN=auto` for auto-download)
- SQLite 3
- golangci-lint v1.60.1

```bash
# Ensure Go auto-fetches matching compiler
export GOTOOLCHAIN=auto

# Install golangci-lint locally (./bin)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b ./bin v1.60.1
```

### Code Standards

Project follows strict Go language standards:
- All exported types and functions must have doc comments
- Use `golangci-lint` for static analysis
- Follow [Effective Go](https://go.dev/doc/effective_go) guide
- Run `make test && make lint` before commits

---

## Contributing

We welcome all forms of contributions!

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Submit Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## Roadmap

- [x] Phase 1: Control Plane (Provider/Tool/Binding management) - **100% Complete** ✅
- [x] Phase 2: HTTP Proxy & Usage monitoring - **100% Complete** ✅
- [x] Phase 3: Intelligent routing & Budget control & Pricing auto-fetch - **100% Complete** ✅
- [ ] Phase 4: Web Dashboard (Optional feature, TUI is already powerful)
- [ ] Phase 5: Multi-user collaboration (Enterprise features)

**🎉 Current Status**: All core features fully implemented, project at **100% completion**!

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for TUI
- Pricing data powered by [OpenRouter](https://openrouter.ai/)
- Inspired by microservice orchestration and API gateway design

---

## Contact

- **Issues**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)
- **Documentation**: [Full Docs](https://royisme.github.io/BobaMixer/)

---

<div align="center">

**Reduce your AI costs by 50% in the time it takes to make a boba tea ☕🧋**

Made with ❤️ by developers, for developers

</div>
