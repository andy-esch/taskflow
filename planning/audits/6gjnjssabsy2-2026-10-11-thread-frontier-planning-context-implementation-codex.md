---
schema: 1
id: 6gjnjssabsy2
bucket: closed
area: thread-frontier-planning-context-implementation-codex
date: "2026-10-11"
---
# Audit: Thread frontier planning context implementation — codex — 2026-10-11

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

Review the Thread frontier planning-context implementation for merge readiness. This is a
contained CLI/wire change, but a shared DTO propagates into multiple read and mutation-receipt
envelopes. Primary lens: contract fidelity and raw-data preservation. Cross-check the human
frontier against the same projection, especially free-form text and unchanged dispatch semantics.
One independent Codex review is requested; do not implement fixes or create a PR.

## Review target

- Read-only source: `/Users/andyeschbacher/git/andy-esch/taskflow`, branch `main`.
- Base: `9bc8d72a9d3adce4b58ebc63eaf5b0bec9cc1b17`. Review the captured dirty snapshot,
  including unstaged and untracked work, not HEAD alone. Record the helper's fingerprint.
- Implements [6g6wdvfp2ksa](../tasks/6g6wdvfp2ksa-make-thread-frontier-help-choose-among-independent-candidates.md)
  in [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md).
- Governing rules: `CLAUDE.md`, the planning-data checklist in `docs/ARCHITECTURE.md`,
  [ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md),
  and `docs/THREADS_COMPATIBILITY.md`.
- Production seams: `core.ThreadTaskView.Task` already contains planning attributes;
  `wire.ThreadTaskJSON.Planning`, `ThreadTaskPlanningJSON`, `toThreadTaskJSON`,
  `ToThreadViewJSON`; `render.ThreadFrontierHuman` and `threadPlanningLabel`;
  `cli.newThreadFrontierCmd`; revision `1.86` in `internal/wire/wire.go`.
- Tests: `TestThreadPlanningContextDoesNotScheduleWork`,
  `TestServiceThreadViewsReadThreadsBeforeTasks`,
  `TestThreadViewPlanningContextPreservesRawValuesAndCanonicalIdentity`,
  `TestThreadPlanningJSONSchemaAndUnsetRepresentation`,
  `TestThreadFrontierPlanningContextAcrossRealAdapter`,
  `TestThreadFrontierHumanShowsInFlightBeforeEligibleMembers`,
  `TestThreadFrontierHumanKeepsContextOnItsOwnSafeLine`,
  `TestThreadPreviewReleaseWireSemanticsRemainCompatible`,
  `TestJSONSchema_ValidatesRealOutput`.
- Fixtures: `internal/cli/testdata/golden/thread_frontier_planning_{human,json}.golden`,
  the existing Thread/dependency-query machine goldens, and generated schema comments/CLI docs.

Verify a consumer inventory: Thread list/show/frontier, graph/plan's embedded view,
`ToThreadUpdateJSON` before/after, and `toThreadProjectionImpactsJSON` used by task/dependency
receipts. Check embedded external-prerequisite DTOs too. Graph nodes and compact
`GraphTaskJSON` are not being widened; do not mistake them for Thread task views.

Known out-of-scope issue:
[6gjn7axv0gmy](../tasks/6gjn7axv0gmy-explain-legacy-dependency-degradation-in-thread-projections.md)
tracks legacy-only degradation with empty Thread cause arrays. It was reproduced in the real-adapter
test and added independently to the Thread. Confirm it is pre-existing; challenge deferral only
with a concrete new consequence. Automatic ranking, TUI redesign, and changed eligibility are
not part of this implementation.

## Intended contract to challenge

1. Candidate status, priority, tier, and effort come from the same core task projection.
   No extra per-task reads, path parsing, or primary-adapter persistence access is introduced.
2. Human output preserves complete candidate slug/stable ID lines, puts planning context beneath
   them, keeps in-flight work first, and does not mutate/reorder the caller's slices.
   Candidates remain stable-ID ordered, not priority-ranked.
3. Raw JSON planning fields are required on each Thread task view: priority and effort are strings,
   tier is an integer. Missing recorded data is `"" / 0 / ""`, not display fallbacks.
   Literal `Unknown` effort and other free-form recorded values survive unchanged.
   Invalid authoring metadata remains a lint concern, not a new graph eligibility policy.
4. Terminal formatting/sanitization belongs only to human rendering. Non-default, multiline,
   control-sequence-bearing values must not be normalized in core or wire output.
5. Frontier membership stays graph-derived: next-up/ready-to-start clear-gated members only.
   In-flight tasks and outside prerequisites are context, not dispatchable candidates.
   Broken/degraded graph or projection evidence still suppresses the frontier.
6. Revision 1.86 is ADDITIVE under the existing reader contract. Old identity/state meanings and
   ordering remain intact; compact dependency-query task shapes, persisted documents, and plan
   schemas do not change. Generated artifacts must match the code, not hide unrelated drift.

## Mandatory evidence floor

- Trace an actual `thread frontier` read from neutral service projection to both output modes.
  Run a real CLI fixture with independent next-up/ready candidates, opposing priorities/tiers,
  literal Unknown effort, unset values, active blocked work, and an outside prerequisite.
  Compare decoded fields and candidate identities/order; keep stdout/stderr/exit observations.
- Validate non-default planning data against the emitted exact-revision JSON Schema. Check one
  mutation receipt with populated before/after or Thread-impact views, not just empty/default
  fixtures. Check raw values and required defaults semantically, not merely schema acceptance.
- Run the new regressions and historical wire-compatibility test. Test degraded and broken cases;
  distinguish missing member, unsorted membership, and broken repository graph causes.
- Execute two compiler-valid mutations, separately, in your sandbox:
  1. Force mapped planning tier to zero in `toThreadTaskJSON`.
     `TestThreadViewPlanningContextPreservesRawValuesAndCanonicalIdentity` and
     `TestThreadPlanningJSONSchemaAndUnsetRepresentation` should fail for nonzero tiers;
     schema acceptance of zero alone cannot prove preservation.
  2. Bypass human sanitization for nonempty values in `threadPlanningLabel`, retaining
     the empty fallback and compilable imports. The hostile-text case of
     `TestThreadFrontierHumanKeepsContextOnItsOwnSafeLine` should fail, not just a build.
  Record exact edit, exact intended test, expected/observed failure, and restored-green result.
  A compile error is invalid evidence. Restore each probe before the next.
- No-findings is valid only with this evidence and an explicit inventory/limits. There is no
  finding quota and no requirement to rerun every unrelated repository gate.

## Required hostile angles

Settle these hypotheses as reproduced, falsified, or unresolved:

- A shared/embedded receipt consumer loses non-default metadata, defaults it, or serializes
  terminal-cleaned values despite frontier tests passing.
- Adding required planning properties to the latest schema breaks the documented additive
  reader promise or silently changes an old meaning. Explain any alleged conflict against
  ADR-0008's exact-revision/schema and unknown-field policy.
- Metadata changes candidate order, promotes active/external work, or obscures stable identity.
  Try equal metadata, reversed input order, sparse attributes, and source/declaration ID drift.
- Free-form effort can forge rows or control terminal output. Exercise ANSI SGR, OSC, newline,
  carriage return, tab, bell, whitespace-only data, Unicode, and long text through a reachable
  parser/reader. Compare no-color and color-enabled output with unchanged raw JSON.
  Narrow streaming output may wrap; it must not silently truncate handles. Distinguish
  existing unsanitized fields from newly introduced metadata before claiming a regression.
- Global golden regeneration conceals unrelated changes. Inventory deltas after accounting for
  the revision and added Thread planning object; historical compatibility fixtures must stay intact.

Then do a separate systemic pass over shared converters, embedded DTOs, fixture defaults, and
health ownership. A checked AC, constructed DTO, or green suite is not by itself a producer trace.
Label preferences/coverage gaps/defense-in-depth separately from reachable product defects.

## Validation and restoration

Use the Mandatory reviewer sandbox protocol above before any code inspection or checks.
Put caches and scratch fixtures outside the captured sandbox tree, or clean only your own artifacts.
Use the sandbox's rebuilt binary or repository runner, not an installed binary of unknown provenance.

Focused checks (set `GOCACHE` to your external scratch cache if needed):

```sh
go test -race ./internal/core ./internal/wire ./internal/cli/render ./internal/cli
go test ./internal/cli -run '^TestThreadPreviewReleaseWireSemanticsRemainCompatible$'
```

Owner evidence: after the cleanup pass, `just test` (full race suite), `just lint`,
`just build`, and full `go test ./internal/cli -update` passed. Planning/audit lint and
documentation regeneration are checked at handoff. `just docs-check` compares against HEAD,
so the intentional uncommitted frontier-help diff is expected; verify generator idempotence
against captured files, not a false claim that HEAD already contains them.
You may inherit these broad gates, but execute the focused hostile probes independently.
On environment/permission/isolation failure, report incomplete and preserve the sandbox.

Restore to the helper's sandbox-only baseline so only this assigned audit differs. Do not
create another commit or push. Inspect the diff, run helper verify, and use guarded transfer
for this audit only. Never fall back to editing the source or copying probe fixes there.

## Deliverable

Keep this brief intact and complete the report below. Lead with verdict/limits, then findings
consolidated by root cause. Each finding needs a reachable producer/caller, consequence,
reproduction, evidence, recommendation, and any uncertainty; leave dispositions open.
Prefer the sandbox CLI's `audit finding new` verb. Severity follows demonstrated impact,
not the number of failed ACs. If no defect reproduces, say so and retain the evidence ledger.

Include: verified consumer inventory; commands/checks run versus inherited; mutation ledger;
bounded surviving-hypothesis ledger; and sandbox path, independent Git directory, baseline,
captured source fingerprint/blob, deliverable path, and verify/transfer attestation.
The only source write permitted is the helper's guarded transfer of your assigned report.

## Reviewer report

### Verdict and limits

**Ready within this review's bounded scope; no new reachable defect reproduced.**
Completed the checklist pass and a separate adversarial systemic pass over shared
converters, embedded receipts, fixture defaults, health ownership, and state changes
between projection and action. No findings were created or closed; the audit remains
in the open bucket. No implementation fixes, extra commits, pushes, or PR were made.

Reviewed the captured dirty snapshot relative to
`9bc8d72a9d3adce4b58ebc63eaf5b0bec9cc1b17`, including the two new planning goldens
and the new wire test file. This is not a HEAD-only review. Inherited the owner's
reported full `just test`, `just lint`, full golden update, and planning lint gates;
independently executed the focused checks and probes below. Did not repeat every
repository gate, inject every partial-durability failure, or exhaustively enumerate
Unicode/terminal grammars. The existing Thread ID declaration meaning is preserved,
rather than silently replaced by the source ID.

### Findings

None in the reviewed change. The reproduced legacy-only diagnostic omission remains
the pre-existing issue tracked by `6gjn7axv0gmy`; no new consequence justifying a
scope expansion was demonstrated.

### Verified producer and consumer inventory

| Seam | Verified implementation and evidence |
| --- | --- |
| Real reader to core | `internal/store/fsstore.go:385` decodes task YAML without normalizing planning strings; `internal/core/dependency_graph.go:453` retains the task under the source identity; `cloneTask` at line 593 preserves scalar planning fields. `internal/core/service_thread.go:349` reads the Thread, then loads one task graph. `ShowThread` at line 326 calls `ProjectLoadedThread`. |
| Frontier command and both modes | `internal/cli/thread.go:335`, particularly lines 344–351, uses that single service view. JSON uses `ToThreadFrontierEnvelope` (`internal/wire/thread.go:199`); human uses `ThreadFrontierHuman` (`internal/cli/render/thread.go:203`). No new per-task reads or persistence capability were introduced. Real rebuilt-binary fixtures exercise the complete path. |
| Shared task mapping | `ThreadTaskJSON.Planning` and `ThreadTaskPlanningJSON` are at `internal/wire/thread.go:44` and line 54. `toThreadTaskJSON` at line 130 copies raw priority, tier, effort from `item.Task`, while the compact handle comes from `item.State.TaskID`. `ToThreadViewJSON` at line 97 maps members, external gates, and frontier using this converter. |
| List/show/frontier | `ToThreadsEnvelope` at `internal/wire/thread.go:157`, `ToThreadShowEnvelope` at line 188, and `ToThreadFrontierEnvelope` at line 199 all consume `ToThreadViewJSON`. Actual `thread list/show/frontier --json` specimens preserve non-default and unset data. |
| Graph and plan | `ToThreadGraphProjectionJSON` at `internal/wire/thread.go:256` embeds `ToThreadViewJSON`. Actual `thread graph/plan --json` specimens preserve planning in `projection.view`; graph nodes remain compact and contain no new planning object. |
| Thread mutation receipts | `ToThreadUpdateJSON` at `internal/wire/thread.go:367` maps both before and after. Producer `internal/store/threadmutation.go:45` loads the guarded graph and Thread set; lines 71–72 project from that graph. Actual `thread remove context 6g0000000003 --dry-run --json` carries populated before/after, including raw hostile effort, external planning, and required unset defaults. Error delivery also calls this converter at `internal/cli/exit.go:191`; that path was inspected, not newly fault-injected. |
| Task-impact receipts | `internal/store/lifecyclemutation.go:94` produces `TaskLifecycleThreadImpacts`; `internal/core/thread_mutation.go:408` projects both snapshots. Status replacement preserves the original task at `internal/core/task_lifecycle.go:388`. `internal/wire/dependency.go:264` maps both sides through `ToThreadViewJSON`; line 288 includes these in `ToTaskLifecycleJSON`. Actual direct-member and external-prerequisite start previews populate the optional `moves[].lifecycle.thread_impacts` branch and preserve raw planning on both sides. |
| Dependency repair receipts | `TaskGraphThreadImpacts` at `internal/core/thread_mutation.go:438` compares graph snapshots; `internal/wire/dependency_repair.go:117` uses the same impact converter. Actual explicit-drop preview and committed repair receipts contain populated before/after views and non-default metadata. Ordinary add/remove dependency receipts remain `DependencyMutationJSON` (`internal/wire/dependency.go:329`) without Thread-view fields; they are not an invented consumer of this DTO. |
| Embedded external prerequisites | `ThreadExternalGateJSON` embeds `ThreadTaskJSON` at `internal/wire/thread.go:62`; lines 113–116 map it with the same converter. Real reads, Thread update, external task impacts, and repair output exercise the branch; the emitted schema validates its flattened planning property. |
| Compact/persisted boundaries | `GraphTaskJSON` at `internal/wire/dependency.go:10` still has only task_id/slug/status/epic. The semantic golden comparison found no compact-query drift. `git diff --name-only BASE -- internal/domain internal/store internal/cli/testdata/thread_compatibility 'internal/core/thread_apply*'` was empty: persisted domain, adapters, retained plans, and historical fixtures were not widened. |

All nine named tests in the brief were located in the captured sandbox:
`internal/core/thread_projection_test.go:103`,
`internal/core/service_thread_test.go:610`,
`internal/wire/thread_test.go:15` and `:55`,
`internal/cli/thread_test.go:119`,
`internal/cli/render/render_test.go:964` and `:999`,
`internal/cli/thread_compatibility_test.go:119`, and
`internal/wire/envelopes_test.go:381`. The last broad schema test constructs
empty Thread projections (lines 405–406); its green result alone cannot establish
non-default preservation. The hostile real-output validation below supplies that
missing evidence for this review.

### Independent evidence ledger

All commands ran inside the isolated sandbox. Scratch data, logs, generated
comparison output, Go cache, and the rebuilt binary are retained outside its tree
at `/private/tmp/thread-frontier-review.DNwktF`. Complete command arguments,
stdout/stderr/exit observations are in `transcripts.json`; individual
`*.stdout`/`*.stderr`, the fixture, and `probe.py` are retained there.

Executed successfully:

- `GOCACHE=$SCRATCH/gocache go test -race ./internal/core ./internal/wire ./internal/cli/render ./internal/cli`
  — all four packages passed (`focused-race.log`), including the nine named regressions.
- `GOCACHE=$SCRATCH/gocache go build -o $SCRATCH/tskflwctl ./cmd/tskflwctl`
  — exit 0; all process probes used this rebuilt binary. Go printed a nonfatal
  denial writing optional module stat-cache metadata under the user's read-only
  module cache; the build and binary invocation succeeded. GOCACHE was external.
- `GOCACHE=$SCRATCH/gocache go test ./internal/cli -count=1 -run '^TestThreadPreviewReleaseWireSemanticsRemainCompatible$'`
  — passed (`compatibility.log`). Historical v0.18.0/revision 1.58 and
  v0.19.0/revision 1.60 fixture paths are declared at
  `internal/cli/thread_compatibility_test.go:21` and lines 35–44; none changed.
- `python3 $SCRATCH/probe.py` — final completed run: 55 actual binary invocations,
  all exit 0 and empty stderr; 30 populated schema specimens (`manifest.json`).
  Earlier invalid probe setup (surrogate-escaped YAML, a missing authoring tag,
  and an inappropriate automatic repair selector) was corrected; those aborted
  attempts are not counted as product failures or successful evidence.
- `$SCRATCH/tskflwctl schema --json-schema > $SCRATCH/schema.json`, followed by
  `GOCACHE=$SCRATCH/gocache go run $SCRATCH/validate.go $SCRATCH` — all 30 specimens
  validated using the emitted exact schema ID
  `https://github.com/andy-esch/taskflow/internal/wire/json-envelopes/1.86`
  (`schema-validation.log`). Python assertions independently compare decoded
  raw values and the exact three planning keys, not merely schema acceptance.
  The wire regression also rejects string tier and omitted effort.
- `GOCACHE=$SCRATCH/gocache go run ./internal/tools/docgen -out $SCRATCH/generated-docs`
  and `diff -ruN docs/cli $SCRATCH/generated-docs` — identical.
  `go run ./internal/tools/schemacomments -out $SCRATCH/generated-schema-comments.json`
  with the same GOCACHE, then `diff -u internal/wire/schema_comments.json $SCRATCH/generated-schema-comments.json`
  — identical; 275 generated comments. Compared captured files rather than
  pretending the intentional help diff was already in HEAD.
- `python3 $SCRATCH/goldens.py` — accounts for all 42 changed golden/marker files:
  the revision marker, two new planning fixtures, and 39 existing JSON goldens.
  Removing only revision 1.86 substitutions and the new Thread planning
  property/definition leaves exact decoded equality with BASE; no unrelated
  semantic delta remains (`golden-inventory.log`). The emitted schema also equals
  the captured schema golden as JSON.
- Additional systemic race check:
  `GOCACHE=$SCRATCH/gocache go test -race ./internal/store -count=1 -run '^(TestThreadMembershipWaitsForDependencyMutationAndUsesFreshGraph|TestThreadCompleteWaitsForTaskLifecycleAndRefusesFreshUndrainedState|TestTaskLifecycleWaitsForThreadMembershipAndReportsFreshImpact|TestThreadMutationRejectsRawRepositoryRaces|TestTaskLifecycleRejectsRawThreadRaceByWholeSnapshotCAS|TestMutateTaskLifecycleRejectsRawPrerequisiteRaceByWholeGraphCAS)$'`
  — passed (`systemic-race.log`). These tests were read in the sandbox; they
  coordinate changes under the transaction and assert fresh evidence or CAS refusal.
- The rebuilt sandbox binary also ran `audit lint 6gjnjssabsy2` after report editing
  — exit 0, `all audit findings pass lint`; `git diff --check` is clean.

The initial real fixture has independent next-up/ready candidates with source
IDs `6g0000000001/2/3`, low/tier-5/XS versus high/tier-1/Unknown, and an omitted
planning record. An active member `...004` is blocked by outside task `...005`.
JSON candidate IDs remain exactly `001,002,003`; human output keeps the active
row first and the complete 120-character slug plus ID on its own line. Required
unset JSON is `{"priority":"","tier":0,"effort":""}`; human fallback is `unset`.

### Mutation ledger

Each mutation was compiler-valid, separate from the next, and restored with
sandbox-only `git restore -- <file>` to the helper baseline. No compile failure
was counted as a killed mutation.

| Probe | Exact edit and intended command | Expected / observed / restored |
| --- | --- | --- |
| Required tier preservation | In `internal/wire/thread.go:133`, change `Tier: item.Task.Tier` to `Tier: 0`. Run `go test ./internal/wire -count=1 -run '^(TestThreadViewPlanningContextPreservesRawValuesAndCanonicalIdentity|TestThreadPlanningJSONSchemaAndUnsetRepresentation)$'` with external GOCACHE. | Expected both named tests to reject nonzero-tier loss. Observed exit 1: the first failed across all five read consumers (wanted tier 5, got 0); the second failed for high/Unknown tier 1 and low/XS tier 5 despite schema acceptance of integer zero. After restoration the same command exited 0. Logs: `mutation-tier.log`, `restored-tier.log`. |
| Required sanitizer bypass | Insert `if value != "" { return value }` at the start of `threadPlanningLabel` (`internal/cli/render/thread.go:245`), retaining the existing empty fallback and all compilable imports. Run `go test ./internal/cli/render -count=1 -run '^TestThreadFrontierHumanKeepsContextOnItsOwnSafeLine$'`. | Expected the hostile-text case to fail. Observed exit 1 at `render_test.go:1022` with raw SGR/newline/tab/bell instead of `S with estimate`; empty and Unknown cases do not supply the kill. Restored command exited 0. Logs: `mutation-sanitizer.log`, `restored-sanitizer.log`. |
| Systemic coordinated health challenge | At `internal/core/thread_projection.go:287`, replace the projection-health guard with `if true`; at line 289 replace `member.State.Eligible` with the explicit next-up-or-ready status condition. Run `go test ./internal/core ./internal/cli -count=1 -run '^(TestThreadPlanningContextDoesNotScheduleWork|TestThreadFrontierPlanningContextAcrossRealAdapter)$'`. | The second edit prevents graph-owned false eligibility from accidentally saving the first bypass. Expected dispatch leaks under unhealthy evidence. Observed exit 1: core rejects dispatch under a missing member; real-adapter subtests reject three leaked candidates under both degraded and broken repository graphs. After restoring the one mutated file, both packages passed the same command. Logs: `mutation-health-coordinated.log`, `restored-health-coordinated.log`. |

### Hostile hypotheses and separate systemic pass

| Hypothesis | Disposition and bounded evidence |
| --- | --- |
| Shared/embedded receipt paths lose, default, or terminal-clean metadata while frontier tests pass | **Falsified for exercised paths.** Actual populated Thread update, direct/external task lifecycle, and repair receipts preserve raw control-bearing effort, nonzero tiers, explicit defaults, and external roles. All validate against the exact emitted schema. Producer traces above establish more than a constructed DTO. Partial-cleanup error delivery uses the same converter; exhaustive failure injection remains a stated limit. |
| Required latest-schema planning properties break the additive reader promise | **Falsified.** Revision is explicitly ADDITIVE at `internal/wire/wire.go:365`. ADR-0008 §3 requires unknown-key tolerance and exact-revision schemas; an old strict schema is expressly not a forward validator. New required output keys do not reinterpret existing identity/state fields or change accepted durable plan inputs. Historical subset compatibility passed, and all old golden meanings survive normalization. |
| Planning metadata ranks candidates, promotes active/external work, or obscures identity | **Falsified.** Opposing metadata, equal metadata, reverse file creation, omitted values, and invalid recorded priority/tier all retain stable source-ID order and graph-derived membership. The constructed renderer regression supplies reversed input slices and verifies that the caller's frontier remains unchanged. Real task ID drift retains canonical task handle `...001` with planning while marking graph/projection broken. Thread declaration drift retains its existing declared-ID meaning and produces `thread-id-drift` with an empty frontier. |
| Free-form effort forges rows or controls the terminal | **Falsified for the required corpus through the real YAML reader.** Tested SGR, OSC with BEL and ST terminators, newline plus forged checkmark/ID text, carriage return, tab, bell, whitespace-only text, Unicode `予定 🧭 café`, mixed controls, and 2,700-character effort. Raw JSON matches every input byte-for-byte as decoded strings; plain output has exactly nine lines and no C0 controls except row newlines. Color output differs only by the renderer's own SGR styling. OSC payloads vanish; Unknown stays literal; long text/handles are complete. `render_test.go:999` separately exercises width 40. Existing slug/title/diagnostic rendering is unchanged; this result does not claim universal sanitization of every old human field. |
| Golden regeneration conceals unrelated changes | **Falsified.** All 42 changed golden/marker files are accounted for by a bounded semantic comparison, generated docs/comments match captured files, and historical fixtures remain byte-unchanged relative to BASE. No global regeneration was used during review. |
| Healthy-looking boundary data bypasses dispatch suppression | **Falsified in the demonstrated cases.** Unsorted membership yields healthy graph/broken projection with `invalid-thread-document`; missing member yields healthy graph/broken projection with `missing-thread-member`; missing repository prerequisite yields broken graph/projection with `missing-dependency`; legacy-only data yields degraded graph/projection. Every real frontier is empty. The coordinated mutation demonstrates that the tests detect both health-owner bypasses instead of being saved accidentally by a nearby guard. |
| Defaults/test helpers or state changes conceal systemic loss | **Falsified for the bounded scope.** `internal/testutil/repo.go:67` injects status into fixtures, so the separate process corpus writes explicit frontmatter and uses production opening/reading. The broad schema fixture's empty Thread views were not credited as raw preservation. Real populated optional branches and semantic validators cover those blind spots. Existing guard producers re-read authoritative snapshots at `internal/store/threadmutation.go:89` and `internal/store/lifecyclemutation.go:105`; the selected concurrent/raw-edit tests pass. This change adds no new action authorized from a displayed frontier. |

The legacy omission is **reproduced and confirmed pre-existing**:
`legacy-degraded.json` has `degraded/degraded`, empty frontier and empty cause
arrays. Comparing `internal/core/thread_projection.go` with BASE after removing
its two new documentation lines yields exact equality; dependency graph/source
implementation is also unchanged. `projectThread` still copies `graph.Problems()`
at line 192, whereas legacy causes live in `LegacyDiagnostics()`
(`internal/core/dependency_graph.go:848`). The planning follow-up is future work,
not shipped diagnostic evidence. No new metadata consequence was found.

No changed repair-advice branch exists in this implementation. As an additional
boundary check, diagnosed a real dangling declaration, previewed and applied its
explicit drop selector, inspected the committed repair receipt, and reread the
healthy frontier (`repair-diagnose-human.stdout`, `receipt-repair*.json`,
`post-repair-frontier.json`). Automatic repair correctly did not guess a
dangling-reference intent; that preliminary no-op was not counted as a repair.

There are no surviving reproduced in-scope defect hypotheses. Remaining limits
are exhaustive terminal/Unicode coverage and unexercised partial-cleanup error
variants; these are coverage limits, not asserted product defects or preferences.

### Isolation, restoration, and guarded delivery attestation

- Source: `/Users/andyeschbacher/git/andy-esch/taskflow`, captured HEAD/base
  `9bc8d72a9d3adce4b58ebc63eaf5b0bec9cc1b17`.
- Sandbox: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.etFyxS`.
- Resolved independent Git directory: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.etFyxS/.git`.
- Helper baseline commit: `4add81a1e126c9ee3f55a33f1f76e07896d54935`.
- Captured source fingerprint: `c3b418479a46d3af78ef81389000c8fc83d93c3a`.
- Captured assigned-audit source blob: `a0b9229bba1ed711752c14cc5e4028c6002f1094`.
- Deliverable: `planning/audits/6gjnjssabsy2-2026-10-11-thread-frontier-planning-context-implementation-codex.md`.
- All mutation files restored to the sandbox-only baseline; before report editing
  `git status --short` was empty. Caches/probe files stayed outside the sandbox tree.
  Final verification allows only this unstaged audit delta; no extra commits or staging.
- Guarded verification/transfer commands:
  `"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"`,
  `git diff -- "$AUDIT_REL"`, then
  `"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"`.
- Delivery result for the received report: **`transfer=succeeded`**, helper exit 0;
  exact helper output retained at `/private/tmp/thread-frontier-review.DNwktF/transfer.log` and verification at
  `/private/tmp/thread-frontier-review.DNwktF/verify.log`. This delivery declaration is effective only if
  the helper succeeds; an untransferred sandbox copy alone does not attest success.
  The helper checks independent Git metadata, unchanged baseline/no staging, one
  permitted delta, source-audit blob equality, atomic replacement, and final byte equality.
- Sandbox and scratch evidence retained for owner-confirmed receipt; no manual source copy.

### Owner disposition (2026-10-11)

Accepted the bounded review after checking the guarded transfer/verification logs and supporting
schema/golden inventories. No findings require implementation changes. The legacy-only omission
is already tracked by `6gjn7axv0gmy`; its pre-existing behavior does not block this metadata slice.

Promoted receipt before/after metadata preservation and additional OSC/multiline/Unicode probes
into permanent tests. These test-only additions do not change the reviewed production snapshot.
Full race tests, lint, and build pass after the additions. The independent report above remains
the reviewer’s historical account, including the open bucket at delivery; owner closure follows.
