package secrets

import (
	"testing"

	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/ui/forms"
)

func TestNewService(t *testing.T) {
	providers := &core.ProvidersConfig{}
	secrets := &core.SecretsConfig{}
	form := &forms.SecretForm{}
	message := ""
	msgNoSelection := "no selection"
	msgInvalid := "invalid"

	svc := NewService(providers, &secrets, form, &message, msgNoSelection, msgInvalid)

	if svc == nil {
		t.Fatal("expected service to be created")
	}
}

func TestStartForm_EmptyIndexes(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	secrets := &core.SecretsConfig{}
	form := &forms.SecretForm{}
	message := ""
	svc := NewService(providers, &secrets, form, &message, "no selection", "invalid")

	result := svc.StartForm([]int{}, 0)

	if result {
		t.Error("StartForm should return false when indexes is empty")
	}
}

func TestStartForm_InvalidSelectedIndex(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	secrets := &core.SecretsConfig{}
	form := &forms.SecretForm{}
	message := ""
	svc := NewService(providers, &secrets, form, &message, "no selection", "invalid")

	indexes := []int{0}
	result := svc.StartForm(indexes, 99)

	if result {
		t.Error("StartForm should return false for invalid selected index")
	}
}

func TestStartForm_NegativeIndex(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	secrets := &core.SecretsConfig{}
	form := &forms.SecretForm{}
	message := ""
	svc := NewService(providers, &secrets, form, &message, "no selection", "invalid")

	indexes := []int{0}
	result := svc.StartForm(indexes, -1)

	if result {
		t.Error("StartForm should return false for negative index")
	}
}

func TestRows_Success(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{
				ID:          "provider1",
				DisplayName: "Provider 1",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceSecrets,
				},
			},
			{
				ID:          "provider2",
				DisplayName: "Provider 2",
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
	svc := NewService(providers, &secrets, nil, nil, "", "")

	indexes := []int{0, 1}
	rows := svc.Rows(indexes)

	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}

	// Check first row (has secret)
	if rows[0].DisplayName != "Provider 1" {
		t.Errorf("Row 0 DisplayName: got %q, want %q", rows[0].DisplayName, "Provider 1")
	}
	if !rows[0].HasKey {
		t.Error("Row 0 should have key")
	}

	// Check second row (no secret)
	if rows[1].DisplayName != "Provider 2" {
		t.Errorf("Row 1 DisplayName: got %q, want %q", rows[1].DisplayName, "Provider 2")
	}
	if rows[1].HasKey {
		t.Error("Row 1 should not have key")
	}
}

func TestRows_NilProviders(t *testing.T) {
	secrets := &core.SecretsConfig{}
	svc := NewService(nil, &secrets, nil, nil, "", "")

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
	secrets := &core.SecretsConfig{}
	svc := NewService(providers, &secrets, nil, nil, "", "")

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
	secrets := &core.SecretsConfig{}
	svc := NewService(providers, &secrets, nil, nil, "", "")

	indexes := []int{0, 99, -1}
	rows := svc.Rows(indexes)

	// Should only return valid row
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row (skipping invalid indexes), got %d", len(rows))
	}
}

func TestRows_NilSecretsConfig(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{ID: "provider1", DisplayName: "Provider 1"},
		},
	}
	var secrets *core.SecretsConfig
	svc := NewService(providers, &secrets, nil, nil, "", "")

	indexes := []int{0}
	rows := svc.Rows(indexes)

	if rows != nil {
		t.Error("Rows should return nil when secrets pointer is nil")
	}
}

func TestRows_EmptySecrets(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{
				ID:          "provider1",
				DisplayName: "Provider 1",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceSecrets,
				},
			},
			{
				ID:          "provider2",
				DisplayName: "Provider 2",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceSecrets,
				},
			},
		},
	}
	secrets := &core.SecretsConfig{
		Secrets: map[string]core.Secret{},
	}
	svc := NewService(providers, &secrets, nil, nil, "", "")

	indexes := []int{0, 1}
	rows := svc.Rows(indexes)

	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}

	// Both should not have keys
	for i, row := range rows {
		if row.HasKey {
			t.Errorf("Row %d should not have key when secrets is empty", i)
		}
	}
}

func TestRows_MultipleProviders(t *testing.T) {
	providers := &core.ProvidersConfig{
		Providers: []core.Provider{
			{
				ID:          "provider1",
				DisplayName: "Provider 1",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceSecrets,
				},
			},
			{
				ID:          "provider2",
				DisplayName: "Provider 2",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceSecrets,
				},
			},
			{
				ID:          "provider3",
				DisplayName: "Provider 3",
				APIKey: core.APIKeyConfig{
					Source: core.APIKeySourceSecrets,
				},
			},
		},
	}
	secrets := &core.SecretsConfig{
		Secrets: map[string]core.Secret{
			"provider1": {APIKey: "key1"},
			"provider3": {APIKey: "key3"},
		},
	}
	svc := NewService(providers, &secrets, nil, nil, "", "")

	indexes := []int{0, 1, 2}
	rows := svc.Rows(indexes)

	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}

	// Provider 1 has key
	if !rows[0].HasKey {
		t.Error("Provider 1 should have key")
	}

	// Provider 2 does not have key
	if rows[1].HasKey {
		t.Error("Provider 2 should not have key")
	}

	// Provider 3 has key
	if !rows[2].HasKey {
		t.Error("Provider 3 should have key")
	}
}
