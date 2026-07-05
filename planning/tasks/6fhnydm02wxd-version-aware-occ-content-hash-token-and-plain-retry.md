---
schema: 1
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: 'Generalize path-CAS into version-CAS: reads return a content hash, writes take ifVersion and return ErrConflict; internal auto-retry for field-level set/append/move. Per epic 24.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [core, storage]
created: "2026-07-01"
id: 6fhnydm02wxd
updated_at: "2026-07-04"
started_at: "2026-07-04"
completed_at: "2026-07-04"
---

# Version-aware OCC: content-hash token and plain retry

## Progress & pivots (2026-07-04)

**Status: implementation complete (steps 1–7).** Green throughout — store + core suites,
vet, golangci-lint (0 issues), live smoke. Two small items are tracked separately so they
outlive this task: the dry-run↔CAS consistency decision
([normalize-dry-run-vs-version-cas-ordering-across-store-writes](6fjt7rm1p50q-normalize-dry-run-vs-version-cas-ordering-across-store-writes.md)) and hardening
fix.go's relocations ([harden-lint-fix-misfiled-move-for-dup-slug-edge-cases](6fjjpfg16ss5-harden-lint-fix-misfiled-move-for-dup-slug-edge-cases.md)).

**⚠ Post-"completion" critical fix (2026-07-04): the flock write-lock.** A concurrency
*smoke test* (16 processes appending to one file) exposed a serious bug the four review
passes missed: version-CAS alone did NOT prevent lost updates — 8–12 of 16 concurrent
writes silently clobbered each other with ZERO conflicts, because the verify→write wasn't
atomic (the verify→rename window, widened to ms by the temp-file fsync — both writers pass
their verify before either renames). Fixed by the planned **advisory `flock`** (`writeLock`,
repo-wide, unix; no-op stub elsewhere) around the verify→write critical section, making the
CAS atomic; the existing retry then heals the now-*detected* conflicts. Post-fix smoke:
LOST=0 (0 of N lost), a couple of surfaced exit-14s under 16-way contention (correct).
Regression pinned by `store.TestConcurrentAppends_NoLostUpdates` (fails without the lock).
Lesson: the hook-based tests only exercised a concurrent edit landing *before* verify (the
detectable case); nothing tested the between-verify-and-rename window until the smoke did.

Pivots taken vs the original plan:
- **The token is fully INTERNAL — the `core.Store` port never changed.** The hybrid
  (tool-carries-the-token) decision means reads don't return a `version` and writes
  don't take an `ifVersion` at the port; the write methods self-source
  `hashContent(content)` and check it internally. So **step 2 (port-threading)
  collapsed into step 3** — no fake/call-site churn. Port-level exposure is deferred to
  `serve` (epic 19).
- **A shared `verifyUnchanged` GUARD, not a `casWrite`-that-writes.** The write shapes
  differ (in-place vs write-then-remove relocate), so the reusable piece is the
  pre-write precondition; each site keeps its own write.
- **Canonical-slug re-resolve (review-driven, pre-existing bug fix).** The guard
  re-resolves by the canonical slug, not the caller's raw fuzzy query — else a
  concurrently-created same-prefix file made the re-resolve `ErrAmbiguous` → a spurious
  conflict on an unmodified file. Found by an independent review; fixed once for all sites.
- **Epic writers GAINED a guard they never had** (MoveEpic/SetEpicFields/EditEpic were
  silent last-write-wins) — a bonus beyond parity.

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

## Design — grounded in prior art (research 2026-07-04)

The whole approach maps 1:1 onto HTTP preconditions (RFC 9110 §13.1) and onto how
etcd / CouchDB / S3-conditional-writes / Git-ref-CAS / Firestore actually work. Locked
decisions, each with its prior-art anchor:

- **The token is hashed ON READ and NEVER stored in the file.** A `version:` frontmatter
  field would be self-referential (the token changes the very bytes it certifies) — the
  classic trap the "ETag without a version column" pattern exists to dodge. So: compute
  `sha256(raw file bytes)` when read; never persist it. (This corrects an early
  research draft that proposed stamping `version:` — wrong for this spec.)
- **Strong validator, not weak.** Whole-file byte equality is HTTP's *strong* comparison
  (`If-Match`). For a lost-update guard that is exactly right — a normalized/"weak" token
  would let two writers who changed *different* fields both believe they're safe and
  silently drop one. Do NOT normalize-before-hash.
- **SHA-256 stays cryptographic.** The token *is* the safety check, so a collision is a
  silent lost update, and cron agents write attacker-influenceable text (task
  titles/bodies). Non-crypto hashes (xxHash/wyhash) are a correctness bug here, not an
  optimization. If body hashing ever shows on a profile, BLAKE3 is the drop-in (same
  256-bit safety, faster) — not now; `crypto/sha256` is stdlib, zero deps.
- **`ifVersion` vocabulary mirrors RFC 9110 exactly:** `ifVersion == "<hash>"` → strong
  `If-Match` (write iff unchanged); `ifVersion == ""` → `If-None-Match: *` → the
  create-must-not-exist / `O_EXCL` case (which `createFileAtomic` already is — creates
  need no new code, just the documented mapping); no precondition → unconditional write.
  Keeps the future HTTP surface (epic 19) a rename, not a redesign.
- **Version-CAS *subsumes* path-CAS; keep both checks in one guard.** Today's re-resolve
  (`curPath != path`) catches a concurrent *relocation*; the content hash catches a
  concurrent *in-place edit*. A move changes both, so a single guard —
  "re-resolve → conflict if moved/gone, then re-hash the source → conflict if hash ≠
  ifVersion" — is a strict superset of the current guard. Path-CAS is NOT deleted (it
  still guards the resurrect-in-old-dir hazard on in-place writes); it's absorbed.
- **HYBRID surface — one internal CAS primitive, two entry points** (the fork the
  research settled). Every surveyed system uses a uniform CAS *primitive* but splits
  *ergonomics* along scriptable-vs-human:
  - **Class (a) — scriptable field ops** (`task set`/`append`/`move`/`defer`, `audit
    append`/`move`, `epic set`/`move`): a **bounded internal auto-retry** loop
    (Firestore's `runTransaction` model). The caller supplies **no** version — forcing an
    agent to fetch+thread `--if-version` would make it reimplement the very read-modify-
    write we're centralizing. Each spin re-reads, re-derives the change on the fresh
    bytes, and CAS-writes; on `ErrConflict` it re-spins, bounded, then surfaces exit 14.
  - **Class (b) — human whole-file `edit`** (`task`/`audit`/`epic edit`): capture the
    version at open (before `$EDITOR`), pass it as `ifVersion` to a **single-shot** write;
    a mismatch surfaces `ErrConflict` with **no retry** (CouchDB/S3/Git model — you edited
    a specific version; silently rebasing your edit onto someone else's is wrong).
  - Both call the **same** `casWrite` primitive, so there is exactly one conflict-check
    code path (the correctness win of uniformity) without inflicting explicit tokens on
    every agent call.
- **Plain retry is safe even for the non-idempotent `append` — because the check is
  fail-BEFORE-mutate.** `writeFileAtomic` (temp+rename) means a write either fully lands
  (no error → no retry) or not at all; the CAS verify sits immediately before the rename.
  So a retry only fires when our write did *not* land, and it re-appends onto freshly-read
  bytes exactly once. No idempotency keys needed (those solve lost-ack-after-commit in
  distributed systems — not a local synchronous rename).
- **Bounded retry + full jitter** (AWS backoff-and-jitter). Two cron agents on the same
  schedule collide *every* round under pure backoff; jitter de-correlates them. Constants
  scale WAY down from the networked-DB literature — local renames are microseconds, so a
  ms-range base + small cap + a hard bound (~4) then a loud exit-14. Make the sleep/jitter
  and bound **injectable** (like `s.now()`) so tests are deterministic.

## Traps / footguns / rakes (read before writing code)

1. **The verify→rename window is real and local-FS-specific.** "Compare hash" and
   "rename" are not one syscall. Re-read + re-hash the source *immediately* before the
   rename, and accept a residual window: it's why class (a) needs the retry *loop* (treat
   a landing between verify and rename as a CAS miss and spin) and why class (b) can still
   occasionally surface a late conflict (fine — the human re-edits). Don't assume "I
   hashed it three lines ago" is safe under concurrent cron writers.
2. **Hash the RAW on-disk bytes, identically on read and on re-verify** — never the
   parsed/re-serialized form. The read side (`GetTask` etc.) and the write side's
   pre-rename verify must feed byte-identical input to `sha256`, or every write
   self-conflicts.
3. **Spurious conflicts = a writer-stability bug, not a token problem.** If our own
   frontmatter writer ever reorders keys or churns whitespace, whole-file hashing fires
   false conflicts and inflates the retry rate. The fix is the CLAUDE.md "surgical edit,
   preserve key order/comments/unknown fields" discipline — NOT weakening the token.
   Add a round-trip test: read → no-op write path must not change bytes.
4. **`ErrConflict` (retry) vs `ErrAmbiguous` (don't).** The move dual-file window
   (write-new-then-remove-old; fsstore.go:246→~254, auditstore.go:184→~191) can, on a
   crash, leave a duplicate slug → `ErrAmbiguous`, which is a *recoverable* state repaired
   by `lint`, NOT a transient conflict. Auto-retry must fire on `ErrConflict` ONLY;
   `ErrAmbiguous` propagates. version-CAS does not change this window — don't try to
   "fix" it here (it's Phase-B / OCC-adjacent but out of scope).
5. **Epics currently have NO write guard** (MoveEpic epicstore.go:65, SetEpicFields:114,
   EditEpic:155 — no re-resolve, because epics never relocate). version-CAS *adds*
   in-place-edit protection they lack today (concurrent epic edits are silently
   last-write-wins right now). This is a genuine improvement, not just parity — include
   epics.
6. **Don't grow `--json`/wire.** The `version` token is INTERNAL to store/core for this
   task; the HTTP `ETag`/`If-Match` surface is epic 19. So NO `version` field in
   TaskJSON/AuditJSON/EpicMetaJSON, NO `SchemaVersion` bump, NO golden churn (confirmed:
   the acceptance criteria never ask to surface it). Huge scope-limiter — resist the urge.
7. **List methods stay unchanged.** `ListTasks`/`ListAudits`/`ListEpics` are bulk display
   reads, not RMW entry points. Only the single-entity `Get*` reads return a `version`.
8. **The `core.Store` port does NOT change** (see Resolved decision). The token is
   internal to package `store`; the field ops self-source it and the edit path captures it
   internally, so no port method grows a `version`/`ifVersion` param and no fake needs
   touching. Port-level exposure waits for `serve` (epic 19).

## Sequenced implementation (each step builds green + is revertible)

1. **[done] Hash + `verifyUnchanged` guard (no behavior change).** Added
   `hashContent([]byte) string` (`crypto/sha256` → hex) + the single internal guard
   `verifyUnchanged(resolve, slug, path, ifVersion, noun, op)` that re-resolves the slug
   (conflict if gone/moved — preserves today's `curPath != path`) and re-reads + re-hashes
   the source (conflict if `≠ ifVersion`, when non-empty). Unit-tested in isolation. The
   per-site *write* stays in each method (in-place vs relocate differ); the guard is the
   shared precondition.
2. **[folded into 3 — see Resolved decision] Version on reads is INTERNAL.** No
   `core.Store` port change: the write methods self-source `hashContent(content)` from the
   bytes they already read. Port-level `version` returns are deferred to `serve` (epic 19).
3. **[done] Route the existing writes through `verifyUnchanged`.** All write sites now
   route through the one guard, sourcing `ifVersion` from the bytes each method just read:
   tasks (moveTask, SetFields, EditTask, EditBody), audits (MoveAudit, EditAudit,
   AppendAuditBody), and the three epic writers (MoveEpic/SetEpicFields/EditEpic — which
   had NO guard before; they gain in-place-edit protection). Adapters: `resolvePath`,
   `resolveAuditPath`, `resolveEpicPath`. Behavior identical PLUS in-place-edit detection;
   6 hook/callback-interleaved tests assert `ErrConflict` across every surface.
4. **[done] Class (a) auto-retry.** `retryOnConflict` (core/retry.go) wraps the eight
   scriptable ops (task set/append/move/defer, audit move/append, epic move/set) in
   core.Service: on `ErrConflict` it re-calls the store's self-contained RMW (re-read →
   re-derive → rewrite), bounded (`defaultMaxRetries=4`) with capped-exponential + full-
   jitter backoff (`defaultRetrySleep`, injectable via `WithRetry` so tests are instant),
   then surfaces exit 14. Dry-runs aren't retried; non-conflict errors pass through. Safe
   for append because the CAS fails pre-write (re-appends onto fresh content exactly once).
   Tests: heal-after-N, exhaustion→ErrConflict, dry-run-not-retried, non-conflict-
   passthrough, append-retries.
5. **[done — landed with step 3] Class (b) `edit` surfaces the conflict.**
   `EditTask`/`EditAudit`/`EditEpic` capture the version at the pre-editor read (via
   `editFile`'s recheck closure calling `verifyUnchanged`); a mismatch is `ErrConflict`
   (exit 14), no retry (there is no retry wrapper on edit). Epic in-place CAS added. Pinned
   by `TestEditTask_ConflictsOnConcurrentContentEdit`.
6. **[done] Creates: document the `ifVersion == ""` mapping.** `createFileAtomic`'s
   `O_EXCL` IS `If-None-Match: *` (create-must-not-exist) — documented on the helper, so
   the create path needs no separate `verifyUnchanged`.
7. **[done] Docs.** OCC note added to `docs/ARCHITECTURE.md`'s `internal/store` bullet: the
   version token (strong SHA-256, on-read, never stored), the internal/hybrid surface,
   auto-retry in core.Service, exit 14, HTTP deferred to epic 19. No `schema`/`--json`
   doc changes.

## Test pins (before it's "done")

- `hashContent` is stable across a read → no-op → read round-trip (writer-stability guard).
- `casWrite` returns `ErrConflict` on (a) a concurrent relocation and (b) a concurrent
  in-place content edit; succeeds when the source is untouched.
- Field op (`set`/`append`/`move`) auto-retries and SUCCEEDS after ONE concurrent edit
  (via the write hook), within the bound.
- **`append` under retry does not double-apply** (the key idempotency test): a conflict
  that fires before the write, then a successful retry, leaves the text appended once.
- Retry bound exhausted (persistent contention) surfaces `ErrConflict` / exit 14 — loud,
  not a spin.
- `edit` path surfaces `ErrConflict` (no retry) on a concurrent write during the editor.
- Epic in-place edit conflict is now detected (was last-write-wins).
- `ifVersion == ""` create collision → `ErrConflict`; existing path-CAS move-detection
  tests stay green (subsumed, not regressed).

## Review outcomes (2026-07-04 — steps 1/3/5)

Adversarial (3 finders) + a fully independent external review. Net: the concurrency core is
sound; NO HIGH finder alarm survived verification — "dryRun regression" (ordering byte-
identical to main), "no-op epic move fails" (that's lost-update prevention working), "stranded
dual-move clobber" (dup-slug → ErrAmbiguous guard; and out of scope), "wrong-file clobber"
(the `curPath != path` check) all rejected with evidence. Acted on:
- **[fixed] Canonical-slug re-resolve** (independent CRITICAL, pre-existing): `verifyUnchanged`
  re-resolved the raw fuzzy query, so a concurrently-created same-prefix file made it
  ErrAmbiguous → a spurious conflict on an unmodified file (`task edit billing` fails when a
  `billing-*` task appears mid-edit). Now re-resolves the canonical (exact) slug — one fix for
  all 10 sites. Guarded by `TestSetFields_FuzzyQueryDoesNotSpuriouslyConflict`.
- **[fixed] IO errors no longer masked as conflicts**: only `os.IsNotExist` → conflict; a real
  read error (EACCES/EIO) propagates.
- **[decide] dryRun ↔ CAS is inconsistent** (pre-existing): in-place writers (SetFields,
  writeBody) run the CAS on dry-run; the movers/epics skip it. Pick one philosophy; low stakes.
- **[follow-up] fix.go relocations are unguarded** — fold a version check into
  [harden-lint-fix-misfiled-move-for-dup-slug-edge-cases](6fjjpfg16ss5-harden-lint-fix-misfiled-move-for-dup-slug-edge-cases.md).

## Acceptance criteria

- [ ] Store reads return an opaque `version` = **SHA-256 of the file's current bytes**
      (working-tree/live content), never a git blob SHA.
- [ ] Store writes accept `ifVersion` (at the `casWrite` primitive + the wide-window
      edit/create paths); on mismatch return `domain.ErrConflict` (exit 14), the *same*
      sentinel today's path-CAS produces. `ifVersion == ""` means create-must-not-exist
      (the `O_EXCL` case).
- [ ] Scriptable field-level mutations (`set` / `append` / `move`) do a **bounded internal
      auto-retry** (re-read → re-apply → rewrite) so agents don't each reimplement it.
- [ ] The **human whole-file `edit`** path surfaces `ErrConflict` instead of retrying —
      they edited a specific version and should see the clash.
- [ ] Existing path-CAS behavior (concurrent-move detection) is preserved.

## Resolved decision (2026-07-04): the token is fully internal

Confirmed the **hybrid** reading, and its consequence: **the token never touches the
`core.Store` port** — not on reads, not on writes. The write methods `os.ReadFile` the
file anyway, so they self-source `ifVersion = hashContent(content)`; the retry re-calls on
`ErrConflict` with no token; the human `edit` captures the version inside `EditTask`.
Nothing in-tree consumes a `version` *returned* from `GetTask`/`GetAudit`/`GetEpic`, so
that port-level exposure is **deferred to `serve`/HTTP** (epic 19). Net: step 2's
port-threading collapses into step 3, and the `core.Store` interface + its fakes stay
untouched.

## Out of scope

- **Merge / conflict-resolution UX** — plain retry only (per epic 24); "assist a merge"
  is deferred.
- **The HTTP `ETag` / `If-Match` / 412 surface** — that's the web adapter (epic
  [19-web-companion-apps-over-a-shared-core](../epics/19-web-companion-apps-over-a-shared-core.md)), which *carries* this content-hash token
  over HTTP; it is not defined here. **No `version` in `--json`/wire, no `SchemaVersion`
  bump** in this task.
- **Git-sync / remote backends** — this task is the local FS implementation of the
  version-CAS foundation; remotes supply their own preconditions on top of the same port.
- **Closing the move dual-file window** (crash → duplicate slug) — a pre-existing
  recoverable-via-lint state, unchanged by OCC; not this task's to fix.

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)
- [2026-06-24-remote-planning-repos-backends-and-sync](../research/2026-06-24-remote-planning-repos-backends-and-sync.md) — §2 sync/OCC context (its
  per-backend token table is superseded by this content-hash decision).
- Epic [19-web-companion-apps-over-a-shared-core](../epics/19-web-companion-apps-over-a-shared-core.md) — the HTTP surface over this token.

### Prior-art anchors (research 2026-07-04)
RFC 9110 §13.1 (If-Match/If-None-Match/412, strong vs weak); etcd Txn mod_revision;
CouchDB `_rev`/409 + PouchDB delta-vs-full retry; AWS S3 conditional writes (If-Match /
If-None-Match, 2024) + redrive; Git `update-ref` old-oid CAS / `--force-with-lease`;
Firestore `runTransaction` internal bounded retry → `ABORTED`; AWS exponential-backoff-
and-jitter (full jitter); ABA-immunity of whole-content hashing.
