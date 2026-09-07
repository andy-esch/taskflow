---
schema: 1
status: active
description: Implement accepted Threads as initiative views over a planning-space task DAG, with global dependencies, lifecycle gating, bulk linking, and generated projections.
priority: medium
tags: [planning-model, threads, graph, workflow]
created: "2026-08-24"
---
# Threads and task dependency graphs

**Goal.** Implement accepted Threads as initiative views over a planning-space task DAG without overcommitting to speculative graph features.

## Why this is its own epic

Threads are not only a new first-class document. They change task dependency ownership, lifecycle eligibility, repository-wide graph integrity, multi-file composition, CLI and wire projections, and eventually the TUI. That cross-cutting domain deserves a coherent home rather than being split between the generic entity, storage, and CLI epics.

The first work in this epic was deliberately a decision spike. It recommended accepting ADR-0006 and scoping the production implementation; no named risk required another spike. ADR-0006 was accepted on 2026-08-25, and the production tasks below now own delivery.

## Decision gate

ADR-0006 is **accepted**. Work follows the delivery sequence below; this epic does not treat all
Thread work as one implementation task.

## Production task graph

These bootstrap edges were prose until task `6g3q4rt7mgjn` landed the guarded dependency-write
surface. They are now persisted through the production commands, including the 2026-08-29 readiness
split between Thread documents and guarded Thread mutations and the first-Thread dogfood correction.
Prefer task IDs over slice numbers because the ADR slices may be divided at reviewed implementation
boundaries:

```text
6g3q4rst78qy strict reads -----> 6g3q4rt7mgjn dependency operations <----- 6g3q4rt0wzkq portable guard
                                           |
                                           v
                                  6g3q4rte8kc1 eligibility
                                           |
                                           v
                                  6g3q4rtmv4ak Thread documents
                                           |
                                           v
                                  6g5075cga2nt graph-driven eligibility correction
                                           |
                                           v
                                  6g4wm2yf6tyj Thread mutations
                                           |
                                           v
                                  6g3q4rtv8d0a bulk link
                                           |
                                           v
                                  6g3q4rv1w9e2 generated views
                                           |
                                           v
                                  6g5m69wpydzw v0.18 preview

6g5gbk5a5bt0 neutral task diagnostics ---------------+
                                                     v
6g5fy1m967ka portable reads -> 6g5ryqqx5ab7 path -> 6g5rxq1ravd3 Thread diagnostics
                                                             |
6g5rxq1g5mp1 watcher recovery -------------------------------+
                                                             v
                                                    6g5rwjqeh6a6 projection reload
                                                             |
6g5rxq17px59 stable TUI identity ----------------------------+
                                                             v
                                                    6g5rwjqr7rt8 list/detail
                                                             |
                                                             v
                                                    6g5rwjr0dz4p topology UX

6g697mp8s4tx graph-health reporting ----+
                                        +-> 6g6scc9jgxae v0.19 preview
6g63db3sdfrh Atlas recovery ------------+             |
                                                      +-> 6g4g8gatbnrs repair
                                                      +-> 6g5vm4efjcdv neutral lint
                                                      |          |
                                                      |          v
                                                      |   6g6jqqcdehne portable summaries
                                                      +-> 6g6wdvfp2ksa frontier decision context
                                                      +-> 6g6wdvfjdaaa graduation contract
                                                                  |
                                                                  v
                                                         6g7ddeyp773z compatibility fixtures
                                                                  |
                                                                  v
                                                         6g7fhfpmy032 v0.20 preview
                                                                  +-> 6g6dw5js81f3 flagship spatial prototype
                                                                  +-> 6g7ddfhh2jc2 graduation decision
```

The projection loader also depends on the generated views and v0.18 preview above. The stable-ID
and watcher fixes are independent general TUI roots; their placement here is a delivery grouping,
not an invented dependency on the release.

- [6g3q4rst78qy — strict dependency reads](../tasks/6g3q4rst78qy-establish-canonical-task-dependencies-and-strict-graph-reads.md)
- [6g3q4rt0wzkq — portable mutation guard](../tasks/6g3q4rt0wzkq-make-repository-graph-mutations-portable-and-serializable.md)
- [6g3q4rt7mgjn — dependency operations and queries](../tasks/6g3q4rt7mgjn-ship-guarded-dependency-mutations-and-graph-queries.md)
- [6g3q4rte8kc1 — eligibility enforcement](../tasks/6g3q4rte8kc1-enforce-dependency-eligibility-across-every-task-start-path.md)
- [6g3q4rtmv4ak — Thread documents, creation, and read projections](../tasks/6g3q4rtmv4ak-add-thread-documents-guarded-creation-and-read-projections.md)
- [6g5075cga2nt — graph-driven eligibility for queued and ready tasks](../tasks/6g5075cga2nt-make-dependency-eligibility-graph-driven-for-queued-and-ready-tasks.md)
- [6g4wm2yf6tyj — guarded Thread membership and lifecycle](../tasks/6g4wm2yf6tyj-ship-guarded-thread-membership-and-lifecycle-mutations.md)
- [6g3q4rtv8d0a — resumable bulk linking](../tasks/6g3q4rtv8d0a-bulk-link-existing-tasks-into-threads-with-resumable-apply.md)
- [6g3q4rv1w9e2 — generated graph views](../tasks/6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md)
- [6g5m69wpydzw — v0.18.0 CLI preview](../tasks/6g5m69wpydzw-cut-v0.18.0-as-a-cli-threads-preview.md)
- [6g5rxq17px59 — stable TUI entity identity](../tasks/6g5rxq17px59-make-tui-entity-navigation-use-stable-identities.md)
- [6g5rxq1g5mp1 — watcher directory recovery](../tasks/6g5rxq1g5mp1-keep-tui-live-reload-healthy-when-entity-directories-appear.md)
- [6g5ryqqx5ab7 — split local Thread paths from portable reads](../tasks/6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md)
- [6g5rxq1ravd3 — neutral Thread read diagnostics](../tasks/6g5rxq1ravd3-make-thread-read-diagnostics-adapter-neutral.md)
- [6g5rwjqeh6a6 — contention-safe TUI projection loading](../tasks/6g5rwjqeh6a6-wire-thread-projections-into-the-tui-with-contention-safe-reloads.md)
- [6g5rwjqr7rt8 — Thread list/detail TUI](../tasks/6g5rwjqr7rt8-add-thread-list-and-detail-views-to-the-tui.md)
- [6g5rwjr0dz4p — dogfooded topology presentation](../tasks/6g5rwjr0dz4p-add-dogfooded-thread-graph-presentation-to-the-tui.md)
- [6g697mp8s4tx — graph degradation in status and lint](../tasks/6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md)
- [6g63db3sdfrh — coherent Atlas recovery](../tasks/6g63db3sdfrh-preserve-coherent-atlas-summaries-across-transient-per-space-refresh-failures.md)
- [6g6scc9jgxae — v0.19.0 TUI preview](../tasks/6g6scc9jgxae-cut-v0.19.0-as-a-tui-threads-preview.md)
- [6g4g8gatbnrs — guarded graph repair](../tasks/6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md)
- [6g5vm4efjcdv — neutral repository-lint diagnostics](../tasks/6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
- [6g6jqqcdehne — portable board/status diagnostics](../tasks/6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md)
- [6g6dw5js81f3 — spatial Thread graph prototype](../tasks/6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
- [6g6wdvfp2ksa — frontier decision context](../tasks/6g6wdvfp2ksa-make-thread-frontier-help-choose-among-independent-candidates.md)
- [6g6wdvfjdaaa — preview graduation contract](../tasks/6g6wdvfjdaaa-define-the-thread-preview-graduation-and-compatibility-contract.md)
- [6g7ddeyp773z — historical Thread compatibility fixtures](../tasks/6g7ddeyp773z-pin-thread-document-and-plan-backward-compatibility.md)
- [6g7fhfpmy032 — v0.20.0 compatibility-hardened preview](../tasks/6g7fhfpmy032-cut-v0.20.0-as-a-compatibility-hardened-threads-preview.md)
- [6g7ddfhh2jc2 — explicit preview graduation decision](../tasks/6g7ddfhh2jc2-graduate-threads-from-preview.md)
- Deprecated combined scope: [6g3q4rv89vzw](../tasks/6g3q4rv89vzw-add-usage-informed-thread-views-to-the-tui.md)

## Delivery sequence and gates

```text
strict reads -> guarded edge writes -> eligibility -> Thread documents/read projections
              -> graph-driven eligibility correction -> Thread mutations -> bulk linking
              -> generated views -> CLI preview
              -> TUI identity/watcher/path/diagnostic foundations -> projection reload
              -> list/detail dogfood -> smallest useful topology view
              -> graph-health and Atlas hardening -> TUI preview
              -> guarded repair / portable diagnostics / frontier context
              -> compatibility fixtures -> v0.20 preview soak
              -> flagship spatial prototype / explicit graduation decision
```

Eligibility enforcement and Threads share the same graph foundation, but implementation is
deliberately serialized after guarded writes stabilize. Eligibility establishes the first
non-dependency guarded mutation seam; Thread creation establishes the document and materializer;
the first real Thread corrects the queued-versus-ready eligibility contract before Thread
membership/lifecycle settles the second mutation family; bulk linking composes the task and Thread
materializers under one outer guard.

| Order | Slice | Exit gate | Highest-value stress tests |
|---|---|---|---|
| 1 | `6g3q4rst78qy`: strict dependency reads, derived state, and legacy diagnosis | One deterministic strict snapshot/analysis contract with problems available to diagnostic readers; no graph write yet | malformed/unreadable tasks, ID drift, unknown status, duplicate/self/missing edges, cycles, legacy slug resolution, reconvergent diamonds |
| 2 | `6g3q4rt0wzkq` + `6g3q4rt7mgjn`: portable guard, dependency writes/queries, and guarded legacy migration | Final scan, pure planning/validation, and write share one store-owned critical section on every supported platform | nested acquisition, concurrent opposite edges, direct write versus bulk apply, stale CAS, idempotent repeats, guarded slug-to-ID migration |
| 3 | Eligibility enforcement | Every route into `in-progress` uses one policy and produces the same blocker/force result | all task statuses, direct/transitive blockers, withdrawn/missing prerequisites, reopen after downstream completion, forced inconsistent work |
| 4 | `6g3q4rtmv4ak`: Thread documents, guarded creation, and read projections | One first-class Thread document/materializer and one shared projection; creation is unstarted and authoritative | shared tasks, external gates and rollup denominators, empty Threads, cross-kind IDs, creation versus task mutation |
| 5 | `6g5075cga2nt`: graph-driven pending-work eligibility correction | Clear-gated `next-up` and `ready-to-start` tasks share frontier/start eligibility while their lifecycle roles remain distinct | both pending roles across clear/blocked/broken gates, forced starts, Thread/task selector parity |
| 6 | `6g4wm2yf6tyj`: guarded Thread membership and lifecycle | Membership/lifecycle use one guarded snapshot, retain committed outcomes, and augment task receipts with affected Threads | empty/all-withdrawn start/complete, cancelled/completed immutability, post-commit cleanup, real cooperating-writer races |
| 7 | Existing-task bulk linking | One literal-YAML manifest can create a Thread, add memberships and global edges, and converge after interruption | failure after every write prefix, retry/idempotency, wrong planning-space identity, edited/stale plan, concurrent edge mutation |
| 8 | Generated Mermaid/DOT and explanatory UX | Stable ordering and explicit member/external roles; nothing generated is persisted | snapshot/golden output, escaping hostile titles, large/deep/wide readable graphs |
| 9a | TUI foundations | Stable-ID UI state, recoverable directory watches, optional local paths, neutral Thread diagnostics, and bounded conflict reloads are independently proven | duplicate slugs, missing/replaced directories, pathless adapters, mutation-event bursts, stale async results |
| 9b | Thread list/detail | A read-only first-class Thread destination consumes core projections and remains useful on narrow terminals | CLI/core parity, completed drift, missing members, shared membership, filters and reloads |
| 9c | Dogfooded topology UX | Recorded list/detail usage justifies the smallest extra graph presentation; no second traversal engine appears | deep/wide/fan-in graphs, external gates, incomplete topology, resize and reload |
| 9d | Release hardening | Graph failures are visible outside mutations and transient Atlas contention preserves the last coherent per-space view | degraded/broken graphs, advisory legacy debt, mixed healthy/contended spaces, stale generations |
| 10 | v0.19.0 TUI preview | Clean-main CLI/TUI dogfood and release validation pass with preview boundaries documented | shared members, external gates, fan-in/out waves, task navigation, watcher refresh, release artifacts |
| 11a | Guarded graph repair | Every accepted durable prefix monotonically reduces deterministic graph damage without creating an escape hatch | cycles, dangling/self/duplicate edges, concurrent edits, interruption after every prefix |
| 11b | Portable multi-entity diagnostics | One adapter-neutral vocabulary survives lint, board, status, Atlas, CLI, TUI, and wire projections | pathless and misleading locations, mixed entity kinds, deterministic ordering, scan counts |
| 11c | Frontier decision context | Independent eligible candidates expose shared planning metadata without changing graph truth or inventing a scheduler | mixed priorities/tiers, unknown effort, active work, external gates, unhealthy graphs |
| 11d | Preview graduation contract | Public and persisted surfaces have explicit compatibility classifications and observable graduation gates | old preview data, CLI/JSON consumers, pathless adapters, TUI/web evolution, unmet gates |
| 11e | Historical compatibility fixtures | Current readers, surgical writers, and resumable apply preserve the concrete v0.18.0/v0.19.0 artifacts and meanings | additive frontmatter, old wire fields, schema 0/1 manifests, strict plans, interrupted prefix, legacy repo identity |
| 12 | v0.20.0 hardened preview | Compatibility and recovery improvements ship through the release pipeline while the preview notice remains explicit | clean candidate, installed CLI/TUI use, retained plans, repair, archives/checksums |
| 13a | Flagship spatial graph design/prototype | A bounded high-priority pass proves deterministic two-dimensional comprehension and navigation, then cuts production-grade slices for the next feature release | fan-in/out, crossings, skipped waves, narrow terminals, large graphs, reload stability |
| 13b | Explicit graduation decision | Fresh gates plus accumulated preview evidence support either a preview-removal candidate or named further work | all G1–G7 evidence, unresolved preview findings, exact candidate/tag provenance |

### Design attention

The mutation guard and strict-versus-resilient repository read split are the highest-risk design
work. They protect graph truth; no library can supply them. Lifecycle consistency is next: recursive
sound completion, forced starts, reopen behavior, and every status transition need one authoritative
policy. Bulk apply is a convergence protocol rather than a transaction, so interruption testing must
inject failure after every operation and prove that the same plan repairs the prefix.

The graph library is lower risk. Keep a taskflow-owned interface and run a bounded contract-test
bake-off between the spike's small implementation and `dominikbraun/graph` during slice 1. Do not
create another open-ended research spike, and do not let library features pull critical path, slack,
or other deferred graph analysis into V1.

The v0.19.0 release intentionally preceded guarded repair; repair and its CAS/source-declaration
foundations have now shipped. v0.20.0 carried that recovery work and the historical compatibility
suite while deliberately retaining the preview label. Graduation is a later explicit decision after
installed-release evidence, not an automatic consequence of green tests. Portable diagnostic tasks
remain serialized around one shared value and explicit wire-version boundary. With v0.20 shipped,
the spatial view is the tier-1 flagship candidate for the next feature
release. Its removable presentation-extension boundary over the existing projection prevents a
prominent UX experiment from becoming a hidden renderer, graph-library, or core dependency.

## Dogfood checkpoints

This epic is the first production consumer of its own capabilities:

1. Task `6g3q4rst78qy` proves a clean strict snapshot over the real repository, records scan timing,
   and reports the six legacy-field resolutions without pretending an edge-free graph exercises
   blocker or topology queries.
2. Task `6g3q4rt7mgjn` uses production dependency commands to persist the bootstrap edges and then
   exercises explanatory queries against those real relationships.
3. Slice 4 creates a real Thread for the remaining initiative and observes its frontier and external
   gates during normal implementation work.
4. Task `6g5075cga2nt` records and corrects the first semantic dogfood finding before later Thread
   behavior builds on it.
5. The guarded mutation slice manages that Thread through production membership and lifecycle verbs.
6. The bulk-linking slice uses the feature on the next naturally suitable initiative rather than a
   synthetic demo.
7. Every dogfood finding is recorded in the active task; contract changes also amend ADR-0006.
8. The first TUI stage fixes general identity and watcher seams plus Thread path/diagnostic
   portability and contention handling before adding visuals. Dogfood list/detail on a real active
   Thread before selecting a topology layout; a terminal diagram is not presumed necessary.
9. After the linear wave view lands, status/lint graph reporting and Atlas partial-refresh recovery
   close the v0.19.0 release boundary. A clean-main dogfood pass must use both CLI and TUI surfaces.
10. The post-v0.19 frontier fans out instead of manufacturing dependencies: graph repair is the
    highest-priority integrity foundation, portable diagnostics are one serialized contract
    migration, and frontier metadata makes independent choices visible. The graduation contract
    defines which outcomes actually gate the end of preview.
11. v0.20.0 shipped with the preview notice intact as the third installed dogfood checkpoint. Its
    completed release evidence unlocks the tier-1 spatial design/prototype as the flagship candidate for the next feature
    release, alongside the independent decision to graduate Threads or name another preview
    checkpoint. The spatial work does not gate v0.20 or silently join core.

The experimental spike binary is limited to disposable planning spaces and does not satisfy these
checkpoints. Dogfooding begins when the corresponding production slice passes its exit gate.

## Out of scope

- Treating the spike as production implementation.
- Critical-path, slack, forecasting, transitive reduction, or scheduler features.
- Autonomous multi-agent or worktree orchestration.
- Treating the spatial prototype as a production renderer before its layout and navigation evidence
  is evaluated.

## Sequencing amendment — guarded multi-kind writes (2026-08-27)

The portable-guard audits proved the dependency boundary but also made the next extension point
explicit: `TaskGraphMutationStore` deliberately materializes dependency writes only, while lifecycle,
Thread, and bulk operations each need an authoritative read/validate/write decision under the same
canonical-root exclusion contract. This amendment supersedes the earlier suggestion that eligibility
and Thread persistence may be implemented independently after dependency operations.

```text
dependency operations
        |
        v
eligibility lifecycle boundary    (first non-dependency guarded write)
        |
        v
Thread documents + creation       (first additional entity kind/materializer)
        |
        v
graph-driven eligibility correction (first production Thread dogfood finding)
        |
        v
Thread membership/lifecycle       (second guarded mutation family)
        |
        v
compound bulk apply -> generated views -> CLI preview
                    -> TUI foundations -> list/detail -> dogfooded topology UX
```

This is implementation coordination, not a new domain dependency: the pure eligibility and Thread
projections remain independently testable. Implementation is serialized so each slice reuses one
reviewed guard-extension pattern rather than inventing incompatible callbacks or nesting guarded
operations, which root-wide callback exclusion correctly rejects.

Keep the public capabilities use-case-specific and share private store mechanics:

- dependency commands use the existing guarded task-dependency capability;
- lifecycle enforcement adds a narrow guarded status-transition capability;
- Thread creation adds the document and lock-free internal materializer through a narrow guarded
  capability;
- Thread lifecycle/membership follows with its own narrow guarded mutation capability;
- bulk apply owns one deliberate compound capability that takes the guard once and composes the
  internal task and Thread materializers. It never orchestrates by nesting the narrower ports.

Generated views remain unchanged and read-only. The TUI remains last, but its implementation is now
split: general stable identity and recoverable watches plus optional local Thread paths and neutral
Thread diagnostics precede one contention-safe projection loader; list/detail ships before topology
presentation. The loader must retry/debounce the documented transient `ErrConflict` when a watcher
refresh overlaps the planner-exclusive phase without hiding durable errors.

## Related

- [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md)
- [Vertical MVP decision spike](../tasks/6g3a1wtx4zrr-spike-a-vertical-threads-and-global-task-dag-mvp.md)
- [First-class entities epic](28-first-class-entities-new-planning-nouns.md).
