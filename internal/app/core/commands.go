package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"project-hub/internal/github"
)

func FetchProjectCmd(client github.Client, projectID, owner string, itemLimit int, iterationFilters []string) tea.Cmd {
	return func() tea.Msg {
		proj, items, err := client.FetchProject(context.Background(), projectID, owner, github.BuildIterationQuery(iterationFilters), itemLimit)
		if err != nil {
			return NewErrMsg(err)
		}
		return FetchProjectMsg{Project: proj, Items: items}
	}
}

func DismissNotificationCmd(id int, duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return DismissNotificationMsg{ID: id}
	})
}

// Make exec.Command and exec.LookPath replaceable for unit tests.
var execCommand = exec.Command
var lookPath = exec.LookPath

// ErrNoTool is returned when a required external tool is not available.
var ErrNoTool = errors.New("required tool not found")

// OpenBrowserCmd opens the provided URL in the user's default browser.
// If no browser opener is available, it returns an ActionResultMsg containing
// the URL so the UI can show it to the user for manual copy.
func OpenBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			if _, err := lookPath("open"); err != nil {
				return ActionResultMsg{Message: fmt.Sprintf("Cannot open browser: 'open' not available. URL: %s", url)}
			}
			cmd = execCommand("open", url)
		case "windows":
			// assume rundll32 exists on Windows; attempt to run it
			cmd = execCommand("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			if _, err := lookPath("xdg-open"); err != nil {
				return ActionResultMsg{Message: fmt.Sprintf("Cannot open browser: 'xdg-open' not available. URL: %s", url)}
			}
			cmd = execCommand("xdg-open", url)
		}
		if err := cmd.Start(); err != nil {
			return NewErrMsg(err)
		}
		return ActionResultMsg{Message: "Opened in browser"}
	}
}

// CopyToClipboardCmd copies the given text to the system clipboard.
// If no clipboard tool is available, it returns an ActionResultMsg that
// contains the URL so the UI can display it for manual copy by the user.
func CopyToClipboardCmd(text string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			if _, err := lookPath("pbcopy"); err != nil {
				return ActionResultMsg{Message: fmt.Sprintf("No clipboard utility found; URL: %s", text)}
			}
			cmd = execCommand("pbcopy")
		case "windows":
			if _, err := lookPath("clip"); err != nil {
				return ActionResultMsg{Message: fmt.Sprintf("No clipboard utility found; URL: %s", text)}
			}
			cmd = execCommand("cmd", "/c", "clip")
		default:
			// prefer wl-copy, then xclip
			if _, err := lookPath("wl-copy"); err == nil {
				cmd = execCommand("wl-copy")
			} else if _, err := lookPath("xclip"); err == nil {
				cmd = execCommand("xclip", "-selection", "clipboard")
			} else {
				return ActionResultMsg{Message: fmt.Sprintf("No clipboard utility found; URL: %s", text)}
			}
		}
		in := &bytes.Buffer{}
		in.WriteString(text)
		cmd.Stdin = in
		if err := cmd.Run(); err != nil {
			return NewErrMsg(err)
		}
		return ActionResultMsg{Message: "Copied URL to clipboard"}
	}
}
