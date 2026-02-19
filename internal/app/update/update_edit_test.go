package update

import (
	"context"
	"strings"
	"testing"

	"project-hub/internal/app/core"
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

func (m *mockClient) FetchIssueDetail(ctx context.Context, repo string, number int) (string, error) {
	return "", nil
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
