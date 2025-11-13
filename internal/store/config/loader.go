// Package config manages configuration loading, validation and storage.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Profile struct {
	Temperature float64
	Tags        []string
	CostPer1K   Cost
	Env         map[string]string
	Params      map[string]string
	Key         string
	Name        string
	Adapter     string
	Provider    string
	Endpoint    string
	Model       string
	MaxTokens   int
}

type Cost struct {
	Input  float64
	Output float64
}

type Profiles map[string]Profile

type Secrets map[string]string

type RoutesConfig struct {
	SubAgents map[string]SubAgent
	Rules     []RouteRule
}

type SubAgent struct {
	Triggers   []string
	Conditions map[string]interface{}
	Profile    string
}

type RouteRule struct {
	ID       string
	If       string
	Use      string
	Fallback string
	Explain  string
}

type PricingTable struct {
	Models  map[string]ModelPrice
	Sources []PricingSource
	Refresh PricingRefresh
}

type ModelPrice struct {
	InputPer1K  float64
	OutputPer1K float64
}

type PricingSource struct {
	Type     string
	URL      string
	Path     string
	Priority int
}

type PricingRefresh struct {
	IntervalHours int
	OnStartup     bool
}

func LoadProfiles(home string) (Profiles, error) {
	data, err := readFileIfExists(filepath.Join(home, "profiles.yaml"))
	if err != nil {
		return nil, err
	}
	root, err := parseYAML(data)
	if err != nil {
		return nil, err
	}
	raw, ok := root["profiles"].(map[string]interface{})
	if !ok {
		return nil, errors.New("profiles key missing")
	}
	result := make(Profiles)
	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		node, ok := raw[key].(map[string]interface{})
		if !ok {
			continue
		}
		prof := Profile{Key: key, Env: map[string]string{}, Params: map[string]string{}}
		prof.Name = stringValue(node["name"])
		prof.Adapter = stringValue(node["adapter"])
		prof.Provider = stringValue(node["provider"])
		prof.Endpoint = stringValue(node["endpoint"])
		prof.Model = stringValue(node["model"])
		prof.MaxTokens = intValue(node["max_tokens"])
		prof.Temperature = floatValue(node["temperature"])
		prof.Tags = stringSlice(node["tags"])
		if cost := toMap(node["cost_per_1k"]); cost != nil {
			prof.CostPer1K = Cost{
				Input:  floatValue(cost["input"]),
				Output: floatValue(cost["output"]),
			}
		}
		if env := toMap(node["env"]); env != nil {
			for ek, ev := range env {
				prof.Env[ek] = stringValue(ev)
			}
		}
		if params := toMap(node["params"]); params != nil {
			for pk, pv := range params {
				prof.Params[pk] = stringValue(pv)
			}
		}
		result[key] = prof
	}
	return result, nil
}

func LoadSecrets(home string) (Secrets, error) {
	data, err := readFileIfExists(filepath.Join(home, "secrets.yaml"))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return Secrets{}, nil
	}
	root, err := parseYAML(data)
	if err != nil {
		return nil, err
	}
	raw, ok := root["secrets"].(map[string]interface{})
	if !ok {
		return Secrets{}, nil
	}
	out := make(Secrets, len(raw))
	for k, v := range raw {
		out[k] = stringValue(v)
	}
	return out, nil
}

func LoadRoutes(home string) (*RoutesConfig, error) {
	data, err := readFileIfExists(filepath.Join(home, "routes.yaml"))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return &RoutesConfig{SubAgents: map[string]SubAgent{}, Rules: nil}, nil
	}
	root, err := parseYAML(data)
	if err != nil {
		return nil, err
	}
	cfg := &RoutesConfig{SubAgents: map[string]SubAgent{}, Rules: []RouteRule{}}
	if subs := toMap(root["sub_agents"]); subs != nil {
		for name, raw := range subs {
			entry := toMap(raw)
			sa := SubAgent{}
			sa.Profile = stringValue(entry["profile"])
			sa.Triggers = stringSlice(entry["triggers"])
			if cond := toMap(entry["conditions"]); cond != nil {
				sa.Conditions = cond
			}
			cfg.SubAgents[name] = sa
		}
	}
	if rulesRaw, ok := root["rules"].([]interface{}); ok {
		for _, item := range rulesRaw {
			m := toMap(item)
			rule := RouteRule{
				ID:       stringValue(m["id"]),
				If:       stringValue(m["if"]),
				Use:      stringValue(m["use"]),
				Fallback: stringValue(m["fallback"]),
				Explain:  stringValue(m["explain"]),
			}
			cfg.Rules = append(cfg.Rules, rule)
		}
	}
	return cfg, nil
}

func LoadPricing(home string) (*PricingTable, error) {
	data, err := readFileIfExists(filepath.Join(home, "pricing.yaml"))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return &PricingTable{Models: map[string]ModelPrice{}}, nil
	}
	root, err := parseYAML(data)
	if err != nil {
		return nil, err
	}
	table := &PricingTable{Models: map[string]ModelPrice{}}
	if models := toMap(root["models"]); models != nil {
		for name, raw := range models {
			entry := toMap(raw)
			table.Models[name] = ModelPrice{
				InputPer1K:  floatValue(entry["input_per_1k"]),
				OutputPer1K: floatValue(entry["output_per_1k"]),
			}
		}
	}
	if sources, ok := root["sources"].([]interface{}); ok {
		for _, item := range sources {
			entry := toMap(item)
			table.Sources = append(table.Sources, PricingSource{
				Type:     stringValue(entry["type"]),
				URL:      stringValue(entry["url"]),
				Path:     stringValue(entry["path"]),
				Priority: intValue(entry["priority"]),
			})
		}
	}
	if refresh := toMap(root["refresh"]); refresh != nil {
		table.Refresh = PricingRefresh{
			IntervalHours: intValue(refresh["interval_hours"]),
			OnStartup:     boolValue(refresh["on_startup"]),
		}
	}
	return table, nil
}

func readFileIfExists(path string) ([]byte, error) {
	// #nosec G304 -- path is from safe home directory structure
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []byte{}, nil
		}
		return nil, err
	}
	return data, nil
}

func stringValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func intValue(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case nil:
		return 0
	default:
		return 0
	}
}

func floatValue(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case nil:
		return 0
	default:
		return 0
	}
}

func boolValue(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true"
	default:
		return false
	}
}

func stringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	var result []string
	switch val := v.(type) {
	case []interface{}:
		for _, item := range val {
			result = append(result, stringValue(item))
		}
	case []string:
		result = append(result, val...)
	default:
		result = []string{stringValue(val)}
	}
	return result
}

func toMap(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return nil
}
