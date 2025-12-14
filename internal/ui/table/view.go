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

	// Header
	head := lipgloss.JoinHorizontal(lipgloss.Top,
		components.TableHeaderCellStyle.Render("ID"),
		components.TableHeaderCellStyle.Render("Title"),
		components.TableHeaderCellStyle.Render("Status"),
		components.TableHeaderCellStyle.Render("Assignee"),
		components.TableHeaderCellStyle.Render("Priority"),
		components.TableHeaderCellStyle.Render("Updated"),
	)

	var rows []string
	for _, it := range items {
		rowStyle := components.TableRowBaseStyle
		if it.ID == focusedID {
			rowStyle = components.TableRowSelectedStyle
		}

		idCell := components.TableCellIDStyle.Width(8).Render(it.ID)
		titleCell := lipgloss.NewStyle().Width(28).Render(it.Title)
		statusCell := components.TableCellStatusStyle.Width(12).Render(it.Status)
		assignee := ""
		if len(it.Assignees) > 0 {
			assignee = "@" + it.Assignees[0]
		}
		assigneeCell := components.TableCellAssigneeStyle.Width(12).Render(assignee)
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
		priorityCell := priorityStyle.Width(10).Render(priority)
		updated := ""
		if it.UpdatedAt != nil {
			updated = it.UpdatedAt.Format("2006-01-02")
		}
		updatedCell := components.TableCellUpdatedStyle.Width(12).Render(updated)

		row := lipgloss.JoinHorizontal(lipgloss.Top, idCell, titleCell, statusCell, assigneeCell, priorityCell, updatedCell)
		rows = append(rows, rowStyle.Render(row))
	}

	return lipgloss.JoinVertical(lipgloss.Left, append([]string{head}, rows...)...)
}
