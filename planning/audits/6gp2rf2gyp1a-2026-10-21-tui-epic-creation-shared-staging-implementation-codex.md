---
schema: 1
id: 6gp2rf2gyp1a
bucket: closed
area: tui-epic-creation-shared-staging-implementation-codex
date: "2026-10-21"
updated_at: "2026-10-22"
---
# Audit: TUI epic creation and shared staging implementation — codex — 2026-10-21

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

Challenge the epic-creation slice and the task/epic shared staged-form host before
it becomes the basis for other entity forms. Can a real operator create exactly
one intended epic, retain useful recovery evidence and land on its exact identity,
without regressing task creation or crossing workspace/input ownership boundaries?
Do not substitute a hypothetical universal-form redesign for reachable evidence.

Reviewer allocation (the assignment in this audit selects your primary lens):

- **Codex:** typed requests/receipts, exact identities, retained service/session
  binding, delayed outcomes and test helpers that can conceal replay or writes.
  Cross-check one supported-small-screen recovery and delayed-paste interaction.
- **Antigravity:** actual key/paste flows, modal precedence, narrow rendering,
  truthful recovery/actions, discovery and shared widget ownership. Cross-check
  one real committed-cleanup receipt through inspection, executing returned commands.

Both own the shared contract below. Independently settle falsifiable hypotheses;
no quota of findings and no requirement to agree. Do not read the other new report
until your independent report is final.

## Review target

Source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`, branch
`feat/tui-epic-creation-shared-forms`. Base commit:
`e18f6c656abedb1a34cae345c8cadda12a1b0897`. **The implementation is dirty and
partly untracked. HEAD alone is not the target.** Capture the full source state
using the mandatory independent-clone protocol above.

- Implementing [epic/shared-host task](../tasks/6gp1ytfvm652-create-epics-in-the-tui-with-shared-staged-form-mechanics.md).
- [Integration owner](../tasks/6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md)
  and [authoring Thread](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Interaction checkpoint](../research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md),
  [TUI architecture](../../docs/architecture/tui.md#epic-creation-and-shared-staging)
  and [planning-data checklist](../../docs/architecture/planning-data.md#change-checklist).

Build a verified consumer inventory, not a copied file list. Starting points:

- `internal/tui/epic_create.go`: `epicCreateForm`, `submitEpicCreate`,
  `handleEpicCreated`, `landCreatedEpic`, `recoveryText`.
- `staged_creation.go`: `creationState`, `creationControls`, `stagedCreation`,
  `routeCreationWidget`, `creationDraft`, `landCreatedEntity`.
- `staged_creation_view.go`: `creationModal`, `creationHeader`, `creationFrame`,
  `creationFieldLayout`, `creationErrorLines` and `creationFieldsView`.
- Task consumers: `task_create.go`, `task_create_commands.go`, `task_create_view.go`.
- Actual callers: `Model.update`/`handleKey` in `model.go`, palette dispatch in
  `command_dispatch.go`, modal precedence in `overlay.go`, displacement in
  `session.go`, help/footer in `keys.go`/`view.go`. The adjacent `edit.go` and
  `palette.go` changes clarify comments; they do not add paste capability.
- Existing core boundary: `Service.NewEpic` in `internal/core/service_epic.go`,
  `Service.NewTaskForSelectedEpic` in `selected_task_creation.go`, typed creation
  receipts in `creation_receipt.go`. No core port, CLI metadata or wire change.
- Evidence: `epic_create_test.go`, `staged_creation_test.go` and existing
  `task_create_test.go`. Verify every referenced test/symbol in your sandbox.

Explicitly deferred: audit/research forms (`6gp1ytg50tev`), guarded Thread creation
(`6gp1ytgea8dt`), template/bulk authoring, registry management and the adjacent
[older-input terminal-paste task](../tasks/6gp2qmse4prt-route-terminal-paste-to-active-tui-text-input-owners.md).
That last task records reproduced paste loss in `e` description, Ctrl+P query and
`:` command bar; config embedding still needs inventory. Re-report it only for a
distinct in-scope consequence or meaningful gap in that follow-up's scope.

Do not modify or review the unrelated unfinished task-creation Antigravity audit
`6gnqmm1yqhxz` or owner-created AC walkthrough task `6gnqadwhhz54`. Their presence
in the captured source is not part of this delivery.

## Intended contract to challenge

1. Epics-page `c` and New epic palette open the same staged form, from list or
   detail and an empty list. New task in this epic remains a distinct selected-ID
   action. Atlas/opening states do not invent a destination.
2. Nothing persists before explicit Create. Epic inputs use `NewEpicParams`,
   active/medium are visible initial choices (not core defaults), description is
   required and tags optional. Core/adapter own validation, numbering, scaffolding
   and persistence. Enum cancellation does not stage a candidate.
3. The common host shares UI mechanics, not an entity schema, domain-default map,
   universal receipt or filesystem authority. Task widgets/default-to-empty,
   parent picker and typed receipt policy remain task-specific. Model value-copy
   ownership must survive shared borrowed slices/callbacks.
4. Every write runs inside a tea.Cmd bound to the original service and copied
   inputs. Pending/uncertain/finalization/committed outcomes cannot replay or
   silently retarget. Explicit retry remains possible after known uncommitted
   refusals. Test commands must actually execute before counting side effects.
5. Late origin results may enrich recovery after a workspace switch but must not
   navigate, reload or write the incoming space. Opaque adapter IDs, not labels,
   `Location` or receipt paths, determine selection and inspection.
6. Terminal paste belongs to the first active visible text owner. Delayed paste
   rechecks kind, form/input generations and modal visibility; blink/cursor
   messages must not lose their update chain. Hidden/enum/pending/recovery fields
   cannot accept text or interpret it as global hotkeys.
7. Supported-small-screen recovery keeps destination, bounded record identity,
   non-replay notice and truthful actions visible ahead of verbose diagnostics.
   Full draft/receipt/error remain copyable; huge/Unicode identities and paths
   may be visibly abbreviated, never treated as authority. Tiny screens retain a
   non-authoring resize notice. Do not promise inspect for preview/unconfirmed data.
8. Success lands on the exact receipt ID despite status filters/watch refresh,
   including a non-active epic. Pathless capability does not promise an editor.

These are hypotheses, not a requested verdict. Checked ACs do not prove them.

## Mandatory evidence floor

Independently run focused task/epic/shared-host tests, then add hostile evidence
through a production caller (Model.Update and executed tea.Cmd, or the running
Bubble Tea program). Keep fixtures and probes in sandbox-local code or external
scratch. For your primary lens, include:

- A valid non-default epic (Unicode/punctuation title, tags, non-active status and
  non-medium priority), plus a refused submission. Check decoded durable values,
  generated scaffold and unchanged pre-existing bytes/modes independently; do not
  derive expectations from the implementation's enum/default helper.
- A reached post-submit boundary: committed cleanup or unknown outcome, then
  recovery Enter and/or a workspace swap. Execute returned commands, check write
  count and exact selected identity, and snapshot the non-target space. Constructed
  receipts need a real producer/wrapper trace or an explicit defense-in-depth label.
- An actual paste/resize/modal interaction with rendered evidence. Distinguish
  reducer event injection from physical-terminal decoding, direct PasteMsg from
  native clipboard results, and program-driven interaction from a unit-only state.

Select **one compiler-valid critical mutation** in your primary lens and require
the intended regression, not merely an unrelated test, to fail. Suggested targets:

- Codex: remove the per-kind displaced-origin receipt guard or alter the captured
  service; use `TestEpicCreateOriginAndWidgetBinding` and explain whether the outer
  session wrapper masks the edit. Coordinate a caller mutation if necessary.
- Antigravity: remove only the late-paste visibility/owner route in Model.update
  and require `TestCreationDelayedPasteRequiresVisibleOwner` to fail, or restore
  diagnostics-before-record rendering and require
  `TestEpicCreateVerboseRecoveryKeepsIdentityAndActionsVisible` to fail.

Record invariant, exact edit, command, intended assertion, killed/survived/invalid
outcome and restored green result. Compile/vet errors do not count as kills.
A surviving mutation is evidence about the test, not automatically a product bug.

Owner's prior self-review: a recovery-replay mutation **survived** the old epic
test because it discarded the returned command. The strengthened
`TestEpicCreatePortableAndNonReplayableReceipts` now executes commands, including
real committed-record inspection. Replacing locked Enter with unlock/resubmit
kills it with `create calls=2`; ordinary restored tests pass. This is a specific
warning about helper trust, not assurance that other counters are meaningful.

## Required hostile angles

- **Receipt truth:** false/empty success, pathless committed identity, DryRun,
  finalization wrapping a known sentinel, local committed-path evidence and
  read-only/unsupported/validation refusal. Differentiate the branches actually
  produced by core from malformed-fake defense in depth. Does every enabled action
  match the command it returns? Can inspection become another create?
- **Races/ownership:** swap before dispatch, after command creation, before late
  delivery, same ID in two spaces; stale field and cross-kind widget results;
  watcher arriving before landing loads. Test the real guard order, not fabricated
  model state alone. Higher-priority help/config ownership tests are useful
  defensive cases; label them if ordinary creation keys cannot reach them.
- **Interaction/rendering:** selector Enter/Tab/Shift+Tab/Esc, dirty Keep/Discard,
  empty epic bootstrap then task-in-new-epic, editable hotkey characters, pending
  navigation, narrow fields/enum/recovery and very verbose error. At 48x20 and the
  supported threshold 46x18, can identity/actions survive? At 30x10, is input inert?
  Verify fit and content, not just ANSI string length. Do not run OS clipboard tools
  on the user's clipboard; validate recoveryText/commands with a safe fake if needed.
- **Second systemic pass:** can shared value slices/callbacks mutate a prior Model,
  can helper settling skip precisely the I/O whose absence it claims, does a common
  mode/presentation accidentally impose task policy on epics, or can duplicated
  visible/input thresholds drift? Seek a demonstrated class of failure, not a new
  abstraction merely because more entity forms are planned.
- **Scope/accuracy:** README/help and owning comments must distinguish delivered
  task/epic flows from queued other forms; do not demand unrelated TUI editing
  capabilities or wire revisions for an internal-only presentation extraction.

## Validation and restoration

Owner ran build, full race suite, lint, planning/link and audit lint. Independently
run the focused suite and repeated runtime tests (commands below); inherit the full
gates unless your findings/fixes or observed baseline instability justify rerunning.
Report inherited checks as inherited, and any unavailable check as a limit.

```sh
go test -race ./internal/tui -run 'Test(EpicCreate|Creation|TaskCreate)' -count=1
go test -race ./internal/tui -run '^Test(EpicCreateRuntimePaste|TaskCreateInRealProgram|TaskCreateTerminalPasteInRealProgram)$' -count=3
```

Use task-specific cache/scratch directories outside the captured tree; set
GOCACHE/GOLANGCI_LINT_CACHE there if required. Do not change project toolchain pins
or weaken gates to accommodate the host. No new dependencies are needed. Normal
source build currently emits a nonfatal module stat-cache permission warning;
distinguish an actual failed command from that warning.

After probes, restore your sandbox to its captured baseline with **only your audit
changed**. Do not restore source HEAD (which lacks this implementation), touch the
source index/refs or push. Run audit/planning lint using a sandbox-built binary if
needed. Follow the injected verify/transfer/receipt protocol; retain the sandbox
until the owner confirms receipt. Stop/report a transfer conflict, never hand-merge
or overwrite the source deliverable. Reviewers do not implement production fixes.

## Deliverable

Preserve this brief; fill only the report below in your assigned audit. Findings
remain open using the injected grammar, with reachable producer/caller trace,
severity, minimal reproduction, expected/actual behavior and a bounded proposed
fix. Separate product defects, coverage gaps and design preferences. Consolidate
root causes. Include verdict/limits, verified consumer inventory, executed versus
inherited checks, hostile/mutation ledger, a few tested-negative or unresolved
hypotheses, and the mandatory isolation attestation. A blocked/incomplete review
is not an unqualified clean verdict. No findings is valid only with the selected
hostile evidence and restored baseline. Never fabricate finding placeholders.

## Reviewer report

Reviewed 2026-10-22 against the helper's complete dirty-source checkpoint, not source HEAD. **Verdict: changes requested.** The executed origin/write/receipt probes held, including the selected critical mutation. Native delayed clipboard delivery still edits a hidden creation form (M1). The second pass also disproved the stated prior-Model text ownership guarantee (coverage/contract gap L1). Both findings remain open; no production fixes or planning dispositions were made.

### Verified consumers and boundaries

Line numbers refer to the captured sandbox baseline. This inventory follows callers and producers rather than treating the brief's names or checked planning ACs as shipped evidence.

| Consumer / owner | Verified implementation and evidence |
| --- | --- |
| Epic controller | `internal/tui/epic_create.go:29` owns `epicCreateForm`; constructor at 57 visibly initializes active/medium. `submitEpicCreate` at 157 constructs `NewEpicParams`, copies strings/tags and captures the origin Service; `handleEpicCreated` at 177 owns typed receipt policy; `landCreatedEpic` at 208 selects the receipt ID; `recoveryText` at 213 retains draft and durable ID/path/error evidence. |
| Shared host | `internal/tui/staged_creation.go:30` owns `creationState`; 83 routes widget results with kind/form/input tokens; 102 defines synchronous borrowed `creationControls`; 117 defines `stagedCreation`; 230 selects the concrete `creationDraft`; 256 lands an exact `entityRef`. The shared host has no persistence opening or universal receipt. Fixed widget arrays do not imply deep ownership of Bubbles internals: L1. |
| Shared view / input | `internal/tui/staged_creation_view.go:14` owns `creationModal`, visibility at 17 and modal recheck at 19; 61/67 define `creationFrame`/`creationHeader`; 114 bounds diagnostics; 135/142 define field layout/rendering. `Model.Update` in `internal/tui/model.go:286` unwraps sessions; `update` at 318 routes native widget results at 348, including the faulty message-type split in M1. |
| Actual opening and shell consumers | Page-local dispatch: `internal/tui/model.go:906`; named palette dispatch: `internal/tui/command_dispatch.go:233`; modal priority: `internal/tui/overlay.go:37`; displacement: `internal/tui/session.go:117` and `closeTransientUI` at 192. Help/footer: `internal/tui/keys.go:78`, `internal/tui/view.go:361` and 417. `edit.go`/`palette.go` have comment changes, not added terminal-paste capability, confirmed by the captured diff. |
| Task consumer | `internal/tui/task_create.go:36` keeps the task widget array, selected parent and catalog; 140 adapts task-only default-to-empty enum policy; 159 handles text/query widgets. `internal/tui/task_create_commands.go:74` captures the selected-parent request and origin; 113 owns task receipt policy. `internal/tui/task_create_view.go:5` supplies task presentation and at 41 owns catalog recovery advice. |
| Core and real producer | `internal/core/service_epic.go:12`/25 own `NewEpicParams`/`Service.NewEpic`; `internal/core/selected_task_creation.go:46` owns `Service.NewTaskForSelectedEpic`. Entity-specific receipts and optional path evidence are in `internal/core/creation_receipt.go:18`, 35 and 42. `internal/store/create.go:70` owns the guarded create and at 103 wraps cleanup errors; `FS.CreateEpic` at 474 retains committed identity/evidence even when cleanup fails. No core port, CLI metadata or wire file changed in this slice. |
| Regression consumers | `internal/tui/epic_create_test.go:49`, 110, 155, 193, 276, 351, 398, 432 and 468 cover durable overrides, palette/watcher landing, refusals, bootstrap, receipts, origin binding, runtime paste and recovery rendering. `internal/tui/staged_creation_test.go:14` and 44 cover empty-value ownership and synthetic delayed terminal paste. `internal/tui/task_create_test.go:45` is the settling helper; 407/448 are running-program tests. The native result and mid-string cases below expose what these green tests do not establish. |

### Executed checks and hostile ledger

All Go commands used `GOCACHE=/private/tmp/codex-6gp2-review-gocache`. Scratch fixtures came from test temporary directories. None used the source planning tree as a fixture.

- `just build`: passed in the sandbox; the documented module stat-cache permission warning was nonfatal. Its fresh `bin/tskflwctl` was used for findings and lint.
- `go test -race ./internal/tui -run 'Test(EpicCreate|Creation|TaskCreate)' -count=1`: passed (45.598s), independently of the owner's run. Repeated after restoring all probes; final result below.
- `go test -race ./internal/tui -run '^Test(EpicCreateRuntimePaste|TaskCreateInRealProgram|TaskCreateTerminalPasteInRealProgram)$' -count=3`: passed (13.758s), independently. These running programs inject `tea.PasteMsg`; their names do not establish physical-terminal decoding or native Ctrl+V coverage.
- `go test -race ./internal/store -run '^TestOrdinaryCreateReportsCommittedGuardReleaseFailure$/epic$' -count=1 -v`: passed. This is the existing real guarded producer with an injected unlock failure, not a fabricated committed receipt (`internal/store/create_test.go:102`).
- `go test ./internal/tools/architecturecheck -count=1`: passed (5.965s). Planning lint and this assigned audit's lint also passed; final audit lint was repeated after report completion.
- Full `just test`, `just lint`, generated-doc and broader owner checks were **inherited from the assignment's stated owner results**, not rerun or represented as independent evidence. No command metadata changed. No physical terminal/PTY decode test or user clipboard access was performed.

The independent scratch probe source is retained outside the tree at `/private/tmp/codex-6gp2-evidence/codex_audit_probe_test.go`, SHA-256 `6f5477414687ede84e652ae33a588147b753c4ab88ef2b0a10e68593f9026739`. It was temporarily installed as `internal/tui/codex_audit_probe_test.go` in this sandbox, then removed before verification. Logs are retained beside it. These are reviewer-created probes, not committed project regressions. To repeat them, copy that file into a **new independent review sandbox** and run the commands recorded below; the two defect assertions intentionally fail on this captured implementation.

| Hypothesis / probe | Execution and observed result |
| --- | --- |
| Non-default durable inputs and copied request | `TestCodexAuditIndependentDurableAndCapturedRequest` passed. Through `Model.Update` keys/paste and an executed Create command, submitted `Hostile / 世界: q?!`, description `Description: α and punctuation!`, tags `alpha, 世界, audit`, retired status and high priority. Hard-coded expectations did not use the enum/default helper. Raw file bytes were independently YAML-decoded; the scaffold retained the original punctuation/Unicode title. Exact selected ID was `02-hostile-世界-q`. Mutating draft title/tags after command construction did not alter the issued request. A complete pre/post bytes/modes/directory snapshot differed only by the one new epic file. |
| Refused submissions and explicit retries | Existing required-description and real read-only adapter tests passed without durable changes. `TestCodexAuditMalformedReceiptDefenses` additionally executed result/recovery/retry commands for eight table cases: nil success with no evidence; committed with empty ID; local committed-path-only evidence plus conflict; finalization wrapping validation; unknown transport error; read-only; validation-labelled unsupported capability; and committed DryRun preview. Ambiguous/progress/finalization cases stayed locked with one create call; known refusals allowed an explicit second attempt; all fixture bytes/modes stayed unchanged. The malformed/inconsistent receipt cases are **defense in depth**, not asserted FS outputs. |
| Cross-space command binding despite identical IDs | `TestCodexAuditOriginSwapWithSameIDAndExecutedCommands` passed. A wrapped real alpha FS create was issued, beta was activated and its returned loads executed, then the old session-stamped write/result was executed/delivered. Both spaces contained `02-same-identity`. Alpha created once; beta create calls stayed zero. The real committed alpha receipt and appended cleanup diagnostic enriched displaced recovery. The result returned no incoming reload/navigation command. Recovery Enter was executed and stayed on beta's dashboard; beta's complete snapshot was unchanged. This probes command construction before dispatch and delivery after activation, not merely fabricated reducer state. |
| Real committed recovery actions and supported sizes | `TestCodexAuditLongIdentityRecovery` passed using the FS-backed wrapper in `epic_create_test.go:263` with appended `CreateFinalizationError` wrapping conflict. Forty Unicode title characters, a real long destination and 300 diagnostic repetitions retained Save to/Record/Create disabled/inspect/y copy/esc at 48×20 and 46×18. Width/height used Lip Gloss cell measurements and rendered content assertions. At 30×10 the resize notice appeared. The full recovery string retained the unabridged ID and diagnostic; after resizing, executing Enter's returned commands selected the exact receipt ID with one create call. The wrapper's post-commit error is injected; the independent store test above establishes the actual cleanup producer branch. |
| Native delayed clipboard after resize | `go test -race ./internal/tui -run '^TestCodexAudit' -count=1 -v` reproduced M1 and L1 while the two initial durable/recovery probes passed. The native production Ctrl+V command executed an isolated fake `pbpaste`, returning the genuine private `textinput.pasteMsg`. Reducer probes failed for task and epic resize; injected help/config cases also failed (defensive precedence states, not ordinary creation-key reachability). A running Bubble Tea wrapper held only result delivery until 30×10 rendered Resize, delivered it, resized to 46×18, and rendered the unintended title `CLIP-qhjkn-世界`. The runtime reproduction also failed in all three repetitions. No OS clipboard executable was invoked. |
| Empty-catalog repair advice | `TestCodexAuditExecuteEmptyCatalogAdvice` passed. Reached the empty task catalog's shipped “Create one on Epics (c)” advice, cancelled the clean task form, used page navigation and `c` to create the first epic, returned to Tasks, opened its picker, executed Ctrl+R, selected that exact new parent and created a task. This advice is executable when no draft needs preserving; it does not demonstrate carrying a dirty task draft between forms. |
| Prior Model ownership / systemic second pass | `TestCodexAuditPriorModelValueOwnership` failed for both kinds: paste `abcd`, Left, retain the prior Model, then paste `X` through `Model.Update`; prior title became `abcX` while next title was `abcXd`. This is L1. The existing shared ownership regression only pastes into an empty widget, so it misses Bubbles' mutable rune-buffer aliasing. No durable corruption or issued-command retargeting was demonstrated. |

Rendered evidence from the program-driven M1 reproduction after restoring 46×18:

```text
││  New epic                                ││
││  Save to: local                          ││
││  › Title                                 ││
││    CLIP-qhjkn-世界                    …  ││
││  [Cancel]  [Create]  ↕ more fields       ││
```

### Critical mutation and restoration

Invariant: a completed create from a departed workspace may enrich recovery but must not land/reload in the incoming workspace. In `internal/tui/epic_create.go:183`, removed exactly this entire result guard, leaving the function compiler-valid:

```go
if f.displaced || f.origin != m.svc || f.session != m.sessionGen {
    f.displace()
    if msg.err != nil {
        f.err += " Origin result: " + msg.err.Error()
    }
    return m, nil
}
```

Ran `go test -race ./internal/tui -run '^TestEpicCreateOriginAndWidgetBinding$' -count=1 -v`. **Killed** (exit 1), specifically `.../issued` at `epic_create_test.go:372`: “departed result lost receipt or issued incoming-space work.” It built and ran; no compiler/vet failure was counted. The draft subtest passed. A coordinated caller mutation was unnecessary: `Model.Update` at 289 deliberately passes old-session epic creation results to this handler, so the outer session wrapper does not mask this guard's removal. Restored the entire file from the sandbox checkpoint, then reran the same command: **passed** (3.393s), both draft/issued subtests green. Exact logs: `codex-6gp2-mutation-origin.log` and `codex-6gp2-mutation-restored.log` in the evidence directory.

No other production mutation was made. Findings were reproduced against restored production code, not the mutant. The scratch test file was removed before the final normal-suite run and helper verification.

### Tested negatives, second-pass conclusions and limits

- No service retarget, repeated Create, identity-by-path fallback, non-active-epic landing loss or beta-space write was observed in the executed probes. Existing palette, empty-list bootstrap, status-filter/watch ordering, stale input generation and cross-kind tests passed. The pathless portable receipt test at `epic_create_test.go:276` preserved the opaque `urn:epic:opaque` pending identity without promising an editor; that synthetic adapter does not prove a real remote backend exists.
- No evidence requires a universal form/receipt redesign. Per-kind requests and receipt policies are concrete; only mechanical controls/rendering are shared. README's delivered task/epic and queued other-kind distinction matches the captured dispatch. The older editor/palette/command-bar terminal-paste task `6gp2qmse4prt` is separate: M1 concerns delayed **native** clipboard delivery inside these creation forms, not its already-recorded immediate terminal-paste loss.
- `settleCreate` executes commands and batches before counting effects, but deliberately discards every `creationWidgetMsg` at `task_create_test.go:61`. It therefore cannot verify clipboard-result delivery. Native probes bypassed that discard with direct `Model.Update` and a running-program wrapper. Its skip is a coverage limitation, not a demonstrated skipped-write defect. Empty-value test helpers also concealed L1; the new epic consumer inherits this shallow-copy defect from task widgets rather than proving a newly introduced task-only regression.
- Normal direct terminal-paste tests and the generation guards remain useful but do not cover the private native-message branch. Cursor/blink messages still follow the non-paste update chain; a repair should preserve that chain. Help/config injection was explicitly labelled defensive; resize is the demonstrated ordinary runtime route.
- Actual filesystem cleanup failure was exercised by the existing store fault hook; the TUI recovery fixtures delegate real persistence then append a cleanup error. Unknown outcomes and malformed receipts are controlled adapter simulations. Recovery-text checks avoided the OS clipboard. No claim is made about physical keyboard decoding, a real remote service or complete platform coverage.
- Procedural disclosure: before `create`, I also ran read-only Git status/HEAD/file-name discovery and read `scripts/README.md` alongside the assignment. Those extra preflight reads exceed the brief's brief-and-copy-only rule. No implementation inspection, build, test, generator, source mutation or Git state change ran in the source checkout; all substantive review and probes ran in the independent sandbox. The helper's final attestation proves the captured/restored state and guarded single-file delivery, but does not erase that read-only protocol deviation.

### Final checks and isolation attestation

Final restored normal runs passed: the full focused race command above (46.166s) and repeated runtime command above (14.148s). `./bin/tskflwctl -C . lint` and `./bin/tskflwctl -C . audit lint 6gp2rf2gyp1a` passed. The assigned brief body matches the sandbox baseline exactly. Both findings were created with the sandbox-built `audit finding new`, remain open, and their managed date stamp is the only frontmatter addition. `git diff --stat` contains only this audit; no probe code remains in the captured tree.

The trusted source helper verified the independent Git directory, unchanged checkpoint/index, restored non-deliverable files, push guard and unchanged source deliverable. Its output is copied verbatim below. The report truthfully records `transfer=pending`; delivery is established only by the helper-owned receipt in the handoff. Keep `/private/tmp/isolated-review.aBsebl` until the owner confirms with the source helper's `receipt --sandbox /private/tmp/isolated-review.aBsebl` command.

```text
sandbox_path=/private/tmp/isolated-review.aBsebl
git_dir=/private/tmp/isolated-review.aBsebl/.git
baseline_commit=bae27fcb995cca36a25d8bfdae49ee129c1b0f92
source_blob=1d3e2143412c0f49c4f123650337cfeceb7675e9
source_fingerprint=b1b6f7586d9ae774ebfd26e207e1957478bccf17
deliverable=planning/audits/6gp2rf2gyp1a-2026-10-21-tui-epic-creation-shared-staging-implementation-codex.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Findings

#### M1. Native delayed clipboard results edit hidden creation forms · **Status:** fixed

**File:** internal/tui/model.go:350 | **Component:** tui
**Effort:** S · **Urgency:** soon

**Class / severity:** product defect, medium. A delayed native clipboard result bypasses visible text ownership for both creation kinds. This is distinct from the deferred older-input terminal-paste task.

**Reachable producer/caller trace:** open a creation form with `c`; send Ctrl+V through `Model.Update`. The focused Bubbles widget returns its native paste command, wrapped by `routeCreationWidget` (`internal/tui/staged_creation.go:83`). In pinned `charm.land/bubbles/v2 v2.2.1`, `textinput/textinput.go:788` calls the clipboard provider and returns private `textinput.pasteMsg` (declared at line 23), not `tea.PasteMsg`. `internal/tui/model.go:350` recognizes only the latter before invoking `creationModal.handlePaste`; the native result goes directly through line 353 to the kind's `updateInput`. `editable()` checks form mode/saving/lock state, not terminal size or top-modal ownership. Resize does not advance the field generation. Thus a matching native result edits an invisible field.

**Minimal reproduction:** open Epics `c` at 120×40; issue Ctrl+V and hold command-result delivery; deliver WindowSize 30×10 and verify the rendered Resize notice; execute/deliver the pending native result; return to 46×18. Expected: title remains empty throughout the hidden interval. Actual: the retained title and rendered field become `CLIP-qhjkn-世界`. The command has no persistence side effect, but changes the next intended Create's draft without visible input ownership.

**Executed evidence:** reviewer-local `TestCodexAuditNativeClipboardHiddenOwner` fails for epic/task resize. Help/config injections fail too, explicitly defensive precedence cases. `TestCodexAuditRuntimeNativePasteAfterResize` reproduces the ordinary resize route in a running Bubble Tea program, holding only delivery of the executed production command; it fails in all three repeated runs. A test-local PATH contains only a fake `pbpaste`, so the actual Bubbles result type is exercised without reading or changing the user's clipboard. Logs and replayable probe code are retained at `/private/tmp/codex-6gp2-evidence/`; use `go test -race ./internal/tui -run '^TestCodexAudit(NativeClipboardHiddenOwner|RuntimeNativePasteAfterResize)$' -count=1 -v` after placing the retained probe file in a new independent sandbox. These are program-driven native-provider probes, not physical-terminal decoding.

**Why current green evidence misses it:** `TestCreationDelayedPasteRequiresVisibleOwner` (`internal/tui/staged_creation_test.go:44`) manufactures `tea.PasteMsg`, and existing runtime tests inject that public terminal message as well. Neither reaches the private native branch. `settleCreate` deliberately drops widget-result delivery, so it is unsuitable for verifying this path.

**Recommendation:** Apply visible text-owner checks to native delayed editing results as well as terminal PasteMsg; preserve the separate cursor/blink update chain and add a real native-provider resize regression.

**Resolution:** Accepted. Model.update now gates every non-blink widget result
    on visible creation input ownership, including private native clipboard
    messages; cursor ticks retain their maintenance chain. Darwin regressions
    execute real Ctrl+V through a fake pbpaste with visible, resized,
    higher-modal and stale-field cases plus held delivery in a running program.
    Restoring the old public-PasteMsg-only split fails the intended test; normal
    native/ownership/runtime regressions pass three race repetitions.

#### L1. The shared ownership test misses non-empty widget buffer aliasing · **Status:** fixed

**File:** internal/tui/staged_creation_test.go:14 | **Component:** tui
**Effort:** S · **Urgency:** eventually

**Class / severity:** coverage/contract gap, low. The new shared ownership regression and owning guide overstate what copying the widget array guarantees. The probe demonstrates observable mutation of a retained prior Model, but does not demonstrate corruption in the single-current-Model runtime or in an issued creation request.

**Producer/caller trace:** both forms hold fixed arrays of `textinput.Model`, and `Model.Update` copies the containing Model. Epic input updates at `internal/tui/epic_create.go:102` and task updates at `internal/tui/task_create.go:170` invoke Bubbles' value-receiver `Update`; the copied widget still shares its internal rune slice. In pinned Bubbles v2.2.1 `textinput/textinput.go:330`, insertion takes `head := m.value[:m.pos]` and appends into that backing buffer at line 337. Fixed-size arrays detach the widget structs, not those internal buffers. The common host's value-ownership comment (`internal/tui/staged_creation.go:100`) and `docs/architecture/tui.md:108` need evidence covering non-empty edits.

**Minimal reproduction / expected versus actual:** for either form, paste `abcd` through `Model.Update`, send Left, keep `prior := m`, then deliver `tea.PasteMsg{Content: "X"}` to get the next Model. Expected under the advertised ownership contract: prior title `abcd`, next title `abcXd`. Actual: prior title `abcX`, next title `abcXd`.

**Executed evidence:** reviewer-local `TestCodexAuditPriorModelValueOwnership` fails for task and epic with `prior="abcX" next="abcXd"`; its source and original failure log are in `/private/tmp/codex-6gp2-evidence/`. Run `go test -race ./internal/tui -run '^TestCodexAuditPriorModelValueOwnership$' -count=1 -v` after installing that retained probe in a new independent sandbox. Existing `TestCreationSharedStatePreservesValueOwnershipAndRecovery` at `internal/tui/staged_creation_test.go:14` passes because it only inserts into an empty input, which allocates a fresh buffer and conceals this class of aliasing. This is inherited from task widgets and now shared by the new epic consumer, not evidence of a newly introduced task-only regression.

**Bounded impact:** the independently executed captured-request probe still persisted its originally copied title and tags after later form changes. Do not infer write replay, cross-workspace mutation or durable corruption from this snapshot result. The remaining work is to make the claimed prior-state ownership true and regression-tested, or explicitly reconcile the contract with the actual supported ownership semantics before further form consumers rely on it.

**Recommendation:** Pin prior-model immutability with a non-empty mid-string edit and detach mutable widget contents where that contract requires it; reconcile owning comments with the supported semantics.

**Resolution:** Accepted as a bounded snapshot/coverage gap, not durable
    corruption. updateCreationInput detaches mutable text through the public
    SetValue API before editing, preserving cursor/focus/settings/error; pure
    blink ticks avoid copying. Non-empty mid-string paste, typing, backspace and
    deletion regressions cover task, epic and parent-query inputs. Omitting the
    detachment fails all twelve cases; owning comments distinguish draft text
    from shared animation bookkeeping.
