package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/github"
	boardPkg "project-hub/internal/ui/board"
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

	currentStatus := item.Status
	currentStatusIndex := -1
	for i, status := range boardPkg.ColumnOrder {
		if status == currentStatus {
			currentStatusIndex = i
			break
		}
	}

	if currentStatusIndex == -1 {
		// Current status not found in defined order, cannot move
		return a, func() tea.Msg {
			return NewErrMsg(fmt.Errorf("current status '%s' not found in column order", currentStatus))
		}
	}

	newStatusIndex := currentStatusIndex
	if msg.Direction == github.DirectionLeft {
		newStatusIndex--
	} else if msg.Direction == github.DirectionRight {
		newStatusIndex++
	}

	if newStatusIndex < 0 || newStatusIndex >= len(boardPkg.ColumnOrder) {
		// Cannot move further in this direction
		return a, nil
	}

	newStatus := boardPkg.ColumnOrder[newStatusIndex]

	// Call the GitHub client to update the status
	cmd := func() tea.Msg {
		updatedItem, err := a.github.UpdateStatus(
			context.Background(),
			a.state.Project.ID,
			a.state.Project.Owner,
			item.ID,
			newStatus,
		)
		if err != nil {
			return NewErrMsg(err)
		}
		return ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	return a, tea.Batch(cmd, a.refreshBoardCmd())
}
