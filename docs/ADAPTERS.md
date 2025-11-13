# Adapter Development Guide

This guide explains how to build custom adapters for BobaMixer to integrate with new AI/LLM providers or tools.

## Adapter Interface

All adapters must implement the `Adapter` interface defined in `internal/adapters/adapter.go`:

```go
type Adapter interface {
    Name() string
    Execute(ctx context.Context, req Request) (Result, error)
}
```

## Available Adapter Types

### 1. HTTP Adapter
For REST API-based providers (Anthropic, OpenAI, OpenRouter, etc.)

**Example Configuration:**
```yaml
my-provider:
  adapter: http
  provider: custom
  endpoint: https://api.example.com/v1/generate
  model: my-model-v1
  headers:
    Authorization: "Bearer secret://my_api_key"
    Content-Type: "application/json"
```

### 2. Tool Adapter
For CLI tools and executables

**Example Configuration:**
```yaml
my-tool:
  adapter: tool
  command: /path/to/my-tool
  args: ["--mode", "generate"]
  env:
    - MY_API_KEY=secret://tool_key
```

### 3. MCP Adapter
For Model Context Protocol integrations

**Example Configuration:**
```yaml
my-mcp:
  adapter: mcp
  command: npx
  args: ["@modelcontextprotocol/server-example"]
  transport: stdio
```

## Creating a Custom HTTP Adapter

### Step 1: Implement the Provider

```go
package myadapter

import (
    "context"
    "encoding/json"
    "net/http"
    
    "github.com/royisme/bobamixer/internal/adapters"
    "github.com/royisme/bobamixer/internal/adapters/http"
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
    // Build provider-specific request
    providerReq := map[string]interface{}{
        "model": req.Model,
        "prompt": string(req.Payload),
        "max_tokens": 2048,
    }
    
    payload, _ := json.Marshal(providerReq)
    req.Payload = payload
    
    // Use base HTTP client
    return a.Client.Execute(ctx, req)
}
```

### Step 2: Register in Factory

Add to `internal/adapters/factory.go` (if exists) or handle in configuration loader.

### Step 3: Test

```go
func TestMyAdapter(t *testing.T) {
    adapter := New("test", "https://api.test.com", map[string]string{
        "Authorization": "Bearer test-key",
    })
    
    result, err := adapter.Execute(context.Background(), adapters.Request{
        Payload: []byte("test prompt"),
        Model: "test-model",
    })
    
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

## Usage Tracking

BobaMixer tracks usage with three estimation levels:

1. **Exact** - From API response `usage` field
2. **Mapped** - From pricing configuration
3. **Heuristic** - Character-based estimation

### Returning Usage from Adapter

```go
return adapters.Result{
    Success: true,
    Output: responseData,
    Usage: adapters.Usage{
        InputTokens: parsed.Usage.InputTokens,
        OutputTokens: parsed.Usage.OutputTokens,
        Estimate: adapters.EstimateExact,
        LatencyMS: elapsedMs,
    },
}
```

### Supporting JSONL Usage Events

For CLI tools, output usage as JSON Lines:

```json
{"event":"usage","input_tokens":100,"output_tokens":50}
```

The Tool Adapter will automatically parse this format.

## Best Practices

1. **Error Handling**: Return errors with context
   ```go
   return adapters.Result{}, fmt.Errorf("api call failed: %w", err)
   ```

2. **Timeout**: Respect context cancellation
   ```go
   select {
   case <-ctx.Done():
       return adapters.Result{}, ctx.Err()
   case result := <-resultChan:
       return result, nil
   }
   ```

3. **Retries**: Implement exponential backoff for transient failures

4. **Secrets**: Use `secret://` prefix for sensitive values

5. **Testing**: Include both unit and integration tests

## Configuration Schema

### Profile Configuration

```yaml
profile-name:
  adapter: http|tool|mcp
  provider: provider-name  # optional
  endpoint: url            # for HTTP
  command: path            # for tool/mcp
  args: [...]              # optional
  model: model-name
  max_tokens: 4096
  temperature: 0.7
  headers:                 # for HTTP
    key: value
  env:                     # for tool/mcp
    - KEY=value
  params:                  # adapter-specific
    key: value
```

## Example: Custom Streaming Adapter

```go
func (a *StreamingAdapter) ExecuteStream(
    ctx context.Context,
    req adapters.Request,
    callback func(chunk []byte) error,
) error {
    // Create streaming request
    resp, err := a.client.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    reader := bufio.NewReader(resp.Body)
    for {
        line, err := reader.ReadBytes('\n')
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        if err := callback(line); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Debugging

Enable debug logging:
```bash
export BOBA_LOG_LEVEL=debug
boba doctor
```

Check adapter health:
```bash
boba doctor
```

Test connectivity:
```bash
# Will show detailed adapter diagnostics
boba doctor
```

## Contributing

When contributing a new adapter:

1. Add implementation in `internal/adapters/<adapter-name>/`
2. Add tests with >80% coverage
3. Update documentation
4. Add example configuration to `configs/examples/`
5. Submit PR with description of use case

## Resources

- [HTTP Adapter Source](../internal/adapters/http/)
- [Tool Adapter Source](../internal/adapters/tool/)
- [MCP Adapter Source](../internal/adapters/mcp/)
- [Example Configurations](../configs/examples/)
