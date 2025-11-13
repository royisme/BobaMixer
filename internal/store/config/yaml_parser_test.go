package config

import (
	"reflect"
	"testing"
)

func TestParseYAMLMap(t *testing.T) {
	data := []byte("profiles:\n  work-heavy:\n    name: \"Work\"\n    max_tokens: 123\n")
	got, err := parseYAML(data)
	if err != nil {
		t.Fatalf("parseYAML: %v", err)
	}
	profiles, ok := got["profiles"].(map[string]interface{})
	if !ok {
		t.Fatalf("profiles not map: %#v", got["profiles"])
	}
	wh, ok := profiles["work-heavy"].(map[string]interface{})
	if !ok {
		t.Fatalf("work-heavy not map: %#v", profiles["work-heavy"])
	}
	if name := wh["name"].(string); name != "Work" {
		t.Fatalf("name=%s", name)
	}
	if tokens := intFromAny(wh["max_tokens"]); tokens != 123 {
		t.Fatalf("tokens=%d", tokens)
	}
}

func TestParseYAMLListOfMaps(t *testing.T) {
	data := []byte("rules:\n  - id: format\n    use: quick\n  - id: deep\n    fallback: quick\n")
	got, err := parseYAML(data)
	if err != nil {
		t.Fatalf("parseYAML: %v", err)
	}
	rules, ok := got["rules"].([]interface{})
	if !ok {
		t.Fatalf("rules not list: %#v", got["rules"])
	}
	if len(rules) != 2 {
		t.Fatalf("rules len=%d", len(rules))
	}
	first := rules[0].(map[string]interface{})
	if first["id"].(string) != "format" || first["use"].(string) != "quick" {
		t.Fatalf("unexpected first: %#v", first)
	}
	second := rules[1].(map[string]interface{})
	if second["fallback"].(string) != "quick" {
		t.Fatalf("unexpected second: %#v", second)
	}
}

func intFromAny(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		panic(reflect.TypeOf(v))
	}
}
