# Frequently Asked Questions (FAQ)

## General Questions

### What is BobaMixer?

BobaMixer is a CLI and TUI tool for tracking, analyzing, and optimizing AI/LLM API usage and costs. It helps developers understand their AI spending patterns and make data-driven decisions about model selection and usage.

### Why the name "BobaMixer"?

Like mixing the perfect boba tea drink, BobaMixer helps you find the right blend of AI models for your needs - balancing cost, performance, and quality.

### What providers does BobaMixer support?

BobaMixer supports any provider with a REST API (Anthropic, OpenAI, OpenRouter, etc.), CLI tools, and MCP (Model Context Protocol) servers. See the [Adapter Guide](ADAPTERS.md) for details.

### Is BobaMixer free?

Yes! BobaMixer is open-source under the MIT license. You only pay for the API usage from your chosen providers.

## Installation & Setup

### How do I install BobaMixer?

**macOS/Linux with Homebrew:**
```bash
brew tap royisme/tap
brew install bobamixer
```

**From source:**
```bash
go install github.com/royisme/bobamixer/cmd/boba@latest
```

See [README](../README.md) for more options.

### Where does BobaMixer store data?

All data is stored in `~/.boba/`:
- `profiles.yaml` - Profile configurations
- `routes.yaml` - Routing rules
- `pricing.yaml` - Model pricing
- `secrets.yaml` - API keys
- `usage.db` - SQLite database with usage records
- `logs/` - Application logs

### How do I migrate from another tool?

Currently, you'll need to manually configure profiles. We're considering import tools for future releases. If you have specific migration needs, please open an issue.

## Configuration

### How do I add a new API provider?

1. Add profile to `~/.boba/profiles.yaml`:
```yaml
my-provider:
  adapter: http
  endpoint: https://api.example.com/v1/chat
  model: model-name
  headers:
    Authorization: "Bearer secret://my_key"
```

2. Add key to `~/.boba/secrets.yaml`:
```yaml
my_key: your-api-key-here
```

3. Test it:
```bash
boba use my-provider
boba doctor
```

### Can I use multiple API keys for the same provider?

Yes! Create separate profiles:
```yaml
anthropic-personal:
  adapter: http
  endpoint: https://api.anthropic.com/v1/messages
  headers:
    x-api-key: "secret://anthropic_personal"

anthropic-work:
  adapter: http
  endpoint: https://api.anthropic.com/v1/messages
  headers:
    x-api-key: "secret://anthropic_work"
```

### How do routing rules work?

Rules are evaluated in order. First match wins. Example:
```yaml
rules:
  - id: large-context
    if: "ctx_chars > 50000"
    use: high-capacity-model
    
  - id: simple-task
    if: "text.contains('format') || text.contains('lint')"
    use: fast-model
```

See [Routing Cookbook](ROUTING_COOKBOOK.md) for more examples.

## Usage & Features

### Does BobaMixer intercept my API calls?

No! BobaMixer is a tracking and management tool. You explicitly invoke it when you want to track usage. It doesn't intercept or proxy your existing API calls.

### What data does BobaMixer collect?

BobaMixer only stores:
- Token counts (input/output)
- Cost calculations
- Latency measurements
- Session metadata (project, branch, timestamps)
- Estimate accuracy level

It **never** stores:
- API keys
- Request/response content
- User prompts or model outputs

### Can I use BobaMixer in CI/CD?

Yes! BobaMixer works great in CI/CD pipelines:
```bash
# In CI
export BOBA_HOME=/tmp/boba-ci
boba use ci-profile
# ... use your AI tool ...
boba stats --today
```

### How accurate are the cost estimates?

Accuracy depends on the estimation level:
- **Exact** (from API response): 100% accurate
- **Mapped** (from pricing config): 95-99% accurate
- **Heuristic** (character-based): 70-90% accurate

Check estimate level in reports:
```bash
boba report --format json | jq '.[] | select(.estimate != "exact")'
```

### What is epsilon-greedy exploration?

BobaMixer can automatically test different profiles (3% of requests by default) to discover better options. This helps you find cost-performance sweet spots without manual testing.

Disable if desired:
```yaml
# In routes.yaml
exploration:
  enabled: false
```

## Troubleshooting

### "Database is locked" error

This means another boba process is accessing the database:
```bash
# Find and kill the process
ps aux | grep boba
pkill -f boba

# Or wait for it to finish
```

### "secrets.yaml must have 0600 permissions"

Fix permissions:
```bash
chmod 600 ~/.boba/secrets.yaml
```

This is a security feature to protect your API keys.

### My TUI looks garbled

Check your terminal:
```bash
echo $TERM  # Should be xterm-256color or similar
TERM=xterm-256color boba
```

### API calls are failing

1. **Check connectivity:**
```bash
boba doctor
```

2. **Verify API key:**
```bash
# Make sure key is in secrets.yaml
grep your_key_name ~/.boba/secrets.yaml

# Check key format
# Should be: key_name: actual-key-value
```

3. **Test with curl:**
```bash
curl -v https://api.anthropic.com/v1/messages \
  -H "x-api-key: $YOUR_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"claude-3-5-sonnet-20241022","max_tokens":10,"messages":[{"role":"user","content":"test"}]}'
```

### How do I reset everything?

```bash
# Backup first!
cp -r ~/.boba ~/.boba.backup

# Remove all data
rm -rf ~/.boba

# Reinitialize
boba doctor
```

## Budget & Costs

### How do budgets work?

Budgets are advisory only - they warn but don't block:
- **Daily budget**: Spending limit per day
- **Hard cap**: Maximum spending per period

When approaching limits, you'll see alerts in TUI and via `boba action`.

### Can budgets block API calls?

No. BobaMixer uses an "alert, don't interrupt" philosophy. It will warn you but never block your work.

### How do I set a budget?

In `~/.boba/profiles.yaml` or `.boba-project.yaml`:
```yaml
budget:
  daily_usd: 10.00
  hard_cap: 100.00
  period_days: 30
```

Or via CLI:
```bash
boba budget --set daily 10 --project my-app
```

### Where does pricing data come from?

1. Remote sources (if configured)
2. Local `pricing.yaml`
3. Profile-specific `cost_per_1k` values

BobaMixer falls back gracefully if remote sources are unavailable.

## Advanced

### Can I write custom adapters?

Yes! See [Adapter Guide](ADAPTERS.md) for details. Adapters are Go plugins that implement a simple interface.

### Can I export data to other tools?

Yes! Export to JSON or CSV:
```bash
# JSON
boba report --format json --output report.json

# CSV
boba report --format csv --output report.csv
```

Then analyze with your preferred tools (Excel, Python, etc.).

### Does BobaMixer support streaming?

The tool adapter supports streaming for CLI tools. HTTP adapter support is planned.

### Can multiple users share a configuration?

Yes! Use shared pricing/routes with individual secrets:
```bash
# Shared configs
ln -s /team/boba/pricing.yaml ~/.boba/pricing.yaml
ln -s /team/boba/routes.yaml ~/.boba/routes.yaml

# Individual secrets (never shared!)
cp my-secrets.yaml ~/.boba/secrets.yaml
chmod 600 ~/.boba/secrets.yaml
```

### How do I contribute?

We welcome contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines. Areas we need help:
- Additional adapter types
- UI/UX improvements
- Documentation
- Testing
- Feature requests

## Privacy & Security

### Is my data sent anywhere?

No. BobaMixer runs entirely locally. The only network calls are:
1. Your API calls to providers (you control this)
2. Optional pricing updates (if configured)

### Are API keys secure?

Yes:
- Stored in `secrets.yaml` with 0600 permissions (readable only by you)
- Never logged
- Never included in reports or exports
- Referenced by name (`secret://key_name`) in configs

### Can I use BobaMixer in air-gapped environments?

Yes! All features work offline except:
- Remote pricing updates (use local pricing.yaml)
- API calls (obviously)

### What data is in the database?

Only usage metrics:
```sql
SELECT * FROM usage_records LIMIT 1;
-- Returns: tokens, cost, timestamp, session_id, profile, model
-- Never: API keys, prompts, responses
```

## Performance

### How much disk space does BobaMixer use?

- Installation: ~10MB
- Configurations: <1MB
- Database growth: ~1KB per API call
- Logs: ~10MB (auto-rotated)

Typical usage: 50-100MB total.

### Does BobaMixer slow down my API calls?

No. BobaMixer only tracks metadata, not the actual API call. Overhead is negligible (<1ms).

### How do I limit database growth?

Purge old records:
```bash
# Delete records older than 90 days
sqlite3 ~/.boba/usage.db "DELETE FROM usage_records WHERE ts < strftime('%s', 'now', '-90 days');"
sqlite3 ~/.boba/usage.db "VACUUM;"
```

See [Operations Guide](OPERATIONS.md) for automated cleanup.

## Still Have Questions?

- **Documentation**: Check [docs/](.)
- **Issues**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)
- **Health Check**: Run `boba doctor`

---

**Didn't find your answer?** Open an issue and we'll add it to this FAQ!
