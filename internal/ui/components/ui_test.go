package components

import (
	"strings"
	"testing"
)

func TestStatusDotRendersDot(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{name: "Open status", status: "OPEN"},
		{name: "Closed status", status: "closed"},
		{name: "Unknown status", status: "Backlog"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := StatusDot(tt.status)
			if !strings.Contains(out, "●") {
				t.Fatalf("StatusDot(%q) should render dot, got: %q", tt.status, out)
			}
		})
	}
}
