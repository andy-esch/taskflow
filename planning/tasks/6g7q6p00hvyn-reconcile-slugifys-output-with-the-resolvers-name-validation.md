---
schema: 1
id: 6g7q6p00hvyn
status: next-up
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: A title containing '..' mints a task whose slug the resolver refuses, so the record is unreachable by name and the CLI's own next-step hint fails.
effort: 1-2 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, store, robustness, storage]
created: "2026-09-07"
---
# Reconcile Slugify's output with the resolver's name validation

`Slugify` can emit a slug that `validQueryName` rejects, so a task is created
successfully and then cannot be addressed by its own slug.

## Reproduction

```
$ tskflwctl task new "Wait... what happens here" --epic 01-probe --tags p --description d
created tasks/6g7q6bmxwpht-wait...-what-happens-here.md
→ next: tskflwctl task start wait...-what-happens-here

$ tskflwctl task start wait...-what-happens-here
✘ wait...-what-happens-here: validation failed: task name "wait...-what-happens-here"
  must be a plain name (no path separators)
error: 1 of 1 transitions failed: ...                                      # exit 11
```

The record is reachable only by its 12-char id (`task show 6g7q6ehqt04b` and
`task start 6g7q6ehqt04b` both work). `tskflwctl lint` reports the tree clean, so
nothing surfaces the problem.

Note the failing command is the one **the tool itself just printed** as the next step.

## Cause

Two rules that were never reconciled:

- `domain.Slugify` (`internal/domain/slug.go:22`) keeps `.` on purpose — the doc
  comment cites version numbers — and trims dots only at the ends. A run of dots in
  the middle survives intact.
- `store.validQueryName` (`internal/store/resolve.go`) rejects any query containing
  `..`, as a path-traversal guard.

So `"Wait... what"` → `wait...-what`, which is a legal filename and an illegal query.

`..` is the *only* gap. Slugify's output alphabet is letters, digits, marks, `.`,
and `-`; `/` and `\` are already word breaks, and an all-punctuation title is
rejected up front (`title produced an empty slug`). So this is a one-character-class
fix, not an open-ended class of defects.

## Scope — single dots are fine

A single mid-slug dot resolves correctly:

```
$ tskflwctl task show rename-foo.md-to-bar.md      # exit 0
```

So a title like "Consolidate agent instructions on AGENTS.md across tracked repos"
(→ `...-on-agents.md-across-tracked-repos`) is **cosmetic only** — ugly in a
filename, functionally sound. That case is what prompted this report; the `..` case
is the actual defect found while checking it.

## Fix options

1. **Collapse runs of `.` to one in `Slugify`.** Minimal, preserves the documented
   version-number intent, kills the only gap. Preferred.
2. **Make `.` a word break unless between digits.** Strictest reading of "dots are
   for version numbers", but changes more existing slugs.
3. **Narrow `validQueryName` to reject `..` only as a path segment.** Riskier — that
   guard is a traversal defense and weakening it deserves its own justification.

Whichever is chosen, the durable protection is a property test asserting that
`validQueryName(Slugify(s))` holds for arbitrary `s`, so the two rules cannot drift
apart again.

## Acceptance criteria

- [ ] A title containing `..` or `...` produces a slug that resolves by name.
- [ ] Version-number slugs still round-trip (`"Upgrade to Go 1.24"` →
  `upgrade-to-go-1.24`, resolvable).
- [ ] A property/table test asserts every `Slugify` output is a valid resolver query,
  covering ellipses, digit ranges (`1..2`), leading/trailing dots, and non-ASCII.
- [ ] Existing slugs in `planning/` are unaffected, or the change is called out as a
  rename with the affected files listed.

## Out of scope

- Cosmetics of single dots in filenames (`agents.md-across-...`) — intentional per
  the Slugify policy, and it resolves correctly.
- Any change to how ids are minted or to the flat id-led filename layout.
