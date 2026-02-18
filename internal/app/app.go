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
	"project-hub/internal/ui/roadmap"
	"project-hub/internal/ui/settings"
	"project-hub/internal/ui/table"
)

// App implements the Bubbletea Model interface and holds root state.
type App struct {
	state           state.Model
	github          github.Client
	itemLimit       int
	boardModel      boardPkg.BoardModel
	textInput       textinput.Model
	statusSelector  components.StatusSelectorModel
	settingsModel   settings.SettingsModel
	detailPanel     components.DetailPanelModel
	roadmapViewport *viewport.Model
	tableViewport   *viewport.Model
}

// New creates an App with an optional preloaded state.
func New(initial state.Model, client github.Client, itemLimit int) App {
	boardModel := boardPkg.NewBoardModel(initial.Items, initial.Project.Fields, initial.View.Filter, initial.View.FocusedItemID)
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	roadmapVP := viewport.New(0, 0)
	tableVP := viewport.New(0, 0)
	settingsModel := settings.New(initial.Project.ID, initial.Project.Owner)
	return App{
		state:           initial,
		github:          client,
		itemLimit:       itemLimit,
		boardModel:      boardModel,
		textInput:       ti,
		settingsModel:   settingsModel,
		roadmapViewport: &roadmapVP,
		tableViewport:   &tableVP,
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
		if len(a.state.Items) > 0 {
			a.state.View.FocusedItemID = a.state.Items[0].ID
		}
		a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.Project.Fields, a.state.View.Filter, a.state.View.FocusedItemID)
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
		a.boardModel = boardPkg.NewBoardModel(a.state.Items, a.state.Project.Fields, a.state.View.Filter, a.state.View.FocusedItemID)
		if !a.state.DisableNotifications {
			notif := state.Notification{Message: "Item updated successfully", Level: "info", At: time.Now(), DismissAfter: 3 * time.Second}
			a.state.Notifications = append(a.state.Notifications, notif)
			cmds = append(cmds, dismissNotificationCmd(len(a.state.Notifications)-1, notif.DismissAfter))
		}
	case DetailReadyMsg:
		a.detailPanel = components.NewDetailPanelModel(m.Item, a.state.Width, a.state.Height)
		if !a.state.DisableNotifications {
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
		// Save settings to config
		configPath, err := config.ResolvePath()
		if err == nil {
			cfg := config.Config{
				DefaultProjectID: m.ProjectID,
				DefaultOwner:     m.Owner,
			}
			if saveErr := config.Save(configPath, cfg); saveErr == nil {
				if !a.state.DisableNotifications {
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
		if !a.state.DisableNotifications {
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
	case "3", "r":
		return a.handleSwitchView(SwitchViewMsg{View: state.ViewRoadmap})
	case "4":
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
			if !a.state.DisableNotifications {
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
		return a.handleEnterEditMode(EnterEditModeMsg{})
	case "a":
		return a.handleEnterAssignMode(EnterAssignModeMsg{})
	case "w":
		if a.state.View.CurrentView == state.ViewTable {
			return a.handleEnterStatusSelectMode(EnterStatusSelectModeMsg{})
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

	editTitle := ""
	if a.state.View.Mode == "edit" {
		editTitle = a.textInput.Value()
	}
	footer := components.RenderFooter(string(a.state.View.Mode), string(a.state.View.CurrentView), width, editTitle)
	notif := components.RenderNotifications(a.state.Notifications)

	body := ""
	bodyHeight := a.bodyViewportHeight(header, footer, notif)
	frameVertical := components.FrameStyle.GetVerticalFrameSize()
	switch a.state.View.CurrentView {
	case state.ViewTable:
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
	case state.ViewRoadmap:
		statusProgress := deriveStatusProgress(a.state.Project.Fields)
		content, focusTop, focusBottom := roadmap.Render(a.state.Project.Iterations, items, a.state.View.FocusedItemID, statusProgress, innerWidth)
		contentHeight := lipgloss.Height(content)
		if contentHeight <= 0 {
			focusTop = -1
			focusBottom = -1
		} else {
			if focusTop >= contentHeight {
				focusTop = contentHeight - 1
			}
			if focusBottom >= contentHeight {
				focusBottom = contentHeight - 1
			}
			if focusBottom < focusTop {
				focusBottom = focusTop
			}
		}
		viewportHeight := bodyHeight - frameVertical
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		if a.roadmapViewport != nil {
			a.ensureRoadmapViewportSize(innerWidth, viewportHeight)
			a.roadmapViewport.SetContent(content)
			a.scrollRoadmapViewportToFocus(focusTop, focusBottom)
			body = a.roadmapViewport.View()
		} else {
			body = content
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

	// Handle edit/assign modes
	if a.state.View.Mode == "edit" || a.state.View.Mode == "assign" {
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

func applyFilter(items []state.Item, fs state.FilterState) []state.Item {
	if fs.Query == "" && len(fs.Labels) == 0 && len(fs.Assignees) == 0 && len(fs.Statuses) == 0 && len(fs.Iterations) == 0 {
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
		if len(fs.Iterations) > 0 && !state.MatchesIterationFilters(it, fs.Iterations, time.Now()) {
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
	if !a.state.DisableNotifications {
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

func (a App) handleEnterDetailMode() (tea.Model, tea.Cmd) {
	a.syncFocusedItem()

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
	if !a.state.DisableNotifications {
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

func (a App) ensureRoadmapViewportSize(width, height int) {
	if a.roadmapViewport == nil {
		return
	}
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if a.roadmapViewport.Width != width {
		a.roadmapViewport.Width = width
	}
	if a.roadmapViewport.Height != height {
		a.roadmapViewport.Height = height
	}
}

func (a App) scrollRoadmapViewportToFocus(focusTop, focusBottom int) {
	if a.roadmapViewport == nil || focusTop < 0 || focusBottom < 0 {
		return
	}
	vp := a.roadmapViewport
	visibleHeight := vp.Height - vp.Style.GetVerticalFrameSize()
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	if focusBottom < focusTop {
		focusBottom = focusTop
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
}

func deriveStatusProgress(fields []state.Field) map[string]int {
	for _, field := range fields {
		if !strings.EqualFold(field.Name, "Status") {
			continue
		}
		total := len(field.Options)
		if total == 0 {
			return nil
		}
		progress := make(map[string]int, total*2)
		divisions := total - 1
		for idx, opt := range field.Options {
			pct := 100
			if divisions > 0 {
				if idx == total-1 {
					pct = 100
				} else {
					pct = idx * (100 / divisions)
				}
			}
			storeStatusProgressKey(progress, opt.Name, pct)
			storeStatusProgressKey(progress, opt.ID, pct)
		}
		return progress
	}
	return nil
}

func storeStatusProgressKey(m map[string]int, key string, pct int) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	m[key] = pct
	lower := strings.ToLower(key)
	m[lower] = pct
}
