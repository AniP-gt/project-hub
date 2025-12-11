package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app"
	"project-hub/internal/github"
	"project-hub/internal/state"
)

func main() {
	projectArg := flag.String("project", "", "GitHub Project ID or URL")
	ownerFlag := flag.String("owner", "", "Owner (org/user) for the project")
	flag.Parse()

	if *projectArg == "" {
		fmt.Fprintln(os.Stderr, "--project is required")
		flag.Usage()
		os.Exit(1)
	}

	projID, urlOwner := parseProjectArg(*projectArg)
	owner := *ownerFlag
	if owner == "" {
		owner = urlOwner
	}
	if projID == "" {
		projID = *projectArg
	}

	client := github.NewCLIClient()
	initial := state.Model{
		Project: state.Project{ID: projID, Name: "GitHub Projects TUI"},
		Items: []state.Item{
			{ID: "1", Title: "Design board", Status: "Backlog", Labels: []string{"design"}},
			{ID: "2", Title: "Wire table view", Status: "InProgress", Labels: []string{"ui"}},
			{ID: "3", Title: "Roadmap draft", Status: "Review", Labels: []string{"roadmap"}},
		},
		View: state.ViewContext{CurrentView: state.ViewBoard, Mode: state.ModeNormal, FocusedIndex: 0, FocusedItemID: "1"},
	}

	// Try to load real project data via gh; fallback to sample on error.
	if proj, items, err := client.FetchProject(context.Background(), projID, owner); err == nil {
		if proj.Name != "" {
			initial.Project = proj
		}
		if len(items) > 0 {
			initial.Items = items
			initial.View.FocusedIndex = 0
			initial.View.FocusedItemID = items[0].ID
			initial.Notifications = append(initial.Notifications, state.Notification{Message: fmt.Sprintf("Loaded %d items from project", len(items)), Level: "info", At: time.Now()})
		} else {
			initial.Notifications = append(initial.Notifications, state.Notification{Message: "No items fetched (gh returned 0)", Level: "warn", At: time.Now()})
		}
	} else {
		fmt.Fprintln(os.Stderr, "warning: gh fetch failed, using sample data:", err)
	}

	p := tea.NewProgram(app.New(initial, client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start program:", err)
		os.Exit(1)
	}
}

func parseProjectArg(arg string) (projectID string, owner string) {
	u, err := url.Parse(arg)
	if err == nil && u.Scheme != "" {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		for i := 0; i < len(parts); i++ {
			if parts[i] == "projects" && i > 0 && i+1 < len(parts) {
				return parts[i+1], parts[i-1]
			}
		}
	}
	return arg, ""
}
