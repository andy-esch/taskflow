---
status: ready-to-start
epic: 17-pm-go-cli
description: Mutually-exclusive audit list flags, case-insensitive audit Status regex, term.IsTerminal for color, and runewidth column alignment
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [go, cli, polish]
created: "2026-06-11"
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

- [ ] `audit list --closed --deferred` errors instead of silently ignoring one.
- [ ] Tables stay aligned with a wide-char (CJK/emoji) cell.
- [ ] Whichever of 2/3/5 the reviewer accepts have a test. Suite + lint green.

## Out of scope

- Anything behavioral beyond these nits; the broader audit finding-level
  commands (their own deferred work).

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/cli/audit.go`, `internal/store/auditstore.go`,
  `internal/cli/color.go`, `internal/cli/render/style.go`,
  `internal/domain/validate.go`.
