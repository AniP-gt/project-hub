package components

import "fmt"

// EmptyState renders a friendly message when no items match.
func EmptyState(reason string) string {
	if reason == "" {
		reason = "No items"
	}
	return fmt.Sprintf("%s. Press / to adjust filters.", reason)
}
