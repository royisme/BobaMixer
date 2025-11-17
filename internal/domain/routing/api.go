// Package routing provides TDD-spec aligned routing DSL APIs.
package routing

import (
	"context"
	"fmt"
	"regexp"

	"github.com/royisme/bobamixer/internal/store/config"
)

// Features represents the routing context for decision making.
type Features struct {
	Intent       string   // Task intent (e.g., "format", "review", "refactor")
	TextSample   string   // Sample of input text
	CtxChars     int      // Context size in characters
	Branch       string   // Git branch name
	ProjectTypes []string // Project types (e.g., ["go", "web"])
	TimeOfDay    string   // Time in "HH:MM" format
	BudgetHint   string   // "near_cap" | "normal" | "over_cap"
}

// RoutingDecision represents a routing decision (TDD-spec aligned).
//
//nolint:revive // RoutingDecision is the established API name
type RoutingDecision struct {
	Profile  string // Selected profile key
	Fallback string // Fallback profile if primary unavailable
	Explore  bool   // Whether this is an exploration decision
}

// Trace contains routing decision explanation.
type Trace struct {
	RuleID  string // ID of the rule that matched
	Explain string // Human-readable explanation
	Matched bool   // Whether the rule matched
}

// Engine is the routing decision engine.
type Engine struct {
	router *Router
	rules  []config.RouteRule
}

// Compile validates and compiles routing rules into an Engine.
// Returns ErrConfig if rules contain invalid patterns or syntax.
func Compile(rules []config.RouteRule) (*Engine, error) {
	// Validate all rules
	for i, rule := range rules {
		if err := validateRule(rule); err != nil {
			return nil, fmt.Errorf("rule %d (%s): %w", i, rule.ID, err)
		}
	}

	// Create routes config
	routesConfig := &config.RoutesConfig{
		Rules: rules,
		Explore: config.ExploreConfig{
			Enabled: true,
			Rate:    0.03,
		},
	}

	// Create router
	router := NewRouter(routesConfig)

	return &Engine{
		router: router,
		rules:  rules,
	}, nil
}

// Match determines the routing decision based on features.
// Returns the decision and a trace explaining how the decision was made.
func (e *Engine) Match(ctx context.Context, f Features) (*RoutingDecision, *Trace, error) {
	// Convert Features to routing.Context
	routingCtx := Context{
		Intent:      f.Intent,
		Text:        f.TextSample,
		CtxChars:    f.CtxChars,
		Branch:      f.Branch,
		ProjectType: f.ProjectTypes,
		TimeOfDay:   f.TimeOfDay,
	}

	// Use empty active profile for pure rule-based routing
	activeProfile := ""

	// Execute routing
	decision := e.router.Route(routingCtx, activeProfile)

	// Build trace
	trace := &Trace{
		RuleID:  decision.RuleID,
		Explain: decision.Explain,
		Matched: decision.RuleID != "",
	}

	// Build decision result
	result := &RoutingDecision{
		Profile:  decision.ProfileKey,
		Fallback: decision.Fallback,
		Explore:  decision.Explore,
	}

	return result, trace, nil
}

// validateRule validates a routing rule for syntax errors.
func validateRule(rule config.RouteRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	if rule.If == "" {
		return fmt.Errorf("rule condition (if) is required")
	}

	if rule.Use == "" {
		return fmt.Errorf("rule target profile (use) is required")
	}

	// Validate regex patterns in conditions
	if err := validateConditionPatterns(rule.If); err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	return nil
}

// validateConditionPatterns validates regex patterns in rule conditions.
func validateConditionPatterns(condition string) error {
	// Extract and validate text.matches() patterns
	textMatchesPattern := regexp.MustCompile(`text\.matches\('([^']+)'\)`)
	matches := textMatchesPattern.FindAllStringSubmatch(condition, -1)
	for _, match := range matches {
		if len(match) > 1 {
			pattern := match[1]
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
			}
		}
	}

	// Extract and validate branch.matches() patterns
	branchMatchesPattern := regexp.MustCompile(`branch\.matches\('([^']+)'\)`)
	matches = branchMatchesPattern.FindAllStringSubmatch(condition, -1)
	for _, match := range matches {
		if len(match) > 1 {
			pattern := match[1]
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
			}
		}
	}

	// Extract and validate task.matches() patterns
	taskMatchesPattern := regexp.MustCompile(`task\.matches\('([^']+)'\)`)
	matches = taskMatchesPattern.FindAllStringSubmatch(condition, -1)
	for _, match := range matches {
		if len(match) > 1 {
			pattern := match[1]
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
			}
		}
	}

	return nil
}
