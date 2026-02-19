# INTERNAL/UI KNOWLEDGE BASE

## OVERVIEW
Terminal UI rendering and interaction models for board, table, settings, and shared components.

## STRUCTURE
```
internal/ui/
├── board/          # kanban view (model, view, logic, render, layout)
├── table/          # table view
├── settings/       # settings panel
└── components/     # shared panels, selectors, empty state
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Board model | board/model.go | board state and Bubble Tea model |
| Board view | board/view.go | columns/cards rendering, navigation |
| Board logic | board/logic.go | business logic for board operations |
| Board render | board/render.go | string rendering helpers |
| Board layout | board/layout.go | layout calculations |
| Table view | table/view.go | row/column rendering |
| Settings panel | settings/settings.go | defaults editor + save flow |
| Detail panel | components/detail_panel.go | item detail popup |
| Status selector | components/status_selector.go | status selection UI |
| Field selector | components/field_selector.go | field toggle UI |
| Edit panel | components/edit_panel.go | item editing UI |
| Empty state | components/empty_state.go | no-items placeholder |
| Shared UI | components/ui.go | common UI helpers |
| UI tests | **/*_test.go | rendered output assertions |

## CONVENTIONS
- UI rendering is string-based; tests compare rendered output with `strings.Contains()`.
- Keep view-specific logic within its subdir (board/table/settings).
- Board is split across multiple files: model, view, logic, render, layout.

## ANTI-PATTERNS
- Avoid putting app state transitions here; keep those in `internal/app`.
- Avoid direct `gh` CLI or config access in UI; route through app/model.
