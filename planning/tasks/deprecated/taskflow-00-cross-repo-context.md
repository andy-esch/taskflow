---
status: deprecated
epic: taskflow-v1-core
effort: 2-3 hours
tier: 2
priority: high
project: taskflow-bootstrap
tags: [architecture, config, multi-repo]
deprecated_at: 2026-06-07
deprecated_reason: Superseded by the scoped tskflwctl design (epic 17-pm-go-cli); old over-ambitious taskflow vision (docker/api/semantic/AI-gen) retired 2026-06-07. Research mined + kept.
updated_at: 2026-06-07
---

# TaskFlow 00: Cross-Repo & Documentation Strategy

**Goal**: Define how TaskFlow handles split repositories (Code vs Planning) and integrates implementation documentation.

## The Scenarios
1.  **Monorepo**: Code and Planning in one repo (e.g. `docs/planning`).
2.  **Split Repo**: Code in `desirelines`, Planning in `desirelines-planning`.

## Requirements
- **Config**: `.taskflow/config.yaml` must explicitly define where the "Code" lives if it's separate.
- **Context Injection**: When asking the Semantic Engine "how does auth work?", it needs to know if it should look in `planning/research/auth.md` OR `../desirelines/docs/auth.md`.

## Proposed Config Structure
```yaml
project:
  name: "Desirelines"
  mode: "split" # or "mono"

paths:
  planning_root: "." # Current repo (desirelines-planning)
  implementation_root: "../desirelines" # Relative path to code

docs:
  # Which folders should the Vector Store index?
  include:
    - "${planning_root}/tasks"
    - "${planning_root}/research"
    - "${implementation_root}/docs" # Index public docs too!
    - "${implementation_root}/README.md"
```

## Decisions to Make
1.  **Doc Indexing**: Should TaskFlow index the *code* documentation by default?
    - *Recommendation*: Yes. The Semantic Search is more powerful if it knows about both the *Plan* (Task) and the *Reality* (Docs).
2.  **References**: How to link them?
    - *Standard*: Use relative links if possible, or TaskFlow-specific links `taskflow://task/123`.

## Deliverables
- [ ] Define the `paths` schema in the Config design.
- [ ] Update `taskflow-06-embedding-pipeline.md` to support multi-root indexing.
