# 故障排除

BobaMixer 的常见问题、解决方案和常见问题解答。

## 快速诊断

遇到问题时,始终从以下开始:

```bash
boba doctor
```

这会检查:
- 配置文件语法
- 文件权限
- 数据库连接
- API 端点可访问性
- 配置文件有效性

## 常见问题

### 安装问题

#### 找不到命令

**问题**: `bash: boba: command not found`

**解决方案**:

1. 检查是否已安装:
   ```bash
   which boba
   ```

2. 如果使用 Go install,将 GOPATH 添加到 PATH:
   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   # 添加到 ~/.bashrc 或 ~/.zshrc 以使其永久生效
   ```

3. 如果使用手动安装,验证位置:
   ```bash
   ls -l /usr/local/bin/boba
   ```

4. 使其可执行:
   ```bash
   sudo chmod +x /usr/local/bin/boba
   ```

---

### 配置问题

#### Secrets 文件权限错误

**问题**: `secrets.yaml must have 0600 permissions`

**解决方案**:

```bash
chmod 600 ~/.boba/secrets.yaml

# 验证
ls -l ~/.boba/secrets.yaml
# 应显示: -rw-------
```

#### YAML 语法错误

**问题**: `error parsing config: yaml: line X: ...`

**解决方案**:

1. 验证 YAML 语法:
   ```bash
   yamllint ~/.boba/profiles.yaml
   ```

2. 常见 YAML 错误:
   - **缩进不正确** (使用 2 个空格,不是制表符)
   - **特殊字符周围缺少引号**
   - **键或值中的无效字符**

#### Secret 未找到

**问题**: `secret not found: anthropic_key`

**解决方案**:

1. 检查 secret 是否存在:
   ```bash
   grep "anthropic_key" ~/.boba/secrets.yaml
   ```

2. 验证引用格式:
   ```yaml
   # 正确
   x-api-key: "secret://anthropic_key"

   # 错误
   x-api-key: "secret://anthropic_key/"
   x-api-key: "secret:anthropic_key"
   ```

3. 添加缺失的 secret:
   ```bash
   boba edit secrets
   # 添加: anthropic_key: sk-ant-your-key
   ```

---

### 数据库问题

#### 数据库被锁定

**问题**: `database is locked`

**解决方案**:

1. 查找使用数据库的进程:
   ```bash
   lsof ~/.boba/usage.db
   ```

2. 如有必要终止:
   ```bash
   pkill -f boba
   ```

3. 移除陈旧锁:
   ```bash
   rm -f ~/.boba/usage.db-shm ~/.boba/usage.db-wal
   ```

4. 测试:
   ```bash
   boba stats --today
   ```

---

### API 问题

#### API 调用失败

**问题**: `API call failed: connection refused`

**解决方案**:

1. 检查互联网连接:
   ```bash
   ping api.anthropic.com
   ```

2. 手动测试端点:
   ```bash
   curl -v https://api.anthropic.com/v1/messages
   ```

3. 验证 API 密钥:
   ```bash
   # 检查密钥是否存在
   grep "anthropic_key" ~/.boba/secrets.yaml

   # 使用 curl 测试
   curl -X POST https://api.anthropic.com/v1/messages \
     -H "x-api-key: YOUR-KEY" \
     -H "anthropic-version: 2023-06-01" \
     -H "content-type: application/json" \
     -d '{"model":"claude-3-5-sonnet-20241022","max_tokens":10,"messages":[{"role":"user","content":"test"}]}'
   ```

---

### 路由问题

#### 规则不匹配

**问题**: 选择了错误的配置文件

**解决方案**:

1. 测试路由:
   ```bash
   boba route test "你的文本在这里"
   ```

2. 启用详细模式:
   ```bash
   boba route test --verbose "你的文本在这里"
   ```

3. 检查规则顺序 (第一个匹配获胜):
   ```yaml
   # 正确顺序
   rules:
     - if: "ctx_chars > 50000"
       use: 昂贵
     - if: "ctx_chars > 0"
       use: 便宜
   ```

4. 验证路由配置:
   ```bash
   boba route validate
   ```

---

### 预算问题

#### 预算未追踪

**问题**: 预算状态显示 $0

**解决方案**:

1. 检查数据库有数据:
   ```bash
   sqlite3 ~/.boba/usage.db "SELECT COUNT(*) FROM usage_records;"
   ```

2. 验证成本计算:
   ```bash
   boba stats --today
   ```

3. 查看估算准确性:
   ```bash
   boba stats --by-estimate
   ```

---

## 常见问题解答

### 一般

#### 什么是 BobaMixer?

BobaMixer 是一个用于追踪、分析和优化 AI/LLM API 使用和成本的 CLI 工具。

#### BobaMixer 免费吗?

是的!BobaMixer 是 MIT 许可下的开源软件。你只需为提供商的 API 使用付费。

#### BobaMixer 会拦截我的 API 调用吗?

不会。BobaMixer 是一个追踪工具。你明确地在想要追踪使用时调用它。

### 安装

#### 我应该使用哪种安装方法?

- **Homebrew**: macOS/Linux 最简单
- **Go install**: 如果已安装 Go 很好
- **二进制文件**: 适用于服务器或没有 Go
- **源代码**: 用于开发或自定义构建

### 配置

#### 配置存储在哪里?

默认: `~/.boba/`

使用以下方式覆盖: `export BOBA_HOME=/custom/path`

#### 多个用户可以共享配置吗?

可以!对共享配置使用符号链接,保留个人 secrets:

```bash
ln -s /shared/boba/profiles.yaml ~/.boba/profiles.yaml
cp my-secrets.yaml ~/.boba/secrets.yaml
chmod 600 ~/.boba/secrets.yaml
```

### 使用

#### 成本估算有多准确?

- **精确** (API 响应): 100%
- **映射** (定价配置): 95-99%
- **启发式** (基于字符): 70-90%

使用以下检查: `boba stats --by-estimate`

#### 预算会阻止 API 调用吗?

不会。BobaMixer 使用"警报,不中断"理念。你会收到警告,但工作永不被阻止。

---

## 获取帮助

### 自助资源

1. **运行诊断**:
   ```bash
   boba doctor --verbose
   ```

2. **检查日志**:
   ```bash
   tail -f ~/.boba/logs/boba.log
   ```

3. **查看文档**:
   - [快速开始](/zh/guide/getting-started)
   - [配置指南](/zh/guide/configuration)
   - [CLI 参考](/zh/reference/cli)

### 社区支持

1. **GitHub 讨论**: [提问](https://github.com/royisme/BobaMixer/discussions)
2. **GitHub Issues**: [报告错误](https://github.com/royisme/BobaMixer/issues)
3. **文档**: [完整文档](https://royisme.github.io/BobaMixer/)

## 下一步

- **[运维指南](/zh/advanced/operations)** - 生产最佳实践
- **[CLI 参考](/zh/reference/cli)** - 命令文档
- **[配置](/zh/guide/configuration)** - 设置指南
