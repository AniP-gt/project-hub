# INTERNAL/APP KNOWLEDGE BASE

## OVERVIEW
Core application state and update loop (Bubble Tea model/update/view).

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| App lifecycle | app.go | Init/Update/View and main model wiring |
| Assign flow | update_assignee.go | assignee updates + commands |
| Status changes | update_status.go | column/status transitions |
| Editing flow | update_edit.go | item edit mode + persistence |
| View switching | update_view_switch.go | board/table/settings transitions |
| Filters | update_filter.go | filter input handling |
| Metrics | metrics.go | timing/telemetry helpers |
| Tests | *_test.go | flow-level assertions |

## CONVENTIONS
- Update handlers are split by feature: keep new behaviors in `update_*.go` rather than growing `app.go`.
- App updates should emit commands rather than mutating external state directly; keep side effects in `github.Client` calls.

## ANTI-PATTERNS
- Avoid placing UI rendering code here; keep view rendering in `internal/ui`.
- Avoid adding direct `gh` CLI calls here; use `internal/github.Client`.

## NOTES
- Most user-facing behavior changes touch both `app.go` and one or more `update_*.go` handlers.
