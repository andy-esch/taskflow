---
status: completed
epic: 17-pm-go-cli
description: tskflwctls strict YAML rejects unquoted colon values the no-PyYAML pm wrote; quote in pm serialize + one-time backfill before tskflwctl reads a repo.
effort: Unknown
tier: 2
priority: high
autonomy_level: 3
tags: [pm-tooling, migration, yaml, data-integrity]
created: 2026-06-07
updated_at: "2026-06-09"
completed_at: "2026-06-09"
---

# Strict-YAML frontmatter: fix pm serialize quoting + backfill invalid values

## Objective

The no-PyYAML Python `pm` wrote non-conformant YAML (unquoted-colon values,
comma-string lists) that tskflwctl's strict YAML rejects. Make tskflwctl read
such repos and provide a clean migration.

## Resolution (2026-06-09)

Functionally satisfied — the goal (tskflwctl reads pm repos; invalid values get
backfilled; new writes are clean) is met, and the "fix pm serialize" sub-goal is
moot now that pm is retired:

- **Resilient reads + actionable errors:** unreadable files are skipped and
  reported with a field-level fix hint (`store/diagnose.go`), not a fatal abort.
- **One-shot backfill:** `tskflwctl lint --fix` quotes unquoted-colon values and
  normalizes comma-string lists to YAML lists (text-level, works on files that
  don't even parse) — the "backfill before tskflwctl reads a repo" step.
- **Conformant at the source:** tskflwctl *writes* valid YAML (surgical
  `updateFrontmatter` + `createFileAtomic` quote correctly), so nothing new is
  malformed.
- **"Fix pm serialize" is obsolete:** pm is retired for the daily loop; we no
  longer write through it.

Remaining is operational, not a tooling gap: running `tskflwctl lint --fix -C
../desirelines-planning` to clean that legacy repo is the user's call (separate
pm-managed repo).
