---
schema: 1
id: 6gdpcag50p97
bucket: closed
area: portable-board-status-diagnostics-implementation-codex
date: "2026-09-25"
updated_at: "2026-09-26"
---
# Audit: Portable board and status diagnostics implementation — codex — 2026-09-25

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Perform an independent adversarial implementation and architecture review of task `6g6jqqcdehne`,
`preserve-portable-load-diagnostics-in-board-and-status`. Treat every checked acceptance criterion,
documentation statement, and green test as a claim to falsify. Reconstruct the diagnostic flow from
secondary adapters through core projections, retained cross-space state, wire DTOs, CLI policy, and
TUI attention. Then try to make a plausible adapter or local repository produce silent data loss,
invented filesystem semantics, the wrong exit result, an incoherent snapshot, or a misleading clean
dashboard.

This review must distinguish runtime defects, regression-test gaps, and deferred architecture work.
Do not report the already-sequenced portable entity-read redesign merely because `SummaryStore`
still has a transitional `FileProblem` contract; demonstrate a current contract violation or a
concrete obstacle before treating that boundary as a finding.

## Review target

Review the complete uncommitted implementation on branch
`feat/portable-board-status-diagnostics` relative to base
`06546ca99c300db8b4e971aa8236aaafb7c6c551`. The primary planning target is
`planning/tasks/6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md`.
Include source, tests, generated schema/comments and machine goldens, ADR/architecture amendments,
and task bookkeeping.

Build a repository-wide consumer inventory and trace at least these seams:

- `TaskGraphLoadProblem`, `TaskGraphRead`, local-path compatibility, opaque location, source
  revisions, sorting, cloning, graph health, and `SameSourceSnapshot`;
- conversion to `LintLoadProblem` in `Board`, `Summary`, repository lint, and compatibility list
  reads, including which layer may recover identity;
- `SummaryStore`, `PlanningSummarySource`, `SpaceOverviewService`, retained-summary cloning, and
  current versus cross-space status;
- filesystem task/audit scanner identity and legacy epic filename identity recovery;
- `BoardEnvelope`, `SummaryJSON`, `StatusAllEnvelope`, `LintLoadProblemJSON`, schema revision 1.75,
  generated JSON Schema/comments, and machine goldens;
- board/status human rendering, partial-result errors, exit code behavior, stdout/stderr, and local
  repair-location usefulness;
- TUI overview and Atlas attention, including refresh/retention paths and pathless diagnostics;
- remaining list, lint, repair, graph mutation, and guarded-CAS consumers that intentionally retain
  local `FileProblem` or `Path` contracts.

Do not implement fixes or edit any file other than this assigned audit.

## Intended contract to challenge

1. A pathless task load failure retains kind, stable ID, slug, opaque location, and message through
   Board, current Summary, cross-space Summary, typed wire output, CLI policy, and TUI attention.
2. Identity is authoritative. Core never derives it from an opaque or local location; a misleading
   location cannot replace or relabel an explicit ID/slug.
3. Local filesystem failures remain at least as actionable as before: identity and repair filename
   survive human errors, required wire `path` and `message` remain compatible, and partial reads
   retain exit code 11 after rendering available results.
4. `LocationIsPath` is semantic, not decorative. Opaque locations populate `location` but never
   `path`; local locations populate both. An empty or identity-only diagnostic invents neither.
5. Mixed task, epic, and audit problems preserve deterministic kind/identity/location attribution
   and task→epic→audit ordering without an extra entity scan. Epic filename knowledge remains in
   the filesystem adapter.
6. Task graph health and guarded source equality remain safe. Adding neutral location cannot hide
   an unreadable record, weaken CAS evidence, leak `SourceVersion`, or cause local repair paths to
   disappear from graph diagnostics.
7. Long-lived consumers clone and retain the complete portable problem values. TUI overview and
   Atlas treat pathless failures as attention without assuming a path exists.
8. Schema 1.75 is an honest additive revision: old `path`/`message` keys remain, new optional keys
   are correctly described, all affected envelopes use the shared DTO, and regeneration introduced
   no unrelated semantic drift.
9. The implementation does not prematurely redesign unrelated list/entity ports. Any transitional
   compatibility conversion is explicit, bounded, documented, and unable to parse a path in core.

## Mandatory evidence floor

- Produce a consumer inventory for every `TaskGraphLoadProblem`, `LintLoadProblem`, and affected
  `Summary.Problems` read/write. Classify each as portable semantic data, deliberate local repair
  evidence, guarded snapshot evidence, presentation mapping, or unresolved leakage.
- Create an independent field-lineage table for `kind`, ID, slug, location, `LocationIsPath`, local
  path, message, and source revision. Cite the exact producer, each conversion, each consumer, and
  whether omission is legal. Do not infer lineage from type names alone.
- Exercise at least these diagnostic inputs without reusing a production converter as the expected
  oracle: identity-only; pathless ID+slug; opaque location; explicit identity plus contradictory
  opaque location; local path with recovered identity; invalid local filename without identity;
  empty location; mixed task/epic/audit failures; and multiple failures whose input order differs.
- For Board, current status, and cross-space status, inspect core values, JSON, human output, and
  exit results. Verify pathless wire output and local repair output separately. Use a sandbox-built
  binary for real filesystem cases and focused independent fakes for pathless adapters.
- Measure calls/opens for Board and Summary. Prove one task graph read and one epic/audit sweep,
  and look for hidden conversion-time rescans or body rereads. Test both success and mixed-failure
  snapshots.
- Verify retained `SpaceOverview` summaries are deep enough for the new diagnostic slice and that a
  pathless failure remains visible through stale-summary reconciliation and the Atlas statistic.
- Inspect schema 1.75 semantically, not just textually. Validate non-default unreadable branches for
  Board, Summary, and StatusAll against the generated schema; compare old `{path,message}` tolerant
  consumers; confirm `path` stays present inside each emitted diagnostic; independently regenerate
  schema comments and goldens and classify every non-version diff.
- Run and restore at least ten targeted mutations. Required mutations include: collapse graph
  problems back to `FileProblem`; reconstruct identity from location; serialize every location as a
  path; drop ID/slug in Board only; misclassify epic/audit kinds; reorder mixed problems; add a
  second task scan; omit problem cloning during stale retention; remove location fields from source
  equality/sort; move epic filename parsing into core; and make TUI attention conditional on a
  non-empty path. At least three must be coordinated cross-layer mutations rather than trivial
  compile failures. Record the exact focused test that fails for the intended reason, or record a
  surviving mutant as a test-gap finding.
- Audit test independence. Identify fakes and assertions that call production conversion helpers,
  use only counts, or use zero-value diagnostics in ways that could let identity/location loss pass.
- Run focused tests, `go test -race ./...`, `golangci-lint run ./...`, `go mod tidy -diff`, planning
  lint, audit lint, schema-comment freshness, machine-golden/schema drift checks, and
  `git diff --check`. Report anything skipped and its exact environmental blocker.

## Required hostile angles

- **Dual location representation:** challenge every state where `Location`, `LocationIsPath`, and
  legacy/local `Path` disagree. Determine whether precedence is explicit and stable across graph
  diagnosis, CAS comparison, dashboards, and wire mapping. Use actual contradictory values.
- **Identity authority:** attempt to make a plausible location masquerade as another task while the
  explicit ID/slug names the correct task. Check error prose, JSON, sorting, and graph resolution.
- **Snapshot equality:** vary only location, path classification, message, source revision, and
  ordering. Confirm changes that matter invalidate equality, ordering alone does not, and missing
  revision evidence still fails closed.
- **Compatibility versus portability:** approach once as a remote-adapter author with no paths, and
  once as an existing machine consumer that only knows required `path`/`message`. A solution that
  satisfies only one side is defective.
- **Aggregate summary boundary:** test identity-bearing and identityless epic/audit `FileProblem`
  inputs. Verify the filesystem owns epic filename semantics and core merely copies supplied
  fields. Do not demand opaque epic/audit locations from a port that this task explicitly leaves to
  the next design slice unless the current code falsely claims to support them.
- **Partial-result policy:** verify every affected command renders the usable result before failing,
  emits valid JSON even on a nonzero exit, and does not turn graph-health warnings into unrelated
  validation failures.
- **Long-lived UI behavior:** test refresh, stale retention, and attention counting with an
  identity-only problem. Look for aliasing, dropped state, false all-clear displays, or navigation
  code that dereferences a path.
- **Generated-contract blast radius:** global schema-version regeneration can hide unrelated drift.
  Compare structures, required fields, and descriptions instead of accepting a changed golden
  count as proof.
- **Future adapter pressure:** use a minimal `TaskGraphSource` and a split `PlanningSummarySource`.
  Identify any compile-time or runtime demand for filesystem facilities not required by the named
  port, but require a runnable example before filing an architectural finding.

## Validation and restoration

Run all probes only in the mandatory independent sandbox. Before each mutation, record the baseline
test result and expected kill. Restore source/tests/generated artifacts to the sandbox baseline after
each probe. End with the assigned audit as the only diff, inspect it, run the isolation helper's
verification, and transfer through the helper. Never format, generate, stage, commit, or repair the
shared source checkout.

## Deliverable

Preserve this brief. Add a severity-ranked report with exact paths/symbols, consumer inventory,
field-lineage table, probe matrix, call counts, mutation table, schema-diff assessment, command
results, and isolation attestation. Add each surviving defect under `## Findings` using exact
repository grammar and leave every finding open for owner triage. A no-findings verdict is valid
only when the hostile evidence settling every intended-contract item is recorded; “the suite is
green” or a checklist of inspected files is not sufficient.

## Reviewer report

### Verdict

The implementation is not ready to close. The main portability path is sound: explicit identity
survives Board, current Summary, cross-space Summary, wire output, exit policy, and TUI attention;
local filesystem scans recover identity at the adapter boundary; source revisions do not leak; the
new DTO remains additive for old `{path,message}` readers; and task/epic/audit reads stay single-pass.
However, three runtime defects and one regression-test gap remain open:

1. Board and Summary expose adapter input order rather than a canonical diagnostic order.
2. A task diagnostic that legitimately carries both an opaque location and a local repair path
   loses the local path at the dashboard boundary.
3. Human `status --all` reduces every portable diagnostic to a count, so neither remote identity nor
   a local repair filename/message reaches the operator.
4. The complete committed core suite still permits conditional identity reconstruction from a
   path-looking opaque location.

Pass one reconstructed every producer, conversion, retained state, wire mapper, CLI policy, and TUI
consumer. Pass two independently attacked shared converters, optional fields, snapshot equality,
category ordering, retained aliases, filesystem ownership, output compatibility, and test-helper
independence. The findings below are the defects that survived both passes; they are left open for
implementation-owner triage.

### Isolation attestation

| Item | Value |
| --- | --- |
| Workspace | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.u4dTIl` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.u4dTIl/.git` |
| Baseline commit | `90224a61042b8ee72d203e4ed85a4448f10b325c` |
| Baseline parent | `06546ca99c300db8b4e971aa8236aaafb7c6c551` |
| Captured source blob | `683bf00a7ca7e7ac1170fa5ad69def9bbff8de88` |
| Captured source fingerprint | `fb29b3d40bf78894f317565ed0828450691ce0c3` |
| Deliverable | `planning/audits/6gdpcag50p97-2026-09-25-portable-board-status-diagnostics-implementation-codex.md` |
| Pre-transfer verification | Recorded below; the helper must report this audit as the sole delta. |
| Transfer result | Pending at report freeze; the authoritative helper result is included in the reviewer handoff because a successful one-shot transfer cannot update the file it just copied. |

All inspection, builds, tests, generated checks, hostile fixtures, and source mutations ran inside
this independent `--no-hardlinks` clone or in its disposable `/tmp` build/fixture outputs. After the
helper created the clone, no inspection, mutation, test, stage, commit, branch switch, stash, reset,
clean, or write-capable project command ran in the shared checkout. Every probe and generated output
was removed or restored manually before this report was written; no reviewer commit was created.

Expected final helper attestation, to be checked verbatim after the report is complete:

```ini
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.u4dTIl
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.u4dTIl/.git
baseline_commit=90224a61042b8ee72d203e4ed85a4448f10b325c
source_blob=683bf00a7ca7e7ac1170fa5ad69def9bbff8de88
source_fingerprint=fb29b3d40bf78894f317565ed0828450691ce0c3
deliverable=planning/audits/6gdpcag50p97-2026-09-25-portable-board-status-diagnostics-implementation-codex.md
deliverable_changed=true
transfer=pending
```

### Consumer inventory

| Boundary or consumer | Classification | Verified behavior |
| --- | --- | --- |
| `store.FS.ReadTaskGraph` (`internal/store/fsstore.go:138-152`) | Portable producer plus guarded snapshot evidence | One `scanDirWithSourceVersions` pass. `TaskGraphLoadProblemFromFile` receives recovered filename identity, local path, message, and the opaque content hash. |
| `taskStoreGraphSource` and `TaskGraphReadFromFiles` (`internal/core/service_task.go:135-171`) | Local compatibility adapter | Legacy `ListTasks` implementations become neutral graph reads. Local path populates `Location`, `LocationIsPath`, and guarded `Path`; no source revision is invented. |
| `TaskGraph` construction (`internal/core/dependency_graph.go:330-380`) | Guarded snapshot and graph-health evidence | Clones and canonically sorts load problems, attributes unreadable IDs without parsing location, uses only `taskGraphLocalPath` in graph diagnostics, and marks valid unreadable IDs as hard-broken. |
| `SameSourceSnapshot` and repair equality (`internal/core/dependency_graph.go:941-979`; `dependency_repair.go:1186`) | Guarded CAS evidence | Compares identity, neutral location, path classification, local path, message, and a non-empty source revision after canonical sorting. Ordering alone is ignored. |
| `Service.ListTasks` (`internal/core/service_task.go:20-90`) | Deliberate local list compatibility | Converts graph problems back to `domain.FileProblem`; `taskGraphLocalPath` keeps repair paths and does not expose opaque locations as paths. |
| `Service.Board` (`internal/core/board.go:42-81`) | Portable semantic projection | Reads one task snapshot, maps problems with `taskGraphLoadProblems`, and builds graph health from the same read. It does not rescan, but currently preserves adapter problem order. |
| `Service.Summary` (`internal/core/service.go:360-460`) | Portable task projection plus transitional local aggregate seam | Reads one task graph, one epic sweep, and one body-aware audit sweep. Task problems are neutral; epic/audit `FileProblem` identity is copied at the documented compatibility boundary. Output is grouped task, epic, audit but not canonically ordered within each group. |
| `Service.Lint` (`internal/core/service.go:493-727`) | Portable lint semantic data | Converts task `LintLoadProblem` back into graph input without filling guarded `Path`; the graph derives a local path only when `LocationIsPath` is true. Other entity kinds and Threads retain their neutral diagnostics. |
| Filesystem epic and generic scanners (`internal/store/resolve.go:36-80`; `epicstore.go:18-43`; `lintsource.go:10-62`) | Local identity recovery | ID-led tasks/audits/research recover ID+slug in the generic scan. Legacy epic identity is the filename stem and is recovered in `ListEpics`, not in core. Invalid local filenames retain a path/message without invented identity. |
| `cloneSpaceSummary` and reconciliation (`internal/core/space_overview.go:145-209`) | Retained long-lived state | Deep-copies the diagnostic slice along with other mutable summary slices; conflict retention preserves pathless problems and marks the result stale. |
| `BoardEnvelope`, `SummaryJSON`, and `StatusAllEnvelope` (`internal/wire/envelopes.go:48-92,246-367`) | Presentation mapping | All use `ToLintLoadProblemsJSON`. Status-all embeds the same versionless Summary payload beneath one top-level schema revision. |
| `ToLintLoadProblemsJSON` (`internal/wire/envelopes.go:938-968`) | Shared machine mapping | Preserves kind/ID/slug/location/message; required `path` is populated only from a path-classified `Location`. This is correct for ordinary pathless/local inputs but cannot preserve a separate guarded `Path`. |
| Board/current-status CLI (`internal/cli/board.go:19-35`; `status.go:118-134`; `problems.go:74-110`) | Partial-result policy | Renders usable output first and then returns `ErrValidation`/exit 11. Identity leads the error name; a local filename is appended when available. |
| Cross-space status CLI (`internal/cli/status.go:53-115`; `render/status.go:151-220`) | Partial-result policy with presentation loss | JSON retains every field and exit 11 remains correct. Human rendering and the final error count problems but never render identity, location, or message. |
| TUI Overview and Atlas (`internal/tui/dashboard.go:185-214`; `atlas.go:950-987`) | Attention-only presentation | Both use `len(Summary.Problems)`, so identity-only/pathless problems remain attention. Neither dereferences or requires a path. |
| Graph mutation, repair, list, path, lint-fix, and entity detail ports | Deliberate local/guarded contracts | They continue to use `Path`, `FileProblem`, source declarations, and revisions where a local repair or guarded write needs them. No unrelated entity-read port was widened by this task. |

### Field lineage

| Field | Producer and conversions | Consumers and omission rules |
| --- | --- | --- |
| Entity kind | `taskGraphLoadProblems` assigns task; Summary assigns epic/audit at `service.go:435-438`; lint adapters assign their entity kind. | Required semantically in dashboard values, optional on wire for additive compatibility. Used in human labels and schema consumers; never derived from location. |
| Stable ID | Generic scanner recovers ID-led filename identity (`resolve.go:67-71`); `ListEpics` recovers the legacy full stem; remote adapters supply it directly. | Copied through graph→lint→wire and used before location in errors. Legal to omit only when the adapter cannot recover identity. |
| Slug | Generic scanner recovers the suffix for ID-led names; remote adapters may supply it. Epics have no separate slug here. | Copied unchanged; preferred human label. Optional and never reconstructed in core. |
| Neutral location | Local adapter sets it to the path; remote adapters may use URI/key; `taskGraphLoadProblems` falls back to guarded `Path` only when Location is empty. | Preserved in core/wire/human lint output. Optional. Core does not parse it. |
| `LocationIsPath` | Local file converters set true; remote adapters own the declaration. | Controls graph local-path fallback and public wire `path`. Empty/false must not invent filesystem semantics. Included in source equality and sort keys. |
| Guarded/local path | Filesystem task graph sets `Path` beside the neutral location. It survives `taskGraphLocalPath`, graph diagnostics, repair, and CAS. | Intentionally absent from `LintLoadProblem`; ordinary equal local inputs recover it through Location. A simultaneous opaque Location causes the runtime loss in M2. |
| Message | Parse/read producer supplies it; every conversion copies it. | Required on the wire and used in graph health and detailed lint rendering. No parsing or classification is based on its prose. |
| Source revision | `scanDirWithSourceVersions` hashes exact unreadable bytes; non-filesystem adapters supply an opaque token. | Compared by `SameSourceSnapshot` and repair equality; non-empty evidence is required for equality. Explicitly excluded from user-facing JSON/YAML. |

### Hostile input and behavior matrix

| Input or angle | Observed result |
| --- | --- |
| Identity only | Board/Summary retained task kind and ID with empty location/path. |
| Pathless ID + slug | Core, Board JSON, current status JSON, cross-space Summary, and errors retained both fields. |
| Opaque location | `location` was retained and compatibility `path` stayed the required empty string. |
| Explicit identity plus contradictory opaque location | Explicit ID/slug controlled error prose and graph addressability; the misleading location never relabelled the record. |
| Path-looking opaque location with no identity | Current code correctly left identity empty. A conditional location parser mutation passed all committed `internal/core` tests and was killed only by a review probe (L1). |
| Legacy local `Path` only | `taskGraphLoadProblems` promoted it to a path-classified Location; wire `path` and human repair filename survived. |
| Opaque Location plus distinct local `Path` | Graph health/CAS retained the local path, but Board/Summary dropped it and wire `path` became empty (M2). |
| Invalid local filename without identity | Path/message survived and no identity was fabricated. |
| Empty location | No location or path was invented; identity/message remained. |
| Mixed task/epic/audit failures | Kind and category grouping were correct. Reversing multiple inputs within a category changed Board/Summary output order (M1). |
| Snapshot equality | Location, classification, local path, message, and revision-only changes all invalidated equality; input reordering did not; missing revision evidence failed closed. |
| Retained stale Summary | Problems were copied independently and remained visible to Atlas attention after contention reconciliation. |
| Human partial results | Board and current status rendered first, named local filenames in their error, and exited 11. Cross-space status exited 11 but supplied counts only (M3). |

### Calls, opens, and coherence

The committed counters and independent fakes established these exact boundaries:

| Projection | Success/mixed-failure reads |
| --- | --- |
| Board | One `TaskGraphSource.ReadTaskGraph`; no epic/audit reads and no conversion-time rescan. |
| Current Summary | One task graph read, one `ListEpics`, and one `ListAuditsWithFindings`; audit metadata and findings share the same body sweep. |
| Cross-space Summary | `OpenPlanningStore` returns one composite `PlanningSummarySource`; `summarize(planningStore, planningStore, ...)` performs the same 1/1/1 reads. |
| Filesystem task graph | One resilient directory scan; each readable or unreadable source is opened once and unreadable bytes are hashed before parsing. |

The mixed-failure counter test observed task/epic/audit `1/1/1`. The duplicate-task-scan mutation
changed it to `2/1/1` and failed immediately. A real sandbox-built binary over a local malformed task
produced full Board and current-status JSON, then exit 11; human output rendered the dashboard,
graph repair detail, and `broken (6g0000000001-broken.md)` before the same exit.

### Mutation table

| Mutation | Result and focused evidence |
| --- | --- |
| Collapse neutral graph diagnostics to path/message-only data | Killed by `TestBoard_PreservesPathlessTaskLoadProblemIdentity`; ID, slug, and opaque location disappeared. |
| Conditionally reconstruct missing identity from a path-looking location | Survived the complete committed `go test ./internal/core`; a review-only `db://tasks/6g0000000009-wrong.md` probe killed it. Finding L1. |
| Serialize every Location as compatibility `path` | Killed by `TestToLintLoadProblemsJSONKeepsOpaqueLocationsOutOfPath`. |
| Drop ID/slug in Board only | Killed by `TestBoard_PreservesPathlessTaskLoadProblemIdentity`. |
| Misclassify epic/audit problems as tasks | Killed by `TestService_Summary_PreservesMixedPortableLoadDiagnostics`. |
| Reorder the task/epic/audit category groups | Killed by the same mixed-summary test. Independently reversing two problems inside one category exposed the untested runtime defect M1. |
| Add a second task read in Summary | Killed by the mixed-summary `2/1/1` call assertion. |
| Omit diagnostic-slice cloning during stale retention | Killed by `TestSpaceOverviewRetainedSummaryOwnsMutableSnapshotData`; the retained message aliased the prior snapshot. |
| Remove Location and path-classification from unreadable-source sort/equality | Killed by `TestTaskGraphSameSourceSnapshotComparesOpaqueUnreadableRevisions`; independent probes also changed classification, Path, message, and revision one at a time. |
| Remove epic identity recovery from the store and recreate it in core | The Summary test still passed, but `TestFS_ListEpics_MissingFrontmatterIsLoud` failed with empty adapter identity, proving filesystem ownership. |
| Make TUI Overview and Atlas attention require a non-empty local path | Killed by `TestDashboardNeedsAttentionReportsPathlessUnreadableRecord` and `TestAtlasAttentionFoldsOnlyWhatWantsAPerson`. |

The store/core epic mutation and the paired Overview/Atlas mutation were coordinated across adjacent
layers rather than compile-only edits. The shared graph→lint collapse likewise exercised Board,
Summary, wire, and CLI consumers through one abstraction. All production mutations were manually
reversed and a clean race run followed.

### Test-independence assessment

- The new Board, Summary, cross-space, wire, and TUI tests construct non-default diagnostic values
  directly and assert fields; none uses `taskGraphLoadProblems` or `ToLintLoadProblemsJSON` as its
  expected oracle.
- The mixed Summary test asserts category order but has only one diagnostic per category, which is
  why adapter-order drift survives (M1).
- The Board identity test always supplies explicit identity. The older no-inference test covers the
  `TaskGraphReadFromFiles` compatibility conversion, not the new dashboard converter, so a
  conditional location parser survives the full core suite (L1).
- The schema validator uses production constructors and the production generated schema, but
  independent field assertions, byte goldens, complete regeneration, and the old-reader `jq`
  projection prevent a shared omission from being accepted merely because validation is green.
- TUI tests now use identity-bearing pathless values rather than zero-value diagnostics; they kill
  path-dependent attention logic. Count-only assertions are appropriate there because TUI scope is
  explicitly attention, not repair rendering.

### Schema 1.75 and compatibility assessment

Schema 1.75 is additive for the affected envelopes. `BoardEnvelope` and both Summary placements now
reference the already-shared `LintLoadProblemJSON`; its required set remains exactly `path,message`,
while kind, ID, slug, and location are optional. A tolerant old consumer independently projected
`{path,message}` from real malformed-task Board and status output successfully. Opaque non-default
branches for Board, Summary, and StatusAll passed the generated Draft 2020-12 schema validator.

The complete schema diff contained the global 1.74→1.75 revision/const changes plus exactly three
`FileProblem`→`LintLoadProblemJSON` references: Board unreadable, SummaryEnvelope unreadable, and the
reusable SummaryJSON used by status-all. No unrelated required fields, structures, descriptions, or
envelope registrations changed. `go run ./internal/tools/schemacomments` regenerated 269 comments
with no diff; `go test ./internal/cli -update` regenerated the complete machine-golden set with no
diff. The committed global golden churn is therefore version stamping, not hidden semantic drift.

### Repair-command execution

Every repair command emitted by the challenged graph states was executed against a state that
recommended it:

- `lint` ran against the real malformed local task, retained identity/location/message, and exited
  11 as expected.
- `task depend migrate` ran in `TestTaskDependMigratePreservesContentAndIsIdempotent`, applied the
  recommended legacy migration, and then proved idempotent.
- `task depend repair` diagnosis, its emitted `--auto`, and its explicit `--drop` remedy ran in
  `TestTaskDependRepairDiagnosesThenAppliesAutoAndExplicitIntent`; final graph health was healthy.

### Validation and restoration

- `go test -race ./...` — passed after every mutation and review-only test was removed.
- `golangci-lint run ./...` with isolated caches — `0 issues.`
- `go mod tidy -diff` — passed with no output.
- `go test ./internal/wire -run '^TestJSONSchema_ValidatesRealOutput$'` — passed with populated
  Board, Summary, and StatusAll unreadable branches.
- `go test ./internal/cli -update` — passed; complete machine-golden regeneration left no diff.
- `go run ./internal/tools/schemacomments` — generated 269 comments; no diff.
- Sandbox-built real CLI Board/status, JSON/human, and local malformed task — usable output rendered;
  all four commands exited 11 and preserved the repair filename for current-repo commands.
- Old `{path,message}` `jq` projections over real Board/status JSON — passed.
- `tskflwctl -C . lint` — `all planning entities and dependency links pass lint`.
- `tskflwctl -C . audit lint 6gdpcag50p97` — passed after all four open findings were recorded.
- `git diff --check` — clean.
- No validation was skipped. The first race/build invocation attempted the host Go build cache and
  was refused by the sandbox; rerunning with `GOCACHE=/tmp/taskflow-review-gocache` passed. This was
  an environment-only retry, not a product failure.

## Findings

#### M1. Board and Summary diagnostics inherit adapter input order · **Status:** fixed

**File:** internal/core/service.go:365 | **Component:** core
**Effort:** S · **Urgency:** soon

Board and summarize convert read.Problems before NewTaskGraphRead canonicalizes its private clone. A review fixture supplied the same two task failures in opposite orders and observed opposite Board/Summary problem order; epic and audit compatibility slices are likewise appended in returned order. This makes machine output and the first graph warning disagree under an otherwise equivalent unordered remote result. The committed mixed-kind test has one problem per kind, so it cannot detect within-kind order drift.

**Recommendation:** Canonically sort portable diagnostics by entity-kind rank, identity, location classification, location, and message before Board/Summary publication; add reversed multi-problem fixtures for each category.

**Resolution:** Canonicalized Board and Summary diagnostics by entity-kind rank
and explicit identity/location fields; reversed multi-record fixtures pin stable
within-kind order.

#### M2. Dashboard conversion drops a distinct local repair path · **Status:** fixed

**File:** internal/core/service_task.go:195 | **Component:** core
**Effort:** S · **Urgency:** soon

TaskGraphLoadProblem intentionally carries both adapter-neutral Location and guarded local Path. taskGraphLoadProblems chooses Location whenever non-empty and LintLoadProblem has no second path field. With Location=db://tasks/6, LocationIsPath=false, and Path=/planning/tasks/6g0000000006-six.md, graph diagnostics and CAS retain the local path while Board/Summary discard it; ToLintLoadProblemsJSON must then emit an empty compatibility path. If LocationIsPath is incorrectly true instead, the opaque location is mislabeled as a path. This violates the dual-location and legacy path compatibility contract.

**Recommendation:** Carry local compatibility path independently through the dashboard diagnostic and map wire location and path from their distinct sources; pin contradictory-value behavior in core and wire tests.

**Resolution:** Added an independent local repair path to LintLoadProblem,
preserved it beside opaque locations, and mapped the two fields separately
through wire and human output with contradictory-value regressions.

#### M3. Human cross-space status hides every diagnostic detail · **Status:** fixed

**File:** internal/cli/status.go:94 | **Component:** cli
**Effort:** S · **Urgency:** soon

StatusAll JSON preserves full portable diagnostics and the command correctly exits 11, but StatusAllHuman only appends an unreadable-record count and statusAllProblemsError also reports only the aggregate count. A fixture combining an identity-only remote task with a local audit showed no ID, slug, filename, location, or message anywhere in human output or the final error. Board and current status name the first records through portableProblemsError, so cross-space status is strictly less actionable.

**Recommendation:** Render portable problem details under each loaded space and include bounded space-qualified identities or filenames in the final partial-result error.

**Resolution:** Cross-space human status now renders each portable diagnostic
under its owning space and returns a bounded space-qualified partial-result
error.

#### L1. No committed dashboard test forbids identity parsing from location · **Status:** fixed

**File:** internal/core/board_test.go:95 | **Component:** tests
**Effort:** XS · **Urgency:** eventually

A coordinated mutation in taskGraphLoadProblems conditionally parsed the first 12 characters of a path-looking location only when TaskID was empty. The complete committed internal/core suite passed. The new Board test always supplies explicit identity, while the older no-inference test exercises TaskGraphReadFromFiles rather than the dashboard conversion. A review-only opaque input db://tasks/6g0000000009-wrong.md killed the mutation.

**Recommendation:** Add Board and Summary tests with empty identity plus a path-looking opaque Location and require EntityID and EntitySlug to remain empty through core output.

**Resolution:** Added Board and Summary regressions with empty identity and
path-looking opaque locations; neither projection invents an ID or slug.
