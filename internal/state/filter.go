package state

import "strings"

// ParseFilter converts a raw query into structured filter fields.
func ParseFilter(query string) FilterState {
	fs := FilterState{Query: strings.TrimSpace(query)}
	if fs.Query == "" {
		return fs
	}
	tokens := strings.Fields(fs.Query)
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
		// Unknown token: keep it in Query but ignore structured fields.
	}
	return fs
}
