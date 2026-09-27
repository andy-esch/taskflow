---
schema: 1
id: 6ge9gcfjnwte
bucket: closed
area: ordinary-create-guard-release-implementation-codex
date: "2026-09-27"
updated_at: "2026-09-27"
---
# Audit: Ordinary create guard-release recovery — codex — 2026-09-27

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
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

Adversarial implementation review of ordinary entity creation after a repository-guard release failure. Treat the checked task criteria and green tests as claims to falsify. Review independently, then take a second systemic pass. Do not edit implementation or other planning files; write only your assigned audit.

## Review target

Task `6ge7qn9ptaxv-report-post-commit-guard-release-failures-from-ordinary-entity-creation` on branch `fix/ordinary-create-post-commit-release-errors` versus `main`. The source checkout has uncommitted code, generated schema/goldens, architecture notes, and task lifecycle edits; the isolated sandbox snapshot, not `HEAD` alone, is the target. Inventory every ordinary create consumer and every `createEntityFile` call site before judging coverage. Focus on `internal/store/create.go`, lock implementations, core `NewTask`/`NewEpic`/`NewAudit`/`NewResearch`, CLI `new` handlers and `WriteError`, wire `CreatedRecoveryJSON`, tests, schema and goldens. Check the related receipt task and broader entity-integrity design task for scope only.

## Intended contract to challenge

- A dry run writes nothing and claims only a planned path. A pre-commit failure never claims a created file; a release error after the atomic file write returns the exact committed kind-specific receipt plus an error. No success stdout is emitted on that error.
- Error classification and OS detail survive wrapping. Human output says a file committed, identifies its path if local, and tells the caller to inspect rather than blindly retry. `--json` carries stable created identity, committed=true, optional relative path, and workspace. A pathless adapter does not invent a filesystem path.
- Research regeneration applies only to a pre-commit ID collision; even a conflict-classified cleanup failure after commit cannot mint a second document. The existing create-and-start lifecycle failure path remains distinct.
- JSON revision 1.78 is truthfully additive; success envelopes and exit-code meanings do not drift. Source-set composition and command-safety enforcement are not bypassed.

## Mandatory evidence floor

Provide a consumer inventory of task, epic, audit, research, and create-and-start flows, with exact paths/lines. Perform mutation evidence: disable or corrupt one central committed flag/path propagation and show which tests fail; restore it in the sandbox. Inject a release failure with `ErrConflict` and one with an OS error, and inspect classification, receipt, stdout/stderr, and durable files. Check dry run, pre-write failure, and post-write failure separately. A no-findings verdict needs these probes and a hostile example that would have produced an unsafe retry before this change.

## Required hostile angles

1. Can a deferred release overwrite or join the wrong named return values, lose a prepared path, or claim commit when `createFileAtomic` failed? Explore first-create root setup, exact-path collision, and a preparation error combined with release failure.
2. Can any service or CLI handler discard a committed kind-specific receipt, especially Research's conflict retry or task `--start`? Probe all four commands, not just a shared helper.
3. Does `errors.Join` preserve `errors.Is`, OS classification, exit status, and useful JSON without suggesting retry? If an OS unlock failure is not reproducible, say so and use an injected equivalent.
4. Does the new embedded created-item wire shape actually validate with the emitted JSON Schema for non-default data? Check schema revision/golden policy, pathless output, and success-envelope compatibility.
5. Look for systemic risks masked by tests: a fake adapter that violates source-set rules, a lock test hook that does not exercise the real release boundary, or another legacy `writeLock` user mistakenly claimed as fixed. Separate regressions from pre-existing/out-of-scope debt.

For Antigravity: build the call graph independently before reading implementation notes, and try to create one end-to-end counterexample crossing store, core, CLI, and wire. Do not count repeated assertions on the same helper as independent proof.

## Validation and restoration

Run focused tests and `just test`/`just lint` if feasible in the isolated sandbox. Capture commands and outcomes, and restore every mutation/probe before transfer. Do not run generators or write-capable commands in the source checkout.

## Deliverable

Report only reproducible findings with severity, exact evidence, impact, and the smallest viable correction; keep statuses open for implementation-owner triage. Explicitly settle each challenged contract even if no finding. Include the isolation helper attestation and transfer result. Do not make code changes.

## Reviewer report

**Verdict: two medium findings, both open for triage.** I reviewed the isolated baseline (which includes the uncommitted handoff state) in a consumer pass and an adversarial pass. The ordinary committed-error path generally retains its receipts; the two failures below occur where other recovery rules interpret those errors.

### Verified consumer and boundary inventory

| Surface | Verified implementation | Result |
| --- | --- | --- |
| Shared transaction | `internal/store/create.go:54-116,123-137`; `internal/store/lock.go:100-139`; `internal/store/lock_unix.go:21-40` | `createEntityFile` is the sole ordinary file-create transaction. Dry run prepares and checks without taking a lock; a real create makes the root, takes `checkedWriteLock`, prepares, writes with `createFileAtomic`, then joins a release error through the named error return. The Unix OS lock is `flock` on the **planning-root directory handle**, with a same-process root mutex. `writeLock` still discards release errors for other legacy mutations; those callers are outside this task. |
| Store consumers | `internal/store/create.go:185-230,292-325,348-386,444-469` | `CreateTask`, `CreateAudit`, `CreateResearch`, and `CreateEpic` are all four `createEntityFile` callers and map its prepared path and committed flag into kind-specific receipts. Epic numbering and task/audit/research identity scans occur inside the guard. |
| Core consumers | `internal/core/service_task.go:745-835`; `internal/core/service_epic.go:25-64`; `internal/core/service_audit.go:24-62`; `internal/core/service_research.go:34-105`; `internal/core/creation_receipt.go:5-52` | Task, epic, and audit pass through the store receipt. Research regenerates IDs on `ErrConflict`, with a committed-receipt stop guard. Task `--start` uses `runTaskLifecycleMutation` and its distinct `TaskLifecycleMutationFailure` (`service_task.go:532-548,826-835`). Local paths come from operation receipts, never the domain record's `Path`. |
| CLI and wire consumers | `internal/cli/task.go:167-189`; `internal/cli/epic.go:231-249`; `internal/cli/audit.go:70-88`; `internal/cli/research.go:215-233`; `internal/cli/creation_failure.go:19-37`; `internal/cli/exit.go:136-158`; `internal/wire/envelopes.go:695-735,1117-1149` | Each ordinary `new` handler checks `receipt.Committed` before any success render. A committed error carries kind, stable ID, slug, state, optional relative local path, and workspace in `error.created`; human prose names the committed file and says inspect before retry. `--start` selects `error.task_lifecycle` instead. `WriteError` also attaches generic filesystem details when the joined cause contains an OS path error. |
| Composition, schema, scope | `internal/cli/root.go:440-459`; `internal/core/service.go:228-276`; `internal/store/fsstore.go:107-120`; `internal/wire/wire.go:325-328`; `internal/cli/testdata/golden/schema_jsonschema.golden:706-743`; `planning/tasks/6ge7qn9ptaxv-report-post-commit-guard-release-failures-from-ordinary-entity-creation.md:27-49` | Production CLI builds one FS, checks `NewService`, and uses its source-set ID and mutation authorization. `new` handlers retain mutating safety annotations. Wire 1.78 and the baseline JSON Schema golden add `error.created`; success `CreatedEnvelope` is unchanged. The task explicitly leaves legacy edit/fix/body `writeLock` users for later design work. |

#### M1. Research hides a guard-release error when its first ID collides · **Status:** fixed

**Severity:** Medium. **Affected contract:** Release errors and OS detail must survive wrapping; Research should regenerate only for a clean pre-commit ID collision.

**Reproduction:** In a temporary sandbox test, seed `research/6ge7qn9ptaa1-existing.md`, make the ID generator return that ID then `6ge7qn9ptaa2`, and have `testHookRepositoryUnlockError` return `*os.PathError{Op:"flock", Err:syscall.EIO}` on the first release and nil on the second. `go test ./internal/store -run '^TestReviewResearchDoesNotDiscardUnlockFailureOnIDCollision$' -count=1 -v` failed the safety assertion: `err=<nil> minted=2 releases=2`; the returned receipt claims the second document committed and the directory contains the seeded file plus that new document. The temporary test was removed. A direct phase probe showed the first attempt's result is uncommitted and its error retains both `domain.ErrConflict` and the OS failure.

**Cause:** `createEntityFile` joins the preparation collision with the release error at `internal/store/create.go:102-111`. `NewResearch` sees an uncommitted receipt, then retries solely on `errors.Is(err, domain.ErrConflict)` at `internal/core/service_research.go:87-104`; the joined OS error is silently discarded when the next attempt succeeds. This release failure is introduced by this change even though the generic Research collision loop predates it.

**Bounded correction:** Distinguish a pure ID-collision error from one joined with a guard-release failure. Stop and return the original error whenever release failed, even if `ErrConflict` is also present. Add a regression that seeds one collision, injects a first-release OS error, and requires one mint/attempt and preservation of `errors.Is(err, syscall.EIO)` with no new document.

**Resolution:** Research now recognizes a create-finalization failure even when
joined with a pre-commit ID collision; a real-store EIO regression requires one
mint, one release, original error preservation, and no new file.

#### M2. Committed-create JSON can advertise an unsafe retry · **Status:** fixed

**Severity:** Medium. **Affected contract:** A post-commit release error must direct callers to inspect the created document, not suggest rerunning creation.

**Reproduction:** A temporary CLI probe wrapped an injected unlock `*os.PathError{Op:"flock", Err:syscall.EAGAIN}` in `committedCreateFailure` and passed it to `WriteError`. It failed the no-retry assertion because the 1.78 envelope contained both `"created":{"committed":true,...}` and `"filesystem":{"class":"io","operation":"flock","retryable":true}`. An independent `EINTR` injection produced the same contradiction. The error still matched the OS cause and exited 1; the prose said to inspect. The temporary probes were removed. A real-store post-write OS injection also left exactly one durable task file and returned its committed path, so the JSON condition concerns a genuinely committed state.

**Cause:** `internal/cli/exit.go:148-158` adds generic `filesystemDetails(err)` without considering the committed receipt. `internal/cli/fserror.go:41-58` marks `EAGAIN`, `EINTR`, and `EBUSY` retryable; `internal/wire/envelopes.go:1145-1148` describes this field as whether repeating the call unchanged could succeed. In a create error envelope, no field scopes that advice to the low-level unlock call, so an automated caller may retry the whole create and produce another document.

**Bounded correction:** For a committed created-item recovery, force the emitted filesystem `retryable` value to false (or explicitly scope low-level retryability so it cannot be read as command retryability). Pin `EINTR`/`EAGAIN` in a CLI regression that asserts `created.committed=true`, retained OS details, and no retry hint.

**Resolution:** Committed-create JSON now retains OS class and operation but
forces filesystem.retryable=false. EAGAIN and EINTR regressions pin
command-level safety; other mutation envelopes are explicitly queued for the
entity-integrity design pass.

### Negative controls and review effectiveness

- `TestOrdinaryCreateReceiptsSeparatePlannedAndCommittedPaths`, `TestOrdinaryCreateReportsCommittedGuardReleaseFailure`, and `TestCreateEntityFileKeepsDryRunAndPreCommitFailuresDistinct` passed (`internal/store/create_test.go:56-171`). A separate probe with a missing planning root showed dry run created neither root nor file; a post-write `EAGAIN` release error retained `CommittedPath` and a durable file; a later same-ID collision plus release error returned no committed receipt. An exact-path O_EXCL collision left the prior bytes intact and returned `ErrConflict` with `committed=false`. A preparation error joined with release failure is also covered by `create_test.go:147-171`.
- All four ordinary CLI handlers passed the existing conflict injection with empty success stdout (`internal/cli/creation_receipt_test.go:105-146`) and a temporary OS `EIO` injection with exit 1, empty success stdout, kind-specific `error.created`, and filesystem operation/class. The injected fake publishes a nonzero `SourceSetID` (`creation_receipt_test.go:71-103`); it proves CLI projection, while the separate real-FS tests prove durable writes. A temporary `task new --start` command probe returned `error.task_lifecycle` and no `error.created` or success stdout, matching `internal/cli/task.go:169-175` and the existing lifecycle test at `internal/cli/transition_test.go:182-202`. Source-set construction and command-safety checks are not bypassed.
- A temporary schema validator used the emitted `JSONSchema()` and non-default `CreatedRecoveryJSON` values for both a relative local path and an empty path; both validated against `#/$defs/ErrorEnvelope` with `committed=true`. The baseline validator already covers a populated created branch (`internal/wire/envelopes_test.go:521-585`), and the golden/revision checks passed. Pathless projection did not invent a path (`internal/cli/creation_receipt_test.go:148-165`, `internal/cli/root.go:521-530`). The new ordinary recovery prose emits no executable repair command to run; it instructs inspection.
- Targeted mutation probes killed release-error joining, committed phase, committed path, dry-run branch, missing-root setup, all four store receipt gates, Research's committed stop guard, all four CLI committed branches, JSON created mapping, JSON committed flag, and corrupt empty-path projection. Each named test failed on its assertion, then the source bytes were restored. Removing only `app.rel`'s empty-path check was an equivalent mutation on this platform (the fallback still returns empty); corrupting that branch to return `"."` failed the pathless regression. The first release-join mutation caused an unused-local compile failure, so it was rerun in compiling form and the named store test failed. The two findings above were demonstrated by temporary failing safety assertions, not inferred from mutation-only failures.
- Focused store/core/CLI/wire tests passed. `just test` (`go test -race ./...`) passed; `just lint` reported `0 issues`. `git diff --check` was clean after every probe restoration. The emitted 1.78 golden and schema tests passed without regeneration. The separate receipt task `6gdx7mcqq67d` and entity-integrity task `6g7wsr8yvt1a` were consulted for scope; their future legacy-write work is not counted as shipped here.

### Isolation attestation

- Workspace path: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.wrA5Pn`
- Resolved Git directory: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.wrA5Pn/.git` (independent in-tree Git directory; helper verified no worktree link or object alternates)
- Sandbox baseline commit: `6f151a87e662604d3f4ab68ef2ac588b6ebf922d`
- Captured source blob: `0710311aba6169d681c2e96f2318b38357d34a3a`
- Captured source fingerprint: `0203514512282e6eb2d786883504475215162d7a`
- Sole deliverable: `planning/audits/6ge9gcfjnwte-2026-09-27-ordinary-create-guard-release-implementation-codex.md`
- Transfer result: `succeeded` via `scripts/isolated-review-workspace.sh transfer` after fail-closed verification; sandbox retained for owner confirmation.
