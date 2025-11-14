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
	name, ok := wh["name"].(string)
	if !ok || name != "Work" {
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
	first, ok := rules[0].(map[string]interface{})
	if !ok {
		t.Fatalf("first rule not map: %#v", rules[0])
	}
	if id, ok := first["id"].(string); !ok || id != "format" {
		t.Fatalf("unexpected first id: %#v", first)
	}
	if use, ok := first["use"].(string); !ok || use != "quick" {
		t.Fatalf("unexpected first use: %#v", first)
	}
	second, ok := rules[1].(map[string]interface{})
	if !ok {
		t.Fatalf("second rule not map: %#v", rules[1])
	}
	if fallback, ok := second["fallback"].(string); !ok || fallback != "quick" {
		t.Fatalf("unexpected second: %#v", second)
	}
}

func TestParseYAMLNestedListStructures(t *testing.T) {
	data := []byte(`list:
  - name: nested
    values:
      - 1
      - 2
  - simple-value
`)
	got, err := parseYAML(data)
	if err != nil {
		t.Fatalf("parseYAML: %v", err)
	}
	items, ok := got["list"].([]interface{})
	if !ok || len(items) != 2 {
		t.Fatalf("unexpected items: %#v", got["list"])
	}
	first, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map entry: %#v", items[0])
	}
	vals, ok := first["values"].([]interface{})
	if !ok {
		t.Fatalf("values not list: %#v", first["values"])
	}
	if len(vals) != 2 {
		t.Fatalf("expected array nested: %#v", vals)
	}
	if val, ok := items[1].(string); !ok || val != "simple-value" {
		t.Fatalf("expected scalar list item")
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
