package app

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"project-hub/internal/github"
	"project-hub/internal/state"
	"project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
	"project-hub/internal/ui/roadmap"
	"project-hub/internal/ui/table"
)

// App implements the Bubbletea Model interface and holds root state.
type App struct {
	state state.Model
	gh    github.Client
}

// New creates an App with an optional preloaded state.
func New(initial state.Model, client github.Client) App {
	return App{state: initial, gh: client}
}

// Init loads initial project data (placeholder until gh wiring is added).
func (a App) Init() tea.Cmd {
	return func() tea.Msg {
		// TODO: dispatch async load of project/items via gh client
		return nil
	}
}

// Update routes incoming messages to state transitions.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.state.Width = m.Width
		a.state.Height = m.Height
		return a, nil
	case tea.KeyMsg:
		return a.handleKey(m)
	case SwitchViewMsg:
		return a.handleSwitchView(m)
	case MoveFocusMsg:
		return a.handleMoveFocus(m)
	case StatusMoveMsg:
		return a.handleStatusMove(m)
	case EnterFilterModeMsg:
		return a.handleEnterFilterMode(m)
	case ApplyFilterMsg:
		return a.handleApplyFilter(m)
	case ClearFilterMsg:
		return a.handleClearFilter(m)
	case EnterEditModeMsg:
		return a.handleEnterEditMode(m)
	case SaveEditMsg:
		return a.handleSaveEdit(m)
	case CancelEditMsg:
		return a.handleCancelEdit(m)
	case AssignMsg:
		return a.handleAssign(m)
	default:
		return a, nil
	}
}

func (a App) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "1", "b":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewBoard})
	case "2", "t":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewTable})
	case "3", "r":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewRoadmap})
	case "R", "ctrl+r":
		return a.handleReload()
	case "j":
		return a.handleMoveFocus(MoveFocusMsg{Delta: 1})
	case "k":
		return a.handleMoveFocus(MoveFocusMsg{Delta: -1})
	case "h":
		return a.handleStatusMove(StatusMoveMsg{Direction: github.DirectionLeft})
	case "l":
		return a.handleStatusMove(StatusMoveMsg{Direction: github.DirectionRight})
	case "/":
		return a.handleEnterFilterMode(EnterFilterModeMsg{})
	case "esc":
		return a.handleClearFilter(ClearFilterMsg{})
	case "i", "enter":
		return a.handleEnterEditMode(EnterEditModeMsg{})
	default:
		return a, nil
	}
}

func (a App) handleReload() (tea.Model, tea.Cmd) {
	ctx := context.Background()
	projID := a.state.Project.ID
	owner := a.state.Project.Owner
	proj, items, err := a.gh.FetchProject(ctx, projID, owner)
	if err != nil {
		a.state.Notifications = append(a.state.Notifications, state.Notification{Message: fmt.Sprintf("Reload failed: %v", err), Level: "error", At: time.Now()})
		return a, nil
	}
	if proj.Name != "" {
		a.state.Project.Name = proj.Name
	}
	if proj.Owner != "" {
		a.state.Project.Owner = proj.Owner
	}
	if len(items) > 0 {
		a.state.Items = items
		a.state.View.FocusedIndex = 0
		a.state.View.FocusedItemID = items[0].ID
		a.state.Notifications = append(a.state.Notifications, state.Notification{Message: fmt.Sprintf("Reloaded %d items", len(items)), Level: "info", At: time.Now()})
	} else {
		a.state.Items = items
		a.state.View.FocusedIndex = -1
		a.state.View.FocusedItemID = ""
		a.state.Notifications = append(a.state.Notifications, state.Notification{Message: "Reloaded: 0 items", Level: "warn", At: time.Now()})
	}
	return a, nil
}

// View renders the current UI.
func (a App) View() string {
	width := a.state.Width
	if width == 0 {
		width = 100
	}
	header := components.RenderHeader(a.state.Project, a.state.View, width)
	items := applyFilter(a.state.Items, a.state.View.Filter)

	body := ""
	switch a.state.View.CurrentView {
	case state.ViewTable:
		body = table.Render(items, a.state.View.FocusedItemID)
	case state.ViewRoadmap:
		body = roadmap.Render(a.state.Project.Iterations, items, a.state.View.FocusedItemID)
	default:
		body = board.Render(items, a.state.View.FocusedItemID)
	}

	frameWidth := width
	if frameWidth <= 0 {
		frameWidth = 100
	}
	innerWidth := frameWidth - components.FrameStyle.GetHorizontalFrameSize()
	if innerWidth < 40 {
		innerWidth = 40
	}
	bodyRendered := lipgloss.NewStyle().Width(innerWidth).Render(body)
	framed := components.FrameStyle.Width(frameWidth).Render(bodyRendered)

	footer := components.RenderFooter(string(a.state.View.Mode), string(a.state.View.CurrentView), width)
	notif := components.RenderNotifications(a.state.Notifications)
	return fmt.Sprintf("%s\n%s\n%s\n%s", header, framed, footer, notif)
}

// LoadInitialState is a helper to fetch project data using the gh client.
func (a *App) LoadInitialState(ctx context.Context, projectID string, owner string) error {
	project, items, err := a.gh.FetchProject(ctx, projectID, owner)
	if err != nil {
		return err
	}
	a.state.Project = project
	a.state.Items = items
	if a.state.View.CurrentView == "" {
		a.state.View.CurrentView = state.ViewBoard
		a.state.View.Mode = state.ModeNormal
	}
	return nil
}

func applyFilter(items []state.Item, fs state.FilterState) []state.Item {
	if fs.Query == "" {
		return items
	}
	var out []state.Item
	for _, it := range items {
		if len(fs.Labels) > 0 && !containsAny(it.Labels, fs.Labels) {
			continue
		}
		if len(fs.Assignees) > 0 && !containsAny(it.Assignees, fs.Assignees) {
			continue
		}
		if len(fs.Statuses) > 0 && !containsAny([]string{it.Status}, fs.Statuses) {
			continue
		}
		out = append(out, it)
	}
	return out
}

func containsAny(haystack []string, needles []string) bool {
	for _, n := range needles {
		for _, h := range haystack {
			if h == n {
				return true
			}
		}
	}
	return false
}
