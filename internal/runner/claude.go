package runner

import (
	"fmt"

	"github.com/royisme/bobamixer/internal/domain/core"
)

// ClaudeRunner handles running Claude Code CLI with proper configuration
type ClaudeRunner struct {
	BaseRunner
}

// Prepare prepares the environment for Claude
func (c *ClaudeRunner) Prepare(ctx *RunContext) error {
	if ctx.Env == nil {
		ctx.Env = make(map[string]string)
	}

	// Get API key
	apiKey, err := ResolveAPIKey(ctx.Provider, ctx.Secrets)
	if err != nil {
		return fmt.Errorf("failed to resolve API key: %w", err)
	}

	// Determine which environment variables to set based on provider kind
	switch ctx.Provider.Kind {
	case core.ProviderKindAnthropic:
		// Official Anthropic provider
		// Use ANTHROPIC_API_KEY and official base URL
		ctx.Env["ANTHROPIC_API_KEY"] = apiKey
		// Only set base URL if it's not the default
		if ctx.Provider.BaseURL != "https://api.anthropic.com" {
			ctx.Env["ANTHROPIC_BASE_URL"] = ctx.Provider.BaseURL
		}

	case core.ProviderKindAnthropicCompatible:
		// Anthropic-compatible provider (e.g., Z.AI)
		// Use ANTHROPIC_AUTH_TOKEN and custom base URL
		ctx.Env["ANTHROPIC_AUTH_TOKEN"] = apiKey
		ctx.Env["ANTHROPIC_BASE_URL"] = ctx.Provider.BaseURL

		// For Z.AI specifically, might need additional env vars
		if ctx.Provider.ID == "claude-zai" {
			// Z.AI uses same token
			ctx.Env["ANTHROPIC_API_KEY"] = apiKey
		}

	default:
		return fmt.Errorf("unsupported provider kind for Claude: %s", ctx.Provider.Kind)
	}

	// Handle model mapping if specified in binding options
	if ctx.Binding.Options.ModelMapping != nil {
		// Set default model env vars based on mapping
		for tier, model := range ctx.Binding.Options.ModelMapping {
			envVar := fmt.Sprintf("ANTHROPIC_DEFAULT_%s_MODEL", tier)
			ctx.Env[envVar] = model
		}
	}

	// Set default model if specified
	if ctx.Binding.Options.Model != "" {
		// This could be used to override the default model
		// Claude CLI might not have a direct env var for this, but we can set it
		ctx.Env["ANTHROPIC_MODEL"] = ctx.Binding.Options.Model
	}

	// Handle proxy mode
	if ctx.Binding.UseProxy {
		// Route requests through local proxy
		ctx.Env["ANTHROPIC_BASE_URL"] = "http://127.0.0.1:7777/anthropic/v1"
		// Preserve the API key for proxy authentication
		// The proxy will forward it to the actual provider
	}

	return nil
}

func init() {
	// Register Claude runner
	Register(core.ToolKindClaude, &ClaudeRunner{})
}
