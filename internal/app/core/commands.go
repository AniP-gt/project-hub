package core

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"project-hub/internal/github"
)

func FetchProjectCmd(client github.Client, projectID, owner string, itemLimit int) tea.Cmd {
	return func() tea.Msg {
		proj, items, err := client.FetchProject(context.Background(), projectID, owner, itemLimit)
		if err != nil {
			return NewErrMsg(err)
		}
		return FetchProjectMsg{Project: proj, Items: items}
	}
}

func DismissNotificationCmd(id int, duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return DismissNotificationMsg{ID: id}
	})
}
