package roadmap

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

func renderProgressBar(percent int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	total := 10
	filled := (percent * total) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", total-filled)
	return components.RoadmapProgressBarStyle.Render(bar)
}

// Render presents items with their iteration if any, with styled timeline and progress.
func Render(timelines []state.Timeline, items []state.Item, focusedID string) string {
	var blocks []string

	for _, tl := range timelines {
		header := components.RoadmapTimelineStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Bold(true).Render(tl.Name), lipgloss.NewStyle().Foreground(components.ColorGray500).Render("  (timeline)")))
		blocks = append(blocks, header)

		for _, it := range items {
			if it.IterationID != tl.ID {
				continue
			}
			isSelected := it.ID == focusedID
			rowStyle := components.RoadmapItemBaseStyle
			if isSelected {
				rowStyle = components.RoadmapItemSelectedStyle
			}

			percent := 0
			// try to parse percent from Timeline.Progress or Item.Position heuristics
			if tl.Progress != "" {
				// placeholder: if tl.Progress contains digits like "60%" extract them
				p := strings.TrimRight(tl.Progress, "%")
				fmt.Sscanf(p, "%d", &percent)
			}

			bar := renderProgressBar(percent)
			percentLabel := components.RoadmapProgressPercentStyle.Render(fmt.Sprintf("%d%%", percent))

			title := lipgloss.NewStyle().Foreground(components.ColorGray200).Render(it.Title)
			meta := components.RoadmapItemSprintStyle.Render(it.Status)
			content := lipgloss.JoinVertical(lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, title, lipgloss.NewStyle().PaddingLeft(1).Render(meta)), lipgloss.JoinHorizontal(lipgloss.Top, bar, percentLabel))
			blocks = append(blocks, rowStyle.Render(content))
		}
	}

	// Unscheduled
	var unscheduled []state.Item
	for _, it := range items {
		if it.IterationID == "" {
			unscheduled = append(unscheduled, it)
		}
	}
	if len(unscheduled) > 0 {
		blocks = append(blocks, components.RoadmapTimelineStyle.Render("[unscheduled]"))
		for _, it := range unscheduled {
			isSelected := it.ID == focusedID
			rowStyle := components.RoadmapItemBaseStyle
			if isSelected {
				rowStyle = components.RoadmapItemSelectedStyle
			}
			percent := 0
			bar := renderProgressBar(percent)
			percentLabel := components.RoadmapProgressPercentStyle.Render(fmt.Sprintf("%d%%", percent))
			title := lipgloss.NewStyle().Foreground(components.ColorGray200).Render(it.Title)
			content := lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.JoinHorizontal(lipgloss.Top, bar, percentLabel))
			blocks = append(blocks, rowStyle.Render(content))
		}
	}

	body := strings.TrimRight(lipgloss.JoinVertical(lipgloss.Left, blocks...), "\n")
	// Return unframed roadmap body; App will apply the frame
	return body
}
