# 📝 GitHub Projects TUI Management Tool — Design Document

## 1. Overview

| Item             | Details                                                                                                                                                                                                                                      |
| :--------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Project Name** | GitHub Projects TUI Management Tool                                                                                                                                                                                                          |
| **Purpose**      | A TUI (Text-based User Interface) application for operating GitHub Projects (Board, Table, Roadmap) directly in the terminal. It eliminates mouse operations and enables **fast, efficient project management** with Vim-style key bindings. |
| **Target Users** | Developers, engineers, and product managers who use GitHub Projects.                                                                                                                                                                         |

---

## 2. Scope & Features

### 2.1. Required Features (MVP: Minimum Viable Product)

1. **Multi-view support:** The following three views should be supported and switchable via keyboard:
   - **Board View**
   - **Table View**
   - **Roadmap View**

2. **Vim-style operation:** All operations are controlled using keyboard shortcuts following Vim conventions such as `j`, `k`, `h`, `l`.
3. **Fast filtering:** Quickly filter boards and tables by labels, assignees, or issue states (triggered with `/`).
4. **Issue operations:** Change status (move between columns) and assign assignees to the selected issue.

### 2.2. Killer Features

- **Progress visualization:** Display iteration-based task placement and progress using text-based graphs in the Roadmap view.
- **High-speed inline editing:** Edit titles and descriptions of the selected issue inline, similar to Vim insert mode.

### 2.3. Iteration Filter CLI (Planned)

- **Goal:** Allow `project-hub` to launch with iteration-scoped data using a new `--iteration` flag that accepts multiple values (e.g., `current next previous`).
- **Accepted inputs:** Literal iteration titles (e.g., `Sprint 28`) and shorthand keywords with or without `@` prefixes (`current`, `@current`, `next`, etc.).
- **Resolution flow:**
  1. CLI parses the flag values in `cmd/project-hub/main.go` and stores them in the initial `FilterState.Iterations` so the UI applies them consistently.
  2. `internal/github/client.go` always fetches the full item list, but augments each item with iteration metadata (ID, title, start date, duration) extracted from the JSON payload.
  3. Keywords resolve locally by comparing `time.Now()` to the iteration window (`@current` matches items whose iteration is in progress, `@next` matches iterations with a start date in the future, `@previous` matches iterations that ended in the past).
  4. Literal values (e.g., `Sprint 12`) are matched case-insensitively against the iteration title or ID.
- **UI impact:** All views (Board/Table/Roadmap) evaluate the same in-memory filters, so reloading or switching views preserves the CLI-provided iteration context.

---

## 3. Technology Stack

| Category             | Technology                    | Role                                                                                         |
| :------------------- | :---------------------------- | :------------------------------------------------------------------------------------------- |
| **Programming Lang** | **Go (Golang)**               | Core application logic, fast processing, single binary distribution.                         |
| **TUI Framework**    | **Bubbletea**                 | TUI rendering and state management (Elm architecture).                                       |
| **TUI Styling**      | **Lipgloss** (planned)        | Layout, colors, styling of TUI components.                                                   |
| **Data Integration** | **`gh` command (GitHub CLI)** | **Data retrieval and authentication handling**. The app invokes `gh` and parses JSON output. |
| **JSON Parsing**     | `encoding/json`               | Map JSON output from `gh` into Go structs.                                                   |

---

## 4. Architecture & Data Flow

### 4.1. Architecture Pattern

- **Bubbletea Model-Update-View (MUV):** The application's state is centrally managed. User inputs update the state (`Model`), and the resulting UI is rendered (`View`).

### 4.2. Data Retrieval Flow

1. **Authentication:** The user must run `gh auth login` beforehand (the TUI tool does not implement authentication).
2. **Command execution:** When the TUI launches, it uses Go’s `os/exec` to run commands such as
   **`gh project view [project-id] --json`**.
3. **JSON retrieval:** The `gh` command outputs project data (issues, fields, columns) in JSON format.
4. **Parsing & storage:** JSON is parsed into common Go structs and stored in the root Bubbletea `Model`.

---

## 5. User Interface (UI) Design

### 5.1. Layout

The UI consists of three sections:

- **Header:** Project name, current view mode, active filters.
- **Footer:** Current mode, hints for essential Vim-style shortcuts.
- **Main View:** Switchable between Board, Table, and Roadmap.

### 5.2. Multi-view Details

| View Mode      | Purpose                       | Key Characteristics                                                                                                                   |
| :------------- | :---------------------------- | :------------------------------------------------------------------------------------------------------------------------------------ |
| **1. Board**   | Status-based flow management  | Issues placed in columns by status. Cards display **assignees and labels** concisely.                                                 |
| **2. Table**   | Sorting & comparing details   | Issues displayed as **rows and columns** with all fields (created/updated dates, estimates, etc.). Future plan: customizable columns. |
| **3. Roadmap** | Iteration & timeline tracking | Tasks positioned on a **timeline** (sprints, months) similar to a text-based Gantt chart.                                             |

---

### 5.3 Mock

- `docs/moc/github-project-hub-mock.tsx`
- Although the mock is written in TSX, it will be implemented in Go.

---

## 6. Interaction Design (Vim Key Bindings)

### 6.1. Modes

- **Normal Mode (default):** Movement, actions, switching views.
- **Filtering Mode:** Enter search queries (triggered with `/`).
- **Editing Mode:** Text input (triggered with `i` or `Enter`).

### 6.2. Main Shortcuts (Normal Mode)

| Category          | Action                    | Key                                  |
| :---------------- | :------------------------ | :----------------------------------- |
| **Navigation**    | Move up/down              | `k` / `j`                            |
| **Status Change** | Move to left/right column | `h` / `l`                            |
| **View Switch**   | Board / Table / Roadmap / Settings | `1` or `b` / `2` or `t` / `3` or `r` / `4` |
| **Editing**       | Enter editing mode        | `i` / `Enter`                        |
| **Filter**        | Enter filtering mode      | `/`                                  |
| **Assign**        | Assign assignee           | `a`                                  |
| **Quit**          | Exit application          | `q` / `Ctrl+c`                       |

---

### 6.3. Startup Defaults (Project/Owner)

- Users can open **Settings** (key `4`) and save default `project` and `owner` values.
- Saved config path (Linux): `~/.config/project-hub/project-hub.json`.
- Precedence rule at startup: **CLI flags override saved config**.
- If config is malformed, the app prints a warning and continues.
