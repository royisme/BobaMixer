---
title: "BobaMixer"
---

{{< blocks/cover title="BobaMixer: 智能AI适配器路由器" image_anchor="top" height="full" >}}
<a class="btn btn-lg btn-primary me-3 mb-4" href="/docs/">
  了解更多 <i class="fas fa-arrow-alt-circle-right ms-2"></i>
</a>
<a class="btn btn-lg btn-secondary me-3 mb-4" href="https://github.com/royisme/BobaMixer">
  下载 <i class="fab fa-github ms-2 "></i>
</a>
<p class="lead mt-5">智能AI适配器路由器，支持智能路由、预算追踪和成本优化。</p>
{{< blocks/link-down color="info" >}}
{{< /blocks/cover >}}


{{% blocks/lead color="primary" %}}
BobaMixer 是一款强大的命令行工具，管理多个 AI 提供商，追踪成本，并优化您的 AI 工作负载路由。

**核心功能**：多提供商支持、智能路由、实时预算追踪和全面的使用分析。
{{% /blocks/lead %}}


{{% blocks/section color="dark" type="row" %}}
{{% blocks/feature icon="fa-lightbulb" title="智能路由" %}}
基于上下文、成本和性能，使用 epsilon-greedy 探索算法将提示路由到最佳 AI 提供商。
{{% /blocks/feature %}}


{{% blocks/feature icon="fa-chart-line" title="预算追踪" %}}
在全局、项目和配置文件级别追踪成本，支持每日/每月限额和实时警报。
{{% /blocks/feature %}}


{{% blocks/feature icon="fab fa-github" title="多提供商支持" %}}
统一接口支持 HTTP API、命令行工具和 MCP（模型上下文协议）服务器。
{{% /blocks/feature %}}


{{% /blocks/section %}}


{{% blocks/section %}}
## 准备开始了吗？

按照我们的[快速入门指南](/docs/getting-started/)安装 BobaMixer 并设置您的第一个 AI 提供商。

{{% /blocks/section %}}


{{% blocks/section type="row" %}}

{{% blocks/feature icon="fab fa-app-store-ios" title="简易安装" %}}
通过 Homebrew、Go 安装，或下载 macOS 和 Linux 的预构建二进制文件。
```bash
brew install royisme/tap/boba
```
{{% /blocks/feature %}}


{{% blocks/feature icon="fa-cog" title="灵活配置" %}}
简单的 YAML 配置，支持配置文件、路由规则、密钥和定价。
{{% /blocks/feature %}}


{{% blocks/feature icon="fa-tachometer-alt" title="实时监控" %}}
漂亮的 TUI 仪表板，显示使用情况、成本和性能指标。
{{% /blocks/feature %}}

{{% /blocks/section %}}
