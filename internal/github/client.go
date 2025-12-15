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
	UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, assigneeFieldID string, userLogins []string) (state.Item, error)
	UpdateItem(ctx context.Context, projectID string, owner string, item state.Item, title string, description string) (state.Item, error)
	FetchRoadmap(ctx context.Context, projectID string, owner string) ([]state.Timeline, []state.Item, error)
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

func (c *CLIClient) UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, assigneeFieldID string, userLogins []string) (state.Item, error) {
	args := []string{
		"project", "item-edit",
		"--id", itemID,
		"--project-id", projectID,
		"--field-id", assigneeFieldID,
		"--format", "json",
	}

	if len(userLogins) > 0 && userLogins[0] != "" {
		args = append(args, "--text", userLogins[0])
	} else {
		// To clear the assignee, use the --clear flag
		args = append(args, "--clear")
	}

	out, err := c.runGh(ctx, args...)
	if err != nil {
		return state.Item{}, fmt.Errorf("gh project item-edit for assignees failed: %w", err)
	}

	var rawItem map[string]any
	if err := json.Unmarshal(out, &rawItem); err != nil {
		return state.Item{}, fmt.Errorf("parse gh project item-edit json for assignees: %w", err)
	}

	item, ok := parseItemMap(rawItem)
	if !ok {
		return state.Item{}, fmt.Errorf("failed to parse updated item from gh project item-edit output for assignees")
	}

	return item, nil
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
	}
	if desc, ok := m["body"].(string); ok && item.Description == "" {
		item.Description = desc
	}
	if repo, ok := m["repository"].(string); ok && item.Repository == "" {
		item.Repository = repo
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

	// Also check "fieldValues" array for status
	if item.Status == "" { // Only if status not found yet
		if fv, ok := m["fieldValues"].([]any); ok {
			for _, v := range fv {
				if fm, ok := v.(map[string]any); ok {
					if fname, ok := fm["fieldName"].(string); ok && fname == "Status" {
						if opt, ok := fm["singleSelectOption"].(map[string]any); ok {
							if name, ok := opt["name"].(string); ok && name != "" {
								item.Status = name
								break // Found status, no need to check other fieldValues
							} else if id, ok := opt["id"].(string); ok && id != "" {
								item.Status = id
								break // Found status, no need to check other fieldValues
							}
						}
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
	if labels, ok := m["labels"].([]any); ok {
		for _, lv := range labels {
			if lm, ok := lv.(map[string]any); ok {
				if name, ok := lm["name"].(string); ok {
					item.Labels = append(item.Labels, name)
				}
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
