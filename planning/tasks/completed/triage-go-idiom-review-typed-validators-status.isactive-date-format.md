---
status: completed
epic: 17-pm-go-cli
description: Adopt the idiom feedback worth taking; reject ctx and DTO-separation as over-engineering for a sync local CLI
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [pm-tooling, go, refactor]
created: "2026-06-09"
completed_at: "2026-06-09"
updated_at: "2026-06-09"
id: 6fakbec00ahf
---

# Triage Go-idiom review: typed validators, Status.IsActive, date format

## Objective

Triage a Go-idioms review against the project's stated philosophy (pragmatic
hexagonal — "don't blindly copy Java/C#", YAGNI). Adopt the genuine improvements;
reject the ceremony with rationale.

## Adopted

- **Typed validators (no string round-trip)** (`domain/validate.go`): `NewTask`
  was doing `strconv.Itoa(tier)` → `ValidateField` → `Atoi` to range-check.
  Added `ValidateTier/Autonomy/Priority/Description/Date` on native types;
  `NewTask`/`NewEpic` call them directly; `ValidateField` (the string `task set`
  path) now delegates to them. Rules live in one place.
- **`Status.IsActive()`** (`domain/status.go`): moved the lifecycle "active"
  invariant out of a private `core` map onto the domain type; `core` calls
  `t.Status.IsActive()`. The invariant is now reusable and lives where it
  belongs.
- **Date format validation** (`domain/lint.go`, `validate.go`): `lint` now flags
  a `created` that isn't `YYYY-MM-DD` (not just empty), and `task set` rejects a
  malformed date field. `created: yesterday` no longer passes.

## Rejected (over-engineering for a sync local CLI — verified against philosophy)

- **`context.Context` on every store/service method**: off the mark here. These
  are synchronous local-fs reads in a short-lived CLI — nothing to cancel, no
  deadline, no request scope; Go's own `os` API takes no ctx. The
  "future-proofing for DB/HTTP/MCP" case is explicitly out-of-scope-by-a-long-
  shot and would thread an ignored param through every layer + test now. It's a
  mechanical add **if/when** a real trigger lands — a TUI background refresh, a
  `watch` mode, or an actual network adapter. Deferred consciously, not ignored.
- **Separate yaml-tagged DTOs in `store` + domain translation**: the reviewer
  concedes it's not worth it for a small CLI, and so does the project. Struct
  tags only affect *reads* (surgical writes use `yaml.Node`, not struct marshal),
  so the coupling is minimal; parallel DTOs + translation would be pure
  boilerplate. Revisit only if the on-disk schema needs to diverge from the
  domain shape (e.g. versioned frontmatter).
- A custom `Date` type (vs the lint/`ValidateField` string check) — same YAGNI;
  a format check covers it without yaml marshal/unmarshal plumbing.

## Acceptance

- [x] Typed validators + `IsActive` + date checks have tests; `ValidateField`
      delegates (no behavior change for `task set`); full suite + lint green.
- [x] Real `planning/` still lints clean (date check found no false positives).

## Related

- Epic [[17-pm-go-cli]]; follows
  [[triage-external-review-fix-fixer-comment-o-excl-config-anchor-diagnostics]].
