package app

import (
	tea "github.com/charmbracelet/bubbletea"

	boardPkg "project-hub/internal/ui/board"
)

// EnterEditModeMsg switches to edit mode for focused item.
type EnterEditModeMsg struct{}

// SaveEditMsg applies title/description changes to focused item.
type SaveEditMsg struct {
	Title       string
	Description string
}

// CancelEditMsg cancels edit mode.
type CancelEditMsg struct{}

// EnterAssignModeMsg switches to assign mode for focused item.
type EnterAssignModeMsg struct{}

// SaveAssignMsg applies assignee changes to focused item.
type SaveAssignMsg struct {
	Assignee string
}

// CancelAssignMsg cancels assign mode.
type CancelAssignMsg struct{}

func (a App) handleEnterEditMode(msg EnterEditModeMsg) (tea.Model, tea.Cmd) {
	a.state.View.Mode = "edit"
	return a, nil
}

func (a App) handleSaveEdit(msg SaveEditMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]
	if msg.Title != "" {
		item.Title = msg.Title
	}
	if msg.Description != "" {
		item.Description = msg.Description
	}
	a.state.Items[idx] = item
	a.state.View.Mode = "normal"
	return a, nil
}

func (a App) handleCancelEdit(msg CancelEditMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "edit" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}

func (a App) handleEnterAssignMode(msg EnterAssignModeMsg) (tea.Model, tea.Cmd) {
	a.state.View.Mode = "assign"
	return a, nil
}

func (a App) handleSaveAssign(msg SaveAssignMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]
	if msg.Assignee != "" {
		item.Assignees = []string{msg.Assignee}
	}
	a.state.Items[idx] = item
	a.state.View.Mode = "normal"
	// Update board model with new data
	a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.View.Filter, a.state.View.FocusedItemID)
	return a, nil
}

func (a App) handleCancelAssign(msg CancelAssignMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "assign" {
		a.state.View.Mode = "normal"
	}
	return a, nil
}
