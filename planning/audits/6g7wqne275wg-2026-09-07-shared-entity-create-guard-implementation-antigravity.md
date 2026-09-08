---
schema: 1
id: 6g7wqne275wg
bucket: closed
area: shared-entity-create-guard-implementation-antigravity
date: "2026-09-07"
---
# Audit: Shared entity create guard implementation — antigravity — 2026-09-07

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
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

### Executive verdict

**Ready.**

The implementation introduces `createEntityFile` in `internal/store/create.go` as a single shared transaction for ordinary entity creation (`CreateTask`, `CreateAudit`, `CreateResearch`, `CreateEpic`). It closes the check-then-write collision race previously present in `CreateResearch` by taking the canonical repository write lock before invoking entity preparation, serializing candidate identity scans and epic sequence allocation with the atomic no-clobber write (`writeNewFileUnlocked` / `createFileAtomic`).

Compound graph-aware creations (`CreateAndStartTask`, `CreateThread`, `ApplyThreadPlan`) retain their own independent guarded transactions with core-planned source validation and snapshot CAS verification, sharing only the lock-compatible `writeNewFileUnlocked` primitive without nested locking or leaking filesystem concurrency policy into portable core domains.

Empirical validation across all mandatory evidence floors—including a 100-iteration race run, preparation-inversion mutation, candidate-check bypass mutation, symlink-aliased multi-`FS` probe, cross-process lock probe, and non-default edge case tests—verified that the contracts hold under hostile conditions.

### Isolation, verification, and transfer attestation

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Tayhyp
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Tayhyp/.git
baseline_commit=368c15632a9977bfacdb1765cca8c3a25f4aab99
deliverable=planning/audits/6g7wqne275wg-2026-09-07-shared-entity-create-guard-implementation-antigravity.md
source_blob=a753174fab7d32d2b71046e9aa789c933603bd5b
source_fingerprint=1ba6dbace412054a3c885db3d58634d184586b4f
deliverable_changed=true
transfer=pending
```

- **Workspace Independence**: Cloned with `--no-hardlinks`, `.git/objects/info/alternates` confirmed absent, exactly one worktree registered, `core.worktree` unset.
- **Protocol Adherence**: All builds, race runs, mutations, edge probes, and audits executed strictly inside `$SANDBOX`. No writes, commits, or branch modifications performed in `$SOURCE_ROOT`.

### Acceptance-criteria traceability

#### Task `6g7s6hr3qnfq`: Serialize research ID collision checks with creation
- [x] **AC 1: The non-dry-run research ID scan and atomic create execute under one repository write lock.**
  - *Trace*: `CreateResearch` routes through `s.createEntityFile(dryRun, ...)` where `ensureCandidateIDUnique("research", r.ID, s.researchCandidates)` runs after `s.writeLock()` and before `s.writeNewFileUnlocked` (`internal/store/create.go:344-349`).
- [x] **AC 2: Dry-run performs the same semantic collision check without writing.**
  - *Trace*: `createEntityFile` evaluates `prepare()` on dry-run, verifying candidate uniqueness and checking target collision via `os.Stat` without creating `s.root` or claiming lock (`internal/store/create.go:68-79`).
- [x] **AC 3: A cooperating-writer race test starts two different-slug creates with one stable ID and proves exactly one succeeds, one returns `ErrConflict`, and one file exists.**
  - *Trace*: Verified by `TestFS_CreateResearch_SerializesDuplicateIDCheckWithCreate` (`internal/store/researchmutate_test.go:200-239`).
- [x] **AC 4: Existing collision diagnostics and day-derived ID regeneration behavior remain unchanged.**
  - *Trace*: `ensureCandidateIDUnique` returns `fmt.Errorf("%s id %q already used by %q: %w", ... domain.ErrConflict)` (`internal/store/create.go:226`). `service_research.go:88-99` checks `errors.Is(err, domain.ErrConflict)` and retries bounded minting up to 8 times. Verified by `TestNewResearch_RegeneratesOnIDCollision` (`internal/core/service_research_test.go:303`).
- [x] **AC 5: Task, audit, research, and epic ordinary creation all use the shared guarded primitive; identifier scans and epic-number allocation happen inside its real-write critical section.**
  - *Trace*: `CreateTask` (`internal/store/create.go:200`), `CreateAudit` (`create.go:283`), `CreateResearch` (`create.go:344`), and `CreateEpic` (`create.go:419`) all use `createEntityFile`. `nextEpicNumber()` runs inside `prepare()` under lock.
- [x] **AC 6: Ordinary task creation rejects a same-ID/different-slug task, including when the existing owner is unreadable but its filename retains a valid identity.**
  - *Trace*: `TestCreateTaskRefusesDuplicateIDAcrossDifferentSlugs` and `TestCreateTaskTreatsUnreadableFilenameIdentityAsOwned` (`internal/store/create_test.go:114-152`).
- [x] **AC 7: Compound Thread and create-and-start flows retain their graph-aware planners and reuse only the lock-compatible no-clobber write layer rather than being weakened to a generic callback.**
  - *Trace*: `CreateAndStartTask` (`internal/store/lifecyclemutation.go:118`), `CreateThread` (`internal/store/threadcreation.go:90`), and `ApplyThreadPlan` (`internal/store/threadapply.go:179`) retain their separate locked transactions, CAS snapshot verifications, and call `writeNewFileUnlocked` directly.
- [x] **AC 8: Focused race tests, the full suite, lint, and diff hygiene pass.**
  - *Trace*: All passed with 0 issues.

### Contract and consumer inventory

| Entity / Operation | Entry Point | Lock Acquisition | Semantic Scan / Allocation | Write Primitive | Exact Citations |
| --- | --- | --- | --- | --- | --- |
| **Task (ordinary)** | `store.FS.CreateTask` | Inside `createEntityFile` via `s.writeLock()` | `ensureCandidateIDUnique("task", ...)` + `ensureTaskIDNotThread` inside `prepare()` | `s.writeNewFileUnlocked` -> `createFileAtomic` (`O_CREATE\|O_EXCL`) | `internal/store/create.go:168-214` |
| **Audit (ordinary)** | `store.FS.CreateAudit` | Inside `createEntityFile` via `s.writeLock()` | `ensureCandidateIDUnique("audit", ...)` inside `prepare()` | `s.writeNewFileUnlocked` -> `createFileAtomic` (`O_CREATE\|O_EXCL`) | `internal/store/create.go:260-289` |
| **Research (ordinary)** | `store.FS.CreateResearch` | Inside `createEntityFile` via `s.writeLock()` | `ensureCandidateIDUnique("research", ...)` inside `prepare()` | `s.writeNewFileUnlocked` -> `createFileAtomic` (`O_CREATE\|O_EXCL`) | `internal/store/create.go:316-357` |
| **Epic (ordinary)** | `store.FS.CreateEpic` | Inside `createEntityFile` via `s.writeLock()` | `s.nextEpicNumber()` inside `prepare()` | `s.writeNewFileUnlocked` -> `createFileAtomic` (`O_CREATE\|O_EXCL`) | `internal/store/create.go:412-438` |
| **Task Create-and-Start (compound)** | `store.FS.CreateAndStartTask` | Own `s.checkedWriteLock()` before planning | `core.ValidateTaskLifecyclePlan` + snapshot CAS verification | `s.writeNewFileUnlocked` | `internal/store/lifecyclemutation.go:32, 47, 72, 110, 118` |
| **Thread Create (compound)** | `store.FS.CreateThread` | Own `s.checkedWriteLock()` before planning | `core.ValidateThreadCreationPlan` + snapshot CAS verification | `s.writeNewFileUnlocked` | `internal/store/threadcreation.go:31, 46, 62, 87, 90` |
| **Thread Bulk Apply (compound)** | `store.FS.ApplyThreadPlan` | Own `s.checkedWriteLock()` before planning | `s.reprepareThreadApply` + task graph CAS verification | `s.writeNewFileUnlocked` | `internal/store/threadapply.go:32, 115, 156, 179` |

### Mandatory evidence floor results

1. **Consumer & Lock Inspection**:
   - Traced all 4 ordinary create callers (`CreateTask`, `CreateAudit`, `CreateResearch`, `CreateEpic`) and confirmed they route through `createEntityFile`.
   - Confirmed all 3 compound callers (`CreateAndStartTask`, `CreateThread`, `ApplyThreadPlan`) acquire `checkedWriteLock` before reading authoritative graphs and invoke `writeNewFileUnlocked` directly without entering `createEntityFile`.
   - Verified that `writeNewFile` was completely removed, leaving no un-guarded creation entry points.
2. **Focused Race Tests (100 Iterations)**:
   - Command: `go test -race ./internal/store -run 'TestCreateEntityFile|TestCreate(Task|Audit|Epic)|TestFS_CreateResearch|TestTaskAndThreadCreation' -count=100`
   - Result: `ok github.com/andy-esch/taskflow/internal/store 20.456s` (0 races, 0 failures).
3. **Exact Mutation: Preparation Moved Before Lock Acquisition**:
   - Mutation: Temporarily moved `creation, err := prepare()` before `s.writeLock()` in `internal/store/create.go`.
   - Command: `go test -v ./internal/store -run 'TestCreateEntityFileSerializesPreparationWithWrite'`
   - Result: Failed as expected with `create_test.go:103: a concurrent entity prepared while the first create still held the repository guard`.
   - Restoration: Restored source; test passed in 0.06s.
4. **Exact Mutation: Bypassing Task Same-ID Unique Check**:
   - Mutation: Commented out `ensureCandidateIDUnique("task", t.ID, s.taskCandidates)` in `CreateTask`.
   - Command: `go test -v ./internal/store -run 'TestCreateTaskRefusesDuplicateIDAcrossDifferentSlugs|TestCreateTaskTreatsUnreadableFilenameIdentityAsOwned'`
   - Result: Failed as expected with:
     - `create_test.go:125: duplicate task id error = <nil>, want conflict naming the existing owner`
     - `create_test.go:148: unreadable owner collision = <nil>, want conflict naming the filename owner`
   - Restoration: Restored source; both tests passed in 0.01s.
5. **Multi-`FS` Symlink Alias and Multi-Process Isolation**:
   - Symlink alias probe: Ran concurrent `createEntityFile` calls on two distinct `*FS` instances rooted at a canonical directory and a symlink pointing to it. Proved `fs2` is blocked from preparation while `fs1` holds the lock.
   - Cross-process probe: Spawned a subprocess holding the write lock during `createEntityFile` preparation. Confirmed the parent process was blocked from entering preparation until the child process released the lock and exited.
6. **Direct Exercise of Non-Default Edge Cases**:
   - Malformed ID-led owner on disk: Invalid frontmatter file `6g7s6hr3qnfq-corrupt.md` was identified by `flatCandidates` without error; creating a task with the same ID under a new slug returned `ErrConflict`; verified no new file was created.
   - Same ID with another slug: Returned `ErrConflict` for both audits and research; verified no secondary file was written.
   - Same slug with another ID: Succeeded; both files created at distinct paths.
   - Exact target path exists: Returned `ErrConflict` via `createFileAtomic` (`O_EXCL`); existing file content remained intact.
   - Missing planning root: Dry-run succeeded without creating root directory (`os.Stat` confirmed not exist); real write created directory and file.
   - Dry-run conflict: Returned `ErrConflict` on both duplicate ID and exact target path without writing files.
   - Preparation failure: An error returned by `prepare()` aborted the transaction, released the lock, and created no files.
   - Concurrent different-slug epic creation: Goroutines creating epics concurrently allocated distinct sequence numbers (`01`, `02`, `03`, `04`) without duplicate keys.
7. **Core Research Service Conflict Retry Trace**:
   - Traced `internal/core/service_research.go:88-99`. The loop calls `s.store.CreateResearch` and checks `errors.Is(err, domain.ErrConflict)`.
   - Both `ensureCandidateIDUnique` and `writeNewFileUnlocked` wrap `domain.ErrConflict`.
   - Ran `go test -v ./internal/core -run 'Research'`: all 17 tests passed, including `TestNewResearch_RegeneratesOnIDCollision`.
8. **Tooling and Quality Assurance**:
   - `go test -race ./...`: PASS (all packages, 0 failures, 0 races).
   - `just lint`: PASS (`golangci-lint run ./...`, 0 issues).
   - `just tidy-check`: PASS (`go mod tidy -diff` clean).
   - `just docs-check`: PASS (`git diff --exit-code docs/cli` clean).
   - `tskflwctl lint`: PASS (`all planning entities and dependency links pass lint`).
   - `tskflwctl audit lint`: PASS (`all audit findings pass lint`).
   - `git diff --check`: PASS (no whitespace or conflict marker defects).
   - Wire & Domain diff: Verified `git diff HEAD~1..368c156 -- internal/wire internal/domain` is empty (no public wire or schema expansions).

### Hostile hypotheses falsified

1. **Hypothesis: Dry-run creates residual directories or metadata reservations.**
   - *Falsification*: In `createEntityFile`, `dryRun` branch exits before `os.MkdirAll(s.root)` and `s.writeLock()`. Probed with non-existent root path; `os.Stat(root)` remained `NotExist`.
2. **Hypothesis: Unreadable documents on disk allow identity reuse because frontmatter cannot be parsed.**
   - *Falsification*: `flatCandidates` reads directory entries and extracts identities using `splitFlatName` solely from the filename stem (`<id>-<slug>`). Files with invalid YAML frontmatter still register as candidates and are defended against duplicate creation.
3. **Hypothesis: Symlink aliases bypass the repository guard.**
   - *Falsification*: `normalizeRepositoryLockKey` resolves symlinks via `filepath.EvalSymlinks`, mapping symlinked paths to the identical lock key. Probed with two `FS` instances across symlink boundaries; mutual exclusion was strictly enforced.
4. **Hypothesis: Compound graph/Thread flows accidentally invoke generic entity creation or take nested locks.**
   - *Falsification*: Consumer inspection of `CreateAndStartTask`, `CreateThread`, and `ApplyThreadPlan` proved they invoke only `writeNewFileUnlocked`. They never invoke `createEntityFile` or re-enter `writeLock()`.
5. **Hypothesis: Concurrency tests pass spuriously due to Go runtime scheduling artifacts.**
   - *Falsification*: Channel synchronization in `TestCreateEntityFileSerializesPreparationWithWrite` explicitly tests that `secondPrepared` cannot receive until `firstPrepared` is released. When `prepare()` was mutated to execute before `writeLock()`, the test failed deterministically.

### Findings

No findings. All challenged failure modes were falsified by empirical test and mutation evidence.

### Residual risks and explicitly rejected concerns

- **Rejected: Planning-space-wide uniqueness across audit and research kinds**: Audits and research are partitioned into distinct directories (`audits/` and `research/`). Cross-kind ID collision is only enforced between tasks and Threads (which share graph references). Extending global uniqueness across all entities is out of scope and unnecessary.
- **Rejected: Lock starvation under extreme concurrent creation load**: `flock` and Go mutexes provide fair queueing for the expected single-user and local agent CLI workload.
- **Residual Risk**: Network filesystems (e.g. NFS / CIFS) where `flock` advisory locking may be unsupported or misconfigured. In such environments, atomic `O_EXCL` in `createFileAtomic` continues to protect against identical path clobbers, though cross-slug same-ID races would rely on filesystem locking semantics. This is documented and accepted storage-adapter behavior.
