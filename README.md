# project-hub

<img width="1669" height="936" alt="Board" src="https://github.com/user-attachments/assets/c48a75c9-eeb9-4fbb-bb07-af1fe1677b48" />


`project-hub` is a TUI for browsing GitHub Projects via the `gh` CLI.

## Requirements

- Go (for building/installing the CLI)
- GitHub CLI (`gh`) installed and authenticated (`gh auth login`)
- Access to a GitHub Project (Project V2)

## Installation

Local install (from repository root):

```bash
git clone https://github.com/AniP-gt/project-hub.git
go install ./cmd/project-hub
```

This places the binary in `$GOBIN` if set, otherwise `$(go env GOPATH)/bin`. Make sure that directory is in your `PATH`. Example:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Remote install (recommended for users after you push to GitHub and tag a release):

1. Ensure `module` in go.mod is set to the repo import path, e.g. `module github.com/AniP-gt/project-hub`.
2. Create and push a tag, e.g. `git tag v0.1.0 && git push origin v0.1.0`.

Then users can run:

```bash
go install github.com/AniP-gt/project-hub/cmd/project-hub@v0.1.0
```

## Usage

```bash
project-hub --project <project-id-or-url> [--owner <org-or-user>] [options]
```

`--project` is required unless you set a default in the config file.

### Options

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `--project` / `-p` | Yes | — | GitHub Project ID or URL (shorthand: `-p`) |
| `--owner` / `-o` | No | — | Owner (org/user) for the project; can be inferred from a project URL (shorthand: `-o`) |
| `--gh-path` / `-g` | No | `gh` | Path to the `gh` CLI executable (shorthand: `-g`) |
| `--item-limit` / `-il` | No | `100` | Maximum number of items to fetch (shorthand: `-il`) |
| `--iteration` / `-i` | No | — | Iteration filters (repeat the flag or pass values after it). Shorthand: `-i` behaves like `--iteration` and can be repeated.

## Features & Operations

### Views & Navigation

| Action | Keys | Notes |
| --- | --- | --- |
| Board view | `1` / `b` | Switch to kanban board |
| Table view | `2` / `t` | Switch to table view |
| Settings | `3` | Open settings panel |
| Move focus | `h` / `l` / `k` / `j` | Left / right / up / down |
| Reload | `R` / `Ctrl+r` | Refresh items |

### Item Actions

| Action | Keys | Notes |
| --- | --- | --- |
| Edit title | `i` / `Enter` | Type → `Enter` to save, `Esc` to cancel |
| Assign | `a` | Type assignee → `Enter` to save, `Esc` to cancel |
| Detail panel | `o` | `j/k` to scroll, `Esc`/`q` to close |
| Status select | `w` | `j/k` to move, `Enter` to confirm, `Esc` to cancel |
| Open in browser | `O` | Uses system opener (macOS: `open`, Linux: `xdg-open`, Windows: `rundll32`). If unavailable, shows the URL in a notification. |
| Copy URL | `y` | Uses clipboard tool (macOS: `pbcopy`, Windows: `clip`, Linux: `wl-copy` or `xclip`). If unavailable, shows the URL in a notification. |

 (Board-specific and Table-specific shortcuts moved to their respective sections below.)

### Filter Mode

| Action | Keys | Notes |
| --- | --- | --- |
| Enter filter | `/` | Type filters → `Enter` to apply, `Esc` to clear |

Supported tokens:

| Token | Description | Notes |
| --- | --- | --- |
| `label:` / `labels:` | Filter by labels | Comma/semicolon-separated values |
| `assignee:` / `assignees:` | Filter by assignees | Multiple values supported |
| `status:` | Filter by status | Use quotes for spaces |
| `iteration:` | Filter by iteration | Shorthand tokens below |
| `group:` / `group-by:` / `groupby:` | Set table grouping | `status`, `assignee`, `iteration` |
| `FieldName:Value` | Any project field | Use quotes for spaces |

Shorthand iteration tokens: `@current`, `@next`, `@previous`, `current`, `next`, `previous`

Examples:

- `status:"In Progress" assignee:alice`
- `iteration:@current,@previous,@next`
- `@current next previous`
- `Sprint:Q1`
- `"Iteration Name":"Q1 Sprint"`

## Board

<img width="1669" height="936" alt="Board" src="https://github.com/user-attachments/assets/89880a68-1c20-45c7-a647-a0ea9d3d0ac7" />

The default view is a kanban-style board. Use the shortcuts in **Features & Operations** above for navigation, item actions, and card field toggles.

### Board shortcuts

| Action | Keys | Notes |
| --- | --- | --- |
| Move between cards | `j` / `k` | Navigate up/down between cards |
| Switch views | `1` / `2` / `3` | Board / Table / Settings |
| Open filter input | `/` | Press `Enter` to apply, `Esc` to clear |
| Open item detail panel | `o` | `j/k` to scroll, `Esc`/`q` to close |
| Edit focused item | `i` | `Enter` to save, `Esc` to cancel |
| Change status | `w` | Select status option with `j/k`, `Enter` to confirm |
| Toggle card fields (field toggle mode) | `f` | In field toggle mode: `m` Milestone, `r` Repository, `l` Labels, `s` Sub-issues, `p` Parent; `Esc` to exit |

## Table Filter

<img width="1674" height="927" alt="Table" src="https://github.com/user-attachments/assets/c28fd58e-f326-4bcd-8cf8-270f8d6ce86c" />


Press `/` to enter Filter Mode in the TUI. The footer shows `FILTER MODE <input>` while you type. Press `Enter` to apply or `Esc` to clear.

### Table shortcuts

| Action | Keys | Notes |
| --- | --- | --- |
| Sort mode | `s` | Then press: `t` Title, `s` Status, `r` Repository, `l` Labels, `m` Milestone, `p` Priority, `n` Number, `c` CreatedAt, `u` UpdatedAt. Press `Esc` to cancel |
| Jump to top/bottom | `g` / `G` | Move focus to the first/last row |
| Group toggle | `m` | Cycles `status → assignee → iteration → none` |

Iteration semantics:

- `@current` matches iterations where **start ≤ now < end** (end = start + duration days)
- `@next` matches iterations with a start date in the future
- `@previous` matches iterations that have ended (now ≥ end)
- Literal values match iteration **name or ID** (case-insensitive)

## Examples

Use a project ID:

```bash
project-hub --project 12345
```

Use a project URL (owner can be inferred from the URL path):

```bash
project-hub --project https://github.com/acme-org/projects/7
```

Provide an explicit owner:

```bash
project-hub --project 12345 --owner acme-org
```

Limit items fetched:

```bash
project-hub --project 12345 --item-limit 50
```

Filter by iteration (repeat the flag):

```bash
project-hub --project 12345 --iteration @current --iteration "Iteration 1"
```

Filter by iteration (values after the flag):

```bash
project-hub --project 12345 --iteration @current "Iteration 1" "Iteration 2"
```

## Configuration

`project-hub` reads a JSON config file for defaults. CLI flags take precedence over config values.

Config file path:

- macOS: `~/Library/Application Support/project-hub/project-hub.json`
- Linux: `~/.config/project-hub/project-hub.json`
- Windows: `%APPDATA%\project-hub\project-hub.json`

Example config:

```json
{
  "defaultProjectID": "12345",
  "defaultOwner": "acme-org"
}
```

## Settings

<img width="1671" height="930" alt="Settings" src="https://github.com/user-attachments/assets/cd84a94d-ce9f-442a-8d33-d5db69986221" />


Press `3` in the TUI to open the Settings panel. Here you can view and manage your defaults:

- **Default Project**: The default project ID saved in your config file
- **Default Owner**: The default owner (org/user) saved in your config file

Changes made in Settings are persisted to `project-hub.json`. Note that **CLI overrides config**—if you pass `--project` or `--owner` on the command line, those values take precedence.

### Configuration Errors

If the config file cannot be loaded, you will see a warning message:

```
warning: failed to load config: <error details>
```

The application will continue to run normally; you can still use the CLI without a config file by providing values via command-line flags.

## Notes

- The binary name is the last path element in the cmd directory (`cmd/project-hub` -> `project-hub`).
- If `--project` is missing and no default is configured, the CLI prints usage and exits.
- Iteration filters accept `@current`, `@next`, `@previous`, or explicit iteration names. The `iteration:` prefix is also accepted (e.g. `iteration:@current`).
