package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/github"
	"project-hub/internal/state"
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
	item := a.state.Items[idx]

	// Find the "Status" field from the project's fields
	var statusField state.Field
	found := false
	for _, field := range a.state.Project.Fields {
		if field.Name == "Status" {
			statusField = field
			found = true
			break
		}
	}
	if !found {
		return a, func() tea.Msg {
			return NewErrMsg(fmt.Errorf("status field not found in project"))
		}
	}

	// Find the current status option index
	currentStatusIndex := -1
	for i, option := range statusField.Options {
		if option.Name == item.Status {
			currentStatusIndex = i
			break
		}
	}

	if currentStatusIndex == -1 {
		return a, func() tea.Msg {
			return NewErrMsg(fmt.Errorf("current status '%s' not found in status field options", item.Status))
		}
	}

	// Calculate the new status index
	newStatusIndex := currentStatusIndex
	if msg.Direction == github.DirectionLeft {
		newStatusIndex--
	} else if msg.Direction == github.DirectionRight {
		newStatusIndex++
	}

	// Check if the new index is valid
	if newStatusIndex < 0 || newStatusIndex >= len(statusField.Options) {
		return a, nil // Cannot move further
	}

	newStatusOption := statusField.Options[newStatusIndex]

	// Call the GitHub client to update the status
	cmd := func() tea.Msg {
		updatedItem, err := a.github.UpdateStatus(
			context.Background(),
			a.state.Project.ID,
			a.state.Project.Owner,
			item.ID, // Use the item's node ID for field updates
			statusField.ID,
			newStatusOption.ID,
		)
		if err != nil {
			return NewErrMsg(err)
		}
		return ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	return a, tea.Batch(cmd)
}
