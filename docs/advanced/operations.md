# Operations Guide

Best practices for running, maintaining, and troubleshooting BobaMixer in production environments.

## Production Installation

### System Requirements

- **OS**: Linux (recommended), macOS, Windows (WSL)
- **Go**: 1.22+ (if building from source)
- **SQLite**: 3.x
- **Disk Space**: 50MB + database growth (estimate ~1KB per API call)
- **Memory**: 50MB typical, 200MB peak
- **Git**: Optional, for hooks integration

### Installation Methods

#### Using Homebrew (Recommended)

```bash
brew tap royisme/tap
brew install bobamixer

# Verify installation
boba version
boba doctor
```

#### Manual Installation

```bash
# Download latest release
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_Linux_x86_64.tar.gz

# Extract
tar -xzf bobamixer_Linux_x86_64.tar.gz

# Install
sudo mv boba /usr/local/bin/
sudo chmod +x /usr/local/bin/boba

# Verify
boba version
```

#### From Source

```bash
git clone https://github.com/royisme/BobaMixer.git
cd BobaMixer
make build
sudo make install

# Or
go install github.com/royisme/bobamixer/cmd/boba@latest
```

### Initial Setup

```bash
# Create configuration directory
mkdir -p ~/.boba/logs
chmod 700 ~/.boba

# Generate default configs
boba doctor

# Secure secrets file
chmod 600 ~/.boba/secrets.yaml

# Verify setup
boba doctor
```

## Configuration Management

### Directory Structure

```
~/.boba/
├── profiles.yaml         # Profile definitions
├── routes.yaml          # Routing rules
├── pricing.yaml         # Model pricing
├── secrets.yaml         # API keys (0600)
├── usage.db            # SQLite database
├── logs/               # Application logs
│   └── boba.log
├── git-hooks/          # Git hooks
└── pricing.cache.json  # Cached pricing (auto-generated)
```

### Version Control

Track configurations (except secrets):

```bash
cd ~/.boba
git init
git add profiles.yaml routes.yaml pricing.yaml
cat > .gitignore << EOF
secrets.yaml
*.db
*.db-*
logs/
pricing.cache.json
EOF
git commit -m "Initial BobaMixer configuration"

# Push to private repo
git remote add origin git@github.com:youruser/boba-config.git
git push -u origin main
```

### Shared Configuration

For team environments:

```bash
# Shared configs in team repo
git clone git@github.com:team/boba-shared-config.git ~/boba-shared

# Symlink shared configs
ln -sf ~/boba-shared/profiles.yaml ~/.boba/profiles.yaml
ln -sf ~/boba-shared/routes.yaml ~/.boba/routes.yaml
ln -sf ~/boba-shared/pricing.yaml ~/.boba/pricing.yaml

# Keep individual secrets
cp ~/.boba/secrets.yaml.example ~/.boba/secrets.yaml
chmod 600 ~/.boba/secrets.yaml
# Edit with your personal API keys
```

## Database Management

### Database Location

```
~/.boba/usage.db
```

### Backup Strategy

#### Automated Daily Backup

```bash
# Create backup script
cat > /usr/local/bin/backup-boba.sh << 'EOF'
#!/bin/bash
BACKUP_DIR=~/backups/boba
mkdir -p $BACKUP_DIR

# Hot backup (safe during use)
sqlite3 ~/.boba/usage.db ".backup $BACKUP_DIR/usage-$(date +%Y%m%d).db"

# Compress old backups
find $BACKUP_DIR -name "usage-*.db" -mtime +7 -exec gzip {} \;

# Delete backups older than 90 days
find $BACKUP_DIR -name "usage-*.db.gz" -mtime +90 -delete

echo "Backup completed: $BACKUP_DIR/usage-$(date +%Y%m%d).db"
EOF

chmod +x /usr/local/bin/backup-boba.sh

# Schedule daily backup (add to crontab)
crontab -e
# Add line:
0 2 * * * /usr/local/bin/backup-boba.sh
```

#### Manual Backup

```bash
# Hot backup (safe during use)
sqlite3 ~/.boba/usage.db ".backup /tmp/usage-backup.db"

# Or simple copy (only when boba is not running)
pkill -f boba
cp ~/.boba/usage.db ~/.boba/usage.db.backup
```

#### Restore from Backup

```bash
# Stop all boba processes
pkill -f boba

# Restore
cp /path/to/backup.db ~/.boba/usage.db

# Verify integrity
sqlite3 ~/.boba/usage.db "PRAGMA integrity_check;"

# Test
boba stats --today
```

### Database Maintenance

#### Weekly Maintenance Script

```bash
cat > /usr/local/bin/boba-maintenance.sh << 'EOF'
#!/bin/bash
DB=~/.boba/usage.db

echo "BobaMixer Database Maintenance"
echo "==============================="

# Check integrity
echo "Checking integrity..."
if ! sqlite3 $DB "PRAGMA integrity_check;" | grep -q "ok"; then
  echo "ERROR: Database integrity check failed!"
  exit 1
fi
echo "✓ Integrity OK"

# Vacuum
echo "Vacuuming database..."
sqlite3 $DB "VACUUM;"
echo "✓ Vacuum complete"

# Analyze
echo "Analyzing database..."
sqlite3 $DB "ANALYZE;"
echo "✓ Analyze complete"

# Show size
echo "Database size: $(du -h $DB | cut -f1)"

# Show record count
echo "Total records: $(sqlite3 $DB 'SELECT COUNT(*) FROM usage_records;')"

echo "Maintenance complete!"
EOF

chmod +x /usr/local/bin/boba-maintenance.sh

# Schedule weekly (add to crontab)
crontab -e
# Add line:
0 3 * * 0 /usr/local/bin/boba-maintenance.sh
```

#### Vacuum Database

```bash
sqlite3 ~/.boba/usage.db "VACUUM;"
```

#### Check Integrity

```bash
sqlite3 ~/.boba/usage.db "PRAGMA integrity_check;"
```

#### Analyze for Query Optimization

```bash
sqlite3 ~/.boba/usage.db "ANALYZE;"
```

#### View Database Size

```bash
du -h ~/.boba/usage.db
```

### Data Cleanup

#### Purge Old Records

```bash
# Delete records older than 90 days
sqlite3 ~/.boba/usage.db "DELETE FROM usage_records WHERE ts < strftime('%s', 'now', '-90 days');"

# Reclaim space
sqlite3 ~/.boba/usage.db "VACUUM;"
```

#### Archive Before Purging

```bash
# Export to CSV
sqlite3 -header -csv ~/.boba/usage.db \
  "SELECT * FROM usage_records WHERE ts < strftime('%s', 'now', '-90 days');" \
  > ~/.boba/archive-$(date +%Y%m%d).csv

# Compress archive
gzip ~/.boba/archive-$(date +%Y%m%d).csv

# Then purge
sqlite3 ~/.boba/usage.db "DELETE FROM usage_records WHERE ts < strftime('%s', 'now', '-90 days');"
sqlite3 ~/.boba/usage.db "VACUUM;"
```

#### Automated Cleanup Script

```bash
cat > /usr/local/bin/boba-cleanup.sh << 'EOF'
#!/bin/bash
DB=~/.boba/usage.db
ARCHIVE_DIR=~/.boba/archives
DAYS_TO_KEEP=90

mkdir -p $ARCHIVE_DIR

# Archive old records
echo "Archiving records older than $DAYS_TO_KEEP days..."
sqlite3 -header -csv $DB \
  "SELECT * FROM usage_records WHERE ts < strftime('%s', 'now', '-$DAYS_TO_KEEP days');" \
  > $ARCHIVE_DIR/archive-$(date +%Y%m%d).csv

if [ -s $ARCHIVE_DIR/archive-$(date +%Y%m%d).csv ]; then
  # Compress archive
  gzip $ARCHIVE_DIR/archive-$(date +%Y%m%d).csv

  # Delete from database
  echo "Purging old records..."
  sqlite3 $DB "DELETE FROM usage_records WHERE ts < strftime('%s', 'now', '-$DAYS_TO_KEEP days');"

  # Vacuum
  echo "Vacuuming database..."
  sqlite3 $DB "VACUUM;"

  echo "Cleanup complete. Archived to: $ARCHIVE_DIR/archive-$(date +%Y%m%d).csv.gz"
else
  echo "No records to archive."
  rm $ARCHIVE_DIR/archive-$(date +%Y%m%d).csv
fi
EOF

chmod +x /usr/local/bin/boba-cleanup.sh

# Schedule monthly (add to crontab)
crontab -e
# Add line:
0 4 1 * * /usr/local/bin/boba-cleanup.sh
```

## Logging

### Log Location

```
~/.boba/logs/boba.log
```

### Log Levels

Set via environment variable:

```bash
export BOBA_LOG_LEVEL=debug  # trace|debug|info|warn|error
```

### Log Rotation

#### Using logrotate (Linux)

```bash
# Create logrotate config
sudo cat > /etc/logrotate.d/bobamixer << EOF
/home/*/.boba/logs/boba.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 $USER $USER
}
EOF
```

#### Manual Script

```bash
cat > /usr/local/bin/rotate-boba-logs.sh << 'EOF'
#!/bin/bash
LOG_DIR=~/.boba/logs
LOG_FILE=$LOG_DIR/boba.log

if [ -f "$LOG_FILE" ]; then
  # Rotate
  mv $LOG_FILE $LOG_FILE.$(date +%Y%m%d)

  # Compress old logs
  find $LOG_DIR -name "boba.log.*" ! -name "*.gz" -mtime +1 -exec gzip {} \;

  # Delete logs older than 30 days
  find $LOG_DIR -name "boba.log.*.gz" -mtime +30 -delete

  # Create new log file
  touch $LOG_FILE

  echo "Log rotation complete"
fi
EOF

chmod +x /usr/local/bin/rotate-boba-logs.sh

# Schedule daily
crontab -e
# Add:
0 0 * * * /usr/local/bin/rotate-boba-logs.sh
```

### Viewing Logs

```bash
# Tail logs
tail -f ~/.boba/logs/boba.log

# Last 100 lines
tail -n 100 ~/.boba/logs/boba.log

# Search for errors
grep -i error ~/.boba/logs/boba.log

# View with timestamps
cat ~/.boba/logs/boba.log | grep "$(date +%Y-%m-%d)"
```

## Monitoring

### Health Checks

```bash
# Basic health check
boba doctor

# Detailed check
boba doctor --verbose

# Exit code check (for monitoring systems)
if boba doctor > /dev/null 2>&1; then
  echo "OK"
else
  echo "CRITICAL"
  exit 2
fi
```

### Usage Monitoring Script

```bash
cat > /usr/local/bin/boba-monitor.sh << 'EOF'
#!/bin/bash

# Thresholds
DAILY_BUDGET=50.00
WARN_THRESHOLD=75  # percent

# Get today's spending
CURRENT=$(boba stats --today --format json | jq '.total_cost_usd')

# Calculate percentage
PERCENT=$(echo "scale=2; $CURRENT / $DAILY_BUDGET * 100" | bc)

# Check threshold
if (( $(echo "$PERCENT > $WARN_THRESHOLD" | bc -l) )); then
  echo "WARNING: Daily spending at ${PERCENT}% ($CURRENT / $DAILY_BUDGET)"
  # Send alert (example: email)
  echo "BobaMixer spending alert: ${PERCENT}% of daily budget" | \
    mail -s "BobaMixer Budget Alert" admin@example.com
fi
EOF

chmod +x /usr/local/bin/boba-monitor.sh

# Run every hour
crontab -e
# Add:
0 * * * * /usr/local/bin/boba-monitor.sh
```

### Integration with Monitoring Systems

#### Prometheus Metrics Export

```bash
# Export metrics to file
cat > /usr/local/bin/boba-export-metrics.sh << 'EOF'
#!/bin/bash
METRICS_FILE=/var/lib/prometheus/node-exporter/bobamixer.prom

cat > $METRICS_FILE << METRICS
# HELP bobamixer_daily_cost Daily cost in USD
# TYPE bobamixer_daily_cost gauge
bobamixer_daily_cost $(boba stats --today --format json | jq '.total_cost_usd')

# HELP bobamixer_daily_requests Daily request count
# TYPE bobamixer_daily_requests counter
bobamixer_daily_requests $(boba stats --today --format json | jq '.total_requests')

# HELP bobamixer_daily_tokens Daily token count
# TYPE bobamixer_daily_tokens counter
bobamixer_daily_tokens $(boba stats --today --format json | jq '.total_tokens')
METRICS

EOF

chmod +x /usr/local/bin/boba-export-metrics.sh

# Update every 5 minutes
crontab -e
# Add:
*/5 * * * * /usr/local/bin/boba-export-metrics.sh
```

## Upgrading

### Pre-Upgrade Checklist

```bash
# 1. Backup configuration
cp -r ~/.boba ~/.boba.backup.$(date +%Y%m%d)

# 2. Backup database
sqlite3 ~/.boba/usage.db ".backup ~/.boba/usage.db.backup.$(date +%Y%m%d)"

# 3. Check current version
boba version

# 4. Review changelog
# https://github.com/royisme/BobaMixer/releases
```

### Upgrade Process

#### Using Homebrew

```bash
# Update tap
brew update

# Upgrade BobaMixer
brew upgrade bobamixer

# Verify new version
boba version

# Run health check
boba doctor
```

#### Manual Upgrade

```bash
# Backup current binary
sudo cp /usr/local/bin/boba /usr/local/bin/boba.old

# Download new version
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_Linux_x86_64.tar.gz

# Extract and install
tar -xzf bobamixer_Linux_x86_64.tar.gz
sudo mv boba /usr/local/bin/
sudo chmod +x /usr/local/bin/boba

# Verify
boba version

# Test
boba doctor

# If issues, rollback
# sudo mv /usr/local/bin/boba.old /usr/local/bin/boba
```

### Post-Upgrade

```bash
# Run health check
boba doctor

# Check for config updates needed
boba doctor --verbose

# Test functionality
boba stats --today
boba route test "test"

# Review and apply suggestions
boba action
```

### Database Migrations

BobaMixer automatically migrates the database schema when needed.

If manual migration is required:

```bash
# Backup first
sqlite3 ~/.boba/usage.db ".backup ~/.boba/usage.db.pre-migration"

# Run migration (if provided)
sqlite3 ~/.boba/usage.db < migration.sql

# Verify
sqlite3 ~/.boba/usage.db "PRAGMA integrity_check;"
```

## Performance Tuning

### Database Performance

```bash
# Enable WAL mode (better concurrency)
sqlite3 ~/.boba/usage.db "PRAGMA journal_mode=WAL;"

# Set cache size (in KB, default ~2MB)
sqlite3 ~/.boba/usage.db "PRAGMA cache_size=-8000;"  # 8MB

# Set page size (before data is added)
sqlite3 ~/.boba/usage.db "PRAGMA page_size=4096;"
```

### Disk Space Management

```bash
# Check usage
du -sh ~/.boba

# Breakdown by component
du -sh ~/.boba/*

# Find large log files
find ~/.boba/logs -type f -size +10M

# Cleanup recommendations
boba-cleanup.sh  # Run cleanup script
```

## Security

### File Permissions

```bash
# Verify permissions
ls -la ~/.boba

# Correct permissions
chmod 700 ~/.boba
chmod 600 ~/.boba/secrets.yaml
chmod 644 ~/.boba/profiles.yaml
chmod 644 ~/.boba/routes.yaml
chmod 644 ~/.boba/pricing.yaml
```

### API Key Rotation

```bash
# 1. Add new key to secrets.yaml
vi ~/.boba/secrets.yaml
# Add: anthropic_key_new: sk-ant-new-key

# 2. Test with new key
# Update one profile temporarily
sed -i 's/anthropic_key/anthropic_key_new/g' ~/.boba/profiles.yaml

# 3. Test
boba doctor

# 4. If successful, update all profiles
# 5. Remove old key from secrets.yaml

# 6. Verify
boba doctor
```

### Secrets Backup

```bash
# Encrypted backup of secrets
gpg --symmetric --cipher-algo AES256 ~/.boba/secrets.yaml

# This creates ~/.boba/secrets.yaml.gpg

# Restore:
gpg --decrypt ~/.boba/secrets.yaml.gpg > ~/.boba/secrets.yaml
chmod 600 ~/.boba/secrets.yaml
```

## Troubleshooting

### Database Locked Error

```bash
# Find processes using database
lsof ~/.boba/usage.db

# Kill if necessary
pkill -f boba

# If persists, check for stale locks
rm ~/.boba/usage.db-shm ~/.boba/usage.db-wal

# Restart
boba doctor
```

### High Disk Usage

```bash
# Check database size
du -h ~/.boba/usage.db

# Check log size
du -h ~/.boba/logs/

# Cleanup
boba-cleanup.sh
rotate-boba-logs.sh

# Vacuum database
sqlite3 ~/.boba/usage.db "VACUUM;"
```

### Performance Issues

```bash
# Analyze database
sqlite3 ~/.boba/usage.db "ANALYZE;"

# Check for missing indexes
sqlite3 ~/.boba/usage.db ".schema"

# Enable WAL mode
sqlite3 ~/.boba/usage.db "PRAGMA journal_mode=WAL;"

# Increase cache
sqlite3 ~/.boba/usage.db "PRAGMA cache_size=-10000;"
```

## Best Practices

### 1. Regular Backups

- Daily automated database backups
- Weekly configuration backups
- Monthly verification of backup restoration

### 2. Monitoring

- Set up health checks
- Monitor disk usage
- Track API spending
- Alert on budget thresholds

### 3. Maintenance

- Weekly database maintenance
- Monthly cleanup of old records
- Quarterly review of configurations
- Regular log rotation

### 4. Security

- Maintain correct file permissions
- Rotate API keys periodically
- Never commit secrets to version control
- Use encrypted backups for secrets

### 5. Documentation

- Document custom configurations
- Track changes to routing rules
- Maintain runbooks for common issues
- Share knowledge with team

## Production Checklist

- [ ] BobaMixer installed and verified
- [ ] Configuration files created and validated
- [ ] Secrets file secured (0600 permissions)
- [ ] Database backup scheduled
- [ ] Log rotation configured
- [ ] Monitoring set up
- [ ] Cleanup scripts scheduled
- [ ] Team has access to documentation
- [ ] Runbooks created for common issues
- [ ] Upgrade process documented

## Next Steps

- **[Troubleshooting](/advanced/troubleshooting)** - Common issues and solutions
- **[CLI Reference](/reference/cli)** - Command documentation
- **[Configuration](/guide/configuration)** - Setup guide
