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

// FetchProject calls `gh project view <projectID> --format json` and extracts basic fields.
func (c *CLIClient) FetchProject(ctx context.Context, projectID string, owner string) (state.Project, []state.Item, error) {
	args := []string{"project", "view", projectID, "--format", "json"}
	if owner != "" {
		args = append(args, "--owner", owner)
	}
	out, err := runGh(ctx, args...)
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
	items := extractItems(raw)
	return proj, items, nil
}

func (c *CLIClient) FetchItems(ctx context.Context, projectID string, owner string, filter string) ([]state.Item, error) {
	_ = filter
	_ = owner
	return nil, errors.New("FetchItems not implemented yet")
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
	out, err := cmd.Output()
	if err != nil {
		return nil, err
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
		m, ok := r.(map[string]any)
		if !ok {
			continue
		}
		item := state.Item{}
		if id, ok := m["id"].(string); ok {
			item.ID = id
		}
		if title, ok := m["title"].(string); ok {
			item.Title = title
		}
		if desc, ok := m["body"].(string); ok {
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
		items = append(items, item)
	}
	return items
}
