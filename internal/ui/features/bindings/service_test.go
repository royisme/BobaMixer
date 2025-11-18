package bindings

import (
	"testing"

	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/ui/forms"
)

func TestNewService(t *testing.T) {
	bindings := &core.BindingsConfig{}
	tools := &core.ToolsConfig{}
	providers := &core.ProvidersConfig{}
	form := &forms.BindingForm{}
	msgNoSelection := "no selection"
	msgInvalid := "invalid"

	svc := NewService(bindings, tools, providers, form, msgNoSelection, msgInvalid)

	if svc == nil {
		t.Fatal("expected service to be created")
	}
	if svc.bindings != bindings {
		t.Error("bindings not set correctly")
	}
	if svc.tools != tools {
		t.Error("tools not set correctly")
	}
	if svc.providers != providers {
		t.Error("providers not set correctly")
	}
	if svc.form != form {
		t.Error("form not set correctly")
	}
	if svc.msgNoSelection != msgNoSelection {
		t.Error("msgNoSelection not set correctly")
	}
	if svc.msgInvalid != msgInvalid {
		t.Error("msgInvalid not set correctly")
	}
}

func TestStartForm_NilForm(t *testing.T) {
	bindings := &core.BindingsConfig{}
	tools := &core.ToolsConfig{}
	providers := &core.ProvidersConfig{}
	svc := NewService(bindings, tools, providers, nil, "no selection", "invalid")

	result := svc.StartForm(true, []int{}, 0)

	if result {
		t.Error("StartForm should return false when form is nil")
	}
}

func TestRows_Success(t *testing.T) {
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{
				ToolID:     "tool1",
				ProviderID: "provider1",
				UseProxy:   true,
				Options: core.BindingOptions{
					Model: "gpt-4",
				},
			},
			{
				ToolID:     "tool2",
				ProviderID: "provider2",
				UseProxy:   false,
				Options: core.BindingOptions{
					Model: "claude-3",
				},
			},
		},
	}
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1", Name: "Tool 1"},
			{ID: "tool2", Name: "Tool 2"},
		},
	}
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
			{ID: "provider2", DisplayName: "Provider 2"},
		},
	}
	svc := NewService(bindings, tools, providers, nil, "", "")

	indexes := []int{0, 1}
	rows := svc.Rows(indexes)

	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}

	// Check first row
	if rows[0].ToolName != "Tool 1" {
		t.Errorf("Row 0 ToolName: got %q, want %q", rows[0].ToolName, "Tool 1")
	}
	if rows[0].ProviderName != "Provider 1" {
		t.Errorf("Row 0 ProviderName: got %q, want %q", rows[0].ProviderName, "Provider 1")
	}
	if !rows[0].UseProxy {
		t.Error("Row 0 UseProxy should be true")
	}

	// Check second row
	if rows[1].ToolName != "Tool 2" {
		t.Errorf("Row 1 ToolName: got %q, want %q", rows[1].ToolName, "Tool 2")
	}
	if rows[1].UseProxy {
		t.Error("Row 1 UseProxy should be false")
	}
}

func TestRows_NilBindings(t *testing.T) {
	svc := NewService(nil, &core.ToolsConfig{}, &core.ProvidersConfig{}, nil, "", "")

	indexes := []int{0}
	rows := svc.Rows(indexes)

	if rows != nil {
		t.Error("Rows should return nil when bindings is nil")
	}
}

func TestRows_EmptyIndexes(t *testing.T) {
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{ToolID: "tool1", ProviderID: "provider1"},
		},
	}
	svc := NewService(bindings, &core.ToolsConfig{}, &core.ProvidersConfig{}, nil, "", "")

	rows := svc.Rows([]int{})

	if rows != nil {
		t.Error("Rows should return nil when indexes is empty")
	}
}

func TestRows_InvalidIndex(t *testing.T) {
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{ToolID: "tool1", ProviderID: "provider1"},
		},
	}
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1", Name: "Tool 1"},
		},
	}
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	svc := NewService(bindings, tools, providers, nil, "", "")

	indexes := []int{0, 99, -1}
	rows := svc.Rows(indexes)

	// Should only return valid row
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row (skipping invalid indexes), got %d", len(rows))
	}
}

func TestRows_ToolNotFound(t *testing.T) {
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{ToolID: "tool1", ProviderID: "provider1"},
		},
	}
	tools := &core.ToolsConfig{
		Tools: []core.Tool{}, // Empty tools
	}
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	svc := NewService(bindings, tools, providers, nil, "", "")

	indexes := []int{0}
	rows := svc.Rows(indexes)

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	if rows[0].ToolName != "tool1" {
		t.Errorf("ToolName should be tool1 (ID), got %q", rows[0].ToolName)
	}
}

func TestRows_ProviderNotFound(t *testing.T) {
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{ToolID: "tool1", ProviderID: "provider1"},
		},
	}
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1", Name: "Tool 1"},
		},
	}
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{}, // Empty providers
	}
	svc := NewService(bindings, tools, providers, nil, "", "")

	indexes := []int{0}
	rows := svc.Rows(indexes)

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	if rows[0].ProviderName != "provider1" {
		t.Errorf("ProviderName should be provider1 (ID), got %q", rows[0].ProviderName)
	}
}

func TestDetails_Success(t *testing.T) {
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{
				ToolID:     "tool1",
				ProviderID: "provider1",
				Options: core.BindingOptions{
					Model: "gpt-4",
				},
			},
		},
	}
	svc := NewService(bindings, &core.ToolsConfig{}, &core.ProvidersConfig{}, nil, "", "")

	indexes := []int{0}
	details := svc.Details(indexes, 0)

	if details == nil {
		t.Fatal("Details should not be nil")
	}

	if details.ToolID != "tool1" {
		t.Errorf("ToolID: got %q, want %q", details.ToolID, "tool1")
	}
	if details.ProviderID != "provider1" {
		t.Errorf("ProviderID: got %q, want %q", details.ProviderID, "provider1")
	}
}

func TestDetails_NilBindings(t *testing.T) {
	svc := NewService(nil, &core.ToolsConfig{}, &core.ProvidersConfig{}, nil, "", "")

	indexes := []int{0}
	details := svc.Details(indexes, 0)

	if details != nil {
		t.Error("Details should return nil when bindings is nil")
	}
}

func TestDetails_EmptyIndexes(t *testing.T) {
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{ToolID: "tool1", ProviderID: "provider1"},
		},
	}
	svc := NewService(bindings, &core.ToolsConfig{}, &core.ProvidersConfig{}, nil, "", "")

	details := svc.Details([]int{}, 0)

	if details != nil {
		t.Error("Details should return nil when indexes is empty")
	}
}

func TestDetails_InvalidSelectedIndex(t *testing.T) {
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{ToolID: "tool1", ProviderID: "provider1"},
		},
	}
	svc := NewService(bindings, &core.ToolsConfig{}, &core.ProvidersConfig{}, nil, "", "")

	indexes := []int{0}
	details := svc.Details(indexes, 99)

	if details != nil {
		t.Error("Details should return nil for invalid selected index")
	}
}

func TestDetails_NegativeSelectedIndex(t *testing.T) {
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{ToolID: "tool1", ProviderID: "provider1"},
		},
	}
	svc := NewService(bindings, &core.ToolsConfig{}, &core.ProvidersConfig{}, nil, "", "")

	indexes := []int{0}
	details := svc.Details(indexes, -1)

	if details != nil {
		t.Error("Details should return nil for negative selected index")
	}
}

func TestSave_NilForm(t *testing.T) {
	bindings := &core.BindingsConfig{}
	svc := NewService(bindings, &core.ToolsConfig{}, &core.ProvidersConfig{}, nil, "", "")

	err := svc.Save("/tmp")

	if err == nil {
		t.Error("Save should fail when form is nil")
	}
}

func TestSave_NilBindings(t *testing.T) {
	form := &forms.BindingForm{}
	svc := NewService(nil, &core.ToolsConfig{}, &core.ProvidersConfig{}, form, "", "")

	err := svc.Save("/tmp")

	if err == nil {
		t.Error("Save should fail when bindings is nil")
	}
}

func TestEmptyStateMessage_Empty_NoSearch(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, "", "")

	msg := svc.EmptyStateMessage(true, false)

	if msg != "No bindings configured." {
		t.Errorf("Message: got %q, want %q", msg, "No bindings configured.")
	}
}

func TestEmptyStateMessage_Empty_WithSearch(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, "", "")

	msg := svc.EmptyStateMessage(true, true)

	if msg != "No bindings match the current filter." {
		t.Errorf("Message: got %q, want %q", msg, "No bindings match the current filter.")
	}
}

func TestEmptyStateMessage_NotEmpty(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, "", "")

	msg := svc.EmptyStateMessage(false, false)

	if msg != "" {
		t.Errorf("Message: got %q, want empty string", msg)
	}
}
