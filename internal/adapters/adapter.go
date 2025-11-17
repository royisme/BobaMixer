// Package adapters defines the common interface for executing requests through various providers.
package adapters

import "context"

// EstimateLevel represents the confidence level of token usage estimation.
type EstimateLevel string

// Token estimation confidence levels
const (
	EstimateExact     EstimateLevel = "exact"     // Exact count from API response
	EstimateMapped    EstimateLevel = "mapped"    // Estimated using model mapping
	EstimateHeuristic EstimateLevel = "heuristic" // Estimated using heuristics
)

// Request represents a request to execute through an adapter.
type Request struct {
	Metadata  map[string]string
	SessionID string
	Profile   string
	Tool      string
	Model     string
	Payload   []byte
}

// Usage contains token usage and latency metrics for a request.
type Usage struct {
	Estimate     EstimateLevel
	LatencyMS    int64
	InputTokens  int
	OutputTokens int
}

// Result contains the outcome of executing a request through an adapter.
type Result struct {
	Usage   Usage
	Error   string
	Output  []byte
	Success bool
}

// Adapter is the interface that all provider adapters must implement.
type Adapter interface {
	Name() string
	Execute(ctx context.Context, req Request) (Result, error)
}
