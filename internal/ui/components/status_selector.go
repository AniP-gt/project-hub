package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
)

// StatusSelectedMsg is sent when a status option is chosen by the user.
type StatusSelectedMsg struct {
	OptionID      string
	OptionName    string
	StatusFieldID string // New: ID of the status field itself
	Canceled      bool
}

// StatusSelectorModel is the model for the status selection UI.
type StatusSelectorModel struct {
	focusedItem state.Item
	statusField state.Field
	cursor      int // which status option is selected
	width       int
	height      int
}

// NewStatusSelectorModel creates a new model for the status selection UI.
func NewStatusSelectorModel(item state.Item, statusField state.Field, width, height int) StatusSelectorModel {
	// Initialize cursor to the currently selected status of the item, if found.
	initialCursor := 0
	for i, option := range statusField.Options {
		if option.Name == item.Status {
			initialCursor = i
			break
		}
	}
	return StatusSelectorModel{
		focusedItem: item,
		statusField: statusField,
		cursor:      initialCursor,
		width:       width,
		height:      height,
	}
}

func (m StatusSelectorModel) Init() tea.Cmd {
	return nil
}

func (m StatusSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.statusField.Options)-1 {
				m.cursor++
			}
		case "enter":
			selectedOption := m.statusField.Options[m.cursor]
			return m, func() tea.Msg {
				return StatusSelectedMsg{
					OptionID:      selectedOption.ID,
					OptionName:    selectedOption.Name,
					StatusFieldID: m.statusField.ID, // Set the status field ID
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return StatusSelectedMsg{Canceled: true}
			}
		}
	}
	return m, nil
}

func (m StatusSelectorModel) View() string {
	if m.statusField.Name == "" {
		return "Error: Status field not found."
	}

	var s strings.Builder
	s.WriteString(fmt.Sprintf("Select status for '%s':\n\n", m.focusedItem.Title))

	for i, option := range m.statusField.Options {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursor, option.Name))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(m.width / 3).                     // Adjust width as needed
		Height(len(m.statusField.Options) + 5). // Adjust height as needed
		Render(s.String())
}
