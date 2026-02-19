package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"project-hub/internal/github/parse"
	"project-hub/internal/state"
)

type Direction string

const (
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

type Client interface {
	FetchProject(ctx context.Context, projectID string, owner string, limit int) (state.Project, []state.Item, error)
	FetchItems(ctx context.Context, projectID string, owner string, filter string, limit int) ([]state.Item, error)
	UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string) (state.Item, error)
	UpdateField(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string, fieldName string) (state.Item, error)
	UpdateLabels(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, labels []string) (state.Item, error)
	UpdateMilestone(ctx context.Context, projectID string, owner string, itemID string, milestone string) (state.Item, error)
	UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, userLogins []string) (state.Item, error)
	UpdateItem(ctx context.Context, projectID string, owner string, item state.Item, title string, description string) (state.Item, error)
	FetchIssueDetail(ctx context.Context, repo string, number int) (string, error)
}

type CLIClient struct {
	GhPath string
}

func NewCLIClient(ghPath string) *CLIClient {
	if ghPath == "" {
		ghPath = "gh"
	}
	return &CLIClient{GhPath: ghPath}
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

	item, ok := parse.ParseItemMap(rawItem)
	if !ok {
		return state.Item{}, fmt.Errorf("failed to parse updated item from gh project item-edit output for status")
	}

	return item, nil
}

func (c *CLIClient) UpdateField(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string, fieldName string) (state.Item, error) {
	if itemID == "" {
		return state.Item{}, fmt.Errorf("item ID is required")
	}
	if fieldID == "" {
		return state.Item{}, fmt.Errorf("field ID is required")
	}
	if optionID == "" {
		return state.Item{}, fmt.Errorf("option ID is required")
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
		return state.Item{}, fmt.Errorf("gh project item-edit for %s failed: %w", fieldName, err)
	}

	var rawItem map[string]any
	if err := json.Unmarshal(out, &rawItem); err != nil {
		return state.Item{}, fmt.Errorf("parse gh project item-edit json for %s: %w", fieldName, err)
	}

	item, ok := parse.ParseItemMap(rawItem)
	if !ok {
		return state.Item{}, fmt.Errorf("failed to parse updated item from gh project item-edit output for %s", fieldName)
	}

	switch fieldName {
	case "Priority":
		item.Priority = optionID
	case "Milestone":
		item.Milestone = optionID
	case "Labels":
		item.Labels = []string{optionID}
	}

	return item, nil
}

func (c *CLIClient) UpdateLabels(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, labels []string) (state.Item, error) {
	if itemType != "Issue" && itemType != "PullRequest" {
		return state.Item{}, fmt.Errorf("cannot edit labels for item of type: %s (only Issues and PullRequests can have labels)", itemType)
	}

	if repo == "" || number == 0 {
		return state.Item{}, fmt.Errorf("cannot edit labels: missing repository or issue number")
	}

	args := []string{"issue", "edit", strconv.Itoa(number), "--repo", repo}

	if len(labels) > 0 {
		args = append(args, "--add-label")
		args = append(args, strings.Join(labels, ","))
	}

	_, err := c.runGh(ctx, args...)
	if err != nil {
		return state.Item{}, fmt.Errorf("gh issue edit for labels failed: %w", err)
	}

	return state.Item{ID: itemID, Labels: labels}, nil
}

func (c *CLIClient) UpdateMilestone(ctx context.Context, projectID string, owner string, itemID string, milestone string) (state.Item, error) {
	if itemID == "" {
		return state.Item{}, fmt.Errorf("item ID is required")
	}

	args := []string{
		"project", "item-edit",
		"--id", itemID,
		"--project-id", projectID,
		"--format", "json",
	}

	if milestone != "" {
		args = append(args, "--milestone", milestone)
	}

	out, err := c.runGh(ctx, args...)
	if err != nil {
		return state.Item{}, fmt.Errorf("gh project item-edit for milestone failed: %w", err)
	}

	var rawItem map[string]any
	if err := json.Unmarshal(out, &rawItem); err != nil {
		return state.Item{}, fmt.Errorf("parse gh project item-edit json for milestone: %w", err)
	}

	item, ok := parse.ParseItemMap(rawItem)
	if !ok {
		return state.Item{}, fmt.Errorf("failed to parse updated item from gh project item-edit output for milestone")
	}

	item.Milestone = milestone
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

		item.Title = title
		item.Description = description
		return item, nil
	}

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

	updatedItem, ok := parse.ParseItemMap(rawItem)
	if !ok {
		return state.Item{}, fmt.Errorf("failed to parse updated item from gh project item-edit output")
	}

	return updatedItem, nil
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
