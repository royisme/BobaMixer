# 适配器

适配器是 BobaMixer 与 AI 提供商之间的桥梁。它们处理通信、追踪使用情况并规范不同提供商类型的响应。

## 概览

BobaMixer 支持三种适配器类型:

1. **HTTP 适配器** - 用于 REST API 提供商 (Anthropic、OpenAI 等)
2. **Tool 适配器** - 用于 CLI 工具和可执行文件
3. **MCP 适配器** - 用于模型上下文协议集成

## HTTP 适配器

HTTP 适配器与 REST API 提供商通信。

### 支持的提供商

- **Anthropic** - Claude 模型
- **OpenAI** - GPT 模型
- **OpenRouter** - 多提供商网关
- **自定义提供商** - 任何 REST API

### 配置示例

**Anthropic:**
```yaml
claude-sonnet:
  adapter: http
  provider: anthropic
  endpoint: https://api.anthropic.com/v1/messages
  model: claude-3-5-sonnet-20241022
  headers:
    anthropic-version: "2023-06-01"
    x-api-key: "secret://anthropic_key"
```

**OpenAI:**
```yaml
gpt4-turbo:
  adapter: http
  provider: openai
  endpoint: https://api.openai.com/v1/chat/completions
  model: gpt-4-turbo-preview
  headers:
    Authorization: "Bearer secret://openai_key"
```

## Tool 适配器

Tool 适配器执行 CLI 工具和脚本,追踪它们的使用情况。

### 用例

- **claude-code** - Claude 的 CLI 接口
- **自定义脚本** - 你自己的 AI 包装器
- **本地模型** - 通过 ollama 的 LLaMA、Mistral
- **专用工具** - 领域特定的 AI 工具

### 配置示例

```yaml
claude-code:
  adapter: tool
  command: /usr/local/bin/claude
  args: ["--session", "work", "--format", "jsonl"]
  env:
    - ANTHROPIC_API_KEY=secret://anthropic_key
```

## MCP 适配器

MCP (模型上下文协议) 适配器与 MCP 兼容的工具集成。

### 配置示例

```yaml
mcp-filesystem:
  adapter: mcp
  command: npx
  args: ["@modelcontextprotocol/server-filesystem", "/path/to/docs"]
  transport: stdio
```

## 使用追踪

BobaMixer 使用三个估算级别追踪使用情况:

1. **精确** - 从 API 响应 `usage` 字段
2. **映射** - 从定价配置
3. **启发式** - 基于字符的估算

### 检查估算级别

```bash
# 在报告中查看估算
boba report --format json | jq '.[] | select(.estimate != "exact")'

# 按估算级别统计
boba stats --by-estimate
```

## 最佳实践

### 1. 选择正确的适配器

- **HTTP**: 用于生产 API 提供商
- **Tool**: 用于本地模型、自定义脚本
- **MCP**: 用于工具增强工作流

### 2. 保护 API 密钥

```yaml
# 始终使用 secret:// 引用
headers:
  x-api-key: "secret://anthropic_key"  # ✅ 好

# 永不硬编码
headers:
  x-api-key: "sk-ant-xxxxx"  # ❌ 坏
```

### 3. 测试适配器

```bash
# 验证适配器配置
boba doctor

# 测试配置文件
boba use 配置文件名称

# 检查连接性
boba route test "测试消息"
```

## 下一步

- **[路由](/zh/features/routing)** - 设置智能路由
- **[配置指南](/zh/guide/configuration)** - 详细配置选项
- **[CLI 参考](/zh/reference/cli)** - 命令文档
