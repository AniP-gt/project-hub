package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
)

type FieldSelectedMsg struct {
	OptionID   string
	OptionName string
	FieldID    string
	FieldName  string
	Canceled   bool
}

type FieldSelectorModel struct {
	focusedItem state.Item
	field       state.Field
	cursor      int
	width       int
	height      int
}

func NewFieldSelectorModel(item state.Item, field state.Field, width, height int) FieldSelectorModel {
	initialCursor := 0
	for i, option := range field.Options {
		if option.Name == getItemFieldValue(item, field.Name) {
			initialCursor = i
			break
		}
	}
	return FieldSelectorModel{
		focusedItem: item,
		field:       field,
		cursor:      initialCursor,
		width:       width,
		height:      height,
	}
}

func getItemFieldValue(item state.Item, fieldName string) string {
	switch fieldName {
	case "Status":
		return item.Status
	case "Priority":
		return item.Priority
	case "Milestone":
		return item.Milestone
	default:
		return ""
	}
}

func (m FieldSelectorModel) Init() tea.Cmd {
	return nil
}

func (m FieldSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 && len(m.field.Options) > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.field.Options)-1 && len(m.field.Options) > 0 {
				m.cursor++
			}
		case "enter":
			if len(m.field.Options) == 0 {
				return m, func() tea.Msg {
					return FieldSelectedMsg{Canceled: true}
				}
			}
			selectedOption := m.field.Options[m.cursor]
			return m, func() tea.Msg {
				return FieldSelectedMsg{
					OptionID:   selectedOption.ID,
					OptionName: selectedOption.Name,
					FieldID:    m.field.ID,
					FieldName:  m.field.Name,
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return FieldSelectedMsg{Canceled: true}
			}
		}
	}
	return m, nil
}

func (m FieldSelectorModel) View() string {
	if m.field.Name == "" {
		return "Error: Field not found."
	}

	var s strings.Builder
	s.WriteString(fmt.Sprintf("Select %s for '%s':\n\n", m.field.Name, m.focusedItem.Title))

	for i, option := range m.field.Options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursor, option.Name))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(m.width / 3).
		Height(len(m.field.Options) + 5).
		Render(s.String())
}
