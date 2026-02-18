package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app"
	"project-hub/internal/config"
	"project-hub/internal/github"
	"project-hub/internal/state"
)

// resolveStartupOptions merges CLI arguments and config defaults with proper precedence.
// CLI non-empty values always win over config defaults.
// Precedence: CLI project/owner > config project/owner.
func resolveStartupOptions(cliProject, cliOwner string, cfg config.Config) (project, owner string) {
	// CLI project takes precedence; fall back to config
	if cliProject != "" {
		project = cliProject
	} else {
		project = cfg.DefaultProjectID
	}

	// CLI owner takes precedence; fall back to config
	if cliOwner != "" {
		owner = cliOwner
	} else {
		owner = cfg.DefaultOwner
	}

	return project, owner
}

func loadStartupConfig(errOut io.Writer) config.Config {
	cfg := config.Config{}
	configPath, err := config.ResolvePath()
	if err != nil {
		fmt.Fprintln(errOut, "warning: failed to resolve config path:", err)
		return cfg
	}
	loadedCfg, err := config.Load(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintln(errOut, "warning: failed to load config:", err)
		}
		return cfg
	}
	return loadedCfg
}

func main() {
	projectArg := flag.String("project", "", "GitHub Project ID or URL")
	ownerFlag := flag.String("owner", "", "Owner (org/user) for the project")
	ghPathFlag := flag.String("gh-path", "", "Path to the gh CLI executable (default: \"gh\")")
	itemLimitFlag := flag.Int("item-limit", 100, "Maximum number of items to fetch (default: 100)")
	disableNotificationsFlag := flag.Bool("disable-notifications", false, "Suppress info-level notifications in the UI")
	var iterationFlag multiValueFlag
	flag.Var(&iterationFlag, "iteration", "Iteration filters (repeat flag or pass values after it)")
	flag.Parse()

	iterationTokens := append([]string{}, iterationFlag...)
	if len(iterationTokens) > 0 {
		iterationTokens = append(iterationTokens, flag.Args()...)
	} else if len(flag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", flag.Args())
		os.Exit(1)
	}
	iterationFilters := normalizeIterationFilters(iterationTokens)

	// Load and use configuration
	cfg := loadStartupConfig(os.Stderr)

	// Resolve final project and owner values (CLI > config defaults)
	resolvedProject, resolvedOwner := resolveStartupOptions(*projectArg, *ownerFlag, cfg)

	// Resolve notification setting: CLI flag takes precedence over config
	disableNotifications := cfg.DisableNotifications
	if *disableNotificationsFlag {
		disableNotifications = true
	}

	// Check that project is now satisfied (either from CLI or config)
	if resolvedProject == "" {
		fmt.Fprintln(os.Stderr, "--project is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse the resolved project argument to extract ID and infer owner from URL if needed
	projID, urlOwner := parseProjectArg(resolvedProject)
	owner := resolvedOwner
	if owner == "" {
		owner = urlOwner
	}
	if projID == "" {
		projID = resolvedProject
	}

	client := github.NewCLIClient(*ghPathFlag)
	initial := state.Model{
		Project: state.Project{ID: projID, Owner: owner, Name: "GitHub Projects TUI"},
		Items: []state.Item{
			{ID: "1", Title: "Design board", Status: "Backlog", Labels: []string{"design"}},
			{ID: "2", Title: "Wire table view", Status: "InProgress", Labels: []string{"ui"}},
			{ID: "3", Title: "Roadmap draft", Status: "Review", Labels: []string{"roadmap"}},
		},
		View:                 state.ViewContext{CurrentView: state.ViewBoard, Mode: state.ModeNormal, FocusedIndex: 0, FocusedItemID: "1"},
		ItemLimit:            *itemLimitFlag,
		DisableNotifications: disableNotifications,
	}
	initial.View.Filter.Iterations = iterationFilters

	// Try to load real project data via gh; fallback to sample on error.
	if proj, items, err := client.FetchProject(context.Background(), projID, owner, *itemLimitFlag); err == nil {
		if proj.Name != "" {
			initial.Project = proj
		}
		if len(items) > 0 {
			initial.Items = items
			initial.View.FocusedIndex = 0
			initial.View.FocusedItemID = items[0].ID
			if !disableNotifications {
				initial.Notifications = append(initial.Notifications, state.Notification{Message: fmt.Sprintf("Loaded %d items from project", len(items)), Level: "info", At: time.Now()})
			}
		} else {
			initial.Notifications = append(initial.Notifications, state.Notification{Message: "No items fetched (gh returned 0)", Level: "warn", At: time.Now()})
		}
	} else {
		fmt.Fprintln(os.Stderr, "warning: gh fetch failed, using sample data:", err)
	}

	p := tea.NewProgram(app.New(initial, client, *itemLimitFlag), tea.WithAltScreen())
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

type multiValueFlag []string

func (m *multiValueFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiValueFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func normalizeIterationFilters(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var filters []string
	for _, raw := range values {
		val := strings.TrimSpace(raw)
		if val == "" {
			continue
		}
		lower := strings.ToLower(val)
		if strings.HasPrefix(lower, "iteration:") {
			val = strings.TrimSpace(val[len("iteration:"):])
			lower = strings.ToLower(val)
		}
		if val == "" {
			continue
		}
		relative := strings.TrimPrefix(lower, "@")
		switch relative {
		case "current", "next", "previous":
			rel := "@" + relative
			if _, ok := seen[rel]; ok {
				continue
			}
			seen[rel] = struct{}{}
			filters = append(filters, rel)
			continue
		}
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		filters = append(filters, val)
	}
	return filters
}
