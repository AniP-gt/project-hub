package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/state"
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
	if fs.Query == "" {
		a.state.View.Mode = state.ModeNormal
	}
	return a, nil
}

func (a App) handleClearFilter(msg ClearFilterMsg) (tea.Model, tea.Cmd) {
	a.state.View.Filter = state.FilterState{}
	if a.state.View.Mode == state.ModeFiltering {
		a.state.View.Mode = state.ModeNormal
	}
	return a, nil
}
