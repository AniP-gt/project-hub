package app

import (
	"context"
	"testing"

	"project-hub/internal/app/core"
	"project-hub/internal/state"
)

// mock client that implements github.Client but does nothing.
type noopClient struct{}

func (n *noopClient) FetchProject(ctx context.Context, projectID string, owner string, limit int) (state.Project, []state.Item, error) {
	return state.Project{}, nil, nil
}
func (n *noopClient) FetchItems(ctx context.Context, projectID string, owner string, filter string, limit int) ([]state.Item, error) {
	return nil, nil
}
func (n *noopClient) UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string) (state.Item, error) {
	return state.Item{}, nil
}
func (n *noopClient) UpdateField(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string, fieldName string) (state.Item, error) {
	return state.Item{}, nil
}
func (n *noopClient) UpdateLabels(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, labels []string) (state.Item, error) {
	return state.Item{}, nil
}
func (n *noopClient) UpdateMilestone(ctx context.Context, projectID string, owner string, itemID string, milestone string) (state.Item, error) {
	return state.Item{}, nil
}
func (n *noopClient) UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, itemType string, repo string, number int, userLogins []string) (state.Item, error) {
	// Return an item with the assignees populated, simulate partial update.
	return state.Item{ID: itemID, Assignees: userLogins}, nil
}
func (n *noopClient) UpdateItem(ctx context.Context, projectID string, owner string, item state.Item, title string, description string) (state.Item, error) {
	return item, nil
}
func (n *noopClient) FetchIssueDetail(ctx context.Context, repo string, number int) (string, error) {
	return "", nil
}

func TestAssignUpdateDoesNotRemapCards(t *testing.T) {
	// Create 3 items in different statuses
	items := []state.Item{
		{ID: "1", Title: "A", Status: "Todo"},
		{ID: "2", Title: "B", Status: "In Progress"},
		{ID: "3", Title: "C", Status: "Done"},
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

	// Simulate updating assignee for item ID "2"
	updated := state.Item{ID: "2", Assignees: []string{"alice"}, Status: "In Progress", Title: "B"}
	// Call Update by invoking App.Update on the model interface
	model, _ := app.Update(core.ItemUpdatedMsg{Index: 1, Item: updated})
	updatedApp := model.(App)

	// After update, ensure item at index 1 has assignee
	if updatedApp.state.Items[1].Assignees == nil || len(updatedApp.state.Items[1].Assignees) == 0 {
		t.Fatalf("expected assignees to be populated on item index 1")
	}

	// Ensure the board model still contains the card with ID "2" in same column
	var newColIdx, newCardIdx int = -1, -1
	for ci, col := range updatedApp.boardModel.Columns {
		for cj, card := range col.Cards {
			if card.ID == "2" {
				newColIdx = ci
				newCardIdx = cj
			}
		}
	}
	if newColIdx != origColIdx {
		t.Fatalf("card was remapped to different column: orig=%d new=%d", origColIdx, newColIdx)
	}
	if newCardIdx != origCardIdx {
		t.Fatalf("card index changed: orig=%d new=%d", origCardIdx, newCardIdx)
	}
}
