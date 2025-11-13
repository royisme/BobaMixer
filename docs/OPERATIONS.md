# Operations Guide

Best practices for running, maintaining, and troubleshooting BobaMixer in production.

## Installation & Setup

### System Requirements
- Go 1.22+ (if building from source)
- SQLite 3
- 50MB disk space (plus growth for usage database)
- Git (for hooks integration)

### Production Installation

**Using Homebrew (Recommended for macOS/Linux):**
```bash
brew tap royisme/tap
brew install bobamixer
```

**Manual Installation:**
```bash
# Download release
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_$(uname -s)_$(uname -m).tar.gz

# Extract
tar -xzf bobamixer_*.tar.gz

# Move to PATH
sudo mv boba /usr/local/bin/

# Verify
boba doctor
```

### Initial Configuration

1. **Create home directory:**
```bash
mkdir -p ~/.boba/{logs,git-hooks}
chmod 700 ~/.boba
```

2. **Generate default configs:**
```bash
cp /usr/local/share/bobamixer/examples/* ~/.boba/
```

3. **Set secrets permissions:**
```bash
chmod 600 ~/.boba/secrets.yaml
```

4. **Validate setup:**
```bash
boba doctor
```

## Database Management

### Database Location
```
~/.boba/usage.db
```

### Backup Strategy

**Daily Backup (Cron):**
```bash
# Add to crontab: crontab -e
0 2 * * * /usr/local/bin/backup-boba.sh

# /usr/local/bin/backup-boba.sh
#!/bin/bash
BACKUP_DIR=~/backups/boba
mkdir -p $BACKUP_DIR
sqlite3 ~/.boba/usage.db ".backup $BACKUP_DIR/usage-$(date +%Y%m%d).db"
# Keep last 30 days
find $BACKUP_DIR -name "usage-*.db" -mtime +30 -delete
```

**Manual Backup:**
```bash
# Hot backup (safe during use)
sqlite3 ~/.boba/usage.db ".backup /tmp/usage-backup.db"

# Or simple copy (only when boba is not running)
cp ~/.boba/usage.db ~/.boba/usage.db.backup
```

### Restore from Backup
```bash
# Stop all boba processes first
pkill -f boba

# Restore
cp /path/to/backup.db ~/.boba/usage.db

# Verify integrity
sqlite3 ~/.boba/usage.db "PRAGMA integrity_check;"
```

### Database Maintenance

**Vacuum (reclaim space):**
```bash
sqlite3 ~/.boba/usage.db "VACUUM;"
```

**Check integrity:**
```bash
sqlite3 ~/.boba/usage.db "PRAGMA integrity_check;"
```

**Analyze (optimize queries):**
```bash
sqlite3 ~/.boba/usage.db "ANALYZE;"
```

**View size:**
```bash
du -h ~/.boba/usage.db
```

## Cleanup & Purging

### Purge Old Records

**Delete records older than 90 days:**
```bash
sqlite3 ~/.boba/usage.db "DELETE FROM usage_records WHERE ts < strftime('%s', 'now', '-90 days');"
sqlite3 ~/.boba/usage.db "VACUUM;"
```

**Archive before deletion:**
```bash
# Export to CSV
sqlite3 ~/.boba/usage.db <<EOF
.headers on
.mode csv
.output ~/.boba/archive-$(date +%Y%m%d).csv
SELECT * FROM usage_records WHERE ts < strftime('%s', 'now', '-90 days');
.quit
