package update

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app/core"
	"project-hub/internal/config"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
)

func EnterDetailMode(s State) (State, tea.Cmd) {
	if s.Model.View.CurrentView == state.ViewBoard {
		s = SyncFocusedItem(s)
	}

	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	focusedItem := s.Model.Items[idx]

	if focusedItem.Repository != "" && focusedItem.Number > 0 {
		s.DetailPanel = components.NewDetailPanelModel(focusedItem, s.Model.Width, s.Model.Height)
		fetchDescCmd := func() tea.Msg {
			body, err := s.Github.FetchIssueDetail(context.Background(), focusedItem.Repository, focusedItem.Number)
			if err != nil {
				return core.NewErrMsg(err)
			}
			return core.DetailReadyMsg{Item: state.Item{
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
		s.Model.View.Mode = state.ModeDetail
		return s, tea.Cmd(fetchDescCmd)
	}

	s.DetailPanel = components.NewDetailPanelModel(focusedItem, s.Model.Width, s.Model.Height)
	s.Model.View.Mode = state.ModeDetail
	if !s.Model.SuppressHints {
		detailNotif := state.Notification{
			Message:      "Detail mode: j/k to scroll, esc/q to close",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, detailNotif)
		return s, tea.Batch(s.DetailPanel.Init(), core.DismissNotificationCmd(len(s.Model.Notifications)-1, detailNotif.DismissAfter))
	}
	return s, s.DetailPanel.Init()
}

func EnterFieldToggleMode(s State) (State, tea.Cmd) {
	s.Model.View.Mode = state.ModeFieldToggle
	if !s.Model.SuppressHints {
		notif := state.Notification{
			Message:      "Field toggle mode: m=milestone r=repository l=labels s=sub-issue p=parent (toggle on/off, esc to cancel)",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}
	return s, nil
}

func FieldToggle(s State, key string) (State, tea.Cmd) {
	vis := &s.Model.View.CardFieldVisibility
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
		s.Model.View.Mode = state.ModeNormal
		return s, nil
	default:
		return s, nil
	}

	s.Model.View.Mode = state.ModeNormal
	s.BoardModel = boardPkg.NewBoardModel(s.Model.Items, s.Model.Project.Fields, s.Model.View.Filter, s.Model.View.FocusedItemID, s.Model.View.CardFieldVisibility)

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

	if !s.Model.SuppressHints {
		notif := state.Notification{
			Message:      "Card fields preference saved",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}
	return s, nil
}

func ToggleGroupBy(s State) (State, tea.Cmd) {
	current := strings.ToLower(strings.TrimSpace(s.Model.View.TableGroupBy))
	switch current {
	case "":
		s.Model.View.TableGroupBy = core.GroupByStatus
	case core.GroupByStatus:
		s.Model.View.TableGroupBy = core.GroupByAssignee
	case core.GroupByAssignee:
		s.Model.View.TableGroupBy = core.GroupByIteration
	default:
		s.Model.View.TableGroupBy = ""
	}

	if s.TableViewport != nil && s.Model.View.TableGroupBy != "" {
		s.TableViewport.YOffset = 0
	}

	if !s.Model.SuppressHints {
		label := s.Model.View.TableGroupBy
		if label == "" {
			label = "none"
		}
		notif := state.Notification{
			Message:      fmt.Sprintf("Group: %s", label),
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}
	return s, nil
}

func SyncFocusedItem(s State) State {
	colIndex := s.BoardModel.FocusedColumnIndex
	cardIndex := s.BoardModel.FocusedCardIndex

	if colIndex < 0 || colIndex >= len(s.BoardModel.Columns) {
		return s
	}
	column := s.BoardModel.Columns[colIndex]

	if cardIndex < 0 || cardIndex >= len(column.Cards) {
		return s
	}
	focusedCard := column.Cards[cardIndex]

	for i, item := range s.Model.Items {
		if item.ID == focusedCard.ID {
			s.Model.View.FocusedIndex = i
			s.Model.View.FocusedItemID = item.ID
			return s
		}
	}
	return s
}
