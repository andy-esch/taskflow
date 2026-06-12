---
status: ready-to-start
epic: 17-pm-go-cli
description: 'DRAFT: ids-only -q and stable --plain (TSV, absolute dates) on list commands, move-failure lines to stderr, xargs recipes; needs plan review'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, pipelines, agents, draft]
created: "2026-06-12"
---
# Pipeline output modes (`-q`, `--plain`, stderr discipline)

> 🚧 **DRAFT — not yet integrated into the overall plan.** Filed from the
> 2026-06-12 CLI-design discussion. Needs a planning pass on the conflicts
> below; defer to existing/completed contract decisions where they disagree.

## Objective

Make classic Unix pipelines first-class alongside `--json`+jq. Human tables
are for eyes; these are for pipes.

1. **`-q/--quiet` on list commands** (task/epic/audit list): ids only, one
   per line — unlocks `task list -q --tag tui | xargs tskflwctl task start`.
   Highest composability-per-line-of-code item.
2. **`--plain`:** documented-STABLE machine-text mode — TSV columns,
   absolute dates (never "yesterday"), no truncation/padding/ANSI. The
   git-porcelain concept: scripts may rely on it across versions.
3. **Stderr discipline sweep:** per-item transition failure lines (`✘ …`)
   currently go to stdout; diagnostics belong on stderr uniformly (lint's
   problems were already moved). Audit every command against the rule:
   stdout = data, stderr = diagnostics/prompts.
4. **README "pipelines" section:** xargs recipes, jq recipes, the
   -q/--plain/--json decision table. Prefer documenting `xargs` over
   building stdin-slug plumbing (revisit only if real friction shows).

## ⚠️ Conflicts to resolve before starting

- **`--plain` becomes a second versioned contract** next to schema 1.1
  (decision D7 chose ONE global schema version) — decide whether --plain
  stability is covered by the same version or documented separately, and
  record it where SchemaVersion is documented.
- Item 3 changes observable behavior of the just-shipped transition summary
  ([[json-and-output-contract-fidelity]], completed) — coordinate so the
  smoke test and any user scripts are updated deliberately, not silently.
- A future `events` NDJSON change-stream (noted in the design discussion as
  a differentiator, NOT filed) would build on the watcher port work in
  [[put-storage-layout-knowledge-back-behind-the-port]] — keep this task's
  scope clear of it.

## Acceptance criteria (draft)

- [ ] Planning conflicts above resolved; task de-drafted.
- [ ] `task list -q | xargs tskflwctl task complete` round-trips.
- [ ] `--plain` output is byte-stable under TTY/no-TTY and documented as a
      contract.
- [ ] No data on stderr, no diagnostics on stdout, anywhere (test sweep).

## Related

- Epic [[17-pm-go-cli]] · [[agent-facing-cli-ergonomics-batch]] ·
  [[2026-06-12-pending-decisions]] (D7).