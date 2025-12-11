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
		repo := it.Repository
		b.WriteString(fmt.Sprintf("%s %-20s | %-10s | %-20s | %s\n", marker, it.Title, it.Status, repo, strings.Join(it.Labels, ",")))
	}
	return strings.TrimRight(b.String(), "\n")
}
