---
schema: 1
status: in-progress
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Resolve epic refs on the id prefix; convert body cross-links from wikilinks to standard relative-path markdown; add a rename verb that cascades inbound links; lint flags danglers. Per ADR-0003.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [core, cli]
created: "2026-07-01"
id: 6fhnydm021es
updated_at: "2026-07-05"
started_at: "2026-07-05"
---
# Scheme 2 references and rename verb with link cascade

## Objective

Finish the Scheme-2 reference model (ADR-0003 / epic 24 decision 2026-06-30): references
resolve on the stable **id/NN prefix** (not the cosmetic slug); body cross-links are
**GitHub-clickable relative-path markdown** (not `[[wikilinks]]`); and a **`rename`** verb
re-titles an entity while cascading its inbound links, with `lint` flagging danglers.

## Pieces (sequenced)

1. **[done 2026-07-05] Epic refs resolve on the id/NN prefix.** A task's `epic:` ref now
   joins its epic by the leading NN (`domain.EpicRefKey`), so a drifted slug or a bare NN
   still resolves — across the rollup, `ShowEpic`, `epicExists`, and lint's epic-existence
   check. Fixes the `epic show <NN>` empty-roster bug. (`EpicRefKey` + tests.)
2. **Body cross-links → relative-path markdown.** Change the body scaffold/template to emit
   `[display](<id>-<slug>.md)` instead of `[[slug]]`, and one-time-convert existing `[[…]]`
   in the planning bodies (both repos) to markdown links. Wikilinks are dropped (GitHub
   renders the markdown form, not `[[…]]`).
3. **`rename` verb with link cascade.** Re-title a task/audit (new slug; id unchanged) and
   rewrite every inbound body link (`](…/<id>-*.md)`) to the new filename in one atomic pass.
4. **Dangler lint.** `lint` flags a body relative-path markdown link whose target file
   doesn't exist, fail-open like the other lints.

## Sub-items / refinements

- Lint refreshes an epic ref's **cosmetic slug suffix** when it has drifted (a `lint --fix`
  nicety; resolution already tolerates the drift).
- On create, store the **canonical stem** for an `epic:` ref so fresh tasks carry a current suffix.

## Acceptance criteria

- [x] Epic refs resolve on the NN prefix; `epic show <NN>` reports the roster.
- [ ] The body scaffold emits markdown links; no `[[wikilinks]]` remain in planning bodies.
- [ ] `tskflwctl rename <slug> "<new title>"` renames the file and cascades inbound links.
- [ ] `lint` flags a body link to a non-existent file.

## Out of scope

- The flatten itself (Phase B — shipped) and its one-time migration.
- Changing the epic `NN-<slug>` scheme (epics keep NN; only id-led tasks/audits use the id).

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]
- Follows [[flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory]] (Phase B).
