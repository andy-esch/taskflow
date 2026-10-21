# Planning data and projections

[Architecture index](../ARCHITECTURE.md) · [Persistence](persistence.md)

This guide owns the relationship between semantic records, source evidence,
snapshots and application projections. Detailed contracts live in
[core ports](../../internal/core/store.go) and their owning types.

## Identity is evidence, not a display label

Domain entities contain semantic data, not filenames, paths or guarded revisions.
[LoadedRecord and VersionedRecord](../../internal/core/entity_read.go) keep those
concerns separate:

- `RecordSource.ID` is adapter-established identity for actions. A frontmatter
  declaration or display slug is not a substitute.
- `Location` is opaque diagnostic context, never parsed to recover identity.
  `LocationIsPath` is a presentation hint, not permission to open a file.
- Local navigation requires an explicit entity-specific path capability.
  Local repair requires a guarded adapter's explicit `LocalPath`.
- Opaque source revisions cover exact readable/unreadable source evidence but
  never enter domain values or ordinary wire output. Missing evidence fails closed
  for guarded equality.

A URI-like location remains useful to a remote adapter without becoming a repair
target. Pathless records remain readable and diagnosable. Readable declaration drift,
within-kind duplicate IDs, and recoverable unreadable task/Thread cross-kind
collisions retain per-occurrence diagnostics; ownership evidence does not establish
valid membership.

Task body/criterion/field writes resolve a reference once through parse-free
`TaskStore.ResolveTaskSource`, then carry its source ID in a
[TaskMutationTarget](../../internal/core/store.go). Persistence requires exact, unique
ownership before callbacks, previews and no-ops; deletion cannot trigger slug fallback.
Selection and write conflicts share one retry budget and timestamp. Whole-body
replacement keeps the strict target but is not retried.

Parse-free selection lets `task set` repair malformed typed fields without borrowing
local-path authority. Persistence validates the updated document, including previews;
unrelated corruption remains an attributed validation failure.
See [CLI repair regressions](../../internal/cli/task_set_repair_test.go).

Criterion handoffs resolve both endpoint IDs before self-tracking checks. Equal IDs
receive a canonical resolution check so a duplicate-ID ambiguity is not misreported
as self-tracking. Endpoint-resolution errors remain unretried; body-write conflicts
retry only the selected source. This is not
an atomic cross-task transaction or a guarantee that the destination stays present.
Task lifecycle selection remains inside the guarded graph; after the first plan,
retries keep its task ID while revalidating policy and impacts against fresh snapshots.
See [portable target tests](../../internal/core/task_target_test.go) and
[real-store drift tests](../../internal/store/task_target_test.go).

Already-selected interfaces use explicit Service entry points, skipping reference
matching while sharing ordinary CLI validation/retry rules. An ID alone is not
cross-space authority: the [TUI](tui.md#shell-and-entity-extension-seams) captures the
originating Service and invalidates unsubmitted intents on workspace activation.
Staged creation retains a non-submittable recovery copy rather than rebinding or
discarding typed input; see [TUI authoring](tui.md#staged-task-creation).
See [portable application tests](../../internal/core/selected_task_action_test.go)
and [TUI action tests](../../internal/tui/selected_task_action_test.go), including
[open-form activation regressions](../../internal/tui/selected_task_session_test.go).

Audit, research and epic scriptable writes follow the same rule with entity-specific
targets and parse-free source lookups. Audit header repair uses each snapshot
occurrence's source ID, not its slug. An epic's ID is its complete filename stem:
changing it removes that identity rather than selecting a similarly named epic.
See [portable entity tests](../../internal/core/entity_target_test.go),
[real-store retry tests](../../internal/store/entity_target_test.go) and
[TUI non-task actions](../../internal/tui/selected_entity_action_test.go).

[Selected-epic task creation](../../internal/core/selected_task_creation.go) uses
an optional sibling capability, not an ID passed back through ordinary epic
reference matching. The selected full source ID becomes the task's epic reference;
the adapter requires a unique readable parent in the same guarded create operation.
The application shares ordinary defaults, validation and scaffolds, makes one
attempt, and preserves its exact creation receipt/error. Selected create-and-start
is not supported; it cannot borrow graph authorization from another capability.
No local path or aggregate Store is required by the new port. Source-set witnesses
still bind composition, while the primary adapter retains the originating Service.
See [portable tests](../../internal/core/selected_task_creation_test.go) and
[guarded-store tests](../../internal/store/selected_task_creation_test.go).

This does not add enduring foreign keys or same-ID delete/recreate detection. Epic
identity is still its whole source stem, so a renamed stem removes that selection;
ordinary content edits at the same identity need not invalidate it. Ordinary CLI
reference-based creation keeps its existing matching and lifecycle behavior.

Dependency add/remove bind both endpoints inside the guarded planner, then recompute
the requested change against each fresh graph. Thread creation, membership and
lifecycle likewise retain successful selection even when validation conflicts before
the adapter returns a plan. Only intent survives: membership, policy, impacts and
write sets are derived again. Pre-planner contention establishes no selection;
committed outcomes and durable prefixes are never automatically replayed.
See [portable retry tests](../../internal/core/graph_retry_target_test.go) and
[real-store graph tests](../../internal/store/graph_retry_target_test.go).
Explicit graph repair binds each selected source ID **and local path** at the first
successful guarded plan. Broken graphs may have duplicate IDs, so retries never
follow a rename by ID alone or resolve the original slug/path again. A moved,
missing, unreadable or identity-changed occurrence is refused; inspect and rerun.
Current content, defects, labels and effects are still derived afresh at that
address. This is logical source identity, not inode continuity: replacing bytes
at the identical ID/path is not distinguishable from another content edit.
Auto-only repair deliberately rediagnoses the current source set. See
[portable repair retries](../../internal/core/dependency_repair_retry_test.go) and
[real-store drift tests](../../internal/store/graphrepair_target_test.go).

Already-selected local navigation uses the optional `*PathByID` Service methods,
not the ordinary reference-based path methods. Each adapter must resolve a unique
exact source ID without typed decoding, preserving access to malformed documents
for manual repair. Ordinary CLI reference matching is unchanged. Source-set binding
and TUI session guards still apply; an ID alone does not authorize another workspace.
See [selected path ports](../../internal/core/selected_entity_path.go) and
[queued local-action tests](../../internal/tui/selected_local_path_test.go).

See [record tests](../../internal/core/entity_read_test.go),
[source-boundary tests](../../internal/core/thread_source_boundary_test.go) and
[wire mappings](../../internal/wire/envelopes_test.go).

## Capability binding and snapshots

[NewService](../../internal/core/service.go) checks every selected planning-data
capability against a stable, nonzero
[SourceSetID](../../internal/core/source_set.go). The witness proves composition
agreement—not a transaction, source revision, repository identity or security boundary.
A value satisfying multiple ports must still be supplied through each required option.

Complete adapters can expose capabilities through the aggregate store. Split adapters
may provide task graph and semantic Thread reads separately. Replacing Thread reads
detaches an implicitly discovered local path source unless it is also supplied
explicitly; an explicit path source is an affirmative override. No hidden local fallback
may pair remote records with unrelated local files.
[Source-set tests](../../internal/core/source_set_test.go) and
[path-capability tests](../../internal/core/service_entity_path_test.go) pin this.

Joined Thread reads obtain Thread data first and tasks second; paired adapters ensure
the later task snapshot is no older, or coordinate compatible snapshots themselves.
Read projections are point-in-time diagnostics, never mutation authorization.
[Workspace opening](configuration-and-spaces.md#workspace-opening) currently supplies
the complete entity service and local Layout required by the TUI; narrower read-only
consumers compose Service directly.

## Shared read models

Task lists, board, blockers, downstream queries and Thread projections use
`TaskGraphSource`, not independently rescanned aggregate-store tasks.
[TaskGraph](../../internal/core/dependency_graph.go) is immutable and owns graph health,
cycle attribution, derived role/gate, sound completion, topology, impact and distinct
causal-blocker/action-frontier projections. Task references resolve inside that
snapshot, not before a guarded planner takes its authoritative read.

Exactly resolved legacy edges participate in diagnostic gates before migration.
Present-but-empty legacy keys remain degraded. Eligibility uses explicit derived
state, not an empty blocker list: both `next-up` and `ready-to-start` are eligible
with clear gates in a healthy snapshot. `task list --unblocked` fails closed on
unhealthy evidence; readiness is a handoff signal, not another authorization gate.

Recovery queries retain the complete readable-source multiset and unreadable evidence.
`SourceDeclarations` records physical occurrences, including invalid/legacy/duplicate
evidence; `CanonicalDependencies` describes canonical deduplicated values;
`Prerequisites` retains unresolved evidence so behavior fails closed.
Duplicate-ID shadows cannot inject edges into the representative graph but still
participate in snapshot equality and repair simulation. Source-incomplete projections
refuse repair queries. Removing a last legacy value does not implicitly remove its
now-empty key. See [repair policy/tests](../../internal/core/dependency_repair_test.go).

Lint uses its dedicated `LintSource`, returning neutral records and `LoadProblem`,
not filesystem `FileProblem` types. Audit metadata, parsed findings and near-miss
evidence come from one `AuditSnapshotSource`; core never rereads a location.
An explicitly injected audit source takes precedence over the broad lint default
independently of option order. Dashboard metadata and task/audit reads likewise use
portable snapshots and preserve diagnostics without another entity scan.
[Lint/audit-source regressions](../../internal/core/lint_source_test.go) pin source
precedence; [summary](../../internal/core/usecases_test.go) and
[board](../../internal/core/board_test.go) tests cover scan reuse and diagnostic joins.

Maintenance and completion are narrow sibling capabilities, not mandatory methods
on every Store. [RepairPlanning/LintWithLinks](../../internal/core/planning_maintenance.go)
keep frontmatter-before-body ordering, preview behavior and completed prefixes on
later failure; repair is never retried wholesale.
An interrupted ID spelling move keeps local action evidence in the core repair
receipt, separate from completed/proposed `FixResult` rows and semantic entities.
The optional wire relocation is recovery data, never path authority; see
[frontmatter repair evidence](../../internal/core/frontmatter_repair.go) and
[receipt regressions](../../internal/core/planning_maintenance_test.go).
[CompleteEntities](../../internal/core/completion.go) retains damaged records and
counts aliases before state exclusion. A label is executable only when the adapter
certifies it, respecting case-folded ambiguity and ID precedence; already-selected
identity does not hide same-label siblings. See
[maintenance tests](../../internal/core/planning_maintenance_test.go) and
[completion tests](../../internal/core/completion_test.go).

## Threads and graph presentation

[ThreadView](../../internal/core/thread_projection.go) derives membership projections from
that same graph. It separates repository health from local projection health, nominal
completion from sound drainage, members from direct external prerequisites, and
dispatchable pending frontier from in-flight work. Completed inconsistency has stable
reason codes; pending frontier requires healthy combined graph/projection evidence.
List results hoist shared structural/legacy evidence even on filtered
or empty pages.

Thread documents persist initiative metadata and task-ID membership only. Task files
own edges and lifecycle. A task transition compares every affected Thread, including
indirect/external effects, without transferring task state into Thread documents.

[ThreadGraphProjection](../../internal/core/thread_graph.go) supplies ordered nodes,
induced prerequisite-to-dependent edges, labels/titles, explanatory member ranks,
the Thread view, and topology completeness. Its bounded selector treats one/two-hop
proximity as undirected while preserving edge direction, original rank numbering,
focal identity, shown/hidden counts and crossing-edge evidence. It never rescans data.
Ordering through external prerequisites does not make them Thread-owned work.
Dependency ranks are explanatory, not execution barriers or lifecycle statuses.

CLI, TUI and future adapters consume the same projection. Mermaid/DOT formatting
belongs to [graphfmt](../../internal/graphfmt/graphfmt.go); TUI geometry, focus caches,
scope cards and keyboard routing belong to [the TUI](tui.md). Neither format is
reparsed to obtain graph semantics. Graph-library/framework/filesystem types do not
enter core or wire contracts. Persisted and machine compatibility is owned by
[Threads compatibility](../THREADS_COMPATIBILITY.md), not the current Go interface
or terminal layout.

## Change checklist

- Put use cases, graph/impact/recovery policy and consumer-owned ports in core;
  controllers use injected services and adapters own decoding/guarded persistence.
  Follow [named exceptions](boundaries.md#composition-and-local-exceptions).
- Preserve source identity separately from declarations and opaque context; request
  explicit navigation/repair capabilities.
- Check split source-set witnesses, compatible snapshots, lazy opening, path-source
  detach/override behavior and identity revalidation.
- Supply explicit [mutation policy](../../internal/core/mutation_policy.go);
  authorization includes previews/no-ops before callbacks/effects.
- Retain committed/partial receipts on failure; proposed impacts are not durability.
  Core owns recovery; never blindly retry committed work or expose revision/repair
  handles in wire output.

For changed behavior, combine portable and real-adapter evidence:

| Boundary | Start with |
| --- | --- |
| Pathless reads, drift, duplicates, unreadable occurrences | [Core reads](../../internal/core/entity_read_test.go), [source boundary](../../internal/core/thread_source_boundary_test.go) |
| Missing/mismatched/unstable capabilities | [Source sets](../../internal/core/source_set_test.go), [local paths](../../internal/core/service_entity_path_test.go) |
| Direct/pointer opening and late authorization | [Opening parity](../../internal/appwiring/thread_apply_test.go), [explicit spaces](../../internal/cli/space_selection_test.go) |
| No effects, stale refusal and durable failures | [Authorization](../../internal/store/mutation_authorization_test.go), [bulk writes](../../internal/store/threadapply_test.go), [recovery](../../internal/core/dependency_recovery_test.go) |
| Stable TUI targets across refresh | [Identity](../../internal/tui/identity_test.go), [local actions](../../internal/tui/local_path_test.go) |

Run the normal code gates. Import rules alone do not prove these behavioral contracts.
