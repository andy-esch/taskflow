---
schema: 1
id: 6gptxh01kx8x
bucket: closed
area: tui-empty-thread-creation-implementation-codex
date: "2026-10-24"
updated_at: "2026-10-24"
---
# Audit: TUI empty Thread creation implementation — codex — 2026-10-24

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

No actionable product defects reproduced. No findings were created or dispositioned; there is no overlap to triage. This is a bounded clean verdict for empty, unstarted Thread creation, not approval of membership, lifecycle or graph-authoring UI.

### Verdict, snapshot and limits

Both passes and the Codex evidence floor completed independently on Darwin. Reviewed the captured dirty implementation at sandbox baseline `bb6f66046cd53ce45c7a725b3d64f4b9bd5a6f8b` against `18b89bdcb6a7f3d5126786440a028858062a067c`. The capture includes the previously untracked Thread form/tests. Only the assigned brief was read before isolation; the sibling report was not read. Planning scope splits and the unrelated manual-AC task were excluded from implementation judgment. Checked planning criteria were not used as proof.

Independent evidence supports guarded creation, literal optional metadata, empty membership, proposal/durability classification, exact-ID landing and origin-bound recovery. No unresolved blocking hypothesis remains within that scope. Other operating systems were not executed. Full project race/lint gates are inherited from the owner statement in this brief, not independently rerun. The running native-result resize fixture remains epic-specific; the five-kind native ownership matrix and independent Thread native probes actually ran. Running-program output predicates inspect renderer chunks, not full terminal screenshots.

### Verified consumer and authority inventory

Paths and line numbers refer to the restored captured baseline, not the source's base commit or a planned file.

| Boundary | Verified implementation/evidence |
| --- | --- |
| Concrete Thread consumer | `internal/tui/thread_create.go:15` defines five literal text fields plus Cancel/Create; `:41` initializes empty inputs; `:91` checks pending workspace opening, active draft, Service and Threads page. It does not independently check every higher modal: shell routing supplies that boundary. |
| Command and typed outcome | `internal/tui/thread_create.go:140` captures copied `NewThreadParams` and originating Service/session, leaving Tasks/Body/Template/DryRun absent; `:162` defines `threadCreatedMsg`; `:175` classifies results; `:102` owns recovery keys; `:203` lands on receipt Thread ID and optional path capability, never the receipt path. |
| Fifth host slot/actions | `internal/tui/creation_host.go:27` stores five value forms; `:37`, `:66`, `:84` expose page and palette actions; `:185` opens Thread staging; `:202` requires active saving state and matching kind/attempt/origin-session tokens. `:13` derives `pageCreationHint` while skipping the epic-context task action. |
| Shell, session and input | `internal/tui/model.go:285` permits stale-session typed creation results to reach the guarded host; `:338` routes results and `:342` separately gates native/widget results. `internal/tui/session.go:117` switches Services and returns legitimate loads; `:188` displaces drafts. Already issued writes retain their origin and are not cancelled. `internal/tui/overlay.go:43` orders help/config before creation; `internal/tui/staged_creation_view.go:29` checks visible input ownership. |
| Shared value controls/presentation | `internal/tui/staged_creation.go:31` holds value state; `:41` locks displacement; `:60` advances input generations; `:92` detaches widget buffers; `:107` routes delayed input; `:178` owns shared navigation/discard. `internal/tui/staged_creation_view.go:82` renders pending/recovery/resize states and `:157` pins actions. Result tokens and native input tokens are distinct protocols. |
| Projection/selection | `internal/tui/staged_creation.go:261` lands through registry reload and exact `entityRef`; `internal/tui/entity.go:238` stamps reload/restore identity, `:265` selects canonical keys, `:280` preserves pending landing across reload, `:464` registers Thread list/detail consumers. `internal/tui/model.go:711` rejects stale/invalid list loads. `internal/tui/thread_projection.go:24` consumes `ListThreadViews`; `:62` consumes graph detail plus an independently optional path. |
| Discoverability/docs | `internal/tui/view.go:496` consumes the host hint; `:448` and `:476` route non-entity footers separately. `internal/tui/keys.go:78`, `README.md:573` and `docs/architecture/tui.md:174` expose the new surface. Independent literal footer/action probes passed on five list pages at 120x40 and 46x18, with dashboard/atlas suppression. |
| Unchanged core owners | `internal/core/service_thread.go:26` owns `NewThread`, required metadata, scaffold and unstarted defaults; `:86` owns bounded conflict retry and committed-failure wrapping. `internal/core/thread_creation.go:26`, `:35`, `:43` own Thread-specific mutation result/receipt/failure. Only `LocalCreateOutcome` is shared (`internal/core/creation_receipt.go:18`); ordinary-create receipt definitions are not Thread receipt ownership. `internal/core/store.go:289` defines the optional creation capability; `internal/core/service.go:378`, `:410` discover it and validate source-set composition. |
| Unchanged guarded producer | `internal/store/threadcreation.go:17` authorizes, reads authoritative graph/Threads, validates, materializes a proposal, rechecks snapshots and writes one new file. `:78` sets proposal fields; `:102` sets committed evidence; `:43` retains release errors. `internal/store/lock.go:115` owns the checked guard; on this Darwin run `internal/store/lock_unix.go:18` locks the planning root directory via flock, not a named lock file. |
| Shipped evidence | `internal/tui/thread_create_test.go:40` real-store durable defaults/metadata/bytes/modes/edge/scaffold, `:101` validation/graph refusals, `:153` read facade, `:178` receipt stub, `:240` receipt/retry matrix, `:294` missing/source binding, `:315` execution-time revalidation, `:346` palette/filter/stale/duplicate landing, `:397` displacement, `:452` submission authority, `:471` narrow modes, `:498` live program. `internal/tui/creation_host_test.go:18` includes five kinds in host matrices; `internal/tui/staged_creation_test.go:37` includes Thread retained text; `internal/tui/staged_creation_clipboard_darwin_test.go:34` runs native ownership over that five-kind matrix. |

The diff command `git diff --name-only 18b89bdcb6a7f3d5126786440a028858062a067c HEAD -- internal/core internal/store internal/wire internal/cli` returned no paths. Optional wire fields, CLI command metadata and graph-repair advice are unchanged; wire branch mutation/schema generation/repair execution gates are therefore inapplicable to this adapter-only change.

### Pass one: hostile hypothesis ledger

| Hypothesis | Disposition and executed evidence |
| --- | --- |
| Proposal mistaken for durability / durable failure replayed | **Falsified by probe.** Independent `TestCodexAuditProposalAndDurability` obtains Changed+PlannedPath from an actual FS preview, adds a known validation refusal at the adapter seam, and verifies no durable bytes and editable deliberate correction. This combination is an adapter-defense construction, not a claim that FS emits that exact refusal. Actual real-FS commit plus conflict-wrapping cleanup becomes `ThreadCreationMutationFailure`, remains one call and cannot replay. Unknown outcome locks. Pure conflict uses two core attempts with retry budget one, then permits one deliberate UI correction. Shipped `TestThreadCreatePortableReceiptsAndCoreRetry` also executed pathless, marked failure, committed-path-only, preview and unconfirmed branches; these fabricated adapter cases are not durable-store fixtures. |
| Empty creation inherits selected work | **Falsified by independent probe.** `TestCodexAuditSelectedContext` starts with a selected task and a selected existing Thread containing alpha, plus a real beta→alpha dependency. Literal c/paste/Enter/Tab creates a different ID with empty Tasks, unstarted status, absent start/end/update timestamps, description `Literal summary β`, goal `Literal goal`, target `2027-03-04` and tags `[雪, α, tui]`. It compares the entire literal default scaffold (including Context placeholder) and every preexisting file/directory byte/mode. No implicit members, task status/edge changes or child tasks survive that oracle. Shipped durable tests independently executed both optional/default branches. |
| Identity/capability authority stale after dispatch | **Falsified in executed shipped regressions and verified producer trace.** `TestThreadCreateRevalidatesAfterDispatch` holds actual commands, introduces a task with the minted ID or revokes authorization, executes, and observes unchanged fixture bytes/modes plus editable refusal. Missing creation capability does not discover a local fallback and conservatively locks on the unclassified error. Foreign source-set composition refuses; portable pathless success lands exact ID without advertising editor support. No error-string inference was introduced. |
| Recovery escapes through reads | **Falsified independently, with mutation sensitivity proved below.** `TestCodexAuditLateOrigin` issues a literal UI create, activates a separate workspace, executes all legitimate activation commands, then executes/delivers the held origin command. Success and committed cleanup have a different durable same-ID incoming Thread. Unknown outcome verifies no origin change. Every case executes duplicate delivery, recovery Enter and all finite follow-ups. Each activation performs 3 counted semantic reads; late delivery/duplicate/recovery add 0 reads, 0 incoming creates, and no incoming byte/mode or selected/detail/dashboard changes. Origin command count stays 1; committed origin content is independently read and matches literal metadata/scaffold. |
| Interaction helpers encode the form's mistake | **Falsified within the bounded evidence.** Shipped `fillThreadCreate` (`thread_create_test.go:25`) uses literal keyboard progression, but `submitThreadCreate` (`:35`) jumps via production controls; `creationContractInputs/Controls` (`creation_host_test.go:57`, `:73`) are mechanics inventories, not independent field oracles. `threadCreatePort` fabricates several receipts/snapshots; only success/cleanup delegate real persistence. Independent scratch tests bypass all four helpers and `settleCreate`, using a separate bounded FIFO drain that executes every finite command and skips only typed cursor blink ticks. The mutation detects a read-only escape even though write counters remain unchanged. |
| Input/result guards conflated | **Falsified by separate native and result probes.** Darwin `TestCodexAuditNativeProvider` invokes actual Ctrl+V and an isolated `pbpaste` script that writes an invocation marker and returns `NATIVE-雪`. Visible input receives exactly that value; field-change, help, resize, close/reopen, pending and displaced cases preserve all five input values. Every case verifies the invocation marker. The shipped native matrix and epic live held-result resize fixture also executed in the focused race command. Generic delayed-paste/value-copy regressions ran there independently of these native cases. No user's clipboard was read. |
| Footer abstraction advertises another action | **Falsified for current pages.** Independent `TestCodexAuditHints` uses literal expected hints for all five kinds and actual c opening, including 46x18; dashboard and atlas contain no creation hint. The host skips epic-context task creation when naming page creation, and the existing selected-epic action keeps its separate Ctrl+P route. Narrow form boundaries and real keyboard Keep/Discard/reopen/Create ran in shipped regressions. This is not a claim that detail/immersive footers display every available command. |

### Pass two: systemic adversarial review

Re-reviewed the per-kind result policies, shared host/controls/footer seams, session routing and adjacent projection consumers after the checklist. Concrete attack: a displaced result can preserve draft locks and avoid writes while still returning a reload against the incoming Service. The independent counter oracle measures after deferred follow-ups and a compiler-valid mutation demonstrates that it catches this entire read-authority failure class. A duplicate cannot restore authority because host application requires saving state; retained recovery requires non-displacement before inspection. Matching IDs in different workspaces do not supply Service authority.

Compared Thread's classifier/recovery (`thread_create.go:109`, `:175`) with epic (`epic_create.go:185`), audit (`audit_create.go:157`), research (`research_create.go:171`) and task recovery (`task_create.go:197`). Thread correctly uses its own marked failure type; its proposal Changed/PlannedPath are deliberately excluded from durability. Duplication alone is not a defect or a reason to invent a universal create port. No current FS producer was found that emits committed destination evidence while denying its committed bit; contradictory custom-adapter receipts are bounded defensive cases, not demonstrated production durability failures. Core conflict retry was verified separately from UI replay.

Challenged common test oracles with literal metadata, the entire scaffold, preexisting tree snapshots, actual native provider execution and independent command draining. Verified snapshot helper implementation at `internal/testutil/snapshot.go:20` captures entry modes, regular bytes, empty directories and link targets. This is a test oracle, not a transaction guarantee. Host and footer changes do not add a per-kind root-model result/displacement branch; current per-kind policy remains in concrete consumers. No reachable systemic product defect reproduced. Future membership/sequence design and cosmetic/comment cleanup remain outside this verdict.

### Executed checks and mutation ledger

All Go commands used `GOCACHE=/private/tmp/taskflow-thread-create-review-codex-go`, `GOMAXPROCS=4`, `GOFLAGS=-p=4`; the prescribed reviewer lint cache was also supplied to the focused suite.

- Independent required commands: `go test -race ./internal/tui -run '^Test(ThreadCreate|Creation)' -count=1` — exit 0, 114.346s; `go test -race ./internal/tui -run '^TestThreadCreateInRealProgram$' -count=3` — exit 0, 4.920s; `go test -race ./internal/store -run '^TestThreadCreationAttributesReleaseFailureAfterCommit$' -count=1 -v` — exit 0, 1.184s, both ordinary and conflict-wrapping release failures pass with committed evidence and no core retries.
- Independent scratch probes: `go test -race ./internal/tui -run '^TestCodexAudit' -count=1 -v` — exit 0, 50.705s (late origin, proposal/durability, native provider); `go test -race ./internal/tui -run '^TestCodexAudit(SelectedContext|Hints)$' -count=1 -v` — exit 0, 10.355s. The selected-context probe was then strengthened to whole literal scaffold equality and rerun separately; exit 0, 7.421s (`scaffold.log`).
- Independent compiler-valid mutation: in `internal/tui/thread_create.go:183`, replace the displaced-result `return nil` with `return m.reloadAll()`. Exact command: `go test -race ./internal/tui -run '^TestCodexAuditLateOrigin$' -count=1 -v`. Intended failure: incoming semantic read count changes after all result/recovery follow-ups even though no incoming create occurs. Observed exit 1: all three cases fail `late/recovery escaped authority: reads 3 -> 8, creates=1/0`. This is a behavioral failure, not a compiler failure. Restore that file from captured baseline HEAD; identical command exits 0, 18.055s, all three cases show 3 activation reads and 0 additional reads. This mutation is independent of the owner's Changed-as-commit, implicit Tasks and recovery-key guard mutations.
- Reviewer evidence is retained outside the worktree at `/private/tmp/taskflow-codex-review-evidence.vRNWGO/`: `zz_codex_audit_test.go`, `independent.log`, `program.log`, `producer.log`, `systemic.log`, `scaffold.log`, `mutation-red.log`, `mutation-restored.log`. The scratch test is reviewer evidence, not shipped regression coverage, and was removed from the sandbox worktree before verification. No implementation file remains modified.
- Full owner build/race/lint and planning/link/audit checks listed in the brief remain attributed to the owner. Reviewer build/planning/report validation and isolation results follow below.

### Final validation and isolation attestation

Reviewer `just build` exited 0 with the fresh CLI (the known nonfatal module stat-cache permission warning appeared). Fresh `./bin/tskflwctl -C . lint`, `lint --links`, and `audit lint 6gptxh01kx8x` passed. Assigned audit relative links passed (2 targets); `git diff --check` passed. Full project test/lint and docs generation were not repeated: the delivered change is audit prose, CLI metadata is unchanged, and the required focused race/probe gates passed. The fresh ignored binary was removed after validation. All probes were restored to the helper's captured baseline; only this audit differs.

Trusted-source helper verification output (copied verbatim; report bytes intentionally retain `transfer=pending`):

```text
sandbox_path=/private/tmp/isolated-review.vRNWGO
git_dir=/private/tmp/isolated-review.vRNWGO/.git
baseline_commit=bb6f66046cd53ce45c7a725b3d64f4b9bd5a6f8b
source_blob=fc4c7da4dd72f95e9022a04e0b3518eb62659d5e
source_fingerprint=79c9baad08334ab8f7ad7a68da3bf514e6329c4b
deliverable=planning/audits/6gptxh01kx8x-2026-10-24-tui-empty-thread-creation-implementation-codex.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

The independent sandbox and reviewer evidence remain in place for owner receipt confirmation. The helper's transfer receipt, returned in the handoff, is delivery evidence; this pre-transfer attestation is not.

## Owner disposition — 2026-10-24

Accepted the bounded no-findings verdict after confirming the trusted helper receipt (verify=passed, transfer=succeeded, delivery=confirmed). Independent literal selected-context/scaffold and whole-tree preservation, real guarded cleanup, proposal/refusal, native-input and same-ID incoming-space probes support the empty/unstarted creation contract. The compiler-valid displaced-result reload mutation failed after executing follow-ups and counting incoming reads, then passed after restoration. Full project gates remain inherited evidence; other platforms and every asynchronous interleaving are not certified. No implementation fix or new follow-up is warranted. Closing this audit.
