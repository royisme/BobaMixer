package config

import (
	"errors"
	"os"
	"path/filepath"
)

// ProjectConfig represents the optional .boba-project.yaml file.
type ProjectConfig struct {
	Budget  *BudgetSettings `yaml:"budget"`
	Project ProjectInfo     `yaml:"project"`
}

// ProjectInfo describes repository metadata.
type ProjectInfo struct {
	Name              string   `yaml:"name"`
	Type              []string `yaml:"type"`
	PreferredProfiles []string `yaml:"preferred_profiles"`
}

// BudgetSettings controls per-project budgets.
type BudgetSettings struct {
	DailyUSD float64 `yaml:"daily_usd"`
	HardCap  float64 `yaml:"hard_cap"`
}

// FindProjectConfig searches upward from start dir for .boba-project.yaml.
func FindProjectConfig(start string) (*ProjectConfig, string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, "", err
	}
	for {
		path := filepath.Join(dir, ".boba-project.yaml")
		// #nosec G304 -- path is constructed from directory traversal for project config
		data, err := os.ReadFile(path)
		if err == nil {
			cfg, err := parseProjectConfig(data)
			return cfg, path, err
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, "", nil
		}
		dir = parent
	}
}

func parseProjectConfig(data []byte) (*ProjectConfig, error) {
	if len(data) == 0 {
		return &ProjectConfig{}, nil
	}
	root, err := parseYAML(data)
	if err != nil {
		return nil, err
	}
	cfg := &ProjectConfig{}
	if proj := toMap(root["project"]); proj != nil {
		cfg.Project.Name = stringValue(proj["name"])
		cfg.Project.Type = stringSlice(proj["type"])
		cfg.Project.PreferredProfiles = stringSlice(proj["preferred_profiles"])
	}
	if budgetNode := toMap(root["budget"]); budgetNode != nil {
		cfg.Budget = &BudgetSettings{
			DailyUSD: floatValue(budgetNode["daily_usd"]),
			HardCap:  floatValue(budgetNode["hard_cap"]),
		}
	}
	return cfg, nil
}
