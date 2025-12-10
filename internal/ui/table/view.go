package table

import (
	"fmt"
	"strings"

	"project-hub/internal/state"
)

// Render shows items in rows with focus indicator.
func Render(items []state.Item, focusedID string) string {
	var b strings.Builder
	for _, it := range items {
		marker := " "
		if it.ID == focusedID {
			marker = ">"
		}
		b.WriteString(fmt.Sprintf("%s %-15s | %-10s | %s\n", marker, it.Title, it.Status, strings.Join(it.Labels, ",")))
	}
	return strings.TrimRight(b.String(), "\n")
}
