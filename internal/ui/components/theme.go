package components

import "github.com/charmbracelet/lipgloss"

var (
	colorSurface = lipgloss.Color("#0f172a")
	colorPanel   = lipgloss.Color("#111827")
	colorText    = lipgloss.Color("#e5e7eb")
	colorMuted   = lipgloss.Color("#9ca3af")
	colorAccent  = lipgloss.Color("#7c3aed")
	colorGood    = lipgloss.Color("#22c55e")
	colorWarn    = lipgloss.Color("#f59e0b")
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Background(colorPanel).
			Foreground(colorText).
			Padding(0, 1).
			Bold(true)

	FooterStyle = lipgloss.NewStyle().
			Background(colorPanel).
			Foreground(colorMuted).
			Padding(0, 1)

	FrameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(1, 2)

	BadgeActive = lipgloss.NewStyle().
			Background(colorAccent).
			Foreground(colorText).
			Padding(0, 1).
			Bold(true)

	BadgeMuted = lipgloss.NewStyle().
			Background(colorMuted).
			Foreground(colorSurface).
			Padding(0, 1)

	BadgeInfo = lipgloss.NewStyle().
			Background(colorSurface).
			Foreground(colorText).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorMuted)

	NotifInfo = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorSurface).
			Padding(0, 1)

	NotifWarn = lipgloss.NewStyle().
			Foreground(colorSurface).
			Background(colorWarn).
			Padding(0, 1)
)

func InlineGap() string {
	return lipgloss.NewStyle().PaddingRight(1).Render("")
}
