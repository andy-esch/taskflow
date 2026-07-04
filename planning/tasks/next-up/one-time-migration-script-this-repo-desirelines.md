---
schema: 1
status: next-up
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: 'Throwaway script for the destructive cutover only: id-led renames, status/bucket to frontmatter, ref rewrites. Run on a copy; git is the undo. id-assignment split to its own lint --fix task.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [migration, storage]
created: "2026-07-01"
updated_at: "2026-07-04"
id: 6fhnydm029gt
---

# One-time migration script (this repo + desirelines)

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]

Backfill-timestamp policy (decided 2026-07-02, ADR-0003 amendment): embed the file's created: date; if absent use another frontmatter date; if neither exists, ERROR so the operator adds the date key. Sub-day ordering = a random low tail via id.NewAt(unixMilli) (stateless, no sequence counter), with dedupe-and-regenerate on the rare same-day id collision.

## Coordination note (2026-07-04) — sweep loose files into `meta/`

Added by the carveout decision
([[curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate]]).
The flat scan errors on any non-entity `.md` left in a scanned bucket (a stray becomes a
`FileProblem`; `README.md` is the one carve). So the one-time migration must **sweep known
loose files out of the entity buckets into a top-level `planning/meta/`** as part of the
cutover, so buckets start clean:

- `audits/HOWTO-execute.md` -> `planning/meta/` (audit execute-doc).
- routine specs / execute-docs -> `planning/meta/routines/`.
- any other non-entity `.md` sitting in `tasks/`/`epics/`/`audits/` -> `planning/meta/`.

**Acceptance criterion:** after migration, `tskflwctl lint` reports zero stray-file
`FileProblem`s in the entity buckets.

Open (migration-author's call, not blocking): the exact `meta/` substructure — flat
(`meta/HOWTO-execute.md`) vs grouped by origin (`meta/audits/...`). "Free internal structure"
was the decision; pick per file.
