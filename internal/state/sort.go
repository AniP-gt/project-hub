package state

import (
	"fmt"
	"sort"
	"strings"
)

func ApplyTableSort(items []Item, ts TableSort) []Item {
	if ts.Field == "" {
		return items
	}
	switch ts.Field {
	case "Title":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Title < items[j].Title
			}
			return items[i].Title > items[j].Title
		})
	case "Status":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Status < items[j].Status
			}
			return items[i].Status > items[j].Status
		})
	case "Repository":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Repository < items[j].Repository
			}
			return items[i].Repository > items[j].Repository
		})
	case "Labels":
		sort.SliceStable(items, func(i, j int) bool {
			iLabels := strings.Join(items[i].Labels, ",")
			jLabels := strings.Join(items[j].Labels, ",")
			if ts.Asc {
				return iLabels < jLabels
			}
			return iLabels > jLabels
		})
	case "Milestone":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Milestone < items[j].Milestone
			}
			return items[i].Milestone > items[j].Milestone
		})
	case "SubIssueProgress":
		parseRatio := func(s string) float64 {
			if s == "" {
				return 0.0
			}
			parts := strings.Split(s, "/")
			if len(parts) != 2 {
				return 0.0
			}
			var a, b float64
			if _, err := fmt.Sscan(parts[0], &a); err != nil {
				return 0.0
			}
			if _, err := fmt.Sscan(parts[1], &b); err != nil || b == 0 {
				return 0.0
			}
			return a / b
		}
		sort.SliceStable(items, func(i, j int) bool {
			ri := parseRatio(items[i].SubIssueProgress)
			rj := parseRatio(items[j].SubIssueProgress)
			if ts.Asc {
				return ri < rj
			}
			return ri > rj
		})
	case "Priority":
		priorityRank := func(p string) int {
			switch strings.ToLower(p) {
			case "high":
				return 3
			case "medium":
				return 2
			case "low":
				return 1
			default:
				return 0
			}
		}
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return priorityRank(items[i].Priority) < priorityRank(items[j].Priority)
			}
			return priorityRank(items[i].Priority) > priorityRank(items[j].Priority)
		})
	default:
		if ts.Field == "Assignees" {
			sort.SliceStable(items, func(i, j int) bool {
				iAssignees := strings.Join(items[i].Assignees, ",")
				jAssignees := strings.Join(items[j].Assignees, ",")
				if ts.Asc {
					return iAssignees < jAssignees
				}
				return iAssignees > jAssignees
			})
			break
		}
		switch ts.Field {
		case "Number":
			sort.SliceStable(items, func(i, j int) bool {
				if ts.Asc {
					return items[i].Number < items[j].Number
				}
				return items[i].Number > items[j].Number
			})
		case "CreatedAt":
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].CreatedAt == nil || items[j].CreatedAt == nil {
					return items[i].CreatedAt == nil
				}
				if ts.Asc {
					return items[i].CreatedAt.Before(*items[j].CreatedAt)
				}
				return items[j].CreatedAt.Before(*items[i].CreatedAt)
			})
		case "UpdatedAt":
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].UpdatedAt == nil || items[j].UpdatedAt == nil {
					return items[i].UpdatedAt == nil
				}
				if ts.Asc {
					return items[i].UpdatedAt.Before(*items[j].UpdatedAt)
				}
				return items[j].UpdatedAt.Before(*items[i].UpdatedAt)
			})
		default:
		}
	}
	return items
}
