# BobaMixer 🧋

> **Intelligent Router & Cost Optimizer for AI Workflows**
> **AI工作流的智能路由器与成本优化引擎**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/royisme/BobaMixer)](https://github.com/royisme/BobaMixer/releases)
[![golangci-lint](https://img.shields.io/badge/lint-passing-brightgreen)](https://golangci-lint.run/)

[**📚 English Docs**](https://royisme.github.io/BobaMixer/) | [**🚀 Quick Start**](#quick-start) | [**中文文档**](https://royisme.github.io/BobaMixer/zh/)

---

## Why BobaMixer?

In daily AI development, have you encountered these pain points?

**在AI开发的日常工作中,你是否遇到过这些痛点:**

- 🔑 **API Key Chaos** - Multiple AI service credentials scattered everywhere, switching providers requires config file changes
  **密钥管理混乱** - 多个AI服务的API密钥散落在各处,切换provider需要修改配置文件

- 💸 **Runaway Costs** - API bills skyrocket without warning, no real-time monitoring or budget control
  **成本失控** - 不知不觉中API调用费用飙升,缺乏实时监控和预算控制

- 🎯 **Routing Decisions** - Which model for which task? How to balance cost vs quality?
  **路由决策困难** - 不同任务应该用哪个模型?如何在成本和效果之间平衡?

- 📊 **Missing Usage Data** - Cannot track token consumption or cost distribution, lack of optimization insights
  **使用数据缺失** - 无法追踪token消耗、成本分布,缺乏优化依据

- 🔄 **High Switching Cost** - Moving from Claude to OpenAI requires code changes, no flexible orchestration
  **切换成本高** - 从Claude切到OpenAI需要修改代码,无法灵活调度

**BobaMixer was born to solve these problems** — It's your AI workflow control plane, letting you orchestrate AI models like microservices.

**BobaMixer 就是为解决这些问题而生的** —— 它是你的AI工作流控制平面,让你像调度微服务一样调度AI模型。

---

## Core Capabilities | 核心能力

### 1. Unified Control Plane | 统一控制平面

No more hardcoded API keys and endpoints in code - everything is configuration-driven:

不再需要在代码中硬编码API密钥和endpoint,一切配置化:

```bash
# View all available AI providers | 查看所有可用的AI provider
$ boba providers

Provider              Kind        Endpoint                      Status
claude-anthropic      anthropic   https://api.anthropic.com    ✓ Ready
claude-zai            anthropic   https://api.z.ai/api/...      ✓ Ready
openai-official       openai      https://api.openai.com        ✓ Ready
gemini-official       gemini      https://generativelanguage... ✓ Ready

# Bind local CLI tool to provider | 绑定本地CLI工具到provider
$ boba bind claude claude-zai

# Auto-inject config at runtime | 运行时自动注入配置
$ boba run claude "Write a function to calculate fibonacci"
```

**Core Value**: Decoupled configuration from code, configure once, apply globally.

**核心价值**: 配置与代码解耦,一次配置,全局生效。

### 2. Local HTTP Proxy | 本地HTTP Proxy (流量拦截与监控)

Start an intelligent proxy locally to intercept all AI API calls:

在你的本地启动一个智能代理,拦截所有AI API调用:

```bash
# Start proxy server (127.0.0.1:7777) | 启动代理服务器
$ boba proxy serve &

# All requests through proxy are automatically logged
# 所有经过proxy的请求都会被自动记录
# Supports both OpenAI and Anthropic API formats
# 支持 OpenAI 和 Anthropic 两种API格式
```

**Technical Highlights**:
- **Zero-intrusion integration** - Just modify `ANTHROPIC_BASE_URL` env var
  零侵入式集成 - 只需修改环境变量
- **Automatic token parsing** - Extract precise input/output tokens from responses
  自动Token解析 - 从响应中提取精确的tokens
- **Real-time cost calculation** - Calculate per-request cost based on latest pricing
  实时成本计算 - 基于最新定价表计算成本
- **Thread-safe** - Concurrent request support with `sync.RWMutex`
  线程安全 - 使用sync.RWMutex保护共享状态

### 3. Intelligent Routing Engine | 智能路由引擎 (Context-Aware)

Automatically select optimal model based on task characteristics:

根据任务特征自动选择最优模型:

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

Test routing decisions | 测试路由决策:

```bash
$ boba route test "Please review this PR and check for security issues"

=== Routing Decision ===
Profile: openai-gpt4
Rule ID: code-review
Explanation: Code review tasks use GPT-4 for best results
Fallback: claude-anthropic
```

**Core Algorithm**: Epsilon-Greedy exploration + Rule engine, auto-balancing between cost optimization and quality exploration.

**核心算法**: Epsilon-Greedy探索 + 规则引擎,在成本优化和效果探索之间自动平衡。

### 4. Budget Management & Alerts | 预算管理与告警

Multi-level budget control to prevent cost overruns:

多层级预算控制,防止成本失控:

```bash
# View current budget status | 查看当前预算状态
$ boba budget --status

Budget Scope: project (my-chatbot)
========================================
Today:  $2.34 of $10.00 (23.4%)
Period: $45.67 of $300.00 (15.2%)
Days Remaining: 23

# Set budget limits | 设置预算限制
$ boba budget --daily 10.00 --cap 300.00

# Auto-switch to cheaper provider when over budget
# 超预算时自动切换到更便宜的provider
$ boba action --auto
```

**Technical Implementation**:
- Pre-request budget check (`checkBudgetBeforeRequest`)
- Conservative token estimation (1000 input, 500 output)
- HTTP 429 response when budget exceeded
- Graceful degradation - allows pass-through without budget config

**技术实现**: 请求前预算检查、保守Token估算、HTTP 429响应、优雅降级

### 5. Usage Analytics & Cost Tracking | 使用分析与成本追踪

Precise token-level tracking with multi-dimensional analysis:

精确的Token级别追踪和多维度分析:

```bash
# View today's stats | 查看今日统计
$ boba stats --today

Today's Usage
=============
Tokens:   45,678
Cost:     $1.23
Sessions: 12

# 7-day trend analysis | 7天趋势分析
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

# Export report | 导出报告
$ boba report --format json --output monthly-report.json
```

**Data Schema**:
- `sessions` table - Records session metadata (project, branch, profile, latency)
- `usage_records` table - Precise token & cost records, 3 estimation levels (exact/mapped/heuristic)
- SQLite storage - Local, no external database dependency

### 6. Real-time Pricing Updates | 实时定价更新 (Pricing Auto-Refresh)

Auto-fetch latest model pricing from OpenRouter API:

从OpenRouter API自动获取最新模型定价:

```bash
# Configure pricing refresh strategy
# 配置定价刷新策略
# ~/.boba/pricing.yaml
refresh:
  interval_hours: 24
  on_startup: false

# Manually verify pricing data | 手动验证定价数据
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

**加载策略** (多层Fallback): OpenRouter API → 本地缓存 → Vendor JSON → pricing.yaml → profiles.yaml

---

## Technical Architecture | 技术架构

### Modular Design | 模块化设计

```
BobaMixer
├── cmd/boba              # CLI entry point | CLI入口
├── internal/cli          # Command implementations | 命令实现
├── internal/domain       # Core domain logic | 核心领域逻辑
│   ├── budget           # Budget tracking | 预算追踪
│   ├── pricing          # Pricing mgmt (OpenRouter) | 定价管理
│   ├── routing          # Routing engine | 路由引擎
│   ├── stats            # Statistical analysis | 统计分析
│   └── suggestions      # Optimization suggestions | 优化建议
├── internal/proxy        # HTTP proxy server | HTTP代理服务器
├── internal/store        # Data storage | 数据存储
│   ├── config           # Config loading | 配置加载
│   └── sqlite           # SQLite operations | SQLite操作
└── internal/ui           # TUI Dashboard (Bubble Tea)
```

### Key Tech Stack | 关键技术选型

- **Language**: Go 1.22+ (Type-safe, concurrency-friendly, single-binary deployment)
  **语言**: Go 1.22+ (类型安全, 并发友好, 单文件部署)
- **TUI**: Bubble Tea (Modern terminal UI framework)
  **TUI**: Bubble Tea (现代化终端UI框架)
- **Storage**: SQLite (Zero-config, local, SQL analytics support)
  **存储**: SQLite (零配置, 本地化, 支持SQL分析)
- **Linting**: golangci-lint (Strict code quality standards)
  **Lint**: golangci-lint (严格代码质量标准)
- **API Integration**: OpenRouter Models API (1000+ model pricing)
  **API集成**: OpenRouter Models API (1000+ 模型定价)

### Go Best Practices | Go最佳实践

Project strictly follows Go language standards:

项目严格遵循Go语言规范:

- ✅ **golangci-lint verified** - 0 issues
- ✅ **Documentation** - All exported types/functions have doc comments
  文档注释 - 所有导出类型/函数都有规范注释
- ✅ **Error handling** - Complete error wrapping & graceful degradation
  错误处理 - 完整的error wrapping和优雅降级
- ✅ **Concurrency safety** - `sync.RWMutex` protects shared state
  并发安全 - 使用sync.RWMutex保护共享状态
- ✅ **Security** - All exceptions marked with `#nosec` after audit
  安全编码 - 通过#nosec标记审计所有例外

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

### Initialize Configuration | 初始化配置

```bash
# Initialize config files | 初始化配置文件
$ boba init

✅ BobaMixer initialized successfully

Configuration directory: ~/.boba

Created files:
  - providers.yaml  (AI service providers)
  - tools.yaml      (Local CLI tools)
  - bindings.yaml   (Tool ↔ Provider bindings)
  - secrets.yaml    (API keys)
  - settings.yaml   (UI preferences)

Next steps:
  1. Add your API keys to environment variables or secrets.yaml
  2. Run 'boba tools' to see detected CLI tools
  3. Run 'boba providers' to see available providers
  4. Run 'boba bind <tool> <provider>' to create bindings
  5. Run 'boba doctor' to verify your setup
```

### Configure API Keys | 配置API密钥

```bash
# Method 1: Environment variables (Recommended)
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."

# Method 2: secrets.yaml
$ boba edit secrets
```

```yaml
# ~/.boba/secrets.yaml
secrets:
  anthropic-key: "sk-ant-..."
  openai-key: "sk-..."
  gemini-key: "..."
```

### Launch TUI Dashboard | 启动TUI Dashboard

```bash
$ boba
```

**Interactive controls** | **交互式操作**:
- `↑/↓` Select tool | 选择工具
- `B` Switch Provider binding | 切换Provider绑定
- `X` Toggle Proxy | 切换Proxy开关
- `V` Stats view | 统计视图
- `R` Run tool | 运行工具
- `q` Quit | 退出

---

## Use Cases | 使用场景

### Case 1: Team Collaboration - Unified API Management
### 案例1: 团队协作 - 统一API管理

**Problem**: Team members manage API keys separately, easy to leak and hard to audit.
**问题**: 团队成员各自管理API密钥,容易泄露且难以审计。

**Solution**:
```bash
# 1. Create .boba-project.yaml in project root
# 1. 在项目根目录创建 .boba-project.yaml
$ cat .boba-project.yaml
project:
  name: "my-chatbot"
  type: ["backend", "ai"]
  preferred_profiles: ["claude-anthropic", "openai-gpt4"]

budget:
  daily_usd: 20.0
  hard_cap: 600.0

# 2. Each team member configures ~/.boba/secrets.yaml
# 2. 团队成员各自配置 ~/.boba/secrets.yaml

# 3. Project-level budget auto-applies
# 3. 项目级预算自动生效
$ cd my-chatbot
$ boba budget --status  # Auto-detects project budget
```

### Case 2: Cost Optimization - Auto Downgrade
### 案例2: 成本优化 - 自动降级

**Problem**: Development uses expensive models, costs skyrocket during testing.
**问题**: 开发环境使用昂贵模型,测试时成本飙升。

**Solution**:
```yaml
# routes.yaml - Auto-select model based on branch
# routes.yaml - 根据分支自动选择模型
rules:
  - id: "production"
    if: "branch == 'main'"
    use: "claude-opus"

  - id: "development"
    if: "branch.matches('dev|feature')"
    use: "claude-haiku"  # 80% cheaper | 便宜80%

  - id: "test"
    if: "project_type contains 'test'"
    use: "gemini-flash"  # Cheapest | 最便宜
```

### Case 3: Multi-Model Comparison - A/B Testing
### 案例3: 多模型对比 - A/B测试

**Problem**: Want to evaluate different models on real workloads.
**问题**: 想评估不同模型在真实工作负载下的效果。

**Solution**:
```bash
# Enable exploration mode (3% random routing)
# 开启探索模式(3%流量随机路由)
$ boba init --explore-rate 0.03

# After 7 days, view analysis | 7天后查看分析
$ boba stats --7d --by-profile

By Profile:
- openai-gpt4: avg_latency=1200ms cost=$6.20 usage=70%
- claude-sonnet: avg_latency=980ms cost=$1.80 usage=27%
- gemini-flash: avg_latency=650ms cost=$0.76 usage=3% (explore)

# View optimization suggestions | 查看优化建议
$ boba action

💡 Suggestion: Switch to claude-sonnet for 40% cost reduction
   Impact: -$30/month, <5% quality difference
   Command: boba use claude-sonnet
```

---

## Advanced Features | 高级功能

### Git Hooks Integration | Git Hooks集成

Auto-track AI calls during commits:

在commit过程中自动追踪AI调用:

```bash
# Install hooks
$ boba hooks install

# Auto-record AI usage on each commit
# 自动记录每次commit时的AI使用
$ git commit -m "feat: add authentication"
[BobaMixer] Tracked: 3 AI calls, 12K tokens, $0.34
```

### Suggestion Engine | 建议引擎

Generate optimization suggestions based on historical data:

基于历史数据生成优化建议:

```bash
$ boba action

💡 High-priority suggestions:
  1. [COST] Switch 'openai-gpt4' to 'claude-sonnet' for code tasks
     → Save $45/month (current: $120/mo → projected: $75/mo)

  2. [PERF] Enable caching for repetitive queries
     → Reduce latency by 60% (avg: 1200ms → 480ms)

  3. [BUDGET] Daily spending on track to exceed monthly cap
     → Action needed: Reduce usage or increase cap

# Auto-apply high-priority suggestions | 自动应用高优先级建议
$ boba action --auto
```

---

## Command Reference | 命令参考

```bash
# Control Plane | 控制平面
boba providers                           # List all providers
boba tools                               # List local CLI tools
boba bind <tool> <provider>              # Create binding
boba run <tool> [args...]                # Run tool

# HTTP Proxy
boba proxy serve                         # Start proxy
boba proxy status                        # Check status

# Usage & Statistics | 使用统计
boba stats [--today|--7d|--30d]         # View statistics
boba report --format json --out file     # Export report

# Budget Management | 预算管理
boba budget --status                     # View budget
boba budget --daily 10 --cap 300        # Set limits

# Routing | 路由
boba route test "Your prompt here"       # Test routing
boba route test @prompt.txt              # Test from file

# Optimization | 优化
boba action                              # View suggestions
boba action --auto                       # Auto-apply

# Configuration | 配置
boba init                                # Initialize config
boba edit <profiles|routes|pricing|secrets>
boba doctor                              # Health check

# Advanced | 高级
boba hooks install                       # Install Git hooks
boba completions install --shell bash    # Shell completion
```

---

## Config File Structure | 配置文件结构

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

## Developer Guide | 开发者指南

### Build | 构建

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

### Requirements | 环境要求

- Go 1.22+ (set `GOTOOLCHAIN=auto` for auto-download)
- SQLite 3
- golangci-lint v1.60.1

```bash
# Ensure Go auto-fetches matching compiler
export GOTOOLCHAIN=auto

# Install golangci-lint locally (./bin)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b ./bin v1.60.1
```

### Code Standards | 代码规范

Project follows strict Go language standards:

项目遵循严格的Go语言规范:
- All exported types and functions must have doc comments
  所有导出类型和函数必须有文档注释
- Use `golangci-lint` for static analysis
  使用golangci-lint进行静态分析
- Follow [Effective Go](https://go.dev/doc/effective_go) guide
  遵循Effective Go指南
- Run `make test && make lint` before commits
  提交前运行make test && make lint

---

## Contributing | 贡献指南

We welcome all forms of contributions!

我们欢迎所有形式的贡献!

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Submit Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## Roadmap | 路线图

- [x] Phase 1: Control Plane (Provider/Tool/Binding management) - **100% Complete** ✅
- [x] Phase 2: HTTP Proxy & Usage monitoring - **100% Complete** ✅
- [x] Phase 3: Intelligent routing & Budget control & Pricing auto-fetch - **100% Complete** ✅
- [ ] Phase 4: Web Dashboard (Optional feature, TUI is already powerful)
- [ ] Phase 5: Multi-user collaboration (Enterprise features)

**🎉 Current Status**: All core features fully implemented, project at **100% completion**!

**🎉 当前状态**: 所有核心功能已完整实现,项目达到 **100% 完成度**！

---

## License | 开源协议

MIT License - See [LICENSE](LICENSE) file for details.

---

## Acknowledgments | 致谢

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for TUI
- Pricing data powered by [OpenRouter](https://openrouter.ai/)
- Inspired by microservice orchestration and API gateway design

---

## Contact | 联系方式

- **Issues**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)
- **Documentation**: [Full Docs](https://royisme.github.io/BobaMixer/)

---

<div align="center">

**Reduce your AI costs by 50% in the time it takes to make a boba tea ☕🧋**

**用一杯珍珠奶茶的时间,让AI成本降低50%**

Made with ❤️ by developers, for developers

</div>
