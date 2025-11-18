package reports

import (
	"testing"
)

func TestNewService(t *testing.T) {
	svc := NewService()

	if svc == nil {
		t.Fatal("expected service to be created")
	}
	if len(svc.options) == 0 {
		t.Error("options should not be empty")
	}
	if svc.commandHelp == "" {
		t.Error("commandHelp should not be empty")
	}
}

func TestOptionCount(t *testing.T) {
	svc := NewService()

	count := svc.OptionCount()

	if count == 0 {
		t.Fatal("OptionCount should return non-zero value")
	}

	expectedCount := 6
	if count != expectedCount {
		t.Errorf("OptionCount: got %d, want %d", count, expectedCount)
	}
}

func TestOptions(t *testing.T) {
	svc := NewService()

	options := svc.Options()

	if len(options) == 0 {
		t.Fatal("Options should return non-empty list")
	}

	if len(options) != len(svc.options) {
		t.Errorf("Options length: got %d, want %d", len(options), len(svc.options))
	}

	for i, opt := range svc.options {
		if options[i].Label != opt.Label {
			t.Errorf("Option[%d].Label: got %q, want %q", i, options[i].Label, opt.Label)
		}
		if options[i].Desc != opt.Desc {
			t.Errorf("Option[%d].Desc: got %q, want %q", i, options[i].Desc, opt.Desc)
		}
	}
}

func TestOptions_AllFieldsNonEmpty(t *testing.T) {
	svc := NewService()

	options := svc.Options()

	for i, opt := range options {
		if opt.Label == "" {
			t.Errorf("Option[%d].Label should not be empty", i)
		}
		if opt.Desc == "" {
			t.Errorf("Option[%d].Desc should not be empty", i)
		}
	}
}

func TestCommandHelp(t *testing.T) {
	svc := NewService()

	help := svc.CommandHelp()

	if help == "" {
		t.Error("CommandHelp should return non-empty string")
	}

	expectedHelp := "Use CLI: boba report --format <json|csv|html> --days <N> --out <file>"
	if help != expectedHelp {
		t.Errorf("CommandHelp: got %q, want %q", help, expectedHelp)
	}
}

func TestOptions_ExpectedOptions(t *testing.T) {
	svc := NewService()

	options := svc.Options()

	expectedLabels := []string{
		"Last 7 Days Report",
		"Last 30 Days Report",
		"Custom Date Range",
		"JSON Format",
		"CSV Format",
		"HTML Format",
	}

	if len(options) != len(expectedLabels) {
		t.Fatalf("Expected %d options, got %d", len(expectedLabels), len(options))
	}

	for i, expected := range expectedLabels {
		if options[i].Label != expected {
			t.Errorf("Option[%d].Label: got %q, want %q", i, options[i].Label, expected)
		}
	}
}

func TestOptions_Consistency(t *testing.T) {
	svc := NewService()

	// Call Options multiple times
	options1 := svc.Options()
	options2 := svc.Options()

	if len(options1) != len(options2) {
		t.Error("Options should return consistent results")
	}

	for i := range options1 {
		if options1[i].Label != options2[i].Label {
			t.Errorf("Inconsistent Label at index %d", i)
		}
		if options1[i].Desc != options2[i].Desc {
			t.Errorf("Inconsistent Desc at index %d", i)
		}
	}
}

func TestOptionCount_MatchesOptionsLength(t *testing.T) {
	svc := NewService()

	count := svc.OptionCount()
	options := svc.Options()

	if count != len(options) {
		t.Errorf("OptionCount (%d) should match Options length (%d)", count, len(options))
	}
}

func TestNewService_DefaultOptions(t *testing.T) {
	svc := NewService()

	// Verify specific expected options exist
	expectedOptions := map[string]bool{
		"Last 7 Days Report":  false,
		"Last 30 Days Report": false,
		"JSON Format":         false,
		"CSV Format":          false,
		"HTML Format":         false,
	}

	for _, opt := range svc.options {
		if _, exists := expectedOptions[opt.Label]; exists {
			expectedOptions[opt.Label] = true
		}
	}

	for label, found := range expectedOptions {
		if !found {
			t.Errorf("Expected option %q not found", label)
		}
	}
}

func TestOptions_TimeRangeOptions(t *testing.T) {
	svc := NewService()

	options := svc.Options()

	// Check that time range options exist
	hasTimeRange := false
	for _, opt := range options {
		if opt.Label == "Last 7 Days Report" || opt.Label == "Last 30 Days Report" || opt.Label == "Custom Date Range" {
			hasTimeRange = true
			if opt.Desc == "" {
				t.Errorf("Time range option %q should have description", opt.Label)
			}
		}
	}

	if !hasTimeRange {
		t.Error("Should have time range options")
	}
}

func TestOptions_FormatOptions(t *testing.T) {
	svc := NewService()

	options := svc.Options()

	// Check that format options exist
	formats := []string{"JSON Format", "CSV Format", "HTML Format"}
	foundFormats := make(map[string]bool)

	for _, opt := range options {
		for _, format := range formats {
			if opt.Label == format {
				foundFormats[format] = true
			}
		}
	}

	if len(foundFormats) == 0 {
		t.Error("Should have format options")
	}
}
