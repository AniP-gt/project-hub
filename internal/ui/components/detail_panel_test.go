package components

import (
	"strings"
	"testing"

	"project-hub/internal/state"
)

func TestDetailPanelViewShowsIssueRelationships(t *testing.T) {
	model := NewDetailPanelModel(state.Item{
		Title:            "Parent issue",
		ParentIssue:      "Epic issue",
		SubIssueProgress: "2/3",
		SubIssueTitles:   []string{"Sub issue A", "Sub issue B"},
		Description:      "Body",
	}, 120, 40)

	view := model.View()

	for _, expected := range []string{"Parent: ", "Epic issue", "Sub-issue progress: ", "2/3", "Sub-issues:", "Sub issue A", "Sub issue B"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("expected detail view to contain %q, got: %q", expected, view)
		}
	}
}

func TestDetailPanelViewShowsBoxedComments(t *testing.T) {
	model := NewDetailPanelModel(state.Item{
		Title:       "Issue with comments",
		Description: "Body",
		Comments: []state.Comment{
			{Author: "alice", Body: "First comment"},
			{Author: "bob", Body: "Second comment"},
		},
	}, 120, 40)

	view := model.View()

	for _, expected := range []string{"Comments (2)", "@alice", "First comment", "@bob", "Second comment", "╭", "╰"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("expected detail view to contain %q, got: %q", expected, view)
		}
	}
}
