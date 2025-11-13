# Getting Started

Welcome to BobaMixer! This guide will help you get started with BobaMixer, a comprehensive CLI tool for managing multiple AI providers, tracking costs, and optimizing your AI workload routing.

## What is BobaMixer?

BobaMixer is a smart AI adapter router that helps you:

- **Track Usage**: Monitor tokens, costs, and latency across multiple AI providers
- **Route Intelligently**: Automatically select the best model based on context and task type
- **Manage Budgets**: Set spending limits and receive proactive alerts
- **Optimize Costs**: Get AI-powered suggestions to reduce spending
- **Analyze Patterns**: Understand your AI usage with comprehensive analytics

## Quick Start

### 1. Install BobaMixer

Choose your preferred installation method:

**Using Go:**
```bash
go install github.com/royisme/bobamixer/cmd/boba@latest
```

**Using Homebrew (macOS/Linux):**
```bash
brew tap royisme/tap
brew install bobamixer
```

**Download Binary:**
Download the latest release from [GitHub Releases](https://github.com/royisme/BobaMixer/releases).

For detailed installation instructions, see the [Installation Guide](/guide/installation).

### 2. Initialize Configuration

Run the doctor command to create default configurations:

```bash
boba doctor
```

This creates the `~/.boba/` directory with example configuration files:
- `profiles.yaml` - Profile definitions
- `routes.yaml` - Routing rules
- `pricing.yaml` - Model pricing
- `secrets.yaml` - API keys (0600 permissions)
- `usage.db` - SQLite database for tracking

### 3. Configure Your First Profile

Edit `~/.boba/profiles.yaml` and add your first profile:

```yaml
default:
  adapter: http
  provider: anthropic
  endpoint: https://api.anthropic.com/v1/messages
  model: claude-3-5-sonnet-20241022
  headers:
    anthropic-version: "2023-06-01"
    x-api-key: "secret://anthropic_key"
```

### 4. Add Your API Key

Edit `~/.boba/secrets.yaml` and add your API key:

```yaml
anthropic_key: sk-ant-your-actual-key-here
```

Make sure the secrets file has correct permissions:
```bash
chmod 600 ~/.boba/secrets.yaml
```

### 5. Activate the Profile

Set the default profile as active:

```bash
boba use default
```

### 6. Verify Setup

Check that everything is configured correctly:

```bash
boba doctor
```

You should see green checkmarks for all configuration items.

### 7. Launch the TUI Dashboard

Start the interactive dashboard:

```bash
boba
```

The dashboard shows:
- Current active profile
- Today's usage statistics
- Budget status
- Recent notifications
- Quick actions

## Basic Usage

### View All Profiles

List all configured profiles:

```bash
boba ls --profiles
```

### Switch Profiles

Activate a different profile:

```bash
boba use <profile-name>
```

### Check Usage Statistics

View today's statistics:

```bash
boba stats --today
```

View last 7 days:

```bash
boba stats --7d
```

View breakdown by profile:

```bash
boba stats --7d --by-profile
```

### Test Routing Rules

Test which profile would be selected for a given input:

```bash
boba route test "Write a function to sort an array"
```

Test with file content:

```bash
boba route test @path/to/file.txt
```

### Check Budget Status

View budget status and alerts:

```bash
boba budget --status
```

View pending actions and suggestions:

```bash
boba action
```

## Next Steps

Now that you have BobaMixer up and running, explore these topics:

- **[Configuration Guide](/guide/configuration)** - Learn about all configuration options
- **[Adapters](/features/adapters)** - Understand different adapter types (HTTP, Tool, MCP)
- **[Intelligent Routing](/features/routing)** - Set up smart routing rules
- **[Budget Management](/features/budgets)** - Configure budgets and alerts
- **[Analytics](/features/analytics)** - Analyze your usage patterns

## Getting Help

If you run into any issues:

1. Run `boba doctor` to check configuration health
2. Check the [Troubleshooting Guide](/advanced/troubleshooting)
3. Review the [FAQ](/advanced/troubleshooting#faq)
4. Open an issue on [GitHub](https://github.com/royisme/BobaMixer/issues)

## Example Workflow

Here's a typical workflow using BobaMixer:

```bash
# Morning: Check yesterday's stats
boba stats --yesterday

# Set up for work session
boba use work-heavy

# Work on your project...
# BobaMixer tracks usage automatically via adapters

# Check current session
boba stats --today

# Review suggestions
boba action

# Apply a suggested optimization
boba action apply suggestion-id

# End of day: Generate report
boba report --format json --output daily-report.json
```

## Community and Support

- **Documentation**: [https://royisme.github.io/BobaMixer/](https://royisme.github.io/BobaMixer/)
- **GitHub**: [https://github.com/royisme/BobaMixer](https://github.com/royisme/BobaMixer)
- **Issues**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)

Welcome to the BobaMixer community!
