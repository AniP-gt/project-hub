package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"project-hub/internal/github/parse"
	"project-hub/internal/state"
)

func (c *CLIClient) FetchProject(ctx context.Context, projectID string, owner string, limit int) (state.Project, []state.Item, error) {
	viewArgs := []string{"project", "view", projectID, "--format", "json"}
	if owner != "" {
		viewArgs = append(viewArgs, "--owner", owner)
	}
	out, err := c.runGh(ctx, viewArgs...)
	if err != nil {
		return state.Project{}, nil, fmt.Errorf("gh project view failed: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return state.Project{}, nil, fmt.Errorf("parse gh project view json: %w", err)
	}
	proj := state.Project{ID: projectID}
	if id, ok := raw["id"].(string); ok && id != "" {
		proj.NodeID = id
	}
	if own, ok := raw["owner"].(map[string]any); ok {
		if login, ok := own["login"].(string); ok && login != "" {
			proj.Owner = login
		}
	}
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

	fieldArgs := []string{"project", "field-list", projectID, "--format", "json"}
	if owner != "" {
		fieldArgs = append(fieldArgs, "--owner", owner)
	}
	fieldsOut, err := c.runGh(ctx, fieldArgs...)
	if err != nil {
	} else {
		var rawFields struct {
			Fields []struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Options []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"options"`
			} `json:"fields"`
		}
		if err := json.Unmarshal(fieldsOut, &rawFields); err == nil {
			for _, rf := range rawFields.Fields {
				field := state.Field{
					ID:   rf.ID,
					Name: rf.Name,
				}
				for _, ro := range rf.Options {
					field.Options = append(field.Options, state.Option{
						ID:   ro.ID,
						Name: ro.Name,
					})
				}
				proj.Fields = append(proj.Fields, field)
			}
		}
	}

	items, err := c.FetchItems(ctx, projectID, owner, "", limit)
	if err != nil {
		return proj, nil, err
	}
	return proj, items, nil
}

func (c *CLIClient) FetchItems(ctx context.Context, projectID string, owner string, filter string, limit int) ([]state.Item, error) {
	limitStr := strconv.Itoa(limit)
	args := []string{"project", "item-list", projectID, "--format", "json", "--limit", limitStr}
	if owner != "" {
		args = append(args, "--owner", owner)
	}
	if filter != "" {
		args = append(args, "--search", filter)
	}
	out, err := c.runGh(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("gh project item-list failed: %w", err)
	}
	items, parseErr := parse.ParseItemList(out)
	if parseErr != nil {
		return nil, parseErr
	}

	if hierarchy, err := c.fetchProjectHierarchy(ctx, owner, projectID); err == nil {
		for i := range items {
			key := issueKey(items[i].Repository, items[i].Number)
			if key == "" {
				continue
			}
			info, ok := hierarchy[key]
			if !ok {
				continue
			}
			if items[i].SubIssueProgress == "" && info.SubIssueTotal > 0 {
				items[i].SubIssueProgress = fmt.Sprintf("%d", info.SubIssueTotal)
			}
			if items[i].ParentIssue == "" {
				if info.ParentTitle != "" {
					items[i].ParentIssue = info.ParentTitle
				} else if info.ParentNumber > 0 {
					items[i].ParentIssue = fmt.Sprintf("#%d", info.ParentNumber)
				}
			}
		}
	}
	return items, nil
}

type projectHierarchy struct {
	SubIssueTotal int
	ParentTitle   string
	ParentNumber  int
}

func issueKey(repo string, number int) string {
	if number <= 0 {
		return ""
	}
	cleanRepo := normalizeRepo(repo)
	if cleanRepo == "" {
		return ""
	}
	return fmt.Sprintf("%s#%d", strings.ToLower(cleanRepo), number)
}

func normalizeRepo(repo string) string {
	trimmed := strings.TrimSpace(repo)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.TrimPrefix(trimmed, "https://github.com/")
	trimmed = strings.TrimPrefix(trimmed, "http://github.com/")
	trimmed = strings.TrimPrefix(trimmed, "github.com/")
	return strings.Trim(trimmed, "/")
}

type projectHierarchyResponse struct {
	Data struct {
		User         *projectHierarchyNode `json:"user"`
		Organization *projectHierarchyNode `json:"organization"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"errors"`
}

type projectHierarchyNode struct {
	ProjectV2 *struct {
		Items struct {
			Nodes []struct {
				Content *struct {
					Number     int `json:"number"`
					Repository struct {
						NameWithOwner string `json:"nameWithOwner"`
					} `json:"repository"`
					SubIssues struct {
						TotalCount int `json:"totalCount"`
					} `json:"subIssues"`
					Parent *struct {
						Title  string `json:"title"`
						Number int    `json:"number"`
					} `json:"parent"`
				} `json:"content"`
			} `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"items"`
	} `json:"projectV2"`
}

func (c *CLIClient) fetchProjectHierarchy(ctx context.Context, owner, projectID string) (map[string]projectHierarchy, error) {
	projectNumber, err := strconv.Atoi(strings.TrimSpace(projectID))
	if err != nil || projectNumber <= 0 {
		return nil, fmt.Errorf("project number required for hierarchy fetch")
	}

	query := `query($owner:String!,$number:Int!,$after:String){user(login:$owner){projectV2(number:$number){items(first:100, after:$after){nodes{content{... on Issue{number repository{nameWithOwner} subIssues{totalCount} parent{title number}}}} pageInfo{hasNextPage endCursor}}}}}`
	res, err := c.fetchProjectHierarchyForOwner(ctx, "user", owner, projectNumber, query)
	if err == nil {
		return res, nil
	}

	query = `query($owner:String!,$number:Int!,$after:String){organization(login:$owner){projectV2(number:$number){items(first:100, after:$after){nodes{content{... on Issue{number repository{nameWithOwner} subIssues{totalCount} parent{title number}}}} pageInfo{hasNextPage endCursor}}}}}`
	return c.fetchProjectHierarchyForOwner(ctx, "organization", owner, projectNumber, query)
}

func (c *CLIClient) fetchProjectHierarchyForOwner(ctx context.Context, ownerType, owner string, projectNumber int, query string) (map[string]projectHierarchy, error) {
	var out map[string]projectHierarchy
	var after string
	for {
		args := []string{"api", "graphql", "--field", fmt.Sprintf("query=%s", query), "-F", fmt.Sprintf("owner=%s", owner), "-F", fmt.Sprintf("number=%d", projectNumber)}
		if after != "" {
			args = append(args, "-F", fmt.Sprintf("after=%s", after))
		}
		respBytes, err := c.runGh(ctx, args...)
		if err != nil {
			return nil, err
		}
		var resp projectHierarchyResponse
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			return nil, err
		}
		if len(resp.Errors) > 0 {
			return nil, fmt.Errorf("hierarchy query error: %s", resp.Errors[0].Message)
		}

		var node *projectHierarchyNode
		switch ownerType {
		case "user":
			node = resp.Data.User
		case "organization":
			node = resp.Data.Organization
		}
		if node == nil || node.ProjectV2 == nil {
			return nil, fmt.Errorf("hierarchy data not found for %s", ownerType)
		}
		if out == nil {
			out = make(map[string]projectHierarchy)
		}
		for _, item := range node.ProjectV2.Items.Nodes {
			if item.Content == nil {
				continue
			}
			key := issueKey(item.Content.Repository.NameWithOwner, item.Content.Number)
			if key == "" {
				continue
			}
			parentTitle := ""
			parentNumber := 0
			if item.Content.Parent != nil {
				parentTitle = item.Content.Parent.Title
				parentNumber = item.Content.Parent.Number
			}
			out[key] = projectHierarchy{
				SubIssueTotal: item.Content.SubIssues.TotalCount,
				ParentTitle:   parentTitle,
				ParentNumber:  parentNumber,
			}
		}
		if !node.ProjectV2.Items.PageInfo.HasNextPage {
			break
		}
		after = node.ProjectV2.Items.PageInfo.EndCursor
	}
	return out, nil
}
