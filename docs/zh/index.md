---
layout: home

hero:
  name: "BobaMixer"
  text: "智能AI适配器路由器"
  tagline: 智能AI适配器路由器，支持智能路由、预算追踪和成本优化
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/getting-started
    - theme: alt
      text: 在 GitHub 上查看
      link: https://github.com/royisme/BobaMixer

features:
  - icon: 🧠
    title: 智能路由
    details: 基于上下文、成本和性能，使用 epsilon-greedy 探索算法将提示路由到最佳 AI 提供商
  - icon: 📊
    title: 预算追踪
    details: 在全局、项目和配置文件级别追踪成本，支持每日/每月限额和实时警报
  - icon: 🔌
    title: 多提供商支持
    details: 统一接口支持 HTTP API、命令行工具和 MCP（模型上下文协议）服务器
  - icon: 📱
    title: 简易安装
    details: 通过 Homebrew、Go 安装，或下载 macOS 和 Linux 的预构建二进制文件
  - icon: ⚙️
    title: 灵活配置
    details: 简单的 YAML 配置，支持配置文件、路由规则、密钥和定价
  - icon: 📈
    title: 实时监控
    details: 漂亮的 TUI 仪表板，显示使用情况、成本和性能指标
---

## 快速开始

### 安装

```bash
# Homebrew (推荐)
brew install royisme/tap/boba

# Go 安装
go install github.com/royisme/BobaMixer/cmd/boba@latest
```

### 基本使用

```bash
# 初始化配置
boba init

# 提问
boba ask "用 Python 写一个 hello world"

# 查看使用统计
boba stats
```

## 核心功能

### 🧠 智能路由

BobaMixer 会根据以下因素自动将您的提示路由到最合适的 AI 提供商：
- 上下文和复杂度
- 成本优化
- 性能要求
- 自定义路由规则

### 📊 预算管理

通过以下功能控制您的 AI 成本：
- 全局、项目和配置文件级别的预算
- 每日和每月限额
- 实时成本追踪
- 使用分析和建议

### 🔌 灵活集成

连接到任何 AI 服务：
- HTTP REST API（OpenAI、Anthropic 等）
- 命令行工具（Ollama、本地模型）
- MCP（模型上下文协议）服务器
- 自定义适配器

## 文档

- [快速上手指南](/zh/getting-started) - 5 分钟快速上手
- [配置指南](/zh/configuration) - 学习如何配置 BobaMixer
- [适配器](/ADAPTERS) - 连接到不同的 AI 提供商
- [路由手册](/ROUTING_COOKBOOK) - 高级路由策略
- [常见问题](/FAQ) - 常见问题解答

## 社区

- [GitHub 仓库](https://github.com/royisme/BobaMixer)
- [问题 & Bug](https://github.com/royisme/BobaMixer/issues)
- [贡献指南](https://github.com/royisme/BobaMixer/blob/main/CONTRIBUTING.md)
