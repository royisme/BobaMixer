package logger

import (
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "API key in header",
			input:    `x-api-key: sk-ant-1234567890abcdef`,
			expected: `x-api-key: ***REDACTED***`,
		},
		{
			name:     "Bearer token",
			input:    `Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9`,
			expected: `Authorization: Bearer ***REDACTED***`,
		},
		{
			name:     "Secret value",
			input:    `"secret": "my-secret-password"`,
			expected: `secret: ***REDACTED***`,
		},
		{
			name:     "Large request body",
			input:    `"messages": "this is a very long message that should be truncated because it exceeds fifty characters"`,
			expected: `"messages": "***TRUNCATED***"`,
		},
		{
			name:     "OpenAI API key",
			input:    `OPENAI_API_KEY=sk-1234567890abcdefghijklmnopqrstuvwxyz`,
			expected: `OPENAI_API_KEY: ***REDACTED***`,
		},
		{
			name:     "Safe content",
			input:    `Processing request with model claude-3-5-sonnet`,
			expected: `Processing request with model claude-3-5-sonnet`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.input)
			if !strings.Contains(result, "***REDACTED***") && strings.Contains(tt.expected, "***REDACTED***") {
				t.Errorf("Expected redaction but got: %s", result)
			}
			if !strings.Contains(result, "***TRUNCATED***") && strings.Contains(tt.expected, "***TRUNCATED***") {
				t.Errorf("Expected truncation but got: %s", result)
			}
			// For safe content, should remain unchanged
			if tt.expected == tt.input && result != tt.input {
				t.Errorf("Expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key       string
		sensitive bool
	}{
		{"api_key", true},
		{"apiKey", true},
		{"x-api-key", true},
		{"secret", true},
		{"password", true},
		{"token", true},
		{"authorization", true},
		{"bearer", true},
		{"payload", true},
		{"request_body", true},
		{"model", false},
		{"temperature", false},
		{"max_tokens", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := isSensitiveKey(tt.key)
			if result != tt.sensitive {
				t.Errorf("isSensitiveKey(%q) = %v, want %v", tt.key, result, tt.sensitive)
			}
		})
	}
}
