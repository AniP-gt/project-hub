package app

import (
	"context"
	"strings"
	"testing"

	"project-hub/internal/state"
)

type mockClient struct{}

func (m *mockClient) FetchProject(ctx context.Context, projectID string, owner string, limit int) (state.Project, []state.Item, error) {
	return state.Project{}, nil, nil
}

func (m *mockClient) FetchItems(ctx context.Context, projectID string, owner string, filter string, limit int) ([]state.Item, error) {
	return nil, nil
}

func (m *mockClient) UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string) (state.Item, error) {
	return state.Item{}, nil
}

func (m *mockClient) UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, assigneeFieldID string, userLogins []string) (state.Item, error) {
	return state.Item{}, nil
}

func (m *mockClient) UpdateItem(ctx context.Context, projectID string, owner string, item state.Item, title string, description string) (state.Item, error) {
	return state.Item{}, nil
}

func (m *mockClient) FetchRoadmap(ctx context.Context, projectID string, owner string) ([]state.Timeline, []state.Item, error) {
	return nil, nil, nil
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

	app := New(initialState, &mockClient{}, 100)

	updatedApp, _ := app.handleEnterAssignMode(EnterAssignModeMsg{})
	appModel := updatedApp.(App)

	if appModel.textInput.Value() != "alice" {
		t.Errorf("expected text input to be 'alice', got %q", appModel.textInput.Value())
	}

	if appModel.state.View.Mode != state.ViewMode("assign") {
		t.Errorf("expected mode to be 'assign', got %q", appModel.state.View.Mode)
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

	app := New(initialState, &mockClient{}, 100)

	updatedApp, _ := app.handleEnterAssignMode(EnterAssignModeMsg{})
	appModel := updatedApp.(App)

	if appModel.textInput.Value() != "" {
		t.Errorf("expected text input to be empty, got %q", appModel.textInput.Value())
	}

	if appModel.state.View.Mode != state.ViewMode("assign") {
		t.Errorf("expected mode to be 'assign', got %q", appModel.state.View.Mode)
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

	app := New(initialState, &mockClient{}, 100)

	updatedApp, _ := app.handleEnterAssignMode(EnterAssignModeMsg{})
	appModel := updatedApp.(App)

	if appModel.textInput.Value() != "alice" {
		t.Errorf("expected text input to be 'alice' (first assignee), got %q", appModel.textInput.Value())
	}

	if appModel.state.View.Mode != state.ViewMode("assign") {
		t.Errorf("expected mode to be 'assign', got %q", appModel.state.View.Mode)
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

	app := New(initialState, &mockClient{}, 100)

	updatedApp, _ := app.handleCancelAssign(CancelAssignMsg{})
	appModel := updatedApp.(App)

	if appModel.state.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to be 'normal', got %q", appModel.state.View.Mode)
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

	app := New(initialState, &mockClient{}, 100)

	updatedApp, _ := app.handleEnterAssignMode(EnterAssignModeMsg{})
	appModel := updatedApp.(App)

	if appModel.state.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to remain 'normal' on invalid index, got %q", appModel.state.View.Mode)
	}
}

// Test: Board view "w" key triggers status select mode entry
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

	app := New(initialState, &mockClient{}, 100)

	updatedApp, _ := app.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
	appModel := updatedApp.(App)

	if appModel.state.View.Mode != state.ModeStatusSelect {
		t.Errorf("expected mode to be 'statusSelect', got %q", appModel.state.View.Mode)
	}
}

// Test: Status field not found shows error notification
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

	app := New(initialState, &mockClient{}, 100)

	updatedApp, _ := app.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
	appModel := updatedApp.(App)

	if appModel.state.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to remain 'normal' when status field not found, got %q", appModel.state.View.Mode)
	}

	if len(appModel.state.Notifications) != 1 {
		t.Errorf("expected 1 notification, got %d", len(appModel.state.Notifications))
	}

	if len(appModel.state.Notifications) > 0 {
		notif := appModel.state.Notifications[0]
		if notif.Level != "error" {
			t.Errorf("expected notification level to be 'error', got %q", notif.Level)
		}
		if !strings.Contains(notif.Message, "Status field not found") {
			t.Errorf("expected notification message to contain 'Status field not found', got %q", notif.Message)
		}
	}
}

// Test: Focused index out of range guards against panic
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

	app := New(initialState, &mockClient{}, 100)

	updatedApp, _ := app.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
	appModel := updatedApp.(App)

	if appModel.state.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to remain 'normal' when focused index is out of range, got %q", appModel.state.View.Mode)
	}

	if len(appModel.state.Notifications) != 0 {
		t.Errorf("expected 0 notifications for out of range index, got %d", len(appModel.state.Notifications))
	}
}

// Test: Negative focused index is handled correctly
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

	app := New(initialState, &mockClient{}, 100)

	updatedApp, _ := app.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
	appModel := updatedApp.(App)

	if appModel.state.View.Mode != state.ModeNormal {
		t.Errorf("expected mode to remain 'normal' when focused index is negative, got %q", appModel.state.View.Mode)
	}

	if len(appModel.state.Notifications) != 0 {
		t.Errorf("expected 0 notifications for negative index, got %d", len(appModel.state.Notifications))
	}
}
