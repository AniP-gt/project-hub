package board

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

const (
	maxTitleLines = 3
	maxMetaLines  = 1
)

func (m BoardModel) renderCard(c state.Card, isSelected bool) string {
	contentWidth := m.ColumnWidth - 4
	if contentWidth < 12 {
		contentWidth = 12
	}
	wrap := func(value string, maxLines int, isSelected bool) string {
		if value == "" {
			return ""
		}
		bg := components.ColorGray800
		if isSelected {
			bg = components.ColorGray700
		}
		rendered := lipgloss.NewStyle().Width(contentWidth).MaxWidth(contentWidth).Background(bg).Render(value)
		return clampRenderedLines(rendered, maxLines, contentWidth)
	}

	title := wrap(c.Title, maxTitleLines, isSelected)

	var contentBlocks []string
	if title != "" {
		contentBlocks = append(contentBlocks, title)
	}

	if c.Assignee != "" {
		assignee := wrap("@"+c.Assignee, maxMetaLines, isSelected)
		if assignee != "" {
			contentBlocks = append(contentBlocks, assignee)
		}
	}

	if m.FieldVisibility.ShowLabels && len(c.Labels) > 0 {
		labels := wrap("["+strings.Join(c.Labels, ", ")+"]", maxMetaLines, isSelected)
		if labels != "" {
			contentBlocks = append(contentBlocks, labels)
		}
	}

	if c.Priority != "" {
		priorityStyle := components.CardPriorityStyle
		switch c.Priority {
		case "High":
			priorityStyle = priorityStyle.Foreground(components.ColorRed400)
		case "Medium":
			priorityStyle = priorityStyle.Foreground(components.ColorYellow400)
		case "Low":
			priorityStyle = priorityStyle.Foreground(components.ColorGreen400)
		}
		priority := wrap(priorityStyle.Render(c.Priority), maxMetaLines, isSelected)
		if priority != "" {
			contentBlocks = append(contentBlocks, priority)
		}
	}

	if m.FieldVisibility.ShowMilestone && c.Milestone != "" {
		milestone := wrap("M: "+c.Milestone, maxMetaLines, isSelected)
		if milestone != "" {
			contentBlocks = append(contentBlocks, milestone)
		}
	}

	if m.FieldVisibility.ShowRepository && c.Repository != "" {
		repo := wrap("R: "+c.Repository, maxMetaLines, isSelected)
		if repo != "" {
			contentBlocks = append(contentBlocks, repo)
		}
	}

	if m.FieldVisibility.ShowSubIssueProgress && c.SubIssueProgress != "" {
		subIssueStyle := lipgloss.NewStyle().Foreground(components.ColorCyan400)
		subIssue := wrap(subIssueStyle.Render("S: "+c.SubIssueProgress), maxMetaLines, isSelected)
		if subIssue != "" {
			contentBlocks = append(contentBlocks, subIssue)
		}
	}

	if m.FieldVisibility.ShowParentIssue && c.ParentIssue != "" {
		parentStyle := lipgloss.NewStyle().Foreground(components.ColorPurple400)
		parent := wrap(parentStyle.Render("P: "+c.ParentIssue), maxMetaLines, isSelected)
		if parent != "" {
			contentBlocks = append(contentBlocks, parent)
		}
	}

	if len(contentBlocks) == 0 {
		contentBlocks = append(contentBlocks, "(no title)")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, contentBlocks...)

	style := components.CardBaseStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
	if isSelected {
		style = components.CardSelectedStyle.Copy().Width(m.ColumnWidth).MaxWidth(m.ColumnWidth)
	}

	return style.Render(content)
}

func clampRenderedLines(rendered string, maxLines, width int) string {
	if maxLines <= 0 {
		return rendered
	}
	lines := strings.Split(rendered, "\n")
	if len(lines) <= maxLines {
		return rendered
	}
	clamped := lines[:maxLines]
	clamped[maxLines-1] = truncateLineWithEllipsis(clamped[maxLines-1], width)
	return strings.Join(clamped, "\n")
}

func truncateLineWithEllipsis(line string, width int) string {
	ellipsis := "..."
	if width <= lipgloss.Width(ellipsis) {
		return ellipsis
	}
	trimmed := strings.TrimRight(line, " ")
	runes := []rune(trimmed)
	for lipgloss.Width(strings.TrimRight(string(runes), " "))+lipgloss.Width(ellipsis) > width && len(runes) > 0 {
		runes = runes[:len(runes)-1]
	}
	trimmed = strings.TrimRight(string(runes), " ")
	if trimmed == "" {
		return ellipsis
	}
	return trimmed + ellipsis
}
