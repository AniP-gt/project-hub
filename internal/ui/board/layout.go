package board

import (
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
