// Package proxy provides the service layer for proxy view data and logic.
package proxy

import (
	"fmt"
)

// Status values used by the proxy service.
const (
	StatusRunning  = "running"
	StatusStopped  = "stopped"
	StatusChecking = "checking"
)

// ViewData holds all state needed to render the proxy page.
type ViewData struct {
	StatusIcon     string
	StatusText     string
	AdditionalNote string
	ShowConfig     bool
	InfoLines      []string
	ConfigLines    []string
	CommandHelp    string
	Address        string
}

// Service produces proxy-specific view data.
type Service struct {
	address   string
	infoLines []string
}

// NewService creates a proxy service with the provided address.
func NewService(address string) *Service {
	return &Service{
		address: address,
		infoLines: []string{
			"The proxy server intercepts AI API requests from CLI tools",
			"and routes them through BobaMixer for tracking and control.",
		},
	}
}

// ViewData builds all proxy page content based on the proxy status.
func (s *Service) ViewData(status string) ViewData {
	data := ViewData{
		InfoLines: s.infoLines,
		ConfigLines: []string{
			fmt.Sprintf("Tools with proxy enabled automatically use HTTP_PROXY=%s", s.address),
			fmt.Sprintf("and HTTPS_PROXY=%s", s.address),
		},
		CommandHelp: "[S] Refresh Status",
		Address:     s.address,
	}

	switch status {
	case StatusRunning:
		data.StatusIcon = "●"
		data.StatusText = "Running"
		data.ShowConfig = true
	case StatusStopped:
		data.StatusIcon = "○"
		data.StatusText = "Stopped"
		data.AdditionalNote = "Note: Use 'boba proxy serve' in terminal to start the proxy server"
	default:
		data.StatusIcon = "⋯"
		data.StatusText = "Checking..."
	}

	return data
}
