---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Render the detail-pane body with glamour (cached, never in View), with an R toggle between raw and pretty; persist the preference
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-11"
updated_at: "2026-06-13"
started_at: "2026-06-12"
completed_at: "2026-06-13"
id: 6fb7ym4017sj
---

# TUI: glamour markdown rendering with raw/pretty toggle

## Objective

The detail pane shows raw markdown. Render it with **glamour** for comfortable
reading, but keep a **raw toggle** — the CLI `task show` deliberately stayed raw
to preserve fidelity + avoid a heavy dep, so the TUI should offer both rather than
replace raw. This was always the plan (epic decision: "plain first, then glamour").

## Approach

- Add `glamour` as a dep. Render in the detail **load Cmd / `SetContent`**, cache
  the compiled string on the pane — **never call glamour in `View()`** (it's
  CPU-heavy; the documented footgun). Re-render on `SetSize` (width reflow).
- Cache both representations per item: the raw (current `renderTaskDetail`-style)
  string and the glamoured body. A toggle just swaps which the viewport shows; no
  re-render needed to flip.
- **`R` key** toggles raw ⇄ pretty; a **global preference** (on the Model) so it
  persists across selections and tabs. Show the mode in the detail title or footer.
- Frontmatter fields stay as the existing styled key/value block; glamour applies
  to the **body** only (the fields aren't markdown).

## Interaction with detail find (important)

S2b's `/` find highlights by line index over the rendered content. Glamour reflows
and indents, so: (a) match line indices differ between raw and pretty — recompute
matches on toggle; (b) the "rebuild matched line from stripped text" highlight drops
glamour styling on matched lines (the #4 issue in
[tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort](6fb7ym4023q0-tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort.md), amplified).
Coordinate with that task — ideally land the ANSI-aware highlighter first.

## Acceptance criteria

- [x] The detail body renders as styled markdown (headings, lists, code, emphasis)
      by default; `R` flips to raw and back; the mode is visible.
- [x] Toggling and resizing never lag — glamour runs in `Update`/load, not `View`;
      flipping mode does no re-compile.
- [x] The preference persists across selection and tab changes.
- [x] Find still works in both modes (matches recomputed on toggle). Suite + lint green.

## Out of scope

- Glamour for epic/audit bodies beyond the same body-rendering path (apply uniformly
  via `detailContent`, but no per-entity markdown features).
- A custom glamour theme (use a sensible built-in; theming is a later polish).

## Progress Log

### 2026-06-12 — implemented (suite + race + lint green)

- **`glamour.go`** — `glamourBody(md, width)` compiles a glamour renderer (dark
  style, word-wrapped to the pane), called only from `Update` (SetContent/SetSize),
  never `View`. Falls back to plain `wrap()` on any error.
- **`detail.go` split** — `detailContent` now exposes `meta(width)` (the styled
  frontmatter block + epic task list, rendered once) and `rawBody()` (the markdown).
  The pane caches **both** compositions per item — `rawStyled` (meta + wrapped raw)
  and `prettyStyled` (meta + glamour) — and `styled` points at the active one. So
  `render()` runs glamour once per content/width change; the toggle just swaps.
- **`R` toggles** raw ⇄ pretty (`detailPane.toggleMode`, `keys.RawToggle`), default
  **pretty**. The mode is a pane field, so it **persists across selections + tabs**.
  Raw is flagged as `slug · raw` in the detail title; pretty (default) is unlabeled.
- **Find composes for free** — `refreshFind` already highlights over `styled` with
  the ANSI-aware highlighter (from the polish pass), so find works over glamour
  output; toggling recomputes matches against the new line structure.
- **Width-only re-render** — `SetSize` re-runs glamour only when the width changes
  (height-only resize skips it).
- Tests (`glamour_test.go`): `glamourBody` rendering/fallback, the toggle +
  `· raw` indicator, the preference persisting across selection, and find finding
  the same matches in both modes. The existing layout-invariant tests now exercise
  glamour-rendered bodies (the final width clamp keeps lines within the terminal).
- New direct dep: `github.com/charmbracelet/glamour v1.0.0`.

### 2026-06-12 — fast-follow fixes (from a review pass)

- **Renderer cached per width** (`detailPane.prettyBody`, `glamW`) — was building a
  `glamour.NewTermRenderer` (goldmark + chroma setup) on every selection; now a
  same-width render reuses it, rebuilt only when the width changes. Test:
  `TestDetailPane_GlamourRendererCachedByWidth`.
- `clear()`/`showEmpty()`/`SetError()` now also reset the `rawStyled`/`prettyStyled`
  caches (was a latent stale-cache trap, unreachable today).
- Discoverability: `R raw/pretty` added to the detail footer hint, plus a help note
  that **find matches the rendered text on screen, not the markdown source**.
- Docs: `docs/ARCHITECTURE.md` + `README.md` updated for S4 (actions) and S5
  (glamour/raw toggle).

**Dependency note:** glamour v1.0.0 pulls `lipgloss` to an untagged pseudo-version
via MVS and adds `x/net` + `bluemonday`. Nothing broke, but the core styling dep is
now pinned to a commit — the CI `govulncheck` step will start covering these once
pushed.

**Deferred (per out-of-scope):** custom glamour theme + light/dark auto-style →
[tui-glamour-theming-auto-light-dark](6fbj87002ex5-tui-glamour-theming-auto-light-dark.md); epic/audit bodies use the same path but no
per-entity markdown features.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Interacts with [tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort](6fb7ym4023q0-tui-s2b-polish-find-occurrences-highlight-fidelity-per-entity-sort.md)

## Coordination note from the container agent (2026-06-12, pre-start)

Read in preparation for this task starting. The "Interaction with detail
find" section is now PARTLY STALE — the merged polish pass landed today and
changes this task's footing, almost entirely favorably:

1. **The precondition is met.** "Ideally land the ANSI-aware highlighter
   first" — done. Find is now occurrence-level (`matchPos` in `find.go`),
   matching via `foldMatches` (rune-by-rune fold, offsets always index the
   original string) and highlighting via `highlightLine`, which rebuilds
   matched lines over the *styled* text with display-column `ansi.Cut` — so
   glamour's colors survive everywhere except the matched span itself. The
   "rebuild from stripped text drops styling" concern in the section above
   no longer exists.
2. **Recompute-on-toggle is nearly free.** `refreshFind()` recomputes all
   matches from `d.styled` (ANSI-stripped per line) every call. A toggle
   that swaps `d.styled` and calls `d.refreshFind()` gets correct matches in
   the new mode for free. Note the semantic: find matches what's ON SCREEN
   (in pretty mode, searching `bold` matches rendered text; searching
   literal `**` won't) — that's the right UX, but worth a line in help.
3. **Current `detail.go` seams** (changed since this task was written):
   `SetContent` now preserves the viewport scroll when the SAME item
   refreshes (live-reload) — a mode toggle on the same item will keep
   YOffset and the viewport clamps it; decide whether toggle should instead
   GotoTop (recommend: keep position, clamp handles shorter content).
   `SetSize` re-renders content on resize — glamour re-compile on resize
   lands there naturally (resize is rare; fine).
4. **Body wrapping:** `renderTaskDetail`/epic/audit funnel everything
   through `wrap()` (lipgloss width re-wrap). Glamour does its own
   word-wrap (`WithWordWrap(width)`) — the glamoured body must NOT also go
   through `wrap()` or long lines double-wrap. Split: fields block wrapped
   as today, body joined after, per-mode.
5. **`R` key is free** (`o/O` sort, `r` refresh, `a` action menu are taken).
   Add it to keys.go AND to the help overlay's Detail section — help is now
   scrollable (`helpLines`/`helpBox(w,h,scroll)` in help.go); just append
   the entry.
6. **Test determinism:** glamour's auto style detection queries the
   terminal/env. In tests force a fixed style (e.g. `WithStandardStyle`) —
   mirror how `TestHighlightLine` forces a lipgloss color profile.
   `TestModel_ViewFitsTerminal` will exercise glamoured content sizes; keep
   it green.
7. **Detail loads are generation-guarded now** (`detailGen` + (kind,id) in
   model.go) — if the toggle re-renders via a Cmd instead of synchronously,
   stamp it like `loadDetail` does. Synchronous swap in Update (from the
   cached strings) avoids the question entirely — recommended.

Suite/vet/golangci-lint were all green at this note's writing; the TUI suite
is the contract — keep it green.
