---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Extend -c/--columns projection to the --json format and make --json compact (no indent; NDJSON for lists) — the biggest agent token win.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [cli, json, agent]
created: "2026-06-22"
updated_at: "2026-06-22"
started_at: "2026-06-22"
completed_at: "2026-06-22"
id: 6fes83r03bbm
---
# Honor `-c/--columns` and compact output for `--json`

**Source.** Product feedback (2026-06-22) from an AI agent actively driving
`tskflwctl`. Its headline token-waste finding: the *discoverable default*
(`--json`) is the *most expensive* output — prominently advertised, but it emits
**pretty-printed (2-space-indented) full frontmatter**, while field projection
(`-o table -c`) is buried in `--help` examples and doesn't apply to JSON at all.
An agent pattern-matches "machine-readable ⇒ `--json`" and pays for every
`tags[]`, `tier`, `priority`, `autonomy_level`, `created`, `updated`… on every
row when it wanted three fields.

This builds directly on the completed
[[consolidate-output-flags-into-output-and-columns]], which gave the list
commands `-o/--output` + a completable `-c/--columns` projection — but only for
`table`/`csv`. This task extends that projection to the format agents actually
parse, and stops pretty-printing for machines.

## Scope

1. **Honor `-c/--columns` for `--json`** (today table/csv only). `--json -c
   slug,status,description` is called out as "the single biggest token win" —
   field projection on the format agents consume. Same completable column
   registry as table/csv.
2. **Compact `--json`**: drop the 2-space indentation for machine output; emit
   **NDJSON** for list commands (one object per line) so consumers can stream
   rows. Keep the envelope's `schema_version` contract intact.

## Acceptance criteria

- [ ] `task list --epic X --json -c slug,status,description` emits only those
      fields, per row.
- [ ] `--json` output is compact (no indentation); list commands emit NDJSON.
- [ ] The `--json` envelope contract (`schema_version`, exit codes) is preserved.
- [ ] Column registry / completion shared with the existing table/csv projection.
- [ ] Suite + lint green; docs/cli regenerated (docs-check gate).

## Related

- Extends [[consolidate-output-flags-into-output-and-columns]] (table/csv `-c`).
- Epic [[20-cli-ux-and-ergonomics]] (output modes, column projection).

## Resolution (2026-06-22)

Shipped: `-c/--columns` now projects `--json` (column-named string fields, in `-c` order, schema_version-first envelope, `unreadable` omitempty), and ALL `--json` output (incl. the error envelope) is compact.

**Deviation from scope — NDJSON dropped on purpose.** Per-line bare rows can't carry the versioned envelope, and `--json everywhere with a schema_version` is a repo non-negotiable. Compact single-envelope delivers the token win AND keeps the contract + published JSON schema intact. NDJSON would be a schema-breaking change for a separate discussion.

Column registry stays the single source of truth (table/csv/json share it). Goldens + docs/cli regenerated.
