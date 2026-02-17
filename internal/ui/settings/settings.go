package settings

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SaveMsg is sent when user confirms settings save
type SaveMsg struct {
	ProjectID string
	Owner     string
}

// CancelMsg is sent when user cancels settings
type CancelMsg struct{}

// SettingsModel manages the settings form UI
type SettingsModel struct {
	projectInput textinput.Model
	ownerInput   textinput.Model
	focusedField int // 0=project, 1=owner
	width        int
	height       int
}

// New creates a new SettingsModel with initial values
func New(projectID, owner string) SettingsModel {
	// Project input
	pi := textinput.New()
	pi.Placeholder = "Project ID (e.g., 9 or PVT_xxx)"
	pi.SetValue(projectID)
	pi.Focus()
	pi.CharLimit = 100
	pi.Width = 50

	// Owner input
	oi := textinput.New()
	oi.Placeholder = "Owner (username or org)"
	oi.SetValue(owner)
	oi.CharLimit = 50
	oi.Width = 50

	return SettingsModel{
		projectInput: pi,
		ownerInput:   oi,
		focusedField: 0,
	}
}

// Init initializes the model
func (m SettingsModel) Init() tea.Cmd {
	return m.projectInput.Focus()
}

// Update handles messages
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			// Switch focus to next field
			m.focusedField = (m.focusedField + 1) % 2
			if m.focusedField == 0 {
				m.projectInput.Focus()
				m.ownerInput.Blur()
			} else {
				m.projectInput.Blur()
				m.ownerInput.Focus()
			}
			return m, nil

		case tea.KeyShiftTab:
			// Switch focus to previous field
			m.focusedField = (m.focusedField - 1 + 2) % 2
			if m.focusedField == 0 {
				m.projectInput.Focus()
				m.ownerInput.Blur()
			} else {
				m.projectInput.Blur()
				m.ownerInput.Focus()
			}
			return m, nil

		case tea.KeyEnter:
			// Save settings
			return m, func() tea.Msg {
				return SaveMsg{
					ProjectID: m.projectInput.Value(),
					Owner:     m.ownerInput.Value(),
				}
			}

		case tea.KeyEscape:
			// Cancel settings
			return m, func() tea.Msg {
				return CancelMsg{}
			}
		}
	}

	// Update focused input
	if m.focusedField == 0 {
		m.projectInput, _ = m.projectInput.Update(msg)
	} else {
		m.ownerInput, _ = m.ownerInput.Update(msg)
	}

	return m, tea.Batch(cmds...)
}

// View renders the settings form
func (m SettingsModel) View() string {
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7c3aed")).
		MarginBottom(1)

	// Label style
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ca3af"))

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		MarginTop(1)

	// Build form
	var projectLabel, ownerLabel string
	if m.focusedField == 0 {
		projectLabel = labelStyle.Foreground(lipgloss.Color("#facc15")).Render("▸ Project ID:")
		ownerLabel = labelStyle.Render("  Owner:")
	} else {
		projectLabel = labelStyle.Render("  Project ID:")
		ownerLabel = labelStyle.Foreground(lipgloss.Color("#facc15")).Render("▸ Owner:")
	}

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("⚙ Settings"),
		"",
		projectLabel,
		m.projectInput.View(),
		"",
		ownerLabel,
		m.ownerInput.View(),
		"",
		helpStyle.Render("tab: switch field • enter: save • esc: cancel"),
	)

	// Center the form
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		form,
	)
}

// SetSize updates the model dimensions
func (m *SettingsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetValues returns current input values
func (m SettingsModel) GetValues() (projectID, owner string) {
	return m.projectInput.Value(), m.ownerInput.Value()
}
