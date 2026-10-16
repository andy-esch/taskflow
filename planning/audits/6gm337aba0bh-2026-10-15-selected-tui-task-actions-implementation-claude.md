---
schema: 1
id: 6gm337aba0bh
bucket: closed
area: selected-tui-task-actions-implementation-claude
date: "2026-10-15"
updated_at: "2026-10-15"
---
# Audit: Selected TUI task actions implementation — claude — 2026-10-15

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
AUDIT_REL="planning/audits/6gm337aba0bh-2026-10-15-selected-tui-task-actions-implementation-claude.md"
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

Determine whether an already-selected TUI task remains the action target through
queued execution and Service retries, without weakening validation, persistence
authorization, lifecycle policy, repair behavior, or asynchronous result handling.
The crucial interval starts **before the first Service call**, not just after a
resolver has chosen a record. Challenge both the implementation and its evidence.

Claude is the sole reviewer for this round. Perform a focused boundary review plus
a separate systemic pass; do not implement fixes or seek a quota of findings.

## Review target

- Source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`.
- Branch: `fix/selected-tui-task-actions`; base HEAD:
  `dfd29d0b496dc5c4cd02673cfc9720512f3677e8`.
- The implementation is **uncommitted**. Inspect the captured staged, unstaged,
  untracked and deleted state, not HEAD alone. In particular, both new
  `selected_task_action_test.go` files are untracked at handoff.
- Implementing task: [selected TUI task identities](../tasks/6gkt3fq0pv26-preserve-selected-tui-task-identities-through-service-actions.md).
  Its checked ACs record owner evidence, not independent approval.
- Owning contracts: [planning-data guide](../../docs/architecture/planning-data.md),
  [TUI guide](../../docs/architecture/tui.md), `TaskMutationTarget` in
  [core ports](../../internal/core/store.go), and the repository `AGENTS.md`.

Verified entry points to re-check in the sandbox:

- [Service task actions](../../internal/core/service_task.go):
  `SetSelectedTaskFields`, `MoveSelectedTask`, `DeferSelectedTask`, shared
  `prepareTaskFieldUpdates`, and their reference-based counterparts.
- [TUI field commands](../../internal/tui/edit.go): `setFieldCmd`, `unsetFieldCmd`;
  [entity commands](../../internal/tui/entity.go): `moveTask`, `deferTaskCmd`.
- [Core regressions](../../internal/core/selected_task_action_test.go) and
  [TUI regressions](../../internal/tui/selected_task_action_test.go), including
  `TestQueuedSelectedTaskActionsRetainIdentityBeforeExecution`,
  `TestSelectedTaskUIFormsRetainIdentityAndResultEnvelopes`, and
  `TestSelectedTaskCommandsPreserveCommittedFailureReceipt`.
- [CLI repair regressions](../../internal/cli/task_set_repair_test.go), particularly
  `TestTaskSetRepairsMalformedTypedFields` and
  `TestTaskSetPreservesPreexistingCorruptionAttribution`.
- Existing action, model, session and identity tests: inspect the actual action
  registry, `scopeMutation`, session envelopes and result reducers. Do not assume
  direct helper tests establish that edit forms, command-bar actions, menus and
  palettes all choose the same route. Verify the changed `testTaskRef` fixtures
  still reach lifecycle policy instead of failing earlier on a bogus identity.

Produce a bounded **consumer inventory**: selected row/source identity → actual UI
submission → captured Service and target → retry planner/persistence → result
envelopes/reload. Include initial CLI reference callers as a compatibility check.
Follow relevant unchanged callees; do not broaden into a full application audit.

Known neighboring work, not silently claimed fixed here:

- [Exact selected E/Y path actions](../tasks/6gm30jat0xsy-keep-selected-tui-path-actions-bound-to-exact-identities.md):
  fuzzy local-path selection can accept an impostor after deletion. Owner constructed
  effects but did not run an editor or clipboard operation. Detail result identity
  is separately checked; fuzzy lookup alone does not prove wrong-detail acceptance.
- [Non-task selection/retry pinning](../tasks/6gkt3fphx1ts-pin-non-task-mutation-retries-to-selected-source-identities.md):
  includes a confirmed selected-audit action exposure and inventory of other callers.
- [Graph/Thread retry endpoints](../tasks/6gkt3fpsfdd7-keep-graph-and-thread-retry-endpoints-fixed-during-replanning.md):
  separate identity-preserving fresh-replanning work.

Do not refile those known defects. Challenge their scope boundary only with a
concrete consequence for this task. No universal entity-action API, incarnation
token, wire/schema change, or command cancellation is being proposed here.

## Intended contract to challenge

1. Adapter-established source ID is authoritative. Selected task actions must never
   reinterpret it as a human slug, declaration, path or fresh fuzzy reference,
   including on their first attempt. Renames preserve identity; missing/duplicate
   targets and ID-shaped/case-folded impostor slugs cannot redirect writes.
2. The four TUI commands use explicit selected Service APIs. Ordinary CLI calls
   retain their existing ergonomic reference selection; an ID-shaped human slug
   is not globally forbidden or reclassified as a strict selected target.
3. Field APIs share coercion, unset/force rules, graph-owned/status rejection,
   epic validation, caller-map ownership and timestamp/retry behavior. Repairing
   malformed typed fields remains possible through the existing CLI path.
4. Lifecycle validation and exact graph identity remain inside guarded mutation;
   no pre-lock scan substitutes for authorization. Dependency/AC gates, override
   policy, defer dates, impacts and committed/partial receipts are preserved.
5. A command captures its originating Service as well as its selected identity.
   Selection/load/session envelopes protect result application. An already-issued
   command may finish in its old space; this is not cancellation. It must not
   silently mutate the newly displayed space or apply a stale editor/flash result.
6. I/O stays in `tea.Cmd`. Existing read-only capabilities and persistence/CAS
   ownership remain intact. No extra filesystem knowledge belongs in TUI or core.

These are hypotheses to falsify, not a checklist whose wording establishes truth.

## Mandatory evidence floor

Independently execute the following bounded checks. If blocked, say which evidence
is missing and return an incomplete verdict rather than substituting a green suite.

1. **Reachable queued-action probe.** Drive at least one actual edit-form submission
   and one actual lifecycle submission through the production UI route, constructing
   the command before the selected file is deleted and another source ID takes its
   ID-shaped slug (include case folding). Check target identity, error/receipt,
   and before/after source bytes for both intended task and sibling. Include a
   successful rename/old-alias-reuse control. Use only disposable planning spaces.
2. **Two compiler-valid regression mutations, separately restored.** In a field
   command replace `SetSelectedTaskFields(typedTarget, ...)` with
   `SetFields(ref.key, ...)`; in `moveTask` replace `MoveSelectedTask(typedTarget, ...)`
   with `Move(ref.key, ...)` (or make the analogous defer mutation). Keep arguments
   and compilation valid. Require the intended impostor assertions in
   `TestQueuedSelectedTaskActionsRetainIdentityBeforeExecution` or the actual-form
   regression to fail. Run a compile gate, record the exact edit/test/failure,
   restore to the captured baseline and prove the original test green. Compilation
   errors, unrelated fixture failures and tests that never reach the write boundary
   are not killed mutations. Investigate any survivor before drawing a conclusion.
3. **Field compatibility through a real producer.** Build a fresh sandbox binary
   and independently exercise malformed typed-field repair with write and dry-run,
   checking preserved unrelated corruption/bytes. Also challenge selected set/unset
   and force rejection for owned fields. Existing repair tests can supplement, not
   replace, explaining the Service→store path. Prior strict-target review found that
   an apparently safer typed read silently blocked repair despite green lower-level
   tests; this is the reason for this explicit compatibility check.
4. **Non-default policy and result behavior.** Independently exercise a blocked
   selected lifecycle transition or completion guard, and trace committed-error
   propagation/retry termination. Check stale selection/load/session results and
   same IDs in different spaces, including the captured Service. Distinguish a
   real-store policy probe from a port-double receipt test: the existing committed
   cleanup-conflict fixture does not inject a filesystem cleanup failure. A
   cancellation policy change is not an acceptable substitute for preserving intent.

Run the focused core/TUI/CLI suite independently. No optional wire fields or repair
recommendation text changes are introduced here, so the generic preamble's wire
and emitted-advice checks apply only if your verified inventory reveals such a path.

## Required hostile angles

- Is identity ever lost before entering the new methods, through an alternate
  registry delegate, editor callback or a helper that uses a display label as key?
  Is a refusal caused by intended exact selection or merely malformed test data?
- Do ref-based and selected field paths diverge for unknown fields, force, unset,
  epic canonicalization/detachment, caller-owned maps, metadata or dry-run? Are
  blank evidence and invalid dates refused before adapter access without inventing
  unnecessary ID syntax restrictions for portable field adapters?
- Can contention replay a newly resolved target or retry a committed operation?
  Are lifecycle plans revalidated against fresh state without authorizing a sibling?
  Verify actual error sentinels and receipt fields rather than guessed names.
- Can stale successes erase another selection's editor or warning, or fail to
  refresh a legitimately changed source? Does workspace switching accidentally
  bind an old target to a newly current Service?
- **Separate systemic pass:** inspect the common field helper, lifecycle retry
  engine, entity/action infrastructure and test fixtures for broad blind spots.
  Try a second consumer or real producer before treating a synthetic malformed-port
  shape as a product bug. Distinguish safety promised by core from adapter contracts;
  avoid demanding redundant unguarded scans that weaken the architecture.
- Check source comments and focused guides against behavior. Do they overclaim
  cross-space authority, cancellation, all-entity protection or local-path safety?

## Validation and restoration

Use the mandatory source-helper sandbox protocol above. Keep build outputs, probes
that are not restored, fixture directories and caches outside the sandbox tree.
For example, after successful sandbox creation:

```sh
cd "${SANDBOX:?create failed: stop and report the blocker}" || exit
REVIEW_SCRATCH="$(mktemp -d /private/tmp/taskflow-claude-selected-actions.XXXXXX)"
export GOCACHE="$REVIEW_SCRATCH/go-cache"
go test -race ./internal/core ./internal/tui ./internal/cli -count=1
go build -o "$REVIEW_SCRATCH/tskflwctl" ./cmd/tskflwctl
"$REVIEW_SCRATCH/tskflwctl" -C . audit finding new --help
```

Owner reports full race tests, lint, build, docs-check, planning lint with links and
focused repeated tests green. Those are inherited evidence unless you rerun them.
Owner reported eight impostor failures before the fix; independently establish the
bounded mutation evidence rather than repeating that report as your measurement.
No interactive visual smoke, new real-process concurrency stress, or real cleanup
fault injection is claimed. Do not run a real editor or clipboard side effect.

Restore every implementation/probe change to the **captured sandbox baseline**,
not the source HEAD, and remove only your own disposable artifacts. Preserve the
brief and change only this audit for transfer. Run fresh-binary `audit lint` on
this audit and planning `lint --links`; distinguish any baseline issue from new
report problems. Use the source helper's verify/transfer protocol, retaining the
workspace and printed transfer receipt. Do not commit implementation or push.

## Deliverable

Fill only `## Reviewer report`; keep the brief intact. Report findings first with
canonical open-status headings, root cause, reachable producer, exact reproduction,
consequence, source evidence, recommendation and uncertainty. Prefer the sandbox's
fresh `audit finding` verbs when available. Leave all dispositions open for owner
triage; do not implement fixes, create unrelated tasks or assert a settled verdict.

Include the exact snapshot/baseline, verified consumer inventory, independently run
versus inherited checks, bounded hostile-probe/mutation ledger (including restored
green results), and surviving hypotheses classified tested-negative/unresolved.
Coverage gaps are normally Low; preferences and already-tracked defects are not
new product failures. No-findings requires the evidence floor, not a finding quota.

Include the source helper's verification attestation with `transfer=pending` before
final guarded transfer; put the helper's printed delivery receipt in your handoff
message. Missing execution or transfer evidence means incomplete. Do not edit the
source audit directly or claim that verification alone proves delivery.

## Reviewer report

Reviewer: claude (sole reviewer), 2026-10-15. Independent sandbox; both passes complete.

**Findings first.** There is one open finding, and no product defect was reproduced in the captured implementation.

- **L1** (coverage gap): open task forms and menus bind to the *current* `m.svc` when submitted. Only `closeTransientUI` stops a form opened in one checkout from writing a same-ID task in a newly activated checkout. That line is untested: a compiler-valid mutation removing it leaves the whole TUI suite green, and a reviewer probe then shows the cross-checkout write. The full block is under *Findings* below.

The mandatory evidence floor is met. Each of the four selected commands was mutated to its reference-based twin, and each mutation was killed by the intended impostor assertions, then restored green. Real UI routes, real-store policy and committed-receipt probes, and fresh-binary repair and compatibility probes all behaved as the brief's contract describes. Dispositions are left open for owner triage.

### Snapshot and baseline

- Source `/Users/andyeschbacher/git/andy-esch/taskflow`, branch `fix/selected-tui-task-actions`, HEAD `dfd29d0b496dc5c4cd02673cfc9720512f3677e8`.
- Sandbox `/private/tmp/isolated-review.mJ36Ii`: independent clone with its own `.git`; push URL is `isolated-review-push-disabled://source-is-read-only`.
- Captured baseline commit `3b7fc92452c4dce0300d385949f6c98fdfa40e59` ("chore: capture isolated review baseline"). It contains the dirty and untracked overlay (15 files, +1104/−41), including both untracked `selected_task_action_test.go` files and this brief.
- Review scratch (outside the sandbox): `/private/tmp/taskflow-claude-selected-actions.Lfp0yB`, holding `GOCACHE`, `GOTMPDIR`, the fresh binary, disposable planning space `space-repair`, and an isolated `HOME`/`XDG_CONFIG_HOME` so no machine registry was touched.

### Verified consumer inventory

Each step below was checked in the sandbox at the cited path and line.

| Stage | Verified path |
| --- | --- |
| Selected row → identity | Task rows key on adapter `record.Source.ID` (`internal/tui/commands.go:81`, `:102`). `taskItem.ref()` returns `entityRef{key: sourceID, label: slug}` (`internal/tui/item.go:56`). Lifecycle reads go through `selectedLifecycle()`; edits use `selectedTask()` plus `selectedRef()`. Both refuse while `identityInvalid` or `coherentGen != loadGen` (`internal/tui/model.go:1227-1265`). |
| UI submission: edit form | `e` (`model.go:894`) → `edit.open(selectedRef)` with `sourceGen` → `handleEditKey` re-checks generation, coherence and ref (`internal/tui/edit.go:255-263`) → `submitEdit` (`edit.go:323-340`) → `setFieldCmd` / `unsetFieldCmd` (`edit.go:357`, `:369`). |
| UI submission: lifecycle | Menu `m` (`model.go:883`) → `handleActionKey` gates (`model.go:1127-1135`) → Enter (`:1195`), destructive `y` confirm (`:1166`), or defer date (`:1146`). Command bar (`internal/tui/command_dispatch.go:36`) and palette (`:248`) → `beginTransition` with a `ref == selectedRef` gate (`model.go:1207-1222`). For tasks the registry delegate is `applyMove: moveTask` (`internal/tui/entity.go:450`). `deferTaskCmd` is at `entity.go:380`. |
| Other TUI writers | A production grep of `svc.<Method>(` under `internal/tui` finds exactly four task writers: two `SetSelectedTaskFields`, one `MoveSelectedTask`, one `DeferSelectedTask`. The remaining writers are epic/audit, tracked in `6gkt3fphx1ts`. Residual `svc.SetFields` / `svc.Move` calls are test setup only (`identity_test.go:346`, `dashboard_test.go:442`). The Atlas owns its keys (`model.go:820`) and has no action route; its work-row jump only stamps a restore key (`internal/tui/atlas.go:424`). |
| Captured Service and target | Each command closure captures its `svc` argument at construction and passes `core.TaskMutationTarget{ID: ref.key}`. |
| Service | `SetSelectedTaskFields` (`internal/core/service_task.go:772`): `requireSourceID` → shared `prepareTaskFieldUpdates` (`:787`) → `retryOnConflict` with a fixed target (`internal/core/retry.go:62`). `MoveSelectedTask` (`:600`) and `DeferSelectedTask` (`:634`) use constant planners, run through `runTaskLifecycleMutation` (`:648`); its retry stops on `Committed`. |
| Persistence | Fields: `FS.SetFields` (`internal/store/fsstore.go:257`) → `resolveTaskMutationTarget` (`:344`): `validEntityID`, then `resolvePath` → `resolveExactID` (`internal/store/resolve.go:288`), which is case-sensitive and never matches slugs. CAS goes through `verifyUnchanged` (`internal/store/cas.go:123`). Lifecycle: `FS.MutateTaskLifecycle` (`internal/store/lifecyclemutation.go:20`) loads the graph under the guard → `validateExistingLifecyclePlan` (`internal/core/task_lifecycle.go:307`: `id.Valid` plus exact `graph.Task` map lookup) → snapshot and target CAS. |
| Result envelopes and reload | `scopeMutation` (`internal/tui/action.go:115`) → `scopeSession` (`internal/tui/session.go:20`) → `Update` session gate (`model.go:283-290`) → stale `mutationResultMsg` triggers a reload only, with no flash or editor side effects (`model.go:448-462`). |
| Initial CLI references (compatibility) | `SetFields` → `retryTaskMutation` → `selectTaskMutationTarget` (`service_task.go:498`) → fuzzy `ResolveTaskSource` (`internal/store/entity_read.go:76`). `Move` / `DeferTask` → `graph.ResolveTaskID` (`internal/core/dependency_graph.go:937`). Unchanged. The fresh binary accepted the fuzzy ref `ALP` and the ID-shaped human slug `k7m2p9q4r6s8`. |

### Checks: independently run vs inherited

**Independently run in the sandbox:**
- `go test -race ./internal/core ./internal/tui ./internal/cli -count=1`: ok (1.484s, 10.436s, 8.081s).
- Full `go test -race ./... -count=1`: rc=0.
- `golangci-lint` 2.14.0 `run ./...`: 0 issues.
- `go build -o <scratch>/tskflwctl ./cmd/tskflwctl`; `audit finding new --help` confirmed.
- Fresh-binary `lint --links` and `audit lint 6gm337aba0bh`: both pass on baseline; re-run after this report (see the end of this report).
- The brief's named CLI tests `TestTaskSetRepairsMalformedTypedFields` (`internal/cli/task_set_repair_test.go:15`) and `TestTaskSetPreservesPreexistingCorruptionAttribution` (`:71`) passed inside the CLI suite.

**Inherited, not rerun:** `just docs` / `docs-check` (no command metadata changed), `just build` with release ldflags, diff hygiene, and the owner's repeated focused runs. No interactive terminal smoke, real-process concurrency stress, editor or clipboard effect was run.

### Mutation ledger

Every mutation was compiler-valid (`go vet ./internal/tui` passed), and each was restored with `git checkout -- <file>` to the captured baseline, verified clean, and re-run green.

| # | Edit | Result |
| --- | --- | --- |
| M1 | `setFieldCmd`: `SetSelectedTaskFields(TaskMutationTarget{ID: ref.key}, …)` → `SetFields(ref.key, …)` | **Killed.** `TestQueuedSelectedTaskActionsRetainIdentityBeforeExecution/set/impostor` and `/set/case-folded_impostor` fail at line 107 ("reported success for a vanished/colliding ID": `editedMsg`). `TestSelectedTaskUIFormsRetainIdentityAndResultEnvelopes/set` fails at line 226 ("retargeted an impostor"). Controls (unchanged, renamed, missing, duplicate) pass. Restored; both tests ok. |
| M2 | `moveTask`: `MoveSelectedTask(...)` → `Move(ref.key, …)` | **Killed.** `move/impostor`, `move/case-folded_impostor` (line 107: `movedMsg` to in-progress), and form `move` (line 226). Restored; ok. |
| M3 | `deferTaskCmd` → `DeferTask(ref.key, …)` (supplementary) | **Killed.** `defer/impostor`, `defer/case-folded_impostor`, and form `defer`. Restored; ok. |
| M4 | `unsetFieldCmd` → `SetFields(ref.key, unset)` (supplementary) | **Killed.** `unset/impostor`, `unset/case-folded_impostor`, and form `unset`. Restored; ok. |
| M5 | Coordinated M1–M4 against the reviewer UI-route probe | **15 of 25 failed**: impostor, case-folded and prefix-extended for each of the five routes. Declaration and rename controls pass under both APIs, as expected; they are non-discriminating controls. Restored. |
| M6 | `closeTransientUI`: delete `m.action.close()` and `m.edit.close()` (`session.go:192`, `:195`) | **Survived** the full `go test ./internal/tui -count=1` (ok 3.054s). Killed only by the reviewer probe. This is **L1**. Restored; clean. |

Measured: each of the four commands has two queued impostor subtests plus one actual-form subtest that fail on reversion. This matches the owner's "eight failures" report for the queued test, extended with the form test.

### Hostile-probe ledger

All probes were temporary sandbox files, run on the captured baseline and then deleted. Their sources are retained under `/private/tmp/taskflow-claude-selected-actions.Lfp0yB/probes/`. Results are on baseline unless marked.

1. **Queued real UI routes** (`tui_queued_ui_routes_probe_test.go`): 25/25 pass.
   - Routes: menu `start`; menu `deprecate` with `y` confirm; menu `defer` with a date; edit `tags` (text); edit `description` (long text). Each command was built through `Update` before the change below.
   - Modes:
     - delete + exact ID-shaped slug impostor;
     - case-folded slug impostor;
     - `<ID>-x` unique-prefix fuzzy impostor, which is not in the owner's tests;
     - a declaration impostor (another file whose frontmatter `id:` is the selected ID);
     - rename to `<ID>-renamed` with a new task reusing the old `chosen` alias.
   - Hostile modes refuse with `task identity "<id>": not found` (fields) or `task "<id>": not found` (lifecycle). For the declaration impostor, lifecycle refuses as a broken graph (frontmatter id disagrees with source id). Bytes of impostor, sibling and alias-reuse files are unchanged.
   - The rename control commits to the exact ID: receipt `Task.ID` equals the selected ID, `Committed` is set, there is no warning, and the UI flashes and reloads.
   - Session (`gen`) and mutation (`ref`) envelopes are present on every result.
2. **Real-FS committed failure** (`store_selected_probe_test.go`): `testHookRepositoryUnlockError` returns `fmt.Errorf("injected release failure: %w", domain.ErrConflict)` after a real write. For `MoveSelectedTask` and `DeferSelectedTask`:
   - `errors.As` gives a `*core.TaskLifecycleMutationFailure`; `errors.Is` still reports both `ErrConflict` and the cause;
   - receipt `Committed` is set and `Task.ID` is the selected ID;
   - hook calls = 1 and retry sleeps = 0, so a conflict-classed post-commit error does not re-run;
   - the file is transitioned and the sibling is unchanged.

   This is a real-adapter injection at the release seam, not the owner's port double. It is not a real OS-level cleanup failure.
3. **Contention swap between attempts**: `testHookBeforeLifecycleVerify` / `testHookBeforeSetFieldsWrite` swap in an exact or case-folded impostor on the first attempt. Exactly one retry (sleeps=1), then `ErrNotFound`, with impostor bytes unchanged. The rename control succeeds on the renamed file.
4. **Real-store lifecycle policy through the selected API:**
   - an unmet dependency gives a `*TaskEligibilityError` with `DependencyGateOverrideAllowed()`, no write and no retry;
   - the `DependencyGate` override commits with `Forced`, `Remedy` "resolve the outstanding blockers…" and a non-empty `OutstandingBlockers`;
   - the AC override aimed at `next-up` gives `ErrValidation`;
   - an unticked AC on completion refuses with "no reason" and no write;
   - a dry-run AC override is `Forced`, not committed, and makes no write;
   - defer dates `2099-13-01`, `tomorrow` and `␠2099-01-01` give `ErrValidation` before adapter I/O, with no write.
5. **Fresh binary, disposable space** (`init --no-register`, isolated `HOME`):
   - Repair of `tier: three` (with an unknown key and comments): `--dry-run task set ALP --tier 2` prints "would update alpha" with identical hashes. The write repairs `tier`, preserves `x-custom`, comments and key order, appends `updated_at`, leaves sibling `beta`'s hash unchanged, and `task info` reads the result.
   - Unrelated duplicate-key corruption: both write and dry-run refuse with "already has malformed frontmatter (not caused by this update)", rc=11, bytes unchanged.
   - Refusal matrix (write and dry-run, 20 rows, all rc=11, bytes unchanged): `status` set (±`--force`), `depends_on` set (±`--force`), `--unset updated_at`, `--unset status`, unknown set and unset, unknown epic, `tier=nine`.
   - Positive: `--force` custom field; epic detach; `epic=01` canonicalises to `01-repair-probe`; an ID-shaped human slug works for both `task set` and `task start`.
   - Lint is clean afterwards.
6. **Cross-activation open form** (`tui_session_probe_test.go`): see L1. Baseline passes; under M6 a form opened in A writes B's same-ID task.
7. **Reachability**: Atlas Enter on beta → Esc → `]` → `e` → beta's still-current open lands under alpha's form. Baseline closes the form; under M6 it survives.

### Surviving hypotheses

**Tested-negative:**
- **Identity lost before entering the new methods** (alternate delegate, editor callback, display-label key): all task routes use `item.ref()` from `Source.ID`; the registry delegate is `moveTask`; no other production task writer exists.
- **Refusal caused by bogus fixtures**: the `testTaskRef` fixtures equal `TaskFixture`'s `TaskID(stem)` (`internal/testutil/repo.go:64-69`). The edited `action_test.go` cases assert policy text ("no reason", "outstanding blockers", impact remedy). No `testEntityRef` reaches a task command.
- **Ref-based and selected field paths diverge**: both use one `prepareTaskFieldUpdates`, and the owner test compares error strings for eight invalid maps. The CLI matrix above holds. `updateFrontmatter` never writes to the shared update map (`internal/store/frontmatter.go:112`). Blank or whitespace IDs are refused in core before adapter access. Other malformed IDs are left to adapter validation (`validEntityID`, `id.Valid`), so core adds no ID syntax rule.
- **Contention replays a re-resolved target, or retries a committed operation**: refuted by probes 2–3.
- **Lifecycle revalidation authorises a sibling**: refuted by probes 3–4.
- **Stale results erase another selection's editor or flash, or skip a needed refresh**: covered by the owner test and the reducer. A stale success still reloads; a session-stale result is dropped, and returning to that space reloads its restored session.
- **An issued command is retargeted to a new Service**: refuted; closures capture `svc`.
- **Read-only and authorisation bypass**: `authorizeMutation` runs on both write paths, and the owner read-only test passes.
- **I/O outside `tea.Cmd`**: the command constructors only build closures.
- **Guide and comment overclaims**: cancellation, all-entity protection and local-path safety are explicitly disclaimed. The one Service-capture overstatement is folded into L1.

**Unresolved / not filed (pre-existing or no consequence demonstrated):**
- A stale `mutationResultMsg` carrying a committed-but-cleanup-failed `movedMsg` reloads but drops the "WARNING: committed, but cleanup failed" flash (`model.go:448-458`). This reducer policy predates this change; the owner may want to surface it regardless of selection.
- Error-precedence order differs: the selected path checks identity before fields. Both errors are `ErrValidation`.
- YAML comment spacing is normalised on repair, and `epic=1` is refused (only `01` resolves). Both are pre-existing.

### Restoration and isolation

Every probe and mutation was restored to the captured baseline; `git status --porcelain` was empty before this report was edited. The only sandbox change is this audit: the tool-written `L1`, the tool-stamped `updated_at`, and this report prose. Disposable scratch caches, binary and fixture space remain outside the sandbox. Probe sources are retained as evidence.

Post-report checks with the fresh sandbox binary and this report in place:
- `audit lint 6gm337aba0bh`: "all audit findings pass lint", rc=0;
- `lint --links`: "all planning entities and dependency links pass lint", rc=0;
- `audit info --json` parses `findings: {total: 1, open: 1}`.

The baseline gave the same lint results, so neither the baseline nor this report introduces a lint issue.

Isolation attestation (helper `verify` output, copied verbatim):

```text
sandbox_path=/private/tmp/isolated-review.mJ36Ii
git_dir=/private/tmp/isolated-review.mJ36Ii/.git
baseline_commit=3b7fc92452c4dce0300d385949f6c98fdfa40e59
source_blob=ab419b10449e44e2db44db8bd1cd7cbc4a8c94a0
source_fingerprint=796a798c66a6fd61918b36a56dd44c716099ab97
deliverable=planning/audits/6gm337aba0bh-2026-10-15-selected-tui-task-actions-implementation-claude.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Findings

#### L1. Open task forms rely on an untested activation close for Service binding · **Status:** fixed

**File:** internal/tui/session.go:192 | **Component:** tui
**Effort:** S · **Urgency:** soon

**Root cause.** Issued commands capture their Service, but open task forms and menus do not. `editMenu` and `actionMenu` keep only `ref` and `sourceGen` (`internal/tui/edit.go:120`, `internal/tui/action.go:100`). Submission builds the command from the model's *current* `m.svc` (`internal/tui/edit.go:336`, `:340`; `internal/tui/model.go:1146`, `:1166`, `:1195`). For a form opened in one checkout, the only thing that stops it from binding its selected ID to a newly activated Service is the call `activateWorkspace` → `closeTransientUI` (`internal/tui/session.go:134`), which runs `m.action.close()` and `m.edit.close()` (`:192`, `:195`). No test pins those two lines.

**Reachable producer.** On the Atlas, Enter on another entry point starts `openWorkspace` and bumps `openGen` (`internal/tui/atlas.go:392`). Esc runs `exitAtlas` (`:356`), which neither cancels that command nor bumps `openGen`. `handleWorkspaceOpened` checks only `msg.gen != m.atlas.openGen` (`:224-225`). So the open lands after the user has returned to the old space and pressed `e` or `m`. Reviewer probe `TestReviewProbeAtlasOpenLandsUnderOpenForm` drives this exact route through `atlasTestModel`: Enter on beta → Esc → `]` → `e` → deliver beta's open. Result: the workspace switches to beta while alpha's form is active. On baseline the form is then closed (PASS).

**Reproduction (compiler-valid mutation).** Delete `m.action.close()` and `m.edit.close()` from `closeTransientUI`. `go vet` passes, and the **entire** `go test ./internal/tui -count=1` suite stays green (`ok … 3.054s`). Reviewer probe `TestReviewProbeOpenFormAcrossActivation` uses two real FS checkouts with identical task IDs and slugs (a second checkout or worktree of in-repo planning):
1. Open the edit form in A (`sourceGen=2`).
2. Activate B. Its restored session reloads to `loadGen=2` with the same ID and label selected. The generations coincided naturally; no refresh alignment was needed.
3. The edit gate (`internal/tui/edit.go:260`) passes and the submit writes **B's** task: `A changed=false B changed=true`, flash `set priority on chosen`.

On baseline the same probe passes, because activation closes the form.

**Consequence.** No product failure today: baseline behaviour is correct. Contract 5's "must not silently mutate the newly displayed space" for open forms and menus rests on one untested line, though. A future refactor of `closeTransientUI`, a new modal route, or deferring the close would silently let a same-ID form or menu (field set/unset, move, confirm, defer prompt) write the newly displayed checkout. The owner's `TestQueuedSelectedTaskActionsRetainCapturedServiceAcrossSessionSwitch` (`internal/tui/selected_task_action_test.go:272-277`) hand-builds `scopeSession(1, scopeMutation(…, action.cmd(old.svc, ref)))` against a separate model. It never runs activation, so it cannot detect this. `docs/architecture/planning-data.md:56` ("The TUI captures its selected service along with the ID") is true for issued commands only.

**Uncertainty.** The write consequence needs the mutation, and the generation coincidence was shown in one natural sequence; per-tab counters may not always coincide. The epic edit form shares the same guard, but non-task callers are out of scope here. No real terminal was run. Probe sources are retained at `/private/tmp/taskflow-claude-selected-actions.Lfp0yB/probes/tui_session_probe_test.go`.

**Recommendation:** Add a real-route regression that opens task edit form, action menu, confirm and defer prompt, delivers a current-generation workspace open, and asserts all close with no write through the new Service; qualify the planning-data guide sentence or bind forms to their opening session.

**Resolution:** Accepted as a coverage gap, not a baseline product defect.
    TestSelectedTaskFormsCloseWhenPendingAtlasOpenActivates drives a real
    pending Atlas open under edit/menu/confirm/defer intents with identical
    selected IDs and matching list generations; activation closes each and both
    trees stay unchanged. Separately deleting edit.close or action.close
    produces the intended edit or lifecycle test failures; restored cases pass
    five race-enabled repetitions. Service-capture wording is qualified in
    owning comments and architecture guides. Full race tests, lint, build and
    docs-check pass. Stale committed-warning presentation remains in navigation
    design task 6g8vxcnezktm.
