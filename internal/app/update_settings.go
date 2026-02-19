package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"project-hub/internal/config"
	"project-hub/internal/state"
	"project-hub/internal/ui/settings"
)

func (a App) handleSettingsSave(msg settings.SaveMsg) (tea.Model, tea.Cmd) {
	configPath, err := config.ResolvePath()
	if err != nil {
		notif := state.Notification{
			Message:      fmt.Sprintf("Failed to resolve config path: %v", err),
			Level:        "error",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		a.state.View.CurrentView = state.ViewBoard
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	cfg := config.Config{
		DefaultProjectID:   msg.ProjectID,
		DefaultOwner:       msg.Owner,
		SuppressHints:      msg.SuppressHints,
		DefaultItemLimit:   msg.ItemLimit,
		DefaultExcludeDone: msg.ExcludeDone,
	}
	saveErr := config.Save(configPath, cfg)
	if saveErr != nil {
		notif := state.Notification{
			Message:      fmt.Sprintf("Failed to save settings: %v", saveErr),
			Level:        "error",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		a.state.View.CurrentView = state.ViewBoard
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	a.state.SuppressHints = msg.SuppressHints
	a.state.ItemLimit = msg.ItemLimit
	a.state.ExcludeDone = msg.ExcludeDone
	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Settings saved successfully",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		a.state.View.CurrentView = state.ViewBoard
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	a.state.View.CurrentView = state.ViewBoard
	return a, nil
}

func (a App) handleSettingsCancel(_ settings.CancelMsg) (tea.Model, tea.Cmd) {
	a.state.View.CurrentView = state.ViewBoard
	return a, nil
}
