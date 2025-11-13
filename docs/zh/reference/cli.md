# CLI 参考

BobaMixer 命令行界面所有命令和选项的完整参考。

## 全局选项

这些选项可用于所有命令:

```bash
--help, -h           显示帮助消息
--version, -v        显示版本信息
--verbose           启用详细输出
--quiet             禁止非错误输出
--no-color          禁用彩色输出
--config PATH       自定义配置目录 (默认: ~/.boba)
```

## 命令

### boba

启动交互式 TUI 仪表板。

```bash
boba [选项]
```

**选项:**
- `--profile PROFILE` - 使用特定配置文件启动
- `--refresh-rate SECONDS` - 仪表板刷新率 (默认: 5)

**示例:**
```bash
# 启动仪表板
boba

# 使用特定配置文件启动
boba --profile claude-sonnet
```

---

### boba doctor

检查配置健康并诊断问题。

```bash
boba doctor [选项]
```

**选项:**
- `--verbose` - 显示详细诊断
- `--fix` - 尝试修复常见问题

**执行的检查:**
- 配置文件语法
- 文件权限 (特别是 secrets.yaml)
- 数据库连接
- API 端点可访问性
- 配置文件有效性
- Secret 引用

**示例:**
```bash
# 基本健康检查
boba doctor

# 详细诊断
boba doctor --verbose

# 自动修复问题
boba doctor --fix
```

---

### boba use

设置活动配置文件。

```bash
boba use PROFILE
```

**参数:**
- `PROFILE` - 要激活的配置文件名称

**示例:**
```bash
# 激活配置文件
boba use claude-sonnet

# 验证活动配置文件
boba ls --current
```

---

### boba ls

列出配置文件、项目或会话。

```bash
boba ls [选项]
```

**选项:**
- `--profiles` - 列出所有配置文件
- `--projects` - 列出所有项目
- `--sessions` - 列出最近的会话
- `--current` - 显示当前活动配置文件
- `--tag TAG` - 按标签过滤
- `--verbose` - 显示详细信息

**示例:**
```bash
# 列出所有配置文件
boba ls --profiles

# 当前配置文件
boba ls --current

# 具有特定标签的配置文件
boba ls --profiles --tag work
```

---

### boba stats

查看使用统计。

```bash
boba stats [选项]
```

**时间范围选项:**
- `--today` - 今天的统计
- `--yesterday` - 昨天的统计
- `--7d` - 最近 7 天
- `--30d` - 最近 30 天
- `--from DATE --to DATE` - 自定义日期范围 (YYYY-MM-DD)

**细分选项:**
- `--by-profile` - 按配置文件分组
- `--by-project` - 按项目分组
- `--by-session` - 按会话分组
- `--by-estimate` - 按估算准确性级别分组

**示例:**
```bash
# 今天的统计
boba stats --today

# 最近 7 天按配置文件
boba stats --7d --by-profile

# 比较配置文件性能
boba stats --7d --by-profile --compare

# 延迟分析
boba stats --7d --latency --percentiles
```

---

### boba route

管理和测试路由规则。

```bash
boba route SUBCOMMAND [选项]
```

**子命令:**
- `test TEXT` - 使用文本或文件测试路由
- `list` - 列出所有路由规则
- `validate` - 验证路由配置

**测试选项:**
- `@FILE` - 使用文件内容测试
- `--verbose` - 显示详细评估
- `--explain` - 解释匹配过程

**示例:**
```bash
# 使用文本测试
boba route test "编写一个排序函数"

# 使用文件测试
boba route test @prompts/example.txt

# 详细输出
boba route test --verbose "格式化这段代码"

# 列出所有规则
boba route list
```

---

### boba budget

管理预算并查看支出。

```bash
boba budget [选项]
```

**查看选项:**
- `--status` - 显示预算状态
- `--detailed` - 详细预算细分
- `--project NAME` - 特定项目预算

**设置选项:**
- `--set TYPE AMOUNT` - 设置预算 (daily|weekly|monthly|cap)
- `--project NAME` - 设置项目特定预算

**示例:**
```bash
# 查看状态
boba budget --status

# 详细视图
boba budget --status --detailed

# 设置每日预算
boba budget --set daily 50

# 设置硬上限
boba budget --set cap 1000
```

---

### boba action

查看和管理警报和建议。

```bash
boba action [选项]
```

**选项:**
- `--type TYPE` - 按类型过滤 (budget|suggestion|alert)
- `apply ID` - 应用建议
- `dismiss ID` - 关闭操作
- `preview ID` - 应用前预览

**示例:**
```bash
# 查看所有操作
boba action

# 仅预算警报
boba action --type budget

# 应用建议
boba action apply suggestion-123

# 先预览
boba action preview suggestion-123
```

---

### boba report

导出使用数据。

```bash
boba report [选项]
```

**格式选项:**
- `--format FORMAT` - 导出格式 (json|csv)
- `--output FILE` - 输出文件路径

**过滤选项:**
- `--from DATE --to DATE` - 日期范围
- `--profile PROFILE` - 特定配置文件
- `--project PROJECT` - 特定项目

**示例:**
```bash
# 导出到 JSON
boba report --format json --output usage.json

# 导出到 CSV
boba report --format csv --output usage.csv

# 最近 30 天
boba report --format json --from $(date -d '30 days ago' +%Y-%m-%d) --output last-month.json
```

---

### boba edit

编辑配置文件。

```bash
boba edit CONFIG
```

**参数:**
- `profiles` - 编辑 profiles.yaml
- `routes` - 编辑 routes.yaml
- `pricing` - 编辑 pricing.yaml
- `secrets` - 编辑 secrets.yaml

**示例:**
```bash
# 编辑配置文件
boba edit profiles

# 编辑路由规则
boba edit routes

# 编辑 secrets
boba edit secrets
```

---

### boba hooks

管理 git hooks 集成。

```bash
boba hooks SUBCOMMAND [选项]
```

**子命令:**
- `install` - 在当前仓库安装 git hooks
- `remove` - 移除 git hooks
- `status` - 显示 hook 安装状态

**示例:**
```bash
# 在当前仓库安装 hooks
cd my-project
boba hooks install

# 检查状态
boba hooks status

# 移除 hooks
boba hooks remove
```

---

### boba version

显示版本信息。

```bash
boba version [选项]
```

**选项:**
- `--check-update` - 检查更新版本

**示例:**
```bash
# 显示版本
boba version

# 检查更新
boba version --check-update
```

---

## 环境变量

```bash
# 自定义配置目录
export BOBA_HOME=/custom/path

# 日志级别 (trace|debug|info|warn|error)
export BOBA_LOG_LEVEL=debug

# 自定义数据库路径
export BOBA_DB_PATH=/custom/usage.db

# 用于 boba edit 的编辑器
export EDITOR=vim

# 禁用颜色
export NO_COLOR=1

# API 超时 (秒)
export BOBA_API_TIMEOUT=30
```

## 退出代码

```bash
0   # 成功
1   # 一般错误
2   # 配置错误
3   # 数据库错误
4   # API 错误
5   # 权限错误
10  # 用户取消
```

## 下一步

- **[配置文件参考](/zh/reference/config-files)** - 详细配置架构
- **[快速开始](/zh/guide/getting-started)** - 基本使用指南
- **[故障排除](/zh/advanced/troubleshooting)** - 常见问题
