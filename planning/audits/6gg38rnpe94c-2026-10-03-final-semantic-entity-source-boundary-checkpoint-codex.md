---
schema: 1
id: 6gg38rnpe94c
bucket: closed
area: final-semantic-entity-source-boundary-checkpoint-codex
date: "2026-10-03"
updated_at: "2026-10-03"
---
# Audit: Final semantic-entity source-boundary checkpoint — codex — 2026-10-03

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

Adversarially review the final semantic-entity source-boundary checkpoint for task
`6gcwcf88z57p`, `split-local-path-capabilities-from-semantic-entity-reads`, in draft PR #274.
Task, Thread, Epic, Audit, and Research now contain no `Path`, `FilenameID`, or
`SourceVersion`; Thread's `CanonicalID()` fallback is removed too. Decide whether
this task can be completed after triage, not merely whether the removed fields compile.

Do not implement fixes. Preserve this brief, update only your assigned audit,
and leave findings open. Distinguish regressions, pre-existing problems exposed by
this migration, and optional improvements. Earlier checkpoint findings have been
triaged; reopen them only with new failing evidence.

## Review target

Branch: `refactor/split-local-entity-source-capabilities`.
Implementation freeze: `506b7990c222a577c751ad87685a90a437d1fb04`.
Primary delta: `2d4d9f1..506b799` (Task path removal, Thread source-boundary
migration, and Thread callback isolation). The sandbox baseline additionally captures
the final planning update and these briefs; it must contain no later implementation.

Relevant implementation commits are `a31bd42`, `7afaecf`, and `506b799`.
The earlier local-capability and Task-identity checkpoint audits are
`planning/audits/6gfrtba5jky2-*`, `6gfrtbae3bp4-*`, `6gfyn1n6wyn6-*`, and
`6gfyn1nhawqq-*`. Read their owner dispositions as scope context, not as proof
that this final delta is correct.

Build a producer/consumer inventory for source ID, declared ID, diagnostic location,
path presentation hint, executable local handle, and source revision. Trace normal
and bulk-apply Thread scans, selected reads, semantic planner snapshots, source
validators, materialization, snapshot CAS, repair/lifecycle impact receipts, ordinary
CLI/wire projections, and TUI selection/local actions. Check production call sites,
not just the names or assertions of the new tests.

## Intended contract to challenge

1. Domain IDs are declarations. `RecordSource.ID` is the adapter's resolution
   identity; it remains independent when a declaration is missing or wrong.
   Ordinary reads/lint retain malformed records and attribute defects honestly.
2. `Location` is diagnostic context. `LocationIsPath` affects presentation only.
   Ordinary local actions request the optional path port; guarded local writers
   require `VersionedRecord.LocalPath` and the original `SourceVersion`.
   Neither path-shaped text nor a path hint creates write authority.
3. `ValidateThreadCreationSource`/`ValidateThreadMutationSource` validate the
   complete source-aware read before semantic planner values are constructed.
   Missing/duplicate source IDs, declaration drift, invalid documents, cross-kind
   collisions, missing members, and unreadable Threads block ordinary guarded work.
   Pure apply-snapshot validation does not substitute for this adapter source gate.
4. Thread source materialization and complete-snapshot CAS fail closed on changed
   handles, renamed sources, missing revisions, and changed readable/unreadable
   bytes. Receipts obtain local paths from operation metadata, never domain values.
5. Broken-graph repair can proceed past malformed Thread evidence. Readable Thread
   impact projections must retain their independent source identity/location,
   including declaration drift, and must not become healthy merely because the
   underlying task graph was repaired. Repair does not edit Thread documents.
6. Callback-owned Thread slices, nested tags/tasks, and apply-body maps cannot
   rewrite the owner's authorization snapshot. A cancelled Thread cannot be
   revived by a callback pretending its starting status was unstarted.
7. Valid-file CLI/wire behavior and persisted Markdown contracts remain compatible.
   Pathless adapters do not manufacture filenames, source IDs, or revision tokens.
   Bare Task/Thread compatibility projections do not impersonate a source-aware read.

## Mandatory evidence floor

- Inventory the actual producers/consumers with exact path/line evidence. Verify
  each removed-field test migration retained the hostile state it originally tested;
  equal-ID or pathless happy paths do not prove drift/duplicate behavior.
- Exercise matching IDs, missing declarations, different valid declarations,
  duplicate source IDs at distinct paths, and distinct sources sharing a declaration.
  Check Thread list/show, lint, source-ID path recovery, and at least one guarded
  operation. Include an unrelated malformed Thread and a pathless loaded record.
- Demonstrate the source gate at creation, membership/lifecycle, task lifecycle,
  compose, initial bulk apply, and final apply convergence. Prove failed preflight
  invokes no planner/write where applicable; do not assume one validator call
  implies every caller is safe. Check no-op and dry-run paths separately.
- Verify repair receipts with a readable drifting Thread in both dry-run and
  committed repair: task graph improves, Thread remains broken, canonical impact
  ID and defect location survive, and Thread bytes remain unchanged.
- Test a path-shaped location with the path hint enabled but no local handle;
  then test an opaque location with a separately supplied valid local handle.
  Include missing/stale revisions and readable/unreadable representation changes.
- Perform at least three restored, compiling mutation probes. Mandatory targets:
  (a) guarded local-path authority or handle CAS; (b) source identity retained in
  repair impacts; (c) callback snapshot isolation. Name the exact changed guard,
  focused regression, and observed behavioral failure. A compile error is not a kill.
  A green mutant must be investigated, not counted as success because another
  layer was assumed to protect it.
- Compare representative valid human/JSON outputs with the base, inspect generated
  schema comments, and verify revisions/local handles never leak through public
  projections. Report any actual compatibility change rather than silently
  approving an updated fixture.

## Required hostile angles

- Could a semantic-only constructor or the new pure snapshot validator be reached
  in production where source identity is required? Find the route and test it.
- Does `ThreadRead.LoadedThreads()` preserve every readable occurrence and own
  mutable slices while excluding guarded metadata? Can repair impacts collapse
  duplicate sources, lose drift, or use the declared ID as an actionable fallback?
- Does materialization compare against the original read revision or accidentally
  mint a fresh token after a race? Can an idempotent result report a fabricated
  snapshot because a callback or helper changed the validation inputs?
- Could a renamed/deleted record, changed body, or contradictory source/location
  pass one path while failing another? Separate guarded-write safety from the
  documented best-effort external-editor pathname race.
- Challenge shared test helpers: `graphFixturePath`, `semanticThreadRead`, and
  Thread fakes can only represent certain source states. Introduce independent
  evidence rather than relying on a helper that couples declaration and source ID.
- Second pass: assume a future database adapter with an opaque key, no filename,
  and no local path. Find remaining assumptions that would force it to invent
  local evidence or weaken its complete-snapshot guard.

Codex emphasis: trace architecture and compatibility routes end to end, including
all consumers of `SemanticThreads()`, `LoadedThreads()`, bare `ProjectThread`, and
repair/lifecycle Thread impacts. Test any fallback or inconsistent authority gate.

Antigravity emphasis: disprove these three concrete claims before broadening scope:
(1) a path hint cannot replace a local write handle; (2) graph repair cannot hide
readable Thread drift; (3) no Thread planner can mutate its own authorization inputs.
Use `internal/store/thread_source_boundary_test.go`,
`internal/core/thread_source_boundary_test.go`, and
`internal/store/thread_planner_snapshot_test.go` as entry points, not as proof.
Remove the relevant protection in your sandbox and require a behavioral test failure;
for callback isolation, challenge creation and apply nested values/body maps too,
not only the mutation scalar status. Then inspect call sites those tests do not reach.

## Validation and restoration

Use only the mandatory independent sandbox. Build a fresh sandbox binary for CLI
probes; do not use an installed binary that predates this branch. Run focused tests,
`go test ./...`, `go test -race ./internal/core ./internal/store ./internal/tui ./internal/wire`,
`just lint`, fresh-binary planning lint, schema-comment regeneration, and
`git diff --check` where feasible. Report checks not run and why.

Restore every probe to the sandbox baseline before the next probe and before transfer.
Transfer only your assigned audit through the helper. Do not push, stage, or create
additional commits, and never run mutations in the shared source checkout.

## Deliverable

Add a severity-ranked verdict, producer/consumer inventory, hostile fixture and
restored-mutation tables, commands/results, compatibility observations, and sandbox
attestation. Add actionable findings under `## Findings` with exact repository
grammar and open statuses. Include confidence limits and any untested critical seam.
Explicitly answer: can the implementing task now be completed, and what must change
first? A partial review must say partial rather than declare readiness.

## Reviewer report

### Verdict and review scope

**Both technical review passes are complete. Owner triage is required before an unconditional completion verdict:** no demonstrated implementation regression in `2d4d9f1..506b799`, and one medium, pre-existing repair-projection defect (`M1`) exposed by the final contract. No high findings or speculative optional improvements are entered. Leave `M1` open. Task `6gcwcf88z57p` can be completed after duplicate-source repair impacts are corrected and covered, or after the owner explicitly accepts a narrower completion scope and records the outstanding defect in a linked follow-up. This audit does not make that owner decision.

Reviewed the frozen implementation `506b7990c222a577c751ad87685a90a437d1fb04`, including `a31bd42`, `7afaecf`, and `506b799`. `git diff --name-only 506b799..HEAD` in the sandbox lists only the two final audit briefs and the implementing planning task; there is no later implementation. Earlier audit dispositions were read as scope context. In particular, the settled declaration-preserving Thread wire contract is retained, and the previously fixed repair reload uses an explicit filesystem source rather than the bare graph constructor. Neither settled finding is reopened.

### First pass: producer and consumer inventory

All paths and line numbers below refer to the captured sandbox implementation. These are shipped symbols, not capabilities inferred from the planning task.

| Evidence | Producers | Consumers and boundary behavior |
| --- | --- | --- |
| Declared IDs and semantic values | Task YAML parser `internal/store/fsstore.go:392`; Thread YAML parser `internal/store/threadstore.go:145`; semantic structs `internal/domain/task.go:5`, `internal/domain/thread.go:52` | `RecordSource.ID` is separately supplied. `TestSemanticEntitiesDoNotCarrySourceEvidence` at `internal/domain/identity_test.go:8` checks all five entities and the removed `CanonicalID` method. Reintroducing the fields/method produced compiling test failures. |
| Task source ID, location and revisions | `taskSource`, `internal/store/entity_read.go:14`; `scanTaskDocuments`, `internal/store/fsstore.go:163`; `ReadTaskGraph`, `internal/store/fsstore.go:194` | Guarded records carry source ID, exact-byte revision and explicit local handle. Ordinary `ReadTasks`/`ReadTask` use `internal/store/entity_read.go:61` and `:69`. `NewTaskGraphRead`, `internal/core/dependency_graph.go:335`, retains every occurrence; `SameSourceSnapshot`, `:965`, compares complete readable/unreadable evidence. |
| Other entity sources | `epicRecord`, `internal/store/entity_read.go:19`; `auditSource`, `:26`; `researchSource`, `:38`; selected/read ports `:73`, `:85`, `:97` | Epic adapter identity is the filename stem assigned by `parseEpic` at `internal/store/epicstore.go:296`; it is not a generated remote filename requirement. Task/Epic/Audit/Research wire projections use loaded source identity in `internal/wire/dto.go:55`, `:382`, `:272`, `:252`. |
| Normal and selected Thread source | `ReadThreads`, `internal/store/threadstore.go:23`; `threadSource`, `:47`; `threadReadFromSourceFiles`, `:52`; unreadable recovery `:72`; selected `ReadThread`, `:94` | Filename ID belongs to the FS adapter. Readable values with missing/drifting declarations survive. Selected reads recover the independent source ID/location. `resolveThread` at `:119` and the optional path port remain parse-free; duplicate source IDs make selected resolution ambiguous. |
| Bulk Thread source and body | `listThreadApplyThreads`, `internal/store/threadapply.go:228`, source wrapper at `:239`, body map key at `:250` | Every readable occurrence survives in `ThreadRead`; a body-map overwrite cannot bypass the source gate. Initial apply validates at `:57`; final convergence rereads and validates at `:271` before pure `PrepareThreadApply`. |
| Diagnostic context and path hint | `RecordSource`, `internal/core/entity_read.go:29`; `VersionedRecord`, `:56` | `ProjectLoadedThread`, `internal/core/thread_projection.go:97`, uses source identity independently; `:103` uses the hint only for diagnostic presentation. `readableSourceLocation`, `internal/wire/dto.go:390`, suppresses redundant local location. No hint supplies a mutation handle. |
| Semantic planner inputs and source gates | `ThreadRead.ValidateSources`, `internal/core/store.go:169`; `ValidateThreadCreationSource`, `internal/core/thread_creation.go:69`; mutation gate `internal/core/thread_mutation.go:202` | Every production `SemanticThreads()` consumer was traced: Thread creation `internal/store/threadcreation.go:65`, mutation `internal/store/threadmutation.go:57`, initial/final apply `internal/store/threadapply.go:63`, `:276`, and compose `internal/core/service_thread_apply.go:41`. Their gates precede conversion. Task lifecycle validates Thread sources at `internal/store/lifecyclemutation.go:58`. Pure semantic validation at `internal/core/thread_creation.go:88` cannot prove source identity. |
| Local write authority and original revision | `VersionedRecord.LocalPath` and `SourceVersion`, `internal/core/entity_read.go:56`; FS scanners above | `materializeThreadMutation`, `internal/store/threadmutation.go:153`, requires the explicit handle and original revision, compares resolved handle at `:162`, and checks original bytes at `:169`; `verifyUnchanged` is called at `:104`. Whole-Thread CAS is `internal/store/cas.go:47`, readable comparison `:83`, unreadable comparison `:97`. The Unix repository lock opens/flocks the planning root directory (`internal/store/lock_unix.go:21`), not a claimed lock-file path. |
| Readable projections and impacts | `LoadedThreads`, `internal/core/store.go:144`, clones tags/tasks and keeps each source; `SemanticThreads`, `:157`, carries semantic values only | All production `LoadedThreads()` consumers: source validation `internal/core/thread_creation.go:83`, graph repair `internal/store/graphrepair.go:57`, task lifecycle impacts `internal/store/lifecyclemutation.go:94`. Repair receipts call `TaskGraphThreadImpacts` at `internal/core/dependency_repair.go:675`; readable source retention is correct for drift, but whole-set duplicate qualification is missing (`M1`). |
| Compatibility, public output and local actions | Bare `NewTaskGraph`, `internal/core/dependency_graph.go:329`; bare `ProjectThread`, `internal/core/thread_projection.go:90` | Bare graph records have declaration IDs but no manufactured path/revision (`internal/core/service_task.go:219`, `:228`). Production bare `ProjectThread` calls are semantic mutation analysis at `internal/core/thread_mutation.go:230`, `:337`, plus the compatibility `ProjectThreadGraph` wrapper at `internal/core/thread_graph.go:55`; FS receipts reattach loaded source views. Ordinary service list/show use loaded views (`internal/core/service_thread.go:245`, `:380`, `:394`). TUI rows/selectors use source IDs (`internal/tui/thread_projection.go:32`); local actions call optional path ports (`internal/tui/commands.go:34`) and check selection/generation plus a second lookup (`internal/tui/model.go:392`). |

### Test migration integrity

Inspected the removed-source-field changes in the test delta, retained in `.git/review-evidence/all-test-migrations.diff` and `test-source-field-changes.txt` for this sandbox. No hostile drift/duplicate case was accepted merely because its replacement compiled.

| Migrated family | Hostile state retained or replaced |
| --- | --- |
| Task graph identity, shadow sources, repair and source simulation | `graphFixturePath` (`internal/core/dependency_graph_test.go:60`) is only a matching-ID helper. `localTaskGraphWithSourceIDs` at `:35` explicitly supplies different source IDs for missing/drifted declarations. Duplicate records retain distinct slugs/paths; custom representative/shadow paths use `localGuardedRecordAt`, including `internal/core/dependency_source_test.go:68` and repair shadow fixtures at `internal/core/dependency_repair_test.go:476`. Repeated dependencies, invalid raw strings, legacy field presence, and unreadable revisions remain. |
| Thread drift, collisions and representation changes | Drift/cross-kind projection moved to an explicit loaded record at `internal/core/thread_projection_test.go:121`. Snapshot representation/identity tests use one independent `RecordSource` at `internal/store/thread_source_snapshot_test.go:117`; changing a declaration no longer changes its source implicitly. |
| Shared fakes | `semanticThreadRead` (`internal/core/thread_creation_test.go:74`) intentionally couples IDs and lacks revisions; `threadReadFake` (`internal/core/service_thread_test.go:14`) has a single optional source-ID/location override for its collection. Neither proves arbitrary multi-source states. Independent physical files and explicit opaque wrappers below supply that evidence. |
| Local actions, receipts and pathless lint | Store tests obtain paths from operation metadata or `ResolveThreadPath`, then perform the same byte/race checks. The TUI test now includes a path-shaped location **with** its hint enabled (`internal/tui/local_path_test.go:31`). Contradictory domain-path receipt fixtures became unrepresentable; absence is checked structurally, and explicit local-outcome assertions remain in `internal/core/creation_receipt_test.go`. Opaque duplicate lint uses loaded records in `internal/core/lint_source_test.go`. |
| Removed transitional-path tests and default wire helpers | URI-valued semantic-path fixtures were removed because the field no longer exists. Their authority challenge remains in `TestTaskGraphRepairDoesNotSelectOpaqueLocations` (`internal/core/dependency_repair_test.go:62`), including enabled path hints. Bare wire helpers still exercise default values; they were supplemented with seven non-default schema-validation cases during this review. |

### Hostile fixtures and production routes

Reviewer-written tests/scripts are **scratch evidence, not committed regressions**. Their runnable copies and logs remain under `.git/review-evidence/` in the retained sandbox; temporary files placed under `internal/` were removed. The independent Thread files do not use `semanticThreadRead` to supply source identity.

| Fixture or route | Observed result |
| --- | --- |
| Fresh CLI: matching IDs | Thread list/show/path/lint and dry/real idempotent membership return exit 0; Markdown bytes unchanged. |
| Missing declaration / different valid declaration | Readable list/show preserve the record; source-ID path lookup succeeds; lint and guarded membership return exit 11. Drift show suppresses the frontier and names the source file while retaining declared Thread wire ID. |
| Duplicate source IDs at distinct paths | List retains both broken occurrences; show/path return ambiguity exit 13; lint and guarded membership return 11. |
| Distinct sources sharing one declaration | Both records remain readable; the drifting source is broken. Selected source-ID recovery works; lint and guarded membership reject. |
| Unrelated unreadable Thread | Partial list exits 11, selected valid Thread show/path succeed, and guarded dry/real membership rejects before writing. |
| Creation, membership no-op, Thread lifecycle, task lifecycle no-op, initial apply and compose | `TestReviewAllSourceGateRoutes` independently exercises missing/drifted declarations, duplicate sources, shared declarations, invalid document status, unreadability, missing members and cross-kind collision, with both dry-run values. All store planners remain uncalled. Compose does not invoke its ID generator. CLI probes additionally compare all Markdown bytes. |
| Final apply convergence | `TestReviewFinalApplySourceGate` adds an unrelated readable drifting Thread after a durable dependency prefix. Final source validation rejects; receipt remains incomplete with the committed prefix and no target Thread write. Removing the final gate produces a compiling behavioral failure. Existing unreadable/raw-source apply tests also pass. |
| Readable drifting Thread during graph repair | Shipped `TestGraphRepairReceiptRetainsReadableThreadIdentityDrift` passes in dry and committed modes: task graph becomes healthy, Thread stays broken, canonical impact ID/location survive, and Thread bytes stay identical. The actual emitted `tskflwctl task depend repair --auto` was run in both modes against a duplicate-dependency state before any declaration normalization. |
| Duplicate Threads during repair | Independent store regression probe fails in both modes on the frozen head **and base `2d4d9f1`**. Fresh CLI reproduces two healthy receipt impacts with frontier length 1 and no problems, while the post-repair list reports both broken with `duplicate-thread-id`. Thread bytes remain unchanged. See `M1`. |
| Presentation hint versus write handle | Shipped materializer test accepts an opaque location with a separately supplied valid handle and rejects a hinted filesystem-shaped location without a handle; missing/stale revisions and contradictory IDs/handles reject. |
| Physical rename/delete/body/unreadable races | `TestReviewThreadMaterializerPhysicalRaces` changes files after the captured read, with changed and no-op materialization. Rename/body/unreadability yield conflict; deletion yields not-found. No fresh revision is minted to authorize changed bytes. |
| Opaque adapter simulation | `TestReviewOpaqueCompleteSnapshot` uses an opaque key/location, empty local handle, and independent revision strings. Source validation and unchanged complete CAS work; changed/missing revisions, readable-to-unreadable and changed unreadable revisions fail closed. A service without the path port cannot manufacture a path. This is a test simulation, not a shipped database adapter. |
| Callback-owned nested values and body maps | Shipped dispatcher tests cover mutation, creation and apply nested tags/tasks. An independent apply callback attempts to replace an actually changed body with the planned body to claim idempotent convergence: both modes reject. Removing the body-map clone yields false `Complete=true`. |

CLI fixture artifacts: `cli-matrix-summary.json`, per-command `*.stdout`/`*.stderr`, `duplicate-cli-summary.log`, `duplicate-repair-True.stdout` and `duplicate-repair-False.stdout`. Store fixture source/logs: `review_independent_test.go`, `independent-passing.log`, `independent-tests.log`, `base-duplicate.log`. Fixture initialization uses an explicit `init --path <fixture> --no-register`.

### Restored compiling mutation evidence

Every accepted mutant below compiled, ran the named behavioral assertion, exited 1, and was restored before the next probe. There were **18 compiling kills** and no green mutant left unexplained. Setup compile errors were corrected and are not counted as kills. Exact transformations and complete output are retained in `run_mutations.py`, `additional_mutations.py`, `final_mutations.py`, and `mutation-*.log` under `.git/review-evidence/`.

| Mutant | Exact protection changed | Focused regression and failure |
| --- | --- | --- |
| `path-hint-authority` | Add a `LocationIsPath` fallback into `source.LocalPath` before `internal/store/threadmutation.go:155` | `TestThreadMutationMaterializerRequiresExplicitSourceEvidence/path_hint_without_handle`: materialization erroneously succeeds. |
| `handle-cas` | Remove local-handle equality at `internal/store/cas.go:85` | `TestThreadSourceSnapshotRetainsReadableHandleAndRevision`: removed/changed handles erroneously compare equal. |
| `fresh-revision` | Replace the original revision with the current hash before `internal/store/threadmutation.go:169` | Materializer `stale_revision` case erroneously succeeds. |
| `repair-source-identity`, `impact-core-source` | Coordinately replace both loaded impact projections with bare `ProjectThread` and use declared impact ID at `internal/core/thread_mutation.go:443`, `:455` | Store drift-repair regression and core `TestThreadGraphImpactsRetainSourceDefectsAfterRepair` fail: drift/location disappear and healthy frontier is fabricated. |
| `callback-snapshot` | Remove all three dispatcher Thread copies and the apply body-map copy (`internal/store/threadmutation.go:121`, `threadcreation.go:113`, `threadapply.go:201`) | Terminal authorization and nested dispatcher tests fail; callback changes owner state/body. |
| `nested-thread-clone` | Remove nested Tags/Tasks copies in `internal/store/threadcreation.go:182` | Dispatcher nested-value tests fail on all three routes. |
| `loaded-thread-clone` | Return the original Thread value at `internal/core/store.go:147` | `TestLoadedThreadsOwnProjectionValuesWithoutGuardedEvidence` detects aliasing of guarded slices. |
| `source-validator` | Bypass `ThreadRead.ValidateSources` at `internal/core/store.go:169` | `TestThreadMutationSourceValidatorsRetainIndependentIdentity` fails; missing/drifting independent source evidence can pass or be misattributed. |
| `complete-source-cas` | Remove independent source comparison and declaration/slug comparison at `internal/store/cas.go:84` | Readable snapshot changed-source-ID/location cases erroneously pass CAS. |
| `apply-body-fabrication` | Remove only dispatcher body-map clone at `internal/store/threadapply.go:202` | `TestReviewApplyCallbackCannotFabricateBodyConvergence`: changed body falsely reports complete, unchanged convergence. |
| `all-entry-source-gates` | Remove pre-planner calls in creation, mutation, task lifecycle and initial apply | `TestReviewAllSourceGateRoutes`: malformed sources reach callbacks. Restoring only a nearby semantic validator would not prove this gate. |
| `final-apply-source-gate` | Remove the source gate at `internal/store/threadapply.go:271` while retaining pure apply validation | `TestReviewFinalApplySourceGate`: drifting unrelated source permits target creation and a complete receipt. |
| `opaque-missing-revision` | Remove nonempty revision requirements from readable/unreadable equality in `internal/store/cas.go:88`, `:100` | `TestReviewOpaqueCompleteSnapshot`: two unversioned reads incorrectly pass. |
| `callback-actual-commit` | Remove mutation dispatcher copy; separately run terminal test with real-write iteration first | `TestThreadPlannerCannotRewriteTerminalLifecycleAuthorization`: cancelled Thread is actually committed as in-progress. This confirms a write failure, not only a misleading dry-run receipt. |
| `semantic-entity-fields`, `canonical-id-fallback` | Reintroduce source fields on all five semantic structs; separately restore a Thread `CanonicalID()` method | `TestSemanticEntitiesDoNotCarrySourceEvidence` fails for every field-bearing entity and for the restored method. |
| `coordinated-duplicate-id-gates` | Remove duplicate-source rejection and the nearby duplicate-declaration check (`internal/core/store.go:179`, `internal/core/thread_creation.go:109`) | `TestThreadIdentityDefectsRemainReadableButBlockGuardedPlanners/duplicate_source` fails because the duplicate set reaches a callback. One semantic duplicate check cannot accidentally preserve the invariant in this probe. |

### Second pass: systemic challenges

The adversarial second pass re-traced shared abstractions and production callers after the checklist tests. It found `M1`: `LoadedThreads` preserves every occurrence, but the downstream per-record impact loop assumes each single record is independently healthy. Duplicate qualification exists only in the list route. This is a demonstrated whole-set failure rather than a speculative adapter concern.

Other challenged patterns were settled with hostile evidence:

- Source-erasing semantic constructors cannot establish guarded identity. Every production Thread semantic conversion has a prior source gate. The coordinated initial-gate mutant and the final-convergence mutant show that pure validation alone is insufficient, and that the actual source checks are consequential.
- Callback authorization inputs are copied twice across the owner/dispatcher boundary; nested slices and apply maps are independently owned. Both terminal revival and fabricated idempotent body convergence were killed, including an actual committed terminal-revival mutant. Thread callback values exclude guarded revisions/local handles. The graph deliberately retains source-query APIs for repair: `TaskSource` (`internal/core/dependency_graph.go:871`) returns a source value, while `SourceRecords` (`internal/core/dependency_source.go:90`) returns copied source fields without private revision tokens; neither lets the callback rewrite owner evidence.
- Original-read revisions, handles and the whole readable/unreadable set remain separate evidence. Path-shaped context does not authorize a write. Physical body/representation races and opaque revisions reject; a fresh-token mutant was killed.
- The future opaque adapter fixture needs no filename or local handle to preserve source identity or qualify its complete snapshot. It does need an adapter-observed nonempty revision for every occurrence. No real database adapter was claimed or tested. FS-only materialization correctly requires FS authority.
- TUI selection keys, stale generation checks and the second path-port lookup survive this delta and their focused tests pass. A raw external editor can still race after the last pathname lookup; this is the documented best-effort navigation seam, not the guarded writer CAS. No safety claim is made about an uncooperative editor after the final check.

### Compatibility and validation results

Seven base/head comparisons were byte-identical on the same valid fixture and command environment: human Thread list/show, JSON Thread show/graph, human/JSON Task show, and JSON `status --all`. No persisted Markdown rewrite was required. Guarded operations preserve timestamps, comments/body and idempotent bytes in the migrated store tests. Thread wire values deliberately retain declared IDs; source drift is reported through health/problems, consistent with the prior owner disposition. This does not make a declared ID a repair selector.

`TestReviewNonDefaultPortableSchemaBranches` validated seven actual non-default DTO/envelope instances with the semantic JSON Schema validator: Task/Epic/Audit/Research opaque locations, unreadable Thread location plus a private revision, bounded graph scope with nonzero counts/boundary edges, and a pathless raw repair declaration plus incomplete Thread evidence. All pass. Private revision/handle field names and the private token are absent from those public instances. Public operation-path receipt fields remain intentional metadata supplied by the operation; their existence is not a leaked semantic entity field.

| Command/check | Result |
| --- | --- |
| Fresh `go build -o .git/review-cli ./cmd/tskflwctl` and base binary build inside archived base sources | Passed; no installed binary used. |
| `go test ./...` | Passed; also rerun after mutation restoration. |
| `go test -race ./internal/core ./internal/store ./internal/tui ./internal/wire` | Passed for all four packages. |
| `just lint` | Passed, `0 issues.` |
| `.git/review-cli -C . lint` | Passed: all planning entities/dependency links pass lint. |
| `go run ./internal/tools/schemacomments`; generated-file diff | Passed; wrote 267 comments and no baseline diff. The sole comment-map removal in the target delta is the removed Task.Path comment. |
| Restored focused source/authority/callback/TUI/identity tests | Passed. |
| Independent route/race/body/opaque probes; optional schema probe | Passed. Duplicate repair assertion intentionally fails on unchanged head/base and is entered as `M1`. |
| `git diff --check` and final guarded workspace verification | Passed before transfer. |

Go and lint caches were redirected to `.git/review-cache/` after the default cache location was denied by the filesystem sandbox; those initial setup failures are not counted as test results. Build VCS-stat-cache warnings did not prevent fresh binary generation. Evidence and caches are sandbox-only, with no staged changes or additional reviewer commits. No requested technical check was omitted. Linux/Windows platform locking and a real remote persistence adapter were not exercised; this is a local macOS review with opaque-adapter simulations. Static code and tests are evidence of the stated contract, not proof of every possible external race.

## Findings

#### M1. Graph repair receipts lose duplicate Thread source defects · **Status:** fixed (PR #274)

**Classification:** Medium; pre-existing defect exposed by this migration, independently reproduced on `2d4d9f1` and the frozen head. It violates the final repair-evidence contract; it is not a regression introduced by the three implementation commits.

**Cause and route:** `internal/store/graphrepair.go:57` correctly retains all readable source envelopes. `internal/core/dependency_repair.go:675` passes them to `TaskGraphThreadImpacts`, whose loop at `internal/core/thread_mutation.go:439` projects each record independently (`:443`) and publishes it (`:454`) without whole-set duplicate qualification. The list route alone calls `markDuplicateThreadIDs` at `internal/core/service_thread.go:248`; its implementation at `:292` breaks projection health, clears the frontier and marks completed inconsistency. A single-record projection cannot know a second source occurrence exists.

**Reproduction:** Create a completed prerequisite `6g0000000003`, a next-up member `6g0000000001` with duplicate `depends_on: [6g0000000003, 6g0000000003]`, and two otherwise-valid Thread documents `threads/6g0000000002-alpha.md` and `threads/6g0000000002-beta.md`, both declaring `id: 6g0000000002` and `tasks: [6g0000000001]`. Run `thread list --json`, then `task depend repair --auto --dry-run --json`, then the committed equivalent and list again. Fresh-binary outputs are retained under `.git/review-evidence/duplicate-repair-*.stdout`; independent Go assertion is `TestReviewDuplicateThreadRepairImpacts` in the retained scratch source.

**Observed:** Both repair modes produce a healthy task graph and two Thread impacts with `after.projection_health="healthy"`, frontier length 1, and `problems=[]`. Both Thread files remain byte-identical. The ordinary post-repair list still marks both sources `broken`, with `duplicate-thread-id` and no frontier. Source occurrences and IDs are retained, but their unresolved duplicate defect disappears from the repair receipt and makes ambiguous evidence appear actionable.

**Required change:** Qualify both complete before/after Thread projection sets for duplicate source identity before filtering or publishing repair impacts. Reuse the same duplicate-health/frontier/completed-inconsistency rules as list, retain each occurrence's source ID/location, and do not rewrite the Thread documents. Add permanent dry-run and committed receipt coverage with a repairable task graph plus duplicate Thread sources, asserting broken impact health, suppressed frontier, per-source duplicate attribution, unchanged Thread bytes, and consistency with ordinary list. Leave this finding open for owner scope/severity triage.

**Resolution:** Qualified the complete before/after readable Thread source sets
before filtering repair/lifecycle impacts, sharing duplicate
health/frontier/completed-inconsistency rules with list. Permanent core and
dry-run/committed store regressions reproduce the defect, retain each source
occurrence and diagnostic location, and verify unchanged Thread bytes.

### Sandbox and transfer attestation

The general helper created the independent `--no-hardlinks` clone and its sole reviewer baseline commit. All implementation inspection, builds, tests, generation, fixtures, mutations and audit editing occurred there. Source-checkout protocol deviation: the initial concurrent brief read also ran read-only `git status`/top-level metadata and filename/AGENTS discovery before the brief's literal restriction was known. No implementation content was inspected, test/project mutation run, or source file written in those initial operations. Only the guarded helper performs the final assigned-file transfer.

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Qf3tqq
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Qf3tqq/.git
baseline_commit=0a447a9979831e8f2a7fa83e98db7b3cf42f3389
source_blob=54d85d06637894e3f304023b495ba6946491f4da
source_fingerprint=6764a217ac43e540a9267831552543b674225538
deliverable=planning/audits/6gg38rnpe94c-2026-10-03-final-semantic-entity-source-boundary-checkpoint-codex.md
deliverable_changed=true
transfer=succeeded
```

Delivery is contingent on the helper returning `transfer=succeeded`; its actual verification/transfer transcript is retained in `.git/review-evidence/verify-attestation.txt` and `transfer-attestation.txt`. If that command refuses, this draft is not delivered and the transfer line must be treated as pending/refused. The workspace is retained until implementation-owner receipt. Only this assigned audit is transferred; findings remain open and no implementation fix, task-state change, push, staging or additional commit is included.

## Owner reconciliation — 2026-10-03

Accepted M1 as an in-scope repair-evidence defect, not a regression from field removal. Permanent
store coverage reproduced missing duplicate diagnostics in both dry-run and committed repair before
the fix. List, graph-repair impacts, and lifecycle impacts now use one collection projector which
qualifies the complete readable source set before filtering changed projections. Every occurrence
keeps its source ID/location; duplicate sources remain broken with no frontier, and completed
duplicates remain inconsistent. No Thread document is rewritten.

`TestGraphRepairReceiptRetainsDuplicateThreadSources` covers both repair modes, per-file attribution,
unchanged Thread bytes, and qualification consistent with ordinary list. The core regression
`TestThreadImpactsQualifyDuplicateSourcesBeforeFiltering` also covers opaque locations, an unchanged
duplicate omitted from the receipt, completed consistency, and unchanged adapter record order.
Both tests failed on the frozen implementation and pass after the fix. Full tests, affected-package
race checks, and lint pass. No outstanding finding or narrower completion exception remains.

Owner confirmed the recorded independent Git directory/baseline, a sole assigned-audit sandbox
delta, byte-identical delivery, and the helper's retained `transfer=succeeded` transcript. Reviewer
scratch evidence remains in the sandbox; only the permanent regression tests ship. The original
review verdict above describes the frozen implementation, not this reconciled follow-up.
