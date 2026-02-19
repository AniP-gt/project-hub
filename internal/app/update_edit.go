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

func (a App) handleEnterEditMode(msg EnterEditModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	a.textInput.Width = a.state.Width - 10
	if a.textInput.Width < 30 {
		a.textInput.Width = 30
	}
	a.textInput.SetValue(item.Title)
	a.textInput.Prompt = ""
	a.textInput.Placeholder = ""
	a.state.View.Mode = "edit"
	return a, a.textInput.Focus()
}

func (a App) handleCancelEdit(msg CancelEditMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "edit" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}

func (a App) handleEnterAssignMode(msg EnterAssignModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	assignee := ""
	if len(item.Assignees) > 0 {
		assignee = item.Assignees[0]
	}

	a.textInput.SetValue(assignee)
	a.textInput.Prompt = ""
	a.textInput.Placeholder = "Enter assignee..."
	a.state.View.Mode = "assign"
	return a, a.textInput.Focus()
}

func (a App) handleSaveAssign(msg SaveAssignMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	if item.Type != "Issue" && item.Type != "PullRequest" {
		return a, func() tea.Msg {
			return NewErrMsg(fmt.Errorf("cannot assign to item of type: %s (only Issues and PullRequests can be assigned)", item.Type))
		}
	}

	cmd := func() tea.Msg {
		userLogins := []string{}
		if msg.Assignee != "" {
			userLogins = append(userLogins, msg.Assignee)
		}
		updatedItem, err := a.github.UpdateAssignees(
			context.Background(),
			projectMutationID(a.state.Project),
			a.state.Project.Owner,
			item.ID,
			item.Type,
			item.Repository,
			item.Number,
			userLogins,
		)
		if err != nil {
			return NewErrMsg(err)
		}
		return ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	a.state.View.Mode = "normal"
	return a, tea.Batch(cmd)
}

func (a App) handleCancelAssign(msg CancelAssignMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "assign" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}

func (a App) handleSaveEdit(msg SaveEditMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	updateCmd := func() tea.Msg {
		updatedItem, err := a.github.UpdateItem(
			context.Background(),
			projectMutationID(a.state.Project),
			a.state.Project.Owner,
			item,
			msg.Title,
			msg.Description,
		)
		if err != nil {
			return NewErrMsg(err)
		}
		return ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	a.state.View.Mode = state.ModeNormal
	return a, tea.Batch(updateCmd)
}

func (a App) handleColumnEdit(msg EnterEditModeMsg) (tea.Model, tea.Cmd) {
	colIdx := a.state.View.FocusedColumnIndex
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}

	switch colIdx {
	case state.ColumnTitle:
		return a.handleEnterEditMode(msg)
	case state.ColumnStatus:
		return a.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
	case state.ColumnAssignees:
		return a.handleEnterAssignMode(EnterAssignModeMsg{})
	case state.ColumnLabels:
		return a.handleEnterLabelsInputMode(EnterLabelsInputModeMsg{})
	case state.ColumnMilestone:
		return a.handleEnterMilestoneInputMode(EnterMilestoneInputModeMsg{})
	case state.ColumnPriority:
		return a.handleEnterPrioritySelectMode(EnterPrioritySelectModeMsg{})
	default:
		return a.handleEnterEditMode(msg)
	}
}

func (a App) handleEnterLabelSelectMode(msg EnterLabelSelectModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	focusedItem := a.state.Items[idx]

	var labelField state.Field
	found := false
	for _, field := range a.state.Project.Fields {
		if field.Name == "Labels" {
			labelField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: "Labels field not found in project", Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	a.fieldSelector = components.NewFieldSelectorModel(focusedItem, labelField, a.state.Width, a.state.Height)
	a.state.View.Mode = state.ModeLabelSelect

	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Label select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, tea.Batch(a.fieldSelector.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	}
	return a, a.fieldSelector.Init()
}

func (a App) handleEnterMilestoneSelectMode(msg EnterMilestoneSelectModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	focusedItem := a.state.Items[idx]

	var milestoneField state.Field
	found := false
	for _, field := range a.state.Project.Fields {
		if field.Name == "Milestone" {
			milestoneField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: "Milestone field not found in project", Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	a.fieldSelector = components.NewFieldSelectorModel(focusedItem, milestoneField, a.state.Width, a.state.Height)
	a.state.View.Mode = state.ModeMilestoneSelect

	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Milestone select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, tea.Batch(a.fieldSelector.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	}
	return a, a.fieldSelector.Init()
}

func (a App) handleEnterPrioritySelectMode(msg EnterPrioritySelectModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	focusedItem := a.state.Items[idx]

	var priorityField state.Field
	found := false
	for _, field := range a.state.Project.Fields {
		if field.Name == "Priority" {
			priorityField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: "Priority field not found in project", Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	a.fieldSelector = components.NewFieldSelectorModel(focusedItem, priorityField, a.state.Width, a.state.Height)
	a.state.View.Mode = state.ModePrioritySelect

	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Priority select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, tea.Batch(a.fieldSelector.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	}
	return a, a.fieldSelector.Init()
}

func (a App) handleEnterLabelsInputMode(msg EnterLabelsInputModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	a.textInput.Width = a.state.Width - 10
	if a.textInput.Width < 30 {
		a.textInput.Width = 30
	}
	a.textInput.SetValue(strings.Join(item.Labels, ","))
	a.textInput.Prompt = ""
	a.textInput.Placeholder = "Enter labels (comma separated)..."
	a.state.View.Mode = "labelsInput"
	return a, a.textInput.Focus()
}

func (a App) handleSaveLabelsInput(msg SaveLabelsInputMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
		return a, func() tea.Msg {
			return NewErrMsg(fmt.Errorf("invalid item ID format: %s", item.ID))
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
		updatedItem, err := a.github.UpdateLabels(
			context.Background(),
			projectMutationID(a.state.Project),
			a.state.Project.Owner,
			item.ID,
			item.Type,
			item.Repository,
			item.Number,
			trimmedLabels,
		)
		if err != nil {
			return NewErrMsg(err)
		}
		return ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	a.state.View.Mode = "normal"
	return a, tea.Batch(cmd)
}

func (a App) handleCancelLabelsInput(msg CancelLabelsInputMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "labelsInput" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}

func (a App) handleEnterMilestoneInputMode(msg EnterMilestoneInputModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	a.textInput.Width = a.state.Width - 10
	if a.textInput.Width < 30 {
		a.textInput.Width = 30
	}
	a.textInput.SetValue(item.Milestone)
	a.textInput.Prompt = ""
	a.textInput.Placeholder = "Enter milestone title..."
	a.state.View.Mode = "milestoneInput"
	return a, a.textInput.Focus()
}

func (a App) handleSaveMilestoneInput(msg SaveMilestoneInputMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
		return a, func() tea.Msg {
			return NewErrMsg(fmt.Errorf("invalid item ID format: %s", item.ID))
		}
	}

	cmd := func() tea.Msg {
		updatedItem, err := a.github.UpdateMilestone(
			context.Background(),
			projectMutationID(a.state.Project),
			a.state.Project.Owner,
			item.ID,
			msg.Milestone,
		)
		if err != nil {
			return NewErrMsg(err)
		}
		return ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	a.state.View.Mode = "normal"
	return a, tea.Batch(cmd)
}

func (a App) handleCancelMilestoneInput(msg CancelMilestoneInputMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "milestoneInput" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}
