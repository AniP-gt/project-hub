package state

import (
	"time"
)

// ViewMode represents current interaction mode.
type ViewMode string

const (
	ModeNormal       ViewMode = "normal"
	ModeFiltering    ViewMode = "filtering"
	ModeEdit         ViewMode = "edit"
	ModeSort         ViewMode = "sort"
	ModeStatusSelect ViewMode = "statusSelect" // Added for status selection mode
	ModeSettings     ViewMode = "settings"     // Added for settings view mode
)

// ViewType represents the active view.
type ViewType string

const (
	ViewBoard    ViewType = "board"
	ViewTable    ViewType = "table"
	ViewRoadmap  ViewType = "roadmap"
	ViewSettings ViewType = "settings"
)

// FilterState captures parsed filter tokens and raw query.
type FilterState struct {
	Query      string
	Labels     []string
	Assignees  []string
	Statuses   []string
	Iterations []string
}

// TableSort captures table ordering preferences.
type TableSort struct {
	Field string
	Asc   bool
}

// Timeline represents an iteration or timebox for roadmap view.
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

// ViewContext holds transient UI state.
type ViewContext struct {
	CurrentView        ViewType
	FocusedItemID      string
	FocusedIndex       int
	FocusedColumnIndex int // Added for cell-level focus in table view
	Filter             FilterState
	Mode               ViewMode
	TableSort          TableSort
}

// Notification represents a non-blocking message to the user.
type Notification struct {
	Message      string
	Level        string // info, warn, error
	At           time.Time
	Dismissed    bool
	DismissAfter time.Duration // Added: Duration after which the notification should be dismissed
}

// Card represents a project item in the Kanban board view.
type Card struct {
	ID       string
	Title    string
	Assignee string
	Labels   []string
	Status   string
	Priority string
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
	ItemLimit     int // Added
}
