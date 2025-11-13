# BobaMixer 🧋

**AI/LLM Usage Tracking and Cost Optimization Tool**

BobaMixer is a comprehensive CLI and TUI tool for tracking, analyzing, and optimizing your AI/LLM API usage and costs. It supports multiple providers, intelligent routing, budget alerts, and actionable insights to help you make data-driven decisions about your AI infrastructure.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/royisme/BobaMixer)](https://github.com/royisme/BobaMixer/releases)

## ✨ Features

### 📊 Usage Tracking & Analytics
- **Real-time monitoring** of tokens, cost, and latency
- **Multi-provider support** (Anthropic, OpenAI, OpenRouter, custom)
- **Historical trends** with 7/30-day analysis
- **Session tracking** with project/branch context
- **Estimate accuracy levels** (exact, mapped, heuristic)

### 🎯 Intelligent Routing
- **Rule-based routing** with DSL expressions
- **Context-aware** (text patterns, size, project type, branch, time)
- **Epsilon-greedy exploration** for automatic optimization
- **Offline testing** with `boba route test`

### 💰 Budget Management
- **Multi-level budgets** (global, project, profile)
- **Proactive alerts** (warning and critical thresholds)
- **Cost projections** and spending trends
- **No blocking** - alerts only, never interrupts workflow

### 🤖 Suggestion Engine
- **Cost optimization recommendations** based on usage patterns
- **Profile switching suggestions** with confidence scores
- **P95 latency comparisons**
- **Auto-apply** or manual review options

### 🛠️ Adapters
- **HTTP Adapter**: REST API providers (Anthropic, OpenAI, etc.)
- **Tool Adapter**: CLI tools (claude-code, custom scripts)
- **MCP Adapter**: Model Context Protocol integrations
- **Extensible**: Easy to add custom adapters

### 📈 TUI Dashboard
- **Beautiful interface** with bubble tea
- **Real-time stats** and trend visualizations
- **Profile switching** and budget status
- **Notification feed** for alerts and suggestions

## 🚀 Quick Start

### Installation

**Using Go:**
```bash
go install github.com/royisme/bobamixer/cmd/boba@latest
```

**Using Homebrew (macOS/Linux):**
```bash
brew tap royisme/tap
brew install bobamixer
```

**Download Binary:**
Download from [Releases](https://github.com/royisme/BobaMixer/releases)

### Initial Setup

1. **Initialize configuration:**
```bash
boba doctor
```

This creates `~/.boba/` with example configurations.

2. **Configure your first profile** in `~/.boba/profiles.yaml`:
```yaml
default:
  adapter: http
  provider: anthropic
  endpoint: https://api.anthropic.com/v1/messages
  model: claude-3-5-sonnet-20241022
  headers:
    anthropic-version: "2023-06-01"
    x-api-key: "secret://anthropic_key"
```

3. **Add your API key** to `~/.boba/secrets.yaml`:
```yaml
anthropic_key: sk-ant-your-key-here
```

4. **Activate the profile:**
```bash
boba use default
```

5. **Launch TUI dashboard:**
```bash
boba
```

## 📖 Documentation

- **[Quick Reference](docs/QUICK_REFERENCE.md)** - Command cheat sheet
- **[Adapter Guide](docs/ADAPTERS.md)** - Building custom adapters
- **[Routing Cookbook](docs/ROUTING_COOKBOOK.md)** - Routing examples
- **[Operations Guide](docs/OPERATIONS.md)** - Backup, cleanup, maintenance
- **[FAQ](docs/FAQ.md)** - Common questions

## 🎮 Usage Examples

### View Profiles
```bash
# List all configured profiles
boba ls --profiles

# Activate a profile
boba use fast-model
```

### Track Usage
```bash
# Today's stats
boba stats --today

# Last 7 days
boba stats --7d

# Breakdown by profile
boba stats --7d --by-profile
```

### Route Testing
```bash
# Test routing with text
boba route test "Write a function to sort an array"

# Test with file content
boba route test @prompt.txt
```

### Budget Management
```bash
# Check budget status
boba budget --status

# View alerts
boba action
```

### Generate Reports
```bash
# Export to JSON
boba report --format json --output usage-report.json

# Export to CSV
boba report --format csv --output usage-report.csv
```

### Git Hooks
```bash
# Install git hooks for project context
cd your-project
boba hooks install

# Remove hooks
boba hooks remove
```

## ⚙️ Configuration

### Directory Structure
```
~/.boba/
├── profiles.yaml    # Profile definitions
├── routes.yaml      # Routing rules
├── pricing.yaml     # Model pricing
├── secrets.yaml     # API keys (0600 permissions)
├── usage.db         # SQLite database
├── logs/            # Application logs
└── pricing.cache.json  # Cached pricing data
```

### Project-Level Config
Create `.boba-project.yaml` in your repo root:
```yaml
project:
  name: my-app
  type: [typescript, react]
  preferred_profiles:
    - fast-model
    - cost-optimized

budget:
  daily_usd: 5.00
  hard_cap: 100.00
```

### Routing Rules
Define in `~/.boba/routes.yaml`:
```yaml
rules:
  - id: large-context
    if: "ctx_chars > 50000"
    use: high-capacity
    explain: "Large context requires high-capacity model"

  - id: code-format
    if: "text.matches('format|prettier|lint')"
    use: fast-model
    explain: "Simple formatting task"

  - id: night-mode
    if: "time_of_day == 'night'"
    use: cost-optimized
    explain: "Off-peak hours, use cheaper model"
```

## 🧪 Development

### Prerequisites
- Go 1.22+
- SQLite 3
- Git

### Build from Source
```bash
git clone https://github.com/royisme/BobaMixer.git
cd BobaMixer
make build
```

### Run Tests
```bash
make test
```

### Run Linter
```bash
make lint
```

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linter
5. Submit a pull request

## 📋 Roadmap

- [x] Phase 0-4: Core infrastructure
- [x] Phase 5: Routing DSL and suggestions
- [ ] Phase 6: Remote pricing sources
- [ ] Phase 7: Enhanced TUI visualizations
- [ ] Phase 8: Shell completion and advanced git integration
- [x] Phase 9: Release and documentation

See [docs/roadmap.md](docs/roadmap.md) for details.

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI
- Inspired by cost optimization needs in AI development
- Community feedback and contributions

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)
- **Documentation**: [docs/](docs/)

---

**Made with ☕ and 🧋 by developers, for developers**
