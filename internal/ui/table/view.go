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
		idWidth       = 8
		titleWidth    = 28
		statusWidth   = 12
		assigneeWidth = 12
		priorityWidth = 10
		updatedWidth  = 12
	)

	// Header cells with explicit widths
	head := lipgloss.JoinHorizontal(lipgloss.Top,
		components.TableHeaderCellStyle.Width(idWidth).Render("ID"),
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

		idCell := components.TableCellIDStyle.Width(idWidth).Render(it.ID)
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

		row := lipgloss.JoinHorizontal(lipgloss.Top, idCell, titleCell, statusCell, assigneeCell, priorityCell, updatedCell)
		rows = append(rows, rowStyle.Render(row))
	}

	tableBody := lipgloss.JoinVertical(lipgloss.Left, append([]string{head}, rows...)...)
	// Wrap table in a border so it looks like a distinct table region
	return components.TableBorderStyle.Render(tableBody)
}
