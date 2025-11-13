# 快速开始

欢迎使用 BobaMixer!本指南将帮助你快速上手 BobaMixer,这是一款全面的命令行工具,用于管理多个 AI 提供商、追踪成本并优化 AI 工作负载路由。

## 什么是 BobaMixer?

BobaMixer 是一款智能 AI 适配器路由器,可帮助你:

- **追踪使用情况**: 监控多个 AI 提供商的令牌、成本和延迟
- **智能路由**: 根据上下文和任务类型自动选择最佳模型
- **管理预算**: 设置支出限制并接收主动警报
- **优化成本**: 获取 AI 驱动的建议以降低支出
- **分析模式**: 通过全面的分析了解你的 AI 使用情况

## 快速开始

### 1. 安装 BobaMixer

选择你偏好的安装方法:

**使用 Go:**
```bash
go install github.com/royisme/bobamixer/cmd/boba@latest
```

**使用 Homebrew (macOS/Linux):**
```bash
brew tap royisme/tap
brew install bobamixer
```

**下载二进制文件:**
从 [GitHub Releases](https://github.com/royisme/BobaMixer/releases) 下载最新版本。

详细安装说明请查看[安装指南](/zh/guide/installation)。

### 2. 初始化配置

运行 doctor 命令创建默认配置:

```bash
boba doctor
```

这会创建 `~/.boba/` 目录及示例配置文件:
- `profiles.yaml` - 配置文件定义
- `routes.yaml` - 路由规则
- `pricing.yaml` - 模型定价
- `secrets.yaml` - API 密钥 (0600 权限)
- `usage.db` - 用于追踪的 SQLite 数据库

### 3. 配置你的第一个配置文件

编辑 `~/.boba/profiles.yaml` 并添加你的第一个配置文件:

```yaml
default:
  adapter: http
  provider: anthropic
  endpoint: https://api.anthropic.com/v1/messages
  model: claude-3-5-sonnet-20241022
  headers:
    anthropic-version: "2023-06-01"
    x-api-key: "secret://anthropic_key"
```

### 4. 添加你的 API 密钥

编辑 `~/.boba/secrets.yaml` 并添加你的 API 密钥:

```yaml
anthropic_key: sk-ant-your-actual-key-here
```

确保 secrets 文件具有正确的权限:
```bash
chmod 600 ~/.boba/secrets.yaml
```

### 5. 激活配置文件

将默认配置文件设为活动:

```bash
boba use default
```

### 6. 验证设置

检查所有内容是否配置正确:

```bash
boba doctor
```

你应该看到所有配置项的绿色对勾标记。

### 7. 启动 TUI 仪表板

启动交互式仪表板:

```bash
boba
```

仪表板显示:
- 当前活动配置文件
- 今日使用统计
- 预算状态
- 最近通知
- 快速操作

## 基本用法

### 查看所有配置文件

列出所有已配置的配置文件:

```bash
boba ls --profiles
```

### 切换配置文件

激活不同的配置文件:

```bash
boba use <配置文件名称>
```

### 检查使用统计

查看今日统计:

```bash
boba stats --today
```

查看最近 7 天:

```bash
boba stats --7d
```

查看按配置文件细分:

```bash
boba stats --7d --by-profile
```

### 测试路由规则

测试给定输入会选择哪个配置文件:

```bash
boba route test "编写一个排序数组的函数"
```

使用文件内容测试:

```bash
boba route test @path/to/file.txt
```

### 检查预算状态

查看预算状态和警报:

```bash
boba budget --status
```

查看待处理操作和建议:

```bash
boba action
```

## 下一步

现在 BobaMixer 已经启动运行,探索这些主题:

- **[配置指南](/zh/guide/configuration)** - 了解所有配置选项
- **[适配器](/zh/features/adapters)** - 理解不同的适配器类型 (HTTP、Tool、MCP)
- **[智能路由](/zh/features/routing)** - 设置智能路由规则
- **[预算管理](/zh/features/budgets)** - 配置预算和警报
- **[分析](/zh/features/analytics)** - 分析你的使用模式

## 获取帮助

如果遇到任何问题:

1. 运行 `boba doctor` 检查配置健康状况
2. 查看[故障排除指南](/zh/advanced/troubleshooting)
3. 查阅 [FAQ](/zh/advanced/troubleshooting#faq)
4. 在 [GitHub](https://github.com/royisme/BobaMixer/issues) 上提出问题

## 示例工作流程

以下是使用 BobaMixer 的典型工作流程:

```bash
# 早上:检查昨天的统计
boba stats --yesterday

# 设置工作会话
boba use work-heavy

# 在你的项目上工作...
# BobaMixer 通过适配器自动追踪使用情况

# 检查当前会话
boba stats --today

# 查看建议
boba action

# 应用建议的优化
boba action apply suggestion-id

# 一天结束:生成报告
boba report --format json --output daily-report.json
```

## 社区和支持

- **文档**: [https://royisme.github.io/BobaMixer/](https://royisme.github.io/BobaMixer/)
- **GitHub**: [https://github.com/royisme/BobaMixer](https://github.com/royisme/BobaMixer)
- **问题**: [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
- **讨论**: [GitHub Discussions](https://github.com/royisme/BobaMixer/discussions)

欢迎来到 BobaMixer 社区!
