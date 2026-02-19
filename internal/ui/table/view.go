package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

type RenderResult struct {
	Header       string
	Rows         []string
	RowHeights   []int
	RowOffsets   []int
	RowsLineSize int
}

// Render renders the table view using lipgloss, matching the moc.go layout.
func Render(items []state.Item, focusedID string, focusedColIndex int, innerWidth int, fieldVisibility state.CardFieldVisibility) RenderResult {
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

	columns := tableColumns(fieldVisibility)
	cols := make([]string, len(columns))
	for i, col := range columns {
		cols[i] = col.Label
	}

	// Assign widths by percentage of innerWidth
	// Reserve a small margin for spacing
	margin := 2
	avail := innerWidth - margin
	if avail < 40 {
		avail = 40
	}

	widths := make([]int, len(columns))
	total := 0
	percentTotal := 0
	for _, col := range columns {
		percentTotal += col.Percent
	}
	if percentTotal == 0 {
		percentTotal = 100
	}
	for i, col := range columns {
		w := avail * col.Percent / percentTotal
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
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, headers...)
	sepLen := innerWidth
	if sepLen < 0 {
		sepLen = 0
	}
	headerView := lipgloss.JoinVertical(lipgloss.Left, headerRow, strings.Repeat("─", sepLen))

	var dataRows []string
	var rowHeights []int
	var rowOffsets []int
	var cumulativeHeight int

	// Build data rows
	for _, it := range items {

		// Base style for the entire row
		rowBaseStyle := cellStyle
		if it.ID == focusedID {
			rowBaseStyle = selectedStyle
		}

		cells := make([]string, len(cols))
		cellValues := make([]string, len(columns))
		for i, col := range columns {
			cellValues[i] = columnValue(it, col.Key)
		}

		for colIdx, val := range cellValues {
			cellStyleToApply := rowBaseStyle.Copy() // Start with the row's base style
			if it.ID == focusedID && colIdx == focusedColIndex {
				cellStyleToApply = focusedCellStyle.Copy()
			}
			if columns[colIdx].Key == state.ColumnStatus {
				status := strings.TrimSpace(it.Status)
				if status != "" {
					val = components.StatusDot(status) + " " + status
				}
			}
			cells[colIdx] = cellStyleToApply.Width(widths[colIdx]).Render(val)
		}

		rowView := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		dataRows = append(dataRows, rowView)
		rowOffsets = append(rowOffsets, cumulativeHeight)
		rowHeight := lipgloss.Height(rowView)
		rowHeights = append(rowHeights, rowHeight)
		cumulativeHeight += rowHeight
	}

	return RenderResult{
		Header:       headerView,
		Rows:         dataRows,
		RowHeights:   rowHeights,
		RowOffsets:   rowOffsets,
		RowsLineSize: cumulativeHeight,
	}
}

type tableColumn struct {
	Key     int
	Label   string
	Percent int
}

func tableColumns(fieldVisibility state.CardFieldVisibility) []tableColumn {
	columns := []tableColumn{
		{Key: state.ColumnTitle, Label: "Title", Percent: 40},
		{Key: state.ColumnStatus, Label: "Status", Percent: 10},
	}

	if fieldVisibility.ShowRepository {
		columns = append(columns, tableColumn{Key: state.ColumnRepository, Label: "Repository", Percent: 15})
	}
	if fieldVisibility.ShowLabels {
		columns = append(columns, tableColumn{Key: state.ColumnLabels, Label: "Labels", Percent: 10})
	}
	if fieldVisibility.ShowMilestone {
		columns = append(columns, tableColumn{Key: state.ColumnMilestone, Label: "Milestone", Percent: 8})
	}
	if fieldVisibility.ShowSubIssueProgress {
		columns = append(columns, tableColumn{Key: state.ColumnSubIssueProgress, Label: "Sub-issues", Percent: 7})
	}
	if fieldVisibility.ShowParentIssue {
		columns = append(columns, tableColumn{Key: state.ColumnParentIssue, Label: "Parent", Percent: 8})
	}

	columns = append(columns, tableColumn{Key: state.ColumnAssignees, Label: "Assignees", Percent: 12})
	return columns
}

func columnValue(item state.Item, columnKey int) string {
	switch columnKey {
	case state.ColumnTitle:
		return item.Title
	case state.ColumnStatus:
		return item.Status
	case state.ColumnRepository:
		return item.Repository
	case state.ColumnLabels:
		return strings.Join(item.Labels, ",")
	case state.ColumnMilestone:
		return item.Milestone
	case state.ColumnSubIssueProgress:
		return item.SubIssueProgress
	case state.ColumnParentIssue:
		return item.ParentIssue
	case state.ColumnPriority:
		return item.Priority
	case state.ColumnAssignees:
		return strings.Join(item.Assignees, ",")
	default:
		return ""
	}
}

func (r RenderResult) RowBounds(index int) (int, int) {
	if index < 0 || index >= len(r.RowOffsets) || index >= len(r.RowHeights) {
		return -1, -1
	}
	top := r.RowOffsets[index]
	height := r.RowHeights[index]
	if height <= 0 {
		height = 1
	}
	bottom := top + height - 1
	return top, bottom
}

// Helper: GetItemURLByFocused returns the URL for the focused item from a list.
// This will be used by the app layer to open/copy the URL when a key is pressed.
func GetItemURLByFocused(items []state.Item, focusedID string) string {
	for _, it := range items {
		if it.ID == focusedID {
			return it.URL
		}
	}
	return ""
}
