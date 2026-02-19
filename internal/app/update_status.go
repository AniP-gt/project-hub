package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

func (a App) handleEnterStatusSelectMode(msg EnterStatusSelectModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	focusedItem := a.state.Items[idx]

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
		notif := state.Notification{Message: fmt.Sprintf("Status field not found in project"), Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	a.statusSelector = components.NewStatusSelectorModel(focusedItem, statusField, a.state.Width, a.state.Height)
	a.state.View.Mode = state.ModeStatusSelect

	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Status select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, tea.Batch(a.statusSelector.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	}
	return a, a.statusSelector.Init()
}
