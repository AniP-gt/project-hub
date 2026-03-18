package board

import (
	"project-hub/internal/config"
	"project-hub/internal/state"
)

func SaveCardFieldVisibility(vis state.CardFieldVisibility) error {
	configPath, err := config.ResolvePath()
	if err != nil {
		return err
	}
	existing, loadErr := config.Load(configPath)
	if loadErr != nil {
		return loadErr
	}
	cfg := config.Config{
		DefaultProjectID:        existing.DefaultProjectID,
		DefaultOwner:            existing.DefaultOwner,
		SuppressHints:           existing.SuppressHints,
		DefaultItemLimit:        existing.DefaultItemLimit,
		DefaultExcludeDone:      existing.DefaultExcludeDone,
		CreateIssueRepoMode:     existing.CreateIssueRepoMode,
		DefaultIterationFilters: existing.DefaultIterationFilters,
		CardFieldVisibility: config.CardFieldVisibility{
			ShowMilestone:        vis.ShowMilestone,
			ShowRepository:       vis.ShowRepository,
			ShowSubIssueProgress: vis.ShowSubIssueProgress,
			ShowParentIssue:      vis.ShowParentIssue,
			ShowLabels:           vis.ShowLabels,
		},
	}
	return config.Save(configPath, cfg)
}
