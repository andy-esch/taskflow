---
schema: 1
id: 6gh86jxtyx5k
bucket: open
area: shared-write-and-audit-safety-implementation-antigravity
date: "2026-10-06"
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

Awaiting the assigned external review. No findings or verdict have been recorded.
