---
schema: 1
id: 6gfyn1n6wyn6
bucket: closed
area: task-source-identity-removal-checkpoint-codex
date: "2026-10-02"
updated_at: "2026-10-02"
---
# Audit: Task source-identity removal checkpoint — codex — 2026-10-02

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

Adversarially review the Task filename-identity removal checkpoint for task
`6gcwcf88z57p`, `split-local-path-capabilities-from-semantic-entity-reads`.
This is a narrow checkpoint, not final acceptance of that task. `Task.Path`
and Thread `Path`/`FilenameID` remain transitional, and Thread guarded mutation
planning is not yet migrated. Do not report their mere presence as a finding;
show a concrete regression or a missing seam that makes the next removal unsafe.

Do not implement fixes. Preserve this brief, write only your assigned audit,
and leave findings open for implementation-owner triage. Separate newly
introduced failures from pre-existing limitations and speculative design ideas.

## Review target

Freeze on branch `refactor/split-local-entity-source-capabilities` at
`50e9ce2819d86c93ccba4e844c2ec47832c890d6`. Compare
`2b04562..50e9ce2`: implementation commit `c845029` removes
`domain.Task.FilenameID`/`Task.CanonicalID()` and moves relevant callers to
source records; `50e9ce2` updates the planning task. Inspect adjacent code
only as needed to establish the behavior of this delta. The earlier checkpoint
audits in `planning/audits/6gfrtba5jky2-*` and `6gfrtbae3bp4-*` have already
been triaged; do not repeat settled findings without new evidence.

Build a producer/consumer inventory for Task identity across the filesystem
scan, ordinary and guarded reads, graph construction, lint, selected reads,
CLI/wire projections, TUI selection/navigation, and mutation/repair entry
points. Identify any compatibility constructor that still has only a semantic
Task and explain whether it can be reached in an authoritative workflow.

## Intended contract to challenge

1. `Task.ID` is the frontmatter declaration. `RecordSource.ID`, derived from
   the filename for the filesystem adapter, is the canonical source identity.
   A missing or drifting declaration remains diagnosable under that source ID.
2. Ordinary reads and projections must neither manufacture a filename ID in
   `domain.Task` nor silently substitute its declared ID for a missing source
   ID. Selected operations remain addressed by the source ID, including when
   frontmatter is wrong or absent.
3. `TaskGraphReadFromFiles`/bare-Task compatibility is read-only. It cannot
   infer guarded repair authority from transitional `Task.Path`, and its
   inability to represent source/declared-ID drift is not a substitute for an
   authoritative loaded record.
4. Guarded graph diagnosis, duplicate-ID attribution, repair targeting, and
   whole-source-snapshot CAS retain record-level source identity, locality,
   and revision evidence. Public wire/schema output does not acquire hidden
   filename or revision fields.
5. The removal leaves the filesystem user experience and machine contracts
   unchanged for valid tasks. It must not weaken behavior for malformed,
   duplicate, or hand-edited records in either active or archived states.

## Mandatory evidence floor

- Inventory each producer and consumer of Task source ID versus declared ID,
  with exact source path/line evidence. Include source location and optional
  local repair path in the inventory; do not conflate the three.
- Run a four-case local fixture matrix: matching IDs, missing frontmatter ID,
  different valid frontmatter ID, and duplicate filename/source ID. For each,
  check ordinary task list/show, graph health/problem attribution, lint, and
  at least one source-ID-addressed local action. Include an archived task and
  an unrelated malformed file to catch filtering and selected-read mistakes.
- Add an isolated pathless loaded-record fixture with declared ID different
  from `Source.ID`, then one with empty `Source.ID` and a misleading opaque
  location. Check Board/list/wire or TUI as applicable and verify an empty
  source ID never becomes an actionable declared-ID fallback.
- Challenge both graph entry routes: explicit guarded records and read-only
  bare-Task compatibility. Supply a path-shaped `Task.Path` to the latter and
  attempt repair, demonstrating where authority is denied. Do not claim the
  compatibility route diagnoses drift when it lacks an independent source ID.
- Check duplicate source IDs at different paths and two different source IDs
  with the same declared ID. Verify problems and lint remain attributable to
  physical records rather than a representative task or shared path.
- Run at least three restored mutation probes. For each, name the exact
  mutation, the focused test that should kill it, and the observed failure.
  One must target graph drift comparison, one ordinary read/projection source
  precedence, and one test helper or adapter boundary. A compile error is not
  a behavioral kill; use coordinated mutations if needed to make the mutant
  compile and exercise the relevant test.
- Compare generated schema comments and representative machine/human output
  with the base. Verify the planned-but-unfinished `Task.Path`/Thread work is
  accurately represented in the task progress, not mistaken for shipped code.

## Required hostile angles

- Assume a future adapter has no filename, no local path, and an opaque
  record key. Find any caller that would still need a fabricated `Task.ID` or
  local path merely to make an ordinary read work.
- Inspect test fixtures that previously set `FilenameID`. Did removing the
  field quietly turn a drift/duplicate regression into an equal-ID happy path?
  Give an exact test and a failing hostile input, not a vague coverage claim.
- Inspect source-ID presence validation at the boundary and downstream. Could
  an empty, duplicate, or stale source key pass one projection but fail another?
- Explore the handoff from source snapshot to action. A renamed file or changed
  content may alter the mapping; distinguish a guarded CAS failure from a
  best-effort local navigation race and from a new regression in this delta.
- In a second pass, challenge the abstraction itself: are source identity and
  semantic ID genuinely separated, or did the refactor simply move an implicit
  fallback into a shared helper? A no-findings verdict needs concrete negative
  evidence, not only passing repository tests.

## Validation and restoration

Use only the mandatory independent sandbox. Run focused tests, `go test ./...`,
`go test -race` on the affected packages where feasible, `just lint`, planning
lint, and `git diff --check`. Run probes only in the sandbox, restore each to
its captured baseline, and transfer only the assigned audit. Report checks
not run and why. Do not push, commit beyond the sandbox helper's baseline,
stage, or mutate the shared checkout.

## Deliverable

Add a severity-ranked verdict, producer/consumer inventory, fixture matrix,
restored mutation table, exact commands/results, and sandbox attestation.
Actionable findings belong under `## Findings` using the repository's exact
finding grammar; leave statuses open. Explicitly answer: is this checkpoint
safe to build the `Task.Path` removal on, and what must be corrected first?

## Reviewer report

Codex-specific emphasis: trace every route that converts a `domain.Task` into
an authoritative graph or selected task action. In particular, decide whether
`TaskGraphReadFromFiles`, `TaskGraphRead.Tasks`, or `sourceRefForTask` can be
reached from a production guarded mutation or repair path with only a declared
ID. Test any suspected route, including an empty declaration and contradictory
source ID. Inspect the new graph test helper's source-ID override: prove it
does not conceal a missing production assertion. Give a concrete migration map
for removing `Task.Path` next, but make a finding only for a demonstrated flaw
in this checkpoint or a blocking contract gap.

## Verdict

**Safe to build the `Task.Path` removal on this checkpoint.** No new high-, medium-, or low-severity findings were demonstrated in `2b04562..50e9ce2`. No correction to this delta is required first. This is checkpoint approval only: the remaining Task/Thread field removal and Thread guarded-mutation migration are unfinished.

Reviewed branch `refactor/split-local-entity-source-capabilities`, implementation commit `c845029`, and planning commit `50e9ce2819d86c93ccba4e844c2ec47832c890d6`, against the independent captured baseline below. The six compiling mutants all failed behavioral assertions. Eighty base/checkpoint command comparisons produced identical stdout, stderr, and exit codes. Repository tests, affected-package race tests, code lint, planning lint, schema regeneration, and restored focused checks passed.

## Findings

None. The already-triaged findings in `6gfrtba5jky2-*` and `6gfrtbae3bp4-*` were read for scope and disposition; their settled defects are not reopened here.

## Producer and consumer inventory

`Task.ID`, canonical source identity, readable location, and local repair authority remain separate:

- `domain.Task.ID` is declared frontmatter (`internal/domain/task.go:18`); `parseTask` unmarshals it without replacing it from the filename (`internal/store/fsstore.go:394`, `:408`). The transitional `Task.Path` is still assigned at `:435`.
- `RecordSource.ID` is the adapter's canonical key. `RecordSource.Location` is opaque explanatory context, and `LocationIsPath` is a presentation hint (`internal/core/entity_read.go:23`). Neither supplies permission to open a file.
- `VersionedRecord.SourceVersion` and `VersionedRecord.LocalPath` are separate guarded evidence (`internal/core/entity_read.go:47`). Graph source references preserve the independently supplied local path, never infer it from the readable location (`internal/core/dependency_source.go:311`).

| Route | Producer / consumer and exact evidence | Identity, location, and authority behavior |
|---|---|---|
| Filesystem scan | `internal/store/entity_read.go:14`; `internal/store/fsstore.go:163`, `:175` | `taskSource(path)` parses the filename ID into `RecordSource.ID`. The task declaration is retained independently. Scan documents hold the body, content hash, and explicit local path outside the Task. |
| Failed filesystem records | `internal/store/resolve.go:60`, `:68`, `:99`; `internal/core/service_task.go:283` | The scanner captures exact unreadable bytes and recovers source ID/slug from the filename. The compatibility diagnostic conversion receives already-recovered identity; core does not parse the path. |
| Guarded task read | `internal/store/fsstore.go:194`, `:206` | `ReadTaskGraph` supplies `GuardedRecords` from scan documents, including each source ID, local path, and revision. Active and archived tasks participate in the same scan. |
| Ordinary task read | `internal/store/entity_read.go:61`, `:109` | `ReadTasks` unwraps the guarded records into loaded records without revisions or repair handles. It retains the source record rather than manufacturing a Task filename-ID field. |
| Selected show read | `internal/store/fsstore.go:240`, `:258`; `internal/core/service_task.go:98` | Filename candidates resolve the requested source ID/slug before parsing one file. `ShowTask` requires a nonblank supplied source ID; a missing declaration is still showable under the filename ID. |
| Resolution / local navigation | `internal/store/fsstore.go:354`, `:368`, `:382`; `internal/store/paths.go:9`; `internal/core/service_task.go:374` | Candidate identity comes from filenames. `TaskPath` uses the optional path port, independent of successful parsing. Duplicate exact IDs remain ambiguous. |
| Service composition | `internal/core/service.go:290`, `:314`, `:347`; `internal/core/service_task.go:166` | A complete FS supplies its explicit graph read. The aggregate-store fallback consumes `TaskRead.Records`, which already have source identity; it does not wrap bare Tasks. Construction validates capability source sets. |
| Source-ID presence | `internal/core/entity_read.go:118`, `:132`; `internal/core/service_task.go:254` | Presence validation trims whitespace. Invalid explicit records become load problems without a declared-ID fallback; the compatibility Tasks projection is cleared after normalization so it cannot reintroduce a rejected record. |
| Strict graph | `internal/core/dependency_graph.go:335`, `:448`, `:458`, `:464` | Graph nodes use source references. Missing declarations, declaration/source drift, and duplicate source IDs are diagnosed against those IDs and physical record references. |
| List and Board | `internal/core/service_task.go:47`, `:89`; `internal/core/board.go:57`, `:61` | Lists and Board retain loaded records. Unblocked filtering uses `record.Source.ID`; invalid source-less records are removed before either projection. Ordinary rows can remain visible while graph health is broken. |
| Summary / combined status | `internal/core/service.go:515`, `:581`; `internal/wire/envelopes.go:281`, `:369` | Summary retains in-progress loaded records. Per-space and combined working-set JSON both use the retained source ID through `ToLoadedTaskJSON`. |
| Lint input and attribution | `internal/store/lintsource.go:12`; `internal/core/service.go:675`, `:693`, `:780`, `:927`, `:950` | Lint uses body-bearing loaded records, compares declarations with source IDs for active and archived tasks, and joins graph issues by physical record reference rather than ID or path. Location is presentation context. |
| CLI / full wire / column views | `internal/cli/task.go:231`, `:294`, `:333`; `internal/cli/render/render.go:60`, `:66`; `internal/cli/render/columns.go:415`, `:508`; `internal/wire/dto.go:54`; `internal/wire/envelopes.go:43`, `:87`, `:111`, `:124` | List/show/Board/info JSON and the ID column use the loaded source ID. Human list rendering unwraps Tasks for display only. Info resolves its optional local path using the source ID. No filename or revision field was added to the wire contract. |
| TUI selection and navigation | `internal/tui/commands.go:75`, `:81`, `:107`, `:135`; `internal/tui/item.go:46`, `:56`; `internal/tui/entity.go:91`; `internal/tui/nav.go:56`, `:84`; `internal/tui/atlas.go:579` | Rows and detail requests use explicit source IDs. The full family is checked for duplicate/empty keys before view filtering. Locations and semantic Task IDs do not become selection keys. |
| TUI editor / copy path | `internal/tui/commands.go:23`, `:30`, `:40`; `internal/tui/model.go:392`, `:1513` | The optional path capability resolves and then rechecks the selected stable ID. Stale selection/list generations cannot authorize the action. A subsequent external rename remains a best-effort navigation race. |
| Dependency and lifecycle mutation | `internal/core/dependency_operations.go:134`; `internal/store/graphmutation.go:53`, `:58`, `:83`, `:128`; `internal/store/lifecyclemutation.go:47`, `:108`, `:157`; `internal/core/task_lifecycle.go:197` | The FS owns the guarded read and planner invocation. Broken graphs fail before ordinary planning. Materialization checks the explicit graph source path against current ID resolution, then compares source evidence before writing. |
| Broken-graph repair | `internal/store/graphrepair.go:58`, `:96`, `:119`; `internal/core/dependency_source.go:281`, `:225` | Repair reads an explicit complete graph, selects an exact physical source, and retains source identity/locality through prospective edits. Opaque location is only a stale-context check. Whole task/Thread evidence is rechecked before replacements. |
| Whole-snapshot CAS and lock | `internal/core/dependency_graph.go:965`, `:984`, `:989`; `internal/store/cas.go:28`, `:122`; `internal/store/lock_unix.go:21` | CAS includes every readable source reference/revision and unreadable source revision. Missing revision evidence fails comparison. The Unix guard locks the planning root directory via `os.Open(s.root)`/`flock`; there is no invented task lock-file path. |
| Direct edits, fixes, and receipts | `internal/store/fsstore.go:263`; `internal/store/fix.go:312`; `internal/store/create.go:229`; `internal/store/rename.go:176`, `:219`; `internal/wire/dto.go:48`; `internal/cli/render/render.go:253` | Direct field writes select a physical file through filename resolution; lint fix derives the filename ID itself. Creation/rename already carry operation-specific local outcome data. Bare mutation JSON still reflects the returned declaration: this is pre-existing behavior, not an ordinary-read fallback introduced by this checkpoint. |
| Completion compatibility | `internal/cli/completion.go:222` | The FS bare task list is used for slug completion. It is not used to supply an authoritative guarded snapshot. |

### Every bare-Task graph conversion

1. `NewTaskGraph` calls `TaskGraphReadFromFiles` (`internal/core/dependency_graph.go:328`, `internal/core/service_task.go:189`). This explicit read-only compatibility route uses `Task.ID` as its key and `Task.Path` as opaque location. It cannot diagnose drift without an independent source ID and has no local repair path or revision.
2. `TaskGraphRead.Tasks` is normalized similarly at `internal/core/service_task.go:220`. No shipped filesystem producer populates it; authoritative FS reads populate `GuardedRecords`. A caller can inject it as a compatibility read, but that does not give the FS mutation adapter repair evidence.
3. `sourceRefForTask` (`internal/core/dependency_source.go:307`) is reached when `newTaskGraph` is given no source references. Its production caller is `taskGraphFromMap` (`internal/core/dependency_graph_mutation.go:123`), used for prospective behavior checks, including `taskGraphWithTask` and `taskGraphAfterDependencyPlan` (`internal/core/task_lifecycle.go:384`, `:404`). Those graphs are explicitly source-incomplete. They are not reinstalled as authoritative snapshots or repair inputs.

Ordinary dependency/lifecycle planners start from a graph whose declaration/source equality has already been checked. Repair instead preserves explicit source references during simulation. Searches for `NewTaskGraph`, `TaskGraphReadFromFiles`, `newTaskGraph`, `sourceRefForTask`, and store `ListTasks`/`GetTask` callers found no production FS guarded-mutation route receiving only a declared ID.

## Local fixture matrix and output comparison

Retained reproducible driver: `$E/matrix.py`; raw command arguments, stdout/stderr, and exit codes: `$E/matrix.json`, where `$E` is the sandbox's `.git/review-evidence` directory. Fixtures use real id-led files and explicit planning config, not domain-value helpers.

Every main case includes completed task `6g0000000003-archived.md` and unrelated malformed task `6g0000000099-broken.md`. The drift case also gives the archived task a contradictory declaration. `matching-clean` removes only the unrelated malformed file to establish healthy behavior.

For each case the driver ran `task list`, `task list --all`, `task show <source-id>`, `board`, `task depend repair`, `lint`, `task info <source-id>`, and `task path <source-id>` with full JSON, then human list/show/Board/repair/lint. It compared both binaries before running a real `task set <source-id> --description 'Changed by canonical source'` and a source-ID-addressed local path read. Mutated fixture contents were restored.

| Fixture | Ordinary reads / filtering | Graph and lint attribution | Source-addressed action |
|---|---|---|---|
| Matching, clean | List/show/Board/info/path all exit 0; archived task appears only with `--all` | Healthy graph; planning lint exit 0 | Actual field write and path lookup exit 0 |
| Matching + malformed unrelated file | Lists and Board retain valid tasks and report the broken source with exit 11; selected show/info/path still exit 0 | Broken graph attributes unreadable source to `6g0000000099` and its exact path | Actual write on `6g0000000001` succeeds despite unrelated malformed file |
| Missing declaration | List/show/info still publish source ID `6g0000000001`; the source file contains no `id:` declaration | `missing-task-id` belongs to source `6g0000000001`, with the target's physical path; lint flags missing declaration | Actual selected field write and local path lookup succeed under source ID |
| Different valid declaration | Source `6g0000000001` declares `6g0000000002`; ordinary selected output uses the source ID | Active and archived drift problems name their own source IDs and paths; lint checks both | Actual selected field write succeeds under source ID, preserving the contradictory declaration |
| Duplicate source ID | Two filenames begin `6g0000000001`; list retains both; selected show/info/path by shared ID exit 13; show by `shadow` succeeds | Two duplicate problems retain both paths. Only `shadow` receives its `bad-shadow-only` dependency defect in graph and lint | Shared-ID write is refused with ambiguity; the emitted exact-path repair commits only the shadow file |
| Different source IDs, same declaration | Sources `6g0000000001` and `6g0000000005` both declare `6g0000000001`; reads retain both source IDs | Drift belongs to source `6g0000000005`, not a false duplicate-source diagnosis; only `second` owns `bad-second-only` | Source-ID write succeeds; exact-path repair commits only `6g0000000005-second.md` |

**Comparison result:** all 80 read/diagnosis command comparisons were byte-identical across the base and checkpoint, including error envelopes and human output. `$E/matrix-differences.json` is `[]`. The fixture epic was given its required priority before the final run; the clean control's final lint result is exit 0.

### Emitted repair commands actually executed

`$E/repair_commands.py` tokenized the complete emitted human commands, ran them against the exact state recommending them, recorded results in `$E/repair-commands.json`, and restored each state:

- Duplicate case: `task depend repair --drop '<absolute shadow path>:depends_on=bad-shadow-only#0'` exited 0 and changed only `tasks/6g0000000001-shadow.md`. Identity duplication and the unrelated malformed file remained honestly residual; final health was broken.
- Shared-declaration case: the analogous emitted command for `bad-second-only#0` exited 0 and changed only `tasks/6g0000000005-second.md`. Declaration drift and the malformed file remained residual.
- Missing declaration: the emitted `lint --fix` advice was executed. It restored `id: 6g0000000001`; exit 11 correctly reflected the unrelated unreadable file still present.

No speculative repair command is presented as validated. No new repair recommendation is added by this audit.

## Pathless fixtures and adversarial second pass

The isolated probe sources are retained in `$E/probes/core_audit_checkpoint_probe_test.go` and `$E/probes/wire_audit_checkpoint_probe_test.go`. They were temporarily placed in their matching internal packages, run, and removed before verification. They are reviewer evidence, not committed regression tests or shipped adapter implementations. Logs are `$E/pathless-and-compatibility.log` and `$E/focused-restored.log`.

| Challenged pattern | Hostile input and observed result |
|---|---|
| Shared helper could erase declaration/source separation | `TestAuditCheckpointPortableIdentity` builds explicit loaded records without the graph helper: declaration `6g0000000002`, source `6g0000000001`, empty Task path, misleading opaque location. Board/list/show retain the source and original declaration independently. The graph reports source-ID drift. Both ordinary and guarded-record branches pass. |
| Empty key could become a declared-ID action | The same probe uses empty and whitespace-only source IDs, plus a misleading location containing another ID. List/Board omit the record and retain a source-less diagnostic; selected show returns `ErrValidation`; no graph node or declared-ID eligibility appears. A missing declaration with a valid source ID instead remains diagnosable as `missing-task-id`. |
| Readable location could become navigation/repair authority | Pathless selected records cannot obtain `TaskPath`. Wire probes exercise opaque and path-shaped locations with `LocationIsPath` false and true. The true hint suppresses redundant location presentation; it grants no repair handle. Existing `TestTaskGraphRepairDoesNotSelectOpaqueLocations` also tests that true-hint branch. |
| Wire fallback or optional branches could escape schema validation | `TestAuditCheckpointPortableWireSemanticValidation` checks task list/show/Board semantic IDs, including a source-less direct DTO input whose ID stays absent. All variants are validated with the JSON Schema semantic validator. Source-less records never reach these ordinary service projections as rows. Hidden filename/revision/local-wrapper fields are absent. |
| Bare Task path could grant repair authority | `TestAuditCheckpointCompatibilityCannotAuthorizeRepair` supplies `Task.Path = tasks/6g0000000001-contradictory.md` while declaring `6g0000000002`. Both compatibility routes report `repair-unavailable`; they do not invent drift diagnosis from the path. An explicit local-path selector returns `ErrNotFound`, because there is no authorized path match. Existing compatibility tests also prove automatic plans contain zero operations. |
| Representative reconstruction could masquerade as authoritative evidence | The probe calls `taskGraphFromMap` with contradictory keyed semantic data. `SourceRecords` and repair diagnosis refuse it with `ErrValidation`; whole-snapshot self-comparison fails. Existing `TestTaskGraphSourceQueriesRejectRepresentativeOnlyDerivedGraphs` passes. |
| Graph test helper could replace hostile drift/missing fixtures with happy paths | The changed helper defaults source ID to the declaration, but the two changed hostile tests explicitly override it (`internal/core/dependency_graph_test.go:35`, `:72`, `:112`). Disabling that override makes both tests fail. Independently, the committed real-filesystem drift test at `internal/store/dependency_source_test.go:58` fails when the production adapter substitutes declarations, even when both scan and selected-read sites are mutated together. |
| Equal IDs or locations could collapse physical attribution | Real-file duplicate and shared-declaration fixtures retain separate targets. Committed `TestLintRecordAttributionDoesNotCollideOnIDOrLocation` (`internal/core/lint_source_test.go:305`) additionally supplies equal opaque locations and duplicate source IDs; its shadow-only issue remains attached to the physical occurrence. `TestTaskGraphSourceSnapshotCASIncludesEveryDuplicateIDRecord` (`internal/core/dependency_source_test.go:442`) includes every duplicate's revision. Both pass. |
| Snapshot-to-action change could evade the guard | Focused raw-edit probes `TestMutateTaskGraphCASRejectsRawEditBeforeApply`, `TestMutateTaskGraphRepairRejectsLateReadableNonTargetTaskByteChange`, and `TestMutateTaskGraphRepairRejectsLateUnreadableTaskByteChange` pass. The adapter's whole-snapshot CAS includes non-target and unreadable records, not only the selected target. |
| Navigation rename could be confused with guarded CAS | `TestLocalPathActionFollowsRenameByStableID`, `TestLocalPathActionReportsMissingStableIDAfterRename`, and `TestDelayedLocalPathActionCannotRetargetSelectionOrReload` pass. TUI navigation re-resolves a stable ID after its first lookup. Opening an external editor after that final lookup cannot be atomic against a later external rename; this is the pre-existing limit documented in task progress, not a regression in this delta. |

The broader fake-store helper at `internal/core/service_epic_test.go:181` still synthesizes test source IDs for fixtures with no declaration. That helper is test-only. It was not used to establish the explicit-empty-source or filesystem-drift conclusions above. An actual remote adapter is not shipped or claimed here.

## Restored mutation probes

The exact replacements and focused commands are retained in `$E/mutations.json` and executable driver `$E/mutations.py`. Each mutated file was restored byte-for-byte in a `finally` block before the next probe. All six compilations succeeded; every observed kill was a test assertion failure, not a compile error.

| Mutant | Exact mutation | Focused tests and observed failure |
|---|---|---|
| Graph drift comparison | At `internal/core/dependency_graph.go:458`, replace `task.ID != record.source.TaskID` with `task.ID != task.ID` | `TestTaskGraphHealthAndDeterministicStructuralProblems` fails because problems omit `task-id-drift`; `TestAuditCheckpointPortableIdentity` fails its independent comparison assertion for both loaded branches. Exit 1. |
| Ordinary wire precedence | At `internal/wire/dto.go:55`, change `toTaskJSON(record.Value, record.Source.ID)` to `toTaskJSON(record.Value, record.Value.ID)` | `TestOrdinaryReadEnvelopesPreferSourceIdentityForEveryEntity` reports `task list id = "declared-id"`; the isolated wire semantic probe reports the contradictory declared ID. Exit 1. |
| Graph test-helper boundary | At `internal/core/dependency_graph_test.go:43`, replace assignment of the explicit source-ID override with `_ = sourceID` | `TestTaskGraphHealthAndDeterministicStructuralProblems` loses drift; `TestTaskGraphDiagnosesEveryIdentityAndEdgeShape` loses `missing-task-id`. Exit 1. |
| Coordinated filesystem adapter boundary | At both `internal/store/fsstore.go:175` and `:258`, replace `taskSource(path)` with a source whose ID is the corresponding semantic Task's declaration, retaining path/hint | `TestFilesystemTaskReadsKeepSourceIdentityDespiteFrontmatterDrift` fails its guarded source/declaration assertion at `dependency_source_test.go:83`. Both producer sites were mutated so a neighboring read could not accidentally preserve the invariant. Exit 1. |
| Empty-source fallback | Before validation at `internal/core/service_task.go:260`, set blank source IDs to the Task declaration | `TestEmptyExplicitTaskSourceIDNeverBecomesEligible` reports a healthy actionable node; the isolated portable-identity probe reports manufactured list identities in all empty/blank-source branches. Exit 1. |
| Compatibility path promotion | At `internal/core/service_task.go:235`, add `LocalPath: record.Value.Path` while wrapping ordinary compatibility records | `TestCompatibilityTaskGraphsDoNotPromoteSemanticPathToRepairAuthority` and the isolated compatibility probe report repairable semantic paths, including URI and path-shaped inputs. Exit 1. |

After restoration, the focused run passed with 36 passing test/subtest entries and zero failures. The standalone pathless/compatibility run passed with 13 passing entries and zero failures. The isolated probe files were preserved under `.git/review-evidence/probes`, then removed from the tracked tree.

## Validation and artifact comparison

Commands ran in the independent sandbox. The locally built CLI is `$E/tskflwctl-current`; the comparison CLI is `$E/tskflwctl-base`, built from `git archive 2b04562` into `$E/base`.

| Command / evidence | Result |
|---|---|
| `GOCACHE=/private/tmp/taskflow-codex-go-cache go build -o /private/tmp/taskflow-codex-audit-current ./cmd/tskflwctl` followed by `go test ./...` with the same cache | Exit 0; all packages passed. The CLI was subsequently rebuilt with `-buildvcs=false` into the sandbox evidence directory for the comparisons. |
| `GOCACHE=/private/tmp/taskflow-codex-go-cache go test -race ./internal/core ./internal/store ./internal/cli/... ./internal/tui ./internal/wire` | Exit 0; every affected package passed. |
| `GOCACHE=/private/tmp/taskflow-codex-go-cache just lint` | Exit 0, `0 issues.` The linter emitted cache-persistence permission warnings against its user cache; these were not lint findings. |
| Current CLI `--no-pager --no-color lint` and `audit lint 6gfyn1n6wyn6` | Exit 0, all planning entities/dependency links and audit findings pass. Repeated after report authoring. |
| `python3 .git/review-evidence/matrix.py` | Six real-file fixture variants, 80 identical base/current read and diagnosis comparisons; required malformed, archived, duplicate, and contradictory cases exercised. |
| `python3 .git/review-evidence/repair_commands.py` | Both emitted exact-path repairs committed only their intended files; missing-ID lint fix was executed; fixture states restored. |
| `python3 .git/review-evidence/mutations.py` | Six behavioral kills, each exit 1; all changed production/test files restored. |
| `go test ./internal/core ./internal/wire -run '^TestAuditCheckpoint' -count=1 -v` with writable cache | Pathless, compatibility, and non-default wire/schema probes passed after correcting the probe's expected local-selector denial to allow `ErrNotFound`. No implementation fix was made. |
| Restored focused `go test ./internal/core ./internal/store ./internal/tui ./internal/wire -run ... -count=1 -v` | Exit 0; exact command is retained in `$E/focused-command.txt`, with results in `$E/focused-restored.log`. Includes graph attribution/CAS, raw-edit races, renamed TUI paths, real schema-output validation, and schema-comment drift checks. |
| Base/current `schema --json-schema` and `schema --json` | Both outputs are byte-identical. Captured as `$E/json-schema-{base,current}.json` and `$E/command-schema-{base,current}.json`; schema revision remains `1.80`. |
| `GOCACHE=/private/tmp/taskflow-codex-go-cache go run ./internal/tools/schemacomments` | Generated 268 comments and no working-tree delta. The committed comment delta removes Task's FilenameID entry, describes ID as declared, and documents transitional Path. No wire field or schema revision change. |
| `git diff --check`; helper `verify`; final assigned-audit diff inspection | Passed; only the assigned audit differs from the captured baseline before transfer. |

The first build attempt using the default Go cache failed with `operation not permitted`; the explicit writable cache resolved the build failure. No mandatory validation was omitted. Interactive external-editor behavior after its final path lookup was not treated as an atomic filesystem guarantee; it remains the documented navigation limit.

## Migration map for removing `Task.Path` next

1. Remove the parser assignment at `internal/store/fsstore.go:435` while preserving `taskSourceDocument.localPath`, source location/hint, and `VersionedRecord.LocalPath`. These adapter-owned values already exist independently.
2. Remove the creation assignment at `internal/store/create.go:229`; retain `TaskCreationReceipt.Local` and the existing operation-specific rename paths (`internal/store/rename.go:219`). Ordinary info and TUI actions already request optional path ports.
3. Remove Task-derived diagnostic location from `TaskGraphReadFromFiles`, the bare `Tasks` normalization (`internal/core/service_task.go:196`, `:222`), and `sourceRefForTask` (`internal/core/dependency_source.go:308`), or require an explicit loaded envelope when that context matters. Preserve their read-only/source-incomplete restrictions; no path-shaped semantic value may become a local repair handle.
4. Update local graph fixtures to supply source location and repair path separately. The explicit source-ID override remains necessary for drift/missing-declaration cases. Preserve the real-filesystem identity regression and physical-record attribution/CAS assertions.
5. Complete the separate Thread guarded-record migration before removing Thread's transitional fields. `internal/domain/thread.go:53`, `:56`, `:73` still define Path/FilenameID/CanonicalID, and `internal/store/threadmutation.go:134` still materializes from `before.Path`. Their presence is planned unfinished work, not this audit's finding.

Task progress at `planning/tasks/6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md:114` and `:122` accurately describes that remaining work. Acceptance criteria at `:57` and `:64` are still unchecked. Passing the Task CAS checks here does not complete the task's broader final Task/Thread acceptance criterion.

## Isolation and guarded transfer attestation

| Helper field | Captured value |
|---|---|
| Workspace | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.nIu1dV` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.nIu1dV/.git` |
| Sandbox baseline commit | `cbf9c8219b178923beec188bd1ffa2a2d685ac09` |
| Captured source audit blob | `a9ca121d9929c46c860a21e9140ff3f4771c29e4` |
| Captured source fingerprint | `89c7ae489a5d9c54506809f11edab7c1c4d27ef5` |
| Deliverable | `planning/audits/6gfyn1n6wyn6-2026-10-02-task-source-identity-removal-checkpoint-codex.md` |
| Final verification | Helper `verify` exit 0; independent in-tree `.git`, baseline HEAD unchanged, no staging, no unrelated changes, source deliverable unchanged. |
| Transfer result | `succeeded`; final helper attestation retained in the sandbox at `.git/review-evidence/transfer.log`. |

Creation used the required `scripts/isolated-review-workspace.sh create --source ... --deliverable ... --print-path`. Its sole sandbox baseline commit captures both supplied audit briefs and the handoff state. Builds/tests/probes/report edits were confined to the independent workspace and temporary build caches/binaries; no implementation changes or probe files were transferred. The report was authored with the Taskflow audit CLI, and the final copy uses only the helper's guarded one-file transfer. No push, reviewer commit, branch switch, or staging was performed.

The workspace and its evidence are retained pending implementation-owner confirmation of receipt. The final helper verification and transfer transcripts are `.git/review-evidence/verify.log` and `.git/review-evidence/transfer.log` in that workspace.

## Implementation-owner reconciliation

The no-findings verdict applies to the captured checkpoint and the probes above. Antigravity's
[L1](6gfyn1nhawqq-2026-10-02-task-source-identity-removal-checkpoint-antigravity.md) identified a
production `NewTaskGraph` call in repair reload verification that this inventory missed. Owner
reproduction confirmed that dependency deduplication failed for a file with a valid filename ID and
missing frontmatter ID, in both dry-run and committed repair. Executing `lint --fix` before repair
normalizes that declaration and therefore does not test this combination.

Commit `e87fa0b` reloads the repaired value with the filesystem source record and checks source
identity alongside declaration fields. The regression also covers drifting declarations, residual
identity diagnostics, preservation of unrelated source content, and dry-run bytes. Full tests,
focused repair race tests, and code lint passed. There are no unresolved findings in this audit;
the broader Task/Thread migration remains tracked by task `6gcwcf88z57p`.
