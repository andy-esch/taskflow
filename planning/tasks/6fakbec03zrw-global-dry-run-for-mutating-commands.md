---
status: completed
epic: 17-pm-go-cli
description: Add --dry-run to task move/set/new, epic new, and audit close/reopen/defer so agents/scripts can preview a mutation without writing
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, agent-ux]
created: "2026-06-09"
updated_at: "2026-06-13"
started_at: "2026-06-13"
completed_at: "2026-06-13"
id: 6fakbec03zrw
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

- [x] Persistent `--dry-run` bool on the root (like `--json`), surfaced on
      `*cli.App`.
- [x] Thread it into the store mutators as a flag, or (cleaner) add a parallel
      "compute the result without writing" path. Preference: pass `dryRun` to
      `Move`/`SetFields`/`CreateTask`/`CreateEpic`/`MoveAudit` and short-circuit
      the write after building the new content/destination — so the returned
      `domain.Task`/`Epic`/path reflect what *would* happen.
- [x] Render notes the preview (e.g. `would move alpha -> in-progress`,
      `would create tasks/ready-to-start/x.md`). JSON envelope carries
      `"dry_run": true`.
- [x] Decide validation semantics: `--dry-run` must still run all validation
      (epic exists, conflict check, transition legality) so the preview is
      truthful — i.e. a dry-run that *would* fail should return the same error.
- [x] Tests: each mutation under `--dry-run` leaves the filesystem unchanged but
      reports the intended outcome; a would-fail dry-run still errors.

## Out of scope

- A diff/patch preview of file contents — a one-line intent summary is enough.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md); raised in the agent-readiness / sync-resilience review.
  Sibling of the deferred advisory-`flock` work.

## Closure (2026-06-13)

Implemented the cleaner option from the sketch: `dryRun bool` threaded into the
store mutators (`Move`/`SetFields`/`CreateTask`/`CreateEpic`/`MoveAudit`) plus
`config.Init`, each short-circuiting AFTER all validation (resolve,
parse-before-commit, CAS/collision, epic existence, transition target) but
before the disk write — so a would-fail dry-run returns the identical error
(verified: bad epic → exit 11 under `--dry-run`). Persistent `--dry-run` on the
root, surfaced on `*cli.App`, honored by task new/set/move + the six transition
verbs, epic new, audit close/reopen/defer, and init (also folds into
`lint --fix`). Human output says "would create / would update / would move /
would initialize"; the JSON envelopes gained `dry_run: true` (schema bumped
1.1 → **1.2**). The `→ next:` hint is suppressed on previews. Tests:
`internal/cli/dryrun_test.go` (each mutation leaves the FS unchanged but
reports the outcome; would-fail still errors). README + ARCHITECTURE updated
(dropped `--dry-run` from the "remaining" list, and the already-shipped JSON
error envelope with it). Suite, vet, golangci-lint, gofmt all green;
live-verified against the planning repo.
