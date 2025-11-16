package runner

import (
	"fmt"

	"github.com/royisme/bobamixer/internal/domain/core"
)

// GeminiRunner handles running Gemini CLI with proper configuration
type GeminiRunner struct {
	BaseRunner
}

// Prepare prepares the environment for Gemini
func (g *GeminiRunner) Prepare(ctx *RunContext) error {
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
	case core.ProviderKindGemini:
		// Official Google Gemini provider
		// Gemini supports both GEMINI_API_KEY and GOOGLE_API_KEY
		// Set both for maximum compatibility
		ctx.Env["GEMINI_API_KEY"] = apiKey
		ctx.Env["GOOGLE_API_KEY"] = apiKey

		// Only set base URL if it's not the default
		if ctx.Provider.BaseURL != "" && ctx.Provider.BaseURL != "https://generativelanguage.googleapis.com/v1" {
			ctx.Env["GEMINI_BASE_URL"] = ctx.Provider.BaseURL
		}

	default:
		return fmt.Errorf("unsupported provider kind for Gemini: %s", ctx.Provider.Kind)
	}

	// Set model if specified in binding options
	if ctx.Binding.Options.Model != "" {
		// Gemini CLI uses GEMINI_MODEL environment variable
		ctx.Env["GEMINI_MODEL"] = ctx.Binding.Options.Model
	} else if ctx.Provider.DefaultModel != "" {
		// Use provider's default model
		ctx.Env["GEMINI_MODEL"] = ctx.Provider.DefaultModel
	}

	// Handle model mapping if specified
	if ctx.Binding.Options.ModelMapping != nil {
		// For Gemini, we might want to set tier-based models
		// This could be used for different model tiers (e.g., flash, pro, ultra)
		for tier, model := range ctx.Binding.Options.ModelMapping {
			envVar := fmt.Sprintf("GEMINI_%s_MODEL", tier)
			ctx.Env[envVar] = model
		}
	}

	// TODO: If use_proxy is true, modify base URL to point to local proxy
	// For Phase 1.5, we're not implementing proxy yet

	return nil
}

func init() {
	// Register Gemini runner
	Register(core.ToolKindGemini, &GeminiRunner{})
}
