package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
)

var DetailPanelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorGray700).
	Foreground(ColorGray200)

var DetailTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorBlue300)

var DetailLabelStyle = lipgloss.NewStyle().
	Foreground(ColorGray500)

var DetailValueStyle = lipgloss.NewStyle().
	Foreground(ColorGray300)

var DetailSectionTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorGray200)

var DetailCommentMetaStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorBlue300)

var DetailCommentTimeStyle = lipgloss.NewStyle().
	Foreground(ColorGray500)

var DetailCommentBodyStyle = lipgloss.NewStyle().
	Foreground(ColorGray200)

var DetailCommentBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorGray700).
	Background(ColorGray900).
	Padding(0, 1)

type DetailPanelModel struct {
	item     state.Item
	width    int
	height   int
	viewport viewport.Model
}

func NewDetailPanelModel(item state.Item, width, height int) DetailPanelModel {
	panelWidth := width * 4 / 5
	if panelWidth < 70 {
		panelWidth = 70
	}
	panelHeight := height * 3 / 4
	if panelHeight < 15 {
		panelHeight = 15
	}
	if panelHeight > height-6 {
		panelHeight = height - 6
	}

	vp := viewport.New(panelWidth-4, panelHeight-2)
	vp.Style = DetailPanelStyle

	m := DetailPanelModel{
		item:     item,
		width:    width,
		height:   height,
		viewport: vp,
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
		case "esc", "q":
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
	contentWidth := m.viewport.Width
	if contentWidth <= 0 {
		contentWidth = 60
	}

	if m.item.Title == "" {
		m.item.Title = "(No title)"
	}
	s.WriteString(DetailTitleStyle.Render(m.item.Title))
	s.WriteString("\n\n")

	if m.item.Number > 0 && m.item.Repository != "" {
		s.WriteString(DetailLabelStyle.Render("#: "))
		s.WriteString(DetailValueStyle.Render(fmt.Sprintf("%d", m.item.Number)))
		s.WriteString(" | ")
	}

	if m.item.Status == "" {
		m.item.Status = "No status"
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

	if m.item.ParentIssue != "" {
		s.WriteString(DetailLabelStyle.Render("Parent: "))
		s.WriteString(DetailValueStyle.Render(m.item.ParentIssue))
		s.WriteString("\n")
	}

	if m.item.SubIssueProgress != "" {
		s.WriteString(DetailLabelStyle.Render("Sub-issue progress: "))
		s.WriteString(DetailValueStyle.Render(m.item.SubIssueProgress))
		s.WriteString("\n")
	}

	if len(m.item.SubIssueTitles) > 0 {
		s.WriteString(DetailLabelStyle.Render("Sub-issues:"))
		s.WriteString("\n")
		for _, title := range m.item.SubIssueTitles {
			s.WriteString(DetailValueStyle.Render("- " + title))
			s.WriteString("\n")
		}
	}

	if m.item.URL != "" {
		s.WriteString(DetailLabelStyle.Render("URL: "))
		s.WriteString(DetailValueStyle.Render(m.item.URL))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(strings.Repeat("─", 40))
	s.WriteString("\n\n")

	s.WriteString(DetailSectionTitleStyle.Render("Description"))
	s.WriteString("\n")
	if m.item.Description != "" {
		descriptionStyle := DetailValueStyle.Copy().Width(contentWidth).MaxWidth(contentWidth)
		s.WriteString(descriptionStyle.Render(m.item.Description))
	} else {
		s.WriteString(DetailValueStyle.Render("(no description)"))
	}

	s.WriteString("\n\n")
	s.WriteString(DetailSectionTitleStyle.Render(fmt.Sprintf("Comments (%d)", len(m.item.Comments))))
	s.WriteString("\n")
	if len(m.item.Comments) == 0 {
		s.WriteString(DetailValueStyle.Render("(no comments)"))
	} else {
		for i, comment := range m.item.Comments {
			if i > 0 {
				s.WriteString("\n\n")
			}
			s.WriteString(renderDetailComment(comment, contentWidth))
		}
	}

	m.viewport.SetContent(s.String())
}

func (m DetailPanelModel) View() string {
	panelWidth := m.width * 4 / 5
	if panelWidth < 70 {
		panelWidth = 70
	}
	panelHeight := m.height * 3 / 4
	if panelHeight < 15 {
		panelHeight = 15
	}
	if panelHeight > m.height-6 {
		panelHeight = m.height - 6
	}

	m.viewport.Width = panelWidth - 4
	m.viewport.Height = panelHeight - 2
	m.updateContent()

	content := m.viewport.View()
	styled := DetailPanelStyle.Width(panelWidth).Height(panelHeight).Render(content)
	return styled
}

func renderDetailComment(comment state.Comment, width int) string {
	boxWidth := width
	if boxWidth < 20 {
		boxWidth = 20
	}
	innerWidth := boxWidth - DetailCommentBoxStyle.GetHorizontalFrameSize()
	if innerWidth < 10 {
		innerWidth = 10
	}

	author := strings.TrimSpace(comment.Author)
	if author == "" {
		author = "unknown"
	}
	meta := DetailCommentMetaStyle.Render("@" + author)
	if comment.CreatedAt != nil {
		meta = lipgloss.JoinHorizontal(lipgloss.Left, meta, DetailCommentTimeStyle.Render("  "+comment.CreatedAt.In(time.Local).Format("2006-01-02 15:04")))
	}
	body := strings.TrimSpace(comment.Body)
	if body == "" {
		body = "(empty comment)"
	}
	bodyView := DetailCommentBodyStyle.Copy().Width(innerWidth).MaxWidth(innerWidth).Render(body)

	return DetailCommentBoxStyle.Copy().Width(boxWidth).MaxWidth(boxWidth).Render(lipgloss.JoinVertical(lipgloss.Left, meta, "", bodyView))
}

type DetailCloseMsg struct{}
