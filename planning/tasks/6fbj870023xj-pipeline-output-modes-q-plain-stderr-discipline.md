---
status: completed
epic: 20-cli-ux-and-ergonomics
description: ids-only -q and a stable headered-table --plain (one schema_version) on list commands; move-failure lines to stderr; xargs/jq recipes
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, pipelines, agents]
created: "2026-06-12"
updated_at: "2026-06-19"
started_at: "2026-06-19"
completed_at: "2026-06-19"
id: 6fbj870023xj
---
# Pipeline output modes (`-q`, `--plain`, stderr discipline)

## Decided (2026-06-17)

De-drafted. Resolutions:
- `--plain` is a **headered table** (column-header row, stable columns, no
  ANSI/truncation), covered by the **one global `schema_version`** (D7) — a
  column add/reorder is a schema bump, documented next to the JSON schema; not a
  second version number.
- `-q/--quiet` (ids-only) and the stderr-discipline sweep stand as written.
- The gcloud-style `--format table(col1,col2,…)` column projection is **split
  out** to [[column-projection-format-table-cols-for-list-commands]] (reused by
  the audit-findings query).

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

## Acceptance criteria

- [x] Planning conflicts resolved; de-drafted (see Decided block).
- [x] `task list -q | xargs tskflwctl task complete` round-trips (verified).
- [x] `--plain` is byte-stable TTY/no-TTY (no ANSI, no width truncation — it
      ignores Style entirely) and documented as a contract (README "Pipelines"
      + under the one `schema_version`).
- [x] Move *failures* go to stderr (the called-out gap); list problems already
      did; tested with split stdout/stderr buffers.

## Progress Log

- **2026-06-19**: Shipped. `-q/--quiet` (ids-only) + `--plain` (headered TSV,
  absolute dates) on task/epic/audit `list` via a shared `listMode` (mutually
  exclusive with `--json`; new `render/pipeline.go`). Move failures routed to
  stderr (`MovesHuman` now takes out+errw). README "Pipelines" section (decision
  table + xargs/jq/awk recipes). The `--format table(cols)` projection is its own
  task ([[column-projection-format-table-cols-for-list-commands]]). 5 new tests;
  suite + lint green.

## Related

- Epic [[20-cli-ux-and-ergonomics]] ·
  [[column-projection-format-table-cols-for-list-commands]] ·
  [[2026-06-12-pending-decisions]] (D7).