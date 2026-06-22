---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: Guard the Move write-then-remove window (SIGINT/SIGTERM) or ship a lint dedup pass for one-slug-in-two-dirs; sweep atomic .tmp orphans.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [store, robustness]
created: "2026-06-22"
updated_at: "2026-06-22"
completed_at: "2026-06-22"
---
# Crash-recovery signal guard and duplicate-slug repair

## Objective

A kill in Move's write-then-remove window leaves a permanent duplicate slug
(ErrAmbiguous forever) and there is no recovery tooling. Install a SIGINT/SIGTERM
guard around the two-step relocation so the remove always completes, OR ship the
dedup repair the create/set/move comments assume (a lint pass that resolves one slug
in two status dirs). Also sweep/test the atomic-write .tmp cleanup branches.

## Audit reference

planning/audits/open/2026-06-22-code-quality-architecture.md — **M11** (signal guard / dedup repair) and relatedly **L18** (atomic-write
.tmp orphans + untested cleanup branches).

## Acceptance criteria

- [x] Ctrl-C during a move cannot leave a permanent duplicate (guarded or repairable via lint).
- [x] Atomic-write failure/cleanup branches are tested; no stray .tmp accumulation.
- [x] just test + just lint green.

## Outcome (2026-06-22)

Shipped **detection + reporting**, not auto-delete or a signal guard — by design.
The Move-crash duplicate has BOTH copies matching their folders (Move rewrites the
new file's `status:` to its bucket; the old file keeps its own), so it's the
*ambiguous* case the plan says to "report, don't guess" (a destructive auto-delete on
a tie violates the never-lose-data stance). So `core.Lint` now flags any slug in >1
status dir (plain `lint` exits 11 naming the dirs) — making the otherwise-silent
permanent `ErrAmbiguous` discoverable and hand-repairable, satisfying "repairable via
lint" / "reported when ambiguous." A SIGINT/SIGTERM guard was skipped (fragile —
SIGKILL/power-loss defeats it — and the store stays free of process-signal concerns).
L18 done in full: `createFileAtomic` Close-path cleanup, negative atomic-write tests,
and a conservative age+prefix-guarded `.tmp` sweep on `lint --fix`.

## Implementation plan

**Approach.** Ship the **dedup repair pass** rather than (or in addition to) a signal
guard. The window in `Move` is two adjacent syscalls (`writeFileAtomic(newPath)` then
`os.Remove(path)`, fsstore.go:168–173) and the comment already calls the result a
"recoverable duplicate" — but no recovery exists, so `lint`/`lint --fix` should learn
to detect and repair one slug in two status dirs. A SIGINT/SIGTERM guard around the
two-step relocation is a reasonable belt-and-suspenders but is fragile (SIGKILL/power
loss defeats it) and the codebase deliberately keeps the store free of process-signal
concerns; the durable fix is repair tooling. Pair this with the L18 atomic-write
cleanup tests + the `.tmp` orphan sweep. Implement in two parts: (A) duplicate-slug
detect+repair, (B) atomic-write hardening/tests.

**Steps — (A) duplicate-slug repair (M11).**
1. **Detect.** The duplicate already surfaces today: `resolveID` (store/resolve.go)
   returns `ErrAmbiguous` listing both locations when the same slug exists in two
   status dirs. Add a *lint-level* detector that doesn't depend on resolving: in the
   store, a helper `DuplicateSlugs() []domain.DuplicateSlug` (or fold into the existing
   candidate scan) groups `taskCandidates()` by `id` and reports any id with >1
   candidate, with its paths. Surface it through a new lint issue so plain
   `tskflwctl lint` flags it (add to `core.Lint`/`domain` lint output — it's a
   *file-level* problem, so a `domain.FileProblem`-style entry or a dedicated
   `LintResult` issue keyed by slug fits the existing `--json` `unreadable`/issues
   shape).
2. **Repair.** Extend `lint --fix` (`store.FixFrontmatter` / a new `FS` method the
   `Fixer` exposes) with a dedup pass: for a slug in two dirs, the **authoritative
   copy is the one whose frontmatter `status:` matches its folder** (status==directory
   — the un-misfiled one); remove the other. If both agree or both disagree, do NOT
   guess — report it as a manual-fix lint issue (a destructive auto-delete on
   ambiguity violates the tool's "never lose data" stance). Use the atomic helpers'
   sibling (`os.Remove` of the losing copy is fine; the winner is untouched). Gate the
   delete behind non-`--dry-run` and report it in the `FixResult`.
3. **(Optional) signal guard.** If also doing the guard: install a `signal.Notify`
   for SIGINT/SIGTERM in `cmd/tskflwctl/main.go` that, during a move, defers the
   process exit until the in-flight `os.Remove` completes — but scope it tightly and
   document it as best-effort. Treat as a stretch; the repair pass is the acceptance
   bar.

**Steps — (B) atomic-write hardening (L18).**
4. Align `createFileAtomic`'s Close-error path (atomic.go:94–96) with its siblings:
   it returns without `os.Remove(path)` unlike the Write/Sync paths — add the cleanup
   so a failed close doesn't leave a partial new file.
5. Add a stale-`.tmp` sweep: a helper that removes `.tskflwctl-*.tmp` orphans older
   than some threshold, called on `lint`/`lint --fix` (or startup). The `.md` filter
   in `markdownDoc` already keeps listings clean, so this is housekeeping, not
   correctness; keep it conservative (only the tool's own temp prefix).

**Tests.**
- **M11:** `internal/store` — write the same slug into two status dirs, assert the
  detector reports it; run the dedup repair, assert the misfiled copy is removed and
  the folder-matching one survives and now resolves cleanly (no `ErrAmbiguous`); assert
  the both-agree/both-disagree case is reported, not auto-deleted. A CLI-level test
  that `lint` exits non-zero and `lint --fix` repairs it.
- **L18:** negative tests per atomic helper forcing a write/rename/close failure
  (read-only dir via `chmod 0555`, `t.Skip` when running as root since root bypasses
  perms), asserting no orphan `.tmp` remains and no partial target file; a test for the
  fixed `createFileAtomic` Close path; a sweep test seeding a stale `.tmp` and
  asserting it's removed. Mirror the existing harden_test.go temp-repo + `writeTask`
  helpers and the testHook pattern.

**Risks / gotchas.** (a) The repair must respect status==directory and NEVER
auto-delete on a tie — losing the wrong copy is worse than leaving the duplicate;
report-don't-guess on ambiguity. (b) chmod-based failure injection no-ops as root —
skip there (CI may run as root in a container). (c) The dedup detector runs on the
hot read path if folded into resolve — keep it lint-only so normal reads aren't slowed.
(d) If the Fixer/Layout split (sibling task M16/L12) lands first, add the dedup pass
behind the new `Fixer` interface, not the `Store` port. (e) The `.tmp` sweep must only
match the tool's own `.tskflwctl-*.tmp` prefix — never a user file.

**Done when.** A slug duplicated across two status dirs is flagged by `lint` and
repaired by `lint --fix` (or reported when ambiguous), `createFileAtomic`'s Close path
cleans up like its siblings, the atomic-write failure/cleanup branches have negative
tests, stray `.tmp` files are swept, and `go build ./...`, `go test ./...`,
`golangci-lint run ./...` are green.
