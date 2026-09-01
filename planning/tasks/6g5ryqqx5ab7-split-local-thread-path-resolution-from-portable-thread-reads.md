---
schema: 1
id: 6g5ryqqx5ab7
status: in-progress
epic: 30-threads-and-task-dependency-graphs
description: Move ResolveThreadPath behind an optional local-path capability so portable Thread list/show adapters need not counterfeit filesystem behavior.
effort: 1 day
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, architecture, ports, filesystem]
created: "2026-09-01"
depends_on: [6g5fy1m967ka]
updated_at: "2026-09-01"
started_at: "2026-09-01"
---

# Split local Thread path resolution from portable Thread reads

## Objective

Make the independently injectable Thread read capability genuinely persistence-neutral. A consumer
that can list and show Threads from a database, API, or cache must not implement or fabricate
`ResolveThreadPath`; explicit local path lookup belongs behind a separate optional capability used
only by adapters that can provide one.

## Starting constraint

At task start, `ThreadStore` combined `ListThreads`, `GetThread`, and `ResolveThreadPath`. The first
two were semantic document reads; the third promised an absolute local repair/open path and was used
by `thread path`. That made every otherwise portable Thread adapter filesystem-shaped even before
its list diagnostics were made neutral.

## Scope

- Remove `ResolveThreadPath` from the portable Thread document-read interface and define a narrow
  consumer-owned path capability for the explicit path use case.
- Let `core.Service` receive Thread reads and Thread paths independently. Complete filesystem stores
  retain ergonomic defaults, while read-only/split construction can omit local paths without losing
  list, show, compose, projection, or graph behavior.
- Carry the optional path capability through `WorkspaceSource`/`WorkspaceService` without silently
  borrowing it from a different persistence source. A split workspace must not pair remote Thread
  contents with an unrelated aggregate store's path resolver.
- Make `Service.ThreadPath` and the CLI return an explicit typed unavailable-capability error when no
  path source exists; do not invent a URI or return an empty successful path.
- Preserve local filesystem resolution, malformed-document repair lookup, ambiguity behavior, and
  existing human/JSON command output when the capability is present.

## Acceptance criteria

- [x] A minimal pathless `ThreadStore` fake compiles and supports list, show, compose, projection,
  plan, and graph reads without implementing filesystem methods.
- [x] `thread path` remains byte-compatible for local FS workspaces and fails explicitly through the
  normal domain error classification when the optional path capability is absent.
- [x] Workspace composition tests prove Thread reads and paths are injected
  independently, implicit aggregate path fallback is detached when reads are
  replaced, and explicitly paired sources are a documented composition-root
  responsibility.
- [x] Malformed local Thread documents remain path-resolvable for repair even when semantic
  `GetThread` cannot parse them.
- [x] Typed-nil and missing-capability cases do not panic, and complete aggregate-store construction
  retains its current defaults.
- [x] Architecture and ADR guidance distinguish semantic Thread reads, optional local path lookup,
  layout watches, and guarded mutation ports.

## Stress tests

- Pathless remote fake, complete FS default, independently injected read/path fakes, missing and
  typed-nil path source, mismatched workspace sources, ambiguous slug, malformed frontmatter,
  filename/frontmatter ID drift, and a symlinked local planning root.

## Out of scope

- Designing general URLs for remote entities, splitting every other entity's path method, changing
  Thread list diagnostics, implementing HTTP/database adapters, or changing path-command semantics
  when a local resolver exists.

## Sequencing

Runs before adapter-neutral Thread list diagnostics so that change can redefine `ThreadStore`
without preserving a separate filesystem obligation. Both precede TUI Thread projection loading.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Foundation [portable Thread graph reads](6g5fy1m967ka-decouple-thread-graph-reads-from-the-aggregate-planning-store.md)
- Downstream [adapter-neutral Thread diagnostics](6g5rxq1ravd3-make-thread-read-diagnostics-adapter-neutral.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)

## Implementation progress (2026-09-01)

Split semantic Thread document reads from optional local path lookup with a new consumer-owned
`ThreadPathSource`. Service and Workspace composition now inject the capabilities independently; an
explicit `ThreadStore` override detaches any aggregate-store path default unless a path source is
also supplied, independent of option order. The local FS remains the complete adapter and preserves
parse-free repair lookup.

Coverage pins a minimal pathless `ThreadStore`, independent and typed-nil sources, complete aggregate
defaults, no-implicit-fallback workspace construction, malformed frontmatter, filename identity
drift, ambiguous references, symlinked planning roots, CLI unavailable-capability classification,
and unchanged local human/JSON path behavior. Architecture, README, and ADR guidance now distinguish
semantic reads, local paths, watcher layout, and guarded mutations.

Validation: the full race suite passes; golangci-lint reports 0 issues; normal planning lint and audit
finding lint pass; `git diff --check` is clean.

## Adversarial review closeout (2026-09-01)

Claude returned ready with tracked follow-ups. Corrected the architecture overclaim about diagnostic
neutrality and documented that a value implementing both read and path ports must be supplied
through both options. Rejected the proposed reverse workspace validation: an explicit path source
is an affirmative override that may legitimately decorate aggregate-discovered reads, while the
detached aggregate fallback was implicit. The ports expose no corpus identity with which to validate
an explicit pair honestly. Service, `WorkspaceSource`, ADR, architecture guidance, and a regression
test now pin that intentional asymmetry and the composition-root compatibility obligation.

All three findings are resolved: L2 and L3 fixed; L1 wontfix with rationale. The full race suite,
focused tests, golangci-lint, planning lint, audit lint, and `git diff --check` pass.
