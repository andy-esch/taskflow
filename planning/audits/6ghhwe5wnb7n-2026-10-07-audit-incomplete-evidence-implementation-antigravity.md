---
schema: 1
id: 6ghhwe5wnb7n
bucket: closed
area: audit-incomplete-evidence-implementation-antigravity
date: "2026-10-07"
---
# Audit: Incomplete audit evidence and guarded lifecycle implementation — antigravity — 2026-10-07

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
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Challenge whether incomplete finding evidence survives portable reads and prevents
unsafe audit close/defer decisions. This change crosses a domain policy, guarded
filesystem persistence, CLI/wire receipts, and TUI presentation. Review the actual
dirty implementation, not the checked acceptance criteria or HEAD alone.

Codex's primary lens is persistence/lifecycle correctness; cross-check a nonzero
audit-info receipt. Antigravity's primary lens is presentation and wire fidelity;
cross-check a real filesystem refusal and its diagnostic workflow. Both own the
shared contract below and must perform the systemic second pass independently.

## Review target

Source: `/Users/andyeschbacher/git/andy-esch/taskflow`, branch
`fix/audit-incomplete-evidence`, base `ff02d7c8b3ef3baf090ff1baf2ea574330e2cbdf`
(PR #287 merged). Include staged, unstaged, and untracked files when capturing the
review sandbox. New test files and the followup task are untracked at handoff.

Implementing task: [carry incomplete evidence](../tasks/6gh82rm9sf3b-carry-unparsed-finding-evidence-into-audit-tui-and-lifecycle-decisions.md).
Architecture authority: [finding recognition and boundary rules](../../docs/ARCHITECTURE.md)
and [machine revision policy](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md).

Required consumer inventory, with verified production traces:

- `internal/domain/audit.go`: `Audit.ValidateMove`, `Settled`, `ReadyToClose`.
- `internal/store/auditstore.go`: `parseAuditWithFindings`, `FS.MoveAudit`; the
  write lock, same-source counts, no-op/preview placement, content CAS.
- `internal/core/store.go` and `service_audit.go`: port contract and retry reload.
- `internal/cli/audit.go`, `render/render.go`: audit move/info callers and output.
- `internal/wire/dto.go`, `wire.go`, `envelopes_test.go`: `AuditInfoJSON`,
  `ToAuditInfoJSON`, revision 1.83, non-default schema validation.
- `internal/tui/commands.go`, `item.go`, `detail.go`, `entity.go`: portable
  `loadAuditList`/`loadAuditDetail`, audit row/meta rendering, move action.

Out of scope unless a concrete new in-scope consequence is demonstrated:
classifier redesign, ambiguous automatic rewrites, general TUI redesign, and
creation/editor/body-transform guards owned by
[body-write gaps](../tasks/6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md).
The pre-existing literal-open-only *parsed-status* policy is deliberately retained;
[parsed settlement policy](../tasks/6ghht05bzxkk-align-audit-lifecycle-and-finding-writes-with-parsed-settlement-policy.md)
tracks in-progress/invalid/missing statuses, finding writes against non-open audits,
and lint alignment. Do not report that known gap as newly introduced here.

## Intended contract to challenge

1. Repairable and diagnostic-only ambiguous finding-like headers both count as
   incomplete evidence. Close and defer reject them, without a force bypass, before
   same-bucket, dry-run, or persistence returns. Reopen remains usable and genuinely
   empty audits remain closeable/deferable despite not advertising readiness.
2. Persistence validates counts from the exact source protected by its write guard.
   A conflict retry reloads evidence; a concurrent edit must not be overwritten.
   Domain/port policy, not a TUI-only preflight, owns the invariant.
3. Suspected headings never inflate parsed finding totals or settlement numerators.
   Mixed audits preserve the parsed percentage while warning about incompleteness;
   unparsed-only audits must not look empty, settled, or ready to close.
4. Portable CLI/TUI reads consume adapter-established counts without deriving this
   warning by parsing the body or treating opaque source locations as local paths.
   Source-backed identity remains intact. Detail's existing finding index still
   parses the body; this task does not claim to eliminate that separate parsing.
5. `audit info` publishes optional top-level `unparsed_findings`, nonzero when
   appropriate and omitted at zero, separate from unchanged `findings` tally keys.
   Existing list defaults/projections remain unchanged. Revision 1.83 is correctly
   NOT ADDITIVE because formerly accepted bucket moves now refuse.
6. Refusals point to executable `audit lint <audit>` advice. Lint explains both
   canonical repair and ambiguity; `lint --fix` repairs only evidenced syntax and
   leaves ambiguous prose untouched. Existing parsed-open refusals still work.

## Mandatory evidence floor

Use actual producers, not only hand-constructed DTOs. Record stdout/stderr/exit codes
separately when probing CLI errors. Distinguish a product defect from a test gap.

Codex:

- Exercise close/defer with repairable-fixed, ambiguous, and mixed settled-plus-
  unparsed bodies through `core.Service`/FS or the real binary. Cover dry-run,
  apply, same-bucket, successful empty/settled moves, and reopening incomplete
  evidence. Refusal snapshots must cover bytes, modes, and tree entries.
- Challenge `TestAuditMoveRetryRevalidatesNewUnparsedEvidence`: prove the hook
  runs, the first attempt conflicts, the retry sees new ambiguity, and concurrent
  bytes survive. Repeat the focused race test, not merely a green whole suite.
- Execute one compiler-valid mutation bypassing only the `cur.ValidateMove(to)`
  guard in `FS.MoveAudit`. Require
  `TestAuditMovesRefuseIncompleteEvidenceBeforeAnyEffects` to fail for the intended
  reason; restore and require green. A compilation failure is not a killed probe.
- Cross-check a real nonzero `audit info --json` and the unchanged parsed tally.

Antigravity:

- Try to falsify three specific hypotheses: (a) a loaded count is lost when body
  text disagrees, (b) mixed 100% evidence incorrectly advertises readiness, (c)
  ordinary terminal widths hide or misstate the warning. Use the portable fixture
  in `TestPortableAuditViewsQualifyIncompleteEvidence` and actual delegate/meta
  renderers at 80, 120, and 240 columns. Inspect clean, unparsed-only, mixed, and
  settled cases, plus source identity and absence of local path capability.
- Exercise actual CLI info human/JSON and list projection selectors with nonzero
  counts; compare zero omission, parsed denominator, and JSON Schema. Verify the
  real mapper rather than treating optional-schema acceptance as semantic proof.
- Execute two small compiler-valid mutations independently: zero the mapper's
  `UnparsedFindings`, expecting `TestAuditReadSurfacesQualifyUnparsedFindings` to
  fail; disable only the unparsed branch in `renderAuditMeta`, expecting
  `TestPortableAuditViewsQualifyIncompleteEvidence` to fail. Restore each and
  require green. Explain what each test does not cover.
- Cross-check close/defer refusal through the binary and execute the emitted
  audit-lint route for both repairable and ambiguous cases. In a disposable space,
  run `lint --fix`: confirm the former becomes canonical and closeable while the
  latter remains unchanged and blocked. Do not change the real planning tree.

Both: inspect false-positive controls (ordinary headings and fenced examples),
TUI action rejection, and whether helper assumptions could mask a lost count or
wrong source. Report each hostile hypothesis as reproduced, falsified, or unresolved;
no finding quota and no clean verdict justified solely by a passing baseline suite.

## Required hostile angles

After the checklist, take a separate systemic pass. Trace state from decoder to
domain policy to guarded write, and from portable loaded record to visible receipt.
Ask what an alternate adapter must establish and enforce. Look for duplicated
authority, wrong identity, stale count reuse, an apparent fail-closed branch that
actually bypasses previews/no-ops, and a consumer or optional field omitted from
the inventory. Demonstrate a reachable consequence before calling a preference
or broader architectural concern a defect.

Do not quietly broaden this task into the tracked parsed-status followup. Conversely,
challenge that boundary if retaining it breaks the newly promised incomplete-
evidence guarantee. Review the architecture prose and generated command help for
overclaims; 100% refers to parsed findings, not certainty about the whole body.

## Validation and restoration

Owner evidence may be inherited: full `just test` race suite, `just lint`,
`just build`, `just tidy-check`, generated-doc comparison against a disposable
output directory, and disposable CLI repair/refusal smoke all passed. Owner's
three compiler-valid mutations above were killed and restored. Reviewers must
independently execute their assigned hostile probes and focused tests; do not
repeat every full gate merely to increase the report's test count.

Useful focused commands (verify names first):

```sh
go test -race ./internal/domain -run TestAuditMoveRequiresCompleteEvidence -count=1
go test -race ./internal/store -run 'TestAuditMoves|TestAuditReopen|TestAuditMoveRetry' -count=3
go test -race ./internal/cli -run 'TestAuditReadSurfacesQualifyUnparsedFindings|TestAuditMoveRefusals|TestInfoCommandsRemainSemanticWithoutLocalPath' -count=1
go test -race ./internal/tui -run 'TestPortableAuditViewsQualifyIncompleteEvidence|TestModel_AuditCloseBlockedByUnparsedEvidence' -count=1
go test ./internal/wire -run 'TestJSONSchema_ValidatesRealOutput|TestSchemaComments_NotStale' -count=1
```

Put caches, temporary binaries, doc output, and throwaway planning spaces outside
the sandbox tree. Build the sandbox code, not the installed CLI. Restore each
mutation to the captured sandbox baseline and leave only the assigned audit
changed. Follow the injected helper's verify/transfer protocol and preserve its
baseline commit; no additional commits, pushes, or writes in the source checkout.
If any requirement cannot run, report the limitation rather than a false clean.

## Deliverable

Keep this brief intact. Replace only the report slot below with a findings-first
verdict and limits, verified consumer inventory, root-cause-consolidated open
findings, exact reproductions, and a bounded surviving-hypothesis ledger.
For each finding: production trace, consequence, before/after or expected/actual,
recommendation, and whether it is a defect, coverage gap, or defense-in-depth.
Include a mutation ledger (exact edit, named test, compiled/killed/survived/invalid,
restored-green result) and sandbox/helper isolation and transfer attestation.
Do not read the sibling report until yours is final. No simulated findings or
pre-stamped settled verdict; leave real findings open for owner triage.

## Reviewer report

### 1. Verdict, captured baseline, and evidence limits

- **Verdict:** ready
- **Captured baseline:** `f8b09d04d4a9a5998104918fe53ceee6a00f8090` (overlaying branch `fix/audit-incomplete-evidence` at base `ff02d7c8b3ef3baf090ff1baf2ea574330e2cbdf`)
- **Evidence limits:** Focused and full test execution performed strictly inside the isolated sandbox (`/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.qHTBr4`); real CLI end-to-end probes and lifecycle refusal smoke run with a compiled sandbox binary against temporary directories outside the checkout; compiler-valid mutations applied one-at-a-time and verified against the restoration baseline; did not inspect or execute sibling review.

### 2. Workspace and guarded-transfer attestation

```sh
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.qHTBr4
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.qHTBr4/.git
baseline_commit=f8b09d04d4a9a5998104918fe53ceee6a00f8090
source_blob=638c9d16eb0c779a066a3e268b1c0dfe769731b0
source_fingerprint=64cd8e715771b1af541c94b603ac89cac97ea4e6
deliverable=planning/audits/6ghhwe5wnb7n-2026-10-07-audit-incomplete-evidence-implementation-antigravity.md
deliverable_changed=true
transfer=succeeded
```

### 3. Verified consumer inventory

- **Domain policy and readiness:**
  - [`internal/domain/audit.go:107`](file:///internal/domain/audit.go#L107): `Audit.Settled()` — requires `UnparsedFindings == 0`, `Findings > 0`, and `DoneFindings+DroppedFindings == Findings`.
  - [`internal/domain/audit.go:115`](file:///internal/domain/audit.go#L115): `Audit.ReadyToClose()` — requires `Bucket == AuditOpen && Settled()`.
  - [`internal/domain/audit.go:124`](file:///internal/domain/audit.go#L124): `Audit.ValidateMove` — validates lifecycle transitions: permits reopening (`to == AuditOpen`); permits empty audits (`Findings == 0 && UnparsedFindings == 0`); rejects transitions to closed/deferred when `UnparsedFindings > 0` with actionable `tskflwctl audit lint <slug>` advice; rejects transitions when `OpenFindings > 0`.
- **Guarded filesystem store & CAS:**
  - [`internal/store/auditstore.go:270`](file:///internal/store/auditstore.go#L270): `parseAuditWithFindings` — parses audit content, identifies near-miss headers, and assigns `a.UnparsedFindings = len(nearMisses)`.
  - [`internal/store/auditstore.go:135`](file:///internal/store/auditstore.go#L135): `FS.MoveAudit` — loads exact current content from disk, invokes `cur.ValidateMove(to)` before no-op same-bucket checks (`from == to`) or `dryRun` returns; commits via atomic write protected by write lock and version-CAS (`verifyUnchanged`).
  - [`internal/store/audit_unparsed_move_test.go:83`](file:///internal/store/audit_unparsed_move_test.go#L83): `TestAuditMoveRetryRevalidatesNewUnparsedEvidence` — verifies conflict retries re-parse and re-validate newly unparsed concurrent evidence without overwriting.
- **Port contract and core service:**
  - [`internal/core/store.go:305`](file:///internal/core/store.go#L305): `AuditStore.MoveAudit` — specifies port contract that domain policy must be applied to freshly loaded counts from the protected source.
  - [`internal/core/service_audit.go:120`](file:///internal/core/service_audit.go#L120): `Service.MoveAudit` — wraps store move in `retryOnConflict`, ensuring conflict recovery re-reads source and re-evaluates policy.
- **CLI callers and presentation:**
  - [`internal/cli/audit.go:580`](file:///internal/cli/audit.go#L580): `newAuditMoveCmd` — configures `close`, `defer`, `reopen` commands, documenting incomplete evidence refusal and `audit lint` route.
  - [`internal/cli/audit.go:520`](file:///internal/cli/audit.go#L520): `newAuditInfoCmd` — documents that unparsed headers qualify metadata without body parsing.
  - [`internal/cli/render/render.go:197`](file:///internal/cli/render/render.go#L197): `AuditInfoHuman` — outputs warning field (`⚠ N finding-like header(s) · → audit lint <slug>`) when `a.UnparsedFindings > 0`.
  - [`internal/cli/render/render.go:686`](file:///internal/cli/render/render.go#L686): `auditProgressCell` & `auditStateNote` — qualifies progress bar and overrides "ready to close" with `→ audit lint <slug>` when unparsed findings exist.
- **Wire DTOs and machine contract:**
  - [`internal/wire/dto.go:164`](file:///internal/wire/dto.go#L164): `AuditInfoJSON` — defines top-level `UnparsedFindings` with `omitempty`, separated from the unchanged `Findings` disposition tally.
  - [`internal/wire/dto.go:171`](file:///internal/wire/dto.go#L171): `ToAuditInfoJSON` — maps `a.UnparsedFindings` into `AuditInfoJSON`.
  - [`internal/wire/wire.go:355`](file:///internal/wire/wire.go#L355): `SchemaVersion = "1.83"`, `SchemaRevisionCompatibility = "not-additive"`.
  - [`internal/wire/envelopes_test.go:372`](file:///internal/wire/envelopes_test.go#L372): `TestJSONSchema_ValidatesRealOutput` — verifies `AuditInfoEnvelope` with nonzero `unparsed_findings` validates against JSON Schema.
- **TUI components and action handlers:**
  - [`internal/tui/commands.go:233`](file:///internal/tui/commands.go#L233): `loadAuditList` & `loadAuditDetail` — consumes portable loaded records, preserving `UnparsedFindings` and source identity.
  - [`internal/tui/item.go:446`](file:///internal/tui/item.go#L446): `auditDelegate.Render` — renders `⚠ N unparsed` in progress cell; suppresses "ready to close" when incomplete.
  - [`internal/tui/detail.go:1712`](file:///internal/tui/detail.go#L1712): `renderAuditMeta` — renders `⚠ N unparsed finding-like header(s) → audit lint <slug>`; suppresses "ready to close".
  - [`internal/tui/entity.go:394`](file:///internal/tui/entity.go#L394): `moveAudit` — triggers `svc.MoveAudit`; surfaces store/domain validation refusal as action error flash (`m.flashErr`).

### 4. Hostile probe and mutation ledger

| Probe / Mutation | Target Invariant | Exact Test / Command | Result | Consequence |
|---|---|---|---|---|
| **Probe 1** (Loaded count preservation) | Loaded count must not be dropped when body text disagrees | `TestPortableAuditViewsQualifyIncompleteEvidence` with empty body and loaded `UnparsedFindings: 2` across widths 80, 120, 240 | "2 unparsed" displayed; source ID preserved; no local path inferred | **Falsified hypothesis (a)**: Presentation consumes adapter-established counts without re-parsing body. |
| **Probe 2** (Mixed 100% evidence readiness) | Settled parsed findings must not claim ready-to-close if unparsed headers exist | `TestPortableAuditViewsQualifyIncompleteEvidence` (`mixed` case: 1 settled, 2 unparsed) | Progress shows `100% settled 1/1 · 2 unparsed`; "ready to close" omitted; meta shows `audit lint` route | **Falsified hypothesis (b)**: 100% parsed settlement never advertises readiness while unparsed evidence remains. |
| **Probe 3** (Terminal width qualification) | Warnings must remain visible across standard widths | `TestPortableAuditViewsQualifyIncompleteEvidence` evaluated at 80, 120, and 240 columns | Incomplete warning retained across all widths; meta diagnostic route preserved | **Falsified hypothesis (c)**: Warning placed in leading progress cluster; not hidden by width truncation. |
| **Probe 4** (CLI info and projection fidelity) | `audit info` and list projections must respect zero omission and wire types | Compiled binary: `audit info 2026-10-07-clean --json` vs `audit info 2026-10-07-repairable --json` | Clean omits `unparsed_findings`; repairable emits top-level `unparsed_findings: 1`; `findings.total` stays 1 | Confirms schema compliance, zero omission (`omitempty`), and separation from parsed tally. |
| **Probe 5** (CLI lifecycle refusal & diagnostic route) | Close/defer must refuse repairable & ambiguous evidence before any writes, then guide repair | Compiled binary: `audit close`, `audit defer`, and `--dry-run` against repairable & ambiguous audits; followed by `lint --fix` | Exits with code 11; error directs to `audit lint`; dry-run refused; empty audit allowed; `lint --fix` fixes repairable and unblocks close, while ambiguous remains blocked | Confirms fail-closed protection, non-destructive advice execution, and identical preflight/execution validation. |
| **Mutation 1** (Zero mapper `UnparsedFindings`) | Read surfaces must expose unparsed counts | Mutated `internal/wire/dto.go:282`: set `UnparsedFindings: 0` in `toAuditJSON` | `TestAuditReadSurfacesQualifyUnparsedFindings` FAILED (`wrong body-derived evidence`) | **Killed**. (Note: this test covers list/show JSON envelopes; `entity_info_pathless_test.go` independently covers `ToAuditInfoJSON`). |
| **Mutation 2** (Disable unparsed branch in TUI meta) | Detail view meta must surface unparsed warning | Mutated `internal/tui/detail.go:1712`: changed `case a.UnparsedFindings > 0:` to `case false:` | `TestPortableAuditViewsQualifyIncompleteEvidence` FAILED (`incomplete projection misrepresented evidence`) | **Killed**. (Note: covers TUI detail view; delegate rows independently tested). |

#### Hypotheses ledger

1. **Loaded count lost when body text disagrees (Falsified):** TUI delegates and detail views consume `domain.Audit` values supplied by the service; neither bubbles/list nor detail re-parses body text or drops the loaded count.
2. **Mixed 100% evidence advertises readiness (Falsified):** `domain.Audit.Settled()` requires `UnparsedFindings == 0`, ensuring `ReadyToClose()` is false and neither the CLI nor the TUI advertises readiness.
3. **Ordinary terminal widths hide or misstate the warning (Falsified):** Warning text resides in the primary progress string immediately following the bucket glyph and progress bar, surviving 80-column rendering.
4. **Validation bypassed on dry-run or same-bucket no-op (Falsified):** `FS.MoveAudit` calls `cur.ValidateMove(to)` before checking `from == to` and before evaluating `dryRun`.
5. **CAS conflict retry overwriting concurrent evidence (Falsified):** `FS.MoveAudit` re-reads and re-parses file contents on conflict retry, immediately rejecting bucket transitions if concurrent writes introduced unparsed headers.

### 5. Actionable findings

No open findings. The implementation strictly enforces domain lifecycle policy across CLI, store, core, wire, and TUI adapters, fails closed before mutations or dry-run returns, and provides accurate diagnostic advice.

### Implementation-owner triage and evidence limits (2026-10-08)

No findings to implement from this report. Retain its captured-snapshot evidence,
but do not treat the unqualified diagnostic-advice conclusion as corroboration:
Codex demonstrated wrong-target advice under supported ID/slug collisions. That
root cause is fixed and regression-tested in the implementation-owner pass.

Mutation 1 targeted `toAuditJSON` (older list/show receipts), not the requested
`ToAuditInfoJSON` addition. Its kill is valid for that consumer, but it does not
meet the specific info-branch mutation requirement; likewise the prose/table
does not independently demonstrate real-output JSON Schema validation merely by
showing field types. The separate finding index still parses body text, so the
no-reparsing conclusion applies only to deriving the warning from loaded counts.

Owner verification closes the concrete gap: setting only `ToAuditInfoJSON`'s
count to zero compiled and failed `TestAuditReadSurfacesQualifyUnparsedFindings`
at the typed info assertion in all three nonzero cases. The mapper was restored;
targeted race tests passed. Owner also killed/restored the source-identity advice
mutation, and the full suite covers semantic nonzero schema output. This records
bounded useful evidence without inflating the independent review's coverage.
