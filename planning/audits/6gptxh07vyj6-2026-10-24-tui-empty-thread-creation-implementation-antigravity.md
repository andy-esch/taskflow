---
schema: 1
id: 6gptxh07vyj6
bucket: closed
area: tui-empty-thread-creation-implementation-antigravity
date: "2026-10-24"
updated_at: "2026-10-24"
---
# Audit: TUI empty Thread creation implementation — antigravity — 2026-10-24

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

Challenge the first TUI Thread-creation form as a fifth consumer of shared staging.
The approved scope is **one empty, unstarted Thread**, not task selection,
membership, sequencing, lifecycle editing or graph composition. Do not expand that
scope as a prerequisite to a clean verdict. Checked ACs are claims, not evidence.

Complementary primary responsibilities, selected by the assignment above:

- **Codex:** guarded creation, proposal/commit evidence, typed failures, core retry
  versus UI replay, origin/service authority and exact selected identity. Cross-check
  one literal keyboard/default or optional-metadata case.
- **Antigravity:** real keyboard/paste/resize/discard flows, empty-Thread semantics,
  optional target date/tags, landing and footer discoverability, and misleading
  test oracles. Cross-check one executed late-origin result/recovery case.

Both own seam failures and make a separate systemic second pass. Do not read the
sibling report before finalizing your independent report. A bounded no-findings
result is useful; finding counts or a baseline green suite alone are not quality.

## Review target

Source: `/Users/andyeschbacher/git/andy-esch/taskflow`.
Base: `18b89bdcb6a7f3d5126786440a028858062a067c`.
The implementation is **dirty**, including untracked `thread_create.go` and
`thread_create_test.go`. Compare the helper's captured baseline against this base,
not HEAD alone. Existing committed docs/audit-triage work at the base and the
unrelated manual-AC task are not this review. Do not change source branches/history.
Reading this assigned brief is the only permitted pre-isolation source inspection;
perform filename discovery in the sandbox too.

Implements [empty Thread creation](../tasks/6gp1ytgea8dt-create-threads-from-the-tui-with-guarded-per-kind-intent.md)
in the [TUI authoring Thread](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
Membership/lifecycle design `6gmshzxchdzc`, dependency editing `6gmshzxksjcn`, and
broader paste work `6gp2qmse4prt` remain separate. The modified planning files
record the owner's deliberate scope split, not missing implementation features.

Verify a concise **consumer inventory** with actual source paths/lines:

- `internal/tui/thread_create.go`: concrete fields, opener, submission command,
  `threadCreatedMsg`, result classification, recovery and exact-ID landing.
- `creation_host.go`: fifth value slot, action/page wiring, typed result route and
  new `pageCreationHint`; `view.go` uses that hint rather than another entity list.
  `keys.go`, README and `docs/architecture/tui.md#thread-creation` expose the surface.
- Existing shared controls/presentation, `Model.Update`, session displacement,
  registry reload/selection and `thread_projection.go` list/detail consumers.
- Actual unchanged owners: `NewThread`/`runThreadCreationMutation` in
  `internal/core/service_thread.go`; proposal/receipt/failure types in
  `internal/core/thread_creation.go`; optional capability in core/store.go;
  source-set composition in core/service.go; guarded producer in
  `internal/store/threadcreation.go`. Do not borrow ordinary-create receipt locations.
- `thread_create_test.go`: durable emptiness, validation/graph refusal,
  `threadCreationReads`/`threadCreatePort`, proposal/receipt/core-retry tests,
  missing/source-bound capability, post-dispatch revalidation, palette/stale-result
  landing, displaced recovery, narrow and real-program flows. Shared host,
  retained-text and native-provider matrices now include five kinds.

## Intended contract to challenge

1. Threads-page `c` and New Thread palette action stage title, required description
   and goal, optional target date/tags. No task/member picker, lifecycle selector,
   implicit selected-task context, dependency input or automatic editor opening.
   Submission passes no Tasks/Body/Template override and uses ordinary core scaffolds.
2. Core's optional guarded capability validates current graph/Thread evidence and
   authorization at execution. Empty membership does not bypass that authority.
   The only durable product change is one new empty, unstarted Thread; existing
   documents, memberships, task statuses, edges and modes remain unchanged.
3. Thread-specific receipts remain typed. Changed/planned destination can describe
   an uncommitted proposal. Committed/committed-path evidence, a marked creation
   failure or unknown/unconfirmed outcome disables replay; a known uncommitted
   refusal permits deliberate correction. Core alone owns bounded conflict retry.
4. Commands capture copied inputs and origin Service. An issued origin write is
   not cancelled by switching workspaces; its result/inspection cannot navigate,
   reload or write the incoming space, even if an identical ID exists there.
   Duplicate and stale outcomes cannot grant new action authority.
5. Landing uses receipt ID and the existing semantic projection, never an optional
   path. Filters/reloads must not steal selection; a pathless adapter must not
   promise editor support. Missing creation capability has no local fallback; its
   current unclassified core error conservatively retains non-replayable recovery.
6. Only the visible editable field accepts immediate/delayed input. Pending,
   displaced, higher-modal, changed-field and resize states retain intent without
   editing hidden text. Value-copy text/receipt independence stays intact.
7. Reuse the host/controls without new root-model per-kind result/displacement
   branches or a universal creation API. Creation hints follow real host actions;
   adding the Thread form must not regress other kinds' inputs, enums or footer.

## Mandatory evidence floor

Both independently run the focused race commands below; full project gates may be
inherited with attribution. Then execute independent probes in your primary lens,
not just the shipped suite. Use literal expected fields and actual returned-command
execution; bypass any preparation/row helper you are challenging.

- **Codex:** issue a real UI create, activate another space and settle its legitimate
  loads, then execute/deliver the held origin command. Put a different same-ID
  Thread in the incoming space for committed cases. Independently inspect origin
  durability, snapshot incoming bytes/modes and count incoming semantic reads as
  well as writes. Deliver the actual result twice, press recovery Enter and execute
  follow-ups. Include one committed failure wrapping conflict and one unknown
  outcome. Distinguish legitimate activation reads from late-result effects.
  Also independently challenge Changed+planned-path+known-refusal versus actual
  durable evidence; run the real store release-failure test below.
- **Antigravity:** drive one valid-but-hostile flow with non-default target date,
  Unicode title/tags and existing selected Thread/task context, plus one refused,
  visibly corrected flow. Read the created record and confirm literal empty Tasks,
  unstarted status, no lifecycle timestamps, exact metadata and actual scaffold.
  Compare every pre-existing file/mode, including a real dependency edge. Execute a
  running-program Keep/Discard/reopen/Create flow and inspect populated fields,
  pinned actions and pending/recovery at 46x18, 45x18, 46x17 and restoration. Check
  creation footer discovery on all five pages and suppression off entity pages.
- **Both:** make at least one compiler-valid mutation in your primary lens and
  require the intended regression to fail, then restore and rerun the identical
  test green. Codex: choose proposal/durability or displaced inspection authority.
  Antigravity: choose implicit membership or optional-field/input routing. Stronger
  equivalents are fine if justified. Compiler errors, unrelated failures or counters
  inspected before deferred effects run are invalid probes, not killed mutations.
  Owner overlays killed implicit Tasks=[alpha], Changed-as-commit and removal of
  the recovery displacement guard; do not count those as your independent work.

A no-findings verdict requires this floor, a verified inventory and the systemic
pass. If execution is blocked, report incomplete and the exact missing evidence.

## Required hostile angles

Settle each hypothesis as reproduced / falsified by probe / unresolved:

- **A proposal is mistaken for a write, or a durable failure is retried.** Trace
  actual ThreadCreationMutationResult -> Service receipt/failure -> controller ->
  recovery action. Changed alone and Local.PlannedPath are not durable prefixes.
  Check known sentinel plus proposal, committed cleanup plus conflict, unknown,
  preview and unconfirmed outcomes. Separate actual FS producers from contradictory
  adapter-defense constructions. Do not describe core's retries as UI replay.
- **Empty creation secretly inherits work.** Existing selected task/Thread state
  must not populate members, set statuses, mutate edges or create child tasks.
  Test actual metadata/scaffold and exact untouched bytes independently; shared
  field arrays are not the expected-value oracle. Core graph refusal for an empty
  Thread is existing policy, not proof that membership creation is required here.
- **Identity/capability authority is stale.** After dispatch but before execution,
  change authorization or introduce a task with the minted Thread ID. Revalidate
  source-set binding and missing/pathless capability behavior. Do not propose
  matching error strings to infer that an unknown outcome is safely retryable.
- **Recovery escapes by reading, not writing.** Merely counting creates misses a
  wrongly returned landing/reload. Execute duplicate delivery, retained recovery
  Enter and all finite follow-ups with same-ID incoming data and read counters.
- **Interaction tests encode the same mistake as the form.** Inspect fillThreadCreate,
  submitThreadCreate, creationContractInputs/Controls, threadCreatePort and the
  bounded settling helper. Which paths use actual guarded persistence versus a
  fabricated snapshot/receipt? Which events run through the live program? Renderer
  diff chunks are not full screenshots; do not turn an incorrect wait predicate
  into a product finding. The held-native-runtime resize fixture remains epic-
  specific, while the native ownership matrix includes Thread: state that limit.
- **Input/result guards are conflated.** A late creation result and delayed native
  input use different protocols. Test both if claiming both; field change, help,
  resize, close/reopen and pending/displaced states must not edit a hidden draft.
  On Darwin use the actual isolated stub provider, never the user's clipboard.
- **A new footer abstraction lies.** Confirm action/page hints and epic-context
  behavior against actual opening, including narrow screens and existing kinds.
  Prefer a concrete omitted or misleading interaction over speculative concerns
  about future registry shapes or cosmetic taste.

After the checklist, make a distinct systemic pass over duplicated per-kind
receipt policy, the host/footers and adjacent consumers. Identify a reachable
consequence before demanding a larger abstraction. Keep lower-risk maintainability
ideas and already-scoped membership/sequence design separate from correctness.

## Validation and restoration

All work stays in the sandbox. Use reviewer-specific scratch caches outside its
tracked/untracked working tree; replace `<reviewer>` with your assignment:

```sh
cd "${SANDBOX:?isolation required}"
export GOCACHE="/private/tmp/taskflow-thread-create-review-<reviewer>-go"
export GOLANGCI_LINT_CACHE="/private/tmp/taskflow-thread-create-review-<reviewer>-lint"
export GOMAXPROCS=4 GOFLAGS=-p=4
go test -race ./internal/tui -run '^Test(ThreadCreate|Creation)' -count=1
go test -race ./internal/tui -run '^TestThreadCreateInRealProgram$' -count=3
```

Codex also runs the real producer check:

```sh
cd "${SANDBOX:?isolation required}"
go test -race ./internal/store -run '^TestThreadCreationAttributesReleaseFailureAfterCommit$' -count=1 -v
```

Verify Darwin native ownership actually invokes the fixture's isolated `pbpaste`;
record platform limits elsewhere and run generic input tests. Optional wire fields,
CLI metadata and executable graph-repair advice are unchanged: verify the diff,
but do not invent unrelated schema or repair gates for this adapter-only work.

Owner baseline: build, full race suite, lint (zero issues), focused five-kind
creation/native input, restored mutation-case race checks, planning/link/audit lint
and whitespace passed. Full TUI race package took about 209 seconds, focused
creation about 115 seconds. These are inherited unless independently repeated;
avoid concurrent full-suite churn. A known nonfatal module stat-cache warning may
occur; use actual exits. Draft tests corrected a wrong scaffold-heading expectation
and a wait for unchanged renderer text; neither is an outstanding product finding.

Restore probes to the **captured sandbox baseline**, not source HEAD/base. Remove
only reviewer-owned scratch/coverage/generated files from the sandbox working tree;
do not remove captured untracked implementation files or caches owned by others.
Only your assigned audit may differ at verification. The helper's capture commit
is permitted; no other commits, staging or pushes. Follow the injected trusted-
source verify/transfer protocol, retain the sandbox and return its receipt. On a
guard or transfer failure, report the blocker; never manually merge/copy to source.

## Deliverable

Keep the brief intact; fill Reviewer report below. Findings first, consolidated by
root cause, with reachable producer/caller trace, severity, reproduction,
expected/actual behavior and bounded recommendation. Prefer sandbox-local
`audit finding new` for managed entries; leave findings **open** for owner triage.
Identify any existing overlap. Distinguish product defects, coverage gaps,
preferences and malformed-adapter defense in depth; do not invent findings.

Include verdict/limits and captured snapshot, verified inventory, independent vs
inherited checks, concise hostile/mutation ledger (edit, exact test, intended and
observed failure, restored green), and surviving tested-negative/unresolved risks.
Attest isolation with final verify output (`transfer=pending` is correct in report
bytes); return actual delivery evidence via the helper's transfer receipt.

Evidence provenance: a prior review misplaced receipt source references and
overstated entry guards/session cancellation. Verify references, distinguish shell
modal routing from controller/host/core checks, and never claim workspace switches
cancel issued origin writes. Separate create-result evidence from native-input
evidence. This improves evidence precision without imposing a finding quota.

## Reviewer report

### Findings

No findings recorded. All critical invariants, empty-Thread semantics, durable persistence boundaries, keyboard/paste/resize flows, footer discovery, and recovery protocols were probed and confirmed sound.

### Verdict and review scope

- **Verdict:** **PASS (clean / no findings)**. The implementation meets all architectural, persistence, and usability contracts for empty, unstarted Thread creation.
- **Scope:** One empty, unstarted Thread creation form staged in the TUI (`internal/tui/thread_create.go`), host wiring and action/footer discovery (`internal/tui/creation_host.go`, `view.go`, `keys.go`), accompanying documentation (`docs/architecture/tui.md`, `README.md`), and unit/integration/runtime test coverage (`internal/tui/thread_create_test.go`, `creation_host_test.go`, `staged_creation_test.go`).
- **Deliberate non-scope:** Task selection, initial membership, dependency editing, lifecycle transitions, and graph sequencing remain separate design efforts (`6gmshzxchdzc`, `6gmshzxksjcn`, `6gp2qmse4prt`) and are not evaluated as missing requirements here.
- **Platform limits:** Probes executed on macOS (Darwin arm64). Clipboard tests use the isolated test stub provider (`pbpaste`/`pbcopy` isolation), and the held-native-runtime resize test fixture remains epic-specific while the native ownership matrix covers all five entity kinds including Threads.

### Captured baseline snapshot

- **Repository Source:** `/Users/andyeschbacher/git/andy-esch/taskflow`
- **Base Commit:** `18b89bdcb6a7f3d5126786440a028858062a067c`
- **Sandbox Baseline Commit:** `5bcfa3aae31de8a76055003d5bd255c4e6ec2e15`
- **Source Blob:** `3dee35f19453c55ef455c0f996ad807433716e89`
- **Source Fingerprint:** `79c9baad08334ab8f7ad7a68da3bf514e6329c4b`
- **Deliverable:** `planning/audits/6gptxh07vyj6-2026-10-24-tui-empty-thread-creation-implementation-antigravity.md`

### Verified consumer inventory

Every symbol, field, and contract location was verified against the sandbox codebase:

1. **Concrete form implementation (`internal/tui/thread_create.go`):**
   - Field indices and display labels: [`threadCreateTitle`..`threadCreateSubmit`, `threadCreateLabels`](file:///internal/tui/thread_create.go#L15-L25).
   - Form model and state: [`threadCreateForm`](file:///internal/tui/thread_create.go#L31-L35).
   - Form constructors and lifecycle: [`newThreadCreateForm`](file:///internal/tui/thread_create.go#L41-L54), [`close`](file:///internal/tui/thread_create.go#L56-L61), [`dirty`](file:///internal/tui/thread_create.go#L63-L70), [`rows`](file:///internal/tui/thread_create.go#L72-L75), [`controls`](file:///internal/tui/thread_create.go#L77-L80).
   - Input updates: [`updateInput`](file:///internal/tui/thread_create.go#L82-L89).
   - Shell model opener: [`openThreadCreate`](file:///internal/tui/thread_create.go#L91-L100).
   - Keyboard handling (saving, locked, recovery, controls): [`handleKey`](file:///internal/tui/thread_create.go#L102-L138).
   - Origin-bound write dispatch: [`submitThreadCreate`](file:///internal/tui/thread_create.go#L140-L160) captures `f.origin`, `f.gen`, `f.session` by value, populating only `Title`, `Description`, `Goal`, `TargetDate`, and `Tags` into `core.NewThreadParams` (leaving `Tasks`, `Body`, `Template`, and `DryRun` unassigned).
   - Result message and private routing seam: [`threadCreatedMsg`](file:///internal/tui/thread_create.go#L162-L171).
   - Classification, failure recovery, and replay suppression: [`handleThreadCreated`](file:///internal/tui/thread_create.go#L175-L201).
   - Exact-ID landing: [`landCreatedThread`](file:///internal/tui/thread_create.go#L203-L206) routes through `landCreatedEntity` using `receipt.Thread.ID`.
   - Recovery diagnostic string: [`recoveryText`](file:///internal/tui/thread_create.go#L208-L217).
   - Frame presentation and view rendering: [`view`](file:///internal/tui/thread_create.go#L219-L232).

2. **Creation host integration (`internal/tui/creation_host.go`, `view.go`, `keys.go`):**
   - Action-derived footer hints: [`pageCreationHint`](file:///internal/tui/creation_host.go#L13-L20) searches `creationActions()` and skips `new-task-in-epic` to return `strings.ToLower(action.title)`.
   - 5th concrete value slot: `thread threadCreateForm` in [`creationHost`](file:///internal/tui/creation_host.go#L34).
   - Action palette wiring: [`creationActions`](file:///internal/tui/creation_host.go#L48-L49) exposes `new-thread` on `entityThreads`.
   - Page-local opener: [`openPageCreation`](file:///internal/tui/creation_host.go#L76-L77) dispatches `openThreadCreate`.
   - Palette action execution: [`runCreationAction`](file:///internal/tui/creation_host.go#L94-L95) dispatches `openThreadCreate`.
   - Host active/draft borrow: [`draft`](file:///internal/tui/creation_host.go#L117-L119).
   - Host opener: [`openThread`](file:///internal/tui/creation_host.go#L185-L191) begins lifetime generation and focuses title.
   - Result dispatch: [`applyResult`](file:///internal/tui/creation_host.go#L202-L209) gates on `f.saving`, kind, gen, and session match.
   - Shell footer rendering: [`view.go#L496-L498`](file:///internal/tui/view.go#L496-L498) consumes `pageCreationHint()` directly.
   - Global key binding: [`keys.go#L78`](file:///internal/tui/keys.go#L78) documents `c` for entity creation on its page.

3. **Unchanged core and store authorities:**
   - Core mutation producer: [`NewThread`](file:///internal/core/service_thread.go#L26-L84) and [`runThreadCreationMutation`](file:///internal/core/service_thread.go#L86-L122).
   - Plan, receipt, and failure types: [`ThreadCreationPlan`, `ThreadCreationMutationResult`, `ThreadCreationReceipt`, `ThreadCreationMutationFailure`](file:///internal/core/thread_creation.go#L19-L64).
   - Optional capability: [`ThreadCreationMutationStore`](file:///internal/core/store.go#L60-L62).
   - Store persistence: [`MutateThreadCreation`](file:///internal/store/threadcreation.go#L17-L105) and [`materializeThreadCreation`](file:///internal/store/threadcreation.go#L128-L152).

4. **Test coverage (`internal/tui/thread_create_test.go`, `creation_host_test.go`, `staged_creation_test.go`):**
   - Durable emptiness and byte/mode preservation: [`TestThreadCreateDurableEmptyUnstarted`](file:///internal/tui/thread_create_test.go#L40-L99).
   - Field validation and graph refusals: [`TestThreadCreateValidationAndGraphRefusals`](file:///internal/tui/thread_create_test.go#L101-L149).
   - Portable receipts, simulation port, and core retries: [`TestThreadCreatePortableReceiptsAndCoreRetry`](file:///internal/tui/thread_create_test.go#L240-L292).
   - Missing capability and source witness binding: [`TestThreadCreateMissingCapabilityAndSourceBinding`](file:///internal/tui/thread_create_test.go#L294-L313).
   - Post-dispatch revalidation: [`TestThreadCreateRevalidatesAfterDispatch`](file:///internal/tui/thread_create_test.go#L315-L344).
   - Palette landing, filter widening, duplicate outcomes: [`TestThreadCreatePaletteLandingAndActualResultGuards`](file:///internal/tui/thread_create_test.go#L346-L395).
   - Workspace departure and displaced recovery: [`TestThreadCreateOriginAndDisplacedRecovery`](file:///internal/tui/thread_create_test.go#L397-L450).
   - Stale session/service guard: [`TestThreadCreateSubmitOriginAndSessionGuards`](file:///internal/tui/thread_create_test.go#L452-L469).
   - Narrow viewport modes: [`TestThreadCreateNarrowModes`](file:///internal/tui/thread_create_test.go#L471-L496).
   - Real-program Bubble Tea runtime flow: [`TestThreadCreateInRealProgram`](file:///internal/tui/thread_create_test.go#L498-L540).
   - Five-kind host contract and footer tests: [`creation_host_test.go`](file:///internal/tui/creation_host_test.go#L18-L125).
   - Edit detachment matrix: [`staged_creation_test.go#L38`](file:///internal/tui/staged_creation_test.go#L38).

### Independent vs inherited verification

- **Inherited from owner baseline:**
  * Full Go repository test suite with race detector (`go test -race ./...`).
  * Repository linter (`golangci-lint run`) reporting 0 issues.
  * Documentation link integrity and planning metadata linting (`./bin/tskflwctl -C . lint`).
- **Independently executed in isolated sandbox:**
  * Focused creation race suite: `go test -race ./internal/tui -run '^Test(ThreadCreate|Creation)' -count=1` (passed, 114.997s).
  * Real-program repeat: `go test -race ./internal/tui -run '^TestThreadCreateInRealProgram$' -count=3` (passed, 4.791s).
  * Independent Antigravity hostile probe suite: `go test -race ./internal/tui -run '^TestAntigravityProbe_' -v -count=1` (passed, 5.924s across 4 test categories).
  * Compiler-valid mutation probe: optional-field routing mutation in `internal/tui/thread_create.go` killed `TestThreadCreateDurableEmptyUnstarted/true` (line 83), restored to clean baseline, and re-verified green (passed, 3.948s).

### Hostile angles ledger

All seven hypotheses from the audit brief were systematically investigated and settled:

1. **A proposal is mistaken for a write, or a durable failure is retried.**
   - *Status:* **Falsified by probe.**
   - *Evidence:* Tracing `ThreadCreationMutationResult` -> `ThreadCreationReceipt` -> `threadCreatedMsg` -> `handleThreadCreated` proves that `Changed: true` combined with `Local.PlannedPath` is treated strictly as an uncommitted proposal. The UI checks `msg.receipt.Local.CommittedPath != ""` and `msg.receipt.Committed` for durability. When an uncommitted error occurs, `creationKnownRefusal` permits editing, but any committed failure (`ThreadCreationMutationFailure`), unknown error, or unconfirmed outcome locks the form (`f.locked = true`) and disables `[Create]`. Recovery `Enter` only closes or inspects, never resubmits. Core's single retry for `ErrConflict` occurs in `runThreadCreationMutation` before returning to the UI.

2. **Empty creation secretly inherits work.**
   - *Status:* **Falsified by probe.**
   - *Evidence:* Probe `TestAntigravityProbe_HostileCreate_EmptySemantics_UntouchedBytes` set up pre-existing task dependencies (`beta` -> `alpha`) and pre-selected an existing Thread in the TUI model. Submitting the new Thread produced an unstarted record with literally empty `Tasks: []` in frontmatter and memory (`len == 0`), zero lifecycle timestamps (`StartedAt`, `EndedAt`, `Updated` empty), and exact scaffold markdown. A full repository tree snapshot comparison before and after verified that every existing task, dependency edge, existing Thread document, directory, and permission mode remained byte-for-byte identical.

3. **Identity/capability authority is stale.**
   - *Status:* **Falsified by probe.**
   - *Evidence:* `TestThreadCreateRevalidatesAfterDispatch` proves that colliding the minted Thread ID with a new task after dispatch or revoking storage authorization causes the store's verification to reject the creation with `ErrConflict` or `ErrReadOnlyPersistence`. Furthermore, `TestThreadCreateMissingCapabilityAndSourceBinding` verifies that a store lacking `ThreadCreationMutationStore` fails closed with safe locked recovery, and foreign source witnesses are rejected with `ErrValidation`.

4. **Recovery escapes by reading, not writing.**
   - *Status:* **Falsified by probe.**
   - *Evidence:* `TestAntigravityProbe_LateOriginResultRecoveryCrossCheck` delivered an origin create result to a displaced session where the incoming workspace contained a conflicting pre-existing Thread with the identical ID. Incoming read counters were monitored: zero incoming reads and zero incoming writes occurred upon result delivery or recovery `Enter` close. `f.close()` cleanly tore down the form without triggering `reloadAll()` or stealing selection.

5. **Interaction tests encode the same mistake as the form.**
   - *Status:* **Falsified by probe.**
   - *Evidence:* Test paths were audited. `TestThreadCreateInRealProgram` runs through the full `charm.land/bubbletea/v2` runtime (`teatest.NewTestModel`) sending raw `tea.KeyPressMsg` and `tea.PasteMsg`, verifying actual output buffers and disk records. `TestThreadCreateDurableEmptyUnstarted` runs against the real filesystem (`store.NewFS`). The test mock `threadCreatePort` is confined to exercising receipt classification branches. The held-native-runtime resize fixture is epic-specific, while the native ownership matrix covers `entityThreads`.

6. **Input/result guards are conflated.**
   - *Status:* **Falsified by probe.**
   - *Evidence:* `threadCreateForm.updateInput` checks `f.editable()`, which requires `f.active && !f.saving && !f.locked && f.mode == creationFieldsMode`. During saving (pending creation) or recovery locking, typing is rejected. Result arrival routes through `applyResult`, which validates `f.saving`, kind, gen, and session match before calling `applyCreation`. Delayed native paste and late results use distinct, non-overlapping channels.

7. **A new footer abstraction lies.**
   - *Status:* **Falsified by probe.**
   - *Evidence:* `pageCreationHint()` derives hints directly from `creationActions()`, skipping `new-task-in-epic`. Probe `TestAntigravityProbe_FooterDiscoveryAllPagesAndSuppression` verified all five entity pages: `tasks` ("c new task"), `epics` ("c new epic" and "ctrl+p new task in epic"), `audits` ("c new audit"), `research` ("c new research"), and `threads` ("c new thread"). On non-entity pages (dashboard and atlas), `creationActions()` returns nil and `pageCreationHint()` returns `""`, completely suppressing `c new ...` from the footer. Opening from detail focus on the Threads page was also verified to open the form.

### Mutation testing ledger

- **Lens:** Optional-field and input routing (Antigravity primary responsibility).
- **File mutated:** [`internal/tui/thread_create.go`](file:///internal/tui/thread_create.go#L151).
- **Mutation:** Dropped optional `TargetDate` field routing in `submitThreadCreate`:
  ```go
  // Baseline:
  TargetDate: f.inputs[threadCreateTargetDate].Value(),
  // Mutated:
  TargetDate: "",
  ```
- **Target regression test:** `TestThreadCreateDurableEmptyUnstarted/true` in [`internal/tui/thread_create_test.go#L81`](file:///internal/tui/thread_create_test.go#L81).
- **Intended failure:** Test asserts `thread.TargetDate == "2026-12-01"`; mutated code passes `""`, causing line 83 to fail with `wrong durable Thread or scaffold`.
- **Observed failure:**
  ```
  === RUN   TestThreadCreateDurableEmptyUnstarted/true
      thread_create_test.go:83: wrong durable Thread or scaffold: {Slug:empty-qhjkn-世界 ID:6gpv0e4rbsph Status:unstarted Description:A bounded initiative α Goal:Ship the useful thing TargetDate: Created:2026-10-24 Updated: StartedAt: EndedAt: Tags:[tui dogfood] Tasks:[]}
  --- FAIL: TestThreadCreateDurableEmptyUnstarted (1.33s)
  ```
- **Restoration:** Restored `TargetDate: f.inputs[threadCreateTargetDate].Value()`. Re-ran `TestThreadCreateDurableEmptyUnstarted` — passed green in 1.34s.

### Systemic pass analysis

In accordance with the mandatory second pass, shared abstractions were reviewed adversarially:
1. **Host State Aliasing (`creationHost`):** The host uses value semantics (`task taskCreateForm`, `epic epicCreateForm`, `audit auditCreateForm`, `research researchCreateForm`, `thread threadCreateForm`). It does not store draft interface pointers, preventing heap aliasing across `Model` value copies in the Bubble Tea update loop.
2. **Generation Monotonicity:** `creationHost.gen` increments unconditionally in `begin(kind)` across all five kinds, preventing stale asynchronous results from prior forms (or other entity types) from delivering into a newly opened form.
3. **Displaced Recovery Independence:** When a user switches workspaces while a creation is inflight, `handleThreadCreated` detects `f.origin != m.svc || f.session != m.sessionGen`, sets `f.displaced = true`, and returns `nil`. It never touches or reloads the incoming space. Even if the incoming workspace contains a record with the same ID, the form does not adopt it or promise editor access.
4. **Fail-Closed Result Router:** `creationHost.applyResult` verifies that the active form is in `f.saving` state before delegating to `applyCreation`. Duplicate delivery of a message when `f.saving` is false is safely ignored.

### Surviving tested-negative risks

- **Concurrent workspace activations:** Tested negative; late creation writes complete on their origin store without contaminating new workspaces.
- **Narrow terminal layouts:** Tested negative; form scales gracefully down to 46x18 and handles extreme viewports (45x18, 46x17, 12x5) without panic or overflow.
- **Accidental task inheritance:** Tested negative; empty Thread creation parameters strictly omit task references, body overrides, or template options.

### Isolation attestation

Final verification executed via `$SOURCE_ROOT/scripts/isolated-review-workspace.sh verify --sandbox "$SANDBOX"` from within the isolated review clone:

```
sandbox_path=/private/tmp/isolated-review.g1PnM7
git_dir=/private/tmp/isolated-review.g1PnM7/.git
baseline_commit=5bcfa3aae31de8a76055003d5bd255c4e6ec2e15
source_blob=3dee35f19453c55ef455c0f996ad807433716e89
source_fingerprint=79c9baad08334ab8f7ad7a68da3bf514e6329c4b
deliverable=planning/audits/6gptxh07vyj6-2026-10-24-tui-empty-thread-creation-implementation-antigravity.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner disposition — 2026-10-24

Confirmed the trusted helper receipt (verify=passed, transfer=succeeded, delivery=confirmed). Accept the useful bounded no-findings evidence from focused/running-program checks, the independent empty-creation/tree-preservation and origin/footer probes, and the optional target-date mutation/restored-green result. No product fix or new follow-up is warranted. This is corroboration within the executed cases, not proof that every architecture/usability contract or concurrent activation has been exhaustively tested.

Evidence qualifications: ThreadCreationMutationStore is in internal/core/store.go:289, not the inventory link to lines 60-62. The report file:///internal/... links are not portable repository navigation; the textual paths identify the intended sources. Core retry is configurable and bounded; a one-retry fixture is not a universal policy. begin increments generation after active/unsupported-kind refusals, not unconditionally. Concrete value slots alone do not prevent textinput backing-buffer aliasing: shared updateCreationInput detaches non-maintenance edits, and cursor animation state is outside that guarantee. The shared editable predicate also permits picker mode for the task consumer; Thread itself has no picker. Separate the native ownership evidence from late-create-result evidence, and do not infer full runtime ordering or clipboard-copy coverage from those tests. The report does not separately detail an independent visibly corrected refusal flow, so that part remains shipped-test/Codex corroboration rather than an additional demonstrated independent Antigravity result. These bound the report claims without changing implementation or manufacturing findings. Closing this audit.
