# Adapters

Adapters are the bridge between BobaMixer and AI providers. They handle communication, track usage, and normalize responses across different provider types.

## Overview

BobaMixer supports three adapter types:

1. **HTTP Adapter** - For REST API providers (Anthropic, OpenAI, etc.)
2. **Tool Adapter** - For CLI tools and executables
3. **MCP Adapter** - For Model Context Protocol integrations

Each adapter implements a common interface, allowing BobaMixer to work seamlessly with any provider.

## HTTP Adapter

The HTTP adapter communicates with REST API providers.

### Supported Providers

- **Anthropic** - Claude models
- **OpenAI** - GPT models
- **OpenRouter** - Multi-provider gateway
- **Custom providers** - Any REST API

### Configuration

```yaml
profile-name:
  adapter: http
  provider: anthropic|openai|openrouter|custom
  endpoint: https://api.example.com/v1/endpoint
  model: model-name
  max_tokens: 4096
  temperature: 0.7
  headers:
    Header-Name: value
    Authorization: "Bearer secret://api_key"
```

### Anthropic Example

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
  cost_per_1k:
    input: 0.015
    output: 0.075
```

### OpenAI Example

```yaml
gpt4-turbo:
  adapter: http
  provider: openai
  endpoint: https://api.openai.com/v1/chat/completions
  model: gpt-4-turbo-preview
  max_tokens: 4096
  temperature: 0.7
  headers:
    Authorization: "Bearer secret://openai_key"
    Content-Type: "application/json"
  cost_per_1k:
    input: 0.01
    output: 0.03
```

### OpenRouter Example

```yaml
openrouter-claude:
  adapter: http
  provider: openrouter
  endpoint: https://openrouter.ai/api/v1/chat/completions
  model: anthropic/claude-3-5-sonnet
  headers:
    Authorization: "Bearer secret://openrouter_key"
    HTTP-Referer: "https://github.com/yourusername"
  cost_per_1k:
    input: 0.015
    output: 0.075
```

### Custom Provider Example

For any custom REST API:

```yaml
custom-llm:
  adapter: http
  provider: custom
  endpoint: https://custom-api.example.com/v1/generate
  model: custom-model-v1
  headers:
    X-API-Key: "secret://custom_key"
    X-Custom-Header: "value"
  params:
    custom_param: value
```

### HTTP Adapter Features

- Automatic retry with exponential backoff
- Request/response logging
- Timeout handling
- Usage extraction from API responses
- Streaming support (coming soon)

## Tool Adapter

The Tool adapter executes CLI tools and scripts, tracking their usage.

### Use Cases

- **claude-code** - CLI interface for Claude
- **Custom scripts** - Your own AI wrappers
- **Local models** - LLaMA, Mistral via ollama
- **Specialized tools** - Domain-specific AI tools

### Configuration

```yaml
profile-name:
  adapter: tool
  command: /path/to/executable
  args: ["--arg1", "value1", "--arg2"]
  env:
    - VAR_NAME=value
    - API_KEY=secret://key_name
  working_dir: /path/to/workdir
  timeout_seconds: 300
```

### Claude Code Example

```yaml
claude-code:
  adapter: tool
  command: /usr/local/bin/claude
  args: ["--session", "work", "--format", "jsonl"]
  env:
    - ANTHROPIC_API_KEY=secret://anthropic_key
    - CLAUDE_OUTPUT_FORMAT=jsonl
  cost_per_1k:
    input: 0.015
    output: 0.075
```

### Custom Script Example

```yaml
my-ai-script:
  adapter: tool
  command: /home/user/scripts/ai-wrapper.sh
  args: ["--model", "best", "--verbose"]
  env:
    - API_KEY=secret://my_api_key
    - LOG_LEVEL=debug
  working_dir: /home/user/workspace
  timeout_seconds: 600
```

### Ollama Example

```yaml
ollama-llama:
  adapter: tool
  command: ollama
  args: ["run", "llama2", "--format", "json"]
  cost_per_1k:
    input: 0.0
    output: 0.0  # Free local model
```

### Usage Tracking for Tools

Tools can report usage in two ways:

#### 1. JSONL Output

Output usage as JSON Lines to stdout:

```json
{"event":"usage","input_tokens":100,"output_tokens":50}
```

Example script:

```bash
#!/bin/bash
# Your AI logic here
result=$(call_ai_api "$@")

# Output usage
echo '{"event":"usage","input_tokens":100,"output_tokens":50}'

# Output result
echo "$result"
```

#### 2. Exit Code + Stderr

Return token counts via stderr and exit code:

```bash
#!/bin/bash
# Your AI logic here

# Output to stderr (parsed by BobaMixer)
echo "INPUT_TOKENS=100" >&2
echo "OUTPUT_TOKENS=50" >&2

exit 0
```

### Tool Adapter Features

- Stdin/stdout/stderr handling
- Environment variable injection
- Working directory control
- Timeout management
- Automatic usage parsing
- Error handling

## MCP Adapter

The MCP (Model Context Protocol) adapter integrates with MCP-compatible tools.

### What is MCP?

Model Context Protocol is a standard for connecting AI models with external tools and data sources. It enables:

- File system access
- Database queries
- API integrations
- Custom tool execution

Learn more at [Model Context Protocol](https://modelcontextprotocol.io/).

### Configuration

```yaml
profile-name:
  adapter: mcp
  command: command-name
  args: ["arg1", "arg2"]
  transport: stdio|sse
  env:
    - VAR=value
```

### Filesystem MCP Example

```yaml
mcp-filesystem:
  adapter: mcp
  command: npx
  args: ["@modelcontextprotocol/server-filesystem", "/path/to/docs"]
  transport: stdio
  env:
    - NODE_ENV=production
```

### Database MCP Example

```yaml
mcp-postgres:
  adapter: mcp
  command: npx
  args: ["@modelcontextprotocol/server-postgres"]
  transport: stdio
  env:
    - DATABASE_URL=secret://postgres_url
```

### Custom MCP Server Example

```yaml
my-mcp-server:
  adapter: mcp
  command: /path/to/my-mcp-server
  args: ["--config", "/path/to/config.json"]
  transport: stdio
```

### MCP Adapter Features

- Stdio and SSE transport
- Tool discovery
- Resource management
- Prompt templates
- Automatic context building

## Creating Custom Adapters

You can create custom adapters by implementing the Adapter interface.

### Adapter Interface

```go
type Adapter interface {
    Name() string
    Execute(ctx context.Context, req Request) (Result, error)
}
```

### HTTP-Based Custom Adapter

```go
package myadapter

import (
    "context"
    "encoding/json"
    "github.com/royisme/bobamixer/internal/adapters"
    httpadapter "github.com/royisme/bobamixer/internal/adapters/http"
)

type MyAdapter struct {
    *httpadapter.Client
}

func New(name, endpoint string, headers map[string]string) *MyAdapter {
    return &MyAdapter{
        Client: httpadapter.New(name, endpoint, headers),
    }
}

func (a *MyAdapter) Execute(ctx context.Context, req adapters.Request) (adapters.Result, error) {
    // Transform request for your API
    providerReq := map[string]interface{}{
        "model": req.Model,
        "prompt": string(req.Payload),
        "max_tokens": 2048,
    }

    payload, _ := json.Marshal(providerReq)
    req.Payload = payload

    // Execute via base HTTP client
    return a.Client.Execute(ctx, req)
}
```

### Tool-Based Custom Adapter

```go
package mytooladapter

import (
    "context"
    "os/exec"
    "github.com/royisme/bobamixer/internal/adapters"
)

type MyToolAdapter struct {
    command string
    args    []string
}

func New(command string, args []string) *MyToolAdapter {
    return &MyToolAdapter{
        command: command,
        args:    args,
    }
}

func (a *MyToolAdapter) Name() string {
    return "my-tool"
}

func (a *MyToolAdapter) Execute(ctx context.Context, req adapters.Request) (adapters.Result, error) {
    cmd := exec.CommandContext(ctx, a.command, a.args...)
    cmd.Stdin = bytes.NewReader(req.Payload)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return adapters.Result{Success: false}, err
    }

    return adapters.Result{
        Success: true,
        Output:  output,
        Usage: adapters.Usage{
            // Parse usage from output
        },
    }, nil
}
```

### Testing Custom Adapters

```go
func TestMyAdapter(t *testing.T) {
    adapter := New("https://api.test.com", map[string]string{
        "Authorization": "Bearer test-key",
    })

    result, err := adapter.Execute(context.Background(), adapters.Request{
        Payload: []byte("test prompt"),
        Model:   "test-model",
    })

    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.NotEmpty(t, result.Output)
}
```

## Usage Estimation

BobaMixer tracks usage with three estimation levels:

### 1. Exact Estimation

From API response `usage` field:

```json
{
  "usage": {
    "input_tokens": 100,
    "output_tokens": 50
  }
}
```

**Accuracy**: 100%

### 2. Mapped Estimation

From pricing configuration:

```yaml
cost_per_1k:
  input: 0.015
  output: 0.075
```

**Accuracy**: 95-99% (depends on pricing accuracy)

### 3. Heuristic Estimation

Character-based estimation when no other data available:

- Input tokens ≈ characters / 4
- Output tokens ≈ characters / 4

**Accuracy**: 70-90%

### Checking Estimation Level

```bash
# View estimates in reports
boba report --format json | jq '.[] | select(.estimate != "exact")'

# Statistics by estimation level
boba stats --by-estimate
```

## Best Practices

### 1. Choose the Right Adapter

- **HTTP**: For production API providers
- **Tool**: For local models, custom scripts
- **MCP**: For tool-augmented workflows

### 2. Secure API Keys

```yaml
# Always use secret:// references
headers:
  x-api-key: "secret://anthropic_key"  # ✅ Good

# Never hardcode
headers:
  x-api-key: "sk-ant-xxxxx"  # ❌ Bad
```

### 3. Set Appropriate Timeouts

```yaml
# For quick tasks
timeout_seconds: 30

# For complex tasks
timeout_seconds: 300

# For very long tasks
timeout_seconds: 600
```

### 4. Monitor Usage Accuracy

```bash
# Check estimation accuracy
boba stats --by-estimate --7d

# If too many heuristics, add pricing config
```

### 5. Test Adapters

```bash
# Verify adapter configuration
boba doctor

# Test profile
boba use profile-name

# Check connectivity
boba route test "test message"
```

### 6. Handle Errors Gracefully

```yaml
# Use fallback profiles
rules:
  - id: primary-with-fallback
    if: "ctx_chars > 0"
    use: primary-profile
    fallback: backup-profile
```

## Troubleshooting

### HTTP Adapter Issues

```bash
# Check endpoint connectivity
curl -v https://api.anthropic.com/v1/messages

# Verify headers
boba doctor --verbose

# Check API key
grep "anthropic_key" ~/.boba/secrets.yaml
```

### Tool Adapter Issues

```bash
# Verify command exists
which claude-code

# Test command manually
/path/to/command --help

# Check permissions
ls -l /path/to/command
```

### MCP Adapter Issues

```bash
# Test MCP server
npx @modelcontextprotocol/server-filesystem /path --test

# Check logs
tail -f ~/.boba/logs/boba.log

# Verify transport
# stdio: uses stdin/stdout
# sse: uses Server-Sent Events
```

### Usage Tracking Issues

```bash
# Check estimation level
boba stats --by-estimate

# View detailed usage
boba report --format json | jq '.[] | {profile, estimate, tokens}'

# Add pricing config if needed
boba edit pricing
```

## Examples

### Multi-Provider Setup

```yaml
# Production: Anthropic
prod-claude:
  adapter: http
  provider: anthropic
  endpoint: https://api.anthropic.com/v1/messages
  model: claude-3-5-sonnet-20241022
  headers:
    x-api-key: "secret://anthropic_key"

# Development: OpenAI
dev-gpt:
  adapter: http
  provider: openai
  endpoint: https://api.openai.com/v1/chat/completions
  model: gpt-4o-mini
  headers:
    Authorization: "Bearer secret://openai_key"

# Local: Ollama
local-llama:
  adapter: tool
  command: ollama
  args: ["run", "llama2"]
  cost_per_1k:
    input: 0.0
    output: 0.0
```

### Hybrid Workflow

```yaml
# API for production
api-profile:
  adapter: http
  provider: anthropic
  # ...

# CLI tool for development
tool-profile:
  adapter: tool
  command: claude-code
  # ...

# MCP for data access
mcp-profile:
  adapter: mcp
  command: npx
  # ...
```

## Next Steps

- **[Routing](/features/routing)** - Set up intelligent routing
- **[Configuration Guide](/guide/configuration)** - Detailed config options
- **[CLI Reference](/reference/cli)** - Command documentation
- **[Operations](/advanced/operations)** - Production best practices

## Resources

- [HTTP Adapter Source](https://github.com/royisme/BobaMixer/tree/main/internal/adapters/http)
- [Tool Adapter Source](https://github.com/royisme/BobaMixer/tree/main/internal/adapters/tool)
- [MCP Adapter Source](https://github.com/royisme/BobaMixer/tree/main/internal/adapters/mcp)
- [Model Context Protocol](https://modelcontextprotocol.io/)
