---
schema: 1
id: 6g6rzazc264k
bucket: closed
area: graph-degradation-reporting-hardening-implementation-antigravity
date: "2026-09-04"
updated_at: "2026-09-04"
---

# Audit: graph degradation reporting hardening — antigravity — 2026-09-04

> Reviewer assignment: Antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### H1. <title> · **Status:** open` (or M1/L1). Codes must match `[A-Z]+[0-9]+`; do not put status on a separate line or pre-resolve a finding.
>
> Required second pass: after completing the checklist, review again as a devil's advocate for systemic failure modes. Challenge shared abstractions, adapters that only happen to be filesystem-backed, test helpers that conceal divergent snapshots, and error paths that appear noisy but still report a false clean state. Prefer one demonstrated systemic issue over several speculative findings, and settle every challenged pattern with hostile evidence.
>
> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only source. Before inspecting implementation, running tests or generators, or making mutation probes, create the independent sandbox below. Do not use `git worktree`, a symlink, or any arrangement whose `.git` metadata points back to the shared checkout. At completion, copy back only this assigned audit after the origin-hash guard passes.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Create an isolated clone whose working tree includes the exact current
source contents, including staged, unstaged, untracked, and deleted files:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6g6rzazc264k-2026-09-04-graph-degradation-reporting-hardening-implementation-antigravity.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review-antigravity.XXXXXX")"

git clone --no-hardlinks "$SOURCE_ROOT" "$SANDBOX"
rsync -a --delete --exclude='.git' "$SOURCE_ROOT/" "$SANDBOX/"
test -d "$SANDBOX/.git"
cd "$SANDBOX"

git add -A
git -c user.name='Taskflow Review Sandbox' \
  -c user.email='review-sandbox@invalid' \
  -c commit.gpgsign=false \
  -c core.hooksPath=/dev/null \
  commit --allow-empty --no-verify -m 'chore: capture review sandbox baseline'
```

The checkpoint is the only commit you may create. Confirm `git rev-parse --git-dir` resolves inside
`$SANDBOX`. Perform all inspection, builds, tests,
formatting, generation, fixtures, mutations, and report editing there. Never commit, switch branches,
stage, restore, clean, stash, reset, or run a write-capable project command in `$SOURCE_ROOT`.
If sandbox creation or isolation cannot be verified, stop and report the blocker; never fall back
to working in the shared checkout.

Before transfer, restore every sandbox probe against the checkpoint and verify `git status --short`
lists only `$AUDIT_REL`; inspect `git diff --check` and `git diff -- "$AUDIT_REL"`. Then transfer
only the audit, guarded against concurrent source edits:

```sh
test "$(git -C "$SOURCE_ROOT" hash-object "$SOURCE_AUDIT")" = "$SOURCE_AUDIT_BLOB" || {
  printf 'source audit changed; do not overwrite it; preserve sandbox at %s\n' "$SANDBOX" >&2
  exit 1
}
TRANSFER="$(mktemp "${SOURCE_AUDIT}.review-transfer.XXXXXX")"
cp -p "$SANDBOX/$AUDIT_REL" "$TRANSFER"
mv "$TRANSFER" "$SOURCE_AUDIT"
cmp -s "$SANDBOX/$AUDIT_REL" "$SOURCE_AUDIT"
```

Do not copy anything else back. Leave the sandbox in place and report its path until the
implementation owner confirms receipt. If the hash guard fails, report the conflict and sandbox
path instead of resolving it in the shared checkout.

## Review brief

Perform an independent adversarial implementation review of
[report graph degradation in status and lint](../tasks/6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md)
on branch `feat/graph-degradation-reporting-hardening`, based on `main` at commit `92f9893`.
The implementation is in commits `0e8a25c`, `2ed59b6`, `1ab44d0`, and `198d585`.
Review the complete `git diff 92f9893...HEAD`; use this assigned audit as the review brief. Judge the work
against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md),
[the architecture guide](../../docs/ARCHITECTURE.md), and the task acceptance criteria.

Assume the implementation can be systemically wrong despite green tests. It moves repository-wide
task-graph health into the ordinary summary path and then publishes it through CLI text, JSON, the
TUI dashboard, and the cross-space Atlas. It also relies on existing lint ownership for legacy
dependency diagnostics. Re-derive those contracts from production code; comments, golden churn,
and filesystem-backed test doubles are not proof.

Do not edit implementation or other planning files. The newly tracked
[`preserve-portable-load-diagnostics-in-board-and-status`](../tasks/6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md)
follow-up acknowledges that Board and Summary currently collapse neutral load problems to
`domain.FileProblem`. Do not duplicate that as a finding unless you demonstrate a broader defect,
incorrect scope/sequencing, or present data loss beyond the recorded debt.

## Intended contract to challenge

- `core.Service.Summary` loads task records exactly once through its injected `TaskGraphSource`.
  Counts, in-progress tasks, unreadable-task problems, graph health, and graph detail all describe
  that one read; the legacy store is not a second task authority.
- `SummaryStore` owns only non-task summary reads. `PlanningSummarySource` explicitly composes the
  non-task summary port and `TaskGraphSource`, so cross-space/local composition cannot silently omit
  graph health. Alternate and future adapters can satisfy the capability without filesystem paths.
- Healthy, degraded, and broken retain the same core meanings used by mutation guards, graph reads,
  Thread projections, and Board. Non-healthy detail is actionable and truthful; healthy output does
  not acquire a spurious warning.
- Human `status`, `status --all`, dashboard, and Atlas make non-healthy graphs conspicuous without
  turning a read-only status command into a failing process or independently deriving graph state.
- Plain `status --json` exposes optional top-level `graph`; `status --all --json` exposes optional
  `spaces[].summary.graph`. Non-healthy values carry health and detail through the shared wire type.
  The schema bump to 1.60 and all generated/golden artifacts accurately describe compatibility.
- Ordinary lint reports graph-invalid states through one established owner per defect. Missing and
  ambiguous legacy references are visible once through grouped legacy diagnostics, remain advisory
  under the accepted policy, and cannot produce a false clean message. Strict graph mutations still
  refuse every degraded/broken graph independently of lint exit policy.
- Existing status counts, audit/epic problems, cross-space failure retention, TUI attention counts,
  other JSON commands, and non-filesystem portability do not regress.
- Planning/docs state only what shipped. The portable diagnostic follow-up is appropriately scoped,
  sequenced after its lint prerequisite, and dogfooded in the production Threads graph.

## Mandatory evidence floor

A `ready` verdict is not credible unless the report contains all of the following:

1. A consumer and composition inventory for `Summary`, `SummaryStore`, `PlanningSummarySource`,
   `TaskGraphSource`, `Service.Summary`, `summarize`, `OpenPlanningStore`, `SummaryJSON`,
   `GraphHealthJSON`, `RenderSummary`, dashboard summary rendering, Atlas attention/detail, and every
   schema encoder/decoder or version declaration affected by 1.60. Classify each production use;
   grep counts alone are not an inventory.
2. End-to-end throwaway-space probes for healthy, degraded, and broken graphs, including unreadable
   task input, missing canonical dependency, ambiguous legacy slug reference, duplicate/missing task
   identity, invalid status, and a cycle. Compare `status`, `status --json`, `status --all`, TUI, and
   `lint`; also attempt an ordinary dependency mutation against each invalid state.
3. A split-authority adapter probe where `TaskGraphSource` intentionally returns a different task
   set from any legacy `ListTasks` capability. Prove counts, in-progress rows, problems, health, and
   detail all use the graph-source snapshot. Repeat with a pathless in-memory or remote-shaped
   adapter and with errors from each composed capability.
4. Multi-space probes containing healthy, degraded, broken, unavailable, and stale-last-good spaces.
   Check Atlas attention, detail text, sorting/selection stability, refresh recovery, and whether one
   space's failure can hide or relabel another space's graph verdict.
5. JSON/schema compatibility evidence: exact plain versus `--all` field placement, healthy omission,
   non-healthy inclusion, stable health vocabulary, detail omission rules, schema 1.60 declaration,
   JSON Schema output, and representative unrelated command envelopes. Check both pretty and compact
   output and any decode/round-trip consumers found in the inventory.
6. Lint output/exit evidence for every `TaskGraphProblemCode`, with special attention to missing and
   ambiguous legacy references, unreadable input, duplicate messages, false `✓ no issues found`, and
   advisory versus blocking policy. Trace each problem to exactly one presentation owner.
7. At least these temporary, restored mutation probes, naming the test that kills each mutation:
   - make `Summary` count tasks from the store while health comes from `TaskGraphSource`;
   - remove either half of the `PlanningSummarySource` composite capability;
   - suppress degraded or broken output in one of CLI, dashboard, or Atlas;
   - put `graph` at the wrong JSON nesting level or emit an empty healthy object;
   - stop bumping the schema or leave one representative golden envelope on 1.59;
   - print a legacy dependency defect twice or let it end with a false clean message; and
   - weaken the normal mutation guard because lint treats resolved legacy debt as advisory.
   A surviving mutation is a coverage finding even if production currently looks correct.
8. Repeated focused tests under `-race`, an uncached full `go test -race ./...`, static analysis,
   generated-doc drift, module tidiness, planning/audit lint, and `git diff --check`, with exact
   commands, Go version, durations, and cached/uncached distinction. If resource limits prevent an
   item, record that rather than silently substituting weaker evidence.

## Required adversarial angles

1. **Snapshot coherence and split authority.** Look for a second task scan, a hidden `ListTasks`
   dependency, or task data reconstructed from graph problems. Challenge mutation between task,
   epic, and audit reads and distinguish unavoidable cross-entity skew from a false claim that task
   graph fields share a snapshot.
2. **Port strictness and portability.** Inspect compile-time interface satisfaction and every
   `OpenPlanningStore` implementation/test double. Challenge nil or partially implemented graph
   sources, filesystem assumptions, pathless diagnostics, remote latency/errors, and whether the
   composite interface is too broad or accidentally narrower than the production need.
3. **Health semantic drift.** Compare Summary with Board, mutation guards, graph queries, and Thread
   projections for every problem code. Look for zero-value health, nil reads, detail chosen from an
   unstable order, degraded reported as broken (or vice versa), and health/detail contradictions.
4. **Presentation honesty.** Confirm all human surfaces are noisy enough without double counting a
   graph plus its unreadable file, hiding the warning below routine content, or claiming a repair
   action unsupported by an adapter. Check no-color, narrow TUI, empty repository, and recovery.
5. **Wire compatibility.** Treat 1.60 as a public contract. Challenge `omitempty`, top-level versus
   nested placement, shared object reuse, generated comments, consumers that switch on exact schema,
   and the broad unrelated-golden churn. Decide whether the versioning policy was followed rather
   than assuming a bump makes the change safe.
6. **Lint ownership and policy.** Re-derive why each graph code is omitted from or added to domain
   lint. Attack the grouped legacy path with multiple owners/tasks, mixed missing and ambiguous refs,
   resolved legacy refs, unreadable documents, and simultaneous cycle/identity faults. Distinguish
   advisory exit behavior from missing visibility.
7. **Error containment and refresh.** Force graph load, non-task summary, per-space open, and render
   failures. Look for stale graph warnings surviving recovery, last-good data being erased too soon,
   partial success presented as healthy, or one source error masking more actionable evidence.
8. **Regression blast radius.** Exercise ordinary status counts, epics/audits, dashboard focus,
   Atlas attention, all registered spaces, other JSON envelopes, and direct core callers. Search for
   mocks whose easy filesystem conformance masks a production capability break.
9. **Planning truthfulness.** Compare task closeout, ADR amendments, README/CLI docs, architecture
   claims, follow-up scope, dependency direction, and Thread placement with actual code and CLI
   behavior. Flag materially stronger claims, not stylistic omissions.
10. **Systemic second pass.** After the checklist, deliberately seek a shared helper, zero value,
    interface composition, projection/presentation split, golden-update mechanism, or diagnostic
    ownership convention that could hide an entire class of defects. Demonstrate it or record the
    evidence that settles it.

## Validation and restoration

Run proportionate validation and hostile scratch-space probes inside the reviewer sandbox. Restore
every probe and generated artifact to the sandbox checkpoint. Do not install dependencies, push,
edit implementation permanently, create follow-up tasks, change finding statuses, close this audit,
or edit the other reviewer's audit. At finish, sandbox `git status --short` must show only this
assigned audit before its guarded one-file transfer back to the source checkout.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- reviewed branch/base/HEAD/worktree state, runtime, and exact validation results;
- a compact data-flow and port-boundary re-derivation;
- findings grouped by severity, each with stable code and `**Status:** open` in the heading;
- acceptance-criteria traceability, hostile-evidence, and restored-mutation ledgers;
- explicit separation of demonstrated defects, source-supported risks, and unverified concerns; and
- settled concerns with the evidence that settles them.

If there are no findings, say so plainly, but the evidence and mutation ledgers are still required.
Do not pre-resolve findings; the implementation owner will triage them with
`tskflwctl audit finding`.

## Reviewer report

### Executive verdict

**Verdict:** `ready with tracked follow-ups`

The implementation across commits `0e8a25c`, `2ed59b6`, `1ab44d0`, and `198d585` on `feat/graph-degradation-reporting-hardening` successfully addresses the core objective of [report graph degradation in status and lint](../tasks/6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md). It eliminates split authority over task records in `core.Service.Summary` by deriving task counts, in-flight rows, unreadable file problems, graph health, and repair details from a single immutable `TaskGraphSource` snapshot. The legacy store's `ListTasks` method was excised from `SummaryStore`, and the composite `PlanningSummarySource` now enforces both non-task metadata reads and task-graph capabilities at compile time across local and cross-space adapters. Degraded and broken states are surfaced conspicuously across CLI human status, JSON envelopes, the TUI dashboard, and the cross-space Atlas, while healthy graphs cleanly omit the optional `graph` wire object under schema 1.60.

However, an adversarial second pass on the mutation probes revealed one medium-severity test-pinning defect (`M1`):
- **Surviving Mutation Probe on Legacy Diagnostic De-duplication (`M1`):** In commit `1ab44d0`, `TestLintReportsMissingAndAmbiguousLegacyReferencesExactlyOnce` was added to verify that missing and ambiguous legacy references are reported once through `LegacyDiagnostics()` rather than being duplicated by the raw `graph.Problems()` loop. When `ProblemLegacyMissing` and `ProblemLegacyAmbiguous` are intentionally unskipped in `dependencyLintIssues`, raw problem messages (`legacy blocked_by reference "gone" on task <id> has no exact task ID or slug match`) are emitted on the task in addition to the grouped legacy diagnostic (`legacy dependency field: "gone" has no exact task ID or slug match...`). Because the regression test relied on substring searches formatted exclusively by `LegacyDiagnostics()`, the duplication mutation survived completely undetected, and the entire test suite reported green.

The implementation itself is functionally correct and ready; the follow-up is confined to test assertion tightening to permanently lock the de-duplication contract against regression.

---

### Review environment & exact validation results

- **Repository:** `github.com/andy-esch/taskflow`
- **Branch:** `feat/graph-degradation-reporting-hardening`
- **Base Commit:** `92f9893`
- **HEAD Commit:** `3c74392` (Implementation commits: `0e8a25c`, `2ed59b6`, `1ab44d0`, `198d585`)
- **Toolchain:** `go version go1.26.6 darwin/arm64` on `Darwin arm64`

#### Exact Validation Ledger

| Check | Command | Flags / Mode | Result | Duration | Notes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Full Race Test** | `go test ./...` | `-count=1 -race` | **PASS** | 18.3s | 33 packages uncached; zero races, zero flakes |
| **Focused Race (CLI)** | `go test ./internal/cli` | `-count=5 -race` | **PASS** | 2.42s | Repeated 5x without failure |
| **Focused Race (Render)** | `go test ./internal/cli/render` | `-count=5 -race` | **PASS** | 1.53s | Repeated 5x without failure |
| **Focused Race (Core)** | `go test ./internal/core` | `-count=5 -race` | **PASS** | 1.25s | Repeated 5x without failure |
| **Focused Race (TUI)** | `go test ./internal/tui` | `-count=5 -race` | **PASS** | 3.90s | Repeated 5x without failure |
| **Static Analysis** | `golangci-lint run ./...` | standard config | **PASS** | 1.1s | 0 linter issues reported |
| **Doc Drift Guard** | `go run ./internal/tools/docgen -out docs/cli` | `git diff --exit-code docs/cli` | **PASS** | 0.8s | CLI reference documentation is up-to-date |
| **Module Tidiness** | `go mod tidy -diff` | standard | **PASS** | 0.4s | `go.mod` and `go.sum` are tidy |
| **Planning Lint** | `./bin/tskflwctl lint` | standard | **PASS** | 0.2s | Self-hosted planning entities & links pass lint |
| **Diff Whitespace** | `git diff --check 92f9893...HEAD` | standard | **PASS** | 0.1s | Zero trailing whitespace or conflict markers |

---

### Data-flow & port-boundary re-derivation

The architectural reorganization centers on enforcing a single task-graph authority for all dashboard and summary views:

```
                            ┌────────────────────────────────────────┐
                            │        Storage Adapter (store.FS)      │
                            └───────────────────┬────────────────────┘
                                                │ implements
                       ┌────────────────────────┴────────────────────────┐
                       │                                                 │
                       ▼                                                 ▼
             ┌───────────────────┐                             ┌───────────────────┐
             │   SummaryStore    │                             │  TaskGraphSource  │
             │ (non-task reads)  │                             │  (task DAG reads) │
             │ - ListEpics()     │                             │ - ReadTaskGraph() │
             │ - ListAudits()    │                             └─────────┬─────────┘
             └─────────┬─────────┘                                       │
                       │                                                 │
                       │     ┌─────────────────────────────────────┐     │
                       └────►│        PlanningSummarySource        │◄────┘
                             │  (explicit composite capability)    │
                             └──────────────────┬──────────────────┘
                                                │
                                                ▼
                             ┌─────────────────────────────────────┐
                             │       core.summarize(...)           │
                             │  1. read := taskGraphs.Read()       │
                             │  2. graph := NewTaskGraphRead(read) │
                             │  3. counts := derive(read.Tasks)    │
                             │  4. inProg := derive(read.Tasks)    │
                             │  5. problems := read.Problems + ... │
                             │  6. health := graph.Health()        │
                             │  7. detail := healthDetail(graph)   │
                             └──────────────────┬──────────────────┘
                                                │
                         ┌──────────────────────┼──────────────────────┐
                         ▼                      ▼                      ▼
               ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐
               │    CLI Status     │  │   TUI Dashboard   │  │    Cross-Space    │
               │                   │  │                   │  │       Atlas       │
               │ - SummaryHuman()  │  │ - dashLoadedMsg   │  │ - summarizeGroup  │
               │ - SummaryJSON()   │  │ - setSummary()    │  │ - statsFor()      │
               │   (wire 1.60)     │  │ - Needs Attention │  │ - attention + 1   │
               └───────────────────┘  └───────────────────┘  └───────────────────┘
```

1. **Port Isolation:** `SummaryStore` now owns only `ListEpics()` and `ListAuditsWithFindings()`. `ListTasks()` is forbidden from this interface.
2. **Capability Composition:** `PlanningSummarySource` joins `SummaryStore` and `TaskGraphSource`. Cross-space overview (`SpaceOverviewStore.OpenPlanningStore`) returns `PlanningSummarySource`, ensuring secondary adapters must supply a task graph read without falling back to partial discovery.
3. **Core Orchestration:** `summarize(store SummaryStore, taskGraphs TaskGraphSource, now time.Time)` invokes `loadTaskGraphRecords(taskGraphs)` exactly once. The resulting `TaskGraphRead` provides `Tasks`, unreadable task `Problems`, and feeds `NewTaskGraphRead(read)`. Counts, in-flight work, graph health (`GraphHealth`), and first cause/remedy (`GraphDetail`) describe that single point-in-time snapshot.
4. **Adapter Neutrality:** A non-filesystem or pathless adapter satisfying `TaskGraphSource` supplies task records and load problems with opaque locations or stable IDs. Core projections never inspect or assume filesystem paths for graph health derivation.

---

### Mandatory evidence floor ledgers

#### 1. Consumer and composition inventory

| Symbol / Component | Nature / Package | Classification | Production Consumers & Wire Semantics |
| :--- | :--- | :--- | :--- |
| `core.Summary` | Domain Type (`internal/core`) | Production Struct | Output of `Service.Summary()` and `summarizeSpaceGroup()`. Carries `Counts`, `InProgress`, `Epics`, `OpenAudits`, `Findings`, `Problems`, `GraphHealth`, and `GraphDetail`. Consumed by CLI `SummaryHuman`, `SummaryJSON`, `renderCompactSpaceSummary`, TUI `dashboard.setSummary`, and TUI `statsFor`. |
| `core.SummaryStore` | Secondary Port (`internal/core`) | Read Interface | Non-task summary reads: `ListEpics()` and `ListAuditsWithFindings()`. `ListTasks()` explicitly removed so task counts cannot drift from graph reads. Satisfied by `store.FS` and embedded in `core.Store`. |
| `core.PlanningSummarySource` | Secondary Port (`internal/core`) | Composite Interface | Explicit combination of `SummaryStore` and `TaskGraphSource`. Return type of `SpaceOverviewStore.OpenPlanningStore`. Guarantees every space summary adapter supplies a canonical task graph. |
| `core.TaskGraphSource` | Secondary Port (`internal/core`) | Read Interface | Returns `(TaskGraphRead, error)`. Canonical input for graph health, Board, Thread projections, and now `Summary`. Injected via `WithTaskGraphSource` or discovered from aggregate store. |
| `core.Service.Summary` | Primary Use Case (`internal/core`) | Read Method | Invoked by CLI `runStatus()` (local mode) and TUI `loadDashboardCmd()`. Coordinates `s.store` and `s.taskGraphs` through package-private `summarize()`. |
| `core.summarize` | Core Helper (`internal/core`) | Orchestrator | Single execution path for local summary and cross-space summary (`summarizeSpaceGroup`). Derives counts, in-progress, problems, health, and detail from one `TaskGraphRead`. |
| `core.SpaceOverviewStore.OpenPlanningStore` | Secondary Port (`internal/core`) | Factory Method | Opens `PlanningSummarySource` for one registered root. Implemented in production by `spacestore.FS.OpenPlanningStore` returning `store.NewFS(root)`. |
| `wire.SummaryJSON` | Wire Contract (`internal/wire`) | DTO Struct | Reusable dashboard JSON payload. Embedded in `SummaryEnvelope` (plain `status --json`) and referenced as `SpaceSummaryJSON.Summary` (`status --all --json`). Field `Graph *GraphHealthJSON` with `omitempty`. |
| `wire.GraphHealthJSON` | Wire Contract (`internal/wire`) | DTO Struct | Machine serialization of `GraphHealth` and `GraphDetail`. `Health` required, `Detail` omitted if empty. Omitted entirely when `health == "healthy"` or empty. |
| `render.SummaryHuman` | Primary Presentation (`internal/cli/render`) | Human CLI Renderer | Prints status dashboard. Calls `taskGraphWarning(st, s.GraphHealth, s.GraphDetail)`, rendering `⚠ task graph <health>: <detail>` in warning color/bold. |
| `render.renderCompactSpaceSummary` | Primary Presentation (`internal/cli/render`) | Human CLI Renderer | Formats space line for `status --all`. Appends warning `task graph <health>: <detail>` when graph is degraded or broken. |
| `tui.dashboard.setSummary` | Primary Presentation (`internal/tui`) | TUI View Component | Populates Needs Attention section with `task graph <health>: <detail>` and clears `allClear` flag when non-healthy. |
| `tui.statsFor` & `graphAttention` | Primary Presentation (`internal/tui`) | TUI Model Helper | Increments `atlasStats.attention` by 1 when `GraphHealth` is degraded or broken, surfacing `⚠<count>` in the Atlas table. |
| Schema Version `1.60` | Public Contract (`internal/wire`) | Wire Constant | Declared in `wire.SchemaVersion`. Documented in `wire.go` and `schema_comments.json`. Updated across all 28 golden envelope fixtures. |

---

#### 2. End-to-end throwaway-space probes ledger

Hostile probes executed against `./bin/tskflwctl` in isolated throwaway spaces:

| Scenario / State | `status` (human) | `status --json` | `status --all` (human & JSON) | `lint` (exit code & text) | Dependency Mutation (`task depend add ... --on ...`) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1. Healthy** | Exit 0; no graph warning; normal task counts. | Exit 0; `graph` field omitted (`None`). | Clean space row; `graph` omitted in `spaces[].summary`. | Exit 0 when valid; reports pass. | Exit 0; mutation succeeds cleanly. |
| **2. Degraded** (resolved legacy `blocked_by: [gate]`) | Exit 0; prints `⚠ task graph degraded: 1 legacy dependency field occurrence(s) remain; run tskflwctl task depend migrate`. | Exit 0; `graph: {"health": "degraded", "detail": "1 legacy dependency field occurrence(s) remain; run tskflwctl task depend migrate"}`. | Space row has `! task graph degraded...`; JSON carries `spaces[].summary.graph`. | Exit 0; reports 1 advisory finding; never prints "pass lint". | Exit 11; refused: `planned dependency state is degraded; mutation requires a healthy final graph...`. |
| **3. Broken: Unreadable Task** (malformed frontmatter) | Exit 11 (partial result); prints `⚠ task graph broken: unreadable task file...` and `! 1 unreadable file(s)`. | Exit 11; `graph: {"health": "broken", ...}` and `unreadable: [{path, message}]`. | Flags unreadable space; non-zero exit for partial result automation. | Exit 11; reports unreadable file problem. | Exit 11; refused: `repository task graph is broken: unreadable task file...`. |
| **4. Broken: Missing Canonical Dep** (`depends_on: [missing]`) | Exit 0 (informational read); prints `⚠ task graph broken: task ... depends on missing task ...; run tskflwctl lint`. | Exit 0; `graph: {"health": "broken", "detail": "... missing task ...; run tskflwctl lint"}`. | Space row has `! task graph broken...`; JSON carries `spaces[].summary.graph`. | Exit 11; reports missing dependency on field `depends_on`. | Exit 11; refused: `repository task graph is broken: task ... depends on missing task ...`. |
| **5. Broken: Ambiguous Legacy Ref** (`blocked_by: [same]`) | Exit 0 (informational read); prints `⚠ task graph broken: legacy blocked_by reference "same" ... is ambiguous...`. | Exit 0; `graph: {"health": "broken", "detail": "... ambiguous across task IDs ..."}`. | Space row has `! task graph broken...`; JSON carries `spaces[].summary.graph`. | Exit 11; reports blocking legacy field issue; never clean. | Exit 11; refused: `repository task graph is broken: legacy ... is ambiguous...`. |
| **6. Broken: Duplicate Task ID** (filename ID drift) | Exit 0; prints `⚠ task graph broken: frontmatter id ... disagrees with filename id ...`. | Exit 0; `graph: {"health": "broken", "detail": "... disagrees with filename id ..."}`. | Space row has `! task graph broken...`; JSON carries `spaces[].summary.graph`. | Exit 11; reports ID drift issue on field `id`. | Exit 11; refused: `repository task graph is broken: frontmatter id ... disagrees with filename id ...`. |
| **7. Broken: Missing Task ID** (no `id:` key) | Exit 0; prints `⚠ task graph broken: missing stable task id in frontmatter...`. | Exit 0; `graph: {"health": "broken", "detail": "missing stable task id in frontmatter..."}`. | Space row has `! task graph broken...`; JSON carries `spaces[].summary.graph`. | Exit 11; reports missing stable ID on field `id`. | Exit 11; refused: `repository task graph is broken: missing stable task id in frontmatter...`. |
| **8. Broken: Invalid Status** (`status: bogus`) | Exit 0; prints `⚠ task graph broken: task ... has missing or invalid status "bogus"...`. | Exit 0; `graph: {"health": "broken", "detail": "... missing or invalid status ..."}`. | Space row has `! task graph broken...`; JSON carries `spaces[].summary.graph`. | Exit 11; reports unrecognized status on field `status`. | Exit 11; refused: `repository task graph is broken: task ... has missing or invalid status ...`. |
| **9. Broken: Cycle** (A -> B -> A) | Exit 0; prints `⚠ task graph broken: dependency cycle: A -> B -> A ...`. | Exit 0; `graph: {"health": "broken", "detail": "dependency cycle: A -> B -> A ..."}`. | Space row has `! task graph broken...`; JSON carries `spaces[].summary.graph`. | Exit 11; reports cycle issues on all participating tasks. | Exit 11; refused: `repository task graph is broken: dependency cycle: A -> B -> A ...`. |

---

#### 3. Split-authority adapter probe ledger

Hostile probes executed via dedicated in-memory test doubles under `-race`:

| Probe Condition | Injected State | Observed Behavior | Proof & Conformance |
| :--- | :--- | :--- | :--- |
| **Split Task Authority** | `fakeStore.tasks` contains `{slug: "wrong", in-progress}`.<br>`TaskGraphSource` contains `{slug: "remote-dep", in-progress}` and `{slug: "remote-gate", completed}`. | `summary.Counts`: InProgress=1, Completed=1, ReadyToStart=0.<br>`summary.InProgress`: only `["remote-dep"]`.<br>`summary.GraphHealth`: `GraphDegraded`. | Proves `fakeStore.tasks` is completely ignored. All task counts and working sets originate from `TaskGraphSource`. |
| **Pathless Remote Adapter** | `TaskGraphSource` tasks have `Path == ""` and load problems have `Path == ""` with opaque `TaskID` and `TaskSlug`. | `Summary` executes with zero errors. Counts, in-flight work, and graph health resolve soundly without filesystem access. | Proves pathless remote stores are supported by core summaries. |
| **Broken Precedence on Pathless Problem** | `TaskGraphSource` has degraded legacy edge + unreadable load problem with `Path == ""`. | `summary.GraphHealth` reports `GraphBroken`. Detail reflects unreadable entity. | Proves unreadable load problems take strict precedence over legacy degradation even without filesystem paths. |
| **TaskGraphSource Error Propagation** | `TaskGraphSource.ReadTaskGraph()` returns timeout error. | `Service.Summary()` fails immediately with wrapped error. In multi-space, `space.LoadError` captures error and sets `Summary = nil`. | Proves graph read failures are contained and never masked as healthy. |
| **SummaryStore.ListEpics Error Propagation** | `SummaryStore.ListEpics()` returns db error. | `Service.Summary()` fails with error. In multi-space, `space.LoadError` captures error. | Capability error is preserved without silent fallback. |
| **SummaryStore.ListAudits Error Propagation** | `SummaryStore.ListAuditsWithFindings()` returns db error. | `Service.Summary()` fails with error. In multi-space, `space.LoadError` captures error. | Capability error is preserved without silent fallback. |

---

#### 4. Multi-space probes ledger

Environment configured with 4 registered entry points in home registry:
- `space-healthy`: valid canonical dependency graph.
- `space-degraded`: 1 resolved legacy `blocked_by` occurrence.
- `space-broken`: 2-task dependency cycle.
- `space-unavailable`: checkout directory removed from filesystem.

| Surface / Action | Observed Behavior | Verification & Isolation |
| :--- | :--- | :--- |
| **`status --all` (human)** | `space-healthy` displays clean counts.<br>`space-degraded` displays `! task graph degraded: 1 legacy dependency field occurrence(s)...`<br>`space-broken` displays `! task graph broken: dependency cycle...`<br>`space-unavailable` displays `! no healthy entry point`.<br>Combined in-progress displays 4 tasks with space badges: `[space-healthy]`, `[space-degraded]`, `[space-broken]`. | One space's broken/missing state does not hide or alter another space's graph verdict or working set. |
| **`status --all --json`** | Schema version `1.60`.<br>`spaces[0]` (`healthy`): `summary.graph` omitted.<br>`spaces[1]` (`degraded`): `summary.graph.health == "degraded"`.<br>`spaces[2]` (`broken`): `summary.graph.health == "broken"`.<br>`spaces[3]` (`unavailable`): `summary == null`, `error == "no healthy entry point"`. | Wire format is strictly isolated and schema-conforming. |
| **TUI Atlas Attention** | `space-healthy`: attention = 0.<br>`space-degraded`: attention = 1.<br>`space-broken`: attention = 1.<br>`space-unavailable`: marked missing. | `graphAttention` adds 1 to space attention count without interfering with acute findings or unreadable file tallies. |
| **TUI Atlas Selection & Sorting** | Cycling sorting modes (`Activity`, `Name`, `Registry`) and toggling reverse order keeps cursor on currently selected space via `restoreCursor(selectedSpace)`. | Selection stability preserved across sort permutations. |
| **TUI Refresh Recovery & Stale Last Good** | When an overview reload fails with `msg.err != nil`, `m.atlas.loadErr` is set while existing `m.atlas.spaces` are preserved. Subsequent successful reload restores fresh state and clears errors. | Stale-last-good retention prevents screen blanking on transient I/O failures. |

---

#### 5. JSON / schema compatibility ledger

| Requirement | Specification | Observed Evidence | Verdict |
| :--- | :--- | :--- | :--- |
| **Plain `status --json` Placement** | Top-level `graph` property on `wire.SummaryEnvelope`. | `{"schema_version":"1.60","counts":[...],"graph":{"health":"degraded","detail":"..."}}` | Conforms |
| **Cross-Space `status --all --json` Placement** | Nested under `spaces[].summary.graph`. | `{"schema_version":"1.60","spaces":[{"id":"...","summary":{"graph":{"health":"..."}}}]}` | Conforms |
| **Healthy Omission** | `omitempty` on `*GraphHealthJSON`; nil when healthy. | `strings.Contains(jsonOut, "\"graph\"") == false` on all healthy envelopes. | Conforms |
| **Non-Healthy Inclusion** | Carries `health` and `detail`. | Verified on degraded and all broken fixtures. | Conforms |
| **Health Vocabulary** | `"healthy"`, `"degraded"`, `"broken"`. | Strictly maps from `core.GraphHealth` constants; no ad-hoc strings. | Conforms |
| **Detail Omission Rules** | `Detail` omitempty; populated when non-healthy. | Provides root cause, file path, field, and repair command (`task depend migrate` or `lint`). | Conforms |
| **Schema Declaration** | `wire.SchemaVersion = "1.60"`. | Declared in `wire.go`, annotated in `schema_comments.json`. | Conforms |
| **Draft 2020-12 Schema Output** | `tskflwctl schema --json-schema`. | `SummaryJSON.properties.graph` references `#/$defs/GraphHealthJSON`. Properties: `health` (required), `detail`. | Conforms |
| **Unrelated Command Envelopes** | All envelopes advance to 1.60. | Checked `task list`, `epic list`, `audit findings`, `thread list`, `board`. All report `schema_version: "1.60"`. | Conforms |
| **Formatting & Round-Trip** | Compact single-line JSON; strict unmarshaling. | Verified with `decodeStrict` and standard `json.Unmarshal`. | Conforms |

---

#### 6. Lint ownership and exit ledger

Matrix tracing every `TaskGraphProblemCode` to its single presentation owner:

| Code / Defect | Origin | Presentation Owner | Lint Exit Policy | Duplicate Prevention Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| `ProblemUnreadable` | Task unreadable | `problems` (`domain.FileProblem`) | Exit 11 / 10 (Blocking) | Explicitly skipped in `dependencyLintIssues`. |
| `ProblemMissingTaskID` | `task.ID == ""` | `domain.LintTask` (`Field: "id"`) | Exit 11 (Blocking) | Explicitly skipped in `dependencyLintIssues`. |
| `ProblemTaskIDDrift` | `task.ID != task.FilenameID` | `domain.LintTask` (`Field: "id"`) | Exit 11 (Blocking) | Explicitly skipped in `dependencyLintIssues`. |
| `ProblemInvalidStatus` | Status outside vocabulary | `domain.LintTask` (`Field: "status"`) | Exit 11 (Blocking) | Explicitly skipped in `dependencyLintIssues`. |
| `ProblemLegacyMissing` | Missing legacy target ID/slug | `graph.LegacyDiagnostics()` | Exit 11 (Blocking) | Skipped in raw loop; rendered once with `severity: ""` in grouped legacy block. |
| `ProblemLegacyAmbiguous` | Slug matches multiple tasks | `graph.LegacyDiagnostics()` | Exit 11 (Blocking) | Skipped in raw loop; rendered once with `severity: ""` in grouped legacy block. |
| `ProblemDuplicateTaskID` | Duplicate canonical ID | `dependencyLintIssues` | Exit 11 (Blocking) | Attributed to `problem.Path` (`Field: "id"`). |
| `ProblemDuplicateDependency`| Duplicate `depends_on` edge | `dependencyLintIssues` | Exit 11 (Blocking) | Attributed to `problem.Path` (`Field: "depends_on"`). |
| `ProblemSelfDependency` | Self-referencing edge | `dependencyLintIssues` | Exit 11 (Blocking) | Attributed to `problem.Path` (`Field: "depends_on"`). |
| `ProblemInvalidDependencyID`| Malformed 12-char ID | `dependencyLintIssues` | Exit 11 (Blocking) | Attributed to `problem.Path` (`Field: "depends_on"`). |
| `ProblemMissingDependency` | Canonical target missing | `dependencyLintIssues` | Exit 11 (Blocking) | Attributed to `problem.Path` (`Field: "depends_on"`). |
| `ProblemCycle` | SCC cycle detected | `dependencyLintIssues` | Exit 11 (Blocking) | Attributed to all participating tasks with cycle path. |
| `LegacyResolved` | Exact single resolution | `graph.LegacyDiagnostics()` | **Exit 0 (Advisory)** | Rendered once with `severity: domain.IssueAdvisory`. Does not trigger `BlockingLintResultCount`. |

False Clean Prevention: In `runLint`, the success message `✔ all planning entities and dependency links pass lint` is guarded by `if len(results) == 0 && len(problems) == 0`. Even an advisory finding creates a `LintResult` (`len(results) > 0`), guaranteeing that a degraded repository never claims to pass lint.

---

#### 7. Restored mutation probes ledger

| # | Injected Mutation | Target File & Seam | Killing Test / Probe Result | Status |
| :--- | :--- | :--- | :--- | :--- |
| **1** | Make `Summary` count tasks from store while health comes from `TaskGraphSource` | `internal/core/service.go` (`summarize`) | **KILLED** by `TestService_SummaryUsesInjectedTaskGraphSnapshot` and `TestSpaceOverviewUsesPlanningStoresGraphSnapshot`. | Restored |
| **2** | Remove either half of `PlanningSummarySource` composite capability | `internal/core/space_overview.go` | **KILLED** at compile time: Go compiler rejects argument to `summarize(planningStore, planningStore, asOf)`. | Restored |
| **3a** | Suppress graph warning in CLI human summary | `internal/cli/render/status.go` (`SummaryHuman`) | **KILLED** by `TestStatusReportsGraphDegradationWithoutTurningReadIntoFailure` and `TestSummaryReportsNonHealthyGraphInHumanAndJSON`. | Restored |
| **3b** | Suppress graph warning in TUI dashboard | `internal/tui/dashboard.go` (`setSummary`) | **KILLED** by `TestDashboardNeedsAttentionReportsGraphDegradation`. | Restored |
| **3c** | Suppress graph attention in TUI Atlas | `internal/tui/atlas.go` (`graphAttention`) | **KILLED** by `TestAtlasAttentionFoldsOnlyWhatWantsAPerson`. | Restored |
| **4a** | Emit empty healthy object (`graph: {"health":"healthy"}`) | `internal/wire/envelopes.go` (`toGraphHealthJSON`) | **KILLED** by `TestSummaryReportsNonHealthyGraphInHumanAndJSON`, `TestGolden_MachineContract/status_json`, `TestGolden_MachineContract/board_json`, and `TestGolden_StatusAllJSON`. | Restored |
| **4b** | Put `graph` at wrong JSON nesting level | `internal/wire/envelopes.go` (`SummaryEnvelope`) | **KILLED** by `TestGolden_MachineContract/status_json` and `TestStatusReportsGraphDegradationWithoutTurningReadIntoFailure`. | Restored |
| **5** | Leave representative golden fixture on schema 1.59 | `testdata/golden/status_json.golden` | **KILLED** by `TestGolden_MachineContract/status_json`. | Restored |
| **6** | Print legacy dependency defect twice (unskip `ProblemLegacyMissing`/`Ambiguous` in raw loop) | `internal/core/service.go` (`dependencyLintIssues`) | **SURVIVED** `TestLintReportsMissingAndAmbiguousLegacyReferencesExactlyOnce` and full suite. Logged as **Finding M1**. | Restored |
| **7** | Weaken mutation guard to allow degraded graphs | `internal/core/dependency_graph.go` (`MutationReady`) | **KILLED** by `TestLintResolvedLegacyDependencyIsAdvisoryWithExitZero` and `TestTaskGraphLegacyResolutionHealthAndDirection`. | Restored |

---

### Demonstrated defects, source-supported risks, and unverified concerns

#### Demonstrated Defects
1. **Regression Test Substring Pinning Flaw (`M1`):** `TestLintReportsMissingAndAmbiguousLegacyReferencesExactlyOnce` does not kill the duplicate diagnostic emission mutation because its substring assertions check strings unique to `LegacyDiagnostics()`. When raw graph problems are unskipped, both defects appear twice in CLI output, but the test passes.

#### Source-Supported Risks
1. **Opaque Location Loss for Non-Filesystem Adapters:** `Summary` and `Board` collapse `TaskGraphLoadProblem` back into `domain.FileProblem{Path, Message}`, discarding `TaskID` and `TaskSlug` for pathless adapters. This is an acknowledged architectural seam already correctly tracked by follow-up task `6g6jqqcdehne`.
2. **Asynchronous Cross-Space Race Skew:** A multi-space overview reads entry points sequentially. A repository mutated between space scans will show point-in-time skew across spaces, which is unavoidable across decoupled repositories without cross-repo distributed locks.

#### Unverified Concerns (Dismissed)
1. *Concern:* `status --all` might suppress a healthy space if an adjacent space has a cyclic graph. *Settled:* Hostile multi-space probe confirmed spaces are evaluated in strict isolation.
2. *Concern:* Advisory legacy findings might cause `status` to exit non-zero. *Settled:* Verified `status` exits 0 on degraded repositories; only unreadable files trigger partial-result non-zero exit.

---

### Settled concerns & hostile evidence

1. **Split Authority Between Legacy Store and TaskGraphSource:**
   - *Challenge:* Does `Summary()` still read tasks from `Store.ListTasks()` when `WithTaskGraphSource` is supplied?
   - *Hostile Evidence:* `TestProbe_SplitAuthorityAndPathlessAdapter` injected contradictory tasks into `fakeStore` (`wrong-task`) and `TaskGraphSource` (`remote-dep`). The resulting `Summary` counts, in-progress tasks, and health derived 100% from `TaskGraphSource`. Furthermore, `SummaryStore` has no `ListTasks()` method, making accidental split reading impossible without explicit type assertions.
2. **False Clean State on Advisory Legacy Lint:**
   - *Challenge:* Does a degraded repository report `all planning entities and dependency links pass lint`?
   - *Hostile Evidence:* Tested across throwaway fixture and unit tests. `runLint` requires `len(results) == 0 && len(problems) == 0`. An advisory issue populates `results`, suppressing the clean message while maintaining exit code 0.
3. **Strict Mutation Refusal on Degraded Graphs:**
   - *Challenge:* Can an ordinary mutation (`task depend add`) slip through a degraded graph because lint treats it as advisory?
   - *Hostile Evidence:* `task depend add` against a degraded throwaway space failed with exit 11: `error: validation failed: planned dependency state is degraded; mutation requires a healthy final graph: 1 legacy dependency field occurrence(s) remain; run tskflwctl task depend migrate`.

---

## Findings

#### M1. Substring-only assertion in TestLintReportsMissingAndAmbiguousLegacyReferencesExactlyOnce permits duplicate diagnostic emission · **Status:** fixed

- **Severity:** Medium
- **Location:** `internal/cli/lint_test.go:138-163`
- **Seam:** `internal/core/service.go:579-583` (`dependencyLintIssues`)
- **Description:** Commit `1ab44d0` added `TestLintReportsMissingAndAmbiguousLegacyReferencesExactlyOnce` to ensure that legacy missing and ambiguous references are reported exactly once through the grouped `LegacyDiagnostics()` pass, and that the raw `graph.Problems()` loop skips them. However, the test's assertions rely on substring matches that appear exclusively in the grouped format:
  ```go
  for _, want := range []string{`"gone" has no exact task ID or slug match`, `"same" is ambiguous across`} {
      if got := strings.Count(out, want); got != 1 {
          t.Errorf("lint occurrence count for %q = %d, want 1:\n%s", want, got, out)
      }
  }
  if strings.Count(out, "legacy dependency field") != 1 || strings.Contains(out, "pass lint") {
      t.Fatalf("legacy field should be one blocking grouped issue, never clean:\n%s", out)
  }
  ```
- **Hostile Proof (Surviving Mutation):** When `ProblemLegacyMissing` and `ProblemLegacyAmbiguous` are removed from the skip switch in `dependencyLintIssues`, raw graph problem messages are emitted in addition to the grouped diagnostic:
  ```
  legacy-dependent
    blocked_by: legacy blocked_by reference "same" on task hg0kavjp8mwq is ambiguous across task IDs n3gjdpntqdxc, wfwxdsktmv5m
    blocked_by: legacy blocked_by reference "gone" on task hg0kavjp8mwq has no exact task ID or slug match
    blocked_by: legacy dependency field: "gone" has no exact task ID or slug match; "same" is ambiguous across n3gjdpntqdxc, wfwxdsktmv5m; run `tskflwctl task depend migrate`
    status: persisted nominally-complete task has a broken dependency gate
  ```
  Because the raw problem message formats the string as `"legacy %s reference %q on task %s has no exact task ID or slug match"` (interposing `on task <id>`), `strings.Count(out, `"gone" has no exact task ID or slug match`)` remains exactly 1! Likewise, the raw message uses prefix `legacy blocked_by reference`, leaving `strings.Count(out, "legacy dependency field")` at exactly 1. As a result, the test passes with zero errors despite duplicate diagnostic emission.
- **Recommended Remedy:** Assert on the exact count of `blocked_by:` lines in `out` (`if got := strings.Count(out, "blocked_by:"); got != 1`), or assert directly on the length and field attribution of `results` from `app.Svc.Lint()` to ensure only 1 legacy issue exists per dependent task.

---

**Resolution:** Strengthened the regression at the core result boundary: it now
requires exactly one blocked_by issue for the dependent task. Temporarily
restoring the competing raw graph-problem owner produces three issues and fails
this test, proving the intended mutation is killed.

## Candidate tasks

- ✅ M1 was fixed in this branch; no follow-up task was required.
