package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
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
	FetchProject(ctx context.Context, projectID string, owner string) (state.Project, []state.Item, error)
	FetchItems(ctx context.Context, projectID string, owner string, filter string) ([]state.Item, error)
	UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, direction Direction) (state.Item, error)
	UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, assigneeIDs []string) (state.Item, error)
	UpdateItem(ctx context.Context, projectID string, owner string, itemID string, title string, description string) (state.Item, error)
	FetchRoadmap(ctx context.Context, projectID string, owner string) ([]state.Timeline, []state.Item, error)
}

// CLIClient is a thin wrapper intended to call `gh` and parse JSON.
type CLIClient struct{}

// NewCLIClient constructs a default CLI client.
func NewCLIClient() *CLIClient {
	return &CLIClient{}
}

// FetchProject calls `gh project view ...` for metadata and `gh project item-list ...` for items.
func (c *CLIClient) FetchProject(ctx context.Context, projectID string, owner string) (state.Project, []state.Item, error) {
	viewArgs := []string{"project", "view", projectID, "--format", "json", "--fields", "title,views,number,owner,public,shortDescription"}
	if owner != "" {
		viewArgs = append(viewArgs, "--owner", owner)
	}
	out, err := runGh(ctx, viewArgs...)
	if err != nil {
		return state.Project{}, nil, fmt.Errorf("gh project view failed: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return state.Project{}, nil, fmt.Errorf("parse gh project view json: %w", err)
	}
	proj := state.Project{ID: projectID}
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

	items, err := c.FetchItems(ctx, projectID, owner, "")
	if err != nil {
		return proj, nil, err
	}
	return proj, items, nil
}

func (c *CLIClient) FetchItems(ctx context.Context, projectID string, owner string, filter string) ([]state.Item, error) {
	args := []string{"project", "item-list", projectID, "--format", "json", "--limit", "100"}
	if owner != "" {
		args = append(args, "--owner", owner)
	}
	if filter != "" {
		args = append(args, "--search", filter)
	}
	out, err := runGh(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("gh project item-list failed: %w", err)
	}
	items, parseErr := parseItemList(out)
	if parseErr != nil {
		return nil, parseErr
	}
	return items, nil
}

func (c *CLIClient) UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, direction Direction) (state.Item, error) {
	_ = projectID
	_ = owner
	_ = itemID
	_ = direction
	return state.Item{}, errors.New("UpdateStatus not implemented yet")
}

func (c *CLIClient) UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, assigneeIDs []string) (state.Item, error) {
	_ = projectID
	_ = owner
	_ = itemID
	_ = assigneeIDs
	return state.Item{}, errors.New("UpdateAssignees not implemented yet")
}

func (c *CLIClient) UpdateItem(ctx context.Context, projectID string, owner string, itemID string, title string, description string) (state.Item, error) {
	_ = owner
	// TODO: call `gh project item-edit` or equivalent once wiring is added.
	return state.Item{ID: itemID, Title: title, Description: description}, nil
}

func (c *CLIClient) FetchRoadmap(ctx context.Context, projectID string, owner string) ([]state.Timeline, []state.Item, error) {
	_ = projectID
	_ = owner
	return nil, nil, errors.New("FetchRoadmap not implemented yet")
}

func runGh(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func extractItems(raw map[string]any) []state.Item {
	var items []state.Item
	rawItems, ok := raw["items"].([]any)
	if !ok {
		return items
	}
	for _, r := range rawItems {
		if it, ok := parseItemMap(r); ok {
			items = append(items, it)
		}
	}
	return items
}

func parseItemList(out []byte) ([]state.Item, error) {
	var items []any
	// item-list returns either an array or an object with "items" field.
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
		if t, ok := content["title"].(string); ok && item.Title == "" {
			item.Title = t
		}
		if body, ok := content["body"].(string); ok {
			item.Description = body
		}
	}
	if desc, ok := m["body"].(string); ok && item.Description == "" {
		item.Description = desc
	}
	if status, ok := m["status"].(string); ok {
		item.Status = status
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
