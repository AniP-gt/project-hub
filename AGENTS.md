# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-19
**Commit:** f5614c9
**Branch:** main

## OVERVIEW
Go-based terminal UI (Bubble Tea) for GitHub Projects via `gh` CLI. Core logic in `internal/`, entrypoint in `cmd/project-hub`.

## STRUCTURE
```
project-hub/
├── cmd/project-hub/        # CLI entrypoint (main.go, 250 lines)
├── internal/app/           # App model/update loop, flows
│   ├── update/             # Feature handlers (edit, assign, status, filter, view)
│   └── core/               # Commands, messages, metrics, helpers
├── internal/ui/            # Board/Table/Settings views + components
├── internal/github/        # gh CLI client + validators
│   └── parse/              # JSON parsing helpers
├── internal/state/         # Domain types, filters, keymap
├── internal/config/        # Config load/save
├── docs/                   # Design doc
├── .specify/               # Spec/plan tooling scripts + templates
├── .opencode/              # Tooling metadata
├── build/                  # Local build artifacts
└── bin/                    # Local tool binaries
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| CLI flags/startup | cmd/project-hub/main.go | parses flags, loads config, starts app |
| App lifecycle | internal/app/app.go | Init/Update/View, wrapper around update.State |
| Feature handlers | internal/app/update/*.go | edit/assign/status/filter/view switch |
| GitHub data | internal/github/client.go | `gh` CLI calls, fetch/update |
| Validation | internal/github/validator.go | input checks for fields/IDs |
| UI board | internal/ui/board/*.go | model, view, logic, render, layout |
| UI table | internal/ui/table/view.go | table rendering |
| UI settings | internal/ui/settings/settings.go | config editing |
| UI components | internal/ui/components/*.go | panels, selectors, empty state |
| Types/state | internal/state/types.go | domain structs + enums (Item, Project, FilterState) |
| Filters | internal/state/filter.go | filter parsing/logic |
| Config | internal/config/config.go | config path/load/save |
| Tests | **/*_test.go | Go stdlib testing, table-driven |

## CONVENTIONS
- **Feature branches:** `NNN-short-slug` (validated by `.specify/scripts/bash/common.sh`).
- **Feature specs:** live under `specs/NNN-...` with `spec.md`, `plan.md`, `tasks.md`.
- **Tests:** table-driven, stdlib `testing` only, no third-party test libs.
- **UI tests:** assert rendered string output with `strings.Contains()`.
- **Import grouping:** stdlib → external (charmbracelet) → internal, blank line separators.
- **Module path:** `project-hub` (local only; must change to `github.com/AniP-gt/project-hub` for remote install).

## ANTI-PATTERNS (THIS PROJECT)
- "DO NOT keep these sample items in the generated checklist file." — `.specify/templates/checklist-template.md`
- "DO NOT keep these sample tasks in the generated tasks.md file." — `.specify/templates/tasks-template.md`
- "UNDER NO CIRCUMSTANCES EVER CREATE ISSUES IN REPOSITORIES THAT DO NOT MATCH THE REMOTE URL" — `.opencode/command/speckit.taskstoissues.md`
- "DO NOT create any checklists that are embedded in the spec. That will be a separate command." — `.opencode/command/speckit.specify.md`
- "NEVER modify files (this is read-only analysis)" — `.opencode/command/speckit.analyze.md`
- "NEVER hallucinate missing sections (if absent, report them accurately)" — `.opencode/command/speckit.analyze.md`

## UNIQUE STYLES
- Bubble Tea model/update/view architecture; app logic split into `update/*.go` handlers.
- `internal/app/app.go` wraps `update.State` with conversion helpers (toUpdateState/applyUpdateState/fromUpdateState).
- UI rendering is string-based; UI tests assert rendered output.
- `main.go` contains sample data fallback (items "1", "2", "3") loaded before real gh fetch.

## COMMANDS
```bash
go install ./cmd/project-hub
go test ./...
go test ./internal/app
go test ./internal/ui/...
go test ./internal/github
```

## NOTES
- `go.mod` module path is `project-hub`; README notes it must be `github.com/AniP-gt/project-hub` for remote `go install`.
- Config file path: macOS `~/Library/Application Support/project-hub/project-hub.json`, Linux `~/.config/project-hub/project-hub.json`, Windows `%APPDATA%\project-hub\project-hub.json`.
- Ignore `.opencode/node_modules`, `build/`, `bin/` for code search unless troubleshooting tooling.
- 56 Go files, ~9,324 lines of code, 18 directories.
