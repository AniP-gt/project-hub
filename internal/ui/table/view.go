package table

import (
	"strings"

	"project-hub/internal/state"
)

// Render renders an ASCII boxed table showing items with columns:
// Title | Status | Repository | Labels | Milestone | Priority
func Render(items []state.Item, focusedID string, innerWidth int) string {
	// Column headers
	headers := []string{"Title", "Status", "Repository", "Labels", "Milestone", "Priority"}

	// Collect column contents (as strings)
	rows := make([][]string, len(items))
	for i, it := range items {
		rows[i] = []string{
			it.Title,
			it.Status,
			it.Repository,
			strings.Join(it.Labels, ","),
			it.Milestone,
			it.Priority,
		}
	}

	// Determine column widths based on headers and content, leaving room for separators and marker
	markerWidth := 2                                     // one for marker + one space
	sepWidth := 3 * (len(headers) - 1)                   // " | " between columns
	available := innerWidth - markerWidth - sepWidth - 2 // 2 for frame borders '|' at ends
	if available < len(headers) {
		available = len(headers)
	}

	colWidths := make([]int, len(headers))
	// start with header widths
	for i := range headers {
		colWidths[i] = len(headers[i])
	}
	// expand to fit content up to available
	for _, r := range rows {
		for c, cell := range r {
			if len(cell) > colWidths[c] {
				colWidths[c] = len(cell)
			}
		}
	}

	// If sum of colWidths exceeds available, shrink columns with priority:
	// shrinkable order: Title, Labels, Repository, Milestone, Priority, Status
	order := []int{0, 3, 2, 4, 5, 1}
	sum := 0
	for _, w := range colWidths {
		sum += w
	}
	if sum > available {
		extra := sum - available
		for _, idx := range order {
			if extra <= 0 {
				break
			}
			canReduce := colWidths[idx] - 5 // minimum width 5
			if canReduce <= 0 {
				continue
			}
			reduction := canReduce
			if reduction > extra {
				reduction = extra
			}
			colWidths[idx] -= reduction
			extra -= reduction
		}
	}

	// helper to pad or truncate
	pad := func(s string, w int) string {
		if len(s) > w {
			if w <= 1 {
				return s[:w]
			}
			return s[:w-1] + "…"
		}
		return s + strings.Repeat(" ", w-len(s))
	}

	var b strings.Builder

	// top border (use '=' to avoid long '-' runs)
	b.WriteString("+" + strings.Repeat("=", innerWidth-2) + "+\n")

	// header row
	b.WriteString("| ")
	for i, h := range headers {
		b.WriteString(pad(h, colWidths[i]))
		if i < len(headers)-1 {
			b.WriteString(" | ")
		}
	}
	rem := innerWidth - 2 - (markerWidth - 1) - (func() int {
		s := 0
		for _, w := range colWidths {
			s += w
		}
		return s + sepWidth
	}())
	// pad to full width
	if rem > 0 {
		b.WriteString(strings.Repeat(" ", rem))
	}
	b.WriteString(" |")
	b.WriteString("\n")

	// header separator (use '=' as well)
	b.WriteString("+" + strings.Repeat("=", innerWidth-2) + "+\n")

	// data rows
	for _, r := range rows {
		b.WriteString("| ")
		for i, cell := range r {
			b.WriteString(pad(cell, colWidths[i]))
			if i < len(r)-1 {
				b.WriteString(" | ")
			}
		}
		// pad remainder
		rem := innerWidth - 2 - (func() int {
			s := 0
			for _, w := range colWidths {
				s += w
			}
			return s + sepWidth
		}())
		if rem > 0 {
			b.WriteString(strings.Repeat(" ", rem))
		}
		b.WriteString(" |")
		b.WriteString("\n")
	}

	// bottom border (match header separator)
	b.WriteString("+" + strings.Repeat("=", innerWidth-2) + "+")

	// add focus marker at line start for focused row(s)
	out := b.String()
	if focusedID == "" {
		return out
	}
	// Mark focused line: find the row index
	for i, it := range items {
		if it.ID == focusedID {
			// locate start of data row: after top border and header (3 lines)
			// compute line to replace
			lines := strings.Split(out, "\n")
			// data rows start at index 3 (0-based): top border(0), header(1), sep(2)
			lineIdx := 3 + i
			if lineIdx >= 0 && lineIdx < len(lines) {
				lines[lineIdx] = ">" + lines[lineIdx][1:]
				return strings.Join(lines, "\n")
			}
		}
	}
	return out
}
