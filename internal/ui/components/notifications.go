package components

import (
	"strings"
	"time"

	"project-hub/internal/state"
)

// RenderNotifications renders non-blocking notifications inline.
func RenderNotifications(notifs []state.Notification) string {
	if len(notifs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, n := range notifs {
		if n.Dismissed {
			continue
		}
		b.WriteString(formatNotification(n))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func formatNotification(n state.Notification) string {
	age := time.Since(n.At).Round(time.Second)
	if age < 0 {
		age = 0
	}
	style := NotifInfo
	if n.Level == "error" {
		style = NotifWarn
	}
	return style.Render(n.Message + " (" + age.String() + " ago)")
}
