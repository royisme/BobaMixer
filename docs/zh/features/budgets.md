# 预算管理

BobaMixer 的预算管理系统帮助你追踪支出、设置限制并在成本失控之前接收主动警报。

## 概览

关键功能:

- **多级预算** - 全局、项目和配置文件级别
- **主动警报** - 警告和关键阈值
- **成本预测** - 预测未来支出
- **不阻断** - 仅警报,从不中断工作流程
- **支出趋势** - 分析一段时间的模式

## 理念: 警报,不中断

BobaMixer 预算是**建议性的,非限制性的**:

- ✅ 接近限制时你会收到警报
- ✅ 你会获得优化支出的建议
- ✅ 你保持对工作流程的完全控制
- ❌ API 调用永不被阻止
- ❌ 工作永不被中断

这种设计确保生产力不受干扰,同时让你保持知情。

## 预算级别

### 1. 全局预算

适用于所有项目和配置文件的所有使用。

**在 `~/.boba/profiles.yaml` 中配置:**

```yaml
global:
  budget:
    daily_usd: 50.00
    hard_cap: 1000.00
    period_days: 30
    alert_at_percent: 75
    critical_at_percent: 90
```

### 2. 项目预算

适用于特定项目。

**在 `.boba-project.yaml` 中配置 (项目根目录):**

```yaml
budget:
  daily_usd: 10.00
  hard_cap: 200.00
  period_days: 30
  alert_at_percent: 80
  critical_at_percent: 95
```

### 3. 配置文件预算

适用于特定配置文件。

**在 `~/.boba/profiles.yaml` 中配置:**

```yaml
expensive-model:
  adapter: http
  budget:
    daily_usd: 5.00
    monthly_usd: 100.00
```

## 设置预算

### 通过 CLI 快速设置

```bash
# 设置全局每日预算
boba budget --set daily 50

# 设置项目预算
boba budget --set daily 10 --project my-project

# 设置硬上限
boba budget --set cap 1000
```

## 检查预算状态

### 查看当前状态

```bash
# 总体预算状态
boba budget --status

# 详细细分
boba budget --status --detailed

# 特定项目
boba budget --status --project my-project
```

## 警报阈值

### 警告警报 (默认: 75%)

当支出达到预算的 75% 时触发。

### 关键警报 (默认: 90%)

当支出达到预算的 90% 时触发。

### 超预算 (100%+)

预算超支但工作继续。

## 支出分析

### 按时间段

```bash
# 今天的支出
boba stats --today

# 最近 7 天
boba stats --7d

# 最近 30 天
boba stats --30d
```

### 按配置文件

```bash
# 按配置文件细分
boba stats --by-profile --7d

# 找出成本最高的配置文件
boba stats --by-profile --30d --sort-by cost
```

## 成本优化

### 自动建议

BobaMixer 分析使用情况并建议优化:

```bash
# 查看建议
boba action --type suggestion

# 应用建议
boba action apply <suggestion-id>
```

## 最佳实践

### 1. 从保守开始

```yaml
# 从低预算开始
budget:
  daily_usd: 5.00
  hard_cap: 100.00

# 根据实际使用增加
```

### 2. 设置多个级别

```yaml
# 全局兜底
global:
  budget:
    daily_usd: 50.00

# 项目特定
budget:
  daily_usd: 10.00
```

### 3. 定期监控

```bash
# 每日检查
boba budget --status

# 每周审查
boba stats --7d --by-profile
```

## 下一步

- **[分析](/zh/features/analytics)** - 分析支出模式
- **[路由](/zh/features/routing)** - 使用路由规则优化
- **[CLI 参考](/zh/reference/cli)** - 预算命令详情
