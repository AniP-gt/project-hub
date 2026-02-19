package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
)

// EnterFilterModeMsg switches mode to filtering.
type EnterFilterModeMsg struct{}

// ApplyFilterMsg applies a new filter query.
type ApplyFilterMsg struct {
	Query string
}

// ClearFilterMsg clears current filters.
type ClearFilterMsg struct{}

func (a App) handleEnterFilterMode(msg EnterFilterModeMsg) (tea.Model, tea.Cmd) {
	a.state.View.Mode = state.ModeFiltering
	if a.state.Width > 0 {
		a.textInput.Width = a.state.Width - 10
		if a.textInput.Width < 30 {
			a.textInput.Width = 30
		}
	}
	a.textInput.Prompt = "FILTER MODE "
	a.textInput.Placeholder = "Enter filters..."
	a.textInput.SetValue(a.state.View.Filter.Raw)
	return a, a.textInput.Focus()
}

func (a App) handleApplyFilter(msg ApplyFilterMsg) (tea.Model, tea.Cmd) {
	fs := state.ParseFilter(msg.Query)
	a.state.View.Filter = fs
	if fs.GroupBy != "" {
		a.state.View.TableGroupBy = fs.GroupBy
	}
	a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.Project.Fields, fs, a.state.View.FocusedItemID, a.state.View.CardFieldVisibility)
	a.state.View.Mode = state.ModeNormal
	a.textInput.Prompt = ""
	return a, nil
}

func (a App) handleClearFilter(msg ClearFilterMsg) (tea.Model, tea.Cmd) {
	a.state.View.Filter = state.FilterState{}
	a.state.View.TableGroupBy = ""
	a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.Project.Fields, state.FilterState{}, a.state.View.FocusedItemID, a.state.View.CardFieldVisibility)
	if a.state.View.Mode == state.ModeFiltering {
		a.state.View.Mode = state.ModeNormal
	}
	a.textInput.SetValue("")
	a.textInput.Prompt = ""
	return a, nil
}
