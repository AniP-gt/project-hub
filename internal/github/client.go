package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"project-hub/internal/github/parse"
	"project-hub/internal/state"
)

type Direction string

const (
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

type Client interface {
	FetchProject(ctx context.Context, projectID string, owner string, filter string, limit int) (state.Project, []state.Item, error)
	FetchItems(ctx context.Context, projectID string, owner string, filter string, limit int) ([]state.Item, error)
	CreateIssue(ctx context.Context, projectID string, owner string, repo string, title string, body string) (state.Item, error)
	UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string) (state.Item, error)
	UpdateField(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string, fieldName string) (state.Item, error)
	UpdateLabels(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, labels []string) (state.Item, error)
	UpdateMilestone(ctx context.Context, projectID string, owner string, itemID string, milestone string) (state.Item, error)
	UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, userLogins []string) (state.Item, error)
	UpdateItem(ctx context.Context, projectID string, owner string, item state.Item, title string, description string) (state.Item, error)
	UpdateIssueBody(ctx context.Context, repo string, number int, body string) error
	AddIssueComment(ctx context.Context, repo string, number int, body string) error
	FetchIssueDetail(ctx context.Context, repo string, number int) (state.Item, error)
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

func (c *CLIClient) CreateIssue(ctx context.Context, projectID string, owner string, repo string, title string, body string) (state.Item, error) {
	projectNumber, err := strconv.Atoi(strings.TrimSpace(projectID))
	if err != nil || projectNumber <= 0 {
		return state.Item{}, fmt.Errorf("project number required to create issue: %q", projectID)
	}

	repo = strings.TrimSpace(repo)
	if repo == "" {
		return state.Item{}, fmt.Errorf("repository is required")
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return state.Item{}, fmt.Errorf("issue title is required")
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return state.Item{}, fmt.Errorf("issue body is required")
	}

	issueURLBytes, err := c.runGh(ctx, "issue", "create", "--repo", repo, "--title", title, "--body", body)
	if err != nil {
		return state.Item{}, fmt.Errorf("gh issue create failed: %w", err)
	}

	issueURL := strings.TrimSpace(string(issueURLBytes))
	if issueURL == "" {
		return state.Item{}, fmt.Errorf("gh issue create returned empty issue URL")
	}

	args := []string{"project", "item-add", strconv.Itoa(projectNumber), "--url", issueURL, "--format", "json"}
	if owner != "" {
		args = append(args, "--owner", owner)
	}

	out, err := c.runGh(ctx, args...)
	if err != nil {
		return state.Item{}, fmt.Errorf("gh project item-add failed: %w", err)
	}

	var rawItem map[string]any
	if err := json.Unmarshal(out, &rawItem); err != nil {
		return state.Item{}, fmt.Errorf("parse gh project item-add json: %w", err)
	}

	item, ok := parse.ParseItemMap(rawItem)
	if !ok {
		return state.Item{}, fmt.Errorf("failed to parse created project item from gh project item-add output")
	}

	if item.URL == "" {
		item.URL = issueURL
	}
	if item.Title == "" {
		item.Title = title
	}
	if item.Repository == "" {
		item.Repository = repo
	}
	if item.Type == "" {
		item.Type = "Issue"
	}

	return item, nil
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

func (c *CLIClient) UpdateIssueBody(ctx context.Context, repo string, number int, body string) error {
	repo = strings.TrimSpace(repo)
	if repo == "" || number <= 0 {
		return fmt.Errorf("cannot edit issue body: missing repository or issue number")
	}

	return c.runGhWithStdin(ctx, body, "issue", "edit", strconv.Itoa(number), "--repo", repo, "--body-file", "-")
}

func (c *CLIClient) AddIssueComment(ctx context.Context, repo string, number int, body string) error {
	repo = strings.TrimSpace(repo)
	if repo == "" || number <= 0 {
		return fmt.Errorf("cannot add issue comment: missing repository or issue number")
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("comment body is required")
	}

	return c.runGhWithStdin(ctx, body, "issue", "comment", strconv.Itoa(number), "--repo", repo, "--body-file", "-")
}

func (c *CLIClient) FetchIssueDetail(ctx context.Context, repo string, number int) (state.Item, error) {
	args := []string{"issue", "view", strconv.Itoa(number), "--repo", repo, "--json", "body,comments"}
	out, err := c.runGh(ctx, args...)
	if err != nil {
		return state.Item{}, fmt.Errorf("gh issue view failed: %w", err)
	}

	var result struct {
		Body     string          `json:"body"`
		Comments json.RawMessage `json:"comments"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return state.Item{}, fmt.Errorf("parse gh issue view json: %w", err)
	}

	comments, err := parseIssueComments(result.Comments)
	if err != nil {
		return state.Item{}, fmt.Errorf("parse gh issue comments json: %w", err)
	}

	return state.Item{Description: result.Body, Comments: comments}, nil
}

func parseIssueComments(raw json.RawMessage) ([]state.Comment, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	type rawComment struct {
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		Body      string     `json:"body"`
		CreatedAt *time.Time `json:"createdAt"`
	}

	toStateComments := func(rawComments []rawComment) []state.Comment {
		comments := make([]state.Comment, 0, len(rawComments))
		for _, comment := range rawComments {
			author := strings.TrimSpace(comment.Author.Login)
			if author == "" {
				author = "unknown"
			}
			comments = append(comments, state.Comment{
				Author:    author,
				Body:      comment.Body,
				CreatedAt: comment.CreatedAt,
			})
		}
		return comments
	}

	var flat []rawComment
	if err := json.Unmarshal(raw, &flat); err == nil {
		return toStateComments(flat), nil
	}

	var nested struct {
		Nodes []rawComment `json:"nodes"`
	}
	if err := json.Unmarshal(raw, &nested); err == nil {
		return toStateComments(nested.Nodes), nil
	}

	return nil, fmt.Errorf("unsupported comments shape")
}

func (c *CLIClient) runGh(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.GhPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func (c *CLIClient) runGhWithStdin(ctx context.Context, stdin string, args ...string) error {
	cmd := exec.CommandContext(ctx, c.GhPath, args...)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
