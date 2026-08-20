---
schema: 1
id: 6g21bv4mjvz0
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: 14 of 30 docs have no tags, so research list --tag silently misses half the corpus — and that query is the actual answer to the 'does research need a status' question.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [planning, discovery]
created: "2026-08-20"
---

# Backfill tags so the corpus is discoverable by topic

## Objective

This is the real answer to "does research need a status". The reader question a status
would have answered — *what is the current guidance on X?* — is already answered by
`research list --tag X`, which is newest-first. No new field needed. The gap is that
**14 of 30 docs have no usable tags**, so that query silently misses half the corpus.

    $ research list --tag tui -o table -c created,slug
    2026-08-15  multi-space-home-registry-and-the-atlas
    2026-06-28  color-palette-and-theming-overhaul
    …

Works today; just incomplete.

## The 14 without tags

All the migrated legacy docs plus the 2026-06-06 cohort: `workspaces-vs-root-modules`,
`configuration-and-onboarding-flow`, `hybrid-search-architecture`,
`documentation-integration-strategy`, `ai-file-editing-strategy`,
`ai-interaction-interfaces`, `cli-language-options`, `monorepo-structure-proposal`,
`ai-provider-abstraction`, `pm-cli-architecture-and-go-port`,
`go-cli-foundation-architecture`, `project-concept-cross-cutting-initiatives`,
`tskflwctl-command-spec`, `dashboard-extension-ideas`.

Same shape of job as the description backfill: read each doc, write tags from what it is
actually about, apply with `research set --tags`.

## Acceptance criteria

- [ ] Every research doc has at least one tag; `research list --tag <t>` reaches the whole
      corpus.
- [ ] Tags are drawn from a consistent vocabulary — reuse what the tagged 16 already use
      rather than inventing a parallel set per doc. Worth listing the existing vocabulary
      first and reconciling.
- [ ] Spot-check the currency query on two or three topics (tui, cli, storage) and confirm
      the newest doc really is the current guidance.
- [ ] Diff touches only `tags` and `updated_at`.

## Out of scope

- A declared tag vocabulary or lint rule enforcing one — decide after seeing what the
  reconciled set looks like.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Decision record: [formalize-the-research-frontmatter-contract](6g0493z6xwc7-formalize-the-research-frontmatter-contract.md)
- Precedent: the description backfill (commit `92bd793`), same method
