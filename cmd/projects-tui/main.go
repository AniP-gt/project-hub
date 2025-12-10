package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app"
	"project-hub/internal/github"
	"project-hub/internal/state"
)

func main() {
	projectID := flag.String("project", "", "GitHub Project ID")
	flag.Parse()

	if *projectID == "" {
		fmt.Fprintln(os.Stderr, "--project is required")
		flag.Usage()
		os.Exit(1)
	}

	client := github.NewCLIClient()
	initial := state.Model{
		Project: state.Project{ID: *projectID, Name: "GitHub Projects TUI"},
		Items: []state.Item{
			{ID: "1", Title: "Design board", Status: "Backlog", Labels: []string{"design"}},
			{ID: "2", Title: "Wire table view", Status: "InProgress", Labels: []string{"ui"}},
			{ID: "3", Title: "Roadmap draft", Status: "Review", Labels: []string{"roadmap"}},
		},
		View: state.ViewContext{CurrentView: state.ViewBoard, Mode: state.ModeNormal, FocusedIndex: 0, FocusedItemID: "1"},
	}

	// Try to load real project data via gh; fallback to sample on error.
	if proj, items, err := client.FetchProject(context.Background(), *projectID); err == nil {
		if proj.Name != "" {
			initial.Project = proj
		}
		if len(items) > 0 {
			initial.Items = items
			initial.View.FocusedIndex = 0
			initial.View.FocusedItemID = items[0].ID
		}
	} else {
		fmt.Fprintln(os.Stderr, "warning: gh fetch failed, using sample data:", err)
	}

	p := tea.NewProgram(app.New(initial, client))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start program:", err)
		os.Exit(1)
	}
}
