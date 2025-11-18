// Package routing provides the service layer for routing view data and logic.
package routing

// Service manages routing view data and logic.
type Service struct{}

// NewService creates a new routing service.
func NewService() *Service {
	return &Service{}
}

// ViewData returns all static data for the routing view.
func (s *Service) ViewData() ViewData {
	return ViewData{
		Title:           "BobaMixer - Routing Rules Tester",
		TestTitle:       "🧪 Test Routing Rules",
		HowToTitle:      "💡 How to Use",
		ExampleTitle:    "📋 Example",
		ContextTitle:    "ℹ️  Context Detection",
		TestDescription: "Test how routing rules would apply to different queries.",
		HowToSteps: []string{
			"1. Prepare a test query (text or file)",
			"2. Run: boba route test \"your query text\"",
			"3. Or: boba route test @path/to/file.txt",
		},
		ExampleLines: []string{
			"$ boba route test \"Write a Python function\"",
			"→ Profile: claude-sonnet-3.5",
			"→ Rule: short-query-fast-model",
			"→ Reason: Query < 100 chars",
		},
		ContextLines: []string{
			"Query length and complexity",
			"Current project and branch",
			"Time of day (day/evening/night)",
			"Project type (go, web, etc.)",
		},
		CommandHelpLine: "Use CLI: boba route test <text|@file>",
	}
}

// ViewData holds all data needed to render the routing view.
type ViewData struct {
	Title           string
	TestTitle       string
	HowToTitle      string
	ExampleTitle    string
	ContextTitle    string
	TestDescription string
	HowToSteps      []string
	ExampleLines    []string
	ContextLines    []string
	CommandHelpLine string
}
