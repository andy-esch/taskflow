---
status: completed
epic: 17-pm-go-cli
description: Rune-safe slug truncation, .git-as-file boundary (worktrees/submodules), and inline-comment-tolerant .tskflwctl.toml root parsing
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [go, cli, robustness]
created: "2026-06-11"
updated_at: "2026-06-13"
started_at: "2026-06-13"
completed_at: "2026-06-13"
---

# Discovery and slug edge-case robustness

> ⚠️ **Externally proposed — needs independent review before implementing.**
> Filed from an outside code-review pass, verified against the code by the
> filing agent. Low-severity edge cases; the implementing agent should confirm
> each is worth its change before coding.

## Objective

Three small, independent robustness gaps in discovery + slug generation. None
bite the current self-hosted repo (English titles, a normal `.git` dir,
own-line config comments), but each is a cheap, correct hardening.

1. **Rune-unsafe slug truncation** (`internal/domain/slug.go:27`). `Slugify`
   caps with `text[:80]` — a byte index. `slugPunct` strips only ASCII
   punctuation, so a non-ASCII title (CJK/accents/emoji) keeps multibyte runes;
   slicing at byte 80 can land mid-rune → an invalid-UTF-8 slug → bad filename.
   *(Note: contrary to the review, this does not panic — Go slices/prints invalid
   UTF-8 without panicking; it just produces a corrupt slug.)* Fix: truncate on a
   rune boundary (then still trim back to the last `-`).

2. **`.git` boundary misses worktrees/submodules** (`internal/config/config.go:39`).
   The climb stops at `isDir(".git")`, but in a git **worktree or submodule**
   `.git` is a *file* pointing elsewhere, so the boundary isn't seen and
   `Discover` over-climbs past the repo — risking a false-positive match on a
   parent dir's `tasks/`/`.tskflwctl.toml`. Only triggers when no planning root
   is found at/under the start. Fix: treat `.git` as a boundary if it exists as
   **file or dir** (`os.Stat`, not `IsDir`).

3. **TOML root parse breaks on an inline comment** (`internal/config/config.go:58`
   `taskflowRoot`). `taskflow_root = "./planning" # note` yields the path
   `./planning" # note` (trailing-quote trim doesn't reach it). The default
   config writes comments on their own lines, so this only bites a hand-edited
   file — but inline comments are valid TOML. Fix: strip a `#`-to-EOL comment
   (outside quotes) before extracting the value.

Folded in from the 2026-06-12 review
([[2026-06-12-critical-code-review-multi-lens]], findings B2/M2 + a slug low —
same files, same theme):

4. **`configuredRoot` doesn't enforce the containment its comment promises**
   (`internal/config/config.go:48-54`). The comment says the result is "kept
   within dir's tree" but the code is plain `filepath.Join` —
   `taskflow_root = "../../elsewhere"` escapes. Reject a cleaned path that
   escapes `dir`, or fix the comment.
5. **A misconfigured `taskflow_root` silently forks the data tree**
   (`config.go:51-54` + `store/fsstore.go:50-53`). `Discover` never checks
   the configured root exists or contains `tasks/`; missing dirs read as
   empty, and `task new` `MkdirAll`s a fresh tree at the bogus root —
   planning data split across two roots with no signal. Verify the
   configured root looks like a planning root (or warn loudly).
6. **Slugs/epic-ids aren't sanitized against path separators**
   (`store/fsstore.go:181`, `epicstore.go:47`, `auditstore.go:112`).
   `epic show ../tasks/in-progress/x` reads outside the intended directory.
   Reject names containing separators or `..` in `resolve`/`GetEpic`.

## Acceptance criteria

- [x] A non-ASCII title over the cap produces a valid-UTF-8 slug (rune test).
- [x] Discovery stops at a `.git` *file* boundary (worktree/submodule sim).
- [x] `taskflow_root = "x" # c` resolves to `x`. Suite + lint green.
- [x] An escaping or nonexistent `taskflow_root` errors (or warns) instead of
      silently presenting an empty project.
- [x] A slug containing `/` or `..` is rejected with `ErrValidation`.

## Out of scope

- Reworking slug rules beyond the truncation fix (pm parity stays).
- Pulling in a real TOML library for one string key.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/domain/slug.go`, `internal/config/config.go`.

## Closure (2026-06-13)

All six items shipped: rune-boundary slug truncation (long CJK titles produce
valid-UTF-8 slugs, dash-trimmed; tested); `.git`-as-file climb boundary
(worktrees/submodules; tested against a parent-dir trap); proper TOML value
extraction (quoted segment wins, inline `#` comments stripped, unterminated
quote treated as unset; table-tested); B2 containment (an escaping
taskflow_root is a loud error); M2 planning-root validation (a configured
root without tasks/ errors instead of presenting an empty project that
`task new` would fork); slug path-separator rejection landed inside the
shared resolver (see [[fuzzypartial-slug-resolution]]).

### Addendum (2026-06-13, same day)

Dogfooding feedback ("epic new doesn't strip em-dashes — got
`19-…-—-…`") exposed that the denylist approach doesn't converge, so
`Slugify` was REWRITTEN as an allowlist with separator collapse: keep
letters/digits/marks (any script) + interior dots; apostrophes vanish
silently (don't → dont); EVERY other rune — unicode dashes, smart quotes,
symbols, whitespace — is a word break; runs collapse; rune-safe 80-byte cap.
Deliberate behavior change for NEW slugs only: punctuation now breaks words
(UI/UX → ui-ux, formerly uiux) — the old strip-without-break behavior was
retired-Python-pm parity. Tests cover the em-dash report, smart quotes,
symbols, apostrophes, and version dots.
