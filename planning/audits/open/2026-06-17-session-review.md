---
area: session-review
date: "2026-06-17"
---

# Audit: session-review — 2026-06-17

Self-review of the 2026-06-17 tskflwctl work tranche (audited field → TUI audit
bucket-view → `audit new` → `schema` → cross-bucket slug-collision → `task new`
`--body-file`/`--start` → create-envelope status/path → `domain.ParseFindings`).
golangci-lint clean and the full suite green throughout; this captures the
correctness/simplification findings a green build doesn't.

## Coverage

True cross-package coverage (the per-package `render` number is an artifact —
its output funcs are exercised by `cli` tests): **whole tree 84.3%**, domain
90.9%. Closed the real gaps this pass — added `--json` show/transition tests +
`schema` human-mode tests, taking render's true coverage **84.1% → 96.2%**. Only
`Style.Enabled()` remains at 0% — see S1.

## Findings

#### S1. Dead `Style.Enabled()` method  · **Status:** fixed 2026-06-17

**File:** internal/cli/render/style.go:63 | **Component:** render
**Effort:** XS · **Urgency:** soon

`func (s Style) Enabled() bool` has no callers anywhere (incl. tests) — the lone
0%-coverage function in render after this pass. Dead code.

**Recommendation:** remove it. *(Done — deleted in this pass.)*

#### M1. `task new --start` omits the `started_at` stamp  · **Status:** open

**File:** internal/core/service.go:271 | **Component:** core
**Effort:** S · **Urgency:** eventually

`Move` into in-progress stamps `started_at` (`fsstore.go:117`), but `NewTask`
with `--start` creates the file directly in in-progress and stamps only
`created` — so a task *born* in-progress has no `started_at`, unlike one moved
there. Not a lint failure (the field is optional), but an inconsistency any
"time in progress" view would trip on. Stamping needs a small create-frontmatter
addition (`taskFields` is a fixed list; `domain.Task` has no `StartedAt`), so
it's a real change with a design choice attached.

**Recommendation:** either stamp `started_at` on `--start`, or decide a
`--start` task's `created` *is* its start and document it. Owner's call.

#### L1. `ParseFindings` mis-read a literal `**Status:**` in a title  · **Status:** fixed 2026-06-17

**File:** internal/domain/finding.go | **Component:** domain
**Effort:** XS · **Urgency:** soon

`field()` matched the FIRST `**Status:**` anywhere in a finding's section, so a
finding whose title or prose contains a literal `**Status:**` had that grabbed as
its status (and `stripInlineStatus` truncated the title there). **This audit
demonstrated it:** with L1's original title, `audit list` reported **3 open
instead of 4** — the parser read this finding's `· **Status:** open` as a garbage
token because the literal marker in the title matched first.

**Recommendation (done):** anchor the status marker to line-start or the header's
`· ` separator, and key title-stripping on `· **Status:**`. Fixed + table-tested
(`TestParseFindings_LiteralStatusInTitle`).

#### L2. `--body-file` only on `task new`  · **Status:** fixed 2026-06-17

**File:** internal/cli/audit.go, internal/cli/epic.go | **Component:** cli
**Effort:** S · **Urgency:** soon

`task new` gained `--body-file <path|->`, but `audit new` (which had `--body`)
and `epic new` (which had neither) didn't — a consistency gap.

**Recommendation (done):** added `--body-file` to `audit new`, and `--body` +
`--body-file` to `epic new`, all through the shared `resolveBody`. Tested.

#### L3. `audit new` scaffold hardcodes `../HOWTO-execute.md`  · **Status:** open

**File:** internal/core/service.go (auditBodyTemplate) | **Component:** core
**Effort:** XS · **Urgency:** eventually

The scaffold links `[../HOWTO-execute.md]`, which exists in desirelines-planning
but **not** in taskflow's own `planning/audits/` — so this very audit's link is
dead (caught by dogfooding `audit new` here). The convention doc is a per-repo
assumption baked into a shared scaffold.

**Recommendation:** make the link conditional (only when the file exists / a
config points at it), or add a taskflow `planning/audits/HOWTO-execute.md`, or
drop the line from the scaffold.

## Candidate tasks

- ✅ S1 — dead `Style.Enabled()` removed this pass; no task needed.
- ✅ L1 — parser robustness fixed this pass (surfaced by dogfooding this audit).
- ✅ L2 — `--body-file` extended to `audit new`/`epic new` this pass.
- ⏳ M1 — `task new --start` + `started_at`: decide stamp-vs-document, then
  `tskflwctl task new "Stamp started_at on task new --start" --epic 17-pm-go-cli --tags cli,core`.
- ⏳ L3 — low; the audit-scaffold `../HOWTO-execute.md` link, decide when next
  touching the scaffold.
