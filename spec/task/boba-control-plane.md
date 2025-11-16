下面我把刚才那份架构文档拆成了一个「可执行的任务列表」，每个任务都带有开发目标和 Review/验收标准。你可以直接把它当成 backlog / roadmap，用 issue tracker 去落地（比如 GitHub Projects）。

我按阶段分：Phase 1 → Phase 2 → Phase 3。
Phase 1 是"必须先做完才能真正用起来"的部分，会写得最细。

---

## 🎯 进度总结

### Phase 1 状态：✅ **已完成**（Claude Only）

**完成时间**: 2025-11-16
**提交**: `2bd10d4` - feat: complete Phase 1 control plane implementation
**分支**: `claude/redesign-tui-onboarding-017wA8dpXTCb5qoWvaeSNngn`

**调整说明**：
- Phase 1 专注于 **Claude Code CLI** 集成，奠定架构基础
- Epic 4 (Codex) 和 Epic 5 (Gemini) 推迟到 Phase 1.5
- 新增：完整的交互式 Onboarding 向导（未在原任务列表中）

**已交付功能**：
- ✅ 完整的 Domain 模型 (Provider/Tool/Binding/Secrets)
- ✅ 4 个配置文件的加载/保存/校验 (providers.yaml, tools.yaml, bindings.yaml, secrets.yaml)
- ✅ 5 个 CLI 命令 (`providers`, `tools`, `bind`, `run`, `doctor`)
- ✅ Runner 系统 + ClaudeRunner (env 注入: ANTHROPIC_API_KEY, ANTHROPIC_BASE_URL)
- ✅ 交互式 Onboarding 向导 (工具扫描、Provider 选择、API Key 输入)
- ✅ Dashboard TUI (Bubble Tea table 组件，支持查看/编辑绑定/运行工具)

**技术亮点**：
- 使用 Bubble Tea 官方组件 (table, list, textinput, spinner)
- 遵循 CLAUDE.md UI/UX 标准 (研究优先、设计先行、现代 TUI 体验)
- Runner 注册表模式，易扩展到其他 Provider
- 环境变量优先级：env > secrets.yaml
- 安全的密钥管理 (0600 文件权限)

**下一步**: Phase 1.5 或 Phase 2

---

## Phase 1 — 核心控制平面（无 Proxy）

目标：
在不实现本地 Proxy 的前提下，让 Boba 具备：

* 能读写 `providers.yaml / tools.yaml / bindings.yaml / secrets.yaml`
* 能用 `boba run <tool>` 实际影响 `claude / codex / gemini` 这些 CLI 的行为
* 有一个最小 TUI Dashboard 让你看清楚「Tool → Provider」绑定并编辑

### Epic 1：Domain & 配置加载 ✅

**P1-E1-1：定义 Domain 类型与配置 Schema** ✅

* 内容：

  * 在 `internal/domain` 或类似目录中定义核心类型：

    * `Provider` / `Tool` / `Binding` / `SecretsStore`
  * 明确各字段的含义与约束（对照架构文档）。
* 验收点（Review）：

  * 有一个集中定义（一个或几个文件），不在项目各处散落 map[string]any。
  * 类型覆盖文档中所有关键字段（id、kind、base_url、env_var、config_type、config_path、use_proxy 等）。
  * 单元测试可以构造这些类型实例，编译无警告（TS/Go idiomatic）。

---

**P1-E1-2：实现 `providers.yaml` 的加载与校验** ✅

* 内容：

  * 从 `~/.boba/providers.yaml` 读取配置。
  * YAML → 强类型 `Provider` slice。
  * 基本校验：

    * `id` 唯一；
    * `kind` 在支持枚举内；
    * `api_key.source` 和 `env_var` 字段合法。
* 验收点：

  * 提供一个简单测试 YAML 文件，`boba providers` 可以打印出所有 Provider。
  * 校验失败时会给出清晰错误（哪一行 / 哪个 id 出的问题）。

---

**P1-E1-3：实现 `tools.yaml` 的加载与校验** ✅

* 内容：

  * 加载 `~/.boba/tools.yaml` → `Tool` 列表。
  * 校验 `exec` 是否在 PATH（不强制，但至少可以给 warning）。
* 验收点：

  * `boba tools` 可以列出：

    * 工具 id；
    * 实际 exec 名；
    * config_type / config_path；
  * 如果手动写错 `config_type`，能得到明确错误，不是 panic。

---

**P1-E1-4：实现 `bindings.yaml` 的加载与校验** ✅

* 内容：

  * 加载 `bindings.yaml` → `Binding` 列表。
  * 校验：

    * `tool_id` 必须存在于已加载的 Tools；
    * `provider_id` 必须存在于 Providers。
* 验收点：

  * `boba doctor` 在 bindings 不合法时会报告“某 binding 引用了不存在的 tool/provider”。

---

**P1-E1-5：实现 `secrets.yaml` + env 优先级策略** ✅

* 内容：

  * 加载 `secrets.yaml` → map[provider_id]Secret。
  * 实现统一方法：

    * `ResolveAPIKey(provider Provider) (string, error)`：

      1. 如果 provider.api_key.source == "env" → 读 env；
      2. 否则读 secrets.yaml；
  * 预留扩展点：未来支持 `boba secrets set ...`。
* 验收点：

  * 单元测试覆盖：

    * env 中有 key → 优先；
    * env 无 key 但 secrets 有 → 用 secrets；
    * 都没有 → 返回明确错误。
  * `boba doctor` 可以检测到某 Provider 缺 key 并提示“你可以通过 env 或 secrets.yaml 填充”。

---

### Epic 2：核心 CLI 命令（providers / tools / bind / run / doctor）✅

**P1-E2-1：实现 `boba providers` 命令** ✅

* 内容：

  * 输出 Provider 列表：id / display_name / kind / base_url / enabled / key 状态。
* 验收点：

  * 常用格式：简单 table 输出。
  * 能显示“Key: env / secrets / missing” 这样的标记。

---

**P1-E2-2：实现 `boba tools` 命令** ✅

* 内容：

  * 输出 Tool 列表：id / exec / config_type / config_path / 本地是否存在。
* 验收点：

  * 对不存在的 exec（PATH 中找不到）会在一列中显示 “missing”。

---

**P1-E2-3：实现 `boba bind <tool> <provider> [--proxy=on|off]`** ✅

* 内容：

  * CLI 命令读取当前 bindings，更新或新增一条 binding。
  * 写回 `bindings.yaml`，保持格式整洁。
* 验收点：

  * 可以通过：

    * `boba bind claude claude-zai --proxy=on`
    * 再 `boba bindings`（或者 `boba tools` 中附带显示当前 provider）验证更新。

---

**P1-E2-4：实现 `boba doctor`（基础版）** ✅

* 内容：

  * 对每个 Provider：

    * 检查 key 是否存在（env/secrets）；
  * 对每个 Binding：

    * 检查 tool/ provider 有效；
    * 检查 tool.exec 是否能找到。
* 验收点：

  * `boba doctor` 在健康状态时输出“OK”的 summary；
  * 当有错误时，会给出分项报告（哪一类问题）。

---

### Epic 3：`boba run` 核心管线 + Claude 集成 ✅

**P1-E3-1：定义 `Runner` 抽象与执行上下文** ✅

* 内容：

  * 创建类似：

    ```go
    type RunContext struct {
        Tool      Tool
        Binding   Binding
        Provider  Provider
        Env       map[string]string // 最终子进程的 env override
        Args      []string
    }

    type Runner interface {
        Prepare(ctx *RunContext) error
        Exec(ctx *RunContext) error
    }
    ```

  * 并建立一个 registry：`map[ToolKind]Runner`。
* 验收点：

  * 支持根据 Tool.kind （claude/codex/gemini）选择不同 Runner。
  * 单元测试可以构造一个 fake Tool/Provider/Binding，跑 `Prepare` 输出 Env。

---

**P1-E3-2：实现 Claude Runner（env 注入逻辑）** ✅

* 内容：

  * 具体规则：

    * 根据 Provider.kind：

      * 官方 Anthropic：`ANTHROPIC_API_KEY` / base URL `https://api.anthropic.com`；
      * Z.AI：`ANTHROPIC_AUTH_TOKEN` + `ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic`；
    * 读取 Binding.options.model_mapping，生成 `ANTHROPIC_DEFAULT_*_MODEL` 系列 env（如有配置）。
  * 仅在子进程 env 注入，不修改 `~/.claude/settings.json`（改文件留给后续版本）。
* 验收点：

  * `boba run claude --version` 时：

    * 可以在 debug 模式打印出将注入的 env；
    * 实际启动的 `claude` 在正确 Base URL 下能成功访问（你本机测试）。

---

**P1-E3-3：实现 `boba run` 顶层命令** ✅

* 内容：

  * 实现命令：

    * 解析 `<tool>` + `[args...]`；
    * 从 configs + bindings 解析出 Tool/Provider/Binding；
    * 调用对应 Runner.Prepare → Runner.Exec。
  * Exec 行为：

    * 使用 `os/exec` 或等价方式启动子进程（命令=Tool.exec，参数=args，env override）。
* 验收点：

  * 最小 demo：

    * `boba bind claude claude-anthropic-official`，并确保 env 有 `ANTHROPIC_API_KEY`；
    * `boba run claude --version` → 能正常运行且使用的是 Anthropic 官方 API；
  * 修改 binding 为 `claude-zai` 并设置 Z.AI key：

    * `boba run claude SOME_CMD`时，子进程 env 中 `ANTHROPIC_BASE_URL` 指向 Z.AI。

---

### Epic 4：Codex Runner 集成（基础版）⏸️ **推迟到 Phase 1.5**

**说明**: Phase 1 专注 Claude，为其他 Provider 奠定基础。Codex/OpenAI 集成移到 Phase 1.5。

**P1-E4-1：实现 Codex Runner（env + 可选 config 写入）** ⏸️

* 内容：

  * 从 Provider/Secrets 解出 key → 注入 `OPENAI_API_KEY` 或 Provider 需要的 key env。
  * 先不修改 `~/.codex/config.toml`，只做 env 注入。
  * 预留未来使用 `-c model=...` / 配置文件写入的扩展点。
* 验收点：

  * `boba bind codex openai-official` 后：

    * `boba run codex --version` 可以正常工作；
    * debug 输出中显示 env 包含 `OPENAI_API_KEY`。

---

**P1-E4-2：给 Codex 加最小的 model 覆盖能力（可选）** ⏸️

* 内容：

  * Binding.options.model 存在时，在 `boba run codex` 中自动加 `-c model=<...>` CLI 参数。
* 验收点：

  * 当 Binding 设定不同 model 时，Codex CLI 中 `config show`（如果支持）或请求日志能看到模型变化。

---

### Epic 5：Gemini Runner 集成（基础 env 管理）⏸️ **推迟到 Phase 1.5**

**说明**: Phase 1 专注 Claude，为其他 Provider 奠定基础。Gemini 集成移到 Phase 1.5。

**P1-E5-1：实现 Gemini Runner（env 注入）** ⏸️

* 内容：

  * 从 Provider 获取 key → 注入 `GEMINI_API_KEY` 或 `GOOGLE_API_KEY`。
  * 不尝试 proxy，只做 key 统一管理。
* 验收点：

  * `boba bind gemini gemini-official` 后：

    * `boba run gemini --version` 能正常运行；
    * debug 输出 env 中包含正确的 `GEMINI_API_KEY`。

---

### Epic 6：最小 TUI Dashboard（Bubble Tea）✅

**P1-E6-1：框架搭建：rootModel & mode 切换** ✅

* 内容：

  * 建立 `rootModel`，支持至少两种模式：

    * `modeDashboard`（后续可增加 `modeOnboarding`）。
  * `boba` 启动时直接进入 Dashboard。
* 验收点：

  * `boba` 命令可以启动 Bubble Tea TUI，不崩。

---

**P1-E6-2：Dashboard 列出 Tool ↔ Provider** ✅

* 内容：

  * Dashboard 默认视图显示一个表格：

    ```text
    Tool      Provider             Model (optional)   Proxy
    codex     openai-official      gpt-5.1-codex      off
    claude    claude-zai           glm-4.6            off
    gemini    gemini-official      gemini-2.0         off
    ```

  * 数据来源：`tools + bindings + providers`。
* 验收点：

  * 能在 TUI 中上下移动焦点行；
  * 数据反映当前配置（改 bindings 文件后重新启动，视图更新）。

---

**P1-E6-3：Dashboard 支持绑定编辑（只改 bindings）** ✅

* 内容：

  * 选中一行按某个键（如 `B`）：

    * 弹出 Provider 列表（简单 list）；
    * 选择后更新内存中的 Binding，并写回 `bindings.yaml`。
* 验收点：

  * 在 TUI 中换一个 Provider 后：

    * 退出 TUI；
    * `cat bindings.yaml` 可以看到对应 binding 更新；
    * 下次 `boba run <tool>` 使用的是新 Provider。

---

**P1-E6-4：Dashboard 支持一键 Run（调用 `boba run` pipeline）** ✅

* 内容：

  * 选中 tool 行按 `R`：

    * 在 TUI 中触发 `Runner.Exec`；
    * 简单情况下可以在 TUI 控制台下方打印子进程 stdout/stderr。
* 验收点：

  * 选中 `claude` 行按 `R` 等价于在 shell 中敲 `boba run claude`；
  * 行为能被 Provider/Binding 的修改影响。

---

## Phase 2 — Proxy & 监控（OpenAI/Anthropic）

目标：
引入本地 HTTP Proxy，让所有 OpenAI/Anthropic 风格调用可以统一出口，并开始记录 usage。

这里只列大任务，不逐行拆太细，你可以后续按需要再分解。

### Epic 7：HTTP Proxy 服务（最小可用版）

* P2-E7-1：实现 `boba proxy serve`，监听 `127.0.0.1:7777`。
* P2-E7-2：支持 OpenAI-style endpoint：

  * `POST /openai/v1/...` → 转发到对应 Provider 的 `base_url`。
* P2-E7-3：支持 Anthropic-style endpoint：

  * `POST /anthropic/v1/...` → 转发到对应 Provider 的 `base_url`。
* P2-E7-4：在 Proxy 中记录简单日志（Tool/Provider/path/status_code/duration）。

验收点：

* 手动设置：

  * `OPENAI_BASE_URL=http://127.0.0.1:7777/openai/v1`；
  * `ANTHROPIC_BASE_URL=http://127.0.0.1:7777/anthropic/v1`；
* 用简单 curl / claude 调用可以正常转发并在日志中看到记录。

---

### Epic 8：`boba run` 与 Proxy 集成

* P2-E8-1：当 Binding.use_proxy = true 时，`boba run` 自动将 base_url/env 指向 Proxy。
* P2-E8-2：增加 `boba proxy status`，显示 Proxy 是否在运行。
* P2-E8-3：在 TUI Dashboard 中增加 Proxy 状态栏，以及 per-tool 的 Proxy 开关列。

验收点：

* 在 Dashboard 中把某 Tool 的 Proxy 设置为 on：

  * `boba run` 该 Tool 的请求会经过 Proxy，日志中能看到。
* Proxy 关闭时，`boba run` 会给出合理错误提示或自动启动。

---

### Epic 9：Usage 记录与简单统计（基础）

* P2-E9-1：在 Proxy 内存储基础 usage 数据到 SQLite（`usage.db`）。
* P2-E9-2：实现 `boba stats --today/--7d/--30d` 的最简单版本：

  * 按 Tool / Provider 聚合请求次数。
* P2-E9-3：在 TUI 中加一个简单 Stats 视图（可选）。

验收点：

* 压测 / 多次请求后：

  * `boba stats --by-tool` 能输出一份基础汇总表。

---

## Phase 3 — 高级功能与路由/预算

这一部分可以等 Phase 1 + Phase 2 稳定后再规划，简单列大方向：

* 高级路由：

  * `routes.yaml` + `boba route test`；
  * Proxy 按 routes.yaml 对请求分流到不同 Provider。
* Budget 控制：

  * `pricing.yaml` 定义各 Provider 的单价；
  * Proxy 使用 usage + pricing 估算花费；
  * `boba budget` / `boba action --auto` 做超预算提醒。
* Git Hooks 集成：

  * `boba hooks install` 在 repo 中安装预设 hooks，让 commit 过程中可以自动带上一些 Agent 调用控制。

---

## 📋 下一阶段启动 Prompt

### Phase 1.5: OpenAI/Codex & Gemini 集成

**使用以下 prompt 启动下一阶段**：

```
请实现 BobaMixer Phase 1.5：OpenAI/Codex 和 Gemini CLI 集成

背景：
- Phase 1 (Claude only) 已完成，提交在 2bd10d4
- 架构基础已就绪：Runner 系统、配置加载、TUI Dashboard
- 需要为其他 AI Provider 添加支持

本阶段目标：
1. 实现 OpenAI Runner（支持 codex CLI）
2. 实现 Gemini Runner（支持 gemini CLI）
3. 更新 Onboarding 向导支持多 Provider 检测
4. 确保所有 Runner 遵循统一的 env 注入模式

任务列表（按优先级）：

Epic 4: Codex/OpenAI 集成
- [ ] P1-E4-1: 实现 Codex Runner（env 注入 OPENAI_API_KEY）
- [ ] P1-E4-2: 支持 model 覆盖能力（通过 Binding.options.model）
- [ ] 在 Onboarding 中添加 OpenAI Provider 检测
- [ ] 更新 Dashboard 支持 OpenAI Provider 显示

Epic 5: Gemini 集成
- [ ] P1-E5-1: 实现 Gemini Runner（env 注入 GEMINI_API_KEY/GOOGLE_API_KEY）
- [ ] 在 Onboarding 中添加 Gemini Provider 检测
- [ ] 更新 Dashboard 支持 Gemini Provider 显示

技术要求：
- 遵循 ClaudeRunner 的实现模式（参考 internal/runner/claude.go）
- 使用 Runner 注册表模式，保持扩展性
- 所有新 Provider 定义在 providers.yaml 中（参考现有 OpenAI/Gemini 定义）
- 更新 boba doctor 支持新 Provider 检测
- 遵循 CLAUDE.md UI/UX 标准

验收标准：
1. `boba run codex --version` 可以正常工作
2. `boba run gemini --version` 可以正常工作
3. Onboarding 向导可以检测 PATH 中的 codex/gemini CLI
4. Dashboard 显示所有三种 Provider 的绑定状态
5. `boba doctor` 可以诊断 OpenAI/Gemini Provider 配置问题

参考文档：
- spec/boba-control-plane.md - 架构设计
- spec/task/boba-control-plane.md - 任务列表（本文档）
- CLAUDE.md - UI/UX 开发规范
```

---

### Phase 2: HTTP Proxy & 监控系统

**使用以下 prompt 启动 Phase 2**：

```
请实现 BobaMixer Phase 2：HTTP Proxy 服务和使用监控

背景：
- Phase 1 (Claude) 和 Phase 1.5 (OpenAI/Gemini) 已完成
- 目前通过 env 注入直连各 Provider，缺乏统一出口和监控能力
- 需要引入本地 Proxy 实现请求拦截、转发、监控

本阶段目标：
1. 实现本地 HTTP Proxy 服务（127.0.0.1:7777）
2. 支持 OpenAI 和 Anthropic API 转发
3. 记录 usage 数据到 SQLite
4. 提供基础统计命令和 TUI 视图

任务列表：

Epic 7: HTTP Proxy 服务
- [ ] P2-E7-1: 实现 `boba proxy serve` 监听 127.0.0.1:7777
- [ ] P2-E7-2: 支持 OpenAI-style endpoint 转发（/openai/v1/*）
- [ ] P2-E7-3: 支持 Anthropic-style endpoint 转发（/anthropic/v1/*）
- [ ] P2-E7-4: 实现请求/响应日志记录

Epic 8: boba run 与 Proxy 集成
- [ ] P2-E8-1: Binding.use_proxy=true 时自动使用 Proxy
- [ ] P2-E8-2: 实现 `boba proxy status` 命令
- [ ] P2-E8-3: Dashboard 显示 Proxy 状态和 per-tool 开关

Epic 9: Usage 记录与统计
- [ ] P2-E9-1: 设计 usage.db schema（requests, tokens, cost）
- [ ] P2-E9-2: 实现 `boba stats` 命令（--today, --7d, --30d）
- [ ] P2-E9-3: Dashboard 添加 Stats 视图

技术要求：
- 使用 Go 标准库 net/http 实现 Proxy
- 使用 SQLite 存储 usage 数据（database/sql + modernc.org/sqlite）
- Proxy 支持优雅关闭和错误恢复
- 性能要求：单机支持 1000+ RPS

验收标准：
1. `boba proxy serve` 可以在后台运行
2. 通过 Proxy 的请求正常转发到目标 Provider
3. usage.db 正确记录每次请求
4. `boba stats --by-tool` 显示准确的统计数据
5. Dashboard Proxy 开关可以切换并立即生效

参考文档：
- spec/boba-control-plane.md - Proxy 架构设计
- spec/task/boba-control-plane.md - 任务列表（本文档）
```

---

## 🔄 阶段总结

| Phase | 状态 | 完成时间 | 提交 | 核心功能 |
|-------|------|----------|------|----------|
| Phase 1 | ✅ 已完成 | 2025-11-16 | 2bd10d4 | Claude 集成、Domain 模型、CLI 命令、TUI Dashboard |
| Phase 1.5 | ⏸️ 待启动 | - | - | OpenAI/Codex + Gemini 集成 |
| Phase 2 | ⏸️ 待启动 | - | - | HTTP Proxy + Usage 监控 |
| Phase 3 | 📝 规划中 | - | - | 高级路由、预算控制、Git Hooks |

---


