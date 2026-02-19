package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

func (a App) updateStatusSelectMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	updatedStatusSelector, statusSelectorCmd := a.statusSelector.Update(msg)
	a.statusSelector = updatedStatusSelector.(components.StatusSelectorModel)
	if statusSelectorCmd != nil {
		cmds = append(cmds, statusSelectorCmd)
	}

	switch m := msg.(type) {
	case components.StatusSelectedMsg:
		if m.Canceled {
			a.state.View.Mode = state.ModeNormal
			return a, tea.Batch(cmds...)
		}
		idx := a.state.View.FocusedIndex
		if idx < 0 || idx >= len(a.state.Items) {
			a.state.View.Mode = state.ModeNormal
			return a, tea.Batch(cmds...)
		}
		item := a.state.Items[idx]

		if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
			notif := state.Notification{
				Message:      fmt.Sprintf("Invalid item ID format: %s. Expected project item node ID.", item.ID),
				Level:        "error",
				At:           time.Now(),
				DismissAfter: 5 * time.Second,
			}
			a.state.Notifications = append(a.state.Notifications, notif)
			a.state.View.Mode = state.ModeNormal
			return a, tea.Batch(append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))...)
		}

		updateCmd := func() tea.Msg {
			updatedItem, err := a.github.UpdateStatus(
				context.Background(),
				projectMutationID(a.state.Project),
				a.state.Project.Owner,
				item.ID,
				m.StatusFieldID,
				m.OptionID,
			)
			if err != nil {
				return NewErrMsg(err)
			}
			if strings.EqualFold(strings.TrimSpace(updatedItem.Status), "unknown") && strings.TrimSpace(m.OptionName) != "" {
				updatedItem.Status = m.OptionName
			}
			return ItemUpdatedMsg{Index: idx, Item: updatedItem}
		}

		a.state.View.Mode = state.ModeNormal
		return a, tea.Batch(append(cmds, updateCmd)...)
	default:
		return a, tea.Batch(cmds...)
	}
}

func (a App) updateFieldSelectMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	updatedFieldSelector, fieldSelectorCmd := a.fieldSelector.Update(msg)
	a.fieldSelector = updatedFieldSelector.(components.FieldSelectorModel)
	if fieldSelectorCmd != nil {
		cmds = append(cmds, fieldSelectorCmd)
	}

	switch m := msg.(type) {
	case components.FieldSelectedMsg:
		if m.Canceled {
			a.state.View.Mode = state.ModeNormal
			return a, tea.Batch(cmds...)
		}
		idx := a.state.View.FocusedIndex
		if idx < 0 || idx >= len(a.state.Items) {
			a.state.View.Mode = state.ModeNormal
			return a, tea.Batch(cmds...)
		}
		item := a.state.Items[idx]

		if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
			notif := state.Notification{
				Message:      fmt.Sprintf("Invalid item ID format: %s. Expected project item node ID.", item.ID),
				Level:        "error",
				At:           time.Now(),
				DismissAfter: 5 * time.Second,
			}
			a.state.Notifications = append(a.state.Notifications, notif)
			a.state.View.Mode = state.ModeNormal
			return a, tea.Batch(append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))...)
		}

		updateCmd := func() tea.Msg {
			updatedItem, err := a.github.UpdateField(
				context.Background(),
				projectMutationID(a.state.Project),
				a.state.Project.Owner,
				item.ID,
				m.FieldID,
				m.OptionID,
				m.FieldName,
			)
			if err != nil {
				return NewErrMsg(err)
			}
			return ItemUpdatedMsg{Index: idx, Item: updatedItem}
		}

		a.state.View.Mode = state.ModeNormal
		return a, tea.Batch(append(cmds, updateCmd)...)
	default:
		return a, tea.Batch(cmds...)
	}
}

func (a App) updateDetailMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	updatedDetailPanel, detailPanelCmd := a.detailPanel.Update(msg)
	a.detailPanel = updatedDetailPanel.(components.DetailPanelModel)
	if detailPanelCmd != nil {
		cmds = append(cmds, detailPanelCmd)
	}

	switch msg.(type) {
	case components.DetailCloseMsg:
		a.state.View.Mode = state.ModeNormal
		return a, tea.Batch(cmds...)
	default:
		return a, tea.Batch(cmds...)
	}
}
