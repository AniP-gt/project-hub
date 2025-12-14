package table

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

// TableModel holds state for the table view with scrolling.
type TableModel struct {
	Items        []state.Item
	FocusedIndex int
	FocusedID    string
	Width        int
	Height       int
	RowOffset    int
}

func NewTableModel(items []state.Item, focusedID string) TableModel {
	idx := 0
	for i, it := range items {
		if it.ID == focusedID {
			idx = i
			break
		}
	}
	return TableModel{Items: items, FocusedIndex: idx, FocusedID: focusedID, Width: 0, Height: 0}
}

func (m TableModel) Init() tea.Cmd { return nil }

func (m TableModel) calculateMaxVisibleRows() int {
	if m.Height <= 0 {
		return 6
	}
	// reserve a couple lines for padding within the frame
	available := m.Height - 2
	if available < 3 {
		return 1
	}
	return available
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch tm := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = tm.Width
		m.Height = tm.Height
		// ensure focused index visible
		max := m.calculateMaxVisibleRows()
		if m.FocusedIndex < m.RowOffset {
			m.RowOffset = m.FocusedIndex
		} else if m.FocusedIndex >= m.RowOffset+max {
			m.RowOffset = m.FocusedIndex - (max - 1)
		}
	case tea.KeyMsg:
		s := tm.String()
		switch s {
		case "j", "down":
			if m.FocusedIndex < len(m.Items)-1 {
				m.FocusedIndex++
				max := m.calculateMaxVisibleRows()
				if m.FocusedIndex >= m.RowOffset+max {
					m.RowOffset++
				}
			}
		case "k", "up":
			if m.FocusedIndex > 0 {
				m.FocusedIndex--
				if m.FocusedIndex < m.RowOffset {
					m.RowOffset--
				}
			}
		case "g":
			// goto top
			m.FocusedIndex = 0
			m.RowOffset = 0
		case "G":
			// goto bottom
			m.FocusedIndex = len(m.Items) - 1
			max := m.calculateMaxVisibleRows()
			if m.FocusedIndex-max+1 > 0 {
				m.RowOffset = m.FocusedIndex - max + 1
			} else {
				m.RowOffset = 0
			}
		case "ctrl+u":
			max := m.calculateMaxVisibleRows()
			delta := max / 2
			m.FocusedIndex -= delta
			if m.FocusedIndex < 0 {
				m.FocusedIndex = 0
			}
			if m.FocusedIndex < m.RowOffset {
				m.RowOffset = m.FocusedIndex
			}
		case "ctrl+d":
			max := m.calculateMaxVisibleRows()
			delta := max / 2
			m.FocusedIndex += delta
			if m.FocusedIndex > len(m.Items)-1 {
				m.FocusedIndex = len(m.Items) - 1
			}
			if m.FocusedIndex >= m.RowOffset+max {
				m.RowOffset = m.FocusedIndex - max + 1
			}
		}
	}
	return m, cmd
}

func (m TableModel) View() string {
	// Use fixed widths similar to previous implementation
	const (
		titleWidth    = 36
		statusWidth   = 14
		assigneeWidth = 12
		priorityWidth = 10
		updatedWidth  = 12
	)

	// header
	head := lipgloss.JoinHorizontal(lipgloss.Top,
		components.TableHeaderCellStyle.Width(2).Render(""),
		components.TableHeaderCellStyle.Width(titleWidth).Render("Title"),
		components.TableHeaderCellStyle.Width(statusWidth).Render("Status"),
		components.TableHeaderCellStyle.Width(assigneeWidth).Render("Assignee"),
		components.TableHeaderCellStyle.Width(priorityWidth).Render("Priority"),
		components.TableHeaderCellStyle.Width(updatedWidth).Render("Updated"),
	)

	max := m.calculateMaxVisibleRows()
	start := m.RowOffset
	end := start + max
	if end > len(m.Items) {
		end = len(m.Items)
	}

	var rows []string
	for i := start; i < end; i++ {
		it := m.Items[i]
		marker := " "
		if i == m.FocusedIndex {
			marker = ">"
		}

		assignee := ""
		if len(it.Assignees) > 0 {
			assignee = "@" + it.Assignees[0]
		}

		// priority derivation
		priority := ""
		for _, l := range it.Labels {
			if strings.Contains(strings.ToLower(l), "high") {
				priority = "High"
				break
			} else if strings.Contains(strings.ToLower(l), "low") {
				priority = "Low"
				break
			}
		}
		if priority == "" {
			priority = "Medium"
		}

		updated := ""
		if it.UpdatedAt != nil {
			updated = it.UpdatedAt.Format("2006-01-02")
		}

		title := it.Title
		if runewidth.StringWidth(title) > titleWidth {
			title = runewidth.Truncate(title, titleWidth-1, "...")
		}

		cells := []string{
			components.TableMarkerCellStyle.Render(marker),
			components.TableCellStyle.Width(titleWidth).Render(title),
			components.TableCellStyle.Width(statusWidth).Render(components.TableCellStatusStyle.Render(it.Status)),
			components.TableCellStyle.Width(assigneeWidth).Render(components.TableCellAssigneeStyle.Render(assignee)),
			components.TableCellStyle.Width(priorityWidth).Render(components.CardPriorityStyle.Render(priority)),
			components.TableCellStyle.Width(updatedWidth).Render(components.TableCellUpdatedStyle.Render(updated)),
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		rowStyle := components.TableRowBaseStyle
		if i == m.FocusedIndex {
			rowStyle = components.TableRowSelectedStyle
		}
		rows = append(rows, rowStyle.Render(row))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, append([]string{head}, rows...)...)
	return body
}
