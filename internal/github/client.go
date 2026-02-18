package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"project-hub/internal/state"
)

// Direction represents left/right movement in board context.
type Direction string

const (
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

// Client defines the operations needed from gh CLI for Projects.
type Client interface {
	FetchProject(ctx context.Context, projectID string, owner string, limit int) (state.Project, []state.Item, error)
	FetchItems(ctx context.Context, projectID string, owner string, filter string, limit int) ([]state.Item, error)
	UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string) (state.Item, error)
	UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, userLogins []string) (state.Item, error)
	UpdateItem(ctx context.Context, projectID string, owner string, item state.Item, title string, description string) (state.Item, error)
	FetchRoadmap(ctx context.Context, projectID string, owner string) ([]state.Timeline, []state.Item, error)
	FetchIssueDetail(ctx context.Context, repo string, number int) (string, error)
}

// CLIClient is a thin wrapper intended to call `gh` and parse JSON.
type CLIClient struct {
	GhPath string
}

// NewCLIClient constructs a default CLI client.
func NewCLIClient(ghPath string) *CLIClient {
	if ghPath == "" {
		ghPath = "gh"
	}
	return &CLIClient{GhPath: ghPath}
}

// FetchProject calls `gh project view ...` for metadata and `gh project item-list ...` for items.
func (c *CLIClient) FetchProject(ctx context.Context, projectID string, owner string, limit int) (state.Project, []state.Item, error) {
	viewArgs := []string{"project", "view", projectID, "--format", "json"}
	if owner != "" {
		viewArgs = append(viewArgs, "--owner", owner)
	}
	out, err := c.runGh(ctx, viewArgs...)
	if err != nil {
		return state.Project{}, nil, fmt.Errorf("gh project view failed: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return state.Project{}, nil, fmt.Errorf("parse gh project view json: %w", err)
	}
	proj := state.Project{ID: projectID}
	if id, ok := raw["id"].(string); ok && id != "" {
		proj.NodeID = id
	}
	if own, ok := raw["owner"].(map[string]any); ok {
		if login, ok := own["login"].(string); ok && login != "" {
			proj.Owner = login
		}
	}
	if name, ok := raw["title"].(string); ok {
		proj.Name = name
	}
	if views, ok := raw["views"].([]any); ok {
		for _, v := range views {
			if m, ok := v.(map[string]any); ok {
				if typ, ok := m["type"].(string); ok {
					proj.Views = append(proj.Views, state.ViewType(strings.ToLower(typ)))
				}
			}
		}
	}

	// --- Fetch Fields ---
	fieldArgs := []string{"project", "field-list", projectID, "--format", "json"}
	if owner != "" {
		fieldArgs = append(fieldArgs, "--owner", owner)
	}
	fieldsOut, err := c.runGh(ctx, fieldArgs...)
	if err != nil {
		// Non-fatal, we can continue without field data
	} else {
		var rawFields struct {
			Fields []struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Options []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"options"`
			} `json:"fields"`
		}
		if err := json.Unmarshal(fieldsOut, &rawFields); err == nil {
			for _, rf := range rawFields.Fields {
				field := state.Field{
					ID:   rf.ID,
					Name: rf.Name,
				}
				for _, ro := range rf.Options {
					field.Options = append(field.Options, state.Option{
						ID:   ro.ID,
						Name: ro.Name,
					})
				}
				proj.Fields = append(proj.Fields, field)
			}
		}
	}

	items, err := c.FetchItems(ctx, projectID, owner, "", limit)
	if err != nil {
		return proj, nil, err
	}
	return proj, items, nil
}

func (c *CLIClient) FetchItems(ctx context.Context, projectID string, owner string, filter string, limit int) ([]state.Item, error) {
	limitStr := strconv.Itoa(limit)
	args := []string{"project", "item-list", projectID, "--format", "json", "--limit", limitStr}
	if owner != "" {
		args = append(args, "--owner", owner)
	}
	if filter != "" {
		args = append(args, "--search", filter)
	}
	out, err := c.runGh(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("gh project item-list failed: %w", err)
	}
	items, parseErr := parseItemList(out)
	if parseErr != nil {
		return nil, parseErr
	}
	return items, nil
}

func (c *CLIClient) UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string) (state.Item, error) {
	if err := ValidateStatusUpdateIDs(projectID, itemID, fieldID, optionID); err != nil {
		return state.Item{}, fmt.Errorf("validation failed: %w", err)
	}

	args := []string{
		"project", "item-edit",
		"--id", itemID,
		"--project-id", projectID,
		"--field-id", fieldID,
		"--single-select-option-id", optionID,
		"--format", "json",
	}

	out, err := c.runGh(ctx, args...)
	if err != nil {
		return state.Item{}, fmt.Errorf("gh project item-edit for status failed: %w", err)
	}

	var rawItem map[string]any
	if err := json.Unmarshal(out, &rawItem); err != nil {
		return state.Item{}, fmt.Errorf("parse gh project item-edit json for status: %w", err)
	}

	item, ok := parseItemMap(rawItem)
	if !ok {
		return state.Item{}, fmt.Errorf("failed to parse updated item from gh project item-edit output for status")
	}

	return item, nil
}

func (c *CLIClient) UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, userLogins []string) (state.Item, error) {
	if itemType != "Issue" && itemType != "PullRequest" {
		return state.Item{}, fmt.Errorf("cannot assign to item of type: %s (only Issues and PullRequests can be assigned)", itemType)
	}

	if repo == "" || number == 0 {
		return state.Item{}, fmt.Errorf("cannot edit assignees: missing repository or issue number")
	}

	editCmd := "issue"
	if itemType == "PullRequest" {
		editCmd = "pr"
	}

	editArgs := []string{editCmd, "edit", strconv.Itoa(number), "--repo", repo}

	if len(userLogins) > 0 && userLogins[0] != "" {
		editArgs = append(editArgs, "--add-assignee")
		editArgs = append(editArgs, userLogins...)
	}

	_, err := c.runGh(ctx, editArgs...)
	if err != nil {
		return state.Item{}, fmt.Errorf("gh %s edit for assignees failed: %w", editCmd, err)
	}

	return state.Item{ID: itemID, Assignees: userLogins}, nil
}

func (c *CLIClient) UpdateItem(ctx context.Context, projectID string, owner string, item state.Item, title string, description string) (state.Item, error) {
	// If the item is a real issue, we must use `gh issue edit`.
	if item.Type == "Issue" {
		if item.Number == 0 || item.Repository == "" {
			return state.Item{}, fmt.Errorf("cannot edit issue without number or repository")
		}

		args := []string{
			"issue", "edit",
			strconv.Itoa(item.Number),
			"--repo", item.Repository,
		}
		if title != "" {
			args = append(args, "--title", title)
		}
		if description != "" {
			args = append(args, "--body", description)
		}

		_, err := c.runGh(ctx, args...)
		if err != nil {
			return state.Item{}, fmt.Errorf("gh issue edit failed: %w", err)
		}

		// `gh issue edit` does not return the updated item, so we update it manually.
		item.Title = title
		item.Description = description
		return item, nil
	}

	// For draft issues, use `gh project item-edit`.
	idToUse := item.ID
	if strings.HasPrefix(item.ContentID, "DI_") {
		idToUse = item.ContentID
	}

	args := []string{"project", "item-edit", "--id", idToUse, "--project-id", projectID, "--format", "json"}
	if title != "" {
		args = append(args, "--title", title)
	}
	if description != "" {
		args = append(args, "--body", description)
	}

	out, err := c.runGh(ctx, args...)
	if err != nil {
		return state.Item{}, fmt.Errorf("gh project item-edit failed: %w", err)
	}

	var rawItem map[string]any
	if err := json.Unmarshal(out, &rawItem); err != nil {
		return state.Item{}, fmt.Errorf("parse gh project item-edit json: %w", err)
	}

	updatedItem, ok := parseItemMap(rawItem)
	if !ok {
		return state.Item{}, fmt.Errorf("failed to parse updated item from gh project item-edit output")
	}

	return updatedItem, nil
}

func (c *CLIClient) FetchRoadmap(ctx context.Context, projectID string, owner string) ([]state.Timeline, []state.Item, error) {
	_ = projectID
	_ = owner
	return nil, nil, errors.New("FetchRoadmap not implemented yet")
}

func (c *CLIClient) FetchIssueDetail(ctx context.Context, repo string, number int) (string, error) {
	args := []string{"issue", "view", strconv.Itoa(number), "--repo", repo, "--json", "body"}
	out, err := c.runGh(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("gh issue view failed: %w", err)
	}

	var result struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return "", fmt.Errorf("parse gh issue view json: %w", err)
	}

	return result.Body, nil
}

func (c *CLIClient) runGh(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.GhPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func parseItemList(out []byte) ([]state.Item, error) {
	var items []any
	if err := json.Unmarshal(out, &items); err != nil {
		var obj map[string]any
		if err2 := json.Unmarshal(out, &obj); err2 != nil {
			return nil, fmt.Errorf("parse item-list json: %w", err)
		}
		if arr, ok := obj["items"].([]any); ok {
			items = arr
		} else {
			return nil, fmt.Errorf("parse item-list json: items not found")
		}
	}
	var result []state.Item
	for _, r := range items {
		if it, ok := parseItemMap(r); ok {
			result = append(result, it)
		}
	}
	return result, nil
}

func parseItemMap(r any) (state.Item, bool) {
	m, ok := r.(map[string]any)
	if !ok {
		return state.Item{}, false
	}
	item := state.Item{}
	if id, ok := m["id"].(string); ok {
		item.ID = id
	}
	if title, ok := m["title"].(string); ok {
		item.Title = title
	}
	if content, ok := m["content"].(map[string]any); ok {
		// Extract ContentID if available
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
	// Robustly parse status: it may be string, an object with id/name,
	// or described in fieldValues -> singleSelectOption. Prefer human-readable name.
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

			if fname, ok := fm["fieldName"].(string); ok {
				switch fname {
				case "Status":
					if item.Status == "" {
						if opt, ok := fm["singleSelectOption"].(map[string]any); ok {
							if name, ok := opt["name"].(string); ok && name != "" {
								item.Status = name
							} else if id, ok := opt["id"].(string); ok && id != "" {
								item.Status = id
							}
						}
					}
				case "Priority":
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
				case "Assignees":
					if len(item.Assignees) == 0 {
						if users, ok := fm["users"]; ok {
							item.Assignees = mergeStringsUnique(item.Assignees, extractAssigneeLogins(users))
						} else if text, ok := fm["text"].(string); ok && text != "" {
							item.Assignees = mergeStringsUnique(item.Assignees, splitListField(text))
						}
					}
				case "Labels":
					if len(item.Labels) == 0 {
						if labelsVal, ok := fm["labels"]; ok {
							item.Labels = mergeStringsUnique(item.Labels, extractLabelNames(labelsVal))
						} else if text, ok := fm["text"].(string); ok && text != "" {
							item.Labels = mergeStringsUnique(item.Labels, splitListField(text))
						}
					}
				case "Milestone":
					if item.Milestone == "" {
						if val, ok := fm["milestone"]; ok {
							item.Milestone = parseMilestoneValue(val)
						} else if text, ok := fm["text"].(string); ok && text != "" {
							item.Milestone = text
						} else if opt, ok := fm["singleSelectOption"].(map[string]any); ok {
							item.Milestone = parseMilestoneValue(opt)
						}
					}
				}
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
