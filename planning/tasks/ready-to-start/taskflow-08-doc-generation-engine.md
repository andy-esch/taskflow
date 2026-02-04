---
status: ready-to-start
epic: taskflow-v1-core
effort: 4-6 hours
tier: 2
priority: high
project: taskflow-bootstrap
tags: [documentation, automation, generation]
---

# TaskFlow 08: Documentation Generation Engine

**Goal**: Implement the "Active Generation" feature where TaskFlow automatically maintains high-level project documentation based on task state.

## Features
1.  **Roadmap Generator**:
    - Input: All Epics + `tasks/next-up` + `tasks/ready-to-start`.
    - Output: `docs/ROADMAP.md` (Timeline view, grouped by Epic/Project).
2.  **Changelog Generator**:
    - Input: `tasks/completed` (filtered by date).
    - Output: `CHANGELOG.md` or `docs/HISTORY.md`.
3.  **Status Report**:
    - Output: A weekly markdown summary of progress (completed vs planned).

## Implementation
- **Logic**: Python Service (Semantic Engine) handles the heavy text processing/formatting.
- **CLI**: `taskflow generate <type>` triggers the job.
- **Config**: Templates for the output format should be customizable in `.taskflow/config.yaml`.

## Deliverables
- [ ] `generate_roadmap()` function in Python backend.
- [ ] `generate_changelog()` function.
- [ ] CLI command `taskflow generate`.
- [ ] Default templates for Markdown output.
