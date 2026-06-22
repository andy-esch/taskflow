---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: writeTable only shrinks the last column, so a wide slug/component pushes a human-output row past the terminal width and wraps.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [render]
created: "2026-06-22"
---
# Render table width discipline for wide non-final columns

## Objective

writeTable only caps the last column; a long slug/epic-id/component in a narrow
terminal overflows maxWidth and wraps (FindingsHuman is most exposed: six columns +
free-text Component). When the fixed columns already overflow, elide/truncate
lower-priority middle columns or clamp the composed line with ansi.Truncate,
mirroring the TUI clamp discipline. Machine formats stay uncapped (contract safe).

## Audit reference

planning/audits/open/2026-06-22-code-quality-architecture.md — **M7**. Human output only; the byte-stable table/csv/json contract is unaffected.

## Acceptance criteria

- [ ] No emitted human-table line exceeds the terminal width when a non-final cell is wide (test).
- [ ] Machine formats (table/csv/json) unchanged.
- [ ] just test + just lint green.

## Implementation plan

**Approach.** Add a final "clamp the composed line to maxWidth" pass in `writeTable`
(render/style.go), mirroring the TUI's clamp discipline (`ansi.Truncate` of the whole
joined row), and keep the existing shrink-the-last-column logic as the *first* line of
defense. The simplest correct guarantee is: after building each row's string, if its
visible width still exceeds `maxWidth`, truncate the whole line with `ansi.Truncate(…,
maxWidth, "…")`. This is preferable to the audit's "elide lower-priority middle
columns" option — column elision changes the table's shape/meaning and risks dropping
the slug/code a user needs, whereas a hard clamp guarantees the no-wrap invariant with
zero column-priority policy to bikeshed. `writeTable` is only reached by the *human*
renderers (`TasksHuman`/`FindingsHuman`/`AuditsHuman`/`EpicsHuman`/`SummaryHuman`); the
machine paths go through `WriteTablePlain`/`WriteCSV`/`*JSON` (columns.go) and are
untouched.

**Steps.**
1. In `writeTable` (style.go:153–221), keep the existing last-column shrink
   (lines 176–199). Then in the inner `write` closure (lines 201–214), after building
   `b.String()` and `TrimRight`-ing it, add: `if maxWidth > 0 { line =
   ansi.Truncate(line, maxWidth, "…") }` before `Fprintln`. `ansi.Truncate` is already
   imported (used by `truncate`) and is ANSI-aware, so a colored non-final cell (e.g.
   `st.Bold(slug)`, `st.Status(...)`) is clipped without severing an escape — the
   existing `truncate` helper bails on ANSI strings, but `ansi.Truncate` here is the
   whole-line clamp the TUI uses (`MaxWidth`), so it's safe on a composed colored row.
2. Confirm the header row is clamped the same way (it goes through the same `write`
   closure, so it is). `maxWidth <= 0` (piped output) skips the clamp — full-width rows
   preserved.

**Tests.** Add to `internal/cli/render/style_test.go` a case feeding `writeTable` (or a
`FindingsHuman` call) a row with a very long non-final cell (a long `Component`/`slug`)
and a small `maxWidth`, asserting **every emitted line's `ansi.StringWidth` ≤
maxWidth** (the invariant the audit asks for) — `FindingsHuman` is the most exposed
(six columns before TITLE + free-text Component). Add a colored-cell variant to prove
the ANSI clamp doesn't leave a dangling escape (assert the line ends with a reset or
re-strip and re-measure). Also assert `maxWidth == 0` leaves a wide row untouched
(piped contract).

**Risks / gotchas.** (a) **Golden snapshots** — the byte-stable `--json`/`-o table`/
`csv` goldens are NOT affected (those paths don't call `writeTable`), but any human-mode
golden or output-coverage test that renders a narrow table could shift; run
`go test ./internal/cli` and, if a human-table fixture legitimately changes, regenerate
with `go test ./internal/cli -update` and eyeball the diff (the clipped row is the
intended new behavior). (b) `ansi.Truncate` truncating mid-escape would corrupt color —
it's ANSI-aware so it won't, but the colored-cell test guards against a regression if
the helper is ever swapped. (c) The last-column shrink already truncates the final cell;
the new clamp is a backstop for the *non-final* overflow case the audit describes —
don't remove the shrink (it produces nicer output when only the description is long).
(d) Wide runes (CJK/emoji) — `ansi.StringWidth`/`ansi.Truncate` count display cells, so
the invariant holds for them too; include one in the test if cheap.

**Done when.** No `writeTable`-emitted human line exceeds `maxWidth` even with a wide
non-final cell (locked by a test), the machine table/csv/json formats are byte-for-byte
unchanged, and `go build ./...`, `go test ./...`, `golangci-lint run ./...` are green.
