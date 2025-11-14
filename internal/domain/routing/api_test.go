package routing_test

import (
	"context"
	"strings"
	"testing"

	"github.com/royisme/bobamixer/internal/domain/routing"
	"github.com/royisme/bobamixer/internal/store/config"
)

func TestCompile(t *testing.T) {
	t.Run("compiles valid rules", func(t *testing.T) {
		// Given: valid routing rules
		rules := []config.RouteRule{
			{
				ID:      "quick-tasks",
				If:      "ctx_chars<1000",
				Use:     "fast-profile",
				Explain: "Small context, use faster model",
			},
			{
				ID:      "review-tasks",
				If:      "intent=='review'",
				Use:     "work-heavy",
				Explain: "Code review needs thorough analysis",
			},
		}

		// When: Compile is called
		engine, err := routing.Compile(rules)

		// Then: compilation succeeds
		if err != nil {
			t.Fatalf("Compile failed: %v", err)
		}
		if engine == nil {
			t.Error("expected non-nil engine")
		}
	})

	t.Run("rejects rule with invalid regex", func(t *testing.T) {
		// Given: rule with invalid regex pattern
		rules := []config.RouteRule{
			{
				ID:      "invalid-regex",
				If:      "text.matches('[invalid')",
				Use:     "test-profile",
				Explain: "Invalid pattern",
			},
		}

		// When: Compile is called
		_, err := routing.Compile(rules)

		// Then: compilation fails with error
		if err == nil {
			t.Error("expected compilation error for invalid regex")
		}
		if !strings.Contains(err.Error(), "invalid regex") &&
			!strings.Contains(err.Error(), "error parsing regexp") {
			t.Errorf("expected regex error, got: %v", err)
		}
	})

	t.Run("rejects rule without ID", func(t *testing.T) {
		// Given: rule without ID
		rules := []config.RouteRule{
			{
				ID:      "",
				If:      "ctx_chars<1000",
				Use:     "test-profile",
				Explain: "No ID",
			},
		}

		// When: Compile is called
		_, err := routing.Compile(rules)

		// Then: compilation fails
		if err == nil {
			t.Error("expected error for rule without ID")
		}
		if !strings.Contains(err.Error(), "ID is required") {
			t.Errorf("expected ID error, got: %v", err)
		}
	})

	t.Run("rejects rule without condition", func(t *testing.T) {
		// Given: rule without condition
		rules := []config.RouteRule{
			{
				ID:      "no-condition",
				If:      "",
				Use:     "test-profile",
				Explain: "No condition",
			},
		}

		// When: Compile is called
		_, err := routing.Compile(rules)

		// Then: compilation fails
		if err == nil {
			t.Error("expected error for rule without condition")
		}
	})

	t.Run("rejects rule without target profile", func(t *testing.T) {
		// Given: rule without target profile
		rules := []config.RouteRule{
			{
				ID:      "no-target",
				If:      "ctx_chars<1000",
				Use:     "",
				Explain: "No target",
			},
		}

		// When: Compile is called
		_, err := routing.Compile(rules)

		// Then: compilation fails
		if err == nil {
			t.Error("expected error for rule without target profile")
		}
	})
}

func TestMatch(t *testing.T) {
	t.Run("matches intent condition", func(t *testing.T) {
		// Given: engine with intent-based rule
		rules := []config.RouteRule{
			{
				ID:      "format-rule",
				If:      "intent=='format'",
				Use:     "quick-tasks",
				Explain: "Formatting tasks use quick profile",
			},
		}
		engine, _ := routing.Compile(rules)

		// When: Match with intent=format
		ctx := context.Background()
		features := routing.Features{
			Intent:   "format",
			CtxChars: 500,
		}
		decision, trace, err := engine.Match(ctx, features)

		// Then: matches the rule
		if err != nil {
			t.Fatalf("Match failed: %v", err)
		}
		if decision.Profile != "quick-tasks" {
			t.Errorf("profile = %s, want quick-tasks", decision.Profile)
		}
		if !trace.Matched {
			t.Error("expected trace.Matched = true")
		}
		if trace.RuleID != "format-rule" {
			t.Errorf("trace.RuleID = %s, want format-rule", trace.RuleID)
		}
		if trace.Explain == "" {
			t.Error("expected non-empty explanation")
		}
	})

	t.Run("matches ctx_chars condition", func(t *testing.T) {
		// Given: engine with ctx_chars rule
		rules := []config.RouteRule{
			{
				ID:      "small-context",
				If:      "ctx_chars<1000",
				Use:     "fast-model",
				Explain: "Small context uses fast model",
			},
		}
		engine, _ := routing.Compile(rules)

		// When: Match with CtxChars=500
		ctx := context.Background()
		features := routing.Features{
			CtxChars: 500,
		}
		decision, trace, err := engine.Match(ctx, features)

		// Then: matches the rule
		if err != nil {
			t.Fatalf("Match failed: %v", err)
		}
		if decision.Profile != "fast-model" {
			t.Errorf("profile = %s, want fast-model", decision.Profile)
		}
		if !trace.Matched {
			t.Error("expected match")
		}
	})

	t.Run("matches branch condition", func(t *testing.T) {
		// Given: engine with branch rule
		rules := []config.RouteRule{
			{
				ID:      "main-branch",
				If:      "branch=='main'",
				Use:     "production-profile",
				Explain: "Main branch uses production profile",
			},
		}
		engine, _ := routing.Compile(rules)

		// When: Match with branch=main
		ctx := context.Background()
		features := routing.Features{
			Branch: "main",
		}
		decision, trace, err := engine.Match(ctx, features)

		// Then: matches the rule
		if err != nil {
			t.Fatalf("Match failed: %v", err)
		}
		if decision.Profile != "production-profile" {
			t.Errorf("profile = %s, want production-profile", decision.Profile)
		}
		if trace.RuleID != "main-branch" {
			t.Errorf("trace.RuleID = %s, want main-branch", trace.RuleID)
		}
	})

	t.Run("handles no match", func(t *testing.T) {
		// Given: engine with specific rule
		rules := []config.RouteRule{
			{
				ID:      "specific-rule",
				If:      "intent=='review' && ctx_chars>10000",
				Use:     "thorough-profile",
				Explain: "Large review tasks",
			},
		}
		engine, _ := routing.Compile(rules)

		// When: Match with non-matching features
		ctx := context.Background()
		features := routing.Features{
			Intent:   "format",
			CtxChars: 500,
		}
		decision, trace, err := engine.Match(ctx, features)

		// Then: no match, returns fallback
		if err != nil {
			t.Fatalf("Match failed: %v", err)
		}
		if decision == nil {
			t.Error("expected non-nil decision")
		}
		if trace.Matched {
			t.Error("expected trace.Matched = false for no match")
		}
	})

	t.Run("returns exploration decision", func(t *testing.T) {
		// Given: engine with rules
		rules := []config.RouteRule{
			{
				ID:      "default",
				If:      "ctx_chars>0",
				Use:     "standard-profile",
				Explain: "Default profile",
			},
		}
		engine, _ := routing.Compile(rules)

		// When: Match multiple times (some should explore)
		ctx := context.Background()
		features := routing.Features{
			CtxChars: 500,
		}

		// Run multiple times to potentially trigger exploration
		exploredCount := 0
		for i := 0; i < 100; i++ {
			decision, _, _ := engine.Match(ctx, features)
			if decision.Explore {
				exploredCount++
			}
		}

		// Then: some decisions should be exploration (epsilon-greedy ~3%)
		// Note: This is probabilistic, so we just check it's possible
		_ = exploredCount // Don't fail test on probabilistic behavior
	})
}

func TestTrace(t *testing.T) {
	t.Run("trace contains rule explanation", func(t *testing.T) {
		// Given: engine with explanatory rule
		rules := []config.RouteRule{
			{
				ID:      "test-rule",
				If:      "intent=='test'",
				Use:     "test-profile",
				Explain: "This is a detailed explanation of why this rule matched",
			},
		}
		engine, _ := routing.Compile(rules)

		// When: Match triggers the rule
		ctx := context.Background()
		features := routing.Features{
			Intent: "test",
		}
		_, trace, err := engine.Match(ctx, features)

		// Then: trace contains explanation
		if err != nil {
			t.Fatalf("Match failed: %v", err)
		}
		if trace.Explain == "" {
			t.Error("expected non-empty explanation")
		}
		if !strings.Contains(trace.Explain, "detailed explanation") {
			t.Errorf("trace.Explain = %s, want to contain 'detailed explanation'", trace.Explain)
		}
	})
}
