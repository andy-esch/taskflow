---
schema: 1
id: 6gpb8hr5rjk1
bucket: closed
area: tui-creation-host-contracts-implementation-codex
date: "2026-10-22"
updated_at: "2026-10-22"
---
# Audit: TUI creation host and interaction contracts implementation — codex — 2026-10-22

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
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

Review the shared TUI creation-host foundation before more entity forms rely on
it. Can one host safely own task/epic draft lifetime, delayed typed outcomes and
displacement, while the new test harness actually delivers the events it claims
to cover? This is a Codex-only review. Primary lens: lifetime/identity/result
safety; complementary lens: test integrity and rendered/input-size agreement.

Perform a second systemic pass after the bounded probes. Seek concrete failures
or misleading tests, not a quota of findings or a generic form-framework redesign.
Checked planning ACs are assertions to challenge, not proof of the contract.

## Review target

Source: `/Users/andyeschbacher/git/andy-esch/taskflow`, branch
`refactor/tui-creation-host-contracts`, base
`7f14290e6b2ab7be30a512b53e0b82761390bae8`. The implementation and implementing
task are **dirty and partly untracked**. Capture the complete source state using
the mandatory isolated clone; do not review source HEAD alone.

- [Implementing task](../tasks/6gpb4y5yqmjp-harden-the-shared-tui-creation-host-and-interaction-contracts.md).
- [Authoring Thread](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md)
  and [integration owner](../tasks/6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md).
- [Owning TUI guide](../../docs/architecture/tui.md#epic-creation-and-shared-staging)
  and relevant [planning-data contracts](../../docs/architecture/planning-data.md#change-checklist).

Verify a consumer inventory tracing production callers, not just these starting
points. Cite actual symbols/lines in your captured sandbox:

- `internal/tui/creation_host.go`: `creationHost`, `begin`, `draft`, `openTask`,
  `openEpic`, `displace`, `creationResult`, `applyResult`, `openPageCreation`,
  `runCreationAction`, `creationActions`.
- `model.go`: the single `Model.creation` field, ordinary creation-result dispatch,
  the old-session exception in `Model.Update`, widget routing and page-local c.
  `session.go`: `scopeSession`, `activateWorkspace`, `closeTransientUI`.
- Per-kind controllers: `task_create_commands.go` and `epic_create.go` implement
  `creationRoute`/`applyCreation`; `handleTaskCreated`/`handleEpicCreated` retain
  receipt and origin-Service policy but now rely on the host's routing checks.
  Core `Service.NewTaskForSelectedEpic` and `Service.NewEpic` are unchanged.
- `staged_creation.go`: borrowed draft interface, mechanical controls, widget
  generations and text-buffer detachment. `view.go`/`overlay.go`: `modalSize` and
  shared modal composition; `staged_creation_view.go`: `creationInputFits`,
  `inputVisible`, `inputOwner` and `creationHeader`.
- `staged_creation_helpers_test.go`: moved `createKey`, `settleCreate`,
  `settleCreationCommands`. `creation_host_test.go`: mechanical matrix and the
  explicitly routing-only `creationRouteProbe`. It is not durable-adapter evidence.
- Existing typed/durable tests in `task_create_test.go`, `epic_create_test.go`,
  shared input tests in `staged_creation_test.go`, and Darwin-only native provider
  fixtures in `staged_creation_clipboard_darwin_test.go` remain relevant consumers.

No core API, wire revision, entity schema, new entity form or CLI metadata change
is intended. Audit/research and guarded Thread forms are separately sequenced
after this foundation. Existing editor/palette/command-bar terminal-paste loss is
tracked by [6gp2qmse4prt](../tasks/6gp2qmse4prt-route-terminal-paste-to-active-tui-text-input-owners.md)
and is not a new finding unless you demonstrate a distinct in-scope consequence.

Leave unrelated untracked task-creation audit `6gnqmm1yqhxz` and owner walkthrough
task `6gnqadwhhz54` untouched. They are not assignments for this review. Historical
[epic review](6gp2rf2gyp1a-2026-10-21-tui-epic-creation-shared-staging-implementation-codex.md)
records native-paste and shallow-buffer failures already fixed in the base; verify
preservation without reporting those historical failures as present by assumption.

## Intended contract to challenge

1. One active draft is authoritative. Host entry points cannot replace active,
   pending or displaced recovery intent. Kind changes/closes/workspace activation
   never reuse an attempt generation. Concrete value slots and synchronous borrows
   preserve retained draft text; animation bookkeeping is not promised immutable.
2. Host result routing requires a live pending draft with matching kind, attempt
   and original session. Removing the guards from per-kind handlers is safe only
   if every production delivery traverses that host boundary exactly once.
3. The old-session exception admits typed create outcomes for origin recovery,
   not catalog/widget/read messages. A late origin result cannot land, reload or
   write in the incoming workspace, even when IDs match in both spaces. Ordinary
   current-space outcomes retain exact-ID landing and watcher/filter behavior.
4. Per-kind requests, parent/default choices, typed receipts and recovery decisions
   remain explicit. Known uncommitted refusals may permit deliberate retry;
   pending, finalization, uncertain or committed outcomes cannot replay Create.
   Receipt paths remain evidence, not selection/editor authority.
5. Finite application and widget results run through Model.Update in the harness,
   including opaque native clipboard results. Only recognized cursor maintenance
   is excluded. Batch/follow-up settling has a work budget, not a wall-clock timeout
   or a replacement for asynchronous running-program tests.
6. One child-size rule gates input and resize rendering after the shell gutter is
   removed once. Whole-terminal 46x18 remains supported; 45x18 and 46x17 display
   a non-authoring resize notice. Native and terminal paste cannot edit hidden
   text; matching blink maintenance retains its update chain.
7. New forms should extend host-local wiring and concrete controllers, not add
   result/displacement cases to the root model or introduce a universal core create
   interface. Judge maintainability with actual extension seams and concrete
   consequences; preference for another pattern alone is not a defect.

## Mandatory evidence floor

Independently run the focused suites below, then perform bounded hostile probes
through actual production callers. Do not rely solely on `creationRouteProbe` or
on setters that skip keyboard/widget routing:

- For **both kinds**, issue an actual create command, change workspace before
  delivery, then execute/deliver the old command/result. Use same-ID records in the
  two spaces and capture before/after bytes/modes of the incoming space. Execute
  returned recovery/inspect commands before asserting no replay or reload. Prove
  that the result enriches origin recovery without rebinding. Existing tests may
  supply fixtures, but independently verify their counter/command coverage.
- Reach one known refusal and one committed-cleanup/unknown outcome through the
  per-kind controller boundary. Verify explicit retry versus locked recovery and
  no duplicate create. Explain real producers versus adapter-injected faults or
  malformed-receipt defense in depth. At least one valid create must independently
  check non-default durable fields and scaffold, not derive expected values from
  shared preparation/default helpers.
- Exercise a finite wrapped widget result through the **new harness**, then
  compare one relevant flow against a running Bubble Tea program. On Darwin use
  the shipped stub-provider Ctrl+V tests without touching the user's clipboard;
  on another OS report that native-provider coverage was not executed rather than
  treating public tea.PasteMsg injection as equivalent.
- Render and deliver input at supported/unsupported size boundaries and after a
  resize round trip. Verify content and cell dimensions, not just byte length or
  a matching helper Boolean. Include at least one pending/recovery mode.

Execute two small compiler-valid mutation probes, one for each critical seam:

1. Alter a host lifetime guard (attempt/kind/pending/single-active refusal), or a
   per-kind origin guard reached through the new result interface. Name the exact
   intended regression and require its relevant assertion to fail. If an outer
   caller masks the edit, report survival honestly or make a justified coordinated
   caller mutation; compile/vet failures are invalid, not kills.
2. Restore drop-all-widget behavior in `settleCreationCommands`; require
   `TestCreationHarnessDeliversFiniteWidgetResults` to fail. Use one actual finite
   native result too when the platform supports the existing stub fixture. Then
   restore and rerun the intended tests green.

Owner already killed two overlays: deleting attempt matching fails the task/epic
stale-attempt cases in `TestCreationHostResultGuards`; dropping all widget results
fails both finite-delivery cases. These are inherited observations, not independent
review evidence. Challenge whether synthetic guards and helper-based scenarios
also establish the reachable production consequences above.

Record exact edit, invariant, command, intended assertion, killed/survived/invalid
result and restored run. A surviving mutation is not automatically a product bug.
A no-findings verdict still needs this bounded evidence, not just a green suite.

## Required hostile angles

- **Lifetime/value ownership:** copy a populated Model, mutate/close/displace only
  the next copy, and inspect prior text/mode/receipt state. Does host selection hide
  another live slot? Can cross-kind reopen or a delayed old result recover authority?
  Separate reachable constructor paths from impossible manually fabricated hosts.
- **Result trust:** enumerate every caller of the typed apply methods, including
  normal and old-session paths. Check pending false, closed/reopened attempts,
  matching numbers on the wrong kind, duplicate delivery, workspace changes before
  dispatch and after issuance. Do not equate a test-only result implementation with
  a shipped untrusted producer; label synthetic defense-in-depth cases explicitly.
- **Helper trust:** follow BatchMsg, session wrappers, native private results,
  blink/cancellation messages and generated follow-up commands. Can a skipped event,
  accidental budget exhaustion, alias or synchronous ordering make the harness
  claim that no write/replay happened before the relevant command executes? Keep
  work-bounded versus time-bounded claims distinct. tea.Sequence is already forbidden
  under session scoping; don't invent a requirement to support an unused protocol.
- **Layout/input:** check shell versus child dimensions and border/padding. The
  common sizing helper must not double-subtract the gutter or change other modals
  through shared body composition. Hidden paste and maintenance have different rights.
- **Second systemic pass:** this refactor is meant to improve extensibility. Look
  for a new parallel registration obligation, false generic interface, too-broad
  Model authority, omitted typed policy, or a matrix that only tests itself. Resolve
  each important hypothesis with a caller trace/probe; distinguish useful cleanup
  recommendations from demonstrated bugs. No speculative architecture finding quota.

Prompt provenance: the prior review showed that public terminal-message injection
missed a private native result, and an empty text buffer concealed aliasing. This
brief therefore asks for producer traces, non-empty state, executed commands and
a running-program cross-check rather than trusting names, types or checked ACs.

## Validation and restoration

Owner checks passed: `just build`, full `just test` race suite, `just lint`,
planning/link lint, audit lint and whitespace checks. Targeted routing, sizing and
runtime probes also passed three race repetitions. You may inherit those full gates
as explicitly inherited evidence; independently execute these focused commands:

```sh
go test -race ./internal/tui -run 'Test(Creation|TaskCreate|EpicCreate)' -count=1
go test -race ./internal/tui -run '^Test(TaskCreateInRealProgram|TaskCreateTerminalPasteInRealProgram|EpicCreateRuntimePaste|CreationNativeClipboardAfterRuntimeResize)$' -count=3
```

Also execute the selected hostile probes and mutation/restoration runs. Build your
own sandbox binary if using planning/finding commands. Set task-specific GOCACHE
and GOLANGCI_LINT_CACHE under external scratch if necessary; do not add dependencies,
change toolchain pins, read the real clipboard or use the source planning tree as
a fixture. The known nonfatal Go module stat-cache warning is not a failed build.

The work budget cannot cancel an arbitrary blocking Cmd. Inspect that limitation
truthfully rather than calling it a timeout. Do not claim profiling/performance
improvement without measurement. No generated CLI pages or wire changes are intended.

Restore all probes to the sandbox's captured dirty-source checkpoint, **not source
HEAD**. Only your assigned audit may differ at verification. Follow the injected
trusted-source-helper verify/transfer/receipt protocol; retain the sandbox until
the owner confirms delivery. Never fall back to source edits, copy extra files,
make another commit or push. Report isolation/permission/transfer failure as a blocker.

## Deliverable

Preserve this brief and fill the report below. Leave findings open using the
injected finding grammar (prefer sandbox-built audit finding new). Consolidate
root causes and give reachable producer/caller trace, expected/actual behavior,
reproduction, severity and bounded recommendation. Coverage gaps and preferences
are not corruption merely because an AC is checked.

Include verdict/limits, verified consumer inventory, independent versus inherited
checks, hostile/mutation ledger, a short tested-negative/unresolved hypothesis
list and mandatory isolation attestation. Keep evidence sufficient to reproduce
without transferring scratch code, but avoid raw log dumps. A blocked review must
say incomplete; a no-findings result is acceptable only after the evidence floor.
Return the helper's transfer receipt in your handoff and keep the sandbox intact.

## Reviewer report

### Verdict and limits

The bounded technical review is complete with **no new findings**. The production caller trace, hostile origin/recovery flows, native-provider/runtime checks and compiler-valid mutations support this foundation for the existing task/epic consumers. No finding was closed or dispositioned; the audit remains open. This is not evidence that the separately planned audit/research/Thread forms already exist.

Reviewed the complete captured working state against base `7f14290e6b2ab7be30a512b53e0b82761390bae8`, using sandbox checkpoint `30ec49fe3d212639b77ffa9965d88b0f15ae017d`, rather than source HEAD. Full project race/lint gates are inherited, explicitly distinguished below. The review does not measure performance, guarantee every asynchronous interleaving, or treat synchronous settling as a wall-clock timeout. The isolation procedural deviation below is disclosed rather than claiming exact compliance with the initial source-read allowlist.

### Verified consumer inventory

References are paths/lines in the captured baseline, not proposed names in planning documents.

| Boundary | Verified production consumers and implementation |
| --- | --- |
| Lifetime and opening | `internal/tui/creation_host.go:14` stores concrete task/epic values, kind and generation; `draft` at 72 lends only the selected active slot; `begin` at 93 refuses replacement and allocates generations. `openTask` at 102 and `openEpic` at 110 construct fresh values; `displace` at 118 retains the active draft. `Model.creation` is the single root field at `internal/tui/model.go:102`. |
| Page/palette callers | `model.go:900` calls `openPageCreation` (`creation_host.go:44`). `command_dispatch.go:170` consumes `creationActions` (`creation_host.go:21`); `runPaletteItem` at 233 delegates to `runCreationAction` (`creation_host.go:56`). Task context still rechecks the selected epic before opening (`task_create_commands.go:19`); epic opening is guarded at `epic_create.go:107`. |
| Result trust | The only production implementations of `creationResult` (`creation_host.go:128`) are `taskCreatedMsg.creationRoute/applyCreation` at `task_create_commands.go:111/115` and `epicCreatedMsg.creationRoute/applyCreation` at `epic_create.go:176/180`. Each apply method has one handler call, and `creation_host.go:139` is the only production call of `applyCreation`. Host matching at 133 requires active, pending, kind, attempt and original session. Ordinary delivery is `model.go:338`; the old-session exception is at 284–296. Neither catalog nor widget results implement this interface. `creationRouteProbe` (`creation_host_test.go:92`) is test-only routing evidence. |
| Origin and typed policy | `submitTaskCreate` (`task_create_commands.go:72`) copies task inputs and selected-parent identity into a command capturing the origin Service; `submitEpicCreate` (`epic_create.go:156`) separately captures `NewEpicParams`. `handleTaskCreated` at 120 and `handleEpicCreated` at 185 retain entity-specific receipts and origin/displacement checks. Core `Service.NewTaskForSelectedEpic` (`internal/core/selected_task_creation.go:46`) and `Service.NewEpic` (`internal/core/service_epic.go:25`) are unchanged in the reviewed diff. |
| Workspace lifetime | `scopeSession` (`session.go:21`) recursively scopes Batch children. `activateWorkspace` at 117 replaces Service, increments session at 130 and calls `closeTransientUI` at 192; that calls host displacement. Origin commands capture their Service before this operation. The old-session branch can retain the matching pending origin result, but the per-kind guard prevents incoming-space landing/reload. |
| Mechanics and input | Borrowed `stagedCreation`/`creationControls` are in `staged_creation.go:142/127`; state, focus and mode transitions are at 31/60/171. `routeCreationWidget` at 107 copies kind/form/field tokens and wraps Batch children without retaining the draft/Service. `updateCreationInput` at 92 detaches non-maintenance text buffers. `model.go:342` checks widget tokens and visible ownership. Only `cursor.BlinkMsg` is recognized maintenance (`staged_creation.go:82`). |
| Layout and presentation | `modalSize` (`overlay.go:36`) removes the shell gutter once. `view.go:123` passes those same dimensions to every modal. `creationInputFits`, `inputVisible`, `inputOwner` and `creationHeader` (`staged_creation_view.go:23/17/29/82`) share the child-size threshold and precedence checks. Task/epic presentation stays in `task_create_view.go:5` and `epic_create.go:229`; shared receipt presentation grants no path authority. |
| Harness and independent oracles | `createKey`, `settleCreate`, `settleCreationCommands` (`staged_creation_helpers_test.go:10/31/40`) execute Batch children, unwrap one session envelope for maintenance classification, deliver finite results through Update, then execute follow-ups against the returned Model. Existing typed/durable and watcher tests are in `task_create_test.go:63/181/343/614/658` and `epic_create_test.go:49/193/276/351`. Non-empty text-buffer/hidden-owner regressions are in `staged_creation_test.go:46/91/112/142`. Darwin's real provider stub and runtime hold/release fixture ship at `staged_creation_clipboard_darwin_test.go:25/34/109/131`. |

`git diff <base> -- internal/core internal/wire internal/cli cmd` was empty. No changed wire branch, CLI metadata or schema required a non-default wire validator or generated CLI documentation. Checked ACs were treated as review claims, not implementation evidence.

### Independent checks and hostile ledger

All Go commands below used `GOCACHE=/private/tmp/taskflow-codex-audit-gocache` in the sandbox. Darwin was independently confirmed with `uname -s`.

- Required focused suite: `go test -race ./internal/tui -run 'Test(Creation|TaskCreate|EpicCreate)' -count=1` passed (81.151s).
- Required running-program cross-check: `go test -race ./internal/tui -run '^Test(TaskCreateInRealProgram|TaskCreateTerminalPasteInRealProgram|EpicCreateRuntimePaste|CreationNativeClipboardAfterRuntimeResize)$' -count=3` passed (16.675s). Native tests use the shipped isolated-PATH `pbpaste` stub; no real clipboard was read.
- Additional shared-shell checks: `go test -race ./internal/tui -run '^(TestModel_(ViewFitsTerminal|HelpOverlayFitsTerminal)|TestModal_.*|TestSessionScoping.*)$' -count=1` passed (2.246s). Architecture navigation, sandbox `just build`, `./bin/tskflwctl -C . lint --links`, and `audit lint 6gpb8hr5rjk1` passed. Build emitted the known nonfatal module stat-cache warning. Initial test attempts using the default protected cache failed setup, then succeeded with the scratch cache.
- Independent scratch scenarios: `go test -race ./internal/tui -run '^TestAudit(Origin|RetryAndRecovery|DurableKeyboard|DimensionsAndOwnership|RealValidationRetry)$' -count=1 -v` passed (69.347s). These reviewer-created tests were not shipped fixtures; they were removed from the working tree after review. A copy is retained only as sandbox evidence under `.git/reviewer-evidence/audit_independent_test.go`. An initial scratch compile error and an incorrect expectation of an epic `## Tasks` heading were corrected before the passing run; neither is a mutation kill or product finding.

| Hostile scenario | Executed evidence and result |
| --- | --- |
| Both kinds, late origin execution and same-ID incoming records | `TestAuditOrigin/task` and `/epic` use real page `c`, terminal input, actual task parent picker and Enter submission, then activate another fixture workspace before executing the captured command. Incoming bytes/modes are checked before/after command execution. After the origin commit, install its exact created ID into incoming with different body text and mode 0600, then capture a second tree snapshot. Deliver the actual old-session typed result and execute every returned follow-up. A read-counting incoming Store detects reloads in addition to byte/mode checks. Both retain committed origin recovery, clear pending, keep the incoming page/session, perform zero incoming task/epic reads and preserve the full incoming snapshot. Task ID in the restored mutation run was `6gpbb52hccrc`; epic ID was `02-host-hostile`. |
| Recovery execution, duplicate delivery and retained copies | The same probes deliver the actual result again after pending clears, then press Enter on displaced recovery and settle any returned command before checking counters/snapshots. Exactly one origin create per kind; no incoming selection intent or reload. Retained pre-activation pending Model text/mode remains intact; a populated recovery copy keeps its committed receipt, identity, path, active/locked state and text after the next Model closes. Existing host lifetime and non-empty mid-string insert/delete/query regressions also passed. |
| Known refusal versus locked outcomes | `TestAuditRealValidationRetry` leaves task tags or epic description empty through keyboard routing, submits, verifies a real core validation refusal with zero persistence calls/byte changes, then repairs the visible field and explicitly retries: exactly one durable create. `TestAuditRetryAndRecovery` additionally injects conflict, committed cleanup and unknown transport outcomes through the concrete creation ports for both kinds, settles result-triggered loads and executes recovery Enter. Known conflict allows two deliberately issued attempts; cleanup/unknown stay at one. Cleanup inspection lands on the exact committed record. |
| Producers versus defensive receipts | Real validation/default/scaffold behavior comes from unchanged core and real FS persistence. Cleanup probes perform a real FS create and then inject `CreateFinalizationError`; unknown transport errors are injected faults, not a shipped remote producer. The FS has a real finalization producer in `internal/store/create.go:102` (guard-release error); the probe does not claim to force that OS failure. Incomplete/preview receipts in the existing suites are explicitly malformed-adapter defense in depth. No speculative corruption claim follows from those fixtures. |
| Independent durable non-defaults | `TestAuditDurableKeyboard` uses only page/input/enum/action routing for an epic with literal expected status `retired`, priority `low`, description `Host epic description`, tags `[audit, independent]`, ID `02-host-hostile`, H1 `Host hostile` and the shipped `## Why this is its own epic` scaffold. Expected values do not call preparation/default helpers. Existing task durable tests independently pin high priority, tier 2, autonomy 4, effort and scaffold; watcher/filter landing tests execute their returned loads. |
| Layout/input and round trips | `TestAuditDimensionsAndOwnership` covers both kinds in fields, issued-pending and displaced-pending recovery, visiting 46×18 → 45×18 → 46×17 → 46×18 → 47×19 → 120×40. It checks rendered mode content, line counts and `ansi.StringWidth` cell bounds, plus actual terminal paste. Strengthened rerun also delivers printable `Q` through the key path and checks the current field's exact value: only supported editable fields gain `ZQ`; hidden/pending/recovery fields and held-create fixture bytes stay unchanged. The strengthened rerun passed (31.927s). Native hidden delivery and post-resize restoration are independently exercised by the shipped stub-provider reducer/program tests above. |
| Finite delivery versus asynchronous ordering | The focused suite executes session-wrapped finite widget results through the new harness for both kinds and actual opaque native Ctrl+V results through the Darwin fixture. The running program holds a real native result, renders Resize, delivers it, restores 46×18 and confirms no hidden edit. Matching hidden blink continuation and explicit loop-budget failure pass. The helper decrements before `cmd()` and cannot cancel a blocking command; sequential Batch settling is not a concurrency oracle. |

Inherited only: owner's full `just test`, `just lint`, earlier complete gates and earlier overlays. They were not reclassified as independent results. No broad source fix, dependency/toolchain change, CLI generation or full-project rerun was needed for this one-audit deliverable.

### Compiler-valid mutation and restoration ledger

1. **Origin/result seam — killed for both real controllers.** In `task_create_commands.go:123` and `epic_create.go:188`, replace only `if f.displaced || f.origin != m.svc || f.session != m.sessionGen {` with `if false {`. The exact invariant is that a live pending typed outcome admitted by the host's old-session exception must preserve origin recovery rather than land/reload incoming. Run `go test -race ./internal/tui -run '^TestAuditOrigin$' -count=1 -v`. Both kinds compiled and failed the intended `origin result failed to enrich displaced recovery` assertion after executing the result's follow-up commands (7.778s). Host matching remains intact; no coordinated outer-caller weakening was needed. Restore both files from the captured sandbox HEAD; the identical command passed (8.743s). This independently challenges reachable production results, rather than counting the owner's synthetic stale-attempt overlay as reviewer evidence.
2. **Harness seam — required drop-all-widget mutation.** At `staged_creation_helpers_test.go:69`, replace `if widget, ok := inner.(creationWidgetMsg); ok && creationWidgetMaintenance(widget.msg) {` with `if _, ok := inner.(creationWidgetMsg); ok {`. Run `go test -race ./internal/tui -run '^(TestCreationHarnessDeliversFiniteWidgetResults|TestCreationNativeClipboardOwnership)$' -count=1 -v`. The compiler-valid mutation was **killed**: both kinds failed `settling discarded finite widget input` at `creation_host_test.go:153`; both actual native-provider visible cases failed at `staged_creation_clipboard_darwin_test.go:91`, retaining empty titles instead of `CLIP-qhjkn-世界` (7.717s). Restore the helper from captured sandbox HEAD; the identical command passed (10.697s), including both native visible cases. No actual clipboard was read.

No compiler/vet failure was counted as a killed mutation. Restore means the captured dirty-source checkpoint, never source HEAD.

### Second systemic pass: tested negatives and remaining limits

- **Hidden second slot / pointer ownership:** constructors require successful host `begin`; closing precedes kind changes and generations increase across reopen. The mechanical matrix plus non-empty widget edits, origin displacement and retained receipt copies pass. A manually fabricated host with two live slots is not a reachable constructor path. Cursor animation bookkeeping remains outside the immutable-text promise.
- **Overbroad result interface:** whole-package caller/implementation search found exactly the two production typed outcomes and one host invocation. Both root delivery paths cross that boundary once; only those outcomes receive the late-session exception. Synthetic wrong-kind/attempt/session/not-pending route tests passed, but are routing defense in depth, not hostile persistence producers. Duplicate actual typed result delivery was separately executed above.
- **Harness tests itself:** opaque provider delivery and running-program held-result ordering independently challenge the public-message-only finite test. Recursive Batch/follow-up work is exercised; only recognized cursor ticks are skipped. `TestSessionScopingHasNoSequenceCommandsToLeak` (`session_scope_test.go:33`) passed; no requirement was invented for the forbidden unused Sequence protocol. Arbitrary blocking Cmd cancellation remains unsupported and is not promised.
- **Extensibility and parallel registration:** the actual seam is concrete host slots/openers/draft selection plus per-kind request/receipt policy and host-local action wiring. No new root result/displacement case is required for another form. This is deliberate adapter wiring, not a universal core create API or a second entity schema. No demonstrated harmful registration omission emerged for the two shipped consumers; future forms still require their own independent durable assertions.
- **Shared layout regression / historical bugs:** common modal sizing is algebraically the prior `width-2, paneOuterH-2`; shared-shell/render checks and boundary input tests passed. Historical private-native-paste and non-empty-buffer failures were verified fixed by current tests, not reported again. The older editor/palette/command-bar paste work remains separately tracked; no distinct new in-scope consequence was demonstrated. No changed repair advice exists in this diff.

### Isolation and delivery attestation

All implementation inspection, builds, tests, scratch source changes, mutation/restoration and report editing occurred in `/private/tmp/isolated-review.5JnbZa`, created by the trusted source helper. No worktree, source commit, staging, branch change or manual transfer was used. The brief and all other deliverables are preserved. One procedural deviation: before reading the isolation section, initial discovery ran `pwd` and a filename-only `rg --files` listing in the source, in addition to reading this audit. That exceeded the initial source-read allowlist; it read no implementation contents and performed no source write/test. All subsequent work used the independent sandbox.

Final trusted-helper verification output follows after probe restoration. The report truthfully retains `transfer=pending`; delivery is established only by the helper-owned receipt returned in the handoff.

```text
sandbox_path=/private/tmp/isolated-review.5JnbZa
git_dir=/private/tmp/isolated-review.5JnbZa/.git
baseline_commit=30ec49fe3d212639b77ffa9965d88b0f15ae017d
source_blob=5f263fb8b9916171a3a496dae29a28eabfb73c68
source_fingerprint=52d7c9991d78978141ba09b8e408654827ecf71e
deliverable=planning/audits/6gpb8hr5rjk1-2026-10-22-tui-creation-host-contracts-implementation-codex.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner disposition — 2026-10-22

Accepted the bounded no-findings verdict after inspecting the report and confirming the trusted helper receipt: verify=passed, transfer=succeeded, delivery=confirmed. Independent origin-isolation, recovery, runtime/native-input and restored mutation evidence supports the existing task/epic consumers; full project gates remain inherited evidence. The disclosed preliminary filename-only source discovery was a procedural deviation, not a source mutation or implementation inspection. No code fix or new follow-up is warranted. Closing this audit; the reviewer report above preserves its pre-transfer historical state. The implementation task remains in progress for PR/integration handoff.
