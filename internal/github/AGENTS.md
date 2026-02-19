# INTERNAL/GITHUB KNOWLEDGE BASE

## OVERVIEW
GitHub CLI integration and validation for Projects V2 data.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| gh CLI calls | client.go | fetch/update project items |
| Input validation | validator.go | IDs/fields/assignees checks |
| Tests | *_test.go | JSON parsing and validator coverage |

## CONVENTIONS
- Use `github.Client` methods for side effects; keep parsing inside this package.
- Validate inputs with `validator.go` helpers before updates.

## ANTI-PATTERNS
- Avoid calling `gh` directly from other packages; centralize in `client.go`.
- Avoid duplicating validation logic outside this package.
