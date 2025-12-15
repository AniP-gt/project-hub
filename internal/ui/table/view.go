package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
)

// Render renders the table view using lipgloss, matching the moc.go layout.
func Render(items []state.Item, focusedID string, innerWidth int) string {
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
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	// Columns: Title, Status, Repository, Labels, Milestone, Priority
	cols := []string{"Title", "Status", "Repository", "Labels", "Milestone", "Priority"}

	// Assign widths by percentage of innerWidth
	// Reserve a small margin for spacing
	margin := 2
	avail := innerWidth - margin
	if avail < 40 {
		avail = 40
	}

	percent := []int{40, 10, 20, 15, 8, 7} // sums to 100
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
		style := cellStyle
		if it.ID == focusedID {
			style = selectedStyle
		}

		cells := []string{
			style.Width(widths[0]).Render(it.Title),
			style.Width(widths[1]).Render(it.Status),
			style.Width(widths[2]).Render(it.Repository),
			style.Width(widths[3]).Render(strings.Join(it.Labels, ",")),
			style.Width(widths[4]).Render(it.Milestone),
			style.Width(widths[5]).Render(it.Priority),
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
