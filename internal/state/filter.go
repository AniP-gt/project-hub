package state

import "strings"

// ParseFilter converts a raw query into structured filter fields.
func ParseFilter(query string) FilterState {
	trimmed := strings.TrimSpace(query)
	fs := FilterState{Raw: trimmed, Query: trimmed}
	if trimmed == "" {
		return fs
	}
	var queryTokens []string
	tokens := tokenizeFilterQuery(trimmed)
	for _, t := range tokens {
		if strings.HasPrefix(t, "label:") {
			fs.Labels = append(fs.Labels, splitFilterValues(strings.TrimPrefix(t, "label:"))...)
			continue
		}
		if strings.HasPrefix(t, "labels:") {
			fs.Labels = append(fs.Labels, splitFilterValues(strings.TrimPrefix(t, "labels:"))...)
			continue
		}
		if strings.HasPrefix(t, "assignee:") {
			fs.Assignees = append(fs.Assignees, splitFilterValues(strings.TrimPrefix(t, "assignee:"))...)
			continue
		}
		if strings.HasPrefix(t, "assignees:") {
			fs.Assignees = append(fs.Assignees, splitFilterValues(strings.TrimPrefix(t, "assignees:"))...)
			continue
		}
		if strings.HasPrefix(t, "status:") {
			fs.Statuses = append(fs.Statuses, splitFilterValues(strings.TrimPrefix(t, "status:"))...)
			continue
		}
		if strings.HasPrefix(t, "iteration:") {
			fs.Iterations = append(fs.Iterations, splitFilterValues(strings.TrimPrefix(t, "iteration:"))...)
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
		if fieldName, fieldValue, ok := splitFieldToken(t); ok {
			fs.FieldFilters = addFieldFilter(fs.FieldFilters, fieldName, splitFilterValues(fieldValue))
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
		queryTokens = append(queryTokens, trimQuotes(t))
	}
	fs.Query = strings.Join(queryTokens, " ")
	return fs
}

func splitFieldToken(token string) (string, string, bool) {
	idx := strings.IndexRune(token, ':')
	if idx <= 0 || idx >= len(token)-1 {
		return "", "", false
	}
	fieldName := strings.TrimSpace(token[:idx])
	if fieldName == "" {
		return "", "", false
	}
	fieldName = trimQuotes(fieldName)
	if isReservedFilterKey(fieldName) {
		return "", "", false
	}
	value := strings.TrimSpace(token[idx+1:])
	value = trimQuotes(value)
	if value == "" {
		return "", "", false
	}
	return fieldName, value, true
}

func isReservedFilterKey(fieldName string) bool {
	key := strings.ToLower(strings.TrimSpace(fieldName))
	switch key {
	case "label", "labels", "assignee", "assignees", "status", "iteration", "group", "group-by", "groupby":
		return true
	default:
		return false
	}
}

func tokenizeFilterQuery(query string) []string {
	var tokens []string
	var current strings.Builder
	inQuote := false
	for _, r := range query {
		switch r {
		case '"':
			inQuote = !inQuote
			current.WriteRune(r)
		case ' ', '\t', '\n', '\r':
			if inQuote {
				current.WriteRune(r)
				continue
			}
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func splitFilterValues(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case ',', ';':
			return true
		default:
			return false
		}
	})
	var out []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			out = append(out, trimQuotes(trimmed))
		}
	}
	return out
}

func trimQuotes(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 && strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"") {
		return strings.TrimSuffix(strings.TrimPrefix(trimmed, "\""), "\"")
	}
	return trimmed
}

func addFieldFilter(filters map[string][]string, fieldName string, values []string) map[string][]string {
	if len(values) == 0 {
		return filters
	}
	if filters == nil {
		filters = make(map[string][]string)
	}
	key := strings.TrimSpace(fieldName)
	filters[key] = append(filters[key], values...)
	return filters
}
