---
status: ready-to-start
epic: 17-pm-go-cli
description: Add --dry-run to task move/set/new, epic new, and audit close/reopen/defer so agents/scripts can preview a mutation without writing
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, agent-ux]
created: "2026-06-09"
---

# Global --dry-run for mutating commands

## Objective

`lint --fix` has `--dry-run`, but the mutating verbs don't. Agents and sync
pipelines want to preview a write ("where would this move? what would it set?")
without touching disk. Add a global `--dry-run` to the mutating commands.

## Scope

Mutations to cover: `task move` + the six transition verbs, `task set`,
`task new`, `epic new`, `audit close|reopen|defer`.

## Implementation sketch

- [ ] Persistent `--dry-run` bool on the root (like `--json`), surfaced on
      `*cli.App`.
- [ ] Thread it into the store mutators as a flag, or (cleaner) add a parallel
      "compute the result without writing" path. Preference: pass `dryRun` to
      `Move`/`SetFields`/`CreateTask`/`CreateEpic`/`MoveAudit` and short-circuit
      the write after building the new content/destination — so the returned
      `domain.Task`/`Epic`/path reflect what *would* happen.
- [ ] Render notes the preview (e.g. `would move alpha -> in-progress`,
      `would create tasks/ready-to-start/x.md`). JSON envelope carries
      `"dry_run": true`.
- [ ] Decide validation semantics: `--dry-run` must still run all validation
      (epic exists, conflict check, transition legality) so the preview is
      truthful — i.e. a dry-run that *would* fail should return the same error.
- [ ] Tests: each mutation under `--dry-run` leaves the filesystem unchanged but
      reports the intended outcome; a would-fail dry-run still errors.

## Out of scope

- A diff/patch preview of file contents — a one-line intent summary is enough.

## Related

- Epic [[17-pm-go-cli]]; raised in the agent-readiness / sync-resilience review.
  Sibling of the deferred advisory-`flock` work.
