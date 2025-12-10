package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/github"
)

// StatusMoveMsg requests moving the focused item left/right.
type StatusMoveMsg struct {
	Direction github.Direction
}

func (a App) handleStatusMove(msg StatusMoveMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	// TODO: map direction to actual adjacent status columns. Placeholder updates status label.
	item := a.state.Items[idx]
	item.Status = string(msg.Direction)
	a.state.Items[idx] = item
	return a, nil
}
