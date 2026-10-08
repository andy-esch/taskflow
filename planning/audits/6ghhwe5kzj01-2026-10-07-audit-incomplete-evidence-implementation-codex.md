---
schema: 1
id: 6ghhwe5kzj01
bucket: closed
area: audit-incomplete-evidence-implementation-codex
date: "2026-10-07"
updated_at: "2026-10-08"
---
# Audit: Incomplete audit evidence and guarded lifecycle implementation — codex — 2026-10-07

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

### Verdict and limits

**Needs changes: one medium product defect, M1, left open.** The guarded filesystem lifecycle refused incomplete evidence in the exercised apply, preview, same-bucket, and retry paths. Its new diagnostic route can nevertheless inspect the wrong audit or become ambiguous because it substitutes a display slug for source-backed identity.

Reviewed the captured dirty implementation, including untracked tests, from baseline `de14a98691791c148c63305e41ec095941db1efc`; no sibling report was read. No implementation fixes were made. The pre-existing literal-open-only parsed-status policy remains the explicitly tracked followup, not a new finding here. Full owner gates listed in the brief are inherited; the focused tests and hostile probes below were independently executed. The bounded pass does not certify arbitrary alternate adapters or non-cooperating writers beyond the exercised CAS interleaving.

### Findings

#### M1. Incomplete-evidence advice loses source identity and can lint a different audit · **Status:** fixed

**Classification:** product defect; medium severity. The persistence refusal remains safe, but the promised executable recovery route can silently report a different audit clean.

**Production trace and root cause:** `internal/domain/audit.go:132` formats the new `audit lint` selector with `a.Slug` (`:133`). The filesystem decoder derives that display slug from the filename at `internal/store/auditstore.go:246`. Meanwhile `internal/store/resolve.go:227` deliberately gives exact canonical IDs priority over display slugs (`:232`, `:238`). The advice therefore does not necessarily resolve back to the record whose evidence was validated. The new human-info warning repeats the same substitution at `internal/cli/render/render.go:198`; its caller discards the loaded source when passing the domain value at `internal/cli/audit.go:550`. TUI detail advice also uses the slug at `internal/tui/detail.go:1714`, despite detail loading retaining the source ID at `internal/tui/commands.go:291`.

**Exact reproduction:** the freshly built sandbox binary and disposable planning space are retained below. Creation is recorded in `identity-advice-probe.py`. Both accepted flat files have valid `schema: 1`, their filename IDs in `id`, `bucket: open`, `area: probe`, and `date: "2026-10-07"`:

- `audits/6g0000000011-6g0000000012.md`: body `## Findings\n\n#### H1 Missing metadata\n`.
- `audits/6g0000000012-clean.md`: body `# Empty audit\n`.

These are supported persisted inputs, as demonstrated by actual info, move, and lint reads; this reproduction does not assume the CLI creation command manufactures such slugs.

```sh
B=/tmp/taskflow-review-20261007.HFi3lp/tskflwctl
R=/tmp/taskflow-review-20261007.HFi3lp/id-shadow
"$B" -C "$R" audit close 6g0000000011 --json
"$B" -C "$R" audit lint 6g0000000012 --json
"$B" -C "$R" audit lint 6g0000000011 --json
"$B" -C "$R" audit info 6g0000000011 --color=never
```

| Command | Exit | Actual stdout | Actual stderr |
| --- | --- | --- | --- |
| close original ID | 11 | Moves receipt names input ID `6g0000000011`, but error says `run tskflwctl audit lint 6g0000000012` | Validation error envelope summarizing the refusal |
| lint emitted selector | 0 | Revision 1.83 receipt with `issues: []`, `unreadable: []` | Empty |
| lint original ID | 11 | Actual ambiguous-heading issue; explains that `lint --fix` will not rewrite it | Validation error envelope |
| human info original ID | 0 | One unparsed heading, again recommending `audit lint 6g0000000012` | Empty |

**Expected:** emitted advice selects `6g0000000011` and reports its ambiguous heading. **Actual:** exact-ID precedence selects the clean sibling. The JSON info receipt still correctly carries `id: "6g0000000011"`; the loss is in the diagnostic selector. The two audit files' hashes were unchanged.

A second real-binary fixture with distinct IDs `6g0000000008` and `6g0000000009` sharing slug `probe` confirms the same root cause without an ID-shadowing slug: close by ID refuses with exit 11, emitted `audit lint probe --json` exits 13 with an ambiguity envelope on stderr and empty stdout, while lint by the original ID exits 11 with the intended heading diagnostic. Root/registry tree snapshots are unchanged. These manifestations are consolidated into M1.

**Recommendation:** preserve the actual source-backed reference when constructing remediation selectors in persistence errors and human/TUI receipts. Retain the semantic domain gate while giving its caller enough structured information to add a canonical route. Do not blindly substitute the frontmatter-declared `a.ID`, since portable loaded records can deliberately have a different `Source.ID`. Add an end-to-end assertion that executes the emitted route against ID-shadowing and duplicate-slug fixtures and verifies the intended audit is inspected. The present tests check the literal `audit lint probe/portable` text rather than its identity semantics.

Evidence: `/tmp/taskflow-review-20261007.HFi3lp/identity-advice-ledger.json` records each stream and exit separately; the duplicate-slug commands are in `production-ledger.json`.

**Resolution:** Preserved adapter-established Source.ID in guarded refusal
diagnostics and loaded-record human/TUI advice; executable-route tests cover
duplicate slugs, ID-shadowing slugs, and declared-ID drift. Domain refusal stays
structured and source-neutral.

### Verified consumer inventory

All locations below were inspected in the sandbox; line numbers refer to the captured implementation.

| Consumer/boundary | Verified production behavior and evidence |
| --- | --- |
| Domain policy | `internal/domain/audit.go:107` `Settled` requires zero unparsed and a nonempty settled parsed tally; `:115` `ReadyToClose` additionally requires open bucket. `:124` `ValidateMove` permits reopen before the incomplete-evidence branch and preserves the old parsed-open check. Focused domain test and real empty/settled/reopen controls passed. |
| Filesystem decoding and writes | `internal/store/auditstore.go:226` `parseAuditWithFindings` derives parsed counts at `:256` and near-miss count at `:263` from the same body bytes. `FS.MoveAudit` at `:108` authorizes before reads, validates those freshly decoded counts at `:135`, then reaches same-bucket `:139` or preview `:152`. Apply locks at `:159`, verifies source identity/content at `:166`, and writes at `:169`. `internal/store/cas.go:123` `verifyUnchanged` re-resolves the filename's stable ID at `:138`. The Unix lock opens the repository directory `s.root` (`internal/store/lock_unix.go:22`) and flocks it at `:26`; there is no invented lockfile. |
| Port and service | `internal/core/store.go:302` `AuditStore` explicitly requires fresh guarded counts and validation before no-op/preview (`:305`). `internal/core/service_audit.go:69`/`:98` retain loaded records and enforce source-ID presence; `:122` delegates each move attempt to the store. `internal/core/retry.go:62` retries conflicts by invoking the store again, with no retry for previews. A conforming alternate adapter must establish both counts on reads and enforce fresh-source domain policy within its own guarded mutation; service does not independently classify arbitrary adapter bodies. |
| CLI receipts | `internal/cli/audit.go:535` info uses `ShowAudit`; local path resolution is capability-gated and uses `record.Source.ID` at `:541`. JSON receives the loaded record at `:548`; human advice has M1. Move at `:603` calls service through `runMoves`. `internal/cli/moves.go:38` records per-item errors, emits JSON on stdout at `:75`, and returns the sentinel-bearing summary at `:85`. Real refusals produced both stdout move receipts and stderr validation envelopes, exit 11. |
| Wire | `internal/wire/dto.go:158` `AuditInfoJSON` adds optional top-level `UnparsedFindings` at `:164`; `ToAuditInfoJSON` copies it at `:171` without changing the five tally keys. `internal/wire/envelopes.go:150` restores source-backed ID at `:152`. `internal/wire/wire.go:350` declares revision 1.83 NOT ADDITIVE for the newly refused moves; `:355` sets the actual revision. `internal/wire/envelopes_test.go:371` exercises a nonzero optional value, with compiled definition validation at `:756`/`:771`; optional-schema acceptance alone would not prove the mapper preserves its value. The independent real-producer semantic check below closes that evidence gap. |
| Portable TUI and actions | `internal/tui/commands.go:233` `loadAuditList` transfers adapter values plus source ID into `auditItem` at `:267`; `loadAuditDetail` at `:284` retains source ID and handles unavailable path capability. `internal/tui/item.go:428` actual delegate qualifies progress at `:446` before readiness `:451`. `internal/tui/detail.go:1698` `renderAuditMeta` uses loaded counts; its separate finding index parses the body at `:1729`, as permitted. `internal/tui/entity.go:397` `moveAudit` uses `ref.key` and returns `actionErrMsg` on service rejection at `:400`. Focused portable/action tests passed; detail advice shares M1. |
| Documentation/help | `docs/ARCHITECTURE.md:756` distinguishes recognition from repair authority; `:763`–`:776` describe loaded counts, parsed percentage, lifecycle guard, and the separate parsed-status followup. ADR-0008's monotonic revision policy was checked. Actual move help is defined at `internal/cli/audit.go:580`. Generated CLI docs matched `docs/cli` exactly in a disposable output directory; emitted JSON Schema matched the committed golden. No additional overclaim was demonstrated beyond M1's promised lint route. |

### Independent evidence and mutation ledger

The binary was built from the sandbox with `go build -o /tmp/taskflow-review-20261007.HFi3lp/tskflwctl ./cmd/tskflwctl`. Caches, outputs, and planning fixtures stayed outside its tree.

- **Real lifecycle matrix:** `production-probes.py` completed 202 process calls. Repairable-fixed, ambiguous, mixed parsed-fixed plus repairable-fixed, clean empty, settled fixed/deferred, ordinary S3/V2 headings with fenced malformed examples, and canonical parsed-open controls were exercised from open/closed/deferred buckets, both close/defer, apply/preview, including same-bucket. Incomplete and parsed-open cases refused; clean controls succeeded. Reopen remained usable with counts preserved. Forty-eight refusal comparisons covered bytes (SHA-256), modes, root/empty-directory/tree entries, and symlink targets in the planning and registry trees; all matched. Accepted moves retained body bytes; this is not a claim that successful writes preserve custom file modes.
- **Retry challenge:** temporary instrumentation around the existing hook and store call proved one hook, one retry, two attempts: first `ErrConflict`, second `ErrValidation` for newly added ambiguity. Fresh read showed unparsed=1 and bucket=open; concurrent bytes survived. `go test -race ./internal/store -run '^TestAuditMoveRetryRevalidatesNewUnparsedEvidence$' -count=10 -v` passed both destinations on every repetition (20 traced subcases). The instrumentation was then restored to captured HEAD and the original test passed with `-race -count=3`.
- **Actual nonzero info:** mixed audit output carried `unparsed_findings: 1` beside exactly `findings: {total:1, open:0, in_progress:0, done:1, dropped:0}`, with correct source ID. Zero output omitted the optional field. Both real outputs passed a temporary semantic validator using the emitted revision-1.83 `AuditInfoEnvelope` JSON Schema; string-valued unparsed counts and revision 1.82 failed validation. Emitted schema bytes equaled `wire.JSONSchema()`. List projection aliases `unparsed` and `unparsed_findings` returned "1" for incomplete and blank for zero without altering defaults.
- **Repair advice controls:** executed the actual emitted route in the same selected disposable space. Repairable lint described canonical repair; `lint --fix` canonicalized it and close then succeeded, with an idempotent second fix. Ambiguous lint explained non-rewrite; fix exited 11, preserved bytes, and close remained refused.
- **Portable controls:** temporarily logged actual delegate/meta rendering in `TestPortableAuditViewsQualifyIncompleteEvidence` at widths 80/120/240. Adapter count 2 survived a plain body, declared ID `declared-other`, source ID `source-audit`, and opaque `urn:opaque:audit`. Mixed views retained 1/1 parsed progress with the warning and no readiness. Clean/settled controls had no false warning; no local path was inferred. Logs include the slug-based advice implicated in M1.

| Probe | Exact edit / intended oracle | Result and restoration |
| --- | --- | --- |
| Required persistence mutation | Removed only `if err := cur.ValidateMove(to); err != nil { return domain.Audit{}, err }` from `FS.MoveAudit`. Ran `go test ./internal/store -run '^TestAuditMovesRefuseIncompleteEvidenceBeforeAnyEffects$' -count=1`. | **Compiler-valid; killed.** All 24 subcases failed at `internal/store/audit_unparsed_move_test.go:34` because the error became nil. Fixed-status bodies avoid incidental protection from the parsed-open branch. Restored `auditstore.go` from captured HEAD; the exact test passed, then the focused store race suite passed again. |
| Retry instrumentation | Counted hook/retry events and traced each returned error; no policy bypass. | Compiled; all 20 race subcases passed with the required conflict→fresh-validation sequence. Original test restored and green. This is an execution proof, not a killed mutation. |
| Temporary semantic/view probes | Added a real-output schema/semantic validator and view logging. | Compiled and passed; scratch files removed and original tests restored. Presentation mutations assigned to the other reviewer were not claimed as independently executed here. |

Independent focused commands, all green on restored implementation:

```sh
go test -race ./internal/domain -run TestAuditMoveRequiresCompleteEvidence -count=1
go test -race ./internal/store -run 'TestAuditMoves|TestAuditReopen|TestAuditMoveRetry' -count=3
go test -race ./internal/cli -run 'TestAuditReadSurfacesQualifyUnparsedFindings|TestAuditMoveRefusals|TestInfoCommandsRemainSemanticWithoutLocalPath' -count=1
go test -race ./internal/tui -run 'TestPortableAuditViewsQualifyIncompleteEvidence|TestModel_AuditCloseBlockedByUnparsedEvidence' -count=1
go test ./internal/wire -run 'TestJSONSchema_ValidatesRealOutput|TestSchemaComments_NotStale' -count=1
go test -race ./internal/store -run '^TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects$' -count=1
go test -race ./internal/tui -run 'TestModel_AuditCloseBlockedByOpenFindings|TestModel_ActionMenuMovesAudit|TestModel_CommandVerbMovesAudit' -count=1
```

Evidence directory: `/tmp/taskflow-review-20261007.HFi3lp`. Named logs include `mutation-guard.log`, `restored-guard.log`, `retry-trace.log`, `retry-restored.log`, `info-schema.log`, `portable-view-trace.log`, and restored focused-suite logs. Temporary probe source is retained there, outside the sandbox.

### Separate systemic pass: bounded hypothesis ledger

| Hostile hypothesis | Disposition and evidence |
| --- | --- |
| New ambiguity between projection/validation and write is overwritten, or retry reuses old counts. | **Falsified for the coordinated interleaving:** traced hook, actual first conflict, fresh second refusal, and unchanged concurrent bytes across 20 race subcases. CAS and policy consume the same original read; retry performs a new read. |
| Shared no-op/preview abstraction bypasses incomplete-evidence validation. | **Falsified:** guard precedes both returns, real matrix includes both destinations and all source buckets, and bypass mutation is killed by the intended test. Previews/no-ops describe their read snapshot, not a promise to lock future edits. |
| Test helpers or default optional values mask count loss, false readiness, or a wire omission. | **Falsified in exercised consumers:** inspected `internal/testutil/snapshot.go:20` and actual fixture construction; used independent mode-0640 binary fixtures, nonzero real info plus semantic negative controls, and body-disagreeing opaque-source TUI views at three widths. Schema validation alone would permit optional omission; the semantic assertion is necessary. |
| Remediation retains canonical identity because selection/actions already do. | **Reproduced, M1:** action and JSON identity are intact, but advice substitutes slug; exact-ID precedence silently redirects lint and duplicate slugs make it ambiguous. Existing literal-string advice assertions do not test this contract. |
| Recognition mistakes ordinary/fenced prose for evidence, or UI preflight is the sole close/defer boundary. | **Falsified for selected controls:** ordinary S3/V2 and fenced malformed examples remain zero/closeable; TUI action rejection reaches the service/FS boundary, and direct service/binary calls refuse. Alternate adapters have an explicit enforcement obligation rather than an independently reparsing service fallback. |

### Isolation and guarded-transfer attestation

The general helper created an independent no-hardlinks clone, captured staged/unstaged/untracked source state in its sandbox-only checkpoint, and retained it throughout review. All mutations were restored from that captured HEAD. Immediately before transfer, only this assigned audit had an unstaged delta; no staging, extra commits, or implementation changes remained. The brief above was preserved verbatim and the audit diff was inspected.

Final helper verification/guarded transfer attestation (stdout retained as `verify.log` and `transfer.log` in the evidence directory):

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.jmnLd6
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.jmnLd6/.git
baseline_commit=de14a98691791c148c63305e41ec095941db1efc
source_blob=ac1d81311d80ecf639d04dcf8be5702d6368320d
source_fingerprint=64cd8e715771b1af541c94b603ac89cac97ea4e6
deliverable=planning/audits/6ghhwe5kzj01-2026-10-07-audit-incomplete-evidence-implementation-codex.md
deliverable_changed=true
transfer=succeeded
source_deliverable=/Users/andyeschbacher/git/andy-esch/taskflow/planning/audits/6ghhwe5kzj01-2026-10-07-audit-incomplete-evidence-implementation-codex.md
sandbox_retained=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.jmnLd6
```

The captured fingerprint identifies the handoff snapshot; transfer guards the assigned source blob, not an assertion that every other owner file remains unchanged. Only this audit was transferred, using the helper's guarded atomic rename. Sandbox and evidence remain in place until the implementation owner confirms receipt.

### Implementation-owner triage (2026-10-08)

M1 accepted and fixed through the audit finding verb. The reviewed snapshot's
verdict above is historical, not a claim that it included the later correction.
`AuditIncompleteEvidenceError` carries semantic refusal; `core.AuditMoveError`
attaches the guarded adapter's source identity. CLI info/list/show/status and TUI
detail now retain `RecordSource.ID` for lint advice rather than deriving a selector
from display metadata or the frontmatter declaration.

`TestAuditIncompleteEvidenceAdviceRetainsSourceIdentity` executes advice through
the real CLI controller for duplicate slugs, ID-shadowing slugs, and declaration
drift, and checks no effects. Portable CLI/TUI fixtures use divergent declared
and source IDs. A fresh process-level smoke confirms the lint route inspects the
ambiguous original while the shadowing sibling remains clean. A compiler-valid
mutation substituting the slug in `AuditMoveError` fails all three identity cases;
restoration passes the exact test under the race detector. Existing CAS/reopen/
empty-audit behavior remains covered by the full race suite.

The adjacent pre-existing `audit edit` receipt/advice gap is recorded in
[body-write guard scoping](../tasks/6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md),
not silently folded into these read and lifecycle changes.
