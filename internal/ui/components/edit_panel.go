package components

// RenderEditPanel shows contextual hints for edit/assign actions.
func RenderEditPanel(mode string) string {
	switch mode {
	case "edit":
		return "[edit mode] enter text, save to apply or esc to cancel"
	default:
		return ""
	}
}
