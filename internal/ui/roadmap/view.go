package roadmap

import (
	"fmt"
	"strings"

	"project-hub/internal/state"
)

// Render presents items with their iteration if any.
func Render(timelines []state.Timeline, items []state.Item, focusedID string) string {
	var b strings.Builder
	for _, tl := range timelines {
		b.WriteString(fmt.Sprintf("[%s]\n", tl.Name))
		for _, it := range items {
			if it.IterationID != tl.ID {
				continue
			}
			marker := " "
			if it.ID == focusedID {
				marker = ">"
			}
			b.WriteString(fmt.Sprintf("  %s %s (%s)\n", marker, it.Title, it.Status))
		}
	}
	// Unscheduled items
	var unscheduled []state.Item
	for _, it := range items {
		if it.IterationID == "" {
			unscheduled = append(unscheduled, it)
		}
	}
	if len(unscheduled) > 0 {
		b.WriteString("[unscheduled]\n")
		for _, it := range unscheduled {
			marker := " "
			if it.ID == focusedID {
				marker = ">"
			}
			b.WriteString(fmt.Sprintf("  %s %s (%s)\n", marker, it.Title, it.Status))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
