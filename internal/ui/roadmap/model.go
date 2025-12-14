package roadmap

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"

	"project-hub/internal/state"
	"project-hub/internal/ui/components"
)

// RoadmapModel holds state for the roadmap view with scrolling.
type RoadmapModel struct {
	Timelines    []state.Timeline
	Items        []state.Item
	FocusedIndex int // index into Items slice (filtered by timeline ordering as used in Render)
	Width        int
	Height       int
	RowOffset    int
}

func NewRoadmapModel(timelines []state.Timeline, items []state.Item, focusedID string) RoadmapModel {
	idx := 0
	for i, it := range items {
		if it.ID == focusedID {
			idx = i
			break
		}
	}
	return RoadmapModel{Timelines: timelines, Items: items, FocusedIndex: idx}
}

func (m RoadmapModel) Init() tea.Cmd { return nil }

func (m RoadmapModel) calculateMaxVisibleRows() int {
	if m.Height <= 0 {
		return 6
	}
	available := m.Height - 2
	if available < 3 {
		return 1
	}
	return available
}

func (m RoadmapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch tm := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = tm.Width
		m.Height = tm.Height
		max := m.calculateMaxVisibleRows()
		if m.FocusedIndex < m.RowOffset {
			m.RowOffset = m.FocusedIndex
		} else if m.FocusedIndex >= m.RowOffset+max {
			m.RowOffset = m.FocusedIndex - (max - 1)
		}
	case tea.KeyMsg:
		s := tm.String()
		switch s {
		case "j", "down":
			if m.FocusedIndex < len(m.Items)-1 {
				m.FocusedIndex++
				max := m.calculateMaxVisibleRows()
				if m.FocusedIndex >= m.RowOffset+max {
					m.RowOffset++
				}
			}
		case "k", "up":
			if m.FocusedIndex > 0 {
				m.FocusedIndex--
				if m.FocusedIndex < m.RowOffset {
					m.RowOffset--
				}
			}
		case "g":
			m.FocusedIndex = 0
			m.RowOffset = 0
		case "G":
			m.FocusedIndex = len(m.Items) - 1
			max := m.calculateMaxVisibleRows()
			if m.FocusedIndex-max+1 > 0 {
				m.RowOffset = m.FocusedIndex - max + 1
			} else {
				m.RowOffset = 0
			}
		case "ctrl+u":
			max := m.calculateMaxVisibleRows()
			delta := max / 2
			m.FocusedIndex -= delta
			if m.FocusedIndex < 0 {
				m.FocusedIndex = 0
			}
			if m.FocusedIndex < m.RowOffset {
				m.RowOffset = m.FocusedIndex
			}
		case "ctrl+d":
			max := m.calculateMaxVisibleRows()
			delta := max / 2
			m.FocusedIndex += delta
			if m.FocusedIndex > len(m.Items)-1 {
				m.FocusedIndex = len(m.Items) - 1
			}
			if m.FocusedIndex >= m.RowOffset+max {
				m.RowOffset = m.FocusedIndex - max + 1
			}
		}
	}
	return m, cmd
}

func (m RoadmapModel) View() string {
	// Build ordered list of display items grouped by timeline
	var blocks []string
	for _, tl := range m.Timelines {
		header := components.RoadmapTimelineStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Bold(true).Render(tl.Name), lipgloss.NewStyle().Foreground(components.ColorGray500).Render("  (timeline)")))
		blocks = append(blocks, header)
		for _, it := range m.Items {
			if it.IterationID != tl.ID {
				continue
			}
			blocks = append(blocks, components.RoadmapItemBaseStyle.Render("")) // placeholder, actual rendering moved to flat list below
		}
	}
	// Unscheduled
	for _, it := range m.Items {
		if it.IterationID == "" {
			blocks = append(blocks, components.RoadmapTimelineStyle.Render("[unscheduled]"))
			break
		}
	}

	// Build flat list of items in the same order as Render would show, so offsets map naturally.
	var flat []string
	for _, tl := range m.Timelines {
		// timeline header counts as an item
		flat = append(flat, components.RoadmapTimelineStyle.Render(tl.Name))
		for _, it := range m.Items {
			if it.IterationID != tl.ID {
				continue
			}
			isSelected := false
			// determine selection by comparing titles/IDs in flat list position
			flat = append(flat, formatRoadmapItem(it, isSelected))
		}
	}
	for _, it := range m.Items {
		if it.IterationID == "" {
			flat = append(flat, components.RoadmapTimelineStyle.Render("[unscheduled]"))
			flat = append(flat, formatRoadmapItem(it, false))
		}
	}

	// Now render visible slice based on RowOffset/Height
	max := m.calculateMaxVisibleRows()
	start := m.RowOffset
	end := start + max
	if end > len(flat) {
		end = len(flat)
	}
	visible := flat[start:end]
	body := strings.TrimRight(lipgloss.JoinVertical(lipgloss.Left, visible...), "\n")
	return body
}

func formatRoadmapItem(it state.Item, selected bool) string {
	isSelected := selected
	rowStyle := components.RoadmapItemBaseStyle
	if isSelected {
		rowStyle = components.RoadmapItemSelectedStyle
	}
	percent := 0
	bar := renderProgressBar(percent)
	percentLabel := components.RoadmapProgressPercentStyle.Render("0%")
	title := lipgloss.NewStyle().Foreground(components.ColorGray200).Render(it.Title)
	content := lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.JoinHorizontal(lipgloss.Top, bar, percentLabel))
	return rowStyle.Render(content)
}
