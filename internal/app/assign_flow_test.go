package app

import (
	"testing"

	"project-hub/internal/app/update"
	"project-hub/internal/state"
)

// Test the full assign flow: EnterAssignMode -> SaveAssign triggers UpdateAssignees -> ItemUpdatedMsg applied
func TestFullAssignFlow(t *testing.T) {
	items := []state.Item{
		{ID: "1", Title: "A", Status: "Todo", Type: "Issue"},
		{ID: "2", Title: "B", Status: "In Progress", Type: "Issue"},
		{ID: "3", Title: "C", Status: "Done", Type: "Issue"},
	}
	initial := state.Model{
		Project: state.Project{ID: "proj"},
		Items:   items,
		View:    state.ViewContext{CurrentView: state.ViewBoard, Mode: state.ModeNormal, FocusedIndex: 1, FocusedItemID: "2"},
	}

	app := New(initial, &noopClient{}, 100)

	// Record initial mapping of item "2"
	var origColIdx, origCardIdx int
	for ci, col := range app.boardModel.Columns {
		for cj, card := range col.Cards {
			if card.ID == "2" {
				origColIdx = ci
				origCardIdx = cj
			}
		}
	}

	// Enter assign mode
	model, cmd := app.Update(update.EnterAssignModeMsg{})
	app = model.(App)
	if cmd != nil {
		// run cmd to completion (should be a no-op here)
		_ = cmd()
	}

	// Save assignee 'alice'
	model, cmd = app.Update(update.SaveAssignMsg{Assignee: "alice"})
	app = model.(App)

	// Execute the returned cmd which performs UpdateAssignees (noopClient returns ItemUpdatedMsg)
	if cmd != nil {
		msg := cmd()
		// Feed the resulting message back into Update to apply ItemUpdatedMsg
		model, _ = app.Update(msg)
		app = model.(App)
	}

	// Verify in-memory item updated
	if app.state.Items[1].Assignees == nil || len(app.state.Items[1].Assignees) == 0 || app.state.Items[1].Assignees[0] != "alice" {
		t.Fatalf("assignee not updated on item: %+v", app.state.Items[1])
	}

	// Verify board model still maps card '2' to same column and index
	var newColIdx, newCardIdx int = -1, -1
	for ci, col := range app.boardModel.Columns {
		for cj, card := range col.Cards {
			if card.ID == "2" {
				newColIdx = ci
				newCardIdx = cj
			}
		}
	}
	if newColIdx != origColIdx {
		t.Fatalf("card remapped to different column: orig=%d new=%d", origColIdx, newColIdx)
	}
	if newCardIdx != origCardIdx {
		t.Fatalf("card index changed: orig=%d new=%d", origCardIdx, newCardIdx)
	}
}

func TestFullCreateIssueFlow(t *testing.T) {
	items := []state.Item{
		{ID: "1", Title: "A", Status: "Todo", Type: "Issue", Repository: "owner/repo"},
	}
	initial := state.Model{
		Project: state.Project{ID: "1", Owner: "owner"},
		Items:   items,
		View:    state.ViewContext{CurrentView: state.ViewBoard, Mode: state.ModeNormal, FocusedIndex: 0, FocusedItemID: "1"},
	}

	app := New(initial, &noopClient{}, 100)

	model, cmd := app.Update(update.EnterCreateIssueModeMsg{})
	app = model.(App)
	if cmd != nil {
		_ = cmd()
	}
	if app.state.View.Mode != state.ModeCreateIssueTitle {
		t.Fatalf("expected create issue title mode, got %q", app.state.View.Mode)
	}

	model, cmd = app.Update(update.SaveCreateIssueMsg{Value: "Created from test"})
	app = model.(App)
	if app.state.View.Mode != state.ModeCreateIssueBody {
		t.Fatalf("expected create issue body mode, got %q", app.state.View.Mode)
	}

	model, cmd = app.Update(update.SaveCreateIssueMsg{Value: "Created from test body"})
	app = model.(App)
	if cmd == nil {
		t.Fatalf("expected create issue command")
	}
	msg := cmd()
	model, _ = app.Update(msg)
	app = model.(App)

	if len(app.state.Notifications) == 0 {
		t.Fatalf("expected success notification after issue creation")
	}
	if app.state.View.Mode != state.ModeNormal {
		t.Fatalf("expected mode to return to normal, got %q", app.state.View.Mode)
	}
}
