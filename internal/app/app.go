package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"project-hub/internal/config"
	"project-hub/internal/github"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
	"project-hub/internal/ui/settings"
	"project-hub/internal/ui/table"
)

const (
	groupByStatus    = "status"
	groupByAssignee  = "assignee"
	groupByIteration = "iteration"
)

// App implements the Bubbletea Model interface and holds root state.
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

// New creates an App with an optional preloaded state.
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

// Msg types for asynchronous operations
type FetchProjectMsg struct {
	Project state.Project
	Items   []state.Item
}

type ItemUpdatedMsg struct {
	Index int
	Item  state.Item
}

type ErrMsg struct {
	Err error
}

func NewErrMsg(err error) ErrMsg {
	return ErrMsg{Err: err}
}

type DismissNotificationMsg struct {
	ID int
}

type EnterStatusSelectModeMsg struct{}

type DetailReadyMsg struct {
	Item state.Item
}

// Cmds for asynchronous operations
func FetchProjectCmd(client github.Client, projectID, owner string, itemLimit int) tea.Cmd {
	return func() tea.Msg {
		proj, items, err := client.FetchProject(context.Background(), projectID, owner, itemLimit)
		if err != nil {
			return NewErrMsg(err)
		}
		return FetchProjectMsg{Project: proj, Items: items}
	}
}

func dismissNotificationCmd(id int, duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return DismissNotificationMsg{ID: id}
	})
}

// Init loads initial project data.
func (a App) Init() tea.Cmd {
	return tea.Batch(FetchProjectCmd(a.github, a.state.Project.ID, a.state.Project.Owner, a.itemLimit), textinput.Blink)
}

// Update routes incoming messages to state transitions.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Handle input while in status select mode
	if a.state.View.Mode == state.ModeStatusSelect {
		var statusSelectorCmd tea.Cmd
		var updatedStatusSelector tea.Model
		updatedStatusSelector, statusSelectorCmd = a.statusSelector.Update(msg)
		a.statusSelector = updatedStatusSelector.(components.StatusSelectorModel)
		if statusSelectorCmd != nil {
			cmds = append(cmds, statusSelectorCmd)
		}

		switch m := msg.(type) {
		case components.StatusSelectedMsg:
			if m.Canceled {
				a.state.View.Mode = state.ModeNormal // Exit status select mode
				return a, tea.Batch(cmds...)
			}
			// Handle status update
			idx := a.state.View.FocusedIndex
			if idx < 0 || idx >= len(a.state.Items) {
				a.state.View.Mode = state.ModeNormal // Exit status select mode
				return a, tea.Batch(cmds...)
			}
			item := a.state.Items[idx]

			if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
				notif := state.Notification{
					Message:      fmt.Sprintf("Invalid item ID format: %s. Expected project item node ID.", item.ID),
					Level:        "error",
					At:           time.Now(),
					DismissAfter: 5 * time.Second,
				}
				a.state.Notifications = append(a.state.Notifications, notif)
				a.state.View.Mode = state.ModeNormal
				return a, tea.Batch(append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))...)
			}

			// Call the GitHub client to update the status
			updateCmd := func() tea.Msg {
				updatedItem, err := a.github.UpdateStatus(
					context.Background(),
					projectMutationID(a.state.Project),
					a.state.Project.Owner,
					item.ID,         // Use the item's node ID for field updates
					m.StatusFieldID, // Use the selected status field ID
					m.OptionID,      // Use the selected option ID
				)
				if err != nil {
					return NewErrMsg(err)
				}
				if strings.EqualFold(strings.TrimSpace(updatedItem.Status), "unknown") && strings.TrimSpace(m.OptionName) != "" {
					updatedItem.Status = m.OptionName
				}
				return ItemUpdatedMsg{Index: idx, Item: updatedItem}
			}

			a.state.View.Mode = state.ModeNormal // Exit status select mode
			return a, tea.Batch(append(cmds, updateCmd)...)

		default:
			return a, tea.Batch(cmds...)
		}
	}

	if a.state.View.Mode == state.ModeLabelSelect || a.state.View.Mode == state.ModeMilestoneSelect || a.state.View.Mode == state.ModePrioritySelect {
		var fieldSelectorCmd tea.Cmd
		var updatedFieldSelector tea.Model
		updatedFieldSelector, fieldSelectorCmd = a.fieldSelector.Update(msg)
		a.fieldSelector = updatedFieldSelector.(components.FieldSelectorModel)
		if fieldSelectorCmd != nil {
			cmds = append(cmds, fieldSelectorCmd)
		}

		switch m := msg.(type) {
		case components.FieldSelectedMsg:
			if m.Canceled {
				a.state.View.Mode = state.ModeNormal
				return a, tea.Batch(cmds...)
			}
			idx := a.state.View.FocusedIndex
			if idx < 0 || idx >= len(a.state.Items) {
				a.state.View.Mode = state.ModeNormal
				return a, tea.Batch(cmds...)
			}
			item := a.state.Items[idx]

			if item.ID == "" || !strings.HasPrefix(item.ID, "PVTI_") {
				notif := state.Notification{
					Message:      fmt.Sprintf("Invalid item ID format: %s. Expected project item node ID.", item.ID),
					Level:        "error",
					At:           time.Now(),
					DismissAfter: 5 * time.Second,
				}
				a.state.Notifications = append(a.state.Notifications, notif)
				a.state.View.Mode = state.ModeNormal
				return a, tea.Batch(append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))...)
			}

			updateCmd := func() tea.Msg {
				updatedItem, err := a.github.UpdateField(
					context.Background(),
					projectMutationID(a.state.Project),
					a.state.Project.Owner,
					item.ID,
					m.FieldID,
					m.OptionID,
					m.FieldName,
				)
				if err != nil {
					return NewErrMsg(err)
				}
				return ItemUpdatedMsg{Index: idx, Item: updatedItem}
			}

			a.state.View.Mode = state.ModeNormal
			return a, tea.Batch(append(cmds, updateCmd)...)

		default:
			return a, tea.Batch(cmds...)
		}
	}

	if a.state.View.Mode == state.ModeDetail {
		var detailPanelCmd tea.Cmd
		var updatedDetailPanel tea.Model
		updatedDetailPanel, detailPanelCmd = a.detailPanel.Update(msg)
		a.detailPanel = updatedDetailPanel.(components.DetailPanelModel)
		if detailPanelCmd != nil {
			cmds = append(cmds, detailPanelCmd)
		}

		switch msg.(type) {
		case components.DetailCloseMsg:
			a.state.View.Mode = state.ModeNormal
			return a, tea.Batch(cmds...)
		default:
			return a, tea.Batch(cmds...)
		}
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
			// Pass key events to settings model
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
		// Merge updated fields into the existing item to avoid overwriting
		// other fields when the returned updated item is partial (e.g. only
		// contains Assignees). This prevents transient UI remapping where an
		// item loses its Status or other metadata.
		if m.Index >= 0 && m.Index < len(a.state.Items) {
			existing := a.state.Items[m.Index]
			// Merge assignees if provided
			if len(m.Item.Assignees) > 0 {
				existing.Assignees = m.Item.Assignees
			}
			// Merge labels if provided
			if len(m.Item.Labels) > 0 {
				existing.Labels = m.Item.Labels
			}
			// Merge title if provided
			if m.Item.Title != "" {
				existing.Title = m.Item.Title
			}
			// Merge status if provided
			if m.Item.Status != "" && m.Item.Status != "Unknown" {
				existing.Status = m.Item.Status
			}
			// Merge priority if provided
			if m.Item.Priority != "" {
				existing.Priority = m.Item.Priority
			}
			// Merge repository/number if provided
			if m.Item.Repository != "" {
				existing.Repository = m.Item.Repository
			}
			if m.Item.Number != 0 {
				existing.Number = m.Item.Number
			}
			a.state.Items[m.Index] = existing
		} else {
			// Fallback: out-of-range index, replace if possible
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
	// case StatusMoveMsg: // Removed as per new design
	// 	var model tea.Model
	// 	model, cmd = a.handleStatusMove(m)
	// 	a = model.(App)
	// 	cmds = append(cmds, cmd)
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
	case EnterStatusSelectModeMsg: // New: handle entering status select mode
		var model tea.Model
		model, cmd = a.handleEnterStatusSelectMode(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case settings.SaveMsg:
		configPath, err := config.ResolvePath()
		if err == nil {
			cfg := config.Config{
				DefaultProjectID:   m.ProjectID,
				DefaultOwner:       m.Owner,
				SuppressHints:      m.SuppressHints,
				DefaultItemLimit:   m.ItemLimit,
				DefaultExcludeDone: m.ExcludeDone,
			}
			if saveErr := config.Save(configPath, cfg); saveErr == nil {
				a.state.SuppressHints = m.SuppressHints
				a.state.ItemLimit = m.ItemLimit
				a.state.ExcludeDone = m.ExcludeDone
				if !a.state.SuppressHints {
					notif := state.Notification{
						Message:      "Settings saved successfully",
						Level:        "info",
						At:           time.Now(),
						DismissAfter: 3 * time.Second,
					}
					a.state.Notifications = append(a.state.Notifications, notif)
					cmds = append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
				}
			} else {
				notif := state.Notification{
					Message:      fmt.Sprintf("Failed to save settings: %v", saveErr),
					Level:        "error",
					At:           time.Now(),
					DismissAfter: 5 * time.Second,
				}
				a.state.Notifications = append(a.state.Notifications, notif)
				cmds = append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
			}
		} else {
			notif := state.Notification{
				Message:      fmt.Sprintf("Failed to resolve config path: %v", err),
				Level:        "error",
				At:           time.Now(),
				DismissAfter: 5 * time.Second,
			}
			a.state.Notifications = append(a.state.Notifications, notif)
			cmds = append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
		}
		// Switch back to board view
		a.state.View.CurrentView = state.ViewBoard
	case settings.CancelMsg:
		// Just switch back to board view
		a.state.View.CurrentView = state.ViewBoard
	default:
		// Handle other messages or pass them to sub-models
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

	// Handle field toggle mode
	if a.state.View.Mode == state.ModeFieldToggle {
		return a.handleFieldToggle(k.String())
	}

	// Handle input while in sort mode (ModeSort is a typed constant)
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
			// move focus down
			return a.handleMoveFocus(MoveFocusMsg{Delta: 1})
		case "k", "up":
			// move focus up
			return a.handleMoveFocus(MoveFocusMsg{Delta: -1})
		case "h":
			// move focus left (column)
			a.state.View.FocusedColumnIndex--
			if a.state.View.FocusedColumnIndex < 0 {
				a.state.View.FocusedColumnIndex = 0
			}
			return a, nil
		case "l":
			// move focus right (column)
			// TODO: dynamically get max columns from table.view
			maxCols := state.ColumnCount
			a.state.View.FocusedColumnIndex++
			if a.state.View.FocusedColumnIndex >= maxCols {
				a.state.View.FocusedColumnIndex = maxCols - 1
			}
			return a, nil
		case "esc":
			// cancel sort mode
			a.state.View.Mode = state.ModeNormal
			return a, nil
		default:
			// ignore other keys in sort mode
			return a, nil
		}
		// apply focus to first item after sort
		if len(a.state.Items) > 0 {
			a.state.View.FocusedIndex = 0
			a.state.View.FocusedItemID = a.state.Items[0].ID
		}
		// exit sort mode
		a.state.View.Mode = state.ModeNormal
		// notify
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
			a.syncFocusedItem() // Sync main state with board's focus
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
	case "h": // move focus left (column)
		a.state.View.FocusedColumnIndex--
		if a.state.View.FocusedColumnIndex < 0 {
			a.state.View.FocusedColumnIndex = 0
		}
		return a, nil
	case "l": // move focus right (column)
		// TODO: dynamically get max columns from table.view
		maxCols := state.ColumnCount
		a.state.View.FocusedColumnIndex++
		if a.state.View.FocusedColumnIndex >= maxCols {
			a.state.View.FocusedColumnIndex = maxCols - 1
		}
		return a, nil
	case "/":
		return a.handleEnterFilterMode(EnterFilterModeMsg{})
	case "s":
		// enter sort mode only in table view
		if a.state.View.CurrentView == state.ViewTable {
			a.state.View.Mode = state.ModeSort
			// notify instructions
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
		// In table view, Enter should trigger edit based on focused column
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

	// Find the corresponding item in the main list and update the app's focused index
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

// LoadInitialState is a helper to fetch project data using the gh client.
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

// View renders the current UI.
func (a App) View() string {
	width := a.state.Width
	if width == 0 {
		width = 100
	}
	header := components.RenderHeader(a.state.Project, a.state.View, width)
	items := applyFilter(a.state.Items, a.state.Project.Fields, a.state.View.Filter)
	// Apply table sort if present
	items = applySort(items, a.state.View.TableSort)

	// Compute frame and inner width early so views can size content appropriately
	frameWidth := width
	if frameWidth <= 0 {
		frameWidth = 100
	}
	innerWidth := frameWidth - components.FrameStyle.GetHorizontalFrameSize()
	if innerWidth < 40 {
		innerWidth = 40
	}

	editTitle := ""
	if a.state.View.Mode == "edit" || a.state.View.Mode == "assign" || a.state.View.Mode == "labelsInput" || a.state.View.Mode == "milestoneInput" || a.state.View.Mode == state.ModeFiltering {
		editTitle = a.textInput.Value()
	}
	footer := components.RenderFooter(string(a.state.View.Mode), string(a.state.View.CurrentView), width, editTitle)
	notif := components.RenderNotifications(a.state.Notifications)

	body := ""
	bodyHeight := a.bodyViewportHeight(header, footer, notif)
	frameVertical := components.FrameStyle.GetVerticalFrameSize()
	switch a.state.View.CurrentView {
	case state.ViewTable:
		groupBy := strings.ToLower(strings.TrimSpace(a.state.View.TableGroupBy))
		if groupBy != "" {
			groupedView := renderGroupedTable(groupBy, items, a.state.Project.Fields, a.state.View.FocusedItemID, a.state.View.FocusedColumnIndex, innerWidth)
			headerHeight := lipgloss.Height(groupedView.Header)
			rowsHeight := bodyHeight - headerHeight - frameVertical
			if rowsHeight < 3 {
				rowsHeight = 3
			}
			if a.tableViewport != nil {
				rowsContent := strings.Join(groupedView.Rows, "\n") + "\n"
				a.ensureTableViewportSize(innerWidth, rowsHeight)
				a.tableViewport.SetContent(rowsContent)
				focusedRow := focusedGroupedRowIndex(groupedView.Groups, a.state.View.FocusedItemID)
				focusTop, focusBottom := groupedView.RowBounds(focusedRow)
				a.syncTableViewportToFocus(focusTop, focusBottom, groupedView.RowsLineSize)
				body = lipgloss.JoinVertical(lipgloss.Left, groupedView.Header, a.tableViewport.View())
			} else {
				body = lipgloss.JoinVertical(lipgloss.Left, append([]string{groupedView.Header}, groupedView.Rows...)...)
			}
		} else {
			tableView := table.Render(items, a.state.View.FocusedItemID, a.state.View.FocusedColumnIndex, innerWidth)
			headerHeight := lipgloss.Height(tableView.Header)
			rowsHeight := bodyHeight - headerHeight - frameVertical
			if rowsHeight < 3 {
				rowsHeight = 3
			}
			if a.tableViewport != nil {
				// Add trailing newline to ensure viewport properly calculates all lines
				rowsContent := strings.Join(tableView.Rows, "\n") + "\n"
				a.ensureTableViewportSize(innerWidth, rowsHeight)
				a.tableViewport.SetContent(rowsContent)
				focusedRow := focusedRowIndex(items, a.state.View.FocusedItemID)
				focusTop, focusBottom := tableView.RowBounds(focusedRow)
				a.syncTableViewportToFocus(focusTop, focusBottom, tableView.RowsLineSize)
				body = lipgloss.JoinVertical(lipgloss.Left, tableView.Header, a.tableViewport.View())
			} else {
				body = lipgloss.JoinVertical(lipgloss.Left, append([]string{tableView.Header}, tableView.Rows...)...)
			}
		}
	case state.ViewSettings:
		a.settingsModel.SetSize(innerWidth, bodyHeight)
		body = a.settingsModel.View()
	default:
		a.boardModel.Width = innerWidth
		a.boardModel.Height = bodyHeight
		a.boardModel.EnsureLayout()
		body = a.boardModel.View()
	}

	// Handle edit/assign/labelInput/milestoneInput modes
	if a.state.View.Mode == "edit" || a.state.View.Mode == "assign" || a.state.View.Mode == "labelsInput" || a.state.View.Mode == "milestoneInput" || a.state.View.Mode == state.ModeFiltering {
		body = body + "\n" + a.textInput.View()
	}

	var framed string
	if a.state.View.CurrentView == state.ViewBoard {
		maxHeight := bodyHeight
		if maxHeight < 15 {
			maxHeight = 15
		}
		if maxHeight < 1 {
			maxHeight = 1
		}
		bodyRendered := clampContentHeight(body, maxHeight)
		bodyRendered = lipgloss.NewStyle().Width(innerWidth).AlignHorizontal(lipgloss.Left).Render(bodyRendered)
		framed = components.FrameStyle.Width(frameWidth).Render(bodyRendered)
	} else {
		maxHeight := bodyHeight - frameVertical
		if maxHeight < 1 {
			maxHeight = 1
		}
		bodyRendered := clampContentHeight(body, maxHeight)
		bodyRendered = lipgloss.NewStyle().Width(innerWidth).AlignHorizontal(lipgloss.Left).Render(bodyRendered)
		framed = components.FrameStyle.Width(frameWidth).Render(bodyRendered)
	}

	if a.state.View.Mode == state.ModeStatusSelect {
		selectorView := a.statusSelector.View()
		framed = lipgloss.Place(
			frameWidth,
			bodyHeight,
			lipgloss.Center,
			lipgloss.Center,
			selectorView,
		)
	}

	if a.state.View.Mode == state.ModeLabelSelect || a.state.View.Mode == state.ModeMilestoneSelect || a.state.View.Mode == state.ModePrioritySelect {
		selectorView := a.fieldSelector.View()
		framed = lipgloss.Place(
			frameWidth,
			bodyHeight,
			lipgloss.Center,
			lipgloss.Center,
			selectorView,
		)
	}

	if a.state.View.Mode == state.ModeDetail {
		detailView := a.detailPanel.View()
		framed = lipgloss.Place(
			frameWidth,
			bodyHeight,
			lipgloss.Center,
			lipgloss.Center,
			detailView,
		)
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, framed, footer, notif)
}

type groupedTableView struct {
	Header       string
	Rows         []string
	RowHeights   []int
	RowOffsets   []int
	RowsLineSize int
	Groups       []boardPkg.GroupBucket
}

func renderGroupedTable(groupBy string, items []state.Item, fields []state.Field, focusedID string, focusedColIndex int, innerWidth int) groupedTableView {
	if innerWidth <= 0 {
		innerWidth = 80
	}
	sepLen := innerWidth
	var groups []boardPkg.GroupBucket
	switch groupBy {
	case groupByStatus:
		groups = boardPkg.GroupItemsByStatusBuckets(items, fields)
	case groupByIteration:
		groups = boardPkg.GroupItemsByIteration(items)
	case groupByAssignee:
		groups = boardPkg.GroupItemsByAssignee(items)
	default:
		groups = []boardPkg.GroupBucket{{Name: "Items", Items: items}}
	}
	if len(groups) == 0 {
		groups = []boardPkg.GroupBucket{{Name: "Items", Items: items}}
	}

	var header string
	var rows []string
	var rowHeights []int
	var rowOffsets []int
	var cumulativeHeight int
	groupHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

	for i, group := range groups {
		if i == 0 {
			groupRender := table.Render(group.Items, focusedID, focusedColIndex, innerWidth)
			header = groupRender.Header
		}

		groupHeader := fmt.Sprintf("# %s (%d)", group.Name, len(group.Items))
		groupHeaderRow := groupHeaderStyle.Render(groupHeader)
		separator := ""
		if sepLen > 0 {
			separator = strings.Repeat("─", sepLen)
		}
		groupHeaderView := lipgloss.JoinVertical(lipgloss.Left, groupHeaderRow, separator)
		rows = append(rows, groupHeaderView)
		rowHeights = append(rowHeights, lipgloss.Height(groupHeaderView))
		rowOffsets = append(rowOffsets, cumulativeHeight)
		cumulativeHeight += lipgloss.Height(groupHeaderView)

		groupRender := table.Render(group.Items, focusedID, focusedColIndex, innerWidth)
		rows = append(rows, groupRender.Rows...)
		for _, h := range groupRender.RowHeights {
			rowHeights = append(rowHeights, h)
			rowOffsets = append(rowOffsets, cumulativeHeight)
			cumulativeHeight += h
		}
	}

	return groupedTableView{
		Header:       header,
		Rows:         rows,
		RowHeights:   rowHeights,
		RowOffsets:   rowOffsets,
		RowsLineSize: cumulativeHeight,
		Groups:       groups,
	}
}

func (g groupedTableView) RowBounds(index int) (int, int) {
	if index < 0 || index >= len(g.RowOffsets) || index >= len(g.RowHeights) {
		return -1, -1
	}
	top := g.RowOffsets[index]
	height := g.RowHeights[index]
	if height <= 0 {
		height = 1
	}
	bottom := top + height - 1
	return top, bottom
}

func focusedGroupedRowIndex(groups []boardPkg.GroupBucket, focusedID string) int {
	index := 0
	for _, group := range groups {
		index++
		for _, item := range group.Items {
			if item.ID == focusedID {
				return index
			}
			index++
		}
	}
	return -1
}

func applyFilter(items []state.Item, fields []state.Field, fs state.FilterState) []state.Item {
	if fs.Query == "" && len(fs.Labels) == 0 && len(fs.Assignees) == 0 && len(fs.Statuses) == 0 && len(fs.Iterations) == 0 && len(fs.FieldFilters) == 0 {
		return items
	}
	var out []state.Item
	for _, it := range items {
		if fs.Query != "" && !strings.Contains(strings.ToLower(it.Title), strings.ToLower(fs.Query)) {
			continue
		}
		if len(fs.Labels) > 0 && !containsAny(it.Labels, fs.Labels) {
			continue
		}
		if len(fs.Assignees) > 0 && !containsAny(it.Assignees, fs.Assignees) {
			continue
		}
		if len(fs.Statuses) > 0 && !containsAny([]string{it.Status}, fs.Statuses) {
			continue
		}
		if len(fs.Iterations) > 0 && !state.MatchesIterationFilters(it, fs.Iterations, time.Now()) {
			continue
		}
		if len(fs.FieldFilters) > 0 && !matchesFieldFilters(it, fields, fs.FieldFilters) {
			continue
		}
		out = append(out, it)
	}
	return out
}

func matchesFieldFilters(item state.Item, fields []state.Field, filters map[string][]string) bool {
	if len(filters) == 0 {
		return true
	}
	for name, values := range filters {
		if !matchesSingleFieldFilter(item, fields, name, values) {
			return false
		}
	}
	return true
}

func matchesSingleFieldFilter(item state.Item, fields []state.Field, name string, values []string) bool {
	if len(values) == 0 {
		return true
	}
	fieldName := strings.TrimSpace(name)
	if fieldName == "" {
		return true
	}
	if strings.EqualFold(fieldName, "title") {
		return matchSliceValues([]string{item.Title}, values)
	}
	if strings.EqualFold(fieldName, "status") {
		return matchSliceValues([]string{item.Status}, values)
	}
	if strings.EqualFold(fieldName, "priority") {
		return matchSliceValues([]string{item.Priority}, values)
	}
	if strings.EqualFold(fieldName, "milestone") {
		return matchSliceValues([]string{item.Milestone}, values)
	}
	if strings.EqualFold(fieldName, "labels") || strings.EqualFold(fieldName, "label") {
		return matchSliceValues(item.Labels, values)
	}
	if strings.EqualFold(fieldName, "assignees") || strings.EqualFold(fieldName, "assignee") {
		return matchSliceValues(item.Assignees, values)
	}
	if strings.EqualFold(fieldName, "iteration") {
		return state.MatchesIterationFilters(item, values, time.Now())
	}
	if len(item.FieldValues) > 0 {
		for key, stored := range item.FieldValues {
			if strings.EqualFold(key, fieldName) {
				return matchSliceValues(stored, values)
			}
		}
	}
	for _, field := range fields {
		if strings.EqualFold(field.Name, fieldName) {
			stored := item.FieldValues[field.Name]
			return matchSliceValues(stored, values)
		}
	}
	return false
}

func matchSliceValues(haystack []string, needles []string) bool {
	if len(needles) == 0 {
		return true
	}
	for _, needle := range needles {
		if needle == "" {
			continue
		}
		for _, value := range haystack {
			if value == "" {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(needle)) {
				return true
			}
		}
	}
	return false
}

func containsAny(haystack []string, needles []string) bool {
	for _, n := range needles {
		for _, h := range haystack {
			if h == n {
				return true
			}
		}
	}
	return false
}

// toggleSort toggles the sort direction or sets new field with ascending order.
func toggleSort(ts state.TableSort, field string) state.TableSort {
	if ts.Field == field {
		ts.Asc = !ts.Asc
		return ts
	}
	return state.TableSort{Field: field, Asc: true}
}

// applySort orders items according to TableSort. If Field is empty, no-op.
func applySort(items []state.Item, ts state.TableSort) []state.Item {
	if ts.Field == "" {
		return items
	}
	// use stable sort
	switch ts.Field {
	case "Title":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Title < items[j].Title
			}
			return items[i].Title > items[j].Title
		})
	case "Status":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Status < items[j].Status
			}
			return items[i].Status > items[j].Status
		})
	case "Repository":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Repository < items[j].Repository
			}
			return items[i].Repository > items[j].Repository
		})
	case "Labels":
		sort.SliceStable(items, func(i, j int) bool {
			iLabels := strings.Join(items[i].Labels, ",")
			jLabels := strings.Join(items[j].Labels, ",")
			if ts.Asc {
				return iLabels < jLabels
			}
			return iLabels > jLabels
		})
	case "Milestone":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Milestone < items[j].Milestone
			}
			return items[i].Milestone > items[j].Milestone
		})
	case "Priority":
		// Priority: attempt to order High > Medium > Low when Asc=false
		priorityRank := func(p string) int {
			switch strings.ToLower(p) {
			case "high":
				return 3
			case "medium":
				return 2
			case "low":
				return 1
			default:
				return 0
			}
		}
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return priorityRank(items[i].Priority) < priorityRank(items[j].Priority)
			}
			return priorityRank(items[i].Priority) > priorityRank(items[j].Priority)
		})
	default:
		// try numeric/date fields
		switch ts.Field {
		case "Number":
			sort.SliceStable(items, func(i, j int) bool {
				if ts.Asc {
					return items[i].Number < items[j].Number
				}
				return items[i].Number > items[j].Number
			})
		case "CreatedAt":
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].CreatedAt == nil || items[j].CreatedAt == nil {
					return items[i].CreatedAt == nil
				}
				if ts.Asc {
					return items[i].CreatedAt.Before(*items[j].CreatedAt)
				}
				return items[j].CreatedAt.Before(*items[i].CreatedAt)
			})
		case "UpdatedAt":
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].UpdatedAt == nil || items[j].UpdatedAt == nil {
					return items[i].UpdatedAt == nil
				}
				if ts.Asc {
					return items[i].UpdatedAt.Before(*items[j].UpdatedAt)
				}
				return items[j].UpdatedAt.Before(*items[i].UpdatedAt)
			})
		default:
			// unsupported field: no-op
		}
	}
	return items
}

// handleSaveEdit is moved here from update_edit.go
func (a App) handleSaveEdit(msg SaveEditMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	item := a.state.Items[idx]

	// Call the GitHub client to update the item
	updateCmd := func() tea.Msg {
		updatedItem, err := a.github.UpdateItem(
			context.Background(),
			projectMutationID(a.state.Project),
			a.state.Project.Owner,
			item, // Pass the whole item
			msg.Title,
			msg.Description,
		)
		if err != nil {
			return NewErrMsg(err)
		}
		return ItemUpdatedMsg{Index: idx, Item: updatedItem}
	}

	a.state.View.Mode = state.ModeNormal // Use state.ModeNormal instead of "normal" string
	return a, tea.Batch(updateCmd)
}

func (a App) handleEnterStatusSelectMode(msg EnterStatusSelectModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil // No item focused, do nothing
	}
	focusedItem := a.state.Items[idx]

	// Find the "Status" field from the project's fields
	var statusField state.Field
	found := false
	for _, field := range a.state.Project.Fields {
		if field.Name == "Status" {
			statusField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: fmt.Sprintf("Status field not found in project"), Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	// Initialize statusSelectorModel
	a.statusSelector = components.NewStatusSelectorModel(focusedItem, statusField, a.state.Width, a.state.Height)

	// Set mode to status select
	a.state.View.Mode = state.ModeStatusSelect

	// Optionally provide a notification to the user
	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Status select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, tea.Batch(a.statusSelector.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	}
	return a, a.statusSelector.Init()
}

func (a App) handleColumnEdit(msg EnterEditModeMsg) (tea.Model, tea.Cmd) {
	colIdx := a.state.View.FocusedColumnIndex
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}

	switch colIdx {
	case state.ColumnTitle:
		return a.handleEnterEditMode(msg)
	case state.ColumnStatus:
		return a.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
	case state.ColumnAssignees:
		return a.handleEnterAssignMode(EnterAssignModeMsg{})
	case state.ColumnLabels:
		return a.handleEnterLabelsInputMode(EnterLabelsInputModeMsg{})
	case state.ColumnMilestone:
		return a.handleEnterMilestoneInputMode(EnterMilestoneInputModeMsg{})
	case state.ColumnPriority:
		return a.handleEnterPrioritySelectMode(EnterPrioritySelectModeMsg{})
	default:
		return a.handleEnterEditMode(msg)
	}
}

func (a App) handleEnterLabelSelectMode(msg EnterLabelSelectModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	focusedItem := a.state.Items[idx]

	var labelField state.Field
	found := false
	for _, field := range a.state.Project.Fields {
		if field.Name == "Labels" {
			labelField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: "Labels field not found in project", Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	a.fieldSelector = components.NewFieldSelectorModel(focusedItem, labelField, a.state.Width, a.state.Height)
	a.state.View.Mode = state.ModeLabelSelect

	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Label select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, tea.Batch(a.fieldSelector.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	}
	return a, a.fieldSelector.Init()
}

func (a App) handleEnterMilestoneSelectMode(msg EnterMilestoneSelectModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	focusedItem := a.state.Items[idx]

	var milestoneField state.Field
	found := false
	for _, field := range a.state.Project.Fields {
		if field.Name == "Milestone" {
			milestoneField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: "Milestone field not found in project", Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	a.fieldSelector = components.NewFieldSelectorModel(focusedItem, milestoneField, a.state.Width, a.state.Height)
	a.state.View.Mode = state.ModeMilestoneSelect

	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Milestone select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, tea.Batch(a.fieldSelector.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	}
	return a, a.fieldSelector.Init()
}

func (a App) handleEnterPrioritySelectMode(msg EnterPrioritySelectModeMsg) (tea.Model, tea.Cmd) {
	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	focusedItem := a.state.Items[idx]

	var priorityField state.Field
	found := false
	for _, field := range a.state.Project.Fields {
		if field.Name == "Priority" {
			priorityField = field
			found = true
			break
		}
	}
	if !found {
		notif := state.Notification{Message: "Priority field not found in project", Level: "error", At: time.Now(), DismissAfter: 5 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}

	a.fieldSelector = components.NewFieldSelectorModel(focusedItem, priorityField, a.state.Width, a.state.Height)
	a.state.View.Mode = state.ModePrioritySelect

	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Priority select mode: Use arrow keys to select, enter to confirm, esc to cancel",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, tea.Batch(a.fieldSelector.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	}
	return a, a.fieldSelector.Init()
}

func (a App) handleEnterDetailMode() (tea.Model, tea.Cmd) {
	if a.state.View.CurrentView == state.ViewBoard {
		a.syncFocusedItem()
	}

	idx := a.state.View.FocusedIndex
	if idx < 0 || idx >= len(a.state.Items) {
		return a, nil
	}
	focusedItem := a.state.Items[idx]

	if focusedItem.Repository != "" && focusedItem.Number > 0 {
		a.detailPanel = components.NewDetailPanelModel(focusedItem, a.state.Width, a.state.Height)
		fetchDescCmd := func() tea.Msg {
			body, err := a.github.FetchIssueDetail(context.Background(), focusedItem.Repository, focusedItem.Number)
			if err != nil {
				return NewErrMsg(err)
			}
			return DetailReadyMsg{Item: state.Item{
				Title:       focusedItem.Title,
				Description: body,
				Number:      focusedItem.Number,
				Repository:  focusedItem.Repository,
				Status:      focusedItem.Status,
				Assignees:   focusedItem.Assignees,
				Labels:      focusedItem.Labels,
				Priority:    focusedItem.Priority,
				Milestone:   focusedItem.Milestone,
				URL:         focusedItem.URL,
			}}
		}
		a.state.View.Mode = state.ModeDetail
		return a, tea.Cmd(fetchDescCmd)
	}

	a.detailPanel = components.NewDetailPanelModel(focusedItem, a.state.Width, a.state.Height)
	a.state.View.Mode = state.ModeDetail
	if !a.state.SuppressHints {
		detailNotif := state.Notification{
			Message:      "Detail mode: j/k to scroll, esc/q to close",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, detailNotif)
		return a, tea.Batch(a.detailPanel.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, detailNotif.DismissAfter))
	}
	return a, a.detailPanel.Init()
}

func (a App) handleEnterFieldToggleMode() (tea.Model, tea.Cmd) {
	a.state.View.Mode = state.ModeFieldToggle
	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Field toggle mode: m=milestone r=repository l=labels s=sub-issue p=parent (toggle on/off, esc to cancel)",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 5 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}
	return a, nil
}

func (a App) handleFieldToggle(key string) (tea.Model, tea.Cmd) {
	vis := &a.state.View.CardFieldVisibility
	switch key {
	case "m", "M":
		vis.ShowMilestone = !vis.ShowMilestone
	case "r", "R":
		vis.ShowRepository = !vis.ShowRepository
	case "l", "L":
		vis.ShowLabels = !vis.ShowLabels
	case "s", "S":
		vis.ShowSubIssueProgress = !vis.ShowSubIssueProgress
	case "p", "P":
		vis.ShowParentIssue = !vis.ShowParentIssue
	case "esc":
		a.state.View.Mode = state.ModeNormal
		return a, nil
	default:
		return a, nil
	}

	a.state.View.Mode = state.ModeNormal
	a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.Project.Fields, a.state.View.Filter, a.state.View.FocusedItemID, a.state.View.CardFieldVisibility)

	configPath, err := config.ResolvePath()
	if err == nil {
		existingCfg, loadErr := config.Load(configPath)
		if loadErr == nil {
			cfg := config.Config{
				DefaultProjectID:   existingCfg.DefaultProjectID,
				DefaultOwner:       existingCfg.DefaultOwner,
				SuppressHints:      existingCfg.SuppressHints,
				DefaultItemLimit:   existingCfg.DefaultItemLimit,
				DefaultExcludeDone: existingCfg.DefaultExcludeDone,
				CardFieldVisibility: config.CardFieldVisibility{
					ShowMilestone:        vis.ShowMilestone,
					ShowRepository:       vis.ShowRepository,
					ShowSubIssueProgress: vis.ShowSubIssueProgress,
					ShowParentIssue:      vis.ShowParentIssue,
					ShowLabels:           vis.ShowLabels,
				},
			}
			config.Save(configPath, cfg)
		}
	}

	if !a.state.SuppressHints {
		notif := state.Notification{
			Message:      "Card fields preference saved",
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}
	return a, nil
}

func (a App) handleToggleGroupBy() (tea.Model, tea.Cmd) {
	current := strings.ToLower(strings.TrimSpace(a.state.View.TableGroupBy))
	switch current {
	case "":
		a.state.View.TableGroupBy = groupByStatus
	case groupByStatus:
		a.state.View.TableGroupBy = groupByAssignee
	case groupByAssignee:
		a.state.View.TableGroupBy = groupByIteration
	default:
		a.state.View.TableGroupBy = ""
	}

	if a.tableViewport != nil && a.state.View.TableGroupBy != "" {
		a.tableViewport.YOffset = 0
	}

	if !a.state.SuppressHints {
		label := a.state.View.TableGroupBy
		if label == "" {
			label = "none"
		}
		notif := state.Notification{
			Message:      fmt.Sprintf("Group: %s", label),
			Level:        "info",
			At:           time.Now(),
			DismissAfter: 3 * time.Second,
		}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
	}
	return a, nil
}

func (a App) refreshBoardCmd() tea.Cmd {
	return func() tea.Msg {
		// Re-fetch items to ensure the board is up-to-date after an update
		proj, items, err := a.github.FetchProject(context.Background(), a.state.Project.ID, a.state.Project.Owner, a.itemLimit)
		if err != nil {
			return NewErrMsg(err)
		}
		return FetchProjectMsg{Project: proj, Items: items}
	}
}

func projectMutationID(project state.Project) string {
	if strings.TrimSpace(project.NodeID) != "" {
		return project.NodeID
	}
	return project.ID
}

func (a App) bodyViewportHeight(header, footer, notif string) int {
	height := a.state.Height
	if height <= 0 {
		height = 40
	}
	available := height - lipgloss.Height(header) - lipgloss.Height(footer)
	if strings.TrimSpace(notif) != "" {
		available -= lipgloss.Height(notif)
	}
	available -= components.FrameStyle.GetVerticalFrameSize()
	if available < 5 {
		available = 5
	}
	return available
}

func clampContentHeight(content string, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) >= height {
		return strings.Join(lines[:height], "\n")
	}
	padding := make([]string, height-len(lines))
	return strings.Join(append(lines, padding...), "\n")
}

func (a App) ensureTableViewportSize(width, height int) {
	if a.tableViewport == nil {
		return
	}
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if a.tableViewport.Width != width {
		a.tableViewport.Width = width
	}
	if a.tableViewport.Height != height {
		a.tableViewport.Height = height
	}
}

func (a App) syncTableViewportToFocus(focusTop, focusBottom, totalLines int) {
	if a.tableViewport == nil {
		return
	}
	vp := a.tableViewport
	if totalLines <= 0 || focusTop < 0 || focusBottom < 0 {
		vp.YOffset = 0
		return
	}
	visibleHeight := vp.Height
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	if focusBottom-focusTop+1 > visibleHeight {
		focusBottom = focusTop + visibleHeight - 1
	}
	top := vp.YOffset
	bottom := vp.YOffset + visibleHeight - 1
	if focusTop < top {
		vp.SetYOffset(focusTop)
		return
	}
	if focusBottom > bottom {
		vp.SetYOffset(focusBottom - visibleHeight + 1)
	}
	maxOffset := totalLines - visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if vp.YOffset > maxOffset {
		vp.YOffset = maxOffset
	}
	if vp.YOffset < 0 {
		vp.YOffset = 0
	}
}

func focusedRowIndex(items []state.Item, focusedID string) int {
	if focusedID == "" {
		return -1
	}
	for idx, it := range items {
		if it.ID == focusedID {
			return idx
		}
	}
	return -1
}
