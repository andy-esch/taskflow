---
area: codebase (cli, core, store, domain, tui, theme, render)
date: "2026-06-13"
id: 6fbwhsw01mm7
---

# Audit: code quality & architecture — tskflwctl

Full-codebase audit of `internal/` + `cmd/` (~7.9k non-test LOC), run as five
parallel package-scoped reviews (architecture conformance; store+domain;
core+config; cli+render; tui+theme). Every finding below cites real code; the
two High findings and the one inter-reviewer conflict were re-verified by hand.

**Caveat — branch is mid-flight.** This was audited on `refactor/tui-polish-batch`
with uncommitted work in progress (a `dryRun` parameter being threaded through the
mutators). The half-threaded `dryRun` is **not** reported as a finding. A couple of
findings touch code that is actively changing — flagged inline where relevant.

## Progress log

- **2026-06-13 (1)** — Fixed and regression-tested: **H1** (unset gate), **H2** (rune
  count), **M1** (audit-bucket sentinel), **L1** (slug over-trim), **L3** (regular-
  file scan guard, via a shared `markdownDoc` predicate), **L5** (single exit-code
  table), **L9** (stale docs). Deferred the rest pending the in-flight `dryRun` merge.
- **2026-06-13 (2)** — `dryRun` work merged (PRs #4/#5), so cleared the unblocked
  batch: **M2** (lint flag shadow), **M4** (fix `--json` carries `unreadable`), **M5**
  (help-scroll clamp), **L7** (`dry_run` always-present). Schema bumped 1.2 → 1.3. The
  remaining refactors/decisions (**M3**, **L2**, **L4**, **L6**, **L8**, and optional
  **M6**) are tracked in task
  `address-deferred-code-audit-findings-numbering-dedup-cas-json-layout` under epic
  17. **10 of 17 findings fixed**; full suite green (9/9), `go vet` clean.
- **2026-06-13 (3)** — Closing the audit: every finding is now either fixed inline or
  handed off to the tracking task above, so nothing is left dangling. The audit's job
  — surfacing the findings — is complete; the remaining work lives in the task and is
  scheduled there, not here.

## Verdict

The codebase is **architecturally clean and unusually disciplined**. The package
import graph has zero layering violations (`domain` is pure; `core` imports only
`domain` + pure stdlib — no `os`, no `cobra`; `tui` never imports `store`), DI is
textbook (one `*App`, no globals), all output flows through injected writers, and
`--json`/`schema_version` is consistent. There are **no consistent Go-idiom
violations** — the code reads like one careful author wrote it. Findings are a
short list of correctness bugs (two genuine, both narrow) plus consistency/drift
items. Nothing here is a release blocker.

## Findings

#### H1. `task set --unset <key>` bypasses the unknown-field gate
**Status:** fixed (2026-06-13)
**Severity:** High · **File:** `internal/core/service.go:106-115`

*Resolution:* the unset branch now applies the same `force`/`KnownTaskField` gate as
the set path (via a shared `unknownFieldErr` helper); regression test
`TestSetFields_UnsetRejectsUnknownField`.


The unset branch guards only `status`/`updated_at`, then `continue`s — it never
reaches the `if !force && !domain.KnownTaskField(field)` check at line 116 that the
assignment branch enforces:

```go
if _, unset := val.(domain.UnsetField); unset {
    switch field { case "status": ...; case "updated_at": ... }
    withMeta[field] = val
    continue          // <-- skips the KnownTaskField gate below
}
if !force && !domain.KnownTaskField(field) { ...reject... }
```

The CLI routes any `--unset <key>` straight here (`cli/task.go:180-184`) with no
compensating check. So `tskflwctl task set foo --unset descriptionn` (typo) reports
a clean "updated" without `--force`, and the user believes they cleared
`description` when they didn't — a silent no-op on a destructive operation, and an
asymmetry that makes the unset path weaker than the documented contract ("a typo'd
field name must not silently persist"). **Verified by hand.**

*Fix:* apply the same `KnownTaskField`/`force` check inside the unset branch before
`withMeta[field] = val` (keep the two hard-rejected keys as special cases above it).

#### H2. `lint` measures description length in bytes; `task new`/`set` measures runes
**Status:** fixed (2026-06-13)
**Severity:** High · **File:** `internal/domain/lint.go:64` vs `internal/domain/validate.go:49`

*Resolution:* `LintTask` now counts `utf8.RuneCountInString`, matching
`ValidateDescription`; regression test `TestLintTask_DescriptionLengthInRunes`.


Two enforcement paths for the same documented cap (`MaxDescriptionLen = 150`, whose
comment says "in characters") disagree on non-ASCII input:

```go
// lint.go:64   — bytes
case len(t.Description) > MaxDescriptionLen:
// validate.go:49 — runes
if n := utf8.RuneCountInString(d); n > MaxDescriptionLen {
```

A description of 100 multibyte characters (~250 bytes) passes creation but `lint`
reports "too long (250 > 150)" — and there's no `--fix` for it, so `planning/`
can't be kept lint-clean. This is exactly the drift the typed-validator unification
exists to prevent. **Verified by hand.**

*Fix:* use `utf8.RuneCountInString(t.Description)` in `LintTask`, or factor both
call sites onto one shared helper.

#### M1. `ParseAuditBucket` doesn't wrap `ErrValidation`
**Status:** fixed (2026-06-13)
**Severity:** Medium · **File:** `internal/domain/audit.go:28`

*Resolution:* now wraps `ErrValidation` (→ exit 11), matching `ParseStatus`;
regression test `TestParseAuditBucket_InvalidWrapsValidation`.


```go
return "", fmt.Errorf("invalid audit bucket %q (open|closed|deferred)", s)
```

`ParseStatus` and `ValidateEpicStatus` both wrap `ErrValidation`; this doesn't. Per
CLAUDE.md, sentinels drive exit-code mapping (validation → 11), so a bad audit
bucket escapes as a generic exit 1 — inconsistent with `task`/`epic`. `MoveAudit`
happens to re-wrap `ErrValidation` itself, masking it there, but any direct caller
of `ParseAuditBucket`/`AuditBucket.Valid()` is exposed.

*Fix:* `fmt.Errorf("%w: invalid audit bucket %q (...)", ErrValidation, s)`.

#### M2. `lint` defines a local `--dry-run` that shadows the persistent one
**Status:** fixed (2026-06-13)
**Severity:** Medium · **File:** `internal/cli/lint.go:13,29` vs `root.go:74`

*Resolution:* removed the local `dryRun` var/flag; `lint --fix` now reads the
persistent `app.DryRun` like every other mutating command. `lint --fix --dry-run`
still works via the persistent flag (covered by `TestLintFix_DryRunThenFix`).


`lint` is the only command that re-declares `--dry-run` locally (`var fix, dryRun bool`
+ `BoolVar(&dryRun, "dry-run", …)`) while the root already provides a persistent
`--dry-run` bound to `app.DryRun`. The RunE papers over it with
`runLintFix(app, dryRun || app.DryRun)`, but `lint --help` now lists `--dry-run`
twice, and the `|| app.DryRun` rescue is a smell hiding the duplicate definition.
Every other mutating command reads `app.DryRun` directly. *(Touches in-flight dryRun
work — fold into it.)*

*Fix:* delete the local var + `BoolVar`; use `app.DryRun` throughout `lint.go`.

#### M3. Auto epic-numbering can produce duplicates; `%02d` mis-sorts past 99
**Status:** deferred → tracked in task `address-deferred-code-audit-findings-numbering-dedup-cas-json-layout`
**Severity:** Medium · **File:** `internal/store/create.go:89-108,123-153`

`nextEpicNumber` computes `max(prefix)+1` via `ReadDir`, then `CreateEpic` relies on
`createFileAtomic`'s `O_EXCL` to prevent clobber — but `O_EXCL` only guards an
identical *path*. Two `epic new` calls with different titles compute the same number,
write different filenames, and both succeed → `03-foo.md` and `03-bar.md` with no
error (the number is the user-facing fuzzy-resolve key). Separately, `%02d` zero-pad
means `100-x` sorts lexically before `99-x` once you pass 99 epics.

*Fix:* detect existing `NN-` prefix collisions and bump (retry loop), and widen or
stop zero-padding the number.

#### M4. JSON consumers lose the post-`--fix` residual lint problems
**Status:** fixed (2026-06-13)
**Severity:** Medium · **File:** `internal/cli/lint.go:76-85`

*Resolution:* `FixJSON` now carries an `unreadable` array, and `runLintFix` emits one
combined envelope (fixed + unrepairable) under `--json` instead of only a prose
count; schema bumped to 1.3. Regression test `TestLintFix_JSONReportsUnreadable`.


After `FixJSON` writes its report, the re-lint of still-broken files renders only via
`if !app.JSON { render.ProblemsHuman(...) }`; under `--json` the unrepairable files
survive only as a *count* in the stderr error envelope. Every other JSON command
emits a structured `unreadable` list — this drops it, violating the "agents on
`--json` never parse prose" principle.

*Fix:* when `app.JSON`, emit the residual `problems` as JSON instead of only the
human render.

#### M5. Help-overlay scroll clamp uses the wrong upper bound
**Status:** fixed (2026-06-13)
**Severity:** Medium · **File:** `internal/tui/model.go:291-292`

*Resolution:* the `j` handler now clamps to a new `helpMaxScroll()` (mirroring
`helpBox`'s window math: `len(helpLines()) - innerH`) instead of the total line
count, so scrolling stops at the visible bottom. Regression test
`TestModel_HelpScrollClampedToVisibleMax`.


The model lets `helpScroll` grow up to `len(helpLines())` (total lines), but the
renderer clamps the *visible* offset to `len(lines)-innerH` and never writes back.
Once scrolled to the visual bottom, every further `j` still increments `helpScroll`
(silently, up to ~35) while the view stays put — so the user must press `k` that
many times before upward scrolling visibly resumes. The overlay feels frozen. Low
blast radius (help modal only), but a real UX bug.

*Fix:* clamp against the true max in the model
(`maxScroll := max(len(helpLines())-innerH, 0)`), or have `helpBox` return the
clamped value to store.

#### M6. Mutators thread `force`/`dryRun` as trailing positional booleans
**Status:** deferred → tracked in task `address-deferred-code-audit-findings-numbering-dedup-cas-json-layout`
**Severity:** Medium · **File:** `internal/core/store.go:21-48`, `service.go:100`

Every mutator on the `Store` port and `Service` ends in bare `bool`(s); `SetFields`
ends in two adjacent same-typed positionals (`force, dryRun`). Call sites like
`app.Svc.SetFields(args[0], updates, force, app.DryRun)` have no compile-time guard
against transposing them, and the in-flight work is adding more of this threading.
One trailing flag is a defensible Go idiom; two adjacent bools is the danger point.

*Fix:* before a third bool lands, introduce a small `MutateOpts{DryRun, Force bool}`
value (or functional options) and document the convention on the port. Pragmatic:
acceptable for one flag, fix the `force, dryRun` adjacency.

#### L1. `Slugify` truncation can over-trim a long-first-word title to a stub
**Status:** fixed (2026-06-13)
**Severity:** Low · **File:** `internal/domain/slug.go:96-111`

*Resolution:* the dash-backup now only applies when the dash is past the midpoint of
the cut, so a short-first-word title keeps a usable rune-boundary cut instead of
collapsing; regression test `TestSlugify_LongFirstWordNotOverTrimmed`.


After the ≤80-byte rune-boundary cut, the "back up to the previous `-`" step takes
the *last* dash in the kept prefix. When the title is a short word followed by one
very long unbroken token, the only early dash is near the front, so the slug
collapses: `"ab " + <90-char word>` → `"ab"`. Real but narrow (normal
space-separated titles have a dash near byte 80, making the backup minor), and
`core` only rejects a *fully* empty slug. *(Note: the related em-dash-leaks-into-slug
bug is already FIXED on this branch — `slug.go` was rewritten 2026-06-13 to an
allowlist; verified it no longer reproduces.)*

*Fix:* skip the dash-backup when it would drop more than ~half the cap; add a
long-single-word case to `slug_test.go` (that path is currently untested).

#### L2. Triple-duplicated list/resolve/scan scaffolding across the three stores
**Status:** deferred → tracked in task `address-deferred-code-audit-findings-numbering-dedup-cas-json-layout`
**Severity:** Low · **File:** `fsstore.go:44-74`, `epicstore.go:16-43`, `auditstore.go:29-59` (+ candidate gatherers)

Each `List*` repeats the same `ReadDir → IsNotExist continue → skip dir/non-.md →
ReadFile → parse → FileProblem` loop, and the three have already drifted subtly
(e.g. `ListEpics` returns `nil` problems on a missing dir while task/audit `continue`
per-bucket). Any change to the skip rules (hidden files, symlinks — see L3) must be
made in 3+ places.

*Fix:* extract a generic `scanDir`/`mdEntries` helper (generics are available).

#### L3. `.md` directory entries are matched without a regular-file check
**Status:** fixed (2026-06-13)
**Severity:** Low · **File:** `fsstore.go:57`, `epicstore.go:27`, `auditstore.go:42`, `create.go:98`, `fix.go:30`

*Resolution:* all seven scan sites now route through one `markdownDoc(e)` predicate
(in `resolve.go`) requiring `e.Type().IsRegular()`, so symlinks are skipped;
regression test `TestFS_ListTasks_SkipsSymlinkedMarkdown`. (Also partially addresses
L2 by giving the scans one shared filter.)


The filter is only `e.IsDir() || !strings.HasSuffix(e.Name(), ".md")`. A symlink
`x.md` pointing outside the tree passes, and `os.ReadFile`/`writeFileAtomic`/`fix.go`
follow it. `validQueryName` carefully blocks `..`/separators in *queries*, so
traversal is in the threat model — but a planted symlink is an escape on the write
side. Low (requires existing write access to a status dir).

*Fix:* `if !e.Type().IsRegular() { continue }` (also subsumes the `IsDir` check).

#### L4. `Move` has no compare-and-swap guard, unlike `SetFields`
**Status:** deferred → tracked in task `address-deferred-code-audit-findings-numbering-dedup-cas-json-layout`
**Severity:** Low · **File:** `internal/store/fsstore.go:97-162`

`Move` does resolve → read → `writeFileAtomic(newPath)` → `os.Remove(path)` with no
re-resolve. If a concurrent op relocates/deletes the file between resolve and remove,
the new file is already written and `os.Remove` fails *after* creation — leaving a
recoverable duplicate in two status dirs. `SetFields` has a CAS re-resolve (and a
harden test); `Move` doesn't. Lower severity: it errors rather than silently losing
data, and concurrency on a local CLI is rare.

*Fix:* apply the same CAS re-resolve, or document that `Move` isn't concurrency-safe.

#### L5. Parallel exit-code switches (`ExitCode` / `errorCodeName`) can drift
**Status:** fixed (2026-06-13)
**Severity:** Low · **File:** `internal/cli/exit.go:15-31,35-48`

*Resolution:* both functions now read one `errCodes` table (sentinel → code → name),
so the code and its machine name can't drift; covered by existing `TestExitCode`.


The integer→code and integer→name mappings are two independent `switch` statements
over the same sentinels. Adding a sentinel requires editing both in sync or
`WriteError` emits `code:"error"` for a non-1 exit.

*Fix:* one `[]struct{err error; code int; name string}` table both functions consult.

#### L6. `epic` JSON shape differs between `epic list` and `epic show`
**Status:** deferred → tracked in task `address-deferred-code-audit-findings-numbering-dedup-cas-json-layout`
**Severity:** Low · **File:** `render/render.go:348-358` (`epicJSON`) vs `:534-541` (`epicMetaJSON`)

`epic list` emits `total/done/percent`; `epic show` (`epicMetaJSON`) omits the
rollup. Two `epic` shapes under one `schema_version` weakens the "whole CLI output
schema" claim. Defensible (show returns `tasks[]` separately), but worth aligning.

*Fix:* embed `epicMetaJSON` in `epicJSON` so meta fields are guaranteed identical and
rollup is purely additive; or document the split.

#### L7. `dry_run` JSON field: `omitempty` in some envelopes, always-present in others
**Status:** fixed (2026-06-13)
**Severity:** Low · **File:** `render/render.go:184` (omitempty) vs `:491` (no omitempty)

*Resolution:* dropped `omitempty` from `MovesJSON`/`CreatedJSON`/`InitJSON`, so
`dry_run` is always present (`false` explicit) on every mutation envelope; schema
bumped to 1.3.


`MovesJSON`/`CreatedJSON`/`InitJSON` use `dry_run,omitempty`; `FixJSON` always emits
it. A consumer checking `payload.dry_run` gets present-or-absent depending on the
command. *(Touches in-flight dryRun work.)*

*Fix:* one policy — recommend always-present (`dry_run:false` explicit) for mutation
results.

#### L8. TUI watcher and shell-completion duplicate the store's directory layout
**Status:** deferred → tracked in task `address-deferred-code-audit-findings-numbering-dedup-cas-json-layout`
**Severity:** Low · **File:** `internal/tui/watch.go:44-55`, `internal/cli/completion.go:56-118`

Both reconstruct `tasks/<status>` / `audits/<bucket>` / `epics/` directly via
`filepath.Join` rather than through the store. Both are deliberate and well-reasoned
(the watcher is watch-only and reloads through `core.Service`; completion must work
without a built service and parse no YAML), so this is drift-risk, not a layering
break — if the on-disk layout ever changes, these go stale silently.

*Fix:* expose the canonical watchable-dir set from the store
(`store.WatchDirs(root)`), shared by both. Low priority.

#### L9. Stale doc references
**Status:** partially fixed (2026-06-13)
**Severity:** Low · **Files:** `internal/tui/model.go:36`, `CLAUDE.md` (sentinel list), `render.go` `CreatedJSON:318-332`

*Resolution:* (a) `Model.root` comment corrected, and (b) `ErrInvalidTransition`
dropped from CLAUDE.md's sentinel list (now notes 12 is retired/reserved). (c) The
`CreatedJSON` named-struct nit is **deferred** — `render.go` is in the mid-flight
change set, so left untouched to avoid colliding with concurrent work.


Three minor doc/readability nits: (a) `Model.root` is commented "not read yet" but it
*is* passed to `newWatcher` (`tui.go:15`) — a future reader could wrongly delete it;
(b) CLAUDE.md still lists `ErrInvalidTransition` among wrappable sentinels, but it was
deliberately retired (`domain/errors.go`); (c) `CreatedJSON` re-declares its inner
anonymous struct literal verbatim instead of a named type.

*Fix:* correct the `root` comment, drop `ErrInvalidTransition` from CLAUDE.md, name
the `createdItem` struct.

## What's genuinely solid (verified, not assumed)

- **Architecture.** `go list` confirms a clean import graph: `domain` pure, `core`
  free of `os`/`cobra`, `tui` free of `store`, single `*App` DI populated in
  `PersistentPreRunE`, zero `fmt.Print*` outside the `cmd/` composition root. The
  ports-and-adapters design holds across the board.
- **`store/atomic.go`** — temp-stage → fsync → chmod → rename with cleanup on every
  error branch, `O_EXCL` for exclusive create, best-effort `syncDir`; the O_EXCL-over-
  hardlink choice is documented for container mounts.
- **Surgical frontmatter editing** (`frontmatter.go`) — carries head/line/foot
  comments onto replaced nodes, preserves key order and unknown fields, refuses to
  overwrite a non-mapping; the unterminated-fence handling closes a real corruption
  vector and is tested.
- **`SetFields` parse-before-commit + CAS** (`fsstore.go:184-201`) is the strongest
  correctness work in the repo, with a dedicated harden test.
- **TUI concurrency** — every async load is generation-stamped *and* kind-tagged with
  explicit stale-drop guards; `routeToTab` recurses into `tea.BatchMsg` so a
  background tab can't pollute the active one. Glamour rendering is correctly cached
  (rebuilt only on width change), never invoked from `View`. No I/O in `Update`/`View`.
- **render package** (~580 LOC) is flat-but-disciplined, NOT accreting unmanageably:
  one Human + one JSON fn per entity, all JSON through `encodeJSON`, all tables through
  one ANSI-aware `writeTable` (`ansi.StringWidth`/`ansi.Truncate` — wide-rune and
  escape-safe). Exit-code mapping uses `errors.Is` (catches wrapping); stdout/stderr
  discipline is clean throughout.
- **Slugify allowlist** (rewritten 2026-06-13) and the `domain/fields.go` single-source
  type registry both reflect a deliberate "stop patching denylists one bug at a time"
  philosophy that's paying off.

## Suggested follow-up tasks

All quick/correctness findings are fixed inline (see the progress log). The remaining
open findings — **M3**, **L2**, **L4**, **L6**, **L8**, and optional **M6** — are
larger refactors or design calls, tracked together in
`planning/tasks/ready-to-start/address-deferred-code-audit-findings-numbering-dedup-cas-json-layout.md`
(epic 17). Close this audit once that task lands; nothing here is urgent.

## Related

- `docs/ARCHITECTURE.md` — the ports-and-adapters contract this audit checked against
- Audited at commit `f88bf41` on branch `refactor/tui-polish-batch` (working tree
  mid-flight; see caveat above)
