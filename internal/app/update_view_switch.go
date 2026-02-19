package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
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

	// Check if we're in table view with grouping active
	if a.state.View.CurrentView == state.ViewTable && a.state.View.TableGroupBy != "" {
		return a.handleMoveFocusGrouped(msg.Delta)
	}

	// Default: move within raw items list
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

// handleMoveFocusGrouped moves focus respecting grouped table view order.
func (a App) handleMoveFocusGrouped(delta int) (tea.Model, tea.Cmd) {
	groupBy := strings.ToLower(strings.TrimSpace(a.state.View.TableGroupBy))
	var groups []boardPkg.GroupBucket

	switch groupBy {
	case groupByStatus:
		groups = boardPkg.GroupItemsByStatusBuckets(a.state.Items, a.state.Project.Fields)
	case groupByIteration:
		groups = boardPkg.GroupItemsByIteration(a.state.Items)
	case groupByAssignee:
		groups = boardPkg.GroupItemsByAssignee(a.state.Items)
	default:
		return a.handleMoveFocus(MoveFocusMsg{Delta: delta})
	}

	if len(groups) == 0 {
		return a, nil
	}

	currentID := a.state.View.FocusedItemID

	type rowInfo struct {
		itemID  string
		isGroup bool
		groupID int
	}

	var rowToItem []rowInfo
	for gi, group := range groups {
		rowToItem = append(rowToItem, rowInfo{isGroup: true, groupID: gi})
		for _, item := range group.Items {
			rowToItem = append(rowToItem, rowInfo{itemID: item.ID, isGroup: false, groupID: gi})
		}
	}

	currentRow := -1
	for i, r := range rowToItem {
		if !r.isGroup && r.itemID == currentID {
			currentRow = i
			break
		}
	}
	if currentRow < 0 {
		currentRow = 0
	}

	newRow := currentRow + delta
	for newRow < 0 {
		newRow += len(rowToItem)
	}
	newRow = newRow % len(rowToItem)

	newItemID := ""
	for i := newRow; i < len(rowToItem); i++ {
		if !rowToItem[i].isGroup && rowToItem[i].itemID != "" {
			newItemID = rowToItem[i].itemID
			break
		}
	}
	if newItemID == "" {
		for i := 0; i < newRow; i++ {
			if !rowToItem[i].isGroup && rowToItem[i].itemID != "" {
				newItemID = rowToItem[i].itemID
				break
			}
		}
	}

	if newItemID != "" {
		a.state.View.FocusedItemID = newItemID
		for idx, item := range a.state.Items {
			if item.ID == newItemID {
				a.state.View.FocusedIndex = idx
				break
			}
		}
	}

	if a.tableViewport != nil {
		a.tableViewport.YOffset = 0
	}

	return a, nil
}
