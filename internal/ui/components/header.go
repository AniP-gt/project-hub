package components

import (
	"github.com/charmbracelet/lipgloss"

	"project-hub/internal/state"
)

// RenderHeader shows project name, view, and active filter/mode hints with badges.
func RenderHeader(project state.Project, view state.ViewContext) string {
	filter := view.Filter.Query
	filterStatus := BadgeMuted.Render("filter:off")
	if filter != "" {
		filterStatus = BadgeActive.Render("filter:on")
	} else {
		filter = "(no filter)"
	}
	viewBadge := badgeForView(view.CurrentView)
	modeBadge := BadgeInfo.Render("mode:" + string(view.Mode))
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e5e7eb")).Render(project.Name)

	segments := []string{title, InlineGap() + viewBadge, InlineGap() + modeBadge, InlineGap() + filterStatus, InlineGap() + filter}
	return HeaderStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, segments...))
}

func badgeForView(v state.ViewType) string {
	switch v {
	case state.ViewTable:
		return BadgeActive.Render("table")
	case state.ViewRoadmap:
		return BadgeActive.Render("roadmap")
	default:
		return BadgeActive.Render("board")
	}
}
