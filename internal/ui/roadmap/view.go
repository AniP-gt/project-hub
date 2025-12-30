package roadmap

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

const progressSegments = 10

func Render(timelines []state.Timeline, items []state.Item, focusedID string, statusProgress map[string]int) (string, int) {
	type sectionBlock struct {
		content   string
		focusLine int
	}

	grouped := groupItemsByTimeline(items)
	var blocks []sectionBlock

	appendBlock := func(content string, focusLine int) {
		if strings.TrimSpace(content) == "" {
			return
		}
		blocks = append(blocks, sectionBlock{content: content, focusLine: focusLine})
	}

	for _, tl := range timelines {
		section, relFocus := renderTimelineSection(tl, grouped[tl.ID], focusedID, statusProgress)
		appendBlock(section, relFocus)
		delete(grouped, tl.ID)
	}

	if unscheduled := grouped[""]; len(unscheduled) > 0 {
		section, relFocus := renderUnscheduledSection(unscheduled, focusedID, statusProgress)
		appendBlock(section, relFocus)
		delete(grouped, "")
	}

	for timelineID, orphaned := range grouped {
		if len(orphaned) == 0 {
			continue
		}
		phantomTimeline := state.Timeline{ID: timelineID, Name: fmt.Sprintf("Timeline %s", timelineID)}
		section, relFocus := renderTimelineSection(phantomTimeline, orphaned, focusedID, statusProgress)
		appendBlock(section, relFocus)
	}

	overview := renderOverview(timelines, groupedItemsWithFallback(timelines, items), statusProgress)
	if overview != "" {
		appendBlock(overview, -1)
	}

	if len(blocks) == 0 {
		return "", -1
	}

	var sections []string
	focusLine := -1
	lineCursor := 0
	for i, block := range blocks {
		if i > 0 {
			lineCursor++
		}
		if block.focusLine >= 0 {
			focusLine = lineCursor + block.focusLine
		}
		sections = append(sections, block.content)
		lineCursor += lipgloss.Height(block.content)
	}

	return strings.TrimSpace(lipgloss.JoinVertical(lipgloss.Left, sections...)), focusLine
}

func renderTimelineSection(timeline state.Timeline, items []state.Item, focusedID string, statusProgress map[string]int) (string, int) {
	header := renderTimelineHeader(timeline)
	if len(items) == 0 {
		empty := components.RoadmapItemBaseStyle.Render("No items scheduled")
		return lipgloss.JoinVertical(lipgloss.Left, header, empty), -1
	}

	var cards []string
	focusLineInList := -1
	lineOffset := 0
	for idx, item := range items {
		selected := focusedID == item.ID
		card := renderItemCard(item, timeline.Name, selected, statusProgress)
		cards = append(cards, card)
		if selected {
			focusLineInList = lineOffset
		}
		lineOffset += lipgloss.Height(card)
		if idx < len(items)-1 {
			lineOffset++
		}
	}

	list := lipgloss.JoinVertical(lipgloss.Left, cards...)
	content := lipgloss.JoinVertical(lipgloss.Left, header, list)
	if focusLineInList >= 0 {
		return content, lipgloss.Height(header) + focusLineInList
	}
	return content, -1
}

func renderUnscheduledSection(items []state.Item, focusedID string, statusProgress map[string]int) (string, int) {
	timeline := state.Timeline{Name: "Unscheduled"}
	return renderTimelineSection(timeline, items, focusedID, statusProgress)
}

func renderTimelineHeader(timeline state.Timeline) string {
	name := lipgloss.NewStyle().Foreground(components.ColorBlue300).Bold(true).Render(timeline.Name)
	var metaParts []string
	if timeline.Start != nil && timeline.End != nil {
		metaParts = append(metaParts, fmt.Sprintf("%s → %s", timeline.Start.Format("2006-01-02"), timeline.End.Format("2006-01-02")))
	} else if timeline.Start != nil {
		metaParts = append(metaParts, fmt.Sprintf("Starts %s", timeline.Start.Format("2006-01-02")))
	} else if timeline.End != nil {
		metaParts = append(metaParts, fmt.Sprintf("Ends %s", timeline.End.Format("2006-01-02")))
	}
	if timeline.Progress != "" {
		metaParts = append(metaParts, timeline.Progress)
	}

	meta := lipgloss.NewStyle().Foreground(components.ColorGray500).Render(strings.Join(metaParts, "   "))
	content := lipgloss.JoinVertical(lipgloss.Left, name, meta)
	return components.RoadmapTimelineStyle.Render(content)
}

func renderItemCard(item state.Item, sprintName string, selected bool, statusProgress map[string]int) string {
	title := components.CardTitleStyle.Render(item.Title)
	sprint := components.RoadmapItemSprintStyle.Render(sprintName)
	status := components.TableCellStatusStyle.Render(item.Status)
	repo := components.TableCellIDStyle.Render(item.Repository)

	meta := lipgloss.JoinHorizontal(lipgloss.Left,
		status,
		lipgloss.NewStyle().PaddingLeft(2).Render(repo),
	)
	progressPct := progressPercentForStatus(item.Status, statusProgress)
	progressBar := components.RoadmapProgressBarStyle.Render(renderProgressBar(progressPct))
	progressLine := lipgloss.JoinHorizontal(lipgloss.Left,
		progressBar,
		lipgloss.NewStyle().PaddingLeft(1).Render(components.RoadmapProgressPercentStyle.Render(fmt.Sprintf("%d%%", progressPct))),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, title, lipgloss.NewStyle().PaddingLeft(2).Render(sprint)),
		meta,
		progressLine,
	)

	style := components.RoadmapItemBaseStyle
	if selected {
		style = components.RoadmapItemSelectedStyle
	}

	return style.Render(content)
}

func renderOverview(timelines []state.Timeline, grouped map[string][]state.Item, statusProgress map[string]int) string {
	if len(timelines) == 0 {
		return ""
	}

	var rows []string
	rows = append(rows, components.RoadmapOverviewHeaderStyle.Render("Sprint Progress Overview:"))
	for _, tl := range timelines {
		items := grouped[tl.ID]
		percent := timelineProgressPercent(tl, items, statusProgress)
		row := lipgloss.JoinHorizontal(lipgloss.Left,
			components.RoadmapOverviewSprintNameStyle.Render(fmt.Sprintf("%s:", tl.Name)),
			components.RoadmapProgressBarStyle.Render(renderProgressBar(percent)),
			lipgloss.NewStyle().PaddingLeft(1).Render(components.RoadmapProgressPercentStyle.Render(fmt.Sprintf("%d%%", percent))),
		)
		rows = append(rows, row)
	}

	container := components.RoadmapItemBaseStyle.Copy().
		BorderForeground(components.ColorGray700).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
	return container
}

func groupItemsByTimeline(items []state.Item) map[string][]state.Item {
	grouped := make(map[string][]state.Item)
	for _, item := range items {
		key := item.IterationID
		grouped[key] = append(grouped[key], item)
	}
	return grouped
}

func groupedItemsWithFallback(timelines []state.Timeline, items []state.Item) map[string][]state.Item {
	grouped := groupItemsByTimeline(items)
	for _, tl := range timelines {
		if _, ok := grouped[tl.ID]; !ok {
			grouped[tl.ID] = nil
		}
	}
	return grouped
}

func renderProgressBar(percent int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := (percent*progressSegments + 50) / 100
	if filled > progressSegments {
		filled = progressSegments
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", progressSegments-filled)
}

func timelineProgressPercent(timeline state.Timeline, items []state.Item, statusProgress map[string]int) int {
	if pct := parsePercent(timeline.Progress); pct >= 0 {
		return pct
	}
	if len(items) == 0 {
		return 0
	}
	sum := 0
	for _, item := range items {
		sum += progressPercentForStatus(item.Status, statusProgress)
	}
	return sum / len(items)
}

func progressPercentForStatus(status string, statusProgress map[string]int) int {
	if pct, ok := lookupStatusPercent(status, statusProgress); ok {
		return pct
	}
	switch strings.ToLower(status) {
	case "done", "completed", "closed":
		return 100
	case "review", "in review":
		return 80
	case "in progress", "doing":
		return 60
	case "todo", "backlog", "open", "blocked":
		return 20
	default:
		return 40
	}
}

func lookupStatusPercent(status string, statusProgress map[string]int) (int, bool) {
	if len(statusProgress) == 0 {
		return 0, false
	}
	status = strings.TrimSpace(status)
	if status == "" {
		return 0, false
	}
	if pct, ok := statusProgress[status]; ok {
		return pct, true
	}
	key := strings.ToLower(status)
	pct, ok := statusProgress[key]
	return pct, ok
}

func parsePercent(raw string) int {
	trimmed := strings.TrimSpace(strings.TrimSuffix(raw, "%"))
	if trimmed == "" {
		return -1
	}
	var pct int
	if _, err := fmt.Sscanf(trimmed, "%d", &pct); err != nil {
		return -1
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return pct
}
