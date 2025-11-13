---
layout: home

hero:
  name: "BobaMixer"
  text: "Smart AI Adapter Router"
  tagline: Intelligent routing, budget tracking, and cost optimization for multiple AI providers
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
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
  - icon: 📈
    title: Usage Analytics
    details: Comprehensive tracking with historical trends, P95 latency comparisons, and detailed cost breakdowns
  - icon: 🤖
    title: Smart Suggestions
    details: AI-powered cost optimization recommendations based on your actual usage patterns
  - icon: 🎨
    title: Beautiful TUI
    details: Interactive dashboard with real-time stats, trend visualizations, and profile switching
---

## Quick Example

```bash
# Install
go install github.com/royisme/bobamixer/cmd/boba@latest

# Initialize configuration
boba doctor

# Launch TUI dashboard
boba

# Track usage
boba stats --7d
```

## Why BobaMixer?

### Cost Control
Never overspend on AI providers again. Set budgets at multiple levels and get proactive alerts before hitting limits.

### Smart Routing
Automatically route requests to the most cost-effective provider based on context, without sacrificing quality.

### Complete Visibility
Track every token, every dollar, every millisecond. Understand exactly where your AI spending goes.

## Getting Started

<div class="vp-doc" style="margin-top: 2rem;">
  <div class="custom-block tip">
    <p class="custom-block-title">Quick Start</p>
    <p>New to BobaMixer? Start with our <a href="/guide/getting-started">Getting Started Guide</a> to get up and running in minutes.</p>
  </div>
</div>

## Community

- [GitHub Repository](https://github.com/royisme/BobaMixer)
- [Issues & Bugs](https://github.com/royisme/BobaMixer/issues)
- [Contributing Guide](https://github.com/royisme/BobaMixer/blob/main/CONTRIBUTING.md)

---

**Made with ☕ and 🧋 for developers, by developers**
