# BobaMixer 🧋

> **AI工作流的智能路由器与成本优化引擎**

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/royisme/BobaMixer)](https://github.com/royisme/BobaMixer/releases)
[![golangci-lint](https://img.shields.io/badge/lint-passing-brightgreen)](https://golangci-lint.run/)

[**📚 中文文档**](https://royisme.github.io/BobaMixer/zh/) | [**🚀 快速开始**](#快速开始) | [**English**](README.md)


---

## 功能概览

**核心功能(控制平面 + boba run)**
- 将 Provider / Tool / Binding 作为一等对象管理
- 通过 `boba run` 运行本地 AI CLI 工具,自动注入凭证和端点
- 可选的本地代理来整合请求

**高级功能(遗留/可选)**
- 路由 / Profile 配置
- 预算与定价控制
- 使用统计 & Git hooks

---

## 为什么选择 BobaMixer?

在AI开发的日常工作中,你是否遇到过这些痛点:

- 🔑 **密钥管理混乱** - 多个AI服务的API密钥散落在各处,切换provider需要修改配置文件
- 💸 **成本失控** - 不知不觉中API调用费用飙升,缺乏实时监控和预算控制
- 🎯 **路由决策困难** - 不同任务应该用哪个模型?如何在成本和效果之间平衡?
- 📊 **使用数据缺失** - 无法追踪token消耗、成本分布,缺乏优化依据
- 🔄 **切换成本高** - 从Claude切到OpenAI需要修改代码,无法灵活调度

**BobaMixer 就是为解决这些问题而生的** —— 它是你的AI工作流控制平面,让你像调度微服务一样调度AI模型。

---

## 核心能力

### 1. 统一控制平面

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
$ boba run claude "编写一个计算斐波那契数列的函数"
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
- **零侵入式集成** - 只需修改环境变量
- **自动Token解析** - 从响应中提取精确的tokens
- **实时成本计算** - 基于最新定价表计算成本
- **线程安全** - 使用sync.RWMutex保护共享状态

## 高级能力(遗留/可选)

> 以下模块是高级/遗留功能。它们不是核心控制平面 + `boba run` 路径的一部分,但仍可供高级用户使用。

### [高级] 智能路由引擎 (Context-Aware)

根据任务特征自动选择最优模型:

```yaml
# ~/.boba/routes.yaml
rules:
  - id: "large-context"
    if: "ctx_chars > 50000"
    use: "claude-anthropic"     # 长上下文 → Claude
    explain: "Large context requires Claude's 200K window"

  - id: "code-review"
    if: "text.matches('review|audit|refactor')"
    use: "openai-gpt4"           # 代码审查 → GPT-4
    fallback: "claude-anthropic"

  - id: "budget-conscious"
    if: "time_of_day == 'night' && budget.remaining < 5.0"
    use: "gemini-flash"          # 夜间 + 低预算 → 便宜模型
```

测试路由决策:

```bash
$ boba route test "请审查这个PR并检查安全问题"

=== 路由决策 ===
Profile: openai-gpt4
Rule ID: code-review
说明: 代码审查任务使用 GPT-4 获得最佳结果
Fallback: claude-anthropic
```

**核心算法**: Epsilon-Greedy探索 + 规则引擎,在成本优化和效果探索之间自动平衡。

### [高级] 预算管理与告警

多层级预算控制,防止成本失控:

```bash
# 查看当前预算状态
$ boba budget --status

Budget Scope: project (my-chatbot)
========================================
今日:  $2.34 / $10.00 (23.4%)
周期: $45.67 / $300.00 (15.2%)
剩余天数: 23

# 设置预算限制
$ boba budget --daily 10.00 --cap 300.00

# 超预算时自动切换到更便宜的provider
$ boba action --auto
```

**技术实现**:
- 请求前预算检查 (`checkBudgetBeforeRequest`)
- 保守Token估算 (1000 input, 500 output)
- HTTP 429响应当预算超限
- 优雅降级 - 允许在没有预算配置时通过

### [高级] 使用分析与成本追踪

精确的Token级别追踪和多维度分析:

```bash
# 查看今日统计
$ boba stats --today

今日使用
=============
Tokens:   45,678
成本:     $1.23
会话数: 12

# 7天趋势分析
$ boba stats --7d --by-profile

最近7天使用
=================
总 Tokens:   312,456
总成本:     $8.76
平均每日成本: $1.25

按 Profile 分析:
-----------
- openai-gpt4: tokens=180K cost=$6.20 sessions=45 avg_latency=1200ms usage=57.6% cost=70.8%
- claude-sonnet: tokens=90K cost=$1.80 sessions=23 avg_latency=980ms usage=28.8% cost=20.5%
- gemini-flash: tokens=42K cost=$0.76 sessions=18 avg_latency=650ms usage=13.5% cost=8.7%

# 导出报告
$ boba report --format json --output monthly-report.json
```

**数据结构**:
- `sessions` 表 - 记录会话元数据(项目、分支、profile、延迟)
- `usage_records` 表 - 精确的token & 成本记录,3种估算级别(精确/映射/启发式)
- SQLite 存储 - 本地化,无外部数据库依赖

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

定价验证
==================
✓ OpenRouter API 可访问
✓ 缓存新鲜 (2小时前更新)
✓ 加载了 1,247 个模型
✓ Vendor JSON fallback 可用
```

**加载策略**(多层Fallback):
1. OpenRouter API (15s 超时)
2. 本地缓存 (24h TTL)
3. Vendor JSON (内嵌数据)
4. pricing.yaml (用户定义)
5. profiles.yaml cost_per_1k (最终fallback)

---

## 技术架构

### 模块化设计

```
BobaMixer
├── cmd/boba              # CLI入口
├── internal/cli          # 命令实现
├── internal/domain       # 核心领域逻辑
│   ├── budget           # 预算追踪
│   ├── pricing          # 定价管理(OpenRouter)
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

- **语言**: Go 1.25+ (类型安全, 并发友好, 单文件部署)
- **TUI**: Bubble Tea (现代化终端UI框架)
- **存储**: SQLite (零配置, 本地化, 支持SQL分析)
- **Lint**: golangci-lint (严格代码质量标准)
- **API集成**: OpenRouter Models API (1000+ 模型定价)

### Go最佳实践

项目严格遵循Go语言规范:

- ✅ **golangci-lint 验证** - 0 issues
- ✅ **文档注释** - 所有导出类型/函数都有规范注释
- ✅ **错误处理** - 完整的error wrapping和优雅降级
- ✅ **并发安全** - 使用sync.RWMutex保护共享状态
- ✅ **安全编码** - 通过#nosec标记审计所有例外

---

## 快速开始

### 安装

```bash
# 使用 Go
go install github.com/royisme/bobamixer/cmd/boba@latest

# 或使用 Homebrew
brew tap royisme/tap
brew install bobamixer
```

### 首次设置 - 交互式向导 🎯

BobaMixer 会自动引导你完成所有配置,**无需手动编辑任何配置文件**:

```bash
# 1. 启动 BobaMixer(首次运行会自动进入向导)
$ boba

# Onboarding 向导会自动:
# ✓ 检测本地 CLI 工具 (claude/codex/gemini)
# ✓ 让你选择 Provider
# ✓ 引导输入 API Key(安全输入,自动保存)
# ✓ 创建所有配置文件
# ✓ 验证配置

# 2. 完成后即可使用
$ boba run claude --version
```

### 备选方案: CLI 设置 (适合高级用户)

如果你更喜欢命令行:

```bash
# 1. 初始化配置目录
$ boba init

# 2. 配置 API Key(安全输入,无需编辑 YAML)
$ boba secrets set claude-anthropic-official
Enter API key: ********
✓ API key saved to ~/.boba/secrets.yaml (permissions: 0600)

# 3. 绑定工具到 Provider
$ boba bind claude claude-anthropic-official

# 4. 验证配置
$ boba doctor

# 5. 运行
$ boba run claude --version
```

### 环境变量 (可选)

你也可以使用环境变量(适合 CI/CD 或临时使用):

```bash
# BobaMixer 会优先使用环境变量
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."

# 然后直接运行
$ boba run claude --version
```

### 启动TUI Dashboard

```bash
$ boba
```

**交互式操作**:
- `↑/↓` 选择工具
- `B` 切换Provider绑定
- `X` 切换Proxy开关
- `V` 统计视图
- `R` 运行工具
- `q` 退出

---

## 使用场景

### 案例1: 团队协作 - 统一API管理

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
$ boba budget --status  # 自动检测项目预算
```

### 案例2: 成本优化 - 自动降级

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

### 案例3: 多模型对比 - A/B测试

**问题**: 想评估不同模型在真实工作负载下的效果。

**方案**:
```bash
# 开启探索模式(3%流量随机路由)
$ boba init --explore-rate 0.03

# 7天后查看分析
$ boba stats --7d --by-profile

按 Profile 分析:
- openai-gpt4: avg_latency=1200ms cost=$6.20 usage=70%
- claude-sonnet: avg_latency=980ms cost=$1.80 usage=27%
- gemini-flash: avg_latency=650ms cost=$0.76 usage=3% (explore)

# 查看优化建议
$ boba action

💡 建议: 切换到 claude-sonnet 可降低40%成本
   影响: -$30/月, <5% 质量差异
   命令: boba use claude-sonnet
```

---

## 高级功能

### Git Hooks集成

在commit过程中自动追踪AI调用:

```bash
# 安装 hooks
$ boba hooks install

# 自动记录每次commit时的AI使用
$ git commit -m "feat: add authentication"
[BobaMixer] Tracked: 3 AI calls, 12K tokens, $0.34
```

### 建议引擎

基于历史数据生成优化建议:

```bash
$ boba action

💡 高优先级建议:
  1. [COST] 将代码任务的 'openai-gpt4' 切换到 'claude-sonnet'
     → 节省 $45/月 (当前: $120/月 → 预计: $75/月)

  2. [PERF] 为重复查询启用缓存
     → 减少60%延迟 (平均: 1200ms → 480ms)

  3. [BUDGET] 每日支出有超出月度上限的趋势
     → 需要行动: 减少使用或增加上限

# 自动应用高优先级建议
$ boba action --auto
```

---

## 命令参考

```bash
# 控制平面
boba providers                           # 列出所有 providers
boba tools                               # 列出本地 CLI 工具
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

# 路由
boba route test "您的提示词"              # 测试路由
boba route test @prompt.txt              # 从文件测试

# 优化
boba action                              # 查看建议
boba action --auto                       # 自动应用

# 配置
boba init                                # 初始化配置
boba edit <profiles|routes|pricing|secrets>
boba doctor                              # 健康检查

# 高级
boba hooks install                       # 安装 Git hooks
boba completions install --shell bash    # Shell 补全
```

---

## 配置文件结构

```
~/.boba/
├── providers.yaml      # AI 服务 provider 配置
├── tools.yaml          # 本地 CLI 工具
├── bindings.yaml       # Tool ↔ Provider 绑定
├── secrets.yaml        # API 密钥 (权限: 0600)
├── routes.yaml         # 路由规则
├── pricing.yaml        # 定价配置
├── settings.yaml       # UI 偏好设置
├── usage.db            # SQLite 数据库
└── logs/               # 结构化日志
```

---

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

# Lint 检查
make lint
```

### 环境要求

- Go 1.25+ (设置 `GOTOOLCHAIN=auto` 自动下载)
- SQLite 3
- golangci-lint v1.60.1

```bash
# 确保 Go 自动获取匹配的编译器
export GOTOOLCHAIN=auto

# 本地安装 golangci-lint (./bin)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b ./bin v1.60.1
```

### 代码规范

项目遵循严格的Go语言规范:
- 所有导出类型和函数必须有文档注释
- 使用golangci-lint进行静态分析
- 遵循[Effective Go](https://go.dev/doc/effective_go)指南
- 提交前运行make test && make lint

---

## 贡献指南

我们欢迎所有形式的贡献!

1. Fork 仓库
2. 创建 feature 分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

---

## 路线图

- [x] 阶段 1: 控制平面 (Provider/Tool/Binding管理) - **100% 完成** ✅
- [x] 阶段 2: HTTP Proxy & Usage监控 - **100% 完成** ✅
- [x] 阶段 3: 智能路由 & 预算控制 & 定价自动获取 - **100% 完成** ✅
- [ ] 阶段 4: Web Dashboard (可选功能,TUI已足够强大)
- [ ] 阶段 5: 多用户协作模式 (企业功能)

**🎉 当前状态**: 所有核心功能已完整实现,项目达到 **100% 完成度**!

---

## 开源协议

MIT License - 详见 [LICENSE](LICENSE) 文件。

---

## 致谢

- 使用 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 构建 TUI
- 定价数据由 [OpenRouter](https://openrouter.ai/) 提供
- 受微服务编排和API网关设计启发

---

## 联系方式

- **问题反馈**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **讨论区**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)
- **完整文档**: [文档](https://royisme.github.io/BobaMixer/zh/)

---

<div align="center">

**用一杯珍珠奶茶的时间,让AI成本降低50% ☕🧋**

Made with ❤️ by developers, for developers

</div>
