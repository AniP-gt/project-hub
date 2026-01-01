package state

import (
	"strings"
	"time"
)

// MatchesIterationFilters reports whether the item satisfies any of the provided iteration filters.
// Filters prefixed with "@" or named current/next/previous are treated as relative keywords.
// Literal values are matched case-insensitively against the iteration name or ID.
func MatchesIterationFilters(item Item, filters []string, now time.Time) bool {
	if len(filters) == 0 {
		return true
	}
	for _, f := range filters {
		if matchesIterationFilter(item, f, now) {
			return true
		}
	}
	return false
}

func matchesIterationFilter(item Item, filter string, now time.Time) bool {
	if filter == "" {
		return false
	}
	lower := strings.ToLower(filter)
	if strings.HasPrefix(lower, "iteration:") {
		lower = strings.TrimPrefix(lower, "iteration:")
	}
	lower = strings.TrimSpace(lower)
	if lower == "" {
		return false
	}
	switch lower {
	case "@current", "current":
		return isCurrentIteration(item, now)
	case "@next", "next":
		return isFutureIteration(item, now)
	case "@previous", "previous":
		return isPastIteration(item, now)
	default:
		if item.IterationName != "" && strings.EqualFold(item.IterationName, filter) {
			return true
		}
		if item.IterationID != "" && strings.EqualFold(item.IterationID, filter) {
			return true
		}
		return false
	}
}

func isCurrentIteration(item Item, now time.Time) bool {
	if item.IterationStart == nil || item.IterationDurationDays <= 0 {
		return false
	}
	end := item.IterationStart.Add(time.Duration(item.IterationDurationDays) * 24 * time.Hour)
	return (now.Equal(*item.IterationStart) || now.After(*item.IterationStart)) && now.Before(end)
}

func isFutureIteration(item Item, now time.Time) bool {
	if item.IterationStart == nil {
		return false
	}
	return now.Before(*item.IterationStart)
}

func isPastIteration(item Item, now time.Time) bool {
	if item.IterationStart == nil || item.IterationDurationDays <= 0 {
		return false
	}
	end := item.IterationStart.Add(time.Duration(item.IterationDurationDays) * 24 * time.Hour)
	return now.Equal(end) || now.After(end)
}
