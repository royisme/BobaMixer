# Getting Started

## Prerequisites

Before installing BobaMixer, ensure you have:

- **Operating System**: macOS or Linux (amd64/arm64)
- **Go** (optional): Go 1.22+ if building from source
- **Git** (optional): For cloning the repository

## Installation

Choose one of the following installation methods:

### Homebrew (macOS/Linux)

```bash
brew install royisme/tap/boba
```

### Go Install

If you have Go installed:

```bash
go install github.com/royisme/BobaMixer/cmd/boba@latest
```

### Download Binary

Download pre-built binaries from the [releases page](https://github.com/royisme/BobaMixer/releases):

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

### Build from Source

```bash
git clone https://github.com/royisme/BobaMixer.git
cd BobaMixer
make build
sudo cp bin/boba /usr/local/bin/
```

## Verify Installation

Confirm BobaMixer is properly installed:

```bash
boba version
```

You should see output similar to:

```
BobaMixer version 0.1.0
```

## Initialize Configuration

BobaMixer stores configuration in `~/.boba/`. Initialize with:

```bash
boba init
```

This creates:

```
~/.boba/
├── profiles.yaml     # AI provider configurations
├── routes.yaml       # Routing rules
├── secrets.yaml      # API keys and secrets
├── pricing.yaml      # Pricing information
└── usage.db          # SQLite database for tracking
```

## Configure Your First Profile

Edit `~/.boba/profiles.yaml` to add your first AI provider:

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

## Add Your API Key

Store your OpenAI API key in `~/.boba/secrets.yaml`:

```yaml
secrets:
  OPENAI_API_KEY: "sk-your-api-key-here"
```

Ensure appropriate permissions:

```bash
chmod 600 ~/.boba/secrets.yaml
```

## Test Your Configuration

Run your first prompt:

```bash
boba ask --profile gpt4-mini "What is the capital of France?"
```

Expected output:

```
The capital of France is Paris.

[Usage] Tokens: 25 in, 8 out | Cost: $0.000009 | Latency: 842ms
```

## Set Default Profile

Set a default profile to avoid specifying `--profile` each time:

```bash
boba use gpt4-mini
```

Now you can run:

```bash
boba ask "Tell me a joke"
```

## Enable Shell Completion (Optional)

### Bash

```bash
# Add to ~/.bashrc
source <(boba completion bash)
```

### Zsh

```bash
# Add to ~/.zshrc
source <(boba completion zsh)
```

### Fish

```bash
# Add to ~/.config/fish/config.fish
boba completion fish | source
```

## Next Steps

Now that you have BobaMixer installed and configured:

1. [Set up routing rules](/ROUTING_COOKBOOK) to automatically select profiles based on context
2. [Configure budgets](/en/configuration#budget-control) to track and limit spending
3. [Add more providers](/ADAPTERS) to optimize costs
4. Explore the TUI for visual analytics

## Common Issues

### Command Not Found

If `boba` is not found after installation:

1. Check if `/usr/local/bin` is in your `PATH`:
   ```bash
   echo $PATH
   ```

2. Add to `PATH` if needed:
   ```bash
   export PATH="/usr/local/bin:$PATH"
   ```

### Permission Denied

If you encounter permission errors:

```bash
chmod +x /usr/local/bin/boba
```

### Configuration Errors

Validate your configuration:

```bash
boba doctor
```

This checks for common configuration issues and suggests fixes.

## Getting Help

If you need assistance:

- Run `boba help` for command reference
- Visit our [FAQ](/FAQ)
- Open an issue on [GitHub](https://github.com/royisme/BobaMixer/issues)
