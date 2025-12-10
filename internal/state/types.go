package state

import "time"

// ViewMode represents current interaction mode.
type ViewMode string

const (
	ModeNormal    ViewMode = "normal"
	ModeFiltering ViewMode = "filtering"
	ModeEdit      ViewMode = "edit"
)

// ViewType represents the active view.
type ViewType string

const (
	ViewBoard   ViewType = "board"
	ViewTable   ViewType = "table"
	ViewRoadmap ViewType = "roadmap"
)

// FilterState captures parsed filter tokens and raw query.
type FilterState struct {
	Query     string
	Labels    []string
	Assignees []string
	Statuses  []string
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
	ID          string
	Title       string
	Description string
	Status      string
	Assignees   []string
	Labels      []string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Due         *time.Time
	IterationID string
	Position    int
}

// Project metadata and available capabilities.
type Project struct {
	ID         string
	Name       string
	Views      []ViewType
	Filters    []string
	Iterations []Timeline
	UpdatedAt  *time.Time
}

// ViewContext holds transient UI state.
type ViewContext struct {
	CurrentView   ViewType
	FocusedItemID string
	FocusedIndex  int
	Filter        FilterState
	Mode          ViewMode
	TableSort     TableSort
}

// Notification represents a non-blocking message to the user.
type Notification struct {
	Message   string
	Level     string // info, warn, error
	At        time.Time
	Dismissed bool
}

// Model is the root state for Bubbletea Update/View.
type Model struct {
	Project       Project
	Items         []Item
	View          ViewContext
	Notifications []Notification
	Width         int
	Height        int
}
