---
schema: 1
status: next-up
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: 'Generalize path-CAS into version-CAS: reads return a content hash, writes take ifVersion and return ErrConflict; internal auto-retry for field-level set/append/move. Per epic 24.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [core, storage]
created: "2026-07-01"
id: 6fhnydm02wxd
---

# Version-aware OCC: content-hash token and plain retry

## Objective

Generalize the store's path-CAS into a **version-CAS**: reads return a `version`,
writes take an `ifVersion` precondition and fail with `ErrConflict` (exit 14) on
mismatch. This makes the existing conflict guard honest for *content* edits, not just
concurrent moves, and is the backend-agnostic foundation every later backend (git-sync,
object store, `serve`) rides on.

The version token is a **whole-file content hash (SHA-256)** — decided under epic 24
(2026-07-01) and **content-hash only**. A git **blob/commit SHA is explicitly not used**:
it fingerprints *committed* bytes and misses uncommitted working-tree edits (the serve
write window), so it would yield a token already stale against the file on disk.
`mtime+size` is also rejected (unreliable across machines/containers — a real risk with
the cron agents). Cost is ~zero: the read-modify-write already reads the file.

## Acceptance criteria

- [ ] Store reads return an opaque `version` = **SHA-256 of the file's current bytes**
      (working-tree/live content), never a git blob SHA.
- [ ] Store writes accept `ifVersion`; on mismatch return `domain.ErrConflict` (exit 14),
      the *same* sentinel today's path-CAS produces. `ifVersion == ""` means
      create-must-not-exist (the `O_EXCL` case).
- [ ] Scriptable field-level mutations (`set` / `append` / `move`) do a **bounded internal
      auto-retry** (re-read → re-apply → rewrite) so agents don't each reimplement it.
- [ ] The **human whole-file `edit`** path surfaces `ErrConflict` instead of retrying —
      they edited a specific version and should see the clash.
- [ ] Existing path-CAS behavior (concurrent-move detection) is preserved.

## Out of scope

- **Merge / conflict-resolution UX** — plain retry only (per epic 24); "assist a merge"
  is deferred.
- **The HTTP `ETag` / `If-Match` / 412 surface** — that's the web adapter (epic
  [[19-web-companion-apps-over-a-shared-core]]), which *carries* this content-hash token
  over HTTP; it is not defined here.
- **Git-sync / remote backends** — this task is the local FS implementation of the
  version-CAS foundation; remotes supply their own preconditions on top of the same port.

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]
- [[2026-06-24-remote-planning-repos-backends-and-sync]] — §2 sync/OCC context (its
  per-backend token table is superseded by this content-hash decision).
- Epic [[19-web-companion-apps-over-a-shared-core]] — the HTTP surface over this token.
