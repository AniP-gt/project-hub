package board

import (
	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
)

const (
	minColumnWidth      = 24
	estimatedCardHeight = 6
	minVisibleCards     = 3
)

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

func (m BoardModel) estimateCardHeight() int {
	const maxSampleSize = 10
	sampleCount := 0
	totalHeight := 0

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

	if sampleCount > 0 {
		return totalHeight / sampleCount
	}

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

func (m BoardModel) calculateMaxVisibleCards() int {
	cardHeight := m.estimateCardHeight()
	if cardHeight <= 0 {
		cardHeight = estimatedCardHeight
	}

	available := m.Height
	headerHeight := m.columnHeaderHeight("Column", false, 0)
	if available <= 0 {
		available = cardHeight*minVisibleCards + headerHeight + 2
	}

	available -= headerHeight + 2
	if available < cardHeight {
		available = cardHeight
	}

	maxCards := available / cardHeight
	if maxCards < 1 {
		maxCards = 1
	}

	return maxCards
}

func (m BoardModel) availableCardAreaHeight(headerHeight int) int {
	available := m.Height - headerHeight
	if available < 1 {
		available = 1
	}
	return available
}

func (m BoardModel) visibleCardRange(col state.Column, isFocused bool, headerHeight int) (start, end int, showAbove, showBelow bool) {
	if len(col.Cards) == 0 {
		return 0, 0, false, false
	}

	start = 0
	if isFocused {
		start = m.CardOffset
		if start < 0 {
			start = 0
		}
		if start >= len(col.Cards) {
			start = len(col.Cards) - 1
		}
	}

	remaining := m.availableCardAreaHeight(headerHeight)
	showAbove = isFocused && start > 0
	if remaining < 1 {
		remaining = 1
	}

	end = start
	for end < len(col.Cards) {
		cardHeight := lipgloss.Height(m.renderCard(col.Cards[end], isFocused && end == m.FocusedCardIndex))
		if cardHeight < 1 {
			cardHeight = 1
		}

		needBottomIndicator := isFocused && end < len(col.Cards)-1
		reserved := 0
		if needBottomIndicator {
			reserved = 1
		}

		if end > start && remaining-cardHeight < reserved {
			break
		}

		end++
		remaining -= cardHeight
		if remaining < 0 {
			break
		}
	}

	if end <= start {
		end = start + 1
		if end > len(col.Cards) {
			end = len(col.Cards)
		}
	}

	showBelow = isFocused && end < len(col.Cards)
	return start, end, showAbove, showBelow
}

func (m *BoardModel) ensureFocusedCardVisible() {
	if m.FocusedColumnIndex < 0 || m.FocusedColumnIndex >= len(m.Columns) {
		m.FocusedCardIndex = 0
		m.CardOffset = 0
		return
	}

	currentColumn := m.Columns[m.FocusedColumnIndex]
	if len(currentColumn.Cards) == 0 {
		m.FocusedCardIndex = 0
		m.CardOffset = 0
		return
	}

	if m.FocusedCardIndex < 0 {
		m.FocusedCardIndex = 0
	}
	if m.FocusedCardIndex >= len(currentColumn.Cards) {
		m.FocusedCardIndex = len(currentColumn.Cards) - 1
	}
	if m.CardOffset < 0 {
		m.CardOffset = 0
	}
	if m.CardOffset >= len(currentColumn.Cards) {
		m.CardOffset = len(currentColumn.Cards) - 1
	}

	headerHeight := m.columnHeaderHeight(currentColumn.Name, true, len(currentColumn.Cards))
	for {
		start, end, _, _ := m.visibleCardRange(currentColumn, true, headerHeight)
		if m.FocusedCardIndex < start {
			m.CardOffset = m.FocusedCardIndex
			continue
		}
		if m.FocusedCardIndex >= end {
			m.CardOffset++
			continue
		}
		m.CardOffset = start
		break
	}
}

func (m BoardModel) columnHeaderHeight(name string, isFocused bool, count int) int {
	header := m.renderColumnHeader(name, isFocused, count)
	height := lipgloss.Height(header)
	if height < 1 {
		return 1
	}
	return height
}
