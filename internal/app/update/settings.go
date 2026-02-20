package update

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app/core"
	"project-hub/internal/config"
	"project-hub/internal/state"
	"project-hub/internal/ui/settings"
)

func SettingsSave(s State, msg settings.SaveMsg) (State, tea.Cmd) {
	configPath, err := config.ResolvePath()
	if err != nil {
		notif := state.Notification{
			Message:      fmt.Sprintf("Failed to resolve config path: %v", err),
			Level:        "error",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		s.Model.View.CurrentView = state.ViewBoard
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}

	cfg := config.Config{
		DefaultProjectID:        msg.ProjectID,
		DefaultOwner:            msg.Owner,
		SuppressHints:           msg.SuppressHints,
		DefaultItemLimit:        msg.ItemLimit,
		DefaultExcludeDone:      msg.ExcludeDone,
		DefaultIterationFilters: msg.IterationFilter,
	}
	saveErr := config.Save(configPath, cfg)
	if saveErr != nil {
		notif := state.Notification{
			Message:      fmt.Sprintf("Failed to save settings: %v", saveErr),
			Level:        "error",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		s.Model.View.CurrentView = state.ViewBoard
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}

	s.Model.SuppressHints = msg.SuppressHints
	s.Model.ItemLimit = msg.ItemLimit
	s.Model.ExcludeDone = msg.ExcludeDone
	s.Model.View.Filter.Iterations = msg.IterationFilter
	if !s.Model.SuppressHints {
		notif := state.Notification{
			Message:      "Settings saved successfully",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		s.Model.View.CurrentView = state.ViewBoard
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}

	s.Model.View.CurrentView = state.ViewBoard
	return s, nil
}

func SettingsCancel(s State, _ settings.CancelMsg) (State, tea.Cmd) {
	s.Model.View.CurrentView = state.ViewBoard
	return s, nil
}
