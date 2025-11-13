# 快速上手

## 前提条件

在安装 BobaMixer 之前，请确保您有：

- **操作系统**：macOS 或 Linux（amd64/arm64）
- **Go**（可选）：如果从源代码构建，需要 Go 1.22+
- **Git**（可选）：用于克隆仓库

## 安装

选择以下安装方法之一：

### Homebrew（macOS/Linux）

```bash
brew install royisme/tap/boba
```

### Go 安装

如果您已安装 Go：

```bash
go install github.com/royisme/BobaMixer/cmd/boba@latest
```

### 下载二进制文件

从[发布页面](https://github.com/royisme/BobaMixer/releases)下载预构建的二进制文件：

```bash
# macOS arm64
curl -LO https://github.com/royisme/BobaMixer/releases/download/v0.1.0/boba_darwin_arm64.tar.gz
tar -xzf boba_darwin_arm64.tar.gz
sudo mv boba /usr/local/bin/

# Linux amd64
curl -LO https://github.com/royisme/BobaMixer/releases/download/v0.1.0/boba_linux_amd64.tar.gz
tar -xzf boba_linux_amd64.tar.gz
sudo mv boba /usr/local/bin/
```

### 从源代码构建

```bash
git clone https://github.com/royisme/BobaMixer.git
cd BobaMixer
make build
sudo cp bin/boba /usr/local/bin/
```

## 验证安装

确认 BobaMixer 已正确安装：

```bash
boba version
```

您应该看到类似的输出：

```
BobaMixer version 0.1.0
```

## 初始化配置

BobaMixer 将配置存储在 `~/.boba/` 中。使用以下命令初始化：

```bash
boba init
```

这将创建：

```
~/.boba/
├── profiles.yaml     # AI 提供商配置
├── routes.yaml       # 路由规则
├── secrets.yaml      # API 密钥和秘密
├── pricing.yaml      # 定价信息
└── usage.db          # 用于追踪的 SQLite 数据库
```

## 配置您的第一个配置文件

编辑 `~/.boba/profiles.yaml` 添加您的第一个 AI 提供商：

```yaml
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
```

## 添加您的 API 密钥

在 `~/.boba/secrets.yaml` 中存储您的 OpenAI API 密钥：

```yaml
secrets:
  OPENAI_API_KEY: "sk-your-api-key-here"
```

确保适当的权限：

```bash
chmod 600 ~/.boba/secrets.yaml
```

## 测试您的配置

运行您的第一个提示：

```bash
boba ask --profile gpt4-mini "法国的首都是什么？"
```

预期输出：

```
法国的首都是巴黎。

[Usage] Tokens: 25 in, 8 out | Cost: $0.000009 | Latency: 842ms
```

## 设置默认配置文件

设置默认配置文件以避免每次都指定 `--profile`：

```bash
boba use gpt4-mini
```

现在您可以运行：

```bash
boba ask "讲个笑话"
```

## 启用 Shell 补全（可选）

### Bash

```bash
# 添加到 ~/.bashrc
source <(boba completion bash)
```

### Zsh

```bash
# 添加到 ~/.zshrc
source <(boba completion zsh)
```

### Fish

```bash
# 添加到 ~/.config/fish/config.fish
boba completion fish | source
```

## 下一步

现在您已经安装并配置了 BobaMixer：

1. [设置路由规则](/ROUTING_COOKBOOK) 以基于上下文自动选择配置文件
2. [配置预算](/zh/configuration#预算控制配置) 以追踪和限制支出
3. [添加更多提供商](/ADAPTERS) 以优化成本
4. 探索 TUI 以进行可视化分析

## 常见问题

### 找不到命令

如果安装后找不到 `boba`：

1. 检查 `/usr/local/bin` 是否在您的 `PATH` 中：
   ```bash
   echo $PATH
   ```

2. 如果需要，添加到 `PATH`：
   ```bash
   export PATH="/usr/local/bin:$PATH"
   ```

### 权限被拒绝

如果您遇到权限错误：

```bash
chmod +x /usr/local/bin/boba
```

### 配置错误

验证您的配置：

```bash
boba doctor
```

这将检查常见的配置问题并提供修复建议。

## 获取帮助

如果您需要帮助：

- 运行 `boba help` 查看命令参考
- 访问我们的[常见问题](/FAQ)
- 在 [GitHub](https://github.com/royisme/BobaMixer/issues) 上提出问题
