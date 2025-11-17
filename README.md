# BobaMixer 🧋

> **AI工作流的智能路由器与成本优化引擎**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/royisme/BobaMixer)](https://github.com/royisme/BobaMixer/releases)
[![golangci-lint](https://img.shields.io/badge/lint-passing-brightgreen)](https://golangci-lint.run/)

## 为什么需要 BobaMixer?

在AI开发的日常工作中,你是否遇到过这些痛点:

- 🔑 **密钥管理混乱** - 多个AI服务的API密钥散落在各处,切换provider需要修改配置文件
- 💸 **成本失控** - 不知不觉中API调用费用飙升,缺乏实时监控和预算控制
- 🎯 **路由决策困难** - 不同任务应该用哪个模型?如何在成本和效果之间平衡?
- 📊 **使用数据缺失** - 无法追踪token消耗、成本分布,缺乏优化依据
- 🔄 **切换成本高** - 从Claude切到OpenAI需要修改代码,无法灵活调度

**BobaMixer** 就是为解决这些问题而生的 —— 它是你的AI工作流控制平面,让你像调度微服务一样调度AI模型。

## 核心能力

### 1. 统一控制平面 (Control Plane)

不再需要在代码中硬编码API密钥和endpoint,一切配置化:

```bash
# 查看所有可用的AI provider
$ boba providers

Provider              Kind        Endpoint                      Status
claude-anthropic      anthropic   https://api.anthropic.com    ✓ Ready
claude-zai            anthropic   https://api.z.ai/api/...      ✓ Ready
openai-official       openai      https://api.openai.com        ✓ Ready
gemini-official       gemini      https://generativelanguage... ✓ Ready

# 绑定本地CLI工具到provider
$ boba bind claude claude-zai

# 运行时自动注入配置
$ boba run claude "Write a function to calculate fibonacci"
```

**核心价值**: 配置与代码解耦,一次配置,全局生效。

### 2. 本地HTTP Proxy (流量拦截与监控)

在你的本地启动一个智能代理,拦截所有AI API调用:

```bash
# 启动代理服务器(127.0.0.1:7777)
$ boba proxy serve &

# 所有经过proxy的请求都会被自动记录
# 支持 OpenAI 和 Anthropic 两种API格式
```

**技术亮点**:
- 零侵入式集成 - 只需修改环境变量 `ANTHROPIC_BASE_URL`
- 自动Token解析 - 从响应中提取精确的input/output tokens
- 实时成本计算 - 基于最新定价表计算每次调用成本
- 线程安全 - 支持并发请求,使用 `sync.RWMutex` 保护共享状态

### 3. 智能路由引擎 (Context-Aware Routing)

根据任务特征自动选择最优模型:

```yaml
# ~/.boba/routes.yaml
rules:
  - id: "large-context"
    if: "ctx_chars > 50000"
    use: "claude-anthropic"     # 长上下文用Claude
    explain: "Large context requires Claude's 200K window"

  - id: "code-review"
    if: "text.matches('review|audit|refactor')"
    use: "openai-gpt4"           # 代码审查用GPT-4
    fallback: "claude-anthropic"

  - id: "budget-conscious"
    if: "time_of_day == 'night' && budget.remaining < 5.0"
    use: "gemini-flash"          # 夜间且预算紧张用便宜模型
```

测试路由决策:

```bash
$ boba route test "Please review this PR and check for security issues"

=== Routing Decision ===
Profile: openai-gpt4
Rule ID: code-review
Explanation: Code review tasks use GPT-4 for best results
Fallback: claude-anthropic
```

**核心算法**: Epsilon-Greedy探索 + 规则引擎,在成本优化和效果探索之间自动平衡。

### 4. 预算管理与告警 (Budget Control)

多层级预算控制,防止成本失控:

```bash
# 查看当前预算状态
$ boba budget --status

Budget Scope: project (my-chatbot)
========================================
Today:  $2.34 of $10.00 (23.4%)
Period: $45.67 of $300.00 (15.2%)
Days Remaining: 23

# 设置预算限制
$ boba budget --daily 10.00 --cap 300.00

# 超预算时自动切换到更便宜的provider
$ boba action --auto
```

**技术实现**:
- 请求前预算检查 (`checkBudgetBeforeRequest`)
- 保守Token估算 (1000 input, 500 output)
- HTTP 429响应当预算超限
- 优雅降级 - 无预算配置时允许通过

### 5. 使用分析与成本追踪

精确的Token级别追踪和多维度分析:

```bash
# 查看今日统计
$ boba stats --today

Today's Usage
=============
Tokens:   45,678
Cost:     $1.23
Sessions: 12

# 7天趋势分析
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

# 导出报告
$ boba report --format json --output monthly-report.json
```

**数据Schema**:
- `sessions` 表 - 记录每次会话的元数据(project, branch, profile, latency等)
- `usage_records` 表 - 精确的token和成本记录,支持三种估算级别(exact/mapped/heuristic)
- SQLite存储 - 本地化,无需依赖外部数据库

### 6. 实时定价更新 (Pricing Auto-Refresh)

从OpenRouter API自动获取最新模型定价:

```bash
# 配置定价刷新策略
# ~/.boba/pricing.yaml
refresh:
  interval_hours: 24
  on_startup: false

# 手动验证定价数据
$ boba doctor --pricing

Pricing Validation
==================
✓ OpenRouter API accessible
✓ Cache fresh (updated 2 hours ago)
✓ 1,247 models loaded
✓ Fallback to vendor JSON available
```

**加载策略** (多层Fallback):
1. OpenRouter API (15秒超时)
2. 本地缓存 (24小时TTL)
3. Vendor JSON (内置数据)
4. pricing.yaml (用户自定义)
5. profiles.yaml cost_per_1k (最终兜底)

## 技术架构

### 模块化设计

```
BobaMixer
├── cmd/boba              # CLI入口
├── internal/cli          # 命令实现
├── internal/domain       # 核心领域逻辑
│   ├── budget           # 预算追踪
│   ├── pricing          # 定价管理(OpenRouter集成)
│   ├── routing          # 路由引擎
│   ├── stats            # 统计分析
│   └── suggestions      # 优化建议
├── internal/proxy        # HTTP代理服务器
├── internal/store        # 数据存储
│   ├── config           # 配置加载
│   └── sqlite           # SQLite操作
└── internal/ui           # TUI Dashboard (Bubble Tea)
```

### 关键技术选型

- **语言**: Go 1.22+ (类型安全, 并发友好, 单文件部署)
- **TUI**: Bubble Tea (现代化终端UI框架)
- **存储**: SQLite (零配置, 本地化, 支持SQL分析)
- **Lint**: golangci-lint (严格代码质量标准)
- **API集成**: OpenRouter Models API (1000+ 模型定价)

### Go最佳实践

项目严格遵循Go语言规范:

- ✅ **golangci-lint验证** - 0 issues
- ✅ **文档注释** - 所有导出类型/函数都有规范注释
- ✅ **错误处理** - 完整的error wrapping和优雅降级
- ✅ **并发安全** - 使用 `sync.RWMutex` 保护共享状态
- ✅ **安全编码** - 通过 `#nosec` 标记审计所有例外

## 快速开始

### 安装

```bash
# 使用 Go
go install github.com/royisme/bobamixer/cmd/boba@latest

# 或使用 Homebrew
brew tap royisme/tap
brew install bobamixer
```

### 初始化配置

```bash
# 初始化配置文件
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

### 配置API密钥

```bash
# 方式1: 环境变量(推荐)
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."

# 方式2: secrets.yaml
$ boba edit secrets
```

```yaml
# ~/.boba/secrets.yaml
secrets:
  anthropic-key: "sk-ant-..."
  openai-key: "sk-..."
  gemini-key: "..."
```

### 启动TUI Dashboard

```bash
$ boba
```

在交互式界面中:
- `↑/↓` 选择工具
- `B` 切换Provider绑定
- `X` 切换Proxy开关
- `R` 运行工具
- `q` 退出

## 使用场景

### 场景1: 团队协作 - 统一API管理

**问题**: 团队成员各自管理API密钥,容易泄露且难以审计。

**方案**:
```bash
# 1. 在项目根目录创建 .boba-project.yaml
$ cat .boba-project.yaml
project:
  name: "my-chatbot"
  type: ["backend", "ai"]
  preferred_profiles: ["claude-anthropic", "openai-gpt4"]

budget:
  daily_usd: 20.0
  hard_cap: 600.0

# 2. 团队成员各自配置 ~/.boba/secrets.yaml
# 3. 项目级预算自动生效
$ cd my-chatbot
$ boba budget --status  # 自动识别项目预算
```

### 场景2: 成本优化 - 自动降级

**问题**: 开发环境使用昂贵模型,测试时成本飙升。

**方案**:
```yaml
# routes.yaml - 根据分支自动选择模型
rules:
  - id: "production"
    if: "branch == 'main'"
    use: "claude-opus"

  - id: "development"
    if: "branch.matches('dev|feature')"
    use: "claude-haiku"  # 便宜80%

  - id: "test"
    if: "project_type contains 'test'"
    use: "gemini-flash"  # 最便宜
```

### 场景3: 多模型对比 - A/B测试

**问题**: 想评估不同模型在真实工作负载下的效果。

**方案**:
```bash
# 开启探索模式(3%流量随机路由)
$ boba init --explore-rate 0.03

# 7天后查看分析
$ boba stats --7d --by-profile

By Profile:
- openai-gpt4: avg_latency=1200ms cost=$6.20 usage=70%
- claude-sonnet: avg_latency=980ms cost=$1.80 usage=27%
- gemini-flash: avg_latency=650ms cost=$0.76 usage=3% (explore)

# 查看优化建议
$ boba action

💡 Suggestion: Switch to claude-sonnet for 40% cost reduction
   Impact: -$30/month, <5% quality difference
   Command: boba use claude-sonnet
```

## 高级功能

### Git Hooks集成

在commit过程中自动追踪AI调用:

```bash
# 安装hooks
$ boba hooks install

# 自动记录每次commit时的AI使用
$ git commit -m "feat: add authentication"
[BobaMixer] Tracked: 3 AI calls, 12K tokens, $0.34
```

### 建议引擎

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

# 自动应用高优先级建议
$ boba action --auto
```

## 命令参考

```bash
# 控制平面
boba providers                           # 列出所有provider
boba tools                               # 列出本地CLI工具
boba bind <tool> <provider>              # 创建绑定
boba run <tool> [args...]                # 运行工具

# HTTP Proxy
boba proxy serve                         # 启动代理
boba proxy status                        # 检查状态

# 使用统计
boba stats [--today|--7d|--30d]         # 查看统计
boba report --format json --out file     # 导出报告

# 预算管理
boba budget --status                     # 查看预算
boba budget --daily 10 --cap 300        # 设置限制

# 路由测试
boba route test "Your prompt here"       # 测试路由
boba route test @prompt.txt              # 从文件测试

# 优化建议
boba action                              # 查看建议
boba action --auto                       # 自动应用

# 配置管理
boba init                                # 初始化配置
boba edit <profiles|routes|pricing|secrets>
boba doctor                              # 健康检查

# 高级功能
boba hooks install                       # 安装Git hooks
boba completions install --shell bash    # 安装shell补全
```

## 配置文件结构

```
~/.boba/
├── providers.yaml      # AI服务商配置
├── tools.yaml          # 本地CLI工具
├── bindings.yaml       # 工具↔Provider绑定
├── secrets.yaml        # API密钥(权限: 0600)
├── routes.yaml         # 路由规则
├── pricing.yaml        # 定价配置
├── settings.yaml       # UI偏好
├── usage.db            # SQLite数据库
└── logs/               # 结构化日志
```

## 开发者指南

### 构建

```bash
# 克隆仓库
git clone https://github.com/royisme/BobaMixer.git
cd BobaMixer

# 安装依赖
go mod download

# 构建
make build

# 运行测试
make test

# Lint检查
make lint
```

### 环境要求

- Go 1.22+ (设置 `GOTOOLCHAIN=auto` 自动下载匹配版本)
- SQLite 3
- golangci-lint v1.60.1

```bash
# 确保Go自动获取匹配编译器
export GOTOOLCHAIN=auto

# 安装golangci-lint到本地 ./bin
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b ./bin v1.60.1
```

### 代码规范

项目遵循严格的Go语言规范:
- 所有导出类型和函数必须有文档注释
- 使用 `golangci-lint` 进行静态分析
- 遵循 [Effective Go](https://go.dev/doc/effective_go) 指南
- 提交前运行 `make test && make lint`

## 贡献指南

我们欢迎所有形式的贡献!

1. Fork本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交Pull Request

详见 [CONTRIBUTING.md](CONTRIBUTING.md)

## 路线图

- [x] Phase 1: Control Plane (Provider/Tool/Binding管理) - **100% 完成** ✅
- [x] Phase 2: HTTP Proxy & Usage监控 - **100% 完成** ✅
- [x] Phase 3: 智能路由 & 预算控制 & Pricing自动获取 - **100% 完成** ✅
- [ ] Phase 4: Web Dashboard (可选功能,TUI已足够强大)
- [ ] Phase 5: 多用户协作模式 (企业功能)

**🎉 当前状态**: 所有核心功能已完整实现,项目达到 **100% 完成度**！

## 开源协议

MIT License - 详见 [LICENSE](LICENSE) 文件

## 致谢

- 使用 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 构建TUI
- 定价数据由 [OpenRouter](https://openrouter.ai/) 提供
- 灵感来源于微服务编排和API网关设计

## 联系方式

- **Issues**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)
- **文档**: [完整文档](https://royisme.github.io/BobaMixer/)

---

**用一杯珍珠奶茶的时间,让AI成本降低50% ☕🧋**
