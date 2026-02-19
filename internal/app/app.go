package app

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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
	boardModel := boardPkg.NewBoardModel(initial.Items, initial.Project.Fields, initial.View.Filter, initial.View.FocusedItemID, initial.View.CardFieldVisibility)
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.Width = 50
	ti.CharLimit = 500
	tableVP := viewport.New(0, 0)
	settingsModel := settings.New(initial.Project.ID, initial.Project.Owner, initial.SuppressHints, initial.ItemLimit, initial.ExcludeDone)
	return App{
		state:         initial,
		github:        client,
		itemLimit:     itemLimit,
		boardModel:    boardModel,
		textInput:     ti,
		settingsModel: settingsModel,
		tableViewport: &tableVP,
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(FetchProjectCmd(a.github, a.state.Project.ID, a.state.Project.Owner, a.itemLimit), textinput.Blink)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if a.state.View.Mode == state.ModeStatusSelect {
		updated, statusCmd := a.updateStatusSelectMode(msg)
		cmds = append(cmds, statusCmd)
		return updated, tea.Batch(cmds...)
	}

	if a.state.View.Mode == state.ModeLabelSelect || a.state.View.Mode == state.ModeMilestoneSelect || a.state.View.Mode == state.ModePrioritySelect {
		updated, fieldCmd := a.updateFieldSelectMode(msg)
		cmds = append(cmds, fieldCmd)
		return updated, tea.Batch(cmds...)
	}

	if a.state.View.Mode == state.ModeDetail {
		updated, detailCmd := a.updateDetailMode(msg)
		cmds = append(cmds, detailCmd)
		return updated, tea.Batch(cmds...)
	}

	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.state.Width = m.Width
		a.state.Height = m.Height
		if a.state.View.CurrentView == state.ViewBoard {
			var model tea.Model
			model, cmd = a.boardModel.Update(msg)
			a.boardModel = model.(boardPkg.BoardModel)
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		if a.state.View.CurrentView == state.ViewSettings {
			updatedModel, settingsCmd := a.settingsModel.Update(m)
			a.settingsModel = updatedModel
			if settingsCmd != nil {
				cmds = append(cmds, settingsCmd)
			}
		} else {
			var model tea.Model
			model, cmd = a.handleKey(m)
			a = model.(App)
			cmds = append(cmds, cmd)
		}
	case FetchProjectMsg:
		a.state.Project = m.Project
		a.state.Items = m.Items
		if a.state.ExcludeDone {
			var filtered []state.Item
			for _, item := range a.state.Items {
				if item.Status != "Done" {
					filtered = append(filtered, item)
				}
			}
			a.state.Items = filtered
		}
		if len(a.state.Items) > 0 {
			a.state.View.FocusedItemID = a.state.Items[0].ID
		}
		a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.Project.Fields, a.state.View.Filter, a.state.View.FocusedItemID, a.state.View.CardFieldVisibility)
	case ItemUpdatedMsg:
		if m.Index >= 0 && m.Index < len(a.state.Items) {
			existing := a.state.Items[m.Index]
			if len(m.Item.Assignees) > 0 {
				existing.Assignees = m.Item.Assignees
			}
			if len(m.Item.Labels) > 0 {
				existing.Labels = m.Item.Labels
			}
			if m.Item.Title != "" {
				existing.Title = m.Item.Title
			}
			if m.Item.Status != "" && m.Item.Status != "Unknown" {
				existing.Status = m.Item.Status
			}
			if m.Item.Priority != "" {
				existing.Priority = m.Item.Priority
			}
			if m.Item.Repository != "" {
				existing.Repository = m.Item.Repository
			}
			if m.Item.Number != 0 {
				existing.Number = m.Item.Number
			}
			a.state.Items[m.Index] = existing
		} else {
			if m.Index >= 0 && m.Index <= len(a.state.Items) {
				a.state.Items[m.Index] = m.Item
			}
		}
		a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.Project.Fields, a.state.View.Filter, a.state.View.FocusedItemID, a.state.View.CardFieldVisibility)
		if !a.state.SuppressHints {
			notif := state.Notification{Message: "Item updated successfully", Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
			a.state.Notifications = append(a.state.Notifications, notif)
			cmds = append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
		}
	case DetailReadyMsg:
		a.detailPanel = components.NewDetailPanelModel(m.Item, a.state.Width, a.state.Height)
		if !a.state.SuppressHints {
			detailNotif := state.Notification{Message: "Detail mode: j/k to scroll, esc/q to close", Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
			a.state.Notifications = append(a.state.Notifications, detailNotif)
			cmds = append(cmds, tea.Batch(a.detailPanel.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, detailNotif.DismissAfter)))
		} else {
			cmds = append(cmds, a.detailPanel.Init())
		}
	case ErrMsg:
		notif := state.Notification{Message: fmt.Sprintf("Error: %v", m.Err), Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		cmds = append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	case DismissNotificationMsg:
		if m.ID >= 0 && m.ID < len(a.state.Notifications) {
			a.state.Notifications[m.ID].Dismissed = true
		}
	case SwitchViewMsg:
		var model tea.Model
		model, cmd = a.handleSwitchView(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case MoveFocusMsg:
		var model tea.Model
		model, cmd = a.handleMoveFocus(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case EnterFilterModeMsg:
		var model tea.Model
		model, cmd = a.handleEnterFilterMode(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case ApplyFilterMsg:
		var model tea.Model
		model, cmd = a.handleApplyFilter(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case ClearFilterMsg:
		var model tea.Model
		model, cmd = a.handleClearFilter(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case EnterEditModeMsg:
		var model tea.Model
		model, cmd = a.handleEnterEditMode(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case SaveEditMsg:
		var model tea.Model
		model, cmd = a.handleSaveEdit(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case CancelEditMsg:
		var model tea.Model
		model, cmd = a.handleCancelEdit(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case EnterAssignModeMsg:
		var model tea.Model
		model, cmd = a.handleEnterAssignMode(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case SaveAssignMsg:
		var model tea.Model
		model, cmd = a.handleSaveAssign(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case CancelAssignMsg:
		var model tea.Model
		model, cmd = a.handleCancelAssign(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case AssignMsg:
		var model tea.Model
		model, cmd = a.handleAssign(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case EnterStatusSelectModeMsg:
		var model tea.Model
		model, cmd = a.handleEnterStatusSelectMode(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case settings.SaveMsg:
		updated, settingsCmd := a.handleSettingsSave(m)
		a = updated.(App)
		if settingsCmd != nil {
			cmds = append(cmds, settingsCmd)
		}
	case settings.CancelMsg:
		updated, settingsCmd := a.handleSettingsCancel(m)
		a = updated.(App)
		if settingsCmd != nil {
			cmds = append(cmds, settingsCmd)
		}
	default:
	}
	return a, tea.Batch(cmds...)
}

func (a App) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.state.View.Mode == "edit" || a.state.View.Mode == "assign" || a.state.View.Mode == "labelsInput" || a.state.View.Mode == "milestoneInput" || a.state.View.Mode == state.ModeFiltering {
		switch k.String() {
		case "enter":
			if a.state.View.Mode == "edit" {
				return a.handleSaveEdit(SaveEditMsg{Title: a.textInput.Value()})
			} else if a.state.View.Mode == "assign" {
				return a.handleSaveAssign(SaveAssignMsg{Assignee: a.textInput.Value()})
			} else if a.state.View.Mode == "labelsInput" {
				return a.handleSaveLabelsInput(SaveLabelsInputMsg{Labels: a.textInput.Value()})
			} else if a.state.View.Mode == "milestoneInput" {
				return a.handleSaveMilestoneInput(SaveMilestoneInputMsg{Milestone: a.textInput.Value()})
			} else if a.state.View.Mode == state.ModeFiltering {
				return a.handleApplyFilter(ApplyFilterMsg{Query: a.textInput.Value()})
			}
		case "esc":
			if a.state.View.Mode == "edit" {
				return a.handleCancelEdit(CancelEditMsg{})
			} else if a.state.View.Mode == "assign" {
				return a.handleCancelAssign(CancelAssignMsg{})
			} else if a.state.View.Mode == "labelsInput" {
				return a.handleCancelLabelsInput(CancelLabelsInputMsg{})
			} else if a.state.View.Mode == "milestoneInput" {
				return a.handleCancelMilestoneInput(CancelMilestoneInputMsg{})
			} else if a.state.View.Mode == state.ModeFiltering {
				return a.handleClearFilter(ClearFilterMsg{})
			}
		default:
			var cmd tea.Cmd
			a.textInput, cmd = a.textInput.Update(k)
			return a, cmd
		}
	}

	if a.state.View.Mode == state.ModeFieldToggle {
		return a.handleFieldToggle(k.String())
	}

	if a.state.View.Mode == state.ModeSort {
		switch k.String() {
		case "t", "T":
			a.state.View.TableSort = toggleSort(a.state.View.TableSort, "Title")
		case "s", "S":
			a.state.View.TableSort = toggleSort(a.state.View.TableSort, "Status")
		case "r", "R":
			a.state.View.TableSort = toggleSort(a.state.View.TableSort, "Repository")
		case "L":
			a.state.View.TableSort = toggleSort(a.state.View.TableSort, "Labels")
		case "m", "M":
			a.state.View.TableSort = toggleSort(a.state.View.TableSort, "Milestone")
		case "p", "P":
			a.state.View.TableSort = toggleSort(a.state.View.TableSort, "Priority")
		case "n", "N":
			a.state.View.TableSort = toggleSort(a.state.View.TableSort, "Number")
		case "c", "C":
			a.state.View.TableSort = toggleSort(a.state.View.TableSort, "CreatedAt")
		case "u", "U":
			a.state.View.TableSort = toggleSort(a.state.View.TableSort, "UpdatedAt")
		case "j", "down":
			return a.handleMoveFocus(MoveFocusMsg{Delta: 1})
		case "k", "up":
			return a.handleMoveFocus(MoveFocusMsg{Delta: -1})
		case "h":
			a.state.View.FocusedColumnIndex--
			if a.state.View.FocusedColumnIndex < 0 {
				a.state.View.FocusedColumnIndex = 0
			}
			return a, nil
		case "l":
			maxCols := state.ColumnCount
			a.state.View.FocusedColumnIndex++
			if a.state.View.FocusedColumnIndex >= maxCols {
				a.state.View.FocusedColumnIndex = maxCols - 1
			}
			return a, nil
		case "esc":
			a.state.View.Mode = state.ModeNormal
			return a, nil
		default:
			return a, nil
		}

		if len(a.state.Items) > 0 {
			a.state.View.FocusedIndex = 0
			a.state.View.FocusedItemID = a.state.Items[0].ID
		}
		a.state.View.Mode = state.ModeNormal
		if !a.state.SuppressHints {
			notif := state.Notification{Message: fmt.Sprintf("Sort: %s %s", a.state.View.TableSort.Field, func() string {
				if a.state.View.TableSort.Asc {
					return "↑"
				}
				return "↓"
			}()), Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
			a.state.Notifications = append(a.state.Notifications, notif)
			return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
		}
	}

	if a.state.View.CurrentView == state.ViewBoard {
		switch k.String() {
		case "j", "k", "h", "l":
			model, cmd := a.boardModel.Update(k)
			a.boardModel = model.(boardPkg.BoardModel)
			a.syncFocusedItem()
			return a, cmd
		case "/":
			return a.handleEnterFilterMode(EnterFilterModeMsg{})
		case "i":
			return a.handleEnterEditMode(EnterEditModeMsg{})
		case "a":
			return a.handleEnterAssignMode(EnterAssignModeMsg{})
		case "w":
			return a.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
		}
	}
	switch k.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "1", "b":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewBoard})
	case "2", "t":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewTable})
	case "3":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewSettings})
	case "R", "ctrl+r":
		return a.handleReload()
	case "j":
		return a.handleMoveFocus(MoveFocusMsg{Delta: 1})
	case "k":
		return a.handleMoveFocus(MoveFocusMsg{Delta: -1})
	case "h":
		a.state.View.FocusedColumnIndex--
		if a.state.View.FocusedColumnIndex < 0 {
			a.state.View.FocusedColumnIndex = 0
		}
		return a, nil
	case "l":
		maxCols := state.ColumnCount
		a.state.View.FocusedColumnIndex++
		if a.state.View.FocusedColumnIndex >= maxCols {
			a.state.View.FocusedColumnIndex = maxCols - 1
		}
		return a, nil
	case "/":
		return a.handleEnterFilterMode(EnterFilterModeMsg{})
	case "s":
		if a.state.View.CurrentView == state.ViewTable {
			a.state.View.Mode = state.ModeSort
			if !a.state.SuppressHints {
				notif := state.Notification{Message: "Sort mode: t=Title s=Status r=Repository l=Labels m=Milestone p=Priority n=Number c=CreatedAt u=UpdatedAt (esc to cancel)", Level: "info", At: time.Now(), DismissAfter: 5 * time.Second}
				a.state.Notifications = append(a.state.Notifications, notif)
				return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
			}
			return a, nil
		}
		return a, nil
	case "esc":
		return a.handleClearFilter(ClearFilterMsg{})
	case "i", "enter":
		if a.state.View.CurrentView == state.ViewTable {
			return a.handleColumnEdit(EnterEditModeMsg{})
		}
		return a.handleEnterEditMode(EnterEditModeMsg{})
	case "a":
		return a.handleEnterAssignMode(EnterAssignModeMsg{})
	case "w":
		if a.state.View.CurrentView == state.ViewTable {
			return a.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
		}
		return a, nil
	case "f":
		if a.state.View.CurrentView == state.ViewBoard {
			return a.handleEnterFieldToggleMode()
		}
		return a, nil
	case "g":
		if a.state.View.CurrentView == state.ViewTable {
			return a.handleToggleGroupBy()
		}
		return a, nil
	case "o":
		return a.handleEnterDetailMode()
	default:
		return a, nil
	}
}

func (a *App) syncFocusedItem() {
	colIndex := a.boardModel.FocusedColumnIndex
	cardIndex := a.boardModel.FocusedCardIndex

	if colIndex < 0 || colIndex >= len(a.boardModel.Columns) {
		return
	}
	column := a.boardModel.Columns[colIndex]

	if cardIndex < 0 || cardIndex >= len(column.Cards) {
		return
	}
	focusedCard := column.Cards[cardIndex]

	for i, item := range a.state.Items {
		if item.ID == focusedCard.ID {
			a.state.View.FocusedIndex = i
			a.state.View.FocusedItemID = item.ID
			return
		}
	}
}

func (a App) handleReload() (tea.Model, tea.Cmd) {
	return a, FetchProjectCmd(a.github, a.state.Project.ID, a.state.Project.Owner, a.itemLimit)
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
