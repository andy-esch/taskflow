---
status: proposal
created: 2026-06-09
tags: [tui, ux, bubble-tea, design-spec]
---

# TUI UX Design and Navigation Specification

**Goal**: Establish a robust, keyboard-driven Terminal User Interface (TUI) design system and sprint roadmap for `tskflwctl ui` (Epic 18), functioning strictly as a primary adapter over the shared `core.Service`.

---

## 1. Architectural Boundary & Context

The TUI is a **primary adapter** over `core.Service`. To prevent logic duplication and maintain the single source of truth, it must interact with planning data solely through the Service layer (`ListTasks`, `ShowTask`, `Summary`, etc.) and **never** read or write files directly from the store/filesystem.

### Key Reductions for V1:
*   **No Projects Tab**: Projects are not backed by domain models in the current CLI; the Projects tab is cut from TUI V1. Active entities are limited to **Tasks**, **Epics**, and **Audits**.
*   **No Multi-Select (Sprint 0-2)**: Bulk actions (like bulk moving tasks) are deferred. Selection checkmarks (`Space` key) are cut from the initial sprints to keep the read-only phase lean.
*   **Bubble Tea Virtualization**: To support large planning trees (e.g. 500+ tasks), the list pane will use `bubbles/list` and `bubbles/viewport` out-of-the-box instead of custom scrolling or prefix-indexing.

---

## 2. Shared Render/Theme Extraction

The CLI and TUI have separate output surfaces: the CLI prints ANSI tables to standard writers, while the TUI composes Lip Gloss view components. They cannot share rendering functions directly, but they **must** share the semantic vocabulary. 
*   **Action**: Extract a shared theme package (containing status glyph mappings, priority colors, and progress calculation logic) so status styles remain unified.

---

## 3. UI Layout & Navigation Specification

The TUI maintains a stable double-pane Miller Column grid layout:

```text
┌─[1] Navigation ────────────────────────┐┌─[2] Detail Preview: task-recharts-zoom.md ───────┐
│  Tasks  [Epics]  Audits                ││                                                  │
├────────────────────────────────────────┤│ # Recharts Zoom and Filter Interactions          │
│ Active Epics (3)                       ││                                                  │
│ ⏺ 04-frontend                          ││ **Status:** planning     **Priority:** high      │
│ ○ 05-vector-engine                     ││ **Created:** 2026-06-08  **Tags:** ui, chart     │
│ ○ 06-auth-v2                           ││                                                  │
│                                        ││ ## Objective                                     │
│                                        ││ Build custom zoom brush component in Recharts    │
│ Completed Epics (14)                   ││ to support time-series zooming.                  │
│ ○ 01-skeleton                          ││                                                  │
│ ○ 02-contracts                         ││ ## Epic Tasks (2/3 completed)                    │
│ ○ 03-fs-store                          ││ - [x] task-recharts-setup.md                     │
│                                        ││ - [x] task-brush-slider.md                       │
│                                        ││ - [ ] task-recharts-zoom.md                      │
│                                        ││                                                  │
└────────────────────────────────────────┘└──────────────────────────────────────────────────┘
┌─ Command Output / Status ──────────────────────────────────────────────────────────────────┐
│ 🤖 Ready. Loaded 19 tasks and 3 epics from /Users/andyeschbacher/git/andy-esch/taskflow.    │
├────────────────────────────────────────────────────────────────────────────────────────────┤
│ [Tab] Cycle Tab  [j/k] Move  [h/l] Focus Preview  [Space] Select  [/] Search  [q] Quit     │
└────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Keyboard Navigation Mapping (Refined):
*   **`[` and `]`**: Cycle horizontally through the active entity tabs (`Tasks` $\leftrightarrow$ `Epics` $\leftrightarrow$ `Audits`).
*   **`Tab` or `h` / `l`**: Toggle active focus horizontally between the Left Pane (List) and the Right Pane (Detail Preview).
*   **`1` or `2`**: Absolute focus jumps (`1` = Left Pane, `2` = Right Pane).
*   **`j` / `k` (or Arrows)**: Vertical list navigation or viewport document scrolling.
*   **`Enter` / `l`**: Focus the Right Pane on the selected task and scroll.
*   **`Esc`**: Return focus to the Left Panel.
*   **`r`**: Trigger a manual reload of data from `core.Service` (V1 standard; no auto-watchers).
*   **`q` or `Ctrl+c`**: Quit.

---

## 4. Empty, Error, and Integrity States

*   **No-Repo Error**: If the tool is run outside a planning repository, the TUI must boot into a fullscreen error state: `"Not a taskflow repository. Run 'tskflwctl init' or set '-C' to target a repo."`
*   **Parser Failures (FileProblems)**: Any frontmatter validation or parsing issues returned by `ListTasks()` must be displayed as a warning badge in the status bar (e.g. `⚠️ 2 issues`). Pressing `e` or `i` will open a viewport modal showing the detailed parse errors.

---

## 5. Technical Risks & Remediation (Adversarial Review Findings)

Following a joint critical review of the CLI codebase, the following runtime complexities were resolved:

### A. Cobra Pre-Run Hijack (Sprint 0/1)
*   **The Risk**: The root Cobra command enforces `PersistentPreRunE: func(...) { return app.resolve() }`. If run outside a repo, `resolve()` errors and Cobra aborts immediately before `tskflwctl ui` can render its fullscreen error view.
*   **Remediation**: The `ui` subcommand must override parent pre-run hooks with a custom `PreRunE` that catches and ignores the `app.resolve()` error, passing the failure status into the Bubble Tea initializer (`NewModel(svc, err)`), allowing the TUI to render its own styled error viewport.

### B. Blocking Disk I/O during List Scroll (Sprint 1)
*   **The Risk**: Selecting a list item triggers `app.Svc.ShowTask(slug)` which performs file read operations. Running this synchronously in `Update()` will cause stuttering and input lag when scrolling.
*   **Remediation**: Wrap all file-read calls in asynchronous Bubble Tea commands (`tea.Cmd`). Ensure `Update` verifies that the loaded slug matches the *currently highlighted* slug upon message delivery to prevent race conditions during rapid scrolling.

### C. Input Interception during List Filter (Sprint 2)
*   **The Risk**: `bubbles/list` has a built-in search mode (`list.Filtering`). When active, typing global keys (like `[` or `]` for tabs, or `q` for quit) will trigger parent events instead of typing inside the input.
*   **Remediation**: Implement a strict keyboard routing check inside `Update()`: if `taskList.FilterState() == list.Filtering`, forward key events exclusively to the list component and return early.

### D. Multi-Model Tab State (Sprint 2)
*   **The Risk**: If the TUI cycles `Tasks`/`Epics`/`Audits` tabs using a single `bubbles/list` instance with dynamic data swapping, cursor indexes and scroll offsets are lost.
*   **Remediation**: The TUI parent model must hold three distinct instances of `list.Model` (Tasks, Epics, Audits). Cycling tabs merely routes incoming window sizes and keyboard messages to the active sub-model instance.

### E. Glamour Rendering and Resizing Overhead (Sprint 4)
*   **The Risk**: Compiling glamour markdown inside `View()` is highly CPU intensive and will cause terminal lag.
*   **Remediation**: Cache the compiled Glamour output in the model state during `Update()` only when selection changes or a resize event occurs. Avoid calling Glamour inside `View()`.

---

## 6. Testing Strategy

*   **Unit Testing (`Update`)**: Since Bubble Tea's `Update()` function is a pure state-mutation machine, we can write fast tests that send keyboard or custom messages directly to the model and assert the resulting model fields.
*   **Integration Testing (`x/teatest`)**: We will introduce `charmbracelet/x/teatest` as a dev-only dependency to spin up mock terminal inputs and assert correct layouts against stdout buffers.
*   **Deterministic Testing**: Inject a mock `core.Service` during `teatest` runs to simulate synchronous, instantaneous returns, preventing timing flakiness.

---

## 7. Proposed Sprint Planning (Epic 18)

*   **Sprint 0 — Foundation**: Clear out the stale 53-line `internal/tui` spike; register the `ui` cobra command overriding Cobra `PersistentPreRunE` wired to `app.Svc`; import Bubble Tea and `bubbles` library components; construct a minimal model that loads tasks via `Service`, handles resizing, and exits on `q`; extract the shared CLI/TUI style theme; set up the `Update()` unit-test harness.
*   **Sprint 1 — Read-Only Browser**: Build the double-pane Tasks list + detail preview using Lip Gloss borders, vim navigation, focus highlights, viewport-scrolled document body, footer key legend, and handle empty, no-repo, and unreadable file states.
*   **Sprint 2 — Multi-Entity & Search**: Integrate tabbed navigation for Tasks/Epics/Audits; render epic progress bars and audit finding counts; add `/` search text input for real-time list filtering in memory.
*   **Sprint 3 — Actions**: Define keyboard mutation triggers (`s` to start, `c` to complete, `d` to defer), trigger confirmations, route calls through `Service.Move`/`SetFields`, refresh views, and re-evaluate multi-select bulk actions.
*   **Sprint 4 — Polish**: Render initial dashboard landing views; parse markdown previews using Glamour; implement the `?` help modal; optionally integrate `fsnotify` file watchers for automatic reloading.
