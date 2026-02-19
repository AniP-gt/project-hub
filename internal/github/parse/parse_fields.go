package parse

import (
	"fmt"
	"strings"
	"time"

	"project-hub/internal/state"
)

func applyIterationMetadata(item *state.Item, data map[string]any) bool {
	iterationID, ok := data["iterationId"].(string)
	if !ok || iterationID == "" {
		return false
	}
	item.IterationID = iterationID
	if title, ok := data["title"].(string); ok {
		item.IterationName = title
	}
	if start, ok := data["startDate"].(string); ok && start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			item.IterationStart = &t
		}
	}
	switch dur := data["duration"].(type) {
	case float64:
		item.IterationDurationDays = int(dur)
	case int:
		item.IterationDurationDays = dur
	}
	return true
}

func compactFieldName(name string) string {
	if name == "" {
		return ""
	}
	return strings.ReplaceAll(strings.ReplaceAll(name, " ", ""), "-", "")
}

func isSubIssueField(compactName string) bool {
	if compactName == "" {
		return false
	}
	switch compactName {
	case "subissuecount", "subissues", "subissueprogress", "subissue", "subissuescount":
		return true
	default:
		return false
	}
}

func isParentIssueField(compactName string) bool {
	if compactName == "" {
		return false
	}
	switch compactName {
	case "parentissue", "isparentissue", "parent":
		return true
	default:
		return false
	}
}

func mergeStringsUnique(dst []string, src []string) []string {
	for _, val := range src {
		if val == "" {
			continue
		}
		exists := false
		for _, existing := range dst {
			if existing == val {
				exists = true
				break
			}
		}
		if !exists {
			dst = append(dst, val)
		}
	}
	return dst
}

func extractLabelNames(val any) []string {
	var out []string
	switch data := val.(type) {
	case []any:
		for _, entry := range data {
			out = append(out, extractLabelNames(entry)...)
		}
	case map[string]any:
		if nodes, ok := data["nodes"]; ok {
			out = append(out, extractLabelNames(nodes)...)
		} else if name, ok := data["name"].(string); ok && name != "" {
			out = append(out, name)
		} else if text, ok := data["text"].(string); ok && text != "" {
			out = append(out, splitListField(text)...)
		}
	case string:
		out = append(out, splitListField(data)...)
	}
	return out
}

func extractAssigneeLogins(val any) []string {
	var out []string
	switch data := val.(type) {
	case []any:
		for _, entry := range data {
			out = append(out, extractAssigneeLogins(entry)...)
		}
	case map[string]any:
		if nodes, ok := data["nodes"]; ok {
			out = append(out, extractAssigneeLogins(nodes)...)
		} else if login, ok := data["login"].(string); ok && login != "" {
			out = append(out, login)
		} else if name, ok := data["name"].(string); ok && name != "" {
			out = append(out, name)
		} else if text, ok := data["text"].(string); ok && text != "" {
			out = append(out, splitListField(text)...)
		}
	case string:
		out = append(out, splitListField(data)...)
	}
	return out
}

func splitListField(text string) []string {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		switch r {
		case ',', ';', '\n', '\r':
			return true
		}
		return false
	})
	var out []string
	for _, f := range fields {
		val := strings.TrimSpace(f)
		if val != "" {
			out = append(out, val)
		}
	}
	return out
}

func addFieldValue(values map[string][]string, fieldName string, fieldMap map[string]any) map[string][]string {
	if fieldName == "" {
		return values
	}
	if values == nil {
		values = make(map[string][]string)
	}
	var collected []string
	if iterID, ok := fieldMap["iterationId"].(string); ok && iterID != "" {
		collected = append(collected, iterID)
	}
	if title, ok := fieldMap["title"].(string); ok && title != "" {
		collected = append(collected, title)
	}
	if opt, ok := fieldMap["singleSelectOption"].(map[string]any); ok {
		if name, ok := opt["name"].(string); ok && name != "" {
			collected = append(collected, name)
		}
		if id, ok := opt["id"].(string); ok && id != "" {
			collected = append(collected, id)
		}
	}
	if text, ok := fieldMap["text"].(string); ok && text != "" {
		collected = append(collected, splitListField(text)...)
	}
	if num, ok := fieldMap["number"]; ok {
		switch n := num.(type) {
		case float64:
			collected = append(collected, fmt.Sprintf("%.0f", n))
		case int:
			collected = append(collected, fmt.Sprintf("%d", n))
		}
	}
	if labelsVal, ok := fieldMap["labels"]; ok {
		collected = append(collected, extractLabelNames(labelsVal)...)
	}
	if usersVal, ok := fieldMap["users"]; ok {
		collected = append(collected, extractAssigneeLogins(usersVal)...)
	}
	if msVal, ok := fieldMap["milestone"]; ok {
		if parsed := parseMilestoneValue(msVal); parsed != "" {
			collected = append(collected, parsed)
		}
	}
	if len(collected) == 0 {
		return values
	}
	key := strings.TrimSpace(fieldName)
	values[key] = mergeStringsUnique(values[key], collected)
	return values
}
