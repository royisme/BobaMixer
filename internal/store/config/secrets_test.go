package config

import (
	"testing"
)

func TestResolveEnv(t *testing.T) {
	secrets := Secrets{
		"anthropic":  "sk-ant-test-key",
		"openrouter": "sk-or-test-key",
	}

	tests := []struct {
		env      map[string]string
		expected map[string]string
		name     string
	}{
		{
			name: "resolve secret references",
			env: map[string]string{
				"ANTHROPIC_API_KEY":  "secret://anthropic",
				"OPENROUTER_API_KEY": "secret://openrouter",
			},
			expected: map[string]string{
				"ANTHROPIC_API_KEY":  "sk-ant-test-key",
				"OPENROUTER_API_KEY": "sk-or-test-key",
			},
		},
		{
			name: "keep non-secret values",
			env: map[string]string{
				"PLAIN_VAR":         "plain-value",
				"ANTHROPIC_API_KEY": "secret://anthropic",
			},
			expected: map[string]string{
				"PLAIN_VAR":         "plain-value",
				"ANTHROPIC_API_KEY": "sk-ant-test-key",
			},
		},
		{
			name: "handle missing secrets",
			env: map[string]string{
				"MISSING_KEY": "secret://missing",
			},
			expected: map[string]string{
				"MISSING_KEY": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveEnv(tt.env, secrets)

			// Convert result to map for easier comparison
			resultMap := make(map[string]string)
			for _, kv := range result {
				// Parse "KEY=VALUE" format
				for i := 0; i < len(kv); i++ {
					if kv[i] == '=' {
						key := kv[:i]
						value := kv[i+1:]
						resultMap[key] = value
						break
					}
				}
			}

			// Compare
			for key, expectedValue := range tt.expected {
				if gotValue, ok := resultMap[key]; !ok {
					t.Errorf("missing key %s", key)
				} else if gotValue != expectedValue {
					t.Errorf("key %s: got %q, want %q", key, gotValue, expectedValue)
				}
			}
		})
	}
}

func TestResolveSecretRef(t *testing.T) {
	secrets := Secrets{
		"test": "test-value",
	}

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "resolve secret reference",
			value:    "secret://test",
			expected: "test-value",
		},
		{
			name:     "keep plain value",
			value:    "plain-value",
			expected: "plain-value",
		},
		{
			name:     "missing secret returns empty",
			value:    "secret://missing",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveSecretRef(tt.value, secrets)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}
