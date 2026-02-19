package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"project-hub/internal/config"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
)

func (a App) handleEnterDetailMode() (tea.Model, tea.Cmd) {
	if a.state.View.CurrentView == state.ViewBoard {
		a.syncFocusedItem()
	}

	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	focusedItem := a.state.Items[idx]

	if focusedItem.Repository != "" && focusedItem.Number > 0 {
		a.detailPanel = components.NewDetailPanelModel(focusedItem, a.state.Width, a.state.Height)
		fetchDescCmd := func() tea.Msg {
			body, err := a.github.FetchIssueDetail(context.Background(), focusedItem.Repository, focusedItem.Number)
			if err != nil {
				return NewErrMsg(err)
			}
			return DetailReadyMsg{Item: state.Item{
				Title:       focusedItem.Title,
				Description: body,
				Number:      focusedItem.Number,
				Repository:  focusedItem.Repository,
				Status:      focusedItem.Status,
				Assignees:   focusedItem.Assignees,
				Labels:      focusedItem.Labels,
				Priority:    focusedItem.Priority,
				Milestone:   focusedItem.Milestone,
				URL:         focusedItem.URL,
			}}
		}
		a.state.View.Mode = state.ModeDetail
		return a, tea.Cmd(fetchDescCmd)
	}

	a.detailPanel = components.NewDetailPanelModel(focusedItem, a.state.Width, a.state.Height)
	a.state.View.Mode = state.ModeDetail
	if !a.state.SuppressHints {
		detailNotif := state.Notification{
			Message:      "Detail mode: j/k to scroll, esc/q to close",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, detailNotif)
		return a, tea.Batch(a.detailPanel.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, detailNotif.DismissAfter))
	}
	return a, a.detailPanel.Init()
}

func (a App) handleEnterFieldToggleMode() (tea.Model, tea.Cmd) {
	a.state.View.Mode = state.ModeFieldToggle
	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Field toggle mode: m=milestone r=repository l=labels s=sub-issue p=parent (toggle on/off, esc to cancel)",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}
	return a, nil
}

func (a App) handleFieldToggle(key string) (tea.Model, tea.Cmd) {
	vis := &a.state.View.CardFieldVisibility
	switch key {
	case "m", "M":
		vis.ShowMilestone = !vis.ShowMilestone
	case "r", "R":
		vis.ShowRepository = !vis.ShowRepository
	case "l", "L":
		vis.ShowLabels = !vis.ShowLabels
	case "s", "S":
		vis.ShowSubIssueProgress = !vis.ShowSubIssueProgress
	case "p", "P":
		vis.ShowParentIssue = !vis.ShowParentIssue
	case "esc":
		a.state.View.Mode = state.ModeNormal
		return a, nil
	default:
		return a, nil
	}

	a.state.View.Mode = state.ModeNormal
	a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.Project.Fields, a.state.View.Filter, a.state.View.FocusedItemID, a.state.View.CardFieldVisibility)

	configPath, err := config.ResolvePath()
	if err == nil {
		existingCfg, loadErr := config.Load(configPath)
		if loadErr == nil {
			cfg := config.Config{
				DefaultProjectID:   existingCfg.DefaultProjectID,
				DefaultOwner:       existingCfg.DefaultOwner,
				SuppressHints:      existingCfg.SuppressHints,
				DefaultItemLimit:   existingCfg.DefaultItemLimit,
				DefaultExcludeDone: existingCfg.DefaultExcludeDone,
				CardFieldVisibility: config.CardFieldVisibility{
					ShowMilestone:        vis.ShowMilestone,
					ShowRepository:       vis.ShowRepository,
					ShowSubIssueProgress: vis.ShowSubIssueProgress,
					ShowParentIssue:      vis.ShowParentIssue,
					ShowLabels:           vis.ShowLabels,
				},
			}
			config.Save(configPath, cfg)
		}
	}

	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Card fields preference saved",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}
	return a, nil
}

func (a App) handleToggleGroupBy() (tea.Model, tea.Cmd) {
	current := strings.ToLower(strings.TrimSpace(a.state.View.TableGroupBy))
	switch current {
	case "":
		a.state.View.TableGroupBy = groupByStatus
	case groupByStatus:
		a.state.View.TableGroupBy = groupByAssignee
	case groupByAssignee:
		a.state.View.TableGroupBy = groupByIteration
	default:
		a.state.View.TableGroupBy = ""
	}

	if a.tableViewport != nil && a.state.View.TableGroupBy != "" {
		a.tableViewport.YOffset = 0
	}

	if !a.state.SuppressHints {
		label := a.state.View.TableGroupBy
		if label == "" {
			label = "none"
		}
		notif := state.Notification{
			Message:      fmt.Sprintf("Group: %s", label),
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}
	return a, nil
}
