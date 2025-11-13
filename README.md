# BobaMixer 🧋

> **Smart AI Adapter Router with Cost Tracking and Intelligent Routing**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/royisme/BobaMixer)](https://github.com/royisme/BobaMixer/releases)
[![Documentation](https://img.shields.io/badge/docs-online-blue)](https://royisme.github.io/BobaMixer/)

BobaMixer is a comprehensive CLI tool for managing multiple AI providers, tracking costs, and optimizing your AI workload routing. It features intelligent routing, real-time budget tracking, and comprehensive usage analytics.

[**📚 Full Documentation**](https://royisme.github.io/BobaMixer/) | [**🚀 Quick Start**](https://royisme.github.io/BobaMixer/guide/getting-started) | [**中文文档**](https://royisme.github.io/BobaMixer/zh/)

## ✨ Features

- **🧠 Intelligent Routing** - Context-aware routing with epsilon-greedy exploration
- **📊 Usage Tracking** - Real-time monitoring of tokens, cost, and latency
- **💰 Budget Management** - Multi-level budgets with proactive alerts
- **🤖 Smart Suggestions** - AI-powered cost optimization recommendations
- **🛠️ Flexible Adapters** - Support for HTTP APIs, CLI tools, and MCP servers
- **📈 TUI Dashboard** - Beautiful terminal interface for monitoring

## 🚀 Quick Start

### Installation

```bash
# Using Go
go install github.com/royisme/bobamixer/cmd/boba@latest

# Using Homebrew (macOS/Linux)
brew tap royisme/tap
brew install bobamixer
```

### Initial Setup

```bash
# Initialize configuration
boba doctor

# Configure your first profile (edit ~/.boba/profiles.yaml)
# Add API keys (edit ~/.boba/secrets.yaml)

# Activate profile
boba use default

# Launch TUI dashboard
boba
```

## 📖 Documentation

**Complete documentation available at: https://royisme.github.io/BobaMixer/**

Quick links:
- **[Getting Started](https://royisme.github.io/BobaMixer/guide/getting-started)** - Installation and setup
- **[Configuration](https://royisme.github.io/BobaMixer/guide/configuration)** - Complete configuration reference
- **[Adapters](https://royisme.github.io/BobaMixer/features/adapters)** - HTTP, Tool, and MCP adapters
- **[Routing](https://royisme.github.io/BobaMixer/features/routing)** - Intelligent routing rules
- **[CLI Reference](https://royisme.github.io/BobaMixer/reference/cli)** - All commands and options
- **[Troubleshooting](https://royisme.github.io/BobaMixer/advanced/troubleshooting)** - Common issues and solutions

中文文档: https://royisme.github.io/BobaMixer/zh/

## 🎮 Usage Examples

```bash
# View profiles
boba ls --profiles

# Track usage
boba stats --today
boba stats --7d --by-profile

# Route testing
boba route test "Write a function to sort an array"

# Budget management
boba budget --status

# Generate reports
boba report --format json --output usage-report.json
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

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linter
5. Submit a pull request

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI
- Inspired by cost optimization needs in AI development

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)
- **Documentation**: [Full Docs](https://royisme.github.io/BobaMixer/)

---

**Made with ☕ and 🧋 for developers, by developers**
