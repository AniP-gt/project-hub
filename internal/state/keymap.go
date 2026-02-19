package state

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines primary Vim-like keybindings.
type KeyMap struct {
	Quit         key.Binding
	ViewBoard    key.Binding
	ViewTable    key.Binding
	MoveLeft     key.Binding
	MoveRight    key.Binding
	MoveUp       key.Binding
	MoveDown     key.Binding
	FilterMode   key.Binding
	EditMode     key.Binding
	Assign       key.Binding
	ViewDetail   key.Binding
	ToggleFields key.Binding
	GroupMode    key.Binding
}

// DefaultKeyMap returns canonical bindings per specification.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		ViewBoard:    key.NewBinding(key.WithKeys("1", "b"), key.WithHelp("1/b", "board")),
		ViewTable:    key.NewBinding(key.WithKeys("2", "t"), key.WithHelp("2/t", "table")),
		MoveLeft:     key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "left/status-")),
		MoveRight:    key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "right/status+")),
		MoveUp:       key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "up")),
		MoveDown:     key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "down")),
		FilterMode:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		EditMode:     key.NewBinding(key.WithKeys("i", "enter"), key.WithHelp("i/enter", "edit")),
		Assign:       key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "assign")),
		ViewDetail:   key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "detail")),
		ToggleFields: key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "fields")),
		GroupMode:    key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "group")),
	}
}
