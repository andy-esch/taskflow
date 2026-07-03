---
schema: 1
status: in-progress
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Move status/bucket into frontmatter as source of truth and flatten tasks/audits to one dir each (id-led filenames); update store, layout, WatchPaths, resolution, completion. Per ADR-0003.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [core, storage]
created: "2026-07-01"
updated_at: "2026-07-03"
id: 6fhnydm03edq
started_at: "2026-07-03"
---
# Flatten the layout — Phase A: frontmatter-authoritative status/bucket (keep the dirs)

## Approach: two phases; this task is Phase A

Split ADR-0003 §2+§4 into a reversible sequence (the "preserve the folders, keep
flattening optional" call):

- **Phase A (this plan).** Make frontmatter `status` (tasks) / `bucket` (audits)
  the *read authority*. Keep `tasks/<status>/` and `audits/<bucket>/` on disk as a
  **lock-step mirror** — the tool still relocates the file into the matching dir on
  every transition, so in the happy path dir == frontmatter *always*. The only
  behavioural change: when they disagree (hand-edit, crashed move), **frontmatter
  wins** and the stale dir is re-synced by *moving the file*, instead of today's
  "folder wins, rewrite the frontmatter." No file stops moving; no dir is removed;
  a one-line revert restores today's behaviour.
- **Phase B (deferred, its own task).** Stop moving files, flatten to
  `<id>-<slug>.md` (one dir per entity), delete the status/bucket subdirs, resolve
  by id-prefix. The irreversible cutover; rides the one-time migration script.

Phase A de-risks the flip: it proves the read path works off frontmatter while the
directory safety-net is still there.

**Cheap for tasks, real work for audits.** Tasks *already dual-store* `status` in
frontmatter today (`parseTask` reads it, then overwrites it with the dir) — so
Phase A for tasks is **pure inversion, no task data migration**. Audits do **not**
store bucket in frontmatter (`Audit.Bucket` is `yaml:"-"`, dir-derived only), so
audits need a new `bucket:` field **and a one-time backfill** — same shape as the
id backfill we just shipped.

## The core inversion

- `store.parseTask` (fsstore.go:354-358) today: `t.Declared = t.Status; t.Status =
  dirStatus` — **dir wins**. Flip: `Status` = frontmatter value; record the *folder's*
  status as the mirror (rename `Declared` → `FolderStatus`; its meaning inverts).
- `store.parseAudit` (auditstore.go:225): `a.Bucket = bucket` (the dir). Flip to read
  `bucket:` from frontmatter; keep the dir value as the mirror.
- `domain.Task.Misfiled()` (task.go:37) inverts — today "frontmatter ≠ folder, folder
  wins"; after, "folder ≠ frontmatter, frontmatter wins — the mirror is stale, move
  the file." Add the audit analog `Audit.Misfiled()` (none exists today).

## TRAPS / footguns / rakes (read before writing code)

**1. `lint --fix` repairs the WRONG direction — data-loss rake.** `store.realignStatus`
(fix.go) rewrites *frontmatter to match the folder*. After the flip that is
**corruption** — it overwrites the authoritative status with the stale mirror's. It
MUST invert to *move the file to match frontmatter* (never touch the status field).
This lives in the same `FixFrontmatter` walk as the **id backfill we just shipped**,
so inverting it makes the fix walk **move files between the very dirs it is
iterating** (mutate-while-iterating). Collect the moves during the walk, apply them
after. Until realign is inverted, **do not run `lint --fix` on a Phase-A tree** — it
un-migrates it.

**2. Misfiled semantics invert across a wide contract surface.** `Task.Misfiled()`
flipping ripples to `wire.TaskJSON.misfiled`/`declared_status` (dto.go), the
`core.Summary.Misfiled` count + the `status` envelope, `cli/render` warnings, and the
TUI marker/legend/help. All currently mean "status ≠ folder; folder wins; run `lint
--fix` to rewrite frontmatter"; post-flip they mean "folder ≠ status; frontmatter
wins; run `lint --fix` to move the file." Keep the field *names* (avoid a breaking
rename) but **redefine + document** them and **bump `schema_version`** with a
changelog line — an agent on the old contract reads the field with inverted meaning.

**3. `SetFields` must refuse `status` / `bucket` — the silent-desync rake.** `SetFields`
writes frontmatter *in place, no move* (fsstore.go:237). If it sets `status`,
frontmatter changes but the mirror dir does not → instant drift, and the CAS
re-resolve (checks only `curPath`) won't catch it. Guard: `status`/`bucket` mutate
**only** via `Move`/`MoveAudit` (which relocate the mirror). Verify the field
registry / `task set` already excludes `status`, and add a store-level reject so it
can't be bypassed.

**4. Audits are asymmetric — real work, sequence them second.** No `bucket:`
frontmatter (`Bucket AuditBucket yaml:"-"`), no `Declared`, no `Misfiled()`,
`CreateAudit` never writes bucket, and `GetAuditByPath` (auditstore.go:79) derives
bucket purely from the parent dir — it feeds the hot findings/lint sweep, the
highest-risk read site. Phase A for audits: add the field, write `bucket: open` on
create, backfill `bucket: <dir>` into existing audits (a `lint --fix` repair
mirroring the id backfill), add `Audit.Misfiled()` + an audit-misfiled lint, and
**re-point the bucket↔findings gate** — `domain.LintFindings` and `MoveAudit`'s
open-findings check (auditstore.go:140) receive the *directory* bucket today; they
must use the *frontmatter* bucket or a closed-but-misfiled audit escapes the "no open
findings in a closed bucket" rule. Upside: `MoveAudit` is an atomic `os.Rename`
(auditstore.go:164) — audits have **no** write-then-remove dual-file window.

**5. Concurrency is unchanged by Phase A — don't regress the nets; OCC is the real
fix.** Because the mirror keeps moving (lock-step), the atomicity profile stays as
today: tasks are write-then-remove (fsstore.go:222-229) with a recoverable dual-file
window, guarded by the CAS re-resolve + the `duplicateSlugIssues` lint. Keep both.
The residual races the surveys flagged (CAS checks path not the status the read saw;
`from` read from the dir can be stale) are **not new to Phase A** and belong to the
**`version-aware-occ`** task (content-hash `ifVersion`). Don't solve OCC here; just
don't remove the dup-slug recovery net.

**6. Watcher + completion keep working *because* we chose lock-step — don't drift
into "stop moving files."** `WatchPaths` (fsstore.go:54) and shell completion glob
the status/bucket dirs. They keep working in Phase A only because files still move
(a real fsnotify event fires; the glob still reflects status). The moment a status
change becomes a pure in-place edit (Phase B), live-reload silently stops and
completion lists from stale dirs. That break is Phase B's to own.

**7. Move no-op trap.** `moveTask`/`MoveAudit` early-return when `from == to`
(fsstore.go:179, auditstore.go:122), with `from` read from the dir. Make the no-op
compare against the *authoritative* (frontmatter) from-status, and let the re-sync
(lint --fix move) — not `Move` — own fixing a stale mirror, or "move to X" on a file
already frontmatter-X-but-in-the-wrong-dir is a silent no-op that never relocates.

**8. `resolve` casts the dir string to a typed value** (`domain.Status(c.dir)`
fsstore.go:319, `domain.AuditBucket(c.dir)` auditstore.go:200) and the ambiguity
message shows the dir. Post-flip the resolved status must come from the file's
frontmatter, and drift/ambiguity messages must show the authoritative value.

## Sequenced implementation (each step builds green + is revertible)

1. **Tasks read from frontmatter.** Flip `parseTask` (frontmatter → Status; dir →
   FolderStatus mirror); invert `Task.Misfiled()`; `resolve` returns the frontmatter
   status. Nothing else changes (moves still relocate).
2. **Invert the fixer.** `realignStatus` *moves the file* to the frontmatter's dir
   (collect-then-apply, outside the walk); make it coexist with the id-backfill walk;
   re-point lint messages.
3. **Lock the mutation invariant.** Reject `status` via `SetFields`; confirm `task
   set` can't set it.
4. **Contract + presentation.** Redefine `misfiled`/`declared_status`, bump
   `schema_version` + changelog, flip render/TUI/summary wording; regen goldens + docs.
5. **Audits: the field.** Add `bucket:` to the struct; `CreateAudit` writes
   `bucket: open`.
6. **Audits: backfill + authority.** `lint --fix` repair backfills `bucket: <dir>`;
   flip `parseAudit`/`GetAuditByPath` to frontmatter; add `Audit.Misfiled()` + audit
   lint; re-point `LintFindings` + `MoveAudit`'s gate to the frontmatter bucket;
   backfill this repo's audits.
7. **Docs.** ARCHITECTURE.md: the directory is now a *mirror*, authority is
   frontmatter; leave dir removal to Phase B.

## Test pins (before it's "done")
- parse: frontmatter status/bucket wins over dir; FolderStatus captured.
- misfiled inversion (tasks + audits): a dir-lagged file is flagged, `lint --fix`
  *moves* it (does not rewrite frontmatter).
- `SetFields` rejects status/bucket.
- Move/MoveAudit still relocate + stamp; no-op compares against frontmatter.
- dup-slug recovery net still fires (write-then-remove crash sim via the test hook).
- audit bucket backfill: no-`bucket` audit gets `bucket: <dir>`; idempotent/deduped.
- bucket↔findings gate uses the frontmatter bucket (closed+misfiled audit with an
  open finding is flagged).
- golden/schema: schema_version bump; misfiled meaning documented.

## Reversibility / rollback
- Dirs are never removed in Phase A; files stay where the mirror puts them.
- Tasks need no data migration (status already dual-stored) — revert = restore
  `parseTask`'s dir-wins line + `Misfiled()`. Audit `bucket:` is additive (harmless
  if reverted to dir-authoritative).
- If Phase A misbehaves, revert the parse flip; the tree is still a valid
  dir-authoritative layout.

## Explicitly Phase B (NOT here)
Flatten to `<id>-<slug>.md`; delete the status/bucket subdirs; id-prefix resolution;
stop moving files; retire the mirror + WatchPaths dirs + completion globs; the
one-time migration script; Scheme-2 body-link rewrite + rename cascade; and *removing*
the Misfiled concept (Phase A only inverts it — there's no dir to disagree once flat).

## Related
- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]
- ADR [[0003-stable-key-id-addressed-storage]] §2, §4 (and §6 migration for Phase B)

## Code review corrections (2026-07-03)

An external review of the step-1 diff corroborated the plan's remaining steps (2–7)
and the `lint --fix` data-loss risk — no new blockers, but three corrections to fold
in when executing:

- **Do NOT change `resolve`/`resolveAudit` to return the frontmatter status** (the
  review's headline "fix" — it's wrong). Their return feeds `parseTask`/`parseAudit`
  as the *folder* argument (recorded as `FolderStatus`); returning frontmatter there
  makes `FolderStatus == Status` always and **silently disables misfiled detection**.
  The move no-op trap (#7) is real, but the fix is local to the movers: derive
  `moveTask`/`MoveAudit`'s `from` from the file's **parsed frontmatter** (both already
  read the content right after resolve), not from `resolve`'s folder value. `resolve`
  stays folder-returning.
- **Step 3 is live, not just defense-in-depth:** `status` is a `KnownTaskField`, so
  `task set status=X` currently writes it **in place** (SetFields' `case "status"`
  only blocks the *unset* path) — a real tool-driven drift vector today. Step 3 must
  reject `status`/`bucket` on the SET path too, and replace the now-stale message
  "status is the directory — use `task <verb>`/`task move`" (service_task.go:148).
- **Step 4 add:** `TaskJSON.Status`'s jsonschema tag still reads "equals the task's
  directory under tasks/" (dto.go) — now inaccurate; fix it alongside the
  `declared_status` redefinition + `schema_version` bump.
