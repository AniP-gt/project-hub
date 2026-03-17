package settings

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestRenderSettingsForm verifies that View() contains required form labels
func TestRenderSettingsForm(t *testing.T) {
	m := New("12345", "test-owner", false, 100, false, "auto", nil)
	m.SetSize(80, 24)

	output := m.View()

	if !strings.Contains(output, "Project ID:") {
		t.Errorf("View() output does not contain 'Project ID:' label\nGot:\n%s", output)
	}

	if !strings.Contains(output, "Owner:") {
		t.Errorf("View() output does not contain 'Owner:' label\nGot:\n%s", output)
	}

	if !strings.Contains(output, "Create Issue Repo Mode:") {
		t.Errorf("View() output does not contain 'Create Issue Repo Mode:' label\nGot:\n%s", output)
	}
}

// TestRenderValidationHint verifies that View() contains help text with keyboard instructions
func TestRenderValidationHint(t *testing.T) {
	m := New("", "", false, 100, false, "auto", nil)
	m.SetSize(80, 24)

	output := m.View()

	helpText := "tab: switch field • space: toggle y/n • a/r: repo mode • enter: save • esc: cancel"
	if !strings.Contains(output, helpText) {
		t.Errorf("View() output does not contain help text '%s'\nGot:\n%s", helpText, output)
	}
}

func TestRenderIterationFilterField(t *testing.T) {
	m := New("", "", false, 100, false, "auto", []string{"@current", "@next"})
	m.SetSize(80, 24)

	output := m.View()

	if !strings.Contains(output, "Iteration Filter:") {
		t.Errorf("View() output does not contain 'Iteration Filter:' label\nGot:\n%s", output)
	}
}

func TestIterationFilterPrefilled(t *testing.T) {
	m := New("", "", false, 100, false, "required", []string{"@current", "@previous"})
	m.SetSize(80, 24)

	if m.iterationInput.Value() != "@current,@previous" {
		t.Errorf("expected iteration input value '@current,@previous', got '%s'", m.iterationInput.Value())
	}
}

func TestCreateIssueRepoModeTogglesWithSpace(t *testing.T) {
	m := New("", "", false, 100, false, "auto", nil)
	m.focusedField = 5

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if updated.createIssueRepoModeInput.Value() != "r" {
		t.Fatalf("expected repo mode to change to r, got %q", updated.createIssueRepoModeInput.Value())
	}

	_, _, _, _, _, mode, _ := updated.GetValues()
	if mode != "required" {
		t.Fatalf("expected normalized repo mode required, got %q", mode)
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if updated.createIssueRepoModeInput.Value() != "a" {
		t.Fatalf("expected repo mode to change to a, got %q", updated.createIssueRepoModeInput.Value())
	}

	_, _, _, _, _, mode, _ = updated.GetValues()
	if mode != "auto" {
		t.Fatalf("expected normalized repo mode auto, got %q", mode)
	}
}
