---
status: completed
epic: 17-pm-go-cli
description: Create refuses a slug that already exists in any bucket, not just the target dir; bundles audit-new follow-ups (--body test, area raw-vs-slug)
effort: Unknown
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, core, validation]
created: "2026-06-16"
updated_at: "2026-06-17"
started_at: "2026-06-17"
completed_at: "2026-06-17"
id: 6fcvejg03y38
---

# Reject cross-bucket slug collisions on create (task/audit)

## Objective

`CreateTask`/`CreateAudit` only refuse a clobber of the *exact target path* (via
`createFileAtomic` O_EXCL), i.e. a collision within the one status/bucket dir
they write to. A slug that already exists in a **different** bucket sails through
— producing two files that share a slug, which breaks the slug's job as the
unique id (`resolveID` then returns ErrAmbiguous). Creation should reject a slug
that exists in *any* bucket.

Verified 2026-06-16 against `internal/store/create.go`:

```
task new "Dup Task"      # → tasks/ready-to-start/dup-task.md
task complete dup-task   # → tasks/completed/dup-task.md
task new "Dup Task"      # exit 0 — creates ready-to-start/dup-task.md again
task show dup-task        # exit 13 (ambiguous): two files, one slug
```

Same for `audit new` (a slug in `closed/`/`deferred/` doesn't block a new
`open/` create). Epics are safe from *exact* collision (auto-numbered ids are
always fresh) but duplicate name-slugs still make `epic show <slug>` fuzzy-
ambiguous — decide whether to warn.

## Acceptance criteria

- [x] `task new` / `audit new` reject a slug already present in **any** bucket
      (not just the target dir), returning ErrConflict (exit 14) and writing
      nothing. The check runs under `--dry-run` too.
- [x] The collision scan lives in the store (it owns the on-disk layout) and
      reuses the existing per-entity candidate scan (new `slugCollision` over
      `taskCandidates()` / new `auditCandidates()`) rather than a second path
      convention.
- [x] Epic behavior: **accept** duplicate name-slugs (the auto-numbered id is
      always fresh → no exact collision; only `epic show <bare-slug>` goes
      fuzzy-ambiguous, recoverable via the full NN-slug). Documented in
      `CreateEpic`.
- [x] Tests: create-into-occupied-other-bucket is ErrConflict for task and
      audit; same-slug-different-bucket still resolves through `show`.

## Also in scope (audit-new review follow-ups)

Small items surfaced reviewing the `audit new` work — cheap to fold in here:

- [x] **`audit new --body` override** — now tested (`TestAuditNew_BodyOverride`:
      a provided `--body` replaces the scaffold).
- [x] **`area` raw vs slug** — decided: **keep raw** (mirrors task title-vs-slug;
      fine for the token areas the routines use). No code change.

## Progress Log

- **2026-06-17**: Implemented. `slugCollision` (store/resolve.go) checks an exact
  slug across all of an entity's dirs; `CreateTask`/`CreateAudit` call it before
  writing (reusing `taskCandidates()` / extracted `auditCandidates()`), so a slug
  in any other bucket → ErrConflict (exit 14), under `--dry-run` too. Epics
  unchanged (accept dup name-slugs, documented). Folded in the two `audit new`
  review nits. 4 new tests; suite + lint + vet green.

## Out of scope

- Renaming/auto-suffixing a colliding slug (`-2`) — reject, don't mangle. The
  user picks a different title.
- Cross-bucket collisions arising from a manual `mv` (not a create) — lint's
  job, not create's.

## Related

- Epic [[17-pm-go-cli]]
- Surfaced by the `audit new` parity work (commit b98dfbd) and its review sweep.
- Touches `internal/store/create.go` (`writeNewFile` + the per-entity Create*),
  `internal/store/resolve.go`/`auditstore.go` (candidate enumeration).
