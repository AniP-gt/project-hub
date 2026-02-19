package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"project-hub/internal/app/core"
	"project-hub/internal/app/update"
	"project-hub/internal/github"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
	"project-hub/internal/ui/settings"
)

type App struct {
	state          state.Model
	github         github.Client
	itemLimit      int
	boardModel     boardPkg.BoardModel
	textInput      textinput.Model
	statusSelector components.StatusSelectorModel
	fieldSelector  components.FieldSelectorModel
	settingsModel  settings.SettingsModel
	detailPanel    components.DetailPanelModel
	tableViewport  *viewport.Model
}

func New(initial state.Model, client github.Client, itemLimit int) App {
	updateState := update.NewState(initial, client, itemLimit)
	return fromUpdateState(updateState)
}

func (a App) Init() tea.Cmd {
	return tea.Batch(core.FetchProjectCmd(a.github, a.state.Project.ID, a.state.Project.Owner, a.itemLimit), textinput.Blink)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updatedState, cmd := update.Update(a.toUpdateState(), msg)
	a = a.applyUpdateState(updatedState)
	return a, cmd
}

func (a *App) LoadInitialState(ctx context.Context, projectID string, owner string) error {
	project, items, err := a.github.FetchProject(ctx, projectID, owner, a.itemLimit)
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

func (a App) toUpdateState() update.State {
	return update.State{
		Model:          a.state,
		Github:         a.github,
		ItemLimit:      a.itemLimit,
		BoardModel:     a.boardModel,
		TextInput:      a.textInput,
		StatusSelector: a.statusSelector,
		FieldSelector:  a.fieldSelector,
		SettingsModel:  a.settingsModel,
		DetailPanel:    a.detailPanel,
		TableViewport:  a.tableViewport,
	}
}

func (a App) applyUpdateState(s update.State) App {
	a.state = s.Model
	a.github = s.Github
	a.itemLimit = s.ItemLimit
	a.boardModel = s.BoardModel
	a.textInput = s.TextInput
	a.statusSelector = s.StatusSelector
	a.fieldSelector = s.FieldSelector
	a.settingsModel = s.SettingsModel
	a.detailPanel = s.DetailPanel
	a.tableViewport = s.TableViewport
	return a
}

func fromUpdateState(s update.State) App {
	return App{
		state:          s.Model,
		github:         s.Github,
		itemLimit:      s.ItemLimit,
		boardModel:     s.BoardModel,
		textInput:      s.TextInput,
		statusSelector: s.StatusSelector,
		fieldSelector:  s.FieldSelector,
		settingsModel:  s.SettingsModel,
		detailPanel:    s.DetailPanel,
		tableViewport:  s.TableViewport,
	}
}
