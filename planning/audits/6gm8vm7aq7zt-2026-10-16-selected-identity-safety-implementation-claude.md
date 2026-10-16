---
schema: 1
id: 6gm8vm7aq7zt
bucket: closed
area: selected-identity-safety-implementation-claude
date: "2026-10-16"
updated_at: "2026-10-16"
---
# Audit: Selected identity safety finish — claude — 2026-10-16

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
SOURCE_ROOT="<SOURCE_ROOT from your handoff prompt>"
AUDIT_REL="<AUDIT_REL from your handoff prompt>"
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

Can a source rename, disappearance or alias reuse redirect a selected mutation or
local TUI action after this change? Challenge the whole selection-to-effect path,
not just whether CAS detects the first conflict. This is a three-task safety finish,
not a new entity/transaction framework.

Use complementary primary lenses, while both reviewers check the shared contract:

- Claude: non-task writes and graph/Thread retry binding, including durable-prefix
  boundaries. Cross-check one queued local-path refusal through the actual TUI caller.
- Antigravity: selected local-path capabilities, all entity consumers, source-set and
  asynchronous session ownership. Cross-check the Thread creation collision case.

These assignments divide evidence, not invariant ownership. Independently report any
demonstrated defect in the other lens. Do not read the sibling report before yours
is final, and do not manufacture findings to meet a quota.

## Review target

Source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`.
Branch: `fix/remaining-selected-identity-boundaries`.
Base: `2b8d05fc` (merged selected-task action work, PR #308).
Implementation checkpoint: `89b057f0029bbe7c46f0f61e75436a4d46b47895` plus the final
planning/audit commits present at capture. Review the complete base-to-captured-state
change, including any staged, unstaged, untracked or deleted content captured by the
isolation helper, not HEAD alone. Stop and report unexpected implementation drift.

Implementing tasks:

- [Non-task mutations](../tasks/6gkt3fphx1ts-pin-non-task-mutation-retries-to-selected-source-identities.md).
- [Graph/Thread retries](../tasks/6gkt3fpsfdd7-keep-graph-and-thread-retry-endpoints-fixed-during-replanning.md).
- [Selected local paths](../tasks/6gm30jat0xsy-keep-selected-tui-path-actions-bound-to-exact-identities.md).
- [Safety Thread](../threads/6gm4h71hm1yt-finish-selected-identity-safety.md): three
  independent tasks bundled for delivery, not an artificial prerequisite chain.

Verify this consumer inventory in the sandbox and extend it when a real caller is
missing. Cite owning symbols/lines rather than assuming this list is exhaustive:

- `internal/core/entity_target.go`, `service_audit.go`, `service_research.go`,
  `service_epic.go`, `finding.go`, `finding_create.go`: parse-free initial selection,
  explicit target types, ordinary/selected entry points, audit occurrence repair.
- `internal/store/entity_target.go` and audit/research/epic writers/CAS helpers:
  exact ownership at every callback, preview/no-op and write boundary.
- `internal/core/dependency_operations.go` and `service_thread.go`: binding of
  successful planner intent, empty-result/pre-planner contention and fresh effects.
  `internal/store/graphmutation.go`, `threadcreation.go`, `threadmutation.go` own
  guarded reads, validation, materialization, CAS and committed outcomes.
- `internal/core/store.go`, `selected_entity_path.go`, `service_entity_path.go`,
  `service.go` and `internal/store/paths.go`: all five optional path families,
  reference versus selected semantics, source-set checking/detachment.
- `internal/tui/commands.go`, `thread_projection.go`, `model.go`, `session.go`:
  selected audit/epic mutation calls, both local-path lookups, detail joins,
  source-ID refusal, service capture and selection/list/workspace envelopes.
- Bulk apply (`internal/core/service_thread_apply.go`) already consumes a materialized
  repository-bound plan; graph repair (`dependency_repair.go`) consumes source
  selectors or deliberate auto-cleanup, not fuzzy task aliases. Challenge that
  exclusion if you can show a concrete affected production path. Global dependency
  migration has no first-selected human endpoint.

Start with the owning comments and architecture guides at
`docs/architecture/{planning-data,persistence,tui}.md`. Preserve ADR-0003 filename
identity and ADR-0008 machine-contract promises. No wire/schema revision or ordinary
CLI matching change is intended; internal Go ports are not an external adapter API.

## Intended contract to challenge

1. Ordinary non-task writes resolve a human reference once, parse-free, and retain
   adapter-established source identity across one bounded selection/write budget.
   Already-selected TUI audit/epic calls do not repeat that initial matching.
2. Persistence refuses absent/duplicate source identity before callbacks, preview,
   no-op success or writes. Malformed typed values remain repairable through existing
   field/path workflows. Audit header repair uses each snapshot occurrence, not slug.
3. Dependency and Thread selection begins only inside an authoritative guarded
   snapshot. After successful selection, renamed/missing endpoints never become a
   different record, including when their IDs become another record's human label.
   A conflict before any planner call establishes nothing; an empty result after
   selection does not erase it. Creation-ID validation can conflict before a result
   plan exists, so binding cannot depend solely on a validated receipt.
4. Retain intent only. Recompute current dependencies, memberships, lifecycle policy,
   impacts and source evidence on each attempt; preserve unrelated concurrent changes.
   Preview remains non-retried, and committed/partial results stop retries.
5. Selected local navigation is optional, exact, unique and parse-free for every kind.
   CLI reference methods remain fuzzy by their existing policy. No diagnostic Location,
   foreign source set, missing capability or display declaration grants a local path.
   Both queued lookups and detail path joins use selected methods; existing semantic
   detail source-ID checks must remain intact.
6. Local commands remain in `tea.Cmd`, capture their originating Service and obey
   selection/list/session guards. Same ID and list generation in another workspace
   cannot make an old result current. No new filesystem work enters core or reducers.

Non-goals: interactive mutation editors, ordinary creates other than retained Thread
member intent, global transaction/retry frameworks, new navigation UX, or pretending
advisory guards eliminate raw-editor races. The final lookup-to-effect/verify-to-rename
window and delete/recreate with the identical canonical ID are not solved. Epics use
the complete filename stem: changing it removes the old identity, not a rename the
system can follow. Do not silently change these policies during review.

## Mandatory evidence floor

Verify the named tests first. In your primary lens execute one valid-but-hostile
workflow and one refused workflow through a real producer, with byte/mode snapshots
or observed queued-effect refusal. Reusing an existing test is fine after verifying
what it actually exercises. A fake port call alone does not prove filesystem safety.

Claude's minimum probes:

- `TestNonTaskRetriesRetainSelectedSourceIdentity`: inventory all nine families,
  distinguish stable-ID audit/research renames from whole-stem epic disappearance.
  Check `TestAuditHeaderRepairUsesOccurrenceIdentityWithDuplicateLabels` and
  `TestSelectedAuditPortsRefuseDuplicateIdentityBeforeCallbackOrNoop`.
- `TestDependencyRetriesPinBothEndpointsAndRecomputeCurrentEdges` and
  `TestThreadRetriesKeepSelectedThreadAndMemberIdentities`: challenge fresh current
  state rather than simply expecting a write to the original file.
- `TestThreadCreationBindsMembersBeforeCollisionValidation`: the reachable producer
  resolves members, rejects a colliding generated ID, then retries after a raw editor
  removes the collision and reuses the old member alias. Explain why this differs
  from a post-validation CAS conflict.
- Compiler-valid mutation: make either Thread bound planner forget selection on
  every callback, or remove dependency endpoint binding. Require the intended test
  above or `TestGraphRetriesRetainBindingAcrossUnplannedConflicts` to fail on target
  identity, then restore and pass it. A compiler error is not a killed mutation.
- Bounded cross-check: run a task impostor case from the local-path test below.

Antigravity's minimum probes (settle these concrete hypotheses, not a generic scan):

- HYPOTHESIS A: a missing selected ID can still become an ID-shaped or case-folded
  label at either lookup. Trace actual `Model.Update` E/Y producers through
  `localPathLookup`, the Service and the FS resolver, then the returned effect guard.
  Run `TestSelectedLocalPathActionsRefuseImpostorsAtEitherLookup` for all five kinds.
  Test fixtures deliberately use malformed content: prove refusal is identity-based,
  not accidental decoding failure. Do not execute editor or clipboard effects.
- HYPOTHESIS B: exact lookup fixes refusal by breaking repair or rename. Run
  `TestSelectedLocalActionsKeepParseFreeRepairAndRenameSupport`, inspect both lookup
  results and contrast epic whole-stem semantics. Challenge all five detail joins;
  distinguish a semantic read refused by source mismatch from an accepted wrong path.
- HYPOTHESIS C: a matching ID/list generation allows an old path into another space,
  or selected mode borrows a detached local capability. Run
  `TestSelectedLocalPathResultsCannotCrossSameIDWorkspaceSessions`, inspect the real
  workspace envelope and `TestSelectedPathsUseExactCapabilityWithoutDecodingOrFallback`.
  Existing source-set/option-order/typed-nil tests may supplement but not replace the
  real queued action evidence.
- Compiler-valid mutation: in `localPathLookup`, switch one selected `*PathByID`
  call back to its ordinary reference method. Require the corresponding existing
  impostor test to fail at its intended effect assertion; restore and pass. If a
  sibling caller accidentally protects it, coordinate the mutation and explain why.
- Bounded cross-check: run the creation-ID collision test above and inspect where
  the successful selection is retained before adapter validation.

Both: run the relevant primary-lens race-enabled tests with `-count=3` and record
commands/results. Add one second-pass systemic challenge outside the immediate happy
path: shared helper conflating slug with identity, adapter returning empty/mutated
result evidence, wrong snapshot reused, permission/no-op ordering or an overlooked
selected consumer. State reproduced / falsified / unresolved with evidence; a
finding count is not the goal. Public-contract assertions need reachable CLI/wire
producer evidence, not only a constructed impossible result shape.

## Required hostile angles

- Selection failure versus write conflict: one budget, one timestamp; successful
  binding survives subsequent empty results but failed initial lookup binds nothing.
- Duplicate labels versus duplicate IDs; case folding; ID-shaped aliases; exact
  resolver use at final CAS as well as initial path lookup; denied preview/no-op.
- Fresh authorization after task/Thread edits; preserved unrelated edges/members;
  committed prefix and cleanup failure must not become a second mutation.
- Snapshot occurrence identity for bulk audit repair, malformed finding evidence,
  partial progress and unchanged body/unknown frontmatter preservation.
- Selected TUI entry before the first Service call, service capture after workspace
  activation, stale selection/reload/session delivery and optional pathless adapters.
- Test integrity: fixtures must have distinct physical identities, hooks must really
  run, sibling snapshots must cover bytes/modes and fake signatures must not silently
  resolve targets again. Compile failures and unused tests are not safety evidence.
- Scope: identify a concrete reachable consequence before proposing wider architecture
  work. Separate defects, coverage gaps, defense-in-depth and preference. Track a
  genuine out-of-scope issue as a recommendation; reviewers do not create taskfiles.

## Validation and restoration

The owner ran `just build`, `just test` (full race suite), `just lint`, `just docs-check`
and planning `lint --links` successfully. These are baseline evidence, not independent
review. You may inherit full-suite/build/lint evidence; independently run your focused
hostile/race/mutation checks and a restored green baseline for the affected packages.
No wire branch or repair-diagnostic wording was added, so optional-wire population and
repair-command execution requirements apply only if you discover a changed producer.

Use sandbox-local Go/build paths and scratch/caches outside its tracked/untracked
tree. Do not run local planning commands targeting the source checkout. Existing
test fixtures must stay throwaway; never exercise these races against real planning
records. Do not execute an editor/clipboard effect just to demonstrate its construction.
The helper's capture baseline is the only permitted reviewer-created commit; no other
commits or pushes. Restore mutations to that baseline, leave only the assigned audit
changed, and complete source-helper verify/transfer/receipt as required above. If
permissions/isolation/restoration/evidence fails, retain the sandbox and report
incomplete rather than an unqualified ready verdict.

## Deliverable

Keep this brief intact; fill only the report below. Findings first, consolidated by
root cause and left open for owner triage using the repository's exact finding grammar.
Each needs severity, production caller/producer trace, root cause, reachable consequence,
reproduction, before/after or effect evidence, recommendation and residual uncertainty.
Label unproven hypotheses and defense-in-depth explicitly. Do not turn checked task
criteria into evidence or mark findings fixed/tracked on the owner's behalf.

Include: ready/needs changes/incomplete and captured snapshot; verified consumer
inventory; commands run versus inherited; mutation ledger (exact edit, intended test,
observed assertion failure, killed/survived/invalid, restored result); short surviving
hypothesis ledger; and the helper verification attestation. Put the successful transfer
receipt in your handoff, not a fabricated delivery claim in the pre-transfer report.
A green baseline suite or “no issues found” without the assigned hostile evidence is
not a completed review.

## Reviewer report

Reviewer: claude (Opus 5.5), 2026-10-16. Primary lens: non-task writes and
graph/Thread retry binding, including durable-prefix boundaries, plus a queued
local-path cross-check through the real TUI caller.

**Verdict: needs changes (narrow).** Every hostile identity probe in my lens held:
no rename, disappearance, or alias reuse redirected a selected non-task,
dependency, or Thread mutation. The second pass did find one low-severity, fail-closed
regression in the shared CAS helper, demonstrated through the built CLI: epics
with id-shaped whole stems can no longer be written (L1). No other defect was
demonstrated.

**Captured snapshot.** The sandbox baseline is `b2e6f072`, an empty-delta capture over
source HEAD `792b5a88`; the source tree was clean at capture. Implementation
checkpoint `89b057f0` is an ancestor. I reviewed `2b8d05fc..b2e6f072`: 57 files,
+2463/−199. I saw no implementation drift beyond the listed task, planning, and audit
commits. I did not read the sibling Antigravity report.

### Findings

#### L1. Exact epic CAS re-resolve rejects epics whose whole stem is id-shaped · **Status:** fixed

- **Severity:** Low. It fails closed: no wrong-record write and no byte change. It is
  still a deterministic regression of an existing write capability. The command reports
  the false advice "changed on disk; retry", exits 14 (conflict), and uses up the whole
  retry budget.
- **Production trace:** CLI `epic set <epic>` →
  `Service.SetEpicFields` (`internal/core/service_epic.go:311`) → `retryEpicMutation`
  (`internal/core/entity_target.go:44`) → `FS.SetEpicFields` (`internal/store/epicstore.go:169`).
  The exact `resolveEpicMutationPath` (`internal/store/entity_target.go:52`) succeeds.
  Then `verifyUnchanged(s.resolveEpicMutationPath, …)` runs (`epicstore.go:211`). It
  rebuilds its lookup key from the path basename and keeps only the 12-character prefix
  when `splitFlatName` accepts the stem (`internal/store/cas.go:138-142`). Exact
  whole-stem matching then returns `ErrNotFound`, which becomes a conflict. `epic move`
  hits the same path (`epicstore.go:152`), and so does the TUI through
  `MoveSelectedEpic`/`SetSelectedEpicFields` (`internal/tui/entity.go:413`,
  `internal/tui/edit.go:383`).
- **Root cause:** The shared helper confuses a label with identity. `verifyUnchanged`
  assumes an epic basename is never id-led (its own comment, `cas.go:132-137`). But epic
  identity is the complete filename stem: ADR-0003 2026-07-04 amendment, `parseEpic`
  sets `ID` from the stem, and `epicCandidates` keys on the whole stem
  (`internal/store/resolve.go:126`). The former fuzzy `resolveEpicPath` still matched
  the truncated key in its prefix tier. The new exact resolver, which is correct
  elsewhere, exposes the mismatch.
- **Reachable consequence:** It affects any epic whose stem is 12 Crockford-alphabet
  characters followed by `-`, for example `202601010000-launch.md` (`id.Valid` only
  checks length and alphabet, `internal/id/id.go:158`). `epic new` always produces
  `NN-<slug>`, so only hand-named or legacy epics are affected. Such a repository passes
  `lint`, and `epic list/show/path` work; only field and status writes fail.
- **Reproduction and evidence:**
  - Temporary Go probe in the sandbox, since removed. For `202601010000-launch`,
    `SetEpicFields`, `MoveEpic` and `MoveSelectedEpic` each return
    `epic "202601010000-launch" changed on disk during update; retry: conflict`, with
    the tree unchanged. For `01-ordinary`, all three succeed.
  - Same probe against a `git archive 2b8d05fc` export in my scratchpad, outside the
    sandbox: both stems succeed (set changed bytes; move returned nil). So this is a
    regression.
  - Freshly built sandbox CLI against a throwaway `init` fixture, since deleted:
    - `epic set 202601010000-launch --priority high` exits 14; sha `0b6366698430`
      before and after.
    - `epic move 202601010000-launch retired` exits 14; bytes unchanged.
    - `epic set 01-ordinary --priority high` exits 0 (`607ec52a27c0` → `0e293720633a`),
      and `epic move 01-ordinary retired` exits 0.
    - `lint` reports "all planning entities and dependency links pass lint" (exit 0).
- **Recommendation:** Pass the CAS re-resolve key explicitly instead of deriving it from
  the basename. For example, epic writes could pass
  `func(string) (string, error) { return s.resolveEpicMutationPath(target.ID) }`, or
  `verifyUnchanged` could accept the resolver key. Do not restore the fuzzy fallback.
  Add an id-shaped-stem epic regression covering both the ordinary and selected
  entries.
- **Residual uncertainty:** I don't know how often real corpora use id-shaped epic
  stems; I did not inspect external repositories. By code inspection only, `epic edit`
  still uses the fuzzy `resolveEpicPath` CAS (`epicstore.go:249`) and so is a likely
  workaround. I did not run it, to avoid executing an editor effect.

**Resolution:** Reproduced before the fix through ordinary and selected epic
    field/lifecycle entry points. CAS now accepts the adapter's explicit
    resolver key rather than deriving identity from a path; editor diagnostic
    labels are separate. Whole-stem epic preview/write and resolver-key
    regressions pass, as do the full race suite and focused count=3 checks.
    Exact matching is preserved.

### Verified consumer inventory

All citations were checked in the sandbox at `b2e6f072`. "Exact" means
`resolveExactID` over directory candidates (`internal/store/resolve.go:288`), with no
slug, prefix or case-folding tiers.

**Non-task writes (all nine families in `TestNonTaskRetriesRetainSelectedSourceIdentity`, `internal/store/entity_target_test.go:34`):**

| Family | Service entry → wrapper | Adapter lookup / CAS |
| --- | --- | --- |
| audit move | `MoveAudit` `service_audit.go:122` → `retryAuditMutation` `entity_target.go:8` | `resolveAuditMutationTarget` `store/entity_target.go:38` (`validEntityID` + exact `resolveAuditPath` `auditstore.go:183`); CAS `auditstore.go:166` exact |
| audit append | `AppendAuditBody` `service_audit.go:154` | `body.go:92`; CAS `body.go:135` exact |
| finding edit | `EditFinding` `finding.go:247` | `TransformAuditBody` `body.go:192`; CAS `body.go:236` exact |
| finding create | `NewFinding` `finding_create.go:43` | same `TransformAuditBody` path |
| header repair | `FixFindingHeaders` `finding.go:299`. It checks `requireSourceID` across the whole snapshot before any prefix (`:309-313`), then uses the snapshot occurrence's `Source.ID` (`:325`); no reference wrapper | same exact path |
| research fields | `SetResearchFields` `service_research.go:174` → `retryResearchMutation` `entity_target.go:26` | `resolveResearchMutationTarget` `store/entity_target.go:45` exact; CAS `researchstore.go:198` exact |
| research append | `AppendResearchBody` `service_research.go:240` | `researchstore.go:243`; CAS `researchstore.go:279` exact |
| epic fields | `SetEpicFields` `service_epic.go:311` → `retryEpicMutation` `entity_target.go:44` | `resolveEpicMutationPath` `store/entity_target.go:52` whole-stem exact; CAS `epicstore.go:211` (**L1**) |
| epic move | `MoveEpic` `service_epic.go:276` | `epicstore.go:104`; CAS `epicstore.go:152` (**L1**) |

- Initial selection is parse-free. `FS.Resolve{Audit,Research,Epic}Source`
  (`store/entity_target.go:14/22/30`) uses the ordinary reference resolver and derives
  the ID from the filename (`entity_read.go:26/38`; whole stem for epics).
  `requireSourceID` (`entity_read.go:120`) refuses empty evidence before binding.
- Selected TUI entries are `MoveSelectedAudit` (`service_audit.go:131`, caller
  `tui/entity.go:400`), `MoveSelectedEpic` (`service_epic.go:287`, `tui/entity.go:413`)
  and `SetSelectedEpicFields` (`service_epic.go:323`, `tui/edit.go:383`). Each checks
  `requireSourceID` before the adapter and repeats no initial matching. A grep for TUI
  non-test calls to any ordinary non-task, dependency or Thread entry point returns
  nothing (exit 1). The CLI calls no selected entry.
- Outside the retry contract: the `EditAudit/EditResearch/EditEpic` editor flows are
  single-attempt (`service_audit.go:145`, `service_research.go:233`,
  `service_epic.go:402`).

**Graph and Thread retry binding:**

- **Dependencies:** `runDependencyMutation` (`dependency_operations.go:139`). Binding
  happens at `:168-179` from the planner's `details`, not from `result.Plan`, and only
  after a successful planner call. The bound `planSelectedDependencyEdges` (`:253`) uses
  exact `graph.Task` (`dependency_graph.go:919`) and recomputes from the current
  `DependsOn`. The loop stops on preview, non-conflict, `AppliedTaskIDs`, or an
  exhausted budget (`:162`). Migration is excluded by the operation guard.
- **Graph adapter:** `MutateTaskGraph` loads a fresh graph on every call
  (`graphmutation.go:53`) and calls the planner under the planner sentinel (`:61`).
  Materialization uses exact `resolvePath` (`fsstore.go:373`). A whole-snapshot verify
  (`:83-89`) precedes per-file exact CAS (`:97`) and prefix recording (`:103`).
- **Thread creation:** `runThreadCreationMutation` (`service_thread.go:86`) binds at
  planner success (`:92-107`). The new Thread ID is minted once, outside the planner
  (`:52`). `ValidateThreadCreationPlan` (`thread_creation.go:127`) raises ID-collision
  `ErrConflict` (`:146/:150`) before the adapter assigns `result.Plan`
  (`threadcreation.go:70-73`). Missing members fail exactly with `ErrValidation`
  (`:153-156`).
- **Thread mutation:** `runThreadMutation` (`service_thread.go:189`) binds at
  `:195-208`. The plan contains only `ThreadID/Operation/TaskIDs`. Inside the adapter,
  `ValidateThreadMutationPlan` (`thread_mutation.go:208`) re-derives policy, membership
  and analysis against each fresh snapshot (`threadmutation.go:62`) using exact
  `threadByID` (`thread_mutation.go:378`). `threadMutationSource` matches exact
  `Source.ID`, and materialize/CAS use `resolveThreadPathExact`
  (`threadmutation.go:158`, `:104`; `paths.go:87`). `!result.Committed` stops retries
  (`service_thread.go:110`, `:212`).
- **Brief exclusions, rechecked:** Bulk apply stops on `Committed/Complete`
  (`service_thread_apply.go:69`). Repair re-plans from the original request on each
  attempt (`dependency_repair.go:632-645`); `normalizeRepairEditSource` (`:817`) is
  exact by ID, slug or local path with no fuzzy tiers. See hypothesis H-a.

**Local paths (shared contract; cross-check depth only):**

- Optional ports with ByID methods: `core/store.go` (task, Thread, epic, audit and
  research path sources). Service methods: `selected_entity_path.go:11/24/37/50/63`.
- FS implementations: `paths.go:52-100`. Task uses `resolveTaskMutationTarget`
  (`fsstore.go:344`); epic uses whole-stem exact; audit and research use exact; Thread
  uses `resolveThreadPathExact`.
- TUI callers: `localPathLookup` (`tui/commands.go:40-48`); detail joins at
  `commands.go:137/217/286/388` and `thread_projection.go:67`.
- `Model.Update` handles `localPathResultMsg` at `model.go:392-410`. A yank returns
  the `copyToClipboard` command (`model.go:1583-1589`) and performs no I/O in `Update`.

### Commands run versus inherited

**Inherited, not rerun:** the owner's `just build`, `just test` (full race), `just lint`,
`just docs-check`, and planning `lint --links`. I ran no planning command against the
source checkout.

**Run by me in `$SANDBOX`** (GOCACHE in session scratchpad; go1.27.2 darwin/arm64):

1. `go test -race -count=3 ./internal/{store,core,tui}/ -run '<the 11 named tests>'`:
   all 11 passed 3/3.
2. `go test -race -count=3 ./internal/core/ ./internal/tui/ -run 'TestEntityMutationsShareSelectionAndWriteRetryBudget|TestSelectedNonTaskActionsRefuseEmptyEvidenceBeforeAdapter|TestEntityMutationSelectionRequiresSourceEvidence|TestQueuedSelectedNonTaskActionsDoNotReselect'`:
   passed 3/3.
3. Cross-check:
   `go test -race -count=3 ./internal/tui/ -run 'TestSelectedLocalPathActionsRefuseImpostorsAtEitherLookup/kind=0/'`.
   All 16 task subtests (E/Y × initial/confirmation × missing/impostor/case-folded/duplicate)
   passed 3/3. They drive real `Model.Update` key presses into the queued lookup and
   confirmation. I executed no effect.
4. Restored baseline on a clean tree at `b2e6f072`:
   `go test -race -count=1 ./internal/core/... ./internal/store/... ./internal/tui/... ./internal/cli/... ./internal/wire/...`.
   All ok.
5. Mutation runs M1–M7 (ledger below), each preceded by `go build ./...` and followed
   by `git checkout -- <file>` with a clean-tree check.
6. L1 probes: a sandbox Go probe, the same probe on a `git archive 2b8d05fc` export, and
   the built CLI (`go build -o <scratchpad>/bin/tskflwctl ./cmd/tskflwctl`) against a
   throwaway fixture.
7. Discrimination check: I ported
   `TestAuditHeaderRepairUsesOccurrenceIdentityWithDuplicateLabels` to the base export.
   It fails there with `"same-label" matches 2 audits … ambiguous match`, so the test
   discriminates.

**Harness note:** My first package run put `GOTMPDIR` under the session directory. Its
parent contains a `tasks/` folder that CLI discovery treats as a planning root, so five
`internal/cli` startup and discovery tests failed (for example
`TestRoot_NotAPlanningRepo`). All five pass with the default `GOTMPDIR`, so these are
not attributable to the change.

**Not run:** `just lint`, `just docs-check`, and full `just test`. None of my changes
remain to lint.

### Mutation ledger

Every edit compiled; there were no invalid probes. Every file was restored with
`git checkout --` and a clean `git status` was confirmed.

| # | Exact edit | Intended test | Observed | Result |
| --- | --- | --- | --- | --- |
| M1 | `dependency_operations.go:176` `planner = func(graph …` → `_ = func(graph …` (endpoint binding removed) | `TestGraphRetriesRetainBindingAcrossUnplannedConflicts/dependency`, `TestDependencyRetriesPinBothEndpointsAndRecomputeCurrentEdges` | core: `dependency binding drift: {TaskWrites:[{TaskID:9ss0kg30xkr3 …` (impostor). store: add/remove × dependent/prerequisite rename/missing failed with `endpoint identity drift` or `missing endpoint was retargeted … AppliedTaskIDs:[…] err=<nil>`; at least 7 of 8 shown before output truncation | Killed; restored, green |
| M2 | `service_thread.go:196` `if selected == nil {` → `if selected == nil \|\| selected != nil {` (mutation planner forgets selection on every callback) | `TestGraphRetriesRetainBindingAcrossUnplannedConflicts/{membership,start}`, `TestThreadRetriesKeepSelectedThreadAndMemberIdentities` | core: `Thread binding drift: {ThreadID:kbssa9dcgv1g …}`. store: `start/rename` started the impostor (`Description:impostor`); `start/missing` and `add/missing` gave `missing endpoint was not refused: err=<nil> unchanged=false`; `add/rename` member drift; `remove/rename` stale membership replay | Killed; restored |
| M3 | `service_thread.go:93`, same edit (creation planner forgets selection) | `TestThreadCreationBindsMembersBeforeCollisionValidation` | `collision retry changed member selection: … Tasks:[q1am50pntm9a]` (impostor). Also failed: core `/creation`, `create/rename`, `create/missing` | Killed; restored |
| M4 | Coordinated: the creation planner never binds; the retry loop binds only from a validated `result.Plan` when present | `TestThreadCreationBindsMembersBeforeCollisionValidation` | Killed by the collision test and core `/creation`. All 8 `TestThreadRetriesKeepSelectedThreadAndMemberIdentities` subtests passed, including the post-validation `create/*` verify-hook conflicts. | Killed by intended test; restored |
| M5 | `entity_target.go:11/29/47` `if target.ID == "" {` → `if target.ID == "" \|\| target.ID != "" {` | `TestNonTaskRetriesRetainSelectedSourceIdentity` | 16/18 failed: `selected source did not receive the write` (renamed) and `missing selected identity was reselected: err=<nil> unchanged=false`. `audit_headers/*` passed as expected; it uses the snapshot ID, not the wrapper. | Killed; restored |
| M6 | Coordinated store: `store/entity_target.go:42` `return s.resolveAuditPath(target.ID)` → `return s.resolveAudit(target.ID)` (fuzzy) | store/core write tests | Survived in `internal/core` and `internal/store` (both packages fully green). Killed in `internal/tui` by `TestQueuedSelectedNonTaskActionsDoNotReselect/audit_move/impostor` (real FS store) and 8 audit (`kind=3`) local-path impostor/case-folded subtests, which share the helper | Killed cross-package; restored |
| M7 | Cross-check: `tui/commands.go:40` `svc.TaskPathByID(id)` → `svc.TaskPath(id)` | `TestSelectedLocalPathActionsRefuseImpostorsAtEitherLookup/kind=0` | 8 impostor/case-folded subtests: `selected local action accepted another source: effect=true` (the returned clipboard/editor command was not executed). Missing/duplicate still refused | Killed; restored |

**Why the collision case differs from a post-validation CAS conflict (M4):** In a
verify-hook or CAS conflict, the adapter has already assigned `result.Plan`, so binding
from the validated receipt would be enough. A creation-ID collision conflicts inside
`ValidateThreadCreationPlan`, before `result.Plan` exists (`threadcreation.go:70-73`).
Only planner-time capture survives that. M4 shows that only
`TestThreadCreationBindsMembersBeforeCollisionValidation` and the core fake pin this
case.

### Second-pass systemic challenges

- **A shared helper confusing a label with identity:** Reproduced. This is L1: the
  basename-derived key in `verifyUnchanged`.
- **Selection failure versus write conflict:** Falsified as a defect.
  - One budget: `retryOnConflict` (`retry.go:62`) wraps both selection and write. The
    core budget test passed 3/3.
  - One timestamp: epic and research timestamps are computed before the loop, and the
    test asserts `clocks == 1`.
  - Failed selection binds nothing: `target.ID`, the dependency planner and the Thread
    `selected` are set only after success.
  - An empty result after selection does not erase binding: shown by
    `retryEndpointStore` call 3 and by M2–M4.
- **Fresh authorization and current state:** Falsified as a defect.
  - Thread policy, membership and analysis are recomputed by the adapter on every
    attempt.
  - The dependency planner recomputes from current edges; the store test keeps a
    concurrent `extra` edge, and M1 shows the impostor drift that binding prevents.
- **Committed results and durable prefixes:** Falsified as a defect.
  - Thread loops require `!result.Committed`.
  - Dependency add/remove plans write exactly one dependent, and `AppliedTaskIDs` stops
    the loop.
  - Preview is never retried in any of the three loops.
- **Permission/no-op ordering:** Noted, not a new defect.
  - The new wrappers resolve the human reference before the adapter's
    `authorizeMutation`. On a read-only policy, an unknown reference therefore reports
    not-found rather than a permission refusal.
  - No write occurs, and this matches the merged `retryTaskMutation` precedent
    (`service_task.go:512-524`).
  - Duplicate-ID refusal precedes callback, no-op and preview
    (`TestSelectedAuditPortsRefuseDuplicateIdentityBeforeCallbackOrNoop`).
- **Test integrity:** Held.
  - `SnapshotTree` compares bytes, modes and symlink targets
    (`internal/testutil/snapshot.go:11-58`).
  - Hooks assert that they fired, and fixtures use distinct `testutil.TaskID` seeds.
  - The core fakes do not re-resolve references.

### Surviving hypothesis ledger

- **H-a (unresolved; out of scope by the brief):**
  - `RepairTaskGraph` re-normalizes the original request on each attempt
    (`dependency_repair.go:633-635`).
  - A CLI selector that is neither path- nor ID-shaped becomes an exact `TaskSlug`
    (`internal/cli/task_dependency_repair.go:139-146`), matched by label
    (`dependency_repair.go:827-833`).
  - So a slug swap between attempts could select a different record. However,
    `authorizeRequestedRepair` (`:753-771`) still requires an identical repairable
    defect on that record, and the operation is removal-only.
  - Not demonstrated. If repair ever joins this contract, consider binding the source
    ID after the first successful normalization.
- **H-b (falsified):** No changed selected-write CAS site falls back to fuzzy
  resolution. See the CAS sites cited in the inventory.
- **H-c (falsified):** Bound Thread planners do not replay derived state.
- **H-d (coverage observation, not a defect):** The store's exact audit
  selected-write resolver is pinned only by TUI-package tests that use the real FS store
  (M6). A store-level test with an ID-shaped alias would localize a failure.

### Isolation attestation

Created with `"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create --source
/Users/andyeschbacher/git/andy-esch/taskflow --deliverable <AUDIT_REL> --print-path`,
which was run from the source root. All inspection, builds, tests, mutations and report
edits happened in the sandbox. Base-commit comparisons used a `git archive` export in
the session scratchpad, which has no Git metadata. The CLI fixture was a throwaway
temporary directory, since deleted. All probes were restored, and only the assigned
audit differs. Output of `verify` before transfer:

```text
sandbox_path=/private/tmp/isolated-review.L9oB5P
git_dir=/private/tmp/isolated-review.L9oB5P/.git
baseline_commit=b2e6f072212f977a9bcacda93bf2f5419d0e1f56
source_blob=aa6f6ff47113b9869c966f43448cfdfbe5cc59c9
source_fingerprint=e663a99ac59d3b7c72285d28fdfa53dbe0ab4afc
deliverable=planning/audits/6gm8vm7aq7zt-2026-10-16-selected-identity-safety-implementation-claude.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner triage (2026-10-16)

Delivery confirmed before owner edits with the source helper: sandbox
`/private/tmp/isolated-review.L9oB5P`, `transfer=succeeded`, `delivery=confirmed`,
transferred blob `1726c1a82cc303b8c6d0715bbb41e75a3a10875f`, at
`2026-10-16T12:33:17Z`. The review's production traces and mutation distinctions
match the captured implementation; its narrow needs-changes verdict was actionable.

L1 reproduced in four ordinary/selected field/lifecycle write cases, while previews
and ordinary stems passed. The shared CAS helper now receives an explicit resolver
key; no filename-shape inference or fuzzy epic fallback remains. Interactive editor
callers supply their established source key separately from diagnostic labels.
[Whole-stem write tests](../../internal/store/epic_selected_cas_test.go) and
[resolver-key/label tests](../../internal/store/cas_test.go) pin the fix.

H-d was useful localized coverage advice rather than a product defect. Added
`TestSelectedAuditPortsRefuseIDShapedImpostors` directly in the store package so the
shared exact audit resolver is no longer pinned only by cross-package TUI tests.
Owner temporarily restored fuzzy `resolveAudit` in that helper: the new test compiled
and failed at the intended impostor no-op assertion (`err=<nil>`). Restoring exact
lookup passed with `-race -count=3`.

H-a was initially an unproven, explicitly out-of-scope hypothesis, not a demonstrated
review finding. Owner reproduced it through a real-store pre-verify slug swap and
promoted it to H1 below, tracked in
[the explicit graph-repair follow-up](../tasks/6gm9ny98zyf2-pin-explicit-graph-repair-source-selections-across-retries.md).
That implementation is outside the current three-task PR and is next-up in the
same safety Thread. The original removal-only/strict-improvement gates still hold,
but do not establish that the selected physical task remained the same.

Final owner gates after the CAS fix: `just build`, `just test` (full race suite),
`just lint`, `just docs-check`, focused core/store/TUI `-race -count=3` checks,
planning `lint --links` and both audit lint calls pass.

## Findings

#### H1. Explicit graph-repair retries can repair another task after a slug swap · **Status:** tracked by 6gm9ny98zyf2

**File:** internal/core/dependency_repair.go:632 | **Component:** core,store
**Effort:** M · **Urgency:** soon

Owner follow-up to the review's surviving hypothesis H-a: a compiler-valid real-store race probe seeded distinct A/alpha and B/bravo with the same invalid dependency declaration. The pre-verify hook renamed A to A-renamed and B to B-alpha without changing either ID or contents. The initial attempt detected source drift, but RepairTaskGraph retried the original TaskSlug selector, returned nil, and applied the removal to B instead of A. The remaining graph was strictly improved, so removal-only and defect authorization did not prevent wrong-source selection. Full reproduction and bounded scope are recorded in task 6gm9ny98zyf2; the temporary reproduction probe was removed rather than retained as passing safety coverage.

**Recommendation:** Retain first-selected physical source evidence for explicit repair selectors while recomputing fresh defects and policy; preserve duplicate-ID disambiguation and separate auto-cleanup semantics.

**Resolution:** Owner confirmed H-a with a real-store alias-swap probe. Task
    6gm9ny98zyf2 is next-up in Finish selected identity safety; implementation
    stays outside the current three-task PR.
