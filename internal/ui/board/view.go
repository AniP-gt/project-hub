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

// BoardModel is the Bubbletea model for the Kanban board view.
type BoardModel struct {
	Columns            []state.Column
	FocusedColumnIndex int
	FocusedCardIndex   int
	Width              int
	Height             int
	ColumnWidth        int
	ColumnOffset       int
	CardOffset         int // Vertical scroll offset for cards in the focused column
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
		numVisibleColumns := 4 // Default to 4 visible columns
		if len(m.Columns) < numVisibleColumns {
			numVisibleColumns = len(m.Columns)
		}
		if numVisibleColumns == 0 {
			numVisibleColumns = 1 // Avoid division by zero if no columns
		}

		m.ColumnWidth = (m.Width - components.FrameStyle.GetHorizontalFrameSize() - (numVisibleColumns * components.ColumnContainerStyle.GetMarginRight())) / numVisibleColumns
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
				m.CardOffset = 0       // Reset card offset
				// Adjust offset for scrolling
				if m.FocusedColumnIndex < m.ColumnOffset {
					m.ColumnOffset = m.FocusedColumnIndex
				}
			}
		case "l", "right":
			numVisibleColumns := 4 // Must match the value used in WindowSizeMsg
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
	var renderedColumns []string

	// Height is managed in app.go, no need for local availableHeight

	// Determine visible columns for horizontal scrolling
	numVisibleColumns := 4 // Must match the value used in WindowSizeMsg
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
		headerStyle := components.ColumnHeaderStyle
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

	id := components.CardIDStyle.Render(card.ID)
	// Arrange ID, Title, then metadata (assignee, labels, priority) to match mock
	content := lipgloss.JoinVertical(lipgloss.Left, id, title, lipgloss.JoinHorizontal(lipgloss.Left, assignee, labels, priority))

	style := components.CardBaseStyle
	if isSelected {
		style = components.CardSelectedStyle
	}

	return style.Render(content)
}

// calculateMaxVisibleCards calculates how many cards can fit in a column based on height.
func (m BoardModel) calculateMaxVisibleCards() int {
	// Estimate cards visible based on model height. Reserve space for header and margins.
	if m.Height <= 0 {
		return 3
	}
	available := m.Height - 8 // reserve header/footer and padding
	if available <= 4 {
		return 1
	}
	// assume each card approx 3 lines tall (ID, title, meta)
	max := available / 3
	if max < 1 {
		max = 1
	}
	return max
}

// ColumnOrder defines the order of columns in the Kanban board.
var ColumnOrder = []string{"Backlog", "In Progress", "Review", "Done"}

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
