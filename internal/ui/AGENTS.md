# INTERNAL/UI KNOWLEDGE BASE

## OVERVIEW
Terminal UI rendering and interaction models for board, table, settings, and shared components.

## STRUCTURE
```
internal/ui/
├── board/       # kanban view
├── table/       # table view
├── settings/    # settings panel
└── components/  # shared panels/selectors
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Board view | board/view.go | columns/cards rendering, navigation |
| Table view | table/view.go | row/column rendering |
| Settings panel | settings/settings.go | defaults editor + save flow |
| Shared UI | components/*.go | panels, selectors, empty state |
| UI tests | **/*_test.go | rendered output assertions |

## CONVENTIONS
- UI rendering is string-based; tests compare rendered output.
- Keep view-specific logic within its subdir (board/table/settings).

## ANTI-PATTERNS
- Avoid putting app state transitions here; keep those in `internal/app`.
- Avoid direct `gh` CLI or config access in UI; route through app/model.
