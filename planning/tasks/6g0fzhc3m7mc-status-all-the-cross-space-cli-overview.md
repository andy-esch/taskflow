---
schema: 1
id: 6g0fzhc3m7mc
status: next-up
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Per-space summaries plus a combined in-progress list across the registry. The CLI counterpart of the proposed atlas, and the cheapest test of whether the board is worth building.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, multi-repo]
created: "2026-08-15"
---
# `status --all`: the cross-space CLI overview

## Objective

Answer the question that actually motivated epic 29 — **"what am I in the middle of,
anywhere?"** — in the CLI, before any TUI work. `status` is already the terminal
counterpart of the TUI's Overview dashboard; `status --all` is the counterpart of the
proposed atlas, and the cheapest possible test of whether a cross-space view is worth
building a screen for.

## Notes

- Per-space `core.Summary` over the registry, rendered as a compact per-space block
  plus a **combined in-progress list** with a space badge on each row. The combined
  list is the hypothesis worth testing: cards orient, but the rail may be the real
  payload.
- Broken spaces render their diagnosis inline (from the health task) rather than
  failing the command — one dead entry must not cost you the other three.
- `--json` carries a `schema_version` and an array of per-space envelopes; goldens
  regenerated (`go test ./internal/cli -update`).
- Reads only. No writes, no registration side effects.
- Sequential scanning is fine to start; measure before adding concurrency.

## Acceptance criteria

- [ ] `status --all` summarizes every registered space plus a combined in-progress list
- [ ] A broken space shows its diagnosis inline; the command still exits 0
- [ ] No registry / no spaces → falls back to today's single-repo `status` output
- [ ] `--json` envelope with `schema_version`; goldens regenerated
- [ ] `just test` + `just lint` green

## Out of scope

- The TUI atlas board — this exists partly to decide whether that is worth building
- Caching or persisted stats

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [6g0ajre026c6-multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
