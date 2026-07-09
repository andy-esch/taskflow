---
schema: 1
status: active
description: 'Home for new first-class entities beyond task/epic/audit (routines, projects, ADRs): fields, lifecycle, storage, and CLI surface for each new noun — the ''which nouns exist'' axis.'
priority: low
tags: []
created: "2026-07-07"
---
# First-class entities — new planning nouns

Home for adding new first-class entities BEYOND task/epic/audit — the "which
nouns exist" axis of the model. Distinct from epic 24's storage-model evolution
(stable-key layout / read-model / OCC), which is about HOW entities are stored;
this epic is about WHICH nouns exist and what they mean.

## Charter

Each new noun that earns first-class status is defined through the same machinery
the existing entities use:

- its frontmatter fields + lifecycle (status/bucket vocabulary, if any),
- its on-disk layout (flat, id-led — per ADR-0003),
- its CLI surface (`<noun> new|list|show|…`) over `core.Service`,
- how it relates to existing entities (cross-references, rollups).

## Candidate tenants

- **routines** — a first-class entity tracking the routine↔audit lineage (audit
  gets a `routine:` field; `audit list --routine`); complements Claude Code's
  scheduler. Tenant #1 (the spike moved here from epic 24).
- **projects** — a grouping above/beside epics (TBD).
- **ADRs as tracked entities** — decisions as first-class, queryable records
  rather than freeform docs (TBD).

## Relationship to other epics

- Epic 24 (data-model evolution) — provides the storage foundation (flat id-led
  layout, OCC, projection) every new noun rides on.
- Epic 26 (frontmatter schema) — a new noun's fields should be declared through
  the same field registry, once that lands.
