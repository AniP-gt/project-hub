# Data Model

## Entities

### Project
- Attributes: id (external), name, views (board/table/roadmap availability), filters (supported fields), iterations/periods (for roadmap), updatedAt.
- Relationships: has many Items; has many Timelines/Iterations; owns ViewContext defaults.

### Item
- Attributes: id (external), title, description, status/column, assignees, labels, createdAt, updatedAt, due/iteration, position (within column), metadata (estimates/priority if提供されれば保持)。
- Relationships: belongs to Project; may belong to Timeline/Iteration; referenced by ViewContext focus.
- State transitions: status moves left/right; title/description editable; assignees add/remove; filters include/exclude based on labels/status/assignee.

### ViewContext
- Attributes: currentView (board/table/roadmap), focusedItemId, filters (label/assignee/status query), mode (normal/filtering/edit), sort preferences for table.
- Relationships: references Project and Items; shared across views to keep focus continuity.
- State transitions: view switch (board↔table↔roadmap), mode switch (normal/filter/edit), filter apply/clear, focus move (up/down/left/right with view-aware rules).

### Timeline / Iteration
- Attributes: id/name, start, end, progress summary, lane grouping (if multiple iterations displayed).
- Relationships: contains Items scheduled in the period; belongs to Project.

## Validation & Rules
- Status changes must map to valid adjacent columns for the current project.
- Filter query must parse known fields (label/assignee/status) and gracefully ignore unknown tokens.
- Edits (title/description) require non-empty title; cancel restores previous values.
- Assignment requires selecting from available assignees; updates must reflect immediately in view state.
- Roadmap placement must respect item period; items without period fall into “unscheduled” lane.

## Data Sources
- Primary: gh CLI `project view ... --json ...` outputs for initial load; subsequent updates via gh commands per action.
- In-memory state only; no persistent storage beyond runtime.
