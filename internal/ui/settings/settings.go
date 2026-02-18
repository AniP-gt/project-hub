package settings

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SaveMsg is sent when user confirms settings save
type SaveMsg struct {
	ProjectID     string
	Owner         string
	SuppressHints bool
	ItemLimit     int
	ExcludeDone   bool
}

// CancelMsg is sent when user cancels settings
type CancelMsg struct{}

// SettingsModel manages the settings form UI
type SettingsModel struct {
	projectInput              textinput.Model
	ownerInput                textinput.Model
	itemLimitInput            textinput.Model
	excludeDoneInput          textinput.Model
	disableNotificationsInput textinput.Model
	focusedField              int // 0=project, 1=owner, 2=itemLimit, 3=excludeDone, 4=disableNotifications
	width                     int
	height                    int
}

// New creates a new SettingsModel with initial values
func New(projectID, owner string, disableNotifications bool, itemLimit int, excludeDone bool) SettingsModel {
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

	// Item limit input (GitHub GraphQL API max: 100)
	li := textinput.New()
	li.Placeholder = "Item limit (max 100)"
	li.SetValue(fmt.Sprintf("%d", itemLimit))
	li.CharLimit = 10
	li.Width = 50

	// Exclude done input (toggle)
	ed := textinput.New()
	ed.Placeholder = "Exclude done items (y/n)"
	if excludeDone {
		ed.SetValue("y")
	} else {
		ed.SetValue("n")
	}
	ed.CharLimit = 1
	ed.Width = 50

	// Suppress hints (y/n toggle)
	dn := textinput.New()
	dn.Placeholder = "Suppress hints (y/n)"
	if disableNotifications {
		dn.SetValue("y")
	} else {
		dn.SetValue("n")
	}
	dn.CharLimit = 1
	dn.Width = 50

	return SettingsModel{
		projectInput:              pi,
		ownerInput:                oi,
		itemLimitInput:            li,
		excludeDoneInput:          ed,
		disableNotificationsInput: dn,
		focusedField:              0,
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
			m.focusedField = (m.focusedField + 1) % 5
			m.focusField(m.focusedField)
			return m, nil

		case tea.KeyShiftTab:
			m.focusedField = (m.focusedField - 1 + 5) % 5
			m.focusField(m.focusedField)
			return m, nil

		case tea.KeySpace:
			if m.focusedField == 3 {
				current := m.excludeDoneInput.Value()
				if current == "y" {
					m.excludeDoneInput.SetValue("n")
				} else {
					m.excludeDoneInput.SetValue("y")
				}
			} else if m.focusedField == 4 {
				current := m.disableNotificationsInput.Value()
				if current == "y" {
					m.disableNotificationsInput.SetValue("n")
				} else {
					m.disableNotificationsInput.SetValue("y")
				}
			}
			return m, nil

		case tea.KeyEnter:
			itemLimit := 100
			if val := m.itemLimitInput.Value(); val != "" {
				if parsed, err := fmt.Sscanf(val, "%d", &itemLimit); err != nil || parsed == 0 {
					itemLimit = 100
				}
			}
			excludeDone := m.excludeDoneInput.Value() == "y"
			disableNotifications := m.disableNotificationsInput.Value() == "y"
			return m, func() tea.Msg {
				return SaveMsg{
					ProjectID:     m.projectInput.Value(),
					Owner:         m.ownerInput.Value(),
					ItemLimit:     itemLimit,
					ExcludeDone:   excludeDone,
					SuppressHints: disableNotifications,
				}
			}

		case tea.KeyEscape:
			return m, func() tea.Msg {
				return CancelMsg{}
			}
		}
	}

	m.updateFocusedInput(msg)
	return m, tea.Batch(cmds...)
}

func (m *SettingsModel) focusField(idx int) {
	m.projectInput.Blur()
	m.ownerInput.Blur()
	m.itemLimitInput.Blur()
	m.excludeDoneInput.Blur()
	m.disableNotificationsInput.Blur()

	switch idx {
	case 0:
		m.projectInput.Focus()
	case 1:
		m.ownerInput.Focus()
	case 2:
		m.itemLimitInput.Focus()
	case 3:
		m.excludeDoneInput.Focus()
	case 4:
		m.disableNotificationsInput.Focus()
	}
}

func (m *SettingsModel) updateFocusedInput(msg tea.Msg) {
	switch m.focusedField {
	case 0:
		m.projectInput, _ = m.projectInput.Update(msg)
	case 1:
		m.ownerInput, _ = m.ownerInput.Update(msg)
	case 2:
		m.itemLimitInput, _ = m.itemLimitInput.Update(msg)
	case 3:
		m.excludeDoneInput, _ = m.excludeDoneInput.Update(msg)
	case 4:
		m.disableNotificationsInput, _ = m.disableNotificationsInput.Update(msg)
	}
}

// View renders the settings form
func (m SettingsModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7c3aed")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ca3af"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		MarginTop(1)

	labels := make([]string, 5)
	for i := range labels {
		labels[i] = labelStyle.Render("  ")
	}
	labels[m.focusedField] = labelStyle.Foreground(lipgloss.Color("#facc15")).Render("▸ ")

	projectLabel := labels[0] + "Project ID:"
	ownerLabel := labels[1] + "Owner:"
	itemLimitLabel := labels[2] + "Item Limit:"
	excludeDoneLabel := labels[3] + "Exclude Done:"
	suppressHintsLabel := labels[4] + "Suppress Hints:"

	helpText := "tab: switch field • space: toggle y/n • enter: save • esc: cancel"

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
		itemLimitLabel,
		m.itemLimitInput.View(),
		"",
		excludeDoneLabel,
		m.excludeDoneInput.View(),
		"",
		suppressHintsLabel,
		m.disableNotificationsInput.View(),
		"",
		helpStyle.Render(helpText),
	)

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
func (m SettingsModel) GetValues() (projectID, owner string, suppressHints bool, itemLimit int, excludeDone bool) {
	itemLimit = 100
	if val := m.itemLimitInput.Value(); val != "" {
		if parsed, err := fmt.Sscanf(val, "%d", &itemLimit); err != nil || parsed == 0 {
			itemLimit = 100
		}
	}
	excludeDone = m.excludeDoneInput.Value() == "y"
	suppressHints = m.disableNotificationsInput.Value() == "y"
	return m.projectInput.Value(), m.ownerInput.Value(), suppressHints, itemLimit, excludeDone
}
