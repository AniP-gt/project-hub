---
name: large-file-refactor
description: Split oversized files into smaller, responsibility-focused files or directories. Use when a file exceeds a reasonable size or mixes concerns.
argument-hint: [file-or-dir-path]
allowed-tools: Read, Grep, Glob, Bash
user-invocable: true
---

# Large File Refactor

Refactor oversized files by splitting responsibilities into smaller files and, when appropriate, a dedicated directory. The goal is to keep each file focused, testable, and easy to maintain.

## Usage

/large-file-refactor internal/github/client.go

## Behavior

1. **Assess responsibilities**
   - Identify cohesive units (fetching, parsing, formatting, validation, UI rendering, etc.).
   - Separate side-effecting operations from pure transformations.

2. **Define a target layout**
   - Split by responsibility (e.g., `*_fetch.go`, `*_parse.go`, `*_logic.go`, `*_model.go`).
   - If multiple files form a clear sub-module, create a directory (e.g., `internal/github/parse/`).

3. **Move code safely**
   - Preserve public API behavior and signatures.
   - Keep package boundaries stable unless there is a clear shared use-case.
   - Avoid duplicating logic across files.

4. **Update references and tests**
   - Adjust imports and references after moving files.
   - Keep tests colocated with the code they verify.
   - Run existing tests and fix regressions.

## Checklist

- [ ] Each file has a single, clear responsibility.
- [ ] Large files are split into focused files or a dedicated directory.
- [ ] Public APIs remain consistent.
- [ ] Imports and references updated.
- [ ] Tests pass (`go test ./...`).

## Limitations

- This skill does not change behavior; it only restructures code.
- Avoid cross-package moves unless the code is clearly reusable.

## Related Files (project)

- `internal/github/client.go`
- `internal/github/project_fetch.go`
- `internal/github/parse/`
