# BobaMixer CLI Redesign - TUI优先方案

## 问题分析

当前CLI设计存在严重的身份混乱问题：

### 当前设计的问题

1. **混合范式**：默认启动TUI，但提供了20+个CLI子命令
2. **用户困惑**：不清楚什么时候用TUI，什么时候用CLI
3. **违反Bubble Tea最佳实践**：
   - Bubble Tea应用应该是完全交互式的
   - 所有功能应该在TUI内完成
   - CLI参数应该极少且目的明确

4. **功能重复**：很多CLI命令的功能可以/应该在TUI中实现

## Bubble Tea最佳实践

参考优秀的Bubble Tea应用（如 lazygit, lazydocker, glow）：

- **TUI是主要界面** - 用户99%的时间在TUI中操作
- **CLI参数极少** - 通常只有 --help, --version, --config 等
- **非交互式操作分离** - 如果需要脚本化，提供独立的命令
- **一致的用户体验** - 不要在TUI和CLI之间切换

## 重新设计方案

### 方案A：TUI优先（推荐）

```bash
# 核心命令
boba                    # 启动TUI dashboard（默认行为）
boba --help, -h         # 显示帮助
boba --version, -v      # 显示版本信息

# 初始化和诊断（保留CLI）
boba init               # 初始化配置（首次运行）
boba doctor             # 系统诊断（适合脚本/CI）

# 非交互式操作（保留CLI）
boba run <tool> [args]  # 运行绑定的工具（非交互式）
boba call --profile <p> --data @file.json  # API调用（脚本用）

# 查询命令（保留CLI，但简化）
boba stats [--today|--7d|--30d]  # 快速查看统计
boba version            # 版本详情
```

### 移入TUI的功能

以下功能应该完全在TUI中实现：

1. **Provider管理** - `boba providers` → TUI Provider管理页面
2. **Tool管理** - `boba tools` → TUI Tool管理页面
3. **Binding管理** - `boba bind` → TUI Binding页面
4. **Secrets管理** - `boba secrets` → TUI Secrets管理页面（安全输入）
5. **Proxy管理** - `boba proxy` → TUI Proxy控制面板
6. **Profile管理** - `boba use`, `boba ls --profiles` → TUI Profile页面
7. **Budget配置** - `boba budget` → TUI Budget页面
8. **Routing测试** - `boba route test` → TUI Routing测试器
9. **Hooks管理** - `boba hooks` → TUI Hooks页面
10. **配置编辑** - `boba edit` → TUI配置编辑器
11. **Suggestions** - `boba action`, `boba suggest` → TUI Suggestions页面
12. **Reports** - `boba report` → TUI Report生成器

### 新的TUI布局

```
┌──────────────────────── BobaMixer Dashboard ────────────────────────┐
│                                                                      │
│  [Dashboard] [Providers] [Tools] [Bindings] [Secrets] [Stats] ...  │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │                                                            │    │
│  │              当前页面内容                                    │
│  │                                                            │    │
│  │                                                            │    │
│  └────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  q:退出 Tab:切换页面 ?:帮助 r:刷新                                  │
└──────────────────────────────────────────────────────────────────────┘
```

### TUI导航结构

```
Main Views:
├── Dashboard (默认) - 概览
├── Control Plane
│   ├── Providers - 管理AI提供商
│   ├── Tools - 管理CLI工具
│   ├── Bindings - 工具↔Provider绑定
│   ├── Secrets - 安全管理API密钥
│   └── Proxy - 代理服务器控制
├── Usage & Stats
│   ├── Today - 今日统计
│   ├── Trends - 趋势分析
│   ├── Sessions - 会话历史
│   └── Reports - 生成报告
├── Budget & Optimization
│   ├── Budget Status - 预算状态
│   ├── Suggestions - 优化建议
│   └── Actions - 应用建议
├── Configuration
│   ├── Profiles - Profile管理
│   ├── Routing - 路由规则测试
│   ├── Hooks - Git hooks管理
│   └── Settings - 全局设置
└── Help & Diagnostics
    ├── Help - 使用帮助
    └── Doctor - 系统诊断
```

## 新的Help输出

```
BobaMixer - AI CLI Control Plane

Usage:
  boba                    Launch interactive TUI dashboard
  boba --help             Show this help message
  boba --version          Show version information

Setup & Diagnostics:
  boba init               Initialize ~/.boba configuration
  boba doctor             Run system diagnostics

Non-Interactive Commands:
  boba run <tool> [args]  Run a bound CLI tool
  boba call --profile <p> --data @file  Execute an API call

Quick Stats:
  boba stats [--today|--7d|--30d]  Show usage statistics

All other features are available in the interactive TUI.
Launch with 'boba' to explore:
  • Manage providers, tools, and bindings
  • Configure secrets and proxy settings
  • View detailed statistics and trends
  • Set budgets and apply optimizations
  • Test routing rules and manage hooks

For more information: https://royisme.github.io/BobaMixer/
```

## 实施计划

### Phase 1: 核心重构
1. ✅ 简化 `printUsage()` - 只显示核心命令
2. 🔄 增强TUI - 添加缺失的管理页面
3. 🔄 移除冗余CLI命令 - 或标记为deprecated

### Phase 2: TUI增强
1. 添加Provider管理页面
2. 添加Tool管理页面
3. 添加Binding管理页面
4. 添加Secrets管理页面（安全输入）
5. 添加Proxy控制面板

### Phase 3: 高级功能
1. TUI内的Routing测试器
2. TUI内的配置编辑器
3. TUI内的Report生成器
4. 完整的键盘导航和快捷键

## 优势

1. **用户体验一致** - 所有交互都在TUI中，学习成本低
2. **符合最佳实践** - 遵循Bubble Tea和现代TUI应用设计
3. **更强大** - TUI可以提供更丰富的交互（表单、选择器、实时更新）
4. **更安全** - 密钥管理在TUI中更安全（不会出现在shell历史）
5. **脚本友好** - 保留必要的非交互式命令用于自动化

## 参考

优秀的Bubble Tea应用：
- **lazygit** - 完全TUI，CLI参数极少
- **lazydocker** - 完全TUI
- **glow** - TUI阅读器，CLI用于快速查看
- **soft-serve** - Git服务器TUI

## 总结

当前设计试图"两全其美"，但实际上造成了混乱。应该**明确BobaMixer是一个TUI应用**，CLI只是补充。所有交互式功能都应该在TUI中完成，CLI只保留必要的非交互式操作。
