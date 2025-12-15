package github

import (
	"encoding/json"
	"testing"
	"time"

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw map[string]any
			if err := json.Unmarshal([]byte(tt.inputJSON), &raw); err != nil {
				t.Fatalf("Failed to unmarshal input JSON: %v", err)
			}
			got, ok := parseItemMap(raw)
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
			if got.UpdatedAt != nil && tt.wantItem.UpdatedAt != nil {
				if !got.UpdatedAt.Equal(*tt.wantItem.UpdatedAt) {
					t.Errorf("parseItemMap() UpdatedAt = %v, want %v", *got.UpdatedAt, *tt.wantItem.UpdatedAt)
				}
			} else if (got.UpdatedAt == nil && tt.wantItem.UpdatedAt != nil) || (got.UpdatedAt != nil && tt.wantItem.UpdatedAt == nil) {
				t.Errorf("parseItemMap() UpdatedAt mismatch: got %v, want %v", got.UpdatedAt, tt.wantItem.UpdatedAt)
			}
			// Add more assertions for other fields as needed
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
