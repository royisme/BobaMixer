---
title: "Documentation"
linkTitle: "Documentation"
weight: 20
menu:
  main:
    weight: 20
---

Welcome to the BobaMixer documentation! This guide will help you get started with BobaMixer, a smart AI adapter router with intelligent routing, budget tracking, and cost optimization.

## What is BobaMixer?

BobaMixer is a powerful command-line tool designed to manage multiple AI providers, track costs, and optimize your AI workload routing. It provides:

- **Intelligent Routing**: Automatically route prompts to the best AI provider based on context, cost, and performance
- **Budget Tracking**: Multi-level budget management (global, project, profile) with real-time alerts
- **Multi-Provider Support**: Unified interface for HTTP APIs, command-line tools, and MCP servers
- **Cost Optimization**: Epsilon-greedy exploration and suggestion engine for cost savings
- **Usage Analytics**: Comprehensive tracking of tokens, costs, latency, and success rates

## Getting Started

New to BobaMixer? Start with our [Getting Started Guide](/docs/getting-started/) to:

1. Install BobaMixer on your system
2. Configure your first AI provider
3. Execute your first prompt
4. Set up routing rules and budgets

## Key Concepts

- **Profiles**: AI provider configurations (model, API settings, costs)
- **Adapters**: Connectors to different types of AI services (HTTP, Tool, MCP)
- **Routes**: Rules that determine which profile to use based on context
- **Budgets**: Cost limits at different levels (global, project, profile)
- **Sessions**: Conversation contexts for tracking multi-turn interactions

## Documentation Sections

### [Getting Started](/docs/getting-started/)
Installation, configuration, and first steps

### [User Guide](/docs/user-guide/)
Day-to-day usage, commands, and workflows

### [Configuration](/docs/configuration/)
Detailed configuration reference for all config files

### [Adapters](/docs/adapters/)
Working with different adapter types and custom adapters

### [Routing](/docs/routing/)
Routing rules, patterns, and optimization strategies

### [Troubleshooting](/docs/troubleshooting/)
Common issues and solutions

### [Development](/docs/development/)
Contributing, building from source, and extending BobaMixer
