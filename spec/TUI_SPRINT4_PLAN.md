# TUI Sprint 4 Plan

目标：在保持现有 Phase 1-3 功能的基础上，进一步提升可发现性、交互深度与后端准确性，让控制平面的体验更连贯、更可信。

范围包含三条主线：导航与发现、交互式编辑、Proxy/Routing/Hooks 的数据闭环。

---

## 1. 导航与发现（Navigation & Discoverability）

| 子任务 | 描述 | 产出 | 验收标准 |
|--------|------|------|----------|
| 视图分组 | 将 13 个顶层视图压缩为 4-6 个领域入口（例如 Dashboard、Control Plane、Usage、Optimization、DevOps），内部通过 list/tab 切换子模块。 | 新的 `viewMode` 枚举与视图调度逻辑；更新帮助文案。 | 数字快捷键 1-6 对应领域视图，Tab 循环遵循新顺序；各子模块仍可访问。 |
| 状态栏与帮助 | 在底部统一显示当前视图可用快捷键；实现 `?` 弹窗集中呈现所有全局/局部操作；为分组后的子视图补充文案。 | Footer 模块、`renderHelpOverlay` 逻辑。 | 任意视图按 `?` 可看到帮助弹窗，Esc 关闭；底部提示实时更新。 |
| 搜索/过滤 | 在列表型视图（Providers/Tools/Sessions 等）启用 `/` 搜索过滤；Tab 或箭头键在搜索输入与列表之间切换。 | 复用 bubbles/textinput 实现的搜索框。 | 输入关键字后只显示匹配条目，Esc 清空；不破坏原有导航。 |
| 主题与布局统一 | 调整 lipgloss 主题，使 Dashboard/TUI 旧视图与 Control Plane 新视图风格一致；处理最小宽度提示。 | 主题配置、全局空态文案。 | 80 列下仍可用；所有视图标题/分隔符一致。 |

---

## 2. 交互式编辑能力（Inline Editing Flows）

| 子任务 | 描述 | 产出 | 验收标准 |
|--------|------|------|----------|
| Secrets 视图增强 | 使用 bubbles/textinput 提供 `Set API key`、`Remove`、`Test` 操作；输入框支持 masked、长度校验。 | `handleSecretInput`、`SecretFormState` 等辅助结构。 | 按 `s` 弹出输入框，Enter 保存；错误提示以 toast/消息形式出现。 |
| Providers 编辑 | 在 Control Plane 子视图中支持 `a` 新增、`e` 编辑 Provider：字段包括 display_name、base_url、default_model、api_key source 等；保存前校验。 | Form 流程 + YAML 写回逻辑（沿用 core 配置写入）。 | 操作后 `providers.yaml` 更新，视图刷新；校验失败时提示。 |
| Bindings 编辑 | 支持在 TUI 中编辑绑定：选择 Provider、model、proxy 开关，必要时可新建绑定；引用已有 providers/tools。 | 简化选择器（list+enter）；binding 更新逻辑。 | 成功保存后 Dashboard 行同步变更；proxy 切换即时显示。 |
| 工具路径管理 | Tools 列表中允许 `r` 重新检测、`e` 编辑路径（文本输入，自动补全可选）；保存后触发检测。 | CLI 兼容的检测函数 + TUI 表单。 | 修改路径后 `tools.yaml` 更新且状态刷新。 |
| 轻量 Form 基础 | 在多处使用 textinput/list/modal 后，如确实出现重复，再抽取最小共用层（例如 `PromptModal`）。 | 可选：`ui/components/form.go`。 | 仅在观察到重复模式后实现，避免过早抽象。 |

---

## 3. Proxy / Routing / Hooks 的数据闭环（Backend Alignment）

| 子任务 | 描述 | 产出 | 验收标准 |
|--------|------|------|----------|
| Proxy 状态与日志 | 将 Proxy 运行信息（端口、请求数、最近错误）写入 sqlite/logs，TUI 读取后实时刷新；可选提供 `Start/Stop` 调用。 | Proxy telemetry 结构 + TUI 读取逻辑。 | Proxy 视图显示真实状态；若 Proxy 未运行，支持一键启动或指引。 |
| Routing 决策真实数据 | Routing 视图从 `routes.yaml` + 最近路由日志生成示例，运行 `boba route test` 时将结果写入 DB 供 TUI 展示；可提供最近规则命中统计。 | Routing 日志 schema、`routing.Engine` 更新、TUI 渲染。 | 输入文本后能看到真实匹配的 profile/rule/explanation；统计信息与 CLI 一致。 |
| Hooks Telemetry | Hooks 安装后把事件写入 sqlite；TUI 视图展示当前 repo 状态与最近事件；提供 install/uninstall 快捷操作（调用 `boba hooks`）。 | hooks manager 更新 + TUI UI。 | TUI 中能检测当前仓库并显示 Hooks 状态；执行 install/uninstall 后状态同步；最近活动列表不为空时显示。 |
| 文档同步 | 更新 `spec/TUI_ENHANCEMENT_PLAN.md` 与用户 Facing 文档，说明 TUI 已覆盖的功能、CLI 转换路径、剩余高级命令。 | 文档 PR。 | README/帮助信息与实际功能一致；CLI 输出提示用户首选 TUI。 |

---

## 时间与里程碑（建议）

| Sprint | 目标 | 主要交付 |
|--------|------|----------|
| Sprint 4-A | 完成视图分组 + 帮助/搜索 | 新导航结构、`?`、`/` |
| Sprint 4-B | Secrets + Providers + Bindings 编辑流程 | 三大编辑流可在 TUI 内完成 |
| Sprint 4-C | Proxy/Routing/Hooks 数据闭环 + 文档 | Proxy 状态可信、Routing/Hook 视图展示真实数据 |

每个 Sprint 结束前需运行 `go build ./...`、`go test ./...`、`golangci-lint run`，并在 TUI 中手动验证 80/120/200 列下的渲染效果。

---

## 成功标准

1. 用户无需记忆大量快捷键即可通过顶层视图进入目标功能，`?` 提示覆盖所有操作。
2. 核心 Control Plane 配置（Provider/Tool/Binding/Secret）在 TUI 内即可新增/编辑/测试，无需回落 CLI。
3. Proxy/Routing/Hooks 视图显示的都是可信实时数据，且与 CLI 输出一致。
4. 文档与帮助明确宣告「TUI 优先、CLI 辅助」的最新状态，避免用户混淆。

达到这些标准后即可认为 Sprint 4 完成，进入后续更细分的优化阶段（例如高级过滤、主题自定义、自动化工作流等）。
