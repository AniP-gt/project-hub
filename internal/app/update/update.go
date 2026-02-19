package update

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app/core"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
	"project-hub/internal/ui/settings"
)

func Update(s State, msg tea.Msg) (State, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if s.Model.View.Mode == state.ModeStatusSelect {
		updated, statusCmd := StatusSelectMode(s, msg)
		cmds = append(cmds, statusCmd)
		return updated, tea.Batch(cmds...)
	}

	if s.Model.View.Mode == state.ModeLabelSelect || s.Model.View.Mode == state.ModeMilestoneSelect || s.Model.View.Mode == state.ModePrioritySelect {
		updated, fieldCmd := FieldSelectMode(s, msg)
		cmds = append(cmds, fieldCmd)
		return updated, tea.Batch(cmds...)
	}

	if s.Model.View.Mode == state.ModeDetail {
		updated, detailCmd := DetailMode(s, msg)
		cmds = append(cmds, detailCmd)
		return updated, tea.Batch(cmds...)
	}

	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.Model.Width = m.Width
		s.Model.Height = m.Height
		if s.Model.View.CurrentView == state.ViewBoard {
			var model tea.Model
			model, cmd = s.BoardModel.Update(msg)
			s.BoardModel = model.(boardPkg.BoardModel)
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		if s.Model.View.CurrentView == state.ViewSettings {
			updatedModel, settingsCmd := s.SettingsModel.Update(m)
			s.SettingsModel = updatedModel
			if settingsCmd != nil {
				cmds = append(cmds, settingsCmd)
			}
		} else {
			updated, keyCmd := HandleKey(s, m)
			s = updated
			cmds = append(cmds, keyCmd)
		}
	case core.FetchProjectMsg:
		s.Model.Project = m.Project
		s.Model.Items = m.Items
		if s.Model.ExcludeDone {
			var filtered []state.Item
			for _, item := range s.Model.Items {
				if item.Status != "Done" {
					filtered = append(filtered, item)
				}
			}
			s.Model.Items = filtered
		}
		if len(s.Model.Items) > 0 {
			s.Model.View.FocusedItemID = s.Model.Items[0].ID
		}
		s.BoardModel = boardPkg.NewBoardModel(s.Model.Items, s.Model.Project.Fields, s.Model.View.Filter, s.Model.View.FocusedItemID, s.Model.View.CardFieldVisibility)
	case core.ItemUpdatedMsg:
		if m.Index >= 0 && m.Index < len(s.Model.Items) {
			existing := s.Model.Items[m.Index]
			if len(m.Item.Assignees) > 0 {
				existing.Assignees = m.Item.Assignees
			}
			if len(m.Item.Labels) > 0 {
				existing.Labels = m.Item.Labels
			}
			if m.Item.Title != "" {
				existing.Title = m.Item.Title
			}
			if m.Item.Status != "" && m.Item.Status != "Unknown" {
				existing.Status = m.Item.Status
			}
			if m.Item.Priority != "" {
				existing.Priority = m.Item.Priority
			}
			if m.Item.Repository != "" {
				existing.Repository = m.Item.Repository
			}
			if m.Item.Number != 0 {
				existing.Number = m.Item.Number
			}
			s.Model.Items[m.Index] = existing
		} else {
			if m.Index >= 0 && m.Index <= len(s.Model.Items) {
				s.Model.Items[m.Index] = m.Item
			}
		}
		s.BoardModel = boardPkg.NewBoardModel(s.Model.Items, s.Model.Project.Fields, s.Model.View.Filter, s.Model.View.FocusedItemID, s.Model.View.CardFieldVisibility)
		if !s.Model.SuppressHints {
			notif := state.Notification{Message: "Item updated successfully", Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
			s.Model.Notifications = append(s.Model.Notifications, notif)
			cmds = append(cmds, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter))
		}
	case core.DetailReadyMsg:
		s.DetailPanel = components.NewDetailPanelModel(m.Item, s.Model.Width, s.Model.Height)
		if !s.Model.SuppressHints {
			detailNotif := state.Notification{Message: "Detail mode: j/k to scroll, esc/q to close", Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
			s.Model.Notifications = append(s.Model.Notifications, detailNotif)
			cmds = append(cmds, tea.Batch(s.DetailPanel.Init(), core.DismissNotificationCmd(len(s.Model.Notifications)-1, detailNotif.DismissAfter)))
		} else {
			cmds = append(cmds, s.DetailPanel.Init())
		}
	case core.ErrMsg:
		notif := state.Notification{Message: fmt.Sprintf("Error: %v", m.Err), Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		cmds = append(cmds, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter))
	case core.DismissNotificationMsg:
		if m.ID >= 0 && m.ID < len(s.Model.Notifications) {
			s.Model.Notifications[m.ID].Dismissed = true
		}
	case SwitchViewMsg:
		updated, switchCmd := SwitchView(s, m)
		s = updated
		cmds = append(cmds, switchCmd)
	case MoveFocusMsg:
		updated, moveCmd := MoveFocus(s, m)
		s = updated
		cmds = append(cmds, moveCmd)
	case EnterFilterModeMsg:
		updated, filterCmd := EnterFilterMode(s, m)
		s = updated
		cmds = append(cmds, filterCmd)
	case ApplyFilterMsg:
		updated, filterCmd := ApplyFilter(s, m)
		s = updated
		cmds = append(cmds, filterCmd)
	case ClearFilterMsg:
		updated, filterCmd := ClearFilter(s, m)
		s = updated
		cmds = append(cmds, filterCmd)
	case EnterEditModeMsg:
		updated, editCmd := EnterEditMode(s, m)
		s = updated
		cmds = append(cmds, editCmd)
	case SaveEditMsg:
		updated, editCmd := SaveEdit(s, m)
		s = updated
		cmds = append(cmds, editCmd)
	case CancelEditMsg:
		updated, editCmd := CancelEdit(s, m)
		s = updated
		cmds = append(cmds, editCmd)
	case EnterAssignModeMsg:
		updated, assignCmd := EnterAssignMode(s, m)
		s = updated
		cmds = append(cmds, assignCmd)
	case SaveAssignMsg:
		updated, assignCmd := SaveAssign(s, m)
		s = updated
		cmds = append(cmds, assignCmd)
	case CancelAssignMsg:
		updated, assignCmd := CancelAssign(s, m)
		s = updated
		cmds = append(cmds, assignCmd)
	case AssignMsg:
		updated, assignCmd := Assign(s, m)
		s = updated
		cmds = append(cmds, assignCmd)
	case core.EnterStatusSelectModeMsg:
		updated, statusCmd := EnterStatusSelectMode(s, m)
		s = updated
		cmds = append(cmds, statusCmd)
	case settings.SaveMsg:
		updated, settingsCmd := SettingsSave(s, m)
		s = updated
		if settingsCmd != nil {
			cmds = append(cmds, settingsCmd)
		}
	case settings.CancelMsg:
		updated, settingsCmd := SettingsCancel(s, m)
		s = updated
		if settingsCmd != nil {
			cmds = append(cmds, settingsCmd)
		}
	default:
	}
	return s, tea.Batch(cmds...)
}
