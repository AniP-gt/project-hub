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

type EnterCreateIssueModeMsg struct{}

type SaveCreateIssueMsg struct {
	Value string
}

type CancelCreateIssueMsg struct{}

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

type EnterDetailEditModeMsg struct{}

type SaveDetailEditMsg struct {
	Description string
}

type CancelDetailEditMsg struct{}

type EnterDetailCommentModeMsg struct{}

type SaveDetailCommentMsg struct {
	Body string
}

type CancelDetailCommentMsg struct{}

type DetailEditSavedMsg struct {
	Index       int
	ItemID      string
	Description string
}

type DetailCommentAddedMsg struct{}

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

func EnterCreateIssueMode(s State, _ EnterCreateIssueModeMsg) (State, tea.Cmd) {
	s.CreateIssueRepo = ""
	s.CreateIssueTitle = ""
	s.CreateIssueBody = ""

	if s.Model.CreateIssueRepoMode == state.CreateIssueRepoModeRequired {
		if err := prepareTextInput(&s, "", repoPlaceholder(s.Model)); err != nil {
			return s, nil
		}
		s.Model.View.Mode = state.ModeCreateIssueRepo
		return s, s.TextInput.Focus()
	}

	repo, err := resolveCreateIssueRepo(s.Model)
	if err == nil {
		s.CreateIssueRepo = repo
		return prepareCreateIssueTitleStep(s)
	}

	if err := prepareTextInput(&s, "", repoPlaceholder(s.Model)); err != nil {
		return s, nil
	}
	s.Model.View.Mode = state.ModeCreateIssueRepo
	return s, s.TextInput.Focus()
}

func SaveCreateIssue(s State, msg SaveCreateIssueMsg) (State, tea.Cmd) {
	value := strings.TrimSpace(msg.Value)

	switch s.Model.View.Mode {
	case state.ModeCreateIssueRepo:
		if value == "" {
			return s, func() tea.Msg { return core.NewErrMsg(fmt.Errorf("repository is required")) }
		}
		s.CreateIssueRepo = value
		return prepareCreateIssueTitleStep(s)
	case state.ModeCreateIssueTitle:
		if value == "" {
			return s, func() tea.Msg { return core.NewErrMsg(fmt.Errorf("issue title is required")) }
		}
		s.CreateIssueTitle = value
		return prepareCreateIssueBodyStep(s)
	case state.ModeCreateIssueBody:
		if value == "" {
			return s, func() tea.Msg { return core.NewErrMsg(fmt.Errorf("issue body is required")) }
		}
		s.CreateIssueBody = value
		repo := s.CreateIssueRepo
		title := s.CreateIssueTitle
		body := s.CreateIssueBody

		createCmd := func() tea.Msg {
			item, err := s.Github.CreateIssue(
				context.Background(),
				s.Model.Project.ID,
				s.Model.Project.Owner,
				repo,
				title,
				body,
			)
			if err != nil {
				return core.NewErrMsg(err)
			}
			return core.IssueCreatedMsg{Item: item}
		}

		s = resetCreateIssueState(s)
		s.Model.View.Mode = state.ModeNormal
		return s, tea.Batch(createCmd)
	default:
		return s, nil
	}
}

func CancelCreateIssue(s State, _ CancelCreateIssueMsg) (State, tea.Cmd) {
	if s.Model.View.Mode == state.ModeCreateIssueRepo || s.Model.View.Mode == state.ModeCreateIssueTitle || s.Model.View.Mode == state.ModeCreateIssueBody {
		s = resetCreateIssueState(s)
		s.Model.View.Mode = state.ModeNormal
	}
	return s, nil
}

func prepareCreateIssueTitleStep(s State) (State, tea.Cmd) {
	if err := prepareTextInput(&s, "", fmt.Sprintf("Title for %s...", s.CreateIssueRepo)); err != nil {
		return s, nil
	}
	s.Model.View.Mode = state.ModeCreateIssueTitle
	return s, s.TextInput.Focus()
}

func prepareCreateIssueBodyStep(s State) (State, tea.Cmd) {
	if err := prepareTextInput(&s, "", "Issue body..."); err != nil {
		return s, nil
	}
	s.Model.View.Mode = state.ModeCreateIssueBody
	return s, s.TextInput.Focus()
}

func prepareTextInput(s *State, value string, placeholder string) error {
	s.TextInput.Width = s.Model.Width - 10
	if s.TextInput.Width < 30 {
		s.TextInput.Width = 30
	}
	s.TextInput.SetValue(value)
	s.TextInput.Prompt = ""
	s.TextInput.Placeholder = placeholder
	return nil
}

func prepareDetailTextArea(s *State, value string, placeholder string) {
	width := s.Model.Width - 20
	if width < 40 {
		width = 40
	}
	height := s.Model.Height / 2
	if height < 8 {
		height = 8
	}
	if height > 20 {
		height = 20
	}

	s.TextArea.SetWidth(width)
	s.TextArea.SetHeight(height)
	s.TextArea.Prompt = ""
	s.TextArea.Placeholder = placeholder
	s.TextArea.SetValue(value)
}

func currentDetailItem(s State) (state.Item, int, bool) {
	item := s.DetailItem
	idx := s.Model.View.FocusedIndex
	if item.ID == "" && idx >= 0 && idx < len(s.Model.Items) {
		item = s.Model.Items[idx]
	}
	if item.Repository == "" || item.Number <= 0 {
		return state.Item{}, idx, false
	}
	return item, idx, true
}

func repoPlaceholder(model state.Model) string {
	_ = model
	return "Repository (owner/repo)..."
}

func resetCreateIssueState(s State) State {
	s.CreateIssueRepo = ""
	s.CreateIssueTitle = ""
	s.CreateIssueBody = ""
	return s
}

func resolveCreateIssueRepo(model state.Model) (string, error) {
	idx := model.View.FocusedIndex
	if idx >= 0 && idx < len(model.Items) {
		repo := strings.TrimSpace(model.Items[idx].Repository)
		if repo != "" {
			return repo, nil
		}
	}

	repoSet := map[string]struct{}{}
	for _, item := range model.Items {
		repo := strings.TrimSpace(item.Repository)
		if repo == "" {
			continue
		}
		repoSet[repo] = struct{}{}
	}

	if len(repoSet) == 1 {
		for repo := range repoSet {
			return repo, nil
		}
	}

	if len(repoSet) == 0 {
		return "", fmt.Errorf("cannot create issue: no repository found in current items")
	}

	return "", fmt.Errorf("repository selection required")
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

func EnterDetailEditMode(s State) (State, tea.Cmd) {
	item, _, ok := currentDetailItem(s)
	if !ok {
		return s, func() tea.Msg {
			return core.NewErrMsg(fmt.Errorf("detail editing is only available for issues with a repository and number"))
		}
	}

	prepareDetailTextArea(&s, item.Description, "Edit issue description...")
	s.Model.View.Mode = state.ModeDetailEdit
	return s, s.TextArea.Focus()
}

func SaveDetailEdit(s State, msg SaveDetailEditMsg) (State, tea.Cmd) {
	item, idx, ok := currentDetailItem(s)
	if !ok {
		return s, func() tea.Msg {
			return core.NewErrMsg(fmt.Errorf("cannot edit issue description: missing repository or issue number"))
		}
	}

	cmd := func() tea.Msg {
		err := s.Github.UpdateIssueBody(context.Background(), item.Repository, item.Number, msg.Description)
		if err != nil {
			return core.NewErrMsg(err)
		}
		return DetailEditSavedMsg{Index: idx, ItemID: item.ID, Description: msg.Description}
	}

	s.DetailItem.Description = msg.Description
	s.DetailPanel = components.NewDetailPanelModel(s.DetailItem, s.Model.Width, s.Model.Height)
	s.Model.View.Mode = state.ModeDetail
	return s, tea.Batch(cmd)
}

func CancelDetailEdit(s State) (State, tea.Cmd) {
	if s.Model.View.Mode == state.ModeDetailEdit {
		s.TextArea.Blur()
		s.Model.View.Mode = state.ModeDetail
	}
	return s, nil
}

func EnterDetailCommentMode(s State) (State, tea.Cmd) {
	_, _, ok := currentDetailItem(s)
	if !ok {
		return s, func() tea.Msg {
			return core.NewErrMsg(fmt.Errorf("commenting is only available for issues with a repository and number"))
		}
	}

	prepareDetailTextArea(&s, "", "Add comment...")
	s.Model.View.Mode = state.ModeDetailComment
	return s, s.TextArea.Focus()
}

func SaveDetailComment(s State, msg SaveDetailCommentMsg) (State, tea.Cmd) {
	item, _, ok := currentDetailItem(s)
	if !ok {
		return s, func() tea.Msg {
			return core.NewErrMsg(fmt.Errorf("cannot add issue comment: missing repository or issue number"))
		}
	}

	if strings.TrimSpace(msg.Body) == "" {
		return s, func() tea.Msg {
			return core.NewErrMsg(fmt.Errorf("comment body is required"))
		}
	}

	cmd := func() tea.Msg {
		err := s.Github.AddIssueComment(context.Background(), item.Repository, item.Number, msg.Body)
		if err != nil {
			return core.NewErrMsg(err)
		}
		return DetailCommentAddedMsg{}
	}

	s.Model.View.Mode = state.ModeDetail
	return s, tea.Batch(cmd)
}

func CancelDetailComment(s State) (State, tea.Cmd) {
	if s.Model.View.Mode == state.ModeDetailComment {
		s.TextArea.Blur()
		s.Model.View.Mode = state.ModeDetail
	}
	return s, nil
}

func ColumnEdit(s State, msg EnterEditModeMsg) (State, tea.Cmd) {
	colIdx := s.Model.View.FocusedColumnIndex
	idx := s.Model.View.FocusedIndex
	if idx < 0 || idx >= len(s.Model.Items) {
		return s, nil
	}

	colKey := colIdx
	if s.Model.View.CurrentView == state.ViewTable {
		visible := tableVisibleColumns(s.Model.View.CardFieldVisibility)
		if len(visible) == 0 {
			return EnterEditMode(s, msg)
		}
		if colIdx < 0 {
			colKey = visible[0]
		} else if colIdx >= len(visible) {
			colKey = visible[len(visible)-1]
		} else {
			colKey = visible[colIdx]
		}
	}

	switch colKey {
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
