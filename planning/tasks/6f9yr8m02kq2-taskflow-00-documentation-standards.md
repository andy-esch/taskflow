---
status: deprecated
epic: taskflow-v1-core
effort: 2-3 hours
tier: 2
priority: high
project: taskflow-bootstrap
tags: [documentation, standards, process]
deprecated_at: 2026-06-07
deprecated_reason: Superseded by the scoped tskflwctl design (epic 17-pm-go-cli); old over-ambitious taskflow vision (docker/api/semantic/AI-gen) retired 2026-06-07. Research mined + kept.
updated_at: 2026-06-07
id: 6f9yr8m02kq2
---

# TaskFlow 00: Documentation Standards & Style Guide

**Goal**: Define the official "Language of TaskFlow" to ensure consistency for Humans and AI.

## Scope
We need precise definitions and formatting rules for:

### 1. The Hierarchy
- **Project**: What qualifies? How long does it last? (e.g., "A cross-cutting initiative lasting 2-6 weeks").
- **Epic**: How does it differ from a Project? (e.g., "Long-lived functional area or theme").
- **Task**: Granularity guidelines. (e.g., "Must be completable in < 1 day").

### 2. Document Types
- **ADR (Architecture Decision Record)**: When to write one? Format?
- **Research/RFC**: How does it transition to an Epic/Task?
- **Incidents**: Post-mortem format.

### 3. File Formats
- **Frontmatter**: The definitive list of allowed fields (`status`, `tier`, `effort`, `project`, etc.).
- **Required Sections**: Every task must have:
    - `Description` / `Goal`: High-level summary.
    - `Context` / `Problem Statement`: Why are we doing this?
    - `Research & References`: Links to external docs, research notes, or source URLs.
    - `Acceptance Criteria`: Checklist of what "Done" looks like.
- **Naming Convention**: `kebab-case.md`. Date prefixes for completed items?

## Deliverables
- [ ] `docs/STYLE_GUIDE.md`: The rulebook.
- [ ] `docs/TEMPLATES/`: Standard templates for:
    - `TASK_FEATURE.md` (Standard implementation)
    - `TASK_BUG.md` (Reproduction steps, fix criteria)
    - `TASK_RESEARCH.md` (Hypothesis, experiments)
- [ ] Update `AI_README.md` to reference these standards.
