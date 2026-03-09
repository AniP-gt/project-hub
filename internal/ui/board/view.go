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
				if m.FocusedCardIndex < len(currentColumn.Cards)-1 {
					m.FocusedCardIndex++
					m.ensureFocusedCardVisible()
				}
			}
		case "k", "up":
			if m.FocusedColumnIndex >= 0 && m.FocusedColumnIndex < len(m.Columns) {
				if m.FocusedCardIndex > 0 {
					m.FocusedCardIndex--
					m.ensureFocusedCardVisible()
				}
				// Support 'O' to open the focused issue in the browser and 'y' to copy the URL.
				// The app-level update.HandleKey will handle these keys and issue side-effect commands.
			}
		}
	}
	return m, nil
}

func (m BoardModel) View() string {
	m.ensureLayoutConstraints()
	m.ensureFocusedCardVisible()
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

		header := m.renderColumnHeader(col.Name, i == m.FocusedColumnIndex, len(col.Cards))
		headerHeight := lipgloss.Height(header)
		columnContent = append(columnContent, header)

		startCard, endCard, showAbove, showBelow := m.visibleCardRange(col, i == m.FocusedColumnIndex, headerHeight)
		if i == m.FocusedColumnIndex && showAbove {
			columnContent = append(columnContent, "↑")
		}

		for j := startCard; j < endCard; j++ {
			isCardSelected := i == m.FocusedColumnIndex && j == m.FocusedCardIndex
			cardView := m.renderCard(col.Cards[j], isCardSelected)
			columnContent = append(columnContent, cardView)
		}

		if i == m.FocusedColumnIndex {
			if showBelow {
				columnContent = append(columnContent, "↓")
			}
		} else if showBelow {
			columnContent = append(columnContent, "...")
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

func (m BoardModel) renderColumnHeader(name string, isFocused bool, count int) string {
	headerStyle := components.ColumnHeaderStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
	if isFocused {
		headerStyle = headerStyle.BorderForeground(lipgloss.Color("205"))
	}
	return headerStyle.Render(name + " (" + fmt.Sprintf("%d", count) + ")")
}
