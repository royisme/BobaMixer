---
layout: home

hero:
  name: "BobaMixer"
  text: "Smart AI Adapter Router"
  tagline: A smart AI adapter router with intelligent routing, budget tracking, and cost optimization
  actions:
    - theme: brand
      text: Get Started
      link: /en/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/royisme/BobaMixer

features:
  - icon: 🧠
    title: Intelligent Routing
    details: Route prompts to the best AI provider based on context, cost, and performance with epsilon-greedy exploration
  - icon: 📊
    title: Budget Tracking
    details: Track costs at global, project, and profile levels with daily/monthly limits and real-time alerts
  - icon: 🔌
    title: Multi-Provider Support
    details: Unified interface for HTTP APIs, command-line tools, and MCP (Model Context Protocol) servers
  - icon: 📱
    title: Easy Installation
    details: Install via Homebrew, Go, or download pre-built binaries for macOS and Linux
  - icon: ⚙️
    title: Flexible Configuration
    details: Simple YAML configuration for profiles, routing rules, secrets, and pricing
  - icon: 📈
    title: Real-time Monitoring
    details: Beautiful TUI dashboard showing usage, costs, and performance metrics
---

## Quick Start

### Installation

```bash
# Homebrew (Recommended)
brew install royisme/tap/boba

# Go Install
go install github.com/royisme/BobaMixer/cmd/boba@latest
```

### Basic Usage

```bash
# Initialize configuration
boba init

# Ask a question
boba ask "Write a hello world in Python"

# View usage statistics
boba stats
```

## Key Features

### 🧠 Intelligent Routing

BobaMixer automatically routes your prompts to the most appropriate AI provider based on:
- Context and complexity
- Cost optimization
- Performance requirements
- Custom routing rules

### 📊 Budget Management

Keep your AI costs under control with:
- Global, project, and profile-level budgets
- Daily and monthly limits
- Real-time cost tracking
- Usage analytics and suggestions

### 🔌 Flexible Integration

Connect to any AI service:
- HTTP REST APIs (OpenAI, Anthropic, etc.)
- Command-line tools (Ollama, local models)
- MCP (Model Context Protocol) servers
- Custom adapters

## Documentation

- [Quick Start Guide](/en/getting-started) - Get up and running in 5 minutes
- [Configuration Guide](/en/configuration) - Learn how to configure BobaMixer
- [Adapters](/ADAPTERS) - Connect to different AI providers
- [Routing Cookbook](/ROUTING_COOKBOOK) - Advanced routing strategies
- [FAQ](/FAQ) - Frequently asked questions

## Community

- [GitHub Repository](https://github.com/royisme/BobaMixer)
- [Issues & Bugs](https://github.com/royisme/BobaMixer/issues)
- [Contributing Guide](https://github.com/royisme/BobaMixer/blob/main/CONTRIBUTING.md)
