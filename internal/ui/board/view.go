package board

import (
	"strings"

	"project-hub/internal/state"
)

// Render returns a simple board-like text list with focus marker.
func Render(items []state.Item, focusedID string) string {
	var b strings.Builder
	for _, it := range items {
		marker := " "
		if it.ID == focusedID {
			marker = ">"
		}
		b.WriteString(marker)
		b.WriteString(" [")
		b.WriteString(it.Status)
		b.WriteString("] ")
		b.WriteString(it.Title)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
