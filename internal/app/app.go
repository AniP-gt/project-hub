package app

import (
	"context"

	"github.com/charmbracelet/bubbles/textarea"
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
	state            state.Model
	github           github.Client
	itemLimit        int
	boardModel       boardPkg.BoardModel
	textInput        textinput.Model
	textArea         textarea.Model
	statusSelector   components.StatusSelectorModel
	fieldSelector    components.FieldSelectorModel
	settingsModel    settings.SettingsModel
	detailPanel      components.DetailPanelModel
	detailItem       state.Item
	tableViewport    *viewport.Model
	createIssueRepo  string
	createIssueTitle string
	createIssueBody  string
	textAreaVimMode  string
}

func New(initial state.Model, client github.Client, itemLimit int) App {
	updateState := update.NewState(initial, client, itemLimit)
	return fromUpdateState(updateState)
}

func (a App) Init() tea.Cmd {
	return tea.Batch(core.FetchProjectCmd(a.github, a.state.Project.ID, a.state.Project.Owner, a.itemLimit, a.state.View.Filter.Iterations), textinput.Blink)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updatedState, cmd := update.Update(a.toUpdateState(), msg)
	a = a.applyUpdateState(updatedState)
	return a, cmd
}

func (a *App) LoadInitialState(ctx context.Context, projectID string, owner string) error {
	project, items, err := a.github.FetchProject(ctx, projectID, owner, github.BuildIterationQuery(a.state.View.Filter.Iterations), a.itemLimit)
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
		Model:            a.state,
		Github:           a.github,
		ItemLimit:        a.itemLimit,
		BoardModel:       a.boardModel,
		TextInput:        a.textInput,
		TextArea:         a.textArea,
		StatusSelector:   a.statusSelector,
		FieldSelector:    a.fieldSelector,
		SettingsModel:    a.settingsModel,
		DetailPanel:      a.detailPanel,
		DetailItem:       a.detailItem,
		TableViewport:    a.tableViewport,
		CreateIssueRepo:  a.createIssueRepo,
		CreateIssueTitle: a.createIssueTitle,
		CreateIssueBody:  a.createIssueBody,
		TextAreaVimMode:  a.textAreaVimMode,
	}
}

func (a App) applyUpdateState(s update.State) App {
	a.state = s.Model
	a.github = s.Github
	a.itemLimit = s.ItemLimit
	a.boardModel = s.BoardModel
	a.textInput = s.TextInput
	a.textArea = s.TextArea
	a.statusSelector = s.StatusSelector
	a.fieldSelector = s.FieldSelector
	a.settingsModel = s.SettingsModel
	a.detailPanel = s.DetailPanel
	a.detailItem = s.DetailItem
	a.tableViewport = s.TableViewport
	a.createIssueRepo = s.CreateIssueRepo
	a.createIssueTitle = s.CreateIssueTitle
	a.createIssueBody = s.CreateIssueBody
	a.textAreaVimMode = s.TextAreaVimMode
	return a
}

func fromUpdateState(s update.State) App {
	return App{
		state:            s.Model,
		github:           s.Github,
		itemLimit:        s.ItemLimit,
		boardModel:       s.BoardModel,
		textInput:        s.TextInput,
		textArea:         s.TextArea,
		statusSelector:   s.StatusSelector,
		fieldSelector:    s.FieldSelector,
		settingsModel:    s.SettingsModel,
		detailPanel:      s.DetailPanel,
		detailItem:       s.DetailItem,
		tableViewport:    s.TableViewport,
		createIssueRepo:  s.CreateIssueRepo,
		createIssueTitle: s.CreateIssueTitle,
		createIssueBody:  s.CreateIssueBody,
		textAreaVimMode:  s.TextAreaVimMode,
	}
}
