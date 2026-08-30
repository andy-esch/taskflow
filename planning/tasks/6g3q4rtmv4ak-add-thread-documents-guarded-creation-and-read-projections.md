---
schema: 1
id: 6g3q4rtmv4ak
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Introduce Thread documents, guarded unstarted creation, reusable materialization, and deterministic read projections.
effort: 5-8 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, domain, storage, cli]
created: "2026-08-25"
updated_at: "2026-08-29"
depends_on: [6g3q4rte8kc1]
started_at: "2026-08-29"
completed_at: "2026-08-29"
---
# Add Thread documents, guarded creation, and read projections

## Objective

Introduce durable Thread documents and one deterministic read projection over the repository-global task DAG, with guarded creation and a materialization seam the later mutation and bulk slices can safely reuse.

## Scope

- Add Thread domain validation, the entity descriptor/template, flat ID-addressed Markdown storage, narrow read ports/use cases, CLI completion, wire/schema, init/layout, health, and watch coverage.
- Define the pure Thread projection over the shared task graph: many-valued membership, nominal and sound rollups, frontier, external gates, and completed-Thread inconsistency.
- Ship guarded creation in the explicit `unstarted` state. Creation validates membership IDs and cross-kind identity against one authoritative task/Thread snapshot and writes through a lock-free internal Thread materializer.
- Expose read-first CLI/wire surfaces for create, list, show, path, and frontier. Membership and lifecycle verbs belong to the immediately following task.
- Retire the unused Projects scaffold without deleting non-empty legacy content.

## Acceptance criteria

- [x] Thread files own metadata and a sorted task-ID membership set only; task files remain the dependency source of truth.
- [x] A task can belong to multiple Threads and an outside prerequisite appears as an external gate without entering progress totals while still preventing Thread completion until satisfied.
- [x] Views expose nominal `done / total`, sound `drained / total`, graph health, and exact outstanding external gates without contradictory completion UX.
- [x] CLI and wire create/read projections use the shared graph/eligibility analysis and expose stable member/external roles without reimplementing task graph rules.
- [x] Initialization creates `threads/`, stops creating `projects/`, and handles non-empty legacy Projects safely.
- [x] Task/Thread IDs are checked for cross-kind collisions, and an empty Projects scaffold is defined narrowly enough to permit only `.gitkeep` removal.
- [x] Thread creation always persists `unstarted`, validates member existence and cross-kind ID collisions inside one canonical-root guard, and cannot accept an arbitrary lifecycle state through a generic create port.
- [x] The store provides lock-free internal Thread materialization reusable by a
  later compound bulk capability without nesting guarded mutations.
- [x] A real two-store race between Thread creation and task-graph/task-lifecycle mutation proves canonical-root serialization and fresh-snapshot validation; raw-file CAS coverage remains separate.

## Stress tests

- Shared members, external gates, empty Threads, deferred/deprecated members, completed-thread drift after upstream reopen, ID drift/collision, degraded diagnostic reads, guarded-creation races, and watcher reload.
- Domain, descriptor, store, core fake, CLI, completion, schema-golden, init/layout, health, and wire fan-out is explicitly covered.

## Sequencing

Requires completed dependency eligibility enforcement so guarded creation can reuse the settled committed-outcome and repository-guard rules. The guarded membership/lifecycle task follows this document/materialization contract; bulk linking waits for both.

## Readiness checkpoint

Completed 2026-08-29. The shipped lifecycle seam proved reusable, but the original task combined the full first-class-entity fan-out with a second guarded mutation family and task-lifecycle receipt augmentation. Those are independently reviewable boundaries and no longer fit one 5–8 day slice. This task now owns documents, guarded unstarted creation, materialization, and read projections; `ship-guarded-thread-membership-and-lifecycle-mutations` owns mutable membership/lifecycle and affected-Thread receipts.

## Guarded Thread mutation amendment (2026-08-27)

Diagnostic Thread projections may degrade explicitly, but guarded creation loads the required task graph and current Thread state inside one canonical-root repository guard. It validates task existence and cross-kind identity there and lands the Thread write before release. The following mutation task applies the same rule to membership and lifecycle.

Expose a use-case-specific guarded Thread-creation port backed by private store guard/materialization helpers. Keep a lock-free, creation-only Thread document materializer so bulk apply can compose its final new-Thread write under one outer guard. Existing-Thread membership and lifecycle changes require a separate lock-free surgical-update materializer that preserves timestamps, unknown fields, comments, and key order. Calling `MutateTaskGraph` from a Thread planner, or nesting a guarded Thread method from another planner, is forbidden and will correctly fail callback exclusion.

Implement after the eligibility task establishes the first non-dependency guarded-write pattern. Guarded membership/lifecycle follows this Thread persistence/materialization contract; bulk linking waits for that second slice as well as dependency operations.

## Implementation contract (2026-08-29)

- `thread new <title>` accepts zero or more repeatable `--task <reference>` values. Empty Threads are valid; supplied task references resolve exactly against the guarded authoritative snapshot, then persist as a sorted, duplicate-free stable-ID set. Creation accepts no status or lifecycle timestamps and always materializes `unstarted`.
- The guarded creation port receives a pure semantic creation document and returns a committed-outcome receipt. The filesystem adapter owns the canonical-root guard, one strict task-graph scan, one current-Thread scan, cross-kind identity validation, lock-free materialization, whole-snapshot verification, and exclusive create. A cleanup failure after the exclusive create reports `committed: true` and must not be blindly retried.
- Cross-kind identity is bidirectional. Thread creation rejects an ID already used by a task, and every supported task-creation path rejects an ID already used by a Thread while holding the same canonical-root guard. Raw hand edits remain lint/diagnostic input, not a supported bypass.
- One pure `ThreadView` is shared by list, show, frontier, human rendering, and wire output. Members are sorted by stable task ID and carry the existing task graph state. Direct prerequisites of non-deprecated members outside membership are sorted external gates; they never enter `total`, `done`, or `drained`, but outstanding ones prevent sound closure. A deprecated member is withdrawn from both the denominator and actionable boundary gates.
- `total` excludes deprecated members but includes deferred members. `done` counts nominally completed members in that denominator; `drained` counts soundly completed members. Frontier contains only member tasks whose shared graph state is eligible. A completed Thread is inconsistent when it is empty/effectively empty, `drained != total`, any outstanding external gate exists, or relevant graph/member evidence is broken.
- Diagnostic `thread list` and `thread show` retain readable Threads and attach deterministic problems when the task graph, a Thread document, or a member reference is unsound. Views distinguish repository `graph_health` from combined `projection_health`; completed inconsistency carries stable reason codes rather than an unexplained boolean. Dispatch-oriented `thread frontier` returns no tasks unless the relevant projection is healthy, while still returning the diagnosis. List output hoists repository-global graph problems once, and `thread path` remains parse-free.
- `planning/threads/` joins init, layout/watch, space-health, schema/template, completion, CLI docs, and wire schema fan-out. Automatic Projects retirement removes only an otherwise-empty scaffold containing at most `.gitkeep`; any Markdown or other user content is preserved and reported with an explicit migration remedy.

## Implementation evidence (2026-08-29)

Thread is now a first-class flat, ID-addressed document with strict domain validation, ordinary lint coverage, schema/template/completion support, read ports, and filesystem parsing/resolution. Guarded creation always plans unstarted against one authoritative task-and-Thread snapshot, resolves and sorts initial member IDs, checks the shared task/Thread identity namespace, materializes through a lock-free internal seam, verifies the whole snapshot, and preserves committed outcomes through core, CLI, and JSON error recovery.

One shared pure projection now drives thread list, show, and frontier in human and machine output. It reports stable member versus external-gate roles, nominal and sound rollups, direct external gates, shared graph health, completed-Thread inconsistency, and an eligibility-backed frontier that fails closed on unhealthy graph evidence. Fresh init creates threads instead of projects; explicit scaffold repair removes only an empty projects directory or one regular .gitkeep and preserves all other legacy content with an actionable remedy.

Concurrency coverage separates raw-writer CAS from cooperating-writer serialization. Tests cover raw task and Thread snapshot races, cross-kind task/Thread creation, planner re-entry rejection, post-commit guard-release failure, and a two-store lifecycle-versus-Thread race in which the waiting Thread planner must observe the committed lifecycle state.

Validation is clean: go test -race ./..., golangci-lint, go mod tidy -diff, generated CLI docs and schema/goldens, git diff --check, and repository lint. A real planning-space dry run also resolved this task plus the four following Thread slices into a sorted five-member Thread without writing it; create the durable dogfood Thread after this implementation lands.

## Adversarial review closeout (2026-08-29)

The Antigravity audit confirmed the slice boundary and fail-safe missing-member denominator; those findings are respectively tracked to the guarded membership/lifecycle task and retained by design. The Claude audit found two blocking defects and seven hardening issues. All were resolved and both audits are closed.

Deprecated members no longer contribute external gates. Completed inconsistency now has stable explanatory reason codes. Thread views separate repository graph health from projection health, and list output hoists global graph diagnostics once. Safe lint repair now covers ordinary Thread scalars and filename-owned IDs while refusing membership normalization and cross-kind repair collisions. The creation field builder is explicitly creation-only; the next mutation task now requires surgical existing-document updates. Tag order, neutral legacy-init receipts, and discoverable read-only scaffold-repair guidance were also corrected.

Post-audit validation is clean under the full race suite, static analysis, module tidiness, generated docs/schema/goldens, repository and audit lint, and diff hygiene.
