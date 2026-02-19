package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"project-hub/internal/app/core"
	"project-hub/internal/state"
	boardPkg "project-hub/internal/ui/board"
	"project-hub/internal/ui/components"
	"project-hub/internal/ui/table"
)

func (a App) View() string {
	width := a.state.Width
	if width == 0 {
		width = 100
	}
	header := components.RenderHeader(a.state.Project, a.state.View, width)
	items := state.ApplyFilter(a.state.Items, a.state.Project.Fields, a.state.View.Filter, time.Now())
	items = applySort(items, a.state.View.TableSort)

	frameWidth := width
	if frameWidth <= 0 {
		frameWidth = 100
	}
	innerWidth := frameWidth - components.FrameStyle.GetHorizontalFrameSize()
	if innerWidth < 40 {
		innerWidth = 40
	}

	editTitle := ""
	if a.state.View.Mode == "edit" || a.state.View.Mode == "assign" || a.state.View.Mode == "labelsInput" || a.state.View.Mode == "milestoneInput" || a.state.View.Mode == state.ModeFiltering {
		editTitle = a.textInput.Value()
	}
	footer := components.RenderFooter(string(a.state.View.Mode), string(a.state.View.CurrentView), width, editTitle)
	notif := components.RenderNotifications(a.state.Notifications)

	body := ""
	bodyHeight := a.bodyViewportHeight(header, footer, notif)
	frameVertical := components.FrameStyle.GetVerticalFrameSize()
	switch a.state.View.CurrentView {
	case state.ViewTable:
		groupBy := strings.ToLower(strings.TrimSpace(a.state.View.TableGroupBy))
		if groupBy != "" {
			groupedView := renderGroupedTable(groupBy, items, a.state.Project.Fields, a.state.View.FocusedItemID, a.state.View.FocusedColumnIndex, innerWidth)
			headerHeight := lipgloss.Height(groupedView.Header)
			rowsHeight := bodyHeight - headerHeight - frameVertical
			if rowsHeight < 3 {
				rowsHeight = 3
			}
			if a.tableViewport != nil {
				rowsContent := strings.Join(groupedView.Rows, "\n") + "\n"
				a.ensureTableViewportSize(innerWidth, rowsHeight)
				a.tableViewport.SetContent(rowsContent)
				focusedRow := focusedGroupedRowIndex(groupedView.Groups, a.state.View.FocusedItemID)
				focusTop, focusBottom := groupedView.RowBounds(focusedRow)
				a.syncTableViewportToFocus(focusTop, focusBottom, groupedView.RowsLineSize)
				body = lipgloss.JoinVertical(lipgloss.Left, groupedView.Header, a.tableViewport.View())
			} else {
				body = lipgloss.JoinVertical(lipgloss.Left, append([]string{groupedView.Header}, groupedView.Rows...)...)
			}
		} else {
			tableView := table.Render(items, a.state.View.FocusedItemID, a.state.View.FocusedColumnIndex, innerWidth)
			headerHeight := lipgloss.Height(tableView.Header)
			rowsHeight := bodyHeight - headerHeight - frameVertical
			if rowsHeight < 3 {
				rowsHeight = 3
			}
			if a.tableViewport != nil {
				rowsContent := strings.Join(tableView.Rows, "\n") + "\n"
				a.ensureTableViewportSize(innerWidth, rowsHeight)
				a.tableViewport.SetContent(rowsContent)
				focusedRow := focusedRowIndex(items, a.state.View.FocusedItemID)
				focusTop, focusBottom := tableView.RowBounds(focusedRow)
				a.syncTableViewportToFocus(focusTop, focusBottom, tableView.RowsLineSize)
				body = lipgloss.JoinVertical(lipgloss.Left, tableView.Header, a.tableViewport.View())
			} else {
				body = lipgloss.JoinVertical(lipgloss.Left, append([]string{tableView.Header}, tableView.Rows...)...)
			}
		}
	case state.ViewSettings:
		a.settingsModel.SetSize(innerWidth, bodyHeight)
		body = a.settingsModel.View()
	default:
		a.boardModel.Width = innerWidth
		a.boardModel.Height = bodyHeight
		a.boardModel.EnsureLayout()
		body = a.boardModel.View()
	}

	if a.state.View.Mode == "edit" || a.state.View.Mode == "assign" || a.state.View.Mode == "labelsInput" || a.state.View.Mode == "milestoneInput" || a.state.View.Mode == state.ModeFiltering {
		body = body + "\n" + a.textInput.View()
	}

	var framed string
	if a.state.View.CurrentView == state.ViewBoard {
		maxHeight := bodyHeight
		if maxHeight < 15 {
			maxHeight = 15
		}
		if maxHeight < 1 {
			maxHeight = 1
		}
		bodyRendered := clampContentHeight(body, maxHeight)
		bodyRendered = lipgloss.NewStyle().Width(innerWidth).AlignHorizontal(lipgloss.Left).Render(bodyRendered)
		framed = components.FrameStyle.Width(frameWidth).Render(bodyRendered)
	} else {
		maxHeight := bodyHeight - frameVertical
		if maxHeight < 1 {
			maxHeight = 1
		}
		bodyRendered := clampContentHeight(body, maxHeight)
		bodyRendered = lipgloss.NewStyle().Width(innerWidth).AlignHorizontal(lipgloss.Left).Render(bodyRendered)
		framed = components.FrameStyle.Width(frameWidth).Render(bodyRendered)
	}

	if a.state.View.Mode == state.ModeStatusSelect {
		selectorView := a.statusSelector.View()
		framed = lipgloss.Place(
			frameWidth,
			bodyHeight,
			lipgloss.Center,
			lipgloss.Center,
			selectorView,
		)
	}

	if a.state.View.Mode == state.ModeLabelSelect || a.state.View.Mode == state.ModeMilestoneSelect || a.state.View.Mode == state.ModePrioritySelect {
		selectorView := a.fieldSelector.View()
		framed = lipgloss.Place(
			frameWidth,
			bodyHeight,
			lipgloss.Center,
			lipgloss.Center,
			selectorView,
		)
	}

	if a.state.View.Mode == state.ModeDetail {
		detailView := a.detailPanel.View()
		framed = lipgloss.Place(
			frameWidth,
			bodyHeight,
			lipgloss.Center,
			lipgloss.Center,
			detailView,
		)
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, framed, footer, notif)
}

type groupedTableView struct {
	Header       string
	Rows         []string
	RowHeights   []int
	RowOffsets   []int
	RowsLineSize int
	Groups       []boardPkg.GroupBucket
}

func renderGroupedTable(groupBy string, items []state.Item, fields []state.Field, focusedID string, focusedColIndex int, innerWidth int) groupedTableView {
	if innerWidth <= 0 {
		innerWidth = 80
	}
	sepLen := innerWidth
	var groups []boardPkg.GroupBucket
	switch groupBy {
	case core.GroupByStatus:
		groups = boardPkg.GroupItemsByStatusBuckets(items, fields)
	case core.GroupByIteration:
		groups = boardPkg.GroupItemsByIteration(items)
	case core.GroupByAssignee:
		groups = boardPkg.GroupItemsByAssignee(items)
	default:
		groups = []boardPkg.GroupBucket{{Name: "Items", Items: items}}
	}
	if len(groups) == 0 {
		groups = []boardPkg.GroupBucket{{Name: "Items", Items: items}}
	}

	var header string
	var rows []string
	var rowHeights []int
	var rowOffsets []int
	var cumulativeHeight int
	groupHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

	for i, group := range groups {
		if i == 0 {
			groupRender := table.Render(group.Items, focusedID, focusedColIndex, innerWidth)
			header = groupRender.Header
		}

		groupHeader := fmt.Sprintf("# %s (%d)", group.Name, len(group.Items))
		groupHeaderRow := groupHeaderStyle.Render(groupHeader)
		separator := ""
		if sepLen > 0 {
			separator = strings.Repeat("─", sepLen)
		}
		groupHeaderView := lipgloss.JoinVertical(lipgloss.Left, groupHeaderRow, separator)
		rows = append(rows, groupHeaderView)
		rowHeights = append(rowHeights, lipgloss.Height(groupHeaderView))
		rowOffsets = append(rowOffsets, cumulativeHeight)
		cumulativeHeight += lipgloss.Height(groupHeaderView)

		groupRender := table.Render(group.Items, focusedID, focusedColIndex, innerWidth)
		rows = append(rows, groupRender.Rows...)
		for _, h := range groupRender.RowHeights {
			rowHeights = append(rowHeights, h)
			rowOffsets = append(rowOffsets, cumulativeHeight)
			cumulativeHeight += h
		}
	}

	return groupedTableView{
		Header:       header,
		Rows:         rows,
		RowHeights:   rowHeights,
		RowOffsets:   rowOffsets,
		RowsLineSize: cumulativeHeight,
		Groups:       groups,
	}
}

func (g groupedTableView) RowBounds(index int) (int, int) {
	if index < 0 || index >= len(g.RowOffsets) || index >= len(g.RowHeights) {
		return -1, -1
	}
	top := g.RowOffsets[index]
	height := g.RowHeights[index]
	if height <= 0 {
		height = 1
	}
	bottom := top + height - 1
	return top, bottom
}

func focusedGroupedRowIndex(groups []boardPkg.GroupBucket, focusedID string) int {
	index := 0
	for _, group := range groups {
		index++
		for _, item := range group.Items {
			if item.ID == focusedID {
				return index
			}
			index++
		}
	}
	return -1
}

func toggleSort(ts state.TableSort, field string) state.TableSort {
	if ts.Field == field {
		ts.Asc = !ts.Asc
		return ts
	}
	return state.TableSort{Field: field, Asc: true}
}

func applySort(items []state.Item, ts state.TableSort) []state.Item {
	if ts.Field == "" {
		return items
	}
	switch ts.Field {
	case "Title":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Title < items[j].Title
			}
			return items[i].Title > items[j].Title
		})
	case "Status":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Status < items[j].Status
			}
			return items[i].Status > items[j].Status
		})
	case "Repository":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Repository < items[j].Repository
			}
			return items[i].Repository > items[j].Repository
		})
	case "Labels":
		sort.SliceStable(items, func(i, j int) bool {
			iLabels := strings.Join(items[i].Labels, ",")
			jLabels := strings.Join(items[j].Labels, ",")
			if ts.Asc {
				return iLabels < jLabels
			}
			return iLabels > jLabels
		})
	case "Milestone":
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return items[i].Milestone < items[j].Milestone
			}
			return items[i].Milestone > items[j].Milestone
		})
	case "Priority":
		priorityRank := func(p string) int {
			switch strings.ToLower(p) {
			case "high":
				return 3
			case "medium":
				return 2
			case "low":
				return 1
			default:
				return 0
			}
		}
		sort.SliceStable(items, func(i, j int) bool {
			if ts.Asc {
				return priorityRank(items[i].Priority) < priorityRank(items[j].Priority)
			}
			return priorityRank(items[i].Priority) > priorityRank(items[j].Priority)
		})
	default:
		switch ts.Field {
		case "Number":
			sort.SliceStable(items, func(i, j int) bool {
				if ts.Asc {
					return items[i].Number < items[j].Number
				}
				return items[i].Number > items[j].Number
			})
		case "CreatedAt":
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].CreatedAt == nil || items[j].CreatedAt == nil {
					return items[i].CreatedAt == nil
				}
				if ts.Asc {
					return items[i].CreatedAt.Before(*items[j].CreatedAt)
				}
				return items[j].CreatedAt.Before(*items[i].CreatedAt)
			})
		case "UpdatedAt":
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].UpdatedAt == nil || items[j].UpdatedAt == nil {
					return items[i].UpdatedAt == nil
				}
				if ts.Asc {
					return items[i].UpdatedAt.Before(*items[j].UpdatedAt)
				}
				return items[j].UpdatedAt.Before(*items[i].UpdatedAt)
			})
		default:
		}
	}
	return items
}

func (a App) bodyViewportHeight(header, footer, notif string) int {
	height := a.state.Height
	if height <= 0 {
		height = 40
	}
	available := height - lipgloss.Height(header) - lipgloss.Height(footer)
	if strings.TrimSpace(notif) != "" {
		available -= lipgloss.Height(notif)
	}
	available -= components.FrameStyle.GetVerticalFrameSize()
	if available < 5 {
		available = 5
	}
	return available
}

func clampContentHeight(content string, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) >= height {
		return strings.Join(lines[:height], "\n")
	}
	padding := make([]string, height-len(lines))
	return strings.Join(append(lines, padding...), "\n")
}

func (a App) ensureTableViewportSize(width, height int) {
	if a.tableViewport == nil {
		return
	}
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if a.tableViewport.Width != width {
		a.tableViewport.Width = width
	}
	if a.tableViewport.Height != height {
		a.tableViewport.Height = height
	}
}

func (a App) syncTableViewportToFocus(focusTop, focusBottom, totalLines int) {
	if a.tableViewport == nil {
		return
	}
	vp := a.tableViewport
	if totalLines <= 0 || focusTop < 0 || focusBottom < 0 {
		vp.YOffset = 0
		return
	}
	visibleHeight := vp.Height
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	if focusBottom-focusTop+1 > visibleHeight {
		focusBottom = focusTop + visibleHeight - 1
	}
	top := vp.YOffset
	bottom := vp.YOffset + visibleHeight - 1
	if focusTop < top {
		vp.SetYOffset(focusTop)
		return
	}
	if focusBottom > bottom {
		vp.SetYOffset(focusBottom - visibleHeight + 1)
	}
	maxOffset := totalLines - visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if vp.YOffset > maxOffset {
		vp.YOffset = maxOffset
	}
	if vp.YOffset < 0 {
		vp.YOffset = 0
	}
}

func focusedRowIndex(items []state.Item, focusedID string) int {
	if focusedID == "" {
		return -1
	}
	for idx, it := range items {
		if it.ID == focusedID {
			return idx
		}
	}
	return -1
}
