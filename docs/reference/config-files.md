# Configuration Files Reference

Complete reference for all BobaMixer configuration file schemas and options.

## Configuration Directory

Default location: `~/.boba/`

Override with: `export BOBA_HOME=/custom/path`

```
~/.boba/
├── profiles.yaml       # Profile definitions
├── routes.yaml         # Routing rules
├── pricing.yaml        # Model pricing
├── secrets.yaml        # API keys (0600 permissions)
├── usage.db            # SQLite database
├── logs/               # Application logs
└── pricing.cache.json  # Cached pricing data (auto-generated)
```

## profiles.yaml

Defines AI provider profiles and global settings.

### Schema

```yaml
# Global settings (optional)
global:
  budget:
    daily_usd: float
    weekly_usd: float
    monthly_usd: float
    hard_cap: float
    period_days: int
    alert_at_percent: int
    critical_at_percent: int

# Profile definitions
profile-name:
  # Required
  adapter: http|tool|mcp

  # HTTP adapter (when adapter=http)
  provider: string              # anthropic|openai|openrouter|custom
  endpoint: string              # API endpoint URL
  headers:                      # HTTP headers
    header-name: string

  # Tool adapter (when adapter=tool)
  command: string               # Executable path
  args: [string]               # Command arguments
  working_dir: string          # Working directory
  timeout_seconds: int         # Timeout

  # MCP adapter (when adapter=mcp)
  command: string              # MCP server command
  args: [string]              # Command arguments
  transport: stdio|sse        # Transport type

  # Common settings
  model: string                # Model identifier
  max_tokens: int              # Max tokens per request
  temperature: float           # Temperature (0.0-1.0)
  env: [string]               # Environment variables
  tags: [string]              # Profile tags

  # Cost tracking
  cost_per_1k:
    input: float              # Cost per 1K input tokens (USD)
    output: float             # Cost per 1K output tokens (USD)

  # Profile-specific budget (optional)
  budget:
    daily_usd: float
    monthly_usd: float

  # Additional parameters (optional)
  params:
    key: value                # Adapter-specific parameters
```

### Example: HTTP Adapter (Anthropic)

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
  tags: [work, complex, production]
  cost_per_1k:
    input: 0.015
    output: 0.075
  budget:
    daily_usd: 10.00
```

### Example: HTTP Adapter (OpenAI)

```yaml
gpt-4-turbo:
  adapter: http
  provider: openai
  endpoint: https://api.openai.com/v1/chat/completions
  model: gpt-4-turbo-preview
  max_tokens: 4096
  temperature: 0.7
  headers:
    Authorization: "Bearer secret://openai_key"
    Content-Type: "application/json"
  tags: [work, vision]
  cost_per_1k:
    input: 0.01
    output: 0.03
```

### Example: Tool Adapter

```yaml
claude-code:
  adapter: tool
  command: /usr/local/bin/claude
  args: ["--session", "work", "--format", "jsonl"]
  env:
    - ANTHROPIC_API_KEY=secret://anthropic_key
    - CLAUDE_OUTPUT_FORMAT=jsonl
  working_dir: /home/user/projects
  timeout_seconds: 300
  tags: [development, coding]
  cost_per_1k:
    input: 0.015
    output: 0.075
```

### Example: MCP Adapter

```yaml
mcp-filesystem:
  adapter: mcp
  command: npx
  args: ["@modelcontextprotocol/server-filesystem", "/path/to/docs"]
  transport: stdio
  env:
    - NODE_ENV=production
  tags: [mcp, tools, filesystem]
```

### Example: Global Budget

```yaml
global:
  budget:
    daily_usd: 50.00
    hard_cap: 1000.00
    period_days: 30
    alert_at_percent: 75
    critical_at_percent: 90

claude-sonnet:
  # ... profile config ...

gpt-4-turbo:
  # ... profile config ...
```

### Field Reference

#### adapter

**Type**: `string`
**Required**: Yes
**Values**: `http`, `tool`, `mcp`

Specifies the adapter type to use.

#### provider

**Type**: `string`
**Required**: For HTTP adapter
**Values**: `anthropic`, `openai`, `openrouter`, `custom`

Provider identifier for HTTP adapters.

#### endpoint

**Type**: `string`
**Required**: For HTTP adapter
**Format**: URL

API endpoint for HTTP requests.

#### model

**Type**: `string`
**Required**: Yes

Model identifier (e.g., `claude-3-5-sonnet-20241022`).

#### headers

**Type**: `map[string]string`
**Required**: For HTTP adapter

HTTP headers to send with requests. Use `secret://key_name` for sensitive values.

#### command

**Type**: `string`
**Required**: For tool/MCP adapter
**Format**: Executable path

Path to executable for tool/MCP adapters.

#### args

**Type**: `array[string]`
**Required**: No

Command-line arguments.

#### env

**Type**: `array[string]`
**Required**: No
**Format**: `KEY=value`

Environment variables. Use `secret://key_name` for sensitive values.

#### tags

**Type**: `array[string]`
**Required**: No

Tags for organizing and filtering profiles.

#### cost_per_1k

**Type**: `object`
**Required**: No

Cost per 1,000 tokens in USD.

```yaml
cost_per_1k:
  input: 0.015
  output: 0.075
```

#### budget

**Type**: `object`
**Required**: No

Profile-specific or global budget settings.

---

## routes.yaml

Defines intelligent routing rules.

### Schema

```yaml
# Routing rules
rules:
  - id: string                  # Unique rule identifier
    if: string                  # Condition expression
    use: string                 # Profile to use
    fallback: string            # Fallback profile (optional)
    explain: string             # Explanation

# Exploration settings (optional)
exploration:
  enabled: boolean              # Enable epsilon-greedy exploration
  epsilon: float                # Exploration rate (0.0-1.0)
  min_samples: int              # Min samples before exploring
  cooldown_hours: int           # Hours between re-testing
  exclude_profiles: [string]    # Profiles to exclude from exploration
```

### Example: Complete Routing Configuration

```yaml
rules:
  # Large context
  - id: extra-large-context
    if: "ctx_chars > 100000"
    use: claude-opus
    fallback: claude-sonnet
    explain: "Very large context requires highest capacity model"

  - id: large-context
    if: "ctx_chars > 50000"
    use: claude-sonnet
    fallback: claude-haiku
    explain: "Large context needs capable model"

  # Task type
  - id: code-generation
    if: "text.matches('write.*function|implement|create.*class')"
    use: code-specialist
    fallback: general-purpose
    explain: "Code generation task"

  - id: code-review
    if: "text.matches('review|analyze.*code|refactor')"
    use: code-reviewer
    explain: "Code review task"

  - id: formatting
    if: "text.matches('format|prettier|eslint|lint')"
    use: fast-formatter
    explain: "Simple formatting task"

  # Project type
  - id: frontend-work
    if: "project_types.includes('react') || project_types.includes('vue')"
    use: frontend-specialist
    explain: "Frontend development"

  - id: backend-work
    if: "project_types.includes('go')"
    use: backend-specialist
    explain: "Backend development"

  # Branch-based
  - id: production-branch
    if: "branch.matches('main|master|prod')"
    use: high-accuracy
    fallback: medium-accuracy
    explain: "Production branch needs highest accuracy"

  # Time-based
  - id: night-mode
    if: "time_of_day == 'night'"
    use: cost-optimized
    explain: "Off-peak hours, use cheaper model"

  # Default fallback
  - id: default
    if: "ctx_chars > 0"
    use: balanced-model
    explain: "Default for unmatched cases"

# Exploration configuration
exploration:
  enabled: true
  epsilon: 0.03
  min_samples: 10
  cooldown_hours: 24
  exclude_profiles:
    - production-only
    - critical-tasks
```

### DSL Reference

#### Variables

| Variable | Type | Description |
|----------|------|-------------|
| `ctx_chars` | int | Input character count |
| `text` | string | Input text content |
| `project_types` | array | Project types from `.boba-project.yaml` |
| `branch` | string | Git branch name |
| `time_of_day` | string | Time period (morning\|day\|evening\|night) |
| `intent` | string | Detected intent (if configured) |

#### Functions

| Function | Description | Example |
|----------|-------------|---------|
| `text.matches(pattern)` | Regex match | `text.matches('\\bcode\\b')` |
| `text.contains(str)` | Substring search | `text.contains('review')` |
| `array.includes(item)` | Array contains | `project_types.includes('go')` |

#### Operators

- **Comparison**: `>`, `<`, `>=`, `<=`, `==`, `!=`
- **Logical**: `&&` (and), `||` (or), `!` (not)
- **Grouping**: `(...)` for precedence

### Field Reference

#### rules[].id

**Type**: `string`
**Required**: Yes

Unique identifier for the rule.

#### rules[].if

**Type**: `string`
**Required**: Yes

Condition expression using DSL.

#### rules[].use

**Type**: `string`
**Required**: Yes

Profile to use when condition matches.

#### rules[].fallback

**Type**: `string`
**Required**: No

Fallback profile if primary fails.

#### rules[].explain

**Type**: `string`
**Required**: Yes

Human-readable explanation of the rule.

#### exploration.enabled

**Type**: `boolean`
**Default**: `false`

Enable epsilon-greedy exploration.

#### exploration.epsilon

**Type**: `float`
**Default**: `0.03`
**Range**: 0.0-1.0

Exploration rate (e.g., 0.03 = 3% of requests).

---

## pricing.yaml

Defines model pricing sources and direct pricing.

### Schema

```yaml
# Direct model pricing
models:
  "provider/model-name":
    input_per_1k: float       # Cost per 1K input tokens (USD)
    output_per_1k: float      # Cost per 1K output tokens (USD)

# Remote pricing sources
sources:
  - type: string              # http-json
    url: string               # Source URL
    priority: int             # Priority (higher = preferred)
    cache_hours: int          # Cache duration
```

### Example: Complete Pricing Configuration

```yaml
# Direct model pricing
models:
  # Anthropic models
  "anthropic/claude-3-5-sonnet-20241022":
    input_per_1k: 0.015
    output_per_1k: 0.075

  "anthropic/claude-3-opus-20240229":
    input_per_1k: 0.015
    output_per_1k: 0.075

  "anthropic/claude-3-haiku-20240307":
    input_per_1k: 0.00025
    output_per_1k: 0.00125

  # OpenAI models
  "openai/gpt-4-turbo-preview":
    input_per_1k: 0.01
    output_per_1k: 0.03

  "openai/gpt-4o-mini":
    input_per_1k: 0.00015
    output_per_1k: 0.0006

  # OpenRouter models
  "openrouter/anthropic/claude-3-5-sonnet":
    input_per_1k: 0.015
    output_per_1k: 0.075

# Remote pricing sources
sources:
  # Primary source
  - type: http-json
    url: https://raw.githubusercontent.com/username/pricing-repo/main/pricing.json
    priority: 10
    cache_hours: 24

  # Backup source
  - type: http-json
    url: https://backup-pricing.example.com/pricing.json
    priority: 5
    cache_hours: 24
```

### Remote Pricing JSON Format

```json
{
  "models": {
    "anthropic/claude-3-5-sonnet-20241022": {
      "input_per_1k": 0.015,
      "output_per_1k": 0.075
    },
    "openai/gpt-4-turbo-preview": {
      "input_per_1k": 0.01,
      "output_per_1k": 0.03
    }
  },
  "updated_at": "2024-01-15T10:00:00Z"
}
```

### Pricing Priority

BobaMixer resolves pricing in this order:

1. **Profile-specific** `cost_per_1k` (highest priority)
2. **Remote sources** by priority value (higher first)
3. **Local models** in pricing.yaml
4. **Heuristic estimation** (lowest priority)

---

## secrets.yaml

Stores sensitive data like API keys.

### Schema

```yaml
secret_name: string
another_secret: string
# ... more secrets ...
```

### Security Requirements

```bash
# Must have 0600 permissions
chmod 600 ~/.boba/secrets.yaml
```

BobaMixer will refuse to run if permissions are incorrect.

### Example

```yaml
# API Keys
anthropic_key: sk-ant-api03-xxxxxxxxxxxxxxxxxxxxx
openai_key: sk-proj-xxxxxxxxxxxxxxxxxxxxxxxxxx
openrouter_key: sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxxxx

# Custom secrets
github_token: ghp_xxxxxxxxxxxxxxxxxxxx
database_password: super-secure-password
custom_api_key: custom-key-here
```

### Referencing Secrets

Use `secret://` prefix in other configuration files:

```yaml
# In profiles.yaml
headers:
  x-api-key: "secret://anthropic_key"
  Authorization: "Bearer secret://openai_key"

env:
  - GITHUB_TOKEN=secret://github_token
  - DB_PASSWORD=secret://database_password
```

### Best Practices

1. **Never commit** secrets.yaml to version control
2. **Maintain 0600 permissions** for security
3. **Use descriptive names** for clarity
4. **Rotate keys** periodically
5. **Backup securely** (encrypted backups only)

---

## .boba-project.yaml

Project-specific configuration (optional).

Place in project root directory.

### Schema

```yaml
# Project metadata
project:
  name: string                  # Project name
  type: [string]               # Project types (go, python, react, etc.)
  preferred_profiles: [string] # Preferred profiles for this project

# Project-specific routing rules
routing:
  rules:
    - id: string
      if: string
      use: string
      fallback: string
      explain: string

# Project budget
budget:
  daily_usd: float
  weekly_usd: float
  monthly_usd: float
  hard_cap: float
  period_days: int
  alert_at_percent: int
  critical_at_percent: int
```

### Example: Complete Project Configuration

```yaml
project:
  name: my-awesome-app
  type: [typescript, react, nodejs]
  preferred_profiles:
    - frontend-specialist
    - fast-model

routing:
  rules:
    # Project-specific rules (evaluated before global rules)
    - id: component-work
      if: "text.matches('component|JSX|React')"
      use: react-specialist
      explain: "React component work"

    - id: api-work
      if: "text.matches('API|endpoint|route')"
      use: backend-specialist
      explain: "API development"

    - id: styling
      if: "text.matches('CSS|style|Tailwind')"
      use: fast-model
      explain: "Styling work doesn't need premium model"

budget:
  daily_usd: 10.00
  weekly_usd: 50.00
  hard_cap: 200.00
  period_days: 30
  alert_at_percent: 80
  critical_at_percent: 95
```

### Field Reference

#### project.name

**Type**: `string`
**Required**: Yes

Project identifier.

#### project.type

**Type**: `array[string]`
**Required**: No

Project technology stack. Used in routing rules via `project_types` variable.

Common values: `go`, `rust`, `python`, `javascript`, `typescript`, `react`, `vue`, `nodejs`, `java`, etc.

#### project.preferred_profiles

**Type**: `array[string]`
**Required**: No

Profiles recommended for this project. Used for suggestions.

---

## Configuration Validation

### Validate All Configuration

```bash
boba doctor
```

### Validate YAML Syntax

```bash
yamllint ~/.boba/profiles.yaml
yamllint ~/.boba/routes.yaml
yamllint ~/.boba/pricing.yaml
```

### Test Routing Rules

```bash
boba route validate
boba route test "sample text"
```

### Verify Secret References

```bash
# Check that all referenced secrets exist
grep -r "secret://" ~/.boba/*.yaml | while read line; do
  secret=$(echo "$line" | grep -o 'secret://[a-zA-Z_][a-zA-Z0-9_]*' | cut -d/ -f3)
  if ! grep -q "^$secret:" ~/.boba/secrets.yaml; then
    echo "Missing secret: $secret"
  fi
done
```

## Configuration Templates

### Minimal Configuration

```yaml
# profiles.yaml
default:
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

```yaml
# routes.yaml
rules:
  - id: default
    if: "ctx_chars > 0"
    use: default
    explain: "Default profile for all requests"
```

```yaml
# secrets.yaml
anthropic_key: sk-ant-your-key-here
```

### Multi-Provider Configuration

See [examples directory](https://github.com/royisme/BobaMixer/tree/main/examples/configs) for complete multi-provider setups.

## Next Steps

- **[CLI Reference](/reference/cli)** - Command-line interface
- **[Configuration Guide](/guide/configuration)** - Detailed setup guide
- **[Adapters](/features/adapters)** - Adapter configuration
- **[Routing](/features/routing)** - Routing rules guide
