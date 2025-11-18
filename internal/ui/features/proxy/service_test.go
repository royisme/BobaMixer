package proxy

import (
	"testing"
)

const (
	testAddress    = "http://localhost:8080"
	testStatusIcon = "⋯"
	testStatusText = "Checking..."
)

func TestNewService(t *testing.T) {
	address := testAddress
	svc := NewService(address)

	if svc == nil {
		t.Fatal("expected service to be created")
	}
	if svc.address != address {
		t.Errorf("address: got %q, want %q", svc.address, address)
	}
	if len(svc.infoLines) == 0 {
		t.Error("infoLines should not be empty")
	}
}

func TestViewData_StatusRunning(t *testing.T) {
	address := testAddress
	svc := NewService(address)

	data := svc.ViewData(StatusRunning)

	if data.StatusIcon != "●" {
		t.Errorf("StatusIcon: got %q, want %q", data.StatusIcon, "●")
	}
	if data.StatusText != "Running" {
		t.Errorf("StatusText: got %q, want %q", data.StatusText, "Running")
	}
	if !data.ShowConfig {
		t.Error("ShowConfig should be true when running")
	}
	if data.AdditionalNote != "" {
		t.Error("AdditionalNote should be empty when running")
	}
	if data.Address != address {
		t.Errorf("Address: got %q, want %q", data.Address, address)
	}
	if data.CommandHelp != "[S] Refresh Status" {
		t.Errorf("CommandHelp: got %q, want %q", data.CommandHelp, "[S] Refresh Status")
	}
}

func TestViewData_StatusStopped(t *testing.T) {
	address := testAddress
	svc := NewService(address)

	data := svc.ViewData(StatusStopped)

	if data.StatusIcon != "○" {
		t.Errorf("StatusIcon: got %q, want %q", data.StatusIcon, "○")
	}
	if data.StatusText != "Stopped" {
		t.Errorf("StatusText: got %q, want %q", data.StatusText, "Stopped")
	}
	if data.ShowConfig {
		t.Error("ShowConfig should be false when stopped")
	}
	if data.AdditionalNote == "" {
		t.Error("AdditionalNote should not be empty when stopped")
	}
	if data.Address != address {
		t.Errorf("Address: got %q, want %q", data.Address, address)
	}
}

func TestViewData_StatusChecking(t *testing.T) {
	address := testAddress
	svc := NewService(address)

	data := svc.ViewData(StatusChecking)

	if data.StatusIcon != testStatusIcon {
		t.Errorf("StatusIcon: got %q, want %q", data.StatusIcon, testStatusIcon)
	}
	if data.StatusText != testStatusText {
		t.Errorf("StatusText: got %q, want %q", data.StatusText, testStatusText)
	}
	if data.ShowConfig {
		t.Error("ShowConfig should be false when checking")
	}
}

func TestViewData_UnknownStatus(t *testing.T) {
	svc := NewService(testAddress)

	data := svc.ViewData("unknown")

	if data.StatusIcon != testStatusIcon {
		t.Errorf("StatusIcon: got %q, want %q", data.StatusIcon, testStatusIcon)
	}
	if data.StatusText != testStatusText {
		t.Errorf("StatusText: got %q, want %q", data.StatusText, testStatusText)
	}
}

func TestViewData_InfoLines(t *testing.T) {
	svc := NewService("http://localhost:8080")

	data := svc.ViewData(StatusRunning)

	if len(data.InfoLines) == 0 {
		t.Error("InfoLines should not be empty")
	}

	if len(svc.infoLines) != len(data.InfoLines) {
		t.Errorf("InfoLines length: got %d, want %d", len(data.InfoLines), len(svc.infoLines))
	}

	for i, line := range svc.infoLines {
		if data.InfoLines[i] != line {
			t.Errorf("InfoLines[%d]: got %q, want %q", i, data.InfoLines[i], line)
		}
	}
}

func TestViewData_ConfigLines(t *testing.T) {
	address := "http://localhost:8080"
	svc := NewService(address)

	data := svc.ViewData(StatusRunning)

	if len(data.ConfigLines) == 0 {
		t.Error("ConfigLines should not be empty")
	}

	// Check that config lines contain the address
	foundAddress := false
	for _, line := range data.ConfigLines {
		if len(line) > 0 {
			foundAddress = true
			break
		}
	}
	if !foundAddress {
		t.Error("ConfigLines should contain address information")
	}
}

func TestViewData_DifferentAddresses(t *testing.T) {
	testCases := []string{
		"http://localhost:8080",
		"http://127.0.0.1:9090",
		"http://proxy.local:3128",
	}

	for _, addr := range testCases {
		svc := NewService(addr)
		data := svc.ViewData(StatusRunning)

		if data.Address != addr {
			t.Errorf("Address: got %q, want %q", data.Address, addr)
		}
	}
}

func TestViewData_Consistency(t *testing.T) {
	svc := NewService("http://localhost:8080")

	// Call ViewData multiple times with same status
	data1 := svc.ViewData(StatusRunning)
	data2 := svc.ViewData(StatusRunning)

	if data1.StatusIcon != data2.StatusIcon {
		t.Error("ViewData should return consistent StatusIcon")
	}
	if data1.StatusText != data2.StatusText {
		t.Error("ViewData should return consistent StatusText")
	}
	if data1.ShowConfig != data2.ShowConfig {
		t.Error("ViewData should return consistent ShowConfig")
	}
}

func TestConstants(t *testing.T) {
	if StatusRunning != "running" {
		t.Errorf("StatusRunning: got %q, want %q", StatusRunning, "running")
	}
	if StatusStopped != "stopped" {
		t.Errorf("StatusStopped: got %q, want %q", StatusStopped, "stopped")
	}
	if StatusChecking != "checking" {
		t.Errorf("StatusChecking: got %q, want %q", StatusChecking, "checking")
	}
}

func TestViewData_AllStatuses(t *testing.T) {
	svc := NewService("http://localhost:8080")

	statuses := []struct {
		status     string
		wantIcon   string
		wantText   string
		wantConfig bool
	}{
		{StatusRunning, "●", "Running", true},
		{StatusStopped, "○", "Stopped", false},
		{StatusChecking, "⋯", "Checking...", false},
		{"invalid", "⋯", "Checking...", false},
	}

	for _, tc := range statuses {
		data := svc.ViewData(tc.status)

		if data.StatusIcon != tc.wantIcon {
			t.Errorf("Status %q - StatusIcon: got %q, want %q", tc.status, data.StatusIcon, tc.wantIcon)
		}
		if data.StatusText != tc.wantText {
			t.Errorf("Status %q - StatusText: got %q, want %q", tc.status, data.StatusText, tc.wantText)
		}
		if data.ShowConfig != tc.wantConfig {
			t.Errorf("Status %q - ShowConfig: got %v, want %v", tc.status, data.ShowConfig, tc.wantConfig)
		}
	}
}
