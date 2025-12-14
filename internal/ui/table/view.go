package table

import (
	"github.com/charmbracelet/lipgloss"
	"strings"

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
	// include a small empty header cell for the selection marker column
	markerHeader := lipgloss.NewStyle().Width(2).Render("")
	head := lipgloss.JoinHorizontal(lipgloss.Top,
		markerHeader,
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
		markerCell := lipgloss.NewStyle().Width(2).Render(marker)

		titleCell := lipgloss.NewStyle().Width(titleWidth).Render(it.Title)
		statusCell := components.TableCellStatusStyle.Width(statusWidth).Render(it.Status)
		assignee := ""
		if len(it.Assignees) > 0 {
			assignee = "@" + it.Assignees[0]
		}
		assigneeCell := components.TableCellAssigneeStyle.Width(assigneeWidth).Render(assignee)

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
		priorityCell := priorityStyle.Width(priorityWidth).Render(priority)
		updated := ""
		if it.UpdatedAt != nil {
			updated = it.UpdatedAt.Format("2006-01-02")
		}
		updatedCell := components.TableCellUpdatedStyle.Width(updatedWidth).Render(updated)

		row := lipgloss.JoinHorizontal(lipgloss.Top, markerCell, titleCell, statusCell, assigneeCell, priorityCell, updatedCell)
		rows = append(rows, rowStyle.Render(row))
	}

	tableBody := lipgloss.JoinVertical(lipgloss.Left, append([]string{head}, rows...)...)
	// Wrap table in a bright frame so it looks like a distinct table region
	framed := components.FrameStyle.Width(lipgloss.Width(tableBody) + components.FrameStyle.GetHorizontalFrameSize()).Render(tableBody)
	return framed
}
