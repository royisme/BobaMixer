package tools

import (
	"testing"

	"github.com/royisme/bobamixer/internal/domain/core"
)

func TestNewService(t *testing.T) {
	tools := &core.ToolsConfig{}
	bindings := &core.BindingsConfig{}

	svc := NewService(tools, bindings)

	if svc == nil {
		t.Fatal("expected service to be created")
	}
	if svc.tools != tools {
		t.Error("tools not set correctly")
	}
	if svc.bindings != bindings {
		t.Error("bindings not set correctly")
	}
}

func TestRows_Success(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{
				ID:   "tool1",
				Name: "Tool 1",
				Exec: "tool1",
				Kind: "cli",
			},
			{
				ID:   "tool2",
				Name: "Tool 2",
				Exec: "tool2",
				Kind: "service",
			},
		},
	}
	bindings := &core.BindingsConfig{
		Bindings: []core.Binding{
			{ToolID: "tool1", ProviderID: "provider1"},
		},
	}
	svc := NewService(tools, bindings)

	indexes := []int{0, 1}
	rows := svc.Rows(indexes)

	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}

	// Check first row (bound)
	if rows[0].Name != "Tool 1" {
		t.Errorf("Row 0 Name: got %q, want %q", rows[0].Name, "Tool 1")
	}
	if rows[0].Exec != "tool1" {
		t.Errorf("Row 0 Exec: got %q, want %q", rows[0].Exec, "tool1")
	}
	if rows[0].Kind != "cli" {
		t.Errorf("Row 0 Kind: got %q, want %q", rows[0].Kind, "cli")
	}
	if !rows[0].Bound {
		t.Error("Row 0 should be bound")
	}

	// Check second row (not bound)
	if rows[1].Name != "Tool 2" {
		t.Errorf("Row 1 Name: got %q, want %q", rows[1].Name, "Tool 2")
	}
	if rows[1].Bound {
		t.Error("Row 1 should not be bound")
	}
}

func TestRows_NilTools(t *testing.T) {
	svc := NewService(nil, &core.BindingsConfig{})

	indexes := []int{0}
	rows := svc.Rows(indexes)

	if rows != nil {
		t.Error("Rows should return nil when tools is nil")
	}
}

func TestRows_EmptyIndexes(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1", Name: "Tool 1"},
		},
	}
	svc := NewService(tools, &core.BindingsConfig{})

	rows := svc.Rows([]int{})

	if rows != nil {
		t.Error("Rows should return nil when indexes is empty")
	}
}

func TestRows_InvalidIndex(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1", Name: "Tool 1"},
		},
	}
	svc := NewService(tools, &core.BindingsConfig{})

	indexes := []int{0, 99, -1}
	rows := svc.Rows(indexes)

	// Should only return valid row
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row (skipping invalid indexes), got %d", len(rows))
	}
}

func TestRows_NilBindings(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1", Name: "Tool 1"},
		},
	}
	svc := NewService(tools, nil)

	indexes := []int{0}
	rows := svc.Rows(indexes)

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	// Should have Bound=false when bindings is nil
	if rows[0].Bound {
		t.Error("Row should not be bound when bindings is nil")
	}
}

func TestDetails_Success(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{
				ID:          "tool1",
				Name:        "Tool 1",
				ConfigType:  "yaml",
				ConfigPath:  "/path/to/config.yaml",
				Description: "A test tool",
			},
		},
	}
	svc := NewService(tools, &core.BindingsConfig{})

	indexes := []int{0}
	details := svc.Details(indexes, 0)

	if details == nil {
		t.Fatal("Details should not be nil")
	}

	if details.ID != "tool1" {
		t.Errorf("ID: got %q, want %q", details.ID, "tool1")
	}
	if details.ConfigType != "yaml" {
		t.Errorf("ConfigType: got %q, want %q", details.ConfigType, "yaml")
	}
	if details.ConfigPath != "/path/to/config.yaml" {
		t.Errorf("ConfigPath: got %q, want %q", details.ConfigPath, "/path/to/config.yaml")
	}
	if details.Description != "A test tool" {
		t.Errorf("Description: got %q, want %q", details.Description, "A test tool")
	}
}

func TestDetails_NilTools(t *testing.T) {
	svc := NewService(nil, &core.BindingsConfig{})

	indexes := []int{0}
	details := svc.Details(indexes, 0)

	if details != nil {
		t.Error("Details should return nil when tools is nil")
	}
}

func TestDetails_EmptyIndexes(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1"},
		},
	}
	svc := NewService(tools, &core.BindingsConfig{})

	details := svc.Details([]int{}, 0)

	if details != nil {
		t.Error("Details should return nil when indexes is empty")
	}
}

func TestDetails_InvalidSelectedIndex(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1"},
		},
	}
	svc := NewService(tools, &core.BindingsConfig{})

	indexes := []int{0}
	details := svc.Details(indexes, 99)

	if details != nil {
		t.Error("Details should return nil for invalid selected index")
	}
}

func TestDetails_NegativeSelectedIndex(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1"},
		},
	}
	svc := NewService(tools, &core.BindingsConfig{})

	indexes := []int{0}
	details := svc.Details(indexes, -1)

	if details != nil {
		t.Error("Details should return nil for negative selected index")
	}
}

func TestDetails_InvalidToolIndex(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1"},
		},
	}
	svc := NewService(tools, &core.BindingsConfig{})

	indexes := []int{99}
	details := svc.Details(indexes, 0)

	if details != nil {
		t.Error("Details should return nil for invalid tool index")
	}
}

func TestEmptyStateMessage_Empty_NoSearch(t *testing.T) {
	svc := NewService(nil, nil)

	msg := svc.EmptyStateMessage(true, false)

	if msg != "No tools configured." {
		t.Errorf("Message: got %q, want %q", msg, "No tools configured.")
	}
}

func TestEmptyStateMessage_Empty_WithSearch(t *testing.T) {
	svc := NewService(nil, nil)

	msg := svc.EmptyStateMessage(true, true)

	if msg != "No tools match the current filter." {
		t.Errorf("Message: got %q, want %q", msg, "No tools match the current filter.")
	}
}

func TestEmptyStateMessage_NotEmpty(t *testing.T) {
	svc := NewService(nil, nil)

	msg := svc.EmptyStateMessage(false, false)

	if msg != "" {
		t.Errorf("Message: got %q, want empty string", msg)
	}
}

func TestRows_AllToolKinds(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1", Name: "CLI Tool", Kind: "cli"},
			{ID: "tool2", Name: "Service Tool", Kind: "service"},
			{ID: "tool3", Name: "Script Tool", Kind: "script"},
		},
	}
	svc := NewService(tools, &core.BindingsConfig{})

	indexes := []int{0, 1, 2}
	rows := svc.Rows(indexes)

	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}

	expectedKinds := []string{"cli", "service", "script"}
	for i, expected := range expectedKinds {
		if rows[i].Kind != expected {
			t.Errorf("Row %d Kind: got %q, want %q", i, rows[i].Kind, expected)
		}
	}
}

func TestDetails_AllConfigTypes(t *testing.T) {
	tools := &core.ToolsConfig{
		Tools: []core.Tool{
			{ID: "tool1", ConfigType: "yaml"},
			{ID: "tool2", ConfigType: "json"},
			{ID: "tool3", ConfigType: "toml"},
		},
	}
	svc := NewService(tools, &core.BindingsConfig{})

	expectedTypes := []string{"yaml", "json", "toml"}
	for i, expected := range expectedTypes {
		indexes := []int{i}
		details := svc.Details(indexes, 0)

		if details == nil {
			t.Fatalf("Details should not be nil for tool %d", i)
		}

		if details.ConfigType != expected {
			t.Errorf("Tool %d ConfigType: got %q, want %q", i, details.ConfigType, expected)
		}
	}
}
