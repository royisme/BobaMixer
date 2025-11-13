---
title: "配置指南实战"
linkTitle: "配置指南"
weight: 3
description: >
  实用配置手册：从基础设置到高级场景，掌握BobaMixer的强大配置能力。
---

# 📚 配置指南实战

**从新手到专家的完整配置手册** - 通过真实场景掌握BobaMixer配置艺术。

---

## 🎯 学习路径

### 🌱 新手配置（5分钟）
- [基础配置模板](#基础配置模板) - 开箱即用的配置
- [常用AI服务商配置](#常用ai服务商配置) - OpenAI、Claude等

### 🚀 进阶配置（15分钟）
- [智能路由规则](#智能路由规则实战) - 自动选择最佳模型
- [预算控制策略](#预算控制配置) - 成本管理最佳实践

### 🏆 高级配置（30分钟）
- [多环境配置](#多环境配置管理) - 开发、测试、生产环境
- [团队协作配置](#团队协作配置) - 企业级使用方案

---

## 📁 配置文件总览

### 全局配置 (`~/.boba/`)
```
~/.boba/
├── profiles.yaml     # AI提供商配置 - ★★★★★ (最重要)
├── routes.yaml       # 智能路由规则 - ★★★★☆ (核心功能)
├── secrets.yaml      # API密钥存储 - ★★★★★ (安全关键)
├── pricing.yaml      # 价格信息 - ★★★☆☆ (可选)
└── usage.db          # 使用数据库 (自动生成)
```

### 项目配置 (项目根目录)
```
my-project/
├── .boba-project.yaml  # 项目特定配置
└── .gitignore          # 记得忽略敏感配置
```

---

## ⚡ 基础配置模板

### 最小可用配置 (复制粘贴即可)

编辑 `~/.boba/profiles.yaml`:

```yaml
# 设置默认使用的AI模型
default_profile: gpt4-mini

# AI提供商配置
profiles:
  # OpenAI GPT-4o-mini - 经济实惠的主力模型
  - key: gpt4-mini
    model: gpt-4o-mini
    adapter: http
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
    cost_per_1k_input: 0.00015
    cost_per_1k_output: 0.0006
```

编辑 `~/.boba/secrets.yaml`:
```yaml
secrets:
  OPENAI_API_KEY: "sk-your-openai-api-key-here"
```

**设置安全权限**:
```bash
chmod 600 ~/.boba/secrets.yaml
```

---

## 🤖 常用AI服务商配置

### OpenAI 全家桶

```yaml
profiles:
  # GPT-4o - 最强综合模型
  - key: gpt4
    model: gpt-4o
    adapter: http
    http:
      endpoint: https://api.openai.com/v1/chat/completions
      method: POST
      headers:
        Authorization: "Bearer {{secret://OPENAI_API_KEY}}"
        Content-Type: application/json
      body_template: |
        {
          "model": "{{.Model}}",
          "messages": [{"role": "user", "content": "{{.Text}}"}],
          "temperature": 0.7
        }
      response_path: choices.0.message.content
      usage_input_path: usage.prompt_tokens
      usage_output_path: usage.completion_tokens
    cost_per_1k_input: 0.0025
    cost_per_1k_output: 0.010

  # GPT-4o-mini - 经济实惠主力
  - key: gpt4-mini
    model: gpt-4o-mini
    adapter: http
    http:
      endpoint: https://api.openai.com/v1/chat/completions
      method: POST
      headers:
        Authorization: "Bearer {{secret://OPENAI_API_KEY}}"
        Content-Type: application/json
      body_template: |
        {
          "model": "{{.Model}}",
          "messages": [{"role": "user", "content": "{{.Text}}"}],
          "temperature": 0.5
        }
      response_path: choices.0.message.content
      usage_input_path: usage.prompt_tokens
      usage_output_path: usage.completion_tokens
    cost_per_1k_input: 0.00015
    cost_per_1k_output: 0.0006

  # GPT-4-Turbo - 长文本处理
  - key: gpt4-turbo
    model: gpt-4-turbo-preview
    adapter: http
    http:
      endpoint: https://api.openai.com/v1/chat/completions
      method: POST
      headers:
        Authorization: "Bearer {{secret://OPENAI_API_KEY}}"
        Content-Type: application/json
      body_template: |
        {
          "model": "{{.Model}}",
          "messages": [{"role": "user", "content": "{{.Text}}"}],
          "max_tokens": 4096
        }
      response_path: choices.0.message.content
      usage_input_path: usage.prompt_tokens
      usage_output_path: usage.completion_tokens
    cost_per_1k_input: 0.01
    cost_per_1k_output: 0.03
```

### Anthropic Claude

```yaml
profiles:
  # Claude 3.5 Sonnet - 代码分析专家
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

  # Claude 3 Haiku - 快速响应
  - key: claude-haiku
    model: claude-3-haiku-20240307
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
          "max_tokens": 1024,
          "messages": [{"role": "user", "content": "{{.Text}}"}]
        }
      response_path: content.0.text
      usage_input_path: usage.input_tokens
      usage_output_path: usage.output_tokens
    cost_per_1k_input: 0.00025
    cost_per_1k_output: 0.00125
```

### 本地模型 (Ollama)

```yaml
profiles:
  # 本地 Llama 3.1
  - key: local-llama
    model: llama3.1:8b
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
    cost_per_1k_input: 0.0  # 本地模型免费
    cost_per_1k_output: 0.0

  # 本地 CodeLlama (编程专用)
  - key: local-codellama
    model: codellama:13b
    adapter: tool
    tool:
      bin: ollama
      args:
        - run
        - codellama:13b
      env:
        OLLAMA_HOST: "{{secret://OLLAMA_HOST}}"
      stdin: true
      output_format: raw
    cost_per_1k_input: 0.0
    cost_per_1k_output: 0.0
```

对应的 `secrets.yaml`:
```yaml
secrets:
  OPENAI_API_KEY: "sk-your-openai-key"
  ANTHROPIC_API_KEY: "sk-ant-your-anthropic-key"
  OLLAMA_HOST: "http://localhost:11434"
```

---

## 🧠 智能路由规则实战

编辑 `~/.boba/routes.yaml`:

### 开发者专用路由规则

```yaml
routes:
  # 代码分析任务 - 使用Claude (代码分析能力强)
  - id: code-analysis
    match:
      text_matches: "分析|优化|重构|review|debug"
      ctx_chars_gte: 500
    profile: claude-sonnet
    explain: "代码分析使用Claude，理解更深入"

  # 简单代码任务 - 使用GPT-4o-mini (经济实惠)
  - id: simple-code
    match:
      text_matches: "写个|实现|函数|方法"
      ctx_chars_lt: 1000
    profile: gpt4-mini
    explain: "简单代码任务使用经济模型"

  # 长文本处理 - 使用GPT-4-Turbo (支持长上下文)
  - id: long-context
    match:
      ctx_chars_gte: 10000
    profile: gpt4-turbo
    explain: "长文本使用GPT-4-Turbo，支持更多上下文"

  # 文档写作 - 使用Claude (写作能力强)
  - id: documentation
    match:
      text_matches: "文档|说明|readme|markdown"
    profile: claude-sonnet
    explain: "文档写作使用Claude，表达更清晰"

  # 快速问答 - 使用本地模型 (快速免费)
  - id: quick-chat
    match:
      ctx_chars_lt: 300
      intent: chat
    profile: local-llama
    explain: "简单问答使用本地模型"

  # 夜间时段 - 优先使用本地模型 (节省成本)
  - id: night-shift
    match:
      time_of_day: [night]
    profile: local-llama
    explain: "夜间使用本地模型，节省成本"

# 默认fallback
fallback: gpt4-mini
```

### 时间段定义
- **morning**: 6:00-12:00
- **afternoon**: 12:00-18:00  
- **evening**: 18:00-22:00
- **night**: 22:00-6:00

### 项目类型支持
- **go**: Go语言项目
- **python**: Python项目
- **javascript**/**typescript**: JS/TS项目
- **java**: Java项目
- **rust**: Rust项目

---

## 💰 预算控制配置

### 个人开发者预算

在 `~/.boba/profiles.yaml` 中设置：

```yaml
profiles:
  - key: gpt4-mini
    model: gpt-4o-mini
    # ... 其他配置 ...
    budget:
      daily: 5.0      # 每日预算 $5
      monthly: 100.0  # 每月预算 $100

  - key: claude-sonnet
    model: claude-3-5-sonnet-20241022
    # ... 其他配置 ...
    budget:
      daily: 10.0     # 每日预算 $10 (Claude更贵)
      monthly: 200.0  # 每月预算 $200
```

### 项目级预算控制

创建 `项目根目录/.boba-project.yaml`:

```yaml
project:
  name: "我的AI应用"
  type: go

# 覆盖全局默认设置
default_profile: gpt4-mini

# 项目专用预算
budget:
  daily: 20.0
  monthly: 300.0

# 项目专用路由规则
routes:
  - id: project-specific
    match:
      project_types: [go]
    profile: claude-sonnet
    explain: "Go项目使用Claude进行代码分析"

fallback: gpt4-mini
```

### 企业团队预算配置

```yaml
# 团队配置示例
team:
  name: "AI开发团队"
  members: ["alice", "bob", "charlie"]

# 团队预算
budget:
  daily: 100.0
  monthly: 2000.0
  per_user_monthly: 200.0  # 每人每月限额

# 告警设置
alerts:
  daily_threshold: 0.8     # 达到80%时告警
  monthly_threshold: 0.9   # 达到90%时告警
  notifications:
    - type: email
      recipients: ["team-leader@company.com"]
    - type: slack
      webhook: "{{secret://SLACK_WEBHOOK}}"
```

---

## 🏢 多环境配置管理

### 开发/测试/生产环境分离

**开发环境** (`~/.boba/dev-profiles.yaml`):
```yaml
default_profile: local-llama

profiles:
  - key: local-llama
    model: llama3.1:8b
    adapter: tool
    tool:
      bin: ollama
      args: ["run", "llama3.1:8b"]
      stdin: true
      output_format: raw
    cost_per_1k_input: 0.0
    cost_per_1k_output: 0.0
```

**生产环境** (`~/.boba/prod-profiles.yaml`):
```yaml
default_profile: gpt4-mini

profiles:
  - key: gpt4-mini
    model: gpt-4o-mini
    # ... 完整生产配置 ...
    budget:
      daily: 50.0
      monthly: 1000.0
```

**环境切换脚本** (`switch-env.sh`):
```bash
#!/bin/bash
BOBA_HOME="$HOME/.boba"

case $1 in
  "dev")
    cp "$BOBA_HOME/dev-profiles.yaml" "$BOBA_HOME/profiles.yaml"
    echo "切换到开发环境"
    ;;
  "prod")
    cp "$BOBA_HOME/prod-profiles.yaml" "$BOBA_HOME/profiles.yaml"
    echo "切换到生产环境"
    ;;
  *)
    echo "用法: $0 {dev|prod}"
    exit 1
    ;;
esac

# 验证配置
boba doctor
```

---

## 👥 团队协作配置

### 集中式配置管理

**团队共享配置** (`team-configs.yaml`):
```yaml
# 团队标准配置
team_standard:
  default_profiles:
    junior: gpt4-mini        # 初级开发者使用经济模型
    senior: claude-sonnet    # 高级开发者使用强力模型
    lead: gpt4               # 技术负责人使用最强模型

  cost_limits:
    junior_dev:
      daily: 10.0
      monthly: 200.0
    senior_dev:
      daily: 20.0
      monthly: 400.0
    tech_lead:
      daily: 50.0
      monthly: 1000.0

  routing_rules:
    code_review: claude-sonnet
    documentation: gpt4-mini
    debugging: gpt4
    testing: local-llama
```

**个人配置定制** (`~/.boba/profiles.yaml`):
```yaml
# 继承团队配置
extends: "team-configs.yaml"

# 个人定制
role: "senior_dev"  # 个人角色

# 个人API密钥
secrets_file: "~/.boba/personal-secrets.yaml"

# 项目覆盖
project_overrides:
  ai-experiment:
    default_profile: gpt4
    budget:
      daily: 100.0
```

---

## 🔧 高级配置技巧

### 动态模板配置

```yaml
profiles:
  - key: dynamic-gpt
    model: "{{.Args.model | default \"gpt-4o-mini\"}}"
    adapter: http
    http:
      endpoint: https://api.openai.com/v1/chat/completions
      method: POST
      headers:
        Authorization: "Bearer {{secret://OPENAI_API_KEY}}"
        Content-Type: application/json
      body_template: |
        {
          "model": "{{.Model}}",
          "messages": [
            {
              "role": "system", 
              "content": "{{.Args.system_prompt | default \"你是一个有帮助的AI助手\"}}"
            },
            {
              "role": "user", 
              "content": "{{.Text}}"
            }
          ],
          "temperature": {{.Args.temperature | default 0.7}},
          "max_tokens": {{.Args.max_tokens | default 2048}}
        }
      response_path: choices.0.message.content
      usage_input_path: usage.prompt_tokens
      usage_output_path: usage.completion_tokens
```

### 条件适配器

```yaml
profiles:
  - key: smart-router
    model: auto
    adapter: conditional
    conditional:
      # 根据网络条件选择
      network_check:
        fast: claude-sonnet
        slow: local-llama
        offline: local-llama
      
      # 根据时间选择
      time_based:
        work_hours: gpt4
        after_hours: gpt4-mini
        weekend: local-llama
      
      # 根据负载选择
      load_balancing:
        high_load: local-llama
        medium_load: gpt4-mini
        low_load: claude-sonnet
```

---

## 🛠️ 配置验证和调试

### 完整配置检查

```bash
# 检查配置完整性
boba doctor

# 测试特定profile
boba test --profile gpt4-mini "测试消息"

# 测试路由规则
boba route test "分析这个Go代码的性能"

# 查看当前配置
boba config show
```

### 配置文件模板生成

```bash
# 生成基础配置模板
boba config init --template basic

# 生成团队配置模板
boba config init --template team

# 生成企业配置模板
boba config init --template enterprise
```

---

## 🚨 故障排除

### 常见配置错误

1. **API密钥错误**
   ```bash
   # 检查密钥格式
   boba doctor
   # 测试连接
   curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
   ```

2. **路由规则不生效**
   ```bash
   # 调试路由
   boba route debug --input "测试内容"
   # 查看路由日志
   boba logs --component router
   ```

3. **预算限制问题**
   ```bash
   # 检查预算状态
   boba budget status
   # 重置预算
   boba budget reset --profile gpt4-mini
   ```

### 配置迁移

```bash
# 从旧版本迁移
boba migrate --from-version 0.1.x

# 备份当前配置
boba config backup --output my-config-backup.yaml

# 恢复配置
boba config restore --input my-config-backup.yaml
```

---

## 📈 配置优化建议

### 成本优化

1. **智能路由**: 合理配置路由规则，避免过度使用昂贵模型
2. **本地优先**: 夜间和简单任务优先使用本地模型
3. **预算控制**: 设置合理的日/月预算限制
4. **使用分析**: 定期检查使用模式，优化配置

### 性能优化

1. **并发控制**: 避免同时发送过多请求
2. **缓存策略**: 相似问题使用缓存结果
3. **网络优化**: 配置合适的超时和重试策略
4. **模型选择**: 根据任务复杂度选择合适模型

---

## 🎯 下一步

- **[路由策略深度指南](/docs/routing/)** - 掌握高级路由技巧
- **[适配器开发指南](/docs/adapters/)** - 自定义AI服务集成
- **[预算管理最佳实践](/docs/budgets/)** - 企业级成本控制
- **[性能优化技巧](/docs/performance/)** - 大规模使用优化

> **💡 小贴士**: 好的配置是持续优化的过程。定期查看使用统计，根据实际需求调整配置。
