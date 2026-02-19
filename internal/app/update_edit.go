package app

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// EnterEditModeMsg switches to edit mode for focused item.
type EnterEditModeMsg struct{}

// SaveEditMsg applies title/description changes to focused item.
type SaveEditMsg struct {
	Title       string
	Description string
}

// CancelEditMsg cancels edit mode.
type CancelEditMsg struct{}

// EnterAssignModeMsg switches to assign mode for focused item.
type EnterAssignModeMsg struct{}

// SaveAssignMsg applies assignee changes to focused item.
type SaveAssignMsg struct {
	Assignee string
}

// CancelAssignMsg cancels assign mode.
type CancelAssignMsg struct{}

// EnterLabelSelectModeMsg switches to label select mode for focused item.
type EnterLabelSelectModeMsg struct{}

// LabelSelectedMsg is sent when a label option is chosen by the user.
type LabelSelectedMsg struct {
	LabelID      string
	LabelName    string
	LabelFieldID string
	Canceled     bool
}

// SaveLabelMsg applies label changes to focused item.
type SaveLabelMsg struct {
	Labels []string
}

// CancelLabelMsg cancels label select mode.
type CancelLabelMsg struct{}

// EnterMilestoneSelectModeMsg switches to milestone select mode for focused item.
type EnterMilestoneSelectModeMsg struct{}

// MilestoneSelectedMsg is sent when a milestone option is chosen by the user.
type MilestoneSelectedMsg struct {
	MilestoneID      string
	MilestoneName    string
	MilestoneFieldID string
	Canceled         bool
}

// SaveMilestoneMsg applies milestone changes to focused item.
type SaveMilestoneMsg struct {
	Milestone string
}

// CancelMilestoneMsg cancels milestone select mode.
type CancelMilestoneMsg struct{}

// EnterPrioritySelectModeMsg switches to priority select mode for focused item.
type EnterPrioritySelectModeMsg struct{}

// PrioritySelectedMsg is sent when a priority option is chosen by the user.
type PrioritySelectedMsg struct {
	PriorityID      string
	PriorityName    string
	PriorityFieldID string
	Canceled        bool
}

// SavePriorityMsg applies priority changes to focused item.
type SavePriorityMsg struct {
	Priority string
}

// CancelPriorityMsg cancels priority select mode.
type CancelPriorityMsg struct{}

// EnterLabelsInputModeMsg switches to label input mode for focused item.
type EnterLabelsInputModeMsg struct{}

// SaveLabelsInputMsg applies label changes to focused item.
type SaveLabelsInputMsg struct {
	Labels string
}

// CancelLabelsInputMsg cancels label input mode.
type CancelLabelsInputMsg struct{}

// EnterMilestoneInputModeMsg switches to milestone input mode for focused item.
type EnterMilestoneInputModeMsg struct{}

// SaveMilestoneInputMsg applies milestone changes to focused item.
type SaveMilestoneInputMsg struct {
	Milestone string
}

// CancelMilestoneInputMsg cancels milestone input mode.
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
		return a, nil // Or handle error
	}
	item := a.state.Items[idx]

	// Initialize input with first assignee if present, otherwise empty string
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

	// Call the GitHub client to update the assignees
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
