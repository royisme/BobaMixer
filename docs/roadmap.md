# BobaMixer Roadmap

> 最新状态：**Phase 5 — 路由 DSL 与建议引擎（规划中）**。Phase 0–4 的 DoD 项均已完成，当前聚焦于 Phase 5 的规则解析与建议引擎实现。

---

# Phase 0 — 基线搭建与规范（0.5 周）

**范围**：初始化仓库与约定，保证后续工作有统一脚手架。

## 必须完成（DoD）

* [x] 仓库初始化：`cmd/boba`、`internal/{cli,ui,domain,adapters,store,svc,integration}`、`configs/examples`、`docs/`
* [x] 语言与工具版本固定：`Go 1.22+`、`Makefile`（`make run`, `make test`）
* [x] 代码规范：`golangci-lint` 基线、`pre-commit` 钩子（格式化/静态检查）
* [x] 日志基线：`zap` 结构化日志，默认 info，`~/.boba/logs/` 滚动策略（10MB×5）
* [x] README（Quickstart）骨架 + CONTRIBUTING 草案

## 可选

* [x] CI（GitHub Actions）：lint + unit test
* [x] Issue/PR 模板

## 验证步骤

* [x] `make run` 启动 CLI 输出帮助
* [x] `golangci-lint run` 无阻塞问题
* [x] README 跟随 Quickstart 成功执行

---

# Phase 1 — 配置与本地存储（1 周）

**范围**：四件套配置加载，SQLite 自动引导建表（无迁移工具）。

## 必须完成（DoD）

* [x] 路径约定（XDG）：`~/.boba/`
* [x] 解析并校验：`profiles.yaml`、`routes.yaml`、`secrets.yaml`、`pricing.yaml`
* [x] `secrets.yaml` 权限检查（0600），并支持 `secret://name` → 注入子进程环境变量
* [x] SQLite `usage.db` 自动建表（`PRAGMA user_version=1`），WAL 模式
* [x] `store` 层仓储接口：`SessionRepo`、`UsageRepo`、`BudgetRepo`

## 可选

* [x] `.boba-project.yaml` 项目级配置发现（从 CWD 向上查找）（`config.FindProjectConfig` 已提供）

## 验证步骤

* [x] 缺失任意配置文件时给出可读的错误与示例生成命令
* [x] 首次运行自动创建 `usage.db` 且 `user_version=1`
* [x] 人为写入 1 条 usage 后，`SELECT * FROM v_daily_summary` 能返回汇总

---

# Phase 2 — CLI 基线与 TUI 主屏（1 周）

**范围**：最小命令集与 TUI 总览。

## 必须完成（DoD）

* [x] `boba ls --profiles|--adapters` 显示可用配置
* [x] `boba use <profile>` 激活 profile
* [x] `boba stats --today` 展示今日 tokens/cost/sessions
* [x] 启动 `boba` 进入 TUI 主屏：当前 profile、今日用量卡片、基本趋势字符图
* [x] `boba edit profiles|routes|pricing|secrets` 打开系统编辑器

## 可选

* [x] TUI 主题浅/深切换（奶茶风）

## 验证步骤

* [x] 切换 profile 后，TUI 实时显示当前激活项
* [x] 人工插入 1 日统计数据，`stats --today` 与主屏一致

---

# Phase 3 — HttpAdapter（1–1.5 周）

**范围**：对接 1 个 HTTP 提供商（Anthropic 或 OpenRouter）形成端到端闭环；记录 usage。

## 必须完成（DoD）

* [x] `HttpAdapter`：POST 调用、headers 注入（从 profile/env 与 secrets 解析）
* [x] usage 采集优先级：**响应 usage 字段 > 模型映射估算 > 启发式估算**（落库带 `estimate_level`）
* [x] 失败策略：指数退避（≤2 次），失败记 event
* [x] `boba doctor`：连通性与密钥自检（HTTP 200/401/403/超时分类）

## 可选

* [x] 支持 2 家 HTTP 提供商，以比较 usage 字段差异

## 验证步骤

* [x] 配置有效 key 后，完成一次真实 API 调用（可最小 payload）
* [x] `usage_records` 生成一条 exact 或估算 usage，TUI 今日统计增加

---

# Phase 4 — ToolAdapter（1 周）

**范围**：包装 1 个可执行 CLI（如 `claude-code` 或 `codex`）。

## 必须完成（DoD）

* [x] 以子进程方式运行 CLI，可向 stdin 传入 payload
* [x] 捕获 stdout/stderr；进程退出码判定 success
* [x] 若工具输出 JSON Lines usage 事件：解析为 exact；否则估算
* [x] 子进程仅继承必要 env（由 `secrets.yaml` 解析注入）

## 可选

* [x] 对无 usage 输出的 CLI，提供“输出抓取正则”配置以改进估算

## 验证步骤

* [x] 用假 CLI（测试夹具）模拟 JSONL usage，标记 exact
* [x] 切换为真实 CLI（若可用），能完成一轮执行并落库 usage

---

# Phase 5 — 路由 DSL 与建议引擎（1.5–2 周）

**范围**：规则命中、解释与轻量探索（不自动应用、不熔断）。

## 必须完成（DoD）

* [ ] 规则解析：`intent/ctx_chars/text.matches()/time_of_day/project_types/branch`
* [ ] 命中策略：声明顺序短路；记录 `rule_id` 与 `explain`
* [ ] `boba route test "<text|@file>"` 离线评估命中与解释
* [ ] 探索（Epsilon-Greedy 3% 默认，可关闭），记录 explore 标记
* [ ] 建议生成：基于近 7/30 天单位成功成本与 P95 延迟，输出“替换建议 + 置信度”

## 可选

* [ ] 项目级指标（同一 project/branch 上下文拆分比较）

## 验证步骤

* [ ] 构造三类文本（format/analysis/mixed），`route test` 输出可解释命中
* [ ] 通过造数对比两个 profile 的单位成功成本，TUI 弹出建议卡片

---

# Phase 6 — 价格表在线更新与回退（0.5–1 周）

**范围**：远程 JSON 源优先，本地回退；缓存与刷新策略。

## 必须完成（DoD）

* [ ] 支持 `pricing.yaml` 中 `sources`（按 priority 拉取 http-json / file）
* [ ] 成功拉取后缓存到 `~/.boba/pricing.cache.json`（含时间戳）
* [ ] 刷新策略：启动时刷新、后台 24h 定时；失败回退至本地 JSON → `pricing.yaml` → `profiles.cost_per_1k`
* [ ] `boba doctor` 增加价格源自检项

## 可选

* [ ] 价格源校验（字段完整性、模型名标准化映射）

## 验证步骤

* [ ] 断网场景仍能从缓存/本地读取价格
* [ ] 替换远程 JSON 后，24h 内或手动刷新时，统计成本变化

---

# Phase 7 — 预算提示与 TUI 仪表盘增强（0.5–1 周）

**范围**：提示型预算（不熔断）；TUI 趋势与分布。

## 必须完成（DoD）

* [ ] 预算配置：全局/项目层 `daily_usd`、`hard_cap`（仅提示与建议，不阻断）
* [ ] 超阈提示：TUI 顶部状态条提示“今日/周期接近上限”
* [ ] TUI 增强：7/30 天成本趋势字符图、Profile 占比条、P95 延迟显示

## 可选

* [ ] `boba budget --set daily 5 --project foo` 命令行配置

## 验证步骤

* [ ] 通过造数逼近阈值，TUI 出现清晰提示；`stats --7d` 与 TUI 一致

---

# Phase 8 — Shell/Git 集成与项目发现（0.5–1 周）

**范围**：轻集成，默认不改 PS1。

## 必须完成（DoD）

* [ ] `boba hooks install|remove`：安装 `post-checkout` 提示（仅提示，不自动切换）
* [ ] shell 补全：bash/zsh/fish
* [ ] 项目发现：向上查找 `.boba-project.yaml` 并合并配置

## 可选

* [ ] 可选 PS1 片段脚本（显示当前 profile），默认不启用

## 验证步骤

* [ ] 切换分支触发提示（展示建议 profile）
* [ ] 终端补全生效；在含项目文件的仓库中 `boba use` 推荐值变化

---

# Phase 9 — 稳定化与发行（0.5–1 周）

**范围**：打包发布、文档与示例完善。

## 必须完成（DoD）

* [ ] `goreleaser`：macOS amd64/arm64、Linux amd64/arm64 产物
* [ ] 文档：README 完整版、Adapter 指南、Routing Cookbook、Ops（备份/清理）、FAQ
* [ ] 示例：`configs/examples/` 四件套模板 + 演示项目（ts/go/python 各 1 文件）
* [ ] 版本化：`CHANGELOG.md`、语义化版本标签

## 可选

* [ ] Homebrew Tap 与 Linux 包（.deb/.rpm）

## 验证步骤

* [ ] 本机仅通过二进制即可跑通 Quickstart
* [ ] 文档按步骤拉起 HttpAdapter/ToolAdapter 最小示例

---

# Phase 10 — 扩展与后续（可选）

**范围**：非首发必需的增强能力。

## 可选清单

* [ ] 第二/第三家 HttpAdapter（DeepSeek/自建推理网关）
* [ ] 更丰富的 Tokenizer 映射与估算回归校正
* [ ] `route test` 支持 `git diff` 作为上下文样例
* [ ] MCP Adapter（作为边车订阅执行事件）
* [ ] 导出统计为 CSV/JSON，按标签（task_type/profile）过滤
* [ ] TUI 可交互应用建议（批量操作）

---

# 全局验收准则（Definition of Done, 横切）

* [ ] 所有用户可见错误均为**可读的人类语言**，包含“如何修复”的指引
* [ ] 不在日志与数据库中存储**请求正文**与**API Key**；仅落用量与元数据
* [ ] `secrets.yaml` 权限检查严格；无则阻止执行并提示
* [ ] 单元测试 ≥ 60%（核心解析/计算模块 ≥ 80%），集成测试覆盖两个 Adapter 路径
* [ ] `boba doctor` 全绿才标记为“可发布”状态
* [ ] `boba stats --today/--7d` 与 TUI 显示一致（同一查询口径与取整逻辑）

---

# 阶段性交付物总览（Checklist）

* [ ] 代码：每阶段完成后一个可编译运行的“增量里程碑”标签
* [ ] 文档：每阶段更新对应章节（安装/配置/演示/故障排查）
* [ ] 模板：四件套配置与示例持续可用并随功能增长
* [ ] 演示脚本：每阶段一个 3–5 分钟内可复现的 Demo 脚本（命令序列 + 期望输出）
