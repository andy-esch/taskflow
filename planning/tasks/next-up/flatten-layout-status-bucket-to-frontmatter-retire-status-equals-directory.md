---
schema: 1
status: next-up
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Move status/bucket into frontmatter as source of truth and flatten tasks/audits to one dir each (id-led filenames); update store, layout, WatchPaths, resolution, completion. Per ADR-0003.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [core, storage]
created: "2026-07-01"
updated_at: "2026-07-02"
id: 6fhnydm03edq
---

# Flatten layout: status/bucket to frontmatter, retire status-equals-directory

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]

Note (id-generator adversarial review, 2026-07-02): the create path MUST do an explicit id-collision check, not rely on O_EXCL. With <id>-<slug>.md the exclusive-create guards the whole filename, not the id, so two different-slug tasks that drew the same id would both be created. On create, scan for an existing <id>-* and regenerate on a hit. That still has a cross-process TOCTOU race (two agents both scan clear, both create), so the definitive cross-process guarantee is serve single-writer serialization (ADR-0004); a slipped duplicate id is a recoverable ErrAmbiguous (like a duplicate slug today) that lint can dedup by id.
