package stats

import (
	"testing"

	"github.com/royisme/bobamixer/internal/domain/stats"
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

func TestConvertSummary(t *testing.T) {
	svc := NewService("/test")

	tests := []struct {
		name            string
		title           string
		summary         stats.Summary
		includeAverages bool
		want            components.StatsSummary
	}{
		{
			name:            "without averages",
			title:           "Test Summary",
			summary:         stats.Summary{TotalTokens: 1000, TotalCost: 0.05, TotalSessions: 5},
			includeAverages: false,
			want: components.StatsSummary{
				Title:        "Test Summary",
				Tokens:       1000,
				Cost:         0.05,
				Sessions:     5,
				ShowAverages: false,
			},
		},
		{
			name:  "with averages",
			title: "Weekly Stats",
			summary: stats.Summary{
				TotalTokens:    7000,
				TotalCost:      0.35,
				TotalSessions:  35,
				AvgDailyTokens: 1000.0,
				AvgDailyCost:   0.05,
			},
			includeAverages: true,
			want: components.StatsSummary{
				Title:          "Weekly Stats",
				Tokens:         7000,
				Cost:           0.35,
				Sessions:       35,
				AvgDailyTokens: 1000.0,
				AvgDailyCost:   0.05,
				ShowAverages:   true,
			},
		},
		{
			name:            "empty summary",
			title:           "Empty",
			summary:         stats.Summary{},
			includeAverages: false,
			want: components.StatsSummary{
				Title:        "Empty",
				ShowAverages: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.convertSummary(tt.title, tt.summary, tt.includeAverages)
			if got.Title != tt.want.Title {
				t.Errorf("Title: got %q, want %q", got.Title, tt.want.Title)
			}
			if got.Tokens != tt.want.Tokens {
				t.Errorf("Tokens: got %d, want %d", got.Tokens, tt.want.Tokens)
			}
			if got.Cost != tt.want.Cost {
				t.Errorf("Cost: got %.4f, want %.4f", got.Cost, tt.want.Cost)
			}
			if got.Sessions != tt.want.Sessions {
				t.Errorf("Sessions: got %d, want %d", got.Sessions, tt.want.Sessions)
			}
			if got.ShowAverages != tt.want.ShowAverages {
				t.Errorf("ShowAverages: got %v, want %v", got.ShowAverages, tt.want.ShowAverages)
			}
			if tt.includeAverages {
				if got.AvgDailyTokens != tt.want.AvgDailyTokens {
					t.Errorf("AvgDailyTokens: got %.2f, want %.2f", got.AvgDailyTokens, tt.want.AvgDailyTokens)
				}
				if got.AvgDailyCost != tt.want.AvgDailyCost {
					t.Errorf("AvgDailyCost: got %.4f, want %.4f", got.AvgDailyCost, tt.want.AvgDailyCost)
				}
			}
		})
	}
}

func TestConvertProfiles(t *testing.T) {
	svc := NewService("/test")

	tests := []struct {
		name  string
		stats []stats.ProfileStats
		want  []components.StatsProfile
	}{
		{
			name:  "empty list",
			stats: []stats.ProfileStats{},
			want:  nil,
		},
		{
			name: "single profile",
			stats: []stats.ProfileStats{
				{
					ProfileName:  "default",
					TotalTokens:  1000,
					TotalCost:    0.05,
					SessionCount: 10,
					AvgLatencyMS: 250.5,
					UsagePercent: 100.0,
					CostPercent:  100.0,
				},
			},
			want: []components.StatsProfile{
				{
					Name:       "default",
					Tokens:     1000,
					Cost:       0.05,
					Sessions:   10,
					AvgLatency: 250.5,
					UsagePct:   100.0,
					CostPct:    100.0,
				},
			},
		},
		{
			name: "multiple profiles",
			stats: []stats.ProfileStats{
				{
					ProfileName:  "fast",
					TotalTokens:  500,
					TotalCost:    0.03,
					SessionCount: 5,
					AvgLatencyMS: 100.0,
					UsagePercent: 33.3,
					CostPercent:  30.0,
				},
				{
					ProfileName:  "accurate",
					TotalTokens:  1000,
					TotalCost:    0.07,
					SessionCount: 15,
					AvgLatencyMS: 500.0,
					UsagePercent: 66.7,
					CostPercent:  70.0,
				},
			},
			want: []components.StatsProfile{
				{
					Name:       "fast",
					Tokens:     500,
					Cost:       0.03,
					Sessions:   5,
					AvgLatency: 100.0,
					UsagePct:   33.3,
					CostPct:    30.0,
				},
				{
					Name:       "accurate",
					Tokens:     1000,
					Cost:       0.07,
					Sessions:   15,
					AvgLatency: 500.0,
					UsagePct:   66.7,
					CostPct:    70.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.convertProfiles(tt.stats)
			if len(got) != len(tt.want) {
				t.Fatalf("length: got %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("[%d] Name: got %q, want %q", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].Tokens != tt.want[i].Tokens {
					t.Errorf("[%d] Tokens: got %d, want %d", i, got[i].Tokens, tt.want[i].Tokens)
				}
				if got[i].Cost != tt.want[i].Cost {
					t.Errorf("[%d] Cost: got %.4f, want %.4f", i, got[i].Cost, tt.want[i].Cost)
				}
				if got[i].Sessions != tt.want[i].Sessions {
					t.Errorf("[%d] Sessions: got %d, want %d", i, got[i].Sessions, tt.want[i].Sessions)
				}
				if got[i].AvgLatency != tt.want[i].AvgLatency {
					t.Errorf("[%d] AvgLatency: got %.2f, want %.2f", i, got[i].AvgLatency, tt.want[i].AvgLatency)
				}
				if got[i].UsagePct != tt.want[i].UsagePct {
					t.Errorf("[%d] UsagePct: got %.2f, want %.2f", i, got[i].UsagePct, tt.want[i].UsagePct)
				}
				if got[i].CostPct != tt.want[i].CostPct {
					t.Errorf("[%d] CostPct: got %.2f, want %.2f", i, got[i].CostPct, tt.want[i].CostPct)
				}
			}
		})
	}
}

func TestConvertToView(t *testing.T) {
	svc := NewService("/test")

	data := StatsData{
		Today: stats.Summary{
			TotalTokens:   500,
			TotalCost:     0.02,
			TotalSessions: 3,
		},
		Week: stats.Summary{
			TotalTokens:    3500,
			TotalCost:      0.15,
			TotalSessions:  21,
			AvgDailyTokens: 500.0,
			AvgDailyCost:   0.021,
		},
		ProfileStats: []stats.ProfileStats{
			{
				ProfileName:  "default",
				TotalTokens:  3500,
				TotalCost:    0.15,
				SessionCount: 21,
				AvgLatencyMS: 300.0,
				UsagePercent: 100.0,
				CostPercent:  100.0,
			},
		},
	}

	view := svc.ConvertToView(data)

	// Check Today
	if view.Today.Title != "📅 Today's Usage" {
		t.Errorf("Today.Title: got %q, want %q", view.Today.Title, "📅 Today's Usage")
	}
	if view.Today.Tokens != 500 {
		t.Errorf("Today.Tokens: got %d, want 500", view.Today.Tokens)
	}
	if view.Today.ShowAverages {
		t.Error("Today.ShowAverages should be false")
	}

	// Check Week
	if view.Week.Title != "📊 Last 7 Days" {
		t.Errorf("Week.Title: got %q, want %q", view.Week.Title, "📊 Last 7 Days")
	}
	if view.Week.Tokens != 3500 {
		t.Errorf("Week.Tokens: got %d, want 3500", view.Week.Tokens)
	}
	if !view.Week.ShowAverages {
		t.Error("Week.ShowAverages should be true")
	}

	// Check Profiles
	if len(view.Profiles) != 1 {
		t.Fatalf("Profiles length: got %d, want 1", len(view.Profiles))
	}
	if view.Profiles[0].Name != "default" {
		t.Errorf("Profiles[0].Name: got %q, want %q", view.Profiles[0].Name, "default")
	}
}
