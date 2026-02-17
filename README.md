# project-hub

`project-hub` is a terminal UI for browsing GitHub Projects via the `gh` CLI.

## Requirements

- Go (for building/installing the CLI)
- GitHub CLI (`gh`) installed and authenticated (`gh auth login`)
- Access to a GitHub Project (Project V2)

## Installation

Local install (from repository root):

```bash
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

## Options

- `--project` (required): GitHub Project ID or URL
- `--owner`: Owner (org/user) for the project (can be inferred from a project URL)
- `--gh-path`: Path to the `gh` CLI executable (default: `gh`)
- `--item-limit`: Maximum number of items to fetch (default: 100)
- `--iteration`: Iteration filters (repeat the flag or pass values after it)

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

- macOS: `~/Library/Application Support/project-hub/projects-tui.json`
- Linux: `~/.config/project-hub/projects-tui.json`
- Windows: `%APPDATA%\project-hub\projects-tui.json`

Example config:

```json
{
  "defaultProjectID": "12345",
  "defaultOwner": "acme-org"
}
```

## Notes

- The binary name is the last path element in the cmd directory (`cmd/project-hub` -> `project-hub`).
- If `--project` is missing and no default is configured, the CLI prints usage and exits.
- Iteration filters accept `@current`, `@next`, `@previous`, or explicit iteration names. The `iteration:` prefix is also accepted (e.g. `iteration:@current`).
