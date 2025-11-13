package pricing

import (
	"testing"

	"github.com/royisme/bobamixer/internal/store/config"
)

func TestGetPrice(t *testing.T) {
	table := &Table{
		Models: map[string]ModelPrice{
			"claude-3-5-sonnet": {
				InputPer1K:  0.015,
				OutputPer1K: 0.075,
			},
			"deepseek-chat": {
				InputPer1K:  0.0005,
				OutputPer1K: 0.002,
			},
		},
	}

	profileCost := config.Cost{
		Input:  0.01,
		Output: 0.05,
	}

	tests := []struct {
		name           string
		modelName      string
		expectedInput  float64
		expectedOutput float64
	}{
		{
			name:           "existing model",
			modelName:      "claude-3-5-sonnet",
			expectedInput:  0.015,
			expectedOutput: 0.075,
		},
		{
			name:           "another existing model",
			modelName:      "deepseek-chat",
			expectedInput:  0.0005,
			expectedOutput: 0.002,
		},
		{
			name:           "fallback to profile cost",
			modelName:      "unknown-model",
			expectedInput:  0.01,
			expectedOutput: 0.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := table.GetPrice(tt.modelName, profileCost)

			if price.InputPer1K != tt.expectedInput {
				t.Errorf("InputPer1K: got %f, want %f", price.InputPer1K, tt.expectedInput)
			}
			if price.OutputPer1K != tt.expectedOutput {
				t.Errorf("OutputPer1K: got %f, want %f", price.OutputPer1K, tt.expectedOutput)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	table := &Table{
		Models: map[string]ModelPrice{
			"test-model": {
				InputPer1K:  0.01,
				OutputPer1K: 0.02,
			},
		},
	}

	profileCost := config.Cost{
		Input:  0.01,
		Output: 0.02,
	}

	tests := []struct {
		name               string
		modelName          string
		inputTokens        int
		outputTokens       int
		expectedInputCost  float64
		expectedOutputCost float64
	}{
		{
			name:               "1k tokens each",
			modelName:          "test-model",
			inputTokens:        1000,
			outputTokens:       1000,
			expectedInputCost:  0.01,
			expectedOutputCost: 0.02,
		},
		{
			name:               "500 tokens each",
			modelName:          "test-model",
			inputTokens:        500,
			outputTokens:       500,
			expectedInputCost:  0.005,
			expectedOutputCost: 0.01,
		},
		{
			name:               "zero tokens",
			modelName:          "test-model",
			inputTokens:        0,
			outputTokens:       0,
			expectedInputCost:  0.0,
			expectedOutputCost: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputCost, outputCost := table.CalculateCost(
				tt.modelName,
				profileCost,
				tt.inputTokens,
				tt.outputTokens,
			)

			if inputCost != tt.expectedInputCost {
				t.Errorf("inputCost: got %f, want %f", inputCost, tt.expectedInputCost)
			}
			if outputCost != tt.expectedOutputCost {
				t.Errorf("outputCost: got %f, want %f", outputCost, tt.expectedOutputCost)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		home     string
		expected string
	}{
		{
			name:     "expand tilde",
			path:     "~/.boba/config.yaml",
			home:     "/home/user/.boba",
			expected: "/home/user/.boba/.boba/config.yaml",
		},
		{
			name:     "no tilde",
			path:     "/absolute/path",
			home:     "/home/user/.boba",
			expected: "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandHome(tt.path, tt.home)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}
