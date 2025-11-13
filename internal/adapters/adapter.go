// Package adapters defines the common interface for executing requests through various providers.
package adapters

import "context"

type EstimateLevel string

const (
	EstimateExact     EstimateLevel = "exact"
	EstimateMapped    EstimateLevel = "mapped"
	EstimateHeuristic EstimateLevel = "heuristic"
)

type Request struct {
	Metadata  map[string]string
	SessionID string
	Profile   string
	Tool      string
	Model     string
	Payload   []byte
}

type Usage struct {
	Estimate     EstimateLevel
	LatencyMS    int64
	InputTokens  int
	OutputTokens int
}

type Result struct {
	Usage   Usage
	Error   string
	Output  []byte
	Success bool
}

type Adapter interface {
	Name() string
	Execute(ctx context.Context, req Request) (Result, error)
}
