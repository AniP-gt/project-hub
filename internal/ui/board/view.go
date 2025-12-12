package board

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

// BoardModel is the Bubbletea model for the Kanban board view.
type BoardModel struct {
	Columns            []state.Column
	FocusedColumnIndex int
	FocusedCardIndex   int
	Width              int
	Height             int
	ColumnWidth        int
	ColumnOffset       int
}

// Init initializes the board model.
func (m BoardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the board model.
func (m BoardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Calculate column width based on available screen width and desired number of visible columns
		// Assume 3 visible columns for now, adjust as needed.
		m.ColumnWidth = (m.Width - components.FrameStyle.GetHorizontalFrameSize() - (3 * components.ColumnContainerStyle.GetMarginRight())) / 3
		if m.ColumnWidth < 20 { // Minimum width
			m.ColumnWidth = 20
		}
		components.ColumnContainerStyle = components.ColumnContainerStyle.Width(m.ColumnWidth)

	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			if m.FocusedColumnIndex > 0 {
				m.FocusedColumnIndex--
				m.FocusedCardIndex = 0 // Reset card focus when changing columns
				// Adjust offset for scrolling
				if m.FocusedColumnIndex < m.ColumnOffset {
					m.ColumnOffset = m.FocusedColumnIndex
				}
			}
		case "l", "right":
			if m.FocusedColumnIndex < len(m.Columns)-1 {
				m.FocusedColumnIndex++
				m.FocusedCardIndex = 0 // Reset card focus
				// Adjust offset for scrolling
				if m.FocusedColumnIndex >= m.ColumnOffset+3 { // Assuming 3 visible columns
					m.ColumnOffset = m.FocusedColumnIndex - 2 // Show focused column and two after it
				}
			}
		case "j", "down":
			if m.FocusedColumnIndex >= 0 && m.FocusedColumnIndex < len(m.Columns) {
				currentColumn := m.Columns[m.FocusedColumnIndex]
				if m.FocusedCardIndex < len(currentColumn.Cards)-1 {
					m.FocusedCardIndex++
				}
			}
		case "k", "up":
			if m.FocusedColumnIndex >= 0 && m.FocusedColumnIndex < len(m.Columns) {
				if m.FocusedCardIndex > 0 {
					m.FocusedCardIndex--
				}
			}
		}
	}
	return m, nil
}

// View renders the board.
func (m BoardModel) View() string {
	var renderedColumns []string

	// Determine visible columns for horizontal scrolling
	startCol := m.ColumnOffset
	endCol := startCol + 3 // Assuming 3 columns visible at a time
	if endCol > len(m.Columns) {
		endCol = len(m.Columns)
	}
	if startCol > endCol { // Handle case where there are fewer than 3 columns
		startCol = 0
	}

	for i := startCol; i < endCol; i++ {
		col := m.Columns[i]
		var columnContent []string

		// Column title
		headerStyle := components.ColumnHeaderStyle
		if i == m.FocusedColumnIndex {
			headerStyle = headerStyle.BorderForeground(lipgloss.Color("205")) // Highlight focused column header
		}
		header := headerStyle.Render(col.Name + " (" + fmt.Sprintf("%d", len(col.Cards)) + ")")
		columnContent = append(columnContent, header)

		// Render cards
		for j, card := range col.Cards {
			isCardSelected := i == m.FocusedColumnIndex && j == m.FocusedCardIndex
			cardView := m.renderCard(card, isCardSelected)
			columnContent = append(columnContent, cardView)
		}

		// Apply column style
		currentColumnStyle := components.ColumnContainerStyle
		if i == m.FocusedColumnIndex {
			currentColumnStyle = currentColumnStyle.BorderForeground(lipgloss.Color("205")) // Highlight focused column border
		}
		renderedColumns = append(renderedColumns, currentColumnStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left, columnContent...),
		))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...)
}

// renderCard renders a single card.
func (m BoardModel) renderCard(card state.Card, isSelected bool) string {
	id := components.CardIDStyle.Render(card.ID)
	title := components.CardTitleStyle.Render(card.Title)

	// Assignee
	var assignee string
	if card.Assignee != "" {
		assignee = components.CardAssigneeStyle.Render("@" + card.Assignee)
	}

	// Labels
	var labels string
	if len(card.Labels) > 0 {
		labels = components.CardTagStyle.Render("[" + strings.Join(card.Labels, ", ") + "]")
	}

	// Priority
	var priority string
	if card.Priority != "" {
		priorityStyle := components.CardPriorityStyle
		switch card.Priority {
		case "High":
			priorityStyle = priorityStyle.Foreground(components.ColorRed400)
		case "Medium":
			priorityStyle = priorityStyle.Foreground(components.ColorYellow400)
		case "Low":
			priorityStyle = priorityStyle.Foreground(components.ColorGreen400)
		}
		priority = priorityStyle.Render(card.Priority)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, id, title, assignee, labels, priority)

	style := components.CardBaseStyle
	if isSelected {
		style = components.CardSelectedStyle
	}

	return style.Render(content)
}

// NewBoardModel creates a new BoardModel from items and filter state.
func NewBoardModel(items []state.Item, filter state.FilterState) BoardModel {
	// Apply global filter
	filteredItems := applyFilter(items, filter)

	// Group items into columns by status
	columns := groupItemsByStatus(filteredItems)

	return BoardModel{
		Columns:            columns,
		FocusedColumnIndex: 0,
		FocusedCardIndex:   0,
		ColumnOffset:       0,
	}
}

// applyFilter applies the global filter to items.
func applyFilter(items []state.Item, fs state.FilterState) []state.Item {
	if fs.Query == "" && len(fs.Labels) == 0 && len(fs.Assignees) == 0 && len(fs.Statuses) == 0 {
		return items
	}
	var out []state.Item
	for _, it := range items {
		if fs.Query != "" && !strings.Contains(strings.ToLower(it.Title), strings.ToLower(fs.Query)) {
			continue
		}
		if len(fs.Labels) > 0 && !containsAny(it.Labels, fs.Labels) {
			continue
		}
		if len(fs.Assignees) > 0 && !containsAny(it.Assignees, fs.Assignees) {
			continue
		}
		if len(fs.Statuses) > 0 && !containsAny([]string{it.Status}, fs.Statuses) {
			continue
		}
		out = append(out, it)
	}
	return out
}

// containsAny checks if any needle is in haystack.
func containsAny(haystack []string, needles []string) bool {
	for _, n := range needles {
		for _, h := range haystack {
			if h == n {
				return true
			}
		}
	}
	return false
}

// groupItemsByStatus groups items into columns by status.
func groupItemsByStatus(items []state.Item) []state.Column {
	statusMap := make(map[string][]state.Card)
	for _, item := range items {
		assignee := ""
		if len(item.Assignees) > 0 {
			assignee = item.Assignees[0] // Take first assignee
		}
		priority := "Medium" // Default priority, could be derived from labels or other fields
		if strings.Contains(strings.Join(item.Labels, " "), "high") {
			priority = "High"
		} else if strings.Contains(strings.Join(item.Labels, " "), "low") {
			priority = "Low"
		}
		card := state.Card{
			ID:       item.ID,
			Title:    item.Title,
			Assignee: assignee,
			Labels:   item.Labels,
			Status:   item.Status,
			Priority: priority,
		}
		statusMap[item.Status] = append(statusMap[item.Status], card)
	}

	// Define column order (Backlog, In Progress, Review, Done)
	columnOrder := []string{"Backlog", "In Progress", "Review", "Done"}
	var columns []state.Column
	for _, status := range columnOrder {
		if cards, exists := statusMap[status]; exists {
			columns = append(columns, state.Column{Name: status, Cards: cards})
		}
	}
	// Add any other statuses not in the predefined order
	for status, cards := range statusMap {
		found := false
		for _, col := range columns {
			if col.Name == status {
				found = true
				break
			}
		}
		if !found {
			columns = append(columns, state.Column{Name: status, Cards: cards})
		}
	}
	return columns
}
