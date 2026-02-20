# project-hub

<img width="1669" height="936" alt="Board" src="https://github.com/user-attachments/assets/c48a75c9-eeb9-4fbb-bb07-af1fe1677b48" />

`project-hub` is a terminal UI for browsing and updating GitHub Projects (Project V2) through the `gh` CLI.

## Quick Start

### 1) Requirements

- Go (for build/install)
- GitHub CLI (`gh`) installed and authenticated (`gh auth login`)
- Access to a GitHub Project (Project V2)

### 2) Install

```bash
git clone https://github.com/AniP-gt/project-hub.git
go install ./cmd/project-hub
```

Binary location is `$GOBIN` (if set) or `$(go env GOPATH)/bin`. Ensure it is in `PATH`.

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### 3) Start the app

```bash
project-hub --project <project-id-or-url> [options]
```

If `--project` is omitted, `defaultProjectID` from config is used. If neither exists, the app exits with usage help.

## Startup Commands (common patterns)

Use a project ID:

```bash
project-hub --project 12345
```

Use a project URL (owner inferred from URL path):

```bash
project-hub --project https://github.com/acme-org/projects/7
```

Provide explicit owner:

```bash
project-hub --project 12345 --owner acme-org
```

Limit fetched items:

```bash
project-hub --project 12345 --item-limit 50
```

Filter by iteration (repeat flag):

```bash
project-hub --project 12345 --iteration @current --iteration "Iteration 1"
```

Filter by iteration (multiple values after one flag):

```bash
project-hub --project 12345 --iteration @current "Iteration 1" "Iteration 2"
```

## CLI Options

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `--project`, `-p` | Yes* | — | Project ID or Project URL |
| `--owner`, `-o` | No | inferred/none | Owner (`org` or `user`). Inferred when `--project` is a URL |
| `--gh-path`, `-g` | No | `gh` | Path to GitHub CLI executable |
| `--item-limit`, `-il` | No | `100` | Maximum number of items to fetch |
| `--iteration`, `-i` | No | none | Iteration filters. Repeat flag and/or pass multiple values |

\* `--project` is only optional when `defaultProjectID` exists in config.

## Feature Guide (by function)

### Global navigation and item actions

| Function | Keys | Behavior |
| --- | --- | --- |
| Switch to Board | `1` / `b` | Kanban view |
| Switch to Table | `2` / `t` | Table view |
| Open Settings | `3` | Settings panel |
| Move focus | `h` / `l` / `k` / `j` | Left / right / up / down |
| Reload items | `R` / `Ctrl+r` | Refresh project data |
| Edit title | `i` / `Enter` | `Enter` to save, `Esc` to cancel |
| Assign user | `a` | Type assignee, `Enter` save, `Esc` cancel |
| Open detail panel | `o` | `j/k` scroll, `Esc`/`q` close |
| Change status | `w` | `j/k` select, `Enter` confirm, `Esc` cancel |
| Open in browser | `O` | Uses OS opener; fallback is URL notification |
| Copy URL | `y` | Uses clipboard command; fallback is URL notification |

### Board view

<img width="1669" height="936" alt="Board" src="https://github.com/user-attachments/assets/89880a68-1c20-45c7-a647-a0ea9d3d0ac7" />

Default view is a kanban-style board.

| Function | Keys | Behavior |
| --- | --- | --- |
| Move between cards | `j` / `k` | Navigate focused card |
| Open filter input | `/` | `Enter` apply, `Esc` clear |
| Toggle card fields | `f` | In toggle mode: `m` Milestone, `r` Repository, `l` Labels, `s` Sub-issues, `p` Parent, `Esc` exit |

### Table view

<img width="1674" height="927" alt="Table" src="https://github.com/user-attachments/assets/c28fd58e-f326-4bcd-8cf8-270f8d6ce86c" />

| Function | Keys | Behavior |
| --- | --- | --- |
| Sort mode | `s` | Then: `t` Title, `s` Status, `r` Repository, `l` Labels, `m` Milestone, `p` Priority, `n` Number, `c` CreatedAt, `u` UpdatedAt |
| Jump top/bottom | `g` / `G` | First/last row |
| Group toggle | `m` | `status -> assignee -> iteration -> none` |

### Filter mode

Press `/` in Board/Table. Footer shows `FILTER MODE <input>` while typing. `Enter` applies filters, `Esc` clears.

Supported tokens:

| Token | Meaning | Notes |
| --- | --- | --- |
| `label:`, `labels:` | Label filter | Comma or semicolon separated |
| `assignee:`, `assignees:` | Assignee filter | Multiple values supported |
| `status:` | Status filter | Quote if it contains spaces |
| `iteration:` | Iteration filter | Supports shorthand tokens |
| `group:`, `group-by:`, `groupby:` | Table grouping | `status`, `assignee`, `iteration` |
| `FieldName:Value` | Any project field | Quote field/value with spaces |

Iteration shorthand tokens: `@current`, `@next`, `@previous`, `current`, `next`, `previous`

Examples:

- `status:"In Progress" assignee:alice`
- `iteration:@current,@previous,@next`
- `@current next previous`
- `Sprint:Q1`
- `"Iteration Name":"Q1 Sprint"`

Iteration semantics:

- `@current`: start <= now < end (end = start + duration days)
- `@next`: iteration start is in the future
- `@previous`: now >= end
- Literal value: matches iteration name or ID (case-insensitive)

### Settings view

<img width="1671" height="930" alt="Settings" src="https://github.com/user-attachments/assets/cd84a94d-ce9f-442a-8d33-d5db69986221" />

Press `3` to open Settings.

- Default Project: saved as `defaultProjectID`
- Default Owner: saved as `defaultOwner`

Settings changes are persisted to config. CLI options always override config values.

## Configuration

`project-hub` reads JSON config for defaults.

Config path by OS:

- macOS: `~/Library/Application Support/project-hub/project-hub.json`
- Linux: `~/.config/project-hub/project-hub.json`
- Windows: `%APPDATA%\project-hub\project-hub.json`

Example:

```json
{
  "defaultProjectID": "12345",
  "defaultOwner": "acme-org"
}
```

If config loading fails, warning is shown and app continues:

```text
warning: failed to load config: <error details>
```

## Optional: Remote install for released versions

If publishing tags for users:

1. Set `module` in `go.mod` to the repo path (for example, `github.com/AniP-gt/project-hub`).
2. Push a tag (for example, `git tag v0.1.0 && git push origin v0.1.0`).

Then users can install by version:

```bash
go install github.com/AniP-gt/project-hub/cmd/project-hub@v0.1.0
```
