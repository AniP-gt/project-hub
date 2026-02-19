package update

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app/core"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

func EnterStatusSelectMode(s State, _ core.EnterStatusSelectModeMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	focusedItem := s.Model.Items[idx]

	var statusField state.Field
	found := false
	for _, field := range s.Model.Project.Fields {
		if field.Name == "Status" {
			statusField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: fmt.Sprintf("Status field not found in project"), Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}

	s.StatusSelector = components.NewStatusSelectorModel(focusedItem, statusField, s.Model.Width, s.Model.Height)
	s.Model.View.Mode = state.ModeStatusSelect

	if !s.Model.SuppressHints {
		notif := state.Notification{
			Message:      "Status select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, tea.Batch(s.StatusSelector.Init(), core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter))
	}
	return s, s.StatusSelector.Init()
}
