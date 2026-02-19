package state

import (
	"time"
)

// ViewMode represents current interaction mode.
type ViewMode string

const (
	ModeNormal          ViewMode = "normal"
	ModeFiltering       ViewMode = "filtering"
	ModeEdit            ViewMode = "edit"
	ModeSort            ViewMode = "sort"
	ModeStatusSelect    ViewMode = "statusSelect"
	ModeLabelSelect     ViewMode = "labelSelect"
	ModeMilestoneSelect ViewMode = "milestoneSelect"
	ModePrioritySelect  ViewMode = "prioritySelect"
	ModeSettings        ViewMode = "settings"
	ModeDetail          ViewMode = "detail"
	ModeFieldToggle     ViewMode = "fieldToggle"
)

// ViewType represents the active view.
type ViewType string

const (
	ViewBoard    ViewType = "board"
	ViewTable    ViewType = "table"
	ViewSettings ViewType = "settings"
)

// FilterState captures parsed filter tokens and raw query.
type FilterState struct {
	Raw          string
	Query        string
	Labels       []string
	Assignees    []string
	Statuses     []string
	Iterations   []string
	GroupBy      string
	FieldFilters map[string][]string
}

// TableSort captures table ordering preferences.
type TableSort struct {
	Field string
	Asc   bool
}

// Timeline represents an iteration or timebox.
type Timeline struct {
	ID       string
	Name     string
	Start    *time.Time
	End      *time.Time
	Progress string
}

// Item is a single project card/issue in any view.
type Item struct {
	ID                    string
	ContentID             string // Added: ID of the underlying content (e.g., DI_ for draft issues)
	Type                  string // Type of content, e.g., "Issue", "PullRequest", "DraftIssue"
	Title                 string
	Description           string
	Status                string
	Repository            string
	Number                int    // Issue or PR number
	URL                   string // URL to the issue or PR
	Assignees             []string
	Labels                []string
	Milestone             string
	Priority              string
	CreatedAt             *time.Time
	UpdatedAt             *time.Time
	Due                   *time.Time
	IterationID           string
	IterationName         string
	IterationStart        *time.Time
	IterationDurationDays int
	Position              int
	SubIssueProgress      string // e.g., "2/5" showing completed/total sub-issues
	ParentIssue           string // Parent issue title or reference
	FieldValues           map[string][]string
}

// Project metadata and available capabilities.
type Project struct {
	ID         string
	NodeID     string
	Owner      string
	Name       string
	Views      []ViewType
	Filters    []string
	Iterations []Timeline
	Fields     []Field // Added to store project fields like "Status"
	UpdatedAt  *time.Time
}

// Field represents a project field (e.g., "Status").
type Field struct {
	ID      string
	Name    string
	Options []Option
}

// Option represents a selectable option for a field (e.g., "Todo", "In Progress").
type Option struct {
	ID   string
	Name string
}

// Column indices for table view
const (
	ColumnTitle            = 0
	ColumnStatus           = 1
	ColumnRepository       = 2
	ColumnLabels           = 3
	ColumnMilestone        = 4
	ColumnSubIssueProgress = 5
	ColumnParentIssue      = 6
	ColumnPriority         = 7
	ColumnAssignees        = 8
	ColumnCount            = 9
)

// ViewContext holds transient UI state.
type ViewContext struct {
	CurrentView         ViewType
	FocusedItemID       string
	FocusedIndex        int
	FocusedColumnIndex  int // Added for cell-level focus in table view
	Filter              FilterState
	Mode                ViewMode
	TableSort           TableSort
	TableGroupBy        string
	CardFieldVisibility CardFieldVisibility
}

// Notification represents a non-blocking message to the user.
type Notification struct {
	Message      string
	Level        string // info, warn, error
	At           time.Time
	Dismissed    bool
	DismissAfter time.Duration // Added: Duration after which the notification should be dismissed
}

// CardFieldVisibility controls which optional fields are displayed on cards.
type CardFieldVisibility struct {
	ShowMilestone        bool
	ShowRepository       bool
	ShowSubIssueProgress bool
	ShowParentIssue      bool
	ShowLabels           bool
}

// DefaultCardFieldVisibility returns the default visibility settings.
func DefaultCardFieldVisibility() CardFieldVisibility {
	return CardFieldVisibility{
		ShowMilestone:        false,
		ShowRepository:       false,
		ShowSubIssueProgress: false,
		ShowParentIssue:      false,
		ShowLabels:           true,
	}
}

// Card represents a project item in the Kanban board view.
type Card struct {
	ID               string
	Title            string
	Assignee         string
	Labels           []string
	Status           string
	Priority         string
	Milestone        string
	Repository       string
	SubIssueProgress string // e.g., "2/5" showing completed/total sub-issues
	ParentIssue      string // Parent issue title or reference
}

// Column represents a column in the Kanban board.
type Column struct {
	Name  string
	Cards []Card
}

// Model is the root state for Bubbletea Update/View.
type Model struct {
	Project       Project
	Items         []Item
	View          ViewContext
	Notifications []Notification
	Width         int
	Height        int
	ItemLimit     int
	SuppressHints bool
	ExcludeDone   bool
}
