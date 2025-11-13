package routing

import (
	"testing"

	"github.com/royisme/bobamixer/internal/store/config"
)

func TestRoute(t *testing.T) {
	routes := &config.RoutesConfig{
		Rules: []config.RouteRule{
			{
				ID:      "formatting",
				If:      "intent=='format'",
				Use:     "quick-tasks",
				Explain: "Format tasks use quick profile",
			},
			{
				ID:      "large-context",
				If:      "ctx_chars>3000",
				Use:     "work-heavy",
				Explain: "Large context uses heavy profile",
			},
			{
				ID:      "text-match",
				If:      "text.matches('review')",
				Use:     "work-heavy",
				Explain: "Review tasks use heavy profile",
			},
		},
	}

	router := NewRouter(routes)

	tests := []struct {
		name            string
		ctx             Context
		activeProfile   string
		expectedProfile string
		expectedRuleID  string
	}{
		{
			name: "match intent format",
			ctx: Context{
				Intent: "format",
				Text:   "format this code",
			},
			activeProfile:   "default",
			expectedProfile: "quick-tasks",
			expectedRuleID:  "formatting",
		},
		{
			name: "match large context",
			ctx: Context{
				Intent:   "analyze",
				CtxChars: 5000,
			},
			activeProfile:   "default",
			expectedProfile: "work-heavy",
			expectedRuleID:  "large-context",
		},
		{
			name: "match text pattern",
			ctx: Context{
				Intent: "task",
				Text:   "please review this PR",
			},
			activeProfile:   "default",
			expectedProfile: "work-heavy",
			expectedRuleID:  "text-match",
		},
		{
			name: "no match uses active profile",
			ctx: Context{
				Intent:   "unknown",
				CtxChars: 100,
				Text:     "simple task",
			},
			activeProfile:   "default",
			expectedProfile: "default",
			expectedRuleID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := router.Route(tt.ctx, tt.activeProfile)

			if decision.ProfileKey != tt.expectedProfile {
				t.Errorf("ProfileKey: got %q, want %q", decision.ProfileKey, tt.expectedProfile)
			}
			if decision.RuleID != tt.expectedRuleID {
				t.Errorf("RuleID: got %q, want %q", decision.RuleID, tt.expectedRuleID)
			}
		})
	}
}

func TestCheckSubAgent(t *testing.T) {
	routes := &config.RoutesConfig{
		SubAgents: map[string]config.SubAgent{
			"code_review": {
				Profile:  "work-heavy",
				Triggers: []string{"review", "check", "audit"},
				Conditions: map[string]interface{}{
					"min_ctx_chars": 3000,
				},
			},
			"quick_fix": {
				Profile:  "quick-tasks",
				Triggers: []string{"fix", "typo"},
				Conditions: map[string]interface{}{
					"max_ctx_chars": 1200,
				},
			},
		},
	}

	router := NewRouter(routes)

	tests := []struct {
		name            string
		ctx             Context
		expectedProfile string
		expectedTrigger bool
	}{
		{
			name: "trigger code review with enough context",
			ctx: Context{
				Intent:   "review",
				CtxChars: 5000,
				Text:     "please review this code",
			},
			expectedProfile: "work-heavy",
			expectedTrigger: true,
		},
		{
			name: "not enough context for code review",
			ctx: Context{
				Intent:   "review",
				CtxChars: 1000,
				Text:     "review this",
			},
			expectedProfile: "",
			expectedTrigger: false,
		},
		{
			name: "trigger quick fix with small context",
			ctx: Context{
				Intent:   "fix typo",
				CtxChars: 500,
				Text:     "fix this typo",
			},
			expectedProfile: "quick-tasks",
			expectedTrigger: true,
		},
		{
			name: "no trigger",
			ctx: Context{
				Intent:   "other",
				CtxChars: 500,
				Text:     "some other task",
			},
			expectedProfile: "",
			expectedTrigger: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, triggered := router.CheckSubAgent(tt.ctx)

			if triggered != tt.expectedTrigger {
				t.Errorf("triggered: got %v, want %v", triggered, tt.expectedTrigger)
			}
			if profile != tt.expectedProfile {
				t.Errorf("profile: got %q, want %q", profile, tt.expectedProfile)
			}
		})
	}
}
