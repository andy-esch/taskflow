# Threads compatibility and preview graduation

Threads remain a preview until the gates below pass. Preview means the feature is suitable for real
planning data, but its command ergonomics and presentation may still change. It does not waive data
integrity: guarded writes, stable IDs, explicit plan/wire schemas, and recovery receipts apply
during preview. Passing the gates makes graduation supportable; it does not force it. The project
shipped the compatibility-hardened v0.20.0 checkpoint with the preview label intact on 2026-09-07;
that installed-release evidence now begins the soak before graduation is explicitly reconsidered.

## Compatibility contract

| Surface | Classification after graduation | Promise |
| --- | --- | --- |
| Task-owned `depends_on` and Thread-owned `tasks` | Stable persisted semantics | Dependencies remain a repository-global DAG; membership never implies an edge; tasks may belong to multiple Threads; projections are computed at read time. |
| `threads/<id>-<slug>.md`, the concrete Thread shapes emitted since v0.18.0, stable ID, lifecycle, and timestamps | Compatibility-managed persisted format | Supported readers are pinned against the actual historical shapes. Additive fields are tolerated and preserved by tool-owned updates. The shared `schema` key is advisory and is not a read or write guard. Until a shared enforcement boundary has shipped, no incompatible Thread-document shape may ship; a later incompatible change requires a compatible reader to have been deployed first plus an explicit, dry-runnable, idempotent migration. |
| Thread authoring manifest | Compatibility-managed input | Schema zero (normally represented by an omitted `schema` key) and explicit schema 1 are accepted. Unknown fields and unsupported versions fail before a plan is written or repository mutation is attempted. |
| Materialized Thread apply plan, schema 1 | Compatibility-managed durable artifact | A materialized plan is a retry token, not disposable transport, and requires exactly schema 1. A replacement schema retains the schema-1 reader or ships an explicit converter before support is removed. Planning-repository binding, exact IDs, additive intent, and resumable-prefix semantics do not change silently. |
| Repository configuration used to replay a Thread plan | Required replay context | `planning_repo_id` is checked against the target repository's durable ID. A legacy repository without that identity fails before mutation with the `config migrate` remedy; historical compatibility tests cover both migrated success and pre-migration refusal. The user-scoped spaces registry is not part of the plan format. |
| Existing `thread` commands, mutation verbs, and flags | Stable command vocabulary | Additive commands and flags are allowed. A rename or removal keeps a forwarding alias and warning for at least one minor release unless continued support would be unsafe. Mutating semantic changes require release notes and, where persisted intent is affected, a migration path. |
| Human CLI text, Mermaid/DOT text, shell completion order | Compatibility-managed presentation | Keep output useful and deterministic, but scripts must not parse it. Material layout may evolve without a schema bump; automation uses `--json`. |
| `--json` envelopes, field meanings, vocabularies, and error codes | Stable versioned machine contract | `schema_version` is the compatibility authority, independent of the binary version. Effective from the Thread graduation baseline (`1.61`), additive fields or vocabulary values bump its minor version; removals, renames, or type/meaning changes bump its major version. Earlier 1.x history includes a vocabulary removal at 1.47 and an eligibility reinterpretation at 1.55, so consumers may not infer this rule retroactively. Consumers must tolerate unknown fields and values and check the version they support. Release review verifies that the bump class matches the change class. |
| Core Thread views, graph projection semantics, health, roles, edge direction, waves, bounded scope, and frontier | Stable semantic contract | Every adapter receives the same taskflow-owned values. Member/external-gate roles, prerequisite-to-dependent edges, graph/projection health, sound completion, and frontier eligibility cannot be reinterpreted by a presentation adapter. A neighborhood is an induced one- or two-hop excerpt of the supplied projection and carries its focal identity, shown/hidden counts, and exact crossing edges; it does not relabel partial evidence healthy. |
| Core Go interfaces and concrete types under `internal/` | Internal implementation detail | Source compatibility is not promised. Capability separation, pathless reads, causally compatible task/Thread snapshots, and renderer-neutral projections are architectural invariants even if port names or shapes change. |
| TUI layout, key map, wave presentation, and a future spatial view | Compatibility-managed presentation | Stable-ID navigation, coherent reload, and honest health evidence are correctness requirements. Exact layout and keys may evolve. A two-dimensional view remains an experiment until separately promoted. |
| Filesystem paths, Markdown surgery, repository locks, CAS tokens, and graph algorithms | Adapter/internal detail | The local adapter preserves unknown frontmatter and body content and supports the documented macOS/Linux release matrix. Other adapters preserve semantic identity, snapshot, mutation, and recovery guarantees without emulating local paths or Markdown bytes. |

The promise is forward upgrade compatibility: a newer supported binary must safely read older
supported data. It is not a promise that an older binary understands fields or states written by a
newer one. Today that failure can be silent because document `schema` is advisory, so mixed-version
teams must not introduce an incompatible document shape until a guarding reader has shipped in an
earlier release. Git remains the rollback mechanism for an explicitly migrated Markdown corpus.

## Graduation gates

All required gates must be evidenced on the recorded clean commits in the decision procedure: the
G1–G6 base and the fully revalidated G7a candidate derived from it. A version number or elapsed time
is not evidence.

| Gate | Observable evidence | State / owner |
| --- | --- | --- |
| G1 — graph integrity and recovery | Normal `lint`, Thread reads, and dispatch fail closed on broken evidence; diagnostics identify causes; guarded repair previews exact edits, permits strict improvement, reports residual damage, and retains partial-durability receipts. | Implemented by [`6g697mp8s4tx`](../planning/tasks/6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md) and [`6g4g8gatbnrs`](../planning/tasks/6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md). Candidate evidence must include `TestService_ListTasks_UnblockedUsesStrictGraphAndFailsClosed`, `TestTaskGraphRepairCycleRequiresExplicitEdgeAndCanLeaveOtherDefects`, `TestMutateTaskGraphRepairReportsAndConvergesDurablePrefix`, and `TestTaskDependRepairDiagnosesThenAppliesAutoAndExplicitIntent`. |
| G2 — lifecycle and mutation safety | Dependency, task-lifecycle, Thread-membership/lifecycle, and bulk-apply writers share authoritative guarded snapshots, reject invalid states, preserve committed outcomes, and pass race/partial-write/idempotent-retry coverage. | Implemented by [`6g3q4rt0wzkq`](../planning/tasks/6g3q4rt0wzkq-make-repository-graph-mutations-portable-and-serializable.md), [`6g5075cga2nt`](../planning/tasks/6g5075cga2nt-make-dependency-eligibility-graph-driven-for-queued-and-ready-tasks.md), [`6g4wm2yf6tyj`](../planning/tasks/6g4wm2yf6tyj-ship-guarded-thread-membership-and-lifecycle-mutations.md), and [`6g3q4rtv8d0a`](../planning/tasks/6g3q4rtv8d0a-bulk-link-existing-tasks-into-threads-with-resumable-apply.md). Candidate race evidence must include `TestMutateTaskGraphConcurrentOppositeEdgesCannotCommitCycle`, `TestThreadMembershipWaitsForDependencyMutationAndUsesFreshGraph`, `TestThreadCompleteWaitsForTaskLifecycleAndRefusesFreshUndrainedState`, `TestThreadApplyEveryDurablePrefixRetriesToCompletion`, and `TestThreadApplyRawEditAfterDurablePrefixIsReportedAndResumable`. |
| G3 — persisted upgrade compatibility | Fixtures from the v0.18.0 and v0.19.0 surfaces prove that the current binary can read, project, and safely mutate old Thread documents; replay retained plans against migrated configuration; and refuse a pre-migration repository before mutation with the documented `config migrate` remedy. The tests pin actual historical shapes rather than treating the advisory document `schema` key as an enforcement boundary. | Implemented by [`6g7ddeyp773z`](../planning/tasks/6g7ddeyp773z-pin-thread-document-and-plan-backward-compatibility.md) and exercised in the v0.20.0 checkpoint. |
| G4 — machine and command compatibility | Thread envelopes remain in the reflected JSON Schema; read-side CLI goldens stay stable; command docs are current; and command-level compatibility coverage checks mutation, update, compose, and apply envelopes—including failure receipts—without parsing human text. | Existing `internal/wire` registry/schema tests and `internal/cli/testdata/golden/thread_*` read fixtures, plus the mutation-side assertions required by the G3 follow-up. Release review also confirms that the wire-version bump class matches any changed fields, meanings, vocabularies, or errors. |
| G5 — adapter-neutral semantics | Pathless fake adapters exercise Thread list/show/compose/plan/graph, bounded graph selection, and stable navigation from the same core projections; no core or wire contract requires a local path, Cobra, Bubble Tea, GitHub, or renderer type. | Candidate evidence must run `TestServiceThreadReadsComposeIndependentGraphAndThreadPorts`, `TestServiceComposeThreadApplyRendersDefaultTemplate`, `TestServiceShowThreadGraphDetailUsesOnePairedReadInOrder`, `TestSelectThreadGraphNeighborhoodKeepsDirectedInducedEdgesAndBoundaryEvidence`, `TestTaskGraphReadAttributesUnreadableRecordWithoutFilesystemPath`, and `TestThreadTopologyCursorOpensSelectedTaskByStableIdentity`, backed by [`6g5fy1m967ka`](../planning/tasks/6g5fy1m967ka-decouple-thread-graph-reads-from-the-aggregate-planning-store.md), [`6g5ryqqx5ab7`](../planning/tasks/6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md), [`6g5rxq1ravd3`](../planning/tasks/6g5rxq1ravd3-make-thread-read-diagnostics-adapter-neutral.md), and [`6g9150nrt4p9`](../planning/tasks/6g9150nrt4p9-export-bounded-thread-neighborhoods-around-a-task.md). |
| G6 — real dogfood and preview soak | A clean binary manages a throwaway space containing shared membership, direct external gates, fan-out/fan-in, lifecycle changes, bulk apply retry, a deliberately broken-and-repaired graph, TUI navigation, and live reload. The production implementation Thread is healthy, and findings from multiple installed preview releases have explicit dispositions. | The compatibility-hardened [`v0.20.0`](../planning/tasks/6g7fhfpmy032-cut-v0.20.0-as-a-compatibility-hardened-threads-preview.md) checkpoint shipped with the preview notice intact. At graduation, repeat the v0.18.0/v0.19.0 playbook on a clean candidate and review the accumulated preview findings. |
| G7a — candidate documentation and release checks | Starting from the G1–G6 commit, make the preview-removal and release-note changes, then run generated docs/schema checks, `just release-snapshot`, full race tests, lint, and planning lint on that exact candidate. | Pre-tag half of [`6g7ddfhh2jc2`](../planning/tasks/6g7ddfhh2jc2-graduate-threads-from-preview.md). This candidate is the only commit eligible to tag. |
| G7b — publication verification | The pushed tag identifies the G7a candidate; the release workflow succeeds; published binaries and checksums identify that same commit. A publication failure is retried for the same immutable tag. If the candidate itself is later proven invalid, ship a patch release that restores the preview notice and names the failed gate. | Post-tag half of [`6g7ddfhh2jc2`](../planning/tasks/6g7ddfhh2jc2-graduate-threads-from-preview.md). The task remains open until this evidence is recorded. |

## Explicitly non-blocking work

- [`6g6jqqcdehne`](../planning/tasks/6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md)
  and its lint prerequisite finish adapter-neutral diagnostics in repository-wide convenience
  views. Thread reads and projections are already portable, so these remain important follow-ups,
  not graduation gates.
- [`6g6wdvfp2ksa`](../planning/tasks/6g6wdvfp2ksa-make-thread-frontier-help-choose-among-independent-candidates.md)
  improves selection among equally eligible work without changing eligibility semantics.
- [`6g6dw5js81f3`](../planning/tasks/6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
  is the high-priority flagship presentation candidate immediately after v0.20. Its design/prototype
  remains a separately removable extension over `ThreadGraphProjection`; this is non-blocking only
  for v0.20 and the correctness-based graduation gates, not a statement of low product priority.
- [`6g7f0tqgftg3`](../planning/tasks/6g7f0tqgftg3-enforce-reserved-document-schema-versions-across-entity-writers.md)
  tracks a future shared read/write boundary for the reserved document `schema` marker. Adopting
  that policy for every entity is intentionally not a Thread-only graduation requirement.
- A production remote/database/web adapter, critical-path calculations, forecasting, cross-space
  edges, and the broader Markdown durability reassessment are not required to graduate the local
  feature. Future adapters must meet the same semantic and wire contracts when they ship.

## Decision procedure

1. The v0.20.0 checkpoint shipped from clean `main` with the preview notice intact. Its release task
   records the candidate, compatibility and dogfood evidence, published artifacts, and follow-ups.
2. After that installed release has supplied enough real-use evidence, explicitly open the
   graduation task. Run G1–G6 against a clean commit built from `main` and record the commit,
   commands, versions, fixture results, preview findings, and any exceptions. Passing gates is
   necessary but does not prevent choosing another bounded preview checkpoint.
3. If G1–G6 and the explicit soak review pass, create the G7a candidate from that commit by removing
   the README preview notice and aligning the ADR, architecture, generated docs/schema, and release
   notes. Run the full G7a validation again on this exact candidate. If it fails, do not tag it;
   retain or restore the preview notice and name the failed evidence and owning task.
4. Tag only the passing G7a candidate, then perform G7b publication verification. Retry a failed
   release workflow for the same tag. If post-publication evidence proves the candidate itself was
   invalid, issue a patch release that restores the preview notice and names the failed gate; do
   not rewrite the published tag.
5. Complete the graduation task only after G7b is recorded. Never substitute a version/date
   milestone or an unrelated optional feature for missing evidence.
