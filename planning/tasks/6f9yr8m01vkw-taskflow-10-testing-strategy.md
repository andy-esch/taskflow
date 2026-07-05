---
status: deprecated
epic: taskflow-v1-core
effort: 2-3 hours
tier: 2
priority: medium
project: taskflow-bootstrap
tags: [testing, quality, templates]
deprecated_at: 2026-06-07
deprecated_reason: Superseded by the scoped tskflwctl design (epic 17-pm-go-cli); old over-ambitious taskflow vision (docker/api/semantic/AI-gen) retired 2026-06-07. Research mined + kept.
updated_at: 2026-06-07
id: 6f9yr8m01vkw
---

# TaskFlow 10: Task Verification & Testing Strategy

**Goal**: Define how we verify that created tasks (human or AI generated) meet our quality standards.

## Context
We are mandating "Required Sections" (Acceptance Criteria, Research Links, etc.) in our templates. We need a way to enforce this, effectively running "tests" on our documentation.

## Requirements to Scope
1.  **Template Validation**: Can `taskflow validate` check for specific headers (e.g. "## Acceptance Criteria") in addition to frontmatter?
2.  **Smoke Tests for Tasks**: A concept where a task can define how to verify it is done.
    - *Idea*: A `verification_command` field in frontmatter? (e.g. `pytest tests/test_auth.py`)
    - *Idea*: AI Agent runs the "Acceptance Criteria" as a script?
3.  **UI/Frontend Verification**:
    - For tasks tagged `web-ui` or `frontend`, mandate a "Manual Verification" section.
    - Requirements: "Click X", "Verify Y appears", "Check mobile view".
    - Potential: Require screenshot attachments for PRs?
4.  **CI/CD for Docs**: Should we fail the build if a task in `next-up` is malformed?

## Deliverables
- [ ] Update `taskflow validate` logic to parse Markdown AST and check for required headers.
- [ ] Define "UI Verification Template" extension.
- [ ] Research: "Executable Tasks" - embedding verification commands.
