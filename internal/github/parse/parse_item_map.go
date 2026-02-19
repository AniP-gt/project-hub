package parse

import (
	"fmt"
	"strings"
	"time"

	"project-hub/internal/state"
)

func ParseItemMap(r any) (state.Item, bool) {
	m, ok := r.(map[string]any)
	if !ok {
		return state.Item{}, false
	}
	item := state.Item{}
	item.FieldValues = make(map[string][]string)
	if id, ok := m["id"].(string); ok {
		item.ID = id
	}
	if title, ok := m["title"].(string); ok {
		item.Title = title
	}
	if content, ok := m["content"].(map[string]any); ok {
		if contentID, ok := content["id"].(string); ok {
			item.ContentID = contentID
		}
		if contentType, ok := content["type"].(string); ok {
			item.Type = contentType
		}
		if t, ok := content["title"].(string); ok && item.Title == "" {
			item.Title = t
		}
		if body, ok := content["body"].(string); ok {
			item.Description = body
		}
		if number, ok := content["number"].(float64); ok {
			item.Number = int(number)
		}
		if url, ok := content["url"].(string); ok {
			item.URL = url
		}
		if st, ok := content["state"].(string); ok && item.Status == "" {
			item.Status = st
		}
		if repo, ok := content["repository"].(string); ok && item.Repository == "" {
			item.Repository = repo
		}
		if item.Milestone == "" {
			item.Milestone = parseMilestoneValue(content["milestone"])
		}
		item.Assignees = mergeStringsUnique(item.Assignees, extractAssigneeLogins(content["assignees"]))
		item.Labels = mergeStringsUnique(item.Labels, extractLabelNames(content["labels"]))
	}
	if desc, ok := m["body"].(string); ok && item.Description == "" {
		item.Description = desc
	}
	if repo, ok := m["repository"].(string); ok && item.Repository == "" {
		item.Repository = repo
	}
	if item.Milestone == "" {
		item.Milestone = parseMilestoneValue(m["milestone"])
	}
	if statusVal, ok := m["status"]; ok {
		switch sv := statusVal.(type) {
		case string:
			item.Status = sv
		case map[string]any:
			if name, ok := sv["name"].(string); ok && name != "" {
				item.Status = name
			} else if id, ok := sv["id"].(string); ok && id != "" {
				item.Status = id
			}
		}
	}

	if fv, ok := m["fieldValues"].([]any); ok {
		for _, v := range fv {
			fm, ok := v.(map[string]any)
			if !ok {
				continue
			}

			fieldName := ""
			if fname, ok := fm["fieldName"].(string); ok && fname != "" {
				fieldName = fname
			} else if fieldObj, ok := fm["field"].(map[string]any); ok {
				if nestedName, ok := fieldObj["name"].(string); ok && nestedName != "" {
					fieldName = nestedName
				}
			}

			if fieldName != "" {
				normalizedField := strings.ToLower(strings.TrimSpace(fieldName))
				compactField := compactFieldName(normalizedField)
				switch {
				case normalizedField == "status":
					if item.Status == "" {
						if opt, ok := fm["singleSelectOption"].(map[string]any); ok {
							if name, ok := opt["name"].(string); ok && name != "" {
								item.Status = name
							} else if id, ok := opt["id"].(string); ok && id != "" {
								item.Status = id
							}
						}
					}
				case normalizedField == "priority":
					if item.Priority == "" {
						if opt, ok := fm["singleSelectOption"].(map[string]any); ok {
							if name, ok := opt["name"].(string); ok && name != "" {
								item.Priority = name
							} else if id, ok := opt["id"].(string); ok && id != "" {
								item.Priority = id
							}
						} else if text, ok := fm["text"].(string); ok && text != "" {
							item.Priority = text
						}
					}
				case normalizedField == "assignees":
					if len(item.Assignees) == 0 {
						if users, ok := fm["users"]; ok {
							item.Assignees = mergeStringsUnique(item.Assignees, extractAssigneeLogins(users))
						} else if text, ok := fm["text"].(string); ok && text != "" {
							item.Assignees = mergeStringsUnique(item.Assignees, splitListField(text))
						}
					}
				case normalizedField == "labels":
					if len(item.Labels) == 0 {
						if labelsVal, ok := fm["labels"]; ok {
							item.Labels = mergeStringsUnique(item.Labels, extractLabelNames(labelsVal))
						} else if text, ok := fm["text"].(string); ok && text != "" {
							item.Labels = mergeStringsUnique(item.Labels, splitListField(text))
						}
					}
				case normalizedField == "milestone":
					if item.Milestone == "" {
						if val, ok := fm["milestone"]; ok {
							item.Milestone = parseMilestoneValue(val)
						} else if text, ok := fm["text"].(string); ok && text != "" {
							item.Milestone = text
						} else if opt, ok := fm["singleSelectOption"].(map[string]any); ok {
							item.Milestone = parseMilestoneValue(opt)
						}
					}
				case isSubIssueField(compactField):
					if item.SubIssueProgress == "" {
						if val, ok := fm["number"]; ok {
							switch n := val.(type) {
							case float64:
								if n > 0 {
									item.SubIssueProgress = fmt.Sprintf("%.0f", n)
								}
							case int:
								if n > 0 {
									item.SubIssueProgress = fmt.Sprintf("%d", n)
								}
							}
						} else if text, ok := fm["text"].(string); ok && text != "" {
							item.SubIssueProgress = text
						}
					}
				case isParentIssueField(compactField):
					if item.ParentIssue == "" {
						if text, ok := fm["text"].(string); ok && text != "" {
							item.ParentIssue = text
						} else if title, ok := fm["title"].(string); ok && title != "" {
							item.ParentIssue = title
						}
					}
				}

				item.FieldValues = addFieldValue(item.FieldValues, fieldName, fm)
			}

			for _, sub := range fm {
				if subMap, ok := sub.(map[string]any); ok {
					if applyIterationMetadata(&item, subMap) {
						break
					}
				}
			}
		}
	}

	if item.ParentIssue == "" {
		if content, ok := m["content"].(map[string]any); ok {
			if parentVal, ok := content["parent"]; ok {
				item.ParentIssue = parseParentIssueValue(parentVal)
			} else if parentVal, ok := content["parentIssue"]; ok {
				item.ParentIssue = parseParentIssueValue(parentVal)
			}
		}
		if item.ParentIssue == "" {
			if parentVal, ok := m["parent"]; ok {
				item.ParentIssue = parseParentIssueValue(parentVal)
			}
		}
	}

	if item.SubIssueProgress == "" {
		if content, ok := m["content"].(map[string]any); ok {
			if subVal, ok := content["subIssues"]; ok {
				item.SubIssueProgress = parseSubIssueCount(subVal)
			} else if subVal, ok := content["subIssue"]; ok {
				item.SubIssueProgress = parseSubIssueCount(subVal)
			}
		}
		if item.SubIssueProgress == "" {
			if subVal, ok := m["subIssues"]; ok {
				item.SubIssueProgress = parseSubIssueCount(subVal)
			}
		}
	}

	if item.Status == "" {
		item.Status = "Unknown"
	}

	if priorityVal, ok := m["priority"]; ok {
		switch pv := priorityVal.(type) {
		case string:
			item.Priority = pv
		case map[string]any:
			if name, ok := pv["name"].(string); ok && name != "" {
				item.Priority = name
			} else if id, ok := pv["id"].(string); ok && id != "" {
				item.Priority = id
			}
		}
	}
	item.Assignees = mergeStringsUnique(item.Assignees, extractAssigneeLogins(m["assignees"]))
	item.Labels = mergeStringsUnique(item.Labels, extractLabelNames(m["labels"]))
	for _, val := range m {
		if sub, ok := val.(map[string]any); ok {
			if applyIterationMetadata(&item, sub) {
				break
			}
		}
	}
	if updated, ok := m["updatedAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updated); err == nil {
			item.UpdatedAt = &t
		}
	}
	return item, true
}
