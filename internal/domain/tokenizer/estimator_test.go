package tokenizer

import (
	"strings"
	"testing"
)

func TestEstimate(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		text     string
		minToken int
		maxToken int
	}{
		{
			name:     "simple english text",
			model:    "gpt-4",
			text:     "Hello world, this is a test.",
			minToken: 4,
			maxToken: 10,
		},
		{
			name:     "code snippet",
			model:    "claude-3",
			text:     "func main() { fmt.Println(\"hello\") }",
			minToken: 6,
			maxToken: 15,
		},
		{
			name:     "empty text",
			model:    "gpt-4",
			text:     "",
			minToken: 0,
			maxToken: 0,
		},
		{
			name:     "long text",
			model:    "claude-3",
			text:     "The quick brown fox jumps over the lazy dog. " + strings.Repeat("This is a sentence. ", 10),
			minToken: 35,
			maxToken: 80,
		},
		{
			name:     "code with lots of symbols",
			model:    "gpt-4",
			text:     "const arr = [1, 2, 3].map(x => x * 2).filter(x => x > 2);",
			minToken: 15,
			maxToken: 35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimator := NewEstimator(tt.model)
			tokens := estimator.Estimate(tt.text)

			if tokens < tt.minToken || tokens > tt.maxToken {
				t.Errorf("Estimate() = %d, want between %d and %d", tokens, tt.minToken, tt.maxToken)
			}
		})
	}
}

func TestLooksLikeCode(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "go code",
			text:     "func main() { return nil }",
			expected: true,
		},
		{
			name:     "python code",
			text:     "def hello(): import os",
			expected: true,
		},
		{
			name:     "javascript code",
			text:     "const x = () => { let y = 10; }",
			expected: true,
		},
		{
			name:     "plain text",
			text:     "This is just a regular sentence.",
			expected: false,
		},
		{
			name:     "text with one code keyword",
			text:     "The function called main is important.",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := looksLikeCode(tt.text)
			if result != tt.expected {
				t.Errorf("looksLikeCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "simple sentence",
			text:     "hello world",
			expected: 2,
		},
		{
			name:     "with punctuation",
			text:     "Hello, world!",
			expected: 2,
		},
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "multiple spaces",
			text:     "hello    world",
			expected: 2,
		},
		{
			name:     "with numbers",
			text:     "test 123 456",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countWords(tt.text)
			if result != tt.expected {
				t.Errorf("countWords() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestEstimatePair(t *testing.T) {
	estimator := NewEstimator("gpt-4")

	input := "What is the capital of France?"
	output := "The capital of France is Paris."

	inputTokens, outputTokens := estimator.EstimatePair(input, output)

	if inputTokens <= 0 {
		t.Errorf("inputTokens should be > 0, got %d", inputTokens)
	}

	if outputTokens <= 0 {
		t.Errorf("outputTokens should be > 0, got %d", outputTokens)
	}

	// Output is slightly longer, should have more tokens
	if outputTokens <= inputTokens {
		t.Logf("Warning: outputTokens (%d) <= inputTokens (%d), expected output to be longer", outputTokens, inputTokens)
	}
}

func TestEstimateWithConfidence(t *testing.T) {
	estimator := NewEstimator("claude-3")

	tests := []struct {
		name               string
		text               string
		expectedConfidence Confidence
	}{
		{
			name:               "short text",
			text:               "Hello world",
			expectedConfidence: ConfidenceHigh,
		},
		{
			name:               "medium text",
			text:               "This is a medium length sentence with several words.",
			expectedConfidence: ConfidenceMedium,
		},
		{
			name:               "code text",
			text:               "func main() { var x int; return x }",
			expectedConfidence: ConfidenceLow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimator.EstimateWithConfidence(tt.text)

			if result.Tokens <= 0 {
				t.Errorf("Tokens should be > 0, got %d", result.Tokens)
			}

			if result.Confidence != tt.expectedConfidence {
				t.Errorf("Confidence = %s, want %s", result.Confidence, tt.expectedConfidence)
			}
		})
	}
}

func BenchmarkEstimate(b *testing.B) {
	estimator := NewEstimator("gpt-4")
	text := "The quick brown fox jumps over the lazy dog. " + strings.Repeat("This is a test sentence. ", 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		estimator.Estimate(text)
	}
}
