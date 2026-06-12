---
status: reference
created: 2026-06-11
tags: [review, tui, core, polish, adoption, reference]
---

# Critical review & "make it shine" research

A fresh-eyes engineering review of `tskflwctl` (CLI + the `feat/tui-multi-entity-navigation`
branch) plus competitive research on how to make the project shine. Findings were
verified by tracing the code; the non-TUI items were independently re-checked
against the source. Build, `go test ./...`, and `go vet ./...` were green at the
time of review (golangci-lint not available in the review sandbox).

## Verdict

The bones are excellent: hexagonal layering (`domain` pure → `core` use-cases →
`store` adapter → `cli`/`tui` primary adapters), status==directory as the source
of truth, `--json` + `schema_version` everywhere, semantic exit codes (10–14),
atomic + surgical frontmatter writes, self-hosted planning. `store/diagnose.go`
(git-conflict-marker detection, pinpointed field/line diagnosis) is better than
most production tools. The TUI faithfully implements its own design spec (truncate
discipline, stale-guard by id, no I/O in `Update`/`View`, value-receiver
reassignment). Everything below is edge cases and polish, not foundational.

Top three to fix first: **B1** (epic validation), **A1** (stale detail after
filter), **B4** (`lint --fix` silent on unrepairable files).

---

## A. TUI issues (verified by tracing the code)

### A1 — Detail pane goes stale after applying a `/` filter. [medium, real]
`internal/tui/model.go:181-202`, `137-141`. Two of the three selection-change
paths reload the detail pane; the filter path doesn't. While `/` is active
(`SettingFilter()`), every keystroke routes straight to
`m.cur().list.Update(msg)` (step 2), and the async `FilterMatchesMsg` is
forwarded at the bottom of `Update` — neither compares `selectedID()` to the
previous selection nor calls `loadItem`. So as you type a filter (the list moves
the cursor to the top match) *and* after Enter applies it, the right pane keeps
showing the pre-filter item until you press `j/k`. `updateList` (`model.go:284`)
has exactly the prev/`selectedID()` diff logic needed — the filter paths just
don't run through it. `TestModel_FilterNarrows` (`model_test.go:252`) only
asserts visible-item count, never detail content, so it passes while the bug
exists.
**Fix:** capture `prev := m.selectedID()` before forwarding filter-affecting
messages; if it changed, fire `m.cur().loadItem(...)` (or `showEmpty()` on zero
matches) — i.e. funnel these through the same tail as `updateList`.

### A2 — A failed *initial* load is an unrecoverable dead-end. [low-medium, real]
`internal/tui/model.go:72-81`. `reloadAll` only re-fires loaders for tabs where
`t.loaded`. If the first tasks load fails, `errMsg` sets `m.err` (full-screen
error; `View` returns early at `model.go:495`) and no tab is `loaded` — so both
`r` (`reloadMsg → reloadAll`) and the fsnotify reload path no-op. The only escape
is `ctrl+c`. `TestModel_RecoversFromFatalError` (`model_test.go:604`) masks this
because it recovers via `m.Init()()`, which calls `cur().reload()` directly — a
path the `r` key never takes.
**Fix:** have `reloadAll` (or the `r` handler) always reload the *active* tab
even when `!loaded`.

### A3 — `q` exits the program from single-pane detail instead of popping focus. [low, design]
`internal/tui/model.go:215`. The global `Quit` match in step 3 runs before focus
routing, so in 60–89-col single-pane "drill" mode, `q` while drilled into the
detail pane quits the whole app rather than returning to the list. This
contradicts the locked design intent ("`q` context-quit — close overlay/mode
first") and the tool's own layering of `Esc` under `q`. `Esc`/`h` correctly go
back; `q` skips the layer.
**Fix:** in single-pane + `focusDetail`, treat `q` as "back to list" first.

### A4 — Working-set sort buckets unknown statuses as in-progress. [low]
`internal/tui/commands.go:116-128`. `statusRank` is a map; a foreign/legacy
status word (which the domain explicitly *tolerates*) misses the map and gets
rank `0`, so it sorts *above* real in-progress tasks. Give `sortWorkingSet` a
sentinel default that ranks unknown statuses last.

### A5 — Minor drift / signals worth tightening
- `recomputeLayout` uses `m.width >= 90` for two-pane (`model.go:467`); the
  design doc says ≥100. Harmless, but pick one and make the doc match.
- `Run` attaches the watcher best-effort and silently leaves live-reload **off**
  if `newWatcher` errors (`tui.go:14`) — no user-visible signal. Consider a
  one-time dim footer note ("live reload off").
- The list reserves a pagination-dots line unconditionally (`model.go:465`),
  costing one row even when the list doesn't paginate. A
  `list.Paginator.TotalPages > 1` check would reclaim it.

---

## B. core / store / config issues (independently verified)

### B1 — `task set --epic <bogus>` writes a dangling reference unchecked. [high]
`internal/core/service.go:72-87` → `internal/domain/validate.go`.
`ValidateField`'s switch has **no `epic` case**, so it falls through to
`return nil`. `NewTask` checks epic existence but `SetFields` doesn't — so
`task set my-task --epic 99-nope` succeeds and the inconsistency only surfaces
later under `lint`. The service already has `ListEpics`; validate there, or
document the asymmetry deliberately.

### B2 — `configuredRoot` doesn't enforce the containment its comment promises. [medium]
`internal/config/config.go:48-54`. The comment says the result is "kept within
dir's tree," but the code is just `filepath.Join(dir, rel)` with no check —
`taskflow_root = "../../elsewhere"` resolves to an arbitrary path. It's the
user's own config so the threat is mild, but the code doesn't do what it claims.
Either reject a cleaned path that escapes `dir`, or fix the comment.

### B3 — CRLF files get rewritten to mixed line endings on edit. [medium]
`internal/store/frontmatter.go`. `splitFrontmatter` recognizes CRLF, but on a
surgical edit the YAML encoder re-emits the frontmatter block with `\n` while the
body keeps `\r\n` — a mixed-ending file and a churny "whole frontmatter changed"
first diff, which breaks the minimal-diff promise for Windows/synced repos (the
very repos `diagnose.go` courts). No test covers CRLF round-trips. Detect the
dominant ending and normalize, or document that edits normalize to LF.

### B4 — `lint --fix` exits 0 and says nothing when a file is unrepairable. [medium]
`internal/store/fix.go` + `internal/cli/lint.go`. `FixFrontmatter` only reports
files it *changed*; a malformed file the text-fixer can't repair (and that
`realignStatus` skips because it won't parse) produces no output and a success
exit. Plain `lint` surfaces these with a non-zero exit, but `--fix` silently
leaves them broken. Have the fix path also return/print the still-unreadable
files and exit non-zero.

### B5 — `time.Now()` re-injected in the service breaks the otherwise-clean DI seam. [low]
`internal/core/service.go:66,85`. The store's `Move(slug, to, now)` takes an
injected clock (and is tested with a fixed time), but `Service.Move`/`SetFields`
call `time.Now()` internally, so the service layer's date-stamping isn't
deterministically testable and the TUI can't control it. A `Clock` field on
`Service` would restore symmetry.

---

## C. Planning / process observations

- The planning corpus is unusually disciplined (locked decisions "after two
  research agents," sprint splits when work grew, `[[wikilink]]` cross-refs). A
  strength — keep it.
- Smell: much of `planning/research/*.md` (ai-provider-abstraction, hybrid-search,
  doc-generation, monorepo) describes a *much larger product* than the shipped
  CLI, and a swath of `tasks/deprecated/` are its tombstones. Fine as history, but
  a new contributor reading `planning/research/` will be misled about scope. Add a
  `planning/research/README.md` (or an `archived/` subdir) noting "these predate
  the Go rewrite; current scope is the CLI+TUI."
- `tasks/in-progress/port-pm-to-go-cli-…` is still in-progress while the
  README/ARCHITECTURE declare `pm` retired and the loop complete — a lint-clean
  dogfooding tool with a stale in-progress task is a small credibility ding.
  Close or demote it.

---

## D. Research: how to make it shine

Best-in-class TUIs (k9s, lazygit, gh-dash, taskwarrior-tui, the Charm suite) win
on **discoverability + liveness + a memorable demo**, in that order.

### Tier 1 — high impact, low effort
1. **A VHS demo GIF, auto-regenerated in CI.** Highest-ROI adoption move for a
   Charm-stack tool. Write a `.tape` (settings at top, `Require` for fail-fast,
   `WindowBar "Colorful"` + margin + `LoopOffset` for a compelling preview frame),
   wire up `charmbracelet/vhs-action` so the README GIF never goes stale, and
   reuse the `.ascii` output as golden integration tests.
2. **A teaching status bar / contextual footer (already ~80% there).** Push the
   existing context-sensitive footer further: a count badge ("12 active · 3
   in-progress"), the active sort/view chip inline, and a first-run dim "press ?
   for keys" nudge. The "status bar that tells you what every key does" is the
   thing repeatedly cited as letting people learn k9s/lazygit without docs.
3. **Make `:` a fuzzy command palette (jump + actions + help).** The `?` overlay
   is static text; lazygit users explicitly want a *filterable* keybinding palette
   (lazygit#4846). The `:` bar is already the right seam — make it the k9s-style
   searchable palette.

### Tier 2 — differentiators (this is a *planning* tool — lean in)
4. **An at-a-glance dashboard as the default landing view.** `core.Summary()` and
   the CLI `status` board already exist; surface them as a TUI home tab (counts
   per status, in-progress items, epic progress bars via the existing `miniBar`,
   and a "what should I do next" pick). This is value a git/k8s TUI can't have —
   nobody opens lazygit to feel oriented about their week.
5. **Time-in-status / staleness signals.** Dates are already stamped; compute
   "in-progress for 9 days," surface a ⏳ glyph on stale rows (reuse the `⚠`
   misfiled precedent), add burndown/velocity from completion dates and a "stale
   work" view. Planning-native delight with no analog in the modeled tools.
6. **Cross-link navigation with a back-stack (already epic'd as S6).** Following
   epic↔task references with jump + breadcrumb is exactly the discoverability win
   cited for git TUIs — people discover features by *seeing* them linked.

### Tier 3 — feel & reach
7. **Micro-interactions on mutations (S4):** optimistic row update + a brief toast
   ("✓ moved alpha → in-progress"), spinner only during the write.
8. **Themeability** + the `NO_COLOR`/16-color discipline already present; ship a
   couple of built-in themes that honor terminal background.
9. **Distribution:** Homebrew tap + `goreleaser` + the VHS GIF in a crafted README
   is the standard adoption funnel.

**Resist:** the research corpus hints at AI/semantic-search ambitions. Tools that
shine here win by being fast, legible, and local-first — keyboard-first liveness
beats a smarter backend. Ship the dashboard and the demo GIF first.

---

## Suggested immediate next moves
1. Fix **B1** (epic validation) and **A1** (stale detail after filter) — both
   small, both correctness.
2. Add a **CRLF round-trip test** and a **`lint --fix` unrepairable-file** test
   (B3/B4) — they pin real data-safety promises.
3. Record a **VHS demo** and put it at the top of the README.
4. Build the **at-a-glance dashboard tab** — the real differentiator.

## Sources
- VHS — https://github.com/charmbracelet/vhs
- vhs-action — https://github.com/charmbracelet/vhs-action
- VHS tape best practices — https://tywer.dev/beyond-screenshots-capture-cli-magic-with-charmbracelet-vhs
- Teaching Claude to design TUIs — https://griffen.codes/post/tui-design-skill-claude/
- The k9s TUI for Kubernetes — https://dustymabe.com/2020/07/18/the-k9s-tui-for-kubernetes/
- lazygit searchable keybindings request — https://github.com/jesseduffield/lazygit/issues/4846
