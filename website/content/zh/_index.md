---
title: "BobaMixer"
---

# BobaMixer: 智能AI适配器路由器

<div class="mb-4">
  <a href="/docs/" class="btn btn-primary me-2 mb-2">了解更多</a>
  <a href="https://github.com/royisme/BobaMixer" class="btn btn-secondary me-2 mb-2">下载</a>
</div>

智能AI适配器路由器，支持智能路由、预算追踪和成本优化。

---

## 概述

BobaMixer 是一款强大的命令行工具，管理多个 AI 提供商，追踪成本，并优化您的 AI 工作负载路由。

**核心功能**：多提供商支持、智能路由、实时预算追踪和全面的使用分析。

---

## 核心功能

### 🧠 智能路由
基于上下文、成本和性能，使用 epsilon-greedy 探索算法将提示路由到最佳 AI 提供商。

### 📊 预算追踪
在全局、项目和配置文件级别追踪成本，支持每日/每月限额和实时警报。

### 🔌 多提供商支持
统一接口支持 HTTP API、命令行工具和 MCP（模型上下文协议）服务器。

---

## 快速开始

按照我们的[快速入门指南](/docs/getting-started/)安装 BobaMixer 并设置您的第一个 AI 提供商。

## 安装与使用

### 📱 简易安装
通过 Homebrew、Go 安装，或下载 macOS 和 Linux 的预构建二进制文件。

```bash
brew install royisme/tap/boba
```

### ⚙️ 灵活配置
简单的 YAML 配置，支持配置文件、路由规则、密钥和定价。

### 📈 实时监控
漂亮的 TUI 仪表板，显示使用情况、成本和性能指标。
