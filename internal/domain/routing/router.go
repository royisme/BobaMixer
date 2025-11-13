package routing

import (
	"regexp"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/store/config"
)

// Context represents the routing context
type Context struct {
	Intent      string
	Text        string
	CtxChars    int
	Project     string
	Branch      string
	ProjectType []string
	TimeOfDay   string
}

// Decision represents a routing decision
type Decision struct {
	ProfileKey string
	RuleID     string
	Explain    string
	Fallback   string
	Explore    bool
}

// Router handles profile routing based on rules
type Router struct {
	routes *config.RoutesConfig
}

// NewRouter creates a new router
func NewRouter(routes *config.RoutesConfig) *Router {
	return &Router{routes: routes}
}

// Route determines which profile to use based on context
func (r *Router) Route(ctx Context, activeProfile string) *Decision {
	// If no routes configured, use active profile
	if r.routes == nil || len(r.routes.Rules) == 0 {
		return &Decision{
			ProfileKey: activeProfile,
			Explain:    "No routing rules configured",
		}
	}

	// Try each rule in order
	for _, rule := range r.routes.Rules {
		if r.matchRule(rule, ctx) {
			return &Decision{
				ProfileKey: rule.Use,
				RuleID:     rule.ID,
				Explain:    rule.Explain,
				Fallback:   rule.Fallback,
			}
		}
	}

	// No rule matched, use active profile
	return &Decision{
		ProfileKey: activeProfile,
		Explain:    "No routing rule matched",
	}
}

// matchRule checks if a rule matches the context
func (r *Router) matchRule(rule config.RouteRule, ctx Context) bool {
	if rule.If == "" {
		return false
	}

	// Simple expression evaluator
	// Supports: intent=='value', text.matches('pattern'), ctx_chars>N, etc.
	expr := rule.If

	// Check for intent equality
	if strings.Contains(expr, "intent==") {
		re := regexp.MustCompile(`intent=='([^']+)'`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 1 && ctx.Intent == matches[1] {
			return true
		}
	}

	// Check for text.matches
	if strings.Contains(expr, "text.matches") {
		re := regexp.MustCompile(`text\.matches\('([^']+)'\)`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 1 {
			pattern := regexp.MustCompile(matches[1])
			if pattern.MatchString(ctx.Text) {
				return true
			}
		}
	}

	// Check for text.contains
	if strings.Contains(expr, "text.contains") {
		re := regexp.MustCompile(`text\.contains\('([^']+)'\)`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 1 && strings.Contains(ctx.Text, matches[1]) {
			return true
		}
	}

	// Check for ctx_chars comparison
	if strings.Contains(expr, "ctx_chars>") {
		re := regexp.MustCompile(`ctx_chars>(\d+)`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 1 {
			var threshold int
			if _, err := regexp.MatchString(`\d+`, matches[1]); err == nil {
				// Parse threshold
				threshold = 0
				for _, c := range matches[1] {
					threshold = threshold*10 + int(c-'0')
				}
				if ctx.CtxChars > threshold {
					return true
				}
			}
		}
	}

	// Check for task.matches
	if strings.Contains(expr, "task.matches") {
		re := regexp.MustCompile(`task\.matches\('([^']+)'\)`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 1 {
			pattern := regexp.MustCompile(matches[1])
			if pattern.MatchString(ctx.Intent) {
				return true
			}
		}
	}

	// Check for branch.matches
	if strings.Contains(expr, "branch.matches") {
		re := regexp.MustCompile(`branch\.matches\('([^']+)'\)`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 1 {
			pattern := regexp.MustCompile(matches[1])
			if pattern.MatchString(ctx.Branch) {
				return true
			}
		}
	}

	return false
}

// CheckSubAgent checks if a sub-agent should be triggered
func (r *Router) CheckSubAgent(ctx Context) (string, bool) {
	if r.routes == nil || len(r.routes.SubAgents) == 0 {
		return "", false
	}

	for _, agent := range r.routes.SubAgents {
		// Check triggers
		triggered := false
		for _, trigger := range agent.Triggers {
			if strings.Contains(strings.ToLower(ctx.Text), trigger) ||
				strings.Contains(strings.ToLower(ctx.Intent), trigger) {
				triggered = true
				break
			}
		}

		if !triggered {
			continue
		}

		// Check conditions
		if agent.Conditions != nil {
			if minChars, ok := agent.Conditions["min_ctx_chars"].(int); ok {
				if ctx.CtxChars < minChars {
					continue
				}
			}

			if maxChars, ok := agent.Conditions["max_ctx_chars"].(int); ok {
				if ctx.CtxChars > maxChars {
					continue
				}
			}

			if timeRanges, ok := agent.Conditions["time_of_day"].([]interface{}); ok {
				if !checkTimeRange(timeRanges) {
					continue
				}
			}
		}

		// All conditions met
		return agent.Profile, true
	}

	return "", false
}

// checkTimeRange checks if current time is within specified ranges
func checkTimeRange(ranges []interface{}) bool {
	now := time.Now()
	currentTime := now.Format("15:04")

	for _, r := range ranges {
		rangeStr, ok := r.(string)
		if !ok {
			continue
		}

		parts := strings.Split(rangeStr, "-")
		if len(parts) != 2 {
			continue
		}

		start := strings.TrimSpace(parts[0])
		end := strings.TrimSpace(parts[1])

		if currentTime >= start && currentTime <= end {
			return true
		}
	}

	return false
}
