---
schema: 1
id: 6gh86jxj4sve
bucket: closed
area: shared-write-and-audit-safety-implementation-codex
date: "2026-10-06"
updated_at: "2026-10-07"
---
# Audit: Shared writes, bootstrap targeting, and audit evidence safety — codex — 2026-10-06

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

### Verdict and scope

**Needs changes.** H1 demonstrates unsafe automatic repair and false readiness; M1 demonstrates a selected-target repair receipt that causes wrong-directory writes when followed. Both findings remain open. The focused lens, real YAML cross-check, bounded mutations, non-default semantic/schema validation, and separate systemic pass are complete.

Reviewed the captured working state at sandbox baseline `179ef203c1c60ec637d3338d2b06f42710abe574`, whose parent is the stated source base `af259c017374bed3333a61b0e38bb60e49c2f39e`. `git diff HEAD^ --stat` showed 87 changed files, 2,171 insertions and 264 deletions, including the new classifier, preserving assembler, and regression files. This is the helper's dirty-state checkpoint, not a review of source HEAD alone. The sandbox began clean; implementation mutations were restored to this checkpoint.

Independent `just test` (`go test -race ./...`) passed before probes and again after restoring/removing all probes. Focused suites and a fresh binary build also passed. Generated CLI references, schema comments, and the JSON Schema golden matched disposable outputs byte-for-byte. Owner lint/tidy evidence is inherited, not claimed as independently rerun. Green suites do not invalidate the demonstrated defects.

The sibling review was not consulted. No implementation fixes are delivered. The known TUI/lifecycle and creation/edit-baseline followups remain deferred as stated in the brief. The read task explicitly targets list/show: `audit info` still drops the new count in `internal/wire/dto.go:158` / `:168` and `internal/cli/render/render.go:191` (a real `audit info ambiguous --json` returned only a zero parsed tally). No claim is made that every readout now carries the warning; this unchanged adjacent surface is an inventory limit, not an additional batch finding.

### Open findings

#### H1. Inline examples grant repair authority and manufacture a settled finding · **Status:** fixed

**Location:** `internal/domain/finding_header.go:74` and `:81`; evidence pattern `internal/domain/finding.go:68`; durable consumer `internal/core/finding.go:305`.

The classifier masks fences, then searches the entire heading-owned section for either a line-start Status marker **or a middle-dot marker anywhere**. The latter is not restricted to the heading and does not exclude inline code or quoted examples. A documented example therefore grants automatic repair permission. The existing parser subsequently consumes the same non-authoritative Status text, producing a false terminal finding and readiness.

Minimal supported input: create an open audit with ordinary valid frontmatter and this body:

~~~~markdown
## Findings

#### H1 Example syntax

Inline example: `· **Status:** fixed ` is documentation.
~~~~

Against a freshly initialized isolated space, run:

~~~~sh
"$B" -C "$R" audit lint inline --color=never
"$B" -C "$R" lint --fix --color=never
"$B" -C "$R" audit show inline --frontmatter-only --json
"$B" -C "$R" audit lint inline --color=never
~~~~

**Observed:** lint exits 11 and explicitly recommends the canonical heading `#### H1. Example syntax` and says `lint --fix` applies it. Following that advice exits 0 and rewrites the heading, stamps `updated_at`, and changes the read from `findings:0, unparsed_findings:1` to `findings:1, done_findings:1, ready_to_close:true`. Subsequent audit lint exits 0; human show reports `100% settled 1/1 — ready to close`. A second fix is a byte-identical no-op, so idempotence does not catch the corruption. The exact generated 1.82 schema also accepts the false-ready payload: the defect is authority/semantics, not JSON shape.

**Consequence:** routine automatic repair changes documentation into actionable audit data and advertises completion without a real finding or authoritative Status. A real `audit append` of an H2 example with the same inline-code metadata also exited 11 with canonical-spelling advice and preserved the actual file bytes (`inline-append-ledger.json`). This meets High's materially misleading semantics/safety-failure threshold; it is separate from the already deferred lifecycle gate.

**Correction:** derive repair authority from a visible Status line or the candidate heading's own visible inline metadata. Mask inline code and exclude quoted/example metadata; never search arbitrary prose for a middle-dot Status. Keep the parser and authority scanner consistent enough that repair cannot import example statuses into real findings. Add a real `lint --fix` regression for the fixture above and for the corresponding quoted/plain-prose middle-dot variants, asserting no heading rewrite, no invented findings/readiness, and successful unrelated narrative writes.

**Evidence:** `/private/tmp/taskflow-codex-evidence.tQOqUg/inline-authority-ledger.json` contains the complete before/after document, hashes, argv, stdout/stderr, and exit codes; `inline-authority-probe.py` is the reproducible process driver. A broader seven-document process matrix independently reproduced the same result.

**Resolution:** Classifier and canonical parser now share byte-offset-preserving
Status authority: visible heading metadata or standalone Status lines only,
excluding inline code and body prose/quotes/indented examples. Real lint --fix
regressions leave the reported examples byte-identical across repeated fixes,
invent no finding/readiness, and allow narrative append. Parser tests pin
genuine metadata spans, code-formatted decoration replacement, and CRLF
preservation; full race suite, lint, and generated comparisons pass.

#### M1. Init repair advice loses the selected directory · **Status:** fixed

**Location:** `internal/cli/init.go:159`, emitted in the JSON receipt at `:168` and human advice at `:181`. The newly honored selector flows into this existing receipt producer from `:56` / `:68`.

From an empty caller directory, initialize a distinct selected directory with a `planning` subdirectory, remove only its `planning/threads` scaffold, then request its existing-topology read:

~~~~sh
"$B" init --path "$T" --taskflow-root planning --no-register --json
# Delete only the disposable fixture's planning/threads scaffold.
"$B" -C "$T" init --json
~~~~

The receipt correctly names `root: "$T"` and `scaffold_repair_available:true`, but emits `tskflwctl init --taskflow-root "./planning"` with no selector. Execute that exact advice from the original caller cwd, as recorded in `repair-target-probe.py`.

**Observed:** the selected target remains without `planning/threads`; the caller acquires a new `.tskflwctl.toml` and planning tree, and the isolated registry acquires the caller's entry. The topology read itself was byte/snapshot-identical across caller, selected target, and registry; the unintended writes happen only when its advertised remedy is followed. Evidence: `/private/tmp/taskflow-codex-evidence.tQOqUg/repair-target-ledger.json` records all commands, receipts, and before/after tree hashes.

**Consequence:** recovery guidance defeats the explicit-target intent and creates unrelated planning data. The receipt construction predates this batch; the now-correct `-C`/registered selection exposes its missing target context. This is an adjacent selector/receipt defect, not a claim that the selection resolver itself chooses the wrong directory.

**Correction:** include a safely quoted explicit `-C`/path for the resolved selected directory in scaffold repair advice. Preserve relevant invocation policy such as `--no-register` where applicable. Execute the emitted command in a regression with a different real caller cwd, asserting that only the selected scaffold is repaired and caller/registry snapshots remain unchanged under opt-out.

**Resolution:** Emitted scaffold repair commands now carry a POSIX-shell-quoted
explicit -C target and retain flag/environment registration opt-out.
TestSmoke_InitEmittedRepairCommandRetainsTargetAndPolicy executes the exact
command from a different real cwd for -C, --path, registered --space, and
environment selection, including quote/expansion-like path contents. Only the
selected scaffold changes; caller and registry remain unchanged. Focused and
full race tests pass.

### Verified consumer inventory

Paths/lines below refer to the restored captured sandbox, and were checked by reads and caller searches. Package paths such as `store/...`, `domain/...`, and `cli/...` are relative to `internal/`; bare encoder filenames are relative to `internal/store/`.

| Boundary | Reachable consumers and evidence |
|---|---|
| Existing-document field encoder | `internal/store/frontmatter.go:118` routes to `assembleEditedFile` at `:153`. All production caller sites found by `rg -n -F 'updateFrontmatter(' internal/store -g '*.go' -g '!**/*test*'`: `fsstore.go:282`, `lifecyclemutation.go:221`, `threadmutation.go:202`, `graphmutation.go:158` and `:166`, `epicstore.go:134` and `:194`, `researchstore.go:170`, `auditstore.go:148`, `fix.go:229` and `:316`, `edit.go:109` (all under `internal/store/`). These cover field, lifecycle, graph, thread, epic, research, audit-bucket, identity-repair, and editor-stamp writes. |
| Dependency/body encoders | `internal/store/graphrepair.go:203` calls `updateDependencySourceEdits` (`frontmatter.go:160` / `:256`). `replaceBodyStamped` (`:298` / `:332`) is called by `body.go:129`, `:171`, `:229`, `:289`, and `researchstore.go:275`. Fresh creation uses ordinary `assembleFile` at `create.go:39`; block-scalar-free edits use it at `frontmatter_preserve.go:28`. `fixFrontmatterText` (`fix.go:327`) is a separate line fixer, not a claimed YAML-node encoder consumer; it skips block-scalar indicators at `:411`. |
| Recognition/lint/repair/write guards | `domain/finding_header.go:35` owns classification; `domain/finding.go:263`, `:292`, `:325`, and `:356` derive diagnostics, repair, advice, and introduced identities. `core/finding.go:260` / `:352` and `core/service.go:928` consume snapshot evidence. Repair recomputes through `core/finding.go:305` and `store/body.go:189`. Generic append guards at `store/body.go:123`; whole-file audit edit guards at `store/edit.go:318`. Typed creation validates at `domain/finding_create.go:52` and allocates within the fresh CAS callback at `core/finding_create.go:50`. |
| Read/tally/DTO/projection | `store/auditstore.go:232` derives parsed counts and unparsed evidence together (`:262` / `:269`); `store/lintsource.go:26` builds audit snapshots. Ordinary loaded reads pass through `core/service_audit.go:69` / `:98`. Full DTO mapping is `wire/dto.go:281`, envelopes `wire/envelopes.go:782` / `:798`; readiness is `domain/audit.go:113`. Human list/show share `cli/render/render.go:686` / `:704`. Optional aliases and default-column policy are `cli/render/columns.go:485` / `:538`; info is the separate narrower consumer described above. |
| Init target/topology/registry/authorization/receipts | `cli/init.go:112` and `cli/root.go:387` resolve selectors; `core/space_registry.go:46` validates the exact registered entry. Bootstrap skips discovery through `cli/root.go:162`, binds safety, and authorizes at `cli/init.go:62` before topology effects. `config/config.go:497`, `:808`, `:877` implement scaffold, pointer, link-back; `cli/init.go:425` and `core/space_registry.go:124` register fresh topology identity. Receipt producers are `cli/init.go:163`, `:275`, `:387`. Store CAS locks the repository directory itself (`store/lock_unix.go:22`), and registry locking uses the config directory (`userconfig/lock_unix.go:15`); no invented lock-file capability is assumed. |
| Machine/docs contract | `internal/wire/wire.go:343` labels 1.82 **NOT ADDITIVE**, explaining selector tightening and withheld readiness. `planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md` and `docs/ARCHITECTURE.md:64` were read. Committed revision marker is `internal/cli/testdata/golden/machine_contract_revision.txt` = 1.82; `projection_contract_json.golden:190` includes canonical `unparsed_findings` and alias `unparsed`. The exact typed schema, comments, and generated command references matched fresh disposable generation. |

### Hostile execution and mutation ledger

Evidence root: `/private/tmp/taskflow-codex-evidence.tQOqUg`. `B` was the freshly built `tskflwctl` there. Probe targets/registries and comparison outputs were outside both checkouts. Fixtures are supported hand-authored Markdown/YAML read by the real CLI/local adapter; none relies on fabricated internal YAML nodes. The named `TestReviewer...` probes are temporary reviewer tests, retained as source files in this evidence directory and removed from the sandbox before transfer; they are not shipped regressions. The two temporal probes wrap the real local adapter solely to interleave a supported file edit between snapshot and action.

| Probe / exact driver | Before → after / result |
|---|---|
| `python3 .../probes.py`: real selector processes | From an empty caller, `-C '../selected target' init --taskflow-root planning --no-register --json` with ambient `TSKFLW_SPACE=missing` scaffolded only the fresh selected target; caller and registry remained empty. Identical-path and all three selector conflicts, explicit blanks, unknown/stale registered entries, and dry-run were snapshot-identical. Validation refusals had stdout empty, stderr JSON revision 1.82/code `validation`, exit 11; unknown/missing-config selections exited 10/`not-found`. Explicit/environment registered pointer reads preserved the exact entry point. Relative pointer/link-back initialization preserved caller cwd. The driver completed 53 recorded process calls and all snapshot assertions passed. |
| Same process driver: reads/projections | Empty, wholly unparsed, ambiguous, mixed, settled-incomplete duplicate, ordinary/fenced examples, and nested-header Status ownership were exercised. Before repair, mixed = parsed 1/done 1/unparsed 1 with readiness withheld; percentage remained 100% of parsed findings. Both projection aliases returned `"1"`; zero returned blank. Default table/CSV omitted the optional column. Ordinary reads exited 0 and qualified incomplete counts; lint remained the validation surface. |
| Same process driver: generic append and typed creation | Introduced evidence-backed `H-2` refused without changing bytes. An ambiguous `H2` narrative append succeeded; later appending Status to the unchanged heading refused without changing bytes. Typed creation refused ambiguity/duplicate drift with clarification advice, and genuine repairable drift with `lint --fix` advice; all refusals preserved actual files. |
| Same process driver plus `inline-authority-probe.py`: execute advice | Genuine mixed drift repaired to two parsed findings, one open/one done (50%); second fix was byte-identical. Ordinary/fenced, ambiguity, duplicate, and Status borrowed from a nested heading remained unchanged; leftovers continued to lint. Inline-code Status instead produced H1's false-ready finding, including standalone fix/lint success and a no-op second pass. |
| `REVIEW_EVIDENCE_DIR=... go test ./internal/store -run '^TestReviewerRealYAMLFidelity$' -count=1 -v` | Real `FS.SetFields`, tiers 3→4→2, preserved exact CRLF `>2+` scalar/alias/nested-literal bytes and unchanged body. Decoded notes remained `"first wrapped\n\n  more indented\nlast\n\n\n"`; alias and nested decoded values also matched. Anchor-breaking replacement returned `ErrValidation`; before/after file SHA256 = `d9da83bbca1a1a37855975538f9909d17d2b3ab714b1b21a22b78a78a65ef55f`. Existing `TestFrontmatterEditsPreserveUntouchedBlockScalars` and `TestBlockScalarFieldWritesPreserveValuesAndRefusalLeavesFileUntouched` independently passed the encoder matrix and repeated real writes/refusal. |
| `REVIEW_EVIDENCE_DIR=... go test ./internal/wire -run '^TestReviewerNondefaultAuditSchema$' -count=1 -v` | Five actual CLI show payloads (including nonzero unparsed values) validated against generated `https://github.com/andy-esch/taskflow/internal/wire/json-envelopes/1.82#/$defs/AuditShowEnvelope`. CLI schema equaled freshly generated schema. Wrong optional-field type and wrong revision were rejected; semantic assertions checked incomplete readiness, parsed counts, and zero omission. Projection assertions checked aliases, blanks, and defaults separately, as ADR-0008 excludes dynamic projected rows from typed-envelope schema claims. |
| One-at-a-time mutation: ignore `-C` after init selection | Overrode `path = "."` only when `chdirSet` in `cli/init.go:125`. `go test ./internal/cli -run '^TestInitChdirTargetsFreshDirectoryAndRepairWithoutTouchingCaller$' -count=1` compiled and failed at `init_selection_test.go:43` with wrong-caller topology conflict. Restoring captured `HEAD:internal/cli/init.go` made the exact test pass. Log: `mutation-init.log`. |
| One-at-a-time mutation: repair ambiguous candidates | Changed the `h.Repairable` filter in `domain/finding.go:295` to `true`. `go test ./internal/domain -run '^TestFindingHeaderClassificationMatrix$' -count=1` compiled and failed both ambiguity and S3-without-context subtests at `finding_header_test.go:39`. Restoring captured `HEAD:internal/domain/finding.go` made the exact test pass. Log: `mutation-repair.log`. No compiler failure counted as a kill; no other mutation was active. |

Focused execution also passed the brief's store/domain/CLI/core/render/smoke selections, including `TestDirectWriteCommandsRejectReadOnlyClassification`. Logs are `store-focused.log`, `domain-focused.log`, `cli-focused.log`, `core-focused.log`, `render-focused.log`, and `smoke-focused.log`. Fresh generation was compared using `diff -qr docs/cli .../docs`, `cmp internal/wire/schema_comments.json .../schema_comments.json`, and `cmp internal/cli/testdata/golden/schema_jsonschema.golden .../schema.json`; every comparison returned 0.

### Separate adversarial systemic pass

Three shared-boundary hypotheses were settled after the focused pass:

1. **Shared Status regex appears authoritative but imports example text:** demonstrated H1 through classifier→repair→parser→tally→readiness, with successful lint and schema validation after corruption. Fence and ordinary-prose-only fixtures cannot protect this entire evidence class.
2. **Target context survives in every actionable receipt:** falsified by M1. The resolver and receipt root are correct; recovery serialization drops the selected directory, and executing the receipt writes caller/registry instead.
3. **Snapshot evidence is replayed after document changes:** challenged using `TestReviewerRepairReclassifiesCurrentBody` against an actual FS wrapper. After snapshot selection, removing Status caused zero fixes and byte-identical current content; moving/changing the header to `M-2` repaired only the fresh `M2` heading, not stale `H1` text. Both subtests passed. This settles body/evidence freshness for these interleavings, not all identity-rebinding races. Existing CAS/concurrent-append store tests also passed.

Small adoptable regression suggestions (test names are proposals, not shipped tests):

~~~~text
TestLintFix_InlineStatusExampleDoesNotManufactureFinding
  Run real lint --fix on the H1 fixture; preserve the heading/body;
  assert parsed findings=0 and ready_to_close absent/false.
TestInit_EmittedRepairCommandRetainsSelectedDirectory
  Caller cwd differs from selected target; remove target threads scaffold;
  execute scaffold_repair_command verbatim from caller; target repaired;
  caller unchanged and registry unchanged when --no-register was requested.
~~~~

### Sandbox and guarded-transfer attestation

The helper created an independent `--no-hardlinks` clone with its own in-tree Git directory and no alternates, then overlaid and checkpointed the source's current state. Pre-copy setup also included `pwd` and an AGENTS.md filename-only search in the parent tree (no matches); implementation inspection began only after sandbox creation. All implementation inspection, builds, tests, mutations, scratch tests, generation, and report editing occurred in the sandbox. No reviewer commit followed its baseline. Scratch tests were copied to the external evidence root, then removed; both mutation files were restored. Only this assigned audit differed at verification/transfer. No source project mutation, staging, restoration, or manual file copy-back was used.

Helper attestation for the final guarded one-file transfer:

~~~~text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.swH67U
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.swH67U/.git
baseline_commit=179ef203c1c60ec637d3338d2b06f42710abe574
source_blob=8aad8f5c5ffd5db3e9fdcaaa558efeacdd384a2d
source_fingerprint=f47a05567248bca2bd65a44f4b387f9d89e5e326
deliverable=planning/audits/6gh86jxj4sve-2026-10-06-shared-write-and-audit-safety-implementation-codex.md
deliverable_changed=true
transfer=succeeded
source_deliverable=/Users/andyeschbacher/git/andy-esch/taskflow/planning/audits/6gh86jxj4sve-2026-10-06-shared-write-and-audit-safety-implementation-codex.md
sandbox_retained=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.swH67U
~~~~

The helper's verification/transfer output is retained in the external evidence root as `verify.log` / `transfer.log`. The workspace and evidence are retained pending implementation-owner receipt; findings remain open.
