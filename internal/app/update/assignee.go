package update

import (
	tea "github.com/charmbracelet/bubbletea"
)

type AssignMsg struct {
	Assignees []string
}

func Assign(s State, msg AssignMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	item := s.Model.Items[idx]
	item.Assignees = msg.Assignees
	s.Model.Items[idx] = item
	return s, nil
}
