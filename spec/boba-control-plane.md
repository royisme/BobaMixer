
---

# BobaMixer 架构设计（控制平面 / boba run / Proxy）

版本：v0.1（Draft）
目标读者：BobaMixer 维护者 & AI coding 工具高级用户
定位：这是一个「管理层（Control Plane）」的设计文档，不是用户使用说明。

---

## 1. 产品定位与边界

### 1.1 BobaMixer 是什么

BobaMixer（后简称 boba）不是一个 AI coding tool，而是一个**本地 AI CLI Agents 的「控制平面」**：

* 被管理对象：本机的各类 AI CLI 工具，例如：

  * Claude Code CLI（`claude`）
  * Codex CLI（`codex`）
  * Gemini CLI（`gemini`）
  * 以及未来其它 openai-compatible / anthropic-compatible 工具
* Boba 的职责：

  1. 统一管理 OpenAI / Claude / Gemini / Router 等 **Provider** 的配置（base_url, api_key, 默认模型等）；
  2. 管理「某个 CLI 工具当前使用哪个 Provider」的 **绑定关系（Binding）**；
  3. 提供 `boba run ...` 命令，**在运行 CLI 之前注入正确的 env / 配置**，让 CLI 的行为发生实际变化；
  4. 可选：通过本地 Proxy 把请求统一出口，监控 tokens / 网络 / 错误等。

### 1.2 BobaMixer 不是什么

* 不是一个「再造一个 AI chat/coding 界面」的工具；
* 不负责替代 Claude Code / Codex / Gemini CLI 自己的交互体验；
* 不在 v1 阶段实现复杂的多模型路由决策、预算控制、团队审计等（这些归为高级功能，后续迭代）。

---

## 2. 核心概念与数据模型

Boba 的 Domain 由以下几个一等公民概念构成：

1. **Provider**：一个“上游模型服务提供方”

   * 例如：OpenAI 官方、Anthropic 官方、Z.AI Anthropic 兼容路由、内部自建 Router、Gemini 官方。
   * 提供：base_url, API Key env 名称, 默认 model 名等。

2. **Tool**：一个“被管理的本地 CLI 工具”

   * 例如：`claude` / `codex` / `gemini`。
   * 描述：它的配置文件在哪里，用什么协议读配置（settings.json / config.toml / .env）。

3. **Binding**：Tool ↔ Provider 的绑定关系

   * “Codex 现在用哪个 Provider？”
   * “Claude Code 现在通过 Z.AI 调 Anthropic？”
   * “这些调用是否走 Boba 的本地 Proxy？”

4. **Secrets**：API 密钥等敏感信息

   * 优先从环境变量读取（尊重已有生态习惯）。
   * Boba 的 `secrets.yaml` 只做兜底存储。

5. **Profile**（可选层次）

   * 一个 profile = Provider + model + 默认参数（temperature 等）。
   * 在本设计中是**可选层**，重点仍然是 Provider / Tool / Binding。

---

## 3. 配置文件与目录结构

默认配置目录：

```text
~/.boba/
  providers.yaml    # Provider 列表
  tools.yaml        # Tool 列表（本机有哪些 CLI 以及配置方式）
  bindings.yaml     # Tool ↔ Provider 绑定关系
  secrets.yaml      # API key 等（可选）
  profiles.yaml     # 可选：Profile 描述
  settings.yaml     # UI 偏好等
  routes.yaml       # [高级] 路由策略
  pricing.yaml      # [高级] 定价信息
  usage.db          # [高级] 统计数据（SQLite）
```

### 3.1 `providers.yaml`

```yaml
version: 1

providers:
  - id: openai-official
    kind: openai                   # openai | anthropic | gemini | openai-compatible | anthropic-compatible ...
    display_name: "OpenAI (official)"
    base_url: "https://api.openai.com/v1"
    api_key:
      source: env                  # env | secrets
      env_var: "OPENAI_API_KEY"
    default_model: "gpt-4.1-mini"
    enabled: true

  - id: claude-anthropic-official
    kind: anthropic
    display_name: "Claude (Anthropic official)"
    base_url: "https://api.anthropic.com"
    api_key:
      source: env
      env_var: "ANTHROPIC_API_KEY"  # 也可支持 ANTHROPIC_AUTH_TOKEN
    default_model: "claude-3.7-sonnet"
    enabled: true

  - id: claude-zai
    kind: anthropic-compatible
    display_name: "Claude via Z.AI (GLM-4.6)"
    base_url: "https://api.z.ai/api/anthropic"
    api_key:
      source: env
      env_var: "ANTHROPIC_AUTH_TOKEN"
    default_model: "glm-4.6"
    enabled: true

  - id: gemini-official
    kind: gemini
    display_name: "Gemini (Google)"
    base_url: "https://generativelanguage.googleapis.com"
    api_key:
      source: env
      env_var: "GEMINI_API_KEY"     # 或 GOOGLE_API_KEY
    default_model: "gemini-2.0-flash"
    enabled: true
```

说明：

* `kind` 决定具体请求协议（OpenAI-style / Anthropic-style / Gemini-style）。
* `base_url` + `env_var` 对齐各家官方推荐/env 约定。
* `enabled` 用于临时关闭 Provider。

### 3.2 `tools.yaml`

```yaml
version: 1

tools:
  - id: claude
    name: "Claude Code CLI"
    exec: "claude"                        # 系统上实际执行的命令名
    kind: "claude"
    config_type: "claude-settings-json"   # 内部约定的配置类型
    config_path: "~/.claude/settings.json"

  - id: codex
    name: "Codex CLI"
    exec: "codex"
    kind: "codex"
    config_type: "codex-config-toml"
    config_path: "~/.codex/config.toml"

  - id: gemini
    name: "Gemini CLI"
    exec: "gemini"
    kind: "gemini"
    config_type: "gemini-settings-json"
    config_path: "~/.gemini/settings.json"
```

说明：

* `exec` 是真正被 `boba run` 启动的命令；
* `config_type` 决定如何读写配置文件：

  * `claude-settings-json`：修改 `settings.json` 的 `"env"` 段；
  * `codex-config-toml`：修改 `config.toml` 的 `model` 和 provider 段；
  * `gemini-settings-json`：生成 `settings.json` & `.env` 等。

### 3.3 `bindings.yaml`

```yaml
version: 1

bindings:
  - tool_id: claude
    provider_id: claude-zai
    use_proxy: true
    options:
      model_mapping:
        opus: "glm-4.6"
        sonnet: "glm-4.6"
        haiku: "glm-4.5-air"

  - tool_id: codex
    provider_id: openai-official
    use_proxy: true
    options:
      model: "gpt-5.1-codex"

  - tool_id: gemini
    provider_id: gemini-official
    use_proxy: false
    options: {}
```

说明：

* `tool_id` / `provider_id` 建立一对一绑定；
* `use_proxy` 表示本次绑定是否通过本地 Proxy 出口；
* `options` 留给各 Tool 特殊设置，例如：

  * Claude 的 `ANTHROPIC_DEFAULT_*_MODEL` 映射；
  * Codex 的 `model` / `model_provider`；
  * Gemini 的特定模式。

### 3.4 `secrets.yaml`（可选）

```yaml
version: 1

secrets:
  openai-official:
    api_key: "sk-..."             # 当 provider.api_key.source == "secrets" 时使用

  claude-anthropic-official:
    api_key: "sk-ant-..."

  claude-zai:
    api_key: "zai-..."

  gemini-official:
    api_key: "xxx..."
```

优先级约定（固定）：

1. CLI 显式参数（未来可能支持）；
2. 环境变量（`OPENAI_API_KEY` / `ANTHROPIC_API_KEY` / `GEMINI_API_KEY` 等）；
3. `secrets.yaml` 中的值。

### 3.5 `profiles.yaml`（非必需）

如果后续需要 profile 层，可以采用：

```yaml
version: 1

profiles:
  - name: default
    provider_id: openai-official
    model: "gpt-4.1-mini"
    description: "日常开发 / 对话"
    params:
      temperature: 0.6

  - name: claude-dev
    provider_id: claude-anthropic-official
    model: "claude-3.7-sonnet"
    description: "长上下文代码工作"
    params:
      temperature: 0.4
```

在本架构中，Profile 是可选扩展层，核心仍然是 Tool ↔ Provider 的绑定。

---

## 4. CLI 设计：核心命令 vs 高级命令

### 4.1 核心命令（Core）

面向「单机高级用户」，默认在 `boba --help` 中展示：

```text
boba
    启动 TUI 控制面板（首启进入向导，之后进入主界面）

boba run <tool> [args...]
    按当前 Binding 和 Proxy 配置运行指定的 CLI 工具
    例如：boba run claude --agent=code_reiver

boba providers
    列出已配置的 Provider 及状态

boba tools
    列出已发现/已配置的本地 CLI tools

boba bind <tool> <provider> [--proxy=on|off]
    更新某个工具与 Provider 的绑定

boba doctor
    做基础诊断：检查 Provider 配置和 Bindings 是否可用

boba proxy serve
    启动本地 Proxy（OpenAI/Anthropic 风格）
```

### 4.2 高级命令（Advanced）

只在 `boba help advanced` 中展示：

```text
boba stats [--today|--7d|--30d] [--by-tool|--by-provider]
boba report [--format json|csv] [--out file]

boba budget [--status]
boba action [--auto]

boba route test <text|@file>
boba doctor --pricing

boba hooks install|remove|track

boba init
    初始化 ~/.boba 目录，创建默认 providers.yaml/tools.yaml/bindings.yaml 等
    多用于 CI 或团队模板，不建议普通用户首次使用
```

---

## 5. `boba run` 行为与 CLI 集成细节

### 5.1 `boba run` 执行流水线

`boba run <tool> [args...]` 的逻辑步骤：

1. 解析参数，确定 `tool_id`（例如 `claude`）。
2. 从 `tools.yaml` 读出 Tool 定义；从 `bindings.yaml` 读出其绑定的 Provider 和 `use_proxy`。
3. 从 `providers.yaml` 找到对应 Provider，并解析：

   * base_url
   * api_key 的来源（env/secrets）
   * default_model 等
4. 合成「运行时配置」（runtime config）：

   * 要写入/覆盖的 env：`ANTHROPIC_API_KEY` / `OPENAI_API_KEY` / `GEMINI_API_KEY` 等；
   * 要传入的其它 env：`ANTHROPIC_BASE_URL` / `OPENAI_BASE_URL`（如走 Proxy）；
   * 针对 Tool 的特殊配置：例如 `ANTHROPIC_DEFAULT_*_MODEL`、Codex 模型名。
5. 如果 Binding 要求 `use_proxy = true`：

   * 确认 Proxy 已启动（否则自动启动 `boba proxy serve` 子进程）；
   * 把 base_url/env 改成指向 Proxy 的 URL（见第 6 章）。
6. 根据 `tools.yaml` 中 `config_type` 调用相应集成逻辑：

   * 只在本次进程 env 注入；
   * 或必要时对 Tool 的配置文件（settings.json/config.toml）做最小必要修改（带备份）。
7. 启动子进程：

   * 命令 = `tools.exec`（如 `claude`）
   * 参数 = `[args...]` 原样透传（比如 `--agent=code_reiver`）
   * 环境变量 = 上述合成后的 env
8. 子进程退出后（可选）：

   * 如果走 Proxy，则 usage 由 Proxy 记录；
   * Boba 本体只需要返回子进程 exit code。

### 5.2 Claude Code 集成（`claude`）

Claude Code 关键点（根据官方文档 & Z.AI 文档）：

* API key env：`ANTHROPIC_API_KEY` 或 `ANTHROPIC_AUTH_TOKEN`
* base_url env：`ANTHROPIC_BASE_URL`（Z.AI 配置该变量到 `https://api.z.ai/api/anthropic`）
* 默认模型：通过 `ANTHROPIC_DEFAULT_*_MODEL` 系列 env 或 `settings.json` 中的 `env` 字段。
* 全局 instructions：`~/.claude/CLAUDE.md` + `settings.json` 其它字段。

Boba 的 Claude 集成策略：

* Provider 层维护：

  * 官方 Anthropic：`base_url=https://api.anthropic.com`，env_var=`ANTHROPIC_API_KEY`
  * Z.AI：`base_url=https://api.z.ai/api/anthropic`，env_var=`ANTHROPIC_AUTH_TOKEN`
* `boba run claude ...` 时：

  * 从 Provider/Secrets/env 中获取 Key；
  * 如果不走 Proxy：直接在子进程 env 中设置：

    * `ANTHROPIC_AUTH_TOKEN` / `ANTHROPIC_API_KEY`
    * `ANTHROPIC_BASE_URL`（如 Provider 非官方）；
    * `ANTHROPIC_DEFAULT_OPUS_MODEL` 等（按 Binding.options.model_mapping）；
  * 如果走 Proxy：把 `ANTHROPIC_BASE_URL` 改为 Proxy 提供的 Anthropic-style endpoint（例如 `http://127.0.0.1:7777/anthropic`）。
* 可选：对 `~/.claude/settings.json` 做持久写入：

  * 读取 JSON；
  * 将 Provider/Binding 相关的 env 写入/合并到 `"env"` 字段；
  * 其他字段保持不动；
  * 写前做备份 `settings.json.bak`。

### 5.3 Codex CLI 集成（`codex`）

Codex CLI（Claude 的官方文档）要点：

* 全局 config：`~/.codex/config.toml`
* 默认模型：`model="gpt-5.1-codex"` 等
* 支持 `-c key=value` 形式覆盖单次运行的配置。

Boba 的 Codex 集成：

* Provider 层可以维护：

  * `openai-official` 或开放路由（openai-compatible）
  * `vantagecraft-router`（你将来的自建 router）
* Binding 中指定：

  * 绑定 provider；
  * options 中可指定 `model`、`model_provider`。
* `boba run codex ...` 时：

  * 在子进程 env 里写对应 provider 的 key/env（例如 `OPENAI_API_KEY`）；
  * 方式 A（快速实现）：依赖用户已有 `config.toml`，只用 `-c model=...` 做一次性覆盖；
  * 方式 B（控制强一点）：直接更新 `config.toml` 里的 `model` 和 provider 段，并备份旧文件。

### 5.4 Gemini CLI 集成（`gemini`）

Gemini CLI 官方模式：

* API key env：`GEMINI_API_KEY` / `GOOGLE_API_KEY`；
* settings.json 里可以引用 env 中的 key；
* endpoint 默认指向官方 `generativelanguage.googleapis.com`，目前 CLI 不公开自定义 endpoint 的 official API（需要根据现实情况决定支持程度）。

Boba 的 Gemini 集成：

* Provider：`gemini-official`，定义 base_url & env_var；
* Binding：`tool_id: gemini`，`provider_id: gemini-official`；
* `boba run gemini ...` 时：

  * 在子进程 env 中设置 `GEMINI_API_KEY`；
  * 可生成/更新 `~/.gemini/settings.json`，使其引用 env，而不是硬编码 key；
* Proxy：当前阶段视为「不支持强行 proxy」，只做 key 统一管理（避免与官方 CLI 的行为冲突）。

---

## 6. 本地 Proxy 设计（可选但推荐）

### 6.1 目标

Proxy 的目标是：**统一所有 openai/anthropic 风格的 HTTP 请求出口，做到可观测和可控**。

能力包括：

* 统一日志：记录 tool / provider / model / path / request size / response size / latency / status code；
* 统一路由：将 API 调用按策略分发到不同 Provider（官方 / 第三方 router / 内网实例）；
* 统一限流与预算（后续高级功能）。

### 6.2 接口设计

Proxy 作为一个本地服务：

* 地址：`http://127.0.0.1:7777`
* 暴露的主要接口风格：

  * `POST /openai/v1/...`：OpenAI-style
  * `POST /anthropic/v1/...`：Anthropic-style

Provider 的 `base_url` 可以设为：

* 不走 Proxy：

  * OpenAI 官方：`https://api.openai.com/v1`
  * Anthropic 官方：`https://api.anthropic.com`
* 走 Proxy：

  * OpenAI 风格：`http://127.0.0.1:7777/openai/v1`
  * Anthropic 风格：`http://127.0.0.1:7777/anthropic/v1`

Binding 中的 `use_proxy` 决定了 `boba run` 时是否生成 Proxy base_url。

### 6.3 Proxy 内部逻辑（简化版）

* 根据请求路径和 Provider 配置判断上游：

  * `openai` 风格：转发到某个 OpenAI-compatible Provider；
  * `anthropic` 风格：转发到某个 Anthropic-compatible Provider（如 Z.AI / Anthropic 官方）。
* 在转发前后记录 metrics：

  * `tool_id`（从 header 或 `boba run` 附加）；
  * `provider_id`；
  * `model`；
  * `start_time / end_time`；
  * `tokens 推算`（如有）；
  * `status_code`、错误消息。

---

## 7. TUI（Bubble Tea）架构与交互

Boba 使用 Bubble Tea 作为 TUI 框架，**主视图是“工具绑定矩阵 + Run Launcher”，而不是聊天界面**。

### 7.1 模式与根模型

```go
type appMode int

const (
    modeOnboarding appMode = iota
    modeDashboard
)

type rootModel struct {
    mode       appMode
    onboarding OnboardingModel
    dashboard  DashboardModel
}
```

* `modeOnboarding`：首启引导（扫描工具&Provider，建立初始 Binding，选择是否启用 Proxy）。
* `modeDashboard`：日常控制面板（工具列表 + Provider + Proxy 状态 + 快捷 Run）。

### 7.2 首启引导（Onboarding）

流程：

1. 扫描本机 Tool：

   * 检查 `claude/codex/gemini` 命令是否存在；
   * 检查对应配置文件是否存在。

2. 扫描 Provider：

   * 从 env / 外部配置文件（如 `~/.claude/settings.json`）中推断已经存在的 key/base_url；

3. TUI 展示一个「工具配置向导」：

   ```text
   首次启动 BobaMixer - 检测到以下工具：

   Tool         检测结果              建议操作
   ────────────────────────────────────────────────
   codex        已安装（config.toml）  [绑定 Provider]
   claude       已安装（settings.json）[绑定 Provider]
   gemini       未检测到               [跳过]

   [Enter] 继续
   ```

4. 用户依次为选中的 Tool 绑定 Provider，并选择是否启用 Proxy；

5. 向导完成后，`OnboardingModel` 发出 `OnboardingDoneMsg`，切换到 `DashboardModel`。

### 7.3 Dashboard 主界面

示意：

```text
BobaMixer - Agent CLI Control Plane

Tool      Provider                  Model             Proxy   快捷操作
───────────────────────────────────────────────────────────────────────
codex     OpenAI (official)         gpt-5.1-codex     on      [R]un  [B]ind
claude    Claude via Z.AI           glm-4.6           on      [R]un  [B]ind
gemini    Gemini (Google)           gemini-2.0        off     [R]un  [B]ind

Proxy: enabled at http://127.0.0.1:7777    [P] Proxy 设置   [?] 帮助   [Q] 退出
```

交互：

* 选择一行按 `B`：

  * 打开 Binding 编辑视图：选择 Provider / 模型 / 是否走 Proxy；
* 选择一行按 `R`：

  * 触发 `boba run <tool>`，可以：

    * 在 TUI 内新开一个 panel 显示子进程输出；
    * 或直接在当前终端执行（交互由底层 CLI 接管）；
* `P`：进入 Proxy 的配置视图（端口、日志位置等）。

---

## 8. 实施阶段划分（建议）

### Phase 1：核心控制能力（无 Proxy）

* 完成 `providers.yaml` / `tools.yaml` / `bindings.yaml` 的解析；
* 实现以下命令：

  * `boba providers`
  * `boba tools`
  * `boba bind <tool> <provider>`
  * `boba run <tool> [args...]`
  * `boba doctor`
* 完成 Claude / Codex / Gemini 的基础 env 注入逻辑；
* 完成 Bubble Tea 的最小 Dashboard：显示 Tool ↔ Provider，支持 `B` 编辑绑定。

### Phase 2：Proxy + 监控

* 实现 `boba proxy serve`（OpenAI/Anthropic-style 转发）；
* Binding 支持 `use_proxy`；
* `boba run` 支持通过 Proxy 出口；
* 在 Proxy 内部实现基础 usage 记录（写 SQLite 或 log 文件）。

### Phase 3：高级功能 & 深度集成

* Proxy 内部增加：

  * 路由策略（根据 cost/latency/provider 权重等）；
  * 简单预算/限流机制；
* `boba stats` / `boba report` 等高级命令；
* Git hooks / 项目维度的策略；
* TUI 内增加：

  * Stats 视图；
  * Routing 视图；
  * Budget 视图。

---

## 9. 非目标（当前明确不做）

为避免 scope creep，本版本明确不做：

1. 在 Boba 内提供完整的 Chat / Coding UI（这属于其它产品形态）；
2. 直接替代各家 CLI 原生配置全部功能（只做 Provider 相关的必要修改）；
3. 在 v1 阶段支持所有小众 Provider / 全部 CLI 的深度集成（先打通 Claude/Codex/Gemini + 若干 Router）；
4. 跨机器/团队的集中配置下发与 RBAC 管理（留给未来“团队版/云版”设计）。

---

这份文档可以作为当前的「设计基准」。
* 建一个 `spec/tasks/core-task.md` 放这份内容的任务拆解之后的工作列表；
* 然后在 issue / PR 里引用其中的章节，逐步去实现对应模块（比如 `internal/run`, `internal/proxy`, `internal/tui/dashboard` 等）。

