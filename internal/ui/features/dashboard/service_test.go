package dashboard

import (
	"testing"

	"github.com/royisme/bobamixer/internal/domain/core"
)

const (
	testToolName = "Test Tool"
)

func TestNewService(t *testing.T) {
	tools := &core.ToolsConfig{}
	bindings := &core.BindingsConfig{}
	providers := &core.ProvidersConfig{}
	secrets := &core.SecretsConfig{}

	svc := NewService(tools, bindings, providers, secrets)
	if svc == nil {
		t.Fatal("expected service to be created")
	}
	if svc.tools != tools {
		t.Error("tools not set correctly")
	}
	if svc.bindings != bindings {
		t.Error("bindings not set correctly")
	}
	if svc.providers != providers {
		t.Error("providers not set correctly")
	}
	if svc.secrets != secrets {
		t.Error("secrets not set correctly")
	}
}

func TestBuildTableRows_NoTools(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{Tools: []core.Tool{}},
		&core.BindingsConfig{Bindings: []core.Binding{}},
		&core.ProvidersConfig{Providers: []core.Provider{}},
		&core.SecretsConfig{},
	)

	rows := svc.BuildTableRows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row (empty message), got %d", len(rows))
	}

	if rows[0][0] != "No tools configured" {
		t.Errorf("expected empty message, got %q", rows[0][0])
	}
}

func TestBuildTableRows_ToolNotBound(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{
			Tools: []core.Tool{
				{ID: "tool1", Name: "Test Tool"},
			},
		},
		&core.BindingsConfig{Bindings: []core.Binding{}},
		&core.ProvidersConfig{Providers: []core.Provider{}},
		&core.SecretsConfig{},
	)

	rows := svc.BuildTableRows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row[0] != testToolName {
		t.Errorf("Tool name: got %q, want %q", row[0], testToolName)
	}
	if row[1] != "(not bound)" {
		t.Errorf("Provider: got %q, want %q", row[1], "(not bound)")
	}
	if row[4] != IconWarning+" Not configured" {
		t.Errorf("Status: got %q, want %q", row[4], IconWarning+" Not configured")
	}
}

func TestBuildTableRows_ProviderMissing(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{
			Tools: []core.Tool{
				{ID: "tool1", Name: testToolName},
			},
		},
		&core.BindingsConfig{
			Bindings: []core.Binding{
				{ToolID: "tool1", ProviderID: "missing-provider"},
			},
		},
		&core.ProvidersConfig{Providers: []core.Provider{}},
		&core.SecretsConfig{},
	)

	rows := svc.BuildTableRows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row[0] != testToolName {
		t.Errorf("Tool name: got %q, want %q", row[0], testToolName)
	}
	if row[1] != "(missing: missing-provider)" {
		t.Errorf("Provider: got %q, want %q", row[1], "(missing: missing-provider)")
	}
	if row[4] != IconCross+" Error" {
		t.Errorf("Status: got %q, want %q", row[4], IconCross+" Error")
	}
}

func TestBuildTableRows_FullConfiguration(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{
			Tools: []core.Tool{
				{ID: "tool1", Name: "Test Tool"},
			},
		},
		&core.BindingsConfig{
			Bindings: []core.Binding{
				{
					ToolID:     "tool1",
					ProviderID: "provider1",
					UseProxy:   true,
					Options: core.BindingOptions{
						Model: "gpt-4",
					},
				},
			},
		},
		&core.ProvidersConfig{
			Providers: []core.Provider{
				{
					ID:           "provider1",
					DisplayName:  "Test Provider",
					Kind:         "openai",
					BaseURL:      "https://api.openai.com/v1",
					DefaultModel: "gpt-3.5-turbo",
					APIKey: core.APIKeyConfig{
						Source: core.APIKeySourceSecrets,
					},
				},
			},
		},
		&core.SecretsConfig{
			Secrets: map[string]core.Secret{
				"provider1": {
					APIKey: "sk-test123",
				},
			},
		},
	)

	rows := svc.BuildTableRows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row[0] != "Test Tool" {
		t.Errorf("Tool name: got %q, want %q", row[0], "Test Tool")
	}
	if row[1] != "Test Provider" {
		t.Errorf("Provider: got %q, want %q", row[1], "Test Provider")
	}
	if row[2] != "gpt-4" {
		t.Errorf("Model: got %q, want %q", row[2], "gpt-4")
	}
	if row[3] != ProxyStateOn {
		t.Errorf("Proxy: got %q, want %q", row[3], ProxyStateOn)
	}
	if row[4] != IconCheckmark+" Ready" {
		t.Errorf("Status: got %q, want %q", row[4], IconCheckmark+" Ready")
	}
}

func TestBuildTableRows_DefaultModel(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{
			Tools: []core.Tool{
				{ID: "tool1", Name: "Test Tool"},
			},
		},
		&core.BindingsConfig{
			Bindings: []core.Binding{
				{
					ToolID:     "tool1",
					ProviderID: "provider1",
					Options:    core.BindingOptions{}, // No model override
				},
			},
		},
		&core.ProvidersConfig{
			Providers: []core.Provider{
				{
					ID:           "provider1",
					DisplayName:  "Test Provider",
					Kind:         "openai",
					BaseURL:      "https://api.openai.com/v1",
					DefaultModel: "gpt-3.5-turbo",
					APIKey: core.APIKeyConfig{
						Source: core.APIKeySourceSecrets,
					},
				},
			},
		},
		&core.SecretsConfig{
			Secrets: map[string]core.Secret{
				"provider1": {
					APIKey: "sk-test123",
				},
			},
		},
	)

	rows := svc.BuildTableRows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row[2] != "gpt-3.5-turbo" {
		t.Errorf("Model: got %q, want %q (should use default)", row[2], "gpt-3.5-turbo")
	}
}

func TestBuildTableRows_NoAPIKey(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{
			Tools: []core.Tool{
				{ID: "tool1", Name: "Test Tool"},
			},
		},
		&core.BindingsConfig{
			Bindings: []core.Binding{
				{
					ToolID:     "tool1",
					ProviderID: "provider1",
				},
			},
		},
		&core.ProvidersConfig{
			Providers: []core.Provider{
				{
					ID:           "provider1",
					DisplayName:  "Test Provider",
					Kind:         "openai",
					BaseURL:      "https://api.openai.com/v1",
					DefaultModel: "gpt-3.5-turbo",
					APIKey: core.APIKeyConfig{
						Source: core.APIKeySourceSecrets,
					},
				},
			},
		},
		&core.SecretsConfig{
			Secrets: map[string]core.Secret{},
		},
	)

	rows := svc.BuildTableRows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row[4] != IconWarning+" No API key" {
		t.Errorf("Status: got %q, want %q", row[4], IconWarning+" No API key")
	}
}

func TestBuildTableRows_ProxyOff(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{
			Tools: []core.Tool{
				{ID: "tool1", Name: "Test Tool"},
			},
		},
		&core.BindingsConfig{
			Bindings: []core.Binding{
				{
					ToolID:     "tool1",
					ProviderID: "provider1",
					UseProxy:   false, // Proxy disabled
				},
			},
		},
		&core.ProvidersConfig{
			Providers: []core.Provider{
				{
					ID:           "provider1",
					DisplayName:  "Test Provider",
					Kind:         "openai",
					BaseURL:      "https://api.openai.com/v1",
					DefaultModel: "gpt-3.5-turbo",
					APIKey: core.APIKeyConfig{
						Source: core.APIKeySourceSecrets,
					},
				},
			},
		},
		&core.SecretsConfig{
			Secrets: map[string]core.Secret{
				"provider1": {
					APIKey: "sk-test123",
				},
			},
		},
	)

	rows := svc.BuildTableRows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row[3] != ProxyStateOff {
		t.Errorf("Proxy: got %q, want %q", row[3], ProxyStateOff)
	}
}

func TestBuildTableRows_MultipleTools(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{
			Tools: []core.Tool{
				{ID: "tool1", Name: "Tool 1"},
				{ID: "tool2", Name: "Tool 2"},
				{ID: "tool3", Name: "Tool 3"},
			},
		},
		&core.BindingsConfig{
			Bindings: []core.Binding{
				{ToolID: "tool1", ProviderID: "provider1"},
				{ToolID: "tool2", ProviderID: "provider1"},
				// tool3 not bound
			},
		},
		&core.ProvidersConfig{
			Providers: []core.Provider{
				{
					ID:           "provider1",
					DisplayName:  "Provider",
					Kind:         "openai",
					BaseURL:      "https://api.openai.com/v1",
					DefaultModel: "gpt-4",
					APIKey: core.APIKeyConfig{
						Source: core.APIKeySourceSecrets,
					},
				},
			},
		},
		&core.SecretsConfig{
			Secrets: map[string]core.Secret{
				"provider1": {
					APIKey: "sk-test",
				},
			},
		},
	)

	rows := svc.BuildTableRows()

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	// Tool 1 and 2 should be configured
	if rows[0][0] != "Tool 1" {
		t.Errorf("Row 0 tool name: got %q, want %q", rows[0][0], "Tool 1")
	}
	if rows[1][0] != "Tool 2" {
		t.Errorf("Row 1 tool name: got %q, want %q", rows[1][0], "Tool 2")
	}

	// Tool 3 should show as not configured
	if rows[2][0] != "Tool 3" {
		t.Errorf("Row 2 tool name: got %q, want %q", rows[2][0], "Tool 3")
	}
	if rows[2][1] != "(not bound)" {
		t.Errorf("Row 2 should be not bound, got %q", rows[2][1])
	}
}

func TestDetermineModel_LongModel(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{},
		&core.BindingsConfig{},
		&core.ProvidersConfig{},
		&core.SecretsConfig{},
	)

	provider := &core.Provider{
		DefaultModel: "short",
	}
	binding := &core.Binding{
		Options: core.BindingOptions{
			Model: "this-is-a-very-long-model-name-that-exceeds-twenty-three-characters",
		},
	}

	model := svc.determineModel(provider, binding)

	if len(model) > 23 {
		t.Errorf("Model should be truncated to 23 chars, got %d chars: %q", len(model), model)
	}
	if model[len(model)-3:] != "..." {
		t.Errorf("Truncated model should end with '...', got %q", model)
	}
	expectedPrefix := "this-is-a-very-long-"
	if model[:20] != expectedPrefix {
		t.Errorf("Model prefix: got %q, want %q", model[:20], expectedPrefix)
	}
}

func TestGetNavigationHelp(t *testing.T) {
	svc := NewService(nil, nil, nil, nil)
	help := svc.GetNavigationHelp()
	if help == "" {
		t.Error("GetNavigationHelp should return non-empty string")
	}
	if help != HelpTextNavigation {
		t.Errorf("GetNavigationHelp: got %q, want %q", help, HelpTextNavigation)
	}
}

func TestGetActionHelp(t *testing.T) {
	svc := NewService(nil, nil, nil, nil)
	help := svc.GetActionHelp()
	if help == "" {
		t.Error("GetActionHelp should return non-empty string")
	}
	if help != HelpTextActions {
		t.Errorf("GetActionHelp: got %q, want %q", help, HelpTextActions)
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are defined
	tests := []struct {
		name  string
		value string
	}{
		{"IconCircleFilled", IconCircleFilled},
		{"IconCircleEmpty", IconCircleEmpty},
		{"IconCheckmark", IconCheckmark},
		{"IconCross", IconCross},
		{"IconWarning", IconWarning},
		{"ProxyStateOn", ProxyStateOn},
		{"ProxyStateOff", ProxyStateOff},
		{"HelpTextNavigation", HelpTextNavigation},
		{"HelpTextActions", HelpTextActions},
		{"MsgNoProviderSelected", MsgNoProviderSelected},
		{"MsgInvalidProvider", MsgInvalidProvider},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("%s constant is empty", tt.name)
			}
		})
	}
}

func TestBuildTableRows_ReturnType(t *testing.T) {
	svc := NewService(
		&core.ToolsConfig{
			Tools: []core.Tool{
				{ID: "tool1", Name: "Test Tool"},
			},
		},
		&core.BindingsConfig{
			Bindings: []core.Binding{
				{ToolID: "tool1", ProviderID: "provider1"},
			},
		},
		&core.ProvidersConfig{
			Providers: []core.Provider{
				{
					ID:           "provider1",
					DisplayName:  "Provider",
					Kind:         "openai",
					BaseURL:      "https://api.openai.com/v1",
					DefaultModel: "gpt-4",
					APIKey: core.APIKeyConfig{
						Source: core.APIKeySourceSecrets,
					},
				},
			},
		},
		&core.SecretsConfig{
			Secrets: map[string]core.Secret{
				"provider1": {
					APIKey: "sk-test",
				},
			},
		},
	)

	rows := svc.BuildTableRows()

	// Verify return type is []table.Row
	var _ = rows

	// Verify each row has 5 columns
	for i, row := range rows {
		if len(row) != 5 {
			t.Errorf("Row %d has %d columns, want 5", i, len(row))
		}
	}
}
