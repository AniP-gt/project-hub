package update

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app/core"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
)

type SwitchViewMsg struct {
	View state.ViewType
}

type MoveFocusMsg struct {
	Delta int
}

func SwitchView(s State, msg SwitchViewMsg) (State, tea.Cmd) {
	s.Model.View.CurrentView = msg.View
	s = syncTableColumnIndex(s)
	if len(s.Model.Items) == 0 {
		s.Model.View.FocusedItemID = ""
		s.Model.View.FocusedIndex = -1
	}
	return s, nil
}

func MoveFocus(s State, msg MoveFocusMsg) (State, tea.Cmd) {
	if len(s.Model.Items) == 0 {
		return s, nil
	}

	if s.Model.View.CurrentView == state.ViewTable && s.Model.View.TableGroupBy != "" {
		return MoveFocusGrouped(s, msg.Delta)
	}

	if s.Model.View.CurrentView == state.ViewTable {
		filteredItems := state.ApplyFilter(s.Model.Items, s.Model.Project.Fields, s.Model.View.Filter, time.Now())
		if len(filteredItems) == 0 {
			s.Model.View.FocusedItemID = ""
			s.Model.View.FocusedIndex = -1
			return s, nil
		}
		idx := -1
		for i, item := range filteredItems {
			if item.ID == s.Model.View.FocusedItemID {
				idx = i
				break
			}
		}
		if idx < 0 {
			idx = 0
		}
		idx += msg.Delta
		if idx < 0 {
			idx = 0
		}
		if idx >= len(filteredItems) {
			idx = len(filteredItems) - 1
		}
		newItemID := filteredItems[idx].ID
		s.Model.View.FocusedItemID = newItemID
		for modelIdx, item := range s.Model.Items {
			if item.ID == newItemID {
				s.Model.View.FocusedIndex = modelIdx
				break
			}
		}
		return s, nil
	}

	idx := s.Model.View.FocusedIndex
	if idx < 0 {
		idx = 0
	}
	idx += msg.Delta
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.Model.Items) {
		idx = len(s.Model.Items) - 1
	}
	s.Model.View.FocusedIndex = idx
	s.Model.View.FocusedItemID = s.Model.Items[idx].ID
	return s, nil
}

func MoveFocusGrouped(s State, delta int) (State, tea.Cmd) {
	groupBy := strings.ToLower(strings.TrimSpace(s.Model.View.TableGroupBy))
	var groups []boardPkg.GroupBucket
	filteredItems := state.ApplyFilter(s.Model.Items, s.Model.Project.Fields, s.Model.View.Filter, time.Now())

	switch groupBy {
	case core.GroupByStatus:
		groups = boardPkg.GroupItemsByStatusBuckets(filteredItems, s.Model.Project.Fields)
	case core.GroupByIteration:
		groups = boardPkg.GroupItemsByIteration(filteredItems)
	case core.GroupByAssignee:
		groups = boardPkg.GroupItemsByAssignee(filteredItems)
	default:
		return MoveFocus(s, MoveFocusMsg{Delta: delta})
	}

	if len(groups) == 0 {
		return s, nil
	}

	currentID := s.Model.View.FocusedItemID

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
	if delta < 0 {
		for i := newRow; i >= 0; i-- {
			if !rowToItem[i].isGroup && rowToItem[i].itemID != "" {
				newItemID = rowToItem[i].itemID
				break
			}
		}
		if newItemID == "" {
			for i := len(rowToItem) - 1; i > newRow; i-- {
				if !rowToItem[i].isGroup && rowToItem[i].itemID != "" {
					newItemID = rowToItem[i].itemID
					break
				}
			}
		}
	} else {
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
	}

	if newItemID != "" {
		s.Model.View.FocusedItemID = newItemID
		for idx, item := range s.Model.Items {
			if item.ID == newItemID {
				s.Model.View.FocusedIndex = idx
				break
			}
		}
	}

	return s, nil
}
