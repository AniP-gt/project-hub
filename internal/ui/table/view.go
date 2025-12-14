package table

import (
	"github.com/charmbracelet/lipgloss"
	"strings"

	"github.com/mattn/go-runewidth"

	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

// Render shows items in rows with styled header and focus indicator.
func Render(items []state.Item, focusedID string) string {
	if len(items) == 0 {
		return ""
	}

	// Column widths (keep in sync with row cell widths)
	const (
		titleWidth    = 36
		statusWidth   = 14
		assigneeWidth = 12
		priorityWidth = 10
		updatedWidth  = 12
	)

	// Header cells with explicit widths (ID is not shown)
	// Use bordered header cells to form a grid
	head := lipgloss.JoinHorizontal(lipgloss.Top,
		components.TableHeaderCellStyle.Width(2).Render(""),
		components.TableHeaderCellStyle.Width(titleWidth).Render("Title"),
		components.TableHeaderCellStyle.Width(statusWidth).Render("Status"),
		components.TableHeaderCellStyle.Width(assigneeWidth).Render("Assignee"),
		components.TableHeaderCellStyle.Width(priorityWidth).Render("Priority"),
		components.TableHeaderCellStyle.Width(updatedWidth).Render("Updated"),
	)

	var rows []string
	for _, it := range items {
		rowStyle := components.TableRowBaseStyle
		if it.ID == focusedID {
			rowStyle = components.TableRowSelectedStyle
		}

		// Selection marker instead of ID column
		marker := " "
		if it.ID == focusedID {
			marker = ">"
		}
		assignee := ""
		if len(it.Assignees) > 0 {
			assignee = "@" + it.Assignees[0]
		}

		// Derive a priority label from labels (fallback to empty)
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
		priorityStyle := components.CardPriorityStyle
		switch priority {
		case "High":
			priorityStyle = priorityStyle.Foreground(components.ColorRed400)
		case "Medium":
			priorityStyle = priorityStyle.Foreground(components.ColorYellow400)
		case "Low":
			priorityStyle = priorityStyle.Foreground(components.ColorGreen400)
		}

		updated := ""
		if it.UpdatedAt != nil {
			updated = it.UpdatedAt.Format("2006-01-02")
		}

		// Truncate title to fit into column
		title := it.Title
		if lipgloss.Width(title) > titleWidth {
			title = runewidth.Truncate(title, titleWidth-1, "...")
		}

		// Render each cell with shared TableCellStyle for borders and alignment
		cells := []string{
			components.TableMarkerCellStyle.Render(marker),
			components.TableCellStyle.Width(titleWidth).Render(title),
			components.TableCellStyle.Width(statusWidth).Render(components.TableCellStatusStyle.Render(it.Status)),
			components.TableCellStyle.Width(assigneeWidth).Render(components.TableCellAssigneeStyle.Render(assignee)),
			components.TableCellStyle.Width(priorityWidth).Render(priorityStyle.Render(priority)),
			components.TableCellStyle.Width(updatedWidth).Render(components.TableCellUpdatedStyle.Render(updated)),
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		rows = append(rows, rowStyle.Render(row))
	}

	tableBody := lipgloss.JoinVertical(lipgloss.Left, append([]string{head}, rows...)...)
	// Return unframed table body; App will apply the outer frame so header/footer stay fixed
	return tableBody
}
