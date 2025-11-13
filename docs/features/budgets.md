# Budget Management

BobaMixer's budget management system helps you track spending, set limits, and receive proactive alerts before costs spiral out of control.

## Overview

Key features:

- **Multi-level budgets** - Global, project, and profile levels
- **Proactive alerts** - Warning and critical thresholds
- **Cost projections** - Predict future spending
- **No blocking** - Alerts only, never interrupts workflow
- **Spending trends** - Analyze patterns over time

## Philosophy: Alert, Don't Interrupt

BobaMixer budgets are **advisory, not restrictive**:

- ✅ You receive alerts when approaching limits
- ✅ You get suggestions to optimize spending
- ✅ You maintain full control over your workflow
- ❌ API calls are never blocked
- ❌ Work is never interrupted

This design ensures productivity isn't disrupted while keeping you informed.

## Budget Levels

### 1. Global Budget

Applies to all usage across all projects and profiles.

**Configuration** in `~/.boba/profiles.yaml`:

```yaml
global:
  budget:
    daily_usd: 50.00
    hard_cap: 1000.00
    period_days: 30
    alert_at_percent: 75
    critical_at_percent: 90
```

### 2. Project Budget

Applies to a specific project.

**Configuration** in `.boba-project.yaml` (project root):

```yaml
project:
  name: my-awesome-project
  type: [typescript, react]

budget:
  daily_usd: 10.00
  hard_cap: 200.00
  period_days: 30
  alert_at_percent: 80
  critical_at_percent: 95
```

### 3. Profile Budget

Applies to a specific profile.

**Configuration** in `~/.boba/profiles.yaml`:

```yaml
expensive-model:
  adapter: http
  # ... other config ...
  budget:
    daily_usd: 5.00
    monthly_usd: 100.00
```

## Budget Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `daily_usd` | float | Daily spending limit in USD | - |
| `weekly_usd` | float | Weekly spending limit in USD | - |
| `monthly_usd` | float | Monthly spending limit in USD | - |
| `hard_cap` | float | Absolute maximum in period | - |
| `period_days` | int | Rolling period for hard_cap | 30 |
| `alert_at_percent` | int | Warning threshold (%) | 75 |
| `critical_at_percent` | int | Critical threshold (%) | 90 |

## Setting Up Budgets

### Quick Setup via CLI

```bash
# Set global daily budget
boba budget --set daily 50

# Set project budget
boba budget --set daily 10 --project my-project

# Set hard cap
boba budget --set cap 1000
```

### Configuration File Setup

**Global budget** in `~/.boba/profiles.yaml`:

```yaml
global:
  budget:
    daily_usd: 50.00
    hard_cap: 1000.00
    period_days: 30
```

**Project budget** in `.boba-project.yaml`:

```yaml
budget:
  daily_usd: 10.00
  weekly_usd: 50.00
  hard_cap: 200.00
  alert_at_percent: 80
```

## Checking Budget Status

### View Current Status

```bash
# Overall budget status
boba budget --status

# Detailed breakdown
boba budget --status --detailed

# Specific project
boba budget --status --project my-project
```

**Output example**:
```
Budget Status
─────────────────────────────────────
Global Budget
  Daily: $12.34 / $50.00 (24.7%)
  Period: $234.56 / $1000.00 (23.5%)
  Status: OK ✓

Project: my-awesome-project
  Daily: $3.45 / $10.00 (34.5%)
  Weekly: $18.92 / $50.00 (37.8%)
  Status: OK ✓

Alerts: None
```

### View in TUI Dashboard

```bash
# Launch dashboard
boba

# Budget shown in header
# Alerts shown in notification panel
```

## Alert Thresholds

### Warning Alert (Default: 75%)

Triggered when spending reaches 75% of budget.

**Actions**:
- Yellow indicator in TUI
- Alert in `boba action` output
- Suggestion to review usage

**Example**:
```
⚠️  Warning: Project budget at 78% ($7.80 / $10.00)
Suggestion: Consider switching to more economical profiles
```

### Critical Alert (Default: 90%)

Triggered when spending reaches 90% of budget.

**Actions**:
- Red indicator in TUI
- Urgent alert in `boba action` output
- Specific optimization suggestions

**Example**:
```
🚨 Critical: Project budget at 92% ($9.20 / $10.00)
Suggestions:
  1. Switch to 'economical-model' (saves ~40%)
  2. Reduce context size where possible
  3. Review routing rules for optimization
```

### Over Budget (100%+)

Budget exceeded but work continues.

**Actions**:
- Red flashing indicator
- Persistent alert
- Recommendation to review or adjust budget

**Example**:
```
🔴 Over Budget: Project budget at 112% ($11.20 / $10.00)
Exceeded by: $1.20
Recommendation: Review recent usage or adjust budget limit
```

## Cost Projections

BobaMixer predicts future spending based on current trends.

### View Projections

```bash
# See daily projection
boba budget --project daily

# See weekly projection
boba budget --project weekly

# See monthly projection
boba budget --project monthly
```

**Output example**:
```
Cost Projections
─────────────────────────────────────
Based on last 7 days average: $15.20/day

Daily projection: $15.20
Weekly projection: $106.40
Monthly projection: $456.00

Warning: Monthly projection exceeds hard cap ($456.00 > $200.00)
Recommendation: Reduce usage by 57% or adjust budget
```

### Projection Accuracy

Projections improve over time:

- **1-3 days**: Low accuracy, high variance
- **7 days**: Good accuracy for consistent usage
- **14+ days**: High accuracy, stable predictions

## Budget Notifications

### In TUI Dashboard

```bash
boba
```

Notifications panel shows:
- Recent alerts
- Budget status changes
- Projection warnings
- Optimization suggestions

### Via CLI

```bash
# View all actions
boba action

# View budget-related actions only
boba action --type budget
```

### Example Actions Output

```
Pending Actions
─────────────────────────────────────
ID: budget-alert-001
Type: Budget Warning
Priority: Medium
Message: Project 'my-app' at 78% of daily budget
Suggested Actions:
  1. Review high-cost profiles
  2. Consider routing rules adjustment

ID: budget-suggest-001
Type: Optimization Suggestion
Priority: Low
Message: Switching to 'economical-model' could save $3.20/day
Action: boba action apply budget-suggest-001
```

## Spending Analysis

### By Time Period

```bash
# Today's spending
boba stats --today

# Last 7 days
boba stats --7d

# Last 30 days
boba stats --30d

# Custom date range
boba stats --from 2024-01-01 --to 2024-01-31
```

### By Profile

```bash
# Breakdown by profile
boba stats --by-profile --7d

# Find highest cost profiles
boba stats --by-profile --30d --sort-by cost
```

**Output example**:
```
Spending by Profile (Last 7 Days)
─────────────────────────────────────
Profile              Cost       % of Total
claude-opus         $45.20      58.3%
claude-sonnet       $23.10      29.8%
gpt-4-turbo         $7.80       10.1%
gpt-4o-mini         $1.40       1.8%
─────────────────────────────────────
Total               $77.50      100%
```

### By Project

```bash
# Breakdown by project
boba stats --by-project --7d

# Specific project
boba stats --project my-app --30d
```

### Export for Analysis

```bash
# Export to CSV
boba report --format csv --output spending.csv

# Export to JSON
boba report --format json --output spending.json

# Open in Excel/Numbers
open spending.csv
```

## Cost Optimization

### Automated Suggestions

BobaMixer analyzes usage and suggests optimizations:

```bash
# View suggestions
boba action --type suggestion

# Apply a suggestion
boba action apply <suggestion-id>

# Preview before applying
boba action preview <suggestion-id>
```

### Suggestion Types

#### Profile Switching

```
Suggestion: Switch to 'gpt-4o-mini' for simple tasks
Potential Savings: $15.20/week (32%)
Confidence: 85%
Reason: 67% of your tasks are under 2000 chars and don't need premium model
```

#### Routing Rule Optimization

```
Suggestion: Add routing rule for formatting tasks
Potential Savings: $8.40/week (18%)
Confidence: 92%
Reason: 45 formatting tasks last week used expensive models unnecessarily
Suggested Rule:
  - if: "text.matches('format|prettier|lint')"
    use: fast-model
```

#### Context Reduction

```
Suggestion: Reduce context size for code reviews
Potential Savings: $5.60/week (12%)
Confidence: 78%
Reason: Large contexts detected in 23 code review tasks
Recommendation: Focus context on changed files only
```

### Manual Optimization

#### Review High-Cost Sessions

```bash
# Find expensive sessions
boba report --format json | jq 'sort_by(.cost) | reverse | .[0:10]'

# Analyze what made them expensive
boba session <session-id>
```

#### Optimize Routing Rules

```bash
# Test current routing
boba route test @typical-prompts.txt

# Identify opportunities
# Add rules for common patterns using cheaper models
```

#### Switch to Economical Profiles

```bash
# Compare profile costs
boba stats --by-profile --compare

# Switch default
boba use economical-model
```

## Budget Best Practices

### 1. Start Conservative

```yaml
# Start with low budgets
budget:
  daily_usd: 5.00
  hard_cap: 100.00

# Increase based on actual usage
```

### 2. Set Multiple Levels

```yaml
# Global catch-all
global:
  budget:
    daily_usd: 50.00
    hard_cap: 1000.00

# Project-specific
# In .boba-project.yaml
budget:
  daily_usd: 10.00
  hard_cap: 200.00
```

### 3. Monitor Regularly

```bash
# Daily check
boba budget --status

# Weekly review
boba stats --7d --by-profile

# Monthly analysis
boba stats --30d --by-project
```

### 4. Act on Alerts

```bash
# Check alerts daily
boba action

# Apply suggestions
boba action apply <id>

# Review before critical threshold
```

### 5. Use Cost-Aware Routing

```yaml
rules:
  # Expensive for critical
  - if: "text.contains('critical') || branch.matches('main')"
    use: premium-model

  # Cheap for everything else
  - if: "ctx_chars < 5000"
    use: economical-model
```

### 6. Track Trends

```bash
# Export monthly
boba report --format csv --output monthly-$(date +%Y-%m).csv

# Compare trends
# Use spreadsheet or analysis tool
```

### 7. Set Realistic Budgets

```bash
# Run for a week without budget
boba stats --7d

# Set budget at ~120% of average
# Allows room for spikes while controlling costs
```

## Common Scenarios

### Scenario 1: Exceeded Daily Budget

**Problem**: Hit daily budget at 3 PM

**Solutions**:
1. Check what caused spike: `boba stats --today --by-profile`
2. Switch to economical profile: `boba use economical-model`
3. Adjust tomorrow's routing rules
4. If justified, increase budget: `boba budget --set daily 15`

### Scenario 2: Month-End Budget Crunch

**Problem**: 5 days left in month, 10% of budget remaining

**Solutions**:
1. Review projections: `boba budget --project monthly`
2. Apply all optimization suggestions: `boba action apply --all`
3. Use only economical profiles
4. Defer non-critical work
5. Consider budget increase if needed

### Scenario 3: Unexpected Spike

**Problem**: Spending suddenly doubled

**Solutions**:
1. Identify cause: `boba stats --today --by-session`
2. Check recent sessions: `boba session list --today`
3. Review if justified (e.g., major refactor)
4. Adjust routing to prevent recurrence
5. Set project-specific budget for high-cost work

### Scenario 4: Multiple Projects

**Problem**: Managing budgets across 5 projects

**Solutions**:
1. Set individual project budgets in `.boba-project.yaml`
2. Set global cap as safety net
3. Monitor per-project: `boba stats --by-project --7d`
4. Allocate budget based on project priority
5. Review monthly and adjust allocations

## Integration with Other Features

### With Routing

```yaml
# Route to cheap models when near budget
rules:
  - id: budget-conscious
    if: "ctx_chars < 10000"
    use: economical-model
    explain: "Conserve budget"
```

### With Analytics

```bash
# Analyze budget efficiency
boba stats --by-profile --30d

# Correlate cost with value
# Export and analyze in Excel/Python
```

### With Exploration

```yaml
# Exploration can help find cheaper alternatives
exploration:
  enabled: true
  epsilon: 0.03

# System will suggest if cheaper profile performs well
```

## Troubleshooting

### Budget Not Tracking

```bash
# Verify database
boba doctor

# Check pricing config
boba edit pricing

# Ensure usage is being recorded
boba stats --today
```

### Inaccurate Costs

```bash
# Check estimation levels
boba stats --by-estimate

# Add exact pricing
boba edit pricing

# Verify profile costs
boba ls --profiles --verbose
```

### Alerts Not Showing

```bash
# Check budget config
boba budget --status

# Verify alert thresholds
cat ~/.boba/profiles.yaml | grep -A 5 budget

# Check action queue
boba action
```

### Wrong Budget Applied

```bash
# Check priority order
# 1. Project budget (if in project directory)
# 2. Global budget

# Verify project config
cat .boba-project.yaml

# Check if in project directory
pwd
```

## API Reference

### CLI Commands

```bash
# View status
boba budget --status
boba budget --status --detailed
boba budget --status --project <name>

# Set budgets
boba budget --set daily <amount>
boba budget --set weekly <amount>
boba budget --set monthly <amount>
boba budget --set cap <amount>

# View projections
boba budget --project daily
boba budget --project weekly
boba budget --project monthly

# View actions
boba action
boba action --type budget
boba action apply <id>
```

### Configuration Schema

```yaml
budget:
  daily_usd: float        # Daily limit
  weekly_usd: float       # Weekly limit
  monthly_usd: float      # Monthly limit
  hard_cap: float         # Absolute maximum
  period_days: int        # Period for hard_cap (default: 30)
  alert_at_percent: int   # Warning threshold % (default: 75)
  critical_at_percent: int  # Critical threshold % (default: 90)
```

## Next Steps

- **[Analytics](/features/analytics)** - Analyze spending patterns
- **[Routing](/features/routing)** - Optimize with routing rules
- **[CLI Reference](/reference/cli)** - Budget command details
- **[Operations](/advanced/operations)** - Production budget management
