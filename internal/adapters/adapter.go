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
	Payload   []byte
	Metadata  map[string]string
	SessionID string
	Profile   string
	Tool      string
	Model     string
}

type Usage struct {
	LatencyMS    int64
	InputTokens  int
	OutputTokens int
	Estimate     EstimateLevel
}

type Result struct {
	Output  []byte
	Usage   Usage
	Error   string
	Success bool
}

type Adapter interface {
	Name() string
	Execute(ctx context.Context, req Request) (Result, error)
}
