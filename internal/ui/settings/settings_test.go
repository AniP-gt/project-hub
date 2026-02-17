package settings

import (
	"strings"
	"testing"
)

// TestRenderSettingsForm verifies that View() contains required form labels
func TestRenderSettingsForm(t *testing.T) {
	m := New("12345", "test-owner")
	m.SetSize(80, 24)

	output := m.View()

	// Verify form contains "Project ID:" label
	if !strings.Contains(output, "Project ID:") {
		t.Errorf("View() output does not contain 'Project ID:' label\nGot:\n%s", output)
	}

	// Verify form contains "Owner:" label
	if !strings.Contains(output, "Owner:") {
		t.Errorf("View() output does not contain 'Owner:' label\nGot:\n%s", output)
	}
}

// TestRenderValidationHint verifies that View() contains help text with keyboard instructions
func TestRenderValidationHint(t *testing.T) {
	m := New("", "")
	m.SetSize(80, 24)

	output := m.View()

	// Verify help text is present
	helpText := "tab: switch field • enter: save • esc: cancel"
	if !strings.Contains(output, helpText) {
		t.Errorf("View() output does not contain help text '%s'\nGot:\n%s", helpText, output)
	}
}
