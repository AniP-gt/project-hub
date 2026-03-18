package update

import (
	"context"
	"strings"
	"testing"

	"project-hub/internal/app/core"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

type mockClient struct{}

var mockCreateIssueLastTitle string
var mockCreateIssueLastBody string
var mockUpdateIssueBodyLastBody string
var mockAddIssueCommentLastBody string
var mockFetchIssueDetailResult state.Item

func (m *mockClient) FetchProject(ctx context.Context, projectID string, owner string, filter string, limit int) (state.Project, []state.Item, error) {
	return state.Project{}, nil, nil
}

func (m *mockClient) FetchItems(ctx context.Context, projectID string, owner string, filter string, limit int) ([]state.Item, error) {
	return nil, nil
}

func (m *mockClient) CreateIssue(ctx context.Context, projectID string, owner string, repo string, title string, body string) (state.Item, error) {
	mockCreateIssueLastTitle = title
	mockCreateIssueLastBody = body
	return state.Item{ID: "PVTI_new", Repository: repo, Title: title, Description: body, Type: "Issue"}, nil
}

func (m *mockClient) UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string) (state.Item, error) {
	return state.Item{}, nil
}

func (m *mockClient) UpdateField(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string, fieldName string) (state.Item, error) {
	return state.Item{}, nil
}

func (m *mockClient) UpdateLabels(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, labels []string) (state.Item, error) {
	return state.Item{}, nil
}

func (m *mockClient) UpdateMilestone(ctx context.Context, projectID string, owner string, itemID string, milestone string) (state.Item, error) {
	return state.Item{}, nil
}

func (m *mockClient) UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, userLogins []string) (state.Item, error) {
	return state.Item{}, nil
}

func (m *mockClient) UpdateItem(ctx context.Context, projectID string, owner string, item state.Item, title string, description string) (state.Item, error) {
	return state.Item{}, nil
}

func (m *mockClient) UpdateIssueBody(ctx context.Context, repo string, number int, body string) error {
	mockUpdateIssueBodyLastBody = body
	return nil
}

func (m *mockClient) AddIssueComment(ctx context.Context, repo string, number int, body string) error {
	mockAddIssueCommentLastBody = body
	return nil
}

func (m *mockClient) FetchIssueDetail(ctx context.Context, repo string, number int) (state.Item, error) {
	return mockFetchIssueDetailResult, nil
}

func TestEnterDetailEditMode_PrefillsDescription(t *testing.T) {
	items := []state.Item{{ID: "item1", Title: "Test", Description: "Old body", Repository: "owner/repo", Number: 12, Type: "Issue"}}
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeDetail, FocusedIndex: 0, FocusedItemID: "item1"}
	initialState := state.Model{Project: project, Items: items, View: viewContext, Width: 100, Height: 40}

	stateModel := NewState(initialState, &mockClient{}, 100)
	stateModel.DetailItem = items[0]

	updated, _ := EnterDetailEditMode(stateModel)

	if updated.Model.View.Mode != state.ModeDetailEdit {
		t.Fatalf("expected mode to be %q, got %q", state.ModeDetailEdit, updated.Model.View.Mode)
	}
	if updated.TextArea.Value() != "Old body" {
		t.Fatalf("expected textarea to contain existing description, got %q", updated.TextArea.Value())
	}
}

func TestSaveDetailEditReturnsSavedMessage(t *testing.T) {
	mockUpdateIssueBodyLastBody = ""
	items := []state.Item{{ID: "item1", Title: "Test", Description: "Old body", Repository: "owner/repo", Number: 12, Type: "Issue"}}
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeDetailEdit, FocusedIndex: 0, FocusedItemID: "item1"}
	initialState := state.Model{Project: project, Items: items, View: viewContext, Width: 100, Height: 40}

	stateModel := NewState(initialState, &mockClient{}, 100)
	stateModel.DetailItem = items[0]

	updated, cmd := SaveDetailEdit(stateModel, SaveDetailEditMsg{Description: "New body\nline two"})

	if updated.Model.View.Mode != state.ModeDetail {
		t.Fatalf("expected mode to return to %q, got %q", state.ModeDetail, updated.Model.View.Mode)
	}
	if cmd == nil {
		t.Fatalf("expected save command")
	}
	msg := cmd()
	saved, ok := msg.(DetailEditSavedMsg)
	if !ok {
		t.Fatalf("expected DetailEditSavedMsg, got %T", msg)
	}
	if saved.Description != "New body\nline two" {
		t.Fatalf("expected saved description to round-trip, got %q", saved.Description)
	}
	if mockUpdateIssueBodyLastBody != "New body\nline two" {
		t.Fatalf("expected client to receive updated body, got %q", mockUpdateIssueBodyLastBody)
	}
}

func TestSaveDetailCommentReturnsSuccessMessage(t *testing.T) {
	mockAddIssueCommentLastBody = ""
	mockFetchIssueDetailResult = state.Item{Description: "Issue body", Comments: []state.Comment{{Author: "alice", Body: "First line\nSecond line"}}}
	items := []state.Item{{ID: "item1", Title: "Test", Repository: "owner/repo", Number: 12, Type: "Issue"}}
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeDetailComment, FocusedIndex: 0, FocusedItemID: "item1"}
	initialState := state.Model{Project: project, Items: items, View: viewContext, Width: 100, Height: 40}

	stateModel := NewState(initialState, &mockClient{}, 100)
	stateModel.DetailItem = items[0]

	updated, cmd := SaveDetailComment(stateModel, SaveDetailCommentMsg{Body: "First line\nSecond line"})

	if updated.Model.View.Mode != state.ModeDetail {
		t.Fatalf("expected mode to return to %q, got %q", state.ModeDetail, updated.Model.View.Mode)
	}
	if cmd == nil {
		t.Fatalf("expected comment command")
	}
	msg := cmd()
	added, ok := msg.(DetailCommentAddedMsg)
	if !ok {
		t.Fatalf("expected DetailCommentAddedMsg, got %T", msg)
	}
	if len(added.Item.Comments) != 1 || added.Item.Comments[0].Author != "alice" {
		t.Fatalf("expected refreshed comment data, got %#v", added.Item.Comments)
	}
	if mockAddIssueCommentLastBody != "First line\nSecond line" {
		t.Fatalf("expected client to receive comment body, got %q", mockAddIssueCommentLastBody)
	}
}

func TestUpdateAppliesDetailReadyMsgWhileInDetailMode(t *testing.T) {
	items := []state.Item{{ID: "item1", Title: "Test", Description: "Old body", Repository: "owner/repo", Number: 46, Type: "Issue"}}
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeDetail, FocusedIndex: 0, FocusedItemID: "item1"}
	initialState := state.Model{Project: project, Items: items, View: viewContext, Width: 100, Height: 40, SuppressHints: true}

	stateModel := NewState(initialState, &mockClient{}, 100)
	stateModel.DetailItem = items[0]
	stateModel.DetailPanel = components.NewDetailPanelModel(items[0], 100, 40)

	updated, _ := Update(stateModel, core.DetailReadyMsg{Item: state.Item{
		ID:          "item1",
		Type:        "Issue",
		Title:       "Test",
		Description: "New body",
		Repository:  "owner/repo",
		Number:      46,
		Comments:    []state.Comment{{Author: "alice", Body: "Visible comment"}},
	}})

	if updated.DetailItem.Description != "New body" {
		t.Fatalf("expected detail description to update, got %q", updated.DetailItem.Description)
	}
	if len(updated.DetailItem.Comments) != 1 || updated.DetailItem.Comments[0].Body != "Visible comment" {
		t.Fatalf("expected comments to update, got %#v", updated.DetailItem.Comments)
	}
	if view := updated.DetailPanel.View(); !strings.Contains(view, "Visible comment") {
		t.Fatalf("expected detail panel view to contain refreshed comment, got %q", view)
	}
}

func TestEnterCreateIssueMode_UsesFocusedRepository(t *testing.T) {
	items := []state.Item{{ID: "item1", Title: "Test Item", Repository: "owner/repo", Status: "Todo", Position: 1}}
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeNormal, FocusedIndex: 0, FocusedItemID: "item1"}
	initialState := state.Model{Project: project, Items: items, View: viewContext, Width: 100, CreateIssueRepoMode: state.CreateIssueRepoModeAuto}

	stateModel := NewState(initialState, &mockClient{}, 100)
	updated, _ := EnterCreateIssueMode(stateModel, EnterCreateIssueModeMsg{})

	if updated.Model.View.Mode != state.ModeCreateIssueTitle {
		t.Fatalf("expected mode to be %q, got %q", state.ModeCreateIssueTitle, updated.Model.View.Mode)
	}
	if !strings.Contains(updated.TextInput.Placeholder, "owner/repo") {
		t.Fatalf("expected placeholder to mention repo, got %q", updated.TextInput.Placeholder)
	}
}

func TestEnterCreateIssueMode_UsesRepoStepWhenMultipleRepositoriesAndNoFocusedRepo(t *testing.T) {
	items := []state.Item{
		{ID: "item1", Title: "One", Repository: "owner/repo-a", Status: "Todo", Position: 1},
		{ID: "item2", Title: "Two", Repository: "owner/repo-b", Status: "Todo", Position: 2},
	}
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeNormal, FocusedIndex: -1}
	initialState := state.Model{Project: project, Items: items, View: viewContext, Width: 100, CreateIssueRepoMode: state.CreateIssueRepoModeAuto}

	stateModel := NewState(initialState, &mockClient{}, 100)
	updated, _ := EnterCreateIssueMode(stateModel, EnterCreateIssueModeMsg{})

	if updated.Model.View.Mode != state.ModeCreateIssueRepo {
		t.Fatalf("expected mode to be %q, got %q", state.ModeCreateIssueRepo, updated.Model.View.Mode)
	}
	if updated.TextInput.Placeholder != "Repository (owner/repo)..." {
		t.Fatalf("expected generic repository placeholder, got %q", updated.TextInput.Placeholder)
	}
}

func TestEnterCreateIssueMode_RequiredModeAlwaysStartsWithRepoStep(t *testing.T) {
	items := []state.Item{{ID: "item1", Title: "Test Item", Repository: "owner/repo", Status: "Todo", Position: 1}}
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeNormal, FocusedIndex: 0, FocusedItemID: "item1"}
	initialState := state.Model{Project: project, Items: items, View: viewContext, Width: 100, CreateIssueRepoMode: state.CreateIssueRepoModeRequired}

	stateModel := NewState(initialState, &mockClient{}, 100)
	updated, _ := EnterCreateIssueMode(stateModel, EnterCreateIssueModeMsg{})

	if updated.Model.View.Mode != state.ModeCreateIssueRepo {
		t.Fatalf("expected mode to be %q, got %q", state.ModeCreateIssueRepo, updated.Model.View.Mode)
	}
}

func TestSaveCreateIssueRepoAdvancesToTitle(t *testing.T) {
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeCreateIssueRepo}
	initialState := state.Model{Project: project, View: viewContext, Width: 100}

	stateModel := NewState(initialState, &mockClient{}, 100)
	updated, _ := SaveCreateIssue(stateModel, SaveCreateIssueMsg{Value: "owner/repo"})

	if updated.Model.View.Mode != state.ModeCreateIssueTitle {
		t.Fatalf("expected mode to be %q, got %q", state.ModeCreateIssueTitle, updated.Model.View.Mode)
	}
	if updated.CreateIssueRepo != "owner/repo" {
		t.Fatalf("expected repo to be stored, got %q", updated.CreateIssueRepo)
	}
}

func TestSaveCreateIssueTitleAdvancesToBody(t *testing.T) {
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeCreateIssueTitle}
	initialState := state.Model{Project: project, View: viewContext, Width: 100}

	stateModel := NewState(initialState, &mockClient{}, 100)
	stateModel.CreateIssueRepo = "owner/repo"
	updated, _ := SaveCreateIssue(stateModel, SaveCreateIssueMsg{Value: "New issue"})

	if updated.Model.View.Mode != state.ModeCreateIssueBody {
		t.Fatalf("expected mode to be %q, got %q", state.ModeCreateIssueBody, updated.Model.View.Mode)
	}
	if updated.CreateIssueTitle != "New issue" {
		t.Fatalf("expected title to be stored, got %q", updated.CreateIssueTitle)
	}
}

func TestSaveCreateIssueBodyReturnsCreatedMessage(t *testing.T) {
	mockCreateIssueLastTitle = ""
	mockCreateIssueLastBody = ""
	items := []state.Item{{ID: "item1", Title: "Test Item", Repository: "owner/repo", Status: "Todo", Position: 1}}
	project := state.Project{ID: "1", Owner: "owner"}
	viewContext := state.ViewContext{Mode: state.ModeCreateIssueBody, FocusedIndex: 0, FocusedItemID: "item1"}
	initialState := state.Model{Project: project, Items: items, View: viewContext, Width: 100}

	stateModel := NewState(initialState, &mockClient{}, 100)
	stateModel.CreateIssueRepo = "owner/repo"
	stateModel.CreateIssueTitle = "New issue"
	updated, cmd := SaveCreateIssue(stateModel, SaveCreateIssueMsg{Value: "Issue body"})

	if updated.Model.View.Mode != state.ModeNormal {
		t.Fatalf("expected mode to be %q after save, got %q", state.ModeNormal, updated.Model.View.Mode)
	}
	if cmd == nil {
		t.Fatalf("expected save command")
	}
	msg := cmd()
	created, ok := msg.(core.IssueCreatedMsg)
	if !ok {
		t.Fatalf("expected IssueCreatedMsg, got %T", msg)
	}
	if created.Item.Title != "New issue" {
		t.Fatalf("expected created issue title, got %q", created.Item.Title)
	}
	if created.Item.Description != "Issue body" {
		t.Fatalf("expected created issue body, got %q", created.Item.Description)
	}
	if mockCreateIssueLastTitle != "New issue" {
		t.Fatalf("expected client to receive title, got %q", mockCreateIssueLastTitle)
	}
	if mockCreateIssueLastBody != "Issue body" {
		t.Fatalf("expected client to receive body, got %q", mockCreateIssueLastBody)
	}
}

func TestEnterAssignMode_PrefillsWithExistingAssignee(t *testing.T) {
	items := []state.Item{
		{
			ID:        "item1",
			Title:     "Test Item",
			Assignees: []string{"alice"},
			Status:    "Todo",
			Position:  1,
		},
	}

	project := state.Project{
		ID:    "proj1",
		Owner: "test-owner",
	}

	viewContext := state.ViewContext{
		Mode:          state.ModeNormal,
		FocusedIndex:  0,
		FocusedItemID: "item1",
	}

	initialState := state.Model{
		Project: project,
		Items:   items,
		View:    viewContext,
	}

	stateModel := NewState(initialState, &mockClient{}, 100)

	updated, _ := EnterAssignMode(stateModel, EnterAssignModeMsg{})

	if updated.TextInput.Value() != "alice" {
		t.Errorf("expected text input to be 'alice', got %q", updated.TextInput.Value())
	}

	if updated.Model.View.Mode != state.ViewMode("assign") {
		t.Errorf("expected mode to be 'assign', got %q", updated.Model.View.Mode)
	}
}

func TestEnterAssignMode_PrefillsEmptyWhenNoAssignee(t *testing.T) {
	items := []state.Item{
		{
			ID:        "item1",
			Title:     "Test Item",
			Assignees: []string{},
			Status:    "Todo",
			Position:  1,
		},
	}

	project := state.Project{
		ID:    "proj1",
		Owner: "test-owner",
	}

	viewContext := state.ViewContext{
		Mode:          state.ModeNormal,
		FocusedIndex:  0,
		FocusedItemID: "item1",
	}

	initialState := state.Model{
		Project: project,
		Items:   items,
		View:    viewContext,
	}

	stateModel := NewState(initialState, &mockClient{}, 100)

	updated, _ := EnterAssignMode(stateModel, EnterAssignModeMsg{})

	if updated.TextInput.Value() != "" {
		t.Errorf("expected text input to be empty, got %q", updated.TextInput.Value())
	}

	if updated.Model.View.Mode != state.ViewMode("assign") {
		t.Errorf("expected mode to be 'assign', got %q", updated.Model.View.Mode)
	}
}

func TestEnterAssignMode_UsesFirstAssigneeWhenMultiple(t *testing.T) {
	items := []state.Item{
		{
			ID:        "item1",
			Title:     "Test Item",
			Assignees: []string{"alice", "bob", "charlie"},
			Status:    "Todo",
			Position:  1,
		},
	}

	project := state.Project{
		ID:    "proj1",
		Owner: "test-owner",
	}

	viewContext := state.ViewContext{
		Mode:          state.ModeNormal,
		FocusedIndex:  0,
		FocusedItemID: "item1",
	}

	initialState := state.Model{
		Project: project,
		Items:   items,
		View:    viewContext,
	}

	stateModel := NewState(initialState, &mockClient{}, 100)

	updated, _ := EnterAssignMode(stateModel, EnterAssignModeMsg{})

	if updated.TextInput.Value() != "alice" {
		t.Errorf("expected text input to be 'alice' (first assignee), got %q", updated.TextInput.Value())
	}

	if updated.Model.View.Mode != state.ViewMode("assign") {
		t.Errorf("expected mode to be 'assign', got %q", updated.Model.View.Mode)
	}
}

func TestCancelAssign_ExitsAssignMode(t *testing.T) {
	items := []state.Item{
		{
			ID:        "item1",
			Title:     "Test Item",
			Assignees: []string{"alice"},
			Status:    "Todo",
			Position:  1,
		},
	}

	project := state.Project{
		ID:    "proj1",
		Owner: "test-owner",
	}

	viewContext := state.ViewContext{
		Mode:          state.ViewMode("assign"),
		FocusedIndex:  0,
		FocusedItemID: "item1",
	}

	initialState := state.Model{
		Project: project,
		Items:   items,
		View:    viewContext,
	}

	stateModel := NewState(initialState, &mockClient{}, 100)

	updated, _ := CancelAssign(stateModel, CancelAssignMsg{})

	if updated.Model.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to be 'normal', got %q", updated.Model.View.Mode)
	}
}

func TestEnterAssignMode_InvalidFocusIndex(t *testing.T) {
	items := []state.Item{
		{
			ID:        "item1",
			Title:     "Test Item",
			Assignees: []string{"alice"},
			Status:    "Todo",
			Position:  1,
		},
	}

	project := state.Project{
		ID:    "proj1",
		Owner: "test-owner",
	}

	viewContext := state.ViewContext{
		Mode:          state.ModeNormal,
		FocusedIndex:  999,
		FocusedItemID: "item1",
	}

	initialState := state.Model{
		Project: project,
		Items:   items,
		View:    viewContext,
	}

	stateModel := NewState(initialState, &mockClient{}, 100)

	updated, _ := EnterAssignMode(stateModel, EnterAssignModeMsg{})

	if updated.Model.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to remain 'normal' on invalid index, got %q", updated.Model.View.Mode)
	}
}

func TestEnterStatusSelectMode_SuccessfulEntry(t *testing.T) {
	items := []state.Item{
		{
			ID:        "PVTI_lAHOBYxjlc4AP3t4zgGkVY4",
			Title:     "Test Item",
			Assignees: []string{},
			Status:    "Todo",
			Position:  1,
		},
	}

	project := state.Project{
		ID:    "proj1",
		Owner: "test-owner",
		Fields: []state.Field{
			{
				ID:   "status_field_id",
				Name: "Status",
				Options: []state.Option{
					{ID: "opt1", Name: "Todo"},
					{ID: "opt2", Name: "In Progress"},
					{ID: "opt3", Name: "Done"},
				},
			},
		},
	}

	viewContext := state.ViewContext{
		Mode:          state.ModeNormal,
		FocusedIndex:  0,
		FocusedItemID: "PVTI_lAHOBYxjlc4AP3t4zgGkVY4",
	}

	initialState := state.Model{
		Project: project,
		Items:   items,
		View:    viewContext,
	}

	stateModel := NewState(initialState, &mockClient{}, 100)

	updated, _ := EnterStatusSelectMode(stateModel, core.EnterStatusSelectModeMsg{})

	if updated.Model.View.Mode != state.ModeStatusSelect {
		t.Errorf("expected mode to be 'statusSelect', got %q", updated.Model.View.Mode)
	}
}

func TestEnterStatusSelectMode_StatusFieldNotFound(t *testing.T) {
	items := []state.Item{
		{
			ID:        "PVTI_lAHOBYxjlc4AP3t4zgGkVY4",
			Title:     "Test Item",
			Assignees: []string{},
			Status:    "Todo",
			Position:  1,
		},
	}

	project := state.Project{
		ID:    "proj1",
		Owner: "test-owner",
		Fields: []state.Field{
			{
				ID:   "other_field_id",
				Name: "Priority",
				Options: []state.Option{
					{ID: "opt1", Name: "Low"},
					{ID: "opt2", Name: "High"},
				},
			},
		},
	}

	viewContext := state.ViewContext{
		Mode:          state.ModeNormal,
		FocusedIndex:  0,
		FocusedItemID: "PVTI_lAHOBYxjlc4AP3t4zgGkVY4",
	}

	initialState := state.Model{
		Project: project,
		Items:   items,
		View:    viewContext,
	}

	stateModel := NewState(initialState, &mockClient{}, 100)

	updated, _ := EnterStatusSelectMode(stateModel, core.EnterStatusSelectModeMsg{})

	if updated.Model.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to remain 'normal' when status field not found, got %q", updated.Model.View.Mode)
	}

	if len(updated.Model.Notifications) != 1 {
		t.Errorf("expected 1 notification, got %d", len(updated.Model.Notifications))
	}

	if len(updated.Model.Notifications) > 0 {
		notif := updated.Model.Notifications[0]
		if notif.Level != "error" {
			t.Errorf("expected notification level to be 'error', got %q", notif.Level)
		}
		if !strings.Contains(notif.Message, "Status field not found") {
			t.Errorf("expected notification message to contain 'Status field not found', got %q", notif.Message)
		}
	}
}

func TestEnterStatusSelectMode_FocusedIndexOutOfRange(t *testing.T) {
	items := []state.Item{
		{
			ID:        "PVTI_lAHOBYxjlc4AP3t4zgGkVY4",
			Title:     "Test Item",
			Assignees: []string{},
			Status:    "Todo",
			Position:  1,
		},
	}

	project := state.Project{
		ID:    "proj1",
		Owner: "test-owner",
		Fields: []state.Field{
			{
				ID:   "status_field_id",
				Name: "Status",
				Options: []state.Option{
					{ID: "opt1", Name: "Todo"},
				},
			},
		},
	}

	viewContext := state.ViewContext{
		Mode:          state.ModeNormal,
		FocusedIndex:  999,
		FocusedItemID: "PVTI_lAHOBYxjlc4AP3t4zgGkVY4",
	}

	initialState := state.Model{
		Project: project,
		Items:   items,
		View:    viewContext,
	}

	stateModel := NewState(initialState, &mockClient{}, 100)

	updated, _ := EnterStatusSelectMode(stateModel, core.EnterStatusSelectModeMsg{})

	if updated.Model.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to remain 'normal' when focused index is out of range, got %q", updated.Model.View.Mode)
	}

	if len(updated.Model.Notifications) != 0 {
		t.Errorf("expected 0 notifications for out of range index, got %d", len(updated.Model.Notifications))
	}
}

func TestEnterStatusSelectMode_NegativeFocusedIndex(t *testing.T) {
	items := []state.Item{
		{
			ID:        "PVTI_lAHOBYxjlc4AP3t4zgGkVY4",
			Title:     "Test Item",
			Assignees: []string{},
			Status:    "Todo",
			Position:  1,
		},
	}

	project := state.Project{
		ID:    "proj1",
		Owner: "test-owner",
		Fields: []state.Field{
			{
				ID:   "status_field_id",
				Name: "Status",
				Options: []state.Option{
					{ID: "opt1", Name: "Todo"},
				},
			},
		},
	}

	viewContext := state.ViewContext{
		Mode:          state.ModeNormal,
		FocusedIndex:  -1,
		FocusedItemID: "PVTI_lAHOBYxjlc4AP3t4zgGkVY4",
	}

	initialState := state.Model{
		Project: project,
		Items:   items,
		View:    viewContext,
	}

	stateModel := NewState(initialState, &mockClient{}, 100)

	updated, _ := EnterStatusSelectMode(stateModel, core.EnterStatusSelectModeMsg{})

	if updated.Model.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to remain 'normal' when focused index is negative, got %q", updated.Model.View.Mode)
	}

	if len(updated.Model.Notifications) != 0 {
		t.Errorf("expected 0 notifications for negative index, got %d", len(updated.Model.Notifications))
	}
}
