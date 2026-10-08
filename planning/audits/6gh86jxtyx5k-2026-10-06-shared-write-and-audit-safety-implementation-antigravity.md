---
schema: 1
id: 6gh86jxtyx5k
bucket: closed
area: shared-write-and-audit-safety-implementation-antigravity
date: "2026-10-06"
updated_at: "2026-10-07"
---
# Audit: Shared writes, bootstrap targeting, and audit evidence safety — antigravity — 2026-10-06

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

Adversarially review the uncommitted safety batch: YAML block-scalar preservation,
explicit bootstrap targeting, and evidence-based audit-header recognition/readouts.
The implementation owner's gates are green; establish whether the tests and boundaries
actually protect user intent. Do not implement fixes or read the sibling review.

Use complementary primary lenses, not different definitions of correctness:

| Reviewer | Primary lens | Required cross-check |
|---|---|---|
| Codex | Init selection and audit parsing/repair/read/machine contracts | One real store write checking decoded and byte YAML fidelity |
| Antigravity | YAML source slicing, value preservation, comments, refusal-before-write | One real CLI init selector probe and one audit repair/projection probe |

Both own failures at seams between their primary lens and shared writers. A lens
split is not permission to skip a concrete adjacent consequence. For Antigravity in
particular, prioritize the falsifiable hypotheses below over a generic architectural
essay. Failing to break a hypothesis is useful evidence; a finding quota is not.

## Review target

Source: `/Users/andyeschbacher/git/andy-esch/taskflow`, branch `main`, base commit
`af259c017374bed3333a61b0e38bb60e49c2f39e`. The target is the captured staged,
unstaged, and untracked working state, not that commit alone. Verify the helper's
captured diff and inventory before drawing conclusions. Most new implementations
and regressions are untracked at handoff; a plain `git diff HEAD` will miss them.

Implementing tasks:

- [YAML values and formatting](../tasks/6g1dhhk6721x-a-surgical-frontmatter-write-re-folds-multi-line-block-scalars-onto-one-line.md).
- [Init target selectors](../tasks/6ggjmtmdd54w-prevent-silently-ignored-target-selectors-during-init.md).
- [Shared finding-header classifier](../tasks/6g77rn6b9wf8-narrow-the-near-miss-finding-recognizer-and-re-measure-across-every-entity-type.md).
- [Unparsed audit read evidence](../tasks/6g77rn6em6n8-report-unparsed-findings-on-the-audit-read-surfaces.md).

Verified starting points (trace callers rather than stopping at these files):

- `internal/store/frontmatter_preserve.go`: `assembleEditedFile` and parsed entry
  boundaries; `frontmatter.go`: field, dependency-source, and body encoders.
- `internal/store/frontmatter_fidelity_test.go`: byte/decoded regression matrix and
  real-filesystem repeated field writes/refusal.
- `internal/cli/init.go`: `initDirectory` and bootstrap's style-only pre-run;
  `init_selection_test.go`, `init_registration_test.go`, and real-process smoke tests
  in `cmd/tskflwctl/main_test.go`.
- `internal/domain/finding_header.go`, `finding.go`, `finding_create.go`;
  `finding_header_test.go`, `nearmiss_test.go`, and `nearmiss_corpus_test.go`.
- `internal/store/auditstore.go` and `nearmiss_write_test.go`;
  `internal/core/finding.go` repair/lint consumers;
  `internal/core/audit_unparsed_test.go` portable read evidence.
- `internal/cli/audit_unparsed_test.go`, `lint_fix_findings_test.go`,
  `internal/cli/render/columns.go` and `columns_contract_test.go`;
  `internal/wire/dto.go`, `wire.go`, schema comments and machine-contract goldens.
- `docs/ARCHITECTURE.md`, `CLAUDE.md`, generated init/audit command references,
  and the corrected finding grammar in `scripts/prepare-adversarial-review-audits.sh`.

Build a short, verified consumer inventory: all existing-document encoder callers;
near-miss lint, repair, generic body-write and typed finding-creation consumers;
audit read/tally/DTO/projection consumers; init target selection, topology, registry,
authorization, and emitted receipts. Cite real paths/lines; distinguish private parser
fixtures from reachable production input.

Read [ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
for revision semantics and the architecture planning-data checklist for port boundaries.

## Intended contract to challenge

1. Unrelated writes preserve both decoded YAML values and exact untouched block-scalar
   source. Body edits and graph repair share the same safeguard. Normal fresh creation
   and block-scalar-free files retain their ordinary encoder. Unsafe preservation
   refuses before durable write rather than silently changing intent.
2. Init honors one explicit --path, -C, or --space target. Explicit --path/-C overrides
   TSKFLW_SPACE. Conflicts, blank targets, and invalid registered selection fail before
   writes; no fallback to cwd. Fresh targets need no existing planning discovery.
   Existing repair, pointer-relative paths, dry-run, no-register, and authorization work.
3. Recognition is not repair permission. Canonical, repairable, ambiguous, and ordinary
   classes share one classifier. Auto-repair requires local authoritative Status
   evidence plus the documented finding context. Generic body writes refuse newly
   introduced repairable drift, not ambiguous prose or existing defects. Typed finding
   creation must not allocate a colliding code or give an impossible repair instruction.
4. Unparsed headers remain separate from parsed findings/statuses/percent denominators.
   List/show qualify incomplete reads and name lint; ordinary reads stay informational.
   Portable consumers carry the same evidence without reading local paths. New optional
   projections do not expand default tables/CSV or invent a zero-valued wire field.
5. Machine revision 1.82 is NOT ADDITIVE: selector behavior is deliberately tightened
   and incomplete audit readiness withheld. New fields, aliases, non-default schema
   branches, classified changelog, and generated fixtures must agree.

Known deferred ownership, not hidden completion claims:

- [Audit TUI/lifecycle consumers](../tasks/6gh82rm9sf3b-carry-unparsed-finding-evidence-into-audit-tui-and-lifecycle-decisions.md):
  TUI metadata still lacks the unparsed warning; closure/deferral still gates only
  on parsed open findings. Challenge an in-scope consequence if one exists, but do
  not report these already recorded gaps as newly discovered batch regressions.
- [Creation/edit baseline guard gaps](../tasks/6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md).
- [Remaining behavior pins/body-transform extraction](../tasks/6g77rn6n6b86-pin-the-unpinned-audit-write-and-lint-behaviours-and-unify-the-body-transform.md):
  only declaration-comment ownership AC 7 is settled here; no shared-transform
  extraction or exhaustive operational pinning is claimed.
- Automatic reviewer launching and review analytics in the sibling skills repository
  are not implemented in this batch. The new preparation skill is separate authoring
  work, not a substitute for reviewing the actual code.

## Mandatory evidence floor

- Execute at least three hostile production-path probes in your primary lens and
  the cross-checks above. Include valid-but-hostile and refused/no-write inputs.
  Record input, exact command/test, before/after values/bytes or target snapshots,
  and consequence. Green unit tests alone are not a no-findings verdict.
- Execute at least two compiler-valid, one-at-a-time mutation probes selected from
  the following invariants. Name the exact test expected to fail, its actual failure
  or survival, and a restored green run. A compile error is invalid, not a kill.
  Antigravity must include preservation/refusal; Codex must include target selection
  or repair authority. Do not impose a mutation quota on every new test.
  - Bypass source-preserving assembly in one writer that claims coverage.
  - Disable decoded-value comparison; use a fixture that actually differentiates it
    from parser validity or the broken-anchor syntax check.
  - Ignore -C in init, or let conflicting selectors proceed to bootstrap writes.
  - Apply canonical spelling to ambiguous headings or omit repairability from the
    introduced-header identity check.
  - Drop optional-column policy while lifting loaded records, or drop the non-default
    UnparsedFindings DTO mapping/readiness condition.
- Give constructed fixtures a producer trace. Hand-authored YAML is supported input;
  fabricated internal node graphs require a reachable encoder/parser path or a
  defense-in-depth label. Do not call every unsupported YAML layout a high defect:
  assess the failure's safety, reachability, and operator usability.
- Exercise at least one non-default optional JSON/projection branch and validate it
  against the exact generated schema/semantic contract, not just successful decoding.
  Execute actionable lint/repair advice against the same fixture that emits it.

## Required hostile angles

### YAML preservation hypotheses

- Folded/literal scalars with blank and more-indented lines, keep/strip/default
  chomping, explicit indent indicators, root/nested indentation, LF and CRLF, repeated
  unrelated writes. Valid YAML after encoding can still decode to the wrong value.
- First/middle/last keys, deleting the first key then changing the next, head/inline/foot
  comments, quoted keys, blank-line ownership around keep-chomp scalars, nested blocks.
  Check exact untouched slices separately from decoded equality and body bytes.
- Anchors/aliases/merges, changing an anchored value, deleting a referenced entry,
  replacing the last block scalar, and adding new fields. Does a refused edit leave
  the actual file unchanged? Does a harmless edit become unnecessarily uneditable?
- Whole-entry reuse versus changed nested values, typed YAML tags, flow/block mappings,
  and unusual-but-valid parsed key layouts. Challenge whether source boundaries and
  semantic equality actually justify the safeguard's claims without requiring a new
  generic YAML editor. Trace lifecycle/body/dependency repair, not only direct helpers.

### Init selection hypotheses

- Separate real caller cwd, selected directory, pointer planning target, registry home,
  and environment-selected entry. Existing helpers used to inject -C; prove your
  fixtures cannot mask the bug. Test relative and absent targets, paths with spaces,
  all explicit conflicts (including identical paths), blanks, stale/unknown registry
  entries, ambient defaults, dry-run, no-register, and denied mutation policy.
- Verify refused and accepted paths through real bootstrap callers and snapshots.
  Check pointer/link-back and repair relative to the chosen directory; distinguish
  intentional topology writes from accidental caller/registry writes. Inspect real
  process stdout/stderr, JSON envelope, and exit classification for one refusal.

### Audit recognition/read hypotheses

- Ordinary S3/V2/MP3/Top3 and word-number headings in/outside Findings; valid canonical
  findings; long/lowercase/bold/hyphen/underscore drift; matching code references and
  duplicate definitions; empty/missing Status; Status borrowed from a later heading,
  fenced/blockquoted/inline-code examples; nested sections and mixed line endings.
- Change evidence without changing heading text, insert before existing occurrences,
  add a duplicate occurrence, and mix repairable with ambiguous candidates. Verify
  generic append/edit and typed finding creation independently. A diagnostic-only
  ambiguity must not be silently rewritten or offered a repair that cannot help.
- Run real lint --fix: no phantom findings or follow-on missing-status errors on prose,
  genuine evidence-backed repair, no-op second pass, and leftover ambiguity reported.
  Distinguish live corpus cleanliness from adversarial branch protection. Number-only
  codes and seven-hash headings are intentionally outside this grammar.
- Empty, wholly unparsed, mixed, and settled-but-incomplete read surfaces. Verify counts,
  readiness, parsed denominators, optional omission at zero, aliases, default table/CSV,
  and portable loaded-record evidence. Inspect repeated derivation/drift possibilities
  between snapshots, counts, and the body used by guarded repair.

### Separate systemic pass

After the focused ledger, make a second pass for a shared helper or test convention
that can hide an entire defect class. Consider source slicing/value aliasing, redundant
classification, broader encoder callers, snapshot-to-action drift, authorization before
target side effects, and serializer policy lost through adapter decoration. Settle the
most plausible hypothesis with evidence; do not pad the report with speculative redesign.

## Validation and restoration

The owner ran `just test` (race), `just lint`, `just build`, `just tidy-check`, generated
CLI/schema-comment comparisons in disposable output paths, `lint --links`, `audit lint`,
and `git diff --check` successfully on the implementation handoff. This is inherited
baseline evidence, not independent execution by you.

Both reviewers run focused tests matching their probes and restored mutations. Codex
also independently runs `just test`; Antigravity can reuse that owner baseline for the
full gate but must still execute focused hostile evidence. Run the full gate if a
cross-cutting finding or mutation makes the inherited result irrelevant.

Useful existing focused entry points (verify names in the sandbox):

```sh
go test ./internal/store -run 'TestFrontmatterEdits|TestBlockScalar|TestAuditBodyWrites|TestAppendAuditBody|TestTransformAuditBody' -count=1
go test ./internal/domain -run 'TestFindingHeader|TestNearMissFindingHeaders' -count=1
go test ./internal/cli -run 'TestInit|TestAuditReadSurfaces|TestLintFix' -count=1
go test ./internal/core -run TestPortableAuditReads -count=1
go test ./internal/cli/render -run 'TestLoadedRecordColumns|TestColumnRegistriesMatchFullWireValues' -count=1
go test ./cmd/tskflwctl -run TestSmoke -count=1
```

Build a fresh sandbox binary for CLI probes. Put probe spaces, registries, cached data,
and generated comparison outputs in reviewer-owned temporary directories outside the
checkout; use normal repository commands only inside the independent sandbox. Do not
install over the user's binary, mutate their registry, publish, or run live planning
mutations in the source checkout. Never use a generator's exit status alone as a drift
comparison. Restore probes to the sandbox helper's captured baseline, not source HEAD;
only your assigned audit may differ at guarded transfer. Keep the sandbox until receipt.

## Deliverable

Keep this brief intact and replace only the report placeholder. Aim for a concise
findings-first report, not a retelling of every file. Include:

1. Verdict (ready / needs changes / incomplete), captured baseline, and evidence limits.
2. Required sandbox and guarded-transfer attestation.
3. Verified consumer inventory and a short hostile-probe/mutation ledger, including
   at most five important hypotheses that survived or remain unresolved.
4. Actionable findings consolidated by root cause, with exact code location, reachable
   production consequence, minimal reproduction, result, and recommended correction.

Leave every finding open. Use canonical headings/statuses (prefer the local finding
creation verb). High requires reachable corruption, wrong-target mutation, safety
failure, or materially misleading semantics; bounded product defects are Medium.
Coverage gaps are normally Low unless a demonstrated consequence warrants more.
Separate known deferred work, design preferences, and unreachable defense-in-depth.
No findings is an acceptable result only with the required negative evidence. If a
probe, validator, or sandbox protocol is blocked, report incomplete rather than
inventing success. Put small adoptable test suggestions as fenced text in this audit;
the single-file transfer must not carry implementation patches.

## Reviewer report

### 1. Verdict, captured baseline, and evidence limits

- **Verdict:** needs changes
- **Captured baseline:** `a27c361e890ce6edf2b53c8daa06e2e2550ce183` (overlaying `main` base `af259c017374bed3333a61b0e38bb60e49c2f39e`)
- **Evidence limits:** Focused and full test execution performed strictly inside the isolated sandbox (`/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.KsbbYG`); real binary smoke and CLI end-to-end probes run with a disposable compiled binary against temporary directories outside the checkout; compiler-valid mutations applied one-at-a-time and verified against the captured baseline; sibling review and uncommitted external skill tooling were neither inspected nor executed.

### 2. Workspace and guarded-transfer attestation

```sh
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.KsbbYG
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.KsbbYG/.git
baseline_commit=a27c361e890ce6edf2b53c8daa06e2e2550ce183
source_blob=8654d520132a632916f64863ede7535f0526b565
source_fingerprint=f47a05567248bca2bd65a44f4b387f9d89e5e326
deliverable=planning/audits/6gh86jxtyx5k-2026-10-06-shared-write-and-audit-safety-implementation-antigravity.md
deliverable_changed=true
transfer=succeeded
```

### 3. Verified consumer inventory

- **Existing-document frontmatter encoders (`assembleEditedFile`):**
  - [`internal/store/frontmatter.go:153`](file:///internal/store/frontmatter.go#L153): `updateFrontmatter` — shared entry point for surgical field updates across entities (`fs.SetFields` for tasks, `SetEpicFields`, `SetResearchFields`, `UpdateEpic`, `materializeThreadMutation`, `applyThreadPlanWrites`, `repairInvalidID`, `backfillMissingID`, `moveTask`/`MoveAudit`).
  - [`internal/store/frontmatter.go:256`](file:///internal/store/frontmatter.go#L256): `updateDependencySourceEdits` — dependency graph repair writer invoked by `repairGraphSourcesAtomic` ([`internal/store/graphrepair.go:142`](file:///internal/store/graphrepair.go#L142)).
  - [`internal/store/frontmatter.go:332`](file:///internal/store/frontmatter.go#L332): `replaceBodyWith` / `replaceBodyStamped` — markdown body rewrites preserving frontmatter in `fs.Edit` ([`internal/store/edit.go:133`](file:///internal/store/edit.go#L133)), `TransformAuditBody` ([`internal/store/auditstore.go:175`](file:///internal/store/auditstore.go#L175)), and `AppendAuditBody` / `ReplaceBody` ([`internal/store/body.go:78`](file:///internal/store/body.go#L78)).
- **Fresh document creation writers (`assembleFile`):**
  - [`internal/store/create.go:39`](file:///internal/store/create.go#L39): `createTaskContent` in `fs.CreateTask` — fresh tasks use the ordinary encoder.
  - [`internal/store/frontmatter_preserve.go:29`](file:///internal/store/frontmatter_preserve.go#L29): bypass in `assembleEditedFile` when `!containsBlockScalar(&before)`.
- **Finding header classification and write guards:**
  - [`internal/domain/finding_header.go:35`](file:///internal/domain/finding_header.go#L35): `ClassifyFindingHeaders` — single AST/regex classification for canonical, repairable, ambiguous, and ordinary headings.
  - [`internal/domain/finding.go:263`](file:///internal/domain/finding.go#L263): `NearMissFindingHeaders` — filters classified headers to repairable and ambiguous items.
  - [`internal/domain/finding.go:293`](file:///internal/domain/finding.go#L293): `CanonicalizeFindingHeaders` — rewrites repairable candidates in place while preserving ambiguous and ordinary lines.
  - [`internal/domain/finding.go:327`](file:///internal/domain/finding.go#L327): `NearMissFindingIssues` — produces diagnostic messages distinguishing auto-repairable items (`lint --fix`) from ambiguous ones requiring manual triage.
  - [`internal/domain/finding.go:350`](file:///internal/domain/finding.go#L350): `IntroducedNearMissHeaders` — tracks header text and repairability authority delta between pre- and post-edit bodies.
  - [`internal/domain/finding.go:384`](file:///internal/domain/finding.go#L384): `NearMissWriteError` — write gate for `AppendAuditBody` and `EditAudit`; refuses newly introduced repairable headers while permitting ambiguous prose.
  - [`internal/domain/finding_create.go:52`](file:///internal/domain/finding_create.go#L52): `CreateFinding` — blocks allocation of new finding codes when near misses are present, returning actionable advice (`audit lint` for ambiguous, `lint --fix` for repairable).
  - [`internal/store/auditstore.go:270`](file:///internal/store/auditstore.go#L270): `parseAuditWithFindings` — calculates `a.UnparsedFindings = len(nearMisses)`.
- **Audit read, tally, projection, and DTO consumers:**
  - [`internal/domain/audit.go:106`](file:///internal/domain/audit.go#L106): `Audit.Settled()` — requires `UnparsedFindings == 0` in addition to zero open/in-progress findings.
  - [`internal/domain/audit.go:113`](file:///internal/domain/audit.go#L113): `Audit.ReadyToClose()` — requires `Bucket == AuditOpen && Settled()`.
  - [`internal/wire/dto.go:217`](file:///internal/wire/dto.go#L217): `AuditJSON` — wire DTO defining `UnparsedFindings` (`omitempty`) and `ReadyToClose` (`omitempty`).
  - [`internal/wire/dto.go:280`](file:///internal/wire/dto.go#L280): `toAuditJSON` & `ToLoadedAuditJSON` — maps domain fields to JSON wire envelopes.
  - [`internal/cli/render/columns.go:487`](file:///internal/cli/render/columns.go#L487): `unparsedProjection` — formats `unparsed_findings` (aliased as `unparsed`) with `optIn = true` (omitted from default tables/CSV; outputs `""` for 0).
  - [`internal/cli/render/render.go:686`](file:///internal/cli/render/render.go#L686): `auditProgressCell` & `auditStateNote` — qualifies human progress bar with unparsed warnings and replaces "ready to close" with `→ audit lint <slug>`.
  - [`internal/core/audit_unparsed_test.go:9`](file:///internal/core/audit_unparsed_test.go#L9): `TestPortableAuditReadsRetainIncompleteFindingEvidence` — portable read service preserves unparsed evidence across store boundaries.
- **Init entry-point selection and bootstrap:**
  - [`internal/cli/init.go:109`](file:///internal/cli/init.go#L109): `App.initDirectory` — validates mutual exclusivity across `--path`, `-C`, and `--space`; checks non-empty paths; delegates to `startDir()` when `--path` is unflagged.
  - [`internal/cli/root.go:387`](file:///internal/cli/root.go#L387): `App.startDir()` — shared entry-point resolution; enforces `-C` over `TSKFLW_SPACE`, validates registered `--space`, falls back to cwd.
  - [`internal/cli/init.go:56`](file:///internal/cli/init.go#L56): `initCmd.RunE` — validates selection before `authorizeMutation()`, avoiding spurious writes or side effects on invalid invocations.

### 4. Hostile probe and mutation ledger

| Probe / Mutation | Target Invariant | Exact Test / Command | Result | Consequence |
|---|---|---|---|---|
| **Probe 1** (YAML comment backtracking) | Comments before keys in block-scalar documents must be preserved | Go probe updating sibling key with multi-line comment containing blank line (`\n\n\n`) | Refused with `cannot locate head comment for "tier"` | **Defect (M1)**: Hand-authored documents with paragraph-spaced comments fail comment localization and become uneditable. |
| **Probe 2** (Nested folded scalar refolding) | Decoded YAML values must not silently change | `TestProbe_NestedFoldedScalarCorruptionCaughtByDeepEqual` mutating nested sibling field next to folded scalar with indented line | Refused by `reflect.DeepEqual(want, got)` with `decoded values differ from the requested edit` | Confirms `DeepEqual` guard protects against `yaml.v3` encoder blank-line corruption. |
| **Probe 3** (CLI init selectors) | Target selectors must be mutually exclusive and fail closed before writes | Real binary: `tskflwctl -C $TARGET init --path $TARGET --no-register --json` | Exited with code 11 (`ErrValidation`), zero filesystem writes to caller or target | Confirms explicit conflict validation fails closed. Empty target (`--path ""`) also yields code 11. Unknown `TSKFLW_SPACE` yields code 10 (`ErrNotFound`) without falling back to cwd. |
| **Probe 4** (Audit repair & read projection) | Unparsed findings qualify read completeness, readiness, and advice | Real binary: `audit list`, `audit lint`, `lint --fix`, and `-c slug,unparsed` on test audit with mixed canonical, repairable, and ambiguous findings | Progress shows `100% settled 1/1 · 2 unparsed → audit lint`; `ready_to_close` omitted; `lint --fix` auto-repairs only repairable header; projection emits `"1"` when >0 and `""` when 0 | Confirms full lifecycle qualification, safe non-destructive repair, and wire contract compliance. |
| **Mutation 1** (Bypass source preservation) | Untouched block scalars must not be re-encoded | Mutated `updateFrontmatter` (`frontmatter.go:153`) to call `assembleFile` directly | `TestFrontmatterEditsPreserveUntouchedBlockScalars` FAILED across all scalar styles and line endings | **Killed**. Source preservation is strictly required by the fidelity matrix. |
| **Mutation 2** (Disable decoded-value comparison) | Decoded values must match requested edits | Mutated `frontmatter_preserve.go:96` to bypass `if !reflect.DeepEqual(want, got)` | `go test ./internal/store -count=1` PASSED (all tests green) | **Survived** in existing test suite (**L1**). Killed only by constructed probe `TestProbe_NestedFoldedScalarCorruptionCaughtByDeepEqual`. |

#### Hypotheses resolved or remaining unresolved

1. **Multi-line comment paragraph breaks (Survives / Confirmed Defect M1):** `yaml.v3` condenses multiple blank lines in `HeadComment` into single newlines, breaking the 1-to-1 reverse physical line search in `frontmatterEntrySources` and locking valid files against writes.
2. **Decoded-value comparison test coverage (Survives / Confirmed Gap L1):** No committed test differentiated `DeepEqual` value validation from YAML syntax parsing or broken-anchor errors, allowing the mutation to survive.
3. **Init selector precedence over environment (Resolved):** Both explicit `-C` and `--path` reliably override ambient `TSKFLW_SPACE`, preventing target confusion in CI and scripted environments.
4. **Distinction between recognition and repair authority (Resolved):** Ambiguous headings inside `## Findings` without status markers are flagged by lint as requiring triage, but are not rewritten by `lint --fix` and do not block generic body appends.
5. **CRLF preservation across surgical updates (Resolved):** `assembleEditedFile` normalizes to LF during entry slicing and converts all newlines to CRLF prior to final assembly when `eol == "\r\n"`, preserving byte fidelity.

### 5. Actionable findings

#### M1. Comment-block blank lines cause frontmatterEntrySources to fail comment localization · **Status:** fixed (PR #287)

- **Location:** [`internal/store/frontmatter_preserve.go:174-182`](file:///internal/store/frontmatter_preserve.go#L174-L182)
- **Reachable consequence:** When an existing document containing block scalars has a comment block before any top-level key that includes two or more consecutive blank lines (a common idiom for multi-paragraph comments), any subsequent surgical write to the document—such as `fs.SetFields`, `updateFrontmatter`, `updateDependencySourceEdits`, or `replaceBodyWith`—fails closed with:
  `validation failed: cannot preserve block-scalar frontmatter safely: cannot locate head comment for "<field>"; edit the document explicitly instead`
  The document becomes completely uneditable through automated CLI and store operations even when the modified field is completely unrelated to the comment or the block scalar.
- **Root cause:** In `go.yaml.in/yaml/v3`, `yaml.Unmarshal` condenses multiple consecutive blank lines inside a node's `HeadComment` into a single empty line (`\n\n`). However, `frontmatterEntrySources` performs a strict 1-to-1 reverse-line matching loop against the raw source:
  ```go
  comments := strings.Split(key.HeadComment, "\n")
  for j := len(comments) - 1; j >= 0; j-- {
      if line == 0 || strings.TrimSpace(string(source[lines[line-1]:lines[line]])) != strings.TrimSpace(comments[j]) {
          return nil, nil, 0, fmt.Errorf("cannot locate head comment for %q", key.Value)
      }
      line--
  }
  ```
  When the source contains two blank lines (`lines[line-1]:lines[line]` is empty), matching the preceding comment token against this second blank line fails, returning the error.
- **Minimal reproduction:**
  ```go
  src := []byte("---\nnotes: |\n  hello\n\n# paragraph 1\n\n\n# paragraph 2\ntier: 1\n---\nbody\n")
  _, err := updateFrontmatter(src, map[string]any{"tier": 2})
  // err: cannot locate head comment for "tier"
  ```
- **Recommended correction:** In `frontmatterEntrySources`, permit consecutive blank lines in `source` when backtracking over comment paragraphs, or advance `line--` through empty lines in `source` while matching empty comment segments.

**Resolution:** Reproduced and fixed comment localization by matching non-empty
parsed comment lines through physical paragraph separators, including yaml.v3
CRLF trailing-newline artifacts, without consuming preceding keep-chomp scalar
blanks. TestBlockScalarParagraphCommentsAllowSurgicalWrites covers LF/CRLF, root
indentation, 0/1/2/3/5 separator lines, field/body/dependency edits, and
deletion; TestBlockScalarParagraphCommentsAllowFieldWrites proves actual
filesystem writes preserve values and comments. Focused and full race suites
pass.

#### L1. Decoded-value comparison lacks differentiation test against parser validity · **Status:** fixed (PR #287)

- **Location:** [`internal/store/frontmatter_preserve.go:96-98`](file:///internal/store/frontmatter_preserve.go#L96-L98) and [`internal/store/frontmatter_fidelity_test.go`](file:///internal/store/frontmatter_fidelity_test.go)
- **Reachable consequence:** The critical safeguard in `assembleEditedFile` comparing decoded values (`reflect.DeepEqual(want, got)`) can be disabled or bypassed without failing any test in the test suite. If a future regression alters this check, silent data corruption from `yaml.v3` encoder re-folding of nested multi-line scalars would go undetected by the test suite.
- **Root cause:** The existing refusal tests in `frontmatter_fidelity_test.go` (`TestBlockScalarEditRefusesBrokenAliasBeforeWriting` and `TestBlockScalarFieldWritesPreserveValuesAndRefusalLeavesFileUntouched`) test broken anchors. In both cases, `yaml.Unmarshal(preserved.Bytes(), &got)` at line 93 fails first with an unresolved anchor error. No committed test exercises a fixture where re-encoded YAML is syntactically valid (line 93 succeeds) but the decoded values differ from `want` (line 96 catches the corruption).
- **Minimal reproduction:**
  Mutate line 96 of `internal/store/frontmatter_preserve.go`:
  ```go
  _ = reflect.DeepEqual(want, got)
  // if !reflect.DeepEqual(want, got) { ... }
  ```
  Run `go test ./internal/store -count=1`. All tests pass.
- **Adoptable test suggestion:**
  ```go
  func TestBlockScalarEditRefusesDecodedValueDriftEvenWhenSyntaxValid(t *testing.T) {
      original := []byte("---\ncustom:\n  nested: >\n    first line\n      indented line\n  tier: 1\n---\nbody\n")
      fm, body, err := splitFrontmatterStrict(original)
      if err != nil {
          t.Fatal(err)
      }
      var doc yaml.Node
      if err := yaml.Unmarshal(fm, &doc); err != nil {
          t.Fatal(err)
      }
      mapping, err := documentMapping(&doc)
      if err != nil {
          t.Fatal(err)
      }
      for i := 0; i+1 < len(mapping.Content); i += 2 {
          if mapping.Content[i].Value == "custom" {
              m := mapping.Content[i+1]
              for j := 0; j+1 < len(m.Content); j += 2 {
                  if m.Content[j].Value == "tier" {
                      m.Content[j+1].Value = "2"
                  }
              }
          }
      }
      _, err = assembleEditedFile(fm, mapping, body, "\n")
      if err == nil || !strings.Contains(err.Error(), "decoded values differ from the requested edit") {
          t.Fatalf("expected DeepEqual refusal, got: %v", err)
      }
  }
  ```

**Resolution:** Added
TestBlockScalarDependencyRepairRefusesDecodedDriftWithValidYAML using a
hand-authored dependency sequence and actual dedupe writer, not fabricated YAML
nodes. Syntactically valid re-folded output must fail specifically at
decoded-value comparison; real service dry-run/apply both return validation and
leave original file bytes untouched. Disabling only DeepEqual compiled but
failed the regression with valid YAML and err=nil; restored guard and full race
suite pass.

## Implementation-owner triage (2026-10-07)

Both findings were independently reproduced and fixed; the original needs-changes
verdict above describes the captured baseline, not the repaired tree. M1 additionally
exposed a CRLF trailing-newline artifact in parsed comments, now covered alongside
paragraph spacing. L1 was addressed through the real dependency-repair writer rather
than the suggested manually modified YAML-node fixture. Its compiler-valid single-guard
mutation fails the new regression; the guard is restored. Full race tests, lint, build,
module tidiness, and generated-doc drift checks pass.

The ledger's phrase "full lifecycle qualification" overstates the evidence: this
batch qualifies reads and derived readiness, not every lifecycle verb. TUI, compact
audit info, and actual closure/deferral decisions remain explicitly owned by
[6gh82rm9sf3b](../tasks/6gh82rm9sf3b-carry-unparsed-finding-evidence-into-audit-tui-and-lifecycle-decisions.md).
The reviewer inventory is baseline context, not a blanket owner attestation; the
two reported findings were checked against current callers and production-path tests.
Both external reports have been received, their findings settled, and their audits
closed. Implementing tasks are completed in [PR #287](https://github.com/andy-esch/taskflow/pull/287).
