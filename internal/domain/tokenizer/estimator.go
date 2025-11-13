// Package tokenizer provides token estimation for various language models.
package tokenizer

import (
	"strings"
	"unicode"
)

// Estimator provides token estimation for text
type Estimator struct {
	model string
}

// NewEstimator creates a new token estimator
func NewEstimator(model string) *Estimator {
	return &Estimator{model: model}
}

// Estimate estimates the number of tokens in the given text
// This is a heuristic estimation based on common tokenization patterns
func (e *Estimator) Estimate(text string) int {
	if len(text) == 0 {
		return 0
	}

	// Different models have different tokenization strategies
	// For now, we use a general approximation
	switch {
	case strings.Contains(strings.ToLower(e.model), "gpt"):
		return e.estimateGPTTokens(text)
	case strings.Contains(strings.ToLower(e.model), "claude"):
		return e.estimateClaudeTokens(text)
	default:
		return e.estimateGenericTokens(text)
	}
}

// estimateGPTTokens estimates tokens for GPT models
// Rule of thumb: ~4 characters per token for English text
func (e *Estimator) estimateGPTTokens(text string) int {
	// Count words and special characters
	words := countWords(text)
	specialChars := countSpecialChars(text)

	// GPT typically: ~0.75 tokens per word + special char overhead
	tokens := int(float64(words)*0.85) + (specialChars / 3)

	// Add overhead for code (more special characters)
	if looksLikeCode(text) {
		tokens = int(float64(tokens) * 1.3)
	}

	// Minimum of 1 token for non-empty text
	if tokens == 0 && len(text) > 0 {
		tokens = 1
	}

	return tokens
}

// estimateClaudeTokens estimates tokens for Claude models
// Similar to GPT but slightly different ratios
func (e *Estimator) estimateClaudeTokens(text string) int {
	words := countWords(text)
	specialChars := countSpecialChars(text)

	// Claude: ~0.75 tokens per word
	tokens := int(float64(words)*0.85) + (specialChars / 3)

	if looksLikeCode(text) {
		tokens = int(float64(tokens) * 1.3)
	}

	// Minimum of 1 token for non-empty text
	if tokens == 0 && len(text) > 0 {
		tokens = 1
	}

	return tokens
}

// estimateGenericTokens provides a generic estimation
func (e *Estimator) estimateGenericTokens(text string) int {
	// Generic: ~4 characters per token
	charCount := len([]rune(text))
	tokens := charCount / 4
	if tokens == 0 && len(text) > 0 {
		tokens = 1
	}
	return tokens
}

// countWords counts the number of words in text
func countWords(text string) int {
	words := 0
	inWord := false

	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			if inWord {
				words++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	if inWord {
		words++
	}

	return words
}

// countSpecialChars counts special characters (punctuation, brackets, etc.)
func countSpecialChars(text string) int {
	count := 0
	for _, r := range text {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			count++
		}
	}
	return count
}

// looksLikeCode detects if text looks like code
func looksLikeCode(text string) bool {
	// Heuristics: high ratio of special chars, common keywords
	codeIndicators := []string{
		"func ", "def ", "class ", "import", "const ",
		"var ", "let ", "return", "if (", "for (", "while (",
		"{", "}", "=>", "->", "//", "/*", "*/", "package ",
	}

	textLower := strings.ToLower(text)
	matches := 0

	for _, indicator := range codeIndicators {
		if strings.Contains(textLower, indicator) {
			matches++
		}
	}

	// If we have 2+ code indicators, likely code
	return matches >= 2
}

// EstimateFromBytes estimates tokens from byte array
func (e *Estimator) EstimateFromBytes(data []byte) int {
	return e.Estimate(string(data))
}

// EstimatePair estimates input and output tokens
func (e *Estimator) EstimatePair(input, output string) (inputTokens, outputTokens int) {
	return e.Estimate(input), e.Estimate(output)
}

// Confidence represents the confidence level of a token estimation
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"   // Within 10% typically
	ConfidenceMedium Confidence = "medium" // Within 20% typically
	ConfidenceLow    Confidence = "low"    // Within 50% typically
)

type Estimation struct {
	Tokens     int
	Confidence Confidence
}

// EstimateWithConfidence provides estimation with confidence level
func (e *Estimator) EstimateWithConfidence(text string) Estimation {
	tokens := e.Estimate(text)

	// Confidence depends on text characteristics
	confidence := ConfidenceMedium

	if looksLikeCode(text) {
		// Code is harder to estimate accurately
		confidence = ConfidenceLow
	} else if len(text) < 30 {
		// Short text is easier
		confidence = ConfidenceHigh
	}

	return Estimation{
		Tokens:     tokens,
		Confidence: confidence,
	}
}
