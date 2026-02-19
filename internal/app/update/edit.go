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

type EnterEditModeMsg struct{}

type SaveEditMsg struct {
	Title       string
	Description string
}

type CancelEditMsg struct{}

type EnterAssignModeMsg struct{}

type SaveAssignMsg struct {
	Assignee string
}

type CancelAssignMsg struct{}

type EnterLabelSelectModeMsg struct{}

type LabelSelectedMsg struct {
	LabelID      string
	LabelName    string
	LabelFieldID string
	Canceled     bool
}

type SaveLabelMsg struct {
	Labels []string
}

type CancelLabelMsg struct{}

type EnterMilestoneSelectModeMsg struct{}

type MilestoneSelectedMsg struct {
	MilestoneID      string
	MilestoneName    string
	MilestoneFieldID string
	Canceled         bool
}

type SaveMilestoneMsg struct {
	Milestone string
}

type CancelMilestoneMsg struct{}

type EnterPrioritySelectModeMsg struct{}

type PrioritySelectedMsg struct {
	PriorityID      string
	PriorityName    string
	PriorityFieldID string
	Canceled        bool
}

type SavePriorityMsg struct {
	Priority string
}

type CancelPriorityMsg struct{}

type EnterLabelsInputModeMsg struct{}

type SaveLabelsInputMsg struct {
	Labels string
}

type CancelLabelsInputMsg struct{}

type EnterMilestoneInputModeMsg struct{}

type SaveMilestoneInputMsg struct {
	Milestone string
}

type CancelMilestoneInputMsg struct{}

func EnterEditMode(s State, _ EnterEditModeMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	item := s.Model.Items[idx]

	s.TextInput.Width = s.Model.Width - 10
	if s.TextInput.Width < 30 {
		s.TextInput.Width = 30
	}
	s.TextInput.SetValue(item.Title)
	s.TextInput.Prompt = ""
	s.TextInput.Placeholder = ""
	s.Model.View.Mode = "edit"
	return s, s.TextInput.Focus()
}

func CancelEdit(s State, _ CancelEditMsg) (State, tea.Cmd) {
	if s.Model.View.Mode == "edit" {
		s.Model.View.Mode = "normal"
	}
	return s, nil
}

func EnterAssignMode(s State, _ EnterAssignModeMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	item := s.Model.Items[idx]

	assignee := ""
	if len(item.Assignees) > 0 {
		assignee = item.Assignees[0]
	}

	s.TextInput.SetValue(assignee)
	s.TextInput.Prompt = ""
	s.TextInput.Placeholder = "Enter assignee..."
	s.Model.View.Mode = "assign"
	return s, s.TextInput.Focus()
}

func SaveAssign(s State, msg SaveAssignMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	item := s.Model.Items[idx]

	if item.Type != "Issue" && item.Type != "PullRequest" {
		return s, func() tea.Msg {
			return core.NewErrMsg(fmt.Errorf("cannot assign to item of type: %s (only Issues and PullRequests can be assigned)", item.Type))
		}
	}

	cmd := func() tea.Msg {
		userLogins := []string{}
		if msg.Assignee != "" {
			userLogins = append(userLogins, msg.Assignee)
		}
		updatedItem, err := s.Github.UpdateAssignees(
			context.Background(),
			core.ProjectMutationID(s.Model.Project),
			s.Model.Project.Owner,
			item.ID,
			item.Type,
			item.Repository,
			item.Number,
			userLogins,
		)
		if err != nil {
			return core.NewErrMsg(err)
		}
		return core.ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	s.Model.View.Mode = "normal"
	return s, tea.Batch(cmd)
}

func CancelAssign(s State, _ CancelAssignMsg) (State, tea.Cmd) {
	if s.Model.View.Mode == "assign" {
		s.Model.View.Mode = "normal"
	}
	return s, nil
}

func SaveEdit(s State, msg SaveEditMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	item := s.Model.Items[idx]

	updateCmd := func() tea.Msg {
		updatedItem, err := s.Github.UpdateItem(
			context.Background(),
			core.ProjectMutationID(s.Model.Project),
			s.Model.Project.Owner,
			item,
			msg.Title,
			msg.Description,
		)
		if err != nil {
			return core.NewErrMsg(err)
		}
		return core.ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	s.Model.View.Mode = state.ModeNormal
	return s, tea.Batch(updateCmd)
}

func ColumnEdit(s State, msg EnterEditModeMsg) (State, tea.Cmd) {
	colIdx := s.Model.View.FocusedColumnIndex
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}

	switch colIdx {
	case state.ColumnTitle:
		return EnterEditMode(s, msg)
	case state.ColumnStatus:
		return EnterStatusSelectMode(s, core.EnterStatusSelectModeMsg{})
	case state.ColumnAssignees:
		return EnterAssignMode(s, EnterAssignModeMsg{})
	case state.ColumnLabels:
		return EnterLabelsInputMode(s, EnterLabelsInputModeMsg{})
	case state.ColumnMilestone:
		return EnterMilestoneInputMode(s, EnterMilestoneInputModeMsg{})
	case state.ColumnPriority:
		return EnterPrioritySelectMode(s, EnterPrioritySelectModeMsg{})
	default:
		return EnterEditMode(s, msg)
	}
}

func EnterLabelSelectMode(s State, _ EnterLabelSelectModeMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	focusedItem := s.Model.Items[idx]

	var labelField state.Field
	found := false
	for _, field := range s.Model.Project.Fields {
		if field.Name == "Labels" {
			labelField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: "Labels field not found in project", Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}

	s.FieldSelector = components.NewFieldSelectorModel(focusedItem, labelField, s.Model.Width, s.Model.Height)
	s.Model.View.Mode = state.ModeLabelSelect

	if !s.Model.SuppressHints {
		notif := state.Notification{
			Message:      "Label select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, tea.Batch(s.FieldSelector.Init(), core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter))
	}
	return s, s.FieldSelector.Init()
}

func EnterMilestoneSelectMode(s State, _ EnterMilestoneSelectModeMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	focusedItem := s.Model.Items[idx]

	var milestoneField state.Field
	found := false
	for _, field := range s.Model.Project.Fields {
		if field.Name == "Milestone" {
			milestoneField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: "Milestone field not found in project", Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}

	s.FieldSelector = components.NewFieldSelectorModel(focusedItem, milestoneField, s.Model.Width, s.Model.Height)
	s.Model.View.Mode = state.ModeMilestoneSelect

	if !s.Model.SuppressHints {
		notif := state.Notification{
			Message:      "Milestone select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, tea.Batch(s.FieldSelector.Init(), core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter))
	}
	return s, s.FieldSelector.Init()
}

func EnterPrioritySelectMode(s State, _ EnterPrioritySelectModeMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	focusedItem := s.Model.Items[idx]

	var priorityField state.Field
	found := false
	for _, field := range s.Model.Project.Fields {
		if field.Name == "Priority" {
			priorityField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: "Priority field not found in project", Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter)
	}

	s.FieldSelector = components.NewFieldSelectorModel(focusedItem, priorityField, s.Model.Width, s.Model.Height)
	s.Model.View.Mode = state.ModePrioritySelect

	if !s.Model.SuppressHints {
		notif := state.Notification{
			Message:      "Priority select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		s.Model.Notifications = append(s.Model.Notifications, notif)
		return s, tea.Batch(s.FieldSelector.Init(), core.DismissNotificationCmd(len(s.Model.Notifications)-1, notif.DismissAfter))
	}
	return s, s.FieldSelector.Init()
}

func EnterLabelsInputMode(s State, _ EnterLabelsInputModeMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	item := s.Model.Items[idx]

	s.TextInput.Width = s.Model.Width - 10
	if s.TextInput.Width < 30 {
		s.TextInput.Width = 30
	}
	s.TextInput.SetValue(strings.Join(item.Labels, ","))
	s.TextInput.Prompt = ""
	s.TextInput.Placeholder = "Enter labels (comma separated)..."
	s.Model.View.Mode = "labelsInput"
	return s, s.TextInput.Focus()
}

func SaveLabelsInput(s State, msg SaveLabelsInputMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	item := s.Model.Items[idx]

	if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
		return s, func() tea.Msg {
			return core.NewErrMsg(fmt.Errorf("invalid item ID format: %s", item.ID))
		}
	}

	cmd := func() tea.Msg {
		labels := strings.Split(msg.Labels, ",")
		var trimmedLabels []string
		for _, l := range labels {
			l = strings.TrimSpace(l)
			if l != "" {
				trimmedLabels = append(trimmedLabels, l)
			}
		}
		updatedItem, err := s.Github.UpdateLabels(
			context.Background(),
			core.ProjectMutationID(s.Model.Project),
			s.Model.Project.Owner,
			item.ID,
			item.Type,
			item.Repository,
			item.Number,
			trimmedLabels,
		)
		if err != nil {
			return core.NewErrMsg(err)
		}
		return core.ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	s.Model.View.Mode = "normal"
	return s, tea.Batch(cmd)
}

func CancelLabelsInput(s State, _ CancelLabelsInputMsg) (State, tea.Cmd) {
	if s.Model.View.Mode == "labelsInput" {
		s.Model.View.Mode = "normal"
	}
	return s, nil
}

func EnterMilestoneInputMode(s State, _ EnterMilestoneInputModeMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	item := s.Model.Items[idx]

	s.TextInput.Width = s.Model.Width - 10
	if s.TextInput.Width < 30 {
		s.TextInput.Width = 30
	}
	s.TextInput.SetValue(item.Milestone)
	s.TextInput.Prompt = ""
	s.TextInput.Placeholder = "Enter milestone title..."
	s.Model.View.Mode = "milestoneInput"
	return s, s.TextInput.Focus()
}

func SaveMilestoneInput(s State, msg SaveMilestoneInputMsg) (State, tea.Cmd) {
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}
	item := s.Model.Items[idx]

	if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
		return s, func() tea.Msg {
			return core.NewErrMsg(fmt.Errorf("invalid item ID format: %s", item.ID))
		}
	}

	cmd := func() tea.Msg {
		updatedItem, err := s.Github.UpdateMilestone(
			context.Background(),
			core.ProjectMutationID(s.Model.Project),
			s.Model.Project.Owner,
			item.ID,
			msg.Milestone,
		)
		if err != nil {
			return core.NewErrMsg(err)
		}
		return core.ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	s.Model.View.Mode = "normal"
	return s, tea.Batch(cmd)
}

func CancelMilestoneInput(s State, _ CancelMilestoneInputMsg) (State, tea.Cmd) {
	if s.Model.View.Mode == "milestoneInput" {
		s.Model.View.Mode = "normal"
	}
	return s, nil
}
