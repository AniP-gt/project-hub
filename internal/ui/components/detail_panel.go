package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
)

var DetailPanelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorGray700).
	Background(ColorGray900).
	Foreground(ColorGray200)

var DetailTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorBlue300)

var DetailLabelStyle = lipgloss.NewStyle().
	Foreground(ColorGray500)

var DetailValueStyle = lipgloss.NewStyle().
	Foreground(ColorGray300)

type DetailPanelModel struct {
	item     state.Item
	width    int
	height   int
	viewport viewport.Model
}

func NewDetailPanelModel(item state.Item, width, height int) DetailPanelModel {
	vp := viewport.New(60, 20)
	vp.Style = DetailPanelStyle

	m := DetailPanelModel{
		item:   item,
		width:  width,
		height: height,
	}
	m.updateContent()
	return m
}

func (m DetailPanelModel) Init() tea.Cmd {
	return nil
}

func (m DetailPanelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return DetailCloseMsg{}
			}
		case "up", "k":
			m.viewport.ScrollUp(3)
		case "down", "j":
			m.viewport.ScrollDown(3)
		}
	}
	return m, nil
}

func (m *DetailPanelModel) updateContent() {
	var s strings.Builder

	s.WriteString(DetailTitleStyle.Render(m.item.Title))
	s.WriteString("\n\n")

	if m.item.Number > 0 && m.item.Repository != "" {
		s.WriteString(DetailLabelStyle.Render("#: "))
		s.WriteString(DetailValueStyle.Render(fmt.Sprintf("%d", m.item.Number)))
		s.WriteString(" | ")
	}
	s.WriteString(DetailLabelStyle.Render("Status: "))
	s.WriteString(DetailValueStyle.Render(m.item.Status))
	s.WriteString("\n")

	if len(m.item.Assignees) > 0 {
		s.WriteString(DetailLabelStyle.Render("Assignees: "))
		s.WriteString(DetailValueStyle.Render(strings.Join(m.item.Assignees, ", ")))
		s.WriteString("\n")
	}

	if len(m.item.Labels) > 0 {
		s.WriteString(DetailLabelStyle.Render("Labels: "))
		s.WriteString(DetailValueStyle.Render(strings.Join(m.item.Labels, ", ")))
		s.WriteString("\n")
	}

	if m.item.Priority != "" {
		s.WriteString(DetailLabelStyle.Render("Priority: "))
		s.WriteString(DetailValueStyle.Render(m.item.Priority))
		s.WriteString("\n")
	}

	if m.item.Milestone != "" {
		s.WriteString(DetailLabelStyle.Render("Milestone: "))
		s.WriteString(DetailValueStyle.Render(m.item.Milestone))
		s.WriteString("\n")
	}

	if m.item.URL != "" {
		s.WriteString(DetailLabelStyle.Render("URL: "))
		s.WriteString(DetailValueStyle.Render(m.item.URL))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(strings.Repeat("─", 40))
	s.WriteString("\n\n")

	s.WriteString(DetailLabelStyle.Render("Description"))
	s.WriteString("\n")
	if m.item.Description != "" {
		s.WriteString(DetailValueStyle.Render(m.item.Description))
	} else {
		s.WriteString(DetailValueStyle.Render("(no description)"))
	}

	m.viewport.SetContent(s.String())
}

func (m DetailPanelModel) View() string {
	panelWidth := m.width * 2 / 3
	if panelWidth < 60 {
		panelWidth = 60
	}
	panelHeight := m.height * 2 / 3
	if panelHeight < 15 {
		panelHeight = 15
	}

	content := m.viewport.View()
	styled := DetailPanelStyle.Width(panelWidth).Height(panelHeight).Render(content)
	return styled
}

type DetailCloseMsg struct{}
