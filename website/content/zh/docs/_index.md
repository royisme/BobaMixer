---
title: "文档"
linkTitle: "文档"
weight: 20
menu:
  main:
    weight: 20
---

欢迎来到 BobaMixer 文档！本指南将帮助您开始使用 BobaMixer，一个具有智能路由、预算追踪和成本优化的智能 AI 适配器路由器。

## 什么是 BobaMixer？

BobaMixer 是一款强大的命令行工具，旨在管理多个 AI 提供商、追踪成本并优化您的 AI 工作负载路由。它提供：

- **智能路由**：基于上下文、成本和性能自动将提示路由到最佳 AI 提供商
- **预算追踪**：多级预算管理（全局、项目、配置文件）并提供实时警报
- **多提供商支持**：统一接口支持 HTTP API、命令行工具和 MCP 服务器
- **成本优化**：Epsilon-greedy 探索和建议引擎以节省成本
- **使用分析**：全面追踪令牌、成本、延迟和成功率

## 快速入门

刚接触 BobaMixer？从我们的[快速入门指南](/docs/getting-started/)开始：

1. 在您的系统上安装 BobaMixer
2. 配置您的第一个 AI 提供商
3. 执行您的第一个提示
4. 设置路由规则和预算

## 核心概念

- **配置文件（Profiles）**：AI 提供商配置（模型、API 设置、成本）
- **适配器（Adapters）**：连接到不同类型 AI 服务的连接器（HTTP、Tool、MCP）
- **路由（Routes）**：基于上下文确定使用哪个配置文件的规则
- **预算（Budgets）**：不同级别的成本限制（全局、项目、配置文件）
- **会话（Sessions）**：用于追踪多轮交互的对话上下文

## 文档章节

### [快速入门](/docs/getting-started/)
安装、配置和第一步

### [用户指南](/docs/user-guide/)
日常使用、命令和工作流程

### [配置](/docs/configuration/)
所有配置文件的详细配置参考

### [适配器](/docs/adapters/)
使用不同的适配器类型和自定义适配器

### [路由](/docs/routing/)
路由规则、模式和优化策略

### [故障排除](/docs/troubleshooting/)
常见问题和解决方案

### [开发](/docs/development/)
贡献、从源代码构建和扩展 BobaMixer
