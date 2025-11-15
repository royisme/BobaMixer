package pricing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// VendorJSONAdapter loads pricing from manually maintained vendor JSON files
type VendorJSONAdapter struct {
	dataDir string
}

// NewVendorJSONAdapter creates a new vendor JSON adapter
func NewVendorJSONAdapter(dataDir string) *VendorJSONAdapter {
	return &VendorJSONAdapter{
		dataDir: dataDir,
	}
}

// LoadLocal loads pricing from local vendor JSON file
// The file should be located at ~/.boba/pricing.vendor.json
func (a *VendorJSONAdapter) LoadLocal() (*PricingSchema, error) {
	vendorPath := filepath.Join(a.dataDir, "pricing.vendor.json")

	// Check if file exists
	if _, err := os.Stat(vendorPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("vendor JSON not found at %s", vendorPath)
	}

	// #nosec G304 -- path is from safe home directory structure
	data, err := os.ReadFile(vendorPath)
	if err != nil {
		return nil, fmt.Errorf("read vendor JSON: %w", err)
	}

	var schema PricingSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse vendor JSON: %w", err)
	}

	// Validate schema
	if schema.Version != SchemaVersion {
		return nil, fmt.Errorf("unsupported schema version: %d (expected %d)", schema.Version, SchemaVersion)
	}

	// Mark all models as coming from vendor JSON
	for i := range schema.Models {
		schema.Models[i].Source.Kind = "vendor_json"
	}

	return &schema, nil
}

// Save saves pricing schema to vendor JSON file
// This is useful for manual maintenance
func (a *VendorJSONAdapter) Save(schema *PricingSchema) error {
	vendorPath := filepath.Join(a.dataDir, "pricing.vendor.json")

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}

	if err := os.WriteFile(vendorPath, data, 0600); err != nil {
		return fmt.Errorf("write vendor JSON: %w", err)
	}

	return nil
}

// MergeSchemas merges two pricing schemas
// Priority: schema1 > schema2 (schema1 wins on conflicts)
func MergeSchemas(schema1, schema2 *PricingSchema) *PricingSchema {
	if schema1 == nil {
		return schema2
	}
	if schema2 == nil {
		return schema1
	}

	merged := NewPricingSchema()
	merged.Version = schema1.Version
	merged.Currency = schema1.Currency
	merged.FetchedAt = schema1.FetchedAt

	// Build a map of models from schema1
	modelMap := make(map[string]ModelPricing)
	for _, model := range schema1.Models {
		modelMap[model.ID] = model
	}

	// Add models from schema2 if not already present
	for _, model := range schema2.Models {
		if _, exists := modelMap[model.ID]; !exists {
			modelMap[model.ID] = model
		}
	}

	// Convert map back to slice
	for _, model := range modelMap {
		merged.Models = append(merged.Models, model)
	}

	return merged
}
