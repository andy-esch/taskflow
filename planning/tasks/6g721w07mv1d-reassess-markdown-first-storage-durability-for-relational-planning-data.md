---
schema: 1
id: 6g721w07mv1d
status: ready-to-start
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Revisit whether Markdown frontmatter should remain authoritative as cross-entity constraints grow, without sacrificing Git-native and AI-native workflows.
effort: 1-2 days
tier: 4
priority: low
autonomy_level: 2
tags: [architecture, storage, git-native, research]
created: "2026-09-05"
---

# Reassess Markdown-first storage durability for relational planning data

## Objective

Revisit the boundary between human-readable planning documents and authoritative structured state
now that repository-global dependencies, shared Thread membership, lifecycle constraints, and
concurrent adapters place relational demands on independently editable files. Preserve the goals of
Markdown-first authoring, Git-native review, and AI-native inspectability rather than assuming that
“move it to a database” is automatically the durable answer.

## Scope

- Inventory concrete integrity, merge, concurrency, recovery, and usability failures from the
  shipped storage model, including what guarded graph repair reveals. Separate representation flaws
  from projection and mutation-boundary flaws.
- Compare at least: authoritative Markdown snapshots; Markdown bodies plus structured sidecars;
  append-only Git-native events with generated readable projections; and a database-authoritative
  store with deterministic Markdown import/export.
- Evaluate human diffs and merges, offline use, direct edits, agent context cost, referential
  integrity, multi-record transactions, schema evolution, history, corruption recovery, TUI/web
  adapters, and planning spaces spanning several implementation repositories.
- Define where authority lives in each option and whether a hybrid merely creates two competing
  truths. Include migration, rollback, compatibility, and disaster-recovery consequences.
- Produce a focused ADR recommendation: retain the current model with explicit limits, adopt a
  staged hybrid, or justify a different authority model. A no-change result is valid.

## Acceptance criteria

- [ ] The review starts from observed repository failure modes and scale, not generic database
      benefits or speculative graph features.
- [ ] Each candidate is scored against Markdown-first, Git-native, and AI-native product goals as
      well as integrity and transactional guarantees.
- [ ] The recommendation identifies which constraints can be strengthened within the existing
      source/projection/guard architecture and which fundamentally require a new persistence model.
- [ ] Any proposed hybrid has one unambiguous authority and a deterministic, recoverable projection
      contract; generated files are never silently treated as a second writable source.
- [ ] The resulting ADR names an incremental proof and migration boundary before authorizing an
      implementation epic.

## Out of scope

- Implementing a new backend, blocking current Threads hardening, or treating an opaque local
  database as acceptable without a reviewable Git artifact strategy.

## Sequencing

This is intentionally not on the production Threads critical path. Revisit after guarded repair has
produced real recovery experience, then use that evidence to decide whether the current authority
model needs refinement or replacement.

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)
- Prior research [task storage model: files, logs, or versioned DB](../research/6ffdv9g00d53-task-storage-model-files-logs-or-versioned-db.md)
- ADR [stable-key ID-addressed storage](../adrs/0003-stable-key-id-addressed-storage.md)
- Threads recovery task [guarded broken-graph repair](6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md)
