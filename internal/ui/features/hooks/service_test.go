package hooks

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
	if data.Title != "🪝 Git Hooks Management" {
		t.Errorf("Title: got %q, want %q", data.Title, "🪝 Git Hooks Management")
	}

	if data.RepoTitle == "" {
		t.Error("RepoTitle should not be empty")
	}
	if data.RepoTitle != "Current Repository" {
		t.Errorf("RepoTitle: got %q, want %q", data.RepoTitle, "Current Repository")
	}

	if data.HooksTitle == "" {
		t.Error("HooksTitle should not be empty")
	}
	if data.HooksTitle != "Available Hooks" {
		t.Errorf("HooksTitle: got %q, want %q", data.HooksTitle, "Available Hooks")
	}

	if data.BenefitsTitle == "" {
		t.Error("BenefitsTitle should not be empty")
	}
	if data.BenefitsTitle != "Benefits" {
		t.Errorf("BenefitsTitle: got %q, want %q", data.BenefitsTitle, "Benefits")
	}

	if data.ActivityTitle == "" {
		t.Error("ActivityTitle should not be empty")
	}
	if data.ActivityTitle != "Recent Hook Activity" {
		t.Errorf("ActivityTitle: got %q, want %q", data.ActivityTitle, "Recent Hook Activity")
	}

	if data.CommandHelpLine == "" {
		t.Error("CommandHelpLine should not be empty")
	}
	expectedCmd := "Use CLI: boba hooks install (install) | boba hooks remove (uninstall)"
	if data.CommandHelpLine != expectedCmd {
		t.Errorf("CommandHelpLine: got %q, want %q", data.CommandHelpLine, expectedCmd)
	}
}

func TestGetAvailableHooks_Installed(t *testing.T) {
	svc := NewService()
	hooks := svc.GetAvailableHooks(true)

	if len(hooks) == 0 {
		t.Fatal("GetAvailableHooks should return non-empty list")
	}

	expectedHooks := []HookInfo{
		{
			Name:   "post-checkout",
			Desc:   "Track branch switches and suggest optimal profiles",
			Active: true,
		},
		{
			Name:   "post-commit",
			Desc:   "Record commit events for usage tracking",
			Active: true,
		},
		{
			Name:   "post-merge",
			Desc:   "Track merge events and repository changes",
			Active: true,
		},
	}

	if len(hooks) != len(expectedHooks) {
		t.Fatalf("GetAvailableHooks length: got %d, want %d", len(hooks), len(expectedHooks))
	}

	for i, expected := range expectedHooks {
		if hooks[i].Name != expected.Name {
			t.Errorf("Hook[%d].Name: got %q, want %q", i, hooks[i].Name, expected.Name)
		}
		if hooks[i].Desc != expected.Desc {
			t.Errorf("Hook[%d].Desc: got %q, want %q", i, hooks[i].Desc, expected.Desc)
		}
		if hooks[i].Active != expected.Active {
			t.Errorf("Hook[%d].Active: got %v, want %v", i, hooks[i].Active, expected.Active)
		}
	}
}

func TestGetAvailableHooks_NotInstalled(t *testing.T) {
	svc := NewService()
	hooks := svc.GetAvailableHooks(false)

	if len(hooks) == 0 {
		t.Fatal("GetAvailableHooks should return non-empty list")
	}

	// All hooks should have Active = false
	for i, hook := range hooks {
		if hook.Active != false {
			t.Errorf("Hook[%d].Active: got %v, want false (not installed)", i, hook.Active)
		}
	}
}

func TestGetAvailableHooks_AllFieldsNonEmpty(t *testing.T) {
	svc := NewService()
	hooks := svc.GetAvailableHooks(true)

	for i, hook := range hooks {
		if hook.Name == "" {
			t.Errorf("Hook[%d].Name should not be empty", i)
		}
		if hook.Desc == "" {
			t.Errorf("Hook[%d].Desc should not be empty", i)
		}
	}
}

func TestGetAvailableHooks_StandardGitHooks(t *testing.T) {
	svc := NewService()
	hooks := svc.GetAvailableHooks(true)

	expectedNames := []string{
		"post-checkout",
		"post-commit",
		"post-merge",
	}

	if len(hooks) != len(expectedNames) {
		t.Fatalf("Expected %d hooks, got %d", len(expectedNames), len(hooks))
	}

	for i, expectedName := range expectedNames {
		if hooks[i].Name != expectedName {
			t.Errorf("Hook[%d].Name: got %q, want %q", i, hooks[i].Name, expectedName)
		}
	}
}

func TestConvertToComponents(t *testing.T) {
	svc := NewService()
	hooks := svc.GetAvailableHooks(true)
	components := svc.ConvertToComponents(hooks)

	if len(components) != len(hooks) {
		t.Fatalf("ConvertToComponents length: got %d, want %d", len(components), len(hooks))
	}

	for i, hook := range hooks {
		if components[i].Name != hook.Name {
			t.Errorf("Component[%d].Name: got %q, want %q", i, components[i].Name, hook.Name)
		}
		if components[i].Desc != hook.Desc {
			t.Errorf("Component[%d].Desc: got %q, want %q", i, components[i].Desc, hook.Desc)
		}
		if components[i].Active != hook.Active {
			t.Errorf("Component[%d].Active: got %v, want %v", i, components[i].Active, hook.Active)
		}
	}
}

func TestConvertToComponents_EmptyList(t *testing.T) {
	svc := NewService()
	components := svc.ConvertToComponents([]HookInfo{})

	if len(components) != 0 {
		t.Errorf("ConvertToComponents with empty input: got %d, want 0", len(components))
	}
}

func TestConvertToComponents_ReturnType(t *testing.T) {
	svc := NewService()
	hooks := svc.GetAvailableHooks(true)
	result := svc.ConvertToComponents(hooks)

	// Verify return type is []components.HookInfo
	var _ = result
}

func TestViewData_Consistency(t *testing.T) {
	svc := NewService()

	// Call ViewData multiple times to ensure consistency
	data1 := svc.ViewData()
	data2 := svc.ViewData()

	if data1.Title != data2.Title {
		t.Error("ViewData should return consistent Title")
	}

	if data1.RepoTitle != data2.RepoTitle {
		t.Error("ViewData should return consistent RepoTitle")
	}

	if data1.HooksTitle != data2.HooksTitle {
		t.Error("ViewData should return consistent HooksTitle")
	}

	if data1.CommandHelpLine != data2.CommandHelpLine {
		t.Error("ViewData should return consistent CommandHelpLine")
	}
}

func TestGetAvailableHooks_Consistency(t *testing.T) {
	svc := NewService()

	// Call with same parameter multiple times
	hooks1 := svc.GetAvailableHooks(true)
	hooks2 := svc.GetAvailableHooks(true)

	if len(hooks1) != len(hooks2) {
		t.Error("GetAvailableHooks should return consistent results")
	}

	for i := range hooks1 {
		if hooks1[i].Name != hooks2[i].Name {
			t.Errorf("Inconsistent Name at index %d", i)
		}
		if hooks1[i].Desc != hooks2[i].Desc {
			t.Errorf("Inconsistent Desc at index %d", i)
		}
		if hooks1[i].Active != hooks2[i].Active {
			t.Errorf("Inconsistent Active at index %d", i)
		}
	}
}

func TestHookInfo_Structure(t *testing.T) {
	// Test that HookInfo struct is properly defined
	hook := HookInfo{
		Name:   "test-hook",
		Desc:   "Test description",
		Active: true,
	}

	if hook.Name != "test-hook" {
		t.Errorf("Name: got %q, want %q", hook.Name, "test-hook")
	}
	if hook.Desc != "Test description" {
		t.Errorf("Desc: got %q, want %q", hook.Desc, "Test description")
	}
	if hook.Active != true {
		t.Errorf("Active: got %v, want true", hook.Active)
	}
}
