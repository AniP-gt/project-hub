package parse

import (
	"fmt"
	"strings"
)

func parseMilestoneValue(val any) string {
	switch data := val.(type) {
	case string:
		return strings.TrimSpace(data)
	case map[string]any:
		if title, ok := data["title"].(string); ok && title != "" {
			return title
		}
		if name, ok := data["name"].(string); ok && name != "" {
			return name
		}
		if text, ok := data["text"].(string); ok && text != "" {
			return text
		}
	case []any:
		for _, entry := range data {
			if candidate := parseMilestoneValue(entry); candidate != "" {
				return candidate
			}
		}
	}
	return ""
}

func parseParentIssueValue(val any) string {
	switch data := val.(type) {
	case string:
		return strings.TrimSpace(data)
	case map[string]any:
		if title, ok := data["title"].(string); ok && title != "" {
			return title
		}
		if number, ok := data["number"].(float64); ok && number > 0 {
			return fmt.Sprintf("#%.0f", number)
		}
		if number, ok := data["number"].(int); ok && number > 0 {
			return fmt.Sprintf("#%d", number)
		}
		if content, ok := data["content"].(map[string]any); ok {
			if title, ok := content["title"].(string); ok && title != "" {
				return title
			}
			if number, ok := content["number"].(float64); ok && number > 0 {
				return fmt.Sprintf("#%.0f", number)
			}
			if number, ok := content["number"].(int); ok && number > 0 {
				return fmt.Sprintf("#%d", number)
			}
		}
	}
	return ""
}

func parseSubIssueCount(val any) string {
	switch data := val.(type) {
	case float64:
		if data > 0 {
			return fmt.Sprintf("%.0f", data)
		}
	case int:
		if data > 0 {
			return fmt.Sprintf("%d", data)
		}
	case map[string]any:
		if total, ok := data["totalCount"]; ok {
			switch n := total.(type) {
			case float64:
				if n > 0 {
					return fmt.Sprintf("%.0f", n)
				}
			case int:
				if n > 0 {
					return fmt.Sprintf("%d", n)
				}
			}
		}
		if nodes, ok := data["nodes"].([]any); ok {
			if len(nodes) > 0 {
				return fmt.Sprintf("%d", len(nodes))
			}
		}
	}
	return ""
}
