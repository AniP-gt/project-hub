package board

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

const (
	minColumnWidth      = 24
	estimatedCardHeight = 6
	minVisibleCards     = 3
)

var columnHeaderHeight = lipgloss.Height(components.ColumnHeaderStyle.Render("Column"))

// BoardModel is the Bubbletea model for the Kanban board view.
type BoardModel struct {
	Columns            []state.Column
	FocusedColumnIndex int
	FocusedCardIndex   int
	Width              int
	Height             int
	ColumnWidth        int
	VisibleColumns     int
	ColumnOffset       int
	CardOffset         int // Vertical scroll offset for cards in the focused column
}

func (m *BoardModel) ensureLayoutConstraints() {
	availableWidth := m.Width
	if availableWidth <= 0 {
		availableWidth = minColumnWidth
	}
	columnCount := len(m.Columns)
	if columnCount == 0 {
		columnCount = 1
	}
	maxVisible := availableWidth / minColumnWidth
	if maxVisible < 1 {
		maxVisible = 1
	}
	if maxVisible > columnCount {
		maxVisible = columnCount
	}
	m.VisibleColumns = maxVisible
	if m.VisibleColumns < 1 {
		m.VisibleColumns = 1
	}
	columnWidth := availableWidth / m.VisibleColumns
	if columnWidth < minColumnWidth {
		columnWidth = minColumnWidth
	}
	m.ColumnWidth = columnWidth

	if len(m.Columns) == 0 {
		m.FocusedColumnIndex = 0
		m.ColumnOffset = 0
		return
	}
	if m.FocusedColumnIndex >= len(m.Columns) {
		m.FocusedColumnIndex = len(m.Columns) - 1
	}
	if m.FocusedColumnIndex < 0 {
		m.FocusedColumnIndex = 0
	}
	maxOffset := len(m.Columns) - m.VisibleColumns
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.ColumnOffset > maxOffset {
		m.ColumnOffset = maxOffset
	}
	if m.ColumnOffset < 0 {
		m.ColumnOffset = 0
	}
}

func (m BoardModel) visibleColumnCount() int {
	if m.VisibleColumns > 0 {
		return m.VisibleColumns
	}
	if len(m.Columns) == 0 {
		return 1
	}
	return len(m.Columns)
}

func (m *BoardModel) EnsureLayout() {
	m.ensureLayoutConstraints()
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
		m.ensureLayoutConstraints()

	case tea.KeyMsg:
		m.ensureLayoutConstraints()
		switch msg.String() {
		case "h", "left":
			if m.FocusedColumnIndex > 0 {
				m.FocusedColumnIndex--
				m.FocusedCardIndex = 0 // Reset card focus when changing columns
				m.CardOffset = 0       // Reset card offset
				// Adjust offset for scrolling
				if m.FocusedColumnIndex < m.ColumnOffset {
					m.ColumnOffset = m.FocusedColumnIndex
				}
			}
		case "l", "right":
			numVisibleColumns := m.visibleColumnCount()
			if m.FocusedColumnIndex < len(m.Columns)-1 {
				m.FocusedColumnIndex++
				m.FocusedCardIndex = 0 // Reset card focus
				m.CardOffset = 0       // Reset card offset
				// Adjust offset for scrolling
				if m.FocusedColumnIndex >= m.ColumnOffset+numVisibleColumns {
					m.ColumnOffset = m.FocusedColumnIndex - (numVisibleColumns - 1)
				}
			}
		case "j", "down":
			if m.FocusedColumnIndex >= 0 && m.FocusedColumnIndex < len(m.Columns) {
				currentColumn := m.Columns[m.FocusedColumnIndex]
				maxVisibleCards := m.calculateMaxVisibleCards()
				if len(currentColumn.Cards) > maxVisibleCards {
					// Use relative index for scrolling
					visibleCardIndex := m.FocusedCardIndex - m.CardOffset
					if visibleCardIndex < maxVisibleCards-1 && m.FocusedCardIndex < len(currentColumn.Cards)-1 {
						m.FocusedCardIndex++
					} else if m.FocusedCardIndex < len(currentColumn.Cards)-1 {
						// Scroll down
						m.CardOffset++
						m.FocusedCardIndex++
					}
				} else {
					// No scrolling needed
					if m.FocusedCardIndex < len(currentColumn.Cards)-1 {
						m.FocusedCardIndex++
					}
				}
				// Ensure focused card is within bounds
				if m.FocusedCardIndex >= len(currentColumn.Cards) {
					m.FocusedCardIndex = len(currentColumn.Cards) - 1
				}
			}
		case "k", "up":
			if m.FocusedColumnIndex >= 0 && m.FocusedColumnIndex < len(m.Columns) {
				maxVisibleCards := m.calculateMaxVisibleCards()
				currentColumn := m.Columns[m.FocusedColumnIndex]
				if len(currentColumn.Cards) > maxVisibleCards {
					// Use relative index for scrolling
					visibleCardIndex := m.FocusedCardIndex - m.CardOffset
					if visibleCardIndex > 0 && m.FocusedCardIndex > 0 {
						m.FocusedCardIndex--
					} else if m.FocusedCardIndex > 0 {
						// Scroll up
						m.CardOffset--
						m.FocusedCardIndex--
					}
				} else {
					// No scrolling needed
					if m.FocusedCardIndex > 0 {
						m.FocusedCardIndex--
					}
				}
				// Ensure focused card is within bounds
				if m.FocusedCardIndex < 0 {
					m.FocusedCardIndex = 0
				}
			}
		}
	}
	return m, nil
}

// View renders the board.
func (m BoardModel) View() string {
	m.ensureLayoutConstraints()
	var renderedColumns []string

	// Height is managed in app.go, no need for local availableHeight

	// Determine visible columns for horizontal scrolling
	numVisibleColumns := m.visibleColumnCount()
	startCol := m.ColumnOffset
	endCol := startCol + numVisibleColumns
	if endCol > len(m.Columns) {
		endCol = len(m.Columns)
	}
	if startCol > endCol { // Handle case where there are fewer than numVisibleColumns
		startCol = 0
	}

	for i := startCol; i < endCol; i++ {
		col := m.Columns[i]
		var columnContent []string

		// Column title
		headerStyle := components.ColumnHeaderStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
		if i == m.FocusedColumnIndex {
			headerStyle = headerStyle.BorderForeground(lipgloss.Color("205")) // Highlight focused column header
		}
		header := headerStyle.Render(col.Name + " (" + fmt.Sprintf("%d", len(col.Cards)) + ")")
		columnContent = append(columnContent, header)

		// Render cards with vertical scrolling
		maxVisibleCards := m.calculateMaxVisibleCards()
		startCard := 0
		endCard := len(col.Cards)
		if len(col.Cards) > maxVisibleCards {
			if i == m.FocusedColumnIndex {
				startCard = m.CardOffset
				endCard = startCard + maxVisibleCards
				if endCard > len(col.Cards) {
					endCard = len(col.Cards)
				}
			} else {
				// For non-focused columns, show first maxVisibleCards
				endCard = maxVisibleCards
			}
		}

		for j := startCard; j < endCard; j++ {
			isCardSelected := i == m.FocusedColumnIndex && j == m.FocusedCardIndex
			cardView := m.renderCard(col.Cards[j], isCardSelected)
			columnContent = append(columnContent, cardView)
		}

		// Add scroll indicators if needed
		if len(col.Cards) > maxVisibleCards {
			if i == m.FocusedColumnIndex {
				if m.CardOffset > 0 {
					columnContent = append([]string{"↑"}, columnContent...)
				}
				if endCard < len(col.Cards) {
					columnContent = append(columnContent, "↓")
				}
			} else {
				// For non-focused columns, show "..." if there are more cards
				if len(col.Cards) > maxVisibleCards {
					columnContent = append(columnContent, "...")
				}
			}
		}

		// Apply column style
		currentColumnStyle := components.ColumnContainerStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
		if i == m.FocusedColumnIndex {
			currentColumnStyle = currentColumnStyle.BorderForeground(lipgloss.Color("205")) // Highlight focused column border
		}
		renderedColumns = append(renderedColumns, currentColumnStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left, columnContent...),
		))
	}

	boardContent := lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...)
	if m.Width > 0 {
		boardContent = lipgloss.PlaceHorizontal(m.Width, lipgloss.Left, boardContent)
	}
	return boardContent
}

// renderCard renders a single card.
func (m BoardModel) renderCard(card state.Card, isSelected bool) string {
	contentWidth := m.ColumnWidth - 4
	if contentWidth < 12 {
		contentWidth = 12
	}
	wrap := func(value string) string {
		if value == "" {
			return ""
		}
		return lipgloss.NewStyle().Width(contentWidth).MaxWidth(contentWidth).Render(value)
	}

	title := wrap(card.Title)

	// Assignee
	var assignee string
	if card.Assignee != "" {
		assignee = wrap("@" + card.Assignee)
	}

	// Labels
	var labels string
	if len(card.Labels) > 0 {
		labels = wrap("[" + strings.Join(card.Labels, ", ") + "]")
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
		priority = wrap(priorityStyle.Render(card.Priority))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, assignee, labels, priority)

	style := components.CardBaseStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
	if isSelected {
		style = components.CardSelectedStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
	}

	return style.Render(content)
}

// calculateMaxVisibleCards calculates how many cards can fit in a column based on height.
func (m BoardModel) calculateMaxVisibleCards() int {
	available := m.Height
	if available <= 0 {
		available = estimatedCardHeight * minVisibleCards
	}
	available -= columnHeaderHeight + 2
	if available < estimatedCardHeight*minVisibleCards {
		available = estimatedCardHeight * minVisibleCards
	}
	maxCards := available / estimatedCardHeight
	if maxCards < minVisibleCards {
		maxCards = minVisibleCards
	}
	return maxCards
}

// ColumnOrder defines the order of columns in the Kanban board.
var ColumnOrder = []string{"Todo", "In Progress", "In_Review", "Done", "Unknown"}

// NewBoardModel creates a new BoardModel from items and filter state.
func NewBoardModel(items []state.Item, filter state.FilterState, focusedItemID string) BoardModel {
	// Apply global filter
	filteredItems := applyFilter(items, filter)

	// Group items into columns by status
	columns := groupItemsByStatus(filteredItems)

	// Find the focused item and set initial focus
	focusedColumnIndex := 0
	focusedCardIndex := 0
	for colIdx, col := range columns {
		for cardIdx, card := range col.Cards {
			if card.ID == focusedItemID {
				focusedColumnIndex = colIdx
				focusedCardIndex = cardIdx
				break
			}
		}
	}

	return BoardModel{
		Columns:            columns,
		FocusedColumnIndex: focusedColumnIndex,
		FocusedCardIndex:   focusedCardIndex,
		ColumnOffset:       0,
		CardOffset:         0,
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
	statusMap := make(map[string][]state.Item)
	for _, item := range items {
		statusMap[item.Status] = append(statusMap[item.Status], item)
	}

	// Sort items by Position within each status
	for status, items := range statusMap {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Position < items[j].Position
		})
		statusMap[status] = items
	}

	// Convert to cards
	statusCardMap := make(map[string][]state.Card)
	for status, items := range statusMap {
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
			statusCardMap[status] = append(statusCardMap[status], card)
		}
	}

	// Define column order (Backlog, In Progress, Review, Done)
	var columns []state.Column
	for _, status := range ColumnOrder { // Use exported ColumnOrder
		if cards, exists := statusCardMap[status]; exists {
			columns = append(columns, state.Column{Name: status, Cards: cards})
		}
	}
	// Add any other statuses not in the predefined order
	for status, cards := range statusCardMap {
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
