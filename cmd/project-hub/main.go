package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the current view
type ViewMode int

const (
	BoardView ViewMode = iota
	TableView
	RoadmapView
)

// Issue represents a GitHub issue
type Issue struct {
	ID       string
	Title    string
	Status   string
	Assignee string
	Priority string
	Labels   []string
	Updated  string
}

// Model holds the application state
type Model struct {
	viewMode     ViewMode
	cursor       int
	columnCursor int
	issues       []Issue
	columns      []string
	width        int
	height       int
}

// Sample data
func getSampleIssues() []Issue {
	return []Issue{
		{"#123", "ユーザー認証機能", "Backlog", "@tanaka", "High", []string{"feature", "auth"}, "2024-12-01"},
		{"#124", "API統合", "Backlog", "@sato", "Medium", []string{"api"}, "2024-12-02"},
		{"#125", "ドキュメント更新", "Backlog", "@yamada", "Low", []string{"docs"}, "2024-12-03"},
		{"#126", "TUIデザイン実装", "In Progress", "@tanaka", "High", []string{"feature", "ui"}, "2024-12-05"},
		{"#127", "キーバインド設定", "In Progress", "@yamada", "Low", []string{"config"}, "2024-12-06"},
		{"#128", "テストコード追加", "Review", "@sato", "Medium", []string{"test"}, "2024-12-04"},
		{"#129", "初期セットアップ", "Done", "@tanaka", "High", []string{"setup"}, "2024-11-28"},
		{"#130", "README作成", "Done", "@sato", "Low", []string{"docs"}, "2024-11-29"},
	}
}

func initialModel() Model {
	return Model{
		viewMode:     BoardView,
		cursor:       0,
		columnCursor: 0,
		issues:       getSampleIssues(),
		columns:      []string{"Backlog", "In Progress", "Review", "Done"},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// View switching
		case "1", "b":
			m.viewMode = BoardView
			m.cursor = 0
		case "2", "t":
			m.viewMode = TableView
			m.cursor = 0
		case "3", "r":
			m.viewMode = RoadmapView
			m.cursor = 0

		// Navigation
		case "j", "down":
			if m.cursor < len(m.issues)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "h", "left":
			if m.columnCursor > 0 {
				m.columnCursor--
			}
		case "l", "right":
			if m.columnCursor < len(m.columns)-1 {
				m.columnCursor++
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	// Header
	s.WriteString(m.renderHeader())
	s.WriteString("\n\n")

	// Main content based on view mode
	switch m.viewMode {
	case BoardView:
		s.WriteString(m.renderBoardView())
	case TableView:
		s.WriteString(m.renderTableView())
	case RoadmapView:
		s.WriteString(m.renderRoadmapView())
	}

	s.WriteString("\n\n")

	// Footer
	s.WriteString(m.renderFooter())

	return s.String()
}

func (m Model) renderHeader() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Bold(true)

	projectStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Background(lipgloss.Color("235")).
		Padding(0, 2)

	viewStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	inactiveViewStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	title := headerStyle.Render("█ GitHub Projects TUI")
	project := projectStyle.Render("Project: Web App v2.0")

	board := "[1:Board]"
	table := "[2:Table]"
	roadmap := "[3:Roadmap]"

	if m.viewMode == BoardView {
		board = viewStyle.Render(board)
	} else {
		board = inactiveViewStyle.Render(board)
	}

	if m.viewMode == TableView {
		table = viewStyle.Render(table)
	} else {
		table = inactiveViewStyle.Render(table)
	}

	if m.viewMode == RoadmapView {
		roadmap = viewStyle.Render(roadmap)
	} else {
		roadmap = inactiveViewStyle.Render(roadmap)
	}

	views := board + " " + table + " " + roadmap

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		" | ",
		project,
		strings.Repeat(" ", 20),
		views,
	)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Width(m.width).
		Render(header)
}

func (m Model) renderFooter() string {
	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("235")).
		Padding(0, 2)

	mode := modeStyle.Render("NORMAL MODE")
	help := helpStyle.Render("j/k:移動 h/l:列移動 i:編集 /:フィルタ a:アサイン 1-3:ビュー切替 q:終了")

	footer := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mode,
		strings.Repeat(" ", 10),
		help,
	)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Width(m.width).
		Render(footer)
}

func (m Model) renderBoardView() string {
	var columns []string

	for colIdx, colName := range m.columns {
		var cards []string
		count := 0

		// Get issues for this column
		for issueIdx, issue := range m.issues {
			if issue.Status == colName {
				count++
				cardStyle := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("8")).
					Padding(1).
					Width(30).
					Foreground(lipgloss.Color("7"))

				// Highlight selected card
				if m.viewMode == BoardView && m.columnCursor == colIdx && m.cursor == issueIdx {
					cardStyle = cardStyle.
						BorderForeground(lipgloss.Color("11")).
						Background(lipgloss.Color("236"))
				}

				idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
				labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
				assigneeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

				card := fmt.Sprintf("%s\n%s\n\n%s  %s",
					idStyle.Render(issue.ID),
					issue.Title,
					assigneeStyle.Render(issue.Assignee),
					labelStyle.Render(fmt.Sprintf("[%s]", strings.Join(issue.Labels, ", "))),
				)

				cards = append(cards, cardStyle.Render(card))
			}
		}

		// Column header
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			Width(32).
			Bold(true)

		header := headerStyle.Render(fmt.Sprintf("%s (%d)", colName, count))

		column := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			strings.Join(cards, "\n\n"),
		)

		columns = append(columns, column)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func (m Model) renderTableView() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Bold(true)

	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	// Header row
	headers := []string{
		headerStyle.Width(8).Render("ID"),
		headerStyle.Width(25).Render("Title"),
		headerStyle.Width(15).Render("Status"),
		headerStyle.Width(12).Render("Assignee"),
		headerStyle.Width(10).Render("Priority"),
		headerStyle.Width(12).Render("Updated"),
	}

	var rows []string
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, headers...))
	rows = append(rows, strings.Repeat("─", 90))

	// Data rows
	for idx, issue := range m.issues {
		style := cellStyle
		if m.viewMode == TableView && idx == m.cursor {
			style = selectedStyle
		}

		idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		priorityColor := "10"
		if issue.Priority == "High" {
			priorityColor = "9"
		} else if issue.Priority == "Medium" {
			priorityColor = "11"
		}

		cells := []string{
			style.Width(8).Render(idStyle.Render(issue.ID)),
			style.Width(25).Render(issue.Title),
			style.Width(15).Render(issue.Status),
			style.Width(12).Render(issue.Assignee),
			style.Width(10).Render(lipgloss.NewStyle().Foreground(lipgloss.Color(priorityColor)).Render(issue.Priority)),
			style.Width(12).Render(issue.Updated),
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderRoadmapView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Background(lipgloss.Color("236")).
		Padding(1, 2).
		Bold(true)

	title := titleStyle.Render("Timeline: Sprint 1 - Sprint 2  (2024-12-01 → 2024-12-31)")

	var items []string
	items = append(items, title, "")

	roadmapData := []struct {
		issue    Issue
		sprint   string
		progress int
	}{
		{m.issues[0], "Sprint 1", 50},
		{m.issues[3], "Sprint 1", 70},
		{m.issues[1], "Sprint 2", 0},
		{m.issues[5], "Sprint 2", 40},
	}

	for idx, data := range roadmapData {
		itemStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1, 2).
			Width(80)

		if m.viewMode == RoadmapView && idx == m.cursor {
			itemStyle = itemStyle.
				BorderForeground(lipgloss.Color("11")).
				Background(lipgloss.Color("236"))
		}

		progressBar := strings.Repeat("█", data.progress/10) + strings.Repeat("░", 10-data.progress/10)
		progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		percentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		sprintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))

		item := fmt.Sprintf("%s %s\n%s\n\n%s %s",
			data.issue.ID,
			data.issue.Title,
			sprintStyle.Render(data.sprint),
			progressStyle.Render(progressBar),
			percentStyle.Render(fmt.Sprintf("%d%%", data.progress)),
		)

		items = append(items, itemStyle.Render(item))
	}

	// Sprint overview
	overviewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1, 2).
		Width(80).
		MarginTop(1)

	overview := "Sprint Progress Overview:\n\n"
	overview += lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render("Sprint 1: ")
	overview += lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("████████░░ ")
	overview += lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("60%\n")
	overview += lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render("Sprint 2: ")
	overview += lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("██░░░░░░░░ ")
	overview += lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("20%")

	items = append(items, overviewStyle.Render(overview))

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
