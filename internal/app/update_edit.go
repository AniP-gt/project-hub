package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// EnterEditModeMsg switches to edit mode for focused item.
type EnterEditModeMsg struct{}

// SaveEditMsg applies title/description changes to focused item.
type SaveEditMsg struct {
	Title       string
	Description string
}

// CancelEditMsg cancels edit mode.
type CancelEditMsg struct{}

// EnterAssignModeMsg switches to assign mode for focused item.
type EnterAssignModeMsg struct{}

// SaveAssignMsg applies assignee changes to focused item.
type SaveAssignMsg struct {
	Assignee string
}

// CancelAssignMsg cancels assign mode.
type CancelAssignMsg struct{}

func (a App) handleEnterEditMode(msg EnterEditModeMsg) (tea.Model, tea.Cmd) {
	a.state.View.Mode = "edit"
	return a, nil
}

func (a App) handleCancelEdit(msg CancelEditMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "edit" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}

func (a App) handleEnterAssignMode(msg EnterAssignModeMsg) (tea.Model, tea.Cmd) {
	a.state.View.Mode = "assign"
	return a, nil
}

func (a App) handleSaveAssign(msg SaveAssignMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	// Call the GitHub client to update the assignees
	cmd := func() tea.Msg {
		assigneeIDs := []string{}
		if msg.Assignee != "" {
			assigneeIDs = append(assigneeIDs, msg.Assignee)
		}
		updatedItem, err := a.github.UpdateAssignees(
			context.Background(),
			a.state.Project.ID,
			a.state.Project.Owner,
			item.ID,
			assigneeIDs,
		)
		if err != nil {
			return NewErrMsg(err)
		}
		return ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	a.state.View.Mode = "normal"
	return a, tea.Batch(cmd, a.refreshBoardCmd())
}

func (a App) handleCancelAssign(msg CancelAssignMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "assign" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}
