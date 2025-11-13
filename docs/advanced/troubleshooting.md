# Troubleshooting

Common issues, solutions, and frequently asked questions for BobaMixer.

## Quick Diagnostics

When encountering issues, always start with:

```bash
boba doctor
```

This checks:
- Configuration file syntax
- File permissions
- Database connectivity
- API endpoint accessibility
- Profile validity

## Common Issues

### Installation Issues

#### Command Not Found

**Problem**: `bash: boba: command not found`

**Solutions**:

1. Check if installed:
   ```bash
   which boba
   ```

2. If using Go install, add GOPATH to PATH:
   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   # Add to ~/.bashrc or ~/.zshrc for persistence
   ```

3. If using manual install, verify location:
   ```bash
   ls -l /usr/local/bin/boba
   ```

4. Make executable:
   ```bash
   sudo chmod +x /usr/local/bin/boba
   ```

#### Permission Denied

**Problem**: `permission denied: /usr/local/bin/boba`

**Solutions**:

```bash
# Make executable
sudo chmod +x /usr/local/bin/boba

# Or reinstall with correct permissions
sudo mv boba /usr/local/bin/
sudo chmod +x /usr/local/bin/boba
```

#### macOS Security Warning

**Problem**: "boba cannot be opened because it is from an unidentified developer"

**Solutions**:

1. Open **System Preferences** → **Security & Privacy**
2. Click **Allow Anyway** next to the BobaMixer warning
3. Try running `boba` again

Or remove quarantine:
```bash
xattr -d com.apple.quarantine /usr/local/bin/boba
```

---

### Configuration Issues

#### Secrets File Permission Error

**Problem**: `secrets.yaml must have 0600 permissions`

**Solution**:

```bash
chmod 600 ~/.boba/secrets.yaml

# Verify
ls -l ~/.boba/secrets.yaml
# Should show: -rw-------
```

#### YAML Syntax Error

**Problem**: `error parsing config: yaml: line X: ...`

**Solutions**:

1. Validate YAML syntax:
   ```bash
   yamllint ~/.boba/profiles.yaml
   ```

2. Common YAML mistakes:
   - **Incorrect indentation** (use 2 spaces, not tabs)
   - **Missing quotes** around special characters
   - **Invalid characters** in keys or values

3. Example fixes:
   ```yaml
   # Wrong
   headers:
     x-api-key: secret://anthropic_key

   # Right
   headers:
     x-api-key: "secret://anthropic_key"
   ```

#### Secret Not Found

**Problem**: `secret not found: anthropic_key`

**Solutions**:

1. Check secret exists:
   ```bash
   grep "anthropic_key" ~/.boba/secrets.yaml
   ```

2. Verify reference format:
   ```yaml
   # Correct
   x-api-key: "secret://anthropic_key"

   # Wrong
   x-api-key: "secret://anthropic_key/"
   x-api-key: "secret:anthropic_key"
   ```

3. Add missing secret:
   ```bash
   boba edit secrets
   # Add: anthropic_key: sk-ant-your-key
   ```

#### Configuration Directory Not Found

**Problem**: `config directory not found: ~/.boba`

**Solution**:

```bash
# Create directory
mkdir -p ~/.boba/logs
chmod 700 ~/.boba

# Generate default configs
boba doctor
```

---

### Database Issues

#### Database Locked

**Problem**: `database is locked`

**Solutions**:

1. Find processes using database:
   ```bash
   lsof ~/.boba/usage.db
   ```

2. Kill if necessary:
   ```bash
   pkill -f boba
   ```

3. Remove stale locks:
   ```bash
   rm -f ~/.boba/usage.db-shm ~/.boba/usage.db-wal
   ```

4. Test:
   ```bash
   boba stats --today
   ```

#### Database Corruption

**Problem**: `database disk image is malformed`

**Solutions**:

1. Restore from backup:
   ```bash
   cp ~/.boba/usage.db.backup ~/.boba/usage.db
   ```

2. If no backup, try to recover:
   ```bash
   # Dump to SQL
   sqlite3 ~/.boba/usage.db .dump > dump.sql

   # Create new database
   mv ~/.boba/usage.db ~/.boba/usage.db.corrupt
   sqlite3 ~/.boba/usage.db < dump.sql

   # Verify
   sqlite3 ~/.boba/usage.db "PRAGMA integrity_check;"
   ```

3. If recovery fails, start fresh:
   ```bash
   mv ~/.boba/usage.db ~/.boba/usage.db.corrupt
   boba doctor  # Creates new database
   ```

#### Slow Database Queries

**Problem**: Stats commands are slow

**Solutions**:

1. Vacuum database:
   ```bash
   sqlite3 ~/.boba/usage.db "VACUUM;"
   ```

2. Analyze:
   ```bash
   sqlite3 ~/.boba/usage.db "ANALYZE;"
   ```

3. Enable WAL mode:
   ```bash
   sqlite3 ~/.boba/usage.db "PRAGMA journal_mode=WAL;"
   ```

4. Increase cache:
   ```bash
   sqlite3 ~/.boba/usage.db "PRAGMA cache_size=-10000;"  # 10MB
   ```

---

### API Issues

#### API Call Failed

**Problem**: `API call failed: connection refused`

**Solutions**:

1. Check internet connectivity:
   ```bash
   ping api.anthropic.com
   ```

2. Test endpoint manually:
   ```bash
   curl -v https://api.anthropic.com/v1/messages
   ```

3. Verify API key:
   ```bash
   # Check key exists
   grep "anthropic_key" ~/.boba/secrets.yaml

   # Test with curl
   curl -X POST https://api.anthropic.com/v1/messages \
     -H "x-api-key: YOUR-KEY" \
     -H "anthropic-version: 2023-06-01" \
     -H "content-type: application/json" \
     -d '{"model":"claude-3-5-sonnet-20241022","max_tokens":10,"messages":[{"role":"user","content":"test"}]}'
   ```

4. Check firewall/proxy:
   ```bash
   # If behind proxy
   export HTTP_PROXY=http://proxy.example.com:8080
   export HTTPS_PROXY=http://proxy.example.com:8080
   ```

#### Invalid API Key

**Problem**: `authentication failed: invalid API key`

**Solutions**:

1. Verify key format:
   - Anthropic: `sk-ant-api03-...`
   - OpenAI: `sk-proj-...` or `sk-...`
   - OpenRouter: `sk-or-v1-...`

2. Check for whitespace:
   ```bash
   # Should have no leading/trailing spaces
   grep "anthropic_key" ~/.boba/secrets.yaml | cat -A
   ```

3. Test key directly:
   ```bash
   # Get key
   KEY=$(grep "anthropic_key:" ~/.boba/secrets.yaml | cut -d: -f2 | tr -d ' ')

   # Test
   curl https://api.anthropic.com/v1/messages \
     -H "x-api-key: $KEY" \
     -H "anthropic-version: 2023-06-01" \
     -H "content-type: application/json" \
     -d '{"model":"claude-3-5-sonnet-20241022","max_tokens":10,"messages":[{"role":"user","content":"test"}]}'
   ```

4. Regenerate key if needed

#### Rate Limited

**Problem**: `rate limit exceeded`

**Solutions**:

1. Wait before retrying
2. Check provider's rate limits
3. Use different profile:
   ```bash
   boba use alternate-profile
   ```
4. Contact provider to increase limits

---

### Routing Issues

#### Rule Not Matching

**Problem**: Wrong profile selected

**Solutions**:

1. Test routing:
   ```bash
   boba route test "your text here"
   ```

2. Enable verbose mode:
   ```bash
   boba route test --verbose "your text here"
   ```

3. Check rule order (first match wins):
   ```yaml
   # Wrong order
   rules:
     - if: "ctx_chars > 0"  # Matches everything!
       use: cheap
     - if: "ctx_chars > 50000"  # Never reached
       use: expensive

   # Correct order
   rules:
     - if: "ctx_chars > 50000"
       use: expensive
     - if: "ctx_chars > 0"
       use: cheap
   ```

4. Validate routing config:
   ```bash
   boba route validate
   ```

#### Regex Not Working

**Problem**: `text.matches()` not matching expected text

**Solutions**:

1. Test regex:
   ```bash
   # Python
   python3 -c "import re; print(re.search(r'your.*pattern', 'your test text'))"

   # Or use online tester: regex101.com
   ```

2. Common regex mistakes:
   ```yaml
   # Wrong: escaped pipe
   if: "text.matches('format\\|prettier')"

   # Right: pipe is OR operator
   if: "text.matches('format|prettier')"

   # Wrong: unescaped special characters
   if: "text.matches('function()')"

   # Right: escape parentheses
   if: "text.matches('function\\(\\)')"
   ```

3. Use word boundaries:
   ```yaml
   # Wrong: matches "format" in "information"
   if: "text.contains('format')"

   # Right: matches whole word only
   if: "text.matches('\\bformat\\b')"
   ```

---

### Budget Issues

#### Budget Not Tracking

**Problem**: Budget status shows $0

**Solutions**:

1. Check database has data:
   ```bash
   sqlite3 ~/.boba/usage.db "SELECT COUNT(*) FROM usage_records;"
   ```

2. Verify cost calculations:
   ```bash
   boba stats --today
   ```

3. Check pricing config:
   ```bash
   boba edit pricing
   ```

4. View estimate accuracy:
   ```bash
   boba stats --by-estimate
   ```

#### Inaccurate Cost Estimates

**Problem**: Costs seem wrong

**Solutions**:

1. Check estimation level:
   ```bash
   boba stats --by-estimate
   ```

2. Add exact pricing:
   ```bash
   boba edit pricing
   # Add model pricing
   ```

3. Verify profile costs:
   ```bash
   boba ls --profiles --verbose
   ```

4. Compare with provider bills

#### Alerts Not Showing

**Problem**: No budget alerts despite high spending

**Solutions**:

1. Check budget configured:
   ```bash
   boba budget --status
   ```

2. Verify alert thresholds:
   ```bash
   cat ~/.boba/profiles.yaml | grep -A 5 budget
   ```

3. Check action queue:
   ```bash
   boba action
   ```

4. View in TUI:
   ```bash
   boba
   # Check notifications panel
   ```

---

### TUI Issues

#### Garbled Display

**Problem**: TUI looks broken or garbled

**Solutions**:

1. Check terminal type:
   ```bash
   echo $TERM
   # Should be xterm-256color or similar
   ```

2. Set terminal type:
   ```bash
   export TERM=xterm-256color
   boba
   ```

3. Disable colors if needed:
   ```bash
   export NO_COLOR=1
   boba
   ```

4. Try different terminal emulator

#### TUI Not Refreshing

**Problem**: Dashboard not updating

**Solutions**:

1. Check refresh rate:
   ```bash
   boba --refresh-rate 2  # 2 seconds
   ```

2. Restart TUI:
   - Press `q` to quit
   - Run `boba` again

3. Check for errors in logs:
   ```bash
   tail -f ~/.boba/logs/boba.log
   ```

---

### Performance Issues

#### Slow Commands

**Problem**: All boba commands are slow

**Solutions**:

1. Check database size:
   ```bash
   du -h ~/.boba/usage.db
   ```

2. Vacuum database:
   ```bash
   sqlite3 ~/.boba/usage.db "VACUUM;"
   ```

3. Cleanup old data:
   ```bash
   sqlite3 ~/.boba/usage.db "DELETE FROM usage_records WHERE ts < strftime('%s', 'now', '-90 days');"
   sqlite3 ~/.boba/usage.db "VACUUM;"
   ```

4. Optimize database:
   ```bash
   sqlite3 ~/.boba/usage.db "PRAGMA optimize;"
   sqlite3 ~/.boba/usage.db "ANALYZE;"
   ```

#### High Memory Usage

**Problem**: BobaMixer using too much memory

**Solutions**:

1. Check memory:
   ```bash
   ps aux | grep boba
   ```

2. Reduce TUI refresh rate:
   ```bash
   boba --refresh-rate 10  # 10 seconds
   ```

3. Limit query size:
   ```bash
   # Instead of --30d, use shorter periods
   boba stats --7d
   ```

---

## Frequently Asked Questions

### General

#### What is BobaMixer?

BobaMixer is a CLI tool for tracking, analyzing, and optimizing AI/LLM API usage and costs. It helps you understand spending patterns and make data-driven decisions about model selection.

#### Is BobaMixer free?

Yes! BobaMixer is open-source under the MIT license. You only pay for API usage from your providers.

#### Does BobaMixer intercept my API calls?

No. BobaMixer is a tracking tool. You explicitly invoke it when you want to track usage.

### Installation

#### Which installation method should I use?

- **Homebrew**: Easiest for macOS/Linux
- **Go install**: Good if you have Go installed
- **Binary**: Good for servers or without Go
- **Source**: For development or custom builds

#### Can I install BobaMixer on Windows?

Yes, via WSL (Windows Subsystem for Linux). Native Windows support is planned.

### Configuration

#### Where are configurations stored?

Default: `~/.boba/`

Override with: `export BOBA_HOME=/custom/path`

#### Can multiple users share configurations?

Yes! Use symlinks for shared configs, keep individual secrets:

```bash
ln -s /shared/boba/profiles.yaml ~/.boba/profiles.yaml
cp my-secrets.yaml ~/.boba/secrets.yaml
chmod 600 ~/.boba/secrets.yaml
```

#### How do I migrate from another tool?

Currently manual. Export data from other tool, import to BobaMixer database, or start fresh.

### Usage

#### How accurate are cost estimates?

- **Exact** (API response): 100%
- **Mapped** (pricing config): 95-99%
- **Heuristic** (character-based): 70-90%

Check with: `boba stats --by-estimate`

#### Can budgets block API calls?

No. BobaMixer uses "alert, don't interrupt" philosophy. You get warnings but work is never blocked.

#### How do I backup my data?

```bash
# Backup database
sqlite3 ~/.boba/usage.db ".backup /path/to/backup.db"

# Backup configs
cp -r ~/.boba ~/.boba.backup
```

### Troubleshooting

#### How do I reset everything?

```bash
# Backup first!
cp -r ~/.boba ~/.boba.backup

# Remove all data
rm -rf ~/.boba

# Reinitialize
boba doctor
```

#### How do I enable debug logging?

```bash
export BOBA_LOG_LEVEL=debug
boba doctor

# Check logs
tail -f ~/.boba/logs/boba.log
```

#### Database is too large

```bash
# Check size
du -h ~/.boba/usage.db

# Cleanup old records
sqlite3 ~/.boba/usage.db "DELETE FROM usage_records WHERE ts < strftime('%s', 'now', '-90 days');"
sqlite3 ~/.boba/usage.db "VACUUM;"
```

## Getting Help

### Self-Service Resources

1. **Run diagnostics**:
   ```bash
   boba doctor --verbose
   ```

2. **Check logs**:
   ```bash
   tail -f ~/.boba/logs/boba.log
   ```

3. **Review documentation**:
   - [Getting Started](/guide/getting-started)
   - [Configuration Guide](/guide/configuration)
   - [CLI Reference](/reference/cli)

### Community Support

1. **GitHub Discussions**: [Ask questions](https://github.com/royisme/BobaMixer/discussions)
2. **GitHub Issues**: [Report bugs](https://github.com/royisme/BobaMixer/issues)
3. **Documentation**: [Full docs](https://royisme.github.io/BobaMixer/)

### Reporting Bugs

When reporting issues, include:

1. **Version**:
   ```bash
   boba version
   ```

2. **Diagnostic output**:
   ```bash
   boba doctor --verbose > diagnostics.txt
   ```

3. **Logs** (remove sensitive data):
   ```bash
   tail -n 100 ~/.boba/logs/boba.log > recent-logs.txt
   ```

4. **Steps to reproduce**
5. **Expected vs actual behavior**
6. **Error messages**

### Feature Requests

Open a GitHub issue with:
- **Use case**: Why do you need this feature?
- **Proposed solution**: How should it work?
- **Alternatives**: What have you tried?

## Known Issues

### Issue: SQLite Version Too Old

**Problem**: Some Linux distributions ship with SQLite < 3.0

**Workaround**: Install newer SQLite

```bash
# Ubuntu/Debian
sudo apt-get install sqlite3 libsqlite3-dev

# Fedora/RHEL
sudo dnf install sqlite sqlite-devel
```

### Issue: Git Hooks Not Triggering

**Problem**: Hooks installed but not running

**Workaround**: Verify git hooks path

```bash
git config core.hooksPath
# Should output: .git/hooks or custom path

# Reinstall
boba hooks install --force
```

## Next Steps

- **[Operations Guide](/advanced/operations)** - Production best practices
- **[CLI Reference](/reference/cli)** - Command documentation
- **[Configuration](/guide/configuration)** - Setup guide
- **[GitHub Issues](https://github.com/royisme/BobaMixer/issues)** - Report bugs
