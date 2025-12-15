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
	"project-hub/internal/github"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
	"project-hub/internal/ui/roadmap"
	"project-hub/internal/ui/table"
)

// App implements the Bubbletea Model interface and holds root state.
type App struct {
	state          state.Model
	github         github.Client // Renamed from 'gh' to 'github' for clarity and to match the error
	itemLimit      int
	boardModel     boardPkg.BoardModel
	textInput      textinput.Model                // For edit/assign input
	statusSelector components.StatusSelectorModel // New field for status selection
}

// New creates an App with an optional preloaded state.
func New(initial state.Model, client github.Client, itemLimit int) App {
	boardModel := boardPkg.NewBoardModel(initial.Items, initial.View.Filter, initial.View.FocusedItemID)
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	return App{state: initial, github: client, itemLimit: itemLimit, boardModel: boardModel, textInput: ti}
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

type EnterStatusSelectModeMsg struct{} // New message type for entering status selection mode

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
	return FetchProjectCmd(a.github, a.state.Project.ID, a.state.Project.Owner, a.itemLimit)
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

			// Call the GitHub client to update the status
			updateCmd := func() tea.Msg {
				updatedItem, err := a.github.UpdateStatus(
					context.Background(),
					a.state.Project.ID,
					a.state.Project.Owner,
					item.ID,         // Use the item's node ID for field updates
					m.StatusFieldID, // Use the selected status field ID
					m.OptionID,      // Use the selected option ID
				)
				if err != nil {
					return NewErrMsg(err)
				}
				return ItemUpdatedMsg{Index: idx, Item: updatedItem}
			}

			a.state.View.Mode = state.ModeNormal // Exit status select mode
			return a, tea.Batch(append(cmds, updateCmd)...)

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
		var model tea.Model
		model, cmd = a.handleKey(m)
		a = model.(App)
		cmds = append(cmds, cmd)
	case FetchProjectMsg:
		a.state.Project = m.Project
		a.state.Items = m.Items
		if len(a.state.Items) > 0 {
			a.state.View.FocusedItemID = a.state.Items[0].ID
		}
		a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.View.Filter, a.state.View.FocusedItemID)
		notif := state.Notification{Message: fmt.Sprintf("Loaded %d items", len(a.state.Items)), Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		cmds = append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
	case ItemUpdatedMsg:
		a.state.Items[m.Index] = m.Item
		a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.View.Filter, a.state.View.FocusedItemID)
		notif := state.Notification{Message: "Item updated successfully", Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		cmds = append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
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
	default:
		// Handle other messages or pass them to sub-models
	}
	return a, tea.Batch(cmds...)
}

func (a App) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle input while in edit/assign modes (these modes are stored as raw strings elsewhere)
	if a.state.View.Mode == "edit" || a.state.View.Mode == "assign" {
		switch k.String() {
		case "enter":
			if a.state.View.Mode == "edit" {
				return a.handleSaveEdit(SaveEditMsg{Title: a.textInput.Value()})
			} else {
				return a.handleSaveAssign(SaveAssignMsg{Assignee: a.textInput.Value()})
			}
		case "esc":
			if a.state.View.Mode == "edit" {
				return a.handleCancelEdit(CancelEditMsg{})
			} else {
				return a.handleCancelAssign(CancelAssignMsg{})
			}
		default:
			var cmd tea.Cmd
			a.textInput, cmd = a.textInput.Update(k)
			return a, cmd
		}
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
			maxCols := 7 // Title, Status, Repository, Labels, Milestone, Priority, Assignees
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
		notif := state.Notification{Message: fmt.Sprintf("Sort: %s %s", a.state.View.TableSort.Field, func() string {
			if a.state.View.TableSort.Asc {
				return "↑"
			}
			return "↓"
		}()), Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
		a.state.Notifications = append(a.state.Notifications, notif)
		return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
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
		}
	}
	switch k.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "1", "b":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewBoard})
	case "2", "t":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewTable})
	case "3", "r":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewRoadmap})
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
		maxCols := 7 // Title, Status, Repository, Labels, Milestone, Priority, Assignees
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
			notif := state.Notification{Message: "Sort mode: t=Title s=Status r=Repository l=Labels m=Milestone p=Priority n=Number c=CreatedAt u=UpdatedAt (esc to cancel)", Level: "info", At: time.Now(), DismissAfter: 5 * time.Second}
			a.state.Notifications = append(a.state.Notifications, notif)
			return a, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter)
		}
		return a, nil
	case "esc":
		return a.handleClearFilter(ClearFilterMsg{})
	case "i", "enter":
		return a.handleEnterEditMode(EnterEditModeMsg{})
	case "a":
		return a.handleEnterAssignMode(EnterAssignModeMsg{})
	case "w": // New: enter status select mode
		if a.state.View.CurrentView == state.ViewTable { // Only allow in table view
			return a.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
		}
		return a, nil
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
	items := applyFilter(a.state.Items, a.state.View.Filter)
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

	body := ""
	switch a.state.View.CurrentView {
	case state.ViewTable:
		body = table.Render(items, a.state.View.FocusedItemID, a.state.View.FocusedColumnIndex, innerWidth)
	case state.ViewRoadmap:
		body = roadmap.Render(a.state.Project.Iterations, items, a.state.View.FocusedItemID)
	default:
		body = a.boardModel.View()
	}

	// Handle edit/assign modes
	if a.state.View.Mode == "edit" || a.state.View.Mode == "assign" {
		prompt := "Edit title: "
		if a.state.View.Mode == "assign" {
			prompt = "Assign to: "
		}
		body = body + "\n" + prompt + a.textInput.View()
	} else if a.state.View.Mode == state.ModeStatusSelect { // New: Handle status select mode
		body = lipgloss.JoinVertical(lipgloss.Left, body, a.statusSelector.View())
	}

	// For board view, limit height to prevent header from scrolling out
	var framed string
	if a.state.View.CurrentView == state.ViewBoard {
		maxHeight := a.state.Height - 30 // Reserve more space for header, footer, and margins
		if maxHeight < 15 {
			maxHeight = 15
		}
		bodyRendered := lipgloss.NewStyle().Width(innerWidth).Height(maxHeight).Render(body)
		framed = components.FrameStyle.Width(frameWidth).Render(bodyRendered)
	} else {
		bodyRendered := lipgloss.NewStyle().Width(innerWidth).Render(body)
		framed = components.FrameStyle.Width(frameWidth).Render(bodyRendered)
	}

	footer := components.RenderFooter(string(a.state.View.Mode), string(a.state.View.CurrentView), width)
	notif := components.RenderNotifications(a.state.Notifications)
	return fmt.Sprintf("%s\n%s\n%s\n%s", header, framed, footer, notif)
}

func applyFilter(items []state.Item, fs state.FilterState) []state.Item {
	if fs.Query == "" {
		return items
	}
	var out []state.Item
	for _, it := range items {
		if len(fs.Labels) > 0 && !containsAny(it.Labels, fs.Labels) {
			continue
		}
		if len(fs.Assignees) > 0 && !containsAny(it.Assignees, fs.Assignees) {
			continue
		}
		if len(fs.Statuses) > 0 && !containsAny([]string{it.Status}, fs.Statuses) {
			continue
		}
		out = append(out, it)
	}
	return out
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
			a.state.Project.ID,
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
	notif := state.Notification{
		Message:      "Status select mode: Use arrow keys to select, enter to confirm, esc to cancel",
		Level:        "info",
		At:           time.Now(),
		DismissAfter: 5 * time.Second,
	}
	a.state.Notifications = append(a.state.Notifications, notif)
	return a, tea.Batch(a.statusSelector.Init(), dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
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
