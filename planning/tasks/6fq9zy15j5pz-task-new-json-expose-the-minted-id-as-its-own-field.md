---
schema: 1
id: 6fq9zy15j5pz
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: task new --json created.id is the slug; the minted id is only inside created.path.
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [json, id]
created: "2026-07-18"
---
# task new --json: expose the minted id as its own field

## Objective

`task new --json` returns `created.id` = the **slug**; the real minted id is
only embedded in `created.path` (`tasks/<minted-id>-<slug>.md`). Scripting
`created.id` expecting the id yields the slug. Confirmed:
`{"created":{"id":"zzz-probe","path":"tasks/6fq9xy66fehv-zzz-probe.md"}}`.

## Acceptance criteria

- [ ] The minted id is a distinct top-level field on the created object
- [ ] `id` vs `slug` are unambiguous in the envelope (and consistent across new for task/epic/audit)

## Notes

- Loci: the create envelope in internal/wire (created DTO) + the CLI `new` handlers.
- Source: https://github.com/andy-esch/taskflow/issues/105 (P4, Low)
