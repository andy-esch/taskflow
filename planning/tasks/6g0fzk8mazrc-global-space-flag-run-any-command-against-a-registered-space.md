---
schema: 1
id: 6g0fzk8mazrc
status: next-up
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: --space <id> (and TSKFLW_SPACE) resolves any command against a registered repo instead of the cwd. Small change to App.resolve(); also the cross-repo handle agents lack today.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, multi-repo]
created: "2026-08-15"
---
# Global `--space` flag: run any command against a registered space

## Objective

Give every command a handle to another planning repo without a `cd`: `--space <id>`
(plus `TSKFLW_SPACE`) looks the id up in the registry and discovers from **there**
instead of the cwd. Small change to `App.resolve()`, disproportionate payoff — it is
also the cross-repo handle agents currently lack.

Plausibly **the most valuable single piece of epic 29**, and the reason the CLI slice
ships before any TUI work: if `--space` plus `status --all` turns out to be most of
the value, the board's cost can be re-decided honestly.

## Notes

- Precedence: `--space` > `TSKFLW_SPACE` > cwd discovery. `-C` and `--space` together
  should be a loud error, not a silent winner — they are two answers to one question.
- An unknown id errors with `ErrNotFound` (exit 10) and lists the known ids; a
  registered-but-broken space errors with the same diagnosis text the health task
  produces, not a raw `Discover` error.
- Shell completion for `--space` off the registry (the completion funcs already do
  their own forgiving discovery, so this must not break outside a planning repo).
- Must not perturb the advisory invariant: with no `--space` and no home config,
  resolution is byte-for-byte today's.

## Acceptance criteria

- [ ] `--space <id>` resolves any command against the registered repo, `-C`-style
- [ ] `TSKFLW_SPACE` honored at lower precedence than the flag
- [ ] `--space` + `-C` together is a clear error
- [ ] Unknown id → exit 10 with the known ids listed
- [ ] Completion offers registered ids and stays silent outside a planning repo
- [ ] `just test` + `just lint` green

## Out of scope

- Cross-space *aggregation* (`status --all`) — separate task
- Any TUI switching

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [2026-08-15-multi-space-home-registry-and-the-atlas](../research/2026-08-15-multi-space-home-registry-and-the-atlas.md)
