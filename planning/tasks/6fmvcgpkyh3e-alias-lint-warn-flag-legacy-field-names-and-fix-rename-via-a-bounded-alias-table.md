---
schema: 1
id: 6fmvcgpkyh3e
status: ready-to-start
epic: 26-frontmatter-schema-declared-validation-contract
description: Flag legacy/misspelled frontmatter keys (deprecated_date→deprecated_at) in lint, rename under --fix via a frozen alias table; needs a raw-keys→lint path shared with epic 26 field checks.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [lint, schema, frontmatter, core]
created: "2026-07-10"
updated_at: "2026-07-10"
---
# Alias lint-warn: flag legacy field names + `--fix` rename via a bounded alias table

Implements epic 26's **Q3** decision (auto-migrate legacy field names under
`--fix`, bounded to a fixed alias table) and the **Q2** "misspelled known field"
path. Gated on ADR-0005 ratifying the two foundations below.

## The gap (researched 2026-07-10)

A legacy/misspelled key like `deprecated_date` (vs recognized `deprecated_at`)
today **sails through lint silently**: it's preserved in the file text but dropped
on unmarshal, and the reader uses `deprecated_at` — so the value is effectively
ignored with no signal. `KnownTaskField` exists (`internal/domain/fields.go`) and
`task set --set` already rejects unknown keys, but `lint` has no equivalent guard.

## Why it's a real (small) feature, not a bolt-on

`Service.Lint` (`internal/core/service.go:216`) operates on parsed `Task`/`Epic`
**structs**, which have already dropped unknown keys. So there is **no current
path that surfaces raw frontmatter keys to the reporter** — only the `--fix` text
walk (`fixFrontmatterText`, `internal/store/fix.go`) sees them. Flagging an alias
needs a new "raw keys → lint" path, which is the **same foundation every future
epic-26 field check needs** — so build it once, deliberately, not bolted on.

## Proposed change

1. **Alias table (domain).** Add `fieldAliases map[string]string` beside
   `taskFields` in `internal/domain/fields.go` (seed: `deprecated_date →
   deprecated_at`) + an `AliasFor(key) (canonical string, ok bool)` accessor.
   Frozen/append-only, like the field registry.
2. **Raw-key surfacing (the foundation).** Give lint access to each file's raw
   top-level frontmatter keys. Two candidate shapes — pick in review:
   - carry a `RawKeys []string` (or an `Unknown`/alias slice) off the store read
     alongside the parsed struct, or
   - a store-side scan that returns alias findings as `FileProblem`s (an alias is
     a file-level defect; `FileProblem` is the existing vehicle).
3. **Detect (domain).** `AliasIssues(rawKeys []string) []Issue` → "legacy field
   `deprecated_date` — use `deprecated_at` (`lint --fix` renames it)". Wire into
   `Service.Lint`; coordinate with the audit-lint roster task (`6fm8p1cj11qf`) so
   audits get it too.
4. **Repair (`--fix`).** Rename alias keys at the text level in `fixFrontmatterText`
   (it already iterates raw `key: value` lines) — a bounded key rewrite that
   preserves value, comments, and key order. Only for keys in the alias table.

## Severity — sidestep Q6

Ride the **existing binary lint** as a *fixable issue* (same class as
`MissingIDIssue`/`IDDriftIssue`: hard-but-`--fix`-repairs), so this lands without
waiting on Q6's `error|warn|info` decision. If Q6 later adds a `warn` tier, demote
then. (A silently-ignored field is arguably a real defect, so fixable-error is
defensible now.)

## Acceptance criteria

- [ ] `tskflwctl lint` flags a task/audit whose frontmatter carries a known alias
      key (e.g. `deprecated_date`), naming the canonical field.
- [ ] `lint --fix` renames the alias to its canonical key, preserving value,
      comments, and key order; idempotent on a second run.
- [ ] `lint --json` includes the alias finding.
- [ ] The alias table is a single frozen source of truth in `domain/fields.go`.
- [ ] Tests: a file with an alias is flagged, `--fix` repairs it, and a file with
      only canonical keys is untouched.

## Out of scope / deferred

- General unknown-field policy (closed registry vs `x-*` escape hatch) — that's
  Q2, decided in ADR-0005; this task only handles keys **in the alias table**.
- A `migrate` verb / migration framework — parked (ADR-0005 Q10, YAGNI at
  `schema: 1`).

## Related

- Epic [26-frontmatter-schema-declared-validation-contract](../epics/26-frontmatter-schema-declared-validation-contract.md) — Q2/Q3 own this; gated on ADR-0005.
- ADR survey task `6fkkz41cax80` — Q3 (auto-migrate, bounded) / Q2 (misspelled-known → alias) notes.
- Audit-lint roster task `6fm8p1cj11qf` — share the raw-key path so audits get alias checks too.
- Prior art: `KnownTaskField` / `taskFields` (`domain/fields.go`), `fixFrontmatterText` (`store/fix.go`), `Service.Lint` (`core/service.go:216`).
