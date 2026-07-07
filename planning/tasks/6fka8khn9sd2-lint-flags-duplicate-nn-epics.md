---
schema: 1
id: 6fka8khn9sd2
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Two epics sharing a leading NN co-mingle tasks in the rollup and canonicalEpic silently picks the first; lint should flag duplicate NN keys (an invalid state nothing currently enforces).
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [core, lint, scheme-2-followup]
created: "2026-07-06"
updated_at: "2026-07-07"
started_at: "2026-07-06"
completed_at: "2026-07-07"
---

# Lint flags duplicate-NN epics

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)

## Origin (adversarial review, 2026-07-06)

Scheme-2 resolves epic refs on the leading-NN key (`domain.EpicRefKey`). If two epics share
an NN (e.g. `01-a.md`, `01-b.md`) — an invalid state nothing enforces — the rollup/`ShowEpic`
co-mingle their tasks and `canonicalEpic` silently returns the first. Add a lint check that
flags duplicate NN keys across epics (fail-open flag, like the other epic-status checks).
Both review passes surfaced this; no dup-NN exists today, so it's a guard, not a live bug.
