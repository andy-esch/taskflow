---
schema: 1
id: 6fjw1d2jm6e9
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: 'Explore routines as a first-class entity tracking the routine<->audit lineage (audit gets a routine: field; audit list --routine). Complements Claude Code''s scheduler; maybe its own epic.'
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [core, design]
created: "2026-07-04"
updated_at: "2026-07-07"
---

# Spike: routines as a first-class entity + routine↔audit lineage

## The idea (user, 2026-07-04)

A recurring agent **routine** (spec) *produces* an **audit** (a tracked entity), which links
back to the routine + its execute-doc. That's the same lineage shape as **task↔epic** — and
nothing in the planning-tool space models *recurring agent workflows and their outputs*.
Making it first-class could turn tskflwctl from a task tracker into a **recurring-agent-
workflow tracker**, which fits the cron-AI-agent thesis this project leans on.

## Crucial framing — it composes with Claude Code, doesn't compete

Claude Code's **routines** are the *scheduler/runner*: a routine there is a saved prompt +
repos + trigger, registered **server-side** (web UI / `/schedule`), NOT discovered from a
`routines/` directory (confirmed against the official Routines docs, 2026-07-04). So the
local `routines/*.md` files are the **specs the routine's prompt points agents at**, and
tskflwctl's opportunity is the **other half**: tracking the *routine spec ↔ audit output*
lineage **in the planning repo**. Claude Code fires it; tskflwctl tracks what it produced.

## What to explore

- **Routines as an entity type** (like tasks/epics/audits): `planning/routines/*.md` with a
  routine schema; `tskflwctl routine list/show`. (Contrast with the near-term
  [curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate](6fjvr03mr9zg-curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate.md) path, which
  merely *tolerates* untracked routine files — this would *manage* them.)
- **The linkage:** an audit carries a `routine:` frontmatter field (like `epic:`), so
  `audit list --routine weekly-code-audit`, a "last run / next due" view, and a routine→its
  audits roster fall out. The execute-doc (`HOWTO-execute.md`) becomes the routine's linked
  doc rather than a loose file.
- **State files** routines keep (e.g. `architecture-rotation-state.json`, `no-op-log.md`):
  where they live, whether the tool is aware of them.

## Open questions

- Does this warrant its **own epic** (it's a feature direction, not really data-model 24)?
- Is `routine:` a hard reference (resolved/lint-checked) or a soft tag?
- Overlap with the scaffolded `adr`/`project` entity groups + the entity-registry
  fan-out (ADR-0003) — is "routine" just another `Descriptor` entry?

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md) (nearest home;
  likely deserves its own epic if pursued)
- Near-term counterpart: [curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate](6fjvr03mr9zg-curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate.md)
- Grounded by the desirelines `routines/` ↔ `audits/HOWTO-execute.md` ↔ audit-output workflow.
