// Package tools provides the service layer for tools view data and logic.
package tools

import (
	"github.com/royisme/bobamixer/internal/domain/core"
	"github.com/royisme/bobamixer/internal/ui/components"
)

// Service encapsulates tool list rendering logic.
type Service struct {
	tools    *core.ToolsConfig
	bindings *core.BindingsConfig
}

// NewService wires the configs for tool rendering.
func NewService(tools *core.ToolsConfig, bindings *core.BindingsConfig) *Service {
	return &Service{
		tools:    tools,
		bindings: bindings,
	}
}

// Rows converts filtered tool indexes into table rows.
func (s *Service) Rows(indexes []int) []components.ToolRow {
	if s.tools == nil || len(indexes) == 0 {
		return nil
	}

	result := make([]components.ToolRow, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(s.tools.Tools) {
			continue
		}

		tool := s.tools.Tools[idx]
		bound := false
		if s.bindings != nil {
			if _, err := s.bindings.FindBinding(tool.ID); err == nil {
				bound = true
			}
		}

		result = append(result, components.ToolRow{
			Name:  tool.Name,
			Exec:  tool.Exec,
			Kind:  string(tool.Kind),
			Bound: bound,
		})
	}

	return result
}

// Details returns the tool details for the selected filtered index.
func (s *Service) Details(indexes []int, selectedIndex int) *components.ToolDetails {
	if s.tools == nil || len(indexes) == 0 || selectedIndex < 0 || selectedIndex >= len(indexes) {
		return nil
	}

	actualIdx := indexes[selectedIndex]
	if actualIdx < 0 || actualIdx >= len(s.tools.Tools) {
		return nil
	}

	tool := s.tools.Tools[actualIdx]
	return &components.ToolDetails{
		ID:          tool.ID,
		ConfigType:  string(tool.ConfigType),
		ConfigPath:  tool.ConfigPath,
		Description: tool.Description,
	}
}

// EmptyStateMessage returns descriptive text for an empty tool list.
func (s *Service) EmptyStateMessage(isEmpty bool, hasSearch bool) string {
	if !isEmpty {
		return ""
	}
	if hasSearch {
		return "No tools match the current filter."
	}
	return "No tools configured."
}
