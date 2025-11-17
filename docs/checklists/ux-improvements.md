# BobaMixer UX 改进建议 - 零配置文件编辑

**核心原则**: 用户应该专注于编码和使用 AI 工具，而不是陷入修改各种配置文件！

**设计目标**: 所有核心功能通过 CLI 命令或 TUI 交互完成，只有高级功能才需要编辑配置文件。

---

## 🎯 当前 UX 问题分析

### ❌ 问题 1: README 让用户手动编辑配置文件

**当前 README (不好的体验)**:
```bash
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

**问题**:
- 要求用户理解 YAML 格式
- 要求用户知道 provider ID 的准确名称
- 要求用户手动编辑文件
- 心智负担高，容易出错

---

### ❌ 问题 2: 缺少 CLI 命令管理 secrets

**当前状态**:
- ✅ 有 `boba providers` - 查看 provider
- ✅ 有 `boba tools` - 查看工具
- ✅ 有 `boba bind` - 绑定工具到 provider
- ❌ **缺少** `boba secrets set <provider> <key>` - 设置 API key
- ❌ **缺少** `boba secrets list` - 查看已配置的 secrets

**影响**:
- 用户被迫手动编辑 `~/.boba/secrets.yaml`
- 与工具的"CLI 优先"设计理念不符

---

### ❌ 问题 3: README 示例流程不够流畅

**当前 Quick Start 流程**:
```bash
# 1. 安装
go install ...

# 2. 初始化
boba init

# 3. 配置 API Keys (需要手动编辑文件 ❌)
export ANTHROPIC_API_KEY="sk-ant-..."
# 或者
boba edit secrets  # 打开编辑器，用户需要手动编辑 YAML

# 4. 启动 TUI
boba
```

**问题**:
- 第 3 步要求用户要么设置环境变量，要么编辑 YAML
- 流程不够顺滑
- 没有充分利用 Onboarding 向导

---

## ✅ 改进方案

### 方案 1: 添加 `boba secrets` CLI 命令

#### 1.1 实现 `boba secrets set`

```bash
# 交互式设置（推荐）
$ boba secrets set claude-anthropic-official
Enter API key for claude-anthropic-official: ********
✓ API key saved to ~/.boba/secrets.yaml

# 非交互式（用于脚本）
$ boba secrets set claude-anthropic-official --key "sk-ant-..."
✓ API key saved

# 一次性设置多个
$ boba secrets set openai-official
Enter API key for openai-official: ********
✓ API key saved
```

**实现要点**:
- 使用 `terminal.ReadPassword()` 安全输入
- 自动创建 `~/.boba/secrets.yaml`（如果不存在）
- 自动设置文件权限 0600
- 验证 provider ID 是否存在于 `providers.yaml`

#### 1.2 实现 `boba secrets list`

```bash
$ boba secrets list

Configured Secrets
==================
Provider                        Status    Source
claude-anthropic-official       ✓ Set     secrets.yaml
openai-official                 ✓ Set     env (OPENAI_API_KEY)
gemini-official                 ✗ Missing -
claude-zai                      ✓ Set     secrets.yaml

Legend:
  ✓ Set     - API key configured
  ✗ Missing - API key not found
```

#### 1.3 实现 `boba secrets remove`

```bash
$ boba secrets remove openai-official
✓ Removed API key for openai-official
```

---

### 方案 2: 增强 Onboarding 向导的 API Key 输入

#### 2.1 当前 Onboarding 流程

```
1. 检测工具 (claude/codex/gemini)
2. 选择 Provider
3. [新增] 输入 API Key (如果缺失)
4. 写入配置
```

#### 2.2 改进后的 Onboarding 体验

**Step 1: 工具检测**
```
🔍 Detecting CLI tools...

Found tools:
  ✓ claude  (Claude Code CLI)
  ✓ codex   (OpenAI Codex CLI)
  ✗ gemini  (not found in PATH)

Press Enter to continue...
```

**Step 2: Provider 绑定**
```
📌 Bind 'claude' to a provider:

Available providers:
  1. Claude (Anthropic official)
  2. Claude via Z.AI (GLM-4.6)
  3. Skip

Select provider: 1
```

**Step 3: API Key 检查与输入（关键改进）**
```
🔑 Checking API key for 'claude-anthropic-official'...

Status: ✗ API key not found

Options:
  1. Enter API key now (recommended)
  2. Set environment variable (ANTHROPIC_API_KEY)
  3. Skip (configure later)

Select option: 1

Enter API key: ********

✓ API key saved to ~/.boba/secrets.yaml (permissions: 0600)
```

**Step 4: 完成**
```
✅ Setup complete!

Summary:
  • claude → Claude (Anthropic official)
    API key: ✓ configured
    Proxy: off

Next steps:
  1. Run 'boba run claude --version' to test
  2. Run 'boba' to open dashboard
  3. Press 'R' in dashboard to run tools
```

---

### 方案 3: 重写 README Quick Start

#### 3.1 新的 Quick Start（零配置文件编辑）

```markdown
## Quick Start

### Installation

```bash
# Using Go
go install github.com/royisme/bobamixer/cmd/boba@latest

# Or download from releases
# https://github.com/royisme/BobaMixer/releases
```

### First Time Setup - Interactive Onboarding 🎯

BobaMixer 会自动引导你完成所有配置，**无需手动编辑任何配置文件**：

```bash
# 1. 启动 BobaMixer（首次运行会自动进入向导）
$ boba

# Onboarding 向导会自动：
# ✓ 检测本地 CLI 工具 (claude/codex/gemini)
# ✓ 让你选择 Provider
# ✓ 引导输入 API Key（安全输入，自动保存）
# ✓ 创建所有配置文件
# ✓ 验证配置

# 2. 完成后即可使用
$ boba run claude --version
```

### Alternative: CLI Setup

如果你更喜欢命令行：

```bash
# 1. 初始化配置目录
$ boba init

# 2. 配置 API Key（安全输入）
$ boba secrets set claude-anthropic-official
Enter API key: ********
✓ Saved

# 3. 绑定工具到 Provider
$ boba bind claude claude-anthropic-official

# 4. 验证配置
$ boba doctor

# 5. 运行
$ boba run claude --version
```

### 🚀 That's it! No YAML editing required.

---

## Advanced Configuration (可选)

只有当你需要高级功能时，才需要手动编辑配置文件：

- **Routing rules**: `~/.boba/routes.yaml`
- **Budget limits**: `~/.boba/settings.yaml` (或使用 `boba budget set`)
- **Custom pricing**: `~/.boba/pricing.yaml`
- **Profile settings**: `~/.boba/profiles.yaml`

大部分用户永远不需要碰这些文件。
```

---

### 方案 4: 改进 `boba init` 命令

#### 4.1 当前 `boba init` 行为

```bash
$ boba init

✅ BobaMixer initialized successfully

Configuration directory: ~/.boba

Created files:
  - providers.yaml  (AI service providers)
  - tools.yaml      (Local CLI tools)
  - bindings.yaml   (Tool ↔ Provider bindings)
  - secrets.yaml    (API keys)
  - settings.yaml   (UI preferences)
```

**问题**: 只创建空文件，用户还是不知道下一步做什么。

#### 4.2 改进后的 `boba init`

```bash
$ boba init

✅ BobaMixer initialized

Configuration directory: ~/.boba

Created:
  ✓ providers.yaml  (3 default providers: Anthropic, OpenAI, Gemini)
  ✓ tools.yaml      (ready for auto-detection)
  ✓ bindings.yaml   (empty, use 'boba bind' to create)
  ✓ secrets.yaml    (empty, use 'boba secrets set' to add keys)
  ✓ settings.yaml   (default UI preferences)

Next steps:
  1. Add API keys:
     $ boba secrets set claude-anthropic-official

  2. Bind tools to providers:
     $ boba bind claude claude-anthropic-official

  3. Verify setup:
     $ boba doctor

  4. Or use interactive setup:
     $ boba
```

---

### 方案 5: 添加 `boba quickstart` 命令（一键式设置）

```bash
$ boba quickstart

🚀 BobaMixer Quick Start

This wizard will help you set up BobaMixer in < 2 minutes.
Press Ctrl+C to exit at any time.

Step 1/3: Detecting CLI tools...
  ✓ Found: claude (Claude Code CLI)
  ✗ Not found: codex
  ✗ Not found: gemini

Step 2/3: Configure 'claude'
  Select provider:
    1. Claude (Anthropic official)
    2. Claude via Z.AI
  Choice: 1

  Enter ANTHROPIC_API_KEY: ********
  ✓ Saved

Step 3/3: Test connection
  Testing: boba run claude --version
  ✓ Success! Claude Code CLI is working.

🎉 Setup complete!

You can now:
  • Run: boba run claude [command]
  • Dashboard: boba
  • Stats: boba stats
  • Help: boba --help
```

---

## 📝 实现优先级

### 🔥 P0 - 必须立即实现（影响核心 UX）

1. **实现 `boba secrets set/list/remove` 命令**
   - 工作量: 2-3 小时
   - 优先级: 最高
   - 理由: 这是避免用户手动编辑 YAML 的关键

2. **增强 Onboarding 的 API Key 输入**
   - 工作量: 1-2 小时
   - 优先级: 最高
   - 理由: 首次体验决定用户是否继续使用

3. **重写 README Quick Start**
   - 工作量: 30 分钟
   - 优先级: 最高
   - 理由: 文档是用户的第一印象

### 🔵 P1 - 应该尽快实现（改善 UX）

4. **改进 `boba init` 提示信息**
   - 工作量: 30 分钟
   - 优先级: 高
   - 理由: 提供清晰的下一步指引

5. **添加 `boba quickstart` 命令**
   - 工作量: 2-3 小时
   - 优先级: 中高
   - 理由: 提供最快的上手体验

### 🟢 P2 - 可选实现（锦上添花）

6. **实现 `boba budget set` CLI 命令**
   - 当前: 需要编辑 YAML
   - 改进: `boba budget set --daily 10 --monthly 300`

7. **实现 `boba route add` CLI 命令**
   - 当前: 需要编辑 routes.yaml
   - 改进: `boba route add --if "ctx_chars > 50000" --use claude-opus`

---

## 🎯 改进后的用户旅程

### Journey 1: 新用户首次使用

```
1. 安装: brew install bobamixer
2. 运行: boba
3. Onboarding 自动检测工具 ✓
4. Onboarding 引导选择 Provider ✓
5. Onboarding 引导输入 API Key ✓
6. 完成！立即可用 ✓

总时间: < 2 分钟
编辑配置文件次数: 0 ✅
```

### Journey 2: CLI 爱好者

```
1. 安装: go install ...
2. 初始化: boba init
3. 设置 Key: boba secrets set claude-anthropic-official
4. 绑定工具: boba bind claude claude-anthropic-official
5. 验证: boba doctor
6. 运行: boba run claude --version

总时间: < 1 分钟
编辑配置文件次数: 0 ✅
```

### Journey 3: 高级用户（需要自定义路由）

```
1-5. 同上（基础设置）
6. 编辑路由规则: vi ~/.boba/routes.yaml  # 这是高级功能，可以接受
7. 测试路由: boba route test "large context prompt"

总时间: 5-10 分钟
编辑配置文件次数: 1（仅高级功能）✅
```

---

## 📊 对比：改进前 vs 改进后

| 维度 | 改进前 | 改进后 |
|------|--------|--------|
| **首次上手时间** | 5-10 分钟 | < 2 分钟 |
| **需要编辑的 YAML 文件** | 2-3 个 (providers, secrets, bindings) | 0 个 |
| **需要理解的概念** | Provider, Tool, Binding, YAML 格式 | 只需要选择和输入 |
| **出错可能性** | 高（YAML 格式、ID 名称） | 低（CLI 自动验证） |
| **心智负担** | 高 | 低 |
| **专业感** | 中（需要手动配置） | 高（自动化、引导式） |

---

## 🔧 技术实现建议

### 实现 `boba secrets set`

```go
// internal/cli/secrets.go

func runSecretsSet(home string, args []string) error {
    if len(args) < 1 {
        return fmt.Errorf("usage: boba secrets set <provider-id>")
    }

    providerID := args[0]

    // 1. 验证 provider 存在
    providers, err := core.LoadProviders(home)
    if err != nil {
        return err
    }

    var provider *core.Provider
    for _, p := range providers.Providers {
        if p.ID == providerID {
            provider = &p
            break
        }
    }
    if provider == nil {
        return fmt.Errorf("provider not found: %s\nRun 'boba providers' to see available providers", providerID)
    }

    // 2. 提示用户输入 API key
    fmt.Printf("Enter API key for %s: ", provider.DisplayName)

    // 使用 terminal.ReadPassword 安全输入
    keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        return fmt.Errorf("failed to read password: %w", err)
    }
    fmt.Println() // 换行

    apiKey := string(keyBytes)
    if apiKey == "" {
        return fmt.Errorf("API key cannot be empty")
    }

    // 3. 保存到 secrets.yaml
    secrets, err := core.LoadSecrets(home)
    if err != nil {
        return err
    }

    if secrets.Secrets == nil {
        secrets.Secrets = make(map[string]core.Secret)
    }

    secrets.Secrets[providerID] = core.Secret{
        APIKey: apiKey,
    }

    if err := core.SaveSecrets(home, secrets); err != nil {
        return err
    }

    fmt.Printf("✓ API key saved to ~/.boba/secrets.yaml\n")
    fmt.Printf("  Provider: %s\n", provider.DisplayName)
    fmt.Printf("  File permissions: 0600 (secure)\n")

    return nil
}
```

### 增强 Onboarding API Key 输入

```go
// internal/ui/onboarding.go

type apiKeyInputModel struct {
    provider     *core.Provider
    textInput    textinput.Model
    err          error
}

func (m apiKeyInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyEnter:
            apiKey := m.textInput.Value()
            if apiKey == "" {
                m.err = fmt.Errorf("API key cannot be empty")
                return m, nil
            }

            // 保存到 secrets.yaml
            secrets, _ := core.LoadSecrets(m.home)
            secrets.Secrets[m.provider.ID] = core.Secret{APIKey: apiKey}
            core.SaveSecrets(m.home, secrets)

            // 返回成功消息
            return m, func() tea.Msg {
                return apiKeySavedMsg{providerID: m.provider.ID}
            }
        }
    }

    m.textInput, cmd := m.textInput.Update(msg)
    return m, cmd
}

func (m apiKeyInputModel) View() string {
    s := fmt.Sprintf("🔑 Configure API key for %s\n\n", m.provider.DisplayName)
    s += m.textInput.View() + "\n\n"

    if m.err != nil {
        s += fmt.Sprintf("Error: %s\n\n", m.err)
    }

    s += "Press Enter to save, Ctrl+C to cancel"
    return s
}
```

---

## 📋 检查清单

### Phase 0 改进建议

- [ ] 更新 spec/boba-control-plane.md 强调"零配置文件编辑"原则
- [ ] 在 spec 中标注哪些功能是 Core（不需要编辑配置），哪些是 Advanced（可以编辑配置）

### Phase 1 改进建议

- [ ] ~~已完成~~（Domain 层不需要改动）

### Phase 2 改进建议

- [ ] **实现 `boba secrets set <provider>` 命令**
- [ ] **实现 `boba secrets list` 命令**
- [ ] **实现 `boba secrets remove <provider>` 命令**
- [ ] 改进 `boba init` 的提示信息
- [ ] （可选）实现 `boba quickstart` 一键式设置

### Phase 4 改进建议

- [ ] **增强 Onboarding 的 API Key 输入步骤**
- [ ] 在 Onboarding 完成后显示清晰的"下一步"提示
- [ ] 在 Onboarding 中添加"测试连接"步骤

### Phase 6 改进建议

- [ ] **重写 README Quick Start（零配置文件编辑）**
- [ ] 将"手动编辑配置文件"的示例移到 Advanced Features 部分
- [ ] 添加"用户旅程"示例（展示完整的无缝体验）

---

## 🎯 总结

### 核心原则

1. **Core 功能 = 零配置文件编辑**
   - Control Plane: CLI 命令 + TUI 完成所有操作
   - Proxy: 自动启动，自动配置

2. **Advanced 功能 = 可选配置文件编辑**
   - Routing: routes.yaml
   - Budget: settings.yaml（或 `boba budget set`）
   - Pricing: pricing.yaml

3. **优先级**
   - P0: `boba secrets` 命令（最关键）
   - P0: Onboarding API Key 输入
   - P0: README 重写
   - P1: `boba init` 改进
   - P1: `boba quickstart` 命令

### 预期效果

- **新用户上手时间**: 从 5-10 分钟 → < 2 分钟
- **配置文件编辑次数**: 从 2-3 个 → 0 个（核心功能）
- **用户满意度**: 大幅提升
- **专业感**: 更强（自动化程度高）

---

**文档版本**: v1.0
**创建时间**: 2025-11-17
**优先级**: 🔥 P0 - 立即处理
**预计工作量**: 4-6 小时（核心功能）
