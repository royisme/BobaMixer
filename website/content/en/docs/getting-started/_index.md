---
title: "5分钟快速上手"
linkTitle: "快速上手"
weight: 1
description: >
  5分钟内完成BobaMixer安装配置，立即体验智能AI路由的威力。
---

# 🚀 5分钟快速上手

让多个AI提供商为你智能协作，无需手动切换，自动成本优化。

---

## ⏱️ 时间预估
- **安装**: 1分钟
- **配置**: 2分钟  
- **第一次使用**: 2分钟

---

## 📥 第一步：安装 BobaMixer

选择最适合你的安装方式：

### 🍺 Homebrew (推荐 - 最快)

```bash
brew install royisme/tap/boba
```

### 🔧 Go 安装

```bash
go install github.com/royisme/BobaMixer/cmd/boba@latest
```

### 📦 下载二进制文件

```bash
# macOS (Intel芯片)
curl -LO https://github.com/royisme/BobaMixer/releases/latest/download/boba_darwin_amd64.tar.gz
tar -xzf boba_darwin_amd64.tar.gz && sudo mv boba /usr/local/bin/

# macOS (Apple芯片) 
curl -LO https://github.com/royisme/BobaMixer/releases/latest/download/boba_darwin_arm64.tar.gz
tar -xzf boba_darwin_arm64.tar.gz && sudo mv boba /usr/local/bin/

# Linux
curl -LO https://github.com/royisme/BobaMixer/releases/latest/download/boba_linux_amd64.tar.gz
tar -xzf boba_linux_amd64.tar.gz && sudo mv boba /usr/local/bin/
```

### ✅ 验证安装

```bash
boba version
# BobaMixer version 0.1.0
```

> **💡 小贴士**: 如果提示 `command not found`，请确保 `/usr/local/bin` 在你的 `PATH` 中。

---

## ⚙️ 第二步：配置第一个AI提供商

### 初始化配置

```bash
boba init
```

这会在 `~/.boba/` 创建配置文件：
```
~/.boba/
├── profiles.yaml     # AI提供商配置
├── routes.yaml       # 智能路由规则
├── secrets.yaml      # API密钥（安全存储）
├── pricing.yaml      # 价格信息
└── usage.db          # 使用追踪数据库
```

### 配置OpenAI (最常见的开始)

编辑 `~/.boba/profiles.yaml`：

```yaml
default_profile: gpt4-mini

profiles:
  - key: gpt4-mini
    model: gpt-4o-mini
    adapter: http
    http:
      endpoint: https://api.openai.com/v1/chat/completions
      method: POST
      headers:
        Authorization: "Bearer {{secret://OPENAI_API_KEY}}"
        Content-Type: application/json
      body_template: |
        {
          "model": "{{.Model}}",
          "messages": [{"role": "user", "content": "{{.Text}}"}]
        }
      response_path: choices.0.message.content
    cost_per_1k_input: 0.00015
    cost_per_1k_output: 0.0006

  # 可以同时配置多个提供商
  - key: claude-sonnet
    model: claude-3-5-sonnet-20241022
    adapter: http
    http:
      endpoint: https://api.anthropic.com/v1/messages
      method: POST
      headers:
        x-api-key: "{{secret://ANTHROPIC_API_KEY}}"
        anthropic-version: "2023-06-01"
        Content-Type: application/json
      body_template: |
        {
          "model": "{{.Model}}",
          "max_tokens": 4096,
          "messages": [{"role": "user", "content": "{{.Text}}"}]
        }
      response_path: content.0.text
    cost_per_1k_input: 0.003
    cost_per_1k_output: 0.015
```

### 添加API密钥

编辑 `~/.boba/secrets.yaml`（**不要提交到git**）：

```yaml
secrets:
  OPENAI_API_KEY: "sk-your-openai-key-here"
  ANTHROPIC_API_KEY: "sk-ant-your-anthropic-key-here"
```

**重要**: 设置安全的文件权限：
```bash
chmod 600 ~/.boba/secrets.yaml
```

---

## 🎯 第三步：第一次体验

### 设置默认profile（可选）

```bash
boba use gpt4-mini
```

### 开始使用！

```bash
# 简单对话
boba ask "写一个Python的hello world"

# 代码相关任务
boba ask "帮我优化这个递归函数的性能"

# 分析任务
boba ask "分析这份用户反馈数据的主要问题"
```

**你会看到类似输出**：
```
Here's a simple Python Hello World program:

```python
print("Hello, World!")
```

[Usage] Tokens: 23 in, 15 out | Cost: $0.000006 | Latency: 523ms | Profile: gpt4-mini
```

---

## 🧠 智能路由初体验

编辑 `~/.boba/routes.yaml` 配置智能路由：

```yaml
routes:
  # 复杂分析任务使用更强的模型
  - id: complex-analysis
    match:
      text_matches: "分析|性能|优化|架构"
    profile: claude-sonnet
    explain: "复杂分析任务使用Claude"

  # 简单任务使用经济模型
  - id: simple-tasks  
    match:
      ctx_chars_lt: 2000
      intent: format
    profile: gpt4-mini
    explain: "简单任务使用GPT-4o-mini"

# 默认fallback
fallback: gpt4-mini
```

现在试试不同的任务：

```bash
# 会自动路由到 claude-sonnet
boba ask "分析这个系统架构的性能瓶颈"

# 会自动路由到 gpt4-mini  
boba ask "格式化这个JSON"
```

---

## 📊 查看使用统计

```bash
# 查看今天的使用情况
boba stats

# 查看详细分析
boba analytics

# 获取成本优化建议
boba suggest
```

---

## 🎯 常见使用场景

### 开发工作流

```bash
# 在项目目录下，BobaMixer会自动识别项目类型
cd ~/projects/my-app

# 代码相关 - 自动选择适合的模型
boba ask "为这个React组件写单元测试"

# 文档任务
boba ask "写API文档说明"
```

### 从文件输入

```bash
# 分析日志文件
boba ask "分析这个错误日志" < error.log

# 代码审查
git diff | boba ask "审查这个代码变更"
```

### 管道操作

```bash
# 组合命令
cat config.yaml | boba ask "验证这个配置文件"

# 多步骤处理
ls -la | boba ask "整理成markdown表格"
```

---

## 🆘 常见问题快速解决

### 命令找不到？
```bash
# 检查PATH
echo $PATH | grep -o "/usr/local/bin"

# 手动添加（临时）
export PATH="/usr/local/bin:$PATH"

# 永久添加到 ~/.zshrc 或 ~/.bashrc
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### API密钥错误？
```bash
# 验证配置
boba doctor

# 测试连接
boba ask --profile gpt4-mini "test"
```

### 权限问题？
```bash
# 设置正确的权限
chmod 600 ~/.boba/secrets.yaml
chmod +x /usr/local/bin/boba
```

---

## 🎉 恭喜！你已经完成了基础配置

现在你可以：
- ✅ 使用多个AI提供商
- ✅ 自动智能路由选择
- ✅ 实时成本追踪
- ✅ 项目级别管理

---

## 🚀 下一步学习

### 想要更好用？
- **[配置路由规则](/docs/routing/)** - 让AI选择更智能
- **[设置预算管理](/docs/budgets/)** - 控制成本，避免超支
- **[添加更多AI服务](/docs/adapters/)** - 支持更多AI提供商

### 想要高级功能？
- **[企业团队使用](/docs/enterprise/)** - 团队协作和管理
- **[性能优化](/docs/performance/)** - 大规模使用最佳实践
- **[自定义开发](/docs/development/)** - 扩展和定制

---

## 💡 实用小贴士

### 提高效率
```bash
# 启用shell补全
echo 'source <(boba completion zsh)' >> ~/.zshrc  # zsh
echo 'source <(boba completion bash)' >> ~/.bashrc  # bash

# 创建别名
echo 'alias ba="boba ask"' >> ~/.zshrc
```

### 项目级别配置
```bash
# 在项目根目录创建项目配置
echo "default_profile: claude-sonnet" > .boba-project.yaml
echo "daily_budget: 20" >> .boba-project.yaml
```

### 快速检查
```bash
# 检查配置健康状态
boba doctor

# 查看所有可用的profile
boba profiles list

# 测试路由规则
boba route test "分析这个算法"
```

> **🎯 开始你的智能AI之旅！**  
> 现在你拥有了管理多个AI的超级能力，接下来去探索更多可能性吧！
