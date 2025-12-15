package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
)

// Render renders the table view using lipgloss, matching the moc.go layout.
func Render(items []state.Item, focusedID string, focusedColIndex int, innerWidth int) string {
	if innerWidth <= 0 {
		innerWidth = 80
	}

	headStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Bold(true)

	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Padding(0, 1)

	focusedCellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // Brighter foreground for focused cell
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Bold(true)

	// Columns: Title, Status, Repository, Labels, Milestone, Priority, Assignees
	cols := []string{"Title", "Status", "Repository", "Labels", "Milestone", "Priority", "Assignees"}

	// Assign widths by percentage of innerWidth
	// Reserve a small margin for spacing
	margin := 2
	avail := innerWidth - margin
	if avail < 40 {
		avail = 40
	}

	percent := []int{40, 10, 15, 10, 8, 5, 12} // sums to 100
	widths := make([]int, len(percent))
	total := 0
	for i, p := range percent {
		w := avail * p / 100
		if w < 6 {
			w = 6
		}
		widths[i] = w
		total += w
	}

	// If rounding causes total > avail, shrink title
	if total > avail {
		over := total - avail
		widths[0] -= over
	}

	// Prepare header row
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = headStyle.Width(widths[i]).Render(c)
	}

	var rows []string
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, headers...))
	sepLen := innerWidth
	if sepLen < 0 {
		sepLen = 0
	}
	rows = append(rows, strings.Repeat("─", sepLen))

	// Build data rows
	for _, it := range items {
		// Base style for the entire row
		rowBaseStyle := cellStyle
		if it.ID == focusedID {
			rowBaseStyle = selectedStyle
		}

		cells := make([]string, len(cols)) // cols: Title, Status, ...
		cellValues := []string{
			it.Title,
			it.Status,
			it.Repository,
			strings.Join(it.Labels, ","),
			it.Milestone,
			it.Priority,
			strings.Join(it.Assignees, ","), // Add Assignees here
		}

		for colIdx, val := range cellValues {
			cellStyleToApply := rowBaseStyle.Copy() // Start with the row's base style
			if it.ID == focusedID && colIdx == focusedColIndex {
				cellStyleToApply = focusedCellStyle.Copy()
			}
			cells[colIdx] = cellStyleToApply.Width(widths[colIdx]).Render(val)
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
