---
status: deprecated
epic: taskflow-v1-core
effort: 4-6 hours
tier: 2
priority: high
project: taskflow-bootstrap
tags: [cli, go, ux, bubble-tea]
deprecated_at: 2026-06-07
deprecated_reason: Superseded by the scoped tskflwctl design (epic 17-pm-go-cli); old over-ambitious taskflow vision (docker/api/semantic/AI-gen) retired 2026-06-07. Research mined + kept.
updated_at: 2026-06-07
id: 6f9yr8m02jh2
---

# TaskFlow 09: Interactive 'Project Start' Workflow

**Goal**: Implement the "Wizard" experience for starting projects and curating tasks using the new Go CLI.

## Context
We want a guided, interactive workflow that feels like a modern CLI application. This helps humans efficiently bundle existing tasks into a "Project" (sprint/initiative) and define prompts for AI agents to generate new tasks.

## Desired Workflow
```text
> taskflow project start

? What's the name of the new project?
> Chart Annotations

? Which tasks would you like to associate? (Space to select, Enter to confirm)
  [ ] web-ui-chart-zoom-and-interactions.md
  [x] recharts-styling-refinements.md
  [ ] backend-pipeline-fix.md
  ... (Type to filter) ...

> Added 1 existing task.

? Are there any new tasks to create? (Add prompts for AI)
> Create a task for adding annotation toggle to the UI
> Create a task for persisting annotation state in Firestore
> (Done)

✅ Project created!
   - Tagged existing tasks with `project: Chart Annotations`
   - Created stub tasks for prompts
```

## Requirements

### 1. Interactive Selection (Split-Pane TUI)
- Use **Bubble Tea** to create a rich TUI.
- **Layout**: Three-pane or Sidebar layout.
    - **Main**: Scrollable Task List.
    - **Preview**: Content of highlighted task.
    - **Selected Context**: A running list of selected tasks (the "Bundle").
- **View Toggle**: Press `Tab` or `v` to toggle List Mode:
    - **Compact**: Filename only.
    - **Expanded**: Filename + Title + Summary (Multi-line row).
- **Filtering**: Press `/` to open search bar.
    - Behavior: Real-time fuzzy filtering (telescope/fzf style).
    - Scope: Matches against Filename, Title, Frontmatter (Tags/Epic), and Body content.
- **Navigation (Vim Style)**:
    - `j` / `k` (or Arrows): Navigate list up/down.
    - `Tab` (or `l` / `h`): Switch focus between **Task List** and **Preview Pane**.
    - When Preview is focused, `j`/`k` scrolls the document content.
- **Selection**:
    - `y`: Add to bundle (Select).
    - `n`: Remove from bundle (Deselect).
- **Help**: `?` toggles a popup listing all keybindings.
- **Interaction**: `Enter` to confirm bundle.

### 2. Task Creation from Prompts (AI Loop)
- **Input**: User types natural language prompts (e.g., "Create a task for X").
- **Template Selector**: For each prompt (or batch), allow selecting a type:
    - `[F]eature` (Default)
    - `[B]ug`
    - `[R]esearch`
- **Process (Synchronous)**:
    - Show spinner ("🤖 Generating tasks...").
    - Stream results as they complete: "✅ Created `fix-login.md`: Refactor auth flow".
- **Handoff**: The CLI does NOT generate the file content locally. Instead, it:
    1.  Sends the prompt + project context to the Semantic Engine (Python API).
    2.  The API (via `roles.generation` LLM) generates the full task Markdown.
    3.  The CLI writes the result to `tasks/ready-to-start/`.
- **Status**: Show a spinner ("🤖 Generating task 1 of 3...") during this process.

### 3. "Project" Definition & Ad-Hoc Bundles
- **Named Project**: User enters "Q1-Cleanup". Tasks get `project: Q1-Cleanup`.
- **Anonymous Bundle**: User leaves name blank. Tasks are grouped in the session (e.g. for bulk completion or printing a summary) but NOT tagged with a project field.
- **Single Task**: Select one, confirm. Useful for "Find something to do".

## Implementation Plan
1.  **State Model**: Define the Bubble Tea model for the wizard (Steps: Name -> Select -> Prompt -> Confirm).
2.  **Task Loading**: Use the "Fast Path" (read JSON index) to populate the selection list instantly.
3.  **File Operations**: Update frontmatter of selected tasks; create new files for prompts.

## Deliverables
- `taskflow project start` command implementation in Go.
