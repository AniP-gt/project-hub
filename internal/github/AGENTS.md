# INTERNAL/GITHUB KNOWLEDGE BASE

## OVERVIEW
GitHub CLI integration and validation for Projects V2 data.

## STRUCTURE
```
internal/github/
├── client.go           # All gh CLI calls (CLIClient implements Client interface)
├── validator.go        # ID/field/assignee validation
├── validator_test.go   # Validator tests
├── client_test.go      # Client tests
├── project_fetch.go    # Project fetch helpers
└── parse/              # JSON parsing helpers
    ├── parse_item_map.go
    ├── parse_item_list.go
    ├── parse_fields.go
    └── parse_values.go
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| gh CLI calls | client.go | fetch/update project items, 296 lines |
| Client interface | client.go | Defines all GitHub operations |
| CLIClient | client.go | Implements Client using `gh` CLI |
| Input validation | validator.go | IDs/fields/assignees checks |
| Project fetch | project_fetch.go | FetchProject implementation |
| Item parsing | parse/parse_item_map.go | Parse single item from JSON |
| List parsing | parse/parse_item_list.go | Parse item list from JSON |
| Field parsing | parse/parse_fields.go | Parse project fields |
| Value parsing | parse/parse_values.go | Parse field values |
| Tests | *_test.go | JSON parsing and validator coverage |

## CONVENTIONS
- Use `github.Client` interface methods for side effects; all implementations in `client.go`.
- Validate inputs with `validator.go` helpers before updates.
- Parsing logic lives in `parse/` sub-package; called by `client.go`.

## ANTI-PATTERNS
- Avoid calling `gh` directly from other packages; centralize in `client.go`.
- Avoid duplicating validation logic outside this package.
