package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

type AssignMsg struct {
	Assignees []string
}

func (a App) handleAssign(msg AssignMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]
	item.Assignees = msg.Assignees
	a.state.Items[idx] = item
	return a, nil
}
