package tooladapter

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/royisme/bobamixer/internal/adapters"
)

type Runner struct {
	name string
	bin  string
	env  []string
}

func New(name, bin string, env []string) *Runner {
	return &Runner{name: name, bin: bin, env: env}
}

func (r *Runner) Name() string { return r.name }

func (r *Runner) Execute(ctx context.Context, req adapters.Request) (adapters.Result, error) {
	cmd := exec.CommandContext(ctx, r.bin)
	cmd.Env = append(cmd.Env, r.env...)
	cmd.Stdin = strings.NewReader(string(req.Payload))
	start := time.Now()
	out, err := cmd.CombinedOutput()
	usage := adapters.Usage{Estimate: adapters.EstimateHeuristic, LatencyMS: time.Since(start).Milliseconds()}
	if err != nil {
		return adapters.Result{Success: false, Output: out, Error: err.Error(), Usage: usage}, nil
	}
	return adapters.Result{Success: true, Output: out, Usage: usage}, nil
}
