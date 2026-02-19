package core

import (
	"strings"

	"project-hub/internal/state"
)

func ProjectMutationID(project state.Project) string {
	if strings.TrimSpace(project.NodeID) != "" {
		return project.NodeID
	}
	return project.ID
}
