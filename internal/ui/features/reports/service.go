// Package reports provides the service layer for reports view data and logic.
package reports

import "github.com/royisme/bobamixer/internal/ui/components"

// Option represents a selectable report configuration.
type Option struct {
	Label string
	Desc  string
}

// Service manages read-only report metadata for the reports view.
type Service struct {
	options     []Option
	commandHelp string
}

// NewService returns a service seeded with default report options.
func NewService() *Service {
	return &Service{
		options: []Option{
			{Label: "Last 7 Days Report", Desc: "Generate usage report for the past 7 days"},
			{Label: "Last 30 Days Report", Desc: "Generate monthly usage report"},
			{Label: "Custom Date Range", Desc: "Specify custom start and end dates"},
			{Label: "JSON Format", Desc: "Export report as JSON (default)"},
			{Label: "CSV Format", Desc: "Export report as CSV for spreadsheet tools"},
			{Label: "HTML Format", Desc: "Generate visual HTML report with charts"},
		},
		commandHelp: "Use CLI: boba report --format <json|csv|html> --days <N> --out <file>",
	}
}

// OptionCount returns the number of available report options.
func (s *Service) OptionCount() int {
	return len(s.options)
}

// Options converts the configured options into UI component data.
func (s *Service) Options() []components.ReportOption {
	result := make([]components.ReportOption, len(s.options))
	for i, opt := range s.options {
		result[i] = components.ReportOption{
			Label: opt.Label,
			Desc:  opt.Desc,
		}
	}
	return result
}

// CommandHelp returns the CLI instructions for generating reports.
func (s *Service) CommandHelp() string {
	return s.commandHelp
}
