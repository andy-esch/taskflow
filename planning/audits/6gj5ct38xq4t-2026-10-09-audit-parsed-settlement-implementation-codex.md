---
schema: 1
id: 6gj5ct38xq4t
bucket: closed
area: audit-parsed-settlement-implementation-codex
date: "2026-10-09"
updated_at: "2026-10-09"
---
# Audit: Audit parsed settlement and guarded finding writes — codex — 2026-10-09

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

Challenge whether audit lifecycle, structured finding writes, readiness, and lint
now share a sound parsed-settlement policy without racing or locking out recovery.
This changes previously accepted writes, not only presentation. Do not use green
tests or checked ACs as evidence of correctness.

Codex: primary lens is guarded persistence, CAS revalidation, and repair availability;
cross-check source-backed diagnostics. Antigravity: primary lens is operator/machine
contract and independent status-policy oracles; cross-check both CAS races. Neither
may read the sibling report until its own report is final. The pair receives this
same brief; each assignment selects a complementary evidence path, not a verdict.

## Review target

Source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`.
Branch: `fix/audit-parsed-settlement`. Base: merged PR #288,
`3b70c1909d00` (verify the full base hash in the sandbox).
Review the captured staged, unstaged, untracked, and deleted state, not HEAD alone.

Implementing task:
[parsed settlement policy](../tasks/6ghht05bzxkk-align-audit-lifecycle-and-finding-writes-with-parsed-settlement-policy.md).
Approved policy is in its body. The preceding unparsed-evidence rules must survive.
Known adjacent scope:
[arbitrary body-write guards](../tasks/6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md).
CLI append can already add an in-progress finding to a closed empty audit; it is
tracked there and new lint detects it. Whole-file editing/creation-body guards are
not claimed solved here. Challenge this boundary if there is a distinct in-scope
consequence, but do not relabel that documented pre-existing route as a new defect.

Verify a consumer inventory before conclusions:

- `internal/domain/audit.go`: `UnsettledFindings`, `Settled`, `ReadyToClose`,
  `ValidateMove`, `AuditUnsettledFindingsError`.
- `internal/domain/finding.go`: `TallyFindings`, `TerminalFindingStatus`,
  `LintFindings`, `SetFindingStatus`; token normalization versus decoration.
- `internal/core/finding.go`: `EditFinding` and `FindingEdit.apply`;
  `internal/core/finding_create.go`: existing open-bucket creation gate.
- `internal/core/audit_move_error.go`: `AuditMoveError`; `core/store.go` port
  promises; `store/auditstore.go: MoveAudit`; `store/body.go: TransformAuditBody`.
- CLI `audit.go`, ordinary lint consumers, TUI move/list/detail consumers,
  `wire/dto.go`, revision 1.84 and generated artifacts.

Actual new regression entry points:
`TestParsedSettlementPolicyAgreesAcrossReadinessMovesAndLint`,
`TestEditFindingPortableBucketPolicy`,
`TestAuditMovesRequireTerminalParsedFindingsBeforeEffects`,
`TestAuditMoveRetryRevalidatesUnsettledParsedStatuses`,
`TestFindingEditsCannotReactivateNonOpenAudit`,
`TestFindingEditsAllowTerminalRepairAndMetadataWithoutCertifyingWholeAudit`,
`TestFindingEditRetryObservesConcurrentNonOpenBucket`,
`TestAuditParsedSettlementCLIRefusalsAndLint`,
`TestAuditUnsettledAdviceExecutesAgainstSourceNotDeclarationOrSlug`,
`TestAuditFindingCLIRequiresExplicitReopenToReactivate`.
Existing read/projection tests were expanded, including
`TestPortableAuditViewsQualifyIncompleteEvidence` in the TUI. Verify these symbols.

## Intended contract to challenge

1. Every parsed status must be terminal for close/defer: fixed, tracked, deferred,
   superseded, wontfix. Open, in-progress, missing, invalid statuses refuse, including
   same-bucket/no-op calls and previews. Unparsed evidence still refuses separately.
2. Empty audits may close/defer but are not advertised ready-to-close. Explicit
   reopen remains available for any readable audit. There is no force bypass.
3. Structured explicit-status edits cannot leave their selected finding nonterminal
   in a non-open audit, including no-ops/previews. Case/decorated input is evaluated
   after normalization. Combined status/note/candidate refusals have no partial effects.
4. Terminal corrections and note/candidate-only edits remain available despite
   unrelated baseline defects. They do not certify that the whole audit is settled.
   Terminal-token policy does not absorb independent metadata lint (tracked destination,
   resolution-note shape, candidate projection). Do not expand the agreed policy silently.
5. The write decision uses the exact metadata/body protected by the content CAS.
   A retry rechecks both, preserving a concurrent writer's bytes. No preflight read
   may grant lasting authority or let a cached open bucket bypass a later close.
6. Domain/core remain adapter-neutral; executable advice retains adapter-established
   Source.ID, never a display slug, declared ID, or opaque Location. An open audit with
   in-progress work may lint clean; the findings advice must reveal actual work.
7. Readiness, terminal mapping, and lint agree without changing literal-open counts
   or denominators. Revision 1.84 is NOT ADDITIVE; no new JSON data field is claimed.

## Mandatory evidence floor

Both reviewers must execute a fresh binary workflow in a disposable, unregistered
planning space: active refusal, terminal close, closed/deferred status refusal,
explicit reopen, and empty closure. Snapshot bytes/modes/entries for a refused
compound edit; validate exact untouched data separately from decoded values.
Exercise missing/invalid statuses as well as valid in-progress work. Run at least
one terminal repair and one metadata-only edit with an unrelated pre-existing defect.

Codex must independently reproduce both race directions: a settled source gains an
unsettled status before a bucket-write CAS; an open source is closed/deferred before
a finding-write CAS. Temporary sandbox-only instrumentation may prove the first
attempt conflicts and the next callback sees fresh state. Require one compiler-valid
mutation of the explicit-status guard in `EditFinding`, killed by
`TestFindingEditsCannotReactivateNonOpenAudit`, with restored green evidence.

Antigravity must independently compare explicit expected terminal statuses against
readiness, moves, and BOTH `audit lint` and normal `lint`, rather than deriving its
oracle from `TerminalFindingStatus`. Require one compiler-valid mutation reverting
the non-open `LintFindings` bucket check to literal-open-only, killed by
`TestAuditParsedSettlementCLIRefusalsAndLint`, with restored green evidence. For
cross-checking races, explain the actual guarded source/hash and fresh-callback seam;
running a test name without confirming what interleaves is insufficient.

Both: execute the emitted findings/lint advice with a slug that shadows a sibling ID
and a frontmatter ID that disagrees. Verify selected source, status, stdout/stderr,
exit 11, typed refusal receipt/dry_run, and current schema revision. A missing field
or boolean false alone is not proof of correct mapping: compare nonempty terminal
readiness with active and invalid-status cases against exact-revision schemas.

Mutation evidence must name the exact edit, intended assertion, actual failure,
compiled/killed/survived/invalid classification, and restored command/result. A mutation
of a nearby producer or a compile error is not proof of the named contract.

## Required hostile angles

Explicitly settle these hypotheses as reproduced, falsified, or unresolved:

- A no-op/dry-run shortcut or a decorated/case-normalized token bypasses refusal.
- An independent core caller or pathless port loses the bucket policy implemented
  only by the CLI; a separate metadata read hides a snapshot race.
- An invalid/missing status contributes zero to a tally and accidentally counts as
  settled, or readiness/lint disagrees with the move decision. Distinguish real adapter
  output from manufactured inconsistent count structs (defense-in-depth if unreachable).
- Rechecking an entire damaged audit makes terminal correction/metadata repair impossible,
  or applying part of a compound edit leaves false status/note/candidate evidence.
- Tests use a self-confirming shared oracle or synthetic fixtures that bypass the
  real producer/selector/guard. Identify one such possible blind spot, then probe it.

Take a separate systemic second pass: inspect shared normalization and disposition
helpers, status-source parsing, early returns, retry captures, indirect callers, and
test utilities. Consolidate by root cause. The strongest review is not the longest
report or highest finding count. A preference is not a defect; an unexecuted idea is
uncertainty, not evidence. Do not hide an unresolved high-risk hypothesis behind ready.

## Validation and restoration

Owner evidence in the task: full race suite, lint (zero issues), build, tidy,
planning/audit lint; separate docgen output matched committed-path docs. Goldens
were regenerated under 1.84. Literal-open move and disabled finding-guard mutations
were compiler-valid, killed by the intended real-store tests, and restored green.
Do not present inherited owner checks as your own independent executions.

Run focused domain/core/store/CLI/TUI/wire checks relevant to your probes; a second
complete race suite is optional if time permits. Set caches/scratch outside tracked
and untracked sandbox contents. Do not regenerate shared source artifacts. Generated
CLI docs differ intentionally from base until committed: compare fresh temporary
docgen output with sandbox `docs/cli`, rather than treating `git diff`'s planned
changes as generator drift.

Use the minter's mandatory independent clone protocol. Restore all mutations to its
captured baseline; leave only your assigned audit changed. Preserve the brief, verify,
then perform the guarded one-file transfer. No staging, commits except the helper's
baseline, pushes, source-checkout mutations, or sibling report consultation. If any
required probe/isolation/transfer fails, retain evidence and report incomplete.

## Deliverable

Fill your Reviewer report with: verdict/limits; isolation attestation (sandbox,
independent Git directory, baseline, captured-source fingerprint, verification and
transfer result); verified consumer inventory; findings; concise hostile/mutation
ledger; bounded surviving hypotheses. Leave genuine findings open using the tool's
canonical grammar. Include root cause, reachable producer/caller, consequence,
reproduction, and recommendation. Label coverage gaps, preferences, and hypothetical
hardening separately. Do not invent finding placeholders or settle your own findings.

## Reviewer report

### Verdict and limits

**Changes requested: one open medium finding (M1).** Both review passes and the
assigned bounded evidence floor are complete. The lifecycle and explicit selected
status guards withstand the tested no-op, preview, normalization, and CAS attacks.
The systemic second pass demonstrated a separate structured-note producer that
can introduce new active work into a closed/deferred audit through those same
finding-edit operations.

M1 is a remaining in-scope structured-writer gap, not a newly introduced renderer
regression: the unsafe note wrapper exists in the verified base too. It is distinct
from the expressly deferred arbitrary append/whole-file editing routes. No code
was fixed and no finding was settled. The sibling reviewer report was not read.

### Isolation attestation

- Source: `/Users/andyeschbacher/git/andy-esch/taskflow`; captured branch
  `fix/audit-parsed-settlement`.
- Independent workspace:
  `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bRkpVV`.
- Resolved Git directory:
  `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bRkpVV/.git`.
- Sandbox-only baseline: `0118b404fc85ef9b978c00c8495ab90eeca7854e`.
- Verified base: `3b70c1909d005405127e162b5fb33ce9bf18634d`,
  whose sandbox log identifies merged PR #288.
- Captured source audit blob: `ff714746583973917a4707a93ca71e63f441cfe9`.
- Captured source fingerprint: `400c5958f5edc8fe911b36c60042f8a83cb52654`.
- Deliverable:
  `planning/audits/6gj5ct38xq4t-2026-10-09-audit-parsed-settlement-implementation-codex.md`.
- Creation: the prescribed helper succeeded, overlaying the dirty handoff into its
  independent clone. All implementation inspection, builds, tests, mutations,
  fixtures, generated outputs, and report edits occurred in that workspace or
  its external scratch directory.
- Verification: the prescribed helper succeeded with only the assigned audit
  changed, unchanged baseline HEAD, no staging, independent Git metadata, and
  matching captured source-deliverable blob.
- Guarded transfer result: `succeeded` upon delivery of this prepared report. The
  authoritative helper receipt is retained at
  `/private/tmp/taskflow-codex-audit.fv5Kj0/transfer.log`; a refused transfer instead
  makes this prepared delivery attestation ineffective and the review incomplete.
  Workspace and evidence are retained for owner receipt.
- Protocol deviation: before loading this brief, initial read-only discovery
  included source `pwd`, file-name enumeration, `git status`, root, and branch.
  No implementation content, sibling report content, tests, generators, or writes
  occurred there. Once the brief was loaded, source access was confined to the
  helper's copy/verification/transfer. This narrow discovery deviation is disclosed
  rather than claiming that every source operation was limited to the two allowed
  operations.

Evidence directory: `/private/tmp/taskflow-codex-audit.fv5Kj0`. Temporary Go
probes have been copied there and removed from the sandbox. The helper baseline
is the only reviewer commit; no subsequent staging, commits, branch switches,
pushes, or manual source copies occurred.

### Verified consumer inventory

All paths/lines below refer to the captured sandbox, not planning-only promises.

| Consumer | Verified implementation/evidence |
| --- | --- |
| Domain move/read policy | `internal/domain/audit.go:94` (`UnsettledFindings`), `:112` (`Settled`), `:120` (`ReadyToClose`), `:129` (`ValidateMove`), `:163` (`AuditUnsettledFindingsError`). Empty audit exception and unconditional readable reopen precede unsettled checks. |
| Parsed statuses | `internal/domain/finding.go:147` (`ParseFindings`), `:402` (`LintFindings`), `:464` (`TallyFindings`), `:478` (`TerminalFindingStatus`), `:483` (shared disposition mapping), `:619` (`SetFindingStatus`). Status source authority is `internal/domain/finding_status.go:15`; leading decoration stripping is `internal/domain/resolution.go:116`. |
| Structured edits/create | `internal/core/finding.go:188` routes `SetFindingStatus` through `EditFinding`; `:204` is `FindingEdit.apply`; `:248` enters the guarded callback; `:254` applies the non-open explicit-status gate. Existing `NewFinding` gate is `internal/core/finding_create.go:47`. |
| Persistence port/error | `internal/core/store.go:305` requires guarded move validation and `:327` requires matching callback metadata/body. `internal/core/audit_move_error.go:13` retains source/cause; `:21` adds findings/lint advice using `Source.ID`. `internal/core/service_audit.go:122` retries the complete move; `internal/core/retry.go:67` retries only conflict errors. |
| Actual guarded source | `internal/store/auditstore.go:128` parses the bytes read at `:122`; policy at `:135` precedes no-op `:139` and dry-run `:152`; CAS at `:166` hashes those same bytes. `parseAuditWithFindings` at `:226` derives total from parsed slice length and bands from the tally at `:257`. |
| Finding CAS and lock | `internal/store/body.go:209` parses metadata from the bytes read at `:205`; callback `:217` precedes no-op `:222`; write CAS `:233` uses those bytes. `writeBody` at `:63–79` runs the interleave hook, locks, rechecks, then atomically writes. `internal/store/cas.go:26` hashes exact bytes, `:138` selects filename identity, `:174` compares the fresh hash. `internal/store/lock_unix.go:22` opens the planning root directory for flock: there is no invented per-audit lock file. |
| Source-backed reads/lint | `internal/store/entity_read.go:26` establishes filename `Source.ID`; `internal/store/lintsource.go:26` loads the selected audit snapshot. `internal/core/finding.go:280` (`AuditLintIssues`) is shared by audit lint `:372` and ordinary lint `internal/core/service.go:928`. |
| CLI | `internal/cli/audit.go:244` routes finding edits through core at `:266`; move commands call core at `:608`; lint help states the new invariant at `:432`. Fresh-binary stdout/stderr, exit codes, advice and refusal receipts were exercised directly. |
| TUI/readiness | `internal/tui/entity.go:399` calls core for audit transitions; `internal/tui/item.go:451`, `internal/tui/detail.go:1718`, and CLI `internal/cli/render/render.go:712` use readiness. `internal/core/service.go:596` aggregates the same settlement policy. `TestPortableAuditViewsQualifyIncompleteEvidence` at `internal/tui/audit_unparsed_test.go:30` includes active and missing/invalid projections. |
| Wire/artifacts | `internal/wire/dto.go:228` is optional `ready_to_close`, mapped at `:284`; `internal/wire/wire.go:355–360` declares 1.84 **NOT ADDITIVE**. Captured fresh schema has exact revision `/json-envelopes/1.84`; machine revision fixture is `internal/cli/testdata/golden/machine_contract_revision.txt:1`. Generated schema/goldens and `schema_comments.json` are actual captured files. Fresh docgen output exactly matched sandbox `docs/cli` (`docgen-diff.log` is zero bytes). No new JSON data field is claimed. |

Verified regression entry points, all exercised by the focused package checks:

- `internal/domain/audit_settlement_policy_test.go:11`:
  `TestParsedSettlementPolicyAgreesAcrossReadinessMovesAndLint`.
- `internal/core/finding_settlement_test.go:14`:
  `TestEditFindingPortableBucketPolicy`.
- `internal/store/audit_settlement_policy_test.go:17,48,82,114,150`:
  `TestAuditMovesRequireTerminalParsedFindingsBeforeEffects`,
  `TestAuditMoveRetryRevalidatesUnsettledParsedStatuses`,
  `TestFindingEditsCannotReactivateNonOpenAudit`,
  `TestFindingEditsAllowTerminalRepairAndMetadataWithoutCertifyingWholeAudit`,
  `TestFindingEditRetryObservesConcurrentNonOpenBucket`.
- `internal/cli/audit_settlement_policy_test.go:16,56,90`:
  `TestAuditParsedSettlementCLIRefusalsAndLint`,
  `TestAuditUnsettledAdviceExecutesAgainstSourceNotDeclarationOrSlug`,
  `TestAuditFindingCLIRequiresExplicitReopenToReactivate`.

### Findings

#### M1. Wrapped resolution notes can introduce active findings in non-open audits · **Status:** fixed

**File:** `internal/core/finding.go:254`; producer:
`internal/domain/finding.go:555,671` and `internal/domain/body.go:343`.

**Root cause and reachable caller.** `FindingEdit.apply` invokes
`SetFindingNote` at `internal/core/finding.go:213`. That writer rejects input
newlines, then calls `wrapNote`, whose continuation indent is empty.
`wrapProse` inserts newlines before overflowing words, including Markdown heading
tokens. A single-line note can therefore produce a canonical finding header.
The new explicit-status guard reparses the result but only checks the first
selected code; metadata-only edits skip it. The real store's write tail validates
file/fence readability, not whether a note introduced findings. This route is
reachable through ordinary `audit finding --note`, without arbitrary body flags.

**Demonstrated consequence.** Starting with one fixed H1 in a closed audit,
`audit finding <source-ID> H1 --status fixed --note <single-line-note> --json`
exits 0, reports no stderr, and persists a second H2 with
`Status: in-progress`. The resulting receipt has
`bucket: closed`, `findings: 2`, `in_progress_findings: 1`, and
`done_findings: 1`. The note itself is truncated in parsed output because its
continuation became a separate finding. This bypasses the open-only creation gate
and introduces a new defect into a previously settled audit; it is not permission
to repair an unrelated pre-existing defect. Subsequent audit lint and same-bucket
close correctly fail with exit 11, but the bad write already succeeded.

**Reproduction.** Fresh binary `tskflwctl`, disposable unregistered space
`/private/tmp/taskflow-codex-audit.fv5Kj0/space`; no fixture rewrite helper:

1. Create with body `#### H1. Settled · **Status:** fixed\n`, then close.
2. Supply this single-line note (64 consecutive `x` characters, one space, then
   the heading):
   `xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx #### H2. Injected · **Status:** in-progress`.
3. Run `audit finding <source-ID> H1 --status fixed --note "$NOTE" --json`.
4. Inspect `audit info`, `audit findings`, and `audit lint`.

Actual generated body:

```markdown
#### H1. Settled · **Status:** fixed

**Resolution:** xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
#### H2. Injected · **Status:** in-progress
```

`note-probe.py` and `note-probe.log` retain the direct binary reproduction
(source ID `6gj5fethm8mr`). Independent real-store
`TestReviewerNoteCannotCreateFindingInNonOpenAudit` compiled and failed all
eight combinations: closed/deferred × metadata-only/explicit fixed × write/preview.
Every result had two findings and one active finding; writes persisted H2, previews
left exact source bytes untouched while predicting the same invalid state.
Evidence is `note-invariant.log`; the temporary test is retained externally.

**Scope and recommendation.** The base's note wrapper already has this behavior;
this is a surviving structured-write hole, not a claim that the change introduced
that wrapper or that the documented append gap is new. Make generated note
continuations remain prose and verify that note edits preserve finding/header
structure. Add real-store/CLI regressions for the eight demonstrated combinations.
Use a baseline-aware structural check or safe rendering, preserving terminal and
metadata repairs when unrelated existing findings remain damaged; do not solve it
by globally requiring a lint-clean audit.

**Resolution:** Indented generated note continuations and kept fence tokens
    inline; domain, real-store and CLI regressions preserve finding structure in
    previews and writes while allowing unrelated repairs. Reverting the fix
    compiles and fails the regressions.

### First-pass execution and hostile ledger

- **Fresh binary/real producer — executed.** Built
  `go build -o /private/tmp/taskflow-codex-audit.fv5Kj0/tskflwctl ./cmd/tskflwctl`.
  `workflow.py` initializes its space with `init --no-register` and uses
  `audit new`, rather than only test fixture injection. It exercises active
  refusal, terminal closure, both non-open buckets, normalized decorated active
  input, explicit reopen, empty closure, missing/invalid statuses, terminal
  repair with unrelated active work, and metadata/candidate edits. Additional
  `repair-probe.py` confirms a metadata-only H1 edit succeeds while an unrelated
  M1 remains invalid in a deferred audit.
- **No-op/preview/token bypass — falsified for selected-status edits.** The real
  shipped tests exercise unchanged active/open status as well as transitions.
  Binary compound edits refuse in both buckets and both preview/write modes.
  Refused edits compare raw whole-tree bytes, stat modes, directory entries,
  empty-directory and symlink sentinels independently of decoded JSON. No note,
  candidate, timestamp, or partial status effects landed.
- **Missing/invalid status accidentally settled — falsified for actual adapter
  output.** Real audit bodies with empty or `opne` status retain total=1 while
  all bands are zero, readiness is absent/false, close/defer refuse, and both audit
  lint and ordinary lint report unsettled work in a non-open bucket. Explicit
  independent expected terminal list is fixed/tracked/deferred/superseded/wontfix.
  Nonempty terminal cases advertise readiness; literal-open counts/denominators
  remain unchanged. Manufactured inconsistent count structs are a separate
  defense-in-depth question, not demonstrated filesystem output.
- **Portable caller loses bucket policy — falsified.** The actual pathless
  `findingCreationStore` callback at
  `internal/core/finding_create_test.go:21` supplies body and bucket; the transform
  supplies the usable body/bucket, and `EditFinding` performs no separate read
  that could grant cached authority. Its bucket matrix
  runs through `SetFindingStatus`/`EditFinding` for preview/write.
- **Whole-audit revalidation locks out repair — falsified in the tested cases.**
  Terminal correction preserves residual active work; metadata-only correction
  preserves unrelated invalid work. Same-bucket moves still refuse the residual
  defect. Tracked-token terminal policy remains independent of destination lint,
  pinned by `TestParsedSettlementDoesNotSubsumeMetadataLint` at
  `internal/domain/audit_settlement_policy_test.go:58`.
- **Source advice selects declaration/display slug — falsified.** Raw files use
  source `6g0000000011`, slug `6g0000000012`, declared ID
  `6g0000000012`, plus an empty sibling whose filename ID is
  `6g0000000012`. For active, invalid and missing statuses, close/defer with both
  dry_run values emit findings/lint commands using the source ID. Those exact
  commands are extracted and executed: findings returns the original H1/status;
  lint returns declaration drift and exit 11. Move stdout carries its refusal
  receipt/dry_run; stderr carries a typed JSON validation error, exit 11.
  Findings exposes its existing slug field, not an invented `audit_id` wire field.
- **Unparsed evidence/adjacent authoring boundary — checked.**
  `boundary-probes.json` reproduces near-miss close refusal with total=0 and
  unparsed=1, while reopen remains available. It separately reproduces the known
  closed-empty append route and new lint diagnosis. That route is not M1.

### Independent CAS evidence

`reviewer_parsed_settlement_probe_test.go` and `independent-cas.log` retain
the temporary raw-file probe, independent of `AuditFixture`.

- **Settled source gains unsettled status before bucket CAS — falsified as a
  bypass.** For close and defer, interleave `in-progress`, missing, or `opne`
  into exact source bytes at `testHookBeforeMoveAuditWrite`. An observing port
  records attempt 1 as `ErrConflict`, attempt 2 as typed
  `AuditUnsettledFindingsError`. Exact concurrent bytes survive.
- **Open source closes/defers before finding CAS — falsified as a bypass.**
  At `testHookBeforeBodyWrite`, the real `FS.MoveAudit` closes/defers the
  settled source. Observed transform 1 sees open/fixed, then conflicts; transform
  2 sees fresh closed/deferred metadata with the same fixed body and refuses
  decorated in-progress. Both attempts are counted, exactly one conflicts, no
  edited bytes land, and concurrent bytes survive.
- Both races passed with `go test -race ./internal/store -run
  '^TestReviewerBothCASDirections$' -count=1 -v`. This demonstrates the actual
  recheck/fresh-callback seam, not merely a named-test invocation.

### Separate systemic pass and mutation ledger

Re-inspected shared disposition/normalization, status-source authority, note and
candidate producers, early returns, retry captures, indirect core callers, port
promises, and helper-generated fixtures. The parser/selector/CAS hypotheses above
remain falsified within their stated scope. The structural-note hypothesis is
**reproduced as M1**, consolidated by root cause.

| Probe | Exact edit/assertion | Actual result and classification | Restoration |
| --- | --- | --- | --- |
| Required explicit-status guard mutation | In `internal/core/finding.go:254`, replace predicate with `if false && edit.Status != "" && audit.Bucket != domain.AuditOpen`. Intended assertion: `TestFindingEditsCannotReactivateNonOpenAudit` must return validation refusal and no change for compound selected-status edits. | `go test ./internal/store -run '^TestFindingEditsCannotReactivateNonOpenAudit$' -count=1` compiles, exits 1; first failure at `audit_settlement_policy_test.go:97`: `changed=true err=<nil>`. **Compiled/killed**, not a compile failure or nearby-producer mutation. `mutation-status-guard.log`. | Restored file from sandbox HEAD. `go test -race ./internal/store -run '^(TestFindingEditsCannotReactivateNonOpenAudit\|TestReviewerBothCASDirections)$' -count=1` passes; `restored-status-guard.log`. |
| Shared-oracle blind spot | In `tallyFindingStatus`, change in-progress return from `FindingTally{Active: 1}` to `FindingTally{Done: 1}`. This changes the common policy used by production and the portable test's expected result together. | `TestEditFindingPortableBucketPolicy` **compiled/survived**, confirming that test alone is self-confirming. `TestParsedSettlementPolicyAgreesAcrossReadinessMovesAndLint` **compiled/killed** it at line 21 with `terminal vocabulary drift`, for lower/upper case active tokens. Logs: `mutation-shared-portable.log`, `mutation-shared-oracle.log`. | Restored `internal/domain/finding.go` from sandbox HEAD; both targeted domain/core tests pass in `restored-shared-oracle.log`. |
| Structured-note invariant | No production mutation. Independently assert that a note edit of the raw one-finding non-open source cannot return two findings/one active. | **Compiled/reproduced defect**, eight assertion failures on unmodified captured implementation. This is evidence for M1, not a killed mutation. `note-invariant.log`. | Temporary failing test removed; copy preserved externally. No implementation repair. |

The shared-oracle concern is bounded by an actually killed independent domain
oracle and the separate real-producer CLI matrix, not dismissed because the
portable helper passed. For the required guard mutation there was no nearby
alternative gate accidentally preserving the invariant.

### Wire, checks, and surviving limits

- Temporary `TestReviewerBinarySettlementEnvelopes` compiled schemas from both
  sandbox generator and fresh binary and required them to match. It validated
  **162 captured binary streams** against their exact 1.84 envelope definitions,
  then required all to reject revision 1.83. Seven outputs exercised
  `ready_to_close: true`; independent workflow assertions compared those
  nonempty terminal cases against active, invalid, missing, and empty cases.
  Evidence: `binary-records.json`, `binary-schema.log`, external copy of
  `reviewer_parsed_settlement_schema_test.go`. This is semantic/non-default
  coverage, not just checking that an optional field is absent.
- Independently ran `go test -race ./internal/domain ./internal/core
  ./internal/store ./internal/cli ./internal/tui ./internal/wire`: all passed.
  `focused-race.log` records the initial runs; `final-focused-race.log` records
  checks after restoration/removal (unchanged packages reused Go test cache).
  Fresh docgen comparison passed. Build emitted a nonfatal denied module
  stat-cache write outside scratch, but compilation succeeded.
- Did not run a second repository-wide race suite, static lint, or module tidy;
  inherited owner evidence for those is not claimed as independent work.
- The shipped wire schema validation fixtures at
  `internal/wire/envelopes_test.go:585–591` primarily cover default readiness.
  The separate actual-binary validator supplied the missing true branch for this
  review; this is a coverage limitation, not an additional production finding.
- A manufactured audit count object with terminal bands greater than total can
  produce negative unsettled count and permissive move validation. Actual
  `parseAuditWithFindings` cannot produce that state; no reachable adapter
  defect was demonstrated. Treat consistency validation as hypothetical hardening.
- Advisory flock covers cooperating writers; the review does not claim protection
  against a raw noncooperating write after the final hash check. No new regression
  in that established boundary was demonstrated.
- No other unresolved high-risk hypothesis within the bounded brief remains.
  M1 stays open; documented arbitrary-body guards remain the separate task.

## Owner triage and verification (2026-10-09)

M1 accepted as an in-scope structured-write defect, despite the wrapper predating
this task. `wrapNote` now paragraph-indents its continuations; `wrapProse` also
keeps fence tokens inline because the structural scanner recognizes indented
fences. No global lint-clean precondition was added: note and terminal repairs
remain available when unrelated findings are damaged.

Permanent regressions cover all eight reported CLI combinations and sixteen
real-store combinations (also with unrelated active work), previews with exact
tree preservation, full note round-trip, neighboring prose/candidate sections,
and unknown frontmatter. Domain probes vary wrap boundaries, Unicode, canonical
and near-miss headers, section/status/resolution labels, and both fence forms;
replacement and removal preserve the original structure.

Reverting the continuation indent compiled and failed the intended domain,
store, and CLI regressions. After restoration they passed. The portable core
test now uses an independent expected-status oracle, and a new wire regression
validates actual emitted `AuditShowEnvelope` instances against their exact schema
with five nondefault terminal readiness cases plus active, missing, invalid,
empty, unparsed, and non-open cases. Mutating the shared in-progress mapping to
the done band compiled and was killed by both the portable core and wire tests;
both passed after restoration.

The reviewer report and its original verdict remain historical evidence, not a
claim that the reviewer inspected these fixes. Its independent CAS/binary/schema
probes are useful additional evidence. Inconsistent manufactured count objects
remain hypothetical hardening, not a demonstrated adapter bug. The pre-existing
finding-edit preview body/slug-reread issue is grouped with the existing
[body-write and receipt followup](../tasks/6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md),
which remains sequenced after this task.

Full race tests, static lint (zero issues), build, and module tidy checks passed.
A fresh binary in an unregistered disposable space retained exactly one fixed
finding and the entire note for M1's input; preview left the source untouched.
