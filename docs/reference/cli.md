# CLI Reference

Complete reference for all BobaMixer command-line interface commands and options.

## Global Options

These options are available for all commands:

```bash
--help, -h           Show help message
--version, -v        Show version information
--verbose           Enable verbose output
--quiet             Suppress non-error output
--no-color          Disable colored output
--config PATH       Custom config directory (default: ~/.boba)
```

## Commands

### boba

Launch the interactive TUI dashboard.

```bash
boba [options]
```

**Options:**
- `--profile PROFILE` - Start with specific profile
- `--refresh-rate SECONDS` - Dashboard refresh rate (default: 5)

**Example:**
```bash
# Launch dashboard
boba

# Launch with specific profile
boba --profile claude-sonnet

# Faster refresh
boba --refresh-rate 2
```

---

### boba doctor

Check configuration health and diagnose issues.

```bash
boba doctor [options]
```

**Options:**
- `--verbose` - Show detailed diagnostics
- `--fix` - Attempt to fix common issues

**Checks performed:**
- Configuration file syntax
- File permissions (especially secrets.yaml)
- Database connectivity
- API endpoint accessibility
- Profile validity
- Secret references

**Example:**
```bash
# Basic health check
boba doctor

# Detailed diagnostics
boba doctor --verbose

# Auto-fix issues
boba doctor --fix
```

**Output:**
```
BobaMixer Health Check
───────────────────────────────────────
✓ Configuration directory exists
✓ Database accessible
✓ Secrets file permissions correct (0600)
✓ All profiles valid
✓ 3 profiles configured
✗ Warning: 2 profiles missing pricing info
✓ API endpoints reachable

Status: Healthy with warnings
```

---

### boba use

Set the active profile.

```bash
boba use PROFILE
```

**Arguments:**
- `PROFILE` - Name of the profile to activate

**Example:**
```bash
# Activate profile
boba use claude-sonnet

# Verify active profile
boba ls --current
```

---

### boba ls

List profiles, projects, or sessions.

```bash
boba ls [options]
```

**Options:**
- `--profiles` - List all profiles
- `--projects` - List all projects
- `--sessions` - List recent sessions
- `--current` - Show current active profile
- `--tag TAG` - Filter by tag
- `--verbose` - Show detailed information

**Example:**
```bash
# List all profiles
boba ls --profiles

# Current profile
boba ls --current

# Profiles with specific tag
boba ls --profiles --tag work

# Detailed profile info
boba ls --profiles --verbose
```

**Output:**
```
Configured Profiles
───────────────────────────────────────
Name              Adapter  Model                          Tags
claude-sonnet     http     claude-3-5-sonnet-20241022    [work, complex]
gpt-4o-mini       http     gpt-4o-mini                   [fast, cheap]
local-llama       tool     llama2                        [local, free]

Active: claude-sonnet
Total: 3 profiles
```

---

### boba stats

View usage statistics.

```bash
boba stats [options]
```

**Time Range Options:**
- `--today` - Today's statistics
- `--yesterday` - Yesterday's statistics
- `--7d` - Last 7 days
- `--30d` - Last 30 days
- `--from DATE --to DATE` - Custom date range (YYYY-MM-DD)

**Breakdown Options:**
- `--by-profile` - Group by profile
- `--by-project` - Group by project
- `--by-session` - Group by session
- `--by-estimate` - Group by estimate accuracy level

**Additional Options:**
- `--compare` - Compare profiles (with --by-profile)
- `--latency` - Show latency statistics
- `--percentiles` - Show latency percentiles (P50, P95, P99)
- `--trend PERIOD` - Show trend (daily|weekly)
- `--breakdown` - Detailed cost breakdown
- `--sort-by FIELD` - Sort by field (cost|requests|tokens|latency)

**Example:**
```bash
# Today's stats
boba stats --today

# Last 7 days by profile
boba stats --7d --by-profile

# Compare profile performance
boba stats --7d --by-profile --compare

# Latency analysis
boba stats --7d --latency --percentiles

# Custom date range
boba stats --from 2024-01-01 --to 2024-01-31

# Project breakdown
boba stats --30d --by-project

# Cost breakdown
boba stats --7d --breakdown
```

---

### boba route

Manage and test routing rules.

```bash
boba route SUBCOMMAND [options]
```

**Subcommands:**
- `test TEXT` - Test routing with text or file
- `list` - List all routing rules
- `validate` - Validate routing configuration

**Test Options:**
- `@FILE` - Test with file content
- `--verbose` - Show detailed evaluation
- `--explain` - Explain matching process

**Example:**
```bash
# Test with text
boba route test "Write a sorting function"

# Test with file
boba route test @prompts/example.txt

# Verbose output
boba route test --verbose "Format this code"

# List all rules
boba route list

# Validate config
boba route validate
```

**Test Output:**
```
Routing Test
───────────────────────────────────────
Input: "Write a sorting function"
Input Length: 1,234 characters

Matched Rule: code-generation
  Condition: text.matches('write.*function|implement')
  Profile: code-specialist
  Fallback: general-purpose
  Reason: Code generation task

Would use profile: code-specialist
```

---

### boba budget

Manage budgets and view spending.

```bash
boba budget [options]
```

**View Options:**
- `--status` - Show budget status
- `--detailed` - Detailed budget breakdown
- `--project NAME` - Specific project budget

**Set Options:**
- `--set TYPE AMOUNT` - Set budget (daily|weekly|monthly|cap)
- `--project NAME` - Set project-specific budget

**Projection Options:**
- `--project PERIOD` - Project spending (daily|weekly|monthly)

**Example:**
```bash
# View status
boba budget --status

# Detailed view
boba budget --status --detailed

# Project budget
boba budget --status --project my-app

# Set daily budget
boba budget --set daily 50

# Set project budget
boba budget --set daily 10 --project my-app

# Set hard cap
boba budget --set cap 1000

# View projection
boba budget --project monthly
```

---

### boba action

View and manage alerts and suggestions.

```bash
boba action [options]
```

**Options:**
- `--type TYPE` - Filter by type (budget|suggestion|alert)
- `apply ID` - Apply a suggestion
- `dismiss ID` - Dismiss an action
- `preview ID` - Preview before applying

**Example:**
```bash
# View all actions
boba action

# Budget alerts only
boba action --type budget

# Apply suggestion
boba action apply suggestion-123

# Preview first
boba action preview suggestion-123

# Dismiss action
boba action dismiss alert-456
```

**Output:**
```
Pending Actions
───────────────────────────────────────
ID: budget-alert-001
Type: Budget Warning
Priority: Medium
Message: Daily budget at 78% ($7.80 / $10.00)

ID: suggestion-001
Type: Optimization Suggestion
Priority: Low
Message: Switch to 'gpt-4o-mini' for simple tasks
Potential Savings: $15.20/week (32%)
Confidence: 85%
Action: boba action apply suggestion-001
```

---

### boba report

Export usage data.

```bash
boba report [options]
```

**Format Options:**
- `--format FORMAT` - Export format (json|csv)
- `--output FILE` - Output file path

**Filter Options:**
- `--from DATE --to DATE` - Date range
- `--profile PROFILE` - Specific profile
- `--project PROJECT` - Specific project

**Example:**
```bash
# Export to JSON
boba report --format json --output usage.json

# Export to CSV
boba report --format csv --output usage.csv

# Last 30 days
boba report --format json --from $(date -d '30 days ago' +%Y-%m-%d) --output last-month.json

# Specific profile
boba report --format csv --profile claude-sonnet --output claude-usage.csv

# Specific project
boba report --format json --project my-app --output my-app-usage.json
```

---

### boba session

Manage and view sessions.

```bash
boba session SUBCOMMAND [options]
```

**Subcommands:**
- `list` - List sessions
- `show ID` - Show session details
- `delete ID` - Delete a session

**List Options:**
- `--today` - Today's sessions
- `--7d` - Last 7 days
- `--30d` - Last 30 days
- `--project PROJECT` - Filter by project

**Example:**
```bash
# List recent sessions
boba session list --7d

# Show session details
boba session show session-123

# Today's sessions
boba session list --today

# Project sessions
boba session list --project my-app --30d
```

---

### boba edit

Edit configuration files.

```bash
boba edit CONFIG
```

**Arguments:**
- `profiles` - Edit profiles.yaml
- `routes` - Edit routes.yaml
- `pricing` - Edit pricing.yaml
- `secrets` - Edit secrets.yaml

**Example:**
```bash
# Edit profiles
boba edit profiles

# Edit routing rules
boba edit routes

# Edit pricing config
boba edit pricing

# Edit secrets
boba edit secrets
```

Opens configuration file in `$EDITOR` (defaults to vim/nano).

---

### boba hooks

Manage git hooks integration.

```bash
boba hooks SUBCOMMAND [options]
```

**Subcommands:**
- `install` - Install git hooks in current repository
- `remove` - Remove git hooks
- `status` - Show hook installation status

**Install Options:**
- `--force` - Overwrite existing hooks

**Example:**
```bash
# Install hooks in current repo
cd my-project
boba hooks install

# Force reinstall
boba hooks install --force

# Check status
boba hooks status

# Remove hooks
boba hooks remove
```

---

### boba version

Show version information.

```bash
boba version [options]
```

**Options:**
- `--check-update` - Check for newer version

**Example:**
```bash
# Show version
boba version

# Check for updates
boba version --check-update
```

**Output:**
```
BobaMixer v1.2.3
Commit: abc1234
Built: 2024-01-15T10:30:00Z
Go Version: go1.22.0

Update available: v1.3.0
Run: brew upgrade bobamixer
```

---

## Environment Variables

```bash
# Custom config directory
export BOBA_HOME=/custom/path

# Log level (trace|debug|info|warn|error)
export BOBA_LOG_LEVEL=debug

# Custom database path
export BOBA_DB_PATH=/custom/usage.db

# Editor for boba edit
export EDITOR=vim

# Disable colors
export NO_COLOR=1

# Force TTY mode
export BOBA_FORCE_TTY=1

# API timeout (seconds)
export BOBA_API_TIMEOUT=30
```

## Exit Codes

```bash
0   # Success
1   # General error
2   # Configuration error
3   # Database error
4   # API error
5   # Permission error
10  # User cancelled
```

## Shell Completion

### Bash

```bash
# Add to ~/.bashrc
eval "$(boba completion bash)"

# Or generate to file
boba completion bash > /etc/bash_completion.d/boba
```

### Zsh

```bash
# Add to ~/.zshrc
eval "$(boba completion zsh)"

# Or generate to file
boba completion zsh > "${fpath[1]}/_boba"
```

### Fish

```bash
# Generate completion
boba completion fish > ~/.config/fish/completions/boba.fish
```

## Common Workflows

### Daily Usage Check

```bash
# Morning routine
boba stats --yesterday
boba budget --status
boba action

# If alerts exist
boba action apply <id>
```

### Profile Switching

```bash
# List available
boba ls --profiles

# Switch
boba use <profile>

# Verify
boba ls --current
```

### Weekly Review

```bash
# View stats
boba stats --7d --by-profile

# Export for analysis
boba report --format csv --7d --output weekly-$(date +%Y-%m-%d).csv

# Check budget
boba budget --status --detailed
```

### Testing Routing

```bash
# Create test cases
mkdir route-tests
echo "Generate code" > route-tests/gen.txt
echo "Review code" > route-tests/review.txt

# Test all
for f in route-tests/*.txt; do
  boba route test @"$f"
done
```

### Budget Management

```bash
# Check status
boba budget --status

# View projection
boba budget --project monthly

# Adjust if needed
boba budget --set daily 60

# Apply optimizations
boba action --type suggestion
boba action apply <id>
```

## Tips and Tricks

### Quick Stats

```bash
# Alias for quick checks
alias bs='boba stats'
alias bst='boba stats --today'
alias bs7='boba stats --7d --by-profile'
```

### Budget Monitoring

```bash
# Daily cron job
0 9 * * * boba budget --status | mail -s "Daily BobaMixer Budget" you@example.com
```

### Auto-Export

```bash
# Weekly export
0 0 * * 0 boba report --format csv --7d --output ~/reports/boba-$(date +%Y-%m-%d).csv
```

### Profile Quick Switch

```bash
# Functions in ~/.bashrc
fast() { boba use fast-model; }
quality() { boba use quality-model; }
cheap() { boba use economical-model; }
```

### JSON Processing

```bash
# Install jq for JSON processing
brew install jq

# Find expensive requests
boba report --format json | jq 'select(.cost_usd > 1.0)'

# Top 10 by cost
boba report --format json | jq 'sort_by(.cost_usd) | reverse | .[0:10]'

# Group by profile
boba report --format json | jq 'group_by(.profile) | map({profile: .[0].profile, total: map(.cost_usd) | add})'
```

## Troubleshooting Commands

### Configuration Issues

```bash
# Full diagnostic
boba doctor --verbose

# Check specific config
cat ~/.boba/profiles.yaml
yamllint ~/.boba/profiles.yaml
```

### Database Issues

```bash
# Check database
sqlite3 ~/.boba/usage.db "PRAGMA integrity_check;"

# Vacuum
sqlite3 ~/.boba/usage.db "VACUUM;"

# Analyze
sqlite3 ~/.boba/usage.db "ANALYZE;"
```

### Permission Issues

```bash
# Fix secrets permissions
chmod 600 ~/.boba/secrets.yaml

# Fix directory permissions
chmod 700 ~/.boba
```

### API Issues

```bash
# Test connectivity
boba doctor

# Enable debug logging
export BOBA_LOG_LEVEL=debug
boba route test "test"

# Check logs
tail -f ~/.boba/logs/boba.log
```

## Next Steps

- **[Configuration Files Reference](/reference/config-files)** - Detailed config schemas
- **[Getting Started](/guide/getting-started)** - Basic usage guide
- **[Troubleshooting](/advanced/troubleshooting)** - Common issues
