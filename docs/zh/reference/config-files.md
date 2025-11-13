# 配置文件参考

所有 BobaMixer 配置文件架构和选项的完整参考。

## 配置目录

默认位置: `~/.boba/`

使用以下方式覆盖: `export BOBA_HOME=/custom/path`

```
~/.boba/
├── profiles.yaml       # 配置文件定义
├── routes.yaml         # 路由规则
├── pricing.yaml        # 模型定价
├── secrets.yaml        # API 密钥 (0600 权限)
├── usage.db            # SQLite 数据库
├── logs/               # 应用程序日志
└── pricing.cache.json  # 缓存的定价数据 (自动生成)
```

## profiles.yaml

定义 AI 提供商配置文件和全局设置。

### 架构

```yaml
# 全局设置 (可选)
global:
  budget:
    daily_usd: float
    hard_cap: float
    period_days: int

# 配置文件定义
profile-name:
  adapter: http|tool|mcp
  provider: string
  endpoint: string
  model: string
  headers:
    header-name: string
  cost_per_1k:
    input: float
    output: float
```

### 示例: HTTP 适配器 (Anthropic)

```yaml
claude-sonnet:
  adapter: http
  provider: anthropic
  endpoint: https://api.anthropic.com/v1/messages
  model: claude-3-5-sonnet-20241022
  headers:
    anthropic-version: "2023-06-01"
    x-api-key: "secret://anthropic_key"
  cost_per_1k:
    input: 0.015
    output: 0.075
```

## routes.yaml

定义智能路由规则。

### 架构

```yaml
rules:
  - id: string
    if: string
    use: string
    fallback: string
    explain: string

exploration:
  enabled: boolean
  epsilon: float
  min_samples: int
```

### 示例

```yaml
rules:
  - id: 大上下文
    if: "ctx_chars > 50000"
    use: claude-sonnet
    fallback: claude-haiku
    explain: "大上下文需要强大模型"

  - id: 代码生成
    if: "text.matches('编写.*函数|实现')"
    use: 代码专家
    explain: "代码生成任务"

exploration:
  enabled: true
  epsilon: 0.03
  min_samples: 10
```

## secrets.yaml

存储敏感数据如 API 密钥。

### 架构

```yaml
secret_name: string
another_secret: string
```

### 安全要求

```bash
# 必须具有 0600 权限
chmod 600 ~/.boba/secrets.yaml
```

### 示例

```yaml
anthropic_key: sk-ant-api03-xxxxxxxxxxxxxxxxxxxxx
openai_key: sk-proj-xxxxxxxxxxxxxxxxxxxxxxxxxx
openrouter_key: sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 引用 Secrets

在其他配置文件中使用 `secret://` 前缀:

```yaml
# 在 profiles.yaml 中
headers:
  x-api-key: "secret://anthropic_key"
  Authorization: "Bearer secret://openai_key"
```

## .boba-project.yaml

项目特定配置 (可选)。

放置在项目根目录。

### 架构

```yaml
project:
  name: string
  type: [string]
  preferred_profiles: [string]

routing:
  rules:
    - id: string
      if: string
      use: string
      explain: string

budget:
  daily_usd: float
  hard_cap: float
  alert_at_percent: int
```

### 示例

```yaml
project:
  name: 我的应用
  type: [typescript, react, nodejs]
  preferred_profiles:
    - 前端专家
    - 快速模型

routing:
  rules:
    - id: 组件工作
      if: "text.matches('组件|JSX|React')"
      use: react专家
      explain: "React 组件工作"

budget:
  daily_usd: 10.00
  hard_cap: 200.00
  alert_at_percent: 80
```

## 配置验证

### 验证所有配置

```bash
boba doctor
```

### 验证 YAML 语法

```bash
yamllint ~/.boba/profiles.yaml
yamllint ~/.boba/routes.yaml
```

### 测试路由规则

```bash
boba route validate
boba route test "示例文本"
```

## 下一步

- **[CLI 参考](/zh/reference/cli)** - 命令行界面
- **[配置指南](/zh/guide/configuration)** - 详细设置指南
- **[适配器](/zh/features/adapters)** - 适配器配置
- **[路由](/zh/features/routing)** - 路由规则指南
