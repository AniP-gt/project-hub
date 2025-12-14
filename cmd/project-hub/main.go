package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"project-hub/internal/app"
	"project-hub/internal/github"
	"project-hub/internal/state"
)

func main() {
	// Minimal bootstrap: empty project and a stub github client
	st := state.Model{}
	client := github.NewStubClient()
	a := app.New(st, client, 100)
	p := tea.NewProgram(a)
	if err := p.Start(); err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
