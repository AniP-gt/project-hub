package state

import "strings"

// ParseFilter converts a raw query into structured filter fields.
func ParseFilter(query string) FilterState {
	trimmed := strings.TrimSpace(query)
	fs := FilterState{Query: trimmed}
	if trimmed == "" {
		return fs
	}
	var queryTokens []string
	tokens := strings.Fields(trimmed)
	for _, t := range tokens {
		if strings.HasPrefix(t, "label:") {
			fs.Labels = append(fs.Labels, strings.TrimPrefix(t, "label:"))
			continue
		}
		if strings.HasPrefix(t, "assignee:") {
			fs.Assignees = append(fs.Assignees, strings.TrimPrefix(t, "assignee:"))
			continue
		}
		if strings.HasPrefix(t, "status:") {
			fs.Statuses = append(fs.Statuses, strings.TrimPrefix(t, "status:"))
			continue
		}
		if strings.HasPrefix(t, "iteration:") {
			fs.Iterations = append(fs.Iterations, strings.TrimPrefix(t, "iteration:"))
			continue
		}
		if strings.HasPrefix(t, "group:") {
			fs.GroupBy = strings.TrimSpace(strings.TrimPrefix(t, "group:"))
			continue
		}
		if strings.HasPrefix(t, "group-by:") {
			fs.GroupBy = strings.TrimSpace(strings.TrimPrefix(t, "group-by:"))
			continue
		}
		if strings.HasPrefix(t, "groupby:") {
			fs.GroupBy = strings.TrimSpace(strings.TrimPrefix(t, "groupby:"))
			continue
		}
		if strings.HasPrefix(t, "@") {
			fs.Iterations = append(fs.Iterations, t)
			continue
		}
		lower := strings.ToLower(t)
		if lower == "current" || lower == "next" || lower == "previous" {
			fs.Iterations = append(fs.Iterations, t)
			continue
		}
		queryTokens = append(queryTokens, t)
	}
	fs.Query = strings.Join(queryTokens, " ")
	return fs
}
