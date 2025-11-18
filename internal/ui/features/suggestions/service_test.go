package suggestions

import (
	"testing"

	"github.com/royisme/bobamixer/internal/domain/suggestions"
	"github.com/royisme/bobamixer/internal/ui/components"
)

func TestNewService(t *testing.T) {
	home := "/test/home"
	svc := NewService(home)
	if svc == nil {
		t.Fatal("expected service to be created")
	}
	if svc.home != home {
		t.Errorf("expected home %q, got %q", home, svc.home)
	}
}

func TestConvertType(t *testing.T) {
	svc := NewService("/test")

	tests := []struct {
		name string
		typ  suggestions.SuggestionType
		want string
	}{
		{
			name: "cost optimization",
			typ:  suggestions.SuggestionCostOptimization,
			want: "cost",
		},
		{
			name: "profile switch",
			typ:  suggestions.SuggestionProfileSwitch,
			want: "profile",
		},
		{
			name: "budget adjust",
			typ:  suggestions.SuggestionBudgetAdjust,
			want: "budget",
		},
		{
			name: "anomaly",
			typ:  suggestions.SuggestionAnomaly,
			want: "anomaly",
		},
		{
			name: "usage pattern",
			typ:  suggestions.SuggestionUsagePattern,
			want: "usage",
		},
		{
			name: "unknown type defaults to usage",
			typ:  suggestions.SuggestionType(999),
			want: "usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.convertType(tt.typ)
			if got != tt.want {
				t.Errorf("convertType(%v): got %q, want %q", tt.typ, got, tt.want)
			}
		})
	}
}

func TestConvertToView(t *testing.T) {
	svc := NewService("/test")

	tests := []struct {
		name  string
		suggs []suggestions.Suggestion
		want  []components.Suggestion
	}{
		{
			name:  "empty list",
			suggs: []suggestions.Suggestion{},
			want:  []components.Suggestion{},
		},
		{
			name: "single suggestion",
			suggs: []suggestions.Suggestion{
				{
					Title:       "Reduce costs",
					Description: "Switch to cheaper model",
					Impact:      "Save $5/day",
					ActionItems: []string{"Update profile", "Test changes"},
					Priority:    5,
					Type:        suggestions.SuggestionCostOptimization,
				},
			},
			want: []components.Suggestion{
				{
					Title:       "Reduce costs",
					Description: "Switch to cheaper model",
					Impact:      "Save $5/day",
					ActionItems: []string{"Update profile", "Test changes"},
					Priority:    5,
					Type:        "cost",
				},
			},
		},
		{
			name: "multiple suggestions with different types",
			suggs: []suggestions.Suggestion{
				{
					Title:       "Cost optimization",
					Description: "Use cheaper provider",
					Impact:      "Save $10/day",
					ActionItems: []string{"Switch provider"},
					Priority:    4,
					Type:        suggestions.SuggestionCostOptimization,
				},
				{
					Title:       "Profile switch",
					Description: "Use fast profile for simple queries",
					Impact:      "Reduce latency by 50%",
					ActionItems: []string{"Configure routing rules"},
					Priority:    3,
					Type:        suggestions.SuggestionProfileSwitch,
				},
				{
					Title:       "Usage spike detected",
					Description: "Unusual usage on weekends",
					Impact:      "Review billing",
					ActionItems: []string{"Check logs", "Set alerts"},
					Priority:    5,
					Type:        suggestions.SuggestionAnomaly,
				},
			},
			want: []components.Suggestion{
				{
					Title:       "Cost optimization",
					Description: "Use cheaper provider",
					Impact:      "Save $10/day",
					ActionItems: []string{"Switch provider"},
					Priority:    4,
					Type:        "cost",
				},
				{
					Title:       "Profile switch",
					Description: "Use fast profile for simple queries",
					Impact:      "Reduce latency by 50%",
					ActionItems: []string{"Configure routing rules"},
					Priority:    3,
					Type:        "profile",
				},
				{
					Title:       "Usage spike detected",
					Description: "Unusual usage on weekends",
					Impact:      "Review billing",
					ActionItems: []string{"Check logs", "Set alerts"},
					Priority:    5,
					Type:        "anomaly",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.ConvertToView(tt.suggs)
			if len(got) != len(tt.want) {
				t.Fatalf("length: got %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Title != tt.want[i].Title {
					t.Errorf("[%d] Title: got %q, want %q", i, got[i].Title, tt.want[i].Title)
				}
				if got[i].Description != tt.want[i].Description {
					t.Errorf("[%d] Description: got %q, want %q", i, got[i].Description, tt.want[i].Description)
				}
				if got[i].Impact != tt.want[i].Impact {
					t.Errorf("[%d] Impact: got %q, want %q", i, got[i].Impact, tt.want[i].Impact)
				}
				if got[i].Priority != tt.want[i].Priority {
					t.Errorf("[%d] Priority: got %d, want %d", i, got[i].Priority, tt.want[i].Priority)
				}
				if got[i].Type != tt.want[i].Type {
					t.Errorf("[%d] Type: got %q, want %q", i, got[i].Type, tt.want[i].Type)
				}
				if len(got[i].ActionItems) != len(tt.want[i].ActionItems) {
					t.Errorf("[%d] ActionItems length: got %d, want %d", i, len(got[i].ActionItems), len(tt.want[i].ActionItems))
				}
				for j := range got[i].ActionItems {
					if got[i].ActionItems[j] != tt.want[i].ActionItems[j] {
						t.Errorf("[%d] ActionItems[%d]: got %q, want %q", i, j, got[i].ActionItems[j], tt.want[i].ActionItems[j])
					}
				}
			}
		})
	}
}

func TestConvertToView_ActionItemsCopied(t *testing.T) {
	svc := NewService("/test")

	original := []suggestions.Suggestion{
		{
			Title:       "Test",
			Description: "Test desc",
			Impact:      "Test impact",
			ActionItems: []string{"action1", "action2"},
			Priority:    3,
			Type:        suggestions.SuggestionUsagePattern,
		},
	}

	result := svc.ConvertToView(original)

	// Modify original action items
	original[0].ActionItems[0] = "modified"

	// Result should not be affected
	if result[0].ActionItems[0] == "modified" {
		t.Error("ActionItems were not properly copied, modification affected result")
	}
	if result[0].ActionItems[0] != "action1" {
		t.Errorf("Expected ActionItems[0] to be %q, got %q", "action1", result[0].ActionItems[0])
	}
}

func TestCommandHelp(t *testing.T) {
	svc := NewService("/test")
	help := svc.CommandHelp()
	if help == "" {
		t.Error("CommandHelp should return non-empty string")
	}
	expected := "Use CLI: boba action [--auto] to apply suggestions"
	if help != expected {
		t.Errorf("CommandHelp: got %q, want %q", help, expected)
	}
}
