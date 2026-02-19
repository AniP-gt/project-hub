# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-19
**Commit:** aaa99ff
**Branch:** main

## OVERVIEW
Go-based terminal UI (Bubble Tea) for GitHub Projects via `gh` CLI. Core logic in `internal/`, entrypoint in `cmd/project-hub`.

## STRUCTURE
```
project-hub/
├── cmd/project-hub/        # CLI entrypoint (main.go)
├── internal/app/           # App model/update loop, flows
├── internal/ui/            # Board/Table/Settings views + components
├── internal/github/        # gh CLI client + validators
├── internal/state/         # Domain types, filters, keymap
├── internal/config/        # Config load/save
├── docs/                   # Design doc
├── .specify/               # Spec/plan tooling scripts + templates
├── .opencode/              # Tooling metadata (includes node_modules)
├── build/                  # Local build artifacts
└── bin/                    # Local tool binaries
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| CLI flags/startup | cmd/project-hub/main.go | parses flags, loads config, starts app |
| App update loop | internal/app/app.go | Init/Update/View, message routing |
| Feature handlers | internal/app/update_*.go | edit/assign/status/filter/view switch |
| GitHub data | internal/github/client.go | `gh` CLI calls, fetch/update |
| Validation | internal/github/validator.go | input checks for fields/IDs |
| UI board | internal/ui/board/view.go | kanban view logic |
| UI table | internal/ui/table/view.go | table rendering |
| UI settings | internal/ui/settings/settings.go | config editing |
| UI components | internal/ui/components/*.go | panels/selectors |
| Types/state | internal/state/types.go | domain structs + enums |
| Filters | internal/state/filter.go | filter parsing/logic |
| Config | internal/config/config.go | config path/load/save |
| Tests | **/*_test.go | Go stdlib testing |

## CONVENTIONS
- Feature branches: `NNN-short-slug` (validated by `.specify/scripts/bash/common.sh`).
- Feature specs live under `specs/NNN-...` with `spec.md`, `plan.md`, `tasks.md`.
- `.specify/scripts/bash/create-new-feature.sh` generates branch + spec dir.
- No ESLint/Prettier/pyproject/Makefile configs present.

## ANTI-PATTERNS (THIS PROJECT)
- "DO NOT keep these sample items in the generated checklist file." — `.specify/templates/checklist-template.md`
- "DO NOT keep these sample tasks in the generated tasks.md file." — `.specify/templates/tasks-template.md`
- "UNDER NO CIRCUMSTANCES EVER CREATE ISSUES IN REPOSITORIES THAT DO NOT MATCH THE REMOTE URL" — `.opencode/command/speckit.taskstoissues.md`
- "DO NOT create any checklists that are embedded in the spec. That will be a separate command." — `.opencode/command/speckit.specify.md`
- "NEVER modify files (this is read-only analysis)" — `.opencode/command/speckit.analyze.md`
- "NEVER hallucinate missing sections (if absent, report them accurately)" — `.opencode/command/speckit.analyze.md`

## UNIQUE STYLES
- Bubble Tea model/update/view architecture; app logic split into `update_*.go` handlers.
- UI rendering is string-based; UI tests assert rendered output.

## COMMANDS
```bash
go install ./cmd/project-hub
go test ./...
go test ./internal/app
go test ./internal/ui/...
go test ./internal/github
```

## NOTES
- `go.mod` module path is `project-hub`; README notes it must be the repo import path for remote `go install` by tag.
- Config file path (README): macOS `~/Library/Application Support/project-hub/project-hub.json`, Linux `~/.config/project-hub/project-hub.json`, Windows `%APPDATA%\project-hub\project-hub.json`.
- Ignore `.opencode/node_modules`, `build/`, and `bin/` for code search unless troubleshooting tooling.
