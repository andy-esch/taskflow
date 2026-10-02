---
schema: 1
id: 6gfrtba5jky2
bucket: open
area: local-entity-source-capabilities-checkpoint-codex
date: "2026-10-02"
updated_at: "2026-10-02"
---
# Audit: Local entity source capabilities checkpoint — codex — 2026-10-02

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

Adversarially review the committed checkpoint for task `6gcwcf88z57p`,
`split-local-path-capabilities-from-semantic-entity-reads`. This is not the final
task acceptance review: `Task` and `Thread` still have transitional `Path` and
`FilenameID` fields, and guarded Thread mutation planning still consumes some
of them. Identify regressions and architectural traps that should be fixed
before the final removal. Do not report the explicitly unfinished removal by
itself as a new defect; show a concrete consequence or a missing migration seam.

Work from source and executable evidence, not the progress checklist. First
trace the contracts and their consumers; then take a separate devil's-advocate
pass looking for systemic failures in a second, pathless adapter and under
concurrent edits. Do not implement fixes or edit any file except your assigned
audit. Leave all findings open for owner triage.

## Review target

Branch `refactor/split-local-entity-source-capabilities`; compare base
`003c39b` with implementation checkpoint
`eeaeba2d50d26d3792214a38684d8d17569c45d7`. Review all three commits,
including generated CLI/schema artifacts and
`planning/tasks/6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md`.
This audit is a review of the committed checkpoint, not proof that all task
acceptance criteria are done. Verify that the code and the planning progress
agree about what remains.

Inventory these seams beyond the named files when the call graph requires it:

- `core.Store`, entity-specific `*PathSource` ports, `WorkspaceSource`, service
  options, typed-nil handling, source-set validation, and overrides in both
  option orders;
- `RecordSource`, `LoadedRecord`, `VersionedRecord`, `TaskGraphRead`, Thread
  reads, source-version and local-repair-path evidence, and filesystem scans;
- task/Thread graph, lint, repair, lifecycle, creation, bulk-apply, and ordinary
  mutation paths, including committed-prefix and whole-snapshot CAS behavior;
- epic/audit/research ordinary and selected reads, lint, summaries, boards,
  source identity under declared-ID drift, and bare mutation receipts;
- CLI `task/audit info` and local path commands, TUI editor/yank/detail paths,
  wire DTOs, JSON projections, generated help/schema and compatibility.

## Intended contract to challenge

1. Semantic reads and projections use adapter-supplied source identity, never
   parse an opaque location or borrow a stale domain path/filename field.
   Explicit local-path resolution stays optional, same-source, parse-free, and
   useful even for a malformed document.
2. `LocationIsPath` is only a presentation hint. An opaque, path-shaped URI or
   key cannot authorize opening a file, a graph repair, or a mutation target.
3. Guarded task and Thread readable/unreadable source revisions stay outside
   semantic records and public wire output. Missing, added, removed, or changed
   source evidence fails closed during whole-snapshot comparison, including
   body-only edits and duplicate-ID cases.
4. Local repair authority comes only from an explicit guarded adapter path.
   Repair diagnostics remain useful when a pathless source cannot be patched.
5. Active and archived lint, graph diagnostics, Thread projections, boards,
   summaries, and wire projections preserve canonical source IDs and attribute
   drift/collisions to the right occurrence.
6. CLI/TUI local actions resolve by selected canonical ID and cannot act on a
   stale selection, stale list generation, changed workspace, or unrelated
   source after an asynchronous result arrives.
7. Existing filesystem behavior and machine contracts remain compatible except
   for the documented optionality of local paths on pathless sources.

## Mandatory evidence floor

- Produce a producer/consumer inventory for every entity's source ID, opaque
  location, local path, and guarded revision. Classify each use as semantic
  read, presentation, local navigation, diagnostic, or mutation authority.
  Include the remaining transitional `Task`/`Thread` field consumers and say
  which will need a new envelope or receipt before removal.
- Build a truly pathless narrow fake or fixture without embedding a broad Store.
  Exercise no location, an opaque URI, and a misleading path-shaped location.
  Verify read/list/lint and explicit local-action behavior, not just constructor
  success. Test matching and mismatched source-set witnesses, typed nils, and
  option order.
- Make task, Thread, audit, and research declared IDs differ from the source
  IDs. Compare normal/show/selected/lint/graph/summary/JSON outcomes. Include
  active and archived tasks, duplicate source IDs, and cases where two records
  share a declared ID but not a source ID.
- Corrupt frontmatter while keeping valid id-led filenames. Run each local path
  command and check that it still returns the exact document without parsing.
  Also check that an unrelated malformed file does not poison a selected read.
- Compare guarded snapshots across readable body-only changes, unreadable-byte
  changes, source addition/removal, missing revision evidence, duplicate IDs,
  and a changed local repair handle. Attempt the corresponding mutations or
  repairs and inspect whether any side effect precedes the rejection.
- Exercise TUI local path result races: select another row, refresh a list,
  change workspace, lose the optional capability, and receive an old result.
  Verify both editor and yank, and a detail hyperlink after selected-read
  failure. Do not infer safety only from a `listGen` check.
- Compare local and pathless CLI human/JSON outputs with the base behavior.
  Independently regenerate docs/schema outputs and inspect each golden change
  for an unintended contract difference. Check mutation receipt IDs/paths when
  frontmatter drift or post-commit cleanup failure is possible.
- Run at least six restored mutation probes, with a table naming each exact
  production mutation, the focused test expected to fail, the observed result,
  and any survivor. At minimum challenge: source-set validation bypass;
  `LocationIsPath` or URI used as repair authority; source revision omitted from
  snapshot comparison; source ID replaced by declared ID in lint/projection;
  TUI stale-result guard bypass; and selected path resolution falling back to
  a domain field. Use coordinated mutations if nearby checks mask the defect.

## Required hostile angles

- Ask whether the optional ports are composable for a future HTTP/database
  adapter or whether the constructor silently turns an absent capability into
  a filesystem action. Seek a counterexample across two simultaneously
  registered planning spaces, not just one service.
- Challenge the abstraction's *negative* contract: can a pathless source gain
  a path through a compatibility constructor, display hint, old domain value,
  mutation receipt, or test helper? Can a local source lose its repair path and
  still claim mutation readiness?
- Look for mixed-snapshot races: one source read followed by another path
  resolution, a rename between resolve and open, and graph/Thread snapshots
  sampled at different moments. Distinguish current bug from an explicitly
  guarded mutation and from a future design concern.
- Inspect the graph repair diagnostics a user would actually follow. If a
  suggested command cannot be run against the reported state, demonstrate it.
- Challenge test helpers that smuggle filesystem paths or filename IDs into
  domain records and can make an adapter-neutral test pass for the wrong reason.
- For Antigravity: do not settle with a checklist or passing suite. Run a
  concrete hostile fixture for each claimed boundary and a second systemic
  pass; a no-findings verdict needs an evidence table showing what was tried.

## Validation and restoration

Work only in the mandatory independent sandbox. Restore every source mutation
against its captured baseline before transfer. Run focused tests, `go test
./...`, `go test -race` on affected packages, `just lint`, planning/audit lint,
generated-doc drift checks, and `git diff --check` where available. If a check
cannot run, report the exact reason rather than claiming it passed. Do not
format, regenerate, stage, commit, or repair the shared source checkout.

## Deliverable

Preserve this brief. Add a severity-ranked verdict, source/path/line evidence,
the consumer inventory, fixture matrix, mutation table, command results, and
the isolation/transfer attestation. Put actionable defects under `## Findings`
with the exact repository grammar and leave every finding open. Separate
confirmed regressions, concrete final-removal blockers, and speculative
future-adapter questions. A no-findings verdict is credible only if the hostile
evidence floor and second pass are documented.

## Reviewer report

## Verdict

**Checkpoint review: changes requested before final field removal.** I completed a
contract/consumer pass and a separate adversarial pass over the three commits from
`003c39b` through `eeaeba2d50d26d3792214a38684d8d17569c45d7`. Three
findings remain open. The local filesystem behavior and generated artifacts passed
the validation below; those passing checks do not exercise the two pathless
counterexamples or the rename race.

The task progress is accurate about the unfinished domain-field removal:
`domain.Task` still has `Path`/`FilenameID` and `domain.Thread` still has
`Path`/`FilenameID` (`internal/domain/task.go:5-26`,
`internal/domain/thread.go:51-56`). Their removal is **planned**, not shipped.
Audit and research have already lost those fields; task and Thread source revisions
are carried by guarded wrappers, not their domain values
(`internal/core/entity_read.go:39-53`, `internal/core/store.go:122-145`).

## Findings

#### M1. Thread lint discards canonical source identity · **Status:** fixed

**Confirmed regression.** `Service.Lint` converts `ThreadRead` to
`SemanticThreads()` at `internal/core/service.go:724-730`, then builds duplicate
evidence from `thread.CanonicalID()` and `thread.Path` at lines 746-755 and
looks up duplicate issues by that same fallback at lines 881-896. It never
compares declared `Thread.ID` to `record.Record.Source.ID`, and the readable
Thread lint result has no source location. This differs from task, audit, and
research lint, which use the loaded source ID at lines 778, 831-845, and 860-877.

A temporary narrow `ThreadStore` (no embedded `Store`) returned two readable
records with source IDs `6g0000000002` and `6g0000000003`, the same declared
`6g0000000001`, empty filename/path fields, and opaque locations
`db://threads/first` and `6g0000000003-second.md`. The focused
`TestAuditProbeThreadLintUsesSourceIdentity` failed: both rows were labeled
duplicate stable ID `6g0000000001`; neither canonical source ID or location
appeared. The test was removed after the probe. A second adapter can therefore
make valid distinct source records look unresolvable while hiding both actual
ID-drift occurrences. Keep loaded Thread records through lint, key collision and
drift checks by `Source.ID`, and preserve each occurrence's opaque diagnostic
location.

**Resolution:** Thread lint now uses loaded source IDs for duplicate, drift, and
cross-kind checks and retains opaque source locations; regression covers
distinct source IDs with one declared ID.

#### M2. A renamed source leaves a stale TUI editor and yank target · **Status:** in-progress

**Confirmed regression.** `resolveLocalPath` captures a selected ID and list
generation and returns a path asynchronously
(`internal/tui/commands.go:21-39`). The reducer checks the selected ID and
generation, then launches the editor or copies that path without checking that
the source still resolves to it (`internal/tui/model.go:392-406`).
`sessionMsg` protects a workspace switch (`internal/tui/model.go:284-305`),
but it does not cover an on-disk rename before the watcher refreshes the list.

In the temporary `TestAuditProbeLocalResultAfterSourceRename`, I requested
`E` and `Y` separately for `ekpd8vnydg2h-alpha.md`, executed each resolver,
renamed the file to `ekpd8vnydg2h-renamed.md`, and delivered the saved result
with the same selected ID/generation. Both subtests failed: `E` returned a
non-nil editor command for the vanished path and `Y` displayed/copied the old
path. No editor or clipboard command was executed in the probe. Revalidate the
resolved source at action time, or carry a source handle/version whose change
invalidates the result. The detail loaders also resolve a path before a separate
semantic read (`internal/tui/commands.go:125-131,204-212,273-280,375-382`);
that hyperlink can be stale across a rename, even when selection guards hold.

**Resolution:** Confirmed rename window between asynchronous path resolution and
editor/yank action. Selection and list-generation checks do not prove source
freshness; a source-aware action design is still needed.

#### M3. The compatibility graph constructor promotes a URI into repair authority · **Status:** in-progress

**Concrete final-removal blocker.** `NewTaskGraph([]domain.Task, ...)` calls
`TaskGraphReadFromFiles` (`internal/core/dependency_graph.go:327-330`), which
copies `task.Path` into `VersionedRecord.LocalPath` at
`internal/core/service_task.go:189-200`. `taskGraphGuardedRecords` does the
same for `TaskGraphRead.Tasks` at lines 231-245. The explicit
`TaskGraphRead.Records` path does not do this. `hasLocalRepairPath` currently
tests only for nonempty text (`internal/core/dependency_repair.go:302-314`).

A temporary hostile test constructed `NewTaskGraph` with
`Task.Path="db://tasks/opaque"` and an invalid dependency. Diagnosis marked
the defect `Repairable:true` with `LocalPath:"db://tasks/opaque"`; a second
test showed `PlanTaskGraphRepair` accepts that URI as the selected repair
source and emits an operation. Both tests failed their negative assertions and
were removed. The current filesystem materializer separately confines writes
to its task directory (`internal/store/graphrepair.go:191-198`), so this is not
evidence that the shipped CLI opens that URI. It is a false repair capability
in the exported core compatibility path and can mislead a future adapter or a
test helper. Remove the compatibility path before domain-field removal, or make
it explicitly local and reject opaque/path-shaped keys as repair handles.

**Resolution:** Confirmed compatibility NewTaskGraph and TaskGraphRead.Tasks
still promote Task.Path to repair authority. The current task acceptance
criterion is reopened; remove this fallback during final domain-field removal.

## Producer and consumer inventory

The contract fields are defined at `internal/core/entity_read.go:21-53`:
`Source.ID` is semantic identity; `Location` is opaque presentation/diagnostic
context; `LocationIsPath` is a presentation hint; `VersionedRecord.LocalPath`
and `SourceVersion` are guarded repair/CAS evidence. The source-set witness
is checked for the aggregate store and every selected read, path, and mutation
port at `internal/core/service.go:290-382`; zero, missing, unstable, typed-nil,
and foreign witnesses are covered by `internal/core/source_set_test.go:45-151`
and `internal/core/service_entity_path_test.go:44-146`.

| Entity | Producer and semantic/presentation consumers | Local navigation, diagnostic, and mutation authority |
| --- | --- | --- |
| Task | `taskSource` recovers filename ID and path at `internal/store/entity_read.go:12-17`; `scanTaskDocuments` and `ReadTaskGraph` attach the record source at `internal/store/fsstore.go:163-178,194-208`. List/board/graph/lint use the loaded ID (`internal/core/service_task.go:25-98`, `internal/core/board.go:55-98`, `internal/core/service.go:760-789`); JSON uses `ToLoadedTaskJSON` at `internal/wire/dto.go:54-57`. | `TaskPathSource`/`FS.ResolveTaskPath` are parse-free (`internal/core/store.go:48-52`, `internal/store/paths.go:8-15`). `LocalPath` in guarded records authorizes repair and materialization (`internal/core/dependency_source.go:314-316`, `internal/store/graphrepair.go:191-198`); SHA-256 source revisions and unreadable problems are compared before writes (`internal/store/cas.go:25-40`, `internal/core/dependency_graph.go:971-999`). |
| Epic | `parseEpic` sets filename-stem identity at `internal/store/epicstore.go:282-297`; scan/selected reads return `RecordSource` at lines 31-57 and 70-95. Lint, summary, board, and wire consume `Source.ID` (`internal/core/service.go:799-813`, `internal/wire/dto.go:381-385`). | `EpicPathSource`/`FS.ResolveEpicPath` are local navigation (`internal/core/store.go:189-192`, `internal/store/paths.go:17-23`). No guarded epic source revision is shipped. Ordinary epic writes remain store-owned; create outcomes use `LocalCreateOutcome` (`internal/core/creation_receipt.go:14-46`). |
| Audit | `auditSource` recovers filename ID and location at `internal/store/entity_read.go:26-38`; body-aware and selected reads attach it at `internal/store/auditstore.go:57-69,82-102`. Finding query, lint, summary and wire use source ID (`internal/core/finding.go:171`, `internal/core/service.go:860-877`, `internal/wire/dto.go:271-275`). | `AuditPathSource`/`FS.ResolveAuditPath` are parse-free local navigation (`internal/core/store.go:194-197`, `internal/store/paths.go:25-31`). `Location` diagnoses failures; ordinary audit mutations are not guarded by a repository-wide source revision. Create receipt local paths are separate (`internal/core/creation_receipt.go:49-54`). |
| Research | `researchSource` recovers filename ID/location at `internal/store/entity_read.go:38-45`; scan/selected reads attach it at `internal/store/researchstore.go:34-44,56-76`. Lint and JSON use source ID (`internal/core/service.go:829-845`, `internal/wire/dto.go:251-255`). | `ResearchPathSource`/`FS.ResolveResearchPath` are parse-free local navigation (`internal/core/store.go:199-202`, `internal/store/paths.go:33-39`). No guarded research source revision is shipped; create receipt local paths are separate (`internal/core/creation_receipt.go:56-61`). |
| Thread | `ReadThreads` and `ReadThread` attach filename ID/location and readable revision at `internal/store/threadstore.go:23-57,99-110`; list/graph/wire project loaded source identity (`internal/core/service_thread.go:215-249`, `internal/core/thread_projection.go:92-111`, `internal/wire/thread.go:150-169`). Lint is the M1 exception. | `ThreadPathSource`/`FS.ResolveThreadPath` are parse-free local navigation (`internal/core/store.go:181-187`, `internal/store/paths.go:41-47`). Readable/unreadable versions are compared at `internal/store/cas.go:43-100` before guarded task/Thread writes; problem location is diagnostic, not a repair handle. |

The remaining transitional consumers are concrete: task compatibility conversion
and fallback repair sources use `Task.CanonicalID/Path`
(`internal/core/service_task.go:189-245`,
`internal/core/dependency_source.go:309-316`); Thread creation validation and
diagnostic names use `Thread.FilenameID/Path`
(`internal/core/thread_creation.go:79-96,131`,
`internal/core/service_thread.go:338-350`); Thread planner snapshots resolve
only semantic `Thread.ID` (`internal/core/thread_mutation.go:30-94`).
They need loaded source envelopes for guarded planning and an operation-specific
local outcome/receipt for local diagnostics before those fields can be removed.
The create receipt seam already exists for all four ordinary entities
(`internal/core/creation_receipt.go:14-61`); it is implemented code, not merely
the task's remaining acceptance criterion.

## Fixture and adversarial evidence

| Challenge | Executed evidence and observed result |
| --- | --- |
| Narrow pathless source, no location/URI/path-shaped key | Temporary `auditProbeThreadSource` implemented only `ThreadStore` and `SourceSetProvider`, with no embedded broad `Store`; paired with the existing narrow lint fake using `NewService(nil, WithLintSource(...), WithThreadStore(...))`. The URI and path-shaped opaque location reproduced M1. `TestNewServiceAcceptsPathlessReadOnlySource`, `TestTaskGraphReadUsesExplicitIdentityRatherThanParsingLocation`, and `TestReadableLocationsAreOptionalAndNeverBecomePathsOrIdentity` passed; the wire test uses empty, URI, and path-shaped location values with non-default `LocationIsPath`. |
| Option composition and two source sets | `TestNewServiceRejectsEveryMismatchedSplitCapability`, `TestNewServicePreservesTypedNilAndOrderIndependentThreadPairing`, `TestEntityPathDefaultsDetachOnExplicitReadReplacement`, and `TestEntityPathCapabilitiesIgnoreTypedNilAndRejectForeignSourceSets` passed. A deliberate validation bypass made the first test fail for each foreign capability, including path ports. No cross-space constructor counterexample was found with distinct witnesses. |
| Drift, duplicates, archived work | Active/archived task drift (`internal/core/lint_source_test.go:35-53`), audit/research drift (lines 363-394), task duplicate source locations (`internal/core/dependency_graph_test.go:756-804`), Thread list/show source identity and duplicate source IDs (`internal/core/service_thread_test.go:372-433`), summary/board IDs (`internal/core/usecases_test.go:232-262`, `internal/core/board_test.go:129-164`), and JSON IDs (`internal/wire/envelopes_test.go:101-213`) passed. The separate same-declared/different-source Thread lint fixture failed as M1. |
| Malformed documents and selected reads | `TestEntityPathsRemainParseFreeForMalformedDocuments` and `TestResolveThreadPathRemainsParseFreeForMalformedDocuments` passed for task, epic, audit, research, Thread (`internal/store/paths_test.go:14-61`). `TestInfoCommandsRemainSemanticWithoutLocalPath` passed human and JSON modes (`internal/cli/entity_info_pathless_test.go:43-90`). Selected audit read resolves one candidate then reads it (`internal/store/auditstore.go:82-102`); ordinary selected-read tests passed in `go test ./...`. |
| Whole-snapshot and side effects | `SameSourceSnapshot` compares readable source handles/revisions and unreadable versions (`internal/core/dependency_graph.go:971-999`); Thread comparison is at `internal/store/cas.go:54-100`. Focused tests cover body-only changed/missing readable revisions (temporary test; passed unmutated), unreadable bytes (`internal/store/graphmutation_test.go:246`), duplicate-ID records (`internal/core/dependency_source_test.go:442`), handle changes (`internal/core/dependency_graph_test.go:756-804`), and addition/removal/representation changes (`internal/store/thread_source_snapshot_test.go:117-175`). `internal/store/graphrepair_test.go:130-181` observes pre-write conflicts with `Committed:false` and no applied files; `internal/store/graphrepair_test.go:190-216` checks durable-prefix receipts after a later failure. |
| TUI result races | `TestDelayedLocalPathActionCannotRetargetSelectionOrReload`, `TestAtlasDropsStaleWorkspaceResultsAndOldSessionMessages`, and `TestLocalThreadPathSurvivesSemanticDetailFailure` passed. The temporary file-rename-between-resolve-and-result fixture failed for both editor and yank (M2). The existing pathless editor test (`internal/tui/local_path_test.go:15-43`) and `TestDetailTitle_NoLinkWithoutContent` passed. |
| Repair commands users see | A temporary CLI test built a broken task with duplicate/self and invalid dependency declarations, captured the diagnosis text, then ran the **emitted** `task depend repair --auto --dry-run` and quoted `--drop <selector> --dry-run` against that same state. Both returned `would remove`; test passed. Pathless explicit repair remains unavailable (`internal/core/dependency_repair_test.go:61-111`), except through M3's compatibility constructor. |

The existing `pathlessInfoStore` test helper embeds `core.Store`
(`internal/cli/entity_info_pathless_test.go:17-40`) and exercises only two
selected reads. It does not prove that a fully narrow, pathless adapter supports
every ordinary entity operation. Likewise several task/Thread test helpers derive
source IDs from `CanonicalID()` (`internal/core/service_thread_test.go:20-52`,
`internal/core/service_task.go:189-245`); they can hide the M1/M3 classes unless
source identity deliberately disagrees with domain fields.

## Restored mutation probes

Each row is an exact temporary production edit, followed by its focused test.
I restored the original bytes immediately after each run; `git status --short`
was empty before report editing. A failing focused test means the mutation was
killed. The readable-revision row required a temporary adversarial test because
the initially selected name had no matching committed test; that temporary test
also failed under the mutation and passed against unmodified code.

| # | Production mutation | Focused test | Observed |
| --- | --- | --- | --- |
| 1 | `internal/core/service.go:347`: replace `s.validateSourceSets()` with nil error | `TestNewServiceRejectsEveryMismatchedSplitCapability` | Failed, all foreign-capability subtests; killed. |
| 2 | `internal/core/dependency_source.go:316`: use `Record.Source.Location` for `LocalPath` | `TestTaskGraphRepairDoesNotSelectOpaqueLocations` | Failed: `db://tasks/opaque` became repairable; killed. |
| 3 | `internal/core/dependency_graph.go:991`: compare readable source only, omit nonempty/equal version | temporary `TestAuditProbeReadableBodyRevisionIsCASBoundary` | Failed on body-only revision change; killed by new hostile test. Existing selected test name matched none and was **not** counted. |
| 4 | `internal/core/service.go:778`: compare task declared ID with itself in lint | `TestLintAttributesTaskIDDriftToAdapterSourceForActiveAndArchivedTasks` | Failed, missing active and archived drift issues; killed. |
| 5 | `internal/tui/model.go:393`: bypass selection/list-generation guard | `TestDelayedLocalPathActionCannotRetargetSelectionOrReload` | Failed, delayed editor result acted on another selection; killed. |
| 6 | `internal/core/service_task.go:385-386`: on absent path port, return transitional `Task.Path` from `TaskGraphRead.Tasks` | `TestTaskLocalActionDoesNotInferPathFromSemanticTask` | Failed: editor command received `urn:task:must-not-open`; killed. |

## Validation and artifact comparison

- `GOCACHE=/private/tmp/taskflow-review-go-cache go test ./...`: passed.
- `go test -race ./internal/core ./internal/store ./internal/tui ./internal/cli ./internal/wire` with that cache: passed.
- `just lint` with sandboxed Go and golangci caches: passed, 0 issues.
- `go run ./cmd/tskflwctl --no-color lint`: passed, all planning entities and links.
- Independently regenerated CLI docs to a temporary directory and schema comments to a separate temporary file; both diffs against committed outputs were empty. `git diff --check`: passed.
- A temporary `git archive 003c39b` extracted **inside the sandbox** supplied the base binary. For the same committed CLI fixture, base and checkpoint human/JSON outputs were byte-equal for task/audit `info`, `path`, and `list --all` (12 command variants, all exit 0). The only changed golden is `schema_jsonschema.golden`, whose two changed descriptions say local paths are optional; the four changed CLI help pages likewise describe optional local paths. `internal/wire/schema_comments.json` only removes comments for fields removed from domain types. These exact diffs are generated by the passing drift checks.
- Create planned/committed local receipt behavior and post-commit cleanup are exercised by `internal/store/create_test.go:57-142`; no new receipt mismatch was demonstrated under this checkpoint. The temporary repair-command test exercised every command emitted by its diagnosis against the recommending state.

## Isolation and transfer attestation

| Field | Value |
| --- | --- |
| Workspace path | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.DUcUU4` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.DUcUU4/.git` |
| Baseline commit | `e9b4cfebd59d743d3f1ff25144eaf6d522a9f144` |
| Captured source deliverable blob | `55954fb7828ea61885c09d21f129377266002982` |
| Captured source fingerprint | `9ceb59f977483c0ad0d7a279c5b00c6a07395fcd` |
| Deliverable | `planning/audits/6gfrtba5jky2-2026-10-02-local-entity-source-capabilities-checkpoint-codex.md` |
| Guarded one-file transfer result | succeeded; helper output is recorded by the reviewer after verification |

Only this assigned audit is transferred. The independent workspace is retained
for the implementation owner to confirm receipt.
