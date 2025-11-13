# BobaMixer Quickstart Guide

## Installation

### Prerequisites

- Go 1.22+
- macOS / Linux (Windows via WSL)
- SQLite3 (optional, for database operations)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/vantagecraft-dev/bobamixer.git
cd bobamixer

# Build the binary
go build -o boba ./cmd/boba

# Move to PATH (optional)
sudo mv boba /usr/local/bin/
```

## Initial Setup

### 1. Create Configuration Directory

```bash
mkdir -p ~/.boba/logs
chmod 700 ~/.boba
```

### 2. Copy Example Configurations

```bash
# Copy example configurations
cp configs/examples/profiles.yaml ~/.boba/
cp configs/examples/routes.yaml ~/.boba/
cp configs/examples/pricing.yaml ~/.boba/
cp configs/examples/secrets.yaml ~/.boba/

# Secure secrets file
chmod 600 ~/.boba/secrets.yaml
```

### 3. Edit Secrets

Edit `~/.boba/secrets.yaml` and add your API keys:

```yaml
secrets:
  anthropic: "sk-ant-YOUR-KEY-HERE"
  openrouter: "sk-or-YOUR-KEY-HERE"
  deepseek: "sk-ds-YOUR-KEY-HERE"
```

## Basic Usage

### List Profiles

```bash
boba ls --profiles
```

### Set Active Profile

```bash
boba use work-heavy
```

### View Today's Statistics

```bash
boba stats --today
```

### Edit Configuration

```bash
boba edit profiles
boba edit routes
boba edit pricing
boba edit secrets
```

### Check Configuration Health

```bash
boba doctor
```

### View Budget Status

```bash
boba budget --status
```

## Configuration Files

### profiles.yaml

Defines model profiles with their endpoints, costs, and settings.

```yaml
profiles:
  work-heavy:
    name: "Work Heavy Tasks"
    adapter: "http"
    provider: "anthropic"
    endpoint: "https://api.anthropic.com"
    model: "claude-3-5-sonnet-latest"
    max_tokens: 4096
    temperature: 0.7
    tags: ["work","complex","analysis"]
    cost_per_1k:
      input: 0.015
      output: 0.075
    env:
      ANTHROPIC_API_KEY: "secret://anthropic"
```

### routes.yaml

Defines routing rules for automatic profile selection.

```yaml
rules:
  - id: "formatting"
    if: "intent=='format' || text.matches('\\bformat\\b|\\bprettier\\b')"
    use: "quick-tasks"
    explain: "格式化类任务优先低成本"
```

### pricing.yaml

Configures pricing sources and model costs.

```yaml
models:
  "anthropic/claude-3-5-sonnet-latest":
    input_per_1k: 0.015
    output_per_1k: 0.075

sources:
  - type: "http-json"
    url: "https://raw.githubusercontent.com/vantagecraft-dev/boba-mixer-pricing/main/pricing.json"
    priority: 10
```

## Project-Level Configuration

Create a `.boba-project.yaml` in your project root:

```yaml
project:
  name: "my-project"
  type: ["go","typescript"]
  preferred_profiles: ["work-heavy","quick-tasks"]

routing:
  rules:
    - if: "task.contains('format')"
      use: "quick-tasks"

budget:
  daily_usd: 5.0
  hard_cap: 50.0
```

## Next Steps

- Read the [Architecture Guide](docs/architecture.md)
- Check out [Routing Cookbook](docs/routing-cookbook.md)
- Learn about [Adapters](docs/adapters.md)
