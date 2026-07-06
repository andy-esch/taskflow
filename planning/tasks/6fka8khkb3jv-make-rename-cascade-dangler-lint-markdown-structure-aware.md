---
schema: 1
id: 6fka8khkb3jv
status: ready-to-start
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: 'repointLinks/DanglingLinks use a naive inline-link regex — they miss reference-style [ref]: x.md links and false-positive on links inside fenced code blocks. Make both markdown-aware.'
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [core, cli, scheme-2-followup]
created: "2026-07-06"
updated_at: "2026-07-06"
---

# Make rename cascade + dangler lint markdown-structure-aware

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)

## Origin (adversarial review, 2026-07-06)

Both external and internal review passes flagged `mdLinkRe` (`\[..\]\(..\)`, inline only):
- **Reference-style links** `[label]: ../tasks/<id>-slug.md` aren't matched, so `rename`
  leaves them stale and `lint --links` won't flag them (false negative).
- **Fenced code blocks** — a `[..](x.md)` shown as an EXAMPLE inside ``` ``` gets cascaded /
  dangler-flagged (false positive).

Fix: a markdown-structure-aware pass (skip code fences; handle inline + reference-style) in
`repointLinks` (rename.go) and `DanglingLinks` (danglers.go). Zero reference-style links
exist in either tree today, so this is future-proofing — low priority.
