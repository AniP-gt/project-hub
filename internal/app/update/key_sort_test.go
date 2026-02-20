package update

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/state"
)

// Test that pressing 'a' while in Sort Mode toggles Assignees sort and does not enter Assign Mode.
func TestHandleKey_SortMode_AssigneesKey(t *testing.T) {
	ti := textinput.New()
	vp := viewport.New(0, 0)

	s := State{
		Model: state.Model{
			Items: []state.Item{{ID: "1", Title: "Item 1"}},
			View: state.ViewContext{
				CurrentView:         state.ViewTable,
				Mode:                state.ModeSort,
				TableSort:           state.TableSort{Field: "", Asc: true},
				FocusedIndex:        0,
				FocusedColumnIndex:  0,
				CardFieldVisibility: state.DefaultCardFieldVisibility(),
			},
			Width:  80,
			Height: 24,
		},
		TextInput:     ti,
		TableViewport: &vp,
	}

	k := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	s2, _ := HandleKey(s, k)

	if s2.Model.View.Mode == "assign" {
		t.Fatalf("expected not to enter assign mode, but mode is %q", s2.Model.View.Mode)
	}
	if s2.Model.View.TableSort.Field != "Assignees" {
		t.Fatalf("expected TableSort.Field to be Assignees, got %q", s2.Model.View.TableSort.Field)
	}
}

// Simulate pressing 's' to enter Sort Mode, then pressing 'a'.
func TestHandleKey_EnterSortThenAssignees(t *testing.T) {
	ti := textinput.New()
	vp := viewport.New(0, 0)

	s := State{
		Model: state.Model{
			Items: []state.Item{{ID: "1", Title: "Item 1"}},
			View: state.ViewContext{
				CurrentView:         state.ViewTable,
				Mode:                state.ModeNormal,
				TableSort:           state.TableSort{Field: "", Asc: true},
				FocusedIndex:        0,
				FocusedColumnIndex:  0,
				CardFieldVisibility: state.DefaultCardFieldVisibility(),
			},
			Width:  80,
			Height: 24,
		},
		TextInput:     ti,
		TableViewport: &vp,
	}

	// Press 's' to enter sort mode
	ks := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	s2, _ := HandleKey(s, ks)
	if s2.Model.View.Mode != state.ModeSort {
		t.Fatalf("expected mode to be ModeSort after pressing s, got %q", s2.Model.View.Mode)
	}

	// Now press 'a'
	ka := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	s3, _ := HandleKey(s2, ka)
	if s3.Model.View.Mode == "assign" {
		t.Fatalf("expected not to enter assign mode, but mode is %q", s3.Model.View.Mode)
	}
	if s3.Model.View.TableSort.Field != "Assignees" {
		t.Fatalf("expected TableSort.Field to be Assignees, got %q", s3.Model.View.TableSort.Field)
	}
}

// Simulate pressing 's' to enter Sort Mode, then pressing 's' to toggle Sub-issues sort
func TestHandleKey_EnterSortThenSubIssues(t *testing.T) {
	ti := textinput.New()
	vp := viewport.New(0, 0)

	s := State{
		Model: state.Model{
			Items: []state.Item{{ID: "1", Title: "Item 1"}},
			View: state.ViewContext{
				CurrentView:         state.ViewTable,
				Mode:                state.ModeNormal,
				TableSort:           state.TableSort{Field: "", Asc: true},
				FocusedIndex:        0,
				FocusedColumnIndex:  0,
				CardFieldVisibility: state.DefaultCardFieldVisibility(),
			},
			Width:  80,
			Height: 24,
		},
		TextInput:     ti,
		TableViewport: &vp,
	}

	// Press 's' to enter sort mode
	ks := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	s2, _ := HandleKey(s, ks)
	if s2.Model.View.Mode != state.ModeSort {
		t.Fatalf("expected mode to be ModeSort after pressing s, got %q", s2.Model.View.Mode)
	}

	ks2 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	s3, _ := HandleKey(s2, ks2)
	if s3.Model.View.TableSort.Field != "" {
		t.Fatalf("expected TableSort.Field to remain empty (no Sub-issue sort), got %q", s3.Model.View.TableSort.Field)
	}
}
