---
status: ready-to-start
epic: 17-pm-go-cli
description: Create refuses a slug that already exists in any bucket, not just the target dir; bundles audit-new follow-ups (--body test, area raw-vs-slug)
effort: Unknown
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, core, validation]
created: "2026-06-16"
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

- [ ] `task new` / `audit new` reject a slug already present in **any** bucket
      (not just the target dir), returning ErrConflict (exit 14) and writing
      nothing. The check runs under `--dry-run` too.
- [ ] The collision scan lives in the store (it owns the on-disk layout) and
      reuses the existing per-entity candidate scan (`markdownCandidates` /
      `resolveAudit`-style enumeration) rather than a second path convention.
- [ ] Decide + implement epic behavior: either accept duplicate name-slugs
      (status quo, document it) or warn on create that `<slug>` is now
      fuzzy-ambiguous.
- [ ] Tests: create-into-occupied-other-bucket is ErrConflict for task and
      audit; same-slug-different-bucket no longer round-trips into an ambiguous
      `show`.

## Also in scope (audit-new review follow-ups)

Small items surfaced reviewing the `audit new` work — cheap to fold in here:

- [ ] **`audit new --body` override is untested** — add a case asserting a
      provided `--body` replaces the scaffold (mirrors the untested task/epic
      `--body`; close all three or just audit).
- [ ] **`area` raw vs slug** — `audit new "Arch Data Flow"` stores
      `area: Arch Data Flow` while the slug is `…-arch-data-flow`. Decide: keep
      raw (mirrors task title-vs-slug, fine for token areas) or normalize `area`
      to the slug so they always match. Low stakes; just pick one and pin it.

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
