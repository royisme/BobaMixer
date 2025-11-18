package providers

import (
	"testing"

	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/ui/forms"
)

func TestNewService(t *testing.T) {
	providers := &core.ProvidersConfig{}
	secrets := &core.SecretsConfig{}
	form := &forms.ProviderForm{}
	msgNoSelection := "no selection"
	msgInvalid := "invalid"

	svc := NewService(providers, secrets, form, msgNoSelection, msgInvalid)

	if svc == nil {
		t.Fatal("expected service to be created")
	}
	if svc.providers != providers {
		t.Error("providers not set correctly")
	}
	if svc.secrets != secrets {
		t.Error("secrets not set correctly")
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

func TestStartForm_AddMode(t *testing.T) {
	providers := &core.ProvidersConfig{}
	form := forms.NewProviderForm("> ")
	svc := NewService(providers, &core.SecretsConfig{}, &form, "no selection", "invalid")

	result := svc.StartForm(true, []int{}, 0)

	if !result {
		t.Error("StartForm should return true in add mode")
	}
	if !form.AddMode() {
		t.Error("Form should be in add mode")
	}
}

func TestStartForm_EditMode_Success(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{
				ID:          "provider1",
				DisplayName: "Provider 1",
				Kind:        "openai",
				BaseURL:     "https://api.openai.com/v1",
			},
		},
	}
	form := forms.NewProviderForm("> ")
	svc := NewService(providers, &core.SecretsConfig{}, &form, "no selection", "invalid")

	indexes := []int{0}
	result := svc.StartForm(false, indexes, 0)

	if !result {
		t.Error("StartForm should return true for valid edit")
	}
	if form.AddMode() {
		t.Error("Form should not be in add mode")
	}
}

func TestStartForm_EditMode_NoSelection(t *testing.T) {
	providers := &core.ProvidersConfig{}
	form := forms.NewProviderForm("> ")
	svc := NewService(providers, &core.SecretsConfig{}, &form, "no selection", "invalid")

	result := svc.StartForm(false, []int{}, 0)

	if result {
		t.Error("StartForm should return false when no selection")
	}
}

func TestStartForm_EditMode_InvalidIndex(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	form := forms.NewProviderForm("> ")
	svc := NewService(providers, &core.SecretsConfig{}, &form, "no selection", "invalid")

	indexes := []int{0}
	result := svc.StartForm(false, indexes, 5)

	if result {
		t.Error("StartForm should return false for invalid index")
	}
}

func TestStartForm_EditMode_NilProviders(t *testing.T) {
	form := forms.NewProviderForm("> ")
	svc := NewService(nil, &core.SecretsConfig{}, &form, "no selection", "invalid")

	indexes := []int{0}
	result := svc.StartForm(false, indexes, 0)

	if result {
		t.Error("StartForm should return false when providers is nil")
	}
}

func TestStartForm_NilForm(t *testing.T) {
	providers := &core.ProvidersConfig{}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "no selection", "invalid")

	result := svc.StartForm(true, []int{}, 0)

	if result {
		t.Error("StartForm should return false when form is nil")
	}
}

func TestSave_AddMode(t *testing.T) {
	t.Skip("Skipping form-dependent test - requires form mock or refactor")
}

func TestSave_EditMode(t *testing.T) {
	t.Skip("Skipping form-dependent test - requires form mock or refactor")
}

func TestSave_EmptyID(t *testing.T) {
	t.Skip("Skipping form-dependent test - requires form mock or refactor")
}

func TestSave_NilForm(t *testing.T) {
	tmpDir := t.TempDir()
	providers := &core.ProvidersConfig{}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "no selection", "invalid")

	err := svc.Save(tmpDir)

	if err == nil {
		t.Error("Save should fail when form is nil")
	}
}

func TestSave_NilProviders(t *testing.T) {
	tmpDir := t.TempDir()
	form := forms.NewProviderForm("> ")
	svc := NewService(nil, &core.SecretsConfig{}, &form, "no selection", "invalid")

	err := svc.Save(tmpDir)

	if err == nil {
		t.Error("Save should fail when providers is nil")
	}
}

func TestSave_InvalidIndex(t *testing.T) {
	t.Skip("Skipping form-dependent test - requires form mock or refactor")
}

func TestRows_Success(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{
				ID:           "provider1",
				DisplayName:  "Provider 1",
				BaseURL:      "https://api1.com",
				DefaultModel: "model-1",
				Enabled:      true,
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceSecrets,
				},
			},
			{
				ID:           "provider2",
				DisplayName:  "Provider 2",
				BaseURL:      "https://api2.com",
				DefaultModel: "model-2",
				Enabled:      false,
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceSecrets,
				},
			},
		},
	}
	secrets := &core.SecretsConfig{
		Secrets: map[string]core.Secret{
			"provider1": {APIKey: "key1"},
		},
	}
	svc := NewService(providers, secrets, nil, "", "")

	indexes := []int{0, 1}
	rows := svc.Rows(indexes)

	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}

	// Check first row
	if rows[0].DisplayName != "Provider 1" {
		t.Errorf("Row 0 DisplayName: got %q, want %q", rows[0].DisplayName, "Provider 1")
	}
	if rows[0].BaseURL != "https://api1.com" {
		t.Errorf("Row 0 BaseURL: got %q, want %q", rows[0].BaseURL, "https://api1.com")
	}
	if rows[0].DefaultModel != "model-1" {
		t.Errorf("Row 0 DefaultModel: got %q, want %q", rows[0].DefaultModel, "model-1")
	}
	if !rows[0].Enabled {
		t.Error("Row 0 should be enabled")
	}
	if !rows[0].HasAPIKey {
		t.Error("Row 0 should have API key")
	}

	// Check second row
	if rows[1].DisplayName != "Provider 2" {
		t.Errorf("Row 1 DisplayName: got %q, want %q", rows[1].DisplayName, "Provider 2")
	}
	if rows[1].Enabled {
		t.Error("Row 1 should not be enabled")
	}
	if rows[1].HasAPIKey {
		t.Error("Row 1 should not have API key")
	}
}

func TestRows_NilProviders(t *testing.T) {
	svc := NewService(nil, &core.SecretsConfig{}, nil, "", "")

	indexes := []int{0}
	rows := svc.Rows(indexes)

	if rows != nil {
		t.Error("Rows should return nil when providers is nil")
	}
}

func TestRows_EmptyIndexes(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "", "")

	rows := svc.Rows([]int{})

	if rows != nil {
		t.Error("Rows should return nil when indexes is empty")
	}
}

func TestRows_InvalidIndex(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "", "")

	indexes := []int{0, 99, -1}
	rows := svc.Rows(indexes)

	// Should only return valid row
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row (skipping invalid indexes), got %d", len(rows))
	}
}

func TestRows_NilSecrets(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	svc := NewService(providers, nil, nil, "", "")

	indexes := []int{0}
	rows := svc.Rows(indexes)

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	// Should have no API key when secrets is nil
	if rows[0].HasAPIKey {
		t.Error("Row should not have API key when secrets is nil")
	}
}

func TestDetails_Success(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{
				ID:      "provider1",
				Kind:    "openai",
				BaseURL: "https://api.openai.com/v1",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceEnv,
					EnvVar: "OPENAI_API_KEY",
				},
			},
		},
	}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "", "")

	indexes := []int{0}
	details := svc.Details(indexes, 0)

	if details == nil {
		t.Fatal("Details should not be nil")
	}

	if details.ID != "provider1" {
		t.Errorf("ID: got %q, want %q", details.ID, "provider1")
	}
	if details.Kind != "openai" {
		t.Errorf("Kind: got %q, want %q", details.Kind, "openai")
	}
	if details.APIKeySource != string(core.APIKeySourceEnv) {
		t.Errorf("APIKeySource: got %q, want %q", details.APIKeySource, string(core.APIKeySourceEnv))
	}
	if !details.ShowEnvVar {
		t.Error("ShowEnvVar should be true")
	}
	if details.EnvVar != "OPENAI_API_KEY" {
		t.Errorf("EnvVar: got %q, want %q", details.EnvVar, "OPENAI_API_KEY")
	}
}

func TestDetails_SecretsSource(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{
				ID:   "provider1",
				Kind: "openai",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceSecrets,
				},
			},
		},
	}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "", "")

	indexes := []int{0}
	details := svc.Details(indexes, 0)

	if details == nil {
		t.Fatal("Details should not be nil")
	}

	if details.ShowEnvVar {
		t.Error("ShowEnvVar should be false for secrets source")
	}
	if details.EnvVar != "" {
		t.Error("EnvVar should be empty for secrets source")
	}
}

func TestDetails_NilProviders(t *testing.T) {
	svc := NewService(nil, &core.SecretsConfig{}, nil, "", "")

	indexes := []int{0}
	details := svc.Details(indexes, 0)

	if details != nil {
		t.Error("Details should return nil when providers is nil")
	}
}

func TestDetails_EmptyIndexes(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1"},
		},
	}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "", "")

	details := svc.Details([]int{}, 0)

	if details != nil {
		t.Error("Details should return nil when indexes is empty")
	}
}

func TestDetails_InvalidSelectedIndex(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1"},
		},
	}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "", "")

	indexes := []int{0}
	details := svc.Details(indexes, 99)

	if details != nil {
		t.Error("Details should return nil for invalid selected index")
	}
}

func TestDetails_NegativeSelectedIndex(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1"},
		},
	}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "", "")

	indexes := []int{0}
	details := svc.Details(indexes, -1)

	if details != nil {
		t.Error("Details should return nil for negative selected index")
	}
}

func TestDetails_InvalidProviderIndex(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1"},
		},
	}
	svc := NewService(providers, &core.SecretsConfig{}, nil, "", "")

	indexes := []int{99}
	details := svc.Details(indexes, 0)

	if details != nil {
		t.Error("Details should return nil for invalid provider index")
	}
}

func TestEmptyStateMessage_Empty_NoSearch(t *testing.T) {
	svc := NewService(nil, nil, nil, "", "")

	msg := svc.EmptyStateMessage(true, false)

	if msg != "No providers configured." {
		t.Errorf("Message: got %q, want %q", msg, "No providers configured.")
	}
}

func TestEmptyStateMessage_Empty_WithSearch(t *testing.T) {
	svc := NewService(nil, nil, nil, "", "")

	msg := svc.EmptyStateMessage(true, true)

	if msg != "No providers match the current filter." {
		t.Errorf("Message: got %q, want %q", msg, "No providers match the current filter.")
	}
}

func TestEmptyStateMessage_NotEmpty(t *testing.T) {
	svc := NewService(nil, nil, nil, "", "")

	msg := svc.EmptyStateMessage(false, false)

	if msg != "" {
		t.Errorf("Message: got %q, want empty string", msg)
	}
}
