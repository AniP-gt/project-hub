package update

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app/core"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
)

// pendingKeyTimeout is an internal message emitted by a Tick to signal expiry of a prefix key sequence.
type pendingKeyTimeout struct {
	Key string
	At  time.Time
}

func HandleKey(s State, k tea.KeyMsg) (State, tea.Cmd) {
	// Support vim-style navigation: 'g' -> go to top, 'G' (shift+g) -> go to bottom.
	if s.Model.View.Mode == "edit" || s.Model.View.Mode == "assign" || s.Model.View.Mode == "labelsInput" || s.Model.View.Mode == "milestoneInput" || s.Model.View.Mode == state.ModeFiltering {
		switch k.String() {
		case "enter":
			if s.Model.View.Mode == "edit" {
				return SaveEdit(s, SaveEditMsg{Title: s.TextInput.Value()})
			} else if s.Model.View.Mode == "assign" {
				return SaveAssign(s, SaveAssignMsg{Assignee: s.TextInput.Value()})
			} else if s.Model.View.Mode == "labelsInput" {
				return SaveLabelsInput(s, SaveLabelsInputMsg{Labels: s.TextInput.Value()})
			} else if s.Model.View.Mode == "milestoneInput" {
				return SaveMilestoneInput(s, SaveMilestoneInputMsg{Milestone: s.TextInput.Value()})
			} else if s.Model.View.Mode == state.ModeFiltering {
				return ApplyFilter(s, ApplyFilterMsg{Query: s.TextInput.Value()})
			}
		case "esc":
			if s.Model.View.Mode == "edit" {
				return CancelEdit(s, CancelEditMsg{})
			} else if s.Model.View.Mode == "assign" {
				return CancelAssign(s, CancelAssignMsg{})
			} else if s.Model.View.Mode == "labelsInput" {
				return CancelLabelsInput(s, CancelLabelsInputMsg{})
			} else if s.Model.View.Mode == "milestoneInput" {
				return CancelMilestoneInput(s, CancelMilestoneInputMsg{})
			} else if s.Model.View.Mode == state.ModeFiltering {
				return ClearFilter(s, ClearFilterMsg{})
			}
		default:
			var cmd tea.Cmd
			s.TextInput, cmd = s.TextInput.Update(k)
			return s, cmd
		}
	}

	if s.Model.View.Mode == state.ModeFieldToggle {
		return FieldToggle(s, k.String())
	}

	if s.Model.View.Mode == state.ModeSort {
		switch k.String() {
		case "t", "T":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Title")
		case "S":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Status")
		case "r", "R":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Repository")
		case "L":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Labels")
		case "m", "M":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Milestone")
		case "p", "P":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Priority")
		case "a", "A":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Assignees")
		case "n", "N":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Number")
		case "c", "C":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "CreatedAt")
		case "u", "U":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "UpdatedAt")
		case "j", "down":
			return MoveFocus(s, MoveFocusMsg{Delta: 1})
		case "k", "up":
			return MoveFocus(s, MoveFocusMsg{Delta: -1})
		case "h":
			if s.Model.View.CurrentView == state.ViewTable {
				return moveTableColumn(s, -1), nil
			}
			s.Model.View.FocusedColumnIndex--
			if s.Model.View.FocusedColumnIndex < 0 {
				s.Model.View.FocusedColumnIndex = 0
			}
			return s, nil
		case "l":
			if s.Model.View.CurrentView == state.ViewTable {
				return moveTableColumn(s, 1), nil
			}
			maxCols := state.ColumnCount
			s.Model.View.FocusedColumnIndex++
			if s.Model.View.FocusedColumnIndex >= maxCols {
				s.Model.View.FocusedColumnIndex = maxCols - 1
			}
			return s, nil
		case "esc":
			s.Model.View.Mode = state.ModeNormal
			return s, nil
		default:
			return s, nil
		}

		if len(s.Model.Items) > 0 {
			s.Model.View.FocusedIndex = 0
			s.Model.View.FocusedItemID = s.Model.Items[0].ID
		}
		s.Model.View.Mode = state.ModeNormal
		if !s.Model.SuppressHints {
			notif := state.Notification{Message: fmt.Sprintf("Sort: %s %s", s.Model.View.TableSort.Field, func() string {
				if s.Model.View.TableSort.Asc {
					return "↑"
				}
				return "↓"
			}()), Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
			s.Model.Notifications = append(s.Model.Notifications, notif)
			return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
		}
		// Always return after handling Sort mode so that top-level key handlers
		// do not also process the same key press (which could re-enter assign
		// mode when SuppressHints is true and the notification branch above
		// doesn't return).
		return s, nil
	}

	// Handle board-specific keys, including g/G
	if s.Model.View.CurrentView == state.ViewBoard {
		key := k.String()
		// Immediate navigation keys handled by BoardModel
		switch key {
		case "j", "k", "h", "l":
			model, cmd := s.BoardModel.Update(k)
			s.BoardModel = model.(boardPkg.BoardModel)
			s = SyncFocusedItem(s)
			return s, cmd
		case "/":
			return EnterFilterMode(s, EnterFilterModeMsg{})
		case "i":
			return EnterEditMode(s, EnterEditModeMsg{})
		case "a":
			return EnterAssignMode(s, EnterAssignModeMsg{})
		case "w":
			return EnterStatusSelectMode(s, core.EnterStatusSelectModeMsg{})
		case "G", "shift+G":
			// Shift+G -> go to bottom of current column
			colIdx := s.BoardModel.FocusedColumnIndex
			if colIdx >= 0 && colIdx < len(s.BoardModel.Columns) {
				current := s.BoardModel.Columns[colIdx]
				if len(current.Cards) > 0 {
					s.BoardModel.FocusedCardIndex = len(current.Cards) - 1
					// adjust CardOffset so the focused card is visible at bottom
					maxVisible := 3
					if s.BoardModel.Height > 0 {
						estCard := 6
						maxVisible = s.BoardModel.Height/estCard - 1
						if maxVisible < 1 {
							maxVisible = 1
						}
					}
					s.BoardModel.CardOffset = s.BoardModel.FocusedCardIndex - (maxVisible - 1)
					if s.BoardModel.CardOffset < 0 {
						s.BoardModel.CardOffset = 0
					}
					s = SyncFocusedItem(s)
				}
			}
			return s, nil
		case "g":
			// Single 'g' -> go to top of current column
			colIdx := s.BoardModel.FocusedColumnIndex
			if colIdx >= 0 && colIdx < len(s.BoardModel.Columns) {
				s.BoardModel.FocusedCardIndex = 0
				s.BoardModel.CardOffset = 0
				s = SyncFocusedItem(s)
			}
			return s, nil
		}
	}
	switch k.String() {
	case "ctrl+c", "q":
		return s, tea.Quit
	case "1", "b":
		return SwitchView(s, SwitchViewMsg{View: state.ViewBoard})
	case "2", "t":
		return SwitchView(s, SwitchViewMsg{View: state.ViewTable})
	case "3":
		return SwitchView(s, SwitchViewMsg{View: state.ViewSettings})
	case "R", "ctrl+r":
		return s, core.FetchProjectCmd(s.Github, s.Model.Project.ID, s.Model.Project.Owner, s.ItemLimit)
	case "j":
		return MoveFocus(s, MoveFocusMsg{Delta: 1})
	case "k":
		return MoveFocus(s, MoveFocusMsg{Delta: -1})
	case "g":
		if s.Model.View.Mode != state.ModeNormal {
			return s, nil
		}
		if s.Model.View.CurrentView == state.ViewTable {
			if s.Model.View.TableGroupBy != "" {
				return moveTableFocusToGroupedTop(s), nil
			}
			return moveTableFocusToTop(s), nil
		}
		return s, nil
	case "G", "shift+G":
		if s.Model.View.Mode != state.ModeNormal {
			return s, nil
		}
		if s.Model.View.CurrentView == state.ViewTable {
			if s.Model.View.TableGroupBy != "" {
				return moveTableFocusToGroupedBottom(s), nil
			}
			return moveTableFocusToBottom(s), nil
		}
		return s, nil
	case "h":
		if s.Model.View.CurrentView == state.ViewTable {
			return moveTableColumn(s, -1), nil
		}
		s.Model.View.FocusedColumnIndex--
		if s.Model.View.FocusedColumnIndex < 0 {
			s.Model.View.FocusedColumnIndex = 0
		}
		return s, nil
	case "l":
		if s.Model.View.CurrentView == state.ViewTable {
			return moveTableColumn(s, 1), nil
		}
		maxCols := state.ColumnCount
		s.Model.View.FocusedColumnIndex++
		if s.Model.View.FocusedColumnIndex >= maxCols {
			s.Model.View.FocusedColumnIndex = maxCols - 1
		}
		return s, nil
	case "/":
		if s.Model.View.Mode == state.ModeNormal {
			return EnterFilterMode(s, EnterFilterModeMsg{})
		}
		return s, nil
	case "s":
		if s.Model.View.Mode == state.ModeNormal && s.Model.View.CurrentView == state.ViewTable {
			s.Model.View.Mode = state.ModeSort
			// Enter Sort mode without exposing a keymap.
			return s, nil
		}
		return s, nil
	case "esc":
		return ClearFilter(s, ClearFilterMsg{})
	case "i", "enter":
		if s.Model.View.Mode != state.ModeNormal {
			return s, nil
		}
		if s.Model.View.CurrentView == state.ViewTable {
			return ColumnEdit(s, EnterEditModeMsg{})
		}
		return EnterEditMode(s, EnterEditModeMsg{})
	case "a":
		if s.Model.View.Mode == state.ModeNormal {
			return EnterAssignMode(s, EnterAssignModeMsg{})
		}
		return s, nil
	case "w":
		if s.Model.View.Mode != state.ModeNormal {
			return s, nil
		}
		if s.Model.View.CurrentView == state.ViewTable {
			return EnterStatusSelectMode(s, core.EnterStatusSelectModeMsg{})
		}
		return s, nil
	case "f":
		if s.Model.View.Mode != state.ModeNormal {
			return s, nil
		}
		if s.Model.View.CurrentView == state.ViewBoard || s.Model.View.CurrentView == state.ViewTable {
			return EnterFieldToggleMode(s)
		}
		return s, nil
	case "m":
		if s.Model.View.Mode != state.ModeNormal {
			return s, nil
		}
		if s.Model.View.CurrentView == state.ViewTable {
			return ToggleGroupBy(s)
		}
		return s, nil
	case "o":
		return EnterDetailMode(s)
	case "O":
		// Open the focused item's URL in the browser if available
		idx := s.Model.View.FocusedIndex
		if idx >= 0 && idx < len(s.Model.Items) {
			url := s.Model.Items[idx].URL
			if url != "" {
				return s, core.OpenBrowserCmd(url)
			}
		}
		return s, nil
	case "y":
		// Copy focused item's URL to clipboard
		idx := s.Model.View.FocusedIndex
		if idx >= 0 && idx < len(s.Model.Items) {
			url := s.Model.Items[idx].URL
			if url != "" {
				return s, core.CopyToClipboardCmd(url)
			}
		}
		return s, nil
	default:
		return s, nil
	}
}

func moveTableFocusToTop(s State) State {
	if s.Model.View.CurrentView != state.ViewTable {
		return s
	}
	filteredItems := state.ApplyFilter(s.Model.Items, s.Model.Project.Fields, s.Model.View.Filter, time.Now())
	filteredItems = state.ApplyTableSort(filteredItems, s.Model.View.TableSort)
	if len(filteredItems) == 0 {
		s.Model.View.FocusedItemID = ""
		s.Model.View.FocusedIndex = -1
		return s
	}
	firstID := filteredItems[0].ID
	s.Model.View.FocusedItemID = firstID
	for idx, item := range s.Model.Items {
		if item.ID == firstID {
			s.Model.View.FocusedIndex = idx
			break
		}
	}
	if s.TableViewport != nil {
		s.TableViewport.YOffset = 0
	}
	return s
}

func moveTableFocusToBottom(s State) State {
	if s.Model.View.CurrentView != state.ViewTable {
		return s
	}
	filteredItems := state.ApplyFilter(s.Model.Items, s.Model.Project.Fields, s.Model.View.Filter, time.Now())
	filteredItems = state.ApplyTableSort(filteredItems, s.Model.View.TableSort)
	if len(filteredItems) == 0 {
		s.Model.View.FocusedItemID = ""
		s.Model.View.FocusedIndex = -1
		return s
	}
	lastID := filteredItems[len(filteredItems)-1].ID
	s.Model.View.FocusedItemID = lastID
	for idx, item := range s.Model.Items {
		if item.ID == lastID {
			s.Model.View.FocusedIndex = idx
			break
		}
	}
	if s.TableViewport != nil {
		s.TableViewport.YOffset = s.TableViewport.TotalLineCount() - s.TableViewport.Height
		if s.TableViewport.YOffset < 0 {
			s.TableViewport.YOffset = 0
		}
	}
	return s
}

func moveTableFocusToGroupedTop(s State) State {
	return moveTableFocusToGroupedEdge(s, true)
}

func moveTableFocusToGroupedBottom(s State) State {
	return moveTableFocusToGroupedEdge(s, false)
}

func moveTableFocusToGroupedEdge(s State, top bool) State {
	if s.Model.View.CurrentView != state.ViewTable {
		return s
	}
	groupBy := strings.ToLower(strings.TrimSpace(s.Model.View.TableGroupBy))
	if groupBy == "" {
		return s
	}
	filteredItems := state.ApplyFilter(s.Model.Items, s.Model.Project.Fields, s.Model.View.Filter, time.Now())
	filteredItems = state.ApplyTableSort(filteredItems, s.Model.View.TableSort)

	var groups []boardPkg.GroupBucket
	switch groupBy {
	case core.GroupByStatus:
		groups = boardPkg.GroupItemsByStatusBuckets(filteredItems, s.Model.Project.Fields)
	case core.GroupByIteration:
		groups = boardPkg.GroupItemsByIteration(filteredItems)
	case core.GroupByAssignee:
		groups = boardPkg.GroupItemsByAssignee(filteredItems)
	default:
		return s
	}
	if len(groups) == 0 {
		return s
	}

	var targetID string
	if top {
		for _, group := range groups {
			if len(group.Items) > 0 {
				targetID = group.Items[0].ID
				break
			}
		}
	} else {
		for gi := len(groups) - 1; gi >= 0; gi-- {
			group := groups[gi]
			if len(group.Items) > 0 {
				targetID = group.Items[len(group.Items)-1].ID
				break
			}
		}
	}

	if targetID == "" {
		s.Model.View.FocusedItemID = ""
		s.Model.View.FocusedIndex = -1
		return s
	}

	s.Model.View.FocusedItemID = targetID
	for idx, item := range s.Model.Items {
		if item.ID == targetID {
			s.Model.View.FocusedIndex = idx
			break
		}
	}

	if s.TableViewport != nil {
		if top {
			s.TableViewport.YOffset = 0
		} else {
			s.TableViewport.YOffset = s.TableViewport.TotalLineCount() - s.TableViewport.Height
			if s.TableViewport.YOffset < 0 {
				s.TableViewport.YOffset = 0
			}
		}
	}

	return s
}

func moveTableColumn(s State, delta int) State {
	visible := tableVisibleColumns(s.Model.View.CardFieldVisibility)
	if len(visible) == 0 {
		s.Model.View.FocusedColumnIndex = 0
		return s
	}
	col := s.Model.View.FocusedColumnIndex + delta
	if col < 0 {
		col = 0
	}
	if col >= len(visible) {
		col = len(visible) - 1
	}
	s.Model.View.FocusedColumnIndex = col
	return s
}

func tableVisibleColumns(vis state.CardFieldVisibility) []int {
	columns := []int{state.ColumnTitle, state.ColumnStatus}
	if vis.ShowRepository {
		columns = append(columns, state.ColumnRepository)
	}
	if vis.ShowLabels {
		columns = append(columns, state.ColumnLabels)
	}
	if vis.ShowMilestone {
		columns = append(columns, state.ColumnMilestone)
	}
	if vis.ShowSubIssueProgress {
		columns = append(columns, state.ColumnSubIssueProgress)
	}
	if vis.ShowParentIssue {
		columns = append(columns, state.ColumnParentIssue)
	}
	columns = append(columns, state.ColumnAssignees)
	return columns
}

func syncTableColumnIndex(s State) State {
	if s.Model.View.CurrentView != state.ViewTable {
		return s
	}
	visible := tableVisibleColumns(s.Model.View.CardFieldVisibility)
	if len(visible) == 0 {
		s.Model.View.FocusedColumnIndex = 0
		return s
	}
	if s.Model.View.FocusedColumnIndex >= len(visible) {
		s.Model.View.FocusedColumnIndex = len(visible) - 1
	}
	if s.Model.View.FocusedColumnIndex < 0 {
		s.Model.View.FocusedColumnIndex = 0
	}
	return s
}

func toggleSort(ts state.TableSort, field string) state.TableSort {
	if ts.Field == field {
		ts.Asc = !ts.Asc
		return ts
	}
	return state.TableSort{Field: field, Asc: true}
}
