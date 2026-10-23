---
schema: 1
id: 6gpgy3n31042
bucket: closed
area: tui-audit-research-creation-implementation-codex
date: "2026-10-23"
updated_at: "2026-10-23"
---
# Audit: TUI audit and research creation implementation — codex — 2026-10-23

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

### Findings and verdict

No findings. Both review passes and the Codex evidence floor are complete for the captured implementation. The tested creation, origin/recovery, retry and exact-selection contracts held. No finding dispositions or task state were changed. This is a bounded review of the captured tree, not certification of later owner edits or every asynchronous terminal interleaving.

### Captured snapshot and scope

- Read the assigned brief in the source, then immediately created `/private/tmp/isolated-review.lcd97x` with the trusted source helper. All implementation discovery, commands, probes, builds and report editing occurred there. The sibling report was not read.
- Reviewed captured baseline `fbef13ca7c82ab42890fa5031b1a47f67ab831c9` against base `05055b8377160015fbb61041c4b374c713c18b0a`. `git log -2` confirmed the former is the helper's capture commit directly above that base. The three added implementation/test files are captured in that baseline, not missing untracked evidence.
- `git diff --name-status <base>` confirmed the implementation surface: two new form controllers and `document_create_test.go`; host, shared-control and shared-test changes; keys, README and TUI guidance. Planning records were present in the capture; unrelated review/manual-AC records were excluded from review.
- `git diff --name-only <base> -- internal/core internal/store internal/cli internal/wire scripts` returned empty. There are no changed optional wire branches, schema/CLI contracts or executable repair advice requiring invented gates. Recovery Enter was nevertheless executed below.

### Verified consumer inventory

All references below were verified in the captured sandbox. Line numbers refer to the restored capture, not a mutation.

| Consumer | Verified ownership and trace |
| --- | --- |
| Audit form | `internal/tui/audit_create.go:15` defines Area, Date, Cancel and Create; `:27` owns the fixed input array and `core.AuditCreationReceipt`. `:78` opens only on the audit page; `:127` copies Area/Date into `core.NewAuditParams` and captures origin/gen/session in the command; `:145` defines `auditCreatedMsg`; `:157` classifies actual receipts; `:185` lands by Audit.ID with audit view `all`; `:190` retains recovery evidence. |
| Research form | `internal/tui/research_create.go:15` defines Title, Description, Tags, Created date, Cancel and Create; `:29` owns its array and `core.ResearchCreationReceipt`. `:90` is the page opener; `:139` copies the four inputs into `core.NewResearchParams`, including Created; `:159` defines `researchCreatedMsg`; `:171` handles results; `:199` lands with no lifecycle axis; `:204` renders recovery evidence. Neither form offers parent, bucket or lifecycle inputs. |
| Host and exposed actions | `internal/tui/creation_host.go:14` has four concrete value slots; `:23`, `:50` and `:66` own palette/page dispatch; `:86` borrows the current slot synchronously; `:111` bounds supported-kind allocation. `:167` and `:172` dispatch typed results only for the active pending kind/attempt/origin session. `internal/tui/command_dispatch.go:236` invokes the palette dispatcher; `internal/tui/model.go:900` routes `c`. `internal/tui/keys.go:78` and `README.md:573` expose the surface. |
| Shared controls and visibility | `internal/tui/staged_creation.go:31` owns lifetime/input tokens; `:92` detaches mutable widget buffers; `:107` scopes deferred widget messages. `:152` safely accepts nil enum options, `:178` handles navigation/cancel and `:261` lands by identity, clears filters and returns reload work. `:291` defines known refusal sentinels. `internal/tui/staged_creation_view.go:23`, `:29`, `:44` and `:51` bind visibility, modal ownership, paste and keys; `:82` and `:157` render pending/recovery and scrolling fields with pinned actions. |
| Delivery, activation and follow-ups | `internal/tui/model.go:284` handles the late-session typed-result exception, `:338` routes ordinary typed results and `:342` guards deferred widget input. `internal/tui/session.go:21` scopes command batches; `:117` activates a workspace; `:192` displaces transient creation. `internal/tui/model.go:236` builds reload commands; `:711` consumes generation-matched lists; `:378` rejects mismatched detail identities; `:1440` captures detail/list generations. `internal/tui/entity.go:238` preserves the pending exact target across reloads. Results and recovery follow-ups were executed, not just inspected as closures. |
| Audit/research reads | `internal/tui/commands.go:233` and `:284` load audit list/detail; `:366` and `:386` load research list/detail. Lists use `record.Source.ID`; detail requests carry the selected key. Audit view widening and research's lack of a view axis were exercised below. `internal/tui/entity.go:471` and `:481` register these consumers. |
| Required application contracts | `internal/core/service_audit.go:12` / `:24` own NewAuditParams/NewAudit, validation, submission-time default, open bucket, ID and scaffold. `internal/core/service_research.go:17` / `:34` own NewResearchParams/NewResearch, optional metadata and date-based minting; `:88` begins the bounded retry loop, with limit eight at `:113`. `internal/core/store.go:379` and `:412` require CreateAudit/CreateResearch. `internal/core/creation_receipt.go:9`, `:18`, `:49`, `:56` define finalization, optional local evidence and typed receipts. No optional audit/research create capability was found or assumed. |
| Real producers | `internal/store/create.go:324` and `:379` implement the typed creates through `createEntityFile` at `:70`, with no-clobber persistence through `writeNewFileUnlocked` at `:125`. Durable failures retain Committed and local destination evidence; precommit failures retain finalization classification. `internal/store/lock_unix.go:21` locks the planning root directory itself, not a guessed lock-file path. |
| Shipped regression inventory | `internal/tui/document_create_test.go:45` and `:95` read real durable metadata/scaffolds/defaults/backdating; `:150` tests deliberate correction; `:198` defines documentCreatePort, with real forwarding only in its explicitly selected outcomes (`:227`, `:255`). `:318` tests portable/recovery outcomes; `:380` core collision retry; `:394` submission guards; `:425` actual stale/duplicate results; `:464` palette/watcher/origin; `:578` populated narrow modes; `:607` real programs. `internal/tui/creation_host_test.go:19`, `internal/tui/staged_creation_test.go:16` and `internal/tui/staged_creation_clipboard_darwin_test.go:34` expand the mechanical matrices to four kinds. |
| Owning guidance | Read `docs/architecture/tui.md:144` (audit/research), `:173` (extension guidance) and `docs/architecture/planning-data.md:222` (change checklist). Their required Store, typed intent/receipt, optional path and origin claims agree with the traced code and bounded probes. |

### Independent checks

Commands ran on Darwin with `GOCACHE=/private/tmp/taskflow-document-review-codex-go`, `GOLANGCI_LINT_CACHE=/private/tmp/taskflow-document-review-codex-lint`, `GOMAXPROCS=4` and `GOFLAGS=-p=4`.

| Check | Observed result |
| --- | --- |
| `go test -race ./internal/tui -run '^Test(AuditCreate\|ResearchCreate\|DocumentCreate\|Creation)' -count=1` | Exit 0; `ok .../internal/tui 106.910s`. Includes the actual Darwin clipboard-provider ownership test. |
| `go test -race ./internal/tui -run '^TestDocumentCreateInRealProgram$' -count=3` | Exit 0; `ok .../internal/tui 6.068s`; both kinds complete through running Bubble Tea programs. |
| Reviewer-only `go test -race ./internal/tui -run '^TestCodexDocument' -count=1 -v` | Exit 0; `60.604s` for the initial late-origin, retry and blank-date probes. Expected domain values are independent literals; the scratch executor bypasses fillDocumentCreate, submitDocumentCreate, creationContractInputs, creationContractControls and settleCreate. |
| Reviewer-only `go test -race ./internal/tui -run '^TestCodexDocumentOpaqueFilteredLanding$' -count=1 -v` | Exit 0; `8.406s`; opaque pathless identities survive filtering, closed audit view and newer reload before original landing. Explicit subsequent detail refresh reads the same source key and expected contents. |
| Reviewer-only `go test -race ./internal/tui -run '^TestCodexDocumentNativeOwnership$' -count=1 -v` | Exit 0; `12.116s`. An isolated PATH contains only a reviewer stub pbpaste. Its invocation marker equals exactly twelve `x` bytes: both new forms × visible/field-move/resize/help/config/close-reopen. Only visible current input accepts `NATIVE-qhjkn-世界`; no real clipboard program is reachable. |
| `go test -race ./internal/store -run '^Test(OrdinaryCreateReportsCommittedGuardReleaseFailure\|NewResearchDoesNotRetryCollisionWhenGuardReleaseFails\|CreateEntityFileKeepsDryRunAndPreCommitFailuresDistinct)$' -count=1 -v` | Exit 0; `1.180s`. Exercises the real guarded producers' release-error hook, including both new consumers and research collision plus failed release. |
| `go test ./internal/tools/architecturecheck -run '^TestArchitecture(Navigation\|ForwardingAnchors)$' -count=1` | Exit 0; `0.183s`. |
| Existing enum consumers | `go test -race ./internal/tui -run '^TestTaskCreateConstrainedOptionsStageCancelAndReset$' -count=1` passed (`3.647s`). Epic durable defaults/overrides passed (`3.111s`); the initial selector also contained a nonexistent task-test alternative, which is not counted as a task check. |
| `just build` | Exit 0. A module stat-cache permission warning appeared; it did not fail the build. |

Owner-reported full build/race/lint and planning/link checks remain inherited evidence, not independent reruns. The owner-reported deletion-of-audit-result-guard and omitted-research-Created mutation kills were not repeated or counted here. No full project race/lint, generator or schema gates were run for this report-only deliverable. Final planning/link lint and isolation results follow below.

### First pass: hostile producer and interaction ledger

Reviewer-only scratch tests were written, compiled and executed inside the sandbox, then removed before transfer. Their names here describe executed review probes, not shipped regressions.

- **TestCodexDocumentLateOrigin:** eight cases (two kinds × success, progress-with-conflict, finalization-wrapping-conflict, unknown). Issued the actual UI command before activation, executed activation and its loads, then executed the held create against its captured origin. Independently read origin durability. For committed cases, inserted incoming records with the exact receipt ID and different area/description/body. Captured incoming bytes, directory/file modes, all required Store list/detail read counters and create counters after fixture setup. Delivered the actual scoped result twice, executed all finite follow-ups, pressed recovery Enter and executed its follow-ups. Incoming counters/bytes/modes remained unchanged; each origin saw exactly one Create. Recovery retained committed identity and did not mutate a previously retained Model copy.
- **TestCodexDocumentRetryAndKeyboard:** eight cases with explicit keyboard Create. A known uncommitted read-only refusal retained the draft and allowed one deliberate retry after lifting the injected refusal (two total producer calls). A real FS write followed by injected conflict or finalization wrapping conflict disabled replay; recovery Enter loaded the exact durable document and body. Unknown outcomes stayed locked and closed without another write. Each non-refusal had exactly one producer call. Error injection around real FS creates is a controlled port fault, not a claim that an arbitrary transport adapter ships here.
- **TestCodexDocumentBlankDateTextOnly:** both forms crossed Tab, Shift+Tab and literal Enter row sequences without enum callbacks. Opened at `2026-10-23 23:59 UTC`, advanced the clock two minutes before Create, and read literal durable date `2026-10-24`. Research description/tags remained empty and its ID decoded to that day. The hostile non-default flow separately persisted audit date/research Created `2024-02-29`, description `Description: independently literal`, tags `alpha` and `βeta`, and the research ID decoded to `2024-02-29`.
- **TestCodexDocumentOpaqueFilteredLanding:** a reviewer portable producer returns IDs `urn:codex:audit:created` / `urn:codex:research:created`, with no local-path capability and an untrusted receipt destination `/opaque/never-open-this`. Set a nonmatching filter and closed audit view, delivered success, completed a newer reload before the original landing's commands, then executed another detail refresh. Exact selected source key and literal detail contents held; audit widened to `all`, research retained no status view, no editor promise appeared and fixture bytes/modes stayed unchanged. This is a fabricated portable seam used to challenge path authority, not a shipped remote adapter.
- **Native ownership / sizes:** the independent actual provider probe above rejected stale native deliveries after field movement, hidden higher modal, 45×18 resize and close/reopen. The independently executed shipped four-kind visibility matrix covers 46×18, 45×18 and 46×17, while populated document narrow-mode tests cover fields/discard/pending/recovery and restoration. Nil enum policy permits the actual Enter/Tab/Shift+Tab/cancel paths without panic; the unchanged task/epic enum policies have their own tests above.

The scratch draft initially failed compilation because it assumed a domain Path field; those assumptions were removed. Those compiler failures are not mutation kills or product findings. Domain Path was not used as identity or authority in the executed probes.

### Independent mutation and restored green

Changed only the recovery Enter guard in both new controllers (`internal/tui/audit_create.go:101`, `internal/tui/research_create.go:113`): removed `!f.displaced &&` from the committed-receipt condition. This is a compiler-valid authority mutation, stronger for this lens than merely deleting a result-routing guard: the result was correctly retained in displaced recovery, and the probe had to execute the subsequent action and reads to expose the escape. No extra caller mutation was needed; the host does not recheck origin for recovery Enter.

Intended test, identical before mutation and after restoration:

```sh
go test -race ./internal/tui -run '^TestCodexDocumentLateOrigin$' -count=1 -v
```

- Mutated: exit 1, `26.907s`. Six committed cases failed at the independent counter assertion: `late origin escaped: incoming reads 2->4 writes 0->0 origin writes=1`. Both unknown cases remained green. The failures came after recovery follow-ups actually ran, not compilation or unrelated assertions.
- Restored both controller files from the captured sandbox HEAD: identical command exit 0, `27.964s`, all eight cases green. No implementation changes remain.

### Second pass: systemic challenge and hypothesis disposition

Re-read the host, shared controls, modal ownership, all four consumers' enum binding, per-kind result/recovery duplication, session exception and generation-matched list/detail callers after the first pass. The principal systemic risk is letting valid origin evidence acquire incoming-service authority through a later caller. The combined same-ID/duplicate/recovery probe and the killed recovery mutation demonstrate that this risk was exercised at its effects, and the restored code rejects it. Merely counting Create calls would have missed the mutation's two incoming reads.

The shared test inventories are mechanical accessors (`internal/tui/creation_host_test.go:40`, `:54`), and submitDocumentCreate chooses the last production row (`internal/tui/document_create_test.go:39`). They could jointly conceal a row/order error if used as the domain oracle. The independent probe therefore filled literal keyboard sequences and compared real durable fields, dates, decoded ID dates and detail contents to literals. documentCreatePort deliberately fabricates preview/unconfirmed/path evidence; only its documented success/cleanup/collision branches call a real store. It is useful defense coverage, not proof those malformed outcomes arise from FS.

The finite settling helper (`internal/tui/staged_creation_helpers_test.go:31`) executes batches and follow-ups and suppresses blink maintenance, with an operation budget. The independent probe used a separate bounded command queue with the same explicit maintenance limit. Neither executor is an asynchronous scheduler or a timeout; runtime claims are limited to the actual repeated Bubble Tea test and actual native-provider executions. The shipped held-native-result running-program resize fixture remains epic-specific (`internal/tui/staged_creation_clipboard_darwin_test.go:124`); this review does not relabel it as a new-form runtime interleaving.

| Challenged hypothesis | Disposition and evidence |
| --- | --- |
| Progress collapses into safe retry | Falsified for shipped producer contracts and tested port faults. Actual FS release-error tests preserve committed/finalization evidence. Both independent kinds suppress UI replay after progress+sentinel, finalization+sentinel and unknown outcomes. The focused portable matrix also locks preview, unconfirmed and committed-path evidence. A contradictory adapter returning durable path evidence with neither Committed nor finalization on conflict is not established as a shipped producer; no claim is made about all malformed adapter combinations. |
| Copied task defaults corrupt area/date or research optional values | Falsified by independent literal durable reads, empty optional metadata, midnight defaulting, explicit leap-day backdating and decoded research ID day. |
| Green shared helpers skip effects or share the wrong oracle | Challenged and bounded: independent input/expectation/executor path, actual durable reads, native invocation marker, exact detail checks and an effect-detecting killed mutation. No helper-derived durable expectations. |
| Selection looks right but targets the wrong record | Falsified by opaque pathless source IDs, nonmatching filter, closed audit view, newer reload before landing and a subsequent exact-key detail read; same-ID incoming origin probe also counts reads. |
| Text-only/small-screen controls panic or edit behind the screen | Falsified within the executed visibility, populated narrow-mode, native-owner and literal keyboard matrices. Existing task/epic enum checks remain green. No claim about cosmetic preferences. |
| Late input changes a different field/form/modal | Falsified for the actual stub-native visible/moved/hidden/resized/reopened cases, plus the executed shared buffer and ownership regressions. Every new-form asynchronous runtime interleaving remains unexhausted. |
| Adding the next kind requires another root policy branch | No demonstrated consequence. The concrete host slot/opener and typed route remain explicit; Model.Update, session activation and modal ownership need no new per-kind branch. Per-kind receipt-policy duplication is visible and intentional here; this pass found no reachable divergence warranting an abstraction finding. Thread/membership and broader paste work remain outside this slice. |

No unresolved in-scope defect was demonstrated. Remaining limits are the nonexhaustive scheduler/terminal state space and contradictory adapter outcomes beyond the verified production contract, not evidence that planned tests or adapters have shipped.

### Final validation and isolation attestation

`./bin/tskflwctl -C . lint --links` passed (exit 0: all planning entities and dependency links pass lint). `./bin/tskflwctl -C . audit lint 6gpgy3n31042` passed (exit 0: all audit findings pass lint). `git diff --check` passed. The pre-report brief bytes were compared to captured HEAD and are identical. Reviewer scratch tests were removed, both mutated implementation files match captured HEAD, and the generated sandbox binary was removed after validation. Full project gates remain attributed to the owner as above.

Trusted-source helper verification (verbatim):

```text
sandbox_path=/private/tmp/isolated-review.lcd97x
git_dir=/private/tmp/isolated-review.lcd97x/.git
baseline_commit=fbef13ca7c82ab42890fa5031b1a47f67ab831c9
source_blob=851aa2a94db087253ea230518c6429c7a51785eb
source_fingerprint=7486b074eeef696b3ae49ce916d50fd4b17010bc
deliverable=planning/audits/6gpgy3n31042-2026-10-23-tui-audit-research-creation-implementation-codex.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

Only the assigned audit differs from the captured baseline. The report is final; transfer is pending as required. The helper-owned transfer receipt, returned separately in the handoff, is the delivery evidence. The independent workspace remains in place for owner confirmation.

## Owner disposition — 2026-10-23

Accepted the bounded no-findings verdict after reading both review passes and confirming the trusted source-helper receipt (verify=passed, transfer=succeeded, delivery=confirmed). Independent literal keyboard/default/backdating, opaque-identity landing, native-provider and real producer checks complement the owner baseline. The recovery-Enter mutation exposed incoming reads after actual follow-up execution and passed after restoration; it is not merely a no-second-write assertion. Full project gates remain inherited evidence, and not every asynchronous interleaving is claimed. No implementation fix or new follow-up is warranted. Closing this audit; Antigravity remains an independent pending assignment, not corroboration or a completed review.
