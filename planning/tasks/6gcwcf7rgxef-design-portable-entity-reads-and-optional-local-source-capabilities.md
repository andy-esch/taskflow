---
schema: 1
id: 6gcwcf7rgxef
status: completed
epic: 21-code-quality-architecture-hardening
description: Define shared record, diagnostic, source-location, and local-path contracts for task, epic, audit, and research adapters.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, ports, design, entities]
created: "2026-09-23"
depends_on: [6g6jqqcdehne, 6gcwcf77tvgq, 6gcwcf7gjayh]
updated_at: "2026-09-26"
started_at: "2026-09-26"
completed_at: "2026-09-26"
---

# Design portable entity reads and optional local source capabilities

## Objective

Define the target application boundary for readable and unreadable task, epic, audit, and research
records before mechanically removing their filesystem-shaped contracts. Thread reads and the new
lint port provide useful precedents, but the aggregate `Store` still returns `FileProblem`, embeds
`Path` in domain entities, and combines semantic reads with local path navigation.

## Settled inputs

The completed predecessor slices are constraints for this design rather than questions to reopen:

- A failed-record diagnostic carries entity kind, optional stable ID/slug, optional adapter-neutral
  location, optional local repair path, and message. Neutral location and local path are independent
  and may coexist; neither is an identity source.
- Persistence adapters recover any filename-derived identity during their existing resilient scan.
  Core copies explicit identity and never parses a location or path to manufacture it.
- Aggregate projections publish diagnostics in canonical explicit-field order and retain one
  authoritative read per entity family. Conversion must not hide a rescan.
- Opaque source revisions are adapter-owned guarded-snapshot evidence. They participate in
  fail-closed comparison where optimistic mutation requires them and never enter public wire output.
- Machine contracts evolve additively: the historical `{path,message}` shape remains readable for
  local sources, while pathless adapters leave `path` empty. Human partial-result output leads with
  identity and shows available location/repair context before preserving exit code 11.
- `LintLoadProblem` is now used beyond lint by Board, Summary, retained cross-space state, and audit
  finding reads. Its behavior is proven; whether its name becomes a general application diagnostic
  is deliberately left to this design.

## Design questions

- What entity-specific read-result shapes carry semantic data, portable diagnostics, and only the
  revision evidence that use case actually needs without turning every use case into a generic
  entity API?
- Which operations require source locations, and which require a specifically local filesystem path?
- Should `LintLoadProblem` be promoted/renamed as the general application diagnostic, or should
  ordinary reads use a deliberately compatible wrapper without duplicating semantics?
- How do TUI editing and `<entity> path` obtain optional local capabilities while web, database, and
  service adapters remain pathless?
- Which domain `Path`, `FilenameID`, and `SourceVersion` fields are genuine semantic evidence versus
  adapter metadata that belongs in read envelopes?

## Decision

Use shared, immutable source and diagnostic values inside **entity-specific read ports**. Do not
create a universal entity repository.

The target vocabulary is:

```go
type RecordSource struct {
    ID       string // adapter-resolved canonical stable ID
    Location string // optional opaque source context; never parsed for identity
}

type LoadedRecord[T any] struct {
    Value  T
    Source RecordSource
}

type LoadProblem struct {
    EntityKind string
    EntityID   string
    EntitySlug string
    Location   string // optional opaque source context
    LocalPath  string // optional local repair path; independent of Location
    Message    string
}
```

`LoadedRecord[T]` is a value building block, not a generic port. The application boundary remains
explicit about each use case: `TaskRead`, `EpicRead`, `AuditRead`, and `ResearchRead` contain their
own record type and `[]LoadProblem`; body-aware task and audit snapshots wrap `TaskWithBody` and
`AuditWithFindings` without rereading their sources. The existing `TaskGraphSource`, `LintSource`,
and `AuditSnapshotSource` remain consumer-owned ports rather than being collapsed into a generic
`ReadEntities(kind)` method.

`RecordSource.ID` is authoritative record identity. On the filesystem it is recovered from the
id-led filename; another adapter may obtain it from a primary key or service response. The parsed
domain value keeps the document-declared `id` so lint can report drift, but `FilenameID` leaves the
domain model. Slug remains semantic read/presentation vocabulary: it is used for labels, filtering,
and human references even when an adapter does not use a filename. Location may distinguish and
explain physical records, but it is never an identity fallback and cannot make a duplicate stable
ID selectable. A readable record must have a non-empty source ID; a source whose canonical identity
cannot be recovered reports a `LoadProblem` instead of publishing a record that downstream
navigation cannot address.

`LoadProblem` promotes the already-proven `LintLoadProblem` semantics to general application
vocabulary. The implementation should rename the core type and `LintEntityKind` helpers rather than
introducing a behaviorally identical wrapper. `LocationIsPath` is transitional: the target carries
`Location` and `LocalPath` explicitly, and a filesystem adapter normally sets both to the same
value. The historical wire `path` field and even the internal wire DTO name
`LintLoadProblemJSON` may remain for machine-schema compatibility; renaming a JSON Schema `$defs`
key is not required to make the core name honest.

### Guarded evidence is a stricter wrapper

Ordinary reads do not carry optimistic-concurrency tokens. Only guarded snapshots use versioned
records/problems, conceptually:

```go
type VersionedRecord[T any] struct {
    Record        LoadedRecord[T]
    SourceVersion string
}

type VersionedLoadProblem struct {
    Problem       LoadProblem
    SourceVersion string
}
```

`TaskGraphRead` and `ThreadRead` use those versioned forms because whole-snapshot comparison is
load-bearing for graph, lifecycle, Thread, and repair mutations. Planner/query/public projections
receive the semantic record and source identity only; version tokens remain inside the guarded
boundary. Missing tokens continue to fail closed. Epic, audit, and research ordinary reads do not
gain speculative revisions—their existing mutations obtain and compare fresh evidence inside the
secondary adapter. A read-only graph/Thread adapter may leave versions empty and still serve
queries; it cannot honestly be paired with a guarded mutation capability that relies on snapshot
comparison.

Revision evidence proves that a record has not changed; it does not prove that independently
configured capabilities address the same corpus. Every semantic reader, local-path resolver, and
mutator that can be composed independently therefore also belongs to an opaque, comparable
`SourceSetID`. A complete adapter may expose those capabilities as one source-set bundle; separately
supplied capabilities must publish the same non-empty identity, and service/workspace construction
rejects a mismatch before use. The value identifies an adapter instance/corpus, not an entity; it is
never inferred from `RecordSource.ID`, location, root text, or a revision, and never reaches domain
or wire records. Unsupported capabilities remain absent, so a pathless or read-only adapter does not
have to pretend it can write. The existing implicit-detach and typed-nil rules still apply.

### Local paths are optional capabilities

Readable record envelopes do not carry a local path. Task, epic, audit, research, and Thread each
have a narrow optional local-path source used by `<entity> path`, editor launch, copy-path, and
similar local navigation. The capabilities remain entity-specific (`TaskPathSource`,
`EpicPathSource`, and so on) even if the filesystem adapter implements all of them.

The existing Thread composition rules apply uniformly:

- a complete local adapter may be discovered as both semantic reader and path source;
- explicitly replacing semantic reads detaches any implicit aggregate path source, so remote data
  cannot silently open unrelated local files;
- an explicitly supplied path source is an affirmative composition-root override, and typed-nil or
  absent capabilities fail with the existing typed unavailable-capability classification;
- filesystem resolution remains filename-only and parse-free so malformed documents can be opened
  for repair;
- `Layout` remains the independent directory-watch capability.

The TUI retains semantic browsing in a pathless workspace. It resolves a local path lazily only
when the user invokes an editor/path action and visibly disables that action when the capability is
absent. Selection and restoration use `RecordSource.ID`, never a path or slug. A lazy result is
accepted only when its originating list generation and canonical source ID still match the current
selection, so a refresh cannot retarget a delayed local action.

Malformed records are different: a `LoadProblem.LocalPath` may be the only repair handle when no
stable identity could be recovered, so it deliberately travels with that diagnostic. This does not
make a local path mandatory for a readable record or a remote adapter.

### Duplicate and ambiguous identity

Resilient reads may contain two physical records with the same canonical ID; dropping one would
hide the defect. Slice position or a private snapshot-local index may correlate diagnostics during
one analysis, as task-graph lint does today. Neither that index nor `Location` becomes a selectable
identity. CLI list/lint output attributes every record with whatever location context the adapter
can honestly provide; the TUI entity registry deliberately fails or degrades the affected loaded
list rather than manufacturing separately actionable rows. Durable selection restoration remains
canonical-ID based and clears when that ID becomes ambiguous.

A unique slug/title may locate one physical occurrence for inspection or local repair, but it does
not make a duplicate canonical identity safe to mutate. Mutation authorization must re-resolve and
validate the canonical source ID against the current snapshot immediately before writing. Existing
filesystem CAS already has this property: a slug-selected duplicate reaches `verifyUnchanged` and
returns `ErrAmbiguous`. Portable adapters must preserve that invariant; they need not forbid every
read-only occurrence lookup that helps explain or repair the corruption.

### Mutation results

Domain records should not retain `Path` merely because create/edit/move commands currently print
one. Entity-specific mutation receipts carry semantic records and, only where the command already
promises it, optional local outcome metadata. Create receipts distinguish the planned destination
from a committed resulting path so a local dry run remains explanatory before any record exists.
Rename/relocation receipts retain exact source and destination paths plus their durable stage for
partial-error recovery. Pathless adapters omit those values rather than manufacturing them, and
ordinary flat-layout lifecycle updates do not gain a fictitious move receipt. Post-success path
resolution is acceptable only for a committed non-moving operation whose resolver shares the
mutator's `SourceSetID`; it cannot replace dry-run or durable-prefix evidence. These operation-
specific receipts must not be flattened into the ordinary read envelope.

## Crossing inventory

| Current crossing | Classification | Target / owner |
| --- | --- | --- |
| `TaskStore.ListTasks`, `ListTasksWithBodies`, `EpicStore.ListEpics`, `AuditStore.ListAudits*`, `ResearchStore.ListResearch`, and `SummaryStore` return `domain.FileProblem` | Filesystem-shaped application debt | Entity-specific read snapshots return `[]LoadProblem`; the ordinary-list migration owns this. `SummaryStore` disappears in favor of the same epic/audit sources used elsewhere. |
| `domain.FileProblem` inside resilient filesystem scans and guarded/local repair code | Local adapter evidence | Keep private to `internal/store` while useful; remove it from core/domain contracts after the last compatibility adapter is gone. |
| `taskStoreGraphSource`, `TaskGraphReadFromFiles`, and `NewTaskGraph(...FileProblem)` | Transitional local compatibility | Native sources return identity-bearing/versioned graph records. Retain only a clearly named test/legacy adapter if still needed; no new core path inference. |
| `Path` on task, epic, audit, research, and Thread domain values | Adapter metadata | Remove from semantic domain values. Use optional opaque `RecordSource.Location`, diagnostic `LocalPath`, explicit local-path sources, or an operation-specific receipt as appropriate. |
| `ResolveTaskPath`, `ResolveEpicPath`, `ResolveAuditPath`, and `ResolveResearchPath` on aggregate `Store` | Optional local capability mixed into semantic persistence | Move to independent entity-specific path sources, matching `ThreadPathSource`; update composition and unavailable behavior together. |
| `FilenameID` on task, audit, research, and Thread | Adapter-derived canonical identity | Replace with `LoadedRecord.Source.ID`. The parsed frontmatter `ID` remains available for drift lint. Epic `ID` is already its semantic/canonical value but still enters the uniform envelope. |
| `SourceVersion` on readable task/Thread domain values and unreadable graph/Thread problems | Guarded adapter evidence | Move into `VersionedRecord` / `VersionedLoadProblem` used only by guarded task-graph and Thread snapshots. Never place it in ordinary reads or wire values. |
| `TaskGraphSourceRef.Location` and diagnostic locations | Portable explanatory source context | Retain as opaque context sourced from the envelope; never parse it, require it, or equate it with a local path. |
| `StatusFellBack`, `BucketFellBack`, legacy-field presence, and body-derived task title | Parsed/derived semantic evidence | Keep with semantic read values. They describe document meaning or lint state, not how to reopen the source. |
| Domain-record `Path` used by create/edit/move/lifecycle/Thread receipts | Local outcome metadata accidentally smuggled through the entity | Replace deliberately with local resolution after success or an optional path on the entity-specific receipt; preserve dry-run and partial-durability evidence. |

## Target read shapes

The exact Go field names may follow local conventions, but the boundary must have these properties:

- `TaskGraphRead` owns `[]VersionedRecord[domain.Task]` plus
  `[]VersionedLoadProblem`; one internal loaded/versioned scan supplies compatibility projections
  for existing graph, list, board, filter, show, and lint consumers. Ordinary projections discard
  versions rather than initiating a second scan.
- `ThreadRead` owns `[]VersionedRecord[domain.Thread]` plus
  `[]VersionedLoadProblem`; list/show/compose projections discard versions, while guarded stores
  compare the complete values.
- Ordinary epic and research reads return `[]LoadedRecord[domain.Epic|Research]` with
  `[]LoadProblem` from one resilient scan.
- The body-aware task lint read returns `[]LoadedRecord[TaskWithBody]`; the body-aware audit read
  returns `[]LoadedRecord[AuditWithFindings]`. Finding queries, audit lint, repository lint, and
  Summary reuse those snapshots rather than rereading through a path.
- Single-document show reads return the same loaded record shape plus body. They do not acquire a
  path as a side effect.
- TUI items consume narrow entity projections carrying semantic display data plus source identity;
  generic loaded-record mechanics need not leak through rendering code. Generation-aware async
  messages revalidate source identity before changing detail state or invoking a local capability.

These shapes support local Markdown, database rows, remote service objects, and read-only caches:
all must supply canonical stable identity and semantic data; any may supply an opaque location;
only guarded snapshot adapters supply revisions; only local navigation adapters supply paths.

Guarded planners that authorize against existing identity receive the loaded record after its
version is stripped, not a bare domain value: source ID drift and cross-kind collision checks remain
possible without exposing persistence tokens.

## Compatibility and migration

1. **Promote the diagnostic vocabulary and introduce one authoritative scan.** Rename the core
   diagnostic, add the shared source/record values, adapt the filesystem scan once, and supply
   compatibility projections for existing graph, ordinary list, single-show, lint, Summary, and
   wire consumers. Preserve public `{path,message}` fields while adding already-established
   kind/identity/location fields; advance the machine revision only for observable envelope changes.
2. **Migrate identity consumers and preserve local mutation outcomes.** Move TUI list/detail keys and
   refresh checks to canonical source identity while retaining its fail-closed duplicate behavior.
   Add operation-specific local receipts for dry-run and partial-durability paths. Bind independently
   composable readers, path sources, and mutators to one opaque source set.
3. **Move source metadata out of domain records and split local capabilities.** Only after all
   consumers have an owned replacement, remove `Path`, `FilenameID`, and `SourceVersion` from
   semantic task/epic/audit/research/Thread values. Move every remaining `Resolve*Path` method out of
   aggregate semantic stores, apply the detach/override and source-set rules, and switch CLI/TUI
   local actions. A pathless fake must drive all semantic reads and non-local UI behavior.
4. **Close primary-adapter bypasses.** Route repair, link integrity, and completion through named
   application capabilities after the read/path boundary is stable, then enforce the composition
   import boundary.
5. **Remove compatibility debt.** Delete core-facing `FileProblem` conversions and obsolete domain
   metadata only after filesystem, fakes, CLI, TUI, wire mappings, generated schema, and docs have
   migrated in the same slice that removes each contract.

This is intentionally a staged internal refactor. Public JSON retains the historical `path` field
for local records and leaves it empty for pathless sources; additive identity/location fields keep
their established meaning. Human partial-result output and exit code 11 do not change. No step
changes Markdown storage, graph policy, lifecycle policy, or write atomicity.

## Rejected alternatives

- **One `EntityStore` or `ReadEntities(kind)` port:** erases body-aware, graph, lifecycle, and
  selector semantics and would make every adapter claim unsupported capabilities.
- **Keep `Path`/`FilenameID`/`SourceVersion` optional on domain structs:** compiles for pathless
  adapters but leaves every consumer free to accidentally depend on empty storage metadata; the
  current Thread value demonstrates that partial separation is not a stable end state.
- **Use location as identity or infer identity from it:** breaks remote adapters and repeats the
  path-parsing defect already removed from graph lint.
- **Put a revision on every read:** exposes persistence policy to use cases that neither compare nor
  authorize against it and invites public leakage.
- **Make local path a URI:** obscures capability availability and causes local-only editor/path
  behavior to accept values it cannot use.

## Constraints

- Keep Markdown and stable IDs as product contracts; adapter-neutral does not mean hiding the
  markdown-first model.
- Preserve resilient partial reads, parse-free repair lookup, guarded snapshot comparison, and
  existing public machine compatibility.
- Keep entity-specific use cases explicit; do not introduce a universal `EntityStore` abstraction.
- Reconcile rather than duplicate the broader shared entity-integrity design task.
- Treat canonical-identity corruption as visible but not interactively writable: analytical
  projections retain occurrences, while stable UI/action registries fail closed.

## Acceptance criteria

- [x] A decision table classifies every `domain.FileProblem`, entity `Path`, `Resolve*Path`, and
      source-version crossing as semantic, optional local capability, adapter evidence, or debt.
- [x] The target port and record shapes support filesystem, database, remote-service, and read-only
      adapters without invented paths or ambiguous identity.
- [x] The chosen diagnostic vocabulary preserves every settled input above and gives the existing
      `LintLoadProblem` consumers an explicit migration path.
- [x] The design explains duplicate-ID attribution, malformed-record repair, TUI editing, and
      optimistic-concurrency evidence.
- [x] Compatibility and migration sequencing are explicit for core, wire, CLI, TUI, fakes, and the
      filesystem adapter.
- [x] The implementation tasks behind this design are amended if the chosen boundary invalidates
      their current assumptions.

## Out of scope

- Implementing the port migration.
- Choosing a database or changing the authoritative storage format.
- Generalizing entity creation or mutation behind one generic API.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Shared entity-integrity foundations](6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md)
- [Thread path/read split precedent](6g5ryqqx5ab7-split-local-thread-path-resolution-from-portable-thread-reads.md)
- [Thread diagnostic precedent](6g5rxq1ravd3-make-thread-read-diagnostics-adapter-neutral.md)
- Design review:
  [Codex](../audits/6gdx1bvdf16p-2026-09-26-portable-entity-read-local-source-capability-design-codex.md)
  and
  [Antigravity](../audits/6gdx1bvz683y-2026-09-26-portable-entity-read-local-source-capability-design-antigravity.md)

## Design-pass progress (2026-09-26)

Inventoried the aggregate and narrow ports, every remaining core `FileProblem` crossing, the five
domain record types, task/Thread guarded snapshot comparison, lint/audit snapshot precedents, CLI
path/create output, and TUI list/detail/editor path consumers. The review exposed the remaining
`Path`/`FilenameID`/`SourceVersion` fields on Thread as an incomplete precedent rather than a shape
to copy. The downstream tasks and Thread sequence now introduce neutral read snapshots first, move
source/local metadata second, and only then close direct CLI adapter access. Planning lint and the
re-sequenced Thread projection were healthy before review; the closeout below records the accepted
amendments and deliberately rejected remedies.

## Adversarial review closeout (2026-09-26)

The reviews exposed three valid implementation-sized gaps and one migration ambiguity. The design
now requires source-set identity for independently composed capabilities, operation-specific local
mutation receipts, a one-scan compatibility stage, and an explicit TUI identity/refresh migration
before domain metadata removal. Those concerns are tracked as separate Thread tasks so the final
source/path slice does not become an unreviewable cutover.

Two proposed remedies were deliberately not adopted. The TUI will keep its current fail-closed
registry semantics instead of inventing snapshot-local actionable row handles for duplicate IDs;
CLI list/lint remains the explanatory surface for every corrupt occurrence. Likewise, non-id-led
filesystem files already become file problems before semantic parsing, so the portable contract does
not newly hide a class of lintable domain records. The filesystem CAS tests also demonstrate that a
slug-selected record with a duplicate canonical ID fails `ErrAmbiguous` before mutation, even though
read-only inspection by a unique slug remains useful for repair. The settled wire rule remains
unchanged: historical `path` is emitted as an empty string when no local path exists.
