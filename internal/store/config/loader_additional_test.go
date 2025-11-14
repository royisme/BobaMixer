package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPricing(t *testing.T) {
	dir := t.TempDir()
	data := `models:
  mixtral:
    input_per_1k: 0.3
    output_per_1k: 0.6
sources:
  - type: http
    url: https://example/pricing.json
    priority: 1
refresh:
  interval_hours: 4
  on_startup: true
`
	if err := os.WriteFile(filepath.Join(dir, "pricing.yaml"), []byte(data), 0o600); err != nil {
		t.Fatalf("write pricing: %v", err)
	}
	table, err := LoadPricing(dir)
	if err != nil {
		t.Fatalf("LoadPricing: %v", err)
	}
	price, ok := table.Models["mixtral"]
	if !ok || price.InputPer1K != 0.3 || price.OutputPer1K != 0.6 {
		t.Fatalf("unexpected model price: %#v", price)
	}
	if len(table.Sources) != 1 || table.Sources[0].Type != "http" {
		t.Fatalf("expected source parsed: %#v", table.Sources)
	}
	if table.Refresh.IntervalHours != 4 || !table.Refresh.OnStartup {
		t.Fatalf("refresh not parsed: %#v", table.Refresh)
	}

	empty, err := LoadPricing(filepath.Join(dir, "empty"))
	if err != nil {
		t.Fatalf("LoadPricing empty: %v", err)
	}
	if len(empty.Models) != 0 {
		t.Fatalf("expected empty models: %#v", empty.Models)
	}
}

func TestReadFileIfExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.yaml")
	data, err := readFileIfExists(path)
	if err != nil {
		t.Fatalf("read missing: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty data for missing file")
	}
	if err := os.WriteFile(path, []byte("value"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	data, err = readFileIfExists(path)
	if err != nil || string(data) != "value" {
		t.Fatalf("read existing: %q %v", data, err)
	}
}

type stringerType struct{}

func (stringerType) String() string { return "stringer" }

func TestHelperConversions(t *testing.T) {
	if got := stringValue(stringerType{}); got != "stringer" {
		t.Fatalf("stringValue stringer = %s", got)
	}
	if got := intValue(float64(2)); got != 2 {
		t.Fatalf("intValue float64 = %d", got)
	}
	if got := intValue(nil); got != 0 {
		t.Fatalf("intValue nil = %d", got)
	}
	if got := floatValue(3); got != 3 {
		t.Fatalf("floatValue int = %f", got)
	}
	if got := floatValue(nil); got != 0 {
		t.Fatalf("floatValue nil = %f", got)
	}
	if !boolValue("true") || boolValue("false") {
		t.Fatalf("boolValue string parsing failed")
	}
	if boolValue(123) {
		t.Fatalf("boolValue default should be false")
	}
	slice := stringSlice([]interface{}{1, "two"})
	if len(slice) != 2 || slice[0] != "1" || slice[1] != "two" {
		t.Fatalf("stringSlice mixed = %#v", slice)
	}
	if single := stringSlice("solo"); len(single) != 1 || single[0] != "solo" {
		t.Fatalf("stringSlice scalar = %#v", single)
	}
	if asStrings := stringSlice([]string{"one", "two"}); len(asStrings) != 2 {
		t.Fatalf("stringSlice []string = %#v", asStrings)
	}
	if m := toMap(map[string]interface{}{"k": "v"}); m["k"].(string) != "v" {
		t.Fatalf("toMap returned %#v", m)
	}
	if toMap(nil) != nil {
		t.Fatalf("toMap nil should return nil")
	}
}
