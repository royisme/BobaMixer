# Routing Cookbook

Common routing patterns and examples for BobaMixer.

## Basic Routing Concepts

Routing in BobaMixer allows you to automatically select the optimal profile based on context, task type, text patterns, project characteristics, and more.

### Route Evaluation Order

Rules are evaluated in **declaration order**. The first matching rule wins (short-circuit evaluation).

### Testing Routes

Always test your routing rules:
```bash
boba route test "Your test text here"
boba route test @path/to/file.txt
```

## Common Patterns

### 1. Context Size-Based Routing

Route based on input size:

```yaml
rules:
  - id: extra-large
    if: "ctx_chars > 100000"
    use: claude-opus
    explain: "Very large context requires Opus"
    
  - id: large
    if: "ctx_chars > 50000"
    use: claude-sonnet
    explain: "Large context, use Sonnet"
    
  - id: medium
    if: "ctx_chars > 10000"
    use: claude-haiku
    explain: "Medium context, Haiku is sufficient"
    
  - id: small
    if: "ctx_chars > 0"
    use: gpt-4o-mini
    explain: "Small context, use mini model"
```

### 2. Task Type Recognition

Route based on task keywords:

```yaml
rules:
  - id: code-generation
    if: "text.matches('write.*function|implement.*class|create.*API')"
    use: code-specialist
    explain: "Code generation task"
    
  - id: code-review
    if: "text.matches('review|analyze.*code|find.*bugs')"
    use: code-reviewer
    explain: "Code review task"
    
  - id: formatting
    if: "text.matches('format|prettier|eslint|style')"
    use: fast-formatter
    explain: "Simple formatting, use fast model"
    
  - id: documentation
    if: "text.matches('document|explain|write.*docs|README')"
    use: doc-writer
    explain: "Documentation task"
```

### 3. Project Type Routing

Route based on project characteristics:

```yaml
rules:
  - id: frontend-work
    if: "project_types.includes('react') || project_types.includes('vue')"
    use: frontend-specialist
    explain: "Frontend project"
    
  - id: backend-work
    if: "project_types.includes('go') || project_types.includes('rust')"
    use: backend-specialist
    explain: "Backend project"
    
  - id: ml-work
    if: "project_types.includes('python') && text.matches('tensor|model|train')"
    use: ml-specialist
    explain: "Machine learning task"
```

### 4. Time-Based Routing

Optimize costs during off-peak hours:

```yaml
rules:
  - id: night-mode
    if: "time_of_day == 'night'"
    use: cost-optimized
    explain: "Off-peak hours, using cheaper model"
    
  - id: business-hours
    if: "time_of_day == 'day'"
    use: high-performance
    explain: "Business hours, prioritizing speed"
```

### 5. Branch-Based Routing

Different models for different git branches:

```yaml
rules:
  - id: production-branch
    if: "branch.matches('main|master|prod')"
    use: high-accuracy
    explain: "Production branch, use most accurate model"
    
  - id: feature-branch
    if: "branch.matches('feature/.*')"
    use: balanced
    explain: "Feature development"
    
  - id: experiment-branch
    if: "branch.matches('exp/.*|test/.*')"
    use: fast-experimental
    explain: "Experimental work, speed over accuracy"
```

### 6. Multi-Condition Rules

Combine multiple conditions:

```yaml
rules:
  - id: complex-backend-task
    if: "ctx_chars > 20000 && project_types.includes('go') && text.matches('refactor|optimize')"
    use: senior-engineer
    explain: "Complex backend refactoring"
    
  - id: urgent-fix
    if: "branch.matches('hotfix/.*') && text.contains('urgent')"
    use: fastest-model
    explain: "Urgent hotfix, prioritize speed"
```

### 7. Fallback Strategy

Always provide fallback for robustness:

```yaml
rules:
  - id: specialized-task
    if: "text.matches('specific-pattern')"
    use: specialized-model
    fallback: general-purpose
    explain: "Try specialized model, fallback to general"
```

## Advanced Patterns

### Sub-Agent Triggers

Route to specialized sub-agents:

```yaml
sub_agents:
  - name: security-scanner
    profile: security-expert
    triggers: ["security", "vulnerability", "CVE"]
    conditions:
      min_ctx_chars: 1000
      
  - name: performance-optimizer
    profile: perf-specialist
    triggers: ["slow", "optimize", "performance"]
    conditions:
      time_of_day: ["09:00-17:00"]
```

### Cost-Aware Routing

Balance cost and capability:

```yaml
rules:
  - id: budget-conscious
    if: "ctx_chars < 5000 && !text.matches('critical|urgent')"
    use: mini-model
    explain: "Small, non-critical task, minimize cost"
    
  - id: quality-required
    if: "text.matches('production|customer-facing|critical')"
    use: premium-model
    explain: "Quality critical, use best model"
```

### Content Type Detection

Route based on content type:

```yaml
rules:
  - id: data-analysis
    if: "text.matches('analyze.*data|CSV|JSON|dataframe')"
    use: data-specialist
    explain: "Data analysis task"
    
  - id: image-description
    if: "text.contains('image') || text.contains('screenshot')"
    use: vision-model
    explain: "Image-related task"
    
  - id: text-only
    if: "ctx_chars > 0"
    use: text-model
    explain: "Pure text task"
```

## Testing Strategies

### 1. Create Test Scenarios

```bash
# Create test files
echo "Write a sorting algorithm" > test-cases/code-gen.txt
echo "Review this code for bugs" > test-cases/code-review.txt
echo "Format this code with prettier" > test-cases/format.txt

# Test each scenario
for f in test-cases/*.txt; do
  echo "Testing: $f"
  boba route test @$f
  echo "---"
done
```

### 2. Validate All Rules

```bash
# Test with various context sizes
boba route test "$(head -c 1000 < /dev/urandom | base64)"   # Small
boba route test "$(head -c 50000 < /dev/urandom | base64)"  # Large
boba route test "$(head -c 200000 < /dev/urandom | base64)" # XL
```

### 3. Time-Based Testing

```bash
# Simulate different times
# (requires modifying Context in code or waiting)
boba route test "test prompt"  # Check what time_of_day is selected
```

## Best Practices

1. **Order Matters**: Put more specific rules first
2. **Test Thoroughly**: Use `boba route test` extensively
3. **Provide Explanations**: Always include meaningful `explain` field
4. **Use Fallbacks**: Specify fallback profiles for critical paths
5. **Monitor Usage**: Track which rules are matching via usage logs
6. **Enable Exploration**: Let epsilon-greedy find optimal routes
7. **Review Regularly**: Analyze suggestions to refine rules

## Common Mistakes

❌ **Too Generic Rules**
```yaml
- if: "ctx_chars > 0"
  use: default
```

✅ **Specific Conditions**
```yaml
- if: "ctx_chars > 50000"
  use: high-capacity
```

❌ **Wrong Order**
```yaml
- if: "ctx_chars > 0"      # Matches everything!
  use: small-model
- if: "ctx_chars > 50000"  # Never reached
  use: large-model
```

✅ **Correct Order**
```yaml
- if: "ctx_chars > 50000"
  use: large-model
- if: "ctx_chars > 0"
  use: small-model
```

❌ **No Fallback**
```yaml
- if: "rare_condition"
  use: specialized
  # What if specialized fails?
```

✅ **With Fallback**
```yaml
- if: "rare_condition"
  use: specialized
  fallback: general-purpose
```

## Examples from Real Projects

### Full-Stack Development

```yaml
rules:
  - id: architecture-design
    if: "text.matches('architect|design.*system|scalability')"
    use: senior-architect
    
  - id: api-development
    if: "text.matches('API|endpoint|REST|GraphQL') && ctx_chars > 5000"
    use: backend-specialist
    
  - id: ui-components
    if: "text.matches('component|UI|interface') && project_types.includes('react')"
    use: frontend-specialist
    
  - id: quick-fixes
    if: "ctx_chars < 2000 && text.matches('fix|bug')"
    use: fast-model
```

### Data Science Workflow

```yaml
rules:
  - id: model-training
    if: "text.matches('train|model|neural|tensorflow')"
    use: ml-specialist
    
  - id: data-cleaning
    if: "text.matches('clean|preprocess|transform.*data')"
    use: data-engineer
    
  - id: visualization
    if: "text.matches('plot|visualize|chart|graph')"
    use: viz-specialist
```

## Integration with Budget

Combine routing with budget awareness:

```yaml
rules:
  - id: budget-alert-mode
    if: "ctx_chars < 10000"  # During budget alerts, prefer smaller contexts
    use: mini-model
    explain: "Budget mode: using economical model"
    
  - id: normal-operation
    if: "ctx_chars > 0"
    use: balanced-model
```

## Monitoring Routes

View routing decisions in reports:
```bash
boba report --format json | jq '.[] | select(.explore == true)'
```

Analyze rule effectiveness:
```bash
boba stats --by-profile --7d
```

---

For more information:
- [Configuration Guide](QUICK_REFERENCE.md)
- [Adapter Guide](ADAPTERS.md)
- [Operations Guide](OPERATIONS.md)
