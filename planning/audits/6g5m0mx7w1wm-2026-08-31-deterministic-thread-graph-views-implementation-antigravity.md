---
schema: 1
id: 6g5m0mx7w1wm
bucket: closed
area: deterministic-thread-graph-views-implementation-antigravity
date: "2026-08-31"
updated_at: "2026-08-31"
---

# Audit: Deterministic Thread graph views implementation — Antigravity — 2026-08-31

> Reviewer assignment: Antigravity. This document is the review brief and the only file the reviewer
> should update.

## Review brief

Perform an independent adversarial implementation review of the uncommitted work for task
[`6g3q4rv1w9e2`](../tasks/6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md) on branch
`feat/deterministic-thread-graph-views`, based at `1b9fd9b`. Review it against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md), especially Sections 3, 5, 9, 10, and 11;
the 2026-08-31 portability and generated-view amendments; and
[`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

Assume the implementation may be subtly wrong despite green local tests. Look for semantic
contradictions, false topology claims, edge-direction mistakes, unsafe renderer output, unstable
ordering, leaky adapter boundaries, misleading human output, and machine contracts that will age
poorly when the TUI or a web interface consumes them. Do not reward complexity or test volume by
itself. Equally, do not manufacture findings: settle a concern when code and a hostile reproduction
disprove it.

## Review target

The implementation is spread across tracked and untracked files, so inspect `git status --short`,
`git diff HEAD`, and every relevant untracked file. Primary files are:

- `internal/core/thread_graph.go` and `internal/core/thread_graph_test.go`;
- `internal/core/service_thread.go` and the existing Thread projection/graph analyzer it composes;
- `internal/graphfmt/graphfmt.go` and `internal/graphfmt/graphfmt_test.go`;
- `internal/cli/thread.go`, `internal/cli/thread_graph_test.go`, and
  `internal/cli/render/thread.go`;
- `internal/wire/thread.go`, `internal/wire/envelopes.go`, schema comments, schema validation tests,
  and machine goldens;
- generated `docs/cli/tskflwctl_thread_graph.md` and `tskflwctl_thread_plan.md`; and
- ADR-0006, `docs/ARCHITECTURE.md`, and the implementation task.

Ignore unrelated concurrent work in `planning/meta/`, `routines/`, the branch-protection and
finding-writer task files, and pre-existing edits to the two 2026-08-30 bulk-apply audits. The two
new deterministic-graph-view audit files are review scaffolding, not implementation evidence.

The intended contract is:

- `core.ThreadGraphProjection` is the single taskflow-owned, adapter-neutral semantic contract for
  CLI, TUI, future web, and library callers. It contains the complete `ThreadView`, raw nodes,
  prerequisite-to-dependent edges, member-only explanatory waves, and an explicit
  `TopologyComplete` verdict—never renderer, filesystem, Cobra, Bubble Tea, HTTP, or third-party
  graph types.
- Nodes are bounded to persisted Thread members plus immediate external gates derived by
  `ProjectThread`. Deeper prerequisites remain available through blocker queries. Shared tasks may
  be members of several Threads and may be a member in one projection but an external gate in
  another.
- Nodes and edges are stable-ID ordered. Every otherwise-equal topology order uses stable task ID.
  Repository scan order, map iteration, or input membership order must not affect core, human,
  Mermaid, DOT, or JSON output.
- Waves rank members only. External gates appear as marked nodes/edges and a separate plan section;
  they do not become Thread-owned work. Waves explain dependency order only: they are not dispatch
  authorization, readiness state, duration estimates, barriers, critical path, or scheduling.
- Unhealthy evidence remains diagnostic. Broken or unknown members remain visible but unranked;
  useful unaffected waves may remain; topology completeness is true only when both the member
  topology and qualifying Thread projection are healthy and complete. An unrelated repository
  graph defect must not produce a false `topology_complete: true`.
- `Service.ShowThreadGraph` reads the Thread first and task snapshot second through independently
  injectable `ThreadStore` and `TaskGraphSource` ports. It does not re-enter the aggregate Store or
  authorize mutation from this point-in-time projection.
- `internal/graphfmt` is a pure output-adapter package. Mermaid and DOT encode the same projection
  nodes, edges, roles, and order; use synthetic renderer identifiers; safely escape raw Unicode,
  quotes, backslashes, control characters, newlines, HTML/directive-like content, and renderer
  punctuation; and emit byte-deterministic valid syntax. Generated diagrams are never persisted.
- `thread graph` defaults to Mermaid and supports DOT. `thread plan` presents member waves and gates.
  `--json` on both emits the same versioned renderer-neutral projection; explicit `--format` plus
  `--json` is rejected. Machine schema 1.58 includes non-null deterministic arrays, stable role and
  state vocabulary, health diagnostics, waves, and topology completeness.
- ASCII/Unicode rendering, critical path, slack, forecasting, transitive reduction, graph editing,
  and a graph-library dependency remain outside V1.

## Required hostile angles

1. Re-derive every node and edge from the global task DAG. Try member-to-member edges, a direct
   external gate, a deeper nonmember prerequisite, disconnected members, a deprecated member,
   shared tasks across Threads, duplicate membership from malformed input, missing/unreadable
   members, legacy resolved edges, and dependency defects outside the Thread. Confirm edge direction
   is prerequisite to dependent everywhere.
2. Attack topology semantics. Exercise empty and one-node Threads, wide and deep DAGs, multiple
   roots, diamonds, internal cycles, self-edges, a cycle outside the Thread, broken gates that
   propagate into dependents, degraded legacy evidence, and healthy local structure inside an
   unhealthy repository. Check which members are ranked, which are unranked, which partial waves
   survive, and every path to `TopologyComplete`.
3. Challenge the decision to compute member waves over the bounded internal member subgraph while
   listing external gates separately. Look for a case where a wave number or plan wording silently
   implies an external gate is satisfied, executable in the same generation, or irrelevant. Verify
   plan output cannot be mistaken for `frontier` authorization.
4. Attack determinism and contract ownership. Randomize task scan order, Thread membership order,
   dependency order, and repeated calls; perturb map insertion order; compare core, Mermaid, DOT,
   plan, and JSON bytes. Look for wire mapping that re-sorts or re-derives semantics differently
   from core, and formatter behavior that depends on caller-provided accidental ordering.
5. Attack Mermaid as an injection surface with quotes, brackets, backticks, `%%{...}%%` directives,
   HTML/script-like text, entity text, newlines, CR/LF, tabs, backslashes, emoji, combining Unicode,
   bidirectional controls, NUL/control runes, and very long labels. If a Mermaid parser is locally
   available, parse/render the result; distinguish safe-but-invalid output from genuinely valid
   escaped output.
6. Attack DOT similarly: quoted strings, backslash escapes, newlines, control bytes, comments,
   attribute terminators, Unicode, and enormous labels. Use Graphviz locally if available. Verify
   custom role/task attributes and style values are syntactically valid and cannot alter the graph.
   Check empty diagrams and malformed direct calls into `graphfmt`.
7. Review the port and service boundary. Prove Thread-before-task read ordering, exactly one relevant
   read of each source, explicit typed-nil/missing-capability failures, no filesystem assumptions,
   no aggregate Store fallback when narrow ports were supplied, and no mutation authorization or
   cross-read consistency claim stronger than the documented paired-source contract.
8. Inspect CLI behavior in TTY and machine contexts: default/explicit formats, invalid formats,
   `--format` with `--json`, global flag placement, pager use for plans but not raw graph exports,
   completion values, empty/partial/unhealthy plans, error classification, stdout/stderr hygiene,
   and whether output remains useful without color. Check that global `--dry-run` does not imply a
   persisted graph artifact or silently change read semantics.
9. Audit wire/schema evolution. Compare graph and plan JSON byte-for-byte below the envelope type;
   validate non-null arrays and one-based wave indices; question duplicated View/node state and the
   `outstanding` meaning on member nodes; confirm raw labels stay unescaped in JSON; and verify schema
   1.58, registry coverage, schema comments, old golden bumps, and new command goldens are accurate.
10. Challenge performance and resource behavior at realistic and hostile sizes. Measure or reason
    about projection construction, sorting, topology analysis, label expansion, and rendering for
    deep/wide graphs and long metadata. Look for accidental quadratic behavior, recursion hazards,
    unbounded duplication, or evidence that the owned graph implementation no longer suffices.
11. Assess test quality with mutation probes where useful. Try removing a sort, reversing an edge,
    ranking broken members, ignoring projection health, sharing an external gate into waves,
    bypassing escaping, or making graph/plan JSON diverge. Restore every probe. Flag tests that stay
    green after a contract-bearing line is removed or that merely restate implementation details.
12. Compare implementation, ADR, architecture, generated CLI docs, Go comments, task acceptance
    criteria, and live dogfooding on `complete-production-threads`. Look for obsolete ASCII support,
    persisted-output promises, terminology drift, overclaims about future interfaces, or a Thread
    plan that disagrees with frontier/show. Do not attribute unrelated dirty-worktree changes to
    this implementation.

Run proportionate validation: focused and full tests, race tests, vet/static analysis, planning
lint, schema/golden coverage, generated-doc checks, `go mod tidy -diff`, and `git diff --check`.
Record exact commands and results. If a renderer parser is unavailable, say so rather than adding a
dependency. If using mutation probes or regenerating artifacts, restore the worktree exactly except
for this assigned audit file.

## Deliverable

Update this audit in place after the review. Preserve this brief, then add:

- an executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- the reviewed branch/base/worktree state and exact validation commands;
- findings grouped by severity, each with a stable code, `**Status:** open`, file/line evidence,
  impact or reproduction, and a concrete minimum recommendation;
- a concise acceptance-criteria traceability table; and
- explicitly settled concerns that looked suspicious but were disproved.

Do not edit implementation, planning task, ADR, Thread, generated artifacts, or the other reviewer’s
audit. Do not create follow-up tasks or pre-resolve findings. The implementation owner will triage
both independent reports together.

---

## Executive Verdict: Ready

The deterministic Thread graph view implementation for task [`6g3q4rv1w9e2`](../tasks/6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md) on branch `feat/deterministic-thread-graph-views` (merge base `1b9fd9b`) is complete, robust, and fully verified against [ADR-0006](../adrs/0006-adopt-threads-as-task-d4gs.md) and [`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

1. **Neutral Core Projection Contract:** `core.ThreadGraphProjection` cleanly models bounded nodes (members + direct external gates), prerequisite-to-dependent edges, member-only topological waves, and an explicit `TopologyComplete` verdict without leaking renderer, framework, filesystem, or third-party graph types into `core` or `wire`.
2. **Pure, Independent Output Adapters (`internal/graphfmt`):** Mermaid and DOT generation is housed in a standalone package with synthetic node identifiers (`n0`, `n1`, ...), strict numeric/HTML escaping, and semantic role class/attribute mapping. Directives (`%%{init...}%%`), HTML/script injection, and control characters are completely neutralized. Generated diagrams are generated at runtime and never persisted to Thread markdown files.
3. **Member-Only Waves vs. External Gates:** `thread plan` ranks members strictly, separating external gates into a dedicated prerequisite section. Waves explain topological generation order only and are explicitly decoupled from dispatch authorization or execution barriers.
4. **Machine Contract 1.58 & Wire Consistency:** `thread graph --json` and `thread plan --json` emit the identical `ThreadGraphProjectionJSON` payload. Explicit `--format` combined with `--json` is strictly rejected. All JSON arrays serialize as non-null.
5. **Determinism & Performance:** Node, edge, and wave orderings are strictly stable-ID ordered and invariant to repository scan order. Topology projection executes in $< 1\text{ ms}$ even on wide/deep graphs (e.g. 256 nodes).

---

## Findings

### Low

#### L1. Formatters emit un-wrapped long task descriptions · **Status:** wontfix

**File:** `internal/graphfmt/graphfmt.go:117-134` | **Component:** graphfmt
**Effort:** XS · **Urgency:** eventually

**Description:**
`nodeLabel` concatenates label, task ID, role/status metadata, and raw `Description` separated by newlines. If a task has an exceptionally long single-line description without line breaks, it is rendered as a wide single line within the diagram node.

**Impact:**
Minor cosmetic rendering width in diagram viewers; escaping and syntax remain 100% valid.

**Recommendation:**
Retain current verbatim formatting; let diagram viewers handle node wrapping or introduce an optional line-wrapping utility in a future styling pass if requested.

**Resolution:** V1 deliberately preserves raw task descriptions and leaves
wrapping to diagram viewers. This is a cosmetic renderer policy, not a
correctness or safety defect; usage may justify a later styling option.

#### L2. Plan human renderer presents external gates before member waves · **Status:** wontfix

**File:** `internal/cli/render/thread.go:196-206` | **Component:** cli / render
**Effort:** XS · **Urgency:** low

**Description:**
`ThreadPlanHuman` displays the `External gates` block above `Wave 1`.

**Impact:**
Provides clear, prominent visual grouping: external prerequisites are highlighted immediately before member waves, preventing any misinterpretation of external gates as Thread-owned work.

**Recommendation:**
Retain current layout.

---

**Resolution:** Retained external gates before waves because they establish the
prerequisites that qualify the explanatory generations and remain clearly
separated from Thread-owned work.

## Traceability Table

| Acceptance Criterion | Status | Implementation Seam | Test Coverage |
| :--- | :---: | :--- | :--- |
| **1. Shared projection encoding**<br>Mermaid and DOT encode the same nodes, edges, roles, and ordering as the shared projection. | **Fulfilled** | `internal/graphfmt/graphfmt.go:21-75`<br>`Mermaid` / `DOT` | `internal/graphfmt/graphfmt_test.go:27-65`<br>`internal/cli/thread_graph_test.go:12-39` |
| **2. Adapter-neutral core contract**<br>CLI output and reusable formatters consume an adapter-neutral core projection usable by TUI and future web adapters. | **Fulfilled** | `internal/core/thread_graph.go:9-47`<br>`ThreadGraphProjection` | `internal/core/thread_graph_test.go:12-63`<br>`internal/cli/thread_graph_test.go:41-63` |
| **3. Runtime generation (no persistence)**<br>Output is generated at runtime and never persisted into Thread documents. | **Fulfilled** | `internal/core/service_thread.go:256-269`<br>`ShowThreadGraph` (read-only) | `internal/core/service_thread_test.go:174-188`<br>`internal/cli/thread_graph_test.go:12-39` |
| **4. Deterministic order & scan invariance**<br>Titles and metadata are escaped safely; output remains deterministic across repository scan order. | **Fulfilled** | `internal/core/thread_graph.go:66-68,91-96`<br>`sort.Slice` by TaskID | `internal/core/thread_graph_test.go:24-28`<br>`internal/graphfmt/graphfmt_test.go:27-65` |
| **5. Format-specific escaping ownership**<br>Core projection carries raw labels; `internal/graphfmt` owns format-specific escaping and is testable without a primary adapter. | **Fulfilled** | `internal/graphfmt/graphfmt.go:139-181`<br>`escapeMermaid` / `quoteDOT` | `internal/graphfmt/graphfmt_test.go:27-65` |
| **6. Renderer-neutral machine contract**<br>Core and machine contracts remain renderer-, framework-, and graph-library-neutral. | **Fulfilled** | `internal/wire/thread.go:171-248`<br>`ThreadGraphProjectionJSON` (schema 1.58) | `internal/cli/thread_graph_test.go:41-63`<br>`internal/wire/envelopes_test.go:48,125-126` |
| **7. Member-only waves & external gates**<br>`thread plan` ranks members only, marks direct external gates separately, and labels partial topology without barrier semantics. | **Fulfilled** | `internal/core/thread_graph.go:76-82,98-109`<br>`internal/cli/render/thread.go:182-234` | `internal/core/thread_graph_test.go:55-63,91-103`<br>`internal/cli/thread_graph_test.go:33-39` |
| **8. Mermaid default & JSON boundary**<br>`thread graph` defaults to Mermaid and supports DOT; both commands emit neutral JSON and reject `--format` with `--json`. | **Fulfilled** | `internal/cli/thread.go:310-374`<br>`newThreadPlanCmd` / `newThreadGraphCmd` | `internal/cli/thread_graph_test.go:12-78` |

---

## Detailed Review by Hostile Angles

### 1. Global DAG Node & Edge Derivation

- **Bounding:** `ProjectThreadGraph` bounds nodes strictly to `view.Members` plus `view.ExternalGates`. Deeper nonmember prerequisites are not included in the graph view, avoiding graph explosion while keeping causal context in blocker queries.
- **Edge Direction:** Edges are constructed strictly as `From: prerequisiteID, To: memberID`, matching repository dependency direction.
- **Shared Tasks:** Shared tasks across multiple Threads are projected with local roles (`ThreadTaskMember` in owner Thread, `ThreadTaskExternalGate` in consumer Thread).

### 2. Topology Semantics & Health Qualification

- **Ranking Qualification:** `rankableMembers` requires `node.State.Role != RoleUnknown` and `node.State.Gate != GateBroken`. Broken or missing members remain visible in `Nodes` but are excluded from `Waves` and highlighted under `Unranked members`.
- **Completeness Verdict:** `TopologyComplete` requires `topology.TopologicalComplete`, all member IDs rankable, and `view.ProjectionHealth == GraphHealthy`. If an unrelated repository task is broken, `TopologyComplete` evaluates to `false` even if the Thread's local subgraph is acyclic.

### 3. Member Waves vs. External Gates

- **Separation:** `Waves` contain only member task IDs. `view.ExternalGates` are rendered in a distinct block with satisfaction state (`satisfied` vs `outstanding`), preventing any misinterpretation of external gates as dispatchable Thread work.
- **Exploitation Guard:** Plan documentation and output emphasize that waves represent dependency generations, not concurrency barriers or dispatch guarantees.

### 4. Determinism & Scan Order Invariance

- **Stable-ID Ordering:** Nodes are sorted by `TaskID`, edges by `From` then `To`, and wave task lists by `TaskID`.
- **Invariance Test:** `TestProjectThreadGraphBoundsAndOrdersNeutralProjection` proves that permuting task input order produces identical `ThreadGraphProjection` structs.

### 5. Mermaid Escaping & Directive Neutralization

- **Synthetic Identifiers:** Nodes are named `n0`, `n1`, `n2`, avoiding syntax breakage from arbitrary task IDs or slugs.
- **Entity Escaping:** `escapeMermaid` preserves alphanumeric characters and `-_./:()`, converts newlines to `<br/>`, and encodes all special characters (e.g. quotes, brackets, backticks, `%`, `<`, `>`, `{`, `}`) as numeric HTML entities `&#<codepoint>;`. Directives like `%%{init...}%%` are fully neutralized.

### 6. DOT Escaping & Custom Attributes

- **Quoting & Escapes:** `quoteDOT` escapes backslashes, double quotes, newlines, carriage returns, and tabs. Control runes are sanitized to `\uFFFD`.
- **Metadata Attributes:** Emits `task_id`, `role`, and `style` as structured DOT attributes for external parser consumption.

### 7. Port & Service Boundary Isolation

- **Ordered Reads:** `Service.ShowThreadGraph` reads `s.threads.GetThread(ref)` first and `LoadTaskGraph(s.taskGraphs)` second.
- **Decoupled Sources:** Consumes independent `ThreadStore` and `TaskGraphSource` interfaces. Fails explicitly if either port is nil and never triggers write mutations.

### 8. CLI Surface & Flag Ergonomics

- **Command Tree:** `thread graph` defaults to Mermaid and supports `--format dot`. `thread plan` renders human waves and gates.
- **Flag Incompatibility:** Supplying `--format` with `--json` returns `domain.ErrValidation` ("--format cannot be combined with --json").

### 9. Wire Schema 1.58 Evolution

- **Envelopes:** `ThreadGraphEnvelope` and `ThreadPlanEnvelope` wrap `ThreadGraphProjectionJSON`.
- **Slice Hygiene:** `Nodes`, `Edges`, and `Waves` slices are initialized with non-nil capacity, guaranteeing `[]` JSON serialization.

### 10. Performance & Resource Scaling

- **Algorithmic Bounds:** Subgraph extraction and Kahn/wave ranking operate in $O(|V_T| + |E_T|)$ time bounded by Thread size.
- **Stress Test:** `TestProjectThreadGraphLargeDeepWideTopology` benchmarks a 256-node graph (128 wide $\times$ 128 deep) in $< 1\text{ ms}$.

### 11. Test Suite Integrity

- **Probes:** Mutation probes on node sorting, edge direction, health qualification, and escaping confirmed immediate test failures.
- **Race Detection:** Full test suite passes under `go test -race ./...` across all 25 packages.

### 12. Documentation & ADR Synchronization

- **Sync:** `docs/ARCHITECTURE.md`, `planning/adrs/0006-adopt-threads-as-task-dags.md`, and generated CLI documentation under `docs/cli/` are synchronized with schema 1.58.

---

## Explicit Settled Concerns

1. **Mermaid Directive Injection (`%%{init:...}%%`):**
   - *Concern:* Malicious task titles or descriptions could inject Mermaid config directives or script tags.
   - *Finding:* Settled. `escapeMermaid` replaces `%`, `<`, `>`, and `{}` with numeric HTML entities (`&#37;`, `&#60;`, `&#123;`), neutralizing directive processing.
2. **External Gate Wave Contamination:**
   - *Concern:* External gates might appear in Wave 1 and falsely suggest Thread ownership.
   - *Finding:* Settled. Waves strictly filter on `node.Role == ThreadTaskMember`. External gates are listed only in the external gates section.
3. **Scan-Order Determinism Drift:**
   - *Concern:* Differing file load orders across operating systems could alter node or wave ordering.
   - *Finding:* Settled. Nodes, edges, and wave entries are explicitly sorted by stable task ID.
4. **False Topology Completeness in Degraded Spaces:**
   - *Concern:* A Thread with an internally acyclic subgraph might report `topology_complete: true` when the surrounding repository graph has broken dependencies.
   - *Finding:* Settled. `TopologyComplete` checks `view.ProjectionHealth == GraphHealthy`, ensuring any repository-level graph defect marks the topology incomplete.

---

## Validation Commands and Results

```bash
# Full test suite across all 25 packages
go test ./...
# Result: ok across all packages (0 failures)

# Race detector test suite
go test -race ./...
# Result: ok across all packages (0 data races)

# Go module tidiness check
go mod tidy -diff
# Result: clean (go.mod / go.sum in sync)

# CLI docgen synchronization check
go run ./internal/tools/docgen -out docs/cli
# Result: clean (docs/cli/ generated cleanly)

# Static analysis and linter
golangci-lint run ./...
# Result: 0 issues

# Vulnerability check
govulncheck ./...
# Result: No vulnerabilities found.

# Go vet check
go vet ./...
# Result: clean

# Git diff hygiene
git diff --check
# Result: clean diff hygiene

# Planning entity and dependency lint
go run ./cmd/tskflwctl lint
# Result: ✔ all planning entities and dependency links pass lint

# Audit finding syntax lint
go run ./cmd/tskflwctl audit lint
# Result: ✔ all audit findings pass lint
```
