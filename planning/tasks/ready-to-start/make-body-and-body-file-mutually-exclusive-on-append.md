---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: --body/--body-file are marked exclusive only on `task new`; on `task append` and `audit append` passing both silently prefers --body-file. Enforce the exclusion on every body-taking command.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [cli]
created: "2026-06-28"
---
# Make `--body` / `--body-file` mutually exclusive on `task append` + `audit append`

## Objective

`resolveBody` (`internal/cli/task.go`) documents the two body flags as "mutually
exclusive (enforced by the command), so at most one is set". That's true for
`task new` — it calls `cmd.MarkFlagsMutuallyExclusive("body", "body-file", "template")`.
But `task append` and `audit append` do **not** mark them exclusive, so passing
both `--body` and `--body-file` is silently accepted: `resolveBody` checks
`bodyFile == ""` first, so `--body-file` quietly wins and `--body` is dropped
with no error. The doc comment is a lie on those paths, and silent flag
precedence is a footgun (the user thinks `--body` took effect).

Enforce the exclusivity on every body-taking command (or, if we'd rather keep it
lenient, make `resolveBody` itself reject the both-set case so the contract holds
regardless of which command calls it) — and correct the comment to match.

## Context

Pre-existing; surfaced by the 2026-06-28 adversarial review of the audit-edit
work ([[audit-editing-faces-audit-edit-set-and-append]]) — `audit append`
inherited the gap by mirroring `task append`. Touches the shared task path, so it
was filed rather than folded into that feature PR.

Relates to epic 20 (CLI UX / ergonomics).

## Acceptance criteria

- [ ] Passing both `--body` and `--body-file` to `task append` and `audit append`
      is a clean `ErrValidation` (exit 11), not a silent drop — mirroring
      `task new`'s behaviour.
- [ ] `resolveBody`'s doc comment is accurate for every caller (either the guard
      is universal, or the comment names where it's enforced).
- [ ] Tests pin the both-set rejection on both append commands. go
      build/test/lint green; docs/cli regenerated if help text changes.

## Implementation sketch

- Simplest + consistent: add `cmd.MarkFlagsMutuallyExclusive("body", "body-file")`
  to `newTaskAppendCmd` and `newAuditAppendCmd` (cobra emits the exit-11 error
  for free, same as `task new`).
- Belt-and-suspenders alternative: also have `resolveBody` return
  `ErrValidation` when `body != "" && bodyFile != ""`, so the invariant is
  guaranteed at the single choke point no matter who calls it — then the comment
  is unconditionally true. Pick one; document the choice.

## Risks / gotchas

- `task new` already marks the trio (`body`/`body-file`/`template`) exclusive —
  don't double-register there.
- Keep the message style consistent with the existing exclusive-flag errors so
  the agent-facing contract stays uniform.

## Done when

`--body` + `--body-file` together is a uniform validation error on every command
that takes a body, and `resolveBody`'s comment no longer overstates the guarantee.
