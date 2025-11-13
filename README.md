# BobaMixer 🧋

> **[English](#english)** | **[中文](#中文)**

---

## English

**Smart AI Adapter Router with Cost Tracking and Intelligent Routing**

BobaMixer is a comprehensive CLI tool for managing multiple AI providers, tracking costs, and optimizing your AI workload routing. It features intelligent routing, real-time budget tracking, and comprehensive usage analytics.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/royisme/BobaMixer)](https://github.com/royisme/BobaMixer/releases)
[![Documentation](https://img.shields.io/badge/docs-online-blue)](https://royisme.github.io/BobaMixer/)

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

**📚 [Full Documentation](https://royisme.github.io/BobaMixer/)** - Complete guides in English and Chinese

Quick Links:
- **[Getting Started](https://royisme.github.io/BobaMixer/docs/getting-started/)** - Installation and first steps
- **[Configuration](https://royisme.github.io/BobaMixer/docs/configuration/)** - Complete configuration reference
- **[Adapters](https://royisme.github.io/BobaMixer/docs/adapters/)** - Working with different adapter types
- **[Routing](https://royisme.github.io/BobaMixer/docs/routing/)** - Routing rules and optimization
- **[Troubleshooting](https://royisme.github.io/BobaMixer/docs/troubleshooting/)** - Common issues and solutions

Legacy Docs:
- [Adapter Guide](docs/ADAPTERS.md) | [Routing Cookbook](docs/ROUTING_COOKBOOK.md) | [Operations](docs/OPERATIONS.md) | [FAQ](docs/FAQ.md)

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

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI
- Inspired by cost optimization needs in AI development
- Community feedback and contributions

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)
- **Documentation**: [Full Docs](https://royisme.github.io/BobaMixer/)

---

## 中文

**智能AI适配器路由器，具备成本追踪和智能路由功能**

BobaMixer 是一款全面的命令行工具，用于管理多个 AI 提供商、追踪成本并优化您的 AI 工作负载路由。它具有智能路由、实时预算追踪和全面的使用分析功能。

## ✨ 核心功能

### 📊 使用追踪与分析
- **实时监控** 令牌、成本和延迟
- **多提供商支持**（Anthropic、OpenAI、OpenRouter、自定义）
- **历史趋势** 支持 7/30 天分析
- **会话追踪** 包含项目/分支上下文
- **估算精度级别**（精确、映射、启发式）

### 🎯 智能路由
- **基于规则的路由** 支持 DSL 表达式
- **上下文感知**（文本模式、大小、项目类型、分支、时间）
- **Epsilon-greedy 探索** 实现自动优化
- **离线测试** 使用 `boba route test`

### 💰 预算管理
- **多级预算**（全局、项目、配置文件）
- **主动警报**（警告和关键阈值）
- **成本预测** 和支出趋势
- **不阻断** - 仅警报，从不中断工作流程

### 🤖 建议引擎
- 基于使用模式的**成本优化建议**
- 带置信度分数的**配置文件切换建议**
- **P95 延迟比较**
- **自动应用** 或手动审核选项

### 🛠️ 适配器
- **HTTP 适配器**：REST API 提供商（Anthropic、OpenAI 等）
- **Tool 适配器**：CLI 工具（claude-code、自定义脚本）
- **MCP 适配器**：模型上下文协议集成
- **可扩展**：易于添加自定义适配器

### 📈 TUI 仪表板
- 使用 bubble tea 的**漂亮界面**
- **实时统计** 和趋势可视化
- **配置文件切换** 和预算状态
- **通知流** 用于警报和建议

## 🚀 快速开始

### 安装

**使用 Go：**
```bash
go install github.com/royisme/bobamixer/cmd/boba@latest
```

**使用 Homebrew（macOS/Linux）：**
```bash
brew tap royisme/tap
brew install bobamixer
```

**下载二进制文件：**
从 [Releases](https://github.com/royisme/BobaMixer/releases) 下载

### 初始设置

1. **初始化配置：**
```bash
boba doctor
```

这会在 `~/.boba/` 中创建示例配置。

2. **在 `~/.boba/profiles.yaml` 中配置您的第一个配置文件：**
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

3. **将您的 API 密钥添加到 `~/.boba/secrets.yaml`：**
```yaml
anthropic_key: sk-ant-your-key-here
```

4. **激活配置文件：**
```bash
boba use default
```

5. **启动 TUI 仪表板：**
```bash
boba
```

## 📖 文档

**📚 [完整文档](https://royisme.github.io/BobaMixer/)** - 中英文完整指南

快速链接：
- **[快速入门](https://royisme.github.io/BobaMixer/zh/docs/getting-started/)** - 安装和第一步
- **[配置](https://royisme.github.io/BobaMixer/zh/docs/configuration/)** - 完整配置参考
- **[适配器](https://royisme.github.io/BobaMixer/zh/docs/adapters/)** - 使用不同的适配器类型
- **[路由](https://royisme.github.io/BobaMixer/zh/docs/routing/)** - 路由规则和优化
- **[故障排除](https://royisme.github.io/BobaMixer/zh/docs/troubleshooting/)** - 常见问题和解决方案

## 🎮 使用示例

### 查看配置文件
```bash
# 列出所有配置的配置文件
boba ls --profiles

# 激活一个配置文件
boba use fast-model
```

### 追踪使用情况
```bash
# 今天的统计
boba stats --today

# 最近 7 天
boba stats --7d

# 按配置文件细分
boba stats --7d --by-profile
```

### 路由测试
```bash
# 使用文本测试路由
boba route test "编写一个排序数组的函数"

# 使用文件内容测试
boba route test @prompt.txt
```

### 预算管理
```bash
# 检查预算状态
boba budget --status

# 查看警报
boba action
```

## ⚙️ 配置

### 目录结构
```
~/.boba/
├── profiles.yaml    # 配置文件定义
├── routes.yaml      # 路由规则
├── pricing.yaml     # 模型定价
├── secrets.yaml     # API 密钥（0600 权限）
├── usage.db         # SQLite 数据库
├── logs/            # 应用程序日志
└── pricing.cache.json  # 缓存的定价数据
```

### 项目级配置
在仓库根目录创建 `.boba-project.yaml`：
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

## 🧪 开发

### 前提条件
- Go 1.22+
- SQLite 3
- Git

### 从源代码构建
```bash
git clone https://github.com/royisme/BobaMixer.git
cd BobaMixer
make build
```

### 运行测试
```bash
make test
```

### 运行 Linter
```bash
make lint
```

## 🤝 贡献

欢迎贡献！请参阅 [CONTRIBUTING.md](CONTRIBUTING.md) 了解指南。

1. Fork 仓库
2. 创建功能分支
3. 进行更改
4. 运行测试和 linter
5. 提交 pull request

## 📜 许可证

MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- 使用 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 构建 TUI
- 受 AI 开发中成本优化需求的启发
- 社区反馈和贡献

## 📞 支持

- **问题**：[GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **讨论**：[GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)
- **文档**：[完整文档](https://royisme.github.io/BobaMixer/)

---

**Made with ☕ and 🧋 by developers, for developers**
