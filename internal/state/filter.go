package state

import (
	"strings"
	"time"
)

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

func ApplyFilter(items []Item, fields []Field, fs FilterState, now time.Time) []Item {
	if fs.Query == "" && len(fs.Labels) == 0 && len(fs.Assignees) == 0 && len(fs.Statuses) == 0 && len(fs.Iterations) == 0 && len(fs.FieldFilters) == 0 {
		return items
	}
	var out []Item
	for _, it := range items {
		if fs.Query != "" && !strings.Contains(strings.ToLower(it.Title), strings.ToLower(fs.Query)) {
			continue
		}
		if len(fs.Labels) > 0 && !containsAny(it.Labels, fs.Labels) {
			continue
		}
		if len(fs.Assignees) > 0 && !containsAny(it.Assignees, fs.Assignees) {
			continue
		}
		if len(fs.Statuses) > 0 && !containsAny([]string{it.Status}, fs.Statuses) {
			continue
		}
		if len(fs.Iterations) > 0 && !MatchesIterationFilters(it, fs.Iterations, now) {
			continue
		}
		if len(fs.FieldFilters) > 0 && !matchesFieldFilters(it, fields, fs.FieldFilters, now) {
			continue
		}
		out = append(out, it)
	}
	return out
}

func matchesFieldFilters(item Item, fields []Field, filters map[string][]string, now time.Time) bool {
	if len(filters) == 0 {
		return true
	}
	for name, values := range filters {
		if !matchesSingleFieldFilter(item, fields, name, values, now) {
			return false
		}
	}
	return true
}

func matchesSingleFieldFilter(item Item, fields []Field, name string, values []string, now time.Time) bool {
	if len(values) == 0 {
		return true
	}
	fieldName := strings.TrimSpace(name)
	if fieldName == "" {
		return true
	}
	if strings.EqualFold(fieldName, "title") {
		return matchSliceValues([]string{item.Title}, values)
	}
	if strings.EqualFold(fieldName, "status") {
		return matchSliceValues([]string{item.Status}, values)
	}
	if strings.EqualFold(fieldName, "priority") {
		return matchSliceValues([]string{item.Priority}, values)
	}
	if strings.EqualFold(fieldName, "milestone") {
		return matchSliceValues([]string{item.Milestone}, values)
	}
	if strings.EqualFold(fieldName, "labels") || strings.EqualFold(fieldName, "label") {
		return matchSliceValues(item.Labels, values)
	}
	if strings.EqualFold(fieldName, "assignees") || strings.EqualFold(fieldName, "assignee") {
		return matchSliceValues(item.Assignees, values)
	}
	if strings.EqualFold(fieldName, "iteration") {
		return MatchesIterationFilters(item, values, now)
	}
	if len(item.FieldValues) > 0 {
		for key, stored := range item.FieldValues {
			if strings.EqualFold(key, fieldName) {
				return matchSliceValues(stored, values)
			}
		}
	}
	for _, field := range fields {
		if strings.EqualFold(field.Name, fieldName) {
			stored := item.FieldValues[field.Name]
			return matchSliceValues(stored, values)
		}
	}
	return false
}

func matchSliceValues(haystack []string, needles []string) bool {
	if len(needles) == 0 {
		return true
	}
	for _, needle := range needles {
		if needle == "" {
			continue
		}
		for _, value := range haystack {
			if value == "" {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(needle)) {
				return true
			}
		}
	}
	return false
}

func containsAny(haystack []string, needles []string) bool {
	for _, n := range needles {
		for _, h := range haystack {
			if h == n {
				return true
			}
		}
	}
	return false
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
