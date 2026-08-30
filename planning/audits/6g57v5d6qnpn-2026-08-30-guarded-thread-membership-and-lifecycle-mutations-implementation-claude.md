---
schema: 1
id: 6g57v5d6qnpn
bucket: closed
area: guarded-thread-membership-and-lifecycle-mutations-implementation-claude
date: "2026-08-30"
updated_at: "2026-08-30"
---
# Audit: Guarded Thread membership and lifecycle mutations — Claude — 2026-08-30

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer should update.

## Review brief

Perform an independent adversarial implementation review of the uncommitted work for task
[`6g4wm2yf6tyj`](../tasks/6g4wm2yf6tyj-ship-guarded-thread-membership-and-lifecycle-mutations.md)
on branch `feat/guarded-thread-mutations`, against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md), especially the Thread lifecycle,
guarded-mutation, task-impact, and 2026-08-30 implementation-consequence sections.

Assume the implementation may be subtly wrong even though its local suite is green. Look for
semantic contradictions, authorization gaps, stale-snapshot claims, receipt lies, compatibility
breaks, and future seams that will make bulk apply or TUI work unsafe. Do not reward complexity or
test volume by itself. Equally, do not manufacture findings: settle a concern when the code and an
adversarial reproduction actually disprove it.

## Review target

The branch is based at `d38f83a`. The implementation is split across staged, unstaged, and untracked
files, so inspect `git status --short`, `git diff HEAD`, and the relevant untracked files. In-scope
untracked production/test files are `internal/core/thread_mutation.go`,
`internal/core/thread_mutation_test.go`, `internal/store/threadmutation.go`,
`internal/store/threadmutation_test.go`, and the six generated `docs/cli/tskflwctl_thread_*.md`
pages for add/remove/start/complete/cancel/reopen. Ignore unrelated concurrent work under
`planning/meta/`, `routines/`, and
`planning/tasks/6g54ay3njm8y-adopt-branch-protection-ruleset-as-code-.github-ruleset.json.md`.
The two review-audit documents themselves are review scaffolding, not implementation evidence.

The intended contract is:

- `thread add/remove <thread> <task>...` is atomic per command, with explicit idempotent outcomes.
- Thread lifecycle is `unstarted -> in-progress -> completed`, cancellation may enter from
  unstarted/in-progress, completed may reopen, and cancelled is terminal.
- Start requires live membership; complete requires every live member to be soundly drained with
  healthy member/external-gate evidence. Deferred members block completion; deprecated members do
  not count.
- Thread writes preserve author-owned body/comments/unknown fields and never mutate task dependency
  edges or task lifecycle.
- Task lifecycle receipts name every Thread whose derived projection changes, including indirect
  downstream-member and external-gate effects, without writing Thread documents.
- Cooperating dependency, task-lifecycle, and Thread writers serialize on one canonical-root guard
  and re-authorize from fresh state. Raw editors remain outside the advisory lock and are detected
  by whole-snapshot and target CAS only where possible.
- A durable write followed by cleanup failure is typed as committed and is never retried, even when
  the cleanup error wraps conflict.

## Required hostile angles

1. Re-derive the full lifecycle transition matrix, including same-state requests, empty/all-
   deprecated Threads, deferred members, completed drift, completed-to-cancel sequencing, and
   timestamp behavior. Identify any state the CLI can create that the projection or lint cannot
   explain.
2. Trace user references from Cobra through Service into the guarded planner. Look for pre-lock
   resolution, duplicate-resolution ambiguity, planner re-entry, nested guards, or a path that can
   mutate membership/status without the authoritative policy.
3. Audit the store transaction in exact order: strict snapshot, pure validation, materialization,
   dry-run/no-op behavior, whole-source CAS, immediate-target CAS, atomic replacement, unlock, and
   Service retry. Prove that comments/body/unknown fields survive and that failed batches cannot
   partially change membership.
4. Stress cooperating-writer claims with independent stores, in both orderings where meaningful:
   membership versus dependency mutation, completion versus task lifecycle, and task-impact
   attribution versus membership mutation. Distinguish these guarantees from raw-editor races and
   challenge whether the tests genuinely prove waiting plus fresh re-authorization.
5. Attack `TaskLifecycleThreadImpacts`: shared membership, direct members, transitive downstream
   members, external gates, unrelated Threads, completed inconsistency entering and leaving,
   same-status metadata changes, deterministic ordering, aliasing, and the cost/behavior of scanning
   all Threads. Look for false positives, omissions, or receipts that differ from a subsequent read.
6. Verify the strict global failure policy. A missing/malformed/legacy Thread must be noisy in normal
   lint and must not allow guarded task/Thread writes to claim authoritative impact. Check that the
   `abandoned` to `cancelled` replacement is complete without creating two canonical synonyms.
7. Inspect committed-outcome and policy failure behavior in human and JSON paths, including wrapped
   conflicts, batch-row behavior for task lifecycle, exit classification, non-nil arrays, stable
   IDs, before/after views, applicable remedies, and schema 1.56 accuracy.
8. Challenge package boundaries and future composition. The existing-Thread materializer must be
   lock-free and surgical without becoming a second policy owner; upcoming bulk apply must be able
   to compose it under one outer guard without nested repository calls or rebuilding Thread files.
9. Assess test quality rather than only coverage. Use targeted mutation probes or temporary local
   reversions where useful, restore them afterward, and report which tests actually fail. Look for
   global-hook race tests that pass accidentally, overfit assertions, missing negative cases, and
   generated contracts that were refreshed without behavioral proof.
10. Compare README, architecture, ADR, task acceptance criteria, generated CLI docs, wire comments,
    and actual behavior. Call out any stronger guarantee in prose than the advisory filesystem
    design can provide. Inspect the dogfood change that started `complete-production-threads` and
    confirm it did not mutate member tasks.

Run proportionate validation, including the full tests, focused race tests, vet/static analysis,
planning lint, schema/golden checks, and `git diff --check`. Record exact commands and outcomes.

## Deliverable

Update this audit in place after the review. Preserve this brief, then add:

- an executive verdict (`ready`, `ready with tracked follow-ups`, or `not ready`);
- the reviewed commit/worktree state and commands run;
- findings grouped by severity, each with a stable code, `**Status:** open`, file/line evidence,
  impact or reproduction, and a concrete recommendation;
- a short acceptance-criteria traceability table; and
- explicitly settled concerns that looked suspicious but were disproved.

Do not modify implementation files, the ADR, task, Thread, generated artifacts, or the other
reviewer's audit. Do not create follow-up tasks or pre-resolve findings; the implementation owner
will triage them after both independent reviews arrive.

## Findings

<!-- Intentionally empty. The assigned reviewer adds evidence-backed findings here. -->

## Verdict

**Ready with tracked follow-ups.**

The transaction shape, lifecycle matrix, idempotence, batch atomicity, impact attribution, and
the `abandoned` → `cancelled` replacement all survived direct attack. I could not find an
authorization gap, a receipt that lies, a partial batch write, or a stale-snapshot claim the code
does not actually make good on. Several things I expected to be wrong turned out to be defended —
those are recorded under *Settled concerns* with the disproving evidence, because a review that
only lists complaints is not useful for deciding what to trust.

One real semantic defect (**H1**) and one real coverage gap (**M1**) landed in the same function,
`validateThreadCompletion`, and they are the same story: its two per-member loops are the only part
of this slice with no test that fails when the guard is deleted, and one of those loops is wrong.
H1 makes a legitimately-finished Thread permanently un-completable, in a state that `lint` calls
clean and `thread show` reports as consistent, with no supported remedy that preserves the record
ADR-0006 says to keep. The fix is a one-line scope change plus the tests M1 asks for. I would land
both in this task rather than track them, because the next slice (bulk apply) composes this
validator and will inherit the behaviour on repositories that contain withdrawn work.

Nothing else I found rises above Low.

## Reviewed state

- Branch `feat/guarded-thread-mutations`, base `d38f83a`, worktree dirty by design.
- 56 tracked files changed (+1073/−91) plus the four untracked production/test files and the six
  generated `docs/cli/tskflwctl_thread_*.md` pages named in the brief.
- New production code under review: `internal/core/thread_mutation.go` (445 lines),
  `internal/store/threadmutation.go` (195). New tests: 243 + 515 lines.
- Out-of-scope concurrent work under `planning/meta/`, `routines/`, and the branch-protection task
  was ignored, as instructed.

## Commands run

| Command | Outcome |
| --- | --- |
| `just build` | ok (`v0.17.1-67-gd38f83a-dirty`) |
| `go test -race ./...` | 23 packages ok, 0 failures |
| `go vet ./...` | clean |
| `golangci-lint run ./...` | 0 issues |
| `gofmt -l cmd internal` | empty |
| `go mod tidy -diff` | clean |
| `git diff --check` | clean |
| `./bin/tskflwctl lint` (this repo) | `✔ all planning entities and dependency links pass lint` |
| `just docs` re-run | idempotent — regenerating produced no change beyond the six new pages already present |

Five throwaway planning repositories were driven through the real binary: an ordering/no-op lab, a
deprecated-member lab, a full lifecycle-matrix lab, an impact-attribution lab, and a
frontmatter-preservation lab. Mutation probes were run against `internal/core/thread_mutation.go`
and `internal/store/threadmutation.go` and both files were restored and verified by md5 against a
pre-probe hash (`core: identical ✓`, `store: identical ✓`).

## Findings

### High

#### H1. `complete` is blocked by a *deprecated* member's broken gate, contradicting the live-member rule in the AC, README, and ADR-0006  · **Status:** fixed

**File:** `internal/core/thread_mutation.go:332-338` | **Component:** Thread completion policy
**Effort:** XS · **Urgency:** acute

`validateThreadCompletion` ends with

```go
for _, member := range view.Members {
    if member.State.Inconsistent || member.State.Gate == GateBroken {
        return threadPolicyError(...)
    }
}
```

`view.Members` includes **deprecated** members. Every other completion rule in the same function
deliberately excludes them — `Rollup.Total`/`Drained` skip deprecated members, and (per finding H1
of audit `6g4zw7ptvx32`) external gates are derived only from non-deprecated members. This loop is
the one place that does not, so withdrawn work still gates the initiative.

`Inconsistent` is unreachable for a deprecated member (`role == RoleWithdrawn` never satisfies the
in-flight/nominally-complete predicate), so the biting half is `Gate == GateBroken` — which a
deprecated task reaches whenever one of its own prerequisites is deprecated.

**Reproduction** (supported commands only, no hand edits):

```
task new live / withdrawn / upstream
task depend add withdrawn --on upstream
task start live && task complete live --force
task deprecate upstream          # -> withdrawn's gate becomes broken
task deprecate withdrawn
thread new "Dep probe" --task live --task withdrawn && thread start dep-probe
```

State at that point — every live member finished, nothing outstanding:

```
rollup {'done': 1, 'total': 1, 'drained': 1, 'deprecated': 1}
members [('live','completed','clear',drained=True), ('withdrawn','deprecated','broken',drained=False)]
external gates []          projection healthy      inconsistent False      problems []
tskflwctl lint -> ✔ all planning entities and dependency links pass lint
```

```
$ tskflwctl thread complete dep-probe
error: Thread ... cannot complete from in-progress: member 6g57waezj715 has inconsistent or
broken graph state; inspect the member blockers and restore sound dependency evidence
```

**Why this matters beyond the message.** The Thread is genuinely done: 1/1 live members drained, no
external gates, projection healthy, `inconsistent: false`, `problems: []`, and the repository lints
clean. Nothing tells the user anything is wrong until `complete` refuses. The remedy text asks them
to "restore sound dependency evidence" for a task they deliberately withdrew, which is not a
coherent instruction. The only supported escape is `thread remove <thread> <withdrawn-member>`
(verified working), which deletes exactly the record ADR-0006 keeps deprecated members for —
withdrawn work is meant to *stay* a member and be excluded from the denominator, not be erased to
unblock closure.

This also contradicts the prose this branch itself adds. `README.md` now says "`complete` requires
every **live** member to be soundly drained", and AC #2 says "every **live** member". The
implementation applies the last check to all members.

**Why existing tests did not catch it:** see M1 — this loop has no test that fails when it is
deleted, so no case exercises a deprecated member with a broken gate.

**Recommendation:** scope the loop to live members, matching the two rules above it — skip a member
whose `State.Role == RoleWithdrawn` (or whose task status is `deprecated`). Add the reproduction
above as a regression test asserting `complete` succeeds. If withdrawn members with broken gates are
considered worth *surfacing*, surface them as a `ThreadProblem` in the projection rather than as a
completion refusal, so `thread show` warns before `complete` blocks.

**Resolution:** Completion now excludes RoleWithdrawn members from the final
member-evidence gate, matching the live-member denominator. A projected
deprecated member with a broken prerequisite is retained in membership and no
longer blocks completion.

### Medium

#### M1. Both per-member loops in `validateThreadCompletion` are unverified — deleting either one fails no test  · **Status:** fixed

**File:** `internal/core/thread_mutation.go:325-338` | **Component:** completion policy tests
**Effort:** S · **Urgency:** soon

Mutation probes across `./internal/core/ ./internal/cli/ ./internal/store/`, each build-checked
before running so a non-compiling mutant is never scored as a pass:

| Mutation | Tests that failed |
| --- | --- |
| `Rollup.Drained != Rollup.Total` → `false` | `TestValidateThreadCompleteRequiresLiveSoundlyDrainedMembers`, `TestThreadCompletionFailureIsStructuredAndExplanatory`, `TestThreadCompleteWaitsForTaskLifecycleAndRefusesFreshUndrainedState`, `TestValidateThreadLifecycleNoOpsRevalidateDestinationPolicy` |
| cancel-from-completed refusal → `false` | `TestValidateThreadLifecycleContractsAndNoOps` |
| immediate-target CAS removed (store) | `TestThreadMutationRejectsRawRepositoryRaces` |
| `Direct:` flag → always `false` | `TestTaskLifecycleJSONNamesChangedThreadProjectionWithoutWritingThread`, `TestTaskLifecycleThreadImpactsIncludeDirectDownstreamAndExternalGateEffects`, `TestTaskLifecycleWaitsForThreadMembershipAndReportsFreshImpact` |
| **`if gate.Outstanding` → `if false && …`** | **none** |
| **`member.State.Inconsistent \|\| member.State.Gate == GateBroken` → `false`** | **none** |

The rest of the slice is genuinely well covered — every other guard I attacked has a test that names
it. These two are the exception, and H1 is the direct consequence: the only wrong branch in the file
is also the only one nothing pins.

**Recommendation:** add two focused cases to `TestValidateThreadCompleteRequiresLiveSoundlyDrainedMembers`
— a member with a broken gate that must block (after H1, a *live* one), and the external-gate case
if L1 concludes it is reachable.

**Resolution:** Regression coverage now exercises the deprecated broken-gate
case and directly pins both defensive per-evidence checks against contradictory
live projections.

### Low

#### L1. The outstanding-external-gate check in `validateThreadCompletion` looks unreachable behind the drained check  · **Status:** wontfix

**File:** `internal/core/thread_mutation.go:325-331` | **Component:** completion policy
**Effort:** XS · **Urgency:** eventually

The loop runs only after `Rollup.Drained == Rollup.Total`. "Drained" is `SoundlyCompleted`, which is
recursively defined as completed with every direct prerequisite soundly completed. External gates
are precisely the prerequisites of non-deprecated members. So once every live member is drained,
every external gate is soundly completed and `gate.Outstanding` is false by construction. I could
not build a repository reaching it, and the M1 probe shows no test does either.

That is consistent with it being harmless defensive code rather than a defect — but it is currently
indistinguishable from a guard that silently never runs.

**Recommendation:** either construct the reachable case and pin it with a test, or keep it and say in
a comment that it is defensive and subsumed by the drained check, so a later reader does not assume
completion is protected by a check that cannot fire.

**Resolution:** The check is intentionally retained as defense against future
projection representations. A code comment records that it is currently implied
by the drained invariant, and focused coverage pins its behavior on
contradictory evidence.

#### L2. ARCHITECTURE.md's new "preserves … comments" claim is very slightly overstated  · **Status:** fixed

**File:** `docs/ARCHITECTURE.md` (materializer paragraph); behaviour in `internal/store` `updateFrontmatter` | **Component:** docs vs behaviour
**Effort:** XS · **Urgency:** eventually

This branch adds: "Its lock-free update materializer is surgical: it changes only
membership/status/timestamps and preserves unknown fields, comments, key order, and body." Unknown
fields, key order, and body all verified preserved. Inline comment *spacing* on an otherwise
untouched line is normalized:

```
- owner: andy   # trailing comment
+ owner: andy # trailing comment
```

**Not introduced by this branch** — the same normalization happens on the task path
(`task set one --priority high` rewrites `tier: 3   # keep my spacing` the same way), so it is
pre-existing shared `updateFrontmatter` behaviour. It is called out only because this branch is
what newly asserts the guarantee in prose, and because an unrelated line changing produces noisier
review diffs than "surgical" implies.

**Recommendation:** soften to "preserves unknown fields, comments, key order, and body" →
"…preserves unknown fields, key order, body, and comment content (inline comment spacing is
normalized)", or fix the normalization in the shared helper as separate work.

**Resolution:** Architecture and task scoping now promise preservation of
comment content while explicitly acknowledging that the shared YAML editor may
normalize inline-comment spacing.

## Acceptance-criteria traceability

| AC | Verdict | Evidence |
| --- | --- | --- |
| 1 — atomic multi-member add/remove, idempotent no-ops, no dependency edits | **holds** | `add t-inprog t-unstarted nonexistent-task` → exit 10, file byte-identical (whole request rejected pre-write). `add`-existing and `remove`-absent both byte-identical. Mixed real+no-op batch wrote only the real change. No task file touched. |
| 2 — start/complete need a live member; complete needs drained + healthy evidence | **partial — H1** | Live-member and drained rules verified. The member-evidence loop over-applies to deprecated members. |
| 3 — cancel terminal, completed needs reopen, cancelled cannot reopen | **holds** | Full matrix below; cancel-from-completed and reopen-from-cancelled both exit 11 with actionable remedies. |
| 4 — stable IDs, before/after, changed/committed, remedy, human + JSON | **holds** | `thread_mutation` envelope carries `operation, thread_id, path, before, after, member_outcomes, changed, committed, dry_run`, `schema_version: 1.56`, non-nil arrays. |
| 5 — committed cleanup failure typed, never auto-retried | **holds (by inspection)** | `ThreadMutationFailure` wraps cause + receipt; `Unwrap` preserves classification. Not independently raced. |
| 6 — task lifecycle names every affected Thread without writing Thread docs | **holds** | See settled concern S4 — direct, shared, external-gate and transitive all attributed correctly, no false positives, receipts match fresh reads. |
| 7 — `abandoned` → `cancelled`, no two synonyms, lint diagnoses legacy | **holds** | See S5. |
| 8 — same-state no-ops validate and stay byte-identical | **holds** | All four self-transitions exit 0 and leave the file byte-identical. |
| 9 — real two-store races | **holds (by inspection + probe)** | Race tests exist for all three orderings; the target-CAS probe proves `TestThreadMutationRejectsRawRepositoryRaces` is load-bearing. |
| 10 — raw-file races caught by both CAS layers, boundary documented accurately | **holds** | Both CAS layers present and ordered correctly; ARCHITECTURE.md states the verify-to-rename limit honestly (see S7). |

## Settled concerns

Recorded because each looked like a defect and was disproved by evidence.

**S1 — Unsorted membership could make a no-op rewrite the file.** `after.Tasks` is sorted while
`thread.Tasks` keeps persisted order, so `slices.Equal` looked like it would report a spurious
`Changed` on a hand-authored unsorted list and reorder the author's file. Disproved:
`ValidateThreadDocument` requires sorted task ids (I had only read the first half of the function),
`lint` flags a hand-edited unsorted list, and the guarded write refuses it outright with
`thread task ids must be sorted`. Unsorted membership is unreachable.

**S2 — Full lifecycle matrix.** Derived by exit code against a live repo in every state; matches the
contract with no state the projection or lint cannot explain.

| state | add | remove | start | complete | cancel | reopen |
| --- | --- | --- | --- | --- | --- | --- |
| unstarted | ok | ok | ok | 11 | ok | 11 |
| in-progress | ok | ok | ok (no-op) | ok | ok | ok (no-op) |
| completed | 11 | 11 | 11 | ok (no-op) | 11 | ok |
| cancelled | 11 | 11 | 11 | 11 | ok (no-op) | 11 |

**S3 — Batch atomicity.** A batch mixing a valid and a nonexistent task exits 10 with the Thread file
byte-identical; nothing partial lands.

**S4 — Impact attribution false positives/omissions.** With five Threads (direct / external-gate /
transitive / unrelated / shared-membership), `task start prereq` named exactly `a-direct` (direct),
`b-gate` (indirect), `e-shared` (direct) and correctly omitted the unrelated and the
two-hops-away Thread whose state genuinely did not move. Completing the middle task then correctly
surfaced `c-transitive` with **both** the external gate and the member in `changed_task_ids`.
Receipt `after` was compared field-by-field against a fresh `thread show` for every impacted Thread:
identical in all cases. No receipt lies, no omissions, deterministic ordering.

**S5 — `abandoned` → `cancelled` completeness.** The only production reference is the deliberate
legacy detector at `internal/domain/thread.go:38`; remaining matches are unrelated prose. A
hand-authored `status: abandoned` is flagged by `lint` and blocks guarded writes *on other Threads*
with `legacy thread status "abandoned" was replaced by "cancelled"; update the Thread frontmatter` —
confirming both the actionable repair and the strict global evidence gate.

**S6 — Frontmatter preservation.** An unknown key (`custom_unknown_key`), key order, an author HTML
comment, and body prose with irregular whitespace all survived `add` + `start` untouched; only
`status`, `tasks`, and the appended timestamps changed. The single exception is the inline-comment
spacing in L2.

**S7 — Immediate-target CAS.** My first probe reported this as untested; that was **my** error — the
mutant did not compile, and the harness scored a build failure as "no test failed". Re-run with a
build check, removing the CAS fails `TestThreadMutationRejectsRawRepositoryRaces`. The CAS is real,
load-bearing, and covers a genuinely narrower window than the whole-snapshot check (which compares
`SourceVersion` content hashes and runs earlier). Recorded so the same false alarm is not
re-litigated.

**S8 — Dogfood change.** `planning/threads/6g503c6pfqeb-complete-production-threads.md` changed only
`status: unstarted → in-progress` plus `updated_at`/`started_at`. Membership untouched, and the only
task file modified on the branch is the implementation task's own document. `thread start` did not
mutate member tasks.

**S9 — Prose vs advisory locking.** ARCHITECTURE.md states the limit accurately: "Raw editors do not
join the advisory lock; the two CAS checks detect their changes where possible but cannot make a
cross-process guarantee across the final verify-to-rename window." I looked for an overclaim here and
did not find one.

## Note on scope

One observation that is explicitly *not* a finding: `TaskLifecycleThreadImpacts` projects every
Thread twice on every task lifecycle write, so cost grows with (Threads × graph size) per
transition. At current scale this is invisible and the correctness benefit is real. It is worth a
measurement before bulk apply multiplies the number of transitions per command, but there is nothing
wrong with it today and I did not want to inflate it into a finding.
