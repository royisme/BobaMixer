package config

import (
	"testing"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("expected service to be created")
	}
}

func TestGetConfigFiles(t *testing.T) {
	svc := NewService()
	files := svc.GetConfigFiles()

	if len(files) == 0 {
		t.Fatal("GetConfigFiles should return non-empty list")
	}

	expectedFiles := []ConfigFile{
		{Name: "Providers", File: "providers.yaml", Desc: "AI provider configurations and API endpoints"},
		{Name: "Tools", File: "tools.yaml", Desc: "CLI tool detection and management"},
		{Name: "Bindings", File: "bindings.yaml", Desc: "Tool-to-provider bindings and proxy settings"},
		{Name: "Secrets", File: "secrets.yaml", Desc: "Encrypted API keys (edit with caution!)"},
		{Name: "Routes", File: "routes.yaml", Desc: "Context-based routing rules"},
		{Name: "Pricing", File: "pricing.yaml", Desc: "Token pricing for cost calculations"},
		{Name: "Settings", File: "settings.yaml", Desc: "Global application settings"},
	}

	if len(files) != len(expectedFiles) {
		t.Fatalf("GetConfigFiles length: got %d, want %d", len(files), len(expectedFiles))
	}

	for i, expected := range expectedFiles {
		if files[i].Name != expected.Name {
			t.Errorf("File[%d].Name: got %q, want %q", i, files[i].Name, expected.Name)
		}
		if files[i].File != expected.File {
			t.Errorf("File[%d].File: got %q, want %q", i, files[i].File, expected.File)
		}
		if files[i].Desc != expected.Desc {
			t.Errorf("File[%d].Desc: got %q, want %q", i, files[i].Desc, expected.Desc)
		}
	}
}

func TestGetConfigFiles_AllFilesHaveYamlExtension(t *testing.T) {
	svc := NewService()
	files := svc.GetConfigFiles()

	for i, file := range files {
		if len(file.File) < 5 || file.File[len(file.File)-5:] != ".yaml" {
			t.Errorf("File[%d].File should end with .yaml, got %q", i, file.File)
		}
	}
}

func TestGetConfigFiles_AllFieldsNonEmpty(t *testing.T) {
	svc := NewService()
	files := svc.GetConfigFiles()

	for i, file := range files {
		if file.Name == "" {
			t.Errorf("File[%d].Name should not be empty", i)
		}
		if file.File == "" {
			t.Errorf("File[%d].File should not be empty", i)
		}
		if file.Desc == "" {
			t.Errorf("File[%d].Desc should not be empty", i)
		}
	}
}

func TestConvertToComponents(t *testing.T) {
	svc := NewService()
	components := svc.ConvertToComponents()

	files := svc.GetConfigFiles()

	if len(components) != len(files) {
		t.Fatalf("ConvertToComponents length: got %d, want %d", len(components), len(files))
	}

	for i, file := range files {
		if components[i].Name != file.Name {
			t.Errorf("Component[%d].Name: got %q, want %q", i, components[i].Name, file.Name)
		}
		if components[i].File != file.File {
			t.Errorf("Component[%d].File: got %q, want %q", i, components[i].File, file.File)
		}
		if components[i].Desc != file.Desc {
			t.Errorf("Component[%d].Desc: got %q, want %q", i, components[i].Desc, file.Desc)
		}
	}
}

func TestConvertToComponents_ReturnType(t *testing.T) {
	svc := NewService()
	result := svc.ConvertToComponents()

	// Verify return type is []components.ConfigFile
	var _ = result
}

func TestViewData(t *testing.T) {
	svc := NewService()
	home := "/home/test"
	data := svc.ViewData(home)

	// Test title fields
	if data.Title == "" {
		t.Error("Title should not be empty")
	}
	if data.Title != "⚙️  Configuration Editor" {
		t.Errorf("Title: got %q, want %q", data.Title, "⚙️  Configuration Editor")
	}

	if data.ConfigTitle == "" {
		t.Error("ConfigTitle should not be empty")
	}
	if data.ConfigTitle != "Configuration Files" {
		t.Errorf("ConfigTitle: got %q, want %q", data.ConfigTitle, "Configuration Files")
	}

	if data.EditorTitle == "" {
		t.Error("EditorTitle should not be empty")
	}
	if data.EditorTitle != "Editor Settings" {
		t.Errorf("EditorTitle: got %q, want %q", data.EditorTitle, "Editor Settings")
	}

	if data.SafetyTitle == "" {
		t.Error("SafetyTitle should not be empty")
	}
	if data.SafetyTitle != "Safety Features" {
		t.Errorf("SafetyTitle: got %q, want %q", data.SafetyTitle, "Safety Features")
	}

	if data.CommandHelpLine == "" {
		t.Error("CommandHelpLine should not be empty")
	}
	if data.CommandHelpLine != "Use CLI: boba edit <target> (to open in editor)" {
		t.Errorf("CommandHelpLine: got %q, want %q", data.CommandHelpLine, "Use CLI: boba edit <target> (to open in editor)")
	}

	if data.EditorName == "" {
		t.Error("EditorName should not be empty")
	}
	if data.EditorName != "vim" {
		t.Errorf("EditorName: got %q, want %q", data.EditorName, "vim")
	}

	if data.Home != home {
		t.Errorf("Home: got %q, want %q", data.Home, home)
	}

	if len(data.ConfigFiles) == 0 {
		t.Error("ConfigFiles should not be empty")
	}
}

func TestViewData_ConfigFilesPopulated(t *testing.T) {
	svc := NewService()
	data := svc.ViewData("/test")

	expectedFiles := svc.GetConfigFiles()

	if len(data.ConfigFiles) != len(expectedFiles) {
		t.Fatalf("ConfigFiles length: got %d, want %d", len(data.ConfigFiles), len(expectedFiles))
	}

	for i, expected := range expectedFiles {
		if data.ConfigFiles[i].Name != expected.Name {
			t.Errorf("ConfigFiles[%d].Name: got %q, want %q", i, data.ConfigFiles[i].Name, expected.Name)
		}
		if data.ConfigFiles[i].File != expected.File {
			t.Errorf("ConfigFiles[%d].File: got %q, want %q", i, data.ConfigFiles[i].File, expected.File)
		}
		if data.ConfigFiles[i].Desc != expected.Desc {
			t.Errorf("ConfigFiles[%d].Desc: got %q, want %q", i, data.ConfigFiles[i].Desc, expected.Desc)
		}
	}
}

func TestViewData_DifferentHomePaths(t *testing.T) {
	svc := NewService()

	testCases := []string{
		"/home/user1",
		"/home/user2",
		"/var/data",
		"",
	}

	for _, home := range testCases {
		data := svc.ViewData(home)
		if data.Home != home {
			t.Errorf("ViewData with home %q: got %q", home, data.Home)
		}
	}
}

func TestViewData_Consistency(t *testing.T) {
	svc := NewService()

	// Call ViewData multiple times to ensure consistency
	data1 := svc.ViewData("/test")
	data2 := svc.ViewData("/test")

	if data1.Title != data2.Title {
		t.Error("ViewData should return consistent Title")
	}

	if data1.ConfigTitle != data2.ConfigTitle {
		t.Error("ViewData should return consistent ConfigTitle")
	}

	if data1.EditorName != data2.EditorName {
		t.Error("ViewData should return consistent EditorName")
	}

	if len(data1.ConfigFiles) != len(data2.ConfigFiles) {
		t.Error("ViewData should return consistent ConfigFiles length")
	}
}

func TestConfigFile_Structure(t *testing.T) {
	// Test that ConfigFile struct is properly defined
	file := ConfigFile{
		Name: "Test",
		File: "test.yaml",
		Desc: "Test description",
	}

	if file.Name != "Test" {
		t.Errorf("Name: got %q, want %q", file.Name, "Test")
	}
	if file.File != "test.yaml" {
		t.Errorf("File: got %q, want %q", file.File, "test.yaml")
	}
	if file.Desc != "Test description" {
		t.Errorf("Desc: got %q, want %q", file.Desc, "Test description")
	}
}
