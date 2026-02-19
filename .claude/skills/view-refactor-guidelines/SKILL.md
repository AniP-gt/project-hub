---
name: view-refactor-guidelines
description: Write view files with render-only focus, then refactor to separate non-UI logic. Use when editing internal/ui/* view files.
argument-hint: [view-path]
allowed-tools: Read, Grep, Glob, Bash
user-invocable: true
---

# View Authoring + Refactor Guidelines

Write view files with render-only focus first, then refactor non-UI logic into dedicated files so responsibilities are clear and testable.

## Usage

/view-refactor-guidelines internal/ui/board/view.go

## Behavior

1. **Write view files with clear responsibilities**
   - Keep `View()` limited to rendering: compose strings/styles from already-prepared data.
   - Keep `Update()` limited to UI state transitions (focus, cursor, scroll, key handling).
   - Avoid domain rules (filtering/grouping/inference) in `View()` and `Update()`.

2. **Identify responsibilities for refactor**
   - Rendering: `View()` and rendering helpers that assemble strings/styles.
   - UI input handling: `Update()` and focus/scroll state transitions.
   - Non-UI logic: filtering, grouping, ordering, inference, and data transforms.

3. **File-level separation (recommended layout)**
   - `view.go`: `Init`, `Update`, `View` only (or as close as possible).
   - `render.go`: rendering helpers (string building, truncation, formatting).
   - `layout.go`: sizing/layout calculations (width/height, visible rows/columns).
   - `model.go`: model structs and constructors.
   - `logic.go`: non-UI logic (filtering, grouping, ordering, inference).

4. **Keep view files “render-only”**
   - `View()` should compose already-prepared data.
   - Avoid direct data filtering/grouping inside `View()`.
   - Avoid domain rules (e.g., status order, priority inference) inside `view.go`.

5. **Move non-UI logic out**
   - Filtering helpers (e.g., `applyFilter`, `matchesFieldFilters`) → `logic.go`.
   - Grouping helpers (e.g., `groupItemsByStatus`, `GroupItemsByAssignee`) → `logic.go`.
   - Ordering constants (e.g., `ColumnOrder`) → `logic.go`.

6. **Prefer package-local helpers over cross-package coupling**
   - If the logic is view-specific, keep it under `internal/ui/<view>/`.
   - If it is shared across views, move it to `internal/ui/components` or `internal/state`.

7. **Tests stay with behavior**
   - Rendering behavior → `internal/ui/.../*_test.go`.
   - Logic behavior → `internal/ui/.../logic.go` tests or `internal/state` tests.

## Checklist (before/after refactor)

- [ ] `view.go` only contains `Init/Update/View` and minimal orchestration.
- [ ] Rendering helpers are in `render.go`.
- [ ] Layout calculations are in `layout.go`.
- [ ] Non-UI logic moved to `logic.go` or `internal/state`.
- [ ] Tests updated and passing (`go test ./...`).

## Limitations

- This skill does not change application behavior; it only restructures code.
- Do not move logic across packages unless a shared use-case is confirmed.

## Related Files (project)

- `internal/ui/board/view.go`
- `internal/ui/board/render.go`
- `internal/ui/board/layout.go`
- `internal/ui/board/model.go`
- `internal/ui/board/logic.go`
