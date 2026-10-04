---
schema: 1
id: 6ggee2fx4053
bucket: closed
area: workspace-thread-apply-identity-parity-implementation-codex
date: "2026-10-04"
updated_at: "2026-10-04"
---
# Audit: Workspace Thread-apply identity parity implementation — codex — 2026-10-04

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

Adversarially review workspace-opened Thread-apply identity parity, not merely whether the new
helper compiles. Challenge stale identity, wrong pointer anchors, dropped live authorization,
accidental startup reads, and tests that pass through an unrelated earlier refusal.

Codex: emphasize contract/lifetime analysis and recovery evidence across ordinary and workspace
consumers. Antigravity: emphasize hostile executable scenarios and coordinated mutation probes;
a checklist-only inventory is insufficient. Both reviewers must perform the complete evidence floor.

## Review target

- Task: [workspace identity parity](../tasks/6gg7e594gcms-preserve-planning-identity-revalidation-in-workspace-opened-thread-apply.md).
- Thread: [adapter-neutral planning data](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md).
- Base main: `8064dfc902dc51c2aaa49fc8bf2fd0316b6654fa`, plus the final working-tree overlay
  captured by the mandatory sandbox helper. Review the overlay, including untracked Go files.
- Production delta: `internal/workspacestore/planning.go`, `internal/workspacestore/fs.go`,
  `internal/appwiring/wiring.go`. Tests: `internal/workspacestore/planning_test.go`,
  `internal/appwiring/thread_apply_test.go`; startup regressions remain in `wiring_test.go`.
- Trace unchanged guards in `internal/store/threadapply.go`, core Thread apply and workspace
  composition, and real config discovery. Inspect `docs/ARCHITECTURE.md` and `.golangci.yml`.

## Intended contract to challenge

1. Both ordinary and workspace services can compose, dry-run, and commit a valid Thread plan
   against directly discovered and pointer-selected planning trees.
2. One already-observed corpus supplies the store, identity metadata, paths, graph/Thread reads,
   and watcher layout. Shared construction does not rediscover, read home state, or choose cwd.
3. Apply uses fresh identity from the original marker directory, including the pointer checkout.
   Same-ID root changes, replaced/missing IDs, failed re-reads, and malformed markers fail closed.
4. The invocation's live authorization reaches both dry-run and commit. Read-only operations do
   not require mutation authorization. Nil policy semantics are unchanged, not newly secured.
5. Early refusals do not alter documents in either corpus and retain the attempted durable token
   in a typed failure. Existing store preflight/prefix semantics remain; this is not a new atomic
   transaction guarantee against every non-cooperating edit during a multi-file apply.
6. Config discovery remains local secondary/composition wiring. Core/invocation ports, public
   machine fields, source-set validation, and local watcher contracts do not acquire config types.

## Mandatory evidence floor

- Write a verified consumer inventory: where each helper input originates, when it is observed,
  what is captured versus re-read, and where the constructed capabilities go. Include both
  `OpenPlanning` and `OpenWorkspace` paths, not only the exported helper. Distinguish read-only
  overview stores from the complete mutation service rather than assuming every `NewFS` is a bug.
- Run the restored focused tests first. Prove a real direct and pointer workspace can dry-run and
  commit a plan with a dependency write; compare ordinary opening and verify actual document bytes.
- Perform at least three compiler-valid mutations covering fresh identity, live authorization,
  and pointer anchoring. Require the intended named test to fail. Test helpers must not bypass the
  production opener, substitute a permissive store, or fail because tags/manifest metadata are invalid.
- Include a same-ID, matching-task root replacement. A different-ID target can be rejected by
  config pointer validation while a missing physical-root guard remains undetected.
- Report command, selected tests, exact outcome/error class, before/after document evidence, and
  why each probe reaches its intended guard. A happy-path substring or vague "tests pass" is not proof.
- Record uncertainties and unsupported platforms explicitly. Do not certify all mutation families
  or concurrency windows from this bounded Thread-apply slice.

## Required hostile angles

1. Trace a nested start -> pointer marker -> flat planning target. Repoint the pointer to a corpus
   with the same ID and tasks. If the helper anchors at the resolved target, does an initial apply
   still pass and a later repoint become invisible? Try the exact wrong-anchor mutation.
2. Cache the initial ID/root, bypass the physical-root comparison, and bypass the core plan/ID
   comparison separately. Which new rows kill each mutant? Where a sibling check masks a defect,
   use a coordinated probe and explain the masking rather than claiming the mutant was covered.
3. Drop authorization at the shared constructor and, separately, at workspace composition.
   Exercise denied dry-run and commit with otherwise valid fixtures. Check callback counts, live
   policy changes after opening, no document writes, and preservation of the attempted Thread ID.
4. Return nil configuration, an empty root, an error with partial data, or no durable ID from the
   re-reader. Can cached values be adopted, can an error become success, or can the caller panic?
5. Add eager or repeated discovery to shared construction. Use the real-reader startup tests;
   a counter around a fake adapter does not prove the runtime opener remains lazy. Challenge a
   mismatched initial metadata/store corpus and caller mutation of the retained config object.
6. Look at identity changes between compose and apply versus changes during guarded preflight or
   a durable prefix. Separate pre-write refusal from partial recovery; do not demand a new
   transaction model without demonstrating a regression or tracing it to an existing task.
7. Take a second systemic pass: is this the right secondary-adapter seam, does it duplicate policy,
   create a dependency cycle, weaken capability composition, or expose misleading guarantees to a
   future TUI/web caller? Verify the documentation's direct import graph and actual lint rules.

## Validation and restoration

Run inside your independent sandbox, not the shared source:

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/appwiring ./internal/workspacestore -count=1
GOCACHE=/tmp/taskflow-review-go-cache go test -race ./internal/appwiring ./internal/workspacestore -count=1
GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache just lint
GOCACHE=/tmp/taskflow-review-go-cache just docs-check
```

Owner validation includes build, full tests, full uncached race tests, lint, generated-doc checks,
and five restored mutation probes documented in the task. Treat that as a claim to challenge, not
as independent reviewer evidence. Run additional full tests if your probe changes shared semantics.
Restore every mutation to the captured sandbox baseline before repeating focused checks. Generated
outputs, binaries, scratch files, and changes outside your assigned audit must not transfer back.

## Deliverable

Preserve this brief and fill the Reviewer report below. Create findings through the sandbox's
`audit finding new` verbs, leave them open for owner triage, and name exact verified symbols and
lines. Separate demonstrated defects, substantive design risks, test weaknesses, and out-of-scope
followups. No finding quota; if nothing survives the challenge, show the experiments that rejected
your strongest hypotheses. Do not mark this implementation task complete or change source planning.

## Reviewer report

### Scope and outcome

Completed the brief checklist and a separate adversarial systemic pass against the captured working-tree overlay, including the three untracked Go files. Base main is `8064dfc902dc51c2aaa49fc8bf2fd0316b6654fa`; the restoration/review baseline is the helper's sandbox commit, not base main. Reviewed the production opening delta, its new regressions, unchanged store/core apply guards, real configuration discovery, composition, architecture documentation, and import rules.

The restored implementation passed the bounded runtime challenges: ordinary/workspace and direct/pointer compose, dry-run, dependency persistence, Thread creation, fresh identity refusals, live authorization, and preflight/prefix recovery. No new production persistence defect was demonstrated. **M1** is a demonstrated systemic regression-test gap; **L1** is an inaccurate documented import edge. Both were created with the sandbox CLI's `audit finding new` verb and remain **open** for owner triage. No implementation task, Thread, sibling audit, or source code was changed by the deliverable.

Isolation disclosure: before reading the brief's isolation section, initial read-only bootstrap also listed paths, read Git status/log, and read the audit generator and helper references in the shared checkout. Those administrative reads exceeded the brief's literal initial-read restriction. No implementation was inspected and no project tests, builds, mutations, staging, or writes ran there. The helper then created the independent clone before all substantive review work; all subsequent inspection, probes, tests, report editing, and restoration used that sandbox. The only final source write is the guarded audit transfer.

### Verified consumer and lifetime inventory

| Consumer/input | Origin and observation | Captured versus fresh | Published capabilities and evidence |
| --- | --- | --- | --- |
| Invocation policy | `newRootCmd` supplies `app.authorizeMutation` to `Bindings.Compose` (`internal/cli/root.go:274`); `commandSafetyState.authorizeMutation` reads the invocation's currently bound classification (`internal/cli/command_safety.go:52`). Test policies were live closures with counters. | The closure is retained, not its opening-time decision. Opening and compose make zero policy calls. | `compose` gives the same closure to ordinary opening and `workspacestore.New(WithMutationAuthorization(...))` (`internal/appwiring/wiring.go:26`, `:35`, `:37`). Nil retains existing permissive semantics (`internal/store/fsstore.go:86`); this audit does not claim a new public authorization policy. |
| Ordinary opening | `LocalBindings` selects actual `config.Discover` and `userconfig.Load` without invoking them (`internal/appwiring/local_sources.go:24`). `openPlanning` calls the selected discoverer once on the supplied start (`internal/appwiring/wiring.go:43`). | That one returned config supplies the helper, `repositorySettings`, and `store.FS`. The helper copies the marker/root strings; it does not retain mutable config fields for later redirection. | `core.NewService(fs)` and `ports.Planning{Repository, Service, Layout:fs}` (`internal/appwiring/wiring.go:48`, `:52`, `:56`). The actual service, watcher paths, task reads, and identity metadata were checked with real records, not a permissive replacement store. |
| Workspace opening | `WorkspaceService.Open` requires an explicit start and calls the workspace adapter (`internal/core/workspace.go:78`, `:87`). `FS.OpenWorkspace` calls real `config.Discover(start)` once (`internal/workspacestore/fs.go:33`). | The same observed config supplies `Checkout`, `PlanningRoot`, `PlanningID`, and shared construction. A pointer's checkout is `cfg.Dir`; legacy fallback uses `cfg.Root`. | The same `fs` supplies `Store`, `TaskGraphs`, entity paths, `Threads`, `ThreadPaths`, and `Layout` (`internal/workspacestore/fs.go:46`). Core composes those capabilities (`internal/core/workspace.go:107`); `NewService` checks source sets including Thread apply (`internal/core/service.go:387`, `:404`, `:424`). A nonzero witness is object/corpus composition evidence, not proof that arbitrary forged metadata is truthful. |
| Shared constructor | `NewPlanningStore(cfg, discover, authorize)` receives an already-discovered config, the opener's reader, and the invocation policy (`internal/workspacestore/planning.go:16`). | Initial store root is `cfg.Root`. The discovery anchor is copied from `cfg.Dir`, falling back only to `cfg.Root` (`:23`). At apply it invokes the supplied reader afresh (`:28`), discards partial data on error (`:29`), validates nil/empty root (`:32`), and returns fresh root/ID (`:35`). | One `store.NewFS` installs identity and authorization (`:27`, `:36`). Construction contains no discovery, home lookup, or cwd selection in the restored code. Caller mutation of the original config does not redirect the captured anchor or store root. |
| Apply identity and persistence | `MutateThreadApply` authorizes and locks first (`internal/store/threadapply.go:30`; `internal/store/lock.go:115`), re-reads identity before authoritative graph/Thread reads (`threadapply.go:45`), repeats identity and whole-source verification before commit (`:108`), and re-prepares after a durable dependency prefix (`:256`). | `currentPlanningIdentity` compares normalized physical root keys and requires a durable ID (`internal/store/threadapply.go:206`, `:214`, `:217`). `PrepareThreadApply` separately compares plan and current IDs (`internal/core/thread_apply.go:335`). | Core preserves an attempted token even if refusal precedes the store's accepted plan, and returns `*ThreadApplyFailure` (`internal/core/service_thread_apply.go:66`, `:71`). On tested Darwin, the OS advisory guard opens/flocks the planning root directory (`internal/store/lock_unix.go:21`), not an invented lock-file path. |
| Watchers and overview | `FS.WatchPaths` returns the five directories under the initial store root (`internal/store/fsstore.go:132`). `spacestore.OpenPlanningStore` uses bare `store.NewFS(root)` (`internal/spacestore/fs.go:111`). | Watcher roots remain bound to the original observed corpus. The overview constructor is a distinct read consumer. | `PlanningSummarySource` exposes summary, task-graph, and audit-snapshot reads (`internal/core/space_overview.go:14`), not a complete mutation service. Its bare `NewFS` is not an identity-parity defect. |

Real discovery walks upward from nested starts and resolves the pointer from its marker. An opted-in pointer independently rejects a different or absent target ID (`internal/config/config.go:291`). The same-ID root tests deliberately preserve ID and task contents, so that pointer check cannot mask a missing physical-root guard. Read-only metadata stays a snapshot; apply revalidates it. Config remains a local secondary/composition concern, with no changed public wire fields or core/invocation config types.

### First pass: restored execution and document evidence

Before any mutation, these exact commands succeeded:

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/appwiring ./internal/workspacestore -count=1
GOCACHE=/tmp/taskflow-review-go-cache go test -race ./internal/appwiring ./internal/workspacestore -count=1
GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache just lint
GOCACHE=/tmp/taskflow-review-go-cache just docs-check
```

The first focused result was appwiring `ok` in 0.814s and workspacestore `ok` in 0.410s. Lint reported `0 issues`; generated CLI docs had no diff.

The shipped `TestPlanningOpenersThreadApplyParity` (`internal/appwiring/thread_apply_test.go:20`) exercises four real production routes: direct/ordinary, direct/workspace, pointer/ordinary, pointer/workspace. `newThreadApplyEntry` initializes actual planning trees and pointer markers and starts from `work/nested`; `openThreadApplyEntry` uses `LocalBindings().Compose(...).OpenPlanning` or `.Workspaces.Open` (`:243`). `composeThreadApplyEntry` supplies tagged valid tasks, an explicit nonmember gate (`Member=false`), and a real dependency (`:260`, `:269`). All four dry-runs had two operations and no document changes; all four commits completed the dependency and Thread.

Supplemental **temporary reviewer tests**, archived rather than shipped, independently checked actual bytes and live policy revocation. Command: `GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/appwiring -count=1 -run '^TestAudit' -v`; corrected fixture run exited 0 (`evidence/extra-appwiring.log`). In all four routes the member task SHA-256 changed from `11bd59533fe9cc42204065a052634f136d71518d963d63448051eaa7bbac18d0` to `68487b910b0a0b1bf3479fbff25095d9750ad773f12b1bf9baa4a2938fbc809e`, while the gate stayed `e7417c8b644bc1d19e24912f8262eaf43fbbb5482a45481626c3aa7f21e876c6`. Exactly one Thread document was added, with the materialized plan body. Compose and dry-run preserved all five entity-directory byte maps. The retained policy subsequently changed from allowed to denied: both dry-run and commit returned the denial, kept the typed attempted token, and preserved committed document bytes; total callback count was four.

Restored identity refusal rows (`TestPlanningOpenersRejectChangedIdentityBeforeThreadApply`, `thread_apply_test.go:56`) all passed for four openers, four changes, and both apply modes. Exact guards/classes were:

- Same-ID matching-task root replacement: `domain.ErrConflict`, `instead of guarded root`, with old and alternate corpus byte maps unchanged.
- Direct ID replacement: `domain.ErrConflict`, `apply plan belongs to planning repository`; pointer ID replacement: `domain.ErrConflict`, `this pointer expects`.
- Direct marker removal: `domain.ErrValidation`, `no durable id`; pointer target marker removal: `domain.ErrConflict`, `carries no id`.
- Malformed original marker: `domain.ErrValidation`, `parse .tskflwctl.toml`.

Each refusal is checked for `*core.ThreadApplyFailure`, uncommitted/incomplete receipt, and the original token in both the returned receipt and failure receipt (`thread_apply_test.go:166`). The valid initial dry-run prevents earlier tag/manifest refusals from masquerading as identity protection. `TestPlanningOpenersPreserveThreadApplyAuthorization` (`:113`) similarly checks zero opening/compose calls, exactly one call on denied apply, two after allowing the same live policy, and unchanged bytes on denial.

`GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/workspacestore -count=1 -run '^TestAudit' -v` also exited 0 (`evidence/extra-reader.log`). Real valid initial fixtures were followed by fresh partial data **with an error**, and fresh matching root **without an ID**, in dry-run and commit. The former preserved the sentinel through `errors.Is`; the latter returned `domain.ErrValidation`/`no durable id`. Both preserved typed tokens and all document bytes. Shipped nil-result, empty-root, reader-error, missing-construction-input, legacy, and caller-config-mutation tests were executed and individually challenged below.

No optional wire branch was added by this delta. The non-default `Member=false` manifest branch was exercised in semantic compose/apply, and the recovery workspace tests supplied a nonempty matching `WorkspaceRequest.ExpectedPlanningID`. Existing mismatching expected-ID core tests ran in the restored full suite; this is not a claim of new serializer coverage.

### Compiler-valid mutation matrix

All mutations were temporary and restored from sandbox `HEAD`, including the untracked-at-handoff files captured in that commit. No probe was restored from base main. Exact scripts, commands, selected rows, receipt/error output, and byte observations are retained in `.git/isolated-review-workspace/evidence/` (`mutations.py`, `second_mutations.py`, `byte_mutations.py`, `eager_formatted.py`). They are review artifacts, not delivered project fixtures. For a self-restoring M1 replay after cleanup, run `python3 .git/isolated-review-workspace/evidence/replay-M1.py` inside this sandbox. It materializes the archived temporary fixture, verifies the restored fixture passes, checks that the mutant survives focused tests and lint, requires the hostile fixture to fail, and restores the captured source/removes temporary tests in `finally`. This replay was executed successfully; `replay-M1-*.log` records exits 0, 0, 0, and 1 in that order.

Command notation below is executable substitution:

```sh
R=TestPlanningOpenersRejectChangedIdentityBeforeThreadApply
A=TestPlanningOpenersPreserveThreadApplyAuthorization
H=TestPlanningOpenersThreadApplyParity
S=TestNewPlanningStoreUsesOneObservedCorpusAndDefersIdentityReads
C=TestNewPlanningStoreRejectsMissingConstructionInputs
F=TestNewPlanningStoreIdentityReaderFailuresNeverUseCachedIdentity
L=TestNewPlanningStoreRechecksLegacyRootWithoutInventingIdentity
D=TestLocalBindingReadsAreDeferredUntilTheirHooks
# AW: GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/appwiring -count=1 -run '<pattern>' -v
# WS: GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/workspacestore -count=1 -run '<pattern>' -v
```

| Probe | Exact changed behavior | Command selection | Observed result and guard reached |
| --- | --- | --- | --- |
| P1 | In `planning.go`, replace `discover(discoveryStart)` with `cfg, error(nil)`; retain a blank use of the anchor for compilation. | AW `$R` | Exit 1; all 32 changed-identity rows fail on successful unsafe apply. Initial dry-run is valid. Dry mutant writes no documents; committing mutant changes member bytes and adds a Thread. |
| P2 | `store.WithMutationAuthorization(authorize)` becomes `store.WithMutationAuthorization(nil)` in shared construction. | AW `$A` | Exit 1; all eight rows fail, callback count 0. Denied dry-run succeeds with equal bytes; denied commit succeeds and persists dependency/Thread. |
| P3 | Workspace composition passes nil to `workspacestore.WithMutationAuthorization` in `wiring.go`; ordinary policy stays installed. | AW `$A/.*/workspace` | Exit 1; four workspace rows fail, count 0, unsafe dry/commit success and actual commit byte changes. This separately tests the upstream composition seam. |
| P4 | `discoveryStart := cfg.Dir` becomes `discoveryStart := cfg.Root`. | AW `$R/pointer/.*/repoint-root` | Exit 1; four pointer rows fail after a valid initial apply against a **flat** target. Repointing to a matching-ID/matching-task corpus is invisible to this mutant. Commit changes the old corpus; alternate corpus stays unchanged. |
| P5 | Disable only `repositoryLockKey(root) != repositoryLockKey(s.root)` in `currentPlanningIdentity`. | AW `$R/.*/.*/repoint-root` | Exit 1; all eight root rows fail with `err=<nil>` instead of `ErrConflict`. Same-ID replacement defeats pointer-ID masking. Old corpus gains the dependency/Thread; alternate stays unchanged. |
| P6 | Disable only `plan.PlanningRepoID != snapshot.PlanningRepoID` in `PrepareThreadApply`. | AW `$R/.*/.*/replace-id` | Exit 1; all four direct rows accept a changed ID. Pointer rows still pass because the independent pointer validator rejects first; they do **not** independently kill this core mutant. |
| P7 | Add an eager `discover(discoveryStart)` call to shared construction. | AW + WS `$S|$D` | Exit 1; both named tests detect the extra supplied-reader call. This does not establish protection against a direct real-reader call; P12/P13 settle that distinction. |
| P8 | Re-reader uses `discover(cfg.Dir)` instead of the copied anchor. | WS `$S` | Exit 1; mutation of caller config redirects re-read and produces an error instead of the valid dry-run. Store/watcher root remains original; the retained mutable anchor is the challenged failure. |
| P9 | Disable fresh nil/empty-root validation. | WS `$F` | Exit 1; nil-result dry-run panics in the mutant. Restored code returns typed `ErrValidation`/`returned no root` without writes. |
| P10 | On reader error, return cached `cfg.Root, cfg.ID, nil`. | WS `$F/reader_error` | Exit 1 for dry-run and commit: sentinel is lost and apply succeeds instead of refusing. |
| P11 | Substitute `invented` for an empty durable ID in `currentPlanningIdentity`. | WS `$L` | Exit 1 on wrong `ErrConflict`/plan-ID refusal instead of `ErrValidation`/`no durable id`. Core's sibling comparison still prevents writes; this is not proof that the test alone kills every permissive missing-ID mutant. |
| P12 | Add discarded `config.Discover(discoveryStart)` directly before shared store creation. | AW + WS `.` | Exit 0: all captured focused tests survive forbidden eager real discovery. Demonstrated coverage gap, M1. |
| P13 | Eagerly call real `config.Discover`; when successful, replace local `cfg` with the fresh result before store construction. | AW + WS `^(TestLocal|TestOpened|TestPlanning|TestNewPlanning|TestFS_)`; then full suite/lint commands below | Exit 0 for all shipped tests and formatted lint. Temporary hostile initial-repoint fixture exits 1 for both direct and pointer, showing original metadata root plus alternate task/store root with reader counter still 1. M1. |
| P14 | Coordinated P6 plus bypass `verifyPlanningRepoID`'s expected-ID check. | AW `$R/pointer/.*/replace-id` | Exit 1; all four previously masked pointer rows now fail with successful unsafe dry/commit apply, and commit changes real dependency/Thread bytes. |
| P15 | Remove the constructor's `cfg == nil` check while retaining the root check. | WS `$C/nil_config` | Exit 1; compiler-valid mutant panics, restored constructor returns `ErrValidation` and nil store. |
| P16 | Remove constructor empty-root rejection while retaining nil rejection. | WS `$C/empty_root` | Exit 1; mutant returns a nonnil store without the required validation error. |
| P17 | Disable constructor nil-discoverer rejection. | WS `$C/nil_discovery` | Exit 1; mutant returns a store instead of immediately rejecting missing identity capability. |
| P18 | Remove only the fresh empty-root branch, retain fresh nil check. | WS `$F/empty_root` | Exit 1 for dry-run and commit. Downstream physical-root `ErrConflict` masks persistence, but fails the required `ErrValidation`/`returned no root` contract. |
| P19 | Disable only the store's empty-ID guard. | WS `$L` | Exit 0: core's empty-snapshot-ID validation still rejects. Record this survivor as masking, not as unsafe success or complete coverage. |
| P20 | Coordinated P19 plus disable core's empty-snapshot-ID guard and plan/current-ID comparison in `PrepareThreadApply`. | WS `$L` | Exit 1: dry-run now succeeds against genuinely markerless legacy state, so the test kills the permissive coordinated mutant. |
| P21 | Recreate predecessor workspace wiring: construct `store.NewFS` with authorization but without the identity reader. | AW `$H/.*/workspace` | Exit 1: both workspace rows fail their first valid dry-run with `ErrValidation`/`planning repository identity cannot be re-read for Thread apply`. Ordinary paths are unchanged. |

For P13, the exact broader commands were:

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./... -count=1 -skip '^TestAudit'
GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache just lint
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/appwiring -count=1 -run TestAuditObservedCorpusSurvivesMarkerRepointDuringOpen -v
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/cli -count=1 -run 'TestCommandTreeConstructionDoesNotReadInvocationData|TestInjectedPlanningOpenerRunsAfterCompletionFlagsAndSafety|TestRuntimeWorkspacePreservesResolvedStartupContext' -v
```

They exited 0, 0 (`0 issues` after formatting scratch sources/mutant), 1 (demonstrated mixed corpus), and 0 respectively. `TestAudit...` are temporary reviewer additions; the `-skip` command runs the shipped suite without them. Early exploratory lint failures were formatting of temporary probes, and were corrected before claiming this survivor. The new fake/counted-reader tests and real-context runtime tests do not observe direct real discovery in the helper.

Byte-observing reruns instrumented the shipped refusal rows without changing fixtures or decisions, then restored those test sources. `bytes-*.log` proves P1–P6/P14 are unsafe success rather than invalid-fixture failures: selected dry rows keep the byte map; selected commits change the same member SHA from `11bd5953...18d0` to `68487b91...809e`, add exactly one Thread, and show `err=<nil>`, `Committed=true`, `Complete=true`. Root-repoint maps cover **both** corpora; counts increase from 14 to 15 in the original+alternate snapshot, while alternate documents stay byte-identical. Authorization mutants show callback count zero. The intended named assertions fail with exit 1, not a compilation error. P9/P15 intentionally demonstrate runtime panics only in mutated code.

### Second pass: systemic seam, timing, and recovery

The strongest initial systemic hypothesis was that consolidation could rescan or mix corpora while counters appeared to prove laziness. P7 alone rejected only injected-reader eagerness. P12/P13 survived shipped validation; the temporary production-opener marker-repoint fixture supplied hostile evidence of the metadata/store split under P13, and passed restored code. M1 records the resulting coverage weakness without alleging eager discovery in the actual helper.

The secondary seam itself is appropriate: it constructs a concrete local store using config types while callers publish neutral capability bundles. It installs the existing store policy rather than duplicating mutation decisions. `GOCACHE=/tmp/taskflow-review-go-cache go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/appwiring ./internal/workspacestore ./internal/core ./internal/cli/ports ./internal/spacestore` verified the real graph: no direct appwiring-to-store edge remains; workspacestore directly imports config/core/domain/store; core imports domain/id, and CLI ports core/design. The omitted domain edge in the diagram is L1.

The actual lint protections were challenged, not just inventoried. Temporary config imports and blank type uses in `core/workspace.go`, `cli/ports/runtime.go`, and `appwiring/wiring.go` compiled under `go test ... -run '^$'` (exit 0). `just lint` rejected all three with the exact `core-owns-ports-not-adapters`, `cli-composition-contracts-stay-neutral`, and `appwiring-startup-reads-use-local-sources` depguard rules (`evidence/boundary-lint.log`). Other exploratory scratch formatting/unused diagnostics are separately visible in that log; the named depguard refusals are direct evidence. The legal secondary-adapter direct-discovery P13 mutant, after formatting, passes lint: those rules do not close M1.

Temporary `TestAuditRealOpenerPreflightAndPrefixIdentityRecovery` used actual `LocalBindings` ordinary/workspace services, direct/pointer nested starts, tagged tasks, a dependency, and a nonempty expected workspace ID. Test-only accessors installed existing store hooks; they did not replace the store, identity reader, or opener. Command:

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/store -count=1 -run 'TestAuditRealOpenerPreflightAndPrefixIdentityRecovery|TestThreadApply' -v
```

Exit 0 (`evidence/recovery.log`); all eight timing/opening combinations passed, along with the existing Thread-apply guard/recovery tests. Marker ID replacement just before whole-source verification produced `domain.ErrConflict`, typed token preservation, `Committed=false`, `Complete=false`, no new Thread, and identical member SHA `f967d90e7aec2f1d79e8e738fa2ed709fb8e16ed9d5b3af1e47287bd5fc92cc2`. Replacing identity immediately after the dependency write instead returned `Committed=true`, `Complete=false`, one pending operation, no Thread, and changed dependency bytes. Restoring the exact original marker and retrying the **same** materialized plan completed every case; prefix retries skipped the already-applied dependency. This distinguishes pre-write refusal from partial recovery, rather than assuming a new transaction guarantee.

Recovery recommendations were actually executed by the captured CLI binary against the recommending fixture state, with an isolated `TSKFLW_CONFIG_HOME` and `init --path <fixture> --no-register` to keep initialization explicit:

- Direct markerless apply recommends `tskflwctl config migrate` from unchanged `currentPlanningIdentity` (`internal/store/threadapply.go:218`). Running `tskflwctl -C <fixture> config migrate --no-input --json` exits **11**, machine code `validation`: `no .tskflwctl.toml governs ... — run tskflwctl init first`. This is an existing recovery-message limitation, not a new parity regression; `migrationConfig` requires a marker (`internal/config/migrate.go:143`). Running the emitted next-step `init --path <fixture> --no-register --no-input --json` succeeds and creates only the marker. A newly composed plan dry-runs successfully; the old durable plan remains conflicting with the newly minted ID.
- Pointer target marker loss recommends target `init` or removal of `planning_repo_id` (`internal/config/config.go:295`). Target init succeeds and mints a new ID; the old opted-in pointer correctly still refuses that different identity. Executing the suggested opt-out allows discovery of the new target ID, but core still rejects the old plan's repository ID. Neither alternative silently adopts the stale plan. Semantic document byte maps stay unchanged throughout repair.

The first scratch recovery harness used `-C` as if it controlled init's `--path`; it initialized only fresh scratch directories inside the sandbox's appwiring package and hit the filesystem sandbox when optional home registration was attempted. Those scratch files were removed, and the corrected explicit-path/no-registration fixture run is the reported passing evidence. The real markerless migration refusal remained after correcting the harness; it is documented as an out-of-scope followup rather than attributed to this implementation.

No shared abstraction, source-set fallback, overview constructor, or nil policy was certified from a name alone. The restored helper's pinned scalar fields, real-opener parity tests, hostile repoint fixture, authorization counts, and coordinated masks settled those hypotheses. Unknown callers returning forged/non-discovered configs remain outside the constructor's already-discovered-input premise.

### Validation, restoration, and residual limits

Reviewer-added probes also passed under race detection before removal:

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test -race ./internal/appwiring ./internal/workspacestore ./internal/store -count=1 -run '^TestAudit'
```

Exit 0 (`evidence/extra-race.log`). Temporary test sources and mutation scripts/logs were archived under sandbox-only `.git/isolated-review-workspace/evidence`; package scratch files and mistaken scratch initialization outputs were removed. Every tracked production/test/generated file was compared against sandbox `HEAD`, and `git diff --name-only` was empty before report edits. No further commit, branch switch, or staging occurred. After the final self-restoring M1 replay, focused tests were repeated with `GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/appwiring ./internal/workspacestore -count=1` and passed (`evidence/final-restored-focused.log`).

On the restored captured baseline with temporary tests removed, these commands exited 0:

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./... -count=1
GOCACHE=/tmp/taskflow-review-go-cache go test -race ./internal/appwiring ./internal/workspacestore -count=1
GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache just lint
GOCACHE=/tmp/taskflow-review-go-cache just docs-check
```

Retained outputs: `evidence/restored-full.log`, `restored-focused-race.log`, `restored-lint.log`, `restored-docs.log`. Final focused race results were appwiring `ok` 2.873s, workspacestore `ok` 1.308s; lint `0 issues`; generated docs unchanged. The full suite was warranted because probes touched shared core/config/store guards. The owner validation prose was treated as a claim; the commands above are independent reviewer evidence. A captured CLI build also succeeded for recovery/finding commands; its build emitted an unwritable external Go stat-cache diagnostic, without a build failure.

Tested platform/toolchain: **Darwin arm64, Go 1.27.1**. Linux, Windows, other filesystems, and cross-process adversarial scheduling were not executed. Full uncached race across the entire repository was not repeated; focused production and reviewer timing probes were run under race. The tested hook windows do not prove absence of every non-cooperating edit between individual verifications/writes, root inode replacement, marker read TOCTOU, or arbitrary future UI/web workflow. Existing advisory-guard and durable-prefix semantics remain the scope; nil authorization is unchanged, and no new TUI Thread-apply feature is claimed. The markerless migration suggestion is a verified pre-existing followup. Findings remain open; this report does not close the implementation task or certify every mutation family.

### Mandatory sandbox and guarded-transfer attestation

The independent clone was created by the general helper with `--no-hardlinks` and the working-state overlay. Its resolved Git directory is an in-tree directory with no alternates and exactly one worktree, as checked by the helper. Only the helper's baseline commit was created. State captured by the helper:

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.AGSEPw
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.AGSEPw/.git
baseline_commit=3d6496ed702e3992d97c923668d4e99f2f600ece
source_blob=9f2c1caf35fc7741dd3d18e93614005736d9571e
source_fingerprint=275985c5c2c5c19061e6e06eb87346ee934bd928
deliverable=planning/audits/6ggee2fx4053-2026-10-04-workspace-thread-apply-identity-parity-implementation-codex.md
deliverable_changed=true
transfer=succeeded
```

Before transfer, the sandbox helper verifies unchanged baseline HEAD, no staging, no unrelated tracked/untracked changes, independent Git metadata, nonempty report, and unchanged source-deliverable blob. The assigned audit's diff is inspected. Transfer uses only `scripts/isolated-review-workspace.sh transfer --sandbox <the attested path>` and its atomic one-file rename; no source manual copy or unrelated write is permitted. The successful transfer result is confirmed by the guarded helper stdout for this delivered report. Final verify/transfer stdout is retained in `evidence/final-verify.log` and `evidence/final-transfer.log`. The sandbox is retained for implementation-owner receipt and cleanup confirmation.

## Findings

#### M1. Shared-construction regressions miss direct rediscovery and a mixed initial corpus · **Status:** fixed locally (2026-10-04)

**File:** internal/workspacestore/planning_test.go:22 | **Component:** appwiring
**Effort:** S · **Urgency:** soon

This is a demonstrated regression-test weakness, not a claim that the restored constructor currently rediscovers. `TestNewPlanningStoreUsesOneObservedCorpusAndDefersIdentityReads` counts only its injected callback (`internal/workspacestore/planning_test.go:22`); `TestLocalBindingReadsAreDeferredUntilTheirHooks` likewise wraps only `localSources.discover` (`internal/appwiring/wiring_test.go:65`). Neither observes a direct call to the real reader inside the shared constructor.

Two compiler-valid mutants before the `store.NewFS` return at `internal/workspacestore/planning.go:27` survived the complete captured appwiring/workspacestore tests: a discarded `config.Discover(discoveryStart)` call, and an eager call that adopts the newly discovered configuration. After formatting, the adopting mutant also passed `GOCACHE=/tmp/taskflow-review-go-cache go test ./... -count=1 -skip '^TestAudit'` (all shipped tests) and `GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache just lint` (0 issues). `TestAudit...` names denote temporary reviewer fixtures only, archived in the retained sandbox; none are shipped.

The adopting mutant demonstrably breaks the stated one-observation contract: a temporary fixture used the production `bindingsFor(...).Compose(...).OpenPlanning` path with a wrapper around real `config.Discover`, then repointed the marker to a same-ID copied corpus before returning the initial observation. The ordinary opener returned the original metadata root while `ShowTask` read a distinguishable `replacement-corpus` task from the alternate root, with the injected discovery count still exactly one. The restored implementation passes this fixture; the mutant fails both direct and pointer cases. This settles the hostile hypothesis without relying on malformed manifests, missing tags, or a substitute store.

Reproduction and exact outputs remain under `.git/isolated-review-workspace/evidence/` in the attested sandbox: `replay-M1.py`, `eager_formatted.py`, `eager-formatted-shipped-full.log`, `eager-formatted-lint-clean.log`, `eager-formatted-hostile.log`, and `appwiring-audit_probe_test.go`. Integrate a hostile initial-observation/repoint regression and an executable rule or observation that detects direct real-reader calls in the shared constructor. Wrapping the injected reader alone does not protect the runtime seam.

**Recommendation:** Add the hostile initial-observation/repoint regression and enforce observation of direct real-reader calls in the shared constructor.

**Resolution:** Accepted as a demonstrated regression gap, not a production
rediscovery defect. Added
TestPlanningOpenRetainsInitialCorpusWhenMarkerChangesDuringDiscovery: real
direct/pointer discovery returns the first observation after a same-ID marker
repoint; metadata, readable tasks, watcher roots, compose, and refused
dry/committing applies remain coherent and preserve both corpora.
TestSharedPlanningConstructorUsesConfigOnlyAsData makes planning.go config
access data-only, detecting direct calls and reader-value references under
renamed imports. In an independent no-Git copy, the discarded call fails the
fitness test, the adopting call fails both the behavioral and fitness tests, and
an aliased reader reference fails the fitness test. Restored focused race and
full uncached race tests, lint, build, and docs checks pass. Local and
uncommitted; no generalized ban on all possible indirect I/O is claimed.

#### L1. Direct adapter graph omits the new workspacestore domain dependency · **Status:** fixed locally (2026-10-04)

**File:** docs/ARCHITECTURE.md:49 | **Component:** documentation
**Effort:** XS · **Urgency:** eventually

The direct adapter graph at `docs/ARCHITECTURE.md:49` still lists `workspacestore -> config, core, store`, but the new constructor directly imports `internal/domain` at `internal/workspacestore/planning.go:7` for `ErrValidation`. Independent `GOCACHE=/tmp/taskflow-review-go-cache go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/workspacestore` reports config, core, domain, and store. The dependency direction is appropriate; the published inventory needs the added domain edge. No cycle or lint-rule violation is alleged.

**Recommendation:** Add domain to the workspacestore row in the direct adapter graph.

**Resolution:** Accepted. Added domain to the workspacestore direct dependency
row in docs/ARCHITECTURE.md, matching planning.go and the actual go list import
inventory. The inward dependency is intentional; no port or production import
was changed by this documentation fix.

#### L2. Markerless Thread identity recovery recommends a migration that cannot run · **Status:** tracked by 6ggfd81jg0qg

**File:** internal/store/threadapply.go:218 | **Component:** recovery
**Effort:** S · **Urgency:** eventually

Owner triage of the executed recovery scenario in this review: a direct markerless planning tree refuses Thread apply with a config migrate suggestion, but migrationConfig requires a governing marker and that command instead refuses with an init prerequisite. Related compose/prepare guidance has the same assumption. This predates workspace identity parity; retain the ID guards and handle the diagnostic repair separately.

**Recommendation:** Execute each emitted recovery instruction against marker-backed, markerless, and missing pointer-target identity states; keep initialization and stale-plan invalidation explicit.

**Resolution:** Pre-existing diagnostic issue from the reviewer narrative is
transferred to the bounded ID-less recovery task, with executable repair-command
criteria and stale-plan/pointer protections. Added to the CLI-contract Thread
after the completed configuration lifecycle foundation; independent of
adapter-neutral refactor closeout.

## Owner reconciliation (2026-10-04)

Accepted M1 as a regression-test gap around correct production code and L1 as documentation drift;
both are fixed locally with the finding-specific evidence above. Replayed discarded, adopting,
and aliased-reference direct-discovery mutants in an independent copy, restored them, and reran
focused race tests. Full `go test -race -count=1 ./...`, `just lint`, `just docs-check`, and
`just build` pass. No mutable global discovery hook, new application config dependency, or
production semantic change was needed.

The executed markerless recovery scenario in the report is now recorded as owner-triaged L2,
tracked by [6ggfd81jg0qg](../tasks/6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md).
It belongs to the CLI-contract Thread, not an additional adapter-neutral closeout prerequisite.
All findings are settled locally; audit closure is not a claim of merge or release inclusion.
