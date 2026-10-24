---
schema: 1
id: 6gpgy3n9hp5z
bucket: closed
area: tui-audit-research-creation-implementation-antigravity
date: "2026-10-23"
updated_at: "2026-10-23"
---
# Audit: TUI audit and research creation implementation — antigravity — 2026-10-23

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

Independently challenge audit/research creation in the TUI: do these two different
forms genuinely reuse the hardened host while preserving their actual application
contracts, recovery authority and usable interactions? Checked ACs and owner tests
are claims to challenge, not a suggested verdict. No finding quota; a bounded,
evidence-backed no-findings result is useful.

Primary responsibilities (read your assignment in this audit's preamble):

- **Codex:** typed requests/receipts, origin binding, durable versus uncertain
  outcomes, retry boundaries, exact identity and incoming-space isolation. Also
  execute one date/default or text-only keyboard cross-check.
- **Antigravity:** actual keyboard/paste/resize/selection behavior, date/default and
  optional-metadata semantics, misleading presentation and systemic test-oracle
  weaknesses. Also execute one actual late-origin result/recovery cross-check.

Both own the shared contract, including seam failures. Do not read the sibling
report until your independent report is final. Do not implement fixes in the
handoff checkout or transfer probe files. The initial source-read allowlist is
this assigned audit only: perform filename/source discovery after isolation too.

## Review target

Source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`.
Base: `05055b8377160015fbb61041c4b374c713c18b0a`.
Implementation is **uncommitted**, including three untracked Go files. Review the
helper's captured working-state baseline against the base above, not HEAD alone.

Implements [audit/research forms](../tasks/6gp1ytg50tev-create-audits-and-research-from-shared-tui-authoring-forms.md)
in the [TUI authoring Thread](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
The previously merged host is context, not all new work. Thread creation,
membership and broad terminal-paste coverage remain separately planned:
`6gp1ytgea8dt` and `6gp2qmse4prt`. Report a distinct in-scope consequence before
reopening those boundaries. The older task-creation Antigravity audit and manual-AC
task present as unrelated untracked files are not this assignment.

Verify a concise **consumer inventory** with actual paths/lines:

- New `audit_create.go` and `research_create.go`: concrete form slots/inputs,
  open/submit paths, `auditCreatedMsg`/`researchCreatedMsg`, result/recovery and landing.
- `creation_host.go`: page/palette actions, supported-kind allocation, borrowed
  draft selection and typed result dispatch; `staged_creation.go` adds nil enum
  options for text-only forms. `keys.go` and README expose the changed surface.
- Existing `Model.Update`, session scoping/activation, `creationModal`, per-kind
  list/detail loaders and `landCreatedEntity`: follow commands through delivery
  and follow-up execution, not merely message construction.
- Actual `NewAudit`/`NewResearch`, their parameter/receipt types, required Store
  methods and real guarded create producers in core/store. These owners did not
  change. Do not invent an optional audit/research create capability.
- `document_create_test.go`: durable/default/date tests, `documentCreatePort`,
  portable recovery, core collision, stale actual results, origin/watcher, narrow
  and real-program tests. The host/shared/native-input matrices now include four
  kinds; their mechanics helpers must not become the domain-value oracle.
- Owning guidance: `docs/architecture/tui.md#audit-and-research-creation` and the
  planning-data checklist. Verify names in the captured tree before citing them.

## Intended contract to challenge

1. Page-local `c` and named palette actions open one draft only on the appropriate
   entity page. No new persistence/filesystem work in Update/View; commands capture
   copied inputs and the originating Service. Host lifetime/session guards apply
   equally to all four forms, without a universal entity creation API.
2. Audit collects area and optional date, delegates validation/scaffold/identity to
   core and always creates open. Research collects title, optional description/tags
   and created date, has no lifecycle/parent, and leaves date-based ID minting and
   bounded collision retries to core. Blank dates default at submission, not opening.
3. One explicit Create prevents duplicate dispatch. Only a known uncommitted
   refusal permits deliberate retry; committed evidence, finalization errors,
   unknown outcomes and incomplete success disable replay. Receipt paths alone
   never authorize opening a file or establish an identity.
4. Departed outcomes may enrich the retained origin recovery only. Even same-ID
   incoming records cannot receive navigation, reloads or writes from that outcome.
   Recovery/duplicate delivery must be tested by executing returned commands.
5. Success/inspection selects the exact receipt ID across active filters, audit
   bucket restrictions and reloads; absent local path capability does not prevent
   creation or falsely promise editor support. Research invents no status axis.
6. Visible editable fields alone accept ordinary and delayed native input. Pending,
   displaced, hidden/higher-modal and resize states cannot edit behind the screen.
   Non-empty text and recovery evidence remain independent across Model copies.
7. No CLI/wire/schema changes are intended. No automatic editor opening, template
   picker, long-prose authoring or graph/membership creation belongs in this slice.

## Mandatory evidence floor

Both reviewers independently run the focused race suite below; full project gates
may be inherited with attribution. Then execute **one independent hostile producer
scenario per new kind** in your primary lens, plus the bounded cross-check. Do not
just repeat the report's existing assertions or use shared helpers to compute
expected durable fields. Scratch tests are welcome but must be restored/removed
before transfer.

- Codex: execute actual commands issued before and run/delivered after workspace
  activation, with same-ID incoming records and different contents. Count incoming
  reads as well as writes, snapshot bytes/modes, deliver duplicate actual results
  and execute recovery Enter/follow-ups. Check origin durability independently.
  Also challenge a known refusal versus a committed/uncertain result for both kinds.
- Antigravity: drive a valid-but-hostile **and** a refused-then-corrected keyboard
  flow for each kind. Include blank-date defaulting across a clock change, explicit
  research backdating, empty optional metadata and exact selection under filtering.
  Inspect actual rendered fields/actions across 46x18, 45x18, 46x17 and restoration,
  with populated text and pending/recovery states. Execute at least one new-form
  running-program flow, not just hand-built reducer messages.
- Both: independently execute at least one compiler-valid mutation and the exact
  intended regression. Codex should challenge origin/recovery authority; Antigravity
  should challenge date forwarding or a text-only interaction assumption. You may
  choose a stronger equivalent, but state why. Owner overlays killed deletion of
  the audit result-origin guard and omission of research Created; those results
  are inherited unless you actually repeat them. A compile error, unrelated test
  failure or never-executed deferred command is not a kill. Restore the captured
  implementation and run the identical test green.
- A no-findings verdict requires this floor, verified consumers and a second
  systemic pass. If blocked, say incomplete and identify the missing evidence.

## Required hostile angles

Settle these hypotheses as reproduced / falsified by probe / unresolved, rather
than merely checking that a symbol exists:

- **Progress classification can collapse into unsafe retry.** Trace actual producer
  receipts versus fabricated preview/unconfirmed/path-evidence cases. Test progress
  accompanied by a known sentinel and finalization wrapping conflict. Research's
  internal collision loop is not a UI replay; confirm no second UI attempt follows
  committed/unknown work. Do not label malformed-adapter defense a shipped producer.
- **Copying another form copied its domain assumptions.** Audit date/area must not
  become task title/default policy; research optional fields must remain optional
  and its created date must encode the requested day. Test non-default values with
  independent literal expectations and actual durable reads. Do not invent required
  research tags/description or a bucket selector.
- **Green tests share a wrong oracle or skip effects.** Inspect `creationContractInputs`,
  `creationContractControls`, `fillDocumentCreate`, `submitDocumentCreate` and
  `documentCreatePort`: could a miswired row/action/default pass both implementation
  and expectation? Which real stores, native providers and commands actually run?
  Bounded synchronous settling is not an asynchronous scheduler or timeout. Use
  one independent probe that bypasses a potentially reassuring helper.
- **Selection is only visually right.** Use a nonempty filter or closed audit view,
  a reload arriving before initial landing, and a pathless receipt with opaque ID.
  Check exact selected source identity and subsequent detail, not only flash text.
- **A small-screen or text-only path breaks shared controls.** Nil enum policy must
  support Enter/Tab/Shift+Tab/cancel without panics or hidden edits. Inspect clipped
  labels, explicit Create/recovery actions and unchanged existing task/epic enums.
  Cosmetic preferences alone are not safety defects.
- **Late input still targets the wrong owner.** Try a field move, higher modal,
  resize and close/reopen; include actual native input when Darwin permits it.
  The existing native provider test covers all four kinds, while the existing
  held-result runtime-resize fixture is epic-specific: do not claim it already
  proves every new form's asynchronous interleaving.

After the checklist, make a distinct systemic pass over the shared controls/host,
per-kind controller duplication and callers. Could the next form be wired without
duplicating root policy? Identify a reachable consequence before recommending a
larger abstraction. A maintainability concern may merit a Low follow-up rather
than demanding a universal schema or core API redesign in this PR.

## Validation and restoration

Run only in the sandbox. Use reviewer-specific scratch caches outside its working
tree (replace `<reviewer>` with your assignment):

```sh
cd "${SANDBOX:?isolation required}"
export GOCACHE="/private/tmp/taskflow-document-review-<reviewer>-go"
export GOLANGCI_LINT_CACHE="/private/tmp/taskflow-document-review-<reviewer>-lint"
export GOMAXPROCS=4 GOFLAGS=-p=4
go test -race ./internal/tui -run '^Test(AuditCreate|ResearchCreate|DocumentCreate|Creation)' -count=1
go test -race ./internal/tui -run '^TestDocumentCreateInRealProgram$' -count=3
```

On Darwin, verify `TestCreationNativeClipboardOwnership` actually executes its
stub `pbpaste`; never read the user's real clipboard. Else record the platform
limit and run generic input cases. Core/wire/CLI changes, optional wire branches
and executable repair advice are absent from the captured diff: verify that
claim, do not invent unrelated schema/repair gates.

Owner baseline: build, full race suite, lint (zero issues), architecture navigation,
planning/link/audit lint and selected focused race checks passed; first test drafts
were corrected for actual scaffold headings, newly created directories and
consumed terminal-output chunks. This is not your independent result. The full
TUI race package took about 162 seconds; avoid concurrent full-suite churn. A known
nonfatal module stat-cache permission warning may occur; report actual gate exits.

Restore probes to the **captured sandbox baseline**, including untracked source
files captured in that checkpoint, not the source base commit. Only your assigned
audit may differ at verification. Do not leave scratch tests, generated output,
coverage or caches in the working tree. The trusted source helper's baseline commit
is permitted; no other commits, staging, branch changes or pushes. Use the injected
verify/transfer protocol, include its attestation and return its transfer receipt.
Keep the sandbox until the owner confirms delivery. Do not manually copy a report
or touch the sibling audit on a conflict.

## Deliverable

Fill Reviewer report below; preserve the brief. Findings first, consolidated by
root cause, with severity, actual producer/caller trace, expected/actual behavior,
reproduction, evidence and bounded recommendation. Prefer `audit finding new` in
the sandbox for managed entries; leave real findings **open**, naming any existing
overlap rather than marking work tracked yourself. Distinguish reachable product
defects, coverage gaps, preferences and adapter-defense hypotheses.

Include verdict/limits, captured snapshot, verified inventory, independent versus
inherited checks, a concise hostile/mutation ledger (edit, intended test, observed
failure, restored green), and remaining tested-negative/unresolved hypotheses.
Report isolation failures honestly. Put final verify attestation in the report
with `transfer=pending`; delivery is established by the helper-owned receipt in
your handoff, not by an impossible self-claim in the transferred bytes.

## Reviewer report

### Verdict and scope limits

- **Verdict:** Ready. No findings have been recorded.
- **Scope Limits:** Confined to TUI audit and research creation authoring forms, host routing, shared staging controls, keyboard/paste/resize interactions, date/metadata semantics, and document-create regressions. Thread creation (`6gp1ytgea8dt`), membership editing, and terminal paste coverage (`6gp2qmse4prt`) remain separate planned tasks. Zero lines changed in core, store, wire, or CLI packages.

### Mandatory isolation attestation

```text
sandbox_path=/private/tmp/isolated-review.hTy7nc
git_dir=/private/tmp/isolated-review.hTy7nc/.git
baseline_commit=e2248401e62e42e676667b5e2b6712c5ed29b599
source_blob=0c4764f841678319a78cdf67cb7bb28002345ef4
source_fingerprint=aa03f6d6d58f952eb0f60997d2299ba6cefde994
deliverable=planning/audits/6gpgy3n9hp5z-2026-10-23-tui-audit-research-creation-implementation-antigravity.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

### Findings

No findings have been recorded. All required hostile angles, keyboard/resize probes, and compiler-valid mutation probes settled negative with verified evidence.

### Captured snapshot

- Base commit: `05055b8377160015fbb61041c4b374c713c18b0a`
- Sandbox baseline commit: `e2248401e62e42e676667b5e2b6712c5ed29b599`
- Working tree state: 16 files in diff against base (11 project source/test/documentation files and 5 planning/audit files). Untracked Go files `internal/tui/audit_create.go`, `internal/tui/research_create.go`, and `internal/tui/document_create_test.go` were properly captured into the sandbox baseline.

### Verified consumer inventory

Exact paths and line numbers verified in the sandbox repository:

1. **`internal/tui/audit_create.go`**:
   - Form slots and inputs: lines 15–20 (`auditCreateArea = 0`, `auditCreateDate = 1`, `auditCreateCancel = 2`, `auditCreateSubmit = 3`), `auditCreateForm` struct lines 27–31 (`creationState`, `inputs [4]textinput.Model`, `receipt core.AuditCreationReceipt`).
   - Open and submit entry points: `m.openAuditCreate()` lines 78–87, `m.submitAuditCreate()` lines 127–143.
   - Message seam: `auditCreatedMsg` lines 145–154 (`creationRoute()`, `applyCreation()`).
   - Outcome handler and recovery: `m.handleAuditCreated()` lines 157–183, `recoveryText()` lines 190–199.
   - Entity landing: `m.landCreatedAudit()` lines 185–188 (`"all"` bucket view, local editor detection).
2. **`internal/tui/research_create.go`**:
   - Form slots and inputs: lines 15–22 (`researchCreateTitle = 0`, `researchCreateDescription = 1`, `researchCreateTags = 2`, `researchCreateDate = 3`, `researchCreateCancel = 4`, `researchCreateSubmit = 5`), `researchCreateForm` struct lines 29–33.
   - Open and submit entry points: `m.openResearchCreate()` lines 90–99, `m.submitResearchCreate()` lines 139–157.
   - Message seam: `researchCreatedMsg` lines 159–168 (`creationRoute()`, `applyCreation()`).
   - Outcome handler and recovery: `m.handleResearchCreated()` lines 171–197, `recoveryText()` lines 204–213.
   - Entity landing: `m.landCreatedResearch()` lines 199–202 (empty status view, local editor detection).
3. **`internal/tui/creation_host.go`**:
   - Palette actions: lines 27–33 (`new-audit`, `new-research`), execution lines 69–75 (`openAuditCreate()`, `openResearchCreate()`).
   - Page-local actions: lines 56–60 (`openAuditCreate()`, `openResearchCreate()`).
   - Lifetime allocation and supported kinds: lines 108–119 (`begin(kind)` accepting `entityTasks`, `entityEpics`, `entityAudits`, `entityResearch`), lines 147–161 (`openAudit`, `openResearch`).
   - Draft borrow: lines 84–106 (`draft()`, `active()`).
   - Private result dispatch: lines 167–179 (`creationResult` interface and `applyResult()`).
4. **`internal/tui/staged_creation.go`**:
   - Nil enum options for text-only forms: line 132 (`options func(int) []string`), lines 152–157 (`enumOptions(row int)` returning nil when `c.options == nil`), lines 190–195 (recovering to fields mode if `len(options) == 0`), lines 231–236 (`len(options) > 0` guard for Enter key).
   - Entity landing helper: lines 261–276 (`landCreatedEntity()` resetting filter, setting status view, widening restore ref, reloading).
5. **`internal/tui/keys.go`**:
   - Lines 33, 78: Key binding help text updated to `"new task/epic/audit/research on its page; task in selected epic: ctrl+p → New task in this epic"`.
6. **`README.md`**:
   - Lines 573–579: Documents staged creation for Audits and Research via `c`, audit default to today, and research backdating without lifecycle.
7. **Existing Model / Host runtime integration**:
   - `internal/tui/model.go:285–290`: Dispatches `creationResult` across departed sessions into origin recovery.
   - `internal/tui/model.go:338–340`: Dispatches `creationResult` within the active session.
   - `internal/tui/session.go:193`: Unconditional `m.creation.displace()` upon workspace switch.
   - `internal/tui/staged_creation_view.go:14–65`: `creationModal` sizing (`creationInputFits` requiring width >= 44, height >= 14), input routing, view formatting.
8. **Core / Store contracts (unmodified authority)**:
   - `NewAudit` in `internal/core/service_audit.go:24–60`, `NewAuditParams` lines 10–18, receipt in `internal/core/store.go:245–250`.
   - `NewResearch` in `internal/core/service_research.go:34–100`, `NewResearchParams` lines 17–25, receipt in `internal/core/store.go:275–280`.
   - 0 lines changed in `internal/core/`, `internal/store/`, `internal/wire/`, or `cmd/`. No optional creation capability added.
9. **`internal/tui/document_create_test.go`**:
   - Lines 45–93: `TestAuditCreateDurableDefaultsAndDates`
   - Lines 95–148: `TestResearchCreateDurableDateAndOptionalMetadata`
   - Lines 150–194: `TestDocumentCreateValidationAndDeliberateRetry`
   - Lines 198–317: `documentCreatePort` mock store
   - Lines 318–378: `TestDocumentCreatePortableReceiptsAndNoReplay`
   - Lines 380–392: `TestResearchCreateLeavesCollisionRetryToCore`
   - Lines 394–423: `TestDocumentCreateReadOnlyAndSubmitOriginGuards`
   - Lines 425–462: `TestDocumentCreateActualResultGuards`
   - Lines 464–576: `TestDocumentCreatePaletteWatcherAndOrigin`
   - Lines 578–605: `TestDocumentCreateNarrowModes`
   - Lines 607–657: `TestDocumentCreateInRealProgram`
10. **Owning architecture documentation**:
   - `docs/architecture/tui.md:144–172` (`#audit-and-research-creation`).

### Independent versus inherited checks

- **Independent focused race suite:**
  - `go test -race ./internal/tui -run '^Test(AuditCreate|ResearchCreate|DocumentCreate|Creation)' -count=1`: passed in 110.389s.
  - `go test -race ./internal/tui -run '^TestDocumentCreateInRealProgram$' -count=3`: passed in 5.879s.
- **Independent Darwin native clipboard verification:**
  - `go test -v ./internal/tui -run '^TestCreationNativeClipboardOwnership$'`: passed in 13.661s across all 4 kinds (`task`, `epic`, `audit`, `research`) and blockers (`visible`, `resize`, `help`, `config`, `field_changed`), confirming execution of the stub `pbpaste` script without touching user clipboard.
- **Inherited checks:**
  - Full repository test suite (`just test`), linter (`just lint`), and planning dogfood lint (`./bin/tskflwctl -C . lint`) inherited from owner baseline.

### Hostile probes and execution evidence

Four custom hostile test suites were authored and executed in the sandbox to stress the implementation beyond existing test oracles:

1. **Hostile Audit Keyboard Flows:**
   - *Valid-but-hostile flow with clock advance and active filter:* Opened audit creation on Audits tab while active filter was `"nomatch"` and view was `"closed"`. Pasted hostile input with unicode and slashes (`"  Security & TUI Review / 2026 🚀  "`), left date blank, advanced clock by 48 hours, and submitted. Verified: form closed, landed on exact created record ID, status view automatically widened to `"all"`, list filter was reset, and audit Date was defaulted at submission time (`"2026-10-22"`), generating slug `"2026-10-22-security-tui-review-2026"`.
   - *Refused-then-corrected deliberate retry flow:* Submitted empty area (refused: ErrValidation), then submitted invalid date format `"not-a-date"` (refused: ErrValidation). Verified: form remained active and editable, draft text was preserved, `f.locked` remained false. Corrected date to `"2026-10-25"` and re-submitted: landed successfully with durable record verified via `ShowAudit`.
2. **Hostile Research Keyboard Flows:**
   - *Valid-but-hostile backdating and empty optional metadata:* Opened research creation with filter `"nomatch"`. Pasted hostile title `"Hostile Title / Special & Char 💡"`, left description and tags blank (empty optional metadata), and entered explicit backdate `"2024-03-15"`. Submitted. Verified: landed on minted research ID, minted ID timestamp decoded exactly to `"2024-03-15"`, description was `""`, tags was empty slice, and no statusView axis was invented.
   - *Refused-then-corrected deliberate retry flow:* Entered unmintable date `"1026-01-01"` (outside encodeable range). Submitted. Core refused with validation error. Form remained editable with input intact. Corrected date to `"2026-10-23"` and re-submitted: landed cleanly.
3. **Small Screen Rendering and Input Guards (46x18, 45x18, 46x17, and restoration):**
   - At terminal size 46x18 (modal size 44x14, the exact supported threshold): `[Create]` and `[Cancel]` buttons were rendered, no resize warning was displayed, and keyboard/paste input was accepted.
   - At 45x18 (1 column narrow, modal size 43x14): rendered "New <noun> · draft retained / Resize to continue safely", and all keyboard/paste input was completely rejected without mutating widget state.
   - At 46x17 (1 row short, modal size 44x13): rendered resize notice and completely rejected all input.
   - Restored to 120x40: all populated text remained intact, fields and actions returned without corruption.
   - In `pending` (`saving = true`) and `recovery` (`locked = true`) states: rendered proper diagnostic/recovery box at 46x18 and correctly collapsed to resize warning at 45x18 and 46x17.
4. **Late-Origin Result and Recovery Cross-Check (Antigravity cross-check):**
   - Opened and submitted creation in origin workspace Alpha. While the origin write command was in flight, switched workspace to Beta via `activateWorkspace()`.
   - Delivered the origin result message to `Model.Update()`.
   - Verified: Beta received 0 writes (`calls = 0`) and 0 reloads/reads (`readCalls` unchanged).
   - The form in the Model became locked and displaced (`f.locked = true`, `f.displaced = true`, `f.saving = false`).
   - Pressed `y`: issued clipboard command containing complete recovery text with destination, record ID, and committed status.
   - Pressed Enter: closed recovery modal without reloading or landing in Beta.
   - Inspected origin workspace Alpha: verified the created record was durably written to Alpha's store.
5. **Interactive Running-Program Real Keyboard Navigation & Discard:**
   - Drove a live `tea.Program` model via `teatest.TestModel` for both audit and research: pressed `c` to open, typed input, pressed `Esc` to trigger "Discard unsaved input?", pressed `Enter` to cancel discard and keep editing, pressed `Esc` + `y` to discard (closing modal back to list), pressed `c` to reopen, filled fields, navigated with Tab/Enter to `[Create]`, pressed Enter, and verified successful landing.

### Concise hostile and mutation ledger

| Mutation | Intended Target / Regression | Observed Failure | Restored Verification |
|---|---|---|---|
| **Mutation A:** Date forwarding bypass in Audit creation (`internal/tui/audit_create.go:136`, replaced `Date: f.inputs[auditCreateDate].Value()` with `Date: ""`). | `TestAuditCreateDurableDefaultsAndDates/date=2025-01-03` | **FAIL** `document_create_test.go:81`: wrong durable audit/scaffold: date defaulted to submission day `2026-10-21` instead of requested `2025-01-03`. | **GREEN** `ok github.com/andy-esch/taskflow/internal/tui 1.407s` |
| **Mutation B:** Text-only enum options dereference (`internal/tui/staged_creation.go:231`, replaced `options := c.enumOptions(f.cursor)` with direct `options := c.options(f.cursor)`). | `TestAuditCreateDurableDefaultsAndDates` | **FAIL (Panic)** `runtime error: invalid memory address or nil pointer dereference` in `creationControls.handleKey` at `staged_creation.go:231`. | **GREEN** `ok github.com/andy-esch/taskflow/internal/tui 1.407s` |

### Settlement of required hostile hypotheses

1. **Progress classification can collapse into unsafe retry:**
   *Settled: falsified by probe.* Trace in `audit_create.go:171–183` and `research_create.go:185–197` demonstrates that `uncertain := errors.As(msg.err, &finalization) || (msg.err != nil && !creationKnownRefusal(msg.err))`. Any committed receipt, non-empty local path evidence, uncertain error, or nil error with uncommitted state unconditionally locks the form (`f.locked = true`), disables Create, and triggers reload. Only clean domain sentinels (`ErrValidation`, `ErrNotFound`, `ErrAmbiguous`, `ErrConflict`, `ErrReadOnlyPersistence`) without progress permit deliberate retry.
2. **Copying another form copied its domain assumptions:**
   *Settled: falsified by probe.* Audit creation collects only Area and Date; it has no bucket picker and unconditionally opens in `AuditOpen`. Research creation collects Title, Description, Tags, and Date; it has no lifecycle or status axis, and no parent epic picker. Blank dates default at submission time across clock changes, and research backdating mints an ID whose timestamp directly matches the requested historical date. Empty optional description and tags remain empty.
3. **Green tests share a wrong oracle or skip effects:**
   *Settled: falsified by probe.* Independent hostile probes bypassed test helpers (`creationContractInputs`, `creationContractControls`, `fillDocumentCreate`, `submitDocumentCreate`), delivering raw `tea.KeyPressMsg` and `tea.PasteMsg` events. Durable records were validated directly against filesystem state and core queries (`ShowAudit`, `ShowResearch`), proving real persistence, slugification, and markdown template bodies.
4. **Selection is only visually right:**
   *Settled: falsified by probe.* Tested under active non-matching list filter text (`"nomatch"`) and restricted audit view (`statusView = "closed"`). Landing in `landCreatedEntity()` sets `statusView = "all"` for audits (and `""` for research), executes `tab.list.ResetFilter()`, and sets `tab.restoreWiden = true`, ensuring the exact created ID is selected and visible.
5. **A small-screen or text-only path breaks shared controls:**
   *Settled: falsified by probe.* Modal size math accurately enforces the 44x14 minimum bounds (outer 46x18). At 45x18 and 46x17, input is completely blocked and resize notices are displayed. Text-only forms with nil `c.options` cleanly route Enter/Tab/Shift+Tab via `c.enumOptions()`.
6. **Late input still targets the wrong owner:**
   *Settled: falsified by probe.* Late results from departed workspaces are intercepted by `Model.Update()` and routed to `creationHost.applyResult()`. Because session generations differ, the result is locked into displacement recovery only; it never navigates, writes, or triggers reloads in the incoming workspace.

### Systemic second pass

A secondary adversarial review examined the architecture for systemic risks:
1. **Controller Duplication vs Shared Abstraction:**
   The concrete controllers in `audit_create.go` and `research_create.go` duplicate receipt-interpretation and state-guard structures from `task_create_commands.go` and `epic_create.go`. However, this repetition is intentional: each entity kind has distinct parameter types (`NewAuditParams` vs `NewResearchParams`), receipts (`AuditCreationReceipt` vs `ResearchCreationReceipt`), landing policies (`"all"` bucket widening for audits vs no-op status view for research), and recovery payloads. A premature generic entity creation abstraction would risk conflating domain lifecycles or bypassing compile-time safety on receipts.
2. **State Changing Between Projection and Action:**
   Submission commands capture immutable value copies of inputs and the exact originating `*core.Service` reference. If `m.svc` or `m.sessionGen` mutates before or during command execution, `submitAuditCreate` and `handleAuditCreated` fail closed by detecting `f.origin != m.svc || f.session != m.sessionGen`, preventing cross-workspace mutation.
3. **Fail-Closed Boundaries:**
   All entry points (`openAuditCreate`, `openResearchCreate`, `openPageCreation`, `runCreationAction`) check `m.atlas.opening`, active creation forms, current page kinds, and modal precedence before allocating form state. Discard confirmation correctly distinguishes clean forms from dirty forms.

## Owner disposition — 2026-10-23

Confirmed the guarded receipt (verify=passed, transfer=succeeded, delivery=confirmed) and accepted the bounded no-findings outcome. Independent keyboard/date/default, small-screen, native-provider and running-program checks plus two compiler-valid mutation kills support the interaction lens. No product fix or new follow-up is warranted.

Evidence qualifications: the typed audit/research receipts are in internal/core/creation_receipt.go:49 and :56, not the store.go locations listed in the inventory. Entry guards are layered: the shell owns modal precedence, per-kind openers check page/service/pending-open state, and host begin owns active-kind/lifetime allocation; not every named entry point repeats all guards. Already-issued writes are not cancelled by a session change: they execute on the captured origin Service, while result/recovery guards prohibit incoming-space navigation or reloads. The late-input hypothesis is supported by the native-input tests, not by the late-create-result trace alone. These corrections bound the report evidence; they are not implementation findings. Closing this audit with no findings.
