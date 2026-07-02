---
schema: 1
status: next-up
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: 'Throwaway script: assign ids, rename files, move status/bucket to frontmatter, rewrite refs. Hard cutover; run on a copy and commit; git is the undo. Per ADR-0003 section 6.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [migration, storage]
created: "2026-07-01"
updated_at: "2026-07-02"
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
