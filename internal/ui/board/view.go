package board

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

const (
	minColumnWidth      = 24
	estimatedCardHeight = 6
	minVisibleCards     = 3
	maxTitleLines       = 3
	maxMetaLines        = 1
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
				// Clamp card offset to safe range
				maxCardOffset := len(currentColumn.Cards) - maxVisibleCards
				if maxCardOffset < 0 {
					maxCardOffset = 0
				}
				if m.CardOffset > maxCardOffset {
					m.CardOffset = maxCardOffset
				}
				if m.CardOffset < 0 {
					m.CardOffset = 0
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
				// Clamp card offset to safe range
				maxCardOffset := len(currentColumn.Cards) - maxVisibleCards
				if maxCardOffset < 0 {
					maxCardOffset = 0
				}
				if m.CardOffset > maxCardOffset {
					m.CardOffset = maxCardOffset
				}
				if m.CardOffset < 0 {
					m.CardOffset = 0
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
	return boardContent
}

// renderCard renders a single card.
func (m BoardModel) renderCard(card state.Card, isSelected bool) string {
	contentWidth := m.ColumnWidth - 4
	if contentWidth < 12 {
		contentWidth = 12
	}
	wrap := func(value string, maxLines int, isSelected bool) string {
		if value == "" {
			return ""
		}
		bg := components.ColorGray800
		if isSelected {
			bg = components.ColorGray700
		}
		rendered := lipgloss.NewStyle().Width(contentWidth).MaxWidth(contentWidth).Background(bg).Render(value)
		return clampRenderedLines(rendered, maxLines, contentWidth)
	}

	title := wrap(card.Title, maxTitleLines, isSelected)

	var contentBlocks []string
	if title != "" {
		contentBlocks = append(contentBlocks, title)
	}

	if card.Assignee != "" {
		assignee := wrap("@"+card.Assignee, maxMetaLines, isSelected)
		if assignee != "" {
			contentBlocks = append(contentBlocks, assignee)
		}
	}

	if len(card.Labels) > 0 {
		labels := wrap("["+strings.Join(card.Labels, ", ")+"]", maxMetaLines, isSelected)
		if labels != "" {
			contentBlocks = append(contentBlocks, labels)
		}
	}

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
		priority := wrap(priorityStyle.Render(card.Priority), maxMetaLines, isSelected)
		if priority != "" {
			contentBlocks = append(contentBlocks, priority)
		}
	}

	if len(contentBlocks) == 0 {
		contentBlocks = append(contentBlocks, "(no title)")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, contentBlocks...)

	style := components.CardBaseStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
	if isSelected {
		style = components.CardSelectedStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
	}

	return style.Render(content)
}

func clampRenderedLines(rendered string, maxLines, width int) string {
	if maxLines <= 0 {
		return rendered
	}
	lines := strings.Split(rendered, "\n")
	if len(lines) <= maxLines {
		return rendered
	}
	clamped := lines[:maxLines]
	clamped[maxLines-1] = truncateLineWithEllipsis(clamped[maxLines-1], width)
	return strings.Join(clamped, "\n")
}

func truncateLineWithEllipsis(line string, width int) string {
	ellipsis := "..."
	if width <= lipgloss.Width(ellipsis) {
		return ellipsis
	}
	trimmed := strings.TrimRight(line, " ")
	runes := []rune(trimmed)
	for lipgloss.Width(strings.TrimRight(string(runes), " "))+lipgloss.Width(ellipsis) > width && len(runes) > 0 {
		runes = runes[:len(runes)-1]
	}
	trimmed = strings.TrimRight(string(runes), " ")
	if trimmed == "" {
		return ellipsis
	}
	return trimmed + ellipsis
}

func (m BoardModel) estimateCardHeight() int {
	// Sample up to 10 cards across all columns for accurate height estimation
	const maxSampleSize = 10
	sampleCount := 0
	totalHeight := 0

	// Collect samples from across all columns
	for _, col := range m.Columns {
		for idx := 0; idx < len(col.Cards) && sampleCount < maxSampleSize; idx++ {
			height := lipgloss.Height(m.renderCard(col.Cards[idx], false))
			totalHeight += height
			sampleCount++
		}
		if sampleCount >= maxSampleSize {
			break
		}
	}

	// Return average height if we have samples
	if sampleCount > 0 {
		return totalHeight / sampleCount
	}

	// Fallback: create a sample card with typical content
	sample := state.Card{
		Title:    "Sample card height",
		Assignee: "assignee",
		Labels:   []string{"label"},
		Priority: "High",
	}
	height := lipgloss.Height(m.renderCard(sample, false))
	if height > 0 {
		return height
	}
	return estimatedCardHeight
}

// calculateMaxVisibleCards calculates how many cards can fit in a column based on height.
func (m BoardModel) calculateMaxVisibleCards() int {
	cardHeight := m.estimateCardHeight()
	if cardHeight <= 0 {
		cardHeight = estimatedCardHeight
	}

	available := m.Height
	if available <= 0 {
		available = cardHeight*minVisibleCards + columnHeaderHeight + 2
	}

	available -= columnHeaderHeight + 2
	if available < cardHeight {
		available = cardHeight
	}

	maxCards := available / cardHeight
	if maxCards < 1 {
		maxCards = 1
	}

	return maxCards
}

// ColumnOrder defines the known progression order of columns in the Kanban board.
// Unknown statuses will be inserted before Done, and Done will always be last if present.
var ColumnOrder = []string{"Todo", "Draft", "In Progress", "In_Review"}

// isDoneStatus checks if a status represents completion (case-insensitive, trimmed match for "done")
func isDoneStatus(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "done")
}

// NewBoardModel creates a new BoardModel from items, fields, and filter state.
// The fields parameter is used to determine the status column order from GitHub Projects.
func NewBoardModel(items []state.Item, fields []state.Field, filter state.FilterState, focusedItemID string) BoardModel {
	// Apply global filter
	filteredItems := applyFilter(items, filter)

	// Group items into columns by status
	columns := groupItemsByStatus(filteredItems, fields)

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

	bm := BoardModel{
		Columns:            columns,
		FocusedColumnIndex: focusedColumnIndex,
		FocusedCardIndex:   focusedCardIndex,
		ColumnOffset:       0,
		CardOffset:         0,
	}

	return bm
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
		if len(fs.Iterations) > 0 && !state.MatchesIterationFilters(it, fs.Iterations, time.Now()) {
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

// groupItemsByStatus groups items into columns by status, enforcing progression order with Done last.
// The fields parameter is used to get the status option order from GitHub Projects.
func groupItemsByStatus(items []state.Item, fields []state.Field) []state.Column {
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
			priority := item.Priority
			if priority == "" {
				priority = inferPriorityFromLabels(item.Labels)
			}
			// Don't hardcode default priority - leave empty if no priority is set
			// GitHub Projects may have custom priority names beyond High/Medium/Low
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

	// Get status option order from fields (GitHub Projects configuration)
	var statusOrder []string
	for _, field := range fields {
		if field.Name == "Status" {
			for _, opt := range field.Options {
				statusOrder = append(statusOrder, opt.Name)
			}
			break
		}
	}

	// Fall back to default order if no Status field found
	if len(statusOrder) == 0 {
		statusOrder = ColumnOrder
	}

	// Build columns: status order from GitHub Projects, unknown statuses, then Done last
	var columns []state.Column
	var doneColumn *state.Column

	// Add columns in status order from GitHub Projects
	for _, status := range statusOrder {
		if cards, exists := statusCardMap[status]; exists {
			columns = append(columns, state.Column{Name: status, Cards: cards})
			delete(statusCardMap, status)
		}
	}

	// Separate Done column and remaining unknown statuses
	var unknownStatuses []string
	for status := range statusCardMap {
		if isDoneStatus(status) {
			doneColumn = &state.Column{Name: "Done", Cards: statusCardMap[status]}
		} else {
			unknownStatuses = append(unknownStatuses, status)
		}
	}

	// Sort unknown statuses for deterministic ordering
	sort.Strings(unknownStatuses)

	// Add unknown statuses before Done
	for _, status := range unknownStatuses {
		columns = append(columns, state.Column{Name: status, Cards: statusCardMap[status]})
	}

	// Add Done column last if present
	if doneColumn != nil {
		columns = append(columns, *doneColumn)
	}

	return columns
}

func inferPriorityFromLabels(labels []string) string {
	joined := strings.ToLower(strings.Join(labels, " "))
	switch {
	case strings.Contains(joined, "high"):
		return "High"
	case strings.Contains(joined, "low"):
		return "Low"
	}
	return ""
}
