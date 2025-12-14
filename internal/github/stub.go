package github

import (
	"context"
	"time"

	"project-hub/internal/state"
)

// StubClient is a simple in-memory Client used for local UI development.
type StubClient struct{}

func NewStubClient() *StubClient { return &StubClient{} }

func (s *StubClient) FetchProject(ctx context.Context, projectID string, owner string, limit int) (state.Project, []state.Item, error) {
	proj := state.Project{
		ID:         "proj-1",
		Name:       "Web App v2.0",
		Owner:      owner,
		Views:      []state.ViewType{state.ViewBoard, state.ViewTable, state.ViewRoadmap},
		Iterations: []state.Timeline{{ID: "s1", Name: "Sprint 1", Progress: "60%"}, {ID: "s2", Name: "Sprint 2", Progress: "20%"}},
	}
	now := time.Now()
	items := []state.Item{
		{ID: "#123", Title: "ユーザー認証機能", Status: "Backlog", Repository: "repo1", Assignees: []string{"tanaka"}, Labels: []string{"high"}, UpdatedAt: &now, Position: 1},
		{ID: "#124", Title: "API統合", Status: "Backlog", Repository: "repo1", Assignees: []string{"sato"}, Labels: []string{"medium"}, UpdatedAt: &now, Position: 2},
		{ID: "#126", Title: "TUIデザイン実装", Status: "In Progress", Repository: "repo2", Assignees: []string{"tanaka"}, Labels: []string{"high"}, UpdatedAt: &now, Position: 3},
		{ID: "#127", Title: "キーバインド設定", Status: "In Progress", Repository: "repo2", Assignees: []string{"yamada"}, Labels: []string{"low"}, UpdatedAt: &now, Position: 4},
		{ID: "#128", Title: "テストコード追加", Status: "Review", Repository: "repo1", Assignees: []string{"sato"}, Labels: []string{"medium"}, UpdatedAt: &now, Position: 5},
	}
	return proj, items, nil
}

func (s *StubClient) FetchItems(ctx context.Context, projectID string, owner string, filter string, limit int) ([]state.Item, error) {
	_, items, _ := s.FetchProject(ctx, projectID, owner, limit)
	return items, nil
}

func (s *StubClient) UpdateStatus(ctx context.Context, projectID string, owner string, itemID string, fieldID string, optionID string) (state.Item, error) {
	// no-op: return an item with updated status
	return state.Item{ID: itemID, Status: optionID}, nil
}

func (s *StubClient) UpdateAssignees(ctx context.Context, projectID string, owner string, itemID string, assigneeFieldID string, userLogins []string) (state.Item, error) {
	ass := ""
	if len(userLogins) > 0 {
		ass = userLogins[0]
	}
	return state.Item{ID: itemID, Assignees: []string{ass}}, nil
}

func (s *StubClient) UpdateItem(ctx context.Context, projectID string, owner string, item state.Item, title string, description string) (state.Item, error) {
	if title != "" {
		item.Title = title
	}
	if description != "" {
		item.Description = description
	}
	return item, nil
}

func (s *StubClient) FetchRoadmap(ctx context.Context, projectID string, owner string) ([]state.Timeline, []state.Item, error) {
	proj, items, _ := s.FetchProject(ctx, projectID, owner, 100)
	return proj.Iterations, items, nil
}
