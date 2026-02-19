package update

import (
	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
)

type EnterFilterModeMsg struct{}

type ApplyFilterMsg struct {
	Query string
}

type ClearFilterMsg struct{}

func EnterFilterMode(s State, _ EnterFilterModeMsg) (State, tea.Cmd) {
	s.Model.View.Mode = state.ModeFiltering
	if s.Model.Width > 0 {
		s.TextInput.Width = s.Model.Width - 10
		if s.TextInput.Width < 30 {
			s.TextInput.Width = 30
		}
	}
	s.TextInput.Prompt = "FILTER MODE "
	s.TextInput.Placeholder = "Enter filters..."
	s.TextInput.SetValue(s.Model.View.Filter.Raw)
	return s, s.TextInput.Focus()
}

func ApplyFilter(s State, msg ApplyFilterMsg) (State, tea.Cmd) {
	fs := state.ParseFilter(msg.Query)
	s.Model.View.Filter = fs
	if fs.GroupBy != "" {
		s.Model.View.TableGroupBy = fs.GroupBy
	}
	s.BoardModel = boardPkg.NewBoardModel(s.Model.Items, s.Model.Project.Fields, fs, s.Model.View.FocusedItemID, s.Model.View.CardFieldVisibility)
	s.Model.View.Mode = state.ModeNormal
	s.TextInput.Prompt = ""
	return s, nil
}

func ClearFilter(s State, _ ClearFilterMsg) (State, tea.Cmd) {
	s.Model.View.Filter = state.FilterState{}
	s.Model.View.TableGroupBy = ""
	s.BoardModel = boardPkg.NewBoardModel(s.Model.Items, s.Model.Project.Fields, state.FilterState{}, s.Model.View.FocusedItemID, s.Model.View.CardFieldVisibility)
	if s.Model.View.Mode == state.ModeFiltering {
		s.Model.View.Mode = state.ModeNormal
	}
	s.TextInput.SetValue("")
	s.TextInput.Prompt = ""
	return s, nil
}
