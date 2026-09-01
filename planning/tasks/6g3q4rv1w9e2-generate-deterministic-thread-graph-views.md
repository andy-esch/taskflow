---
schema: 1
id: 6g3q4rv1w9e2
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Render explanatory Mermaid and DOT Thread graphs from shared runtime projections without persisting derived output.
effort: 2-3 days
tier: 3
priority: medium
autonomy_level: 4
tags: [threads, graph, cli, rendering]
created: "2026-08-25"
depends_on: [6g3q4rtv8d0a, 6g5f1d23jy1b, 6g5fy1m967ka, 6g5gbk5a5bt0]
updated_at: "2026-08-31"
started_at: "2026-08-31"
completed_at: "2026-08-31"
---
# Generate deterministic Thread graph views

## Objective

Render explanatory Thread graphs from the shared runtime projection without storing derived diagrams or introducing scheduling semantics.

## Scope

- Add one taskflow-owned, adapter-neutral graph projection in core for CLI, TUI, and future web
  consumers; it must not contain renderer, framework, filesystem, or third-party graph types.
- Add deterministic Mermaid and DOT output with explicit member versus external-gate roles.
- Put Mermaid and DOT formatters in a pure `internal/graphfmt` package with no CLI/TUI/HTTP or
  styling dependencies, so another primary adapter can export the same projection without invoking
  Cobra, importing Lip Gloss, or parsing CLI output.
- Polish topological waves as explanatory planning output.
- Keep ASCII/Unicode optional and exclude critical path, slack, forecasting, and transitive reduction.

## V1 contract

- `tskflwctl thread plan <thread>` presents deterministic, member-only topological waves and lists
  immediate nonmember prerequisites separately as external gates. Waves explain dependency order;
  they do not authorize dispatch, impose execution barriers, or promise duration.
- `tskflwctl thread graph <thread> [--format mermaid|dot]` emits Mermaid by default. ASCII/Unicode
  is deferred rather than exposed as an unimplemented format value.
- The neutral projection contains the Thread view, stable-ID-ordered nodes, prerequisite-to-dependent
  edges, member-only waves, and an explicit topology-completeness verdict. Nodes are bounded to
  members plus their immediate external gates, and edges include every repository dependency whose
  endpoints are both inside that boundary. Member waves preserve ordering paths through included
  external gates without placing those gates in the waves. Deeper causal context remains available
  through blocker queries.
- `--json` on either command emits the versioned neutral projection, not Mermaid or DOT embedded in
  JSON. An explicitly supplied `--format` is incompatible with `--json`, keeping renderer selection
  out of the machine contract.
- Unhealthy repository or Thread evidence remains visible in the projection and may retain useful
  partial waves, but `topology_complete` is false. Rendering is explanatory and does not turn an
  unhealthy projection into permission to start or complete work.

## Acceptance criteria

- [x] Mermaid and DOT encode the same nodes, edges, roles, and ordering as the shared projection.
- [x] CLI output and reusable formatters consume an adapter-neutral core projection that can also
  be used directly by TUI and future web adapters.
- [x] Output is generated at runtime and never persisted into Thread documents.
- [x] Titles and metadata are escaped safely and output remains deterministic across repository scan order.
- [x] The core projection carries raw labels; each `internal/graphfmt` formatter owns its
  format-specific escaping and is directly testable without a primary adapter.
- [x] Core and machine contracts remain renderer-, framework-, and graph-library-neutral.
- [x] `thread plan` ranks members only, marks direct external gates separately, and labels partial
  topology without presenting waves as dispatch authorization or barriers.
- [x] `thread graph` supports Mermaid and DOT with a documented Mermaid default; both graph commands
  emit the same neutral projection under `--json` and reject explicit renderer selection there.

## Stress tests

- Golden fixtures for external gates, shared tasks, disconnected members, hostile labels, empty/minimal graphs, and large deep/wide graphs.

## Sequencing

Requires stable Thread projections and adapter-neutral graph-load diagnostics. It is not a
prerequisite for bulk linking or eligibility enforcement.

## Implementation progress (2026-08-31)

- Added `core.ThreadGraphProjection` and `Service.ShowThreadGraph` over the independent
  `ThreadStore` and `TaskGraphSource` ports. The projection retains the complete `ThreadView`, raw
  stable-ID-ordered nodes, bounded dependency edges, member-only waves, and a health-qualified
  topology verdict; broken members remain visible but unranked.
- Added pure `internal/graphfmt` Mermaid and DOT adapters with synthetic node identifiers,
  format-specific escaping, semantic role metadata, and deterministic output validation. No graph
  library or presentation-framework type enters core or wire.
- Added `thread plan` and `thread graph`, plus one versioned renderer-neutral JSON projection shared
  by both commands. The global machine schema is now 1.58 and the generated CLI reference documents
  the Mermaid default and `--format`/`--json` boundary.
- Stress coverage includes direct external gates, disconnected and shared tasks, missing/broken
  members, hostile multiline labels, empty projections, scan-order invariance, and a 256-node
  deep/wide Thread. Dogfooding against `complete-production-threads` produced five member waves and
  one satisfied external gate. Starting this task still removes it from `thread frontier`; the
  already-filed `show-in-flight-work-before-dispatchable-thread-frontier-members` task owns that
  separate presentation issue.

## Adversarial review closeout (2026-08-31)

The independent reviews found one substantive topology omission and several contract-hardening
issues. The bounded graph now retains every dependency induced by its member/external-gate node set,
including edges into an external gate. Member-only waves contract paths through included gates, so
members separated by such a path cannot be presented in the same explanatory generation while the
gate still remains nonmember work.

The graph-node transport shape no longer repeats the gate-specific `outstanding` boolean on every
member. Mermaid and DOT replace Unicode format controls, including bidi overrides and zero-width
characters, rather than preserving their visual-spoofing behavior. Dedicated regressions pin the
gate-traversing topology and an edge order that differs from natural dependent scan order. Long
description wrapping and the gates-before-waves human layout remain intentional presentation
choices.

The rollout boundary is now a dogfooded v0.18.0 CLI preview after this slice and the in-flight
frontier presentation task merge. Task `6g5m69wpydzw` owns that release checkpoint and precedes the
usage-informed TUI work.

Validation: the full test suite and race-enabled full suite pass; golangci-lint reports zero issues;
Go module tidiness, generated wire-schema comments, generated CLI docs, `git diff --check`, and
planning lint are clean.
