---
schema: 1
area: code-quality-architecture
date: "2026-06-22"
id: 6fes83r00wnj
---

# Audit: code quality & architecture — adversarial sweep (2026-06-22)

> Edit findings in place and flip each `**Status:**` as you work it.

Full-codebase adversarial audit of `internal/` + `cmd/` (~21k non-test LOC). Run
as 11 parallel finders (7 package-scoped + 4 cross-cutting: architecture,
concurrency, error-handling, testing). Every finding was then re-read against the
real code by an independent skeptic on a separate model (try-to-refute stance):
**53 raw → 47 deduped → 45 stood, 2 refuted.** 44 are recorded below.

**Baseline is green:** `go vet`, full `go test ./...`, and `golangci-lint`
(standard set) all pass clean — so everything here is what tooling cannot catch:
design boundaries, invariants, concurrency, edge cases, and growth risk.

## Progress log

- **2026-06-22** — Knocked out the quick-win batch (11 findings), each with a
  regression test; `go build` / `go vet` / `go test ./...` / `golangci-lint` all
  green. Fixed: **H1** (`errBadFrontmatter` now wraps `ErrValidation` → uniform exit
  11 on read/move), **H2** (shared `domain.ActiveTaskFieldErr`, enforced on both
  `NewTask` and the `SetFields` write path), **H3** (`newWatcher` errors when zero
  dirs are watchable → honest `watchOff`), **H4** (`task edit --dry-run` rejected,
  not silently ignored), **M2** (`writeFileAtomic` preserves an existing file's
  mode), **M3** (`status` returns `problemsError` → non-zero on unreadable files),
  **M5** (`MoveAudit` compare-and-swap + test hook → `ErrConflict` on concurrent
  relocation), **M13** (fuzz targets now assert body-preserved / re-parses /
  key-applied on cleanly-parsing input), **M15** (`s/S/o/O/F` gated on list focus so
  they can't wipe the detail pane), **L4** (closing fence tolerates trailing
  whitespace), **L10** (`IntFields`/`ListFields` unexported behind
  `IsIntField`/`IsListField`). **11 of 44 fixed.**
- **Deferred: M16** (O(N²) audit rescan) — the clean fix needs a new read-by-path
  store-port method plus `fakeStore` rework, i.e. a port-interface change, not a
  quick win; left open. All other findings (H5, the architecture/growth themes, and
  the remaining medium/low items) are still open.
- **2026-06-22 (2)** — Second tranche (6 findings), each with a regression test;
  suite + vet + `golangci-lint` green. Fixed: **H5** (post-move reload no longer
  flashes a spurious "<slug> not found" over the success — `movedAway` guard),
  **M4** (`MoveAudit` now refuses to close/defer an audit with open findings,
  matching `audit lint`), **L3** (`applyView` resets the filter like `jumpTo`),
  **L9** (nav back-stack capped + skips consecutive duplicates), **L13** (`runMoves`
  prefers a sentinel-bearing exit code over a generic one, not argv order), **L21**
  (clarified `repoColorScheme`'s intentionally-ignored `LightDarkFunc`). **17 of 44
  fixed.** The remaining 27 are the architecture/growth themes + lower-priority
  localized items — the substantial ones are now tracked as tasks under epic
  `21-code-quality-architecture-hardening`; the rest stay here as the backlog.
- **2026-06-22 (3)** — Third tranche (6 findings); suite + vet + `golangci-lint`
  green. Fixed: **L5** (`SetFields` blames a pre-existing-corrupt file rather than the
  user's update), **L7** (`FixFrontmatter` returns partial results on a write error;
  `runLintFix` reports them), **L11** (shared `App.startDir()` for resolve/completion
  discovery), **L15** (`WriteCSV` neutralizes formula-injection cells), **L22**
  (documented the untagged-message tab-routing invariant), **L23** (task-creation
  defaults applied in `NewTask`, not just CLI flags). L5/L15/L23 carry regression
  tests; L11 is a behavior-preserving refactor; L22 is doc-only. **23 of 44 fixed.**
  Remaining 21 = the 16 in the epic-21 tasks + 5 minor localized items (L2, L6, L14,
  L16, L20) left here as backlog.
- **2026-06-22 (4)** — Fixed **M7** (`writeTable` now clamps each composed human-table
  line to the terminal width with an ANSI-aware `ansi.Truncate`, backstopping the
  last-column shrink so a wide non-final cell — slug/component/id — can't wrap; machine
  formats untouched; tests in `render/style_test.go`). **M1 → in-progress**: the entity
  descriptor (epic 21 task, now complete) collapsed the DOMAIN fan-out, but the
  render/TUI fan-out remains — tracked by **M9** (god-file split) and **M10** (TUI
  registry), so the broad M1 theme stays in-progress. **24 of 44 fixed; 1 in-progress.**
- **2026-06-22 (5)** — Config-robustness tranche (`internal/config/config.go`): **L8**
  (containment now resolves symlinks via `EvalSymlinks` before the `Rel` no-`..` check,
  so a `planning -> /outside` symlink can't slip past), **L19** (`Discover` resolves the
  start dir so the `.git`-boundary walk-up uses physical ancestry), **L17** (the one-key
  TOML scanner now refuses a backslash-bearing basic string rather than mis-decoding
  `"a\"b"` as `a\`; literal `'...'` strings unaffected). New symlink-escape /
  symlinked-worktree / TOML-escape tests; existing Root comparisons made symlink-safe
  for macOS temp dirs. **27 of 44 fixed; 1 in-progress.**
- **2026-06-22 (6)** — God-file split (**M9** + **L1**, the same epic-21 task) landed as
  a PURE, zero-behavior-change file move within each package: `render.go`'s JSON DTOs +
  mappers → `dto.go`, the schema-contract types/funcs → `schema_render.go` (render.go now
  holds only the generic + list/show renderers); `core/service.go`'s per-entity use-cases →
  `service_task.go`/`service_epic.go`/`service_audit.go` (service.go keeps the `Service`
  facade, `NewService`/`WatchPaths`, `Summary`/`Lint`, and shared helpers). No symbol
  renamed, no casing changed, no exported surface or `core.Store` port touched. Proven by a
  before/after output snapshot (54-command battery, color forced) that diffed EMPTY, plus
  goldens passing without `-update`, `schema_comments.json` regenerating with no drift, and
  docs/cli unchanged; `go build`/`go vet`/`gofmt`/`golangci-lint`/full `go test ./...` all
  green. **29 of 44 fixed; 1 in-progress.**
- **2026-06-22 (7)** — Two port/boundary clusters, each with regression tests; suite +
  vet + `golangci-lint` green, docs/cli no-drift. **core.Store port reshape:** **M16**
  (new `GetAuditByPath` accessor on the audit port → `QueryFindings`/`LintAudits` read
  each audit by the path `ListAudits` already resolved, killing the O(N²) re-resolve and
  the concurrent-edit window; `fakeStore`/`nopStore` gained the method, finding seeds
  carry `.Path`), **L12** (split `FixFrontmatter`→`core.Fixer` and `WatchPaths`→
  `core.Layout` off the use-case `Store`; the CLI wires `app.Fixer` and the TUI takes a
  `core.Layout` directly — `*FS` satisfies all three via compile assertions; the fakes
  shed two methods), **M8** (ARCHITECTURE.md re-justified: render imports ~5 core types
  not "two", and the port-purity theme reflects the done Fixer/Layout split). Paired with
  the epic-22 **template port** (a `core.TemplateSource` behind `Service.ListTemplates`/
  `ShowTemplate` + the create paths; `template list/show` resolve repo-best-effort so
  built-ins still work repo-less — readying step 4's repo-local layering as a source
  swap). **TUI cluster:** **M6** (cursor-restore carried per-message + gen-stamped, and a
  reload mid-jump carries the jump target forward — no more single-slot steal/false
  "not found"), **M14** (a `modal` overlay interface + ordered registry the reducer
  loops, ForceQuit handled once in the preamble — a new overlay is one entry, no new
  guard block/`bodyView` case), **M10** (the `a` menu + `:` verbs are registry-driven
  off each entity's transition table + `applyMove`, so audits now close/reopen/defer
  in-TUI; the M4 open-findings guard surfaces as a red flash). **35 of 44 fixed; 1
  in-progress (M1).** Remaining: M11, M12 + the low backlog (L2, L6, L14, L16, L18, L20).
- **2026-06-22 (8)** — Closed the medium correctness backlog + the M1 theme; each with
  tests, suite/vet/`golangci-lint` green, docs/cli no-drift. **M12** (`NewRootCmd(in, out,
  errOut)` is now the single stdin owner — it feeds both `app.In` (gate/prompter/editor)
  and the cobra root via `SetIn`, so `resolveBody`'s `InOrStdin` and the prompter read
  one injected reader; ~44 call sites swept, no stray `os.Stdin` left in cli). **M11 +
  L18** (duplicate-slug DETECTION in `core.Lint`: a Ctrl-C in Move's write-then-remove
  window leaves the same slug in two dirs → `lint` now flags it loudly and exits 11
  naming both dirs, so the otherwise-silent permanent `ErrAmbiguous` is hand-repairable.
  No auto-DELETE: the Move-crash duplicate has BOTH copies matching their folders — an
  ambiguous tie the "never lose data" stance says report-don't-guess, and the task's
  Done-when explicitly allows "reported when ambiguous." Plus L18: `createFileAtomic`'s
  Close path now cleans up like its siblings, negative atomic-write tests (read-only dir,
  skip-as-root), and a conservative `.tskflwctl-*.tmp` orphan sweep on `lint --fix`,
  age+prefix-guarded). **M1** (descriptor + generic seams were the recommendation's
  substance and they're in place; ARCHITECTURE.md now states the honest residual cost of
  a typed multi-entity domain instead of underselling it — the recommendation's stated
  minimum bar). Also added a transition-table validity test guarding M10's stringly-typed
  `to`. **39 of 44 fixed.** Remaining 5 = the low backlog L2 (watcher backoff), L6
  (realign dry-run preview), L14 (lazy glamour), L16 (find across wrap — documented
  tradeoff), L20 (selectByID scan) — all low-value polish or accept-as-documented.
- **2026-06-22 (9)** — Closed out the low backlog and the audit. Fixed **L16**
  (`findStatus` now hints `R: raw (match spans a wrap)` when a pretty-mode search
  finds 0 but the raw render would match — so a wrap-straddling hit isn't read as
  "not present"; the same fold-aware matcher decides, so the hint only fires when R
  actually helps). Accepted (wontfix) the rest as low-value/by-design: **L2** (the
  fsnotify error→reload nudge is *correct* for the dominant error, kernel queue
  overflow → resync; the 200ms debounce already prevents a busy-loop, and flipping
  to `watchOff` on a transient error would regress overflow-resync), **L6** (the
  realign dry-run-preview gap is cosmetic — the real `--fix` re-lints and signals),
  **L14** (the glamour render is cached so the cost is bounded "wasted work"; the
  lazy-render change risks the find/cache state machine for no real gain), **L20**
  (`selectByID` is ≤2 scans per reload — negligible until thousands of items).
  **40 of 44 fixed; 4 accepted (wontfix); 0 open.** Audit closed.

## Verdict

Healthy, well-architected project — **no critical bugs, nothing that corrupts
data or crashes in normal use.** Hexagonal boundaries are real, atomic writes are
disciplined, the read path degrades gracefully. The dominant risk is **not bugs —
it's contract integrity and entity fan-out**, both of which bite precisely as the
project grows. Severity spread: 0 critical · 5 high · 16 medium · 23 low.

## Architecture & growth themes

These cross-cutting patterns matter more than any single line item, and are the
answer to "what will inhibit sustainable growth."

1. **Contract integrity (high).** The tool's value proposition (scriptable,
   agent-driven) rests on stable contracts enforced *inconsistently*. The same
   condition behaves differently per command: semantic exit codes (malformed
   frontmatter → 11 on `list`/`set`, 1 on `show`/`complete`/`move` — **H1**);
   `--dry-run` honored everywhere except `task edit` (**H4**); domain invariants
   enforced in `NewTask` but dropped in `SetFields`/`MoveAudit` (**H2**, **M4**).
   This class of defect *multiplies per new command/entity*. Highest-leverage fix:
   centralize each contract at its narrowest seam.

2. **Entity-add fan-out (medium, growth) — RESOLVED (2026-06-22).** Was a ~15-file
   shotgun edit (**M1**). The data-driven entity `Descriptor` now collapses the
   metadata fan-out (schema/scaffold/conventions), the generic seams (`scanDir[T]`,
   `resolveID`, `Column[T]`/`WriteTablePlain`) carry the mechanics, and TUI lifecycle
   is registry-driven (**M10**). The remaining per-entity surface — a typed `domain`
   struct + `parse*`, thin `*Store`/`Service` wiring, a cli command, and render/TUI
   display delegates — is the irreducible cost of a *typed* domain with three
   genuinely different shapes (status/tier vs rollups vs findings); ARCHITECTURE.md
   now documents it honestly. A full reflection/codegen collapse is deliberately not
   pursued for three not-yet-doubled entities.

3. **God-files growing per entity × use-case (medium, growth).** `render.go`
   (703 LOC, **M9**), `core/service.go` (~693 LOC, **L1**), and `model.go`'s
   `handleKey` (~160-line reducer, **M14**) all grow monotonically and are the
   merge-conflict epicenters of the future.

4. **TUI shared single-slot state → reload/jump/detail races (medium).** State
   that should be generation-stamped lives in single mutable slots shared across
   independent triggers (`tab.restore` — **M6**; the post-move flash — **H5**).
   Message-ordering hazards (Update is serial, not data races) that tangle further
   as reload triggers multiply.

5. **Doc/justification drift (medium).** ARCHITECTURE.md green-lights boundaries
   with justifications that no longer hold: render imports 5 core types not the
   documented "two" (**M8**); the TUI registry's "no new keybindings" is false for
   any lifecycle-bearing entity (**M10**). Documented ceilings already exceeded.

6. **Port purity leak (low).** `FixFrontmatter`/`WatchPaths` are fs/text
   operations sitting on the use-case `core.Store` port (**L12**) — split into a
   `Fixer`/`Layout` interface before a second entity adds its own fix rules.

## Findings

### High (5)

#### H1. Malformed frontmatter exits 1, not 11, on read/move paths  · **Status:** fixed (2026-06-22)

**File:** internal/store/fsstore.go:87,110,127,260 | **Component:** store
**Effort:** S · **Urgency:** soon

`errBadFrontmatter` is a package-local sentinel unrelated to `domain.ErrValidation`
on the `Get*`/`Move` paths, so `ExitCode` falls to generic 1. The same broken file
yields exit 11 from `task list`/`set`/`append` but exit 1 from
`task show`/`complete`/`move` and `epic show` — breaking the semantic exit-code
contract agents route on. The write paths (`SetFields`/`EditBody`) already classify
correctly; the read paths don't.

**Recommendation:** Wrap `errBadFrontmatter` with `domain.ErrValidation` at its
return sites (or classify via `errors.Is` in the `Get*`/`Move` paths). Collapses
every malformed-file outcome to a uniform 11. *(quick win)*

#### H2. SetFields writes a file its own linter rejects (empty tags / no description on active tasks)  · **Status:** fixed (2026-06-22)

**File:** internal/core/service.go:129-177 | **Component:** core
**Effort:** S · **Urgency:** soon

`ValidateField` has no `tags` case and `ValidateDescription` doesn't reject empty,
so `task set` can drive an active (next-up/in-progress) task into a state
`LintTask` immediately flags. `NewTask` guards these explicitly; `SetFields`
silently drops them — the invariant is enforced on one write path and dropped on
the other, with no compensating post-write lint and no test coverage.

**Recommendation:** Factor the active-task field rules (non-empty tags, required
description) into one domain function both `NewTask` and `SetFields` call, so the
write paths cannot diverge. *(quick win)*

#### H3. newWatcher reports success even when every directory Add fails  · **Status:** fixed (2026-06-22)

**File:** internal/tui/watch.go:28-37 | **Component:** tui
**Effort:** S · **Urgency:** soon

`newWatcher` swallows every `fsw.Add` failure and returns a non-nil watcher with
`err==nil` even when zero adds succeeded (inotify ENOSPC, FUSE/overlay mount, all
leaf dirs absent). `tui.go` then leaves `m.watchOff=false`, so the footer shows
normal hints while `Init` waits on a watcher that will never deliver — defeating
the explicit "so the degradation isn't silent" guarantee.

**Recommendation:** Count successful `Add` calls; if none succeeded, return an
error/sentinel so `Run` takes the `watchOff` branch and the footer tells the truth.
Keep ignoring individual missing-dir failures. *(quick win)*

#### H4. `task edit --dry-run` ignores the flag and writes to disk  · **Status:** fixed (2026-06-22)

**File:** internal/cli/edit.go:33-68 | **Component:** cli
**Effort:** S · **Urgency:** soon

`--dry-run` is documented as "preview the mutation without writing" and `edit` is
annotated mutating, but `RunE` never reads `app.DryRun`; `EditTask`/`store.EditTask`
carry no `dryRun` param and `writeFileAtomic` runs unconditionally. Every other
mutator threads the flag. `task edit --dry-run <task>` opens a full editor session
whose save lands on disk — accept-and-ignore of a safety flag, worse than rejecting.

**Recommendation:** Thread `dryRun` through `EditTask` (run editor, parse-validate,
skip the write, report "would update"), or reject `edit --dry-run` with
`ErrValidation`. *(quick win)*

#### H5. A successful lifecycle move flashes a spurious "<slug> not found" error  · **Status:** fixed (2026-06-22)

**File:** internal/tui/model.go:146-152,86-96,260-274 | **Component:** tui
**Effort:** M · **Urgency:** soon

Completing/deferring/deprecating from the default active view (the common case)
moves the task out of the list; `markReload` captures the moved id into
`tab.restore`, then after reload `selectByID` fails and the unfiltered branch
overwrites the green "moved → completed" flash with a red "<slug> not found". The
mutation succeeds but every successful archiving action reports as failure,
undermining trust in the TUI's only write surface.

**Recommendation:** Don't treat a moved id's absence as a dangling reference: skip
the active tab's restore-by-moved-id for the post-move reload, or suppress the
"not found" flash when a fresh success flash already exists.

### Medium (16)

#### M1. Adding a new entity type is a ~15-file shotgun edit across 6 packages  · **Status:** fixed (2026-06-22) — descriptor landed; residual documented honestly

**File:** internal/core/store.go:15-65 | **Component:** architecture
**Effort:** L · **Urgency:** eventually

`ProjectsDir` is scaffolded but lighting up a project/adr entity requires
coordinated edits across domain (struct/layout/schema/AuthoringFields/Conventions/
SchemaKinds), `core/store.go` ports, store, `core/service.go`, scaffold, cli
command + root wiring, render (human+json+columns+envelope), and tui (entity/
detail/commands/item). Each layer re-enumerates the entity set — the per-entity
fan-out a layered design is meant to localize. ARCHITECTURE.md frames it as cheaper
than it is.

**Recommendation:** Introduce an entity descriptor (dir name, field order, parse/
serialize, columns) to drive the already-generic machinery (`scanDir`, `Column[T]`,
`resolveID`, `writeNewFile`) from data; at minimum correct the doc.

**Resolution (2026-06-22).** The metadata fan-out — the recommendation's high-leverage
half — is collapsed into `entity.go`'s `Descriptor` (dir/fields/conventions/scaffold/
placeholders drive `SchemaKinds`/`AuthoringFields`/`Conventions`/`BodyTemplate`, no
`switch kind`). TUI lifecycle is now registry-driven too (M10). The generic seams the
recommendation named are in place: `scanDir[T]`, `resolveID`, and `Column[T]` +
`WriteTablePlain[T]`/`WriteCSV[T]`. The genuine residual for a new entity is a typed
`domain` struct + `parse*`, thin `*Store` port methods, `core.Service` use cases, a cli
command, and per-entity render/TUI *display* delegates — the irreducible cost of a
**typed** domain whose three kinds have different shapes (status/tier vs rollups vs
findings/buckets); the generics remove the mechanics, not the shape. A full data-driven
persistence/render collapse is deliberately NOT pursued: for three heterogeneous,
not-yet-doubled entities it trades clarity for reflection/codegen machinery. ARCHITECTURE.md
now states this cost honestly (no longer "cheaper than it is"), which was the
recommendation's stated minimum bar. Reopen if the entity count actually doubles.

#### M2. writeFileAtomic silently resets file mode to 0644 on every edit  · **Status:** fixed (2026-06-22)

**File:** internal/store/atomic.go:12-53 | **Component:** store
**Effort:** S · **Urgency:** soon

`stageTemp` chmods the temp to the passed perm (0o644, hardcoded by all five
callers) without consulting the existing file mode. A file the user chmod'd to
0600/0444 is silently reset to world-readable 0644 on the next mutation;
`createFileAtomic` respects umask, so a file created under umask 077 starts 0600
then jumps to 0644 on first edit — surprising, undocumented permission widening on
the source-of-truth files.

**Recommendation:** Stat the destination when it exists and reuse its `FileMode`
for the temp chmod; fall back to 0o644 only when the file is absent. *(quick win)*

#### M3. `status` returns exit 0 even when files are unreadable  · **Status:** fixed (2026-06-22)

**File:** internal/cli/status.go:16-26 | **Component:** cli
**Effort:** XS · **Urgency:** soon

`Summary()` collects `Problems` and the render funcs embed an `unreadable` array
but both return nil — no `problemsError` call. An agent running `status --json` to
gate on repo health gets exit 0 plus a non-empty `unreadable` array, contradicting
the `list`/`lint` contract for the identical condition — the silent-empty/forked-
tree failure class the codebase elsewhere works to make loud.

**Recommendation:** Return `problemsError(s.Problems)` from status (consistent with
list commands), or explicitly document status as glance-only in help and schema.
*(quick win)*

#### M4. MoveAudit can close/defer an audit with open findings  · **Status:** fixed (2026-06-22)

**File:** internal/core/service.go:668-671 | **Component:** core
**Effort:** S · **Urgency:** soon

`Service.MoveAudit` is a pass-through; `store.MoveAudit` only checks `to.Valid()`
and never reads the findings body. `LintFindings` flags any non-open-bucket audit
with open findings, but that rule runs only in `LintAudits`, never on the write
path — the same broken-symmetry class as the tags-on-set gap (H2). A store test even
asserts `MoveAudit` succeeds on an open-findings audit.

**Recommendation:** In `MoveAudit`, when transitioning to a non-open bucket, parse
the body and reject (`ErrValidation`/`ErrConflict`) if `CountOpenFindings>0`, or
require `--force`. Reuse `domain.CountOpenFindings` so write and lint share one rule.

#### M5. MoveAudit lacks the compare-and-swap re-resolve every task mutator has  · **Status:** fixed (2026-06-22)

**File:** internal/store/auditstore.go:54-90 | **Component:** store
**Effort:** M · **Urgency:** soon

A single `resolveAudit` precedes ReadFile/parse/os.Rename with no intermediate
re-resolve, unlike `Move`/`SetFields`/`EditBody` which re-resolve and map a mismatch
to `ErrConflict`. A concurrent relocation between resolve and rename fails as a plain
`move audit: %w` → exit 1 instead of the exit-14 retry signal — an inconsistent
concurrency contract. No test hook or concurrency test exists.

**Recommendation:** Add the same pre-rename re-resolve CAS and a test hook, mapping
a mismatch to `ErrConflict` with a retry message. *(quick win)*

#### M6. Per-tab cursor-restore (tab.restore) is a single slot shared by reload + jump  · **Status:** fixed (2026-06-22)

**File:** internal/tui/entity.go:63,90-95 | **Component:** tui
**Effort:** M · **Urgency:** soon

`tab.restore` (one string per tab) is written by both `markReload` (reload) and
`jumpTo` (navigation) and consumed asynchronously after refilter. If `jumpTo` sets
`restore=<target>` and an fsnotify debounce reload calls `markReload` before the
jump's load lands, the target is silently lost, the cursor restores to the wrong row,
and the "not found" flash can fire against the wrong id. The `loadGen` guard protects
the list result but not the restore slot.

**Recommendation:** Make restore intent explicit and generation-stamped: carry the
target id + `loadGen` in the load result/Cmd rather than a mutable tab field.

#### M7. writeTable only shrinks the last column, so a wide non-final cell overflows  · **Status:** fixed (2026-06-22)

**File:** internal/cli/render/style.go:177-199 | **Component:** render
**Effort:** M · **Urgency:** eventually

`used` accumulates raw widths for every non-last column and only `width[last]` is
capped to `avail`; no non-last column is ever truncated. A long slug/epic-id/
component in a narrow terminal pushes the row past `maxWidth` and wraps.
`FindingsHuman` is most exposed (six columns before TITLE, free-text Component).
Machine formats are uncapped and unaffected, so the schema contract is safe — human
output only.

**Recommendation:** When fixed columns already overflow, elide lower-priority middle
columns or clamp the composed line with `ansi.Truncate`, mirroring the TUI clamp
discipline; add a no-line-exceeds-maxWidth test.

#### M8. The cli→render→core diamond is ~5 core types, not the doc's "two"  · **Status:** fixed (2026-06-22)

**File:** internal/cli/render/render.go:208-296,327-359,459-486 | **Component:** architecture
**Effort:** S · **Urgency:** eventually

ARCHITECTURE.md green-lights the boundary by claiming render imports core for "two
view-models"; the non-test render files import 5 (`Summary`, `StatusCount`,
`EpicSummary`, `AuditFinding`, `LintResult`) and grow one-per-entity. The justified
design becomes a finding because the justification is shown false; stats/index/tags
on the roadmap deepen it. (Coherent today — render is the isolation seam, TUI
doesn't touch these — hence medium.)

**Recommendation:** Re-justify the boundary honestly in the doc (count ~5, trending
up per entity), or map core results into render-owned DTOs at the call site as
`taskJSON`/`auditJSON` already do.

#### M9. render.go is a 703-LOC multi-concern god-file growing 4-6 funcs per entity  · **Status:** fixed (2026-06-22)

**File:** internal/cli/render/render.go:1-703 | **Component:** render
**Effort:** M · **Urgency:** eventually

703 lines, 33 exported funcs, plus all JSON DTOs and schema-contract types
co-located. Each entity adds a Human func + JSON func + DTO struct into this one
file, so it grows monotonically and every reviewer/merge touches the same hot file —
the opposite of the by-concern split `envelopes.go`/`columns.go` started. Already
shows copy-drift (`%-12s` vs `%-9s` padding between `TaskShowHuman`/`AuditShowHuman`).

**Recommendation:** Finish the split: move JSON DTOs to `dto.go`, schema types to
`schema.go`, keep `render.go` for generic table renderers. Consider a
field-descriptor list driving `*ShowHuman`.

#### M10. TUI entity-registry's "no new keybindings" claim is false for lifecycle entities  · **Status:** fixed (2026-06-22)

**File:** internal/tui/entity.go:14-23,161-203 | **Component:** architecture
**Effort:** L · **Urgency:** eventually

The registry earns its keep for read/browse, but transitions is a fixed task-only
table, the `a` handler calls `selectedTask()` (no-op on epics/audits), `:`verbs
route through task-only `transitionFor`+`selectedTask`, `applyTransition` calls
task-only `svc.Move`, and `followSelected` switches on task/epic kinds with a
default "no linked entities". Audits already have CLI close/reopen/defer with zero
TUI mutation path. A reader trusting the comment when wiring a lifecycle entity
discovers edits needed in model.go/action.go/nav.go.

**Recommendation:** Lift the action/transition machinery onto an entity's declared
transition table (registry-driven `a` menu + `:` verbs), OR scope the doc's "no new
keybindings" promise to read-only browse.

#### M11. Ctrl-C during Move's write-then-remove window leaves a permanent duplicate slug  · **Status:** fixed (2026-06-22)

**File:** internal/store/fsstore.go:159-172 | **Component:** store
**Effort:** M · **Urgency:** eventually

No SIGINT/SIGTERM handler exists on the verb path. A kill between
`writeFileAtomic(newPath)` and `os.Remove(path)` leaves the same slug in two status
dirs; every later `resolve(slug)` returns `ErrAmbiguous` so the task can't be
shown/moved/set by name, and `lint --fix` has no dedup pass — the comment's
"recoverable duplicate" leans on tooling that doesn't exist. (Window is two adjacent
syscalls, both files intact, error names both locations, recovery is one `rm` —
hence medium, downgraded from high.)

**Recommendation:** Install a SIGINT/SIGTERM guard around the two-step relocation so
the remove always completes, or ship the dedup repair pass; until then document the
kill window honestly.

#### M12. Two divergent stdin sources; NewRootCmd cannot inject stdin  · **Status:** fixed (2026-06-22)

**File:** internal/cli/root.go:60-61 | **Component:** architecture
**Effort:** M · **Urgency:** eventually

`root.go` hard-wires `app.In=os.Stdin` and never calls `SetIn`; `task.go` reads
`cmd.InOrStdin()` for `--body-file -`, while the prompt gate and editor read
`app.In`. One process has two stdin handles. A caller/test injecting via
`cmd.SetIn` feeds `resolveBody` but not prompts/editor. (In real shell use both are
the same `os.Stdin`, so it only bites embedders and interactive-prompt tests — hence
medium.)

**Recommendation:** Pick one stdin owner: add an `io.Reader` param to `NewRootCmd`
(or a `SetIn` on `App` that also updates the cobra root) so `App.In`, the prompt
gate, the editor, and `resolveBody` read one source.

#### M13. Fuzz targets only assert no-panic; the round-trip invariant is never checked  · **Status:** fixed (2026-06-22)

**File:** internal/store/fuzz_test.go:26-39 | **Component:** store
**Effort:** S · **Urgency:** soon

`FuzzSplitFrontmatter`/`FuzzUpdateFrontmatter` discard their outputs (`_,_ = …`), so
a fuzz input that makes `updateFrontmatter` emit invalid YAML or silently truncate
the body passes green. Surgical-frontmatter preservation and "output re-parses" are
stated critical invariants (CLAUDE.md, ARCHITECTURE.md) but are checked only on a
handful of hand-picked unit inputs. (Encode path uses yaml.v3, so probability of a
real corruption is lower — hence medium, a testing-robustness gap not a current bug.)

**Recommendation:** Add property assertions: for `FuzzUpdateFrontmatter`, on
cleanly-parsing input assert the output re-parses as valid YAML, the body is
preserved verbatim, and the requested key has the new value; skip inputs the parser
legitimately rejects. *(quick win)*

#### M14. handleKey is an oversized reducer whose modal precedence stack grows per overlay  · **Status:** fixed (2026-06-22)

**File:** internal/tui/model.go:281-444 | **Component:** architecture
**Effort:** L · **Urgency:** eventually

~160-line guard chain (showHelp, action, follow, cmd, SettingFilter, finding, then
global switch + focus router), with `ForceQuit` re-implemented in 5 places, three
modal bools/structs as direct `Model` fields, and a parallel `bodyView` switch.
Every new modal (peek overlay, confirm, tag picker) means another guard block,
another bool, another `bodyView` case, and another chance to get precedence/
ForceQuit wrong. Functionally correct today; the dispatch layer resists extension.

**Recommendation:** Introduce a modal/overlay interface (`active()`, `handleKey →
(handled, cmd)`, `view(w,h)`) with an ordered stack; `handleKey` loops the stack
then falls through to base routing.

#### M15. Global view/sort/filter keys fire regardless of focus, wiping the detail pane  · **Status:** fixed (2026-06-22)

**File:** internal/tui/model.go:351-404 | **Component:** tui
**Effort:** S · **Urgency:** soon

The global hotkey block has no focus guard: `s/S` (cycleView→applyView), `o/O`,
`F`, `a` all fire in detail focus. The destructive path is `s` → `applyView`, which
sets `focus=focusList`, calls `detail.clear()`, and triggers a reload — so pressing
`s` while reading a task body silently wipes the body and snaps focus back to the
list. The detail-focus footer hint advertises none of these keys.

**Recommendation:** Gate `s/S/o/O/F` (and arguably `a/f`) on `focus==focusList`, or
at minimum document them in the detail-focus hint. *(quick win)*

#### M16. QueryFindings / LintAudits re-resolve + re-read every audit — O(N²) rescan  · **Status:** fixed (2026-06-22)

**File:** internal/core/finding.go:56-69,93-106 | **Component:** core
**Effort:** M · **Urgency:** soon

`ListAudits` already returns audits with `.Path` populated, but the loop calls
`GetAudit(a.Slug)`, which re-resolves across all 3 bucket dirs and re-reads+re-parses
each file: for N audits, 3(N+1) dir scans + 2N file reads. This is the read path
agents hit hardest (`audit findings --status open`) and the TUI runs it on live
reload, scaling quadratically in syscalls. Re-reading also opens a concurrent-edit
window. (Negligible at dozens of audits — hence medium.)

**Recommendation:** Add a read-by-path accessor to the store port (or have
`ListAudits` return bodies in one scan) and read `a.Path` directly instead of
re-resolving `a.Slug`. *(quick win)*

### Low (23)

#### L1. core.Service is a 21-method, ~693-line god-object  · **Status:** fixed (2026-06-22)

**File:** internal/core/service.go:15-25,41-498,663-671 | **Component:** core
**Effort:** M · **Urgency:** eventually

19 methods in service.go + 2 in finding.go, ~693 lines, five entity groupings, many
one-line store pass-throughs, inline body templates. As adr/project/track/stats
arrive it accretes more unrelated clusters. (Downgraded: standard hexagonal facade,
thin delegations are intentional, 4th-largest file, `finding.go` proves it splits
cleanly.)

**Recommendation:** Keep `Service` as the facade but move per-entity use-cases into
`service_task/epic/audit.go`; decide before the adr/project surface doubles methods.

#### L2. fsnotify Errors are treated as reload nudges with no backoff  · **Status:** wontfix (2026-06-22, accepted)

**File:** internal/tui/watch.go:53-66 | **Component:** tui
**Effort:** S · **Urgency:** eventually

Every value on `fsw.Errors` maps to `fsEventMsg` regardless of type. (Downgraded:
the 200ms debounce rate-limits — a continuous error stream fires zero reloads, and
`waitForFS` blocks rather than spinning, so no CPU busy-loop.) The real gap: a
permanently broken watcher never transitions to `watchOff` and may make periodic
futile reloads on a gone directory.

**Recommendation:** Separate the Events and Errors cases: on a real error surface
`watchOff`/a footer note and stop re-listening, or apply backoff before re-arming.

#### L3. applyView (s/S cycle) keeps a stale '/' filter; jumpTo clears it  · **Status:** fixed (2026-06-22)

**File:** internal/tui/model.go:765-775 | **Component:** tui
**Effort:** XS · **Urgency:** eventually

`jumpTo` calls `ResetFilter` ("a jump must not hide the target") but `applyView`
reloads via `SetItems` without resetting, and bubbles/list preserves the filter
across `SetItems`. After `s` with `/foo` applied, the new status view is silently
filtered by foo (chip still reads `filter:foo`), possibly showing an empty view.

**Recommendation:** Pick one policy: also `ResetFilter()` in `applyView`, or document
that view-cycling intentionally preserves the filter.

#### L4. A closing fence with trailing whitespace (`--- `) is unrecognized  · **Status:** fixed (2026-06-22)

**File:** internal/store/frontmatter.go:36-48 | **Component:** store
**Effort:** XS · **Urgency:** soon

The fence comparison only trims `\r`, so `--- ` (trailing space/tab) fails to match
and the file is reported as "no closing ---" even though the fence is present — the
error actively misdescribes the file, and `lint --fix` can't repair it
(`fixFrontmatterText` returns early when no fence is found). A recoverable editor
artifact is refused by every surgical mutator.

**Recommendation:** Trim trailing horizontal whitespace
(`bytes.TrimRight(line, " \t\r")`) when comparing to `---`; add a seed/fuzz case.
*(quick win)*

#### L5. updateFrontmatter only rewrites the first of duplicate keys; cause misattributed  · **Status:** fixed (2026-06-22)

**File:** internal/store/frontmatter.go:186-197 | **Component:** store
**Effort:** S · **Urgency:** eventually

`yaml.Node` parsing accepts duplicate keys; `setMapNode` updates only the first,
re-marshals both, and `parseTask` then rejects the dup. `SetFields` wraps this as
`ErrValidation` "update would not reload", blaming the user's update rather than
pre-existing dup-key corruption — inverting the "don't blame a file that was never
touched" principle. No data is written (safety holds); diagnostic-quality only.

**Recommendation:** When the post-update parse fails but the pre-update parse already
fails the same way, report it via the diagnose path; or detect duplicate top-level
keys up front.

#### L6. FixFrontmatter realignStatus silently no-ops on a misfiled file with a coexisting YAML defect  · **Status:** wontfix (2026-06-22, accepted)

**File:** internal/store/fix.go:73-87 | **Component:** store
**Effort:** S · **Urgency:** eventually

`realignStatus` bails if `yaml.Unmarshal` fails, so a file that is text-fixable AND
has a surviving structural YAML defect AND a misfiled status gets its text fix
reported while the status drift is silently skipped. (Downgraded: the non-dry-run
path re-lints after writing and signals the user; residual gap is the `--dry-run`
preview, which shows the text fix without flagging the skipped realign.)

**Recommendation:** When `realignStatus` declines because frontmatter won't decode,
surface it in the `FixResult` ("status could not be realigned: <reason>").

#### L7. FixFrontmatter aborts the whole run on the first write error, discarding progress  · **Status:** fixed (2026-06-22)

**File:** internal/store/fix.go:19-67 | **Component:** store
**Effort:** S · **Urgency:** eventually

On a mid-run write failure, `fixDir` returns immediately and `FixFrontmatter`
returns `nil,err`, discarding results for files already rewritten in prior status
dirs. The user gets only an error and an empty result set, making the partial
mutation hard to reconcile. Atomic writes prevent half-written files, so this is
observability only — but it contradicts `scanDir`'s resilient read-side philosophy.

**Recommendation:** Return accumulated results alongside the error (or collect
per-file errors and continue).

#### L8. taskflow_root escape guard is purely lexical — a symlink defeats containment  · **Status:** fixed (2026-06-22)

**File:** internal/config/config.go:62-75 | **Component:** config
**Effort:** S · **Urgency:** eventually

`filepath.Rel`/`filepath.Abs` are lexical and never resolve symlinks; a
`taskflow_root` naming a symlink inside `dir` (e.g. `planning -> /etc`) passes the
no-`..` containment check and writes follow the link outside the repo.
(Downgraded from high: exploiting it requires pre-existing repo write access — the
same access needed to edit files directly — and the read path separately rejects
symlink dir-entries.)

**Recommendation:** `filepath.EvalSymlinks` both `dir` and `root` before the `Rel`
check; reject if the evaluated root escapes the evaluated dir. Add a symlinked-root
test.

#### L9. navStack grows unbounded with no cap or cycle detection  · **Status:** fixed (2026-06-22)

**File:** internal/tui/nav.go:135-151 | **Component:** tui
**Effort:** XS · **Urgency:** eventually

`pushLoc` unconditionally appends with no cap and no dedup of consecutive identical
locations; only `ctrl+o` pops. Bouncing between an epic and its tasks grows the slice
for the program's lifetime. (Downgraded: the footer shows depth, the vim-jumplist
design is intentional, and `navLoc` is two strings so memory impact is trivial.)

**Recommendation:** Cap the stack (last N) and/or skip pushing a location identical
to the current top.

#### L10. Domain field registries exported as mutable package-global maps  · **Status:** fixed (2026-06-22)

**File:** internal/domain/fields.go:12-28 | **Component:** domain
**Effort:** S · **Urgency:** eventually

`IntFields`/`ListFields` are exported mutable maps; any sibling package can mutate
the canonical type registry (`domain.ListFields["x"]=true`), corrupting coercion/fix/
diagnose/schema at once with no compile-time guard. The asymmetric `knownTaskFields`
(unexported, accessed via `KnownTaskField()`) shows the right pattern. Read-only
today — a footgun not a live bug.

**Recommendation:** Expose `IsIntField`/`IsListField` accessors like
`KnownTaskField` and unexport the maps. *(quick win)*

#### L11. Planning-root discovery duplicated between resolve() and planningRoot()  · **Status:** fixed (2026-06-22)

**File:** internal/cli/completion.go:49-63 | **Component:** cli
**Effort:** S · **Urgency:** eventually

The "where is the planning repo" start-dir contract (Chdir → Getwd → config.Discover)
is implemented identically in `root.go` (fatal) and `completion.go` (forgiving
`ok=false`). A future change to discovery semantics must be made in both or
completion silently diverges; no test/linter catches divergence.

**Recommendation:** Extract a shared `startDir`/`discoverStart` helper; have both
callers use it with their own error handling.

#### L12. FixFrontmatter sits on the Store port as a leaky, presentation-adjacent operation  · **Status:** fixed (2026-06-22)

**File:** internal/core/store.go:57-59 | **Component:** architecture
**Effort:** M · **Urgency:** eventually

`FixFrontmatter` is an inherently fs/text operation on the use-case-driven `Store`
port (`Service.LintFix` is a pure pass-through), so it bloats the port the core's
purity argument rests on; any second store implementation pays for a method
unrelated to core use-cases. (Downgraded: the doc concedes it as a known wart and
the burden on fakes is one no-op line.)

**Recommendation:** Split `FixFrontmatter` (and possibly `WatchPaths`) into a narrow
`Fixer`/`Layout` interface cli wires directly to the FS.

#### L13. runMoves picks the exit code from whichever failure is first in argv  · **Status:** fixed (2026-06-22)

**File:** internal/cli/moves.go:16-46 | **Component:** cli
**Effort:** S · **Urgency:** eventually

`firstErr` is set only once (first failure) and wrapped for the exit code, so a batch
partial-failure with heterogeneous causes yields a non-deterministic exit code w.r.t.
argument order, and a generic exit-1 error can mask a more meaningful sentinel.
Per-item stderr still shows full detail; only the summarized code is order-dependent.

**Recommendation:** Pick the most actionable sentinel deterministically (prefer
sentinel-bearing errors over generic exit-1), independent of argv order; at minimum
document it.

#### L14. render() recomputes the glamour body on every load even in raw mode  · **Status:** wontfix (2026-06-22, accepted)

**File:** internal/tui/detail.go:84-94 | **Component:** tui
**Effort:** S · **Urgency:** eventually

`render()` unconditionally builds `prettyStyled` via glamour even when
`d.pretty==false`. A user in raw mode arrowing through a long list pays a glamour
render per row; an fsnotify save-storm reload runs glamour for a body nobody sees.
(The renderer is cached so per-call cost is bounded — strictly wasted work.)

**Recommendation:** Render the inactive mode lazily (thunks or a per-mode dirty
flag), materializing the other only on first `toggleMode()`.

#### L15. WriteCSV does not neutralize CSV formula-injection (leading =,+,-,@)  · **Status:** fixed (2026-06-22)

**File:** internal/cli/render/columns.go:90-115 | **Component:** render
**Effort:** S · **Urgency:** eventually

`WriteCSV` delegates to `encoding/csv` (RFC-4180 quoting) but never prefixes cells
beginning with `=,+,-,@` to neutralize spreadsheet formula injection; `FindingColumns`
exposes free-text titles. Low under the local-first threat model, but it undercuts
the "for spreadsheets" claim the moment a repo is shared.

**Recommendation:** If the spreadsheet use case is real, prefix cells whose first
rune is in `{=,+,-,@,\t,\r}` with a leading `'`; otherwise soften the comment.

#### L16. foldMatches/highlightLine do per-line search; '/' find can't match across wrap  · **Status:** fixed (2026-06-22)

**File:** internal/tui/detail.go:272-306 | **Component:** tui
**Effort:** M · **Urgency:** eventually

Matches are computed line-by-line over already-wrapped/glamour-rendered text, so a
query straddling a wrap point yields zero hits; the same query can match in raw mode
but not pretty. Mitigated by help text ("matches the rendered text on screen") and
the R raw fallback — an intentional tradeoff enabling efficient inline highlight.

**Recommendation:** Acceptable as a documented limitation; consider noting in
`findStatus` when 0 matches occur while the raw body contains the query.

#### L17. Hand-rolled taskflow_root TOML parser mis-reads valid escapes  · **Status:** fixed (2026-06-22)

**File:** internal/config/config.go:105-119 | **Component:** config
**Effort:** S · **Urgency:** eventually

`tomlStringValue` scans to the first matching quote with no escape handling, so a
valid basic string like `taskflow_root = "a\"b"` terminates at the escaped quote and
`\\`/`\t` pass through literally — a silent wrong-answer the loud-fail design
elsewhere avoids. Blast radius small (the value is a path; literal single-quote
strings work).

**Recommendation:** Document that only literal/escape-free basic strings are
supported, or decode the one key with a real TOML decoder; at minimum reject basic
strings containing a backslash rather than mis-decoding.

#### L18. Atomic-write helpers leave a .tmp orphan on interruption; cleanup branches untested  · **Status:** fixed (2026-06-22)

**File:** internal/store/atomic.go:12-53,71-91 | **Component:** store
**Effort:** S · **Urgency:** eventually

A SIGKILL between `stageTemp` and `os.Rename` leaves a `.tskflwctl-*.tmp` orphan; the
`.md` filter keeps listings clean but orphans accumulate with no sweep. The crash/
error-rollback guarantees the comments promise are asserted only on the happy path
(stageTemp 50%, writeFileAtomic 62.5%, createFileAtomic 46.7% coverage), and
`createFileAtomic`'s `Close` error path returns without `os.Remove` unlike its
siblings. No corruption.

**Recommendation:** Add a negative test per helper forcing a write/rename failure
(read-only dir via chmod, skip as root); optionally sweep stale temps on startup or
during lint; align `createFileAtomic`'s Close path.

#### L19. Discovery never resolves symlinks; the .git walk-up boundary uses logical paths  · **Status:** fixed (2026-06-22)

**File:** internal/config/config.go:25-55 | **Component:** config
**Effort:** S · **Urgency:** eventually

`filepath.Abs` + lexical `filepath.Dir` climb the logical ancestry.
(Downgraded/overstated: `exists()` uses `os.Stat` which follows symlinks per
component, so the primary claimed mis-termination does NOT occur; the genuine narrow
gap is a symlinked path crossing a repo boundary, which yields "not found" not wrong
discovery.) The test gap is real.

**Recommendation:** `filepath.EvalSymlinks(start)` once at the top of `Discover`
before the climb; add a symlinked-worktree discovery test.

#### L20. selectByID linearly scans VisibleItems during the restore window  · **Status:** wontfix (2026-06-22, accepted)

**File:** internal/tui/entity.go:80-88 | **Component:** tui
**Effort:** S · **Urgency:** eventually

`selectByID` linearly scans `VisibleItems()` (which allocates in the filtered state)
during a pending restore. (Overstated: `SetItems` fires a single one-shot
`filterItems` → one `FilterMatchesMsg`, so the actual call count per reload-with-
filter is ≤2 scans.) Negligible at current scale; only matters at thousands of items.

**Recommendation:** If lists are expected to grow, build an id→index map once per
`SetItems`; fine at current scale.

#### L21. repoColorScheme ignores fang's LightDarkFunc  · **Status:** fixed (2026-06-22)

**File:** cmd/tskflwctl/main.go:84-113 | **Component:** cmd
**Effort:** XS · **Urgency:** eventually

The `LightDarkFunc` parameter fang supplies is dropped. (Overstated: the palette uses
ANSI 16-color indices remapped by the terminal's own theme plus `NoColor`, so it DOES
adapt; the only real issue is an unused parameter that would need wiring if a slot
became truecolor.)

**Recommendation:** Either use `LightDarkFunc` to choose background-appropriate slots,
or name the param `_` with a note that 16-color indices are intentionally
background-agnostic.

#### L22. Non-key messages forward only to the active tab, silently dropping them for background tabs  · **Status:** fixed (2026-06-22)

**File:** internal/tui/model.go:180-185 | **Component:** tui
**Effort:** S · **Urgency:** eventually

The fall-through sends untagged messages only to `m.cur().list` — an unstated
invariant that any future untagged background message (a spinner on a still-loading
background tab) would be misrouted. (Benign now: bubbles' own id/tag guards make
blink/spinner ticks harmless, and background tabs never focus their FilterInput.) The
future risk is real but speculative.

**Recommendation:** Document the invariant (every async list-affecting message must be
tab-tagged), or broadcast the residual to all tabs with per-tab `routeToTab` wrapping
if background components gain ticks.

#### L23. Task creation defaults live in CLI flags, not the core  · **Status:** fixed (2026-06-22)

**File:** internal/core/service.go:219-234,259-327 | **Component:** architecture
**Effort:** M · **Urgency:** eventually

`NewTask` hard-validates priority/tier/autonomy with no internal defaulting; the
defaults (Unknown/medium/3/3) live in the CLI flag definitions. Any second caller of
`NewTask` that doesn't replicate them gets `ErrValidation`. (Heavily downgraded: the
TUI has no task-creation path, and `NewTaskParams` doc already states defaults are the
CLI's job — a latent smell, not a current broken invariant.)

**Recommendation:** Move the defaults into `NewTask` (applied when a field is
zero-valued), keeping CLI flag defaults only as help-text hints.

## Refuted (verification working as intended)

Two reported findings were knocked down on close re-read and are NOT recorded above:

- *Picker dead-ends on zero options* — `fillSelect` (fill.go:42) guards empty option
  sets with `ErrValidation` + guidance before the picker ever opens.
- *Detail-load guard defeated by tab switch reusing an id* — each tab has a distinct
  `entityKind` and `switchTab → refreshDetail` always bumps `detailGen`; the
  `kind`+`id`+`gen` triple can't all match a stale response.

## Candidate tasks

Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do

- ✅ Contract-integrity batch (H1, H2, H4, M3, M4): centralize exit-code
  classification, `--dry-run` on `edit`, and the active-task/audit-bucket invariants
  at their narrowest seams.
- ✅ Entity-descriptor refactor (M1, M9, M10, L1) — descriptor + generic seams landed,
  TUI registry/lifecycle data-driven, the residual cost documented honestly.
- ✅ TUI state-restore hardening (H5, M6, M15) — generation-stamp restore intent and
  focus-gate global keys.
- ✅ Store robustness (M2, M5, M11, L4, L18) — file-mode preservation, MoveAudit CAS,
  duplicate-slug detection, fence whitespace, atomic-write cleanup tests + temp sweep.
  (Signal guard intentionally not shipped — the Move-crash duplicate is detected by
  lint instead; SIGKILL/power-loss defeats a guard anyway.)
- ✅ Doc truth-up (M8, M10) — re-justified the render→core and TUI-registry boundaries
  to match reality.
- ⏳ Low-backlog triage (L2, L6, L14, L16, L20) — all low-value polish or
  accept-as-documented; fix opportunistically or close with rationale.
