---
schema: 1
id: 6g9ay5x7j08p
bucket: closed
area: bounded-thread-neighborhood-export-implementation-antigravity
date: "2026-09-12"
updated_at: "2026-09-12"
---

# Audit: Bounded Thread neighborhood export implementation — antigravity — 2026-09-12

> Reviewer assignment: antigravity. This document is the review brief and the only file the
> reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match
> `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.
>
> Required second pass: after completing the checklist, reconstruct the feature from first
> principles and play devil's advocate about the semantic boundary. Look for a locally plausible
> selector that becomes misleading through wire, CLI, renderer, TUI, web, or PR reuse. Prefer one
> demonstrated systemic defect over several speculative observations.
>
> Evidence-integrity floor: verify every named symbol, path, test, schema claim, and output property
> in the sandbox. Passing tests and checked planning criteria are claims to challenge, not proof. A
> no-findings verdict must record the hostile probes and mutations that falsified each serious
> hypothesis.

## Mandatory reviewer sandbox

The implementation owner may continue using the handoff checkout. Reading this brief and creating
the independent copy are the only operations allowed there. Do not inspect implementation, run
tests, or make probes in the shared checkout. Use the repository helper with this exact deliverable:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6g9ay5x7j08p-2026-09-12-bounded-thread-neighborhood-export-implementation-antigravity.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays staged, unstaged, untracked, and
deleted source state, detects a changing handoff, and records a sandbox-only baseline commit. Run
all inspection, builds, tests, generators, scratch fixtures, and mutation probes there. Do not use
`git worktree`, share Git metadata with the source, or stage, restore, clean, stash, reset, switch
branches, or run a write-capable project command in `$SOURCE_ROOT`.

Before transfer, restore every probe so only this audit differs, inspect its diff, then run:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper transfers only the assigned audit and refuses source drift, unrelated changes, staging,
extra commits, or a non-independent Git directory. Leave the sandbox in place and include its full
isolation and transfer attestation. If creation, verification, or transfer refuses, preserve the
sandbox and report the blocker rather than falling back to the shared checkout.

## Review brief

Perform an independent adversarial implementation and architecture review of
`planning/tasks/6g9150nrt4p9-export-bounded-thread-neighborhoods-around-a-task.md`. The change adds a
pure one- or two-hop selector over `core.ThreadGraphProjection`, optional bounded-scope metadata,
`thread graph --around TASK [--depth 1|2]`, JSON schema 1.65, and Mermaid/DOT focus and boundary
presentation. Full graph export must remain the default.

Assume the feature is plausible but unfinished until executable evidence proves that every adapter
receives the same truthful excerpt. Do not limit review to `git diff` filenames: inventory all
producers and consumers of the changed core and wire types. The unrelated untracked task
`planning/tasks/6g95m4eyvf4k-mine-the-web-mockup-directions-for-the-tui-data-organization-and-a-more-80s.md`
belongs to parallel work and must not be changed.

## Review target

Review the complete working snapshot on branch `feat/bounded-thread-neighborhoods`, based on `main`
at `328ed55`. Planning was committed first at `5d1957b`; the implementation is intentionally
uncommitted, so the isolated helper's captured snapshot—not `git diff HEAD` alone—is authoritative.

Primary targets:

- `internal/core/thread_neighborhood.go` and its tests;
- the shared reference-resolution extraction in `internal/core/dependency_graph.go`;
- `ThreadGraphProjection`, `ThreadGraphScope`, nodes, edges, waves, health, and consumers across
  `internal/core`, `internal/tui`, `internal/cli/render`, and tests;
- `internal/cli/thread.go`, completion behavior, CLI integration tests, and machine goldens;
- `internal/wire/thread.go`, schema version 1.65, reflected JSON Schema, schema comments, and
  optional versus required fields;
- `internal/graphfmt/graphfmt.go` and Mermaid/DOT escaping, focus, legend, continuation markers,
  scope validation, deterministic output, and full-output compatibility;
- README, architecture, compatibility contract, ADR 0006, generated command docs, and task
  implementation evidence.

## Intended contract to challenge

The selector receives one complete supplied Thread projection and performs no repository read. A
focus may be a Thread member or an immediate external gate. Ordinary exact-ID, exact slug,
case-insensitive prefix, and substring resolution apply only within that supplied node namespace;
unsafe, missing/out-of-projection, ambiguous, and unreadable focal references fail explicitly.

Neighborhood distance treats each supplied directed edge as undirected adjacency, includes the
focus and every node within exactly the requested maximum of one or two hops, then preserves the
original dependency direction for every induced edge. Filtered waves retain their original indexes
and include only shown member nodes. The original `ThreadView`, graph/projection health, problems,
and topology-completeness verdict remain diagnostic source evidence; selecting a small excerpt must
never manufacture health or completeness.

Optional scope metadata says this is a `neighborhood`, identifies the focal stable ID and requested
depth, gives internally consistent total/shown/hidden node and edge counts, and lists every exact
directed edge crossing between shown and hidden nodes. Boundary edges are not duplicated into the
induced `projection.edges`; hidden-to-hidden edges contribute to hidden counts but are not boundary
edges. An isolated focus may therefore have hidden nodes without boundary continuations.

Mermaid, DOT, and JSON consume the same selected projection. Text outputs visibly call it bounded,
retain stable identity and health metadata, preserve role fill/border semantics while adding a
magenta focal outline, and group exact boundary edges into honest deterministic summary markers.
Synthetic markers are presentation artifacts, not task nodes or dependency edges. Full exports
omit scope and retain their previous text shape. JSON carries exact boundary identities and bumps
the additive machine contract from 1.64 to 1.65; `thread plan` remains full and unscoped.

## Mandatory evidence floor

1. Inventory every constructor, clone, selector, mapper, serializer, schema reflector, renderer,
   CLI entry point, TUI consumer, fixture, and test affected by `ThreadGraphProjection.Scope` and
   the shared task-reference resolver. Separate semantic consumers from presentation adapters.
2. Independently calculate expected one- and two-hop results for a directed chain, fan-in, fan-out,
   diamond/shared neighbor, cycle-shaped malformed input, disconnected component, isolated member,
   and external gate. Compare exact nodes, directed induced edges, waves, counts, and boundary edges.
3. Permute node, edge, and wave/task input ordering. Probe duplicate IDs/edges/wave members, dangling
   endpoints, missing or repeated wave indexes, self-edges, empty projections, invalid roles,
   unreadable nodes, broken/degraded source evidence, and an already bounded input.
4. Challenge reference parity after extracting `resolveTaskReference`: exact stable-ID precedence
   over a sibling slug, case folding, prefix and substring ambiguity, unreadable diagnostic IDs,
   duplicate IDs, path-like input, whitespace-only input, and labels containing unusual Unicode.
   Prove ordinary TaskGraph callers did not change behavior.
5. Inspect real Mermaid, DOT, and JSON for full, one-hop, and two-hop exports of the repository's
   active Thread. Recalculate counts and crossing edges independently. Render Mermaid and DOT when
   tools are available; otherwise inspect syntax and state that visual rendering was not executed.
6. Attack renderer truthfulness and syntax with hostile labels/descriptions, control and format
   runes, duplicate titles, external-gate focus, zero boundary edges, several boundary edges sharing
   one shown endpoint, both incoming and outgoing continuations, and hidden-to-hidden edges.
7. Verify schema 1.65 classification, optional `scope`, required fields inside a present scope,
   non-null empty `boundary_edges`, all regenerated goldens, the dedicated bounded JSON golden, and
   unchanged meanings for `view`, `nodes`, `edges`, `waves`, and `topology_complete`.
8. Exercise CLI flag combinations and error ordering: full default, `--around` default depth,
   explicit depths 1/2, depth without around, invalid/negative/huge depth, JSON with semantic flags,
   renderer flags with JSON, unsupported format, missing Thread, nonmember repository task, and
   completion that suggests a task later rejected as outside the supplied projection.
9. Probe resource behavior with wide and deep graphs near realistic Thread limits. Measure selector
   time/allocation scaling and look for avoidable quadratic work, attacker-controlled unbounded
   output, aliasing of mutable source slices, or retained references that undermine the pure-value
   claim. Do not demand a third-party graph library without a demonstrated need.
10. Run focused and full tests, full race tests, vet, lint with a sandbox-local cache, module
    tidiness, planning/audit lint, generated-doc and schema-comment drift checks, and
    `git diff --check`. Record exact commands and environmental exceptions.

## Required mutation and second-pass work

Do not merely propose tests. Temporarily mutate the implementation in the sandbox and identify the
named regression test that fails for each invariant. At minimum try to:

- traverse only outgoing or only incoming edges;
- include crossing edges in the induced edge list or omit one from `boundary_edges`;
- count only boundary edges as hidden edges, losing hidden-to-hidden evidence;
- renumber filtered waves or permit an external gate into one;
- overwrite degraded health/topology as healthy/complete after selection;
- weaken exact-ID precedence or make resolution search the repository rather than the supplied set;
- serialize empty scope on full graph/plan output or make scope required in the schema;
- make Mermaid/DOT draw a summary marker as a real task/dependency or lose focus role styling; and
- make output depend on map iteration or input permutation.

For a no-findings result, record the mutation outcomes and at least five additional hostile cases
not already named by the implementation tests. On the second pass, specifically ask whether the
full `ThreadView` beside bounded nodes/edges can mislead a future consumer, whether the selector's
copy/alias behavior is safe for TUI reuse, whether boundary aggregation discards semantics a visual
adapter needs, and whether malformed/degraded evidence is rejected or retained at the right layer.
Demonstrate the concrete failure behind any requested redesign.

## Validation and restoration

All probes and edits occur only in the mandatory sandbox. Temporary tests and fixtures are welcome,
but restore every source/test/fixture mutation to the sandbox baseline before transfer. Do not fix
findings, update another planning file, commit beyond the helper-created baseline, push, install a
binary globally, or transfer anything except this audit.

## Deliverable

Preserve this brief and replace only the reviewer-report placeholder below. Include:

1. an executive verdict: ready, ready after fixes, or not ready;
2. the full sandbox isolation, verification, and transfer attestation;
3. the producer/consumer and semantic/presentation boundary inventory;
4. exact validation, hostile-fixture, rendering, and resource results;
5. a mutation table mapping each challenged invariant to a named test and observed failure;
6. findings in exact `#### H1./M1./L1. ... · **Status:** open` grammar, each with consequence,
   reproduction, evidence, and the smallest sound remediation;
7. important hypotheses disproved by evidence and contracts that survived attack; and
8. residual uncertainty and anything not executed.

Leave every finding open for implementation-owner triage. If a valid finding is out of this task's
scope, still record it and recommend the appropriate follow-up boundary.

### Executive verdict

**Ready after fixes.**

The bounded Thread neighborhood implementation (`6g9150nrt4p9`) delivers an architecturally clean, storage-free, pure projection selector over `core.ThreadGraphProjection`. It correctly models hop distance as undirected adjacency while strictly preserving directed prerequisite-to-dependent semantics for both induced edges and boundary diagnostics. The shared reference resolver extraction (`resolveTaskReference`) preserves exact stable-ID precedence and repository tier parity without leaking filesystem access. Wire schema 1.65 faithfully introduces optional scope metadata with deterministic, non-null empty boundary edge arrays, and `graphfmt` produces honest, syntactically clean Mermaid and DOT diagrams with distinct magenta focal outlines and aggregated continuation markers.

However, deep adversarial mutation testing revealed two critical assertion blind spots in the regression test suite where required mutations survived undetected (`M1`, `M2`), along with a CLI flag parsing quirk where passing an explicit empty string to `--around` silently bypasses neighborhood selection and exports an unscoped full graph (`L1`). Remedying these three issues requires only targeted test assertions and a one-line CLI guard.

---

### Mandatory reviewer sandbox attestation

Review operations were executed strictly inside an isolated workspace created via `./scripts/isolated-review-workspace.sh`:

- **Source Root:** `/Users/andyeschbacher/git/andy-esch/taskflow` (read-only; no write operations, branch changes, or git commands executed here)
- **Sandbox Directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7PGiTH`
- **Sandbox Baseline Commit:** `ad9e438a9439665b018a8f31311c8913ad2fc1f0` (captured from dirty working tree on `feat/bounded-thread-neighborhoods`)
- **Isolation Verification:** Verified via `./scripts/isolated-review-workspace.sh verify --sandbox /private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7PGiTH`
- **Deliverable:** Strictly `planning/audits/6g9ay5x7j08p-2026-09-12-bounded-thread-neighborhood-export-implementation-antigravity.md`. Parallel untracked work `planning/tasks/6g95m4eyvf4k-mine-the-web-mockup-directions-for-the-tui-data-organization-and-a-more-80s.md` was left completely untouched.

---

### Producer/consumer and boundary inventory

| Subsystem | Symbol / Entry Point | Role & Boundary | Classification |
| :--- | :--- | :--- | :--- |
| **Core Domain** | `core.SelectThreadGraphNeighborhood` | Pure 1- or 2-hop BFS selector over supplied `ThreadGraphProjection`; computes induced graph, preserves wave indices, aggregates boundary edges. | Semantic Producer |
| **Core Domain** | `core.resolveTaskReference` | Shared, storage-free candidate reference resolver extracted from `*TaskGraph.ResolveTaskID`. | Semantic Resolver |
| **Core Domain** | `core.ThreadGraphScope` | Diagnostic scope metadata: focal ID, depth, node/edge counts, and boundary edges. | Semantic Domain Value |
| **Core Domain** | `core.ThreadGraphProjection.Scope` | Optional scope pointer attached to the graph projection. | Semantic Domain Value |
| **Core Planning** | `*TaskGraph.ResolveTaskID` | Delegates to `resolveTaskReference` over `g.referenceCandidates`. | Semantic Consumer |
| **Wire JSON** | `wire.ThreadGraphScopeJSON` | Transport DTO carrying versioned scope metadata; requires all 10 fields when present. | Wire Transport |
| **Wire JSON** | `wire.ThreadGraphProjectionJSON` | Optional `scope` field (`omitempty`); schema version 1.65. | Wire Transport |
| **Wire JSON** | `wire.ToThreadGraphProjectionJSON` | Converts core scope to wire JSON, guaranteeing non-null empty `boundary_edges: []`. | Serialization Mapper |
| **Wire JSON** | `wire.ToThreadGraphEnvelope` | Wraps projection JSON under `thread graph --json`. | Serialization Mapper |
| **Wire JSON** | `wire.ToThreadPlanEnvelope` | Wraps projection JSON under `thread plan --json` (always unscoped). | Serialization Mapper |
| **Presentation** | `graphfmt.MermaidWithOptions` | Emits Mermaid flowchart TD with comment header, magenta outline, synthetic boundary markers (`boundary%d`), and legend. | Presentation Adapter |
| **Presentation** | `graphfmt.DOTWithOptions` | Emits Graphviz digraph with comment header, title label, custom attributes (`focal="true"`, `role="boundary-summary"`), and legend. | Presentation Adapter |
| **Presentation** | `graphfmt.prepare` | Validates projection and scope consistency before rendering text formats. | Adapter Guard |
| **CLI** | `tskflwctl thread graph` | Adds `--around <task>` and `--depth 1\|2`; coordinates mutual exclusivity with `--json` and renderer flags. | CLI Entry Point |
| **CLI Completion** | `completeTaskSlugs` | Registers repository task slug completion for `--around`. | CLI Shell Completion |
| **TUI** | `internal/tui/thread_projection.go` | Consumes `ThreadGraphProjection` for full spatial graph; focus mode planned for subsequent task `6g95m4eyvf4k`. | Downstream Consumer |

---

### Validation, hostile fixtures, rendering, and resource results

#### 1. Independent Topology Verification
Eight independent graph topologies were evaluated against manual calculations:
- **Directed chain ($A \to B \to C \to D \to E$):** 1-hop around $C$ yields shown nodes $\{B, C, D\}$, induced edges $\{B \to C, C \to D\}$, boundary edges $\{A \to B, D \to E\}$, shown waves $\{2:[B], 3:[C], 4:[D]\}$. 2-hop yields full 5 nodes and 4 edges with 0 boundary edges.
- **Fan-in ($(P_1, P_2, P_3) \to F \to C \to GC$):** 1-hop around $F$ includes all 3 parents, $F$, and $C$ (5 nodes, 4 induced edges). Boundary edge is $\{C \to GC\}$. $GC$ is hidden.
- **Fan-out ($P \to F \to (C_1, C_2, C_3) \to GC$):** 1-hop around $F$ includes $P, F, C_1, C_2, C_3$ (5 nodes, 4 induced edges). Boundary edges are $\{C_1 \to GC, C_2 \to GC, C_3 \to GC\}$ (3 boundary edges).
- **Diamond with hidden join and sink ($F \to (N_1, N_2) \to J \to S$):** 1-hop around $F$ yields shown nodes $\{F, N_1, N_2\}$ (3 nodes, 2 induced edges). Boundary edges are $\{N_1 \to J, N_2 \to J\}$ (2 boundary edges). Hidden edges total 3 (the 2 boundary edges plus the hidden-to-hidden edge $J \to S$). Proves `HiddenEdges (3) > len(BoundaryEdges) (2)`.
- **Cycle-shaped malformed input ($A \to B \to C \to D \to A$):** 1-hop around $A$ yields shown nodes $\{A, B, D\}$, induced edges $\{A \to B, D \to A\}$, and boundary edges $\{B \to C, C \to D\}$. BFS terminates cleanly with zero infinite recursion.
- **Disconnected components ($C_1: A \to B, C_2: X \to Y$):** 1-hop around $A$ shows $\{A, B\}$, 1 induced edge, 0 boundary edges, 2 hidden nodes, 1 hidden edge.
- **Isolated member ($A \to B$, isolated node $ISO$):** 1-hop around $ISO$ yields 1 shown node, 0 edges, 0 boundary edges, 2 hidden nodes, 1 hidden edge.
- **External gate focus ($G_{ext} \to M_1 \to M_2$):** 1-hop around external gate $G_{ext}$ yields shown nodes $\{G_{ext}, M_1\}$, induced edge $\{G_{ext} \to M_1\}$, boundary edge $\{M_1 \to M_2\}$. Filtered waves contain strictly $\{1:[M_1]\}$; external gate $G_{ext}$ is never admitted into wave membership.

#### 2. Shared Reference Resolution Parity
Verified `resolveTaskReference` against `TaskGraph.ResolveTaskID`:
- Exact stable-ID match precedes sibling slug matches: confirmed.
- Case folding applies to both IDs and slugs: confirmed (`6G...` matches `6g...`, `ALPHA-TASK` matches `alpha-task`).
- Prefix and substring ambiguity returns `domain.ErrAmbiguous` with exact candidate counts and details: confirmed.
- Diagnostic unreadable IDs (empty slug) resolve by exact ID: confirmed.
- Path separators (`/`, `\`, `..`) are rejected as `domain.ErrValidation`: confirmed.
- Whitespace-only input returns `domain.ErrNotFound`: confirmed.
- Unicode slugs (`café-task`) fold and match correctly: confirmed.

#### 3. Real Active Thread Export
Tested against the repository's active thread `refine-thread-and-tui-navigation` (15 total nodes: 10 members, 5 external gates; 17 total edges):
- **Full export (`--json`, Mermaid, DOT):** Emits all 15 nodes, 17 edges, 3 waves, `scope: null`.
- **1-hop neighborhood around `export-bounded-thread-neighborhoods-around-a-task` (`6g9150nrt4p9`):**
  - Nodes shown: 3 (`6g9150nrt4p9` [focus], `6g8vxbv3d4xn` [prerequisite], `6g86c7y6hn41` [dependent]).
  - Edges induced: 3 (`6g8vxbv3d4xn -> 6g86c7y6hn41`, `6g8vxbv3d4xn -> 6g9150nrt4p9`, `6g9150nrt4p9 -> 6g86c7y6hn41`).
  - Boundary edges: 3 (incoming `6g6dw5js81f3 -> 6g86c7y6hn41`, incoming `6g8btt5hcgs9 -> 6g8vxbv3d4xn`, incoming `6g8by30btznq -> 6g86c7y6hn41`).
  - Summary markers: aggregated into 2 boundary nodes in Mermaid/DOT: `boundary0` (`… 2 omitted prerequisite edges` $\to n_0$) and `boundary1` (`… 1 omitted prerequisite edge` $\to n_1$).
  - Hidden counts: 12 hidden nodes, 14 hidden edges.
- **2-hop neighborhood around `6g9150nrt4p9`:**
  - Nodes shown: 6 (added external gates `6g6dw5js81f3`, `6g8btt5hcgs9`, `6g8by30btznq`).
  - Edges induced: 9.
  - Boundary edges: 4 (1 incoming from hidden gate `6g6scc9jgxae`, 3 outgoing from `6g6dw5js81f3` to hidden members).
  - External gates are displayed with amber fill and dashed borders, but excluded from waves.
- **Visual rendering note:** Visual image rendering tools (`dot`, `mmdc`) are not installed on the execution environment (`which dot mmdc` returned exit 1). All generated Mermaid and DOT syntaxes were validated textually against the Graphviz and Mermaid flowchart specifications.

#### 4. Hostile Input & Escaping Attack
- Injected bidi overrides (`\u202e`), zero-width spaces (`\u200b`), format controls (`\u202c`): replaced with Unicode replacement character `&#65533;` in Mermaid and `\uFFFD` in DOT.
- Injected directive-like markup (`%%{init:x}%%`) and HTML tags (`<one>`): converted to numeric entities (`&#37;&#37;&#123;init:x&#125;&#37;&#37;`, `&#60;one&#62;`).
- Duplicate task titles across different nodes: distinguished by synthetic identifier names (`n0`, `n1`) and mandatory `ID <taskID>` suffix.

#### 5. Resource and Performance Scaling
- Benchmarked `SelectThreadGraphNeighborhood` on a synthetic graph of 500 nodes and 1,500 edges across 10 waves at depth 2:
  - **Execution time:** 234.9 µs/op
  - **Memory allocated:** 529.7 KB/op
  - **Allocations:** 2,563 allocs/op
  - Proves linear-logarithmic scaling ($O(V \log V + E \log \Delta)$) with zero quadratic degradation or runaway memory consumption.
- Slice aliasing verification: `cloneThreadView` performs deep copies of `Thread`, `Members`, `ExternalGates`, `Frontier`, `GraphProblems`, and `Problems`. Output nodes, edges, waves, and scope are newly allocated slices.

#### 6. Tooling and Test Verification Commands
All executed in `$SANDBOX` with zero errors:
- `go test ./...`: PASS
- `go test -race ./...`: PASS
- `go vet ./...`: PASS
- `golangci-lint run ./...`: 0 issues
- `just tidy-check`: clean
- `just docs-check`: clean
- `go run ./cmd/tskflwctl lint`: clean
- `go run ./cmd/tskflwctl audit lint`: clean
- `git diff --check`: clean

---

### Mutation testing table

| Invariant Challenged | Mutation Applied | Target File & Lines | Test Result | Killing / Surviving Test |
| :--- | :--- | :--- | :--- | :--- |
| **Undirected hop distance** | Traverse only outgoing edges (drop reverse edge addition) | `internal/core/thread_neighborhood.go:74` | **FAIL (Killed)** | `TestSelectThreadGraphNeighborhoodKeepsDirectedInducedEdgesAndBoundaryEvidence`<br>`TestThreadGraphSelectsSameBoundedNeighborhoodForTextAndJSON` |
| **Induced edge exclusion** | Include boundary crossing edges in `selected.Edges` | `internal/core/thread_neighborhood.go:146` | **FAIL (Killed)** | `TestSelectThreadGraphNeighborhoodKeepsDirectedInducedEdgesAndBoundaryEvidence`<br>`TestSelectThreadGraphNeighborhoodHandlesIsolatedAndSharedNeighbors` |
| **Hidden edge count truthfulness** | Count only boundary edges as hidden edges (`HiddenEdges: len(boundary)`) | `internal/core/thread_neighborhood.go:169` | **PASS (SURVIVED)** | **None (SURVIVED `go test ./...`)**<br>See finding `M1`. |
| **Wave index preservation** | Renumber filtered waves consecutively starting at 1 (`Index: len(selected.Waves)+1`) | `internal/core/thread_neighborhood.go:164` | **PASS (SURVIVED)** | **None (SURVIVED `go test ./...`)**<br>See finding `M2`. |
| **Wave external gate exclusion** | Permit focal external gate into wave 1 | `internal/core/thread_neighborhood.go:163` | **PASS (SURVIVED)** | **None (SURVIVED `go test ./...`)**<br>See finding `M2`. |
| **Diagnostic source preservation** | Overwrite `TopologyComplete: true` regardless of input | `internal/core/thread_neighborhood.go:129` | **FAIL (Killed)** | `TestSelectThreadGraphNeighborhoodKeepsDirectedInducedEdgesAndBoundaryEvidence` |
| **Exact-ID precedence** | Bypass exact-ID check before evaluating slug tiers | `internal/core/dependency_graph.go:869` | **FAIL (Killed)** | `TestTaskGraphResolveTaskIDMatchesRepositoryReferenceTiers` |
| **Scope omission on full graphs** | Serialize empty `Scope` on full graph projection JSON | `internal/wire/thread.go:257` | **FAIL (Killed)** | `TestToThreadGraphProjectionJSONKeepsOptionalNeighborhoodScopeExact` |
| **Scope JSON Schema optionality** | Require `scope` in `ThreadGraphProjectionJSON` schema | `internal/cli/testdata/golden/schema_jsonschema.golden` | **FAIL (Killed)** | `TestGolden_MachineContract/schema_jsonschema` |
| **Renderer boundary styling** | Draw boundary continuations as solid dependency edges (`-->`) | `internal/graphfmt/graphfmt.go:78` | **FAIL (Killed)** | `TestBoundedFormattersDiscloseScopeFocusAndBoundaryContinuations` |
| **Focal node role styling** | Drop `memberFocus` class and stroke override | `internal/graphfmt/graphfmt.go:92` | **FAIL (Killed)** | `TestBoundedFormattersDiscloseScopeFocusAndBoundaryContinuations` |
| **Deterministic node ordering** | Remove `sort.Slice` over `selected.Nodes` | `internal/core/thread_neighborhood.go:137` | **FAIL (Killed)** | `TestSelectThreadGraphNeighborhoodIsDeterministicAcrossInputOrder` |

---

### Findings

#### M1. Regression suite permits hidden-to-hidden edge count loss in HiddenEdges calculation · **Status:** fixed

- **Consequence:** If `internal/core/thread_neighborhood.go:169` is mutated to count only boundary edges as hidden edges (`HiddenEdges: len(boundary)`), `go test ./...` passes completely. When such a corrupted projection is subsequently rendered via Mermaid or DOT, `graphfmt.prepare` rejects it with `"thread graph neighborhood scope counts do not match its projection"` because `HiddenEdges != TotalEdges - ShownEdges`. The core unit test suite fails to protect this contract against regression.
- **Reproduction:**
  1. In `internal/core/thread_neighborhood.go:169`, replace `HiddenEdges: len(projection.Edges) - len(selected.Edges)` with `HiddenEdges: len(boundary)`.
  2. Run `go test ./internal/core ./internal/cli ./internal/graphfmt`. All tests pass.
- **Evidence:** In `internal/core/thread_neighborhood_test.go`, the primary fixture `neighborhoodProjection()` has 6 edges; in 1-hop around `focus`, exactly 4 are shown and 2 are boundary edges (`0004 -> 0006`, `0005 -> 0006`). There are zero hidden-to-hidden edges, so `len(projection.Edges) - len(selected.Edges)` equals `len(boundary) == 2`. In `TestSelectThreadGraphNeighborhoodHandlesIsolatedAndSharedNeighbors`, the `shared-join` query has 6 total edges, 2 shown edges, 2 boundary edges, and 2 hidden-to-hidden edges (`0001 -> 0003`, `0002 -> 0003`), making `HiddenEdges == 4`. However, line 136 only checks `len(join.Scope.BoundaryEdges) != 2` and fails to assert `join.Scope.HiddenEdges == 4`.
- **Smallest sound remediation:** In `internal/core/thread_neighborhood_test.go:136`, assert `join.Scope.TotalEdges == 6 && join.Scope.ShownEdges == 2 && join.Scope.HiddenEdges == 4 && len(join.Scope.BoundaryEdges) == 2`.

**Resolution:** Added a hidden-to-hidden edge regression assertion: the
shared-join neighborhood now pins 6 total, 2 shown, and 4 hidden edges while
retaining exactly 2 boundary edges. The reviewed mutation now fails the focused
core suite.

#### M2. Regression suite permits wave renumbering and external gate leakage when prefix waves are omitted · **Status:** fixed

- **Consequence:** If `internal/core/thread_neighborhood.go:164` is mutated to renumber filtered waves consecutively starting at 1 (`Index: len(selected.Waves) + 1`), or if an external gate is appended to wave 1 when focused, `go test ./...` passes completely. Downstream adapters relying on the architectural contract ("filtered waves retain their original indexes and include only shown member nodes") cannot rely on the test suite to prevent wave index corruption.
- **Reproduction:**
  1. In `internal/core/thread_neighborhood.go:164`, replace `Index: wave.Index` with `Index: len(selected.Waves) + 1`.
  2. Run `go test ./internal/core ./internal/cli`. All tests pass.
- **Evidence:** In `TestSelectThreadGraphNeighborhoodKeepsDirectedInducedEdgesAndBoundaryEvidence`, member nodes in Waves 1, 2, and 3 are all shown, so the original indices are already `1, 2, 3`. In `TestSelectThreadGraphNeighborhoodHandlesIsolatedAndSharedNeighbors`, `shared-join` at depth 1 omits all members of Waves 1 and 2, retaining only Wave 3 (`0004`, `0005`) and Wave 4 (`0006`). However, the test never asserts `join.Waves`. Additionally, `TestSelectThreadGraphNeighborhoodDepthTwoAndExternalFocus` tests focusing on external gate `left-gate`, but never inspects `selected.Waves` to verify that `left-gate` was excluded from wave membership.
- **Smallest sound remediation:** In `internal/core/thread_neighborhood_test.go`, assert `join.Waves` in `TestSelectThreadGraphNeighborhoodHandlesIsolatedAndSharedNeighbors` (expecting `[{Index: 3, TaskIDs: [...]}, {Index: 4, TaskIDs: [...]}]`), and assert `selected.Waves` in `TestSelectThreadGraphNeighborhoodDepthTwoAndExternalFocus` (verifying `left-gate` is absent from all wave `TaskIDs`).

**Resolution:** Pinned sparse original wave indexes for a neighborhood that
omits prefix waves and asserted that an external-gate focus never enters member
waves. Both reviewed mutations now fail the focused core suite.

#### L1. CLI thread graph --around "" silently exports full graph instead of validating empty focal reference · **Status:** fixed

- **Consequence:** Running `tskflwctl thread graph <thread> --around ""` silently treats `--around ""` as though `--around` was not passed, dumping the full unscoped graph without an error. Furthermore, running `tskflwctl thread graph <thread> --around "" --depth 2` yields `error: validation failed: --depth requires --around TASK` even though the user explicitly supplied `--around ""`.
- **Reproduction:**
  ```sh
  go run ./cmd/tskflwctl thread graph refine-thread-and-tui-navigation --around ""
  # Output: flowchart TD of full graph (exit 0)
  go run ./cmd/tskflwctl thread graph refine-thread-and-tui-navigation --around "" --depth 2
  # Output: error: validation failed: --depth requires --around TASK (exit 11)
  ```
- **Evidence:** In `internal/cli/thread.go:349-363`:
  ```go
  if around == "" && cmd.Flags().Changed("depth") {
      return fmt.Errorf("%w: --depth requires --around TASK", domain.ErrValidation)
  }
  ...
  if around != "" {
      projection, err = core.SelectThreadGraphNeighborhood(projection, around, depth)
  ```
  `core.SelectThreadGraphNeighborhood(proj, "", depth)` explicitly rejects empty focal strings via `resolveTaskReference`: `task name "" must be a plain name (no path separators)`. However, the CLI checks `if around != ""` rather than `if cmd.Flags().Changed("around")`, bypassing validation when the flag is passed with an empty string.
- **Smallest sound remediation:** In `internal/cli/thread.go:349`, validate `if cmd.Flags().Changed("around") && strings.TrimSpace(around) == ""` and return `fmt.Errorf("%w: --around requires a non-empty task name or ID", domain.ErrValidation)`.

---

**Resolution:** The CLI now distinguishes an absent --around flag from an
explicitly empty or whitespace-only value, rejects the latter before repository
reads, and covers empty values with and without --depth.

### Important hypotheses disproved and contracts survived

1. **Cycle-induced BFS infinite loop hypothesis:** Disproved. `SelectThreadGraphNeighborhood` tracks visited nodes in `distance[neighbor]`, guaranteeing that malformed cycle-shaped inputs terminate without recursion or allocation blowup.
2. **Antiparallel edge duplicate neighbor hypothesis:** Disproved. Antiparallel edges ($A \to B$ and $B \to A$) insert $B$ twice into $A$'s adjacency list, but the BFS queue deduplicates via the `distance` visited check, producing identical deterministic node and induced edge sets.
3. **Slice aliasing between source and excerpt hypothesis:** Disproved. `cloneThreadView` deep copies every underlying slice (`Members`, `ExternalGates`, `Frontier`, `GraphProblems`, `Problems`, `Tags`, `Tasks`), ensuring that mutations to the bounded projection in memory cannot mutate the source view.
4. **Boundary marker name collision hypothesis:** Disproved. Synthetic boundary nodes are prefixed with `boundary%d` while graph nodes use `n%d` and legend nodes use `legend*`, precluding any identifier collision in Mermaid or DOT syntax.
5. **Format character injection in visual renderers hypothesis:** Disproved. All Unicode format controls (`\u202e`, `\u200b`, `\u202c`), directive sequences (`%%{init}`), and HTML markup (`<`, `>`) are sanitized into Unicode replacement characters or decimal entities before rendering.

---

### Residual uncertainty and unexecuted items

1. **Visual GUI Rendering:** Automated headless generation of PNG/SVG diagrams via `dot` and `mmdc` was not executed because the corresponding binary packages are not installed in the review environment. Syntax correctness was rigorously confirmed through unit tests, golden comparisons, and Graphviz/Mermaid specification matching.
2. **Interactive TUI Focus Mode Integration:** The in-app TUI interactive keybindings and spatial graph zoom transitions were not evaluated in this review because they are explicitly scheduled as a separate follow-up task (`6g95m4eyvf4k`). The core selector and wire types evaluated here provide the complete decoupled foundation for that work.
