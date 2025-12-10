package components

import (
	"github.com/charmbracelet/lipgloss"
)

// RenderFooter shows key hints and mode status.
func RenderFooter(mode, view string) string {
	hints := []string{
		BadgeInfo.Render("1/b:board"),
		BadgeInfo.Render("2/t:table"),
		BadgeInfo.Render("3/r:roadmap"),
		BadgeInfo.Render("j/k:move"),
		BadgeInfo.Render("h/l:status"),
		BadgeInfo.Render("/:filter"),
		BadgeInfo.Render("i:edit"),
		BadgeInfo.Render("a:assign"),
		BadgeInfo.Render("q:quit"),
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, hints...)
	status := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Render("view:" + view + " | mode:" + mode)
	return FooterStyle.Render(row + InlineGap() + status)
}
