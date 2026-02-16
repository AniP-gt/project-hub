package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"project-hub/internal/state"
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
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	a.textInput.Width = a.state.Width - 10
	if a.textInput.Width < 30 {
		a.textInput.Width = 30
	}
	a.textInput.SetValue(item.Title)
	a.textInput.Placeholder = ""
	a.state.View.Mode = "edit"
	return a, a.textInput.Focus()
}

func (a App) handleCancelEdit(msg CancelEditMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "edit" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}

func (a App) handleEnterAssignMode(msg EnterAssignModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil // Or handle error
	}
	item := a.state.Items[idx]

	// Initialize input with first assignee if present, otherwise empty string
	assignee := ""
	if len(item.Assignees) > 0 {
		assignee = item.Assignees[0]
	}

	a.textInput.SetValue(assignee)
	a.textInput.Placeholder = "Enter assignee..."
	a.state.View.Mode = "assign"
	return a, a.textInput.Focus()
}

func (a App) handleSaveAssign(msg SaveAssignMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	// Find the "Assignees" field from the project's fields
	var assigneeField state.Field
	found := false
	for _, field := range a.state.Project.Fields {
		if field.Name == "Assignees" {
			assigneeField = field
			found = true
			break
		}
	}
	if !found {
		return a, func() tea.Msg {
			return NewErrMsg(fmt.Errorf("assignees field not found in project"))
		}
	}

	// Call the GitHub client to update the assignees
	cmd := func() tea.Msg {
		userLogins := []string{}
		if msg.Assignee != "" {
			userLogins = append(userLogins, msg.Assignee)
		}
		updatedItem, err := a.github.UpdateAssignees(
			context.Background(),
			a.state.Project.ID,
			a.state.Project.Owner,
			item.ID, // Use the item's node ID for field updates
			assigneeField.ID,
			userLogins,
		)
		if err != nil {
			return NewErrMsg(err)
		}
		return ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	a.state.View.Mode = "normal"
	return a, tea.Batch(cmd)
}

func (a App) handleCancelAssign(msg CancelAssignMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "assign" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}
