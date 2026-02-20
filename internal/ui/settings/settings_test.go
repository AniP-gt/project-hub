package settings

import (
	"strings"
	"testing"
)

// TestRenderSettingsForm verifies that View() contains required form labels
func TestRenderSettingsForm(t *testing.T) {
	m := New("12345", "test-owner", false, 100, false, nil)
	m.SetSize(80, 24)

	output := m.View()

	if !strings.Contains(output, "Project ID:") {
		t.Errorf("View() output does not contain 'Project ID:' label\nGot:\n%s", output)
	}

	if !strings.Contains(output, "Owner:") {
		t.Errorf("View() output does not contain 'Owner:' label\nGot:\n%s", output)
	}
}

// TestRenderValidationHint verifies that View() contains help text with keyboard instructions
func TestRenderValidationHint(t *testing.T) {
	m := New("", "", false, 100, false, nil)
	m.SetSize(80, 24)

	output := m.View()

	helpText := "tab: switch field • space: toggle y/n • enter: save • esc: cancel"
	if !strings.Contains(output, helpText) {
		t.Errorf("View() output does not contain help text '%s'\nGot:\n%s", helpText, output)
	}
}

func TestRenderIterationFilterField(t *testing.T) {
	m := New("", "", false, 100, false, []string{"@current", "@next"})
	m.SetSize(80, 24)

	output := m.View()

	if !strings.Contains(output, "Iteration Filter:") {
		t.Errorf("View() output does not contain 'Iteration Filter:' label\nGot:\n%s", output)
	}
}

func TestIterationFilterPrefilled(t *testing.T) {
	m := New("", "", false, 100, false, []string{"@current", "@previous"})
	m.SetSize(80, 24)

	if m.iterationInput.Value() != "@current,@previous" {
		t.Errorf("expected iteration input value '@current,@previous', got '%s'", m.iterationInput.Value())
	}
}
