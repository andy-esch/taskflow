---
status: completed
epic: 17-pm-go-cli
description: Acted on the pertinent items from an external code review; rejected the off-the-mark ones with rationale
effort: Unknown
tier: 2
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, robustness]
created: "2026-06-09"
updated_at: "2026-06-09"
completed_at: "2026-06-09"
---

# Triage external review: fix fixer-comment, O_EXCL, config anchor, diagnostics

## Objective

Triage an external code review against the actual code, fix the pertinent
findings, and reject the off-the-mark ones with a recorded rationale.

## Fixed (verified real)

- **`lint --fix` corrupted inline comments** (`store/fix.go`): a colon inside a
  trailing `# comment` made `fixValue` quote the whole remainder. Now splits the
  comment off first (`splitInlineComment`) and re-appends it outside the value.
- **Exclusive create portability** (`store/atomic.go`): `createFileAtomic`
  switched from `os.Link` (hard-link — restricted on many container/NFS/VM
  mounts) to portable `os.OpenFile(O_CREATE|O_EXCL)`. Still race-free; collision
  still → `ErrConflict`.
- **`.tskflwctl.toml` was write-only** (`config/config.go`): `Discover` now uses
  it as the explicit anchor marker and honors `taskflow_root` (minimal scan, no
  TOML dep); `tasks/` heuristic stays as fallback. Default config comments
  clarify `tracked_repos` is reserved.
- **Diagnostic key corruption** (`store/diagnose.go`): `fieldOnLine` now guards
  with `isIdentifier`, so a list item / quoted value with an inner colon falls
  back to a generic line-number message instead of a junk "key".
- **Ambiguous-match error** (`store/fsstore.go`): now names the directories
  holding the duplicates (`matches 2 tasks (in ready-to-start, in-progress)`).
- Stale comment in `main.go` (exit codes now 10–14).

## Rejected (off the mark — verified)

- **"main.go bypasses exit codes with os.Exit(1)"**: false — `main.go:15`
  already calls `os.Exit(cli.ExitCode(err))`; the conflict demo exits 14.
- **"add a `completion` subcommand"**: already shipped (cobra `completion` +
  `just completion-zsh`, with slug completion).
- **"delete the TUI stub"**: intentionally retained — the TUI is a planned phase
  to be rebuilt over the current `core`; documented as a parked sketch.
- **"add a read cache / SQLite index for thousands of files"**: premature
  (YAGNI) — we just deleted exactly this (the `internal/index` JSON fast-index).
  Markdown stays the source of truth; revisit only if scale demands it.

## Deferred (good, but their own work)

- **Advisory `flock` on mutations** (concurrent agent loops) — already on the
  roadmap; needs a dep + careful integration.
- **`schema` exporter** (emit task/epic frontmatter schema as JSON for agents)
  — genuinely useful and agent-aligned; recommend as the next task.

## Acceptance

- [x] Each accepted fix has a test (fixer-comment, config anchor + taskflow_root,
      ambiguous-names-dirs, audit fence/open-ish); full suite + lint green.
- [x] Demoed `lint --fix` keeps an inline comment; discovery still works on the
      config-less `planning/` via fallback.

## Related

- Epic [[17-pm-go-cli]]; follows
  [[harden-create-loop-and-fill-docs-post-pm-retirement-review]].
