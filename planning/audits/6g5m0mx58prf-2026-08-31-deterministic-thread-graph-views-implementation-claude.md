---
schema: 1
id: 6g5m0mx58prf
bucket: closed
area: deterministic-thread-graph-views-implementation-claude
date: "2026-08-31"
updated_at: "2026-08-31"
---

# Audit: Deterministic Thread graph views implementation — Claude — 2026-08-31

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer
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

## Reviewer report

### Executive verdict

**Not ready** — one defect makes the primary deliverable materially wrong, and it is cheap to fix.

Most of this slice is strong. The port boundary is clean, `internal/graphfmt` is genuinely free of
UI dependencies, escaping survived thirteen hostile payloads without a single break-out, determinism
holds under every perturbation I could legitimately produce, the CLI contract (formats, `--json`
exclusivity, exit classification, stdout hygiene) is exact, and performance at 1,500 nodes is 90 ms.
Seven of eight mutation probes on contract-bearing lines were caught by a failing test.

But `ProjectThreadGraph` derives edges only by walking the prerequisites of **member** nodes, so any
repository edge whose *dependent* is an external gate is silently dropped even though both endpoints
are nodes in the same projection. The result is a diagram that omits real sequencing, a plan that
places demonstrably ordered members in the same wave, and `topology_complete: true` asserted over
it. On a three-task repository the exported graph directly contradicts what this tool's own
`task blockers` prints for the same DAG. Since this projection is explicitly the contract that the
TUI and a future web interface will consume, shipping it with a known-incomplete edge set is the one
thing worth blocking on.

The fix is three lines in the edge loop. The wave question beneath it (M1) is a design decision that
deserves an explicit answer, not necessarily a code change.

### Reviewed state and validation

Branch `feat/deterministic-thread-graph-views`, based at `1b9fd9b`, unstaged:
`45 files changed, 524 insertions(+), 44 deletions(-)`, 59 `git status --short` lines including
untracked. Machine: Apple M5, darwin 25.6.0. The worktree was restored to exactly this state after
every probe (`shasum -c` over the three patched sources; probe test files deleted and confirmed
absent).

| Command | Result |
| --- | --- |
| `go build ./...` · `go vet ./...` | clean |
| `go test ./...` | all packages pass |
| `go test -race -count=1 ./internal/core/... ./internal/graphfmt/... ./internal/cli/... ./internal/wire/... ./internal/tui/...` | pass, no race reports |
| `just lint` (golangci-lint) | 0 issues |
| `./bin/tskflwctl lint` · `audit lint` | both clean |
| `go mod tidy -diff` | clean |
| `just docs` + `diff -rq` against a pre-run copy | no drift |
| `git diff --check` | clean |
| CLI matrix (11 invocations) | formats, `--json`, `--format`+`--json`, bad format, `--dry-run`, missing thread — all as specified |

Renderer parsers: **neither Graphviz `dot` nor a Mermaid parser is installed on this machine**, and
per the brief I did not add one. Renderer validity was therefore assessed structurally — by
grammar-level analysis of the escaping and by asserting that no payload could add a line, a label,
or an edge — not by parsing rendered output. That limitation is why L2 is reported as a
display-layer observation rather than a proven rendering defect.

Mutation probes, each applied then reverted and checksum-verified:

| Probe | Caught by |
| --- | --- |
| remove node sort | `TestProjectThreadGraphBoundsAndOrdersNeutralProjection` |
| **remove edge sort** | **nothing — suite stays green (L3)** |
| reverse edge direction | 2 core tests + `TestGolden_MachineContract` |
| rank broken/unknown members | `TestProjectThreadGraphLeavesBrokenMembersUnranked` |
| drop projection health from `TopologyComplete` | `TestProjectThreadGraphQualifiesUsefulTopologyWithHealth` |
| let external gates into waves | 1 core + 2 cli tests |
| bypass Mermaid escaping | `TestMermaidEscapesHostileLabelsAndPreservesProjectionOrder` |
| make plan JSON diverge from graph JSON | 3 cli/wire tests incl. schema validation |

## Findings

### High

#### H1. Edges whose dependent is an external gate are never derived, so the graph omits real sequencing between its own nodes and still reports `topology_complete: true`  · **Status:** fixed

**File:** `internal/core/thread_graph.go:81-88` | **Component:** core/thread-graph
**Effort:** S · **Urgency:** acute

The edge loop walks prerequisites of **members only**:

```go
for _, memberID := range memberIDs {
    for _, prerequisiteID := range graph.Prerequisites(memberID) {
        if included[prerequisiteID] { ...append edge... }
    }
}
```

`memberIDs` excludes external gates, so no edge is ever emitted whose `To` endpoint is a gate — even
when its `From` endpoint is already a projected node. Nothing documents this restriction. The
`ThreadGraphEdge` doc comment says only "Both endpoints always occur in
ThreadGraphProjection.Nodes", ADR-0006's 2026-08-31 amendment item 1 says the projection contains
"prerequisite-to-dependent edges" without qualification, and item 5 bounds the **node** set, not the
edge set.

**Reproduction.** Three tasks, real repository DAG `m2 → gate → m1`; Thread members `{m1, m2}`;
`gate` is a direct prerequisite of `m1` and therefore an external gate:

```
$ tskflwctl thread graph <t>
  n2 --> n0                      # gate -> m1     ... and nothing else
$ tskflwctl thread graph <t> --json
  nodes: [(m1,'member'), (m2,'member'), (gate,'external-gate')]
  edges: [(gate, m1)]            # m2 -> gate is MISSING
  waves: [(1, [m1, m2])]
  topology_complete: True

$ tskflwctl task blockers m1     # the same repository, same tool
  • m2  not-started  transitive
    m1 -> gate -> m2
```

The diagram draws `gate` as a root with no dependencies and `m2` as an isolated node, when in truth
`m2` blocks the gate which blocks `m1`. `task blockers` and `thread graph` describe the same DAG
incompatibly.

A second instance of the same class: with `m` depending on gates `g1` and `g2`, and `g1` depending
on `g2`, the projection emits `g1 → m` and `g2 → m` but never `g2 → g1`, hiding gate sequencing
between two nodes it is already drawing.

No test exercises an external gate that has an included prerequisite, so this is an untested gap
rather than a pinned decision.

**Recommendation:** Derive edges over the included node set rather than the member set — iterate
`graph.Prerequisites` for every node in `projection.Nodes` and keep the existing
`if included[prerequisiteID]` filter. That preserves the bounded node boundary exactly while making
the edge set complete over it, and it is what every consumer already assumes. Add a regression case
for a gate with an included prerequisite. Note that this alone does not change wave numbering — see
M1.

**Resolution:** The projection now derives every dependency induced by its
bounded member and external-gate node set, including edges whose dependent is a
gate. The gate-path regression asserts both bounded edges and a healthy complete
verdict.

### Medium

#### M1. Member-only waves place demonstrably ordered members in the same generation when the ordering runs through an external gate  · **Status:** fixed

**File:** `internal/core/thread_graph.go:95-108`; plan rendering `internal/cli/render/thread.go:206-214` | **Component:** core/thread-graph
**Effort:** M · **Urgency:** soon

`memberEdges` keeps an edge only when `rankableMembers[edge.From] && rankableMembers[edge.To]`, so
every path that traverses an external gate is dropped from the wave computation. In the H1
reproduction this yields:

```
External gates
• gate  <id>  outstanding

Wave 1
• m1  <id>  candidate/blocked
• m2  <id>  candidate/clear
```

Two members that are strictly ordered (`m2` before `gate` before `m1`) are presented as one
generation, and the wave contains a `candidate/blocked` task beside a `candidate/clear` one with
nothing connecting the blockage to the gate listed above. This is exactly the case the brief's angle
3 asks about: the plan silently implies the external gate is irrelevant to member ordering.

Ranking members only is a deliberate, ADR-stated choice ("member-only waves"), and I am not
disputing it — waves should not turn a nonmember into Thread-owned work. The defect is that the
presentation offers no signal that a wave boundary was elided. Fixing H1 improves the *diagram* but
does not change this, because the wave filter still excludes gate-traversing paths.

**Recommendation:** Keep member-only ranking, but make the elision visible. The minimum is to rank
members over paths that may pass **through** included gates — that is, contract the gate vertices
rather than deleting their edges — so `m2` lands in an earlier wave than `m1` while the gate itself
still appears only as a node and in the gates section. If that is judged out of scope, then at
minimum mark members whose wave placement is qualified by an interposed gate, and say plainly in
`thread plan`'s output that waves ignore ordering through external gates.

**Resolution:** Member-only waves now contract ordering paths through included
external gates. The gate remains outside every wave while members on opposite
sides occupy distinct explanatory generations; ADR-0006 records the rule.

### Low

#### L1. `Outstanding` is structurally meaningless on member nodes  · **Status:** fixed

**File:** `internal/core/thread_graph.go:60-63`, `internal/wire/thread.go:178` | **Component:** wire/thread-graph
**Effort:** XS · **Urgency:** eventually

Members are constructed with `threadGraphNode(member, false)`, so `outstanding` is hard-coded false
on every member node regardless of its actual state. A member that is `ready-to-start` — plainly
outstanding work — serialises as `"outstanding": false`. The JSON schema description
("true only when an external gate is not soundly completed") documents the intent, and
`state.soundly_completed` carries the real per-node signal, so this is a naming/shape wart rather
than a data defect. It matters only because this projection is the contract a TUI and a web client
will bind to, where a boolean named `outstanding` on every node invites the wrong reading.

**Recommendation:** Either move the field onto a gate-specific shape, or rename it to something
unambiguous (`gate_outstanding`) so its member-node value cannot be misread.

**Resolution:** Removed the graph-node Outstanding field from core and wire.
Gate outstandingness remains only on the gate-specific Thread view, while node
state carries lifecycle and sound-completion semantics.

#### L2. Escaping neutralises syntax but preserves bidirectional and zero-width characters through numeric entities  · **Status:** fixed

**File:** `internal/graphfmt/graphfmt.go:140-158` | **Component:** graphfmt
**Effort:** S · **Urgency:** eventually

`escapeMermaid` passes letters, digits and ` -_./:()` through and encodes every other rune as
`&#N;`. That is completely effective against syntax injection — thirteen hostile payloads, including
`%%{init:...}%%`, `<script>`, DOT attribute terminators, `-->`, and a 20,000-character label,
produced exactly two label lines and one edge in both formats every time. It is not effective
against *visual* spoofing, because Mermaid renders HTML labels and the browser decodes the entity
back to the original character:

```
payload "safe‮gnirts‬" -> label "safe&#8238;gnirts&#8236;"   (renders reversed)
payload "a​b"               -> label "a&#8203;b"                  (invisible separator)
payload "\x1b[31mred"            -> label "&#65533;&#91;31mred"        (correctly defanged)
```

ANSI/C0 controls are correctly replaced with U+FFFD via `unicode.IsControl`, but bidi overrides and
zero-width characters are category `Cf`, which `IsControl` does not cover. A task description can
therefore make a rendered diagram display a different task name than the one stored. I could not
confirm rendering behaviour directly — no Mermaid parser is installed here and the brief forbids
adding one — so this is reasoned from the entity-decoding semantics of HTML labels.

**Recommendation:** Treat `unicode.Cf` (minus benign joiners if any label needs them) the same way as
`unicode.IsControl` — replacement character rather than a round-tripping entity — in `escapeMermaid`,
and consider the same in `quoteDOT`.

**Resolution:** Mermaid and DOT now replace Unicode format controls, including
bidi overrides and zero-width characters, instead of round-tripping them through
renderer labels. A shared hostile-control regression covers both adapters.

#### L3. The edge sort is observable but no test pins it  · **Status:** fixed

**File:** `internal/core/thread_graph.go:89-94`; tests `internal/core/thread_graph_test.go` | **Component:** testing/thread-graph
**Effort:** XS · **Urgency:** soon

Deleting the `sort.Slice(projection.Edges, ...)` block leaves `go test ./internal/core/
./internal/graphfmt/ ./internal/cli/ ./internal/wire/` **fully green** — the only one of eight
mutation probes not caught. The sort is not redundant: natural derivation order groups edges by
dependent, while the contract orders them by `(From, To)`. Constructed proof, two members whose
prerequisite IDs sort opposite to their own:

```
WITH sort   : [(4gyr -> f992), (tx3b -> k0sp)]
WITHOUT sort: [(tx3b -> k0sp), (4gyr -> f992)]
```

Existing fixtures happen to make the two orders coincide, so "edges are stable-ID ordered" — an
explicit contract line in the ADR and the doc comment — is currently unenforced.

**Recommendation:** Add a case whose natural derivation order differs from `(From, To)` order, and
assert the exact edge sequence.

**Resolution:** Added an adversarial fixture whose dependent scan order differs
from prerequisite-then-dependent order and asserted the exact edge sequence.

## Acceptance-criteria traceability

| # | Criterion (abbreviated) | Verdict | Evidence |
| --- | --- | --- | --- |
| 1 | Mermaid and DOT encode the same nodes, edges, roles, ordering as the shared projection | **Qualified** | Both formats agree with the projection and with each other on every probe. But the projection's own edge set is incomplete (H1), so both faithfully encode a wrong graph. |
| 2 | CLI and reusable formatters consume an adapter-neutral core projection usable by TUI/web | Met | `internal/graphfmt` imports only `fmt`, `strings`, `unicode`, `core`. No Cobra/Bubble Tea/HTTP/filesystem/graph-library type in core or wire. `ShowThreadGraph` uses the narrow ports. |
| 3 | Output generated at runtime, never persisted | Met | No write path; `--dry-run` on both commands is a plain read (exit 0, no artifact). Thread documents unchanged across every probe. |
| 4 | Titles/metadata escaped safely; deterministic across scan order | **Qualified** | 13 hostile payloads produced zero break-outs in either format; output byte-identical across repeated calls and `depends_on` reordering (membership order cannot legitimately vary — the domain requires sorted task IDs). Bidi/zero-width survive as entities (L2). |
| 5 | Core carries raw labels; each formatter owns its escaping and is directly testable | Met | `Label`/`Description` unescaped in core and in JSON; `graphfmt` tested without any primary adapter; probe confirms bypassing escaping fails a test. |
| 6 | Core and machine contracts renderer-, framework-, graph-library-neutral | Met | Schema 1.58 exposes only taskflow types; no renderer text in JSON; `--format` rejected with `--json` (exit 11). |
| 7 | `thread plan` ranks members only, marks gates separately, labels partial topology, no dispatch authorization | **Qualified** | Member-only ranking, separate gates section, `topology partial` label, and `Unranked members` all behave as specified; `frontier` remains the authorization surface. But wave membership can be wrong (M1). |
| 8 | `thread graph` Mermaid default + DOT; both emit the same neutral projection under `--json`; renderer selection rejected there | Met | Default is Mermaid; `--format dot` works; graph and plan JSON are byte-identical below the envelope (verified); `--format`+`--json` → exit 11 in either flag order. |

## Settled concerns

Chased and disproved by code plus a hostile reproduction.

1. **Renderer injection.** The most likely place for a real vulnerability. Thirteen payloads —
   Mermaid `%%{init:...}%%` directives, `<script>`/`<img onerror>`, entity text, `-->` and
   `class n0 externalGate` injections, DOT `label"]; evil [label="pwned`, backslash/quote/newline
   mixes, and a 20,000-character label — each produced exactly two label lines and one edge in both
   Mermaid and DOT. Nothing escaped its own line or altered graph structure. Structural analysis
   only; no parser was available (see validation note).
2. **Determinism.** Five repeated `thread graph` runs and three `--json` runs are byte-identical.
   Reordering a task's `depends_on` (which lint accepts) leaves Mermaid, DOT and JSON byte-identical.
   Reordering a Thread's `tasks:` is not a legitimate perturbation — the domain rejects unsorted
   membership as `invalid-thread-document`, which the projection surfaces rather than silently
   reordering.
3. **`topology_complete` overclaiming on unhealthy evidence.** Internal member cycle → `waves=[]`,
   `topo=false`. Missing member → `proj=broken`, useful partial wave retained, `topo=false`. A
   status defect on an **unrelated** task elsewhere in the repository → `graph=broken`,
   `topo=false`. All three guards work; the probe that drops projection health from the verdict
   fails a test. (H1 is the one path to a false `true`, and it is not a health question.)
4. **Node boundary and shared tasks.** Nodes are exactly members plus immediate external gates.
   A deeper nonmember prerequisite is correctly absent. A task that is a member of one Thread and a
   gate of another projects with the right role in each (`TestProjectThreadGraphSharedTaskRoleIsLocalToEachThread`).
5. **Edge direction.** Prerequisite-to-dependent everywhere; reversing it fails two core tests and
   the machine-contract golden.
6. **Port boundary and mutation safety.** `ShowThreadGraph` reads the Thread first and the task
   snapshot second through `ThreadStore` and `TaskGraphSource`, fails explicitly when either is
   absent, never touches the aggregate `Store`, and cannot authorize a write — it has no mutation
   path at all.
7. **CLI contract and stream hygiene.** Bad format → exit 11 with the valid set named; `--format`
   with `--json` → exit 11 in both flag orders; case-sensitive format values rejected; missing
   Thread → exit 10; completion offers `mermaid`/`dot`; `thread plan` pages while `thread graph`
   does not; stderr is empty on success for both commands (my first measurement suggesting otherwise
   was a zsh MULTIOS artifact in my own harness, not tool behaviour).
8. **Machine contract evolution.** Schema 1.58; `nodes`/`edges`/`waves` are always non-null arrays;
   wave indices are one-based and contiguous; graph and plan projections are byte-identical; raw
   labels stay unescaped in JSON; `go mod tidy -diff`, generated docs, schema comments and goldens
   are all clean and current.
9. **Performance and resource behaviour.** A 1,500-task deep chain with every task a member and
   180-character descriptions — the worst case for wave count — renders in 90 ms (`plan` 70 ms,
   `--json` 70 ms) against a 60 ms `thread show` baseline, producing 1,500 waves and a 434 KB
   Mermaid document. Cycle detection is iterative (explicit stack, not recursion), so deep chains
   carry no stack hazard. No quadratic behaviour; nothing here argues for a graph library.
10. **Dogfooding coherence.** `thread plan complete-production-threads` yields five waves and one
    satisfied external gate; `thread frontier` independently reports the single eligible member.
    Wave 1 mixing completed and candidate members is correct — waves are dependency generations, not
    remaining work — and the command's own help says so.
11. **Package placement.** `internal/graphfmt` is a genuine sibling package, not `internal/cli/render`
    (which imports lipgloss). This resolves the placement concern raised against the previous slice's
    task wording.
