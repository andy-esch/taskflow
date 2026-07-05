---
status: deprecated
epic: taskflow-v1-core
effort: 3-4 hours
tier: 2
priority: high
project: taskflow-bootstrap
tags: [ai, generation, architecture]
deprecated_at: 2026-06-07
deprecated_reason: Superseded by the scoped tskflwctl design (epic 17-pm-go-cli); old over-ambitious taskflow vision (docker/api/semantic/AI-gen) retired 2026-06-07. Research mined + kept.
updated_at: 2026-06-07
id: 6f9yr8m03aws
---

# TaskFlow 09a: AI Task Generation Flow

**Goal**: Design the mechanism for converting user prompts into structured Task Markdown files.

## The Problem
In `project start`, the user types "Create a task for X". We need to turn that 1-line string into a high-quality Markdown file with Frontmatter, Objectives, and Implementation Plans.

## Architecture
1.  **CLI**: Collects prompts list: `["Fix login", "Add auth"]`.
2.  **Request**: `POST /generate/tasks`
    ```json
    {
      "project": "Q1-Cleanup",
      "prompts": ["Fix login", "Add auth"],
      "context_tasks": ["existing-task-1.md"] // Optional: for style matching
    }
    ```
3.  **API (Semantic Engine)**:
    - Loads `roles.generation` LLM (e.g., Claude).
    - Injects "Task Style Guide" (from `taskflow-00-documentation-standards`).
    - Generates Markdown content.
4.  **Response**: JSON list of `{ "filename": "fix-login.md", "content": "..." }`.
5.  **CLI**: Writes files to disk.

## Questions to Answer
- **Latency**: Doing this 1-by-1 might be slow. Should we parallelize?
- **Review**: Does the user get to review the generated task before saving?
    - *Decision*: For V1, save to `tasks/ready-to-start` and let user review in editor or CLI later.

## Deliverables
- [ ] API Endpoint `POST /generate/tasks`.
- [ ] Prompt Template for "Task Generation" (System Prompt).
