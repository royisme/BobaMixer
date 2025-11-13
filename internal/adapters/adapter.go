package adapters

import "context"

type EstimateLevel string

const (
	EstimateExact     EstimateLevel = "exact"
	EstimateMapped    EstimateLevel = "mapped"
	EstimateHeuristic EstimateLevel = "heuristic"
)

type Request struct {
	SessionID string
	Profile   string
	Tool      string
	Model     string
	Payload   []byte
	Metadata  map[string]string
}

type Usage struct {
	InputTokens  int
	OutputTokens int
	LatencyMS    int64
	Estimate     EstimateLevel
}

type Result struct {
	Success bool
	Output  []byte
	Error   string
	Usage   Usage
}

type Adapter interface {
	Name() string
	Execute(ctx context.Context, req Request) (Result, error)
}
