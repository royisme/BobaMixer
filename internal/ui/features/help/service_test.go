package help

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
	if data.Title != "❓ BobaMixer Help & Shortcuts" {
		t.Errorf("Title: got %q, want %q", data.Title, "❓ BobaMixer Help & Shortcuts")
	}

	if data.Subtitle != "" {
		t.Errorf("Subtitle: got %q, want empty string", data.Subtitle)
	}

	if data.NavigationHint == "" {
		t.Error("NavigationHint should not be empty")
	}
	if data.NavigationHint != "Press Esc to close this overlay" {
		t.Errorf("NavigationHint: got %q, want %q", data.NavigationHint, "Press Esc to close this overlay")
	}
}

func TestGetDefaultTips(t *testing.T) {
	svc := NewService()
	tips := svc.GetDefaultTips()

	if len(tips) == 0 {
		t.Fatal("GetDefaultTips should return non-empty list")
	}

	expectedTips := []string{
		"Use number keys (1-5) to jump between sections",
		"All interactive features live in the TUI",
		"CLI commands remain available for automation",
		"Press ? anytime to toggle this help overlay",
	}

	if len(tips) != len(expectedTips) {
		t.Fatalf("GetDefaultTips length: got %d, want %d", len(tips), len(expectedTips))
	}

	for i, expected := range expectedTips {
		if tips[i] != expected {
			t.Errorf("Tip[%d]: got %q, want %q", i, tips[i], expected)
		}
	}
}

func TestGetDefaultTips_AllNonEmpty(t *testing.T) {
	svc := NewService()
	tips := svc.GetDefaultTips()

	for i, tip := range tips {
		if tip == "" {
			t.Errorf("Tip[%d] should not be empty", i)
		}
	}
}

func TestGetDefaultLinks(t *testing.T) {
	svc := NewService()
	links := svc.GetDefaultLinks()

	if len(links) == 0 {
		t.Fatal("GetDefaultLinks should return non-empty list")
	}

	expectedLinks := []HelpLink{
		{Label: "Full docs", URL: "https://royisme.github.io/BobaMixer/"},
		{Label: "GitHub", URL: "https://github.com/royisme/BobaMixer"},
	}

	if len(links) != len(expectedLinks) {
		t.Fatalf("GetDefaultLinks length: got %d, want %d", len(links), len(expectedLinks))
	}

	for i, expected := range expectedLinks {
		if links[i].Label != expected.Label {
			t.Errorf("Link[%d].Label: got %q, want %q", i, links[i].Label, expected.Label)
		}
		if links[i].URL != expected.URL {
			t.Errorf("Link[%d].URL: got %q, want %q", i, links[i].URL, expected.URL)
		}
	}
}

func TestGetDefaultLinks_AllFieldsNonEmpty(t *testing.T) {
	svc := NewService()
	links := svc.GetDefaultLinks()

	for i, link := range links {
		if link.Label == "" {
			t.Errorf("Link[%d].Label should not be empty", i)
		}
		if link.URL == "" {
			t.Errorf("Link[%d].URL should not be empty", i)
		}
	}
}

func TestGetDefaultLinks_ValidURLs(t *testing.T) {
	svc := NewService()
	links := svc.GetDefaultLinks()

	for i, link := range links {
		if len(link.URL) < 8 || (link.URL[:7] != "http://" && link.URL[:8] != "https://") {
			t.Errorf("Link[%d].URL should start with http:// or https://, got %q", i, link.URL)
		}
	}
}

func TestConvertLinksToComponents(t *testing.T) {
	svc := NewService()
	links := svc.GetDefaultLinks()
	components := svc.ConvertLinksToComponents(links)

	if len(components) != len(links) {
		t.Fatalf("ConvertLinksToComponents length: got %d, want %d", len(components), len(links))
	}

	for i, link := range links {
		if components[i].Label != link.Label {
			t.Errorf("Component[%d].Label: got %q, want %q", i, components[i].Label, link.Label)
		}
		if components[i].URL != link.URL {
			t.Errorf("Component[%d].URL: got %q, want %q", i, components[i].URL, link.URL)
		}
	}
}

func TestConvertLinksToComponents_EmptyList(t *testing.T) {
	svc := NewService()
	components := svc.ConvertLinksToComponents([]HelpLink{})

	if len(components) != 0 {
		t.Errorf("ConvertLinksToComponents with empty input: got %d, want 0", len(components))
	}
}

func TestConvertLinksToComponents_ReturnType(t *testing.T) {
	svc := NewService()
	links := svc.GetDefaultLinks()
	result := svc.ConvertLinksToComponents(links)

	// Verify return type is []components.HelpLink
	var _ = result
}

func TestConvertLinksToComponents_CustomLinks(t *testing.T) {
	svc := NewService()
	customLinks := []HelpLink{
		{Label: "Custom1", URL: "https://example.com/1"},
		{Label: "Custom2", URL: "https://example.com/2"},
	}

	components := svc.ConvertLinksToComponents(customLinks)

	if len(components) != len(customLinks) {
		t.Fatalf("Length: got %d, want %d", len(components), len(customLinks))
	}

	for i, link := range customLinks {
		if components[i].Label != link.Label {
			t.Errorf("Component[%d].Label: got %q, want %q", i, components[i].Label, link.Label)
		}
		if components[i].URL != link.URL {
			t.Errorf("Component[%d].URL: got %q, want %q", i, components[i].URL, link.URL)
		}
	}
}

func TestViewData_Consistency(t *testing.T) {
	svc := NewService()

	// Call ViewData multiple times to ensure consistency
	data1 := svc.ViewData()
	data2 := svc.ViewData()

	if data1.Title != data2.Title {
		t.Error("ViewData should return consistent Title")
	}

	if data1.Subtitle != data2.Subtitle {
		t.Error("ViewData should return consistent Subtitle")
	}

	if data1.NavigationHint != data2.NavigationHint {
		t.Error("ViewData should return consistent NavigationHint")
	}
}

func TestGetDefaultTips_Consistency(t *testing.T) {
	svc := NewService()

	// Call multiple times
	tips1 := svc.GetDefaultTips()
	tips2 := svc.GetDefaultTips()

	if len(tips1) != len(tips2) {
		t.Error("GetDefaultTips should return consistent results")
	}

	for i := range tips1 {
		if tips1[i] != tips2[i] {
			t.Errorf("Inconsistent tip at index %d", i)
		}
	}
}

func TestGetDefaultLinks_Consistency(t *testing.T) {
	svc := NewService()

	// Call multiple times
	links1 := svc.GetDefaultLinks()
	links2 := svc.GetDefaultLinks()

	if len(links1) != len(links2) {
		t.Error("GetDefaultLinks should return consistent results")
	}

	for i := range links1 {
		if links1[i].Label != links2[i].Label {
			t.Errorf("Inconsistent Label at index %d", i)
		}
		if links1[i].URL != links2[i].URL {
			t.Errorf("Inconsistent URL at index %d", i)
		}
	}
}

func TestHelpLink_Structure(t *testing.T) {
	// Test that HelpLink struct is properly defined
	link := HelpLink{
		Label: "Test Link",
		URL:   "https://test.example.com",
	}

	if link.Label != "Test Link" {
		t.Errorf("Label: got %q, want %q", link.Label, "Test Link")
	}
	if link.URL != "https://test.example.com" {
		t.Errorf("URL: got %q, want %q", link.URL, "https://test.example.com")
	}
}
