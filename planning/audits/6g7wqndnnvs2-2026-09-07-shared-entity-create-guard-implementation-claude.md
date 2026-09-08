---
schema: 1
id: 6g7wqndnnvs2
bucket: closed
area: shared-entity-create-guard-implementation-claude
date: "2026-09-07"
updated_at: "2026-09-07"
---
# Audit: Shared entity create guard implementation — claude — 2026-09-07

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Adversarially review the implementation of the shared ordinary entity-file creation transaction and
the research stable-ID race fix. Treat green tests as claims to challenge, not proof. First assess the
concrete research task, then ask whether the new abstraction genuinely prevents the same check-then-
write defect for current and future entity kinds without weakening graph-aware creation.

## Review target

Review the complete working-tree diff from the current branch base, including:

- `internal/store/create.go`
- `internal/store/create_test.go`
- `internal/store/researchmutate_test.go`
- `docs/ARCHITECTURE.md`
- `planning/tasks/6g7s6hr3qnfq-serialize-research-id-collision-checks-with-creation.md`

Build a repository-wide consumer inventory for ordinary task, epic, audit, and research creation,
plus every compound caller of `writeNewFileUnlocked`: Thread creation, Thread apply, and guarded
task create-and-start. Inspect the repository-lock implementation and all candidate scanners those
paths rely on. Cite exact paths and lines for every claimed consumer and invariant.

## Intended contract to challenge

- A real ordinary single-file create acquires the canonical repository writer guard before invoking
  its preparation callback and retains the guard through the atomic no-clobber write.
- Identity scans and sequence allocation that authorize a create happen in that callback. Two
  cooperating callers cannot both authorize from the same stale repository state.
- Research and audit reject a same-kind stable ID even when the requested slug differs. Ordinary
  tasks do the same and also preserve the planning-space-wide task/Thread ID constraint. Filename-
  recoverable identity remains owned even when its document is unreadable.
- Concurrent research creates using one ID and different slugs produce exactly one file: one succeeds
  and one returns `ErrConflict`. Existing research ID regeneration behavior is unchanged.
- Concurrent epic creation allocates distinct sequence numbers. Duplicate slugs with distinct stable
  IDs remain legal where the existing contract permits them.
- Dry-run performs the same semantic preparation and exact-path check but writes nothing, creates no
  planning root, and does not claim a durable reservation against a later writer.
- The shared helper is filesystem-adapter policy, not portable domain policy. Thread creation, Thread
  bulk apply, and create-and-start keep their authoritative graph/Thread snapshots, pure core plans,
  validation, CAS, receipt, and partial-durability semantics; they reuse only the already-locked
  no-clobber write primitive.
- Exact-path races with raw/non-cooperating writers remain protected by atomic exclusive creation.
  The change does not claim database-style transactions or planning-space-wide identity across kinds
  other than the already-decided task/Thread namespace.

## Mandatory evidence floor

1. Inventory every ordinary and compound creation entry point and trace where its repository lock,
   semantic identity check/allocation, and final write happen. Flag any scan that still occurs before
   the guard or any helper caller that can write without either owning or receiving the guard.
2. Run the focused tests repeatedly under the race detector, including at least:

   ```sh
   go test -race ./internal/store -run 'TestCreateEntityFile|TestCreate(Task|Audit|Epic)|TestFS_CreateResearch|TestTaskAndThreadCreation' -count=100
   ```

3. Perform the exact mutation the new foundation test claims to kill: temporarily invoke ordinary
   preparation before acquiring the repository lock. Demonstrate that
   `TestCreateEntityFileSerializesPreparationWithWrite` fails, restore the source, and rerun it.
4. Remove or bypass the task same-ID candidate check and demonstrate that the readable and unreadable
   owner regressions fail. Restore it and rerun them.
5. Probe two independent `FS` values rooted through equivalent absolute/symlink spellings and, on a
   supported Unix platform, two processes. Verify the shared canonical-root/platform guard protects
   preparation as claimed rather than only serializing one Go pointer.
6. Exercise non-default cases directly: a malformed id-led owner, same ID with another slug, same
   slug with another ID, an exact target path, a missing planning root, dry-run success and conflict,
   preparation failure, and concurrent different-slug epic creation. Inspect files after every failed
   create rather than trusting only returned errors.
7. Trace the core research service's conflict retry/regeneration path and prove the store's error
   classification and changed diagnostics do not disable it. Distinguish an asserted requirement in
   planning from implemented evidence.
8. Run `go test -race ./...`, `just lint`, `just tidy-check`, `just docs-check`, `tskflwctl lint`,
   `tskflwctl audit lint`, and `git diff --check` after restoring all probes.

There is no new wire/schema branch in this change. Verify that claim rather than manufacturing a
wire test; if a public envelope did change, exercise every new optional/non-default value and report
the undocumented contract expansion.

## Required hostile angles

- **Abstraction honesty:** Can a future ordinary entity call the helper while still doing its
  authoritative scan before the callback? Does the API shape make the correct ordering obvious, and
  are current callers actually migrated? Is the helper too filesystem-specific for `core`, or has
  filesystem locking leaked into portable contracts?
- **Lock completeness:** Look for root creation, callback panic/error, directory creation, platform
  lock failure, and unlock behavior. Check same-process, multi-`FS`, symlink-alias, and cross-process
  callers. Separate cooperative serialization from raw-editor guarantees.
- **Identity semantics:** Challenge same-ID/different-slug tasks as well as research/audits, malformed
  filename owners, task/Thread cross-kind collision, and the distinction between an epic's full ID
  and numeric ordering prefix. Ensure the abstraction does not silently invent global uniqueness for
  audits/research.
- **TOCTOU and dry-run:** Identify every projection-to-action window. Ensure real writes do not release
  the guard after preparation and before `O_EXCL`; ensure dry-run is explicitly non-reserving and does
  not mutate directories or metadata.
- **Compound-path regression:** Prove Thread creation/apply and task create-and-start still have their
  stronger source validation, CAS, and receipts. A generic helper must not become a shortcut around
  those plans or accidentally take a nested lock.
- **Test strength:** Look for concurrency tests that pass under the old broken ordering because the
  scheduler happened to serialize them, time-based false confidence, global hook races, cache-masked
  failures, and helpers that let several entity regressions fail for one unrelated reason.
- **Diagnostics and compatibility:** Confirm conflicts retain `ErrConflict`, name the existing owner
  when known, preserve dry-run behavior, and do not break research's bounded minted-ID retry.
- **Second systemic pass:** Search outside the named files for any other create, import, migration, or
  repair path that performs an identity/number scan followed by a separately locked write. Report a
  demonstrated sibling gap or explicitly record the negative evidence and commands used.

Reject speculative architecture churn. Every finding must name a reproducible failure, violated
contract, missing enforceable boundary, or materially weak test. If a concern belongs to another
task, still record it with evidence and a proposed bounded follow-up rather than silently widening
this implementation.

## Validation and restoration

Perform every inspection, mutation, build, test, and generated-file check only inside the mandatory
independent sandbox. Before transfer, restore all probes to the sandbox baseline so only your assigned
audit differs. Do not commit, stage, push, switch branches, edit the source checkout, or transfer any
other file. Use the isolated-workspace helper's `verify` and `transfer` commands exactly as instructed
above; preserve the sandbox if either refuses.

## Deliverable

Replace the reviewer-report placeholder below with:

- a verdict (`ready`, `ready after fixes`, or `not ready`);
- a concise contract/consumer inventory with exact evidence;
- findings using the required `H1`/`M1`/`L1` grammar, left `open` for owner triage;
- commands and mutation results, including what failed before restoration and passed afterward;
- residual risks and explicitly rejected/speculative concerns; and
- the complete isolation/verification/transfer attestation required by the sandbox protocol.

If no findings survive, say what hostile hypotheses were falsified and show the evidence. Do not use
“all tests pass” as the sole basis for a settled verdict.

## Reviewer report

**Verdict: ready.**

The shared ordinary-create transaction does what it claims. Every contract clause in the brief was
attacked with an executed probe rather than read, including two mutation probes that had to fail and
then pass, a cross-process reproduction of the original corruption, and a symlink/alias probe proving
the guard is keyed on the canonical root rather than one Go pointer. No finding survived verification
against the change itself.

The required second systemic pass did find two **pre-existing** sibling defects of exactly the class
this task closes — an identity/number scan that authorizes a later, separately locked, *clobbering*
write. Both are outside the five reviewed files and outside this task's scope, both are demonstrated
with a reproducing test, and both are recorded below as `open` for owner triage with a bounded
follow-up proposal. They do not block this change; `H1` is deterministic and does not need
concurrency, so it should be triaged on its own merits rather than as a footnote to this one.

### Isolation attestation

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.4jCsFm
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.4jCsFm/.git
baseline_commit=37e0d16d38e0825f225edb354b02ff30671f56ab
source_blob=00883cd6909e0eab3053a227f32c9ae1efd4caa2
source_fingerprint=1ba6dbace412054a3c885db3d58634d184586b4f
deliverable=planning/audits/6g7wqndnnvs2-2026-09-07-shared-entity-create-guard-implementation-claude.md
deliverable_changed=true
transfer=succeeded
```

The workspace is an independent `--no-hardlinks` clone; its `.git` is a real directory inside the
sandbox, not a link, file, or pointer back to `/Users/andyeschbacher/git/andy-esch/taskflow`. Source
branch tip at capture was `de2a246`; the sandbox baseline commit `37e0d16` overlays the handoff's
staged, unstaged, and untracked state and is the only commit this reviewer created. Every
inspection, build, test, generator run, scratch fixture, mutation, and edit below happened inside
`$SANDBOX`. The shared checkout was read exactly twice: to read this brief and to perform the helper's
initial copy. All probe files and source mutations were removed before verification —
`git status --porcelain` reported a clean tree with only this audit modified. The workspace is
retained at the path above pending owner confirmation.

### Contract and consumer inventory

Ordinary single-file creates — all four migrated, all scans and allocations inside the callback:

| Entry point | Guard | Identity work inside the callback | Final write |
|---|---|---|---|
| `CreateTask` (`internal/store/create.go:168`) | `createEntityFile` (`:200`) | `ensureCandidateIDUnique("task", …)` + `ensureTaskIDNotThread` (`:201`, `:204`) | `writeNewFileUnlocked` (`:96`) |
| `CreateAudit` (`:260`) | `createEntityFile` (`:283`) | `ensureAuditIDUnique` → `ensureCandidateIDUnique` (`:296`) | same |
| `CreateResearch` (`:316`) | `createEntityFile` (`:344`) | `ensureCandidateIDUnique("research", …)` (`:345`) | same |
| `CreateEpic` (`:412`) | `createEntityFile` (`:419`) | `nextEpicNumber` sequence allocation (`:373`, called `:420`) | same |

`createEntityFile` (`internal/store/create.go:61`) acquires the guard at `:86` and calls `prepare`
at `:92`, with `defer unlock()` at `:90` — the guard is never released between preparation and
`writeNewFileUnlocked` at `:96`, whose `createFileAtomic` (`:113`) supplies the `O_EXCL` backstop
against non-cooperating writers. Dry-run (`:68`) runs the same `prepare` and an exact-path `os.Stat`
without `MkdirAll` and without a lock, which is the documented non-reserving contract.

Compound callers of `writeNewFileUnlocked` — all three keep their own outer guard, authoritative
snapshot, pure-core plan, and CAS; none was routed through the generic callback:

- Thread creation — `internal/store/threadcreation.go:90`; guard at `:31`, `core.ValidateThreadCreationPlan`
  cross-kind and same-kind ID checks at `internal/core/thread_creation.go:120-127`, snapshot re-verify at `:87`.
- Thread bulk apply — `internal/store/threadapply.go:179`; guard at `:26`, per-task CAS at `:131`.
- Guarded create-and-start — `internal/store/lifecyclemutation.go:118`; guard at `:32`,
  `validateCreateAndStartPlan` ID check at `internal/core/task_lifecycle.go:268`,
  `ensureTaskIDNotThread` at `:78` and again at `:115`, whole-snapshot re-verify at `:110`.

Repository guard: `checkedWriteLock` (`internal/store/lock.go:115`) = process-keyed mutex keyed on
`repositoryLockKey` (`:33`, `EvalSymlinks`→`Abs`→`Clean`, lowercased on Windows) plus
`flock(LOCK_EX)` on the root (`internal/store/lock_unix.go:21`). Candidate scanners are all
`flatCandidates` (`internal/store/resolve.go:144`) — a `ReadDir` plus `splitFlatName`
(`internal/store/flatname.go:28`), i.e. filename-only, which is what lets an unreadable record keep
ownership of its recoverable id.

**No scan authorizes an ordinary create before the guard, and no `writeNewFileUnlocked` caller writes
without either owning or receiving it.**

### Mutation and probe results

1. **Pre-lock preparation (required probe 3).** Moved `prepare()` above `s.writeLock()` in
   `createEntityFile`. `TestCreateEntityFileSerializesPreparationWithWrite` failed **200/200**
   iterations — not a scheduler-luck kill. `TestFS_CreateResearch_SerializesDuplicateIDCheckWithCreate`
   and `TestCreateEpicSerializesNumberAllocation` also failed under the same mutation (`successes=2
   conflicts=0`; `ids = "01-alpha", "01-beta"`). Restored via `git checkout --`; all three pass at
   `-count=200 -race`.
2. **Bypassed task identity check (required probe 4).** Deleted the
   `ensureCandidateIDUnique("task", …)` call. `TestCreateTaskRefusesDuplicateIDAcrossDifferentSlugs`
   and `TestCreateTaskTreatsUnreadableFilenameIdentityAsOwned` both failed with `err = <nil>` — the
   duplicate-id file was actually written, so `O_EXCL` and `ensureTaskIDNotThread` do not
   accidentally preserve the invariant. A full `go test ./internal/store` under the mutation failed
   **only** those two, confirming they are the sole guardians and that they are precise. Restored;
   both pass at `-count=50 -race`.
3. **Multi-`FS` and canonical-root aliasing (required probe 5).** A scratch probe ran
   `createEntityFile` on two independently constructed `FS` values reaching one root through a
   **symlink alias**, a `path/.` segment, and a trailing-separator spelling. In all three, the second
   preparation did not begin within 100ms while the first held the guard; 20 iterations under `-race`.
   The keyed mutex therefore protects preparation across `FS` values, not one pointer.
4. **Cross-process (required probe 5).** Two OS processes re-executing the test binary raced
   `CreateResearch` with one stable id and different slugs against a shared root, coordinated by
   ready/start files. Shipped code: exactly one file, 15 attempts × 5 runs. **Under the pre-lock
   mutation the same probe produced two files for one stable id** (`6g7s6hr3qnfq-alpha.md` and
   `-beta.md`) within two attempts. This reproduces the original corruption across process boundaries
   and proves the fix, not just cooperative in-process serialization.
5. **Non-default cases (required probe 6).** Scratch probes, each inspecting the directory after
   every refusal rather than trusting the returned error:
   - same slug / distinct ids still legal for task, audit, and research — **passes**;
   - same id / different slug refused for audit and research, naming the existing owner, nothing
     written — **passes**;
   - unreadable filename owner (`<id>-broken-owner.md`, no frontmatter) owns the identity for audit
     and research as well as task — **passes**;
   - dry-run creates no planning root, no entity directory, and no file, for task, epic, and research
     against a non-existent root — **passes**;
   - dry-run still refuses both an exact-path clash and a same-id clash, and the tree is byte-identical
     before and after — **passes**;
   - preparation returning an error propagates it, writes nothing, and leaves the guard reusable; a
     `nil` prepare is `ErrValidation` — **passes**;
   - a **panicking** preparation callback does not strand the guard (a subsequent `CreateTask` on
     another goroutine completes; 5s deadline) — **passes**;
   - a non-id-led stray (`notes.md`) and an id-only file with no slug segment (`<id>.md`) correctly
     own no identity and do not block a legitimate create — **passes**;
   - concurrent different-slug epic creation allocates `01`/`02` — **passes** (shipped test).
6. **Nested lock / planner reentry.** An ordinary `CreateTask` invoked from inside a
   `MutateThreadCreation` planner callback returns an attributable `ErrConflict` and writes nothing —
   it does not deadlock on the non-reentrant mutex. A concurrent `CreateResearch` issued while a
   planner callback is parked is refused rather than blocked. Both at `-count=5 -race`.
7. **Research retry path (required probe 7).** `internal/core/service_research.go:87-99` keys its
   bounded regeneration loop purely on `errors.Is(err, domain.ErrConflict)`. The refactor preserves
   both wrappers: `ensureCandidateIDUnique` emits `research id %q already used by %q` and
   `entityAlreadyExistsError` (`internal/store/create.go:102`) emits `research doc %q already exists`,
   both wrapping `domain.ErrConflict` — byte-identical wording to the pre-change messages. The loop is
   unaffected. Verified as implemented code, not as a planning assertion.
8. **Wire/schema claim.** Verified rather than assumed: `git diff --name-only de2a246 37e0d16`
   outside `planning/` is exactly `docs/ARCHITECTURE.md`, `internal/store/create.go`,
   `internal/store/create_test.go`, `internal/store/researchmutate_test.go`. No envelope, schema, or
   `--json` surface changed, so there is no new optional wire branch to exercise. No wire test was
   manufactured.
9. **Brief's test names resolve.** The mandated `-run` pattern matches 26 real tests, including
   `TestTaskAndThreadCreationSerializeCrossKindIdentity`
   (`internal/store/threadcreation_test.go:150`) — it is not a vacuous pattern.

### Findings

#### H1. `lint --fix` id repair omits the same-kind identity check and deterministically bricks the graph · **Status:** fixed

`repairInvalidID` (`internal/store/fix.go:180`) authorizes a filename+frontmatter id rewrite with
three checks: already-valid (`:182`), exact target path (`:204-207`), and `crossKindIdentityOwner`
(`:208-214`). That last helper only ever looks at the *other* kind —
`internal/store/fix.go:225-234` maps `tasksDir→threadsDir` and `threadsDir→tasksDir` and returns
`"", nil` for everything else. Nothing checks the **same kind for the canonical id under a different
slug**, which is precisely the case `ensureCandidateIDUnique` was added to cover on the create path.
The write at `internal/store/fix.go:92` is `writeFileAtomic` — a clobbering write, no `O_EXCL` — so
there is no backstop either.

Unlike M1 this needs no concurrency: a single-threaded `tskflwctl lint --fix` takes a healthy
repository to `broken` and reports the corruption as a successful repair.

**Failure scenario (reproduced).** `planning/tasks/` holds `6g7s6hr3qnf0-beta.md` (healthy) and an
unreferenced `6g7s6hr3qnfo-alpha.md` whose id differs only by the Crockford alias `o→0`
(`internal/id/id.go:219`). `FixFrontmatter(false)` renames the second to `6g7s6hr3qnf0-alpha.md` and
returns `Skipped:false` with the change `id: 6g7s6hr3qnfo → 6g7s6hr3qnf0 (canonical Crockford
spelling)`. The directory then holds two files carrying stable id `6g7s6hr3qnf0`;
`core.LoadTaskGraph` reports `health=broken` (`ProblemDuplicateTaskID`,
`internal/core/dependency_graph.go:399-403`), and `ValidateTaskLifecycleSource` then refuses every
guarded mutation on the repository. `repairInvalidID` already declines for the exact-path and
cross-kind cases with an explanatory note, so the omission reads as an oversight rather than a decision.

**Proposed bounded follow-up.** Extend the pre-rename check to reject when a same-kind candidate other
than the file being repaired already owns `good`, reusing `ensureCandidateIDUnique`'s filename-candidate
rule so an unreadable owner also blocks; report it through the existing `note` refusal channel so the
result stays `Skipped:true` rather than becoming a hard error. One test on the alias-collision fixture.

**Resolution:** Added a same-kind filename-identity owner check before canonical
ID repair. lint --fix now returns a skipped result naming the existing owner,
and a regression test proves it creates no duplicate file or source mutation.

#### M1. `RenameTask` authorizes outside the guard, so concurrent renames duplicate a stable id and lose cooperating writes · **Status:** tracked by 6g7wxs43g7nh

`RenameTask` (`internal/store/rename.go:27`) performs its target-path collision check at `:49-55` and
its whole-tree read/cascade computation at `:66-98`, then takes the repository guard only at `:113`
and writes at `:119` with `writeFileAtomic` — clobbering, and with **no** `verifyUnchanged` anywhere
in the function. This is the same projection-then-separately-locked-write shape the reviewed change
removed from ordinary creation, except that here the write primitive provides no `O_EXCL` backstop.
The function's own comment claims it is "a multi-file write serialized by the repo write-lock but NOT
version-CAS-guarded"; the writes are serialized, but the checks that *authorize* them are not, which
is the part the comment does not say.

**Failure scenario (reproduced, first attempt of 40).** Two concurrent
`RenameTask("old", …)` calls on one task with different titles: both stat their target as absent,
both build edits, then both write. Result: `6fjangd7kva1-alpha-title.md` **and**
`6fjangd7kva1-beta-title.md` — two files, one stable id, `graph health=broken`. One call returned
`nil`; the other returned only `remove old …: no such file or directory`, so the corruption is not
attributable to either caller from its own error.

Second reproduction (attempt 1 of 200): a concurrent `RenameTask` and a cooperating guarded
`SetFields("6fjangd7kvb2", {"priority": "low"})` **both returned success**, yet task B on disk
retained `priority: high` with the rename's cascade applied. Rename rewrote B from the copy it read
before locking, silently discarding a guarded field write from a cooperating writer — the lost-update
class `verifyUnchanged` exists to defeat.

**Proposed bounded follow-up.** Move the target stat and the cascade walk inside the guard, or keep
the walk outside and add a `verifyUnchanged`-equivalent recheck per edited file plus a re-stat of
`newPath` after locking. The "git is the undo, single-user CLI" rationale covers a *lost edit*; it does
not cover producing a duplicate stable id, which is the state the tool itself refuses to mutate.

**Resolution:** Reproduction accepted. The follow-up scopes guarded target
authorization, stale cascade-plan detection, per-file CAS, partial durability,
and deterministic concurrent-rename/lost-update tests; it is broader than the
ordinary-create transaction under review.

### Hostile hypotheses falsified

- *"The helper only serializes one `FS` pointer; two stores or a symlinked root slip through."*
  Falsified by probe 3 (symlink alias, `/.` segment, trailing separator) and probe 4 (two OS
  processes). `repositoryLockKey` canonicalizes before keying (`internal/store/lock.go:37-49`).
- *"The concurrency tests pass under the old ordering because the scheduler happened to serialize
  them."* Falsified by 200/200 and 200/200 kill rates under the pre-lock mutation.
- *"Create-and-start is now weaker than `CreateTask` for the unreadable-owner case."* Falsified:
  an unreadable task record becomes `ProblemUnreadable` (`internal/core/dependency_graph.go:355-358`)
  → `GraphBroken` (`:524`) → `MutationReady()` false (`:806`) → `ValidateTaskLifecycleSource` refuses
  the whole operation (`internal/core/task_lifecycle.go:195`). The compound path fails closed more
  strictly, just with a different diagnostic. Not a gap.
- *"The generic helper became a shortcut around the compound plans, or takes a nested lock."*
  Falsified by inventory (all three compound callers still enter through `checkedWriteLock`, load an
  authoritative snapshot, validate in `core`, re-verify, and CAS) and by probe 6.
- *"Dry-run now claims a reservation, or creates directories."* Falsified — no `MkdirAll`, no lock, no
  file, planning root still absent after task/epic/research dry-runs against a non-existent root.
- *"The changed diagnostics break research's bounded minted-id retry."* Falsified — wording and
  sentinel are byte-identical (item 7).
- *"Other write paths share the defect."* Swept every non-test `writeFileAtomic`/`createFileAtomic`
  call site in `internal/`. `fsstore.go:243`, `epicstore.go:124`/`:185`, `auditstore.go:148`,
  `researchstore.go:171`, `body.go:78`, `edit.go:133` all re-verify inside the guard via
  `verifyUnchanged`/`recheck`; `graphmutation.go:100`, `graphrepair.go:142`, `threadapply.go:134`,
  `threadmutation.go:102`, `lifecyclemutation.go:131` lock first and then load, plan, and CAS;
  `fix.go:117` reads and writes wholly inside the guard (`fix.go:34`); `danglers.go` and the
  `internal/config`/`internal/userconfig` writers are read-only or outside the planning space.
  Only `rename.go` and `fix.go`'s id-repair branch broke the pattern — reported as M1 and H1.

### Residual risks and rejected concerns

- **Not enforceable, only documented:** nothing structurally stops a *future* entity from doing its
  authoritative scan before calling `createEntityFile` — the callback shape makes the correct ordering
  natural and the doc comment (`internal/store/create.go:52-60`) states it, but the compiler does not.
  All four current callers are genuinely migrated, and the shipped foundation test pins the boundary,
  so I am recording this as a residual risk rather than a finding; inventing an enforcement mechanism
  here would be the speculative architecture churn the brief rejects.
- **Rejected — redundant planner check.** `createEntityFile:62` calls `rejectRepositoryPlannerCall`,
  which `checkedWriteLock` (`lock.go:116`) already performs. I confirmed against `de2a246` that all
  four `Create*` entry points already had their own check at function entry, so this is duplication,
  not a behavior change, and it correctly extends the check to the dry-run path. Not worth churn.
- **Rejected — epic dry-run does not reserve its number.** Two concurrent dry-runs report the same
  `NN`. This is the explicitly stated contract ("does not claim a durable reservation against a later
  writer"), the real create re-allocates under the guard, and the previous code behaved identically.
- **Rejected — unbounded `repositoryGuards` map.** One entry per canonical root, never evicted. For a
  CLI process this is one entry, and `lock.go:22-27` already documents the ref-counting a long-lived
  multi-space adapter would need. Pre-existing and acknowledged.
- **Unknown (not claimed either way):** Windows behavior. `normalizeRepositoryLockKey`'s
  case-insensitive branch is unit-tested, but `lock_other.go` and cross-process behavior on Windows
  were not executed — this sandbox is darwin/arm64. I am labelling this unknown rather than asserting
  it works.

### Commands

```sh
# baseline
go test -race ./internal/store -run 'TestCreateEntityFile|TestCreate(Task|Audit|Epic)|TestFS_CreateResearch|TestTaskAndThreadCreation' -count=100   # ok 20.1s

# mutation probes (each: fail before restore, pass after)
#  probe 1 — prepare() hoisted above s.writeLock()
go test ./internal/store -run 'TestCreateEntityFileSerializesPreparationWithWrite' -count=200          # 200 failures
go test -race ./internal/store -run 'TestFS_CreateResearch_SerializesDuplicateIDCheckWithCreate' -count=200
go test -race ./internal/store -run 'TestCreateEpicSerializesNumberAllocation' -count=200
#  probe 2 — ensureCandidateIDUnique("task", …) deleted
go test -race ./internal/store -run 'TestCreateTaskRefusesDuplicateIDAcrossDifferentSlugs|TestCreateTaskTreatsUnreadableFilenameIdentityAsOwned' -count=1
go test ./internal/store                                                                              # only those 2 fail
git checkout -- internal/store/create.go                                                              # after each probe

# scratch probes (added, run, then deleted)
go test -race ./internal/store -run 'TestProbeSymlinkAliasSerializesPreparation' -count=20
go test ./internal/store -run 'TestProbeCrossProcessResearchIdentity' -count=5
go test -race ./internal/store -run 'TestProbe(SameSlug|SameIDDifferentSlug|DryRun|Preparation|NonIdLed|UnreadableOwner)' -count=1
go test ./internal/store -run 'TestProbeConcurrentRenameDuplicatesStableID'                           # H1/M1 repro
go test ./internal/store -run 'TestProbeRenameCascadeLosesCooperatingWrite'
go test ./internal/store -run 'TestProbeFixRepairMintsDuplicateSameKindID'
go test -race ./internal/store -run 'TestProbeOrdinaryCreateInsideGuardedPlannerIsAttributable|TestProbeConcurrentCreateDuringPlannerIsRefused' -count=5

# final validation, all probes restored, tree clean
go test -race ./...          # ok, no failures
just lint                    # 0 issues
just tidy-check              # clean
just docs-check              # clean
just build && ./bin/tskflwctl lint        # ✔ all planning entities and dependency links pass lint
./bin/tskflwctl audit lint                # ✔ all audit findings pass lint
git diff --check                          # clean
git status --porcelain                    # only this audit
```

`go test -race ./...`, `just lint`, `just tidy-check`, `just docs-check`, `tskflwctl lint`,
`tskflwctl audit lint`, and `git diff --check` were all re-run after the last probe was removed, from
the restored baseline tree. "All tests pass" is not the basis of the verdict — the mutation kill rates,
the cross-process reproduction, and the falsified hypotheses above are.
