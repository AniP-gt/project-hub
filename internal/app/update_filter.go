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
	return a, nil
}

func (a App) handleApplyFilter(msg ApplyFilterMsg) (tea.Model, tea.Cmd) {
	fs := state.ParseFilter(msg.Query)
	a.state.View.Filter = fs
	a.boardModel = boardPkg.NewBoardModel(a.state.Items, fs, a.state.View.FocusedItemID)
	if fs.Query == "" {
		a.state.View.Mode = state.ModeNormal
	}
	return a, nil
}

func (a App) handleClearFilter(msg ClearFilterMsg) (tea.Model, tea.Cmd) {
	a.state.View.Filter = state.FilterState{}
	a.boardModel = boardPkg.NewBoardModel(a.state.Items, state.FilterState{}, a.state.View.FocusedItemID)
	if a.state.View.Mode == state.ModeFiltering {
		a.state.View.Mode = state.ModeNormal
	}
	return a, nil
}
