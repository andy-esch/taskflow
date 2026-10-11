---
schema: 1
id: 6gjqmf7c7jhj
bucket: closed
area: thread-legacy-degradation-diagnostics-implementation-codex
date: "2026-10-11"
updated_at: "2026-10-11"
---
# Audit: Thread legacy degradation diagnostics implementation — codex — 2026-10-11

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

Challenge whether legacy dependency degradation is now explainable everywhere a Thread read
claims graph health, without weakening eligibility, mutation authorization, or portable identity.
This is a bounded diagnostic followup, not a dependency-policy or graph-repair redesign.

Codex: primary lens is core ownership, occurrence identity, list snapshots/cursors, and receipt
consistency; cross-check one real CLI diagnosis and migration. Antigravity: primary lens is
operator guidance, emitted JSON/schema, stdout/stderr separation, and TUI comprehension;
cross-check occurrence grouping and cursor invalidation. Both own the shared invariants below.
Review independently; do not read the sibling report before finalizing your own.

## Review target

Source: `/Users/andyeschbacher/git/andy-esch/taskflow`, branch
`fix/thread-legacy-degradation-diagnostics`, base `b090275`. Review the entire captured dirty
implementation, including untracked tests and new goldens, not HEAD alone.

Implementing task: [Explain legacy dependency degradation](../tasks/6gjn7axv0gmy-explain-legacy-dependency-degradation-in-thread-projections.md).
Thread: [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md).
Contract: [ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md).

Verify a consumer inventory and trace actual producers, not merely constructed values:

- `internal/core/dependency_graph.go`: `LegacyDiagnostics`, `MigrationGuidance`, `Summary`,
  `ExplainsGraphProblem`; `thread_projection.go`: `projectThread` and `ThreadListView`.
- `service_thread.go`: `ListThreadViews`; `service_list_query.go`: `QueryThreadViews` cursor
  diagnostic digest; `thread_mutation.go`: `cloneThreadView`.
- `internal/wire/thread.go`: `ToThreadViewJSON`, `ToThreadsEnvelope`; `dependency.go`:
  `ToLegacyDependenciesJSON`; revision 1.87 in `wire.go`. Existing mutation/repair impact
  conversions also consume Thread views.
- `internal/cli/render/thread.go`, `dependency.go`, `list_query.go`; TUI `detail.go`,
  `thread_projection.go`, `view.go`. Compare summary/rank diagnostics and list warning cues.
- New `thread_legacy_test.go` in core, wire, cli, and tui; render dependency tests; new
  `thread_frontier_legacy_{human,json}.golden` and refreshed machine-contract goldens.

Non-goals: storage changes, automatic migration, graph-health/eligibility changes, new graph
navigation or immersive-layout redesign. The old unbounded `--json -c` projection envelope
deliberately remains unchanged; opt-in query pages and full typed lists carry hoisted context.
Challenge that compatibility boundary only with a concrete consequence, not a preference.

## Intended contract to challenge

- Safely resolved and present-empty legacy fields outside membership still degrade the graph,
  suppress dispatch, and explain the cause; they do not become external prerequisites.
- Diagnostics come from the same immutable graph snapshot. Lists carry one repository-wide
  set even when filtering removes every row; outside-context changes invalidate continuations.
- Source identity is authoritative. Optional local paths and opaque locations are distinct;
  duplicate source-ID occurrences stay attributable. No adapter rescans or location-based writes.
- Missing/ambiguous legacy references remain broken. Grouping hides only their corresponding
  raw human/TUI messages, never structural or sibling-occurrence defects. Machine graph problems
  remain complete alongside typed legacy references.
- Guidance names a workflow, not whole-repository mutation permission. Migration and repair
  revalidate their existing preconditions. Diagnostic reads and refused/preview actions do not write.
- New JSON arrays are explicit empty arrays when healthy; non-default opaque locations and
  references survive reads and receipts and validate against the actual emitted schema.
  No persisted document/plan format or historical compatibility fixture is redefined.

## Mandatory evidence floor

Run the focused regressions below, then produce hostile evidence through real callers:

1. In an outside-Thread resolved/empty legacy fixture, inspect frontier/list output and prove
   that no dispatchable row appears. Execute the advertised `task depend migrate` as preview
   then commit; compare source snapshots and the reread frontier. Separately exercise empty
   legacy cleanup with `task depend repair --auto`, checking before/after Thread receipt causes.
2. Use missing and ambiguous references plus an independent structural defect. Prove the
   defect survives display grouping, typed JSON retains every problem, and migration refuses
   without source changes. Check a pathless or duplicate-occurrence producer, not just a struct.
3. Populate optional `location` and reference candidate/edge values; inspect decoded semantics
   and validate with `schema --json-schema`. Check empty selected pages and at least one row
   stream's stdout/stderr separation. Inspect TUI summary/ranks and list warning cues.

Execute at least two selected compiler-valid mutations, restore each, and rerun green:

- Omit `GraphLegacy` from `projectThread`: target
  `TestThreadLegacyDiagnosticsRetainRepositoryScopeAndOwnTheirValues` or the real CLI test.
- Omit `GraphLegacy` from the `QueryThreadViews` diagnostic digest: target
  `TestThreadListLegacyDiagnosticsHoistAndInvalidateContinuation`; hold health and row values
  constant while changing only an outside opaque location.
- Loosen location matching in `ExplainsGraphProblem`: target
  `TestThreadLegacyDiagnosticsDistinguishDuplicateSourceOccurrences` and
  `TestThreadLegacyDiagnosticsPreserveBrokenSeverityAndGrouping`.

Codex must include the cursor mutation; Antigravity must include the grouping mutation.
A compile failure is invalid evidence, not a killed mutation. Record exact edits, intended
tests, observed results, and restored results. No quota of findings; a clean verdict needs
these falsification attempts, not a green baseline alone.

## Required hostile angles

- Hypothesis: global context is lost or repeated in lists, including zero rows and paginated
  projections. Compare full typed output, opt-in columns, and streaming output contracts.
- Hypothesis: field-level guidance misrepresents whole-graph readiness when other defects
  coexist. Diagnose and try the command; do not treat per-field readiness as permission.
- Hypothesis: duplicate IDs or pathless sources make grouped causes hide another occurrence,
  or mutable reference/candidate slices let receipts corrupt the shared graph snapshot.
- Hypothesis: location/control text forges a terminal diagnostic row or core values are
  sanitized before reaching JSON. Probe actual renderers and preserve raw machine evidence.
- Hypothesis: migration/repair receipts disagree with subsequent reads because an adjacent
  converter silently omits the new context. Trace and compare both production paths.
- Second systemic pass: inspect adjacent dependency-query renderers, shared source attribution,
  and compatibility/schema helpers. Separate a reachable product defect from test fragility,
  coverage-only gaps, preferences, and already excluded design work.

## Validation and restoration

Implementation-owner baseline: full `just test` (race), `just lint`, `just build`,
`just docs-check`, schema-comment generation, and planning lint passed. You may inherit
those full gates; independently run focused tests and the required hostile probes:

```sh
GOCACHE=/tmp/taskflow-legacy-review-cache go test -race \
  ./internal/core ./internal/wire ./internal/cli/render ./internal/cli ./internal/tui \
  -run 'ThreadLegacy|ThreadListLegacy|ThreadGraphRepairImpactsCarryLegacy|GraphLegacy'
```

Build your sandbox's binary for real-process probes; do not use an installed/stale binary.
Put fixtures, caches, and generated probe binaries outside the captured sandbox tree.
Restore every mutation to the helper's sandbox baseline, leaving only your assigned audit
changed. Run `audit lint` for that audit, `git diff --check`, and the prescribed sandbox
verify/guarded transfer protocol. Do not manually copy files into the shared checkout.

## Deliverable

Preserve this brief; fill the report slot. Give verdict and limits, verified consumer inventory,
findings by root cause with reachable consequence and reproduction, concise probe/mutation
ledger, and a bounded surviving-hypothesis ledger. Use tool-owned finding creation where
available and leave every finding open for owner triage. Distinguish product defects from
coverage gaps and defense-in-depth. Include isolation attestation, captured snapshot/baseline,
checks run versus inherited, and guarded transfer result. An incomplete review says so.

## Reviewer report

### Verdict and scope

**Review complete; one open medium finding (M1) for owner triage.** The selected core
ownership, occurrence, cursor, eligibility, and receipt invariants survived the required
hostile probes. No newly introduced core/cursor/mutation defect was demonstrated. M1 is a
reachable shared-renderer defect, also present at `b090275`, found through the explicitly
requested terminal-boundary and adjacent dependency-query second pass. It is not claimed
as a regression introduced by this patch.

Reviewed the entire captured implementation against `b090275`, including source-untracked
legacy tests and both new frontier goldens. Their presence in the sandbox checkpoint is
review evidence, not a claim that the implementation owner has committed them. Planning
acceptance marks and the owner's full-gate checkpoint were treated as requirements and
inherited claims; the concrete checks below were executed independently. The sibling audit
was not read. No implementation changes are delivered.

### Verified consumer and producer inventory

All paths/lines refer to the captured sandbox baseline. This inventory was checked in code
and through the named regressions or process probes, rather than inferred from planning prose.

| Boundary | Verified production path and evidence |
| --- | --- |
| Core occurrence producer | `internal/core/dependency_graph.go:691` (`resolveLegacyDiagnostics`) takes the source ID, local path, and opaque location from each record at line 728. `LegacyDiagnostics` at line 901 uses `cloneLegacyDiagnostics` at line 905, including independent reference and candidate-ID slices and stripped snapshot-local handles. |
| Core wording/grouping | `MigrationGuidance` at `internal/core/dependency_graph.go:114`, `Summary` at line 123, and `ExplainsGraphProblem` at line 143; the latter matches ID, field, path, and location and accepts only missing/ambiguous codes. `sameLegacyDiagnostic` at line 1060 includes location. |
| Shared projection | `ThreadView` / `ThreadListView` at `internal/core/thread_projection.go:65` / line 83; `projectThread` at line 165 obtains graph-owned legacy context at line 195. The outside task is not a membership or external-gate producer. |
| List snapshot/cursor | `ListThreadViews` at `internal/core/service_thread.go:210` reads Threads and the graph once, hoists context at line 242, and clears it from rows at line 246. `QueryThreadViews` at `internal/core/service_list_query.go:97` binds `GraphLegacy` into the diagnostic digest at lines 122–123 before selection/pagination. |
| Receipt ownership | `cloneThreadView` at `internal/core/thread_mutation.go:516` clones legacy context at lines 522–523. `TaskGraphThreadImpacts` at line 438 projects before/after graphs. `internal/core/dependency_repair.go:675` invokes it and line 685 clones the impact receipt. Real empty-field cleanup exercised this path. |
| Shared wire | `ToLegacyDependenciesJSON` at `internal/wire/dependency.go:97` retains distinct optional location and explicit reference/candidate arrays. `ToThreadViewJSON` at `internal/wire/thread.go:98` uses it at line 110; `ToThreadsEnvelope` at line 160 hoists it at line 163. Show/frontier/graph/plan converters at lines 193/204/262 reuse the view conversion. Update receipts at line 376 and repair-impact conversion at `internal/wire/dependency.go:270` also reuse it. |
| CLI lists and reads | `internal/cli/thread.go:223` calls `QueryThreadViews`; lines 237/269 select projected pages and stderr diagnostics. `internal/cli/render/thread.go:44`, line 120, and line 385 share global/grouped diagnostics; `internal/cli/render/list_query.go:49` includes `graph_legacy` at line 55 only for opt-in query context. The old unbounded column envelope remains as specified. |
| Adjacent dependency queries | `TaskBlockers` / `TaskUnblocks` at `internal/core/dependency_operations.go:400` / line 443 acquire graph problems and legacy context from the same graph. Their human renderers at `internal/cli/render/dependency.go:263` / line 291 share `graphDiagnosticsHuman` at line 323. M1 affects that shared boundary. |
| TUI | `internal/tui/thread_projection.go:50` carries hoisted legacy context into the registry. `threadDiagnosticLines` at `internal/tui/detail.go:1350` feeds summary/ranks at lines 1345/1482. Empty-list diagnostics and footer cues are at `internal/tui/view.go:221` / line 535. Both new TUI tests exercised these actual registry/detail paths. |
| Machine contract/fixtures | `internal/wire/wire.go:369` classifies revision 1.87 as ADDITIVE; the constant is at line 373. `internal/cli/testdata/golden/schema_jsonschema.golden:1859`, line 5270, and line 5347 define the legacy/view/list shapes. Both `thread_frontier_legacy_human.golden` and `thread_frontier_legacy_json.golden` exist and were exercised. A semantic comparison of 39 existing changed JSON goldens found only revision/new `graph_legacy` changes; the new legacy JSON golden was identified separately. |

The Unix mutation guard is the **repository root directory itself**, opened and flocked in
`internal/store/lock_unix.go:21`; there is no asserted lock-file path. Mutation authorization
and guard entry are in `internal/store/lock.go:115`. Migration rereads the authoritative
snapshot under that guard (`internal/store/graphmutation.go:53`) and rechecks the full
snapshot before writes at line 83. Repair checks graph/Thread source snapshots before writes
at `internal/store/graphrepair.go:95` and revalidates each next step at line 132. These are
existing boundaries; no new mutation permission is inferred from diagnostic guidance.

### Executed probe ledger

Process binary: built this sandbox with
`GOCACHE=/tmp/taskflow-legacy-review-cache go build -buildvcs=false -o /tmp/taskflow-legacy-codex-review ./cmd/tskflwctl`.
All process fixtures, output captures, scripts, overlays, and the binary remain outside the
captured tree. Main output directory: `/tmp/taskflow-legacy-codex-evidence-w31mz8kj`.

| Probe | Observed result |
| --- | --- |
| Resolved and present-empty legacy fields outside membership | `/tmp/taskflow-legacy-real-probes.py`: degraded frontier, no dispatchable row, two repository diagnostics, no external gate. Full typed, limited, projected query, and empty selected lists retain one hoisted set; rows carry empty legacy arrays. Diagnostic reads left all source files byte-identical. |
| Advertised migration | Same script: `task depend migrate --dry-run --json` left source bytes unchanged; commit removed both legacy keys and reread produced healthy graph, empty legacy array, and one eligible independent member. |
| Empty-field repair/receipts | `task depend repair --auto --dry-run --json` then commit: before impact contains outside `blocks` cause and empty frontier; after contains no cause and one member. Global health changes degraded → healthy, indirect impact remains indirect, and durable reread agrees with the receipt. Preview wrote nothing. |
| Missing + ambiguous + structural defect | Actual filesystem fixture with `blocked_by: [gone, shared]`, two exact `shared` slugs, and invalid canonical dependency: raw missing/ambiguous display rows are grouped, structural row remains, and typed JSON contains all three problems. Migration exits 11 without any source change. `task depend repair` diagnoses choices without writing. |
| Guidance versus action/stale state | A resolved field plus independent invalid dependency still gives field-level migration guidance; migration refuses with exit 11. Separately, after a degraded diagnosis, adding an outside structural defect causes the later migration to revalidate/refuse. Changing a diagnosed empty field to `blocks: [gone]` before `repair --auto` leaves the source intact and reports broken residual evidence. |
| Stream/compatibility | Empty opt-in CSV query: source rows/header only on stdout, legacy diagnosis on stderr. Existing CLI regression also covers name/table/CSV modes. Unbounded `thread list --json -c id` retains its old envelope without `graph_legacy`; this is the brief's intentional compatibility boundary. |
| Pathless/duplicate sources and aliases | Committed-in-sandbox core tests produce pathless sources, source-ID drift, and duplicate-ID records with different opaque locations. Both missing-reference occurrences retain exactly one grouped owner; structural duplicate evidence survives. Projection/receipt candidate-ID mutation cannot alter the graph or the original view. These are graph producers, not merely prefilled diagnostics. |
| Optional wire semantics/schema | Review-only `TestCodexLegacyProducedOptionalSemanticsAndActualSchema` uses `NewTaskGraphRead`, `ProjectThread`, and `TaskGraphThreadImpacts` to produce raw newline/ANSI location, resolved edges, ambiguous candidate IDs, and missing references. Decoded values survive all nine read/list/update/repair/dependency-query converter shapes. Integer replacements for candidate/edge fields fail schema validation. Nine real-process outputs also validate using the actual binary's `schema --json-schema` resource and the relevant envelope `$defs`. |
| TUI comprehension | `TestThreadLegacyDiagnosticsReachTUIRegistryAndDetail` verifies actual registry load, summary/rank diagnosis once, migration footer, empty list cause, and no dispatch. `TestThreadLegacyTUIDeduplicatesGroupedProblemsAndSanitizesContext` retains structural defects and terminal-safe grouped causes. |
| Terminal boundary/systemic second pass | A real local duplicate-ID filename containing newline plus ESC `[2J` emits a forged row and screen-clear sequence in frontier/blockers/unblocks. A separate pathless graph producer with missing legacy plus independent invalid dependency reproduces raw opaque-location injection. Legacy rows are safe; the adjacent structural rows bypass sanitization. See M1. |

The optional-wire overlay is `/tmp/taskflow-legacy-wire-overlay.json`, backed by
`/tmp/taskflow-legacy-wire-hostile_test.go`; command:
`GOCACHE=/tmp/taskflow-legacy-review-cache go test -overlay=/tmp/taskflow-legacy-wire-overlay.json ./internal/wire -run '^TestCodexLegacyProducedOptionalSemanticsAndActualSchema$' -count=1` — PASS.
The emitted schema is a registry of concrete envelope definitions; validation selected the
actual envelope definition, not the registry-root object. Initial harness assumptions about
an envelope name and registry-root validation were corrected against shipped definitions;
those harness failures are not findings or mutation kills.

### Compiler-valid mutation ledger

Each code edit was made only in the sandbox, targeted tests compiled and ran, and restoration
used `git restore --source=HEAD -- <exact mutated file>` against the captured checkpoint.
No compile failure was counted as a killed mutation.

| Exact edit | Intended test / hostile evidence | Mutated result | Restored result |
| --- | --- | --- | --- |
| In `internal/core/service_list_query.go`, remove the `GraphLegacy []LegacyDependencyDiagnostic` digest member **and** its corresponding positional literal argument. | `TestThreadListLegacyDiagnosticsHoistAndInvalidateContinuation` at `internal/core/thread_legacy_test.go:148`; row values and degraded health stay constant while only outside opaque location changes. | FAIL, exit 1: `changed outside diagnosis permitted stale continuation: <nil>`. | PASS, exit 0. |
| Remove `view.GraphLegacy = graph.LegacyDiagnostics()` from `projectThread`. | `TestThreadLegacyDiagnosticsRetainRepositoryScopeAndOwnTheirValues` at `internal/core/thread_legacy_test.go:14`. | FAIL, exit 1: `legacy-only cause was lost or dispatch changed`, with empty `GraphLegacy`. | PASS, exit 0. |
| Remove only `|| d.Location != problem.Location` from the grouping ownership condition. | `TestThreadLegacyDiagnosticsPreserveBrokenSeverityAndGrouping` / `TestThreadLegacyDiagnosticsDistinguishDuplicateSourceOccurrences` at `internal/core/thread_legacy_test.go:70` / line 98. | Both FAIL, exit 1: other-occurrence grouping accepted; missing reference has two owners. | Both PASS, exit 0. |

Additional coordinated remediation probe for M1: sanitize **both** `problem.Message` and
`location` at both structural print branches using existing `singleLineText`. The produced
frontier/blockers/unblocks hostile test then passed all three subtests. This probe was
restored; no fix is included. Sanitizing location alone would leave the real local-path
message attack intact. After restoration, the entire prescribed focused race suite passed
again with `-count=1`.

### Bounded surviving-hypothesis ledger

| Challenged hypothesis | Disposition and limits |
| --- | --- |
| Hoisted context disappears on zero rows or is repeated per row | Falsified for tested full typed, empty selected, limited, and opt-in projected pages and streams. Old unbounded projected JSON intentionally excludes it. |
| Outside-context drift preserves stale cursor accidentally | Falsified by real service location-only change and compiler-valid digest mutation; the row and health cannot accidentally provide the invalidation. |
| Guidance grants whole-graph readiness or mutation permission | Falsified by coexisting-defect and between-read/action state changes. Existing guard/source preconditions remain decisive; no automatic migration was observed. |
| Grouping hides structural/sibling causes or receipt slices corrupt graph | Falsified for produced pathless/different-location duplicate occurrences and nested reference/candidate mutation. Identical ID/path/location tuples cannot present distinct opaque labels; no loss of reference content was demonstrated, so no speculative finding was added. |
| Optional machine values are omitted/sanitized or permissive schema tests mask them | Falsified with non-default produced values, decoded edge/candidate assertions, typed negative controls, and actual emitted schema definitions. Pathless mutation-store behavior was not newly implemented here; receipt converters/impact producer were exercised separately from real local repair actions. |
| Terminal-safe legacy grouping makes the shared terminal boundary safe | **Survives with demonstrated product defect M1**, across three public surfaces and two different source-attribution paths. Existing new legacy-only renderer test cannot detect coexisting structural branch injection. |

### Checks, isolation, and transfer attestation

Independently executed: required focused race suite twice (five packages PASS); all three
selected mutation tests and restored reruns; targeted wire/CLI Thread and machine-contract
regressions (PASS); sandbox binary build; real process probes; produced optional-wire semantic
validation; M1's hostile failing test and temporary coordinated passing probe; assigned
`audit lint`; `git diff --check`; sandbox helper verification and guarded transfer.
Full `just test`, lint, build, docs-check, schema-comment generation, and planning lint are
**inherited from the implementation-owner checkpoint**, not claimed as independently rerun.
No live TUI screenshot or exhaustive repository-wide mutation campaign was performed.

Implementation inspection, builds, tests, probes, and report edits all took place after the
independent copy and only inside its workspace. Initial source metadata discovery was
read-only. No source branch switch, staging, commit, reset, stash, or implementation write
was performed. The helper created the sole sandbox checkpoint; every probe was restored.
Only the assigned audit differs at final verification.

Helper-owned attestation:

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.0iJ8Wi
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.0iJ8Wi/.git
baseline_commit=dd75ec63e2e6ab8aa00d198317a53940db6c56c8
source_blob=cc1383a7c022a25df3b2bef607841a84cec55431
source_fingerprint=1b75c5c396aafa66dbf9cc8b54d9e60f968eb856
deliverable=planning/audits/6gjqmf7c7jhj-2026-10-11-thread-legacy-degradation-diagnostics-implementation-codex.md
deliverable_changed=true
transfer=succeeded
```

The final helper stdout is retained at `/tmp/taskflow-legacy-codex-transfer.txt`. The workspace
is retained for owner receipt; nothing besides the assigned audit is transferred. If the
final guarded operation refuses, this attestation must be corrected in the retained sandbox
and the report is incomplete until transfer succeeds.

## Findings

#### M1. Structural diagnostics bypass terminal sanitization · **Status:** fixed

**File:** internal/cli/render/dependency.go:346 | **Component:** cli/render
**Effort:** S · **Urgency:** soon

`graphDiagnosticsHuman` sanitizes `LegacyDependencyDiagnostic.Summary()` and its opaque location, but prints structural `GraphProblem.Message` and `Location` verbatim at `internal/cli/render/dependency.go:346` and line 348. Structural defects intentionally survive grouping, so a coexisting defect reopens the same terminal boundary. A newline can forge an extra diagnostic row; ESC `[2J` can clear the terminal screen and obscure the actual broken-health explanation. Eligibility remains fail-closed; this finding concerns diagnostic integrity.

This is reachable with the shipped filesystem adapter: create duplicate source ID `6g0000000002`, with one task filename `6g0000000002-outside\nFORGED-READY\x1b[2J.md` and a legacy field, then run `thread frontier delivery`, `task blockers member`, or `task unblocks member`. The duplicate diagnostic producer at `internal/core/dependency_graph.go:501` embeds unescaped source paths in its message. Every command exits 0 and contains the actual extra line and ESC sequence. Machine JSON retains the raw evidence and the frontier stays empty.

Reproduction: `python3 /tmp/taskflow-legacy-terminal-local.py` creates independent fixtures and runs all three real-process commands using the sandbox-built binary. Observed fixture: `/tmp/taskflow-legacy-terminal-rsi2qoab`. A second producer-based test uses `NewTaskGraphRead` with pathless source location `db://outside\nFORGED-READY\x1b[2J`, missing legacy intent, and an independent invalid dependency; all three renderer subtests fail. Run `GOCACHE=/tmp/taskflow-legacy-review-cache go test -overlay=/tmp/taskflow-legacy-render-overlay.json ./internal/cli/render -run '^TestCodexProducedStructuralContextCannotForgeTerminalRow$' -count=1 -v`. Its source is `/tmp/taskflow-legacy-render-hostile_test.go`.

A temporary coordinated probe that applies existing `singleLineText` to both structural messages and opaque locations makes all three hostile renderer subtests pass. Sanitizing only the location leaves the local-file message attack. The probe was restored. Add a regression from an actual graph producer with legacy and structural evidence coexisting; the new `TestGraphLegacyDiagnosticsHumanKeepOpaqueContextOnOneSafeLine` at `internal/cli/render/dependency_test.go:126` passes only legacy rows and therefore misses this shared branch.

`git show b090275:internal/cli/render/dependency.go` confirms the raw structural print branches predate this patch. This is an adjacent existing product defect included because the brief explicitly requires challenging terminal-control text and shared dependency-query abstractions; it is not presented as a newly introduced regression or a coverage-only issue.

**Recommendation:** Apply terminal-only single-line sanitization to every structural message and opaque location at the shared renderer boundary; preserve raw core and JSON values.

**Resolution:** Confirmed the pre-existing structural-renderer defect. Apply
    singleLineText to both GraphProblem.Message and opaque location at the
    shared graphDiagnosticsHuman boundary, preserving raw deduplication keys and
    core/wire values.
    TestGraphDiagnosticsHumanKeepProducedStructuralContextOnOneSafeLine produces
    pathless and local duplicate-source graphs with coexisting legacy evidence;
    all frontier/blockers/unblocks subtests fail before the fix and pass
    afterward under race testing. Decoded JSON retains the raw source context.
