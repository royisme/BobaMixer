// Package routing provides request routing logic based on context and rules.
package routing

import (
	"math/rand"
	"regexp"
	"strconv"
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
	epsilonRate := 0.03   // 3% default exploration rate
	enableExplore := true // enabled by default

	// Use configuration if available
	if routes != nil {
		epsilonRate = routes.Explore.Rate
		enableExplore = routes.Explore.Enabled
	}

	return &Router{
		routes:        routes,
		epsilonRate:   epsilonRate,
		enableExplore: enableExplore,
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
//
//nolint:gocyclo // Complex rule matching with multiple conditions and operators
func (r *Router) matchRule(rule config.RouteRule, ctx Context) bool {
	if rule.If == "" {
		return false
	}

	return r.evaluateBooleanExpression(strings.TrimSpace(rule.If), ctx)
}

// evaluateBooleanExpression evaluates complex boolean expressions supporting &&, ||, and parentheses
func (r *Router) evaluateBooleanExpression(expr string, ctx Context) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false
	}

	expr = trimOuterParentheses(expr)

	// Evaluate OR at the top level first (lowest precedence)
	if idx := findTopLevelOperator(expr, "||"); idx >= 0 {
		left := expr[:idx]
		right := expr[idx+2:]
		return r.evaluateBooleanExpression(left, ctx) || r.evaluateBooleanExpression(right, ctx)
	}

	// Then evaluate AND
	if idx := findTopLevelOperator(expr, "&&"); idx >= 0 {
		left := expr[:idx]
		right := expr[idx+2:]
		return r.evaluateBooleanExpression(left, ctx) && r.evaluateBooleanExpression(right, ctx)
	}

	return r.evaluateSingleCondition(expr, ctx)
}

func findTopLevelOperator(expr string, op string) int {
	depth := 0
	opLen := len(op)
	for i := 0; i <= len(expr)-opLen; i++ {
		switch expr[i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth == 0 && strings.HasPrefix(expr[i:], op) {
			return i
		}
	}
	return -1
}

func trimOuterParentheses(expr string) string {
	for {
		expr = strings.TrimSpace(expr)
		if len(expr) < 2 || expr[0] != '(' || expr[len(expr)-1] != ')' {
			return expr
		}
		match := matchingParenIndex(expr, 0)
		if match != len(expr)-1 {
			return expr
		}
		expr = expr[1 : len(expr)-1]
	}
}

func matchingParenIndex(expr string, start int) int {
	depth := 0
	for i := start; i < len(expr); i++ {
		switch expr[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// evaluateSingleCondition evaluates a single condition expression
func (r *Router) evaluateSingleCondition(expr string, ctx Context) bool {
	for _, handler := range conditionHandlers {
		if handler(expr, ctx) {
			return true
		}
	}

	return false
}

var conditionHandlers = []func(string, Context) bool{
	intentEqualsCondition,
	textMatchesCondition,
	textContainsCondition,
	ctxCharsCondition,
	taskMatchesCondition,
	branchMatchesCondition,
	branchEqualsCondition,
	branchEqualityShortcutCondition,
	timeOfDayCondition,
	projectTypeCondition,
}

var (
	intentEqualsPattern           = regexp.MustCompile(`intent=='([^']+)'`)
	textMatchesPattern            = regexp.MustCompile(`text\.matches\('([^']+)'\)`)
	textContainsPattern           = regexp.MustCompile(`text\.contains\('([^']+)'\)`)
	ctxCharsPattern               = regexp.MustCompile(`ctx_chars>(\d+)`)
	taskMatchesPattern            = regexp.MustCompile(`task\.matches\('([^']+)'\)`)
	branchMatchesPattern          = regexp.MustCompile(`branch\.matches\('([^']+)'\)`)
	branchEqualsPattern           = regexp.MustCompile(`branch\.equals\('([^']+)'\)`)
	branchEqualityShortcutPattern = regexp.MustCompile(`branch=='([^']+)'`)
	timeOfDayPattern              = regexp.MustCompile(`time_of_day\.in\('([^']+)'\)`)
	projectTypePattern            = regexp.MustCompile(`project_types\.contains\('([^']+)'\)`)
)

func intentEqualsCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "intent==") {
		return false
	}
	matches := intentEqualsPattern.FindStringSubmatch(expr)
	return len(matches) > 1 && ctx.Intent == matches[1]
}

func textMatchesCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "text.matches") {
		return false
	}
	matches := textMatchesPattern.FindStringSubmatch(expr)
	if len(matches) <= 1 {
		return false
	}
	pattern, err := regexp.Compile(matches[1])
	if err != nil {
		return false
	}
	return pattern.MatchString(ctx.Text)
}

func textContainsCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "text.contains") {
		return false
	}
	matches := textContainsPattern.FindStringSubmatch(expr)
	return len(matches) > 1 && strings.Contains(ctx.Text, matches[1])
}

func ctxCharsCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "ctx_chars>") {
		return false
	}
	matches := ctxCharsPattern.FindStringSubmatch(expr)
	if len(matches) <= 1 {
		return false
	}
	threshold, err := strconv.Atoi(matches[1])
	if err != nil {
		return false
	}
	return ctx.CtxChars > threshold
}

func taskMatchesCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "task.matches") {
		return false
	}
	matches := taskMatchesPattern.FindStringSubmatch(expr)
	if len(matches) <= 1 {
		return false
	}
	pattern, err := regexp.Compile(matches[1])
	if err != nil {
		return false
	}
	return pattern.MatchString(ctx.Intent)
}

func branchMatchesCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "branch.matches") {
		return false
	}
	matches := branchMatchesPattern.FindStringSubmatch(expr)
	if len(matches) <= 1 {
		return false
	}
	pattern, err := regexp.Compile(matches[1])
	if err != nil {
		return false
	}
	return pattern.MatchString(ctx.Branch)
}

func branchEqualsCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "branch.equals") {
		return false
	}
	matches := branchEqualsPattern.FindStringSubmatch(expr)
	return len(matches) > 1 && ctx.Branch == matches[1]
}

func branchEqualityShortcutCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "branch==") {
		return false
	}
	matches := branchEqualityShortcutPattern.FindStringSubmatch(expr)
	return len(matches) > 1 && ctx.Branch == matches[1]
}

func timeOfDayCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "time_of_day.in") {
		return false
	}
	matches := timeOfDayPattern.FindStringSubmatch(expr)
	return len(matches) > 1 && checkTimeRangeString(matches[1])
}

func projectTypeCondition(expr string, ctx Context) bool {
	if !strings.Contains(expr, "project_types.contains") {
		return false
	}
	matches := projectTypePattern.FindStringSubmatch(expr)
	if len(matches) <= 1 {
		return false
	}
	targetType := matches[1]
	for _, pt := range ctx.ProjectType {
		if pt == targetType {
			return true
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

// checkTimeRangeString checks if current time is within a single time range string
func checkTimeRangeString(rangeStr string) bool {
	now := time.Now()
	currentTime := now.Format("15:04")

	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return false
	}

	start := strings.TrimSpace(parts[0])
	end := strings.TrimSpace(parts[1])

	return currentTime >= start && currentTime <= end
}
