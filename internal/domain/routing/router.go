// Package routing provides request routing logic based on context and rules.
package routing

import (
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/store/config"
)

// Context represents the routing context
type Context struct {
	ProjectType []string
	Intent      string
	Text        string
	Project     string
	Branch      string
	TimeOfDay   string
	CtxChars    int
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
	routes        *config.RoutesConfig
	rng           *rand.Rand
	epsilonRate   float64
	enableExplore bool
}

// NewRouter creates a new router
func NewRouter(routes *config.RoutesConfig) *Router {
	return &Router{
		routes:        routes,
		epsilonRate:   0.03, // 3% default exploration rate
		enableExplore: true,
		// #nosec G404 -- weak RNG acceptable for epsilon-greedy exploration
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SetExplorationRate sets the epsilon value for epsilon-greedy exploration
func (r *Router) SetExplorationRate(epsilon float64) {
	r.epsilonRate = epsilon
}

// SetEnableExplore enables or disables exploration
func (r *Router) SetEnableExplore(enable bool) {
	r.enableExplore = enable
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

	// First, determine the normal routing decision
	var normalDecision *Decision

	// Try each rule in order
	for _, rule := range r.routes.Rules {
		if r.matchRule(rule, ctx) {
			normalDecision = &Decision{
				ProfileKey: rule.Use,
				RuleID:     rule.ID,
				Explain:    rule.Explain,
				Fallback:   rule.Fallback,
			}
			break
		}
	}

	// If no rule matched, use active profile
	if normalDecision == nil {
		normalDecision = &Decision{
			ProfileKey: activeProfile,
			Explain:    "No routing rule matched",
		}
	}

	// Apply epsilon-greedy exploration
	if r.enableExplore && r.rng.Float64() < r.epsilonRate {
		// Explore: randomly select a different profile
		allProfiles := r.collectAllProfiles()
		if len(allProfiles) > 1 {
			// Remove the normal choice to ensure exploration is different
			var explorationOptions []string
			for _, p := range allProfiles {
				if p != normalDecision.ProfileKey {
					explorationOptions = append(explorationOptions, p)
				}
			}

			if len(explorationOptions) > 0 {
				exploredProfile := explorationOptions[r.rng.Intn(len(explorationOptions))]
				return &Decision{
					ProfileKey: exploredProfile,
					RuleID:     normalDecision.RuleID,
					Explain:    "Exploration: randomly selected for learning",
					Fallback:   normalDecision.ProfileKey, // Can fallback to normal choice
					Explore:    true,
				}
			}
		}
	}

	return normalDecision
}

// collectAllProfiles collects all profile names mentioned in rules
func (r *Router) collectAllProfiles() []string {
	profileSet := make(map[string]bool)
	for _, rule := range r.routes.Rules {
		if rule.Use != "" {
			profileSet[rule.Use] = true
		}
		if rule.Fallback != "" {
			profileSet[rule.Fallback] = true
		}
	}

	profiles := make([]string, 0, len(profileSet))
	for p := range profileSet {
		profiles = append(profiles, p)
	}
	return profiles
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
