---
status: completed
epic: 17-pm-go-cli
description: Mutually-exclusive audit list flags, case-insensitive audit Status regex, term.IsTerminal for color, and runewidth column alignment
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [go, cli, polish]
created: "2026-06-11"
updated_at: "2026-06-13"
started_at: "2026-06-13"
completed_at: "2026-06-13"
---

# CLI/render polish batch (audit flags, color TTY, wide-char width)

> ⚠️ **Externally proposed — needs independent review before implementing.**
> Filed from an outside code-review pass. These are low-severity polish items
> (and a couple were *deliberate* prior choices); the implementing agent should
> decide which actually earn their change before coding.

## Objective

A grab-bag of low-severity CLI/render correctness nits, batched. Some overlap
the "follow-up cleanup" already noted in
[[port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec]] (audit
regex over-matching, etc.).

1. **`audit list` flags aren't mutually exclusive** (`internal/cli/audit.go:31`).
   `--closed --deferred` (or `--all --closed`) silently honors the first the
   `switch` hits. Fix: `cmd.MarkFlagsMutuallyExclusive("all","closed","deferred")`.

2. **Audit `Status` regex is case-sensitive** (`internal/store/auditstore.go:21`).
   `openFindingRe` matches lowercase `open` only; a hand-edited `**Status:** Open`
   isn't counted. Convention is lowercase, so this is defensive — `(?i)` (the
   `[^-\w]|$` guard still blocks `opened`). *Debatable whether to bother.*

3. **`/dev/null` color pollution** (`internal/cli/color.go:46`). `isTerminal`
   keys on `ModeCharDevice`, which is true for `/dev/null` → ANSI is emitted to
   it. This was a *deliberate* choice (see the comment) and `/dev/null` discards
   the bytes, so impact is ~nil; `golang.org/x/term.IsTerminal` (already imported
   for width) is simply more correct.

4. **Column width counts runes, not display cells** (`internal/cli/render/style.go:128`).
   `visibleWidth` uses `utf8.RuneCountInString`, so CJK/emoji (2 cells wide)
   misalign tables. Fix: `runewidth.StringWidth` (`mattn/go-runewidth`, already a
   transitive dep). Also revisit `truncate` (rune-based) for the same reason.

5. **Description length counts bytes** (`internal/domain/validate.go:48`).
   `len(d) > 150` caps bytes, not characters — non-ASCII hits the cap early, and
   it's a slight parity gap (Python `len` counts code points). Consider
   `utf8.RuneCountInString`. *Lowest priority; confirm intended semantics.*

## Acceptance criteria

- [x] `audit list --closed --deferred` errors instead of silently ignoring one.
- [x] Tables stay aligned with a wide-char (CJK/emoji) cell.
- [x] Whichever of 2/3/5 the reviewer accepts have a test. Suite + lint green.

## Out of scope

- Anything behavioral beyond these nits; the broader audit finding-level
  commands (their own deferred work).

6. **Progress-bar logic is duplicated** (folded in from the S2a review,
   2026-06-11). The fill calc (`filled := pct*width/100`, clamp 0..width) now
   lives in **both** `render.Style.Bar` (CLI → ANSI) and `tui.miniBar` (TUI →
   lipgloss). Per the "shared theme, not shared rendering" decision, extract the
   *arithmetic* into a dependency-free `theme.BarFill(pct, width) int` (returns the
   filled-cell count) and have both renderers draw their own runes from it — so the
   two bars can't silently drift. The color is already shared via `theme.Percent`.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/cli/audit.go`, `internal/store/auditstore.go`,
  `internal/cli/color.go`, `internal/cli/render/style.go`,
  `internal/domain/validate.go`, `internal/theme/`, `internal/tui/style.go`.

## Closure (2026-06-13)

All six items accepted and shipped: (1) `MarkFlagsMutuallyExclusive("all",
"closed", "deferred")` — conflicting flags now error (live-verified + test);
(2) `openFindingRe` is `(?i)` — hand-edited "**Status:** Open" counts, the
word-boundary guard still blocks "opened-" (test); (3) `isTerminal` uses
`term.IsTerminal` (isatty) instead of ModeCharDevice — /dev/null no longer
gets ANSI; (4) `visibleWidth`/`truncate` count DISPLAY CELLS via
`x/ansi.StringWidth`/`Truncate` — CJK/emoji no longer shift columns (tests);
(5) description cap counts characters, not bytes — multibyte descriptions get
the full 150 (test; error message now says "chars"); (6) the bar fill
arithmetic lives once in `theme.BarFill`, drawn by both `render.Style.Bar`
and `tui.miniBar` (table test).
