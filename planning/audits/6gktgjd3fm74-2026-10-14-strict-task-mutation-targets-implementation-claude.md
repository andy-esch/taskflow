---
schema: 1
id: 6gktgjd3fm74
bucket: closed
area: strict-task-mutation-targets-implementation-claude
date: "2026-10-14"
updated_at: "2026-10-15"
---
# Audit: Strict task mutation targets implementation — claude — 2026-10-14

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens or em dash in place of the period. Status may be inline after `·` or on its own `**Status:** open` line within that finding section.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: for the critical regressions selected in the brief, execute a compiler-valid mutation and require the intended test to fail; exercise changed optional wire branches with non-default values in semantic validators; run changed repair advice against the state that recommends it; and use coordinated mutations when a nearby caller would otherwise preserve an invariant accidentally. A compile failure is an invalid probe, not a killed mutation. Apply the brief's bounded evidence floor rather than exhaustively mutating every new test.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Your handoff prompt prints exact `SOURCE_ROOT` and `AUDIT_REL`
values. Substitute them below; never derive them from your current directory, which may be a
different checkout. Then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="/Users/andyeschbacher/git/andy-esch/taskflow"
AUDIT_REL="planning/audits/6gktgjd3fm74-2026-10-14-strict-task-mutation-targets-implementation-claude.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)" && cd "${SANDBOX:?create failed: stop and report the blocker}"
```

Every later block begins from `${SANDBOX:?…}`, so it cannot run in the shared checkout after a
failed `create`. Run the helper from `$SOURCE_ROOT`, not from the sandbox: you may mutate the
sandbox's copy while probing, and only the read-only source copy can be trusted to verify your work.

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. The sandbox's push URLs are disabled; never re-enable them or
push to the source path. If creation fails, report the blocker; never fall back to the shared
checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then run
the helper's fail-closed verification:

```sh
cd "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX" &&
  git diff -- "$AUDIT_REL"
```

Copy that `verify` output into the report's isolation attestation: workspace path, resolved Git
directory, baseline commit, captured source blob/fingerprint, deliverable, push guard,
`verify=passed`, and `transfer=pending`. The report is final from this point, so `transfer=pending`
is the truthful value; never edit it to claim delivery. Re-run `verify` (its fields do not depend on
the report text), then transfer:

```sh
"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent or nested `.git`, a
removed or weakened push guard, source-deliverable drift, and empty reports; it copies back only
the assigned audit through a same-directory atomic rename. A successful transfer records a
helper-owned receipt and prints it (`transfer=succeeded` with the transferred blob). A refusal exits
2 and records nothing. Exit 3 means the report was already delivered when a later step failed: run
`"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" receipt --sandbox "$SANDBOX"` and put its
output, or its refusal, in your handoff. Put the printed receipt in your handoff message. Do not
copy anything else manually. Leave the workspace in place and report its path until the
implementation owner confirms delivery with that same `receipt` command. On refusal, preserve it and
report the conflict rather than resolving it in the shared checkout.

A report without the verification attestation, or a handoff without the transfer receipt, is
incomplete even if its technical findings are otherwise sound.

## Review brief

Can the strict task-target change still write to a different record, silently change
intent on retry, or weaken existing guards? This is an independent implementation
review, not approval of the checked planning criteria. Leave findings open.

Assignments (one shared contract, different primary evidence):
- **Claude:** core selection/retry budgets, source evidence and guarded lifecycle plan
  binding. Cross-check one real FS disappearance/collision path.
- **Antigravity:** exact persistence lookup, callback/no-op/preview/authorization order,
  and actual CLI producer behavior. Cross-check lifecycle's fresh authorization.

Do not read the sibling report before your own final report. Review the dirty capture,
not just HEAD. A strong negative result with killed mutations is useful; no finding quota.

## Review target

- Source checkout `/Users/andyeschbacher/git/andy-esch/taskflow`, branch
  `fix/identity-bound-task-mutations`, base `9cea9eee1a2135000d9847d4789f8007babaaf50`.
  Implementation, test migrations, new untracked tests/planning and both briefs are dirty.
- Task: planning/tasks/6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md.
- Contract owners: internal/core/store.go (`TaskMutationTarget`, the three task-write
  methods); service_task.go (`selectTaskMutationTarget`, `retryTaskMutation`, body
  transforms, SetFields, ReplaceBody/AppendBody, runTaskLifecycleMutation);
  task_lifecycle.go (`TaskLifecycleMutationResult.Plan`).
- Adapter: internal/store/fsstore.go (`resolveTaskMutationTarget`, SetFields,
  resolve/resolvePath); body.go (EditBody/TransformTaskBody/writeBody); cas.go
  (verifyUnchanged); lifecyclemutation.go (plan retention, fresh validation/materialization).
- New regressions: internal/core/task_target_test.go and internal/store/task_target_test.go;
  additionally task_correctness_test.go and the migrated direct-adapter tests. Verify
  their actual functions rather than assuming this inventory proves coverage.
- Guides: docs/architecture/planning-data.md and persistence.md; ADR-0008 governs
  machine contracts. No new wire/schema keys, automatic ID repair or transaction
  framework is intended.

Produce a verified consumer inventory: all callers/implementations of the changed
write signatures; primary Service entry points; first-selection and retry paths;
result/receipt/error consumers; fakes that can conceal missing identity or re-resolution.

Known, reproduced follow-ups (not fixed here):
- 6gkt3fphx1ts: ordinary audit/research/epic retries and audit bulk header repair.
- 6gkt3fpsfdd7: dependency/Thread endpoint reselection during fresh guarded replanning.
- 6gkt3fq0pv26: an ID selected earlier by TUI is passed as a fresh string at Service
  entry; an ID-shaped impostor can be selected after deletion. It depends on this
  task and is a navigation Thread member.
- 6gkkzvr92waq: tracked writes versus lint's explanatory-reason policy; human decision pending.
Do not refile these as new findings without a distinct in-scope consequence or a
concrete reason the boundary makes this implementation unacceptable.

## Intended contract to challenge

1. Initial user references retain existing fuzzy tiers, even ID-shaped human slugs.
   After successful application selection, only RecordSource.ID names the target.
   Declarations, labels and opaque locations cannot replace it. Core allows portable
   opaque IDs; filesystem filename validity remains adapter policy.
2. SetFields/EditBody/TransformTaskBody take an explicit selected target. FS resolves
   it exactly before callbacks, no-ops and previews, then retains content/path CAS.
   A vanished ID never becomes another record's slug; collisions remain actionable ambiguity.
3. Single-task initial selection conflicts and writes share one bounded retry budget
   and clock. Once selected, later failures cannot cause another alias resolution.
   Dry runs and whole-body replacement never retry. Tracked endpoint preflight retains
   its prior unretried policy; body retries never reselect either endpoint.
4. Lifecycle first selection stays inside the guarded graph. Retain selected semantic
   intent, not stale write bytes, policy, eligibility or impacts. Pre-planner contention
   establishes no identity; later empty results must not erase existing binding.
   Never retry a committed receipt, even if the cause includes ErrConflict.
5. Preserve explicit persistence authorization (including callbacks/no-ops/previews),
   source-set binding, surgical YAML/body fidelity, sentinel routing and receipt semantics.

These are falsifiable hypotheses, not a requested clean verdict. The end-to-end TUI
selection gap is explicitly not claimed closed by the narrower application guarantee.

## Mandatory evidence floor

Both reviewers independently:
- Run `go test -race ./internal/core ./internal/store ./internal/cli -count=1` twice,
  with GOCACHE outside the sandbox. Name executed cases and inherited checks separately.
- Build a fresh binary outside the tracked tree. In a fresh unregistered space run a
  valid field/body/criterion workflow through fuzzy and ID refs, tracked destination,
  rename then ID write, start/defer; run refused self/collision and no-write preview
  workflows. Capture actual exit statuses, names and byte hashes. Do not mask an
  expected failure by reporting the status of a following hash command.
- Verify at least one selected-identity drift case through the production Service+FS
  path rather than only the portable fake. Explain the producer and exact interleaving.
- Complete the distinct systemic second pass and a bounded hypothesis ledger.

**Claude's required falsification:**
- Execute a compiler-valid mutation resetting/re-resolving the target each retry.
  Require `TestTaskMutationsSelectSourceOnceAndRetainTargetAndClock` or
  `TestTaskMutationRetryKeepsSelectedIdentityAfterRename` to fail for its intended reason.
- Disable lifecycle plan binding (e.g. make its plan-binding condition false).
  Compile first; require `TestTaskLifecycleRetryKeepsFirstSelectedPlan` to fail with
  the wrong identity, then restore and rerun green.
- Independently test conflict before selection, after selection and between two
  lifecycle planned attempts (empty intervening result), plus a fresh blocked/forced
  lifecycle state. Check one shared budget rather than two separately capped loops.

**Antigravity's required falsification:**
- Replace exact action lookup `resolvePath(target.ID)` with fuzzy `resolve(target.ID)`
  in resolveTaskMutationTarget, compile, then require the missing-ID cases in
  `TestSelectedTaskPortsRefuseMissingDuplicateAndInvalidTargetsBeforeCallbacks` or
  `TestTaskMutationRetryNeverFallsBackAfterSelectedIDDisappears` to fail. Restore green.
- Author one independent real-store or caller-boundary probe (not a renamed existing
  fixture) with a disappeared selected ID and a case-folded ID-shaped impostor label;
  compare siblings as well as the returned error, in write and preview/no-op cases.
- Exercise one denied direct port mutation with a populated record and observable
  callback, and one lifecycle retry where a new prerequisite blocks start. Include
  the forced non-default blocker/receipt values, not only an empty-success shape.

Record exact edits, compile gate, intended regression, observed failure, killed/
survived/invalid classification and restored outcome. Compile failures are not kills;
unrelated failures are not proof of the intended invariant. No exhaustive mutation quota.

## Required hostile angles

- Source.ID differs from declared ID and slug; missing/blank evidence; opaque pathless
  identity; genuinely duplicate IDs versus duplicate labels. Find reachable producers
  or label malformed-port-only scenarios as defense-in-depth.
- Drift after ReadTask but before the first write, during CAS, old-alias reuse,
  a newly unique prefix match, disappearance and case-folded ID-shaped label fallback.
- Refused identity before transform/no-op/preview; a collision appearing AFTER exact
  lookup still needs CAS. Do migrated tests accidentally stop exercising the guard?
- Selection errors versus mutation errors; preview, non-retried replacement, no-ops,
  exhaustion, fixed timestamps, callback errors and preserved concurrent notes.
- Lifecycle missing/colliding target, lost plan on errors, changed eligibility,
  forced blockers and completion policy; inspect actual Plan retention sites before
  calling a hypothetical bad adapter a production bug.
- Source-set/authorization regressions from an extra read or changed fake signature;
  denied calls must not reach filesystem effects or callbacks.
- Public CLI/wire drift versus narrow internal Go port change. Do not globally
  reserve ID-shaped human refs as a purported quick fix without a compatibility decision.
- Systemic pass: shared helpers/fakes can appear fail-closed while another caller
  preserves the invariant accidentally. Trace actual inputs, effects and consumers.
  Challenge owner scope boundaries with evidence, not a list of already tracked bugs.

## Validation and restoration

Owner evidence: full `just test` (race), lint, build, docs-check and planning lint;
red-before-green rename/deletion/lifecycle regressions; fresh-binary throwaway workflow;
out-of-scope neighbor probes. These are inherited, not your independent proof.

Follow the injected isolated-workspace protocol. Keep scratch fixtures, caches and
binaries outside the captured sandbox tree. No source edits, commits, staging, branches
or pushes; the helper's sandbox baseline is the sole permitted commit. Restore every
mutation and temporary probe so only the assigned audit differs. Rerun focused tests,
diff hygiene and audit/planning lint. A permissions/isolation/evidence failure is
incomplete, not no-findings. Retain the sandbox on any refusal.

Run every helper command using the source copy. Put final verify evidence (truthfully
`transfer=pending`) in the report, then guarded transfer. Put its printed receipt in
the handoff; do not edit the frozen report to claim its own delivery. Retain the
sandbox until the owner confirms `receipt`. Never manually copy over a transfer refusal.

## Deliverable

Preserve the brief; fill only your assigned report. Findings first, consolidated by
root cause, with severity, exact producer path, reachable consequence, reproduction,
evidence, recommendation and residual uncertainty. Prefer tool-created canonical
findings, all open. Separate functional defects, coverage gaps, preferences and
malformed-port defenses. Include snapshot/verdict limits, verified inventory, checks
executed vs inherited, mutation/probe ledger, surviving hypotheses and verification
attestation. A clean baseline alone is insufficient for a ready/no-findings verdict.

## Reviewer report

## Findings

#### M1. Strict selection blocks task set repair of malformed typed fields · **Status:** fixed 2026-10-15

**File:** internal/core/service_task.go:496 | **Component:** core/store
**Effort:** S · **Urgency:** soon

**Class:** functional defect (unintended CLI behavior regression), plus a direct-adapter regression that hides it. This is not a wrong-record write.

**Producer path:** `tskflwctl task set <ref> --<field> <value>` → `internal/cli/task.go:676` `app.Svc.SetFields` → `internal/core/service_task.go:792` `retryTaskMutation` → `selectTaskMutationTarget` (`service_task.go:496`) → `FS.ReadTask` → `readTask` (`internal/store/fsstore.go:233`) → `parseTask(content, path)` (`fsstore.go:245`), which runs a strict typed decode of the *current* frontmatter. At base `9cea9eee`, `FS.SetFields` resolved the reference itself and applied `updateFrontmatter` to the raw YAML nodes. It then parsed only the *updated* content (now `fsstore.go:282-300`). So an update replacing the malformed value succeeded. When the original was malformed for an unrelated reason, base blamed the file: "already has malformed frontmatter (not caused by this update)" (`fsstore.go:298`).

**Reachable consequence:** Suppose someone hand-corrupts a typed frontmatter field (`tier: three`, `tags: {a: b}`, `priority: [x]`). `task set` can no longer repair it by setting that field. Both the write and the `--dry-run` preview now exit 11 at selection, before the adapter runs. From Service/CLI, the adapter's pre-existing-corruption attribution branch is now unreachable outside a race. Lint reports exactly this state (`field "tier" must be a whole number, but it is a string ("three")`). `task set` was the non-interactive repair route. What remains is hand editing or the interactive `task edit`, whose code is unchanged and was not exercised here. The owning task's design checkpoint says to stop for human input if the implementation changes promised CLI behavior. The task, the guides and the tests do not record this change.

**Reproduction:** I built fresh binaries outside the sandbox from `9cea9eee` and from the captured dirty tree, then ran them in fresh unregistered spaces. The scripts are `repair.sh` and `repair2.sh` in the reviewer scratchpad. Each case first hand-corrupts one field of a valid task.

| Corruption; command | base exit / file | dirty exit / file |
| --- | --- | --- |
| `tier: three`; `task set --tier 2` | 0 / repaired to `tier: 2` | 11 / unchanged |
| `tier: three`; `--dry-run task set --tier 2` | 0 "would update" / unchanged | 11 / unchanged |
| `tags: {a: b}`; `task set --tags go,core` | 0 / repaired | 11 / unchanged |
| `priority: [x]`; `task set --priority high` | 0 / repaired | 11 / unchanged |
| `tier: three`; `task set --priority high` | 11 "already has malformed frontmatter (not caused by this update)" | 11 "`<path>`: validation failed: malformed frontmatter: …" |
| duplicate `tier` key; `task set --tier 4` | 11, blamed on the file | 11 at selection |
| `tier: three`; `task ac --check 1`, `task append`, `task set --body` | 11 / unchanged | 11 / unchanged |

Body operations were already refused at base by their reload parse. Only field repair, the field preview and the attribution changed. All exit codes come directly from the `tskflwctl` invocation, and file changes were detected with SHA-256.

**Why the suite stays green:** `TestFS_SetFields_PreexistingCorruptionAttributedToFile` (`internal/store/audit_fixes_test.go:152`) was migrated to call `FS.SetFields(core.TaskMutationTarget{…})` directly. It bypasses the new selection, so it passes while the production path never reaches the branch it pins. The portable `selectedTaskStore` fake (`internal/core/task_target_test.go:13-34`) returns a parseable record for every reference. No Service or CLI test covers a task that resolves but fails to parse, and no test pinned field repair at base either.

**Fix options:** Get an explicit human decision on whether malformed tasks stay addressable by field writes. To keep base behavior without weakening identity, select through an identity-only adapter resolution. It would return the resolved occurrence's established `RecordSource` without a typed decode. The FS identity comes from the filename (`taskSource`, `internal/store/entity_read.go:14`), so it does not depend on whether the frontmatter is valid. `FS.SetFields` would keep its post-update reload validation and attribution. Add Service- or CLI-level regressions for repairing a type-malformed field (write and preview) and for the attribution message. If the stricter behavior is intended, record that in the task, the planning-data guide and a CLI regression instead.

**Residual uncertainty:** I did not search agent-facing guidance outside `docs/` and Go sources for an explicit "repair with `task set`" instruction, and lint names no command. An untyped selection must still refuse a blank `RecordSource.ID` and duplicate exact IDs (the current `requireSourceID` and `resolveExactID` behavior).

**Recommendation:** Select by adapter identity without a typed decode, or record an explicit decision; add a Service-level repair regression.

**Resolution:** Confirmed through the production CLI path: six
    repair/write-preview cases and four corruption-attribution cases failed
    before correction. TaskStore.ResolveTaskSource now selects
    adapter-established identity without typed decoding; core retains that ID
    and FS still requires strict ownership, authorization, CAS and validation of
    updated bytes. CLI regressions cover malformed tier/tags/priority repair,
    unchanged previews, unrelated/duplicate-key corruption attribution and
    refusal of malformed body writes. Portable tests now expose a typed-read
    failure; real-store drift and contention tests remain green. Fresh-binary
    repair/preview/error checks and full race, lint, build and docs gates pass.
    This preserves base behavior rather than adopting a stricter repair policy.

## Verdict, snapshot and limits

**Verdict: not ready as captured.** One open Medium functional regression (M1). I found no
path that writes to a different record, and none that changes intent on retry. Five
compiler-valid mutations of the binding, retry-budget and tracked-handoff logic were each
killed by the intended regressions. Seven independent probes also passed. The defect is the
opposite failure: the new selection step refuses a write that base allowed.

- **Snapshot:** sandbox baseline commit `546d88086fc0d25820d6f3510d7f013b7b53d1a0`, which
  captures the dirty and untracked handoff over base `9cea9eee1a2135000d9847d4789f8007babaaf50`.
  I reviewed the captured diff of 32 files, not HEAD alone.
- **Interleavings:** drift evidence comes from deterministic test hooks plus one probe
  with a second, independent FS instance. I made no claim about real multi-process
  frequency. I ran on macOS (APFS, case-insensitive) only.
- **Not reviewed:** I did not read the sibling report. I did not re-review the tracked
  TUI selection gap (`6gkt3fq0pv26`), non-task retries (`6gkt3fphx1ts`), graph/Thread
  replanning (`6gkt3fpsfdd7`) or the tracked-reason policy (`6gkkzvr92waq`). I found
  nothing that changes their boundaries.

## Coverage notes (not filed as findings)

- **Portable-only kills.** Probes C and E show that two rules are pinned only by portable
  fakes: an empty intervening lifecycle result must not erase the binding, and selection
  and writes share one budget. Every real-store test stayed green under both mutations.
  FS has a real producer for both cases: `repositoryPlannerReentryError`
  (`internal/store/lock.go:85`). It returns `ErrConflict` from `rejectRepositoryPlannerCall`
  in `ReadTask` (selection) and from `checkedWriteLock`/`enterRepositoryPlanner` before
  a lifecycle plan exists. It fires whenever another planner in the same process is
  active, such as concurrent TUI commands. A real-store regression would be cheap
  defense; the current portable tests do kill both mutations.
- **Store-only kill.** Probe D shows that tracked-handoff source binding is pinned only by
  the real-store test `TestTaskSelectedIdentitySurvivesDriftBeforeFirstMutation`. The core
  `TestCriteriaTracking*` tests stayed green.
- **Probe P7 sensitivity not demonstrated.** No core mutation can make lifecycle reuse a
  stale analysis, because core holds none; a probe would need an adapter mutation.

## Verified consumer inventory

**Implementations of the three changed write signatures.** All were checked with
`grep 'func (.*) SetFields(|EditBody(|TransformTaskBody('` over `*.go`.

- **Production:** `FS.SetFields` (`internal/store/fsstore.go:257`), `FS.EditBody`
  (`internal/store/body.go:146`) and `FS.TransformTaskBody` (`body.go:252`). All three
  resolve through `resolveTaskMutationTarget` (`fsstore.go:345`). That calls
  `validEntityID` and then `resolvePath` → `resolveExactID`, a case-sensitive exact
  filename-ID match (`internal/store/resolve.go:288`), after `authorizeMutation` and
  before callbacks, no-ops and previews.
- **Test fakes:** `conflictStore` (`internal/core/occ_retry_test.go:44-52`), `nopStore`
  (`internal/core/service_epic_test.go:59-71`), `selectedTaskStore`
  (`internal/core/task_target_test.go:44-54`) and `criterionIdentityStore`
  (`internal/core/task_correctness_test.go:33`, `TransformTaskBody` only).
- **No production decorators.** `core.Store` is embedded only in test doubles
  (`internal/tui/audit_unparsed_test.go:17`, `internal/cli/planning_ports_test.go:19`,
  `internal/cli/entity_info_pathless_test.go:19`, `internal/cli/creation_receipt_test.go:72`).

**Port callers.** The only non-test callers are in core:
`internal/core/service_task.go:532` (`TransformTaskBody`), `:562` (`EditBody` replace),
`:575` (`EditBody` append) and `:793` (`SetFields`).

**Service entry points and their selection and retry paths:**

| Entry | First selection | Retry |
| --- | --- | --- |
| `SetFields` | `retryTaskMutation` → `selectTaskMutationTarget` (`:496`, `ReadTask` + `requireSourceID`) | one shared `retryOnConflict` budget (`:510`); target kept once set |
| `AppendBody` | same | same; `now` captured once |
| `SetAcceptanceCriterion`, `SetCriterionState`, `SetCriteriaState`, `EditCriteria` | `transformTaskBody` → `retryTaskMutation` | same |
| `SetCriteriaTracked` (`:435`) | unretried `ReadTask` preflight of source and destination, plus a canonical self check | `transformTaskBodyAt` (`:536`) with a fixed source target; neither endpoint is re-selected |
| `ReplaceBody` (`:557`) | one selection | none |
| `Move`, `DeferTask`, `NewTask --start` | planner inside the FS guard (`MutateTaskLifecycle`) | `runTaskLifecycleMutation` (`:617`): binds on a non-empty `Plan.TaskID` (`:626`), stops on `Committed`; create-and-start has an empty `TaskID`, so it never binds |

**Producers of those entries:** CLI `internal/cli/task.go:489` (`EditCriteria`), `:538`
(`SetCriteriaTracked`), `:540` (`SetCriteriaState`), `:670` (`ReplaceBody`), `:676`
(`SetFields`) and `:789` (`AppendBody`). Their references come from `resolveOne`
(`internal/cli/fill.go:179`): the raw argument, or a slug from the picker. The TUI calls
`SetFields` with a previously selected `ref.key` (`internal/tui/edit.go:358,370`); that
gap is tracked as `6gkt3fq0pv26`.

**Result and receipt consumers.** CLI `reportTaskMutation`; the lifecycle receipt from
`taskLifecycleReceipt`, plus `TaskLifecycleMutationFailure` for a committed error. I
observed these exit codes: validation 11, ambiguity 13. `git diff 9cea9eee HEAD --stat`
shows no changes under `internal/wire`, `internal/cli`, `internal/tui` or `docs/cli`.

**Fakes that can hide defects.** `selectedTaskStore` returns a parseable record with one
fixed `sourceID` for any reference, which hides selection preconditions (M1).
`lifecycleTargetStore` never returns a validated plan that differs from the raw plan.
The migrated direct-adapter tests pass a pre-selected `TaskMutationTarget`, so Service
selection is bypassed (M1).

## Checks executed and inherited

**Executed by me in the sandbox, with GOCACHE in the reviewer scratchpad:**

- `go test -race ./internal/core ./internal/store ./internal/cli -count=1`, run twice.
  Both exited 0. Each run had 2905 passing tests and subtests, 0 failed and 0 skipped.
  Durations were core 1.48s/1.45s, store 10.28s/10.39s and cli 10.03s/10.22s.
- Executed target tests, with passing test-plus-subtest counts from run 1:
  - `TestTaskMutationsSelectSourceOnceAndRetainTargetAndClock` (9)
  - `TestTaskMutationSelectionRefusesMissingEvidenceAndReadErrors` (33)
  - `TestTaskLifecycleRetryRetainsBindingAcrossUnplannedContention` (1)
  - `TestTaskMutationTargetDoesNotCarryLocationOrDeclaration` (1)
  - `TestTaskMutationSelectionAndWritesShareOneRetryBudget` (1)
  - `TestTaskMutationRetryKeepsSelectedIdentityAfterRename` (17)
  - `TestTaskMutationRetryNeverFallsBackAfterSelectedIDDisappears` (17)
  - `TestTaskLifecycleRetryKeepsFirstSelectedPlan` (5)
  - `TestSelectedTaskPortsRefuseMissingDuplicateAndInvalidTargetsBeforeCallbacks` (31)
  - `TestTaskInitialReferenceStillResolvesAnIDShapedHumanSlug` (1)
  - `TestTaskSelectedIdentitySurvivesDriftBeforeFirstMutation` (55)
  - `TestTaskMutationCASStillRefusesAnIdentityDuplicatedAfterLookup` (9)
  - `TestTaskLifecycleRetryRevalidatesEligibilityFromFreshGraph` (3)
  - `TestCriteriaTrackingUsesSourceIdentity` (13)
  - `TestCriteriaTrackingPreservesCanonicalResolutionErrors` (7)
  - `TestCriteriaTrackingDoesNotMisreportCanonicalFallbackAsSelf` (1)
  - `TestFS_SetFields_PreexistingCorruptionAttributedToFile` (1)
  - `TestSetFields_FuzzyQueryDoesNotSpuriouslyConflict` (1)
  - `TestVerifyUnchanged_*` (5)
- `go build ./...` and `go vet` on the mutated packages, run as the compile gate for
  every probe.
- Fresh `tskflwctl` binaries built outside the sandbox tree: one from the dirty capture
  and one from a `git archive 9cea9eee` export.
- The fresh-binary workflow and the M1 malformed-field matrix, both described below.
- After restoration: diff hygiene, `audit lint` and planning `lint`. The results are in
  the attestation.

**Inherited from the owner and not re-run by me:** full `just test`, `just lint`,
`just build`, `just docs`/`docs-check`, architecture navigation, `lint --links` in the
owner's space, red-before-green runs, and the owner probes under
`/private/tmp/taskflow-correctness-review.*`.

### Fresh-binary CLI workflow

I used a fresh unregistered space (`TSKFLW_NO_REGISTER=1`, isolated `HOME`) with tasks
`alpha` (`6gkypyf03gck`), `alpha-two` (`6gkypyf1gs79`) and `beta` (`6gkypyf36tkn`).
Each exit code is the `tskflwctl` status itself. "Tree" means the SHA-256 of all task
files before and after each step.

| Step | Exit | Result |
| --- | --- | --- |
| `task set alpha --priority high` (exact slug) | 0 | alpha updated |
| `task set BET --tier 2` (case-folded prefix) | 0 | beta updated |
| `task set 6gkypyf1gs79 --tier 4` (ID) | 0 | alpha-two updated |
| `task append 6gkypyf03gck` (ID) | 0 | alpha updated |
| `task ac alpha-t --check 1` (fuzzy) | 0 | alpha-two updated |
| `task ac 6gkypyf03gck --add …` (ID) | 0 | alpha updated |
| `task ac alpha --tracked 2 --to bet` | 0 | the reason records destination ID `6gkypyf36tkn` |
| `task rename alpha Omega`, then `task set 6gkypyf03gck --priority low` | 0, 0 | the ID write landed on `omega` |
| `task set alpha --priority medium` after the rename | 0 | resolved by prefix to `alpha-two`; tree unchanged (value and date already equal) |
| `task start 6gkypyf03gck`; `task defer alpha-two --until 2026-12-01` | 0, 0 | started and deferred |
| `task ac 6gkypyf03gck --tracked 1 --to 6gkypyf03gck`; same with slug `omega` | 11, 11 | self-handoff refused; tree `fe3949971986` unchanged |
| Duplicate `6gkypyf36tkn-beta-copy.md`, then `task set <id>` / `task append <id>` | 13, 13 | "matches 2 tasks by id"; tree unchanged |
| `--dry-run` set, append, ac, `set --body`, start | 0 ×5 | tree `fe3949971986` unchanged |

Final hashes: `76bb3644a55e omega`, `044ca6918b2d alpha-two`, `67345f74f40d beta`.
`lint --links` exited 0.

The stale-alias row is a mistake in my expectation, not a defect. I had expected a
refusal. Initial fuzzy selection is the documented contract (intended contract 1), and
here it chose `alpha-two`.

## Mutation and probe ledger

| ID | Exact edit (sandbox only) | Compile | Intended regression and observed failure | Class | Restored |
| --- | --- | --- | --- | --- | --- |
| A | `retryTaskMutation`: after `mutate(target)`, set `target = TaskMutationTarget{}` | ok | `TestTaskMutationsSelectSourceOnceAndRetainTargetAndClock`: `reads=3`. `…AfterRename`: 14 of 16 subtests fail with "retry changed a sibling" (tracked ×2 pass via the fixed target). `…AfterSelectedIDDisappears`: 14 of 16 fail with `err=<nil>` (impostor written). Shared-budget test: `reads=3`. | killed | `git checkout`; 4 tests green |
| B | `runTaskLifecycleMutation`: `if result.Plan.TaskID != ""` → `if false && …` | ok | `TestTaskLifecycleRetryKeepsFirstSelectedPlan` 4/4: committed receipt for the sibling `alpha-two` (move and defer); with `remove=true`, `nil` instead of `ErrNotFound`. The core unplanned-contention test plans the sibling's ID. | killed | green |
| C | add `else { retryPlanner = planner }` | ok | The core unplanned-contention test fails; real-store lifecycle tests stay green | killed (portable only) | green |
| D | `SetCriteriaTracked`: `transformTaskBodyAt(TaskMutationTarget{ID: sourceRecord.Source.ID}, …)` → `transformTaskBody(slug, …)` | ok | `TestTaskSelectedIdentitySurvivesDriftBeforeFirstMutation`: rename-and-reuse/tracked/write reports "selected renamed task was not changed"; delete/tracked write and preview get `nil` instead of not found. Core tracking tests stay green. | killed (store only) | green |
| E | Split selection and writes into two separately capped `retryOnConflict` loops | ok | `TestTaskMutationSelectionAndWritesShareOneRetryBudget`: "exhausted combined budget should surface conflict: `<nil>`" | killed (portable only) | green |

**Independent probes.** These were temporary files, now deleted; copies are in the
reviewer scratchpad under `probes/`.

| Probe | Path | Result |
| --- | --- | --- |
| P1 conflict before selection | core fake: unplanned conflict, then a fresh graph where `alpha` names another task | pass: 2 calls, 1 sleep, plans from fresh data |
| P2 after selection, then empty, then planned | planned conflict → unplanned → planned conflict → unplanned → planned, with the raw reference finally naming the sibling | pass: all 3 plans keep the first ID, `To` and `dependency-gate`; fails under B, so the probe is sensitive |
| P3 one budget | persistent conflicts, starting unplanned or planned, `maxRetries=3` | pass: exactly 4 calls and 3 sleeps |
| P4 committed conflict | `Committed: true` with `ErrConflict` | pass: 1 call, `TaskLifecycleMutationFailure`, committed receipt |
| P5 dry-run lifecycle | planned conflict on preview | pass: 1 call |
| P6 production Service+FS drift | see the interleaving below | pass for fields, checkbox and append; under A all three fail and the newcomer is written |
| P7 forced, then unblocked | `Move(--force dependency-gate)` blocked by a real edge, with the edge removed before the first verify | pass: retry commits, `Forced=false`, no outstanding blockers, `Before.Gate=clear` |

**P6 producer and exact interleaving.** This is the Service+FS drift case the brief
requires. The second store is an independently constructed `store.NewFS` on the same
root, standing in for another CLI process.

1. `Service.SetFields("alph")` calls `selectTaskMutationTarget`. `FS.ReadTask` resolves
   the unique prefix and returns `Source.ID = ID(alpha-one)`.
2. `FS.SetFields(target)` calls `resolvePath(ID)`, then reads and validates the file.
3. In `testHookBeforeSetFieldsWrite` (body ops use `testHookBeforeBodyWrite`), the second
   FS runs the public `RenameTask(ID, "Gamma")` and creates `<newID>-alpha-zeta.md`. The
   raw reference `alph` now names only the newcomer.
4. `writeLock` runs, then `verifyUnchanged`. `resolvePath(ID)` returns the renamed path,
   which differs from the path read, so the write fails with `ErrConflict`.
5. The retry runs with the retained target, so `resolvePath(ID)` finds the `gamma` path
   and the write lands there. Every other file stays byte-identical. Under mutation A,
   the retry re-resolves `alph` and writes the newcomer instead.

## Bounded hypothesis ledger

| Hypothesis | Evidence | Outcome |
| --- | --- | --- |
| Declared ID or slug can stand in for `Source.ID` | FS identity is the filename (`taskSource`, `internal/store/entity_read.go:14`); the portable fake's declaration and slug are ignored | held |
| Opaque, pathless identity | core passes `urn:task:opaque` through unchanged; FS rejects invalid IDs via `validEntityID` as adapter policy | held |
| A vanished ID falls back to a slug or case-folded slug | `resolveExactID` matches exactly and case-sensitively; mutation A shows the retry tests detect it | held |
| A collision after exact lookup slips past CAS | `verifyUnchanged` returns `ErrAmbiguous`, which is not retried (`…CASStillRefuses…`: 8 ops, 0 sleeps); CLI exit 13 | held |
| Same ID re-created after deletion (incarnation) | a retry targets the new occurrence because the ID is the identity; incarnation policy is out of scope | not filed |
| The retained lifecycle plan carries stale policy | the plan holds only `TaskID`, `To`, `Override`, `RevisitAt` and `Create` (`internal/core/task_lifecycle.go:55`); the store recomputes the analysis from the fresh graph (owner fresh-blocked test, P7) | held |
| A validated plan replaces intent | the store sets `result.Plan = validated` only on success (`internal/store/lifecyclemutation.go:76`), and success returns the input plan; on error the raw plan from `:66` is kept | held |
| A committed receipt is retried | the loop requires `!result.Committed`; P4 | held |
| Selection and writes use two budgets | a single loop per path; mutation E is killed; P3 | held |
| Retries change the timestamp | `now` is captured once per operation (core `clocks==1`) | held |
| Selection read precedes store authorization | only observable by an unauthorized caller of a mutation entry. Production composes `GuardedMutations` (`internal/appwiring/wiring.go:30`); the read-only FS is used only for Atlas summaries (`internal/spacestore/fs.go:105`). No producer found. | not filed (authorization order is Antigravity's area) |
| Strict selection adds read-model preconditions to writes | reproduced with both binaries | **M1** |
| The tracked destination vanishes after preflight | the reason records a dangling ID; planning-data.md documents this as a non-guarantee | not filed |
| Public CLI or wire drift | no wire/CLI/TUI/docs-cli diffs; initial fuzzy tiers unchanged (`TestTaskInitialReferenceStillResolvesAnIDShapedHumanSlug`, workflow) | held, except for M1's behavior change |

## Systemic second pass

1. **The shared selection helper.** `selectTaskMutationTarget` uses the general-purpose
   `ReadTask` as its identity oracle. That makes every task write inherit the read
   model's validity contract, a full typed decode that the write adapters deliberately
   did not require. This is the one systemic issue I demonstrated (M1).
2. **Test helpers that hide defect classes.** Every migrated direct-adapter test, and both
   new portable fakes, start from an already selected or always-parseable record. They
   check port behavior correctly but cannot see preconditions that selection adds. M1
   passes the whole suite for exactly this reason.
3. **State changing between projection and action.** Drift between selection and the
   first write, and during CAS, is covered by the owner's drift tests and by P6 with a
   second store instance. Lifecycle selection stays inside the guard; P1, P2 and P7
   cover pre-planner contention, planned contention and changed eligibility.
4. **Boundaries that only appear to fail closed.** I confirmed that exact lookup runs
   before callbacks, no-ops and previews (`TestSelectedTaskPortsRefuse…`, 31 cases),
   that `requireSourceID` rejects blank evidence, and that a late collision returns
   ambiguity rather than a retried conflict. I found no fail-open path.

## Verification attestation

Output of the helper's fail-closed `verify`, run from the source copy against this
sandbox after every probe was restored and only this report differed:

```text
sandbox_path=/private/tmp/isolated-review.2enrt8
git_dir=/private/tmp/isolated-review.2enrt8/.git
baseline_commit=546d88086fc0d25820d6f3510d7f013b7b53d1a0
source_blob=42aa225def87c96fdc13142a1bde2615ca6f3db6
source_fingerprint=f445e2160591b27cf5563518c9f5f0d2d0396488
deliverable=planning/audits/6gktgjd3fm74-2026-10-14-strict-task-mutation-targets-implementation-claude.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner triage (2026-10-15)

The source helper's `receipt --sandbox /private/tmp/isolated-review.2enrt8` confirmed
delivery before dispositions: `transfer=succeeded`, `delivery=confirmed`, transferred
blob `b1b1d711060431cecb2a766df9fedc0f76b235fb`. The frozen attestation above remains
the reviewer's pre-transfer evidence, not the delivery receipt.

M1 is fixed without adopting a stricter repair policy. Its resolution records the
parse-free source selection and CLI regressions. Both non-finding coverage notes
also gained real-store regressions for shared selection/CAS budgets and intervening
unplanned lifecycle contention; each passes ten race-enabled runs. See the
[task disposition](../tasks/6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md#claude-review-disposition-2026-10-15)
for gates and fresh-binary evidence. The captured-snapshot verdict above is retained;
closure reflects owner verification, not a reviewer rerun against the corrected tree.
