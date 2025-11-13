# 智能路由

BobaMixer 的智能路由系统根据上下文、任务类型、项目特征和使用模式自动选择最优 AI 配置文件。

## 概览

路由规则允许你:

- **优化成本** 通过对简单任务使用更便宜的模型
- **提高质量** 通过对复杂工作使用更好的模型
- **节省时间** 通过在适当时使用更快的模型
- **探索替代方案** 使用 epsilon-greedy 优化
- **适应上下文** 基于项目、分支、时间等

## 配置

路由在 `~/.boba/routes.yaml` 中定义:

```yaml
rules:
  - id: 规则标识符
    if: "条件表达式"
    use: 配置文件名称
    fallback: 备用配置文件
    explain: "为什么存在此规则"

exploration:
  enabled: true
  epsilon: 0.03
```

## 路由 DSL

### 可用变量

| 变量 | 类型 | 描述 | 示例 |
|----------|------|-------------|---------|
| `ctx_chars` | int | 输入字符数 | `ctx_chars > 50000` |
| `text` | string | 输入文本内容 | `text.contains('代码')` |
| `project_types` | array | 项目类型 | `project_types.includes('go')` |
| `branch` | string | Git 分支名称 | `branch.matches('main')` |
| `time_of_day` | string | 时间段 | `time_of_day == 'night'` |

### 可用函数

| 函数 | 描述 | 示例 |
|----------|-------------|---------|
| `text.matches(pattern)` | 正则匹配 | `text.matches('\\b代码\\b')` |
| `text.contains(str)` | 子字符串搜索 | `text.contains('审查')` |
| `array.includes(item)` | 数组成员 | `project_types.includes('react')` |

## 常见路由模式

### 1. 基于上下文大小的路由

根据输入大小路由:

```yaml
rules:
  - id: 超大上下文
    if: "ctx_chars > 100000"
    use: claude-opus
    fallback: claude-sonnet
    explain: "超大上下文需要最高容量模型"

  - id: 大上下文
    if: "ctx_chars > 50000"
    use: claude-sonnet
    explain: "大上下文需要强大模型"

  - id: 小上下文
    if: "ctx_chars > 0"
    use: gpt-4o-mini
    explain: "小上下文,经济模型"
```

### 2. 任务类型识别

根据任务关键词路由:

```yaml
rules:
  - id: 代码生成
    if: "text.matches('编写.*函数|实现|创建.*类')"
    use: 代码专家
    explain: "代码生成任务"

  - id: 代码审查
    if: "text.matches('审查|分析.*代码|重构')"
    use: 代码审查员
    explain: "代码审查任务"

  - id: 格式化
    if: "text.matches('格式|prettier|eslint')"
    use: 快速格式化器
    explain: "简单格式化任务"
```

### 3. 项目类型路由

根据项目特征路由:

```yaml
rules:
  - id: 前端工作
    if: "project_types.includes('react') || project_types.includes('vue')"
    use: 前端专家
    explain: "前端项目"

  - id: 后端工作
    if: "project_types.includes('go')"
    use: 后端专家
    explain: "后端项目"
```

### 4. 基于分支的路由

不同分支使用不同配置文件:

```yaml
rules:
  - id: 生产分支
    if: "branch.matches('main|master|prod')"
    use: 高精度
    fallback: 中精度
    explain: "生产分支需要最高精度"

  - id: 功能分支
    if: "branch.matches('feature/.*')"
    use: 平衡模型
    explain: "功能开发"
```

### 5. 基于时间的路由

在非高峰时段优化成本:

```yaml
rules:
  - id: 夜间模式
    if: "time_of_day == 'night'"
    use: 成本优化
    explain: "非高峰时段,使用更便宜的模型"

  - id: 工作时间
    if: "time_of_day == 'day'"
    use: 高性能
    explain: "工作时间,优先考虑速度"
```

## Epsilon-Greedy 探索

BobaMixer 可以使用探索自动发现更好的配置文件。

### 工作原理

1. **利用** (97%): 根据规则使用已知最佳配置文件
2. **探索** (3%): 尝试替代配置文件以发现更好的选项
3. **学习**: 追踪性能 (成本、延迟、质量)
4. **建议**: 发现更好的配置文件时推荐

### 配置

```yaml
exploration:
  enabled: true
  epsilon: 0.03          # 3% 探索率
  min_samples: 10        # 探索前需要 10+ 样本
  cooldown_hours: 24     # 重新测试配置文件之间等待 24 小时
```

## 测试路由

始终在部署前测试路由规则。

### 使用文本测试

```bash
# 内联文本测试
boba route test "编写一个排序数组的函数"
```

### 使用文件测试

```bash
# 文件内容测试
boba route test @path/to/prompt.txt
```

### 详细测试

```bash
# 查看详细评估
boba route test --verbose "你的提示在这里"
```

## 最佳实践

### 1. 顺序很重要

规则从上到下评估。第一个匹配获胜。

**正确顺序** (先具体规则):
```yaml
rules:
  - if: "ctx_chars > 50000"
    use: 昂贵模型
  - if: "ctx_chars > 0"
    use: 便宜模型
```

### 2. 始终提供解释

```yaml
- id: 大上下文
  if: "ctx_chars > 50000"
  use: 高容量
  explain: "大上下文需要高容量模型以保证质量"
```

### 3. 对关键路径使用回退

```yaml
- id: 生产
  if: "branch.matches('main')"
  use: 最佳模型
  fallback: 良好模型
  explain: "带有回退的生产"
```

## 下一步

- **[预算](/zh/features/budgets)** - 设置预算管理
- **[分析](/zh/features/analytics)** - 分析路由模式
- **[配置](/zh/guide/configuration)** - 高级路由配置
