# 配置指南

本综合指南涵盖所有 BobaMixer 配置选项和最佳实践。

## 配置文件概览

BobaMixer 使用 YAML 文件进行配置,存储在 `~/.boba/`:

```
~/.boba/
├── profiles.yaml       # 配置文件定义
├── routes.yaml         # 路由规则
├── pricing.yaml        # 模型定价信息
├── secrets.yaml        # API 密钥 (0600 权限)
├── usage.db            # SQLite 数据库
├── logs/               # 应用程序日志
└── pricing.cache.json  # 缓存的定价数据
```

## 配置文件 (Profiles)

`profiles.yaml` 文件定义你的 AI 提供商配置文件。

### 基本配置文件结构

```yaml
配置文件名称:
  adapter: http|tool|mcp
  provider: 提供商名称
  endpoint: https://api.example.com
  model: 模型名称
  max_tokens: 4096
  temperature: 0.7
  headers:
    Header-Name: value
  env:
    - VAR_NAME=value
  tags: [tag1, tag2]
  cost_per_1k:
    input: 0.01
    output: 0.03
```

### HTTP 适配器示例 (Anthropic)

```yaml
claude-sonnet:
  adapter: http
  provider: anthropic
  endpoint: https://api.anthropic.com/v1/messages
  model: claude-3-5-sonnet-20241022
  max_tokens: 4096
  temperature: 0.7
  headers:
    anthropic-version: "2023-06-01"
    x-api-key: "secret://anthropic_key"
  tags: [工作, 复杂, 分析]
  cost_per_1k:
    input: 0.015
    output: 0.075
```

### Tool 适配器示例

```yaml
claude-code:
  adapter: tool
  command: /usr/local/bin/claude-code
  args: ["--session", "work"]
  env:
    - ANTHROPIC_API_KEY=secret://anthropic_key
  tags: [开发, 编码]
  cost_per_1k:
    input: 0.015
    output: 0.075
```

## Secrets 管理

`secrets.yaml` 文件存储敏感数据如 API 密钥。

### 安全要求

```bash
# 必须具有 0600 权限
chmod 600 ~/.boba/secrets.yaml
```

如果权限不正确,BobaMixer 将拒绝运行。

### Secrets 格式

```yaml
anthropic_key: sk-ant-api03-xxxxx
openai_key: sk-proj-xxxxx
openrouter_key: sk-or-v1-xxxxx
```

### 引用 Secrets

在配置文件中使用 `secret://` 前缀:

```yaml
headers:
  x-api-key: "secret://anthropic_key"
  Authorization: "Bearer secret://openai_key"
```

## 路由配置

`routes.yaml` 文件定义智能路由规则。

### 基本路由结构

```yaml
rules:
  - id: 规则标识符
    if: "条件表达式"
    use: 配置文件名称
    fallback: 备用配置文件
    explain: "为什么存在此规则"

exploration:
  enabled: true
  epsilon: 0.03
```

### 路由示例

```yaml
rules:
  # 大上下文
  - id: extra-large-context
    if: "ctx_chars > 100000"
    use: claude-opus
    fallback: claude-sonnet
    explain: "超大上下文需要最高容量模型"

  # 任务类型
  - id: code-generation
    if: "text.matches('编写.*函数|实现|创建.*类')"
    use: 代码专家
    explain: "代码生成任务"

  # 默认回退
  - id: default
    if: "ctx_chars > 0"
    use: 平衡模型
    explain: "未匹配情况的默认值"
```

## 项目级配置

在项目根目录创建 `.boba-project.yaml` 用于项目特定设置:

```yaml
project:
  name: 我的项目
  type: [typescript, react, nodejs]
  preferred_profiles:
    - 前端专家
    - 快速模型

routing:
  rules:
    - id: 项目特定规则
      if: "text.contains('组件')"
      use: react专家
      explain: "React 组件工作"

budget:
  daily_usd: 5.00
  hard_cap: 100.00
  period_days: 30
  alert_at_percent: 80
  critical_at_percent: 95
```

## 最佳实践

### 1. 使用标签组织配置文件

```yaml
claude-sonnet:
  tags: [工作, 复杂, 生产]

gpt-4o-mini:
  tags: [开发, 测试, 快速]
```

### 2. 使用回退配置文件

```yaml
rules:
  - id: 生产任务
    if: "branch.matches('main|master')"
    use: 最佳模型
    fallback: 良好模型
    explain: "带有回退的生产工作"
```

### 3. 保护 Secrets

```bash
# 永不提交 secrets
echo "secrets.yaml" >> .gitignore

# 定期权限检查
chmod 600 ~/.boba/secrets.yaml
```

### 4. 监控成本

为所有项目设置预算:

```yaml
budget:
  daily_usd: 10.00
  hard_cap: 200.00
  alert_at_percent: 75
```

## 配置验证

### 检查配置健康

```bash
boba doctor
```

### 测试路由规则

```bash
# 使用文本测试
boba route test "你的测试文本"

# 使用文件测试
boba route test @path/to/file.txt
```

## 下一步

- **[适配器](/zh/features/adapters)** - 了解适配器类型
- **[路由](/zh/features/routing)** - 掌握智能路由
- **[预算](/zh/features/budgets)** - 设置预算管理
- **[CLI 参考](/zh/reference/cli)** - 探索所有命令
