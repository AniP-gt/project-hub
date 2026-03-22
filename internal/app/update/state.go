package update

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"project-hub/internal/github"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
	"project-hub/internal/ui/settings"
)

type State struct {
	Model            state.Model
	Github           github.Client
	ItemLimit        int
	BoardModel       boardPkg.BoardModel
	TextInput        textinput.Model
	TextArea         textarea.Model
	StatusSelector   components.StatusSelectorModel
	FieldSelector    components.FieldSelectorModel
	SettingsModel    settings.SettingsModel
	DetailPanel      components.DetailPanelModel
	DetailItem       state.Item
	TableViewport    *viewport.Model
	LastKey          string
	LastKeyAt        time.Time
	CreateIssueRepo  string
	CreateIssueTitle string
	CreateIssueBody  string
	TextAreaVimMode  string
}

func NewState(initial state.Model, client github.Client, itemLimit int) State {
	boardModel := boardPkg.NewBoardModel(initial.Items, initial.Project.Fields, initial.View.Filter, initial.View.FocusedItemID, initial.View.CardFieldVisibility)
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.Width = 50
	ti.CharLimit = 500
	ta := textarea.New()
	ta.Placeholder = ""
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetWidth(60)
	ta.SetHeight(12)
	tableVP := viewport.New(0, 0)
	settingsModel := settings.New(initial.Project.ID, initial.Project.Owner, initial.SuppressHints, initial.ItemLimit, initial.ExcludeDone, string(initial.CreateIssueRepoMode), initial.View.Filter.Iterations)
	return State{
		Model:         initial,
		Github:        client,
		ItemLimit:     itemLimit,
		BoardModel:    boardModel,
		TextInput:     ti,
		TextArea:      ta,
		SettingsModel: settingsModel,
		TableViewport: &tableVP,
	}
}
