# 使用分析

BobaMixer 提供全面的分析,帮助你了解 AI 使用模式、优化成本并做出数据驱动的模型选择决策。

## 概览

分析功能包括:

- **实时追踪** - 每个请求的令牌、成本和延迟
- **历史趋势** - 7/30/90 天分析
- **多维视图** - 按配置文件、项目、时间、估算级别
- **会话追踪** - 将相关请求分组
- **导出功能** - JSON、CSV 用于外部分析
- **可视化** - 带图表的 TUI 仪表板

## 关键指标

### 令牌使用

**输入令牌**: 发送到 AI 模型的令牌 (提示、上下文)
**输出令牌**: AI 模型生成的令牌 (响应)

```bash
# 查看令牌使用
boba stats --today
```

### 成本

**总成本**: 输入 + 输出成本合计 (美元)

```bash
# 查看成本
boba stats --7d
```

### 延迟

**延迟**: 从请求到响应完成的时间 (毫秒)

```bash
# 查看延迟统计
boba stats --by-profile --7d
```

## 查看统计

### 基于时间的视图

```bash
# 今天
boba stats --today

# 昨天
boba stats --yesterday

# 最近 7 天
boba stats --7d

# 最近 30 天
boba stats --30d
```

### 基于配置文件的视图

```bash
# 按配置文件
boba stats --by-profile --7d

# 特定配置文件
boba stats --profile claude-sonnet --30d

# 比较配置文件
boba stats --by-profile --compare --7d
```

### 基于项目的视图

```bash
# 按项目
boba stats --by-project --7d

# 特定项目
boba stats --project my-app --30d
```

## 导出数据

### CSV 导出

```bash
# 导出到 CSV
boba report --format csv --output usage-report.csv

# 在 Excel/Numbers 中打开
open usage-report.csv
```

### JSON 导出

```bash
# 导出到 JSON
boba report --format json --output usage-report.json

# 美化打印
boba report --format json | jq '.'
```

## 可视化

### TUI 仪表板

```bash
# 启动仪表板
boba

# 交互功能:
# - 实时统计
# - 配置文件细分图表
# - 成本趋势图
# - 最近请求列表
# - 预算状态
```

## 常见分析模式

### 1. 找到成本优化机会

```bash
# 找到最昂贵的配置文件
boba stats --by-profile --30d --sort-by cost

# 识别路由改进
boba stats --by-profile --breakdown
```

### 2. 比较模型性能

```bash
# 延迟比较
boba stats --latency --by-profile --7d

# 成本效率 (每 1K 令牌成本)
boba stats --by-profile --efficiency --7d
```

### 3. 追踪项目成本

```bash
# 按项目的月度成本
boba stats --by-project --30d

# 用于计费导出
boba report --format csv --by-project --output project-costs-$(date +%Y-%m).csv
```

## 最佳实践

### 1. 定期审查

```bash
# 每日检查
boba stats --today

# 每周审查
boba stats --7d --by-profile

# 每月深入分析
boba report --format csv --30d --output monthly-review.csv
```

### 2. 追踪关键指标

关注:
- **每日/周/月成本** - 预算合规性
- **每个配置文件的成本** - 优化机会
- **估算准确性** - 数据质量
- **P95 延迟** - 用户体验

### 3. 长期存储导出

```bash
# 每月导出
boba report --format json --30d --output archive/usage-$(date +%Y-%m).json

# 压缩旧数据
gzip archive/usage-*.json
```

## 下一步

- **[预算](/zh/features/budgets)** - 设置预算追踪
- **[路由](/zh/features/routing)** - 使用路由规则优化
- **[CLI 参考](/zh/reference/cli)** - 命令文档
