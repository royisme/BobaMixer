---
title: "Configuration"
linkTitle: "Configuration"
weight: 3
description: >
  Complete reference for BobaMixer configuration files.
---

BobaMixer uses YAML configuration files stored in `~/.boba/`. This section provides a complete reference for all configuration options.

## Configuration Files

BobaMixer uses four main configuration files:

1. **profiles.yaml** - AI provider configurations
2. **routes.yaml** - Routing rules and strategies
3. **secrets.yaml** - API keys and sensitive data
4. **pricing.yaml** - Cost information for models

Additionally, project-specific settings can be stored in `.boba-project.yaml` in your project directory.

## File Locations

### Global Configuration

```
~/.boba/
├── profiles.yaml
├── routes.yaml
├── secrets.yaml
├── pricing.yaml
└── usage.db
```

### Project Configuration

```
/path/to/your/project/
└── .boba-project.yaml
```

## profiles.yaml

Define AI provider configurations with model settings, adapter type, and costs.

### Structure

```yaml
default_profile: gpt4-mini  # Optional: default profile key

profiles:
  - key: gpt4-mini           # Unique identifier
    model: gpt-4o-mini       # Model name
    adapter: http            # Adapter type: http, tool, or mcp

    # Adapter-specific configuration
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

    # Cost configuration
    cost_per_1k_input: 0.00015
    cost_per_1k_output: 0.0006

    # Budget limits (optional)
    budget:
      daily: 5.0    # Daily limit in USD
      monthly: 100.0  # Monthly limit in USD
```

### Adapter Types

#### HTTP Adapter

For REST API providers:

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

#### Tool Adapter

For command-line tools:

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

#### MCP Adapter

For Model Context Protocol servers:

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

Define rules for automatic profile selection based on context.

### Structure

```yaml
routes:
  - id: large-context          # Unique route identifier
    match:
      ctx_chars_gte: 50000     # Context size >= 50k chars
    profile: claude-sonnet     # Profile to use
    explain: "Large context requires Claude"

  - id: quick-tasks
    match:
      intent: format           # Intent matching
      ctx_chars_lt: 1000       # Context size < 1k chars
    profile: gpt4-mini
    explain: "Quick formatting task"

  - id: code-review
    match:
      text_matches: "review|PR|pull request"  # Regex pattern
      project_types: [go, typescript]         # Project type
    profile: claude-sonnet
    explain: "Code review task"

  - id: night-hours
    match:
      time_of_day: [night]     # Time ranges: morning, afternoon, evening, night
    profile: local-llama
    explain: "Use local model during night hours"

  - id: feature-branch
    match:
      branch_matches: "^feature/.*"  # Git branch pattern
    profile: gpt4-mini
    explain: "Development work on feature branch"

# Fallback profile if no routes match
fallback: gpt4-mini
```

### Match Conditions

All conditions in a `match` block must be satisfied (AND logic):

- **ctx_chars_gte**: Context size greater than or equal to value
- **ctx_chars_lt**: Context size less than value
- **intent**: Matches intent (format, analysis, chat, code, etc.)
- **text_matches**: Regex pattern against input text
- **project_types**: List of project types (go, typescript, python, rust, etc.)
- **time_of_day**: Time ranges (morning: 6-12, afternoon: 12-18, evening: 18-22, night: 22-6)
- **branch_matches**: Regex pattern against git branch name

## secrets.yaml

Store API keys and sensitive configuration.

```yaml
secrets:
  OPENAI_API_KEY: "sk-..."
  ANTHROPIC_API_KEY: "sk-ant-..."
  OLLAMA_HOST: "http://localhost:11434"
```

**Security**:
- File must have `0600` permissions (owner read/write only)
- Never commit this file to version control
- Add `secrets.yaml` to `.gitignore`

Use secrets in configurations with `{{secret://KEY_NAME}}` syntax.

## pricing.yaml

Define or override model pricing.

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

Pricing defined in `profiles.yaml` takes precedence over `pricing.yaml`.

## .boba-project.yaml

Project-specific configuration placed in project root:

```yaml
project:
  name: my-awesome-app
  type: go

# Override global settings for this project
default_profile: claude-sonnet

# Project-specific budget
budget:
  daily: 10.0
  monthly: 200.0
```

## Environment Variables

Override configuration with environment variables:

```bash
export BOBA_HOME=/custom/path      # Default: ~/.boba
export BOBA_PROFILE=gpt4-mini      # Override default profile
export BOBA_DEBUG=true             # Enable debug logging
```

## Configuration Validation

Validate your configuration:

```bash
boba doctor
```

This checks:
- File permissions
- YAML syntax
- Required fields
- Secret references
- Profile references in routes

## Next Steps

- Learn about [Routing Strategies](/docs/routing/)
- Explore [Adapter Types](/docs/adapters/)
- Set up [Budget Tracking](/docs/user-guide/budgets/)
