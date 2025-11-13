# Configuration Guide

Complete configuration handbook from beginner to expert - master BobaMixer configuration through real-world scenarios.

## Configuration Overview

### Global Configuration (`~/.boba/`)

```
~/.boba/
├── profiles.yaml     # AI provider configs - ★★★★★ (most important)
├── routes.yaml       # Smart routing rules - ★★★★☆ (core feature)
├── secrets.yaml      # API keys storage - ★★★★★ (security critical)
├── pricing.yaml      # Pricing info - ★★★☆☆ (optional)
└── usage.db          # Usage database (auto-generated)
```

### Project Configuration (project root)

```
my-project/
├── .boba-project.yaml  # Project-specific config
└── .gitignore          # Remember to ignore sensitive configs
```

## Basic Configuration Template

### Minimal Working Configuration

Edit `~/.boba/profiles.yaml`:

```yaml
# Set default AI model
default_profile: gpt4-mini

# AI provider configurations
profiles:
  # OpenAI GPT-4o-mini - economical workhorse
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

Edit `~/.boba/secrets.yaml`:

```yaml
secrets:
  OPENAI_API_KEY: "sk-your-openai-api-key-here"
```

**Set secure permissions**:

```bash
chmod 600 ~/.boba/secrets.yaml
```

## Common AI Provider Configurations

### OpenAI Suite

```yaml
profiles:
  # GPT-4o - Most capable model
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

  # GPT-4o-mini - Economical workhorse
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
```

### Anthropic Claude

```yaml
profiles:
  # Claude 3.5 Sonnet - Code analysis expert
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

  # Claude 3 Haiku - Fast responses
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

### Local Models (Ollama)

```yaml
profiles:
  # Local Llama 3.1
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
    cost_per_1k_input: 0.0  # Local models are free
    cost_per_1k_output: 0.0
```

Corresponding `secrets.yaml`:

```yaml
secrets:
  OPENAI_API_KEY: "sk-your-openai-key"
  ANTHROPIC_API_KEY: "sk-ant-your-anthropic-key"
  OLLAMA_HOST: "http://localhost:11434"
```

## Smart Routing Rules

Edit `~/.boba/routes.yaml`:

### Developer-Focused Routing Rules

```yaml
routes:
  # Code analysis tasks - Use Claude (strong code analysis)
  - id: code-analysis
    match:
      text_matches: "analyze|optimize|refactor|review|debug"
      ctx_chars_gte: 500
    profile: claude-sonnet
    explain: "Code analysis uses Claude for deeper understanding"

  # Simple code tasks - Use GPT-4o-mini (economical)
  - id: simple-code
    match:
      text_matches: "write|implement|function|method"
      ctx_chars_lt: 1000
    profile: gpt4-mini
    explain: "Simple code tasks use economical model"

  # Long context - Use GPT-4-Turbo (supports long context)
  - id: long-context
    match:
      ctx_chars_gte: 10000
    profile: gpt4-turbo
    explain: "Long text uses GPT-4-Turbo for more context"

  # Documentation - Use Claude (strong writing)
  - id: documentation
    match:
      text_matches: "documentation|docs|readme|markdown"
    profile: claude-sonnet
    explain: "Documentation uses Claude for clearer expression"

# Default fallback
fallback: gpt4-mini
```

### Time Period Definitions

- **morning**: 6:00-12:00
- **afternoon**: 12:00-18:00
- **evening**: 18:00-22:00
- **night**: 22:00-6:00

### Project Type Support

- **go**: Go projects
- **python**: Python projects
- **javascript**/**typescript**: JS/TS projects
- **java**: Java projects
- **rust**: Rust projects

## Budget Control

### Personal Developer Budget

Set in `~/.boba/profiles.yaml`:

```yaml
profiles:
  - key: gpt4-mini
    model: gpt-4o-mini
    # ... other config ...
    budget:
      daily: 5.0      # Daily budget $5
      monthly: 100.0  # Monthly budget $100

  - key: claude-sonnet
    model: claude-3-5-sonnet-20241022
    # ... other config ...
    budget:
      daily: 10.0     # Daily budget $10 (Claude is more expensive)
      monthly: 200.0  # Monthly budget $200
```

### Project-Level Budget Control

Create `project-root/.boba-project.yaml`:

```yaml
project:
  name: "My AI App"
  type: go

# Override global defaults
default_profile: gpt4-mini

# Project-specific budget
budget:
  daily: 20.0
  monthly: 300.0

# Project-specific routing rules
routes:
  - id: project-specific
    match:
      project_types: [go]
    profile: claude-sonnet
    explain: "Go projects use Claude for code analysis"

fallback: gpt4-mini
```

## Configuration Validation and Debugging

### Complete Configuration Check

```bash
# Check configuration integrity
boba doctor

# Test specific profile
boba test --profile gpt4-mini "test message"

# Test routing rules
boba route test "analyze this Go code performance"

# View current configuration
boba config show
```

## Troubleshooting

### Common Configuration Errors

1. **API Key Errors**
   ```bash
   # Check key format
   boba doctor
   # Test connection
   curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
   ```

2. **Routing Rules Not Working**
   ```bash
   # Debug routing
   boba route debug --input "test content"
   # View routing logs
   boba logs --component router
   ```

3. **Budget Limit Issues**
   ```bash
   # Check budget status
   boba budget status
   # Reset budget
   boba budget reset --profile gpt4-mini
   ```

## Configuration Optimization Tips

### Cost Optimization

1. **Smart Routing**: Configure routing rules wisely to avoid overusing expensive models
2. **Local First**: Prioritize local models for nighttime and simple tasks
3. **Budget Control**: Set reasonable daily/monthly budget limits
4. **Usage Analysis**: Regularly review usage patterns and optimize configuration

### Performance Optimization

1. **Concurrency Control**: Avoid sending too many requests simultaneously
2. **Caching Strategy**: Use cached results for similar questions
3. **Network Optimization**: Configure appropriate timeout and retry strategies
4. **Model Selection**: Choose appropriate models based on task complexity

## Next Steps

- [Routing Cookbook](/ROUTING_COOKBOOK) - Master advanced routing techniques
- [Adapters](/ADAPTERS) - Custom AI service integration
- [FAQ](/FAQ) - Frequently asked questions

> **💡 Tip**: Good configuration is a process of continuous optimization. Regularly review usage statistics and adjust configuration based on actual needs.
