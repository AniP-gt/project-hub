package github

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"project-hub/internal/github/parse"
	"project-hub/internal/state"
)

func TestParseItemMap(t *testing.T) {
	tests := []struct {
		name      string
		inputJSON string
		wantItem  state.Item
		wantOK    bool
	}{
		{
			name: "Status as string",
			inputJSON: `{
				"id": "I_kwDOJb9WfM57Wp-n",
				"title": "Test Issue",
				"status": "In Progress",
				"updatedAt": "2023-10-27T10:00:00Z"
			}`,
			wantItem: state.Item{
				ID:        "I_kwDOJb9WfM57Wp-n",
				Title:     "Test Issue",
				Status:    "In Progress",
				UpdatedAt: parseTime("2023-10-27T10:00:00Z"),
			},
			wantOK: true,
		},
		{
			name: "Status as object with name",
			inputJSON: `{
				"id": "I_kwDOJb9WfM57Wp-o",
				"title": "Another Issue",
				"status": {
					"id": "FO_123",
					"name": "Done"
				},
				"updatedAt": "2023-10-27T11:00:00Z"
			}`,
			wantItem: state.Item{
				ID:        "I_kwDOJb9WfM57Wp-o",
				Title:     "Another Issue",
				Status:    "Done",
				UpdatedAt: parseTime("2023-10-27T11:00:00Z"),
			},
			wantOK: true,
		},
		{
			name: "Status as object with ID (fallback)",
			inputJSON: `{
				"id": "I_kwDOJb9WfM57Wp-p",
				"title": "Issue with ID status",
				"status": {
					"id": "FO_456"
				},
				"updatedAt": "2023-10-27T12:00:00Z"
			}`,
			wantItem: state.Item{
				ID:        "I_kwDOJb9WfM57Wp-p",
				Title:     "Issue with ID status",
				Status:    "FO_456",
				UpdatedAt: parseTime("2023-10-27T12:00:00Z"),
			},
			wantOK: true,
		},
		{
			name: "Status in fieldValues with singleSelectOption name",
			inputJSON: `{
				"id": "I_kwDOJb9WfM57Wp-q",
				"title": "FieldValues Issue",
				"fieldValues": [
					{"fieldName": "Labels", "labels": {"nodes": [{"name": "bug"}]}},
					{
						"fieldName": "Status",
						"singleSelectOption": {
							"id": "FO_789",
							"name": "Blocked"
						}
					}
				],
				"updatedAt": "2023-10-27T13:00:00Z"
			}`,
			wantItem: state.Item{
				ID:        "I_kwDOJb9WfM57Wp-q",
				Title:     "FieldValues Issue",
				Status:    "Blocked",
				Labels:    []string{"bug"},
				UpdatedAt: parseTime("2023-10-27T13:00:00Z"),
			},
			wantOK: true,
		},
		{
			name: "Status in fieldValues with singleSelectOption ID (fallback)",
			inputJSON: `{
				"id": "I_kwDOJb9WfM57Wp-r",
				"title": "FieldValues Issue ID fallback",
				"fieldValues": [
					{
						"fieldName": "Status",
						"singleSelectOption": {
							"id": "FO_999"
						}
					}
				],
				"updatedAt": "2023-10-27T14:00:00Z"
			}`,
			wantItem: state.Item{
				ID:        "I_kwDOJb9WfM57Wp-r",
				Title:     "FieldValues Issue ID fallback",
				Status:    "FO_999",
				UpdatedAt: parseTime("2023-10-27T14:00:00Z"),
			},
			wantOK: true,
		},
		{
			name: "No status field",
			inputJSON: `{
				"id": "I_kwDOJb9WfM57Wp-s",
				"title": "No Status Issue",
				"updatedAt": "2023-10-27T15:00:00Z"
			}`,
			wantItem: state.Item{
				ID:        "I_kwDOJb9WfM57Wp-s",
				Title:     "No Status Issue",
				Status:    "Unknown",
				UpdatedAt: parseTime("2023-10-27T15:00:00Z"),
			},
			wantOK: true,
		},
		{
			name: "Full item example with content and status",
			inputJSON: `{
				"id": "PVTI_lADOJb9WfM57Wp-nAAgJpM",
				"content": {
					"id": "I_kwDOJb9WfM57Wp-n",
					"type": "Issue",
					"title": "Bug: Fix auth flow",
					"body": "User reported a bug in the authentication process.",
					"number": 123,
					"url": "https://github.com/owner/repo/issues/123",
					"state": "OPEN",
					"repository": "repo"
				},
				"labels": [],
				"milestone": null,
				"assignees": [],
				"status": {
					"id": "FO_123abc",
					"name": "Todo"
				},
				"priority": {
					"id": "PO_456def",
					"name": "High"
				},
				"updatedAt": "2023-10-27T16:00:00Z"
			}`,
			wantItem: state.Item{
				ID:          "PVTI_lADOJb9WfM57Wp-nAAgJpM",
				ContentID:   "I_kwDOJb9WfM57Wp-n",
				Type:        "Issue",
				Title:       "Bug: Fix auth flow",
				Description: "User reported a bug in the authentication process.",
				Number:      123,
				URL:         "https://github.com/owner/repo/issues/123",
				Status:      "Todo",
				Repository:  "repo",
				Priority:    "High",
				UpdatedAt:   parseTime("2023-10-27T16:00:00Z"),
			},
			wantOK: true,
		},
		{
			name: "Assignees and labels from nested nodes",
			inputJSON: `{
				"id": "I_nested_assignees",
				"title": "Issue with nested data",
				"content": {
					"title": "Issue with nested data",
					"assignees": {
						"nodes": [
							{"login": "alice"},
							{"login": "bob"}
						]
					},
					"labels": {
						"nodes": [
							{"name": "bug"},
							{"name": "High Priority"}
						]
					},
					"milestone": {
						"title": "Sprint Alpha"
					}
				},
				"fieldValues": [
					{
						"fieldName": "Priority",
						"singleSelectOption": {
							"name": "High"
						}
					}
				],
				"updatedAt": "2023-10-27T17:00:00Z"
			}`,
			wantItem: state.Item{
				ID:        "I_nested_assignees",
				Title:     "Issue with nested data",
				Status:    "Unknown",
				Labels:    []string{"bug", "High Priority"},
				Assignees: []string{"alice", "bob"},
				Priority:  "High",
				Milestone: "Sprint Alpha",
				UpdatedAt: parseTime("2023-10-27T17:00:00Z"),
			},
			wantOK: true,
		},
		{
			name: "Milestone from field values",
			inputJSON: `{
				"id": "I_field_milestone",
				"title": "Milestone via field",
				"fieldValues": [
					{
						"fieldName": "Milestone",
						"text": "Release 1"
					}
				]
			}`,
			wantItem: state.Item{
				ID:        "I_field_milestone",
				Title:     "Milestone via field",
				Status:    "Unknown",
				Milestone: "Release 1",
			},
			wantOK: true,
		},
		{
			name: "Parent issue and sub-issue titles from content",
			inputJSON: `{
				"id": "I_relationships",
				"title": "Parent task",
				"content": {
					"title": "Parent task",
					"parent": {"title": "Epic task"},
					"subIssues": {
						"totalCount": 2,
						"nodes": [
							{"title": "Child A"},
							{"title": "Child B"}
						]
					}
				}
			}`,
			wantItem: state.Item{
				ID:               "I_relationships",
				Title:            "Parent task",
				Status:           "Unknown",
				ParentIssue:      "Epic task",
				SubIssueProgress: "2",
				SubIssueTitles:   []string{"Child A", "Child B"},
			},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw map[string]any
			if err := json.Unmarshal([]byte(tt.inputJSON), &raw); err != nil {
				t.Fatalf("Failed to unmarshal input JSON: %v", err)
			}
			got, ok := parse.ParseItemMap(raw)
			if ok != tt.wantOK {
				t.Errorf("parseItemMap() ok = %v, want %v", ok, tt.wantOK)
			}
			if got.ID != tt.wantItem.ID {
				t.Errorf("parseItemMap() ID = %v, want %v", got.ID, tt.wantItem.ID)
			}
			if got.Title != tt.wantItem.Title {
				t.Errorf("parseItemMap() Title = %v, want %v", got.Title, tt.wantItem.Title)
			}
			if got.Status != tt.wantItem.Status {
				t.Errorf("parseItemMap() Status = %v, want %v", got.Status, tt.wantItem.Status)
			}
			if got.Repository != tt.wantItem.Repository {
				t.Errorf("parseItemMap() Repository = %v, want %v", got.Repository, tt.wantItem.Repository)
			}
			if got.Priority != tt.wantItem.Priority {
				t.Errorf("parseItemMap() Priority = %v, want %v", got.Priority, tt.wantItem.Priority)
			}
			if got.Milestone != tt.wantItem.Milestone {
				t.Errorf("parseItemMap() Milestone = %v, want %v", got.Milestone, tt.wantItem.Milestone)
			}
			if got.UpdatedAt != nil && tt.wantItem.UpdatedAt != nil {
				if !got.UpdatedAt.Equal(*tt.wantItem.UpdatedAt) {
					t.Errorf("parseItemMap() UpdatedAt = %v, want %v", *got.UpdatedAt, *tt.wantItem.UpdatedAt)
				}
			} else if (got.UpdatedAt == nil && tt.wantItem.UpdatedAt != nil) || (got.UpdatedAt != nil && tt.wantItem.UpdatedAt == nil) {
				t.Errorf("parseItemMap() UpdatedAt mismatch: got %v, want %v", got.UpdatedAt, tt.wantItem.UpdatedAt)
			}
			if !reflect.DeepEqual(got.Assignees, tt.wantItem.Assignees) {
				t.Errorf("parseItemMap() Assignees = %v, want %v", got.Assignees, tt.wantItem.Assignees)
			}
			if !reflect.DeepEqual(got.Labels, tt.wantItem.Labels) {
				t.Errorf("parseItemMap() Labels = %v, want %v", got.Labels, tt.wantItem.Labels)
			}
			if got.ParentIssue != tt.wantItem.ParentIssue {
				t.Errorf("parseItemMap() ParentIssue = %v, want %v", got.ParentIssue, tt.wantItem.ParentIssue)
			}
			if got.SubIssueProgress != tt.wantItem.SubIssueProgress {
				t.Errorf("parseItemMap() SubIssueProgress = %v, want %v", got.SubIssueProgress, tt.wantItem.SubIssueProgress)
			}
			if !reflect.DeepEqual(got.SubIssueTitles, tt.wantItem.SubIssueTitles) {
				t.Errorf("parseItemMap() SubIssueTitles = %v, want %v", got.SubIssueTitles, tt.wantItem.SubIssueTitles)
			}
		})
	}
}

func parseTime(s string) *time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &t
}

func TestFetchProject_ProjectNodeIDAssignment(t *testing.T) {
	tests := []struct {
		name           string
		inputProjectID string
		rawJSON        string
		wantProjectID  string
		wantNodeID     string
	}{
		{
			name:           "numeric input with canonical id in response",
			inputProjectID: "9",
			rawJSON:        `{"id": "PVT_kwDOABC123", "title": "Test Project"}`,
			wantProjectID:  "9",
			wantNodeID:     "PVT_kwDOABC123",
		},
		{
			name:           "numeric input without id in response uses fallback",
			inputProjectID: "9",
			rawJSON:        `{"title": "Test Project"}`,
			wantProjectID:  "9",
			wantNodeID:     "",
		},
		{
			name:           "canonical input preserved as project id and node id",
			inputProjectID: "PVT_kwDOABC123",
			rawJSON:        `{"id": "PVT_kwDOABC123", "title": "Test Project"}`,
			wantProjectID:  "PVT_kwDOABC123",
			wantNodeID:     "PVT_kwDOABC123",
		},
		{
			name:           "different canonical id in response stored in node id",
			inputProjectID: "9",
			rawJSON:        `{"id": "PVT_xyz789", "title": "Test Project"}`,
			wantProjectID:  "9",
			wantNodeID:     "PVT_xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw map[string]any
			if err := json.Unmarshal([]byte(tt.rawJSON), &raw); err != nil {
				t.Fatalf("Failed to unmarshal test JSON: %v", err)
			}

			proj := state.Project{ID: tt.inputProjectID}
			if id, ok := raw["id"].(string); ok && id != "" {
				proj.NodeID = id
			}

			if proj.ID != tt.wantProjectID {
				t.Errorf("project ID = %q, want %q", proj.ID, tt.wantProjectID)
			}

			if proj.NodeID != tt.wantNodeID {
				t.Errorf("project node ID = %q, want %q", proj.NodeID, tt.wantNodeID)
			}
		})
	}
}

func TestUpdateStatus_ValidationIntegration(t *testing.T) {
	client := NewCLIClient("gh")
	ctx := context.Background()

	tests := []struct {
		name      string
		projectID string
		itemID    string
		fieldID   string
		optionID  string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "invalid numeric project ID fails fast",
			projectID: "9",
			itemID:    "PVTI_test123",
			fieldID:   "PVTF_field456",
			optionID:  "PVTSSO_opt789",
			wantErr:   true,
			errMsg:    "numeric only",
		},
		{
			name:      "invalid numeric item ID fails fast",
			projectID: "PVT_proj123",
			itemID:    "42",
			fieldID:   "PVTF_field456",
			optionID:  "PVTSSO_opt789",
			wantErr:   true,
			errMsg:    "numeric only",
		},
		{
			name:      "empty field ID fails fast",
			projectID: "PVT_proj123",
			itemID:    "PVTI_test123",
			fieldID:   "",
			optionID:  "PVTSSO_opt789",
			wantErr:   true,
			errMsg:    "empty",
		},
		{
			name:      "whitespace option ID fails fast",
			projectID: "PVT_proj123",
			itemID:    "PVTI_test123",
			fieldID:   "PVTF_field456",
			optionID:  "   ",
			wantErr:   true,
			errMsg:    "whitespace",
		},
		{
			name:      "project ID missing PVT_ prefix fails",
			projectID: "ABC123",
			itemID:    "PVTI_test123",
			fieldID:   "PVTF_field456",
			optionID:  "PVTSSO_opt789",
			wantErr:   true,
			errMsg:    "does not start with 'PVT_'",
		},
		{
			name:      "item ID missing PVTI_ prefix fails",
			projectID: "PVT_proj123",
			itemID:    "PVT_item456",
			fieldID:   "PVTF_field456",
			optionID:  "PVTSSO_opt789",
			wantErr:   true,
			errMsg:    "does not start with 'PVTI_'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.UpdateStatus(ctx, tt.projectID, "", tt.itemID, tt.fieldID, tt.optionID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateStatus() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("UpdateStatus() error = %v, want error containing %q", err, tt.errMsg)
				}
				if !strings.Contains(err.Error(), "validation failed") {
					t.Errorf("UpdateStatus() error should indicate validation failure, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("UpdateStatus() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestBuildIterationQuery(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output string
	}{
		{
			name:   "empty",
			input:  nil,
			output: "",
		},
		{
			name:   "single token",
			input:  []string{"@current"},
			output: "iteration:@current",
		},
		{
			name:   "trim and normalize",
			input:  []string{" iteration:@next ", "", "Iteration:  Sprint 1"},
			output: "iteration:@next iteration:Sprint 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildIterationQuery(tt.input)
			if got != tt.output {
				t.Fatalf("BuildIterationQuery() = %q, want %q", got, tt.output)
			}
		})
	}
}

func TestCreateIssueRejectsNonNumericProjectID(t *testing.T) {
	client := NewCLIClient("gh")
	_, err := client.CreateIssue(context.Background(), "PVT_kw123", "owner", "owner/repo", "Test issue", "Body")
	if err == nil {
		t.Fatalf("expected error for non-numeric project id")
	}
	if !strings.Contains(err.Error(), "project number required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateIssueRejectsEmptyTitle(t *testing.T) {
	client := NewCLIClient("gh")
	_, err := client.CreateIssue(context.Background(), "1", "owner", "owner/repo", "   ", "Body")
	if err == nil {
		t.Fatalf("expected error for empty title")
	}
	if !strings.Contains(err.Error(), "issue title is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateIssueRejectsEmptyBody(t *testing.T) {
	client := NewCLIClient("gh")
	_, err := client.CreateIssue(context.Background(), "1", "owner", "owner/repo", "Test issue", "   ")
	if err == nil {
		t.Fatalf("expected error for empty body")
	}
	if !strings.Contains(err.Error(), "issue body is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateIssueBodyRejectsMissingRepository(t *testing.T) {
	client := NewCLIClient("gh")
	err := client.UpdateIssueBody(context.Background(), "", 1, "Body")
	if err == nil {
		t.Fatalf("expected error for missing repository")
	}
	if !strings.Contains(err.Error(), "missing repository or issue number") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddIssueCommentRejectsEmptyBody(t *testing.T) {
	client := NewCLIClient("gh")
	err := client.AddIssueComment(context.Background(), "owner/repo", 1, "   ")
	if err == nil {
		t.Fatalf("expected error for empty comment body")
	}
	if !strings.Contains(err.Error(), "comment body is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseProjectItemAddOutput(t *testing.T) {
	rawJSON := `{
		"id": "PVTI_created123",
		"content": {
			"id": "I_created123",
			"type": "Issue",
			"title": "Created issue",
			"number": 42,
			"url": "https://github.com/owner/repo/issues/42",
			"repository": "owner/repo"
		},
		"fieldValues": []
	}`

	var raw map[string]any
	if err := json.Unmarshal([]byte(rawJSON), &raw); err != nil {
		t.Fatalf("Failed to unmarshal test JSON: %v", err)
	}

	item, ok := parse.ParseItemMap(raw)
	if !ok {
		t.Fatalf("expected parse to succeed")
	}
	if item.ID != "PVTI_created123" {
		t.Fatalf("expected project item id, got %q", item.ID)
	}
	if item.Type != "Issue" {
		t.Fatalf("expected issue type, got %q", item.Type)
	}
	if item.Title != "Created issue" {
		t.Fatalf("expected title, got %q", item.Title)
	}
	if item.Number != 42 {
		t.Fatalf("expected number 42, got %d", item.Number)
	}
	if item.Repository != "owner/repo" {
		t.Fatalf("expected repository owner/repo, got %q", item.Repository)
	}
	if item.URL != "https://github.com/owner/repo/issues/42" {
		t.Fatalf("expected URL to be populated, got %q", item.URL)
	}
}

func TestIsUnknownFlagError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		flag   string
		isTrue bool
	}{
		{
			name:   "nil error",
			err:    nil,
			flag:   "--query",
			isTrue: false,
		},
		{
			name:   "non-flag error",
			err:    fmt.Errorf("some other error"),
			flag:   "--query",
			isTrue: false,
		},
		{
			name:   "unknown flag match",
			err:    fmt.Errorf("exit status 1: unknown flag: --query"),
			flag:   "--query",
			isTrue: true,
		},
		{
			name:   "unknown flag different",
			err:    fmt.Errorf("exit status 1: unknown flag: --foo"),
			flag:   "--query",
			isTrue: false,
		},
		{
			name:   "missing flag",
			err:    fmt.Errorf("exit status 1: unknown flag: --query"),
			flag:   "",
			isTrue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUnknownFlagError(tt.err, tt.flag); got != tt.isTrue {
				t.Fatalf("isUnknownFlagError() = %v, want %v", got, tt.isTrue)
			}
		})
	}
}

func TestParseIssueComments(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantLen int
		want    string
	}{
		{
			name:    "flat array shape",
			raw:     `[{"author":{"login":"alice"},"body":"hello","createdAt":"2026-03-18T00:00:00Z"}]`,
			wantLen: 1,
			want:    "alice",
		},
		{
			name:    "nested nodes shape",
			raw:     `{"nodes":[{"author":{"login":"bob"},"body":"hi","createdAt":"2026-03-18T00:00:00Z"}]}`,
			wantLen: 1,
			want:    "bob",
		},
		{
			name:    "null comments",
			raw:     `null`,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments, err := parseIssueComments([]byte(tt.raw))
			if err != nil {
				t.Fatalf("parseIssueComments() error = %v", err)
			}
			if len(comments) != tt.wantLen {
				t.Fatalf("parseIssueComments() len = %d, want %d", len(comments), tt.wantLen)
			}
			if tt.wantLen > 0 && comments[0].Author != tt.want {
				t.Fatalf("parseIssueComments() author = %q, want %q", comments[0].Author, tt.want)
			}
		})
	}
}
