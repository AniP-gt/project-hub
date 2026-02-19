package update

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app/core"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
)

func HandleKey(s State, k tea.KeyMsg) (State, tea.Cmd) {
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
		case "s", "S":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Status")
		case "r", "R":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Repository")
		case "L":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Labels")
		case "m", "M":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Milestone")
		case "p", "P":
			s.Model.View.TableSort = toggleSort(s.Model.View.TableSort, "Priority")
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
			s.Model.View.FocusedColumnIndex--
			if s.Model.View.FocusedColumnIndex < 0 {
				s.Model.View.FocusedColumnIndex = 0
			}
			return s, nil
		case "l":
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
	}

	if s.Model.View.CurrentView == state.ViewBoard {
		switch k.String() {
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
	case "h":
		s.Model.View.FocusedColumnIndex--
		if s.Model.View.FocusedColumnIndex < 0 {
			s.Model.View.FocusedColumnIndex = 0
		}
		return s, nil
	case "l":
		maxCols := state.ColumnCount
		s.Model.View.FocusedColumnIndex++
		if s.Model.View.FocusedColumnIndex >= maxCols {
			s.Model.View.FocusedColumnIndex = maxCols - 1
		}
		return s, nil
	case "/":
		return EnterFilterMode(s, EnterFilterModeMsg{})
	case "s":
		if s.Model.View.CurrentView == state.ViewTable {
			s.Model.View.Mode = state.ModeSort
			if !s.Model.SuppressHints {
				notif := state.Notification{Message: "Sort mode: t=Title s=Status r=Repository l=Labels m=Milestone p=Priority n=Number c=CreatedAt u=UpdatedAt (esc to cancel)", Level: "info", At: time.Now(), DismissAfter: 5 * time.Second}
				s.Model.Notifications = append(s.Model.Notifications, notif)
				return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
			}
			return s, nil
		}
		return s, nil
	case "esc":
		return ClearFilter(s, ClearFilterMsg{})
	case "i", "enter":
		if s.Model.View.CurrentView == state.ViewTable {
			return ColumnEdit(s, EnterEditModeMsg{})
		}
		return EnterEditMode(s, EnterEditModeMsg{})
	case "a":
		return EnterAssignMode(s, EnterAssignModeMsg{})
	case "w":
		if s.Model.View.CurrentView == state.ViewTable {
			return EnterStatusSelectMode(s, core.EnterStatusSelectModeMsg{})
		}
		return s, nil
	case "f":
		if s.Model.View.CurrentView == state.ViewBoard {
			return EnterFieldToggleMode(s)
		}
		return s, nil
	case "g":
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

func toggleSort(ts state.TableSort, field string) state.TableSort {
	if ts.Field == field {
		ts.Asc = !ts.Asc
		return ts
	}
	return state.TableSort{Field: field, Asc: true}
}
