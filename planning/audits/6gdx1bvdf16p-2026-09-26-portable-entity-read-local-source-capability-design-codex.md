---
schema: 1
id: 6gdx1bvdf16p
bucket: closed
area: portable-entity-read-local-source-capability-design-codex
date: "2026-09-26"
updated_at: "2026-09-26"
---
# Audit: Portable entity read and local source capability design — Codex — 2026-09-26

> Reviewer assignment: Codex. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; do not use a free-standing status line. Preserve this brief and add the report below it.

> This is a design review, not an implementation review. Treat every checked acceptance criterion,
> type sketch, migration claim, and downstream scope as a hypothesis. Do not report merely that the
> implementation has not happened yet. Report a finding when the proposed boundary is internally
> unsound, contradicts shipped behavior, cannot be migrated as sequenced, or omits a load-bearing
> state that the current code demonstrates.

> Shared-worktree isolation is mandatory. Treat the handoff checkout as a read-only source. Create
> an independent clone with the repository helper before inspecting code, running commands, or
> making disposable probes. Do not use `git worktree`, symlinks, or shared Git metadata.

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

The helper captures committed, staged, unstaged, deleted, and untracked handoff state in an
independent `--no-hardlinks` clone and creates a sandbox-only baseline commit. Perform every read,
test, scratch compile, and report edit there. Before transfer, restore every probe so the assigned
audit is the only diff, then run:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

Include the workspace path, resolved Git directory, baseline commit, captured source fingerprint,
deliverable, and transfer result in the report. If isolation or transfer fails, preserve the
sandbox and report the blocker; never fall back to editing the shared checkout.

## Review target

Perform an independent adversarial architecture review of task `6gcwcf7rgxef`,
`design-portable-entity-reads-and-optional-local-source-capabilities`, on branch
`design/portable-entity-read-contracts` relative to base
`7d1e27370745707e79ae3c95d81831246f98becd`.

Review the complete captured planning delta, especially:

- `planning/tasks/6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md`;
- the amended ordinary-read, source/path, CLI-routing, and composition tasks;
- `planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md`;
- current `internal/core`, `internal/domain`, `internal/store`, `internal/cli`, `internal/tui`, and
  `internal/wire` consumers that the design claims to migrate;
- the completed Thread path/read, Thread diagnostic, repository lint, graph attribution, audit
  snapshot, and Board/status planning precedents.

Do not edit implementation or planning files other than this assigned audit.

## Design claims to challenge

1. A small shared `RecordSource`, `LoadedRecord[T]`, `LoadProblem`, `VersionedRecord[T]`, and
   `VersionedLoadProblem` vocabulary can support the real consumers while ports remain
   entity-specific and use-case-owned.
2. `RecordSource.ID` can replace every `FilenameID`/`CanonicalID()` dependency without losing
   frontmatter-drift lint, duplicate-record attribution, mutation authorization, TUI stable
   navigation, or repairability.
3. A readable record must have canonical source identity; an identityless source becomes a load
   problem rather than a half-addressable record. Verify that this matches every entity's tolerant
   parsing and lint policy.
4. Opaque `Location` is useful context but never identity. `LocalPath` is a separately optional
   repair handle. `LocationIsPath` can be retired without creating precedence ambiguity or losing
   compatibility.
5. Ordinary reads need no revision. Task/Thread guarded snapshots retain complete readable and
   unreadable revision evidence outside domain values; read-only adapters may omit it but cannot be
   paired dishonestly with guarded mutation.
6. The current Thread precedent is incomplete because its domain values still carry `Path`,
   `FilenameID`, and `SourceVersion`; finishing that separation is safer than preserving the
   optional fields.
7. Path commands and TUI editor/copy actions can depend on independent entity-specific local-path
   sources while semantic list/show/navigation remains useful when those capabilities are absent.
8. Mutation-returned paths can move to explicit entity-specific receipts or post-success path
   resolution without weakening dry-run output, committed-prefix recovery, or local UX.
9. Promoting `LintLoadProblem` to a general core diagnostic while retaining historical wire fields
   and possibly the `LintLoadProblemJSON` schema definition name is coherent and compatibility-safe.
10. The amended sequence—ordinary read snapshots, then source/local metadata separation, then CLI
    routing, then import enforcement—is technically feasible and does not create an unusable
    half-migrated state.

## Mandatory evidence

### 1. Crossing and consumer inventory

Independently enumerate every current crossing of:

- `domain.FileProblem` through core-facing ports and service results;
- `Path`, `FilenameID`, `CanonicalID()`, `SourceVersion`, and `LocationIsPath` on task, epic, audit,
  research, and Thread flows;
- `ResolveTaskPath`, `ResolveEpicPath`, `ResolveAuditPath`, `ResolveResearchPath`, and
  `ResolveThreadPath`;
- mutation receipts or CLI/TUI paths that obtain local source information indirectly from a domain
  entity.

For each, classify the actual semantic need: canonical identity, declared document value, opaque
source context, local repair/navigation, optimistic-concurrency evidence, presentation-only data,
or accidental leakage. Compare your table to the design's inventory and call out omissions.

### 2. Target type and operation matrices

Produce two matrices:

- For filesystem, database, remote service, and read-only cache adapters, state which fields and
  capabilities must, may, and must not be supplied for list, show, lint, graph query, guarded
  mutation, path, editor, create dry-run, and malformed repair.
- For task, epic, audit, research, and Thread, state where declared ID, canonical source ID, slug,
  location, local path, and source revision live before and after the migration.

Use those matrices to test whether the proposed shared values are minimal and sufficient. Watch for
generic wrappers that merely relocate storage coupling or entity-specific facts that the generic
shape cannot express.

### 3. Identity and malformed-state analysis

Trace these states through list, lint, show/resolution, TUI identity, and mutation authorization:

- valid source ID matching declared ID;
- valid source ID with missing or drifting declared ID;
- duplicate source IDs with distinct slugs and/or locations;
- duplicate source IDs without useful locations;
- missing source ID with otherwise parseable content;
- unreadable content with ID only, slug only, both, or neither;
- explicit identity plus a contradictory path-like or opaque location.

Determine whether slice position/private snapshot references are sufficient for analysis and where
durable selection must instead fail ambiguous. Core must never recover identity from location.

### 4. Snapshot and concurrency analysis

Reconstruct task-graph and Thread whole-snapshot comparison for readable and unreadable records.
Show exactly which code currently depends on domain-carried revisions and how a versioned wrapper
would preserve sorting, cloning, equality, planner stripping, simulation, repair, and CAS. Challenge
the statement that epic/audit/research ordinary mutations need no exposed revision with current
read-modify-write and retry behavior.

### 5. Local capability and mutation-output analysis

Trace every CLI/TUI read, copy, editor, create, rename, move, lifecycle, and partial-durability path
that currently reads `entity.Path`. For each, decide whether lazy path resolution is safe, whether a
receipt needs optional local outcome metadata, or whether the design has missed a race/UX contract.
Pay special attention to dry-run paths, malformed records, create results not yet resolvable,
rename's old/destination state, and split read/path sources from different corpora.

### 6. Compatibility and migration feasibility

Inspect wire DTOs, generated JSON Schema, schema comments, human renderers, exit code 11 policy,
fakes, and Service construction. Determine:

- whether renaming only the core diagnostic is truly wire-neutral;
- whether `$defs` names, required `path`, empty path, location, and local-path coexistence are
  correctly treated;
- whether each proposed stage can compile and preserve one authoritative scan without temporary
  duplicate reads or hidden aggregate fallback;
- whether the downstream task scopes and dependency edges are small enough to review and land
  safely, or require a different split.

## Required adversarial angles

- Look for a false dichotomy between semantic data and adapter metadata: some derived evidence may
  be required by domain validation even if it comes from an adapter.
- Assume `LoadedRecord[T]` spreads mechanically through the codebase and makes application logic
  harder to use. Identify where explicit entity records or use-case projections would be safer.
- Assume removing domain paths breaks more than navigation—diagnostic attribution, duplicate
  explanation, sorting, guarded plans, and mutation receipts all use them today.
- Assume the Thread precedent's option-order/detach rules do not scale cleanly to five independent
  read/path capabilities. Probe configuration ambiguity and typed-nil behavior.
- Assume a read-only source is accidentally paired with a mutation adapter from another corpus.
  Decide whether the proposed ports can detect or merely document that incompatibility.
- Distinguish honest deferred implementation from a missing design decision. Do not demand a
  database implementation, generic repository, or storage-format change.

## Required result

For each of the ten design claims, record `accepted`, `accepted with amendment`, or `rejected`, with
exact code evidence. Provide:

- the independent crossing inventory;
- the adapter/operation and entity/evidence matrices;
- a migration-risk table with compile blast radius and safe intermediate states;
- concrete amendments to the task or downstream scopes;
- the smallest user decision needed for any genuinely unresolved product tradeoff.

A no-findings verdict is valid only if all required matrices and hostile states are settled with
evidence. Green tests or a prose summary are insufficient.

## Findings

#### H1. Explicit read, path, and mutation capabilities have no enforceable common-corpus identity · **Status:** tracked by 6gdx7mcqm371

The proposal correctly detaches an implicit `ThreadPathSource` after an explicit `ThreadStore` is
installed, but the explicit override remains unconstrained: `WithThreadPathSource` accepts any
resolver and `WithThreadMutation` accepts any writer without proving either belongs to the same
corpus as the semantic read source (`internal/core/service.go:164-223`). This limitation is already
documented in the workspace composition code—its narrow ports carry no shared backend identity, so
the composition root must keep them aligned—but `Workspace.Open` validates only layout/root/checkout
and then wires the independent capabilities (`internal/core/workspace.go:17-28,90-112`). The TUI then
combines `ThreadPath(id)` with `ShowThreadGraphDetail(id)` as if the results necessarily describe the
same record (`internal/tui/thread_projection.go:55-75`). A read-only source accidentally paired with
a path or mutation adapter from another corpus can therefore display, copy, edit, authorize, or
mutate a different record that happens to have the same canonical ID. Option ordering and comments
can prevent accidental inheritance, but cannot detect this failure.

Amend the design to introduce an opaque comparable `SourceSetID`/corpus token on every semantic
read, local-path, and mutation capability (or construct them only through a bundle which owns such a
token). Service construction must reject incompatible explicit pairs, including read/mutation pairs,
while preserving the current implicit detach rule. A record's `Source.ID` is entity identity and
must not double as corpus identity. The downstream split-source acceptance criterion already says an
unrelated resolver cannot be silently paired; the proposed interfaces do not yet make that criterion
implementable.

**Resolution:** Accepted. The design now distinguishes corpus identity from
record identity and revision evidence; task 6gdx7mcqm371 owns constructor-time
source-set compatibility across read, path, and mutation capabilities.

#### M1. Duplicate canonical IDs cannot be both retained and addressed by an ID-only TUI key · **Status:** fixed

The design requires ordinary reads to retain both records for a duplicate ID and requires the TUI to
list every returned record, while also declaring canonical ID to be the only durable selection and
restoration key. The current TUI makes `entityRef.key()` equal to canonical ID and rejects an entire
list as invalid when two rows have the same key (`internal/tui/entity.go:46-93`); Thread list rows set
that key from `CanonicalID()` (`internal/tui/thread_projection.go:30-43`). Thus the proposed contract
has no way to render or select the two rows it promises to retain. Using location or slice position as
public identity would violate the design in the other direction.

Add a snapshot-local, opaque row handle for UI addressing. It may encode an index internally but is
valid only for the exact loaded snapshot. Durable/session restoration remains canonical-ID based and
must clear or prompt when that ID is now ambiguous. Any show, path, or mutation action addressed only
by canonical ID must fail ambiguous before touching a local capability; an action from a row handle
must resolve the selected occurrence against the still-current snapshot. Location remains evidence,
never identity.

**Resolution:** Resolved with a stricter variant. Analytical CLI/list/lint
projections retain duplicate occurrences, while the TUI registry continues to
reject or degrade missing/duplicate canonical keys. Snapshot position and
location remain non-actionable; no row handle is introduced.

#### M2. Mutation outcomes do not preserve enough local evidence for dry runs and partial durability · **Status:** tracked by 6gdx7mcqq67d

The proposed optional local receipt is not specified per operation. Today create commands print the
new domain-carried path for task, epic, audit, research, and Thread results
(`internal/cli/task.go:167-175`, `internal/cli/epic.go:231-239`,
`internal/cli/audit.go:70-78`, `internal/cli/research.go:215-224`,
`internal/cli/thread.go:163-177`). A dry-run create has no committed record for a later resolver to
find. Rename is more serious: the store can durably write the destination and then fail to remove the
source (`internal/store/rename.go:138-165`), and its recovery text tells the user to inspect the old
and destination files (`internal/core/task_rename.go:81-88`). The current receipt exposes flags and
slugs but not both exact local paths (`internal/core/task_rename.go:12-43`,
`internal/wire/task_rename.go:8-22`); removing `Task.Path` erases the remaining indirect destination
evidence.

Define a typed `LocalOutcome` for each operation before removing domain paths: create needs the
planned and, when committed, resulting path; rename/move needs source and destination paths plus the
durability state; Thread and lifecycle creation need the created/planned path. Partial-error receipts
must survive wire and human rendering. Post-success lazy resolution is safe only for a committed,
non-moving, single-record operation whose resolver is bound to the same corpus; it is not a substitute
for dry-run or recovery evidence.

**Resolution:** Accepted. Task 6gdx7mcqq67d now owns planned/resulting create
paths and exact rename source/destination plus durable-stage evidence before
domain paths are removed.

#### M3. The staged migration has no specified compilable, one-scan intermediate for shows, lint, and lazy TUI actions · **Status:** tracked by 6gdx7mcrq8s8

The ordinary-list task says it will introduce shared loaded snapshots, yet explicitly leaves removal
of `Path`, `FilenameID`, and `SourceVersion` to the later split-source task. The design task at the
same time treats versioned task-graph reads as an already settled target. Currently task list, board,
and summary all project one `TaskGraphRead` (`internal/core/service_task.go:48-53`,
`internal/core/board.go:43-57`, `internal/core/service.go:364-371`); adding a second ordinary
`TaskRead` without an explicit projection seam risks a duplicate scan, while changing
`TaskGraphRead` early consumes the later task's scope. Single-document ports also still return bare
domain values (`internal/core/store.go:15-22,185-188,237-245,274-280`), so removing
`FilenameID`/`Path` before changing `Get*`/`Show*` shapes loses source identity and diagnostic
evidence outside list commands.

The TUI part is likewise underspecified. Editor and copy actions synchronously read the selected
item's path (`internal/tui/model.go:1447-1516`), while Thread eagerly resolves its path when loading
detail (`internal/tui/thread_projection.go:55-75`). A new lazy/async resolver must capture the list
generation, source ID, and snapshot-local row handle, and revalidate all three before the side effect;
otherwise a response arriving after refresh can act on a different selection.

Amend the downstream scopes and dependencies with one authoritative read implementation that first
returns an internal loaded/versioned record and supplies compatibility projections for existing graph,
list, and show consumers. Move `Get*`/`Show*` port signature changes, lint evidence projection, and
TUI generation-aware resolution into explicit owned stages. Only after all consumers stop reading
the domain fields should the split-source task remove them.

**Resolution:** Accepted. The ordinary-read task now owns one authoritative
loaded scan plus list/show/lint/wire compatibility projections; task
6gdx7mcrq8s8 owns TUI identity and generation-aware refresh migration before
final field removal.

## Reviewer report

### Verdict

**Rework required.** The separation of semantic records, source evidence, versions, and local
capabilities is directionally sound, and the Thread/task-graph precedents support most of the proposed
vocabulary. The design is not implementation-ready because it cannot enforce that independently
configured capabilities refer to one corpus, cannot satisfy its duplicate-ID TUI promises, loses
required recovery evidence in several mutation outcomes, and does not define a safe staged ownership
boundary for ordinary reads, shows, lint, and TUI navigation.

### Ten design claims

| # | Claim | Decision | Evidence and required amendment |
|---|---|---|---|
| 1 | One shared read vocabulary should replace entity-specific transport metadata. | **accepted with amendment** | `TaskGraphRead` and Thread already need source/version evidence (`internal/core/dependency_graph.go:322-330`; `internal/core/store.go:143-162`). Keep the generic vocabulary minimal, but use explicit use-case projections for graph, lint, and UI rather than mechanically spreading `LoadedRecord[T]` through application code. Add corpus identity to capability composition, not to every entity value. |
| 2 | `RecordSource.ID` can replace `FilenameID` and other adapter-derived identity. | **rejected** | Parsers currently retain declared and filename identity separately (`internal/store/fsstore.go:330-369`; `internal/store/researchstore.go:71-97`), and lint consumes the distinction through `IDDriftIssue` (`internal/domain/lint.go:140-180`). The replacement works only after lint receives explicit declared/source identity and the TUI has a snapshot-local duplicate-row handle (M1). |
| 3 | Every readable record has a non-empty source ID; malformed records may lack one. | **accepted** | ID-led recovery obtains an ID/slug without parsing content, while non-ID-led malformed files can remain identityless (`internal/store/resolve.go:36-80`, `internal/store/flatname.go:10-40`). Epic parsing derives its semantic ID from its source filename (`internal/store/epicstore.go:257-273`). Operations must fail identity-required actions when `Source.ID` is absent. |
| 4 | `Location` and optional `LocalPath` should replace `LocationIsPath`. | **accepted with amendment** | The separation is honest and preserves non-filesystem adapters. During compatibility conversion, old `LocationIsPath=true` must populate both `Location` and `LocalPath`; contradictory explicit values must be retained for diagnostics, not normalized into identity. Empty `LocalPath` disables local operations. |
| 5 | Ordinary epic/audit/research mutations need no public revision token. | **accepted with amendment** | Current stores re-read, compare, and retry internally (for example research update at `internal/store/researchstore.go:120-177` and audit move at `internal/store/auditstore.go:61-134`). That remains valid only if read and mutation capabilities are provably same-corpus and guarded/batch APIs continue to expose their own snapshot contract. Do not infer authorization from a caller-supplied local path. |
| 6 | Task graph and Thread whole-snapshot operations keep explicit versions. | **accepted** | Task equality requires complete source/version agreement (`internal/core/dependency_graph.go:934-971`); Thread CAS compares readable and unreadable source evidence and non-empty versions (`internal/store/cas.go:31-98`). A versioned loaded wrapper preserves those semantics without keeping revisions in domain values. |
| 7 | Local path capabilities can be optional, lazily resolved, and keyed only by canonical ID. | **rejected** | Parse-free resolvers exist (`internal/store/paths.go:3-46`), but ID-only lookup is ambiguous for retained duplicates and independent capability pairs can address different corpora (H1, M1). Lazy TUI responses also require snapshot-generation revalidation (M3). |
| 8 | Mutation receipts need local paths only where commands already promise them. | **rejected** | Create dry runs and rename partial durability require planned/source/destination paths even when no resolvable committed record exists (`internal/store/rename.go:138-165`; current create renderers cited in M2). Specify operation-specific local outcomes before removing domain paths. |
| 9 | Promoting `FileProblem` to `LoadProblem` can be wire-neutral. | **accepted with amendment** | A core rename is neutral only while DTOs remain unchanged. Ordinary envelopes currently expose `domain.FileProblem` directly (`internal/wire/envelopes.go:35,725,759,788`), whereas `LintLoadProblemJSON` has a distinct required `path` contract (`internal/wire/envelopes.go:938-968`; `internal/wire/testdata/schema_jsonschema.golden:1894-1921`). Any `$defs`, required-field, or property-shape change requires the repository's schema-version process; `location`/`local_path` should be additive before legacy `path` is retired. |
| 10 | The proposed downstream sequence is independently landable and keeps one authoritative scan. | **rejected** | Existing list, board, and summary projections share `TaskGraphRead`; bare `Get*` ports still serve show/lint consumers. The current task split neither owns all required signature changes nor defines compatibility projections, so the advertised sequence can duplicate scans or strand evidence (M3). |

### Independent crossing inventory

This inventory was reconstructed from code rather than copied from the design task.

| Crossing | Task | Epic | Audit | Research | Thread |
|---|---|---|---|---|---|
| Domain transport fields | `Path`, `SourceVersion`, `FilenameID`, `CanonicalID` in `internal/domain/task.go:7-30,64-73` | `ID`, `Path`, `CanonicalID` in `internal/domain/epic.go:25-42` | `Path`, `ID`, `CanonicalID` in `internal/domain/audit.go:41-56,80-88` | `Path`, `ID`, `FilenameID`, `CanonicalID` in `internal/domain/research.go:31-41,58-66` | `Path`, `SourceVersion`, `CanonicalID` in `internal/domain/thread.go:51-79` |
| Store/list/show port | `ListTasks`/`GetTask` return entity plus `[]FileProblem` (`internal/core/store.go:15-22`) | `ListEpics`/`GetEpic` (`internal/core/store.go:185-188`) | `ListAudits`/`GetAudit` (`internal/core/store.go:237-245`) | `ListResearch`/`GetResearch` (`internal/core/store.go:274-280`) | `ThreadStore` already returns `[]VersionedThread` plus problems; path is a separate port (`internal/core/store.go:143-162`) |
| Parser/scan evidence | Path, hash revision, filename identity assigned at parse (`internal/store/fsstore.go:330-369`) | Epic-specific recovery and filename-derived ID (`internal/store/epicstore.go:16-43,257-273`) | Parser assigns path and source IDs (`internal/store/auditstore.go:188-215`) | Parser assigns path and filename identity (`internal/store/researchstore.go:71-97`) | Parser assigns path and hash revision (`internal/store/threadstore.go:95-119`) |
| Graph/aggregate crossing | `NewTaskGraph` and compatibility projections carry problems and versioned source evidence (`internal/core/dependency_graph.go:322-425`) | Summary and board consume shared service state | Summary consumes audit problems | Summary consumes research problems | Thread snapshot/CAS consumes versioned readable and unreadable records (`internal/store/cas.go:31-98`) |
| Lint/diagnostic crossing | Declared versus filename identity feeds ID-drift findings (`internal/domain/lint.go:140-180`) | Derived source identity must remain attributable | Core finding logic reports audit identity drift (`internal/core/finding.go:250-265`) | Research lint uses filename identity (`internal/domain/lint.go:200-204`) | Service lint reports IDs, paths, and source failures (`internal/core/service.go:500-717`) |
| Path resolution | `Service.TaskPath` delegates to resolver (`internal/core/service_task.go:253-257`) | `Service.EpicPath` delegates to resolver (`internal/core/service_epic.go:384-388`) | `Service.AuditPath` delegates to resolver (`internal/core/service_audit.go:99-103`) | Dedicated resolver/service path crossing | `Service.ThreadPath` delegates to independently configured source (`internal/core/service_thread.go:386-390`) |
| TUI crossing | List refs and details read canonical ID/path (`internal/tui/item.go:50-60`; `internal/tui/detail.go:785`) | Item/detail path (`internal/tui/item.go:112-122`; `internal/tui/detail.go:855`) | Detail path (`internal/tui/detail.go:1693`) | Detail path (`internal/tui/detail.go:1739`) | Projection combines graph detail and eager path (`internal/tui/thread_projection.go:30-75`; `internal/tui/detail.go:967`) |
| CLI/wire/receipt crossing | Create prints path; rename wire receipt lacks old/destination paths (`internal/cli/task.go:167-175`; `internal/wire/task_rename.go:8-22`) | Create prints path (`internal/cli/epic.go:231-239`) | Create prints path (`internal/cli/audit.go:70-78`) | Create prints path (`internal/cli/research.go:215-224`) | Update/create output carries path (`internal/cli/thread.go:123-177`) |

`FileProblem` also crosses ordinary envelopes directly and is rendered under exit code 11. The legacy
renderer assumes a path/base (`internal/cli/problems.go:18-43`); the portable renderer can already
use identity, path, or opaque location (`internal/cli/problems.go:74-117`). Exit classification is
centralized as `ExitProblems=11` (`internal/cli/exit.go:14-21,53-61`). Every ordinary command must
switch renderers in the same stage that permits identityless or non-local problems.

### Adapter/operation matrix

Legend: `SID` = source entity ID, `LOC` = adapter location, `LP` = optional local path, `REV` =
opaque content version, `SET` = required corpus identity introduced by H1.

| Adapter/operation | Ordinary list/show | Graph/snapshot planning | `path`/copy/editor | Ordinary mutation | Guarded/batch mutation | Required contract |
|---|---|---|---|---|---|---|
| Task | `SID, LOC, LP?`; show must return loaded evidence | `SID, LOC, REV` for every readable/unreadable member | same-`SET` resolver; fail missing/ambiguous | local outcome for create/update/lifecycle | `SID, LOC, REV` snapshot plus same-`SET` writer | one scan projected to list/board/summary/graph |
| Epic | `SID, LOC, LP?` | no public revision for ordinary use | same-`SET` resolver | create/move outcome; store-internal CAS | explicit version only if a future multi-record plan needs it | derived filename ID remains source evidence |
| Audit | `SID, LOC, LP?` | no public revision for ordinary use | same-`SET` resolver | create/move outcome; store-internal CAS | as above | declared/source drift projection for lint |
| Research | `SID, LOC, LP?` | no public revision for ordinary use | same-`SET` resolver | create/update/move outcome; store-internal CAS | as above | declared/source drift projection for lint |
| Thread | `SID, LOC, LP?, REV` | `SID, LOC, REV` whole snapshot | same-`SET` resolver; fail missing/ambiguous | create/update outcome | same-`SET` mutation and full CAS snapshot | do not eagerly join path to every semantic read |
| Load problem | `SID?`, slug?, `LOC`, `LP?`, message | included in complete snapshots where health matters | no navigation without unambiguous `LP` | never mutation authority by itself | exact problem evidence compared in Thread/task graph | neither `LOC` nor `LP` may synthesize `SID` |

### Entity/evidence matrix

| State | Domain value | `Source.ID` | Declared ID evidence | `Location` / `LocalPath` | Version | Permitted behavior |
|---|---|---|---|---|---|---|
| Valid Task/Research/Audit with matching IDs | semantic fields only | required | explicit lint projection | location required; local path optional | Task graph only, via wrapper | list/show; mutate through same-corpus capability |
| Valid Epic | semantic fields only | required, derived by adapter | no independent declared ID today | same | none for ordinary read | list/show; internal optimistic retry for mutations |
| Valid Thread | semantic fields only | required | declared semantic ID if format supplies it | same | required for whole snapshot | list/show; guarded mutation via full snapshot |
| Missing/drifting declared ID | readable semantic value if parser permits | required when source name is ID-led | retained separately for lint | evidence retained, never identity | per entity policy | list/lint; identity-sensitive mutation uses source ID and policy |
| Unreadable source | none | optional | unknown | location required, local path optional | included where snapshot CAS needs it | diagnose; no semantic mutation authorization |
| Duplicate source IDs | multiple loaded records | same non-empty ID on each | per-record evidence | distinct or even unhelpful locations | per-record where relevant | list all via row handles; ID-only durable action fails ambiguous |

Explicit projections are preferable at the false semantic/adapter boundary: ID-drift validation,
duplicate explanation, source-attributed diagnostics, task-graph sorting, snapshot equality, and
mutation recovery all legitimately consume adapter-derived evidence. They should not force storage
fields back into domain entities, but they also cannot be reduced to `T` alone.

### Hostile identity states

| State | List/lint outcome | Show/resolution outcome | TUI outcome | Mutation authorization |
|---|---|---|---|---|
| Source ID matches declared ID | normal | unique show allowed | stable restoration by ID if unique | allowed after same-`SET` and snapshot checks |
| Source ID present; declared ID missing/drifting | retain and emit drift/missing evidence | address by source ID, show declared discrepancy | row keyed by private handle; label discrepancy | use source ID, never silently rewrite based on location |
| Duplicate source ID, distinct slugs/locations | retain both and diagnose | ID-only lookup fails ambiguous | display both with evidence; restoration prompts/clears | fail ambiguous unless an exact current snapshot occurrence is authorized |
| Duplicate source ID, useless/equal locations | retain both; position is snapshot-local only | fail ambiguous | private handles can render/select this snapshot | durable mutation fails unless adapter offers another explicit occurrence token |
| Missing source ID, parseable content | retain if use case permits and diagnose | no ID-addressed show | visible but non-restorable/non-actionable | reject identity-required mutation |
| Unreadable: ID only / slug only / both / neither | preserve every available field | no semantic show; problem detail only | diagnostic row, local action only with explicit unambiguous `LP` | reject semantic mutation; repair API needs its own guarded source token |
| Explicit ID contradicts path-like/opaque location | retain contradiction as evidence | ID remains identity | show both, never derive key from location | authorize on ID/snapshot, not location |

Slice position is sufficient only as a private handle inside one immutable snapshot. It is not safe for
session restoration, deferred path completion, CLI addressing, or mutation authorization.

### Snapshot and concurrency reconstruction

| Concern | Current behavior | Required wrapper behavior |
|---|---|---|
| Task parsing | Hash revision and filename/path evidence are written onto the domain task (`internal/store/fsstore.go:330-369`) | Construct `VersionedRecord[Task]{LoadedRecord{Value, Source}, Version}` once at scan time. |
| Task sorting/cloning | Graph construction retains duplicates and sorts with source evidence (`internal/core/dependency_graph.go:376-425`); source refs include ID/slug/location (`internal/core/dependency_source.go:272-298`) | Sort/copy the wrapper by source ID, semantic slug, location, and version. Never fall back to `Value.Path`. |
| Task snapshot equality | Requires complete/healthy snapshots and equal non-empty versions/source refs (`internal/core/dependency_graph.go:934-971`) | Compare wrapper source and version; planner-facing `Task()` may strip version but must preserve source evidence in an explicit projection. |
| Simulation/repair | Simulation copies all tasks/problems and clears revisions only for changed prospective values (`internal/core/dependency_source.go:137-224`); repair comparison tolerates version/update differences only for intended changes (`internal/core/dependency_repair.go:1155-1190`) | Clone wrappers; clear/replace only changed wrapper versions; retain unchanged source and problem evidence. |
| Thread snapshot | CAS compares readable/unreadable membership, source fields, health, and non-empty versions (`internal/store/cas.go:31-98`) | Preserve the same comparison over versioned loaded records/problems. |
| Guarded application | Lifecycle and Thread application re-read and compare before write (`internal/store/lifecyclemutation.go:47-111`; Thread mutation flow uses the same full-snapshot principle) | Re-read from the same `SET`; reject incomplete, changed, or cross-corpus snapshots. |

Epic/audit/research do not need to expose revision on every ordinary read merely because their stores
use optimistic concurrency internally. Their current read-modify-write loops re-read the exact path
and retry on conflict. The boundary changes only for operations that plan from a returned multi-record
snapshot or cross independently configured capabilities; those require an explicit version/snapshot
contract and `SET`, respectively.

### Local operations and output contract

| Operation | Lazy resolver safe? | Required output/evidence |
|---|---|---|
| List/show/info | Yes for an optional navigation affordance; no need to join `LP` eagerly | semantic result plus source evidence; show must preserve `SID/LOC/LP?` |
| CLI `path` | Yes after unique ID resolution and same-`SET` validation | path on success; not-found/ambiguous/non-local are distinct errors |
| TUI copy/editor | Yes only with captured generation + row handle + ID revalidation immediately before side effect | disable when unavailable; ignore stale completion; explain ambiguous/non-local state |
| Create dry run | No committed record exists to resolve | planned local path and `committed=false` in a typed local outcome |
| Create success | Resolver may supplement but must not be the only promised evidence | created local path when adapter promises a local file |
| Field/body/lifecycle update | Usually, after successful same-corpus commit | optional resulting path; explicit receipt if command currently prints it |
| Rename/move | No | exact source/destination local paths and durability state, including partial failure |
| Malformed repair | Not by canonical ID alone | explicit guarded source token/path supplied by the repair API, never inferred from message/location |

### Compatibility, schema, and exit policy

- Keep the wire DTO name and shape stable while renaming the core type. `path` is currently required
  in generated schema, so an empty string remains representable during migration.
- Add `location` and optional `local_path` without overloading `path`; when an old local problem is
  converted, preserve both the legacy path and the new fields. A non-local location may legitimately
  have no local path.
- Treat a changed `$defs` reference, required set, or ordinary-envelope problem type as a schema
  change even if the Go type rename is nominally internal.
- Preserve exit code 11 for usable results accompanied by load problems. Switch task, epic, audit,
  research, summary, and Thread human renderers to the portable formatter before allowing empty local
  paths.
- Update fakes to model missing local capability, duplicate IDs, identityless problems, contradictory
  location/path evidence, typed-nil adapters, explicit option reordering, and mismatched `SET` values.

### Migration-risk table

| Stage | Compile/behavior blast radius | Unsafe intermediate | Safe intermediate and scope amendment |
|---|---|---|---|
| 1. Introduce vocabulary | Core ports, stores, fakes, graph fixtures | Generic wrappers added beside a second filesystem scan | Build one internal loaded/versioned scan and compatibility projections to current `TaskGraphRead`, list, summary, and show interfaces. No domain-field removal. |
| 2. Promote diagnostics | All five stores, envelopes, CLI renderers, schema goldens | Core rename leaks a new `$defs` or empty path breaks legacy formatter | Keep wire DTO stable; add new evidence fields deliberately; preserve exit 11; migrate all renderers together. |
| 3. Ordinary reads/shows | Service ports, board, summary, lint, CLI/TUI details | Lists return loaded values but `Get*` returns bare values, losing evidence | Own both list and `Get*`/`Show*` projections in this stage; verify one authoritative scan and explicit lint records. |
| 4. Path capability split | Service options, workspace composition, every path/copy/editor action | Independent adapters are only documented as compatible; eager Thread join persists | Add `SET` handshake/bundle first; test option permutations and typed nil; add generation-aware TUI resolver. |
| 5. Mutation receipts | Core results, wire DTOs/schema, human renderers, failure tests | Remove domain paths before dry-run/rename recovery is represented | Land typed per-operation `LocalOutcome` and partial-failure rendering while old fields still exist. |
| 6. Domain cleanup | Domain structs, parsers, equality, lint, tests/fixtures | Mechanical deletion silently drops sorting, diagnostic, or CAS evidence | Delete `Path`/`FilenameID`/`SourceVersion` only after searches prove every consumer uses source/version/local-outcome projections. |

### Concrete amendments

1. Add an opaque corpus/source-set identity and require constructor-time compatibility across read,
   path, and mutation capabilities; specify typed-nil and option-order behavior for all five entities.
2. Add a snapshot-local TUI row handle and precise ambiguity/stale-generation rules. Keep canonical ID
   as the only durable identity, but not the only in-snapshot address.
3. Define per-operation `LocalOutcome` types, including planned/committed create paths and rename/move
   source, destination, and partial-durability state; carry them through wire and human renderers.
4. Expand the ordinary-read stage to own `Get*`/`Show*` loaded projections and ID-drift lint inputs.
   Establish a single internal scan with compatibility projections before changing public ports.
5. Assign generation-aware lazy TUI resolution and all synchronous `entity.Path` consumers to the
   path-split task; add stale-selection tests.
6. Make schema evolution and exit-11 renderer migration explicit acceptance criteria, including
   required `path`, additive `location`/`local_path`, and identityless non-local problems.
7. Require migration tests for duplicate IDs with equal/unhelpful locations, malformed sources with
   every ID/slug combination, contradictory location/path evidence, split-corpus adapters, partial
   rename durability, and create dry runs.

### Smallest user decision

No unresolved product choice is required to close these findings. The design already commits to
listing duplicate records, failing ambiguous durable operations, supporting split sources safely,
and preserving promised local outcomes. Those commitments imply snapshot-local UI handles,
same-corpus capability validation, and typed mutation outcomes. If cross-corpus read/path pairing is
actually desired as a product feature, that would be the one decision to revisit; the safe default is
to reject it.

### Verification and isolation attestation

- Isolated workspace: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xWGHW7`
- Isolated git directory: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xWGHW7/.git`
- Captured baseline commit: `d540443a599ff08d5a2a1e7db046e97d012c6050`
- Source fingerprint: `781792c97f14cafc56cd4b04327214200aa6e60f`
- Assigned deliverable: `planning/audits/6gdx1bvdf16p-2026-09-26-portable-entity-read-local-source-capability-design-codex.md`
- Commands/results: initial `go test ./...` could not use the host Go build cache under sandbox
  policy; `GOCACHE=/private/tmp/taskflow-audit-go-cache go test ./...` passed all packages;
  `git diff --check` passed; the isolation helper's `verify` command passed with this audit as the
  sole diff.
- Transfer result: succeeded via `isolated-review-workspace.sh transfer`; only the assigned audit was
  transferred to the handoff checkout.

I inspected the assigned audit in the source checkout only far enough to obtain its isolation
instructions, then performed all repository reads and this edit inside the helper-created independent
clone. I did not inspect another reviewer's deliverable, planning diff, working-tree state, or
conversation. Only the assigned audit is intentionally modified.
