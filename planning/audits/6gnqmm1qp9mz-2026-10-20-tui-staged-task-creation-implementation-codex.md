---
schema: 1
id: 6gnqmm1qp9mz
bucket: closed
area: tui-staged-task-creation-implementation-codex
date: "2026-10-20"
updated_at: "2026-10-20"
---
# Audit: Staged TUI task creation and extensibility — codex — 2026-10-20

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

Adversarially review the first staged TUI task-creation implementation and its
maintainability before extending creation to other entity kinds. The owner has
tried the interaction and approved its overall direction, including edit-style
constrained selectors. This is an implementation/safety/extension-boundary review,
not a request to redesign the approved interaction or implement more entity kinds.

Complementary primary lenses:
- Codex: adapter boundaries, extension seams, per-kind variance, shared-widget and
  test-oracle quality. Cross-check one receipt/session safety path.
- Antigravity: event ordering, wrong-target/replay hazards, real rendered keyboard
  behavior, selection restoration and constrained terminals. Cross-check one
  portable-capability/extension assumption.

Both inspect the shared contract independently. Do not read the sibling report
until your own report is final. No finding quota or presumed clean verdict.

## Review target

SOURCE_ROOT: /Users/andyeschbacher/git/andy-esch/taskflow.
Base: f746f8618c2c62fac6fb958d94e081cacdda2d5b.
Implementation is UNCOMMITTED on main. Capture staged, unstaged and untracked
files; HEAD alone does not contain this feature.

Implementing task:
planning/tasks/6g4ym5y2dkp8-create-tasks-from-the-tui-and-epic-context.md.
Approved design:
planning/research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md.
Future extension scope:
planning/tasks/6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md.
Owning contracts: AGENTS.md, docs/architecture/tui.md (including staged-task-creation
and extending-creation-to-other-kinds), planning-data.md and boundaries.md.

Verify actual symbols before citing them:
- internal/tui/task_create.go: taskCreateForm, taskCreateMode, newTaskCreateForm,
  taskCreateEnumOptions, routeTaskCreateWidget, handleTaskCreateKey, recoveryText.
- internal/tui/task_create_commands.go: taskCreationActions, openTaskCreate,
  loadEpics, applyTaskCreateEpics, submitTaskCreate, handleTaskCreated,
  taskCreateKnownRefusal, landCreatedTask.
- internal/tui/task_create_view.go: taskCreateModal, taskCreateBox, pickerView.
- internal/tui/model.go: Update/update, sessionMsg and widget-result guards, Create key.
- internal/tui/session.go: activateWorkspace, closeTransientUI and scopeSession.
- internal/tui/command_dispatch.go, palette.go, overlay.go, keys.go, help.go, view.go:
  action discovery, modal precedence, keyboard ownership and footer/render integration.
- internal/tui/edit.go: editEnumOptions and enumInline (also consumed by existing edits).
- internal/tui/task_create_test.go: actual regressions named below, createReceiptPort,
  createPortableReads and settleCreate. Its skipped cursor ticks and direct input
  setters are limitations, not substitutes for running-program evidence.

Build a repository-wide consumer inventory with rg: action/key dispatch, modal
ownership, every creation command/result consumer, source IDs, session routing,
core creation use cases/receipts and shared edit widgets. The already-merged
SelectedEpicTaskCreationStore/NewTaskForSelectedEpic is a prerequisite, not newly
implemented here. Other entity creation and a universal form/CreateEntity API do
not ship in this patch. The owner-created 6gnqadwhhz54 AC walkthrough task is not a
shipped test fixture or evidence that a scenario has been executed.

## Intended contract to challenge

1. Page-local c on Tasks opens a staged task form in list/detail, even with no
   selected task. Epic-page c must not create a task. The named epic-context
   palette action preselects an exact adapter-established source ID, not a label.
2. Title/epic/tags/description and optional metadata are in-memory until explicit
   Create. Enter in text advances; enum/picker acceptance only stages values.
   Priority/tier/autonomy share edit-style bounded choices; default leaves core
   defaulting intact. Preserve existing e writes and n/N search semantics.
3. Core owns defaults, validation, selected-parent revalidation, authorization,
   scaffolding and persistence. TUI calls NewTaskForSelectedEpic in tea.Cmd with
   copied inputs and the originating Service. No CLI subprocess, direct planning
   filesystem access, reference fallback or opaque-Location path inference.
4. Pending submission blocks repeat writes and ordinary navigation. Clean cancel
   writes nothing; dirty cancel defaults to Keep editing. Known uncommitted
   refusals retain input for explicit retry; do not guess field attribution by
   error strings. A malformed/unknown receipt is not permission to create again.
5. Committed cleanup failures, CreateFinalizationError (including conflict-wrapped
   ones), unknown outcomes and unconfirmed success disable replay and retain
   receipt/draft evidence. Local paths are display evidence only; inspect/editor
   routes require existing exact-identity/optional-local capabilities.
6. A pending workspace open blocks a new form. Activation invalidates submission
   authority but preserves a read-only recovery copy. Already-issued commands
   retain origin authority; late origin results may update that copy, never
   navigate/reload/write the incoming workspace. Matching labels/IDs/generations
   in two workspaces do not transfer authority.
7. Form, catalog, focused-widget and workspace generations have distinct jobs.
   Late catalog/clipboard/cursor results cannot edit another field/form or replace
   a newer catalog. Watcher/status/filter reloads cannot steal created-task focus.
8. Rendering respects width/height and keeps destination, actionable input,
   errors and explicit actions usable. Tiny resize notices are not invisible
   mutation surfaces. Ctrl+C remains force-quit, not a durable-draft promise.
9. Extension reuses mechanics, not task semantics: private task row IDs/modes,
   typed request/receipt handling and parent/lifecycle assumptions are not a
   universal entity schema. A second real kind should justify shared-host
   extraction under tests. No broad Store or core port expansion is warranted
   just to reuse a modal. Challenge the boundary with actual New* use cases.

## Mandatory evidence floor

Both independently run the focused baseline in the sandbox:
go test -race ./internal/tui -run '^TestTaskCreate|^TestEdit|^TestPalette' -count=1

Each must exercise a valid-but-hostile non-default creation and a refused attempt
through actual key/action -> tea.Cmd -> Service paths. Verify selected source ID,
decoded durable metadata, untouched existing bytes/modes and absence of writes
before Create. Do not derive expected values from the same shared form helper.
Use disposable roots only; no writes to the owner's planning space or registry.

Codex must:
- Independently verify portable/absent/read-only capabilities and opaque selected
  identity using TestTaskCreateUsesPortableCapabilityAndOpaquePickerIdentity and
  TestTaskCreateReadOnlyAndEmptyEpicCatalog, plus a hostile independent fixture.
  No local path may be required or promised for the pathless case.
- Compare real defaults and deliberate metadata overrides against independent
  expected values. Check both creation and existing e enum behavior; a shared
  helper making both wrong must not count as corroboration.
- Perform a compiler-valid mutation removing only the focused-input generation
  comparison in Model.update's taskCreateWidgetMsg branch. Require
  TestTaskCreateDelayedWidgetResultsCannotEditAnotherField to fail, restore green.
- Trace actual NewEpicParams/NewAuditParams/NewResearchParams/NewThreadParams and
  their receipts. Give a bounded change sketch for adding audit area/date and
  Thread guarded member selection. Identify concrete task-specific coupling
  versus safe shared mechanics. Do not implement those forms or demand a generic
  framework; classify a proposed future extraction honestly.
- Cross-check one committed/uncertain receipt and one displaced-origin result.

Antigravity must:
- Run TestTaskCreateInRealProgram with -race -count=3, not only reducer tests.
  Exercise selector stage/cancel/default and dirty cancellation with real keys.
- Independently challenge result order: origin write delivered after activation,
  a stale catalog after reload/reopen, a late widget result after field change,
  and watcher reload superseding a pending created-task landing. Use same IDs in
  different spaces to make accidental reference/session matching observable.
- Perform a compiler-valid mutation discarding landCreatedTask's pending tab.restore
  target. Require TestTaskCreateWatcherCannotSupersedeCreatedLanding to fail,
  restore green. Also attempt a bounded duplicate-submit/recovery mutation; if
  another guard rescues it, identify that guard and coordinate the mutation rather
  than claiming the first edit tested replay prevention.
- Probe committed cleanup, unknown, incomplete-success and uncommitted-finalization
  outcomes via actual command result delivery. Independent port counters and tree
  snapshots must show no remint/replay or incoming-space mutation.
- Render normal, epic/enum picker, dirty-confirm, validation-error and recovery
  modes at narrow sizes with long/Unicode content. Distinguish fit from usable
  visible target/actions; clipping alone is not a correct interaction oracle.
- Cross-check the selected capability/no-path boundary and one actual other-kind
  use case; do not assume an entity-neutral recovery policy from a task receipt.

Keep a short mutation ledger: exact compiler-valid edit, intended invariant/test,
observed killed/survived/invalid result and restored pass. Compiler failures and
unrelated test failures are not killed mutations. No finding-count target.

## Required hostile angles

After the checklist, do a distinct systemic pass. Settle each important hypothesis
as reproduced, falsified with evidence, or unresolved:
- Can duplicated form/command guards or test helpers accidentally mask another
  broken guard? Which booleans are intentionally orthogonal, versus exclusive
  taskCreateMode states? Trace reachable event sequences, not arbitrary private
  state fabrication passed off as a product defect.
- Could added kinds require cloning task parent pickers, ready-to-start behavior,
  TaskCreationReceipt or root-model branches? What is the smallest maintainable
  extraction with two real consumers? Separate current defects from future work
  already assigned to 6gmshzwphadw; code-size taste alone is not a finding.
- Can raw editor drift, same-label siblings, duplicate/missing canonical IDs or
  partial catalogs substitute a displayed selection? Parent whole-stem identity,
  non-reserving previews, later deletion and same-ID replacement limits are owned
  by the merged selected-epic contract, not invented database guarantees.
- Can recovery evidence be lost, normalized into a fresh create, or routed to the
  wrong workspace? Can failure/read-only/empty catalogs cause a hidden write?
- Does a successful landing actually select by canonical key after status/filter,
  async refilter and watcher races? Does a generic error footer promise an editor
  that a pathless adapter cannot supply?
- Could widget key ownership leak q/h/j/k/n into navigation, or could enum accept,
  Enter advance, discard confirmation or resize silently save/discard typed data?

No CLI metadata, machine wire/schema or persisted schema changed here. Do not
invent conditional wire coverage requirements. Existing watcher/session hardening
task 6gmbfxy96tpw is separate; an adjacent issue needs a concrete new-creation
consequence, not duplicate tasking or scope-by-association.

## Validation and restoration

Use the injected independent-clone protocol BEFORE implementation inspection or
commands. Source is read-only; only the assigned final audit may be transferred.
Reviewers do not share .git metadata, read sibling reports, commit beyond the
sandbox baseline, stage/restore the source, push, create owner tasks or manually
copy reports. On protocol/evidence failure retain the sandbox and report incomplete.

Keep GOCACHE/GOLANGCI_LINT_CACHE and runtime fixture roots in reviewer-owned temp
directories outside the sandbox corpus. Never change HOME/CODEX_HOME. Temporary
probe Go files may live in the sandbox, but restore/delete only your artifacts
before verification. All planning fixtures and registry configuration must be
disposable; do not register review spaces in the user's system-wide config.

Owner baseline after the quality pass: fresh just build; full just test (-race);
just lint; planning/link lint; audit lint and diff checks passed. Build emitted a
nonfatal module stat-cache permission warning but exited 0. Full baseline gates may
be inherited; focused probes above must be independent. Run affected wider tests
if your hypothesis touches shared edit/shell behavior. No docs generation is needed
unless command metadata actually changes in a restored/final patch.

## Deliverable

Keep the brief intact; replace only the Reviewer report slot. Findings first,
all OPEN, with root cause, reachable producer/event trace, consequence, exact
reproduction/evidence, recommended direction, scope and residual uncertainty.
Distinguish product defects, coverage gaps, preferences and future extraction.
Preserve valid out-of-scope observations as concrete follow-up proposals in this
audit; do not create other files in the source. A ready/no-findings verdict needs
the evidence floor, not just green tests or a claim the architecture looks clean.

Include verified consumer inventory, commands/results, concise mutation ledger,
extension-boundary assessment and surviving hypotheses. Unexecuted probes must be
reported as unknown. Report the final independent-workspace verification attestation
(transfer=pending); hand back the guarded transfer receipt so the owner can confirm
delivery before triage. Never mark your own findings fixed/tracked/settled.

## Reviewer report

## Findings

#### M1. Terminal paste is dropped across the staged creation form · **Status:** fixed

**File:** internal/tui/model.go:321 | **Component:** tui input routing
**Effort:** S · **Urgency:** soon

**Product defect; reachable runtime input loss.** With the new task form open and
Title focused, a normal bracketed terminal paste arrives as `tea.PasteMsg`.
`Model.update` routes only `tea.KeyPressMsg` through modal ownership
(`internal/tui/model.go:321`) and only command-wrapped `taskCreateWidgetMsg` to
the form (`:331`). Direct paste falls through to the active list (`:636`), never
reaching `taskCreateForm.updateInput` (`internal/tui/task_create.go:218`). The same
event class is lost in Tags, Description, Effort and the epic search query. This
can silently omit optional metadata and leave core defaults in the eventual task,
or cause unexplained title/tag validation refusals. No persistence bypass or
wrong-parent write was demonstrated.

**Executed reproduction:** reviewer-only `TestReviewerBracketedPasteInProgram`
opened a real `teatest.NewTestModel` Program at 120×40 over a disposable independent
fixture, sent actual `]`, `c` keys, waited for `Nothing saved until Create`, sent
`tea.PasteMsg{Content: "Pasted title 世界"}`, then force-quit and checked the final
model. `go test -race ./internal/tui -run '^TestReviewerBracketedPasteInProgram$'
-count=1 -v` compiled and failed: `runtime bracketed paste lost: got "", want
"Pasted title 世界"`. Existing bytes/modes were unchanged. A second literal-key
probe failed separately for all five destinations. Production code was unchanged
for both failures.

Compact reproduction, also executed against restored code as
`TestTaskCreateTerminalPaste` (race run failed, `paste lost: ""`, 0.889s). Append
temporarily to `internal/tui/task_create_test.go`; all imports/helpers exist there:

```go
func TestTaskCreateTerminalPaste(t *testing.T) {
    tm := teatest.NewTestModel(t, newModel(t), teatest.WithInitialTermSize(120, 40))
    tm.Send(press("]"))
    tm.Send(press("c"))
    teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
        return bytes.Contains(out, []byte("Nothing saved until Create"))
    }, teatest.WithDuration(5*time.Second))
    tm.Send(tea.PasteMsg{Content: "Pasted title 世界"})
    tm.Send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
    final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
    if got := final.createTask.inputs[taskCreateTitle].Value(); got != "Pasted title 世界" {
        t.Fatalf("paste lost: %q", got)
    }
}
```

**Causal evidence:** a temporary direct-paste dispatch case forwarding to active
creation's `updateInput` made the runtime and all five destination probes pass
(race run, 8.405s); restoring the original model restores the missing route. This
was a diagnostic intervention, not a final fix. The pinned dependencies confirm
the event is real: `go.mod:7` selects Bubble Tea v2.1.0; its installed `input.go:36`
converts `uv.PasteEvent` to `PasteMsg`; Bubbles v2.2.1 `textinput.go:654` handles
`tea.PasteMsg`. The dependency sources were read locally in the review environment.

**Test-oracle gap and scope:** `fillCreate` directly sets values
(`task_create_test.go:78`); `settleCreate` intentionally skips widget messages
(`:61`); the real-program test types each rune (`:415`, `:422`). None supplies this
terminal event class, so green creation tests do not corroborate terminal paste.
Handle live input events through the active modal owner while preserving pending,
locked/displaced, picker-mode and tiny-terminal restrictions. Do not forward paste
to every widget or confuse an immediate runtime paste with a delayed native
clipboard command that needs generation guards. Add live Program coverage for
text and picker paste plus refusal in blocked modes before sharing the form host.
Existing e/command-bar paste behavior was not independently reproduced; this
finding is bounded to the newly added creation flow. Physical clipboard/terminal
escape parsing remains untested; the Program message path and underlying event
conversion are verified. No OS clipboard tool is needed to reproduce it.

**Recommendation:** Route direct terminal paste to the active creation input with modal, size and submission guards; pin it with runtime-event regressions.

**Resolution:** Confirmed the direct tea.PasteMsg dispatch gap with failing
    regressions. Added first-active-modal paste ownership, visible-input gating
    and editable-mode guards; terminal paste now reaches creation text fields
    and epic search without leaking through blocking modes.
    TestTaskCreateTerminalPasteInRealProgram persists independently expected
    title, tags, description and effort through actual runtime messages;
    TestTaskCreateTerminalPasteOwnership covers search, precedence, no partial
    writes, pending/recovery/displaced/enum/discard/tiny states. Full race suite
    and lint pass; delayed clipboard generation guards remain intact.

## Review evidence

**Verdict: changes requested; one open product defect.** Both technical review
passes are complete. The adapter boundary and deferred extraction direction are
sound for this first kind, but the live input boundary drops terminal paste (M1).
No implementation fix or additional entity form is included in this report.

**Protocol qualification:** before creating the sandbox I also enumerated source
filenames and read `scripts/README.md`. That exceeded the brief's restriction to
reading this brief and performing the initial copy. No implementation content was
inspected, test executed, or source file changed before isolation. All substantive
inspection, builds, probes, restoration and report editing occurred in the
independent clone. Strict protocol compliance is therefore **incomplete**, despite
the helper verification below passing. The workspace is retained for owner review;
this is not an unqualified protocol-complete or ready verdict. I did not read the
sibling report.

### Verified consumer inventory

Repository-wide `rg` searches covered `createTask|taskCreate`,
`NewTaskForSelectedEpic|SelectedEpicTaskCreationStore`, creation result types,
`new-task|new-task-in-epic|keys.Create`, `editEnumOptions|enumInline`, and every
`New*Params`/creation receipt named below. Locations refer to the captured baseline,
with all production probes restored.

| Boundary | Verified producers and consumers |
| --- | --- |
| Action discovery and dispatch | `internal/tui/keys.go:78`, `help.go:45`, `view.go:364`, `view.go:495`; `palette.go:23` adds `palAction`; `command_dispatch.go:170` indexes actions and `:237` dispatches them; `model.go:886` routes page-local c. `task_create_commands.go:19` declares Tasks and selected-epic palette actions, both reaching `openTaskCreate` at `:39`. |
| Modal ownership/rendering | `overlay.go:31` registers `taskCreateModal`; `model.go:796` handles force quit before modal precedence. `task_create_view.go:17` gates keys at tiny sizes, `:33` renders modes, `:158` bounds the box, `:182` renders the epic picker. `view.go:418` supplies creation/recovery footer text. |
| Draft and focused widgets | `task_create.go:19` contains private task rows; `:40` declares mutually exclusive `taskCreateMode`; `:49` owns form state; `:99` supplies bounded choices; `:117` constructs widgets; `:176` increments focused-input generation; `:201` routes delayed widget commands; `:218` updates text/query; `:245` owns keys; `:385` formats recovery evidence. `model.go:331` consumes widget messages. |
| Catalog and selected identity | `task_create_commands.go:65` captures origin Service and form/catalog generations; `:75` consumes results and validates the complete returned key set. `task_create.go:231` filters and `:296` accepts `EpicSummary.Source.ID`. `entity.go:46`, `:91`, `:109` separate keys/labels and validate/disambiguate them. Core's epic display projection itself normalizes the display ID to source ID (`internal/core/service_epic.go:192`); contradictory declared IDs are not independently trusted picker identities. |
| Submission/results | `task_create_commands.go:98` converts to copied `NewTaskParams` and captures `EpicMutationTarget{ID: f.epic.key}` plus origin Service in the command; `:141` consumes `taskCreatedMsg`; `:180` classifies known refusals; `:185` lands by receipt Task.ID, carrying `tab.restore`. `model.go:324` and `:328` are normal catalog/create result consumers; `:288` is the special stale-session create-result consumer. No other production consumer was found by the searches. |
| Session and authority | `session.go:16`/`:21` define `sessionMsg`/`scopeSession`; `:117` activates a workspace, increments generation and invokes `closeTransientUI` at `:192`. That calls form displacement, retaining origin authority/evidence rather than rebinding. The stale-session result branch delegates to the form's gen/session/saving and origin checks. |
| Core/persistence prerequisite | `internal/core/selected_task_creation.go:23`, `:46` define the optional selected-parent port and one-attempt use case. `service.go:369`, `:408` discover/bind capabilities to source sets. `service_task.go:963` owns preparation/defaults. `creation_receipt.go:9`, `:18`, `:35` own finalization errors, optional local evidence and typed task receipts. `internal/store/create.go:188` revalidates exact parent identity under the create guard, and `:213` preserves ordinary guarded creation. These were merged prerequisites, not new APIs in this patch. |
| Existing shared edit consumers | `internal/tui/edit.go:61`, `:64`, `:91` consume `editEnumOptions` (`:98`) for task priority/tier and epic priority. Creation adds default plus priority/tier/autonomy choices. `enumInline` at `:461` is consumed by the existing editor (`:441`) and creation view (`task_create_view.go:117`). Existing e does not offer an autonomy field; no such feature is claimed. |
| Shipped test oracles | `task_create_test.go:272`/`:300` define `createReceiptPort`/`createPortableReads`. `settleCreate` at `:45` drops wrapped widget results; `fillCreate` at `:78` sets inputs and parent directly; `chooseCreateOption` at `:87` consults production options. These are useful reducer/service helpers, not runtime event coverage. The real-program test at `:403` types individual keys, not paste. Independent literal-value and running-program probes supplemented them. |

### First pass: executed evidence

All commands ran from `/private/tmp/isolated-review.OKDTdg`. Go caches and runtime
fixtures were outside the corpus under `/private/tmp/taskflow-codex-review-cache`
and `/private/tmp/taskflow-codex-review-runtime`; no review space was registered.
The first focused run used the OS temporary directory for test roots, also outside
the sandbox. Temporary probe code was removed before final verification; an
unshipped copy remains at
`/private/tmp/taskflow-codex-review-runtime/reviewer_probe_test.go` for inspection.
Its names below identify executed reviewer probes, not committed regression tests.

| Command/probe | Result |
| --- | --- |
| `go test -race ./internal/tui -run '^TestTaskCreate|^TestEdit|^TestPalette' -count=1` | Passed initial baseline (19.709s); rerun after final restoration recorded below. This includes `TestTaskCreateUsesPortableCapabilityAndOpaquePickerIdentity` and `TestTaskCreateReadOnlyAndEmptyEpicCatalog` as required. |
| `go test -race ./internal/tui -skip '^TestReviewer' -count=1` | All shipped TUI tests passed (30.703s), broadening coverage to the actual `TestModel_Edit*` tests that the mandatory regex alone does not select. |
| `go test -race ./internal/core ./internal/store -run '^TestSelectedEpic|^TestTaskCreationPreservesIndependentExpectedValues$|^TestService_NewAudit$|^TestService_NewEpic$|^TestNewResearch_|^TestServiceNewThread' -count=1` | Passed: core 1.284s, store 1.314s. Includes actual other-kind use cases, authoritative Thread membership, selected-parent drift/read-only/finalization and independent expected creation values. |
| `TestReviewerHostileDurable` | Passed defaults, deliberate overrides and a missing-tags refusal through literal keys → returned tea.Cmd → Service → real FS. No direct setters, focus setters, `fillCreate`, `chooseCreateOption`, or production-derived expected option values. |
| `TestReviewerHostilePortable` | Passed pathless success, absent selected-create capability and known uncommitted conflict. Two records had contradictory declared IDs and opaque Locations; the selected source remained `urn:two`, port target and task epic were exactly `urn:two`, and port count was one (zero for absent capability). No local path/editor promise or reference fallback was available. |
| `TestReviewerExistingEditOracle` | Passed literal e-key flow: existing priority low → high and tier 2 → 5, decoded from the durable record. Expected values did not come from the shared enum helper. |
| `TestReviewerReceiptAndDisplacement` | Passed actual committed creation plus injected cleanup failure, both current-space and delivered after activation. Independent selected-create counter stayed one; recovery retained committed ID/path and error. Both roots used the same epic IDs. Late origin delivery returned no incoming command, retained recovery without navigation, and left incoming bytes/modes unchanged. |
| `TestReviewerBracketedPasteInProgram`, `TestReviewerPasteAllCreationInputs` | Failed on restored production code as intended: empty values after terminal paste in the running program and all five text/query destinations. Diagnostic routing intervention made both pass; it was removed. See M1. |
| `just build` | Passed with a nonfatal denied Go module stat-cache write outside the sandbox. The freshly built CLI created M1 and validated the audit; no installation or registry change. |

The hostile durable fixture contained distinct `01-decoy` and `02-target` epics
with matching descriptions/body headings, unknown frontmatter and a comment, both
mode 0640. Literal picker search selected `02-target`. Title was
`Boundary / qhjkn 世界`; description was `Exact durable description α`; tags decoded
as exactly `[one, two]`; created date was independently fixed to `2026-10-20`;
status decoded as ready-to-start; scaffold retained the full title. Defaults
decoded as priority medium, effort Unknown, tier 3, autonomy 3. Deliberate overrides
decoded as low, `2h / α`, 5, 1. Tree snapshots before Create and after command
construction were identical; successful execution added exactly one file and
preserved every pre-existing file's bytes and mode. Missing tags refused with
the title retained and no filesystem delta. Literal typing discarded only its
perpetual blink commands; it is not asserted to simulate runtime scheduling.

The focused baseline also executed the shipped unknown/unconfirmed/finalization
receipt cases (`TestTaskCreateReceiptRecoveryNeverReplays`, `:552`), including
conflict-wrapped finalization; displaced drafts/issued writes (`:590`); invalid
catalog and stale loads (`:356`); parent rename refusal (`:246`); watcher landing
(`:382`); dirty cancellation (`:190`); selectors (`:492`); and small-terminal fit
(`:631`). Fit assertions alone are not evidence of visible action usability in
every recovery/Unicode rendering mode. No changed wire/schema branch or shipped
repair command exists in this patch; none is invented as a coverage obligation.

### Compiler-valid mutation ledger

| Exact bounded edit | Intended oracle and observed result | Restoration |
| --- | --- | --- |
| Removed only `&& msg.inputGen == m.createTask.inputGen` from `Model.update`'s `taskCreateWidgetMsg` condition (`model.go:332`). | `TestTaskCreateDelayedWidgetResultsCannotEditAnotherField` failed at `task_create_test.go:478`: departed field's async result edited newly focused field. **Killed**, not a compiler failure. | Exact original bytes restored; targeted race test passed (2.137s). |
| Replaced the first `if f.saving` in `handleTaskCreateKey` (`task_create.go:247`) with `if false`, leaving the controller guard intact. | `TestTaskCreateFromTasksAndEpicContext` **survived** (3.382s). `submitTaskCreate` at `task_create_commands.go:100` rescued duplicate submission. This does not prove the key-layer navigation guard is individually pinned. | Followed by the coordinated probe below. |
| Retained the previous edit and also removed `f.saving ||` from `submitTaskCreate`'s early return. | Same test failed for both tasks/epic contexts at `:134`, `duplicate Create issued a second write`. **Killed**; both edits compiled. The test observed a second write command, without needing to execute it against persistence. | Both files restored byte-for-byte; the same test passed (1.76s), followed by the full TUI suite and final focused run. |
| Diagnostic intervention: added a direct `tea.PasteMsg` case forwarding to active creation's `updateInput`. | Previously failing runtime and five-destination paste probes passed (8.405s). This localizes M1 to dispatch, rather than widget insertion or a clipboard tool. It is a causal probe, not a reviewed fix: a real fix must also honor modal precedence and tiny-terminal gating. | `model.go` restored byte-for-byte; probe test removed; no production delta retained. |

### Second pass: systemic hypotheses and disposition

- **Reproduced:** the shell's key-only modal interface and wrapped-command route
  omit a whole valid input event class. Direct input setters, helper-selected
  enum values and key-only runtime typing can all remain green while paste is
  lost throughout creation. M1 is one systemic finding, not five field findings.
- **Falsified within the exercised paths:** shared enum preparation makes both
  creation and e wrong together. Literal option expectations in the shipped
  selector test plus separate durable override/default and e probes pin different
  expectations without comparing consumers to each other. Core's independent
  expected-value test also passed; parity with `NewTask` alone was not treated as
  sufficient.
- **Guard masking demonstrated and bounded:** the solo pending mutation survived
  because another real guard preserved the invariant; coordinated removal was
  killed. `mode` selects one field/picker/enum/discard interaction; `saving`,
  `locked` and `displaced` describe independent outcome/authority dimensions.
  A reachable issued → activated → displaced + locked + still-saving state is
  necessary, not evidence for collapsing all booleans into one exclusive enum.
- **Falsified within the exercised paths:** selected-parent display/Location
  substitution, a hidden write on absent/read-only/empty catalogs, stale catalog
  overwrite, or replay on uncertain receipts. Hostile independent ports plus
  baseline/core/store regressions described above refused or retained the exact
  intent. Whole-stem identity, no reservation on preview and same-ID replacement
  limitations remain the prerequisite contract, not stronger database promises.
- **Falsified for the tested landing/session sequences:** a watcher supersedes
  pending canonical created-task focus or a late origin receipt reloads/navigates
  the incoming workspace. The shipped watcher regression and independent delayed
  origin probe passed. Broader adjacent watcher work is not duplicated here.
- **Unresolved beyond these bounds:** physical-terminal escape parsing/clipboard
  integrations, every narrow recovery/enum/Unicode combination, the sister
  reviewer's repeated scheduler/replay probes, and exhaustive status/filter
  interleavings were not independently executed here. The paste reproduction
  used a real running Program with its actual `tea.PasteMsg`, not a physical
  clipboard. Existing key/dirty/selector and real-program baselines passed; they
  do not settle these unexecuted cases. Force quit remains an escape with no
  durable-draft promise.

### Extension-boundary assessment

The current separation is a reasonable first-consumer seam. Task row constants,
epic requirement, `EpicMutationTarget`, ready-to-start copy, `NewTaskParams`,
`TaskCreationReceipt` interpretation and Tasks landing are concretely task-specific.
Form focus generations, bounded scrolling/layout, explicit stage/submit/cancel,
dirty confirmation and delayed widget routing are candidate shared mechanics.
The missing event class in M1 should be pinned before those mechanics spread.
Root Model storage/message branches and page-local action discovery will need a
small host/registry seam when another real kind exists; their current task-specific
shape is not itself a product defect or a reason for a universal core port.

Actual use cases differ:

| Kind | Verified owning contract and consequence |
| --- | --- |
| Epic | `internal/core/service_epic.go:12`/`:25`: title, required description, validated explicit status/priority, tags, body/template, dry-run. Calls adapter `CreateEpic`; numbering is adapter-owned (`internal/store/create.go:474`). Typed `EpicCreationReceipt` is in `creation_receipt.go:42`. Do not borrow task defaults or mint its epic number in the form. |
| Audit | `service_audit.go:12`/`:24`: area/date, body/template, dry-run; empty date defaults in core, bucket always open, date+area slug, typed `AuditCreationReceipt` (`creation_receipt.go:49`). No title/epic/tags requirement. |
| Research | `service_research.go:17`/`:34`: title, created date, description/tags, body/template, dry-run; date affects ID minting; bounded collision loop preserves committed/finalization receipts (`:87`). `ResearchCreationReceipt` is distinct (`creation_receipt.go:56`); no lifecycle or task parent. |
| Thread | `service_thread.go:12`/`:26`: title, required description/goal, optional target date/tags/member references, body/template, dry-run; membership resolves inside the authoritative guarded planner (`:64`). `runThreadCreationMutation` at `:86` binds successful selection over conflicts. `thread_creation.go:35`/`:43` supplies Changed plus typed creation failure/receipt. A task receipt cannot represent this contract. |

A bounded audit slice would add area/date fields and a typed conversion to
`NewAuditParams`, keep blank date for core defaulting, submit through the captured
Service command, interpret `AuditCreationReceipt`, and land on the audit tab by
receipt Audit.ID. Extract only the form-host mechanics now used by task and audit,
with literal per-kind field/default and runtime-event tests; do not transplant the
epic catalog, task lifecycle copy or task numeric rows.

A Thread slice would collect description/goal and optional exact selected task
identities in a multi-select catalog. **Current `NewThreadParams.Tasks` remains
reference-oriented**, calling `ResolveTaskID` at `service_thread.go:69`; passing an
ID-looking string does not establish an exact selected-endpoint guarantee. Add a
bounded selected-member intent contract if offering that picker, revalidate it
under `ThreadCreationMutationStore`, and keep guarded graph validation, conflict
handling and `ThreadCreationReceipt`/`ThreadCreationMutationFailure` in core.
Reuse the host against task, audit and Thread under tests; keep their receipt and
destination policy separate. This is future work already assigned to
`6gmshzwphadw`, not a missing feature or current finding. No forms were implemented
and no broad Store expansion is recommended.

The design checkpoint is intent, and the extension task's unchecked criteria are
future work. The owner AC walkthrough task is not used as a fixture or evidence.
No universal CreateEntity API, other-kind TUI creation, changed machine contract or
persisted schema is claimed to ship in this patch.

### Final validation and isolation attestation

Final restored focused race run passed (18.860s). `audit lint 6gnqmm1qp9mz`,
`lint --links` and `git diff --check` passed. The CLI-created finding remains open;
its normal write stamped `updated_at` in this audit. All brief prose is unchanged.
Full-repository `just test`/`just lint` were inherited from the owner baseline, not
rerun; no command metadata changed, so docs generation was not run.

All findings remain open. Only this audit may be transferred; the report records
`transfer=pending` because delivery is established separately by the helper receipt.

```text
sandbox_path=/private/tmp/isolated-review.OKDTdg
git_dir=/private/tmp/isolated-review.OKDTdg/.git
baseline_commit=8b4435bcdda45078f4812681e8eafade9e026cea
source_blob=cc5181fd59806a0754358f3ff1495fa9e0948c46
source_fingerprint=00db015bcd781bdd2749c6dfc5d7da84ae1b826a
deliverable=planning/audits/6gnqmm1qp9mz-2026-10-20-tui-staged-task-creation-implementation-codex.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner disposition

Delivery was confirmed with the source helper's `receipt` command against
`/private/tmp/isolated-review.OKDTdg` before editing this report. The reviewer
disclosed a limited pre-isolation filename enumeration/scripts-guide read;
this is not recorded as fully protocol-compliant. Its sandboxed technical
evidence is independently actionable: the owner reproduced M1 with failing
regressions and implemented the bounded fix recorded in its resolution.

The other-kind assessment corroborates the existing additional-creation task
`6gmshzwphadw`, including its exact selected-member prerequisite for Threads;
no duplicate task or premature generic form framework is needed. Antigravity's
independent report remains pending and untouched. Closing this audit settles
its finding, not the implementing task or the other review.
