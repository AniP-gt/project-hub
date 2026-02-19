package components

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"project-hub/internal/state"
)

var (
	// Colors
	ColorBlack     = lipgloss.Color("#000000")
	ColorGray900   = lipgloss.Color("#111827") // bg-gray-900
	ColorGray800   = lipgloss.Color("#1F2937") // bg-gray-800
	ColorGray700   = lipgloss.Color("#374151") // border-gray-700, bg-gray-700
	ColorGray500   = lipgloss.Color("#6B7280") // text-gray-500, hover:border-gray-500
	ColorGray400   = lipgloss.Color("#9CA3AF") // text-gray-400
	ColorGray300   = lipgloss.Color("#D1D5DB") // text-gray-300
	ColorGray200   = lipgloss.Color("#E5E7EB") // text-gray-200
	ColorGreen500  = lipgloss.Color("#22C55E") // text-green-500
	ColorGreen400  = lipgloss.Color("#4ADE80") // text-green-400
	ColorBlue400   = lipgloss.Color("#60A5FA") // text-blue-400
	ColorBlue300   = lipgloss.Color("#93C5FD") // text-blue-300
	ColorYellow400 = lipgloss.Color("#FACC15") // text-yellow-400
	ColorRed400    = lipgloss.Color("#F87171") // text-red-400
	ColorCyan400   = lipgloss.Color("#22D3EE") // text-cyan-400
	ColorPurple400 = lipgloss.Color("#C084FC") // text-purple-400
	ColorAccent    = lipgloss.Color("#7c3aed")
	ColorWhite     = lipgloss.Color("#FFFFFF")
	ColorText      = lipgloss.Color("#e5e7eb")
	ColorMuted     = lipgloss.Color("#9ca3af")
	ColorSurface   = lipgloss.Color("#0f172a")

	// Base Styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorGreen400).
			Background(ColorBlack)

	// Header Styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorGreen400).
			Padding(1, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorGray700)

	HeaderTitleStyle = lipgloss.NewStyle().
				Foreground(ColorGreen500).
				Bold(true)

	HeaderProjectStyle = lipgloss.NewStyle().
				Foreground(ColorBlue400)

	HeaderViewSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorYellow400).
				Bold(true)

	HeaderViewUnselectedStyle = lipgloss.NewStyle().
					Foreground(ColorGray500)

	// Footer Styles
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorGray400).
			Padding(1, 2).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ColorGray700)

	FooterModeStyle = lipgloss.NewStyle().
			Foreground(ColorGreen500)

	FooterKeybindsStyle = lipgloss.NewStyle().
				Foreground(ColorWhite)

	// Card Styles (Kanban Board)
	CardBaseStyle = lipgloss.NewStyle().
			Background(ColorGray800).
			Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(ColorGray700).
			Padding(0, 1).
			Foreground(ColorGray300)

	CardSelectedStyle = CardBaseStyle.Copy().
				BorderForeground(ColorYellow400).
				Background(ColorGray700)

	CardHoverStyle = CardBaseStyle.Copy(). // For TUI, hover is often simulated by selection
			BorderForeground(ColorGray500)

	CardIDStyle = lipgloss.NewStyle().
			Foreground(ColorGreen400).
			MarginBottom(0) // text-xs mb-1

	CardTagStyle = lipgloss.NewStyle().
			Foreground(ColorCyan400).
			MarginLeft(1)

	CardAssigneeStyle = lipgloss.NewStyle().
				Foreground(ColorPurple400)

	CardTitleStyle = lipgloss.NewStyle().
			Foreground(ColorGray200)

	// Column Styles (Kanban)
	ColumnHeaderStyle = lipgloss.NewStyle().
				Background(ColorGray800).
				Foreground(ColorBlue300).
				Padding(0, 1). // px-3 py-2
				Bold(true).
				Border(lipgloss.NormalBorder(), true, true, true, true).
				BorderForeground(ColorGray700)

	ColumnContainerStyle = lipgloss.NewStyle().
				Width(24).     // w-64 (approx 24 chars in monospaced font)
				MarginRight(0) // gap-2 (reduced from 2 to make columns closer)

	// Table Styles
	TableBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorGray700)

	TableHeaderCellStyle = lipgloss.NewStyle().
				Background(ColorGray800).
				Foreground(ColorBlue300).
				Padding(0, 1). // px-3 py-2
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorGray700).
				Align(lipgloss.Left)

	TableRowBaseStyle = lipgloss.NewStyle().
				Background(ColorGray900).
				Foreground(ColorGray300)

	TableRowSelectedStyle = lipgloss.NewStyle().
				Background(ColorGray700).
				Foreground(ColorYellow400)

	TableRowHoverStyle = lipgloss.NewStyle(). // Simulated by selection
				Background(ColorGray800)

	TableCellIDStyle = lipgloss.NewStyle().
				Foreground(ColorGreen400)

	TableCellStatusStyle = lipgloss.NewStyle().
				Foreground(ColorCyan400)

	TableCellAssigneeStyle = lipgloss.NewStyle().
				Foreground(ColorPurple400)

	TableCellPriorityHighStyle = lipgloss.NewStyle().
					Foreground(ColorRed400)

	TableCellPriorityMediumStyle = lipgloss.NewStyle().
					Foreground(ColorYellow400)

	TableCellPriorityLowStyle = lipgloss.NewStyle().
					Foreground(ColorGreen400)

	CardPriorityStyle = lipgloss.NewStyle().
				Foreground(ColorYellow400) // Default for medium, can be overridden

	TableCellUpdatedStyle = lipgloss.NewStyle().
				Foreground(ColorGray500)
)

func StatusColor(status string) lipgloss.Color {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized == "" {
		return ColorGray400
	}
	switch {
	case normalized == "open" || normalized == "opened":
		return ColorGreen500
	case normalized == "closed" || normalized == "done" || normalized == "merged":
		return ColorPurple400
	default:
		return ColorGreen500
	}
}

func StatusDot(status string) string {
	return lipgloss.NewStyle().Foreground(StatusColor(status)).Render("●")
}

var (
	FrameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGray700).
			Padding(1, 2)

	BadgeActive = lipgloss.NewStyle().
			Background(ColorAccent).
			Foreground(ColorText).
			Padding(0, 1).
			Bold(true)

	BadgeMuted = lipgloss.NewStyle().
			Background(ColorMuted).
			Foreground(ColorSurface).
			Padding(0, 1)

	BadgeInfo = lipgloss.NewStyle().
			Background(ColorSurface).
			Foreground(ColorText).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted)

	NotifInfo = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorSurface).
			Padding(0, 1)

	NotifWarn = lipgloss.NewStyle().
			Foreground(ColorSurface).
			Background(ColorRed400). // Changed to Red400 for warning
			Padding(0, 1)
)

func InlineGap() string {
	return lipgloss.NewStyle().PaddingRight(1).Render("")
}

// RenderHeader shows project name, view, and active filter/mode hints with badges.
func RenderHeader(project state.Project, view state.ViewContext, width int) string {
	title := HeaderTitleStyle.Render("█ GitHub Projects TUI")
	projectName := HeaderProjectStyle.Render("Project: " + project.Name)

	leftContent := lipgloss.JoinHorizontal(lipgloss.Top, title, lipgloss.NewStyle().Foreground(ColorGray500).Render(" | "), projectName)

	var middleContent string
	if view.CurrentView == state.ViewBoard {
		// Show status summary for board view
		statuses := []string{"Backlog", "In Progress", "Review", "Done"}
		statusParts := []string{}
		for _, status := range statuses {
			statusParts = append(statusParts, HeaderProjectStyle.Render(status))
		}
		middleContent = lipgloss.JoinHorizontal(lipgloss.Top, statusParts...)
		middleContent = lipgloss.NewStyle().Foreground(ColorGray400).Render(" | Status: ") + middleContent
	}

	rightContent := renderViewTabs(view.CurrentView)

	// Calculate widths
	leftWidth := lipgloss.Width(leftContent)
	middleWidth := lipgloss.Width(middleContent)
	rightWidth := lipgloss.Width(rightContent)
	totalContentWidth := leftWidth + middleWidth + rightWidth

	spacerWidth := 0
	if width > totalContentWidth {
		spacerWidth = width - totalContentWidth
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, middleContent, lipgloss.NewStyle().Width(spacerWidth).Render(""), rightContent)

	style := HeaderStyle
	if width > 0 {
		style = style.Width(width)
	}
	return style.Render(content)
}

func renderViewTabs(currentView state.ViewType) string {
	boardTab := HeaderViewUnselectedStyle.Render("[1:Board]")
	tableTab := HeaderViewUnselectedStyle.Render("[2:Table]")
	settingsTab := HeaderViewUnselectedStyle.Render("[3:Settings]")

	switch currentView {
	case state.ViewBoard:
		boardTab = HeaderViewSelectedStyle.Render("[1:Board]")
	case state.ViewTable:
		tableTab = HeaderViewSelectedStyle.Render("[2:Table]")
	case state.ViewSettings:
		settingsTab = HeaderViewSelectedStyle.Render("[3:Settings]")
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, boardTab, tableTab, settingsTab)
}

// RenderFooter shows key hints and mode status.
func RenderFooter(mode, view string, width int, editTitle string) string {
	// Mock's footer keybinds: j/k:移動 h/l:列移動 i:編集 /:フィルタ a:アサイン 1-3:ビュー切替 q:終了
	keybinds := FooterKeybindsStyle.Render("j/k:move i:edit /:filter a:assign o:detail f:fields 1-3:view q:quit")
	var modeLabel string
	modeStyle := FooterModeStyle
	switch strings.ToLower(mode) {
	case "edit":
		if editTitle != "" {
			modeLabel = "INSERT MODE " + editTitle
		} else {
			modeLabel = "INSERT MODE"
		}
		modeStyle = FooterModeStyle.Copy().Foreground(ColorYellow400)
	case "assign":
		if editTitle != "" {
			modeLabel = "ASSIGN MODE " + editTitle
		} else {
			modeLabel = "ASSIGN MODE"
		}
		modeStyle = FooterModeStyle.Copy().Foreground(ColorCyan400)
	case "labelsinput":
		if editTitle != "" {
			modeLabel = "INSERT LABELS " + editTitle
		} else {
			modeLabel = "INSERT LABELS"
		}
		modeStyle = FooterModeStyle.Copy().Foreground(ColorGreen400)
	case "milestoneinput":
		if editTitle != "" {
			modeLabel = "INSERT MILESTONE " + editTitle
		} else {
			modeLabel = "INSERT MILESTONE"
		}
		modeStyle = FooterModeStyle.Copy().Foreground(ColorPurple400)
	case "filtering":
		if editTitle != "" {
			modeLabel = "FILTER MODE " + editTitle
		} else {
			modeLabel = "FILTER MODE"
		}
	case "detail":
		modeLabel = "DETAIL MODE"
	case "fieldtoggle":
		modeLabel = "FIELD TOGGLE MODE"
	default:
		modeLabel = "NORMAL MODE"
	}
	modeStatus := modeStyle.Render(modeLabel)

	// Calculate remaining width for spacing
	totalKeybindsWidth := lipgloss.Width(keybinds)
	totalModeStatusWidth := lipgloss.Width(modeStatus)

	// Ensure there's enough space for both elements and a gap
	// If not, prioritize keybinds and truncate mode status or adjust layout
	availableSpace := width - totalKeybindsWidth - totalModeStatusWidth
	gap := ""
	if availableSpace > 0 {
		gap = strings.Repeat(" ", availableSpace)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, modeStatus, gap, keybinds)

	style := FooterStyle
	if width > 0 {
		style = style.Width(width)
	}
	return style.Render(content)
}

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
