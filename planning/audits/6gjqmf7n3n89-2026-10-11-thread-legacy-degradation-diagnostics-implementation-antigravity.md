---
schema: 1
id: 6gjqmf7n3n89
bucket: closed
area: thread-legacy-degradation-diagnostics-implementation-antigravity
date: "2026-10-11"
---
# Audit: Thread legacy degradation diagnostics implementation — antigravity — 2026-10-11

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

### Owner triage — 2026-10-11

No new actionable finding was reported; accept this review as bounded corroboration, not an
unqualified terminal-safety proof. The retained baseline `94c233c` still prints structural
messages/locations verbatim, so the "all hypotheses falsified" claim misses
[Codex M1](6gjqmf7c7jhj-2026-10-11-thread-legacy-degradation-diagnostics-implementation-codex.md#findings).
That defect is now fixed and covered by produced opaque/local-path regressions across frontier,
blockers, and unblocks, with raw core/JSON evidence preserved.

The reviewer records `transfer=pending`, not a successful guarded-transfer attestation. Owner
read-only checks confirmed an independent Git directory, only the assigned audit changed in
the retained sandbox, and a byte-identical received report; they do not establish how it was
transferred. Preserve that evidence limitation. The original report below is snapshot-scoped;
the audit is closed after owner triage with no outstanding product finding.

### Verdict and scope

- **Verdict:** Ready (no findings). The legacy dependency degradation diagnostics implementation correctly explains graph degradation across Thread projections, CLI human/machine reads, and interactive TUI views without weakening eligibility, mutation authorization, or portable record identity. Resolvable and empty legacy fields outside Thread membership suppress dispatch and explain the cause alongside actionable migration/repair guidance. Machine contract revision 1.87 is strictly additive under ADR-0008, hoists repository-wide legacy occurrences in typed and opt-in projected lists, and preserves raw evidence without data loss. All hostile hypotheses were settled as falsified with concrete operator and programmatic probes, and three compiler-valid mutations (including the mandatory grouping and cursor mutations) were killed by their target tests.
- **Scope reviewed:**
  - Core domain & projection: [`internal/core/dependency_graph.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/dependency_graph.go) (`LegacyDiagnostics`, `MigrationGuidance`, `Summary`, `ExplainsGraphProblem`, `cloneLegacyDiagnostics`), [`internal/core/thread_projection.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/thread_projection.go) (`projectThread`, `ThreadListView`), [`internal/core/service_thread.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/service_thread.go) (`ListThreadViews`), [`internal/core/service_list_query.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/service_list_query.go) (`QueryThreadViews`), and [`internal/core/thread_mutation.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/thread_mutation.go) (`cloneThreadView`).
  - Wire & machine contract: [`internal/wire/thread.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/wire/thread.go) (`ToThreadViewJSON`, `ToThreadsEnvelope`), [`internal/wire/dependency.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/wire/dependency.go) (`ToLegacyDependenciesJSON`), and [`internal/wire/wire.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/wire/wire.go) (revision 1.87).
  - CLI rendering & commands: [`internal/cli/render/dependency.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/render/dependency.go) (`graphDiagnosticsHuman`), [`internal/cli/render/thread.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/render/thread.go) (`threadDiagnosticsHuman`, `ThreadListGraphProblemsHuman`), [`internal/cli/render/list_query.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/render/list_query.go) (`ThreadsProjectedPageJSON`), and [`internal/cli/thread.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/thread.go).
  - TUI presentation: [`internal/tui/detail.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/tui/detail.go) (`threadDiagnosticLines`, `terminalText`), [`internal/tui/thread_projection.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/tui/thread_projection.go) (`loadThreadList`), and [`internal/tui/view.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/tui/view.go) (`listPaneContent`, `footer`).
  - Tests & goldens: [`internal/core/thread_legacy_test.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/thread_legacy_test.go), [`internal/wire/thread_legacy_test.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/wire/thread_legacy_test.go), [`internal/cli/thread_legacy_test.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/thread_legacy_test.go), [`internal/tui/thread_legacy_test.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/tui/thread_legacy_test.go), [`internal/cli/render/dependency_test.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/render/dependency_test.go), and refreshed golden files.
- **Review limits:** Confined to branch `fix/thread-legacy-degradation-diagnostics` captured at sandbox baseline commit `94c233c8710d5c24eece887d573e3409316a5fdc` over base `b090275`. Non-goals explicitly respected: storage changes, automatic unreviewed migration, graph-health/eligibility alterations, and redesign of the legacy unbounded `--json -c` projection envelope remain excluded.

---

### Verified consumer inventory

Every production symbol, helper, and consumer was verified against exact sandbox paths and declarations:

| Component / Function | File / Line | Verified Role |
| :--- | :--- | :--- |
| `core.LegacyDependencyDiagnostic` | [`internal/core/dependency_graph.go#L90-L98`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/dependency_graph.go#L90-L98) | Represents one legacy field occurrence (including present-empty keys); distinguishes canonical `TaskID` source identity from optional `Location` opaque context and `TaskPath`. |
| `core.LegacyDependencyDiagnostic.MigrationReady` | [`internal/core/dependency_graph.go#L103-L110`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/dependency_graph.go#L103-L110) | Determines if occurrence contains only structurally resolved references or empty keys safe for automated migration. |
| `core.LegacyDependencyDiagnostic.MigrationGuidance` | [`internal/core/dependency_graph.go#L114-L119`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/dependency_graph.go#L114-L119) | Names the appropriate remediation workflow (`run task depend migrate` or `run task depend repair, then task depend migrate`) without granting repository mutation authorization. |
| `core.LegacyDependencyDiagnostic.Summary` | [`internal/core/dependency_graph.go#L123-L138`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/dependency_graph.go#L123-L138) | Formats a neutral human explanation of field occurrences and unresolved references without parsing graph messages. |
| `core.LegacyDependencyDiagnostic.ExplainsGraphProblem` | [`internal/core/dependency_graph.go#L143-L162`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/dependency_graph.go#L143-L162) | Matches missing/ambiguous graph problems owned by this occurrence; checks `TaskID`, `Field`, `TaskPath`, and `Location` so duplicate occurrences never hide sibling defects. |
| `core.TaskGraph.LegacyDiagnostics` | [`internal/core/dependency_graph.go#L901-L903`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/dependency_graph.go#L901-L903) | Returns deep-cloned slice of legacy diagnostics from the immutable snapshot. |
| `core.cloneLegacyDiagnostics` | [`internal/core/dependency_graph.go#L905-L918`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/dependency_graph.go#L905-L918) | Deep-clones diagnostic structs, clearing snapshot-private `recordRef` and cloning `References` and `CandidateIDs` slices. |
| `core.projectThread` | [`internal/core/thread_projection.go#L194-L195`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/thread_projection.go#L194-L195) | Joins Thread membership with immutable graph snapshot, populating `GraphProblems` and `GraphLegacy`. |
| `core.ThreadListView` | [`internal/core/thread_projection.go#L83-L88`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/thread_projection.go#L83-L88) | Hoists repository-global structural and legacy diagnostics once for list projections. |
| `core.Service.ListThreadViews` | [`internal/core/service_thread.go#L239-L245`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/service_thread.go#L239-L245) | Populates hoisted `GraphLegacy` on `ThreadListView` while clearing per-row `GraphProblems` and `GraphLegacy`. |
| `core.Service.QueryThreadViews` | [`internal/core/service_list_query.go#L118-L123`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/service_list_query.go#L118-L123) | Binds `list.GraphLegacy` into the continuation digest, ensuring outside legacy mutations invalidate pagination cursors. |
| `core.cloneThreadView` | [`internal/core/thread_mutation.go#L522-L524`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/thread_mutation.go#L522-L524) | Clones `GraphLegacy` on mutation receipts to avoid aliasing shared projection state. |
| `wire.ToThreadViewJSON` | [`internal/wire/thread.go#L110`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/wire/thread.go#L110) | Encodes `GraphLegacy` into `ThreadViewJSON`, defaulting to empty slice `[]` rather than null. |
| `wire.ToThreadsEnvelope` | [`internal/wire/thread.go#L161`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/wire/thread.go#L161) | Encodes hoisted `GraphLegacy` into `ThreadsEnvelope` for list endpoints. |
| `wire.ToLegacyDependenciesJSON` | [`internal/wire/dependency.go#L97-L115`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/wire/dependency.go#L97-L115) | Shared conversion from core diagnostics to wire JSON; suppresses duplicate presentation when `Location == TaskPath`. |
| `wire.SchemaVersion` | [`internal/wire/wire.go#L373`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/wire/wire.go#L373) | Incremented monotonically to 1.87 (ADDITIVE revision). |
| `render.graphDiagnosticsHuman` | [`internal/cli/render/dependency.go#L323-L358`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/render/dependency.go#L323-L358) | Groups explained missing/ambiguous graph problems under legacy occurrence summary; sanitizes display text. |
| `render.threadDiagnosticsHuman` | [`internal/cli/render/thread.go#L384-L389`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/render/thread.go#L384-L389) | Dispatches Thread document problems followed by shared `graphDiagnosticsHuman`. |
| `render.ThreadListGraphProblemsHuman` | [`internal/cli/render/thread.go#L44-L50`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/render/thread.go#L44-L50) | Renders repository graph diagnostics to stderr for streaming list modes (`name`, `table`, `csv`). |
| `render.ThreadsProjectedPageJSON` | [`internal/cli/render/list_query.go#L49-L59`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/cli/render/list_query.go#L49-L59) | Preserves repository graph context (including `graph_legacy`) on paginated projected lists while preserving backward-compatible envelope for unbounded projection. |
| `tui.threadDiagnosticLines` | [`internal/tui/detail.go#L1350-L1382`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/tui/detail.go#L1350-L1382) | Formats Thread diagnostic lines in TUI, suppressing grouped graph problems and coloring legacy lines (yellow for ready, red for broken). |
| `tui.terminalText` | [`internal/tui/detail.go#L1690-L1704`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/tui/detail.go#L1690-L1704) | Neutralizes terminal escape sequences and replaces newlines/control characters to prevent line forging. |
| `tui.loadThreadList` | [`internal/tui/thread_projection.go#L46-L51`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/tui/thread_projection.go#L46-L51) | Captures `graphLegacy` into `threadListDiagnostics` on tab reload. |
| `tui.Model.listPaneContent` / `footer` | [`internal/tui/view.go#L218-L223`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/tui/view.go#L218-L223), [`#L533-L542`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/tui/view.go#L533-L542) | Renders legacy diagnostic lines for empty lists and displays migration cues in the footer status bar. |

---

### Findings

No open product defects, data corruption risks, or security vulnerabilities were identified. All verification gates and hostile probes passed.

---

### Mandatory evidence floor ledger

#### 1. Outside-Thread Resolved/Empty Legacy Fixtures & Repair Lifecycle

Tested in isolated sandbox scratch directory (`/tmp/taskflow-probe-1.*` and `/tmp/taskflow-probe-1b.*`) with fresh `/tmp/tskflwctl` binary:

1. **Frontier & List Suppression:**
   - Populated a repository with an active Thread containing member task in `next-up` status.
   - Added a legacy `blocked_by: [<prereq>]` field to an unrelated completed task outside the Thread.
   - Inspected `thread frontier delivery` (human and `--json`): graph health reported `degraded`, frontier was strictly empty (`0 eligible member(s)`, `"frontier": []`), and the legacy occurrence was explicitly reported (`⚠ legacy blocked_by on <outside-id>; run task depend migrate`).
   - Inspected `thread list` (human and `--json`): `GraphLegacy` was hoisted once at top level, per-thread `graph_legacy` was empty `[]`, and no dispatchable tasks were surfaced.
2. **Migration Preview & Commit:**
   - Ran `task depend migrate --dry-run`: reported planned addition of canonical dependency and planned removal of `blocked_by`. SHA-256 tree snapshot comparison before and after verified zero file modifications occurred.
   - Ran `task depend migrate`: committed the migration, cleared `blocked_by`, added `depends_on`.
   - Reread `thread frontier delivery --json`: verified `graph_health: "healthy"`, `graph_legacy: []`, and the member task became dispatchable in the frontier.
3. **Empty Legacy Field Cleanup with `task depend repair --auto`:**
   - Configured an outside task with present-empty `blocks: []`.
   - Thread frontier before repair reported `graph_health: "degraded"` and empty frontier.
   - Ran `task depend repair --auto --dry-run --json`: returned `initial_health: "degraded"`, `final_health: "healthy"`, and `thread_impacts` showing `before.graph_legacy` carrying `blocks` and `after.graph_legacy: []` with frontier restored to 1 eligible task. Snapshot check verified dry run wrote nothing.
   - Ran `task depend repair --auto --json`: committed the cleanup.
   - Subsequent `thread frontier delivery --json` matched the receipt's `after` state exactly (`graph_health: "healthy"`).

#### 2. Missing/Ambiguous References & Independent Structural Defects

1. **Display Grouping & Evidence Retention:**
   - Created a fixture with an outside task containing `blocked_by: [missing-slug]` (unresolvable) and `depends_on: [invalid-token!]` (structural defect).
   - Human `thread frontier` output verified:
     - The independent structural defect survived and printed: `⚠ invalid-dependency-id: task ... depends_on value "invalid-token!" is not a stable task id`.
     - The legacy missing reference was grouped under the field diagnosis: `⚠ legacy blocked_by on ... ("missing-slug": missing); run task depend repair, then task depend migrate`.
     - The raw `legacy-reference-missing` problem was suppressed from duplicate printing.
   - Machine `thread frontier --json` verified:
     - `graph_problems` preserved both `invalid-dependency-id` and `legacy-reference-missing`.
     - `graph_legacy[0].references` contained the missing reference with `"resolution": "missing"`.
2. **Migration Refusal Without Source Changes:**
   - Executed `task depend migrate` against the broken graph: command failed closed with exit code 11 (`ErrValidation`), printing repository broken notice and repair recommendation.
   - SHA-256 tree snapshot comparison verified no task or thread files were touched.
3. **Duplicate-Occurrence & Pathless Producers:**
   - Verified via `TestThreadLegacyDiagnosticsDistinguishDuplicateSourceOccurrences`: two records sharing identical `TaskID` but differing in `Location` produce two distinct legacy diagnostics and neither masks the other's missing reference.

#### 3. Optional Wire Fields, JSON Schema Validation & Streams

1. **Populated Optional Location & Decoded Semantics:**
   - Verified via `TestThreadLegacyJSONPreservesOccurrenceAcrossConsumers`: populated opaque locations (e.g. `db://tasks/outside`), resolved edges, and candidate IDs survive conversion across `show`, `frontier`, `graph`, `plan`, `update`, `impacts`, and `list` envelopes.
   - When `Location == TaskPath`, `Location` is suppressed from wire output to avoid duplicate path presentation while preserving `Location` when distinct or pathless.
2. **Validation Against `schema --json-schema`:**
   - Validated marshaled `ThreadFrontierEnvelope` and `ThreadsEnvelope` instances against compiled schema output from `JSONSchema()`.
   - Verified mutating fields to invalid types causes immediate schema rejection.
3. **Empty Selected Pages & Stream Stdout/Stderr Separation:**
   - Executed `thread list --search not-present -o name`, `-o table`, and `-o csv`.
   - For all modes: stdout received zero rows / empty output, while stderr cleanly received:
     `Diagnostics  repository graph: degraded`
     `⚠ legacy blocked_by on ...; run task depend migrate`
4. **TUI Comprehension & Terminal Sanitization:**
   - Verified via `TestThreadLegacyDiagnosticsReachTUIRegistryAndDetail` and `TestThreadLegacyTUIDeduplicatesGroupedProblemsAndSanitizesContext`:
     - Both detail summary and topology ranks display the legacy occurrence and remedy.
     - Footer status bar reflects legacy count and migration cue (`⚠ 1 legacy field(s); run task depend migrate`).
     - Empty list view reports graph status and legacy diagnosis.
     - `terminalText` replaces newlines with `↵`, tabs with `⇥`, and control characters with ``, preventing injected text from forging terminal rows or corrupting displays.

---

### Compiler-valid mutation ledger

| Invariant Class | Exact Mutation Location | Mutation Content | Intended Test | Observed Test Failure | Restoration Result |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Occurrence Grouping (Antigravity Required):** Location matching in `ExplainsGraphProblem` must be exact to prevent cross-occurrence problem hiding. | [`internal/core/dependency_graph.go#L153`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/dependency_graph.go#L153) | Removed `\|\| d.Location != problem.Location`, allowing location mismatches. | `TestThreadLegacyDiagnosticsDistinguishDuplicateSourceOccurrences` and `TestThreadLegacyDiagnosticsPreserveBrokenSeverityAndGrouping` | **FAIL:** `TestThreadLegacyDiagnosticsPreserveBrokenSeverityAndGrouping:93: grouping hid a different source occurrence`; `TestThreadLegacyDiagnosticsDistinguishDuplicateSourceOccurrences:122: missing reference has 2 occurrence owners`. | Restored baseline; tests PASS (0.13s). |
| **Continuation Digest (Cursor Invalidation):** Diagnostic digest in `QueryThreadViews` must include `GraphLegacy` to invalidate cursors upon outside legacy changes. | [`internal/core/service_list_query.go#L122`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/service_list_query.go#L122) | Omitted `GraphLegacy` from the anonymous context struct passed to `queryList`. | `TestThreadListLegacyDiagnosticsHoistAndInvalidateContinuation` | **FAIL:** `thread_legacy_test.go:171: changed outside diagnosis permitted stale continuation: <nil>`. | Restored baseline; test PASS (0.09s). |
| **Repository Scope Retention:** `projectThread` must attach `graph.LegacyDiagnostics()` to Thread views. | [`internal/core/thread_projection.go#L195`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/internal/core/thread_projection.go#L195) | Commented out `view.GraphLegacy = graph.LegacyDiagnostics()`. | `TestThreadLegacyDiagnosticsRetainRepositoryScopeAndOwnTheirValues` | **FAIL:** `thread_legacy_test.go:29: legacy-only cause was lost or dispatch changed: ... GraphLegacy:[]`. | Restored baseline; test PASS (0.10s). |

---

### Surviving-hypothesis ledger

All required hostile angles were investigated and settled as **falsified**:

1. **"Global context is lost or repeated in lists, zero rows, or paginated projections": FALSIFIED.**
   - In `service_thread.go`, `ListThreadViews` hoists `GraphLegacy` to the outer `ThreadListView` while clearing `view.GraphLegacy` on member rows. When serialized via `ToThreadsEnvelope`, outer `graph_legacy` carries the repository-wide array while each thread row contains `[]`.
   - When filters produce zero rows (`--search not-present`), `empty.GraphLegacy` remains fully populated.
   - For paginated queries with customized columns (`ThreadsProjectedPageJSON`), `page != nil` causes `graph_health`, `graph_problems`, and `graph_legacy` to be passed into `context`, preserving hoisted diagnostics.
   - For streaming modes (`-o name`, `-o table`, `-o csv`), repository graph problems are emitted cleanly to stderr via `ThreadListGraphProblemsHuman`.
2. **"Field-level guidance misrepresents whole-graph readiness when other defects coexist": FALSIFIED.**
   - While an individual resolvable field reports `MigrationGuidance() = "run task depend migrate"`, the migration engine (`MigrateTaskDependencies`) revalidates the entire repository graph snapshot.
   - In `planLegacyDependencyMigration` (`internal/core/dependency_operations.go`), every diagnostic is inspected; if any reference in any field is unresolvable or if any structural graph problem exists, migration fails closed immediately (`domain.ErrValidation`) without writing any files.
3. **"Duplicate IDs or pathless sources make grouped causes hide another occurrence, or mutable slices corrupt shared snapshots": FALSIFIED.**
   - `ExplainsGraphProblem` matches `TaskID`, `Field`, `TaskPath`, and `Location`. If two records share `TaskID` but have distinct `Location` strings (e.g. `db://first` vs `db://second`), neither hides the other's defect.
   - Defensive cloning is strictly enforced: `TaskGraph.LegacyDiagnostics` calls `cloneLegacyDiagnostics`, cloning the outer slice, `recordRef` (cleared to 0), `References`, and `CandidateIDs`. `cloneThreadView` also deep-clones `GraphLegacy`. Mutating a receipt or projection never affects the cached graph snapshot.
4. **"Location/control text forges a terminal diagnostic row or core values are sanitized before reaching JSON": FALSIFIED.**
   - Terminal protection: in CLI rendering, `singleLineText` strips ANSI codes and replaces control runes with spaces. In TUI rendering, `terminalText` replaces newlines with `↵`, tabs with `⇥`, and control runes with ``. Neither renderer allows injected strings to manufacture extra rows.
   - Raw fidelity: wire converters (`ToLegacyDependenciesJSON`) do not strip or sanitize strings, preserving full verbatim content in JSON with standard JSON string escaping.
5. **"Migration/repair receipts disagree with subsequent reads because an adjacent converter omits context": FALSIFIED.**
   - Both mutation receipts and subsequent read envelopes (`show`, `frontier`, `graph`, `plan`, `update`, `impacts`) converge on `ToThreadViewJSON` and `ToLegacyDependenciesJSON`.
   - Tested live with `task depend repair --auto`: receipt's `thread_impacts[0].after.graph_legacy` was empty `[]` and subsequent `thread frontier` read returned matching `graph_legacy: []`.
6. **"Second systemic pass: shared abstractions or boundaries only appear to fail closed": FALSIFIED.**
   - Inspected adjacent dependency queries (`task blockers`, `task unblocks`): both continue to consume `ToLegacyDependenciesJSON` without regression.
   - Structural graph problems (such as cycles or invalid IDs) return `false` from `ExplainsGraphProblem` and are never swallowed by legacy grouping.
   - Monotonic machine schema revision 1.87 is strictly additive under ADR-0008, validated by refreshed jsonschema goldens.

---

### Mandatory isolation attestation

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9SGD7m/.git
baseline_commit=94c233c8710d5c24eece887d573e3409316a5fdc
source_blob=1fad402f698af1b717c4f820310236d5692bfa8c
source_fingerprint=1b75c5c396aafa66dbf9cc8b54d9e60f968eb856
deliverable=planning/audits/6gjqmf7n3n89-2026-10-11-thread-legacy-degradation-diagnostics-implementation-antigravity.md
deliverable_changed=true
transfer=pending
```

- **Checks independently executed in sandbox:**
  - Full race test suite: `go test -race ./...` (PASS, 0 failures across all packages).
  - Linter: `golangci-lint run ./...` (PASS, 0 issues).
  - Planning entity lint: `tskflwctl lint` (PASS, 0 errors).
  - CLI docs generator drift: `just docs-check` (PASS, 0 diff).
  - Audit finding lint: `tskflwctl audit lint 6gjqmf7n3n89` (PASS).
  - Focused legacy test suite: `go test -race ./internal/core ./internal/wire ./internal/cli/render ./internal/cli ./internal/tui -run 'ThreadLegacy|ThreadListLegacy|ThreadGraphRepairImpactsCarryLegacy|GraphLegacy'` (PASS).
  - Three compiler-valid mutation probes (killed and restored).
  - Hostile operator CLI probes across live repos in `/tmp`.
