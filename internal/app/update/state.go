package update

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"project-hub/internal/github"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
	"project-hub/internal/ui/settings"
)

type State struct {
	Model          state.Model
	Github         github.Client
	ItemLimit      int
	BoardModel     boardPkg.BoardModel
	TextInput      textinput.Model
	StatusSelector components.StatusSelectorModel
	FieldSelector  components.FieldSelectorModel
	SettingsModel  settings.SettingsModel
	DetailPanel    components.DetailPanelModel
	TableViewport  *viewport.Model
	LastKey        string
	LastKeyAt      time.Time
}

func NewState(initial state.Model, client github.Client, itemLimit int) State {
	boardModel := boardPkg.NewBoardModel(initial.Items, initial.Project.Fields, initial.View.Filter, initial.View.FocusedItemID, initial.View.CardFieldVisibility)
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.Width = 50
	ti.CharLimit = 500
	tableVP := viewport.New(0, 0)
	settingsModel := settings.New(initial.Project.ID, initial.Project.Owner, initial.SuppressHints, initial.ItemLimit, initial.ExcludeDone)
	return State{
		Model:         initial,
		Github:        client,
		ItemLimit:     itemLimit,
		BoardModel:    boardModel,
		TextInput:     ti,
		SettingsModel: settingsModel,
		TableViewport: &tableVP,
	}
}
