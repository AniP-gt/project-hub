package board

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/ui/components"
)

func (m BoardModel) Init() tea.Cmd {
	return nil
}

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
				m.FocusedCardIndex = 0
				m.CardOffset = 0
				if m.FocusedColumnIndex < m.ColumnOffset {
					m.ColumnOffset = m.FocusedColumnIndex
				}
			}
		case "l", "right":
			numVisibleColumns := m.visibleColumnCount()
			if m.FocusedColumnIndex < len(m.Columns)-1 {
				m.FocusedColumnIndex++
				m.FocusedCardIndex = 0
				m.CardOffset = 0
				if m.FocusedColumnIndex >= m.ColumnOffset+numVisibleColumns {
					m.ColumnOffset = m.FocusedColumnIndex - (numVisibleColumns - 1)
				}
			}
		case "j", "down":
			if m.FocusedColumnIndex >= 0 && m.FocusedColumnIndex < len(m.Columns) {
				currentColumn := m.Columns[m.FocusedColumnIndex]
				maxVisibleCards := m.calculateMaxVisibleCards()
				if len(currentColumn.Cards) > maxVisibleCards {
					visibleCardIndex := m.FocusedCardIndex - m.CardOffset
					if visibleCardIndex < maxVisibleCards-1 && m.FocusedCardIndex < len(currentColumn.Cards)-1 {
						m.FocusedCardIndex++
					} else if m.FocusedCardIndex < len(currentColumn.Cards)-1 {
						m.CardOffset++
						m.FocusedCardIndex++
					}
				} else {
					if m.FocusedCardIndex < len(currentColumn.Cards)-1 {
						m.FocusedCardIndex++
					}
				}
				if m.FocusedCardIndex >= len(currentColumn.Cards) {
					m.FocusedCardIndex = len(currentColumn.Cards) - 1
				}
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
					visibleCardIndex := m.FocusedCardIndex - m.CardOffset
					if visibleCardIndex > 0 && m.FocusedCardIndex > 0 {
						m.FocusedCardIndex--
					} else if m.FocusedCardIndex > 0 {
						m.CardOffset--
						m.FocusedCardIndex--
					}
				} else {
					if m.FocusedCardIndex > 0 {
						m.FocusedCardIndex--
					}
				}
				if m.FocusedCardIndex < 0 {
					m.FocusedCardIndex = 0
				}
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

func (m BoardModel) View() string {
	m.ensureLayoutConstraints()
	var renderedColumns []string

	numVisibleColumns := m.visibleColumnCount()
	startCol := m.ColumnOffset
	endCol := startCol + numVisibleColumns
	if endCol > len(m.Columns) {
		endCol = len(m.Columns)
	}
	if startCol > endCol {
		startCol = 0
	}

	for i := startCol; i < endCol; i++ {
		col := m.Columns[i]
		var columnContent []string

		headerStyle := components.ColumnHeaderStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
		if i == m.FocusedColumnIndex {
			headerStyle = headerStyle.BorderForeground(lipgloss.Color("205"))
		}
		header := headerStyle.Render(col.Name + " (" + fmt.Sprintf("%d", len(col.Cards)) + ")")
		columnContent = append(columnContent, header)

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
				endCard = maxVisibleCards
			}
		}

		for j := startCard; j < endCard; j++ {
			isCardSelected := i == m.FocusedColumnIndex && j == m.FocusedCardIndex
			cardView := m.renderCard(col.Cards[j], isCardSelected)
			columnContent = append(columnContent, cardView)
		}

		if len(col.Cards) > maxVisibleCards {
			if i == m.FocusedColumnIndex {
				if m.CardOffset > 0 {
					columnContent = append([]string{"↑"}, columnContent...)
				}
				if endCard < len(col.Cards) {
					columnContent = append(columnContent, "↓")
				}
			} else {
				if len(col.Cards) > maxVisibleCards {
					columnContent = append(columnContent, "...")
				}
			}
		}

		currentColumnStyle := components.ColumnContainerStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
		if i == m.FocusedColumnIndex {
			currentColumnStyle = currentColumnStyle.BorderForeground(lipgloss.Color("205"))
		}
		renderedColumns = append(renderedColumns, currentColumnStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left, columnContent...),
		))
	}

	boardContent := lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...)
	return boardContent
}
