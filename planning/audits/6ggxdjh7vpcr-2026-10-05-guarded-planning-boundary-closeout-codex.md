---
schema: 1
id: 6ggxdjh7vpcr
bucket: closed
area: guarded-planning-boundary-closeout-codex
date: "2026-10-05"
---
# Audit: Guarded planning boundary contracts and Thread closeout — codex — 2026-10-05

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

Adversarial implementation review and bounded pre-closeout check for the final task in
`make-planning-data-access-adapter-neutral`. Challenge whether these regressions actually
protect the claimed boundaries. A green suite, an inventory of test names, or repeating the
owner's closeout table is not sufficient evidence. Prefer demonstrated gaps over speculative
redesigns, and distinguish safety protection from useful early diagnostics.

## Review target

- Task: `planning/tasks/6ggdkzv2tnta-pin-guarded-planning-mutation-boundary-contracts.md`.
- Thread: `planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md`.
- Source findings: M1-M3 in `planning/audits/6gfrcytd9n9a-2026-10-02-test-rigour.md`.
- Review the complete handoff diff from main, including five changed test files, merge
  bookkeeping, and the tightened existing YAML-fidelity task. No production change is
  intended. Verify that statement in your captured sandbox.
- Prerequisites merged in #279 (ordinary/workspace identity), #280 (unreadable cross-kind
  collisions), #282 (core-owned impact/recovery), and #283 (explicit persistence policy).
  Read their task-local evidence where needed; do not re-audit all historical changes.
- Build a concise **consumer inventory**: pure compose/prepare, core service creation/apply,
  filesystem mutation implementation, CLI explicit-space presentation, source-set composition,
  and the runtime composition/controller boundary. Locate actual symbols before citing them.

## Intended contract to challenge

1. A malformed manifest or edited/stale durable plan refuses with the appropriate sentinel
   and actionable diagnostic. Preview and refused apply change no planning files. The typed
   failure retains the attempted plan, even when planning stops before the adapter echoes it.
2. A post-commit Thread creation error wrapping `ErrConflict` cannot trigger retry/remint or
   erase the original committed receipt. Generic failures remain valid controls. The actual
   persisted Thread remains readable, with no extra artifact.
3. Explicit bad `--space` refuses even for best-effort theme commands; ambient discovery
   behavior is not being removed. Error classification and exit 10 remain pinned.
4. Source identities/versions remain outside semantic values; optional local paths do not
   become required reads. Split capabilities must address one witnessed corpus. Ordinary
   and workspace paths preserve fresh identity and authorization, including previews.
5. Controllers use injected ports rather than selecting concrete persistence. Exceptions
   stay narrowly named and recursively lint-enforced.
6. Closeout claims are honest: 18 merged/completed members, one local in-flight task, healthy
   graph/projection, and reviewed versus merely tracked versus released work distinguished.

Non-goals: a general conformance framework, second backend, full repeated workspace matrix,
schema/retry redesign, TUI polish, source-audit L1-L3, or fixing the existing folded-YAML
storage bug. H1 from `2026-10-05-arch-data-model-and-storage` was reproduced and handed to
existing task `6g1dhhk6721x`, now high priority; it is not fixed or hidden by this closeout.
Assess whether any claimed migration blocker was incorrectly dismissed, but do not manufacture
a duplicate finding for an honestly tracked pre-existing defect.

## Mandatory evidence floor

- Work only in the mandated independent clone, including temporary test/probe files. Capture
  the exact baseline and inspect the actual net diff. Build the sandbox's own binary for any
  planning or product commands; never use a stale installed binary as implementation evidence.
- Execute the new focused tests before mutating anything. Establish valid fixture/preview
  controls before each hostile case. Record the actual failing assertion and semantic result,
  not only a command exit status.
- Execute these compiler-valid mutations, individually and restore each one:
  - Remove the duplicate local-key refusal in `ComposeThreadApplyPlan`.
  - Remove the missing prerequisite/dependent guards in `PrepareThreadApply` together;
    require the corresponding pure AND real-store rows to fail for the lost diagnostics.
    Explain that later validation can still block writes; do not claim corruption unless reproduced.
  - Remove only `&& !result.Committed` in `runThreadCreationMutation`. Run the generic and
    conflict cases through both core fake and real service/store. Check the original identity,
    call/release/retry counters, typed receipt, and durable document rather than merely error text.
  - Suppress theme's explicit-selection refusal. Require both theme subtests to fail on ambient
    success; template is not a substitute for reaching theme's own branch.
- Add at least one hostile experiment not copied from the owner's list. Suitable targets:
  absent versus explicitly empty authoritative body; stale body/status and converged no-op;
  a missing endpoint after real composition; a forged receipt/plan; a changed guard on a late
  workspace open. State which invariant it tests and whether existing coverage catches it.
- Independently assess the full-tree no-effects oracle: fixture edits are outside the measured
  action, entries/bytes/modes are captured, earlier failures cannot masquerade as target coverage,
  and the test does not claim transactional or root/permission guarantees it cannot demonstrate.
- Run the focused suite with `-race -count=5`, full `go test -race -timeout 180s ./...`, standard
  lint, planning/audit lint, and generated CLI/schema-comment comparisons in disposable outputs.
  Record any unavailable command as a limitation. No new resource-performance behavior is claimed;
  avoid inventing benchmark requirements for this test-only slice.
- Verify the live sandbox Thread rollup, external gates, and completed task ACs. Check that
  review findings with followups have actual task destinations, not merely reassuring prose.

## Required hostile angles

After the checklist, perform a separate systemic pass: could a shared test helper or redundant
validator make many tests green while the contract is broken? Could the receipt lose truth
after commit? Could independently injected capabilities address different corpora? Could a
preflight/preview silently execute user code or persist a no-op? Could a primary adapter open
storage through a new nested package while lint appears protective? Investigate at least two
with executed evidence, using existing foundation tests rather than cloning the whole matrix.

Reviewer emphasis (both still review the same contract):

- **Codex:** prioritize coordinated mutations and semantic recovery/composition gaps. Challenge
  whether the witness/policy/core-owned intent is load-bearing across the actual caller boundary.
- **Antigravity:** prioritize branch reachability and fixture/oracle skepticism. Provide a compact
  claim / hostile input / observed result / protecting assertion matrix. For any ready verdict,
  execute both real-store committed-conflict and malformed/stale-plan cases, not just fakes.
  Try to falsify the owner's clean conclusion; if it survives, report the failed attack and why.
  Do not substitute unverified symbol inventories or planned tests for implemented evidence.

Consider assertions too brittle or incomplete, loss of existing coverage during test rewrites,
global-hook cleanup across subtests, stable-ID minting/collision behavior, empty versus missing
values, and whether the whole-corpus closeout overstates the bounded evidence. A diagnostic
guard survivor can be a test gap without being a production data-safety defect.

## Validation and restoration

Use the captured sandbox and its own baseline. The helper's baseline commit is permitted;
no other commit, push, source checkout mutation, or planning lifecycle changes in the source.
Do not parallelize mutation probes in one sandbox. Do not use destructive cleanup on shared
paths. Restore only your sandbox-local probes, rerun their controls, and verify an empty
implementation diff before completing the one-file audit transfer.

```sh
go test ./internal/core ./internal/store ./internal/cli -run 'TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest|TestComposeThreadApplyPlanRequiresValidClockAndIDGenerator|TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity|TestServiceThreadCommittedFailureIsNotRetried|TestThreadCreationAttributesReleaseFailureAfterCommit|TestThreadApplyMalformedAndStalePlansDoNotPersist|TestGlobalSpace_UnknownListsKnownLabels' -count=1
go test -race ./internal/core ./internal/store ./internal/cli -run 'TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest|TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity|TestServiceThreadCommittedFailureIsNotRetried|TestThreadCreationAttributesReleaseFailureAfterCommit|TestThreadApplyMalformedAndStalePlansDoNotPersist|TestGlobalSpace_UnknownListsKnownLabels' -count=5
go test -race -timeout 180s ./...
just lint
just build
./bin/tskflwctl -C . lint
./bin/tskflwctl -C . audit lint
./bin/tskflwctl -C . thread frontier make-planning-data-access-adapter-neutral
```

Generate docgen and schemacomments output under a disposable directory and compare with
`docs/cli` and `internal/wire/schema_comments.json`. Prefer existing cache locations if your
environment needs them. A compiler failure, unavailable tool, or unrelated invalid fixture
does not count as a killed mutation.

## Deliverable

Update only your assigned audit. Use the repo's finding verbs for real findings and preserve
the exact finding grammar. Each finding needs reproduction, actual/expected behavior, impact,
minimum fix, and in-scope versus followup treatment. No quota; do not mark findings settled
before owner triage. Replace the report placeholder with isolation attestation, executed command
results, mutation/hostile-case evidence, limitations, and one verdict: ready, ready with tracked
followups, or not ready. Explicitly answer whether anything blocks closure of this bounded Thread.
Transfer only after the isolated-workspace helper verifies restored implementation and the
unchanged source audit. If the source changed, retain the sandbox and report the conflict.

## Reviewer report

**Verdict: ready with tracked followups. Nothing demonstrated blocks closure of this bounded adapter-neutral Thread once the owner reconciles this final review.** Both independent review passes are complete. No new finding is asserted and no finding disposition, implementation fix, task completion, or Thread closure was applied. The local test slice is reviewed here; merge and release remain separate milestones. The existing folded-YAML defect is reproduced, remains unfixed and high priority, and should retain its pre-release storage followup.

### Captured scope and verified consumers

`git diff origin/main --name-only` contains **14 files: five test files and nine planning files**, including the two review briefs, merge bookkeeping and the tightened YAML-fidelity task. No production Go, lint configuration, machine contract, or generated document changed. The complete captured diff is retained as `handoff.diff`. Main is at merged PR #283; the independent helper checkpoint, rather than source HEAD, is the restoration baseline.

| Consumer | Actual implementation and bounded evidence |
| --- | --- |
| Pure materialization | `internal/core/thread_apply.go:190` `ComposeThreadApplyPlan` checks manifest/clock/generator, resolves exact IDs and invokes `PrepareThreadApply`. Duplicate local-key refusal is at `:249`. Generation is an explicit input callback; “pure” does not imply that a supplied ID generator is never called. |
| Pure current-state validation | `internal/core/thread_apply.go:322` `PrepareThreadApply` clones/normalizes the durable plan, checks identity/lifecycle/edges and plans additive writes. Missing prerequisite/dependent guards are at `:381`, `:385`; authoritative-body presence is at `:429`. Early endpoint/lifecycle checks improve diagnostics; later validators also enforce safe writes. |
| Core creation recovery | `internal/core/service_thread.go:26` `NewThread` mints once before its planner closure; `:86` `runThreadCreationMutation` retries conflict only before `result.Committed`. A committed failure retains identity/local outcome in `ThreadCreationMutationFailure`. Removing the guard does not remint in this implementation, but it does retry and lose the real committed receipt. |
| Core apply recovery | `internal/core/service_thread_apply.go:12` `ComposeThreadApply` reads injected graph/Thread ports; `:48` `ApplyThreadPlan` performs guarded preparation, retains the caller plan when the adapter cannot echo it, and surfaces typed `ThreadApplyFailure`. The fallback is at `:66`. |
| Filesystem mutation | `internal/store/threadcreation.go:17` `MutateThreadCreation` authorizes, locks, validates, materializes, verifies source snapshots and records commit before release. `internal/store/threadapply.go:18` `MutateThreadApply` reads fresh identity and authoritative graph/Thread/body evidence, calls the planner and re-prepares before persistence. Dependencies are written first, Thread last; it is resumable per-file persistence, not a new transaction guarantee. |
| Explicit-space presentation | `internal/cli/theme.go:21` `newThemeCmd`, `:44` best-effort resolve branch; `internal/cli/root.go:422` `wantsSpace`, `:446` `resolve`. An explicit address failure is fatal even where ambient discovery is optional. Both theme list and deterministic preview reach this branch; template's sibling check is not their substitute. |
| Source-set composition | `internal/core/service.go:321` `NewService`, `:387` constructor refusal and `:404` capability inventory delegate to `internal/core/source_set.go:41` `validateSourceSet`. Nonzero/stable/matching witnesses bind split ports before use; they are wiring evidence, not semantic record IDs or security credentials. `internal/store/fsstore.go:67` exposes the instance witness. |
| Runtime opening/composition | `internal/appwiring/wiring.go:26` `compose` selects guarded policy, `:55` `openPlanning` publishes neutral ports after one discovery. `internal/workspacestore/planning.go:17` `NewPlanningStore` captures the observed root/marker reader and policy; workspace opening uses the same helper. `internal/core/workspace.go:78` `WorkspaceService.Open` assembles the injected capabilities through `NewService`. Existing ordinary/direct/pointer foundation tests were rerun, rather than copied into a new workspace matrix. |
| Controller enforcement | `.golangci.yml` `cli-controllers-use-injected-application` matches direct and recursive CLI Go files and uses exact allowed internal packages. `cli-local-topology-and-checkout` separately names only `init.go`/`workspace.go`, allowing config topology work, not persistence selection. `cli-composition-contracts-stay-neutral` constrains ports. A new nested CLI package importing `store` compiles and is rejected by the actual recursive rule. |
| Semantic/local evidence | `internal/domain/identity_test.go:8` `TestSemanticEntitiesDoNotCarrySourceEvidence`; `internal/core/source_set_test.go:167` `TestNewServiceAcceptsPathlessReadOnlySource`; `internal/core/service_entity_path_test.go:121` `TestEntityPathCapabilitiesIgnoreTypedNilAndRejectForeignSourceSets`. These implemented foundation tests pass. Optional navigation remains optional; the audit does not infer path requirements from a filesystem fixture. |

### First-pass claim / hostile input / result matrix

The following evidence is executed, not inferred from owner checkmarks. Files named without a directory live in `.git/isolated-review-workspace/evidence/` in the retained sandbox.

| Claim | Hostile input and observed result | Protecting assertion |
| --- | --- | --- |
| Manifest guards are reachable | Owner compose rows include distinct task IDs sharing a trimmed local key, valid title/metadata and valid clock/generator. Removing the duplicate-key branch returns success. Removing the zero-clock guard produces a plan dated `0001-01-01` and calls the generator once. | `TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest` (`internal/core/thread_apply_test.go:115`) fails `/duplicate_local_key`; `TestComposeThreadApplyPlanRequiresValidClockAndIDGenerator` (`:204`) fails `/missing_clock`, while its other controls pass. |
| Missing endpoints retain useful diagnostics | Pure and actual service/store cases start from a valid plan/preview, then remove either endpoint from the graph/filesystem. Coordinated removal of both early guards changes errors to later “planned dependency … does not exist” / “planned task … does not exist in the authoritative snapshot”. | `TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity` (`internal/core/thread_apply_test.go:320`) and `TestThreadApplyMalformedAndStalePlansDoNotPersist` (`internal/store/threadapply_test.go:49`) fail both endpoint rows, including preview/commit. Later validation still refuses; no corruption is claimed. |
| Malformed/stale apply has no effects and keeps intent | The real service/store test first proves valid preview, then exercises 11 independent schema/status/slug/date/edge/endpoint/identity/body cases in both modes. Intentional removals/body edits happen before the measured-action snapshot. | Sentinel plus useful diagnostic, `ThreadApplyFailure`, exact attempted plan and typed receipt equality, no changed/committed/complete flags, and equal tree snapshots. Removing caller-plan fallback makes refusal receipts lose their token. |
| Committed conflict cannot be retried | Both generic and conflict-wrapping release errors run through the core fake and real service/store. Original code: one mint, one adapter/release, zero retries, same committed identity/local outcome, typed receipt, one readable Thread with its exact body. | `TestServiceThreadCommittedFailureIsNotRetried` (`internal/core/service_thread_test.go:188`) and `TestThreadCreationAttributesReleaseFailureAfterCommit` (`internal/store/threadcreation_test.go:289`). Removing **only** `&& !result.Committed` leaves generic controls green, but fake conflict calls become **5**, retry sleeps **4**; real conflict release calls become **4**, sleeps **3**, and the returned receipt loses committed truth. |
| Durable evidence remains observable even on a failing mutant | Reviewer `TestAuditCreationFailureDoesNotRemint` uses the real guarded store, an observable ID generator, and nonfatal receipt assertions so it continues to read the document after a failed assertion. | Mutated conflict path: first ID `sym9ma36n7ff` is still readable, artifact count remains **1**, mint count **1**, but returned ID is empty and typed committed failure disappears. This proves receipt loss, not an invented duplicate-file/remint defect (`creation-retries-after-commit.log`). Restored generic/conflict cases pass. |
| Theme's own explicit selection cannot become ambient success | Remove only theme's explicit-selection refusal. Unknown `--space` produces the ambient list and palette output; template remains protected. | `TestGlobalSpace_UnknownListsKnownLabels` (`internal/cli/space_selection_test.go:140`) fails both `/theme/list` and `/theme/preview/--variant/dark`, requiring `ErrNotFound`, exit **10**, diagnostic and empty stdout. Fresh binary separately verifies ambient successes and explicit JSON refusals (`product.json`). |
| Empty body is evidence; missing body is not | Reviewer `TestAuditEmptyBodyConvergenceAndStaleStatus` applies a valid plan whose body is explicitly empty to real storage, reads the authoritative empty body, and retries preview/commit after convergence. It also deletes the body-map entry in a pure snapshot. | Present empty body converges, returns complete/no change/no commit and preserves the tree; absent entry returns validation and no decision. Suppressing the presence guard makes the absent-empty case succeed. Existing owner missing-body row also fails classification. |
| Current lifecycle changes cannot be overwritten by an old plan | The reviewer fixture first creates and converges the empty-body Thread, then uses real `StartThread` as a valid state change before measuring old-plan apply. | Old-plan preview/commit return conflict with typed retained plan and equal post-start tree. The corresponding pure advanced-status control also conflicts (`independent-final.log`). |
| Split ports and late policy are load-bearing | Remove source-set mismatch refusal; independently opened real workspaces can now be cross-wired into a service. Separately substitute unrestricted policy in the shared late-opening helper. | `TestFS_MultiWorkspaceSourceSetsCannotBeCrossWired` (`internal/workspacestore/fs_test.go:89`) fails on a nonnil service/nil error. `TestNewServiceRejectsEveryMismatchedSplitCapability` (`internal/core/source_set_test.go:61`) fails every injected capability row. Existing real ordinary/workspace authorization tests fail direct/pointer and preview/commit cases under unrestricted substitution. |
| Repair instructions are executable evidence | Fresh binary runs a valid composition/preview, removes the governing marker's ID, executes the recommended `config migrate`, then checks the old plan conflicts with the minted identity. A separate markerless apply recommends migrate; that command actually exits **11** and recommends init. | Explicit same-target `init --path … --taskflow-root planning --no-register` succeeds; old plan still conflicts (exit **14**); recomposition applies and a converged repeat writes nothing. The markerless detour remains the already-tracked recovery followup. `space add` from the unknown-space hint is also executed against the isolated registry (`product.json`, **18** commands with asserted exits/tree effects). |

### Executed mutations and restoration

**13 compiler-valid behavior mutations, all killed semantically, plus one compiler-valid nested-import lint probe.** No compiler failure is counted. `mutations.py` retains exact replacements and commands; `mutations.json` records names, failing tests and restored exits. Each behavior probe has `<name>.log` and `<name>-restored.log`; all **13** immediate restored runs exit zero. The nested package was removed and standard lint rerun successfully.

| Probe | Exact claimed regression reached | Observed failure |
| --- | --- | --- |
| `compose-duplicate-key` | `TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest` | Duplicate local key returns nil instead of validation. |
| `prepare-both-endpoints` | `TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity` **and** `TestThreadApplyMalformedAndStalePlansDoNotPersist` | Both pure endpoint rows and four real preview/commit rows lose the endpoint-specific messages; later validators still refuse. |
| `creation-retries-after-commit` | `TestServiceThreadCommittedFailureIsNotRetried`, `TestThreadCreationAttributesReleaseFailureAfterCommit`, reviewer `TestAuditCreationFailureDoesNotRemint` | Conflict controls retry; real receipt loses commit/identity while original document remains. Generic controls pass. |
| `theme-explicit-selection-suppressed` | `TestGlobalSpace_UnknownListsKnownLabels` | Both theme branches emit ambient success; template control passes. |
| `authoritative-body-presence` | `TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity`, reviewer empty-body test | Missing evidence becomes different-body conflict for a nonempty plan, and successful convergence for an empty plan. |
| `planned-thread-status` | Pure unsafe-plan and real malformed/stale-plan tests | Planned-status-specific diagnostic disappears; later creation validation still refuses. |
| `caller-plan-recovery-token` | `TestThreadApplyMalformedAndStalePlansDoNotPersist` | Typed failure no longer retains caller plan on early refusal. |
| `split-corpus-witness` | `TestNewServiceRejectsEveryMismatchedSplitCapability` | Constructor publishes mismatched split capabilities. |
| `preview-persistence-oracle` | `TestThreadApplyMalformedAndStalePlansDoNotPersist` | Injected preview artifact is caught by the **valid-preview** tree comparison, before hostile edits. This tests the oracle and is not misreported as reaching an endpoint refusal. |
| `snapshot-modes-oracle` | `TestSnapshotTreeCapturesBytesEmptyDirectoriesAndSymlinkTargets` (`internal/testutil/snapshot_test.go:10`) | Dropping permission bits from captured modes fails the changed-mode assertion. |
| `real-workspace-source-witness` | `TestFS_MultiWorkspaceSourceSetsCannotBeCrossWired` | Cross-wired real workspace service is nonnil instead of refused. |
| `late-policy-propagation` | `TestPlanningOpenersPreserveThreadApplyAuthorization` (`internal/appwiring/thread_apply_test.go:177`) and `TestWorkspaceOpeningCarriesPolicyAcrossDirectAndPointerEntryPoints` (`internal/workspacestore/fs_test.go:144`) | Real late mutations are allowed/skipping the expected guard after open. |
| `compose-clock-required` | `TestComposeThreadApplyPlanRequiresValidClockAndIDGenerator` | Missing clock returns a materialized plan and invokes ID generator; other rows remain green. |
| Nested lint probe | `internal/cli/reviewnested/audit_probe.go` importing actual `store.FS`; `go test ./internal/cli/reviewnested -run '^$'` then `just lint` | Compilation exits **0**; lint exits **1** solely with the recursive `cli-controllers-use-injected-application` depguard diagnostic (`nested-compile.log`, `nested-lint.log`). |

The coordinated endpoint mutation removes **both** early checks, not merely one backstopped branch. The two preparation passes share the altered validator, and downstream graph/creation validation still protects persistence. The audit distinguishes a caught diagnostic regression from data corruption and does not treat redundant safe refusal as a missing safety barrier.

### Independent oracle assessment and systemic second pass

`internal/testutil/snapshot.go:20` `SnapshotTree` records relative entries including root/empty directories, regular-file bytes, full file modes, and symlink targets without following them; unsupported special files or read errors fail the test. It does not record owner, inode, timestamps, extended attributes or writes outside the selected root, and makes no concurrent/transactional snapshot claim. In the new real-store test, valid compose/preview is measured first; endpoint removals and the changed-body setup are completed before the refusal snapshot. Each hostile case therefore reaches its intended assertion without an unrelated invalid baseline masquerading as coverage.

Two systemic assumptions were actively challenged beyond the mandatory owner list:

- **A shared helper could make all no-effects tests green.** Injecting an actual preview file makes the new real-store test fail its valid-control snapshot; dropping captured permission bits makes the helper's independent expected-value test fail. Reviewer converged empty-body requests compare the full tree in both modes. These settle entries/bytes/modes protection, not crash atomicity, root-container permissions or concurrent edits.
- **Independent ports could address different corpora while the service looks valid.** Both the portable 18-capability constructor test and the two real workspace source sets fail after removing the shared mismatch check. A separate late-policy substitution fails actual ordinary/workspace callers, showing that neither the corpus witness nor policy is merely presentation metadata.

The post-commit probe additionally attacks truth across the fake/store boundary: the fake's repeated committed result can obscure receipt loss, whereas the real adapter returns a new pre-write collision on retry. The reviewer continues durable reads after the expected failure and confirms the first artifact, exposing that difference. One mint is explained by `NewThread` placing generation outside the retry closure, rather than inferred from a constant generator fixture.

Strongest new evidence: the real service/store malformed/stale test's valid-preview control, recaptured hostile fixtures, typed full-plan/receipt equality and tree oracle; the real committed-conflict test's call/release/mint counters plus readable document; and theme preview's forced `dark` variant, which avoids terminal background dependence while reaching theme's own hook.

Weakest evidence in isolation: a message substring would not prove sentinel/durability; pure compose rejection rows do not measure a filesystem; the creation fake alone cannot prove a document exists; and `t.Fatalf` in the owner committed-conflict mutant stops before its durable-read assertions. The paired real tests, sentinel assertions and reviewer nonfatal durable observations close those bounded gaps. Global release hooks are restored by subtest cleanup and cleared before reads; the original generic controls and other existing test coverage remain. No new helper, conformance framework or benchmark requirement was introduced.

### Thread closeout and tracked destinations

Fresh sandbox binary outputs (`thread-show.json`, `thread-frontier.json`) report **done=18, total=19, drained=18**, graph and projection **healthy**, `inconsistent=false`, no graph/problems, no pending frontier, and only task `6ggdkzv2tnta` **in-flight with a clear gate**. All three external gates—`6g5rxq1ravd3`, `6g697mp8s4tx`, `6g6scc9jgxae`—are completed and soundly completed, with none outstanding.

`closeout.py` independently resolves all 19 member IDs to actual task files. The **18 completed members have 101 checked ACs, zero unchecked ACs**, sound completion in the live projection, and `status: completed` in `origin/main` as well as the captured overlay (`closeout-members.json`). This verifies merged/completed bookkeeping without claiming a fresh retrospective review of every historical slice. Git history contains the prerequisite merges: #279 `33a6c1c8`, #280 `289b0bab`, #282 `6f2d8202`, #283 `28cec0ae`. The final task and Thread remain in progress; their checked local implementation ACs are not merge/release evidence.

| Existing followup | Verified actual destination and current status |
| --- | --- |
| Markerless identity recovery (workspace audit L2) | [6ggfd81jg0qg](../tasks/6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md): ready-to-start, four unchecked ACs; executed detour is reproduced here. |
| Thread-only TUI feedback (impact/recovery audit M1) | [6ggkdbg0816h](../tasks/6ggkdbg0816h-preserve-thread-only-lifecycle-recovery-guidance-in-the-tui.md): next-up, five unchecked ACs. Core-owned recovery remains covered; the presentation followup is not represented as implemented. |
| Dependency-policy ADR (hexagonal audit M3) | [6gg7e59mm68g](../tasks/6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md): ready-to-start, four unchecked ACs; acceptance is still required, not implied by executable lint. |
| Folded-YAML decoded-value corruption (storage audit H1) | [6g1dhhk6721x](../tasks/6g1dhhk6721x-a-surgical-frontmatter-write-re-folds-multi-line-block-scalars-onto-one-line.md): ready-to-start, **high priority/tier 2**, four unchecked ACs including decoded-value fidelity. Reviewer `TestAuditTrackedFoldedScalarStillChangesDecodedValue` executes five unrelated tier writes through `updateFrontmatter` and decodes with the actual YAML package: **41→46 bytes** (`independent-final.log`). It asserts the observed existing defect, not a fix. |

The YAML task's strengthened scope is honest and important: decoded-value corruption is more than wrapping churn. The production writer is unchanged in this test-only slice, and the storage defect has an actual destination and independently measured evidence. Deferring its fix outside the port migration does not falsely certify storage fidelity or release readiness. No duplicate finding is created for that already-tracked root cause. Test-rigour L1–L3 remain open and outside this task; no source audit or task state was changed by this review.

### Executed validation

All commands ran in the independent workspace with `GOCACHE=/tmp/taskflow-review-go-cache`; lint used `GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache`. Toolchain/platform: **Go 1.27.1, Darwin arm64**. The fresh build version is `v0.22.0-140-ga4ad85f6`; a denied external module stat-cache write was printed but build/binary execution succeeded.

```sh
go test ./internal/core ./internal/store ./internal/cli -run 'TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest|TestComposeThreadApplyPlanRequiresValidClockAndIDGenerator|TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity|TestServiceThreadCommittedFailureIsNotRetried|TestThreadCreationAttributesReleaseFailureAfterCommit|TestThreadApplyMalformedAndStalePlansDoNotPersist|TestGlobalSpace_UnknownListsKnownLabels' -count=1 -v
go test -race ./internal/core ./internal/store ./internal/cli -run 'TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest|TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity|TestServiceThreadCommittedFailureIsNotRetried|TestThreadCreationAttributesReleaseFailureAfterCommit|TestThreadApplyMalformedAndStalePlansDoNotPersist|TestGlobalSpace_UnknownListsKnownLabels' -count=5
go test -race -timeout 180s ./...
just lint
just build
```

All exit **0** (`focused.log`, `focused-race5.log`, `full-race.log`, `lint.log`, `build.log`). New focused tests ran before any mutation. The exact additional foundation command is retained in `foundation-command.txt`, with package results in `foundations.log`; it reruns semantic/pathless reads, optional paths, corpus mismatch, current policy/no-op/callback/constructor-option cases, real opening/identity, invocation-scoped authorization, and core/wire recovery. It does not claim new coverage of all historical requirements.

| Additional command/check | Result |
| --- | --- |
| Reviewer `go test ./internal/store -run '^TestAudit' -count=1 -v` | Pass (`independent-final.log`); empty/missing body, live advanced status, converged no-op, durable receipt observations, and tracked YAML reproduction. Reviewer file archived privately as `probe-store.go` and removed before transfer. |
| Reviewer `go test -race ./internal/store -run '^TestAudit' -count=5` | Pass (`independent-race5.log`). |
| Restored `go test ./... -count=1` and `just lint` | Pass (`restored-full.log`, `restored-lint.log`); implementation/private probe source restored. |
| `./bin/tskflwctl -C . lint --json --no-input` and `audit lint --json --no-input`, isolated config home | Both exit 0, `unreadable=[]`, `issues=[]` (`planning-lint.json`, `audit-lint.json`), repeated after report. |
| `./bin/tskflwctl -C . thread show make-planning-data-access-adapter-neutral --json --no-input` and `thread frontier … --json --no-input` | Exit 0; live rollup/gates above. |
| `go run ./internal/tools/docgen -out .git/isolated-review-workspace/evidence/generated-cli`, then `diff -qr docs/cli …/generated-cli` | Exit 0; no differences (`docgen.log`, `docs-compare.log`). |
| `go run ./internal/tools/schemacomments -out .git/isolated-review-workspace/evidence/schema_comments.json`, then `cmp internal/wire/schema_comments.json …/schema_comments.json` | Exit 0; 267 comments match (`schema-generate.log`). No update/golden generation was used. |
| Schema/optional branch review | `internal/wire/wire.go:343` remains **1.81**; no wire production change or new optional branch exists in this diff. New-optional-branch semantic validation is therefore not applicable; existing wire tests run in the full race suite. |
| Fresh product commands | **18** commands meet asserted exits, classifications and tree effects (`product.py`, `product.json`); no installed binary was used. |

Limitations: Linux/Windows, uid-0 permission behavior, terminal TUI interaction, every filesystem and arbitrary concurrent/non-cooperating edit schedules were not executed. The tree oracle is bounded as described above. The snapshot-mode and preview-persistence probes are distinct; an earlier valid-control failure is not counted as reaching a later hostile row. One initial preview-mutation text match aborted the harness without running a claimed test, then baseline restoration and the corrected exact branch ran successfully. Product harness corrections used the actual `not-found` spelling and required `--out` even in dry-run; those initial fixture assertions are not mutation kills or production findings. The final logs show the stated results.

### Isolation and guarded transfer attestation

The mandatory helper created an independent clone with independent Git metadata and captured the complete overlay in its sole reviewer baseline commit. Source access for this audit was limited to reading this brief, helper creation and final guarded transfer. All inspection, commands, scratch fixtures, mutations, restoration and report edits occurred inside the clone. No extra commit, staging, source project command, implementation change or lifecycle/finding-state change was made.

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.GuEVOI
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.GuEVOI/.git
baseline_commit=a4ad85f6452bad862c016b91a89600e727d0397e
source_blob=1f0afa4f765abff809816bbbe5d54c35da96411f
source_fingerprint=4476ea56229836353e343554d32c9abeb4145c14
deliverable=planning/audits/6ggxdjh7vpcr-2026-10-05-guarded-planning-boundary-closeout-codex.md
deliverable_changed=true
transfer=succeeded
```

Before transfer the helper verified that only this audit differs from baseline, the implementation/probes are restored, and the source audit has not drifted. The audit diff was inspected and only this file was atomically transferred. Private logs, scripts and archived reviewer tests remain in the retained sandbox. Helper output is `final-transfer.log`; retain this workspace until the owner confirms receipt.

## Owner reconciliation

Accepted the no-new-findings verdict after checking the retained independent Git directory,
baseline, transfer attestation, all 13 mutation records and their successful restored results.
Spot-checked the actual missing-body and committed-conflict logs and archived reviewer tests;
the latter distinguish lost recovery truth from reminting or duplicate persistence. All five
changed implementation test files match the restored reviewed snapshot byte-for-byte, and
the sandbox has no implementation delta. The fresh owner focused race suite also passed.

The additional empty-body/convergence, live-status, source-witness, late-policy, and snapshot
oracle experiments provide corroborating evidence, not new permanent tests or production fixes.
The tracked YAML defect remains unfixed; no duplicate task is needed. There is no demonstrated
in-scope defect requiring an implementation change from this review. Receipt is confirmed;
retain the private evidence until the final review round is reconciled.

This audit is settled and closed. Antigravity's assigned audit is still pending and was not
edited; the implementation snapshot is unchanged, and the task and Thread remain in progress
for that independent pass. Closing this audit does not complete the task, close the Thread, or certify
release readiness.
