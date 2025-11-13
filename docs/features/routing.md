# Intelligent Routing

BobaMixer's intelligent routing system automatically selects the optimal AI profile based on context, task type, project characteristics, and usage patterns.

## Overview

Routing rules allow you to:

- **Optimize costs** by using cheaper models for simple tasks
- **Improve quality** by using better models for complex work
- **Save time** by using faster models when appropriate
- **Explore alternatives** with epsilon-greedy optimization
- **Adapt to context** based on project, branch, time, and more

## How Routing Works

1. **Rule Evaluation**: Rules are evaluated in order (first match wins)
2. **Condition Checking**: Each rule's condition is evaluated against current context
3. **Profile Selection**: When a rule matches, its profile is selected
4. **Fallback Handling**: If the selected profile fails, fallback is used
5. **Exploration**: Occasionally (ε% of time), alternative profiles are tested

## Configuration

Routes are defined in `~/.boba/routes.yaml`:

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

## Routing DSL

### Available Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `ctx_chars` | int | Input character count | `ctx_chars > 50000` |
| `text` | string | Input text content | `text.contains('code')` |
| `project_types` | array | Project types | `project_types.includes('go')` |
| `branch` | string | Git branch name | `branch.matches('main')` |
| `time_of_day` | string | Time period | `time_of_day == 'night'` |
| `intent` | string | Detected intent | `intent == 'format'` |

### Time of Day Values

- `morning` - 6 AM to 12 PM
- `day` - 12 PM to 6 PM
- `evening` - 6 PM to 10 PM
- `night` - 10 PM to 6 AM

### Available Functions

| Function | Description | Example |
|----------|-------------|---------|
| `text.matches(pattern)` | Regex match (case-insensitive) | `text.matches('\\bformat\\b')` |
| `text.contains(str)` | Substring search | `text.contains('review')` |
| `array.includes(item)` | Array membership | `project_types.includes('react')` |

### Operators

- Comparison: `>`, `<`, `>=`, `<=`, `==`, `!=`
- Logical: `&&` (and), `||` (or), `!` (not)
- Parentheses: `(...)` for grouping

## Common Routing Patterns

### 1. Context Size-Based Routing

Route based on input size:

```yaml
rules:
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

  - id: medium-context
    if: "ctx_chars > 10000"
    use: claude-haiku
    explain: "Medium context, balanced model"

  - id: small-context
    if: "ctx_chars > 0"
    use: gpt-4o-mini
    explain: "Small context, economical model"
```

**Why this works**: Larger contexts need models with higher capacity. This pattern optimizes costs while ensuring quality.

### 2. Task Type Recognition

Route based on task keywords:

```yaml
rules:
  - id: code-generation
    if: "text.matches('write.*function|implement|create.*class|build.*API')"
    use: code-specialist
    explain: "Code generation task"

  - id: code-review
    if: "text.matches('review|analyze.*code|find.*bug|refactor')"
    use: code-reviewer
    explain: "Code review task"

  - id: formatting
    if: "text.matches('format|prettier|eslint|lint|style')"
    use: fast-formatter
    explain: "Simple formatting, use fast model"

  - id: documentation
    if: "text.matches('document|explain|write.*docs|README|comment')"
    use: doc-writer
    explain: "Documentation task"

  - id: testing
    if: "text.matches('test|spec|unit test|integration test')"
    use: test-specialist
    explain: "Testing task"
```

**Why this works**: Different tasks have different requirements. Code generation needs creativity, while formatting just needs accuracy.

### 3. Project Type Routing

Route based on project characteristics:

```yaml
rules:
  - id: frontend-react
    if: "project_types.includes('react') || project_types.includes('vue')"
    use: frontend-specialist
    explain: "Frontend framework project"

  - id: backend-go
    if: "project_types.includes('go')"
    use: backend-go-specialist
    explain: "Go backend project"

  - id: backend-rust
    if: "project_types.includes('rust')"
    use: backend-rust-specialist
    explain: "Rust backend project"

  - id: ml-python
    if: "project_types.includes('python') && text.matches('tensor|model|train|dataset')"
    use: ml-specialist
    explain: "Machine learning task"

  - id: data-analysis
    if: "project_types.includes('python') && text.matches('analyze|dataframe|csv|pandas')"
    use: data-specialist
    explain: "Data analysis task"
```

**Why this works**: Specialized models for specific tech stacks can provide better, more accurate responses.

### 4. Branch-Based Routing

Different profiles for different branches:

```yaml
rules:
  - id: production-branch
    if: "branch.matches('main|master|prod')"
    use: high-accuracy
    fallback: medium-accuracy
    explain: "Production branch, use most accurate model"

  - id: staging-branch
    if: "branch.matches('staging|stage')"
    use: medium-accuracy
    explain: "Staging branch, balanced model"

  - id: feature-branch
    if: "branch.matches('feature/.*|feat/.*')"
    use: balanced-model
    explain: "Feature development"

  - id: hotfix-branch
    if: "branch.matches('hotfix/.*|fix/.*')"
    use: fastest-model
    explain: "Urgent fix, prioritize speed"

  - id: experiment-branch
    if: "branch.matches('exp/.*|test/.*|playground/.*')"
    use: experimental-model
    explain: "Experimental work"
```

**Why this works**: Production code needs more careful review, while experiments can use faster/cheaper models.

### 5. Time-Based Routing

Optimize costs during off-peak hours:

```yaml
rules:
  - id: night-mode
    if: "time_of_day == 'night'"
    use: cost-optimized
    explain: "Off-peak hours, use cheaper model for cost savings"

  - id: business-hours
    if: "time_of_day == 'day' || time_of_day == 'morning'"
    use: high-performance
    explain: "Business hours, prioritize speed and quality"

  - id: evening-balanced
    if: "time_of_day == 'evening'"
    use: balanced-model
    explain: "Evening work, balanced approach"
```

**Why this works**: Your budget can stretch further by using cheaper models during non-critical hours.

### 6. Multi-Condition Rules

Combine multiple conditions:

```yaml
rules:
  - id: complex-backend-refactor
    if: "ctx_chars > 20000 && project_types.includes('go') && text.matches('refactor|optimize|improve')"
    use: senior-engineer
    explain: "Complex backend refactoring needs expert model"

  - id: urgent-production-fix
    if: "branch.matches('hotfix/.*') && text.contains('urgent')"
    use: fastest-accurate
    fallback: fastest-model
    explain: "Urgent production fix, speed critical"

  - id: large-frontend-feature
    if: "ctx_chars > 30000 && project_types.includes('react') && text.matches('feature|component')"
    use: frontend-expert
    explain: "Large frontend feature needs specialized model"

  - id: simple-night-task
    if: "ctx_chars < 5000 && time_of_day == 'night'"
    use: mini-model
    explain: "Simple task at night, minimize cost"
```

**Why this works**: Combining conditions allows precise targeting of specific scenarios.

### 7. Cost-Aware Routing

Balance cost and capability:

```yaml
rules:
  - id: critical-quality
    if: "text.matches('production|customer-facing|critical|important')"
    use: premium-model
    fallback: good-model
    explain: "Quality critical work, use best model"

  - id: non-critical-small
    if: "ctx_chars < 5000 && !text.matches('critical|important|urgent')"
    use: mini-model
    explain: "Small, non-critical task, minimize cost"

  - id: bulk-processing
    if: "text.contains('batch') || text.contains('multiple files')"
    use: economical-model
    explain: "Bulk processing, optimize for cost"

  - id: exploratory-work
    if: "text.matches('explore|experiment|try|investigate')"
    use: cheap-model
    explain: "Exploratory work, cost over quality"
```

**Why this works**: Not all tasks need the best model. Save budget for what matters.

### 8. Content Type Detection

Route based on content type:

```yaml
rules:
  - id: data-analysis
    if: "text.matches('analyze.*data|CSV|JSON|dataframe|SQL')"
    use: data-specialist
    explain: "Data analysis task"

  - id: image-work
    if: "text.contains('image') || text.contains('screenshot') || text.contains('visual')"
    use: vision-model
    explain: "Image-related task needs vision capability"

  - id: pure-text
    if: "ctx_chars > 0"
    use: text-model
    explain: "Pure text task"
```

**Why this works**: Different content types need different model capabilities.

## Advanced Patterns

### Intent-Based Routing

```yaml
rules:
  - id: format-intent
    if: "intent == 'format'"
    use: fast-formatter
    explain: "Formatting intent detected"

  - id: review-intent
    if: "intent == 'review'"
    use: code-reviewer
    explain: "Code review intent"

  - id: generate-intent
    if: "intent == 'generate'"
    use: code-generator
    explain: "Code generation intent"
```

**Note**: Intent detection requires configuration or integration with external intent classifier.

### Cascading Rules

```yaml
rules:
  # Highest priority: Production critical
  - id: prod-critical
    if: "branch.matches('main') && text.contains('critical')"
    use: best-model
    explain: "Production critical work"

  # Medium priority: Production non-critical
  - id: prod-normal
    if: "branch.matches('main')"
    use: good-model
    explain: "Production work"

  # Lower priority: Other branches
  - id: dev-work
    if: "ctx_chars > 10000"
    use: medium-model
    explain: "Development work"

  # Fallback: Everything else
  - id: default
    if: "ctx_chars > 0"
    use: economical-model
    explain: "Default for small tasks"
```

**Why this works**: Most specific rules first, gradually broadening to catch everything.

### Dynamic Thresholds

```yaml
rules:
  # Adjust thresholds based on project type
  - id: large-go-project
    if: "project_types.includes('go') && ctx_chars > 30000"
    use: go-specialist
    explain: "Large Go codebase"

  - id: large-python-project
    if: "project_types.includes('python') && ctx_chars > 50000"
    use: python-specialist
    explain: "Large Python codebase (higher threshold for Python)"

  - id: large-js-project
    if: "project_types.includes('javascript') && ctx_chars > 40000"
    use: js-specialist
    explain: "Large JavaScript codebase"
```

**Why this works**: Different languages have different verbosity, so adjust thresholds accordingly.

## Epsilon-Greedy Exploration

BobaMixer can automatically discover better profiles using exploration.

### How It Works

1. **Exploitation** (97%): Use the best known profile based on rules
2. **Exploration** (3%): Try alternative profiles to discover better options
3. **Learning**: Track performance (cost, latency, quality)
4. **Suggestions**: Recommend better profiles when found

### Configuration

```yaml
exploration:
  enabled: true
  epsilon: 0.03          # 3% exploration rate
  min_samples: 10        # Need 10+ samples before exploring
  cooldown_hours: 24     # Wait 24h between re-testing profiles
  exclude_profiles:      # Never explore these
    - critical-only
    - production-only
```

### Benefits

- **Automatic optimization**: Finds cost-performance sweet spots
- **Adaptation**: Adjusts to changing usage patterns
- **Discovery**: Reveals unexpected good matches
- **No manual work**: System learns over time

### Viewing Exploration Results

```bash
# See exploration events
boba report --format json | jq '.[] | select(.explore == true)'

# Review suggestions
boba action

# Apply a suggestion
boba action apply <suggestion-id>
```

## Testing Routes

Always test routing rules before deploying.

### Test with Text

```bash
# Test inline text
boba route test "Write a function to sort an array"

# Shows: which rule matched, which profile selected, why
```

### Test with File

```bash
# Test file content
boba route test @path/to/prompt.txt

# Test multiple files
for file in test-cases/*.txt; do
  echo "Testing: $file"
  boba route test @"$file"
done
```

### Verbose Testing

```bash
# See detailed evaluation
boba route test --verbose "Your prompt here"

# Shows:
# - All rules evaluated
# - Which conditions matched/failed
# - Final selection reasoning
```

### List All Rules

```bash
# View configured rules
boba route list

# Shows rules in evaluation order
```

## Best Practices

### 1. Order Matters

Rules are evaluated top-to-bottom. First match wins.

**Wrong Order** (generic rule first):
```yaml
rules:
  - if: "ctx_chars > 0"      # Matches everything!
    use: cheap-model

  - if: "ctx_chars > 50000"  # Never reached
    use: expensive-model
```

**Correct Order** (specific rules first):
```yaml
rules:
  - if: "ctx_chars > 50000"
    use: expensive-model

  - if: "ctx_chars > 0"
    use: cheap-model
```

### 2. Always Provide Explanations

```yaml
# Good: Clear explanation
- id: large-context
  if: "ctx_chars > 50000"
  use: high-capacity
  explain: "Large context requires high-capacity model for quality"

# Bad: No explanation
- id: rule1
  if: "ctx_chars > 50000"
  use: profile1
```

Explanations help with:
- Debugging routing issues
- Understanding logs
- Team collaboration

### 3. Use Fallbacks for Critical Paths

```yaml
# Production: always have fallback
- id: production
  if: "branch.matches('main')"
  use: best-model
  fallback: good-model
  explain: "Production with fallback"

# Experimental: fallback optional
- id: experiment
  if: "branch.matches('exp')"
  use: experimental-model
  explain: "Experimental work"
```

### 4. Test Thoroughly

```bash
# Create test suite
mkdir route-tests

# Add test cases
echo "Write a function" > route-tests/code-gen.txt
echo "Review this code" > route-tests/code-review.txt
echo "Format with prettier" > route-tests/format.txt

# Test all cases
for f in route-tests/*.txt; do
  boba route test @"$f"
done
```

### 5. Monitor Route Performance

```bash
# Stats by profile
boba stats --by-profile --7d

# Identify which routes are used most
boba report --format json | jq 'group_by(.profile) | map({profile: .[0].profile, count: length})'
```

### 6. Enable Exploration Gradually

```yaml
# Start conservative
exploration:
  enabled: true
  epsilon: 0.01  # 1% exploration

# Increase after confidence
exploration:
  enabled: true
  epsilon: 0.03  # 3% exploration
```

### 7. Document Your Strategy

```yaml
# Add comments in routes.yaml
rules:
  # === PRODUCTION RULES ===
  # These handle production branch work
  # Always use high-quality models with fallbacks

  - id: prod-critical
    # ...

  # === DEVELOPMENT RULES ===
  # These handle feature branch work
  # Optimize for cost while maintaining quality

  - id: dev-work
    # ...
```

## Common Mistakes

### Mistake 1: Overly Generic Rules

**Problem**:
```yaml
- if: "text.contains('code')"
  use: code-model
```

**Solution**:
```yaml
- if: "text.matches('write.*code|implement|create.*function')"
  use: code-generation-model

- if: "text.matches('review.*code|analyze|refactor')"
  use: code-review-model
```

### Mistake 2: No Default Rule

**Problem**: No rule matches, system has no fallback

**Solution**: Always have a catch-all at the end
```yaml
- id: default
  if: "ctx_chars > 0"
  use: balanced-model
  explain: "Default for unmatched cases"
```

### Mistake 3: Regex Errors

**Problem**:
```yaml
# Wrong: unescaped special chars
if: "text.matches('format|prettier')"  # Correct
if: "text.matches('format\\|prettier')"  # Wrong (escaped |)
```

**Solution**: Only escape regex special chars like `\b`, `\d`, `\w`, etc.

### Mistake 4: Circular Fallbacks

**Problem**:
```yaml
profile-a:
  # ...
  fallback: profile-b

profile-b:
  # ...
  fallback: profile-a
```

**Solution**: Use linear fallback chains
```yaml
profile-a:
  fallback: profile-b

profile-b:
  fallback: profile-c

profile-c:
  # No fallback (or different one)
```

## Debugging Routes

### Enable Verbose Logging

```bash
export BOBA_LOG_LEVEL=debug
boba route test "Your test"
```

### Check Logs

```bash
tail -f ~/.boba/logs/boba.log
```

### Verify Rule Syntax

```bash
# Check YAML syntax
yamllint ~/.boba/routes.yaml

# Validate with boba
boba doctor
```

### Test Edge Cases

```bash
# Empty input
boba route test ""

# Very large input
boba route test @large-file.txt

# Special characters
boba route test "Code with 'quotes' and \"escapes\""
```

## Integration with Other Features

### With Budgets

```yaml
# When near budget, use cheaper models
- id: budget-conscious
  if: "ctx_chars < 10000"
  use: economical-model
  explain: "Small task, conserve budget"
```

Check budget status: `boba budget --status`

### With Analytics

```bash
# Analyze routing effectiveness
boba stats --by-profile --30d

# Export for analysis
boba report --format csv --output routing-analysis.csv
```

### With Git Hooks

Install git hooks to use routing automatically:
```bash
cd your-project
boba hooks install
```

## Next Steps

- **[Budgets](/features/budgets)** - Set up budget management
- **[Analytics](/features/analytics)** - Analyze routing patterns
- **[Configuration](/guide/configuration)** - Advanced routing config
- **[CLI Reference](/reference/cli)** - Route testing commands

## Examples Repository

For more examples, see:
- [Routing Cookbook Examples](https://github.com/royisme/BobaMixer/tree/main/examples/routing)
- [Common Patterns](https://github.com/royisme/BobaMixer/tree/main/examples/patterns)
