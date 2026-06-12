---
status: reference
created: 2026-06-12
tags: [review, code-quality, architecture, tui, cli, testing, reference]
---

# Critical code review — multi-lens (2026-06-12)

A follow-up to [[2026-06-11-critical-review-and-polish-research]]. Five
parallel reviewers, each with a different lens — core/store correctness, CLI
code+UX, TUI code+UX, architecture, and testing/tooling/hygiene — all briefed
on yesterday's findings so they verified rather than re-reported them. Every
finding below was confirmed by tracing the actual code (file:line cited); the
three highest-severity new findings were additionally re-verified by hand
(grep / live binary). Reviewed tree: `feat/tui-multi-entity-navigation` plus
the uncommitted `harden-task-set` changes. `go build ./...` and
`go test ./...` green at review time.

## Verdict

The architecture is real, not aspirational: the import graph was verified
clean (`domain` pure → `core` → `store` implements consumer-defined ports;
`cli`/`tui` adapt on top), atomic-write and surgical-frontmatter discipline is
excellent, DI is genuinely global-free, color/TTY handling is exemplary, and
the TUI's Elm discipline largely holds. The new `task set` hardening closes
yesterday's worst data bug (B1).

But this round found things yesterday's pass missed, including two
**high-severity TUI bugs on the current branch** (Esc quits the app from list
focus; reload-while-filtered blanks a tab), a **broken CI lint step** (so
"lint green" is currently unverifiable), and a **workflow contradiction**
(`task new`'s own scaffold fails `tskflwctl lint`). The recurring theme across
lenses: *contracts that are advertised but not enforced* — exit code 12 can
never fire, `--json` silently drops fields the human output shows, silent
empty results where a validation error belongs, and storage-layout knowledge
that has leaked out of the store into the TUI watcher and `config.Init`.

## Status of yesterday's findings

| ID | Finding | Status today |
|----|---------|--------------|
| B1 | `task set --epic <bogus>` dangling ref | **Fixed** in working tree (`service.go` coercion + `epicExists`, store parse-before-commit guard, new tests) — residual edges below |
| A1 | Detail pane stale through `/` filter | Open (`model.go:201-206`, `141-143`) |
| A2 | Failed initial load unrecoverable via `r` | Open (`model.go:75-84`) |
| A3 | `q` quits from single-pane detail | Open (`model.go:219`) |
| A4 | Unknown status ranks as in-progress in working-set sort | Open (`commands.go:116-129`) |
| A5 | 90 vs 100 threshold / silent watcher failure / pagination row | Open |
| B2 | `configuredRoot` containment comment lies | Open (`config.go:48-54`) |
| B3 | CRLF → mixed line endings on edit | Open (`frontmatter.go:86-92`, `fix.go:114-119`) |
| B4 | `lint --fix` silent on unrepairable files | Open (`cli/lint.go:56-66`, `store/fix.go:19-67`) |
| B5 | `time.Now()` inside Service | Open — the new SetFields diff adds a fourth call site (`service.go:66,103,226,280`) |

---

## High severity (new)

### H1 — Esc in list focus quits the entire app. [TUI, verified]
`internal/tui/entity.go:121-132`, `model.go:247-255`. The embedded
`bubbles/list` keeps its default `Quit` binding (`q`, `esc`) and `mk()` never
calls `DisableQuitKeybindings()` (grep-verified: no call anywhere in
`internal/tui`). With list focus and no filter applied, Esc falls through the
model's list branch into `list.Update` → `handleBrowsing`, which returns
`tea.Quit`. The UX spec says Esc returns focus to the left panel. No test
presses Esc in list focus (`TestModel_FocusRouting` only does so
detail-focused). **Fix:** `l.DisableQuitKeybindings()` in `mk()`; add the
missing test.

### H2 — Reload while a background tab has a filter applied blanks that tab. [TUI]
`internal/tui/model.go:139-143`, `167-168`. `reloadAll` (now fired by every
fs event) calls `SetItems` on a filtered background tab, which nils its
`filteredItems` and returns a refilter cmd — but the resulting
`FilterMatchesMsg` is forwarded by the default branch to the **active** tab's
list. The background tab is left in `FilterApplied` state with nil matches
(empty list until the filter is cleared); if the active tab also has a filter,
it receives the *other entity's* match set. **Fix:** route filter cmds/msgs by
tab (wrap the cmd), or clear/reapply filters explicitly on reload. Related:
cursor restore after reload is also lost on filtered tabs
(`entity.go:63-70` — `selectByID` walks `VisibleItems()` while `filteredItems`
is still nil, consuming `restore` with no effect).

### H3 — CI lint step is broken: golangci-lint v1 against a v2 config. [tooling, verified]
`.github/workflows/ci.yml:47` installs
`github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5`, but
`.golangci.yml` is v2-schema (`version: "2"`, `linters.default: standard`,
top-level `formatters:`). v1.x cannot parse v2 configs, so the lint gate
errors out — and locally an installed v1 fails on the Go 1.25 toolchain too.
The repo's own definition of done ("get all three green") is currently
unverifiable for lint. **Fix:** pin v2 (`…/golangci-lint/v2/cmd/…`) or use
`golangci/golangci-lint-action`; do the same wherever `just lint` is run.

### H4 — `task new`'s own scaffold fails `tskflwctl lint`. [UX, verified live]
Fresh `init` → `epic new` → `task new "T1" --epic … --description d` →
`lint` exits 11 with `t1 / tags: missing`. The create verb (`cli/task.go:34-68`,
no required or defaulted `--tags`) produces a file the tool's own linter
rejects, so the documented create→lint-clean workflow is broken out of the
box. **Fix:** `new` and `lint` must agree — either default/require tags at
creation or drop the non-empty-tags lint rule.

### H5 — `task list --status <typo>` (and `--epic <bogus>`) silently returns empty, exit 0. [CLI+core]
`cli/task.go:96`, `core/service.go:44`. The status filter is never run through
`domain.ParseStatus` (which `task move` already uses), so a typo is
indistinguishable from an empty bucket. For the stated agent audience routing
on exit codes this is actively dangerous. **Fix:** validate filter values →
`ErrValidation` (exit 11), and make the error enumerate valid statuses
(`task move`'s message currently doesn't either).

---

## Medium severity

### Core / store correctness

- **M1 — `Move` mutates disk, then can report failure.** `store/fsstore.go:139-145`
  parses the assembled content *after* writing the new file and removing the
  old; a task that round-trips as YAML but fails typed decode (e.g.
  `tier: "4"`) moves on disk yet returns an error — a retrying agent sees a
  "failed" move that succeeded. Same pattern in `MoveAudit`
  (`auditstore.go:102-105`, parse after `os.Rename`). **Fix:** parse before
  committing, exactly as the new `SetFields` guard does.
- **M2 — Misconfigured `taskflow_root` silently forks the data tree.**
  `config.go:51-54` + `fsstore.go:50-53`. A typo'd root shows a clean empty
  project (missing dirs read as empty) and `task new` `MkdirAll`s a brand-new
  tree there — planning data split across two roots with no signal. **Fix:**
  `Discover` should verify the configured root looks like a planning root.
  (Compounds B2: containment isn't enforced either. Also: stray empty
  `tasks/`, `epics/`, `audits/`, `projects/` dirs at the *repo* root are
  exactly this trap materialized — without `.tskflwctl.toml`, walk-up
  discovery anchors to them. Delete them.)
- **M3 — No read-modify-write concurrency control; a concurrent `Move` +
  `SetFields` can resurrect a task in its old status dir.** `fsstore.go:97-174`.
  `SetFields` writes back to the originally-resolved path; if a `Move`
  relocated the file in between, the write recreates the slug in the old
  directory → permanent `ErrAmbiguous` with no repair tooling. Atomic writes
  protect against torn writes, not lost updates. **Fix:** re-resolve before
  rename (compare-and-swap) or the advisory `flock` ARCHITECTURE.md already
  lists as future work. Related: `nextEpicNumber` is a TOCTOU race
  (`create.go:81-132`) — `O_EXCL` only dedupes identical `NN-slug`.
- **M4 — The new SetFields parse-guard returns a non-sentinel, misleading error.**
  `fsstore.go:166-169` wraps `errBadFrontmatter` ("malformed frontmatter…")
  instead of `domain.ErrValidation` — exit 1 instead of 11, and the message
  blames the file when nothing was written. **Fix:** wrap `ErrValidation` with
  an "update would not reload; nothing written" message.
- **M5 — "List fields" knowledge is duplicated and divergent.**
  `core/service.go:114` coerces only `tags`; `store/diagnose.go:14-17` /
  `fix.go:171` treat `related_tasks`, `dependencies`, `blocks`… as lists. So
  `task set --set related_tasks=a,b` writes a string that the project's own
  `lint --fix` then rewrites — the tool generates the drift its linter
  repairs. **Fix:** one canonical table (domain is the natural home), ideally
  derived from `domain.Task`'s yaml tags; this also addresses the
  architecture-level smell that `SetFields(map[string]any)` smears type
  knowledge across three packages with comment-only synchronization
  (`service.go:110-115`, `frontmatter.go:128-143`).
- **M6 — `taskflowRoot` breaks on inline TOML comments.** `config.go:68-74`.
  `taskflow_root = "." # comment` yields garbage (`strings.Trim` only strips
  the leading quote). TOML allows inline comments. **Fix:** strip `#`-comments
  outside quotes before unquoting.

### CLI / JSON contract

- **M7 — `--json` drops fields the CLI itself writes, including the misfiled signal.**
  `render/render.go:21-31`: `taskJSON` omits `effort` and `autonomy_level`
  (both settable via flags — write-only fields), and there's no
  `misfiled`/`declared_status` even though human output renders `⚠`. Agents
  are the stated `--json` audience and they're exactly who should detect
  status/dir drift. `epicJSON` similarly drops `priority`, `created`, `tags`.
  **Fix:** round-trip all frontmatter fields + a misfiled flag (minor schema
  bump). While in there: `"created"` vs `"updated_at"` key asymmetry, and
  decide whether one global `SchemaVersion` for ~10 payload shapes is
  intentional before 1.x consumers exist.
- **M8 — Cobra's own output bypasses the injected writers.** `cli/root.go:45-81`
  never calls `root.SetOut/SetErr`, so help text, usage errors, and completion
  scripts go straight to `os.Stdout/Stderr`, contradicting the package's "all
  output flows through the injected writers" header (tests already patch
  around it, `task_test.go:35-36`). **Fix:** two lines in `NewRootCmd`.
- **M9 — Epic `status` is free text end-to-end.** `cli/epic.go:42` accepts
  `--status bananas`; `domain.Epic.Status` is an unvalidated string,
  `NewEpic` doesn't check it, and `lint` only lints tasks. Tasks got a closed
  enum; epics got nothing. **Fix:** define the epic-status vocabulary in
  domain (or document it as deliberately open and lint it loosely).
- **M10 — `init` is the one command that ignores `--json`.** `cli/init.go:26-45`
  prints human output regardless. **Fix:** an `InitJSON` envelope.
- **M11 — `task move` completion offers task slugs for the `<status>` arg.**
  `cli/task.go:193` uses `completeTaskSlugs` for all positions. The status set
  is closed and small — completion actively misleads here. **Fix:**
  position-aware completer; also register the existing flag completers
  uniformly (`task set --epic`, `task list --epic/--status/--tag` have none).

### TUI

- **M12 — Every fs event yanks the detail pane to the top.** `model.go:170` →
  `refreshDetail` → `SetContent` → `vp.GotoTop()` (`detail.go:62`). With the
  watcher live, any write under `planning/` rescrolls the detail you're
  reading. **Fix:** preserve `YOffset` when the incoming `(kind,id)` matches
  the displayed item.
- **M13 — One failing tab's loader blanks the whole UI; concurrent reloads race on `m.err`.**
  `commands.go:36-38`, `model.go:135-137,157`. During `reloadAll`, one tab
  failing + another succeeding resolves nondeterministically (full-screen
  error flickers or sticks); the failing tab keeps stale rows silently.
  **Fix:** per-tab error state; reserve `m.err` for the nothing-loaded case.
- **M14 — List loads have no generation guard (detail loads do).**
  `model.go:150-171`. Cmds run concurrently; cycling status views fast can
  land an older load last — chip says `view:completed`, rows show the previous
  view. **Fix:** stamp `listLoadedMsg` with a generation or the view it was
  loaded for, drop mismatches (same idea as `isCurrentSelection`).

### Architecture / tooling

- **M15 — Storage-layout knowledge has escaped the store in two places.**
  (a) The TUI watcher reconstructs `<root>/tasks/<status>` itself
  (`tui/watch.go:44-55`) and `cli/ui.go:17` leaks `Cfg.Root` to make that
  possible — contradicting "the TUI never touches the fs" in
  ARCHITECTURE.md. (b) `config.Init` hardcodes the status-dir list as string
  literals (`config.go:104-107`) instead of deriving from
  `domain.AllStatuses()`, with no sync-guard test (the TUI's `statusViews`
  has one). A new status would ship with `init` not scaffolding its dir while
  the watcher watches it. **Fix:** expose `WatchPaths()` from the store
  through core; derive `Init`'s list from domain.
- **M16 — Exit code 12 / `ErrInvalidTransition` is an advertised contract that can never fire.**
  `domain/errors.go:11`, `cli/exit.go:19`, README + ARCHITECTURE document it;
  nothing returns it — `Move` accepts any status→status. A dead documented
  contract is worse than none for the scripting agents it targets. **Fix:**
  implement a transition matrix in domain, or remove the sentinel and the
  documented code.
- **M17 — Every operation is a full-tree rescan; the TUI multiplies it per fs event.**
  `fsstore.go:44-74`, `service.go:315-325,400-416`, `model.go:75-84`. Fine at
  ~200 tasks; at 10k, every editor save costs tens of thousands of
  read+parse calls across loaded tabs. Not urgent — but when it bites, the
  architecture-preserving fix is an mtime-keyed cache *inside* `FS` (behind
  the port), not indexes in core. Record the intent now so nobody adds caching
  without a mutex (FS's safety currently comes from statelessness).
- **M18 — `core.Service` is the weakest-covered critical path: 54.5%, with
  `ListTasks`, `Move`, `NewEpic`, `Lint`, `LintFix`, and all audit use-cases
  at 0% direct coverage** — exercised only through CLI tests. The `fakeStore`
  seam exists and is only used for epic rollup. Coverage elsewhere:
  cli 81.7%, **render 24.9%** (every formatter 0% direct, no golden files),
  store 80.2%, tui 79.3%, domain 85.9%. **Fix:** fakeStore-based units for
  Move/Lint/audits; golden-file tests for render (pure functions over
  view-models — ideal targets).
- **M19 — Docs point at a spec that isn't in git.** README/CLAUDE.md say
  `tests/test_pm.py` is "kept as the historical executable spec", but only
  `bin/pm` is tracked — `tests/` is untracked, so a fresh clone loses it.
  Commit it or delete both relics and update the docs together.
- **M20 — No vulnerability scanning.** No `govulncheck` in Justfile or CI;
  `.golangci.yml` defers gosec as follow-up. Free to add a
  `govulncheck ./...` step.

---

## Low severity (grouped)

**Core/store edges:**
- B1 residuals: clearing an epic is now impossible (`--epic ""` →
  `unknown epic ""`, `service.go:94-102`); `--set updated_at=…` validates then
  is silently clobbered (`service.go:103`).
- Unterminated frontmatter (` --- ` never closed) parses as "no frontmatter"
  with no `FileProblem`; a later `SetFields` then double-fences the file
  (`frontmatter.go:45`, `fsstore.go:202-209`).
- Slugs/epic-ids aren't sanitized against path separators —
  `epic show ../tasks/in-progress/x` reads outside the directory
  (`fsstore.go:181`, `epicstore.go:47`, `auditstore.go:112`).
- `config.Init` writes non-atomically with a check-then-write race
  (`config.go:121-123`) — contradicts the repo's own atomic-helper convention.
- `writeFileAtomic` never fsyncs the parent dir (`atomic.go:42-52`), so
  `Move`'s crash-safety comment overstates the guarantee.
- Audit "finding"/"open" semantics live as regexes in the storage adapter
  (`auditstore.go:15-24`) — a domain invariant invisible to core; consider
  `domain.CountFindings(body)`.
- `Task.Path` et al. put adapter-owned fs paths in the "pure" domain types —
  pragmatic, but acknowledge it in ARCHITECTURE.md rather than claiming
  purity unqualified.
- Frontmatter files carry no version marker; reserve a key in scaffolds now
  as cheap forward-compat insurance.

**CLI polish:**
- Failed transitions print the error twice (`moves.go:13-23` + main.go).
- Load-problem diagnostics: stderr in list commands, stdout in `lint` — pick
  one stream.
- `audit list --closed --deferred` conflicts silently (already tracked in the
  `clirender-polish-batch…` task; `MarkFlagsMutuallyExclusive` is the
  one-liner).
- `list` command bodies triplicated across task/epic/audit (`task.go:78-94`
  etc.); dead exported `Style.Enabled` (`render/style.go:63`).

**TUI polish:**
- `highlightOccurrences` can misalign/panic on runes whose lowercase changes
  byte length (`find.go:43-55` — offsets computed on `ToLower(plain)`, sliced
  from `plain`).
- Task-row date column misaligns for non-ASCII slugs (`item.go:68` pads by
  bytes, not display cells).
- `View` mutates shared state (`model.go:548` sets `t.list.Title` during
  render) — Elm-discipline wart; set it in `Update`.
- Closed/deferred audits are unbrowsable in the TUI
  (`commands.go:90-92` hardcodes the open bucket) — needs a bucket view or an
  explicit scope note.
- Help/keys drift: help claims `ctrl+d/u` half-page in lists (not bound
  there); list paging keys undocumented; Detail help omits `/`, `n/N`.
- Detail loads for the *same* id aren't ordered (stale guard is id-equality
  only); a monotonic request generation would make it airtight.

**Testing/tooling/hygiene:**
- `just test` lacks `-race` while CI has it — the watcher/debounce code is
  exactly where races live; first race report shouldn't be in CI.
- CRLF/unicode handled in code but only fuzz-asserted (no-panic), never
  behaviorally tested for correct round-trip values.
- Four hand-rolled "build a planning tree" test fixtures across packages —
  an `internal/testutil` builder before they sprawl.
- No integration tests of the built binary (exit codes 10–14 never tested
  through a real process exit; the ldflags version stamp unverified).
- No LICENSE, no CHANGELOG; `.gitignore` carries a large dead-Python section
  and ignores `bin/` while `bin/pm` is tracked inside it.
- Planning hygiene: `tui-sprint-3-fsnotify-live-reload` is still in-progress
  though the feature shipped in `f76254a`; `port-pm-to-go-cli…` remains the
  known stale one from yesterday. `lint` itself passes (verified via
  `go run`).
- TUI test gaps that map 1:1 to this review's bugs: Esc in list focus (H1),
  reload-while-filtered (H2), `r` after failed initial load (A2), watcher
  lifecycle, detail scroll across reloads. The teatest harness already used
  for `TestModel_FilterNarrows` is the right tool.

---

## What's genuinely good (keep doing it)

- Verified-clean import graph; consumer-defined ports with a real second
  implementation (`fakeStore`), so the seam isn't speculative.
- temp+fsync+rename writes, `O_EXCL` creates, write-new-before-remove-old
  moves; surgical yaml.Node edits preserving unknown keys/comments/order.
- Per-file `FileProblem`s threaded store→core→adapters so one bad file never
  blinds a listing; filename-based completion that works on broken files.
- Exit-code mapping verified end-to-end; idempotent re-transitions exit 0.
- Color discipline (NO_COLOR/FORCE_COLOR/CLICOLOR_FORCE/TTY precedence,
  documented order).
- The TUI suite is strong in absolute terms (36 tests incl. teatest
  full-program runs and a view-fits-terminal invariant); fuzz targets with
  good seeds on the frontmatter parser.
- Behavioral (not change-detector) tests throughout; the planning corpus
  discipline noted yesterday still holds.

## Prioritized next moves

1. **Branch blockers (before merging `feat/tui-multi-entity-navigation`):**
   H1 (Esc quits app), H2 (filtered-tab blank on reload), M12 (detail scroll
   reset) — all watcher/filter interactions introduced or activated by this
   branch — plus yesterday's still-open A1.
2. **Repair the quality gates:** H3 (CI golangci v2), add `-race` to
   `just test`, then actually run lint and burn down whatever it reports.
3. **Workflow/contract honesty, one batch:** H4 (`new` vs `lint` agreement),
   H5 (validate list filters), M4 (sentinel-wrap the new guard), M7 (JSON
   round-trips + misfiled), M16 (implement or delete exit 12).
4. **Data-safety follow-ups to the harden-task-set work:** M1 (parse before
   commit in Move/MoveAudit — same pattern, two more call sites), M5 (one
   canonical field-type table), B3/B4 from yesterday.
5. **Coverage where it's thinnest:** M18 (core use-case units via fakeStore,
   render golden files) — cheap, the seams already exist.
6. M2/M15 (root validation, layout knowledge back behind the port) when
   touching config/watcher next.
