package app

import (
	"context"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"project-hub/internal/config"
	"project-hub/internal/state"
	"project-hub/internal/ui/settings"
)

type mockClient struct{}

func (m *mockClient) FetchProject(ctx context.Context, projectID string, owner string, filter string, limit int) (state.Project, []state.Item, error) {
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

func TestSwitchToSettingsView(t *testing.T) {
	initialState := state.Model{
		Project: state.Project{ID: "1", Owner: "User"},
		Items:   []state.Item{},
		View: state.ViewContext{
			CurrentView: state.ViewBoard,
			Mode:        state.ModeNormal,
		},
	}

	a := New(initialState, &mockClient{}, 100)
	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	appModel := updated.(App)

	if appModel.state.View.CurrentView != state.ViewSettings {
		t.Fatalf("expected current view %q, got %q", state.ViewSettings, appModel.state.View.CurrentView)
	}
}

func TestSettingsSaveWritesConfigAndReturnsToBoard(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")

	initialState := state.Model{
		Project: state.Project{ID: "1", Owner: "old-owner"},
		Items:   []state.Item{},
		View: state.ViewContext{
			CurrentView: state.ViewSettings,
			Mode:        state.ModeNormal,
		},
	}

	a := New(initialState, &mockClient{}, 100)
	updated, _ := a.Update(settings.SaveMsg{ProjectID: "1", Owner: "User"})
	appModel := updated.(App)

	if appModel.state.View.CurrentView != state.ViewBoard {
		t.Fatalf("expected view to return to %q, got %q", state.ViewBoard, appModel.state.View.CurrentView)
	}

	if len(appModel.state.Notifications) == 0 {
		t.Fatalf("expected a success notification")
	}
	if !strings.Contains(strings.ToLower(appModel.state.Notifications[len(appModel.state.Notifications)-1].Message), "saved") {
		t.Fatalf("expected save notification, got %q", appModel.state.Notifications[len(appModel.state.Notifications)-1].Message)
	}

	configPath, err := config.ResolvePath()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config file to be created: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, `"defaultProjectID": "1"`) {
		t.Fatalf("expected project id in config, got: %s", text)
	}
	if !strings.Contains(text, `"defaultOwner": "User"`) {
		t.Fatalf("expected owner in config, got: %s", text)
	}
}

func TestSettingsCancelReturnsToBoardWithoutWriting(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")

	initialState := state.Model{
		Project: state.Project{ID: "1", Owner: "owner"},
		Items:   []state.Item{},
		View: state.ViewContext{
			CurrentView: state.ViewSettings,
			Mode:        state.ModeNormal,
		},
	}

	a := New(initialState, &mockClient{}, 100)
	updated, _ := a.Update(settings.CancelMsg{})
	appModel := updated.(App)

	if appModel.state.View.CurrentView != state.ViewBoard {
		t.Fatalf("expected view to return to %q, got %q", state.ViewBoard, appModel.state.View.CurrentView)
	}

	configPath, err := config.ResolvePath()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("expected no config file on cancel, got err=%v", err)
	}
}
