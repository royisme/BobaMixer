# 运维指南

在生产环境中运行、维护和故障排除 BobaMixer 的最佳实践。

## 生产安装

### 系统要求

- **操作系统**: Linux (推荐)、macOS、Windows (WSL)
- **Go**: 1.22+ (如果从源代码构建)
- **SQLite**: 3.x
- **磁盘空间**: 50MB + 数据库增长 (估计每个 API 调用 ~1KB)
- **内存**: 典型 50MB,峰值 200MB

### 安装方法

#### 使用 Homebrew (推荐)

```bash
brew tap royisme/tap
brew install bobamixer

# 验证安装
boba version
boba doctor
```

#### 手动安装

```bash
# 下载最新版本
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_Linux_x86_64.tar.gz

# 解压
tar -xzf bobamixer_Linux_x86_64.tar.gz

# 安装
sudo mv boba /usr/local/bin/
sudo chmod +x /usr/local/bin/boba

# 验证
boba version
```

## 数据库管理

### 数据库位置

```
~/.boba/usage.db
```

### 备份策略

#### 自动每日备份

```bash
# 创建备份脚本
cat > /usr/local/bin/backup-boba.sh << 'EOF'
#!/bin/bash
BACKUP_DIR=~/backups/boba
mkdir -p $BACKUP_DIR

# 热备份 (使用时安全)
sqlite3 ~/.boba/usage.db ".backup $BACKUP_DIR/usage-$(date +%Y%m%d).db"

# 压缩旧备份
find $BACKUP_DIR -name "usage-*.db" -mtime +7 -exec gzip {} \;

# 删除 90 天前的备份
find $BACKUP_DIR -name "usage-*.db.gz" -mtime +90 -delete

echo "备份完成: $BACKUP_DIR/usage-$(date +%Y%m%d).db"
EOF

chmod +x /usr/local/bin/backup-boba.sh

# 安排每日备份 (添加到 crontab)
crontab -e
# 添加行:
0 2 * * * /usr/local/bin/backup-boba.sh
```

#### 手动备份

```bash
# 热备份 (使用时安全)
sqlite3 ~/.boba/usage.db ".backup /tmp/usage-backup.db"

# 或简单复制 (仅当 boba 未运行时)
pkill -f boba
cp ~/.boba/usage.db ~/.boba/usage.db.backup
```

### 数据库维护

#### 每周维护

```bash
# Vacuum 数据库
sqlite3 ~/.boba/usage.db "VACUUM;"

# 检查完整性
sqlite3 ~/.boba/usage.db "PRAGMA integrity_check;"

# 分析
sqlite3 ~/.boba/usage.db "ANALYZE;"
```

### 数据清理

#### 清除旧记录

```bash
# 删除 90 天前的记录
sqlite3 ~/.boba/usage.db "DELETE FROM usage_records WHERE ts < strftime('%s', 'now', '-90 days');"

# 回收空间
sqlite3 ~/.boba/usage.db "VACUUM;"
```

## 监控

### 健康检查

```bash
# 基本健康检查
boba doctor

# 详细检查
boba doctor --verbose
```

## 升级

### 预升级检查清单

```bash
# 1. 备份配置
cp -r ~/.boba ~/.boba.backup.$(date +%Y%m%d)

# 2. 备份数据库
sqlite3 ~/.boba/usage.db ".backup ~/.boba/usage.db.backup.$(date +%Y%m%d)"

# 3. 检查当前版本
boba version

# 4. 查看变更日志
# https://github.com/royisme/BobaMixer/releases
```

### 升级过程

#### 使用 Homebrew

```bash
# 更新 tap
brew update

# 升级 BobaMixer
brew upgrade bobamixer

# 验证新版本
boba version

# 运行健康检查
boba doctor
```

## 最佳实践

### 1. 定期备份

- 每日自动数据库备份
- 每周配置备份
- 每月验证备份恢复

### 2. 监控

- 设置健康检查
- 监控磁盘使用
- 追踪 API 支出
- 预算阈值警报

### 3. 维护

- 每周数据库维护
- 每月清理旧记录
- 每季度审查配置
- 定期日志轮转

### 4. 安全

- 维护正确的文件权限
- 定期轮换 API 密钥
- 永不将 secrets 提交到版本控制
- 使用加密备份存储 secrets

## 下一步

- **[故障排除](/zh/advanced/troubleshooting)** - 常见问题和解决方案
- **[CLI 参考](/zh/reference/cli)** - 命令文档
- **[配置](/zh/guide/configuration)** - 设置指南
