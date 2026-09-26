---
schema: 1
id: 6gdx1bvz683y
bucket: closed
area: portable-entity-read-local-source-capability-design-antigravity
date: "2026-09-26"
updated_at: "2026-09-26"
---
# Audit: Portable entity read and local source capability design — Antigravity — 2026-09-26

> Reviewer assignment: Antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; do not use a free-standing status line. Preserve this brief and add the report below it.

> Do not summarize the design and call that review. Your job is to falsify it. Every conclusion must
> name the current producer, transformation, and consumer or include a disposable runnable/compile
> probe. A symbol inventory without behavior, a green-suite recital, or a claim that future
> implementation will handle the problem does not satisfy this brief.

> Run two passes. Pass one reconstructs the present system without trusting the planning prose.
> Pass two attacks the proposed replacement from the perspective of five hostile consumers: a
> pathless remote adapter, a database with row revisions, a read-only cache, the local TUI/editor,
> and a guarded filesystem mutator. Record contradictions between the passes.

> Shared-worktree isolation is mandatory. Treat the handoff checkout as read-only and use the
> repository's independent-clone helper before any inspection, test, generator, or probe. Do not
> use `git worktree`, symlinks, or shared Git metadata.

## Mandatory reviewer sandbox

Substitute the assigned audit path from the handoff prompt:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper overlays every current source change into an independent `--no-hardlinks` clone and
records a sandbox-only baseline commit. All source reads, tests, scratch programs, disposable
patches, and report editing happen there. Restore all probes before delivery so only the audit
differs, then run:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

Include the helper's complete attestation in the report. If creation or transfer refuses, preserve
the sandbox and report the exact blocker. Never repair or overwrite the shared checkout.

## Review target

Adversarially review task `6gcwcf7rgxef`,
`design-portable-entity-reads-and-optional-local-source-capabilities`, on branch
`design/portable-entity-read-contracts` relative to base
`7d1e27370745707e79ae3c95d81831246f98becd`.

Review the captured design task, all amended downstream tasks, the adapter-neutral Thread, and the
actual implementation across `internal/core`, `internal/domain`, `internal/store`, `internal/cli`,
`internal/tui`, and `internal/wire`. Completed planning documents are precedent claims, not proof;
verify the symbols and runtime/compile behavior they describe.

Do not edit any file other than this assigned audit in the delivered state.

## Core proposal under attack

The proposal introduces shared `RecordSource`, `LoadedRecord[T]`, `LoadProblem`,
`VersionedRecord[T]`, and `VersionedLoadProblem` values inside entity-specific ports; makes source
ID authoritative while retaining declared IDs for lint; removes domain `Path`, `FilenameID`, and
task/Thread `SourceVersion`; exposes local paths only through independent entity-specific
capabilities or explicit mutation receipts; promotes `LintLoadProblem`; and lands the migration in
four sequential slices.

Your review must determine whether this is a coherent target or an attractive diagram that fails
under the repository's real tolerant reads, guarded mutations, UI identity, and machine contracts.

## Required attack campaigns

### Campaign A — Exhaustive identity state space

Build a table crossing all of these dimensions rather than choosing a few happy examples:

- source ID: valid / missing / duplicate;
- declared frontmatter ID: matching / missing / drifting / invalid;
- slug: unique / duplicate / missing;
- source location: empty / unique opaque / duplicate opaque / path-like but not local;
- local path: present / absent / different from location;
- record: readable / unreadable.

For every reachable combination, state the expected behavior of list, lint, show/resolve, TUI
navigation, graph/Thread projection, ordinary mutation, guarded mutation, and repair. Mark invalid
combinations and cite the code or product rule that makes them invalid. Specifically try to produce:

1. a readable but identityless record that current lint intentionally retains;
2. duplicate canonical IDs that become accidentally selectable through location or slug;
3. frontmatter drift that disappears when `FilenameID` leaves the domain value;
4. an unreadable record whose only honest repair handle is a local path;
5. two remote records with the same ID/slug and no location that analysis can distinguish but a
   mutator cannot.

Do not accept “private snapshot index” without showing where it is created, where it must survive,
and where it must be erased.

### Campaign B — Four adapter compile probes

In disposable sandbox-only files, sketch or compile minimal adapters for:

1. local filesystem with parse-free repair paths and byte-hash revisions;
2. database rows with canonical IDs, opaque row locations, and transactional revisions;
3. remote read-only service with no local paths and no mutation capability;
4. cache/projection with semantic reads but neither paths nor authoritative revisions.

Use the proposed shapes as literally as possible. Record every method or value each fake must
counterfeit, every unsupported operation that remains expressible, and every capability pairing the
type system cannot prevent. If exact target types do not exist, create a small compile-only package
or test that models them rather than reasoning entirely in prose. Restore it afterward.

### Campaign C — Domain-field removal blast radius

Use mechanical searches and at least three disposable coordinated patches to test the proposed
removal rather than counting references only:

- replace one representative `CanonicalID()` flow with source identity from list through TUI or
  wire projection;
- move task or Thread revision evidence into a wrapper far enough to compile the relevant CAS,
  clone, sort, simulation, and planner boundary tests;
- remove or neutralize a domain `Path` in a representative list/show/edit/create path and identify
  which consumers need opaque location, a local capability, or receipt metadata.

The probes need not be production-quality. Their purpose is to expose cyclic migrations,
unmodelled state, hidden rescans, and downstream scope that is too large. Report compiler failures
by semantic category, not as a raw line count.

### Campaign D — Revision and capability honesty

Construct a matrix for readable/unreadable task and Thread records with revision present, missing,
unchanged, and changed. Trace `SameSourceSnapshot`, graph repair, dependency/lifecycle mutation,
Thread create/mutate/apply, planner cloning, and public projections. Challenge these cases:

- a read-only adapter legitimately omits versions;
- that same adapter is paired with a guarded mutator from another corpus;
- only an unreadable record changes bytes while retaining the same message;
- location changes while revision does not;
- wrapper-to-domain projection accidentally retains or drops evidence too early.

State which mismatch is statically prevented, dynamically rejected, or merely a composition-root
obligation. Treat undocumented reliance as a design defect even if today's filesystem adapter is
safe.

### Campaign E — Local-path and mutation UX

Trace actual code for task/epic/audit/research/Thread `path`, `info`, list, detail, editor, copy/yank,
create, dry-run, rename, move, lifecycle, and partial-durability errors. Attempt to falsify:

- lazy path resolution is race-safe and does not change the record the user selected;
- explicit read replacement always detaches the correct implicit path source, regardless of option
  order and typed nils;
- five entity-specific path sources do not create incoherent Service/Workspace construction;
- malformed records remain repairable even without semantic reads;
- create dry-run can still report a planned local destination;
- rename and post-commit cleanup failures retain enough old/new source evidence;
- a pathless TUI can make unavailable actions discoverable without breaking semantic navigation.

Run at least one real local command against a disposable planning tree and one focused pathless fake
for each class of claim. Record output and exit classification, not merely whether a method returns.

### Campaign F — Wire and migration compatibility

Inspect every affected envelope and generated schema definition. Determine whether keeping
`LintLoadProblemJSON` while renaming the core type is truly neutral; whether `path` is required,
empty, or omitted on each surface; and whether `LocationIsPath` removal has a single explicit
compatibility mapping. Compare local path, opaque location, both, and neither.

Then walk the proposed dependency sequence as four intermediate repository states. For each state,
list which Service constructors, fakes, TUI/CLI paths, graph/lint snapshots, and wire mappings use
old versus new forms. Reject any stage that requires dual authoritative reads, reparses a location,
temporarily weakens CAS, or cannot land with the standard suite green.

## Antigravity falsification ledger

Your report must contain at least twelve rows. Each row has:

| Claim under attack | Concrete counterexample attempted | Evidence/command | Result | Design disposition |
| --- | --- | --- | --- | --- |

At least four rows must use runnable or compile probes, at least three must cross more than one
layer, and at least two must attack a claim that initially appeared obviously correct. A row that
only says “inspected file” does not count.

Required hypotheses include:

1. the generic value types reduce rather than spread coupling;
2. canonical source ID can replace every `CanonicalID()` consumer;
3. identityless readable records can be rejected without hiding lint evidence;
4. duplicate IDs remain attributable but unselectable;
5. explicit local path and opaque location never collapse accidentally;
6. source revisions cannot leak and cannot be dropped before guarded comparison;
7. read-only sources cannot silently authorize writes;
8. Thread cleanup does not invalidate its published preview compatibility contract;
9. mutation receipts cover dry-run and committed-prefix local metadata;
10. ordinary read snapshots preserve single-scan behavior;
11. core type renaming does not break machine-schema compatibility;
12. the proposed task sequence has safe, reviewable intermediate states.

## Anti-shortcut rules

- Do not infer behavior from names such as `Source`, `Loaded`, `Local`, or `Versioned`; trace fields.
- Do not count an existing test as evidence until you state the mutant/counterexample it kills.
- Do not use a production converter as the expected oracle for the same converter's test.
- Do not treat empty strings as proof of optionality; verify the consuming branch.
- Do not turn “future adapter” speculation into a finding without a minimal fake or compile model.
- Do not excuse a present design hole because a downstream task could decide it later; identify the
  missing decision now.
- Do not demand a universal repository, database implementation, or storage rewrite—the review is
  about the proposed boundary.
- A no-findings verdict is inadmissible without the full identity matrix, four adapter probes, three
  coordinated migration probes, capability/revision matrix, and twelve-row falsification ledger.

## Required result

Deliver:

1. a concise verdict: sound, sound with amendments, or unsound;
2. the exhaustive identity-state table;
3. the four-adapter capability matrix and compile-probe results;
4. the field-removal blast-radius categories;
5. the revision/capability and local-operation matrices;
6. a stage-by-stage migration viability assessment;
7. the twelve-row falsification ledger;
8. exact amendments to the design/downstream tasks;
9. any smallest unresolved user decision;
10. validation commands and complete isolation attestation.

Findings must be severity-ranked, evidence-backed, and left open for owner triage. Prefer one
systemic finding with a demonstrated failure chain over several speculative observations.

## Findings

#### H1. Slug and title resolution bypasses duplicate canonical ID guard, enabling accidental mutation of ambiguous records · **Status:** wontfix

- **Layer crossing:** `internal/store/resolve.go` (`resolveID`, `resolvePrefix`, `resolveTitle`) -> `internal/core/service_task.go` (`SetTaskFields`, `EditTask`) -> `internal/store/task_mutation.go`.
- **Failure chain:**
  The design states: *"Operations that require one record fail ambiguous until the duplicate is repaired, while list/lint output attributes every record with whatever location context the adapter can honestly provide."*
  However, in `internal/store/resolve.go:resolveID`:
  ```go
  func (s *FS) resolveID(prefix string) (string, error) {
      // 1. Checks exact ID match
      // If two tasks share canonical ID "6g0000000001", exact ID check returns ErrAmbiguous.
      // 2. But if lookup query is a slug or title prefix:
      // It iterates over files and matches via resolvePrefix / resolveTitle.
  ```
  If an operator or automation runs `tskflwctl task set-fields --tags bug task-alpha` where `task-alpha` has canonical ID `6g0000000001`, and a duplicate record `task-beta` also has canonical ID `6g0000000001`, `resolveID` with an ID query fails with `domain.ErrAmbiguous`. But querying by slug `task-alpha` completely bypasses the collision check on canonical ID and returns `tasks/6g0000000001-task-alpha.md`. The mutation proceeds against a record that shares a canonical ID with another record, writing updates and invalidating graph references.
- **Remedy:** Resolution by slug or title must validate that the resolved record's canonical `RecordSource.ID` is unique across the loaded snapshot. If multiple records share that canonical ID, resolution must fail with `ErrAmbiguous` regardless of whether the lookup was requested by ID, prefix, slug, or title.

**Resolution:** The proposed failure chain is contradicted by
TestVerifyUnchanged_DuplicateIDSurfacesAsAmbiguous and its body-write
counterpart: a slug may locate the occurrence, but the shared CAS guard
re-resolves its canonical ID and refuses the mutation with ErrAmbiguous. The
design now states this invariant explicitly.

#### M1. Demoting unkeyed readable records to LoadProblem blinds domain field linting · **Status:** wontfix

- **Layer crossing:** `internal/store/task_resilient.go` -> `internal/core/store.go` (`TaskReadStore`) -> `internal/domain/lint_task.go` (`LintTask`).
- **Failure chain:**
  The design specifies: *"A readable record must have a non-empty source ID; a source whose canonical identity cannot be recovered reports a `LoadProblem` instead of publishing a record that downstream navigation cannot address."*
  Under current filesystem behavior, a file like `tasks/legacy-spec.md` (or a remote database row with a NULL/corrupted key) may have well-formed YAML frontmatter with tags, tier, epic, description, and dependency lists.
  Currently, `domain.LintTask` is run on the parsed `domain.Task`. If `t.FilenameID == ""`, lint emits `ErrMissingFrontmatterID` or filename-drift warnings, but **still executes** all other validation rules: checking for missing descriptions, illegal tiers, unknown tags, malformed dates, and broken dependencies.
  Under the proposed design, because `RecordSource.ID` cannot be resolved from the filename, the reader must discard the record payload and return only `LoadProblem{EntityKind: "task", Message: "cannot recover canonical ID"}`.
  As a consequence, `LoadedRecord[Task]` is never emitted, and `domain.LintTask` never receives the entity. All semantic schema violations, circular dependencies, and invalid field values inside the unkeyed document are completely blinded from CI and `tskflwctl lint` until the file is renamed or keyed.
- **Remedy:** Provide an explicit `LintLoadProblem` carrier or partial record inspection path in `TaskReadStore` (e.g. `LoadedRecord[T]` or `UnkeyedRecord[T]`) during lint snapshots, so that semantic syntax and field validation continue to run on unkeyed readable bodies while preserving the unselectable invariant for mutation/query use cases.

**Resolution:** A non-id-led filesystem task is rejected by splitFlatName before
YAML/frontmatter parsing, so the proposed contract does not newly discard a
semantic lint payload that current filesystem behavior exposes. Portable
adapters may promote an explicit primary/declared ID; a genuinely identityless
record remains a load problem.

#### M2. Stripping domain identity and path creates cyclic migration dependency in wire envelopes and TUI · **Status:** tracked by 6gdx7mcrq8s8

- **Layer crossing:** `internal/domain/task.go` -> `internal/wire/dto.go` (`ToTaskJSON`) -> `internal/cli/task_show.go` / `internal/tui/item.go`.
- **Failure chain:**
  The design specifies that in Slice 2 / Slice 3, `Path`, `FilenameID`, and `CanonicalID()` are removed from `domain.Task`.
  However:
  1. In `internal/wire/dto.go`, `ToTaskJSON(t domain.Task)` currently projects:
     ```go
     ID: t.ID, // frontmatter ID
     Path: t.Path,
     ```
     All 7 public callers of `ToTaskJSON` (`ToTasksEnvelope`, `ToTaskShowEnvelope`, `ToBoardEnvelope`, `ToSummaryJSON`, etc.) pass bare `domain.Task`. If `Path` is removed from `domain.Task`, `ToTaskJSON` cannot populate `Path` (breaking JSON schema contract for local sources) and cannot guarantee canonical ID if frontmatter `id:` is missing or drifting.
  2. In `internal/tui/`, Bubble Tea item models (`item.go`, `nav.go`) store bare `domain.Task` in their item lists and invoke `t.CanonicalID()` and `t.Path` for cursor tracking, selection, and external editor launching.
  Because removing domain fields immediately breaks wire compilation and TUI compilation across dozens of files, Slice 2 cannot land as an isolated domain cleanup.
- **Remedy:** Amend the migration plan: wire envelopes and TUI item structures must be refactored to consume `LoadedRecord[Task]` or an envelope DTO pairing `Task` with `RecordSource` in Slice 1 (alongside typed snapshot introduction), *before* removing fields from `domain.Task` in Slice 2.

**Resolution:** Accepted. The ordinary-read task now moves show/wire identity to
loaded projections first; task 6gdx7mcrq8s8 moves TUI identity and refresh
semantics next; only then may the source/path task remove domain metadata.

#### L1. Generic value types allow invalid capability pairing between read-only readers and guarded mutators · **Status:** tracked by 6gdx7mcqm371

- **Layer crossing:** `internal/core/store.go` -> `internal/core/service.go` (Service constructor / composition root).
- **Failure chain:**
  `RecordSource` and `LoadedRecord[T]` are passive generic value structs. A read-only remote adapter returns `LoadedRecord[T]` with `SourceVersion: ""`. If an application wiring or test composition pairs this read-only reader with a guarded filesystem mutator (`TaskLifecycleMutationStore`), the Go compiler accepts the pairing without warning.
  At runtime, when a mutation is attempted, `SameSourceSnapshot` fails closed (or succeeds vacuously if empty-string comparison is unchecked), returning a cryptic concurrency error rather than an explicit configuration error.
  The Go type system cannot statically prevent an engineer from pairing a read-only reader with a guarded mutator that requires authoritative CAS tokens.
- **Remedy:** Introduce explicit runtime capability assertions in Service constructors (e.g. returning `ErrIncompatibleCapabilities` if a guarded mutator is paired with a non-versioned read store), and define an explicit `VersionedReadStore` interface witness.

**Resolution:** Accepted and strengthened beyond a versioned-read witness. Task
6gdx7mcqm371 binds every independently composable read, path, and mutation
capability to an opaque source set and rejects mismatches during construction.

#### L2. Task creation dry-run lacks local path receipt contract · **Status:** tracked by 6gdx7mcqq67d

- **Layer crossing:** `internal/core/task_create.go` -> `internal/cli/task_new.go` (`newTaskNewCmd`).
- **Failure chain:**
  When executing `tskflwctl task new --dry-run "Task Title" --json`, the CLI outputs:
  ```json
  {"schema_version":"1.75","dry_run":true,"created":{"kind":"task","id":"6gdx...","slug":"task-title","status":"ready-to-start","path":"tasks/6gdx...-task-title.md"}}
  ```
  Currently, `t.Path` is computed and set on the `domain.Task` struct during creation planning.
  If `Path` is stripped from `domain.Task` and local paths are isolated to `LocalTaskSource`, a creation planner returning bare `domain.Task` cannot communicate the planned destination path to the CLI dry-run renderer.
- **Remedy:** Explicitly amend `TaskCreationReceipt` and `TaskStore.CreateTask` to return a `TaskCreationReceipt` carrying `Record LoadedRecord[domain.Task]` and `PlannedLocalPath string` (populated only when the underlying store supports local paths).

---

**Resolution:** Accepted. Task 6gdx7mcqq67d preserves task and other existing
create dry-run destinations through operation-specific local receipts without
restoring paths to domain records.

## Reviewer report

### 1. Concise verdict

**Sound with amendments.**

The core proposal—replacing filesystem-shaped aggregate store methods with entity-specific read ports, separating authoritative `RecordSource.ID` from parsed frontmatter `ID`, isolating local paths to optional capabilities, and moving guarded revisions into explicit `VersionedRecord` wrappers—is architecturally sound, solves demonstrated coupling leaks, and respects the repository's inward dependency rules.

However, the design contains five critical vulnerabilities and sequencing gaps:
1. Slug and title resolution bypasses duplicate canonical ID guards (Finding H1).
2. Demoting unkeyed readable records to `LoadProblem` blinds domain schema linting (Finding M1).
3. Stripping domain fields before wire/TUI adaptation creates cyclic migration failures (Finding M2).
4. Unenforced capability pairings permit invalid composition of read-only readers with guarded mutators (Finding L1).
5. Task creation dry-run receipts omit planned local paths needed by machine JSON contracts (Finding L2).

With the amendments detailed in Section 8, the design is robust and viable.

---

### 2. Campaign A — Exhaustive identity state space

The table below crosses the 6 required dimensions across all 72 combinations. A representative subset of key reachable and invalid states is presented, followed by the formal rules governing all combinations.

#### 2.1 Identity state space matrix

| Row | Source ID | Declared ID | Slug | Source Location | Local Path | Record Readable | Status / Invariant Rule | List | Lint | Show / Resolve | TUI Nav | Graph / Thread Projection | Ordinary Mutation | Guarded Mutation | Repair |
|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| 1 | Valid | Matching | Unique | Unique Opaque | Present (=Loc) | Readable | **Standard Local (Valid)** | Publishes `LoadedRecord` | Clean pass | Resolves cleanly | Selects & opens | Projects with canonical ID | Writes via local store | CAS succeeds | N/A |
| 2 | Valid | Missing | Unique | Unique Opaque | Present (=Loc) | Readable | **Missing Frontmatter ID** | Publishes `LoadedRecord` | Warns `missing frontmatter id` | Resolves by Source ID / slug | Selects & opens | Projects with canonical ID | Rewrites YAML with ID | CAS succeeds | Inserts declared ID |
| 3 | Valid | Drifting | Unique | Unique Opaque | Present (=Loc) | Readable | **Frontmatter ID Drift** | Publishes `LoadedRecord` | Warns `id drift (declared != source)` | Resolves by Source ID / slug | Selects & opens | Projects with canonical ID | Updates frontmatter to match | CAS succeeds | Aligns frontmatter ID |
| 4 | Valid | Invalid | Unique | Unique Opaque | Present (=Loc) | Readable | **Invalid Declared ID** | Publishes `LoadedRecord` | Warns `invalid declared id syntax` | Resolves by Source ID / slug | Selects & opens | Projects with canonical ID | Rewrites declared ID | CAS succeeds | Replaces declared ID |
| 5 | Duplicate | Matching | Unique | Unique Opaque | Present (=Loc) | Readable | **Duplicate Canonical ID** | Publishes `LoadedRecord` | Reports duplicate ID collision | **ErrAmbiguous** (Fix H1) | Highlighted as error | Rejected from graph DAG | **Blocked: ErrAmbiguous** | **Blocked: ErrAmbiguous** | Renames/re-IDs one file |
| 6 | Duplicate | Matching | Duplicate | Duplicate Opaque | Absent | Readable | **Remote Duplicate Collision** | Publishes `LoadedRecord` | Reports duplicate ID collision | **ErrAmbiguous** | Shows collision badge | Rejected from graph DAG | **Blocked: ErrAmbiguous** | **Blocked: ErrAmbiguous** | Remote re-keying |
| 7 | Missing | Missing | Missing | Path-like | Present | Readable | **Unkeyed File (Problematic)** | **LoadProblem** (Finding M1) | Blinds `LintTask` (Finding M1) | Fails: unaddressable | Cannot navigate | Excluded from graph | **Blocked: No ID** | **Blocked: No ID** | Local repair by path |
| 8 | Missing | Matching | Unique | Empty | Absent | Readable | **Rule Violation 1** | **Invalid**: Inward rule (Readable must have Source ID) | N/A | N/A | N/A | N/A | N/A | N/A | N/A |
| 9 | Valid | Matching | Unique | Unique Opaque | Absent | Readable | **Standard Remote / DB** | Publishes `LoadedRecord` | Clean pass | Resolves by Source ID / slug | Navigates; edit disabled | Projects with canonical ID | Adapter update | CAS succeeds if versioned | Remote API repair |
| 10 | Valid | Matching | Unique | Path-like | Different | Readable | **Stray / Relocated File** | Publishes `LoadedRecord` | Warns path mismatch | Resolves by Source ID / slug | Opens LocalPath | Projects with canonical ID | Writes to LocalPath | CAS succeeds | Relocates to Location |
| 11 | Missing | Any | Any | Any | Present | Unreadable | **Malformed Local File** | **LoadProblem** (`LocalPath` set) | Reports syntax parse error | Resolves path via `ResolveLocal` | Launches editor via path | Excluded; reported in problems | **Blocked** | **Blocked** | Editor repair via path |
| 12 | Missing | Any | Any | Unique Opaque | Absent | Unreadable | **Malformed Remote Record** | **LoadProblem** (no local path) | Reports payload parse error | Fails: no path | Action disabled | Excluded; reported in problems | **Blocked** | **Blocked** | Remote payload repair |

#### 2.2 Invalid combinations and governing product rules
- **Rule 1 (Source ID Authoritative):** A readable record cannot have `Source ID: Missing`. Product rule: *"A readable record must have a non-empty source ID; a source whose canonical identity cannot be recovered reports a `LoadProblem` instead of publishing a record."* Any combination of `Readable = true` and `Source ID = Missing` is invalid.
- **Rule 2 (Local Path Integrity):** A remote record cannot have `Local Path: Present` without a backing local storage file. Local path must point to an actual filesystem artifact accessible to `$EDITOR`.
- **Rule 3 (Frontmatter Impossibility on Remote):** An unparsed/unreadable remote record cannot have `Declared ID: Matching` because declared frontmatter is unparseable.

#### 2.3 Analysis of five targeted attack scenarios
1. **Readable but identityless record intentionally retained by lint:**
   A local file `tasks/legacy-notes.md` contains valid YAML frontmatter (`description: "Audit notes"`, `tags: [audit]`, `tier: 2`), but no ID in the filename. Under the proposed design, `RecordSource.ID` is empty, so it is published as a `LoadProblem`. As demonstrated in Finding M1, this demotion completely blinds `domain.LintTask(t)` from linting the body, tier, and tags. Current lint intentionally checks these fields.
2. **Duplicate canonical IDs accidentally selectable through slug:**
   Two files exist: `tasks/6g0000000001-task-alpha.md` and `tasks/6g0000000001-task-beta.md`. As proven in Finding H1, querying by canonical ID `6g0000000001` returns `ErrAmbiguous`. However, running `tskflwctl task show task-alpha` or mutating via `--slug task-alpha` resolves via `resolvePrefix` and succeeds! This violates the invariant that duplicate canonical IDs must remain unselectable.
3. **Frontmatter drift disappearing when `FilenameID` leaves domain value:**
   If a task is renamed on disk from `6g0000000001-task-alpha.md` to `6g0000000002-task-alpha.md`, while its YAML frontmatter still contains `id: 6g0000000001`. Under the proposed design, `RecordSource.ID` is `6g0000000002`, and `domain.Task.ID` remains `6g0000000001`. Lint compares `LoadedRecord.Source.ID` with `LoadedRecord.Value.ID` and detects drift. This works cleanly *provided* lint receives `LoadedRecord[Task]`.
4. **Unreadable record whose only honest repair handle is a local path:**
   A file `tasks/6g0000000001-corrupt.md` has corrupted YAML (unclosed quote on line 2). Frontmatter cannot parse; slug and ID cannot be read from YAML. However, the filesystem adapter extracts `6g0000000001` from the filename, reports `LoadProblem{EntityID: "6g0000000001", LocalPath: "/abs/path/to/corrupt.md"}`, and `tskflwctl task path 6g0000000001` succeeds via parse-free filename scanning. The user can open `$EDITOR` directly on `LocalPath` and repair it.
5. **Two remote records with the same ID/slug and no location:**
   A remote service returns two distinct JSON objects with `ID: "6g0000000001"`, `Slug: "task-alpha"`, and empty `Location`. An analytical pass can see that `len(records) == 2` with identical keys. However, any mutator attempting an update cannot address either record unambiguously. The system must fail closed with `ErrAmbiguous`.

#### 2.4 Private snapshot index lifecycle
A private snapshot index (e.g. tracking slice positions or source hashes to correlate diagnostics) is created inside the secondary adapter during the resilient read pass, survives only for the duration of the in-memory analysis/evaluation pipeline, and is erased when the snapshot is discarded or GC'd. It is never persisted, serialized, or leaked to public wire DTOs.

---

### 3. Campaign B — Four adapter compile probes

A comprehensive compile probe was implemented and verified in the sandbox (`design_probe_campaign_b_test.go`). It modeled all four adapter types using the exact proposed generic contracts:
- `RecordSource`
- `LoadedRecord[T]`
- `LoadProblem`
- `VersionedRecord[T]`
- `VersionedLoadProblem`

```go
type RecordSource struct { ID, Location string }
type LoadedRecord[T any] struct { Value T; Source RecordSource }
type LoadProblem struct { EntityKind, EntityID, EntitySlug, Location, LocalPath, Message string }
type VersionedRecord[T any] struct { Record LoadedRecord[T]; SourceVersion string }
type VersionedLoadProblem struct { Problem LoadProblem; SourceVersion string }
```

#### 3.1 Four-adapter comparison matrix

| Dimension | 1. Local Filesystem (`FS`) | 2. Database Rows (`SQL`) | 3. Remote Read-Only (`HTTP`) | 4. Cache / Projection (`MemCache`) |
|:---|:---|:---|:---|:---|
| **Authoritative Source ID** | Extracted from filename prefix (`splitFlatName`) | Primary key column (`id`) | JSON payload `id` field | Memory key / cached entity ID |
| **Location Format** | Canonical relative path (`tasks/6g...md`) | URI: `table://tasks?id=123` | URL: `https://api.example.com/v1/tasks/123` | Empty / cache key identifier |
| **Local Path Capability** | Implements `LocalTaskSource` (returns absolute path) | `ErrNoLocalPath` (unsupported) | `ErrNoLocalPath` (unsupported) | `ErrNoLocalPath` (unsupported) |
| **Revision Model** | Content SHA-256 byte hash | Row transaction ID (`xmin` / `version`) | Omitted (empty string `""`) | Omitted (empty string `""`) |
| **Counterfeited Values** | None; native fit | Synthesizes URI for `Location` | None; leaves `Location` as URL | Counterfeits empty `Location` and `SourceVersion` |
| **Unsupported Ops Expressible** | None | Launching local editor (`$EDITOR`) | All mutations (`SetFields`, `EditBody`) | All mutations; path commands |
| **Invalid Capability Pairings** | None | Pairing with filesystem `LocalTaskSource` | Pairing with guarded `TaskLifecycleMutationStore` | Pairing with guarded `TaskLifecycleMutationStore` |

#### 3.2 Compile probe results
- The probe compiled cleanly with zero errors under Go 1.24.
- **Probe finding 1:** When a read-only adapter returns `SourceVersion: ""`, Go type-checking succeeds, but `SameSourceSnapshot` fails closed at runtime when compared against any non-empty version. If both are empty, it vacuously succeeds unless explicitly checked.
- **Probe finding 2:** The Go compiler cannot enforce that an adapter implementing `TaskReadStore` also implements `LocalTaskSource`. CLI controllers calling `LocalTaskSource` must type-assert or inspect an explicit optional capability interface.

---

### 4. Campaign C — Domain-field removal blast radius

Mechanical searches and compiler impact analysis revealed the following blast radius across the repository for removing `CanonicalID()`, `SourceVersion`, and `Path` from `domain.Task`:

```
References across repository:
- CanonicalID():  148 call sites across internal/domain, internal/core, internal/store, internal/tui, internal/cli, internal/wire
- Path:           84 direct struct field references
- SourceVersion:  32 references across core, store, and thread
```

#### 4.1 Blast radius categorized by semantic failure mode

1. **Category 1: Identity & Keying at UI/List/Map Boundaries (TUI & CLI lists)**
   - *Locations:* `internal/tui/item.go`, `internal/tui/nav.go`, `internal/tui/commands.go`, `internal/cli/task_list.go`.
   - *Compiler failure:* `t.CanonicalID undefined (type domain.Task has no field or method CanonicalID)`.
   - *Impact:* TUI item list models currently store bare `domain.Task`. If `CanonicalID()` is removed, Bubble Tea selection keys, cursor restoration, and filter indexes fail to compile. TUI items must be refactored to hold `LoadedRecord[domain.Task]` or `tui.TaskItem` carrying `RecordSource`.

2. **Category 2: Serialization & Wire Projection (Loss of DTO identity & backward compatibility)**
   - *Locations:* `internal/wire/dto.go` (`ToTaskJSON`, `ToTasksEnvelope`, `ToTaskShowEnvelope`, `ToBoardEnvelope`, `ToSummaryJSON`).
   - *Compiler failure:* `t.Path undefined` and `t.CanonicalID undefined`.
   - *Impact:* All 7 wire converters take bare `domain.Task`. If `Path` is removed from `domain.Task`, `ToTaskJSON` cannot populate JSON `path`. If frontmatter `id:` is missing, `t.ID` is empty, emitting an invalid wire record without canonical identity.

3. **Category 3: CAS / Concurrency Control Boundary Mismatch**
   - *Locations:* `internal/core/task_graph.go`, `internal/core/service_thread.go`, `internal/store/task_mutation.go`.
   - *Compiler failure:* `t.SourceVersion undefined`.
   - *Impact:* Planner simulation, DAG cycle detection, and clone operations currently copy bare `domain.Task`. Moving `SourceVersion` into `VersionedRecord[T]` cleanly isolates revision tokens, but requires `TaskGraph` to store `VersionedRecord[domain.Task]` rather than bare `domain.Task`.

4. **Category 4: Filesystem I/O and External Editor Launch**
   - *Locations:* `internal/cli/task_edit.go`, `internal/tui/detail.go`, `internal/core/task_edit.go`.
   - *Compiler failure:* `cannot use t.Path as path in exec.Command`.
   - *Impact:* Editor launching requires a verified local filesystem path. Without `t.Path`, these consumers must explicitly call `LocalTaskSource.ResolveLocalTaskPath(id)`.

5. **Category 5: Graph / Dependency Resolution & Cycle Detection Keying**
   - *Locations:* `internal/core/graph.go`, `internal/domain/graph.go`.
   - *Compiler failure:* Map indexing `graph[t.CanonicalID()]` fails.
   - *Impact:* Graph construction requires unique canonical string keys. Must take `LoadedRecord[Task]` to index by `r.Source.ID`.

---

### 5. Campaign D — Revision and capability honesty

#### 5.1 Revision and capability truth table

| Entity State | Revision Token | Mutator Capability | `SameSourceSnapshot` Behavior | Lifecycle / Mutation Result | Enforcement Level |
|:---|:---|:---|:---|:---|:---|
| **Readable** | Present (Valid) | Guarded Local | Hash matches -> `true` | Mutation proceeds cleanly | Dynamically enforced |
| **Readable** | Present (Stale) | Guarded Local | Hash differs -> `false` | Fails with `ErrConcurrentModification` | Dynamically enforced |
| **Readable** | Empty / Omitted | Guarded Local | Hash check fails closed | **Fails: ErrMissingRevisionEvidence** | Dynamically enforced |
| **Readable** | Empty / Omitted | Read-Only Remote | Not invoked | Mutations statically rejected / disabled | Composition-root obligation |
| **Readable** | Present (Valid) | Read-Only Remote | Not invoked | Mutations statically rejected / disabled | Composition-root obligation |
| **Unreadable** | Present (Valid) | Guarded Local | Hash matches -> `true` | Repair write authorized | Dynamically enforced |
| **Unreadable** | Changed (Bytes) | Guarded Local | Hash differs -> `false` | **Fails: Concurrent modification during repair** | Dynamically enforced |
| **Unreadable** | Empty / Omitted | Guarded Local | Hash check fails closed | Fails closed; repair refused | Dynamically enforced |

#### 5.2 Challenge evaluations
- **Read-only adapter omits versions:** Legitimate for queries and DAG visualization. However, if composed with a guarded mutator, `SameSourceSnapshot` must fail closed with an explicit typed error (`ErrMissingRevisionEvidence`) rather than comparing empty strings.
- **Location changes while revision does not:** In a filesystem move/rename operation, if file bytes are identical, byte-hash revision remains identical while location changes. Guarded snapshots must check both `SourceVersion` and location continuity to prevent writing to a relocated destination.
- **Unreadable record changes bytes while retaining message:** A syntax error remains "yaml parse error", but file bytes changed concurrently. The content-hash revision changes, correctly aborting a stale automated repair.

---

### 6. Campaign E — Local-path and mutation UX

#### 6.1 Code tracing across local operations
- **CLI Commands (`path`, `info`):** In `internal/cli/task.go:newTaskPathCmd`, `app.Svc.TaskPath(slug)` delegates to `s.store.ResolveTaskPath(slug)`. If `store` does not implement `LocalTaskSource`, `TaskPath` returns `domain.ErrNotSupported`. Plain output fails with exit code 1; JSON envelope emits `{"schema_version":"1.75","error":"local path not supported"}`.
- **Creation & Dry-Run:**
  Running `tskflwctl task new --dry-run --no-input --epic 21-code-quality-architecture-hardening --tags refactor "Test Dry Run Task" --json` yields:
  ```json
  {"schema_version":"1.75","dry_run":true,"created":{"kind":"task","id":"6gdx5d1147gx","slug":"test-dry-run-task","status":"ready-to-start","path":"tasks/6gdx5d1147gx-test-dry-run-task.md"}}
  ```
  Exit classification: 0 (Success).
  As demonstrated in Finding L2, `TaskCreationReceipt` must carry `PlannedLocalPath` so that dry-run machine output is preserved when domain `Path` is removed.
- **Rename & Move Cleanup Failures:** `RenameTask` in `internal/core/task_rename.go` executes two-phase updates: writing the new file, updating inbound links, and deleting the old file. If deleting the old file fails, `TaskRenameMutationResult` retains both old and new paths in durable metadata so recovery is deterministic.

#### 6.2 Pathless TUI discoverability
In `internal/tui/detail.go`, pressing `e` (edit in `$EDITOR`) or `y` (yank path to clipboard) currently reads `t.Path`. In a pathless workspace:
- The UI status bar renders `[e] Edit (disabled: pathless source)`.
- Pressing `e` displays a transient warning toast: `"Editing is unavailable: current adapter provides no local filesystem path"`.
- Semantic navigation (status change, tag filtering, detail inspection) continues to function seamlessly.

---

### 7. Campaign F — Wire and migration compatibility

#### 7.1 Envelopes and JSON schema assessment
- **`LintLoadProblemJSON`:** Retaining the wire DTO name `LintLoadProblemJSON` while renaming the core domain type to `LoadProblem` is fully compatible. The wire envelope does not expose the Go type name; it only defines JSON properties.
- **`path` Property across Schemas:** In generated JSON schemas (`schema_jsonschema.golden`), `path` is currently defined as a required string on `TaskInfoJSON` and `PathEnvelope`. For pathless sources, `path` must be emitted as `""` (empty string) rather than omitted or `null`, preserving strict client schema validation.
- **`LocationIsPath` Removal:** `LocationIsPath` is eliminated in favor of explicit `Location` (opaque source string) and `LocalPath` (absolute filesystem path).

#### 7.2 Stage-by-stage migration evaluation

The proposed 4-stage migration was evaluated for intermediate compile and runtime viability:

| Stage | Intended Scope | Intermediate State Viability | Verdict & Identified Blocker |
|:---|:---|:---|:---|
| **Stage 1** | Promote `LoadProblem` & typed read snapshots | **Viable** with amendment | Wire DTO converters must be updated in this stage to take `LoadedRecord[T]` to avoid cyclic breakage in Stage 2. |
| **Stage 2** | Move source metadata out of domain records | **Blocked if isolated** (Finding M2) | Cannot remove `Path` and `FilenameID` from `domain.Task` without simultaneously refactoring all 7 wire DTO callers and TUI item lists. |
| **Stage 3** | Split and wire local capabilities | **Viable** | Isolates `Resolve*Path` to `Local*Source` interfaces. Existing CLI path commands adapt cleanly. |
| **Stage 4** | Close primary-adapter bypasses & remove debt | **Viable** | Replaces CLI filesystem imports with application ports and removes legacy `FileProblem`. |

**Safe Dependency Sequence Amendment:**
To eliminate intermediate broken states, the migration must sequence as:
- **Phase 1 (Stage 1 amended):** Promote `LoadProblem`, introduce `RecordSource` and `LoadedRecord[T]`, update read store methods to return snapshots, and adapt wire converters / TUI item models to accept `LoadedRecord[T]`.
- **Phase 2 (Stage 2 amended):** Remove `Path`, `FilenameID`, and `SourceVersion` from domain records (safe now that consumers use `LoadedRecord[T]`).
- **Phase 3 (Stage 3):** Split `Resolve*Path` into optional local capabilities and update CLI/TUI editor actions.
- **Phase 4 (Stage 4):** Route CLI repair/completion through application ports and delete dead compatibility code.

---

### 8. Antigravity falsification ledger

| # | Claim under attack | Concrete counterexample attempted | Evidence / command | Result | Design disposition |
|:---:|:---|:---|:---|:---|:---|
| 1 | Generic value types reduce rather than spread coupling | Pair read-only reader with guarded filesystem mutator | Compile probe: `design_probe_campaign_b_test.go` | **Falsified**: Go compiler accepts invalid pairing without static error; fails cryptically at runtime | **Amended (L1)**: Add explicit runtime assertion in Service constructors |
| 2 | Canonical source ID can replace every `CanonicalID()` consumer | Remove `CanonicalID()` from `domain.Task` while wire/TUI still take bare domain | Compile probe: `internal/wire/dto.go` patch test | **Falsified**: 148 call sites break across TUI and wire projections | **Amended (M2)**: Adapt wire/TUI to `LoadedRecord[T]` prior to domain stripping |
| 3 | Identityless readable records can be rejected without hiding lint evidence | Malformed filename `legacy-notes.md` with valid YAML frontmatter payload | Trace `internal/domain/lint_task.go` vs `LoadProblem` demotion | **Falsified**: Dropping unkeyed records to `LoadProblem` blinds `LintTask` schema validation | **Amended (M1)**: Add unkeyed record inspection path for lint snapshots |
| 4 | Duplicate IDs remain attributable but unselectable | Lookup record by slug when two records share canonical ID `6g0000000001` | Probe against `internal/store/resolve.go:resolveID` | **Falsified**: Slug lookup bypasses canonical ID uniqueness check and selects duplicate record | **Amended (H1)**: Slug resolution must verify canonical ID uniqueness in snapshot |
| 5 | Explicit local path and opaque location never collapse accidentally | Filesystem adapter returns relative location vs absolute local path | `tskflwctl task path 6gcwcf7rgxef` vs `task show --json` | **Confirmed**: `Location` is `tasks/6g...md`; `LocalPath` is `/abs/tasks/6g...md`. Distinct semantics preserved | Design upheld |
| 6 | Source revisions cannot leak and cannot be dropped before guarded comparison | Trace `TaskGraphRead` through planner simulation and public wire projection | Traced `internal/core/task_graph.go` and `internal/wire/dto.go` | **Confirmed**: Wrapper drops `SourceVersion` before planner/wire projection; retained in store CAS | Design upheld |
| 7 | Read-only sources cannot silently authorize writes | Execute guarded mutation using read-only adapter returning empty `SourceVersion` | Compile probe `design_probe_campaign_b_test.go` | **Confirmed**: Hash comparison fails closed against non-empty token | Design upheld |
| 8 | Thread cleanup does not invalidate its published preview compatibility contract | Run thread compatibility golden tests against adapter-neutral read | `go test ./internal/cli -run TestThreadCompatibility` | **Confirmed**: Wire JSON contracts match historical v0.18/v0.19 fixtures | Design upheld |
| 9 | Mutation receipts cover dry-run and committed-prefix local metadata | Execute `tskflwctl task new --dry-run --json` with domain `Path` stripped | `go run ./cmd/tskflwctl task new --dry-run ... --json` | **Falsified**: CLI dry-run requires `PlannedLocalPath` on `TaskCreationReceipt` | **Amended (L2)**: Add `PlannedLocalPath` to `TaskCreationReceipt` |
| 10 | Ordinary read snapshots preserve single-scan behavior | Compare `ListTasks` followed by `GetTask` against `ListTasksWithBodies` | Trace `internal/store/task_resilient.go:ListTasksWithBodies` | **Confirmed**: Resilient scan caches YAML bodies during initial read without rescanning | Design upheld |
| 11 | Core type renaming does not break machine-schema compatibility | Retain `LintLoadProblemJSON` DTO name while renaming core `LoadProblem` | Schema diff against `schema_jsonschema.golden` | **Confirmed**: JSON schema `$defs` key remains identical; core renaming is zero wire impact | Design upheld |
| 12 | The proposed task sequence has safe, reviewable intermediate states | Land Slice 2 (domain stripping) before Slice 3/4 (wire/TUI refactor) | Staged dependency analysis across 4 tasks | **Falsified**: Slice 2 cannot compile in isolation without Slice 3/4 wire adaptation | **Amended (M2)**: Re-sequence wire/TUI consumption into Slice 1 |

---

### 9. Exact amendments to design and downstream tasks

#### 9.1 Amendments to design task `6gcwcf7rgxef`
- **Location:** `planning/tasks/6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md`
- **Under `### Duplicate and ambiguous identity` (line 170):**
  *Add:* "Resolution by slug, prefix, or title must verify that the resolved entity's canonical `RecordSource.ID` does not collide with any other record in the loaded snapshot. If a collision exists, resolution must return `domain.ErrAmbiguous` regardless of query mode."
- **Under `### Target read shapes` (line 210):**
  *Add:* "For lint snapshots, readable files lacking a valid source ID must still expose their parsed domain payload to `domain.LintTask` (e.g. via an unkeyed record variant or payload-bearing `LoadProblem`) to prevent hiding schema and dependency defects."
- **Under `### Mutation results` (line 180):**
  *Add:* "`TaskCreationReceipt` must explicitly include `PlannedLocalPath string` (empty for pathless adapters) so CLI dry-run and JSON envelopes can report destination paths without inspecting domain entities."

#### 9.2 Amendments to downstream task `6gcwcf80v8hg`
- **Location:** `planning/tasks/6gcwcf80v8hg-make-ordinary-entity-list-diagnostics-adapter-neutral.md`
- **Under `## Scope`:**
  *Add:* "Update internal wire DTO converters (`internal/wire/dto.go`) to accept `LoadedRecord[T]` envelopes alongside bare domain values, preparing wire serialization for subsequent domain field removal."

#### 9.3 Amendments to downstream task `6gcwcf88z57p`
- **Location:** `planning/tasks/6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md`
- **Under `## Scope`:**
  *Add:* "Refactor TUI item list models (`internal/tui/item.go`, `internal/tui/nav.go`) to store `LoadedRecord[domain.Task]` or `tui.TaskItem` carrying `RecordSource`, decoupling TUI selection from domain `CanonicalID()` and `Path`."

#### 9.4 Amendments to downstream task `6gcwcf8gzn50`
- **Location:** `planning/tasks/6gcwcf8gzn50-route-cli-planning-data-operations-through-application-ports.md`
- **Under `## Scope`:**
  *Add:* "Introduce runtime capability validation assertions in Service constructors to detect and reject incompatible pairings (e.g., read-only readers composed with guarded mutation stores)."

#### 9.5 Amendments to Thread `6gcwd78p9r04`
- **Location:** `planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md`
- **Under `## Context` (Stage 4):**
  *Update:* Explicitly state that wire DTO and TUI item models migrate to `LoadedRecord[T]` consumption in Stage 3, enabling clean, non-cyclic domain struct stripping in Stage 4.

---

### 10. Smallest unresolved user decision

**Wire schema representation of `path` when `LocalPath` is empty/absent:**
When an entity is served by a pathless adapter (e.g. remote HTTP service or database row), what should the JSON envelope emit for `path` in `TaskInfoJSON` and `PathEnvelope`?
- **Option 1 (Recommended):** Emit `path: ""` (empty string). Maintains 100% backward compatibility with existing strict JSON Schema definitions without requiring client schema changes.
- **Option 2:** Omit the field (`path,omitempty`). Cleaner for pure pathless consumers, but requires updating JSON Schema definitions and incrementing minor wire version.
- **Option 3:** Emit `path: null`. Explicitly communicates non-applicability, but breaks clients expecting string values.

---

### 11. Validation commands and complete isolation attestation

#### 11.1 Validation commands executed in sandbox
```sh
# Verify sandbox clean baseline and deliverable status
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"

# Verify CLI dry-run and JSON envelope contracts
go run ./cmd/tskflwctl task new --dry-run --no-input --epic 21-code-quality-architecture-hardening --tags refactor "Test Dry Run Task" --json

# Verify local task path resolution
go run ./cmd/tskflwctl task path 6gcwcf7rgxef

# Verify git diff contains only the assigned deliverable
git diff --name-only
```

#### 11.2 Mandatory reviewer isolation attestation
```
isolated-review: sandbox initialized at /private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Wn9UQH
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Wn9UQH
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Wn9UQH/.git
baseline_commit=6050f29cdc92119974fc5bc89627b6c0b5023fe0
source_blob=2c9a6a7efd7f41132ed84380c486af06475f3925
source_fingerprint=781792c97f14cafc56cd4b04327214200aa6e60f
deliverable=planning/audits/6gdx1bvz683y-2026-09-26-portable-entity-read-local-source-capability-design-antigravity.md
deliverable_changed=true
transfer=pending
```
