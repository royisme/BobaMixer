package routing

import (
	"testing"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("expected service to be created")
	}
}

func TestViewData(t *testing.T) {
	svc := NewService()
	data := svc.ViewData()

	// Test title fields
	if data.Title == "" {
		t.Error("Title should not be empty")
	}
	if data.Title != "BobaMixer - Routing Rules Tester" {
		t.Errorf("Title: got %q, want %q", data.Title, "BobaMixer - Routing Rules Tester")
	}

	if data.TestTitle == "" {
		t.Error("TestTitle should not be empty")
	}
	if data.TestTitle != "🧪 Test Routing Rules" {
		t.Errorf("TestTitle: got %q, want %q", data.TestTitle, "🧪 Test Routing Rules")
	}

	if data.HowToTitle == "" {
		t.Error("HowToTitle should not be empty")
	}
	if data.HowToTitle != "💡 How to Use" {
		t.Errorf("HowToTitle: got %q, want %q", data.HowToTitle, "💡 How to Use")
	}

	if data.ExampleTitle == "" {
		t.Error("ExampleTitle should not be empty")
	}
	if data.ExampleTitle != "📋 Example" {
		t.Errorf("ExampleTitle: got %q, want %q", data.ExampleTitle, "📋 Example")
	}

	if data.ContextTitle == "" {
		t.Error("ContextTitle should not be empty")
	}
	if data.ContextTitle != "ℹ️  Context Detection" {
		t.Errorf("ContextTitle: got %q, want %q", data.ContextTitle, "ℹ️  Context Detection")
	}

	if data.TestDescription == "" {
		t.Error("TestDescription should not be empty")
	}
	if data.TestDescription != "Test how routing rules would apply to different queries." {
		t.Errorf("TestDescription: got %q, want %q", data.TestDescription, "Test how routing rules would apply to different queries.")
	}

	if data.CommandHelpLine == "" {
		t.Error("CommandHelpLine should not be empty")
	}
	if data.CommandHelpLine != "Use CLI: boba route test <text|@file>" {
		t.Errorf("CommandHelpLine: got %q, want %q", data.CommandHelpLine, "Use CLI: boba route test <text|@file>")
	}
}

func TestViewData_HowToSteps(t *testing.T) {
	svc := NewService()
	data := svc.ViewData()

	if len(data.HowToSteps) == 0 {
		t.Fatal("HowToSteps should not be empty")
	}

	expectedSteps := []string{
		"1. Prepare a test query (text or file)",
		"2. Run: boba route test \"your query text\"",
		"3. Or: boba route test @path/to/file.txt",
	}

	if len(data.HowToSteps) != len(expectedSteps) {
		t.Fatalf("HowToSteps length: got %d, want %d", len(data.HowToSteps), len(expectedSteps))
	}

	for i, expected := range expectedSteps {
		if data.HowToSteps[i] != expected {
			t.Errorf("HowToSteps[%d]: got %q, want %q", i, data.HowToSteps[i], expected)
		}
	}
}

func TestViewData_ExampleLines(t *testing.T) {
	svc := NewService()
	data := svc.ViewData()

	if len(data.ExampleLines) == 0 {
		t.Fatal("ExampleLines should not be empty")
	}

	expectedLines := []string{
		"$ boba route test \"Write a Python function\"",
		"→ Profile: claude-sonnet-3.5",
		"→ Rule: short-query-fast-model",
		"→ Reason: Query < 100 chars",
	}

	if len(data.ExampleLines) != len(expectedLines) {
		t.Fatalf("ExampleLines length: got %d, want %d", len(data.ExampleLines), len(expectedLines))
	}

	for i, expected := range expectedLines {
		if data.ExampleLines[i] != expected {
			t.Errorf("ExampleLines[%d]: got %q, want %q", i, data.ExampleLines[i], expected)
		}
	}
}

func TestViewData_ContextLines(t *testing.T) {
	svc := NewService()
	data := svc.ViewData()

	if len(data.ContextLines) == 0 {
		t.Fatal("ContextLines should not be empty")
	}

	expectedLines := []string{
		"Query length and complexity",
		"Current project and branch",
		"Time of day (day/evening/night)",
		"Project type (go, web, etc.)",
	}

	if len(data.ContextLines) != len(expectedLines) {
		t.Fatalf("ContextLines length: got %d, want %d", len(data.ContextLines), len(expectedLines))
	}

	for i, expected := range expectedLines {
		if data.ContextLines[i] != expected {
			t.Errorf("ContextLines[%d]: got %q, want %q", i, data.ContextLines[i], expected)
		}
	}
}

func TestViewData_Consistency(t *testing.T) {
	svc := NewService()

	// Call ViewData multiple times to ensure consistency
	data1 := svc.ViewData()
	data2 := svc.ViewData()

	if data1.Title != data2.Title {
		t.Error("ViewData should return consistent results")
	}

	if len(data1.HowToSteps) != len(data2.HowToSteps) {
		t.Error("ViewData should return consistent HowToSteps")
	}

	if len(data1.ExampleLines) != len(data2.ExampleLines) {
		t.Error("ViewData should return consistent ExampleLines")
	}

	if len(data1.ContextLines) != len(data2.ContextLines) {
		t.Error("ViewData should return consistent ContextLines")
	}
}
