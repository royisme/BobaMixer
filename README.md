# BobaMixer

> 面向多种代码/AI CLI 工具的配置编排、智能路由与用量统计工具

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

**BobaMixer** 是一个本地优先的 CLI 工具，用于管理多个 AI 模型配置文件、智能路由请求、跟踪用量和成本。

## Quick Start

👉 **[查看 Quickstart 指南](QUICKSTART.md)** 快速开始使用

```bash
# 安装
go install github.com/royisme/bobamixer/cmd/boba@latest

# 设置配置
mkdir -p ~/.boba/logs
cp configs/examples/*.yaml ~/.boba/

# 使用
boba ls --profiles
boba use work-heavy
boba stats --today
boba doctor
```

`params.command` + `endpoint: stdio` 会驱动 MCP Adapter 通过 STDIN/STDOUT 调用自定义 server。

## 功能特性

- ✅ **Profile 管理** - 配置多个 AI 模型和工具，轻松切换
- ✅ **智能路由** - 基于规则自动选择最合适的 profile
- ✅ **用量统计** - 跟踪 token 使用量、成本和延迟
- ✅ **预算管理** - 设置每日预算和硬性上限
- ✅ **本地优先** - 所有数据存储在本地，不收集遥测
- ✅ **安全** - secrets.yaml 使用 0600 权限保护 API 密钥

## 架构

BobaMixer 采用分层架构设计：

- **CLI Layer** - 命令行接口 (use/ls/stats/doctor/budget/edit)
- **Domain Layer** - 业务逻辑 (Routing/Pricing/Session/Usage)
- **Adapter Layer** - 适配不同的服务 (HTTP/Tool/MCP)
- **Data Layer** - SQLite 数据库和 YAML 配置

## 开发状态

**当前版本**: Phase 4 (v0.4.0)

✅ Phase 1 已完成:
- SQLite 数据库自动引导
- 配置文件加载 (profiles/routes/pricing/secrets)
- HTTP 和 Tool 适配器基础框架
- CLI 命令 (ls/use/stats/edit/doctor/budget)
- Routing 路由引擎
- Pricing 价格管理器

✅ Phase 2 已完成:
- **ToolAdapter 增强** - JSON Lines usage 事件解析，支持参数和流式输出
- **Tokenizer 估算器** - 智能 token 估算（支持 GPT/Claude/通用模型）
- **HttpAdapter 增强** - 自动解析 Anthropic/OpenAI/OpenRouter API 的 usage 信息
- **完整的单元测试** - 所有核心模块测试覆盖

🚀 Phase 3/4 新增:
- ✅ GitHub Actions CI（编译 + go test）
- ✅ `boba release` 版本管理（自动 bump + changelog）
- ✅ 预算跟踪/提醒，支持 `.boba-project.yaml`
- ✅ 7/30 天趋势分析 + 建议引擎（CLI + 报表）
- ✅ TUI 仪表板 + 实时提醒
- ✅ MCP Adapter（面向 MCP Server 的 STDIO Transport）
- ✅ Git Hooks 集成（post-checkout/merge/commit）
- ✅ Goreleaser 配置

---

# BobaMixer 开发方案 v1

> 核心原则：本地优先、可解释、低侵入、可迭代。

---

## 0. 名称与范围

- **名称**：BobaMixer（CLI：`boba`）
- **目标**：对接多类“代码/AI”CLI 或 HTTP 客户端（Anthropic/OpenRouter、Claude Code、Codex CLI、后续 MCP）
- **能力**：Profile 管理、智能路由、用量/成本/延迟统计、预算提醒、项目/分支配置继承、TUI 控制台
- **非目标**：不会做熔断（hard stop）、不会依赖 OS Keychain、不会收集遥测

---

## 1. 架构概览

```
┌─────────────────────────────────────────────┐
│ TUI Layer (Bubble Tea + Lip Gloss + Glamour)│
│ 主屏/切换/统计/建议/项目/设置/诊断           │
└─────────────────────────────────────────────┘
                    ↕
┌─────────────────────────────────────────────┐
│ CLI (Cobra)                                 │
│ use/ls/stats/budget/route/doctor/edit/hooks │
└─────────────────────────────────────────────┘
                    ↕
┌─────────────────────────────────────────────┐
│ Domain & Services                           │
│ Profiles / Routing / Budget / Usage / Project
│ Suggestions / Pricing / Tokenizer           │
└─────────────────────────────────────────────┘
                    ↕
┌─────────────────────────────────────────────┐
│ Adapters                                    │
│ HttpAdapter / ToolAdapter / McpAdapter(后续) │
│ LogTap / Interceptor / Token Estimator      │
└─────────────────────────────────────────────┘
                    ↕
┌─────────────────────────────────────────────┐
│ Data Access                                 │
│ SQLite / YAML 配置 / JSONL 日志              │
└─────────────────────────────────────────────┘
```

---

## 2. 配置与文件布局（XDG ~`~/.boba`）

```
~/.boba/
  profiles.yaml         # 各模型/工具的连接与参数
  routes.yaml           # 路由规则与子代理（sub-agents）
  pricing.yaml          # 本地价格表（可被在线源覆盖）
  secrets.yaml          # API Key 等敏感项（0600 权限）
  usage.db              # SQLite 统计库（自动引导建表）
  logs/
    boba-YYYYMMDD.jsonl # 结构化运行日志
```

`secrets.yaml` 仅本机使用，建议 `chmod 600`；支持可选的本地对称加密（后续可接入 sops/age，首发不必需）。

### 2.1 profiles.yaml 示例

```yaml
profiles:
  work-heavy:
    name: "Work Heavy Tasks"
    adapter: "http"
    provider: "anthropic"
    endpoint: "https://api.anthropic.com"
    model: "claude-3-5-sonnet-latest"
    max_tokens: 4096
    temperature: 0.7
    tags: ["work","complex","analysis"]
    cost_per_1k:
      input: 0.015
      output: 0.075
    env:
      ANTHROPIC_API_KEY: "secret://anthropic"

  quick-tasks:
    name: "Quick Tasks"
    adapter: "http"
    provider: "openrouter"
    endpoint: "https://openrouter.ai/api/v1"
    model: "deepseek/deepseek-chat"
    max_tokens: 2048
    temperature: 0.3
    tags: ["quick","simple","code"]
    cost_per_1k:
      input: 0.0005
      output: 0.002
    env:
      OPENROUTER_API_KEY: "secret://openrouter"

  mcp-tools:
    name: "Local MCP"
    adapter: "mcp"
    provider: "local"
    endpoint: "stdio"
    params:
      command: "./scripts/mcp-server"
      default_tool: "codebase"
```

### 2.2 secrets.yaml 示例

```yaml
secrets:
  anthropic: "sk-ant-***"
  openrouter: "sk-or-***"
  deepseek: "sk-ds-***"
```

`env` 中出现 `secret://name` 时，运行期从 `secrets.yaml` 读取注入环境变量，值不会写入日志。

### 2.3 routes.yaml 示例

```yaml
sub_agents:
  code_review:
    profile: "work-heavy"
    triggers: ["review","check","audit"]
    conditions:
      min_ctx_chars: 3000
      project_types: ["java","go","ts"]

  quick_fix:
    profile: "quick-tasks"
    triggers: ["fix","typo","format"]
    conditions:
      max_ctx_chars: 1200
      time_of_day: ["09:00-18:00"]

rules:
  - id: "formatting"
    if: "intent=='format' || text.matches('\\bformat\\b|\\bprettier\\b')"
    use: "quick-tasks"
    explain: "格式化类任务优先低成本"

  - id: "deep-analysis"
    if: "ctx_chars>3000 || task.matches('architecture|review|audit')"
    use: "work-heavy"
    fallback: "quick-tasks"
```

### 2.4 项目级 `.boba-project.yaml`

```yaml
project:
  name: "codebase-rag"
  type: ["python","neo4j"]
  preferred_profiles: ["work-heavy","quick-tasks"]

routing:
  rules:
    - if: "task.contains('format')"
      use: "quick-tasks"
    - if: "branch.matches('^release/') || pr_size>1000"
      use: "work-heavy"

budget:
  daily_usd: 5.0
  hard_cap: 50.0
```

`boba budget --status` 会自动向上搜索 `.boba-project.yaml` 并为项目创建/同步预算记录，可用 `--daily`、`--cap` 快速调整。

---

## 3. 统计与 SQLite

- `~/.boba/usage.db`
- `PRAGMA user_version` 管理 schema（v1）

```sql
CREATE TABLE IF NOT EXISTS sessions (
  id           TEXT PRIMARY KEY,
  started_at   INTEGER NOT NULL,
  ended_at     INTEGER,
  project      TEXT,
  branch       TEXT,
  profile      TEXT,
  adapter      TEXT,
  task_type    TEXT,
  success      INTEGER,
  latency_ms   INTEGER,
  notes        TEXT
);

CREATE TABLE IF NOT EXISTS usage_records (
  id             TEXT PRIMARY KEY,
  session_id     TEXT NOT NULL,
  ts             INTEGER NOT NULL,
  input_tokens   INTEGER DEFAULT 0,
  output_tokens  INTEGER DEFAULT 0,
  input_cost     REAL DEFAULT 0,
  output_cost    REAL DEFAULT 0,
  tool           TEXT,
  model          TEXT,
  FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS budgets (
  id           TEXT PRIMARY KEY,
  scope        TEXT NOT NULL,
  target       TEXT,
  daily_usd    REAL,
  hard_cap     REAL,
  period_start INTEGER,
  period_end   INTEGER,
  spent_usd    REAL DEFAULT 0
);

CREATE VIEW IF NOT EXISTS v_daily_summary AS
SELECT
  date(ts, 'unixepoch') AS date,
  SUM(input_tokens + output_tokens) AS total_tokens,
  SUM(input_cost + output_cost)     AS total_cost
FROM usage_records
GROUP BY date;
```

引导流程：打开/创建 `usage.db` → 若 `user_version=0` 则执行 DDL 并设为 1 → 后续演进使用 `ALTER TABLE` + `user_version` 增量。

---

## 4. Adapter 设计

### 4.1 HttpAdapter（首发）
- 适配 Anthropic/OpenRouter/DeepSeek
- 若响应无 usage 字段，则用 Tokenizer 估算并标记 `estimate_level`
- 成本优先在线价格表 → 本地 `pricing.yaml` → `profiles.yaml` 兜底

### 4.2 ToolAdapter（首发之一）
- 适配 `claude-code`、`codex` 等 CLI
- 监听 stdout/stderr JSON Lines usage 事件；否则估算 tokens

### 4.3 McpAdapter（后续）
- MCP 客户端交互，采集 usage，不在首发范围

统一事件（JSON Lines）：

```json
{"event":"request","session_id":"...","profile":"quick-tasks","tool":"claude-code","model":"...","ts":"..."}
{"event":"usage","session_id":"...","input_tokens":153,"output_tokens":412,"latency_ms":8312}
{"event":"result","session_id":"...","success":true}
```

---

## 5. 智能路由与建议

```
输入 → 特征提取(intent/ctx_chars/project/branch/time/budget_hint)
     → 规则 DSL 命中（优先级/短路）
     → 未命中：小比例探索（默认 3%）
     → 选择 profile/adapter 执行 → 记录 usage/延迟/成功
     → 汇总出“性价比”与“建议”
```

- 成本优化建议：对比近 7/30 天相似上下文的单位成功成本与延迟
- TUI 提供 `[A]应用 / [I]忽略 / [L]稍后`，仅提示不强制

---

## 6. 价格表策略

```yaml
models:
  "anthropic/claude-3-5-sonnet-latest":
    input_per_1k: 0.015
    output_per_1k: 0.075
  "deepseek/deepseek-chat":
    input_per_1k: 0.0005
    output_per_1k: 0.002
sources:
  - type: "http-json"
    url: "https://raw.githubusercontent.com/vantagecraft-dev/boba-mixer-pricing/main/pricing.json"
    priority: 10
  - type: "file"
    path: "~/.boba/pricing.local.json"
    priority: 5
refresh:
  interval_hours: 24
  on_startup: true
```

优先级：在线 JSON（成功则缓存）> 本地 `pricing.local.json` > `pricing.yaml` > `profiles.yaml` 中 `cost_per_1k`。

---

## 7. TUI 设计

- 导航：Profiles / Routing / Usage / Budget / Projects / Doctor / Settings
- 主题：浅/深双色，奶茶风
- 今日仪表板示例：

```
╭─ BobaMixer · Today ─────────────────────────────────╮
│ Cost  $2.45   Tokens 45.2k   Sessions 15   P95 3.2s │
│                                                     │
│ Cost Trend (7d)  ▂▄█▆▃▂▁                              │
│ Profile Usage                                      │
│ ███████░░  work-heavy  (80%)  $1.96   P95 4.1s      │
│ ██░░░░░░  quick-tasks (20%)  $0.49   P95 1.2s       │
│                                                     │
│ 💡 Suggestion: 将“format”任务路由到 quick-tasks，      │
│   预计节省 ~$0.8/日（置信度 84%）。 [A]应用 [I]忽略     │
╰─────────────────────────────────────────────────────╯
```

---

## 8. CLI 子命令

```
boba use <profile>
boba ls [--profiles|--adapters]
boba stats [--today|--7d|--30d|--json]
boba budget [--status] [--daily 5] [--cap 50]
boba route test "<text|@file>"
boba doctor
boba edit profiles|routes|pricing|secrets
boba hooks install|remove
boba release --bump patch [--notes "..."]

### 8.1 Git Hooks

- `boba hooks install`：自动在当前 Git 仓库注入 `post-checkout/post-merge/post-commit` 脚本
- Hook 会调用 `boba hooks track`，将分支/事件记录到 `~/.boba/git-hooks/*.jsonl`
- `boba hooks remove`：安全删除脚本
boba release --bump patch [--notes "..."]
```

数据库自动引导建表，无 `migrate`。

---

## 9. 错误处理与可靠性

- HTTP/Tool 失败：指数退避重试（≤2 次）→ 失败则按规则 fallback profile（若有）
- 价格源不可用：使用缓存 → 本地定价 → profiles 兜底
- 用量估算等级：`exact|mapped|heuristic`，落库供纠偏

---

## 10. 性能指标

- `boba use` ≤ 150ms
- `stats --7d` ≤ 200ms（索引 `usage_records(ts)`）
- Adapter 默认直连，仅在不可观测时启用拦截

---

## 11. 打包与分发

- Go 1.22+
- `goreleaser` 输出 macOS/Linux 各架构
- 可选 Homebrew Tap；Linux 提供 .deb/.rpm

### 11.1 版本发布流程

- `VERSION` 文件作为单一真相
- `boba release --bump patch --notes "..."` 自动更新 VERSION + `CHANGELOG.md`
- `.goreleaser.yaml` 提供 `goreleaser release --clean` 所需配置
- GitHub Actions CI 在 PR/Push 上跑 `gofmt`、`go vet`、`go test`

---

## 12. 开发计划（8 周）

1. **Phase 1**：SQLite 引导、配置解析、HttpAdapter（1 provider）、`boba use/ls/stats/edit`、TUI 主屏
2. **Phase 2**：ToolAdapter、Tokenizer 估算、预算提示与趋势、价格源拉取
3. **Phase 3**：路由 DSL、探索、建议引擎、`route test`
4. **Phase 4**：Git Hooks/补全、`doctor`、`goreleaser` 发布、文档站

---

## 13. 测试策略

- 单元：profiles/routes 解析、cost 计算、token 估算、价格回退
- 集成：HttpAdapter/ToolAdapter 端到端
- 金样：路由 DSL 解释
- 性能：统计查询、TUI 渲染
- 回归：建议引擎输出稳定性

---

## 14. 安全与隐私

- 不使用 OS Keychain；敏感信息仅存 `secrets.yaml`，权限 0600
- 日志/库不存请求正文，仅元数据
- `boba purge` 支持导出并删除

---

## 15. 参考目录结构

```
cmp/boba/main.go
internal/ui/...
internal/cli/...
internal/domain/...
internal/adapters/...
internal/store/...
internal/integration/...
internal/svc/...
configs/examples/...
docs/...
```

---

## 16. 开放点

- API Key：配置文件管理
- 熔断：不实现
- 价格源：在线拉取接口，若无则使用我们托管静态 JSON，可本地覆盖

---

## 17. 首版交付清单

1. 配置模板（profiles/routes/secrets/pricing）
2. SQLite `bootstrap.go`
3. HttpAdapter（一个 provider）
4. `boba use|ls|stats|edit` + TUI 主屏
5. README Quickstart + Adapter 指南 + Routing Cookbook
```
