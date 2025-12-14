# GitHub Projects TUI Mock — Guidance for Implementers

This file explains how to use `github-projects-tui-mock.tsx` (a React mock) as a visual reference when implementing the TUI with Bubbletea and Lipgloss.

Goals:
- Capture visual hierarchy: header -> body -> footer; keep key information visible.
- Reproduce the mock's color palette, selection emphasis, and density within TUI constraints.

Key mapping from mock to TUI:
- Colors: use defined colors in `internal/ui/components/ui.go` as the source of truth.
- Header: project title (green), project name (blue), view tabs on the right (highlight selected in yellow).
- Board:
  - Column header: blue text on a darker panel with a border.
  - Card: small ID in green, title, then metadata (assignee/tags/priority). Selected card should use a yellow border and slightly brighter background.
  - Column width: target ~24-28 characters depending on terminal width.
- Table:
  - Add a header row styled with `TableHeaderCellStyle` and bordered cells.
  - ID in green, Status in cyan, Assignee in purple, Priority color-coded.
- Roadmap:
  - Timeline header boxed and subdued timestamp.
  - Each task has a progress bar (10 chars) and a percentage.

Practical tips:
- Lipgloss width calculations depend on characters; test at typical terminal widths (80, 120, 160).
- Use `lipgloss.NewStyle().Width(n)` for columns and `JoinHorizontal/JoinVertical` for consistent spacing.
- Reuse styles from `internal/ui/components/ui.go` and extend them only when necessary.

Workflow suggestion:
1. Update `internal/ui/components/ui.go` to add/adjust styles that reflect the mock.
2. Update `internal/ui/board/view.go`, `internal/ui/table/view.go`, and `internal/ui/roadmap/view.go` to consume those styles.
3. Run the app and iterate with different terminal sizes.

Note: perfect pixel parity is impossible in TUI; prioritize information density and emphasis.
