# Phase 2: Operational & Optimization Features

## 已完成的功能

Phase 2 在Phase 1的基础上，实现了3个运营和优化相关的TUI视图，完成从CLI到TUI的进一步迁移。

### 1. Proxy控制视图 (按键: 7)

**功能**：
- 实时显示proxy server状态（Running/Stopped/Checking）
- 显示proxy地址配置
- 说明proxy功能和使用方法
- 提供启动proxy的指引

**界面特性**：
```
🌐 Proxy Status

  Status:   ● Running
  Address:  127.0.0.1:7777

ℹ️  Information
  The proxy server intercepts AI API requests from CLI tools
  and routes them through BobaMixer for tracking and control.

📝 Configuration
  Tools with proxy enabled will automatically use:
  • HTTP_PROXY=127.0.0.1:7777
  • HTTPS_PROXY=127.0.0.1:7777
```

**状态指示器**：
- `●` (绿色) - Proxy正在运行
- `○` (红色) - Proxy已停止
- `⋯` (灰色) - 正在检查状态

**CLI集成**：
- 使用 `boba proxy serve` 在终端启动proxy
- TUI显示当前状态，无需切换窗口

### 2. Routing测试器视图 (按键: 8)

**功能**：
- 说明routing规则的工作原理
- 提供测试routing的使用指南
- 展示示例查询和结果
- 列出routing考虑的上下文因素

**界面特性**：
```
🧪 Test Routing Rules
  Test how routing rules would apply to different queries.

💡 How to Use
  1. Prepare a test query (text or file)
  2. Run: boba route test "your query text"
  3. Or: boba route test @path/to/file.txt

📋 Example
  $ boba route test "Write a Python function"
  → Profile: claude-sonnet-3.5
  → Rule: short-query-fast-model
  → Reason: Query < 100 chars

ℹ️  Context Detection
  Routing considers:
  • Query length and complexity
  • Current project and branch
  • Time of day (day/evening/night)
  • Project type (go, web, etc.)
```

**教育性设计**：
- 帮助用户理解routing决策过程
- 鼓励用户优化routing规则
- 提供清晰的CLI命令示例

### 3. Suggestions视图 (按键: 9)

**功能**：
- 显示基于最近7天usage的优化建议
- 按优先级排序建议（P1-P5）
- 支持列表导航查看详情
- 显示预估影响和推荐行动

**界面特性**：
```
💡 Recommendations (Last 7 Days)

  🔴 ⚠️  [P5] Unusual Spending Spikes Detected
  🟠 💰 [P4] Rising Cost Trend Detected
  🟡 🔄 [P3] High Dependency on Expensive Profile

Details:
  Your daily costs have increased by 35% recently.
  Consider optimizing usage or switching to more cost-effective models.

  Impact: Could save $4.50/day

Recommended Actions:
  1. Review recent high-cost sessions
  2. Consider using smaller models for simple tasks
  3. Implement caching to reduce redundant API calls
  4. Set daily budget limits to control spending
```

**建议类型**：
- 💰 **Cost Optimization** - 成本优化建议
- 🔄 **Profile Switch** - Profile切换建议
- 📊 **Budget Adjust** - 预算调整建议
- ⚠️  **Anomaly Detection** - 异常检测警报
- 📈 **Usage Pattern** - 使用模式洞察

**优先级指示**：
- 🔴 P5 - Critical (严重)
- 🟠 P4 - High (高)
- 🟡 P3 - Medium (中)
- 🟢 P2 - Low (低)
- ⚪ P1 - Info (信息)

**CLI集成**：
- 使用 `boba action` 查看建议
- 使用 `boba action --auto` 自动应用高优先级建议

## 导航增强

### 扩展的数字键导航

现在支持1-9数字键快速切换视图：

| 按键 | 视图 | 描述 |
|------|------|------|
| `1` | Dashboard | 主仪表板 |
| `2` | Providers | AI Provider管理 |
| `3` | Tools | CLI工具管理 |
| `4` | Bindings | 工具↔Provider绑定 |
| `5` | Secrets | API密钥管理 |
| `6` | Stats | 统计数据 |
| `7` | **Proxy** (新) | **Proxy服务器控制** |
| `8` | **Routing** (新) | **Routing规则测试器** |
| `9` | **Suggestions** (新) | **优化建议** |

### Tab键循环

- **Tab** - 从当前视图循环到下一个（1→2→...→9→1）
- 智能加载：切换到需要数据的视图时自动加载

## 与CLI命令的对应关系

| 旧CLI命令 | 新TUI视图 | 快捷键 |
|---------|---------|--------|
| `boba proxy status` | Proxy视图 | `7` |
| `boba route test <text>` | Routing视图 | `8` |
| `boba action` | Suggestions视图 | `9` |

## 技术实现

### 扩展的ViewMode

```go
const (
    viewDashboard viewMode = iota
    viewProviders
    viewTools
    viewBindings
    viewSecrets
    viewStats
    viewProxy       // 新增
    viewRouting     // 新增
    viewSuggestions // 新增
)
```

### 新增的数据加载

**Suggestions加载**：
```go
func (m *DashboardModel) loadSuggestions() tea.Msg {
    engine := suggestions.NewEngine(db)
    suggs, err := engine.GenerateSuggestions(7)
    // ...
}
```

**Proxy状态检查**（已有，复用）：
```go
func checkProxyStatus() tea.Msg {
    // HTTP GET to proxy health endpoint
    // Returns proxyStatusMsg
}
```

### 新增的渲染函数

1. `renderProxyView()` - 显示proxy状态和说明
2. `renderRoutingView()` - 显示routing使用指南
3. `renderSuggestionsView()` - 显示优化建议列表

### 键盘处理更新

```go
case "7": m.currentView = viewProxy; return m, checkProxyStatus
case "8": m.currentView = viewRouting; return m, nil
case "9": m.currentView = viewSuggestions; return m, m.loadSuggestions

case "tab":
    m.currentView = (m.currentView + 1) % 9  // 从6改为9
```

## 用户体验改进

### Phase 2特色

1. **运营可见性** - Proxy状态一目了然，无需切换终端
2. **教育性设计** - Routing视图帮助理解路由机制
3. **主动优化** - Suggestions提供可操作的优化建议
4. **CLI/TUI协同** - 保留CLI命令用于脚本，TUI用于查看

### 信息架构

```
Phase 1 (配置管理)          Phase 2 (运营优化)
├── Providers              ├── Proxy (运营状态)
├── Tools                  ├── Routing (测试工具)
├── Bindings               └── Suggestions (优化建议)
├── Secrets
└── Stats
```

## Suggestions引擎能力

### 分析维度

1. **Cost Trend Analysis** - 成本趋势分析
   - 检测成本上涨（>20%触发）
   - 计算平均vs实际成本差异
   - 提供具体节省估算

2. **Profile Usage Analysis** - Profile使用分析
   - 识别高成本Profile依赖（>60%）
   - 建议更便宜的替代方案
   - 估算切换节省

3. **Anomaly Detection** - 异常检测
   - 识别超出平均2倍的异常日
   - 计算异常成本总额
   - 建议预防措施

4. **Budget Optimization** - 预算优化
   - 分析峰值vs平均成本
   - 建议合理的预算buffer
   - 提供月度预算建议

### 行动建议示例

**Cost Optimization**：
- 为简单任务使用更小的模型
- 实现缓存减少重复调用
- 设置每日预算限制

**Profile Switch**：
- 为routine任务使用GPT-3.5或Claude Haiku
- 创建不同复杂度的Profile
- 优化routing规则

**Anomaly Prevention**：
- 检查高成本会话的异常长对话
- 查找失控进程或循环
- 实施速率限制
- 设置预算警报

## CLI集成命令参考

虽然功能都在TUI中可见，CLI命令仍可用于脚本和自动化：

```bash
# Proxy管理
boba proxy serve          # 启动proxy server
boba proxy status         # 检查proxy状态

# Routing测试
boba route test "query text"
boba route test @file.txt

# Suggestions
boba action               # 查看建议
boba action --auto        # 自动应用高优先级建议
```

## 测试验证

✅ **编译通过** - `go build ./...`
✅ **代码检查通过** - `go vet ./internal/ui`
✅ **Tab键循环** - 9个视图正确循环
✅ **数字键跳转** - 1-9直接切换
✅ **数据加载** - Suggestions按需加载

## 与Phase 1的对比

| 维度 | Phase 1 | Phase 2 |
|------|---------|---------|
| 视图数量 | 6个 | 9个 (+3) |
| 焦点 | 配置管理 | 运营优化 |
| 交互性 | 浏览和切换 | 浏览、测试、建议 |
| 数据源 | 配置文件 | 配置文件 + 数据库分析 |
| CLI替代 | 基础命令 | 高级功能 |

## 下一步 (Phase 3参考)

根据 `docs/TUI_ENHANCEMENT_PLAN.md`，Phase 3将实现：

1. **Reports生成器** - 在TUI中配置和生成使用报告
2. **Hooks管理** - Git hooks的可视化管理
3. **Config编辑器** - TUI内的配置文件编辑

## 总结

Phase 2成功扩展了TUI功能，从配置管理延伸到运营和优化领域。通过Proxy控制、Routing测试和Suggestions视图，用户现在可以：

- 监控proxy运行状态
- 理解和测试routing规则
- 获取基于数据的优化建议
- 在TUI中查看所有关键信息

这进一步强化了BobaMixer作为TUI优先应用的定位，同时保留必要的CLI命令用于自动化场景。
