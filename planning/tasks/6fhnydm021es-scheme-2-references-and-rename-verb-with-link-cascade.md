---
schema: 1
status: completed
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
completed_at: "2026-07-05"
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

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)
- Follows [flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory](6fhnydm03edq-flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory.md) (Phase B).

## Pieces 2a + 2b landed (2026-07-05)

Body cross-links are now GitHub-clickable relative-path markdown:
- **2a (scaffold):** `task new` stores the epic's canonical `<NN>-<slug>` stem (resolved on
  the NN key) and the template emits `- Epic [<stem>](../epics/<stem>.md)` instead of `[[…]]`.
- **2b (migration):** `internal/tools/wikimigrate` converted existing `[[slug]]` wikilinks to
  relative-path markdown across both planning trees — taskflow (531 links / 173 files) and
  desirelines (309 / 115). Byte-safe, 0 broken; placeholders + danglers left untouched (the
  danglers are what piece 4's lint will flag).

Remaining: piece 3 (rename verb with link cascade) + piece 4 (dangler lint).

## Pieces 3 + 4 landed — scheme-2 complete (2026-07-05)

- **3 (rename verb):** `tskflwctl task rename <task> "<new title>"` — derives a new slug
  (12-char id kept), rewrites the body H1, and cascades every inbound relative-path markdown
  link across the tree to the new filename (freshening a link whose display text was the old
  slug). Write-locked, not CAS'd (a rare, deliberate op). `store.RenameTask` + tests + docgen.
- **4 (dangler lint):** opt-in `tskflwctl lint --links` — flags any body `[..](path.md)`
  whose target file is missing (exit 11), skipping external + placeholder links. A narrow
  `core.Linter` port wired to the FS like `Fixer`. Default `lint` is unchanged, so the cron
  routines' gate stays clean despite a tree's pre-existing danglers.

All four pieces done; the `epic show <NN>` roster bug is retired.

Follow-up (not blocking): ~86 pre-existing markdown danglers across both planning trees
(stale old-layout paths, typo'd slugs, deleted refs) a future `lint --links` sweep could
clean up; and `rename` for audits/epics (task-only today).
