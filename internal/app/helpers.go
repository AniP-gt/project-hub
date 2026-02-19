package app

import (
	"strings"

	"project-hub/internal/state"
)

func projectMutationID(project state.Project) string {
	if strings.TrimSpace(project.NodeID) != "" {
		return project.NodeID
	}
	return project.ID
}
