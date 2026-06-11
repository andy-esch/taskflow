---
status: ready-to-start
epic: 17-pm-go-cli
description: Rune-safe slug truncation, .git-as-file boundary (worktrees/submodules), and inline-comment-tolerant .tskflwctl.toml root parsing
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [go, cli, robustness]
created: "2026-06-11"
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

## Acceptance criteria

- [ ] A non-ASCII title over the cap produces a valid-UTF-8 slug (rune test).
- [ ] Discovery stops at a `.git` *file* boundary (worktree/submodule sim).
- [ ] `taskflow_root = "x" # c` resolves to `x`. Suite + lint green.

## Out of scope

- Reworking slug rules beyond the truncation fix (pm parity stays).
- Pulling in a real TOML library for one string key.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/domain/slug.go`, `internal/config/config.go`.
