# Configuration Guide

This comprehensive guide covers all BobaMixer configuration options and best practices.

## Configuration Files Overview

BobaMixer uses YAML files for configuration, stored in `~/.boba/`:

```
~/.boba/
├── profiles.yaml       # Profile definitions
├── routes.yaml         # Routing rules
├── pricing.yaml        # Model pricing information
├── secrets.yaml        # API keys (0600 permissions)
├── usage.db            # SQLite database
├── logs/               # Application logs
└── pricing.cache.json  # Cached pricing data
```

## Profiles Configuration

The `profiles.yaml` file defines your AI provider profiles.

### Basic Profile Structure

```yaml
profile-name:
  adapter: http|tool|mcp
  provider: provider-name
  endpoint: https://api.example.com
  model: model-name
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

### HTTP Adapter Example

For REST API providers like Anthropic, OpenAI, etc.:

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
  tags: [work, complex, analysis]
  cost_per_1k:
    input: 0.015
    output: 0.075

gpt4-turbo:
  adapter: http
  provider: openai
  endpoint: https://api.openai.com/v1/chat/completions
  model: gpt-4-turbo-preview
  max_tokens: 4096
  headers:
    Authorization: "Bearer secret://openai_key"
  tags: [work, vision]
  cost_per_1k:
    input: 0.01
    output: 0.03
```

### Tool Adapter Example

For CLI tools and executables:

```yaml
claude-code:
  adapter: tool
  command: /usr/local/bin/claude-code
  args: ["--session", "work"]
  env:
    - ANTHROPIC_API_KEY=secret://anthropic_key
    - CLAUDE_OUTPUT_FORMAT=jsonl
  tags: [development, coding]
  cost_per_1k:
    input: 0.015
    output: 0.075

custom-script:
  adapter: tool
  command: /path/to/my-ai-script.sh
  args: ["--model", "best"]
  tags: [custom]
```

### MCP Adapter Example

For Model Context Protocol integrations:

```yaml
mcp-filesystem:
  adapter: mcp
  command: npx
  args: ["@modelcontextprotocol/server-filesystem", "/path/to/docs"]
  transport: stdio
  tags: [mcp, tools]
```

### Profile Parameters Reference

| Parameter | Type | Description |
|-----------|------|-------------|
| `adapter` | string | Adapter type: `http`, `tool`, or `mcp` |
| `provider` | string | Provider name (for HTTP adapters) |
| `endpoint` | string | API endpoint URL (for HTTP adapters) |
| `command` | string | Command path (for tool/MCP adapters) |
| `args` | array | Command arguments (for tool/MCP adapters) |
| `model` | string | Model identifier |
| `max_tokens` | int | Maximum tokens per request |
| `temperature` | float | Temperature setting (0.0-1.0) |
| `headers` | map | HTTP headers (for HTTP adapters) |
| `env` | array | Environment variables |
| `tags` | array | Profile tags for organization |
| `cost_per_1k` | object | Cost per 1000 tokens |
| `transport` | string | Transport type (for MCP: stdio, sse) |

## Secrets Management

The `secrets.yaml` file stores sensitive data like API keys.

### Security Requirements

```bash
# Must have 0600 permissions
chmod 600 ~/.boba/secrets.yaml
```

BobaMixer will refuse to run if permissions are incorrect.

### Secrets Format

```yaml
# API Keys
anthropic_key: sk-ant-api03-xxxxx
openai_key: sk-proj-xxxxx
openrouter_key: sk-or-v1-xxxxx

# Custom secrets
my_custom_secret: custom-value
database_password: secure-password
```

### Referencing Secrets

Use the `secret://` prefix in configuration files:

```yaml
headers:
  x-api-key: "secret://anthropic_key"
  Authorization: "Bearer secret://openai_key"

env:
  - API_KEY=secret://my_custom_secret
  - DB_PASS=secret://database_password
```

## Routing Configuration

The `routes.yaml` file defines intelligent routing rules.

### Basic Routing Structure

```yaml
rules:
  - id: rule-identifier
    if: "condition expression"
    use: profile-name
    fallback: backup-profile
    explain: "Why this rule exists"

exploration:
  enabled: true
  epsilon: 0.03
  min_samples: 10
```

### Routing DSL

Available variables and functions:

| Variable | Type | Description |
|----------|------|-------------|
| `ctx_chars` | int | Input context character count |
| `text` | string | Input text content |
| `project_types` | array | Project types from `.boba-project.yaml` |
| `branch` | string | Current git branch |
| `time_of_day` | string | `morning`, `day`, `evening`, or `night` |

| Function | Description | Example |
|----------|-------------|---------|
| `text.matches(pattern)` | Regex match | `text.matches('\\bcode\\b')` |
| `text.contains(str)` | Contains substring | `text.contains('format')` |
| `array.includes(item)` | Array contains | `project_types.includes('go')` |

### Routing Examples

**Context Size-Based:**

```yaml
rules:
  - id: extra-large-context
    if: "ctx_chars > 100000"
    use: claude-opus
    fallback: claude-sonnet
    explain: "Very large context requires highest capacity model"

  - id: medium-context
    if: "ctx_chars > 10000"
    use: claude-sonnet
    fallback: claude-haiku
    explain: "Medium context, balanced model"

  - id: small-context
    if: "ctx_chars > 0"
    use: gpt-4o-mini
    explain: "Small context, economical model"
```

**Task Type-Based:**

```yaml
rules:
  - id: code-generation
    if: "text.matches('write.*function|implement|create.*class')"
    use: code-specialist
    explain: "Code generation task"

  - id: code-review
    if: "text.matches('review|analyze.*code|find.*bug')"
    use: code-reviewer
    explain: "Code review task"

  - id: formatting
    if: "text.matches('format|prettier|eslint')"
    use: fast-model
    explain: "Simple formatting, use fast model"
```

**Project Type-Based:**

```yaml
rules:
  - id: frontend-work
    if: "project_types.includes('react') || project_types.includes('vue')"
    use: frontend-specialist
    explain: "Frontend development"

  - id: backend-work
    if: "project_types.includes('go') || project_types.includes('rust')"
    use: backend-specialist
    explain: "Backend development"
```

**Time-Based:**

```yaml
rules:
  - id: night-mode
    if: "time_of_day == 'night'"
    use: cost-optimized
    explain: "Off-peak hours, use cheaper model"

  - id: business-hours
    if: "time_of_day == 'day'"
    use: high-performance
    explain: "Business hours, prioritize speed"
```

### Exploration Configuration

Epsilon-greedy exploration helps discover optimal profiles:

```yaml
exploration:
  enabled: true
  epsilon: 0.03        # 3% exploration rate
  min_samples: 10      # Minimum samples before exploration
  cooldown_hours: 24   # Hours between profile re-evaluation
```

## Pricing Configuration

The `pricing.yaml` file defines model pricing sources.

### Pricing Structure

```yaml
# Direct model pricing
models:
  "anthropic/claude-3-5-sonnet-20241022":
    input_per_1k: 0.015
    output_per_1k: 0.075

  "openai/gpt-4-turbo-preview":
    input_per_1k: 0.01
    output_per_1k: 0.03

# Remote pricing sources
sources:
  - type: http-json
    url: https://example.com/pricing.json
    priority: 10
    cache_hours: 24

  - type: http-json
    url: https://backup.example.com/pricing.json
    priority: 5
    cache_hours: 24
```

### Pricing Priority

BobaMixer uses pricing in this order:

1. Profile-specific `cost_per_1k` (highest priority)
2. Remote sources (by priority value)
3. Local `models` in pricing.yaml
4. Heuristic estimation (lowest priority)

## Project-Level Configuration

Create `.boba-project.yaml` in your project root for project-specific settings:

```yaml
project:
  name: my-awesome-project
  type: [typescript, react, nodejs]
  preferred_profiles:
    - frontend-specialist
    - fast-model

routing:
  rules:
    - id: project-specific-rule
      if: "text.contains('component')"
      use: react-specialist
      explain: "React component work"

budget:
  daily_usd: 5.00
  hard_cap: 100.00
  period_days: 30
  alert_at_percent: 80
  critical_at_percent: 95
```

### Project Configuration Priority

Project settings override global settings:

- Routes: Project rules evaluated first, then global
- Budgets: Project budgets take precedence
- Preferred profiles: Used for suggestions

## Environment Variables

BobaMixer supports environment variable configuration:

```bash
# Home directory override
export BOBA_HOME=/custom/path

# Log level
export BOBA_LOG_LEVEL=debug  # trace, debug, info, warn, error

# Database path override
export BOBA_DB_PATH=/custom/usage.db

# Disable colors
export NO_COLOR=1

# Force TTY mode
export BOBA_FORCE_TTY=1
```

## Configuration Validation

### Check Configuration Health

```bash
boba doctor
```

This validates:
- File permissions
- YAML syntax
- Secret references
- Profile configurations
- Database connectivity
- API endpoint accessibility

### Test Routing Rules

```bash
# Test with text
boba route test "Your test text here"

# Test with file
boba route test @path/to/file.txt

# Show detailed evaluation
boba route test --verbose "Test text"
```

### Validate Profiles

```bash
# List all profiles
boba ls --profiles

# Test profile activation
boba use profile-name

# Check current profile
boba ls --current
```

## Best Practices

### 1. Organize Profiles with Tags

```yaml
claude-sonnet:
  # ...
  tags: [work, complex, production]

gpt-4o-mini:
  # ...
  tags: [development, testing, fast]
```

Query by tags:
```bash
boba ls --profiles --tag work
```

### 2. Use Fallback Profiles

Always specify fallbacks for critical workflows:

```yaml
rules:
  - id: production-task
    if: "branch.matches('main|master')"
    use: best-model
    fallback: good-model
    explain: "Production work with fallback"
```

### 3. Secure Secrets

```bash
# Never commit secrets
echo "secrets.yaml" >> .gitignore

# Regular permission check
chmod 600 ~/.boba/secrets.yaml

# Rotate keys periodically
```

### 4. Monitor Costs

Set budgets for all projects:

```yaml
budget:
  daily_usd: 10.00
  hard_cap: 200.00
  alert_at_percent: 75
```

### 5. Version Control Configurations

```bash
# Track global configs (without secrets!)
cd ~/.boba
git init
git add profiles.yaml routes.yaml pricing.yaml
echo "secrets.yaml" >> .gitignore
echo "*.db" >> .gitignore
echo "logs/" >> .gitignore
git commit -m "Initial BobaMixer configuration"
```

### 6. Test Before Deploying

```bash
# Validate after changes
boba doctor

# Test routing
boba route test @test-cases.txt

# Dry run with stats
boba stats --today --dry-run
```

## Configuration Examples

### Multi-Provider Setup

```yaml
# profiles.yaml
anthropic-work:
  adapter: http
  provider: anthropic
  endpoint: https://api.anthropic.com/v1/messages
  model: claude-3-5-sonnet-20241022
  headers:
    x-api-key: "secret://anthropic_work_key"

anthropic-personal:
  adapter: http
  provider: anthropic
  endpoint: https://api.anthropic.com/v1/messages
  model: claude-3-5-sonnet-20241022
  headers:
    x-api-key: "secret://anthropic_personal_key"

openai-experiments:
  adapter: http
  provider: openai
  endpoint: https://api.openai.com/v1/chat/completions
  model: gpt-4-turbo-preview
  headers:
    Authorization: "Bearer secret://openai_key"
```

### Development vs Production

```yaml
# Development profile
dev-fast:
  adapter: http
  provider: openai
  model: gpt-4o-mini
  max_tokens: 2048
  temperature: 0.8
  tags: [development, fast, cheap]

# Production profile
prod-quality:
  adapter: http
  provider: anthropic
  model: claude-3-5-sonnet-20241022
  max_tokens: 4096
  temperature: 0.5
  tags: [production, quality, critical]
```

## Troubleshooting Configuration

### YAML Syntax Errors

```bash
# Validate YAML syntax
yamllint ~/.boba/profiles.yaml

# Or use boba doctor
boba doctor
```

### Secret Reference Errors

```bash
# Check secret exists
grep "anthropic_key" ~/.boba/secrets.yaml

# Verify reference format
# Correct: "secret://key_name"
# Wrong: "secret://key_name/"
# Wrong: "secret:key_name"
```

### Routing Not Working

```bash
# Test routing with verbose output
boba route test --verbose "Test input"

# Check rule order (first match wins!)
boba route list

# Validate conditions
boba route test --explain "Test input"
```

## Next Steps

- **[Adapters](/features/adapters)** - Learn about adapter types
- **[Routing](/features/routing)** - Master intelligent routing
- **[Budgets](/features/budgets)** - Set up budget management
- **[CLI Reference](/reference/cli)** - Explore all commands
- **[Config Files Reference](/reference/config-files)** - Detailed schema documentation
