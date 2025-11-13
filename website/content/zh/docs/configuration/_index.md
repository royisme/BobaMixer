---
title: "配置"
linkTitle: "配置"
weight: 3
description: >
  BobaMixer 配置文件完整参考。
---

BobaMixer 使用存储在 `~/.boba/` 中的 YAML 配置文件。本节提供所有配置选项的完整参考。

## 配置文件

BobaMixer 使用四个主要配置文件：

1. **profiles.yaml** - AI 提供商配置
2. **routes.yaml** - 路由规则和策略
3. **secrets.yaml** - API 密钥和敏感数据
4. **pricing.yaml** - 模型成本信息

此外，项目特定的设置可以存储在项目目录中的 `.boba-project.yaml` 中。

## 文件位置

### 全局配置

```
~/.boba/
├── profiles.yaml
├── routes.yaml
├── secrets.yaml
├── pricing.yaml
└── usage.db
```

### 项目配置

```
/path/to/your/project/
└── .boba-project.yaml
```

## profiles.yaml

定义 AI 提供商配置，包括模型设置、适配器类型和成本。

### 结构

```yaml
default_profile: gpt4-mini  # 可选：默认配置文件键

profiles:
  - key: gpt4-mini           # 唯一标识符
    model: gpt-4o-mini       # 模型名称
    adapter: http            # 适配器类型：http、tool 或 mcp

    # 适配器特定配置
    http:
      endpoint: https://api.openai.com/v1/chat/completions
      method: POST
      headers:
        Authorization: "Bearer {{secret://OPENAI_API_KEY}}"
        Content-Type: application/json
      body_template: |
        {
          "model": "{{.Model}}",
          "messages": [{"role": "user", "content": "{{.Text}}"}]
        }
      response_path: choices.0.message.content
      usage_input_path: usage.prompt_tokens
      usage_output_path: usage.completion_tokens

    # 成本配置
    cost_per_1k_input: 0.00015
    cost_per_1k_output: 0.0006

    # 预算限制（可选）
    budget:
      daily: 5.0    # 每日限额（美元）
      monthly: 100.0  # 每月限额（美元）
```

### 适配器类型

#### HTTP 适配器

用于 REST API 提供商：

```yaml
- key: claude-sonnet
  model: claude-3-5-sonnet-20241022
  adapter: http
  http:
    endpoint: https://api.anthropic.com/v1/messages
    method: POST
    headers:
      x-api-key: "{{secret://ANTHROPIC_API_KEY}}"
      anthropic-version: "2023-06-01"
      Content-Type: application/json
    body_template: |
      {
        "model": "{{.Model}}",
        "max_tokens": 4096,
        "messages": [{"role": "user", "content": "{{.Text}}"}]
      }
    response_path: content.0.text
    usage_input_path: usage.input_tokens
    usage_output_path: usage.output_tokens
  cost_per_1k_input: 0.003
  cost_per_1k_output: 0.015
```

#### Tool 适配器

用于命令行工具：

```yaml
- key: local-llama
  model: llama-3.1-8b
  adapter: tool
  tool:
    bin: ollama
    args:
      - run
      - llama3.1:8b
    env:
      OLLAMA_HOST: "{{secret://OLLAMA_HOST}}"
    stdin: true
    output_format: raw
  cost_per_1k_input: 0.0
  cost_per_1k_output: 0.0
```

#### MCP 适配器

用于模型上下文协议服务器：

```yaml
- key: mcp-server
  model: custom-model
  adapter: mcp
  mcp:
    command: npx
    args:
      - -y
      - "@modelcontextprotocol/server-filesystem"
      - /tmp
    env:
      NODE_ENV: production
  cost_per_1k_input: 0.001
  cost_per_1k_output: 0.002
```

## routes.yaml

定义基于上下文自动选择配置文件的规则。

### 结构

```yaml
routes:
  - id: large-context          # 唯一路由标识符
    match:
      ctx_chars_gte: 50000     # 上下文大小 >= 50k 字符
    profile: claude-sonnet     # 要使用的配置文件
    explain: "大上下文需要 Claude"

  - id: quick-tasks
    match:
      intent: format           # 意图匹配
      ctx_chars_lt: 1000       # 上下文大小 < 1k 字符
    profile: gpt4-mini
    explain: "快速格式化任务"

  - id: code-review
    match:
      text_matches: "review|PR|pull request"  # 正则表达式模式
      project_types: [go, typescript]         # 项目类型
    profile: claude-sonnet
    explain: "代码审查任务"

  - id: night-hours
    match:
      time_of_day: [night]     # 时间范围：morning, afternoon, evening, night
    profile: local-llama
    explain: "夜间使用本地模型"

  - id: feature-branch
    match:
      branch_matches: "^feature/.*"  # Git 分支模式
    profile: gpt4-mini
    explain: "在功能分支上的开发工作"

# 如果没有路由匹配时的回退配置文件
fallback: gpt4-mini
```

### 匹配条件

`match` 块中的所有条件必须同时满足（AND 逻辑）：

- **ctx_chars_gte**：上下文大小大于或等于值
- **ctx_chars_lt**：上下文大小小于值
- **intent**：匹配意图（format、analysis、chat、code 等）
- **text_matches**：针对输入文本的正则表达式模式
- **project_types**：项目类型列表（go、typescript、python、rust 等）
- **time_of_day**：时间范围（morning: 6-12, afternoon: 12-18, evening: 18-22, night: 22-6）
- **branch_matches**：针对 git 分支名称的正则表达式模式

## secrets.yaml

存储 API 密钥和敏感配置。

```yaml
secrets:
  OPENAI_API_KEY: "sk-..."
  ANTHROPIC_API_KEY: "sk-ant-..."
  OLLAMA_HOST: "http://localhost:11434"
```

**安全性**：
- 文件必须具有 `0600` 权限（仅所有者可读写）
- 切勿将此文件提交到版本控制
- 将 `secrets.yaml` 添加到 `.gitignore`

在配置中使用 `{{secret://KEY_NAME}}` 语法使用秘密。

## pricing.yaml

定义或覆盖模型定价。

```yaml
models:
  gpt-4o-mini:
    input_per_1k: 0.00015
    output_per_1k: 0.0006

  claude-3-5-sonnet-20241022:
    input_per_1k: 0.003
    output_per_1k: 0.015

  gpt-4o:
    input_per_1k: 0.0025
    output_per_1k: 0.010
```

在 `profiles.yaml` 中定义的定价优先于 `pricing.yaml`。

## .boba-project.yaml

放置在项目根目录的项目特定配置：

```yaml
project:
  name: my-awesome-app
  type: go

# 覆盖此项目的全局设置
default_profile: claude-sonnet

# 项目特定预算
budget:
  daily: 10.0
  monthly: 200.0
```

## 环境变量

使用环境变量覆盖配置：

```bash
export BOBA_HOME=/custom/path      # 默认：~/.boba
export BOBA_PROFILE=gpt4-mini      # 覆盖默认配置文件
export BOBA_DEBUG=true             # 启用调试日志
```

## 配置验证

验证您的配置：

```bash
boba doctor
```

这将检查：
- 文件权限
- YAML 语法
- 必需字段
- 秘密引用
- 路由中的配置文件引用

## 下一步

- 了解[路由策略](/docs/routing/)
- 探索[适配器类型](/docs/adapters/)
- 设置[预算追踪](/docs/user-guide/budgets/)
