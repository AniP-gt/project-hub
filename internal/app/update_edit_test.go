package app

import (
	"context"
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
