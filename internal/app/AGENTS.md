# INTERNAL/APP KNOWLEDGE BASE

## OVERVIEW
Core application state and update loop (Bubble Tea model/update/view).

## STRUCTURE
```
internal/app/
├── app.go              # App struct, Init/Update/View wrapper
├── view.go             # View() rendering (455 lines)
├── update/             # Feature handlers
│   ├── update.go       # Main dispatch
│   ├── state.go        # State struct for handlers
│   ├── edit.go         # Edit mode handlers
│   ├── assignee.go     # Assignee update handlers
│   ├── status.go       # Status change handlers
│   ├── view_switch.go  # View switching handlers
│   ├── view.go         # View-specific updates
│   ├── filter.go       # Filter input handlers
│   ├── key.go          # Key handling
│   └── settings.go     # Settings handlers
├── core/               # Shared infrastructure
│   ├── commands.go     # tea.Cmd helpers
│   ├── messages.go     # Msg types
│   ├── helpers.go      # Utility functions
│   ├── metrics.go      # Timing/telemetry
│   └── constants.go    # Constants
└── *_test.go           # Flow-level tests
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| App lifecycle | app.go | Init/Update/View wrapper, state conversion |
| View rendering | view.go | Main View() function (455 lines) |
| Update dispatch | update/update.go | Message routing to handlers |
| State struct | update/state.go | Unified State for all handlers |
| Edit flow | update/edit.go | item edit mode + persistence |
| Assign flow | update/assignee.go | assignee updates + commands |
| Status changes | update/status.go | column/status transitions |
| View switching | update/view_switch.go | board/table/settings transitions |
| Filter input | update/filter.go | filter input handling |
| Key handling | update/key.go | keyboard shortcuts |
| Settings | update/settings.go | settings panel handlers |
| Commands | core/commands.go | tea.Cmd factories |
| Messages | core/messages.go | Msg type definitions |
| Helpers | core/helpers.go | utility functions |
| Metrics | core/metrics.go | timing/telemetry helpers |
| Tests | *_test.go | flow-level assertions |

## CONVENTIONS
- Update handlers are split by feature in `update/` package; keep new behaviors there rather than growing `app.go`.
- `app.go` wraps `update.State` with conversion helpers: `toUpdateState()`, `applyUpdateState()`, `fromUpdateState()`.
- App updates should emit commands rather than mutating external state directly; keep side effects in `github.Client` calls.
- `view.go` (455 lines) contains View() plus sorting and grouped-table rendering logic.

## ANTI-PATTERNS
- Avoid placing UI rendering code here; keep view rendering in `internal/ui`.
- Avoid adding direct `gh` CLI calls here; use `internal/github.Client`.

## NOTES
- Most user-facing behavior changes touch both `app.go` and one or more `update/*.go` handlers.
- `update.State` holds Model + UI components (BoardModel, TextInput, StatusSelector, etc.) for handler access.
