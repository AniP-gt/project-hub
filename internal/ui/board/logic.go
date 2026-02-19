package board

import (
	"sort"
	"strings"

	"project-hub/internal/state"
)

var ColumnOrder = []string{"Todo", "Draft", "In Progress", "In_Review"}

func isDoneStatus(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "done")
}

func groupItemsByStatus(items []state.Item, fields []state.Field) []state.Column {
	statusMap := make(map[string][]state.Item)
	for _, item := range items {
		statusMap[item.Status] = append(statusMap[item.Status], item)
	}

	for status, items := range statusMap {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Position < items[j].Position
		})
		statusMap[status] = items
	}

	statusCardMap := make(map[string][]state.Card)
	for status, items := range statusMap {
		for _, item := range items {
			assignee := ""
			if len(item.Assignees) > 0 {
				assignee = item.Assignees[0]
			}
			priority := item.Priority
			if priority == "" {
				priority = inferPriorityFromLabels(item.Labels)
			}
			card := state.Card{
				ID:               item.ID,
				Title:            item.Title,
				Assignee:         assignee,
				Labels:           item.Labels,
				Status:           item.Status,
				Priority:         priority,
				Milestone:        item.Milestone,
				Repository:       item.Repository,
				SubIssueProgress: item.SubIssueProgress,
				ParentIssue:      item.ParentIssue,
			}
			statusCardMap[status] = append(statusCardMap[status], card)
		}
	}

	var statusOrder []string
	for _, field := range fields {
		if field.Name == "Status" {
			for _, opt := range field.Options {
				statusOrder = append(statusOrder, opt.Name)
			}
			break
		}
	}

	if len(statusOrder) == 0 {
		statusOrder = ColumnOrder
	}

	var columns []state.Column
	var doneColumn *state.Column

	for _, status := range statusOrder {
		if cards, exists := statusCardMap[status]; exists {
			columns = append(columns, state.Column{Name: status, Cards: cards})
			delete(statusCardMap, status)
		}
	}

	var unknownStatuses []string
	for status := range statusCardMap {
		if isDoneStatus(status) {
			doneColumn = &state.Column{Name: "Done", Cards: statusCardMap[status]}
		} else {
			unknownStatuses = append(unknownStatuses, status)
		}
	}

	sort.Strings(unknownStatuses)

	for _, status := range unknownStatuses {
		columns = append(columns, state.Column{Name: status, Cards: statusCardMap[status]})
	}

	if doneColumn != nil {
		columns = append(columns, *doneColumn)
	}

	return columns
}

type GroupBucket struct {
	Name  string
	Items []state.Item
}

func GroupItemsByAssignee(items []state.Item) []GroupBucket {
	buckets := map[string][]state.Item{}
	for _, item := range items {
		if len(item.Assignees) == 0 {
			buckets["Unassigned"] = append(buckets["Unassigned"], item)
			continue
		}
		for _, assignee := range item.Assignees {
			name := strings.TrimSpace(assignee)
			if name == "" {
				continue
			}
			buckets[name] = append(buckets[name], item)
		}
	}
	return bucketsToOrderedList(buckets, []string{"Unassigned"})
}

func GroupItemsByIteration(items []state.Item) []GroupBucket {
	buckets := map[string][]state.Item{}
	for _, item := range items {
		name := strings.TrimSpace(item.IterationName)
		if name == "" {
			name = "No Iteration"
		}
		buckets[name] = append(buckets[name], item)
	}
	return bucketsToOrderedList(buckets, []string{"No Iteration"})
}

func GroupItemsByStatusBuckets(items []state.Item, fields []state.Field) []GroupBucket {
	statusMap := make(map[string][]state.Item)
	for _, item := range items {
		statusMap[item.Status] = append(statusMap[item.Status], item)
	}

	for status, items := range statusMap {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Position < items[j].Position
		})
		statusMap[status] = items
	}

	var statusOrder []string
	for _, field := range fields {
		if field.Name == "Status" {
			for _, opt := range field.Options {
				statusOrder = append(statusOrder, opt.Name)
			}
			break
		}
	}

	if len(statusOrder) == 0 {
		statusOrder = ColumnOrder
	}

	var buckets []GroupBucket
	var doneBucket *GroupBucket

	for _, status := range statusOrder {
		if items, exists := statusMap[status]; exists {
			buckets = append(buckets, GroupBucket{Name: status, Items: items})
			delete(statusMap, status)
		}
	}

	var unknownStatuses []string
	for status := range statusMap {
		if isDoneStatus(status) {
			bucket := GroupBucket{Name: "Done", Items: statusMap[status]}
			doneBucket = &bucket
		} else {
			unknownStatuses = append(unknownStatuses, status)
		}
	}

	sort.Strings(unknownStatuses)
	for _, status := range unknownStatuses {
		buckets = append(buckets, GroupBucket{Name: status, Items: statusMap[status]})
	}

	if doneBucket != nil {
		buckets = append(buckets, *doneBucket)
	}

	return buckets
}

func bucketsToOrderedList(buckets map[string][]state.Item, trailing []string) []GroupBucket {
	var names []string
	for name := range buckets {
		names = append(names, name)
	}
	sort.Strings(names)
	trailingSet := map[string]struct{}{}
	for _, t := range trailing {
		trailingSet[t] = struct{}{}
	}
	var ordered []GroupBucket
	for _, name := range names {
		if _, ok := trailingSet[name]; ok {
			continue
		}
		ordered = append(ordered, GroupBucket{Name: name, Items: buckets[name]})
	}
	for _, name := range trailing {
		if items, ok := buckets[name]; ok {
			ordered = append(ordered, GroupBucket{Name: name, Items: items})
		}
	}
	return ordered
}

func inferPriorityFromLabels(labels []string) string {
	joined := strings.ToLower(strings.Join(labels, " "))
	switch {
	case strings.Contains(joined, "high"):
		return "High"
	case strings.Contains(joined, "low"):
		return "Low"
	}
	return ""
}
