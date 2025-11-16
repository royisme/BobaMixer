package runner

import (
	"fmt"

	"github.com/royisme/bobamixer/internal/domain/core"
)

// OpenAIRunner handles running OpenAI Codex CLI with proper configuration
type OpenAIRunner struct {
	BaseRunner
}

// Prepare prepares the environment for OpenAI Codex
func (o *OpenAIRunner) Prepare(ctx *RunContext) error {
	if ctx.Env == nil {
		ctx.Env = make(map[string]string)
	}

	// Get API key
	apiKey, err := ResolveAPIKey(ctx.Provider, ctx.Secrets)
	if err != nil {
		return fmt.Errorf("failed to resolve API key: %w", err)
	}

	// Set environment variables based on provider kind
	switch ctx.Provider.Kind {
	case core.ProviderKindOpenAI:
		// Official OpenAI provider
		ctx.Env["OPENAI_API_KEY"] = apiKey

		// Only set base URL if it's not the default
		if ctx.Provider.BaseURL != "" && ctx.Provider.BaseURL != "https://api.openai.com/v1" {
			ctx.Env["OPENAI_BASE_URL"] = ctx.Provider.BaseURL
		}

	case core.ProviderKindOpenAICompatible:
		// OpenAI-compatible provider (e.g., Azure OpenAI, LocalAI)
		ctx.Env["OPENAI_API_KEY"] = apiKey
		ctx.Env["OPENAI_BASE_URL"] = ctx.Provider.BaseURL

	default:
		return fmt.Errorf("unsupported provider kind for OpenAI: %s", ctx.Provider.Kind)
	}

	// Set model if specified in binding options
	if ctx.Binding.Options.Model != "" {
		// Codex CLI uses OPENAI_MODEL environment variable
		ctx.Env["OPENAI_MODEL"] = ctx.Binding.Options.Model
	} else if ctx.Provider.DefaultModel != "" {
		// Use provider's default model
		ctx.Env["OPENAI_MODEL"] = ctx.Provider.DefaultModel
	}

	// Handle model mapping if specified
	if ctx.Binding.Options.ModelMapping != nil {
		// For OpenAI, we might want to set tier-based models
		// This could be used for different model tiers (e.g., fast, balanced, quality)
		for tier, model := range ctx.Binding.Options.ModelMapping {
			envVar := fmt.Sprintf("OPENAI_%s_MODEL", tier)
			ctx.Env[envVar] = model
		}
	}

	// TODO: If use_proxy is true, modify base URL to point to local proxy
	// For Phase 1.5, we're not implementing proxy yet

	return nil
}

func init() {
	// Register OpenAI/Codex runner
	Register(core.ToolKindCodex, &OpenAIRunner{})
}
