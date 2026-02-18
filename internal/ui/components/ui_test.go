package components

import (
	"strings"
	"testing"

	"project-hub/internal/state"
)

// TestRenderViewTabsIncludesSettings verifies that ViewTabs output includes "[3:Settings]"
func TestRenderViewTabsIncludesSettings(t *testing.T) {
	tests := []struct {
		name        string
		currentView state.ViewType
		want        string
	}{
		{
			name:        "Board view shows Settings tab",
			currentView: state.ViewBoard,
			want:        "[3:Settings]",
		},
		{
			name:        "Table view shows Settings tab",
			currentView: state.ViewTable,
			want:        "[3:Settings]",
		},
		{
			name:        "Settings view shows Settings tab",
			currentView: state.ViewSettings,
			want:        "[3:Settings]",
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

	// Check for view switch hint (1-3 for 3 views)
	if !strings.Contains(output, "1-3") {
		t.Errorf("RenderFooter does not contain view switch hint\nGot: %s", output)
	}
}
