package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/state"
)

// SwitchViewMsg requests changing the active view.
type SwitchViewMsg struct {
	View state.ViewType
}

// MoveFocusMsg moves focus by delta within current view list.
type MoveFocusMsg struct {
	Delta int
}

func (a App) handleSwitchView(msg SwitchViewMsg) (tea.Model, tea.Cmd) {
	a.state.View.CurrentView = msg.View
	// Keep focus stable; if no items, clear focus.
	if len(a.state.Items) == 0 {
		a.state.View.FocusedItemID = ""
		a.state.View.FocusedIndex = -1
	}
	return a, nil
}

func (a App) handleMoveFocus(msg MoveFocusMsg) (tea.Model, tea.Cmd) {
	if len(a.state.Items) == 0 {
		return a, nil
	}
	idx := a.state.View.FocusedIndex
	if idx < 0 {
		idx = 0
	}
	idx += msg.Delta
	if idx < 0 {
		idx = 0
	}
	if idx >= len(a.state.Items) {
		idx = len(a.state.Items) - 1
	}
	a.state.View.FocusedIndex = idx
	a.state.View.FocusedItemID = a.state.Items[idx].ID
	return a, nil
}
