---
status: completed
epic: 20-cli-ux-and-ergonomics
description: One -o/--output {human,json,name,table} flag + a completable -c/--columns projection flag; --json/-q stay as aliases
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, output, dx]
created: "2026-06-19"
updated_at: "2026-06-19"
started_at: "2026-06-19"
completed_at: "2026-06-19"
id: 6fdtbb401htr
---
## Objective

The list output surface is fragmenting: `--json` (persistent), `-q`, `--plain`,
and a proposed `--columns`/`--format` are four flags answering one question —
"what shape is the output?" — plus a mutual-exclusion matrix between them, and it
only grows (csv, etc.). Consolidate the **format** axis into a single
`-o/--output` flag and keep **projection** as one orthogonal, completable
`-c/--columns` flag. This supersedes [column-projection-format-table-cols-for-list-commands](6fd5r5c009vr-column-projection-format-table-cols-for-list-commands.md)
(projection becomes `-c`, not a `table(...)` value DSL). Only `--json` has
shipped, so this is a clean replace — `-q`/`--plain` never released.

Design rationale and the research behind it (kubectl `custom-columns=` is not
completable; `=`/parens inside a flag value break bash COMP_WORDBREAKS) live in
the session that produced this task; the short version is below.

## Design

**Scope decision:** `-o`/`-c` are **list-local** flags (on `task`/`epic`/`audit
list`), not persistent. `name`/`table` are inherently list concepts, and `--json`
already gives *every* command the universal human/json split — a persistent `-o`
would force all ~10 non-list commands to accept `name`/`table` only to reject
them, for ~zero gain. `--json` stays the universal selector; `-o json`/`-o human`
work on list commands for symmetry. Promoting `-o` to persistent later is additive.

**Format axis — `-o, --output FORMAT` (list-local, default `human`):**

- `human` — colored, aligned, truncated (the only colored format; the default)
- `json`  — the stable envelope; works on **every** command
- `name`  — ids only, one per line (the old `-q` behavior)
- `table` — headered TSV, all default columns (the old `--plain` behavior)
- `csv`   — headered RFC 4180 comma-separated (for spreadsheets; cells with
  commas/quotes/newlines are quoted via `encoding/csv`)

**Projection axis — `-c, --columns a,b,c`:**

- Its own completable token (space-separated from the flag, comma-joined). No
  parens, no value-internal `=`, no `table(...)` DSL.
- Implies `-o table` when `--output` is unset (`task list -c slug,status` just
  works).
- Compatible with the columnar formats (`table`, `csv`); with `human`/`json`/
  `name` → exit 11 naming the formats it applies to. (Relaxing to allow `json`
  projection later is non-breaking: an error becomes a success.)

**Back-compat aliases:** `--json` → `-o json`, `-q/--quiet` → `-o name`. Both
stay visible/supported (not deprecated). Reconciled in `resolve()` using
`cmd.Flags().Changed("output")` to tell an explicit `-o` from the default:
`--json` + an explicit `-o table` → exit 11; `--json` + default `-o` → json.

**One column registry per entity** (`render`): an ordered
`[]Column{Name, Header, Extract func(item) string}` that is the single source of
truth for (1) the `table` default columns, (2) `-c` validation, (3) `-c`
completion candidates, (4) the projection itself. Lift today's hardcoded
`TasksPlain`/`EpicsPlain`/`AuditsPlain` columns into it.

**Completion (the payoff of the split):** `RegisterFlagCompletionFunc` on both
flags. `-c` uses the incremental comma-list idiom — split `toComplete` on the
last comma, dedup already-chosen columns, return `prefix+col` candidates with
`ShellCompDirectiveNoSpace | NoFileComp | KeepOrder`, each with a `\t`-separated
description. `-o` completes the four format names with descriptions.

**Deliberately not** a `pflag.Value` enum for `-o` (it would error with cobra's
exit 1, not our exit 11) — validate the value in `resolve()` instead. Comment
this so it isn't "fixed" later.

## Acceptance criteria

- [ ] `-o/--output` resolves `human|json|name|table|csv`; unknown value → exit 11
      listing valid formats.
- [ ] `--json` and `-q` behave exactly as before (aliases); `--json` + an
      explicit conflicting `-o` → exit 11.
- [ ] `-c/--columns` projects the columnar formats (`table`/`csv`); implies
      `-o table` when format unset; `-c` with `human`/`json`/`name` → exit 11.
- [ ] `csv` is RFC 4180 (comma cells quoted), header-only on empty, byte-stable.
- [ ] Column names validated against the per-entity registry; unknown column →
      exit 11.
- [ ] `-o <TAB>` completes the five formats; `-c slug,sta<TAB>` completes column
      names, dedups already-chosen ones, leaves the cursor mid-token (NoSpace).
- [ ] Completion is unit-tested via the hidden `__complete` command (formats and
      columns), so it can't silently rot.
- [ ] `-o table` stays byte-stable / ANSI-free under `--color=always` (existing
      contract test still green).
- [ ] README/help and `schema` reflect the new flags; `--plain` is never
      documented as a shipped flag.

## Out of scope

- `json` projection (slim envelopes) and a `yaml` format — the design leaves room
  for both, but they're not shipped here. (`csv` *was* folded in — it reused the
  registry + `-c` almost for free.)
- Filtering/sorting/transforms in the output flags (the gcloud `--format`
  DSL-creep that the split exists to avoid).
- The readiness axis and interactive pickers (separate epic-20 tasks).

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Supersedes [column-projection-format-table-cols-for-list-commands](6fd5r5c009vr-column-projection-format-table-cols-for-list-commands.md)
- Builds on the shipped-but-unreleased `-q`/`--plain` work
  ([pipeline-output-modes-q-plain-stderr-discipline](6fbj870023xj-pipeline-output-modes-q-plain-stderr-discipline.md)) and its `renderList`
  helper.
