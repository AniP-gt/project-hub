package components

import (
	"strings"
	"testing"

	"project-hub/internal/state"
)

// TestRenderViewTabsIncludesSettings verifies that ViewTabs output includes "[4:Settings]"
func TestRenderViewTabsIncludesSettings(t *testing.T) {
	tests := []struct {
		name        string
		currentView state.ViewType
		want        string
	}{
		{
			name:        "Board view shows Settings tab",
			currentView: state.ViewBoard,
			want:        "[4:Settings]",
		},
		{
			name:        "Table view shows Settings tab",
			currentView: state.ViewTable,
			want:        "[4:Settings]",
		},
		{
			name:        "Roadmap view shows Settings tab",
			currentView: state.ViewRoadmap,
			want:        "[4:Settings]",
		},
		{
			name:        "Settings view shows Settings tab",
			currentView: state.ViewSettings,
			want:        "[4:Settings]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := renderViewTabs(tt.currentView)
			if !strings.Contains(output, tt.want) {
				t.Errorf("renderViewTabs(%v) does not contain %q\nGot: %s", tt.currentView, tt.want, output)
			}
		})
	}
}

// TestRenderFooterKeyHints verifies that footer keybinds include view switch hint
func TestRenderFooterKeyHints(t *testing.T) {
	output := RenderFooter("normal", "board", 80, "")

	// Check for view switch hint (either 1-4 or individual hints)
	hasViewSwitch := strings.Contains(output, "1-4") ||
		strings.Contains(output, "[1:") ||
		strings.Contains(output, "[4:")

	if !hasViewSwitch {
		t.Errorf("RenderFooter does not contain view switch hint\nGot: %s", output)
	}
}
