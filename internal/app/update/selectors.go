package update

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app/core"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

func StatusSelectMode(s State, msg tea.Msg) (State, tea.Cmd) {
	var cmds []tea.Cmd
	updatedStatusSelector, statusSelectorCmd := s.StatusSelector.Update(msg)
	s.StatusSelector = updatedStatusSelector.(components.StatusSelectorModel)
	if statusSelectorCmd != nil {
		cmds = append(cmds, statusSelectorCmd)
	}

	switch m := msg.(type) {
	case components.StatusSelectedMsg:
		if m.Canceled {
			s.Model.View.Mode = state.ModeNormal
			return s, tea.Batch(cmds...)
		}
		idx := s.Model.View.FocusedIndex
		if idx < 0 || idx >= len(s.Model.Items) {
			s.Model.View.Mode = state.ModeNormal
			return s, tea.Batch(cmds...)
		}
		item := s.Model.Items[idx]

		if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
			notif := state.Notification{
				Message:      fmt.Sprintf("Invalid item ID format: %s. Expected project item node ID.", item.ID),
				Level:        "error",
				At:           time.Now(),
				DismissAfter: 5 * time.Second,
			}
			s.Model.Notifications = append(s.Model.Notifications, notif)
			s.Model.View.Mode = state.ModeNormal
			return s, tea.Batch(append(cmds, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter))...)
		}

		updateCmd := func() tea.Msg {
			updatedItem, err := s.Github.UpdateStatus(
				context.Background(),
				core.ProjectMutationID(s.Model.Project),
				s.Model.Project.Owner,
				item.ID,
				m.StatusFieldID,
				m.OptionID,
			)
			if err != nil {
				return core.NewErrMsg(err)
			}
			if strings.EqualFold(strings.TrimSpace(updatedItem.Status), "unknown") && strings.TrimSpace(m.OptionName) != "" {
				updatedItem.Status = m.OptionName
			}
			return core.ItemUpdatedMsg{Index: idx, Item: updatedItem}
		}

		s.Model.View.Mode = state.ModeNormal
		return s, tea.Batch(append(cmds, updateCmd)...)
	default:
		return s, tea.Batch(cmds...)
	}
}

func FieldSelectMode(s State, msg tea.Msg) (State, tea.Cmd) {
	var cmds []tea.Cmd
	updatedFieldSelector, fieldSelectorCmd := s.FieldSelector.Update(msg)
	s.FieldSelector = updatedFieldSelector.(components.FieldSelectorModel)
	if fieldSelectorCmd != nil {
		cmds = append(cmds, fieldSelectorCmd)
	}

	switch m := msg.(type) {
	case components.FieldSelectedMsg:
		if m.Canceled {
			s.Model.View.Mode = state.ModeNormal
			return s, tea.Batch(cmds...)
		}
		idx := s.Model.View.FocusedIndex
		if idx < 0 || idx >= len(s.Model.Items) {
			s.Model.View.Mode = state.ModeNormal
			return s, tea.Batch(cmds...)
		}
		item := s.Model.Items[idx]

		if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
			notif := state.Notification{
				Message:      fmt.Sprintf("Invalid item ID format: %s. Expected project item node ID.", item.ID),
				Level:        "error",
				At:           time.Now(),
				DismissAfter: 5 * time.Second,
			}
			s.Model.Notifications = append(s.Model.Notifications, notif)
			s.Model.View.Mode = state.ModeNormal
			return s, tea.Batch(append(cmds, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter))...)
		}

		updateCmd := func() tea.Msg {
			updatedItem, err := s.Github.UpdateField(
				context.Background(),
				core.ProjectMutationID(s.Model.Project),
				s.Model.Project.Owner,
				item.ID,
				m.FieldID,
				m.OptionID,
				m.FieldName,
			)
			if err != nil {
				return core.NewErrMsg(err)
			}
			return core.ItemUpdatedMsg{Index: idx, Item: updatedItem}
		}

		s.Model.View.Mode = state.ModeNormal
		return s, tea.Batch(append(cmds, updateCmd)...)
	default:
		return s, tea.Batch(cmds...)
	}
}

func DetailMode(s State, msg tea.Msg) (State, tea.Cmd) {
	var cmds []tea.Cmd
	updatedDetailPanel, detailPanelCmd := s.DetailPanel.Update(msg)
	s.DetailPanel = updatedDetailPanel.(components.DetailPanelModel)
	if detailPanelCmd != nil {
		cmds = append(cmds, detailPanelCmd)
	}

	switch msg.(type) {
	case components.DetailCloseMsg:
		s.Model.View.Mode = state.ModeNormal
		return s, tea.Batch(cmds...)
	default:
		return s, tea.Batch(cmds...)
	}
}
