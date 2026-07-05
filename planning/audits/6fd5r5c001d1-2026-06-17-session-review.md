---
area: session-review
date: "2026-06-17"
id: 6fd5r5c001d1
bucket: closed
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

#### M1. `task new --start` omits the `started_at` stamp  · **Status:** fixed 2026-06-17

**File:** internal/core/service.go, internal/store/create.go | **Component:** core
**Effort:** S · **Urgency:** soon

`Move` into in-progress stamps `started_at`, but `NewTask --start` created the
file directly in in-progress stamping only `created` — so a task *born*
in-progress had no `started_at`, unlike one moved there.

**Recommendation (done — owner chose "stamp"):** added `StartedAt` to
`domain.Task` (the one lifecycle stamp a create can carry), `NewTask` sets it to
today when status is in-progress, and `taskFields` appends it when set. So "every
in-progress task has a `started_at`" now holds however it got there. Tested.

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

#### L3. `audit new` scaffold hardcodes `../HOWTO-execute.md`  · **Status:** fixed 2026-06-17

**File:** internal/core/service.go (auditBodyTemplate) | **Component:** core
**Effort:** XS · **Urgency:** eventually

The scaffold linked `[../HOWTO-execute.md]`, which exists in desirelines-planning
but **not** in taskflow's own `planning/audits/` — so this very audit's link was
dead (caught by dogfooding `audit new` here). A per-repo assumption baked into a
shared scaffold.

**Recommendation (done — owner chose "drop"):** removed the link; the scaffold is
generic now (kept the "flip `**Status:**`" guidance). A repo with a conventions
doc points at it from its own tooling. `TestAuditNew` asserts the absence.

#### L4. `task new --start`/`--next` could create a lint-failing task  · **Status:** fixed 2026-06-17

**File:** internal/core/service.go | **Component:** core
**Effort:** XS · **Urgency:** soon

`task new` didn't require `--description`, but lint requires one for
`next-up`/`in-progress` — so `--next`/`--start` without it created a task born
immediately lint-failing. Pre-existing for `--next`; `--start` extended it.
Surfaced dogfooding M1.

**Recommendation (done — owner chose "require"):** `NewTask` now rejects a
`--next`/`--start` create with no `--description` (ErrValidation, exit 11), the
same principle as the existing required-tags check. README example + tests
updated.

## Candidate tasks

- ✅ S1 — dead `Style.Enabled()` removed.
- ✅ L1 — parser robustness fixed (surfaced by dogfooding this audit).
- ✅ L2 — `--body-file` extended to `audit new`/`epic new`.
- ✅ M1 — `started_at` now stamped on `task new --start` (owner: stamp).
- ✅ L3 — repo-specific HOWTO link dropped from the scaffold (owner: drop).
- ✅ L4 — `task new --next`/`--start` now requires `--description` (owner: require).

## Closeout

Closed-ready 2026-06-17: all six findings terminal — S1/L1/L2/M1/L3/L4 fixed,
none carried forward. The dogfood paid off (L1 + L3 + L4 were surfaced *by*
writing/working this audit with the tool). Run `tskflwctl audit close
2026-06-17-session-review` to archive.
