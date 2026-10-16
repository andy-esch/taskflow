---
schema: 1
id: 6gm8vm7j1z6e
bucket: open
area: selected-identity-safety-implementation-antigravity
date: "2026-10-16"
---
# Audit: Selected identity safety finish — antigravity — 2026-10-16

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
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

Pending independent review.
