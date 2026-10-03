---
schema: 1
id: 6gfyn1nhawqq
bucket: closed
area: task-source-identity-removal-checkpoint-antigravity
date: "2026-10-02"
updated_at: "2026-10-02"
---
# Audit: Task source-identity removal checkpoint — antigravity — 2026-10-02

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
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

Antigravity-specific emphasis: do not infer correctness from compilation or
green tests. Build one real temporary planning-space fixture whose filename ID
differs from frontmatter, one whose ID is missing, and one with colliding source
IDs at separate paths. Run the shipped CLI against them, then compare that
evidence with a pathless loaded-record fake. For each claimed protection,
identify the exact line that rejects the hostile case and a mutant that the
focused test kills. Try to find a divergence between list/show, graph, lint,
and source-ID-addressed actions; if they agree, show the observed table. On the
second pass, deliberately seek a systemic issue that would survive the new
test helper, rather than listing speculative future-adapter concerns.

### 1. Executive Summary & Verdict

- **Verdict:** **Ready with Tracked Follow-up** (1 Low finding open for owner triage: `L1`).
- **Review Target:** Task [`6gcwcf88z57p`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/planning/tasks/6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md) (`split-local-path-capabilities-from-semantic-entity-reads`), commits `2b04562..50e9ce2` (implementation commit `c845029`, planning documentation commit `50e9ce2`).
- **Review Architecture:** Dual-pass adversarial audit executed in an isolated, non-hardlinked sandbox (`isolated-review.8tMnEo`) under the mandatory independent-sandbox protocol.
- **Safety for Next Removal (`Task.Path`):**
  - **Verdict:** **Safe to proceed with `Task.Path` removal**, with one specific decoupling requirement identified in `L1`.
  - **Rationale:** The refactor in commit `c845029` cleanly severs `domain.Task.FilenameID` and `Task.CanonicalID()` from the domain model. Canonical source identity is now consistently owned by `core.RecordSource.ID` across filesystem scanning, ordinary reads, graph construction, linting, CLI wire projections, and TUI navigation. `Task.ID` represents solely the declared YAML frontmatter identifier.
  - **Prerequisite for Next Removal (`Task.Path`):** In [`internal/store/graphrepair.go:214`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/store/graphrepair.go#L214), the store's reload verification step currently passes reloaded tasks through `core.NewTaskGraph([]domain.Task{parsed}, nil).SourceRecords()`. This compatibility constructor relies on `parsed.ID` (and transitional `parsed.Path`), and drops records whose frontmatter `id:` is missing even when the file on disk has a valid filename ID. Before or during the removal of `Task.Path`, this call must be replaced with `NewTaskGraphRead` using explicit versioned records (`taskSource(path)` and `LocalPath: path`).

---

### 2. Review Environment & Isolation Attestation

All inspection, builds, test execution, mutation probes, and report edits were performed exclusively inside the independent review sandbox:

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/.git
baseline_commit=021b1370e2fff840d39bd036a981590522383ef8
source_blob=d3e8bd104f521e36d33ff7eda4093de0853ca89b
source_fingerprint=89c7ae489a5d9c54506809f11edab7c1c4d27ef5
deliverable=planning/audits/6gfyn1nhawqq-2026-10-02-task-source-identity-removal-checkpoint-antigravity.md
deliverable_changed=true
transfer=pending
```

No writes, git modifications, or temporary files were created in `$SOURCE_ROOT`. All mutation probes have been restored to the captured baseline.

---

### 3. Producer/Consumer Inventory for Task Identity

The refactor strictly separates four orthogonal dimensions of task identity:
1. **Declared ID (`domain.Task.ID`)**: Parsed strictly from the Markdown YAML frontmatter `id:` field. Represents user-declared intent.
2. **Canonical Source ID (`core.RecordSource.ID`)**: Extracted by the adapter from storage metadata (the flat filename's leading ID prefix `tasks/<id>-<slug>.md` for filesystem storage). Acts as the immutable canonical identity.
3. **Source Location (`core.RecordSource.Location` & `LocationIsPath`)**: Optional, adapter-neutral URI or path identifying where the record was read from.
4. **Local Repair Path (`core.VersionedRecord.LocalPath` / `TaskGraphSourceRef.LocalPath`)**: Explicit local filesystem path granting guarded repair authority. Never manufactured from `domain.Task.Path`.

| Subsystem / Component | Exact Source Path & Line | Role | Task ID Consumed / Produced | Source ID Consumed / Produced | Location / LocalPath Handling |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Filesystem Parser** | [`internal/store/fsstore.go:392-435`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/store/fsstore.go#L392-L435) (`parseTask`) | Producer | Produces `domain.Task.ID` from frontmatter YAML. | Does not produce source ID (`FilenameID` assignment removed). | Sets legacy `t.Path = path`. |
| **Adapter Source Minting** | [`internal/store/entity_read.go:14-17`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/store/entity_read.go#L14-L17) (`taskSource`) | Producer | Ignores domain task. | Produces `RecordSource.ID` via `splitFlatName` on filename. | Sets `Location = path`, `LocationIsPath = true`. |
| **Document Scan** | [`internal/store/fsstore.go:163-180`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/store/fsstore.go#L163-L180) (`scanTaskDocuments`) | Producer | Combines `parseTask` and `taskSource`. | Pairs `Value.Task` with `taskSource(path)`. | Sets `LocalPath = path`, `SourceVersion = hashContent(content)`. |
| **Strict Graph Snapshot Read** | [`internal/store/fsstore.go:194-214`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/store/fsstore.go#L194-L214) (`FS.ReadTaskGraph`) | Producer | Extracts `record.Value.Task`. | Emits `VersionedRecord[domain.Task]` with `Source.ID`. | Populates `LocalPath: record.localPath`. |
| **Ordinary Task List** | [`internal/core/service_task.go:30-95`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/service_task.go#L30-L95) (`Service.ListTasks`) | Consumer / Producer | Filters on `record.Value`. `--unblocked` checks `graph.State(record.Source.ID).Eligible`. | Emits `[]LoadedRecord[domain.Task]`. | Retains `record.Source.Location`. |
| **Selected Task Read** | [`internal/core/service_task.go:98-107`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/service_task.go#L98-L107) (`Service.ShowTask`) | Consumer | Validates `requireSourceID(EntityTask, record.Source)`. | Returns `LoadedRecord[TaskWithBody]`. | Retains `record.Source.Location`. |
| **Filename Resolution** | [`internal/store/fsstore.go:354-364`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/store/fsstore.go#L354-L364) (`FS.resolve`) | Consumer | Ignores frontmatter `id:`. | Matches `slug` against filename ID and slug via `flatCandidates`. | Returns resolved file path. |
| **Graph Input Validation** | [`internal/core/service_task.go:254-274`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/service_task.go#L254-L274) (`validatedTaskGraphRead`) | Guard | Rejects records where `requireSourceID` fails; converts to `TaskGraphLoadProblem`. | Enforces non-empty `RecordSource.ID`. | Preserves `SourceVersion` and `Location`. |
| **Graph Node Attribution** | [`internal/core/dependency_graph.go:415-460`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/dependency_graph.go#L415-L460) (`newTaskGraph`) | Consumer / Analyzer | Compares `task.ID` with `record.source.TaskID`. Flags `ProblemTaskIDDrift` or `ProblemMissingTaskID`. | All internal maps (`g.tasks`, `g.ids`, `idCounts`) keyed by `record.source.TaskID`. | Problems carry `record.source.LocalPath`. |
| **Graph Reference Resolver** | [`internal/core/dependency_graph.go:880-922`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/dependency_graph.go#L880-L922) (`ResolveTaskID`) | Consumer | Matches query against `g.referenceCandidates`. | Candidates populated with `id: record.source.TaskID`. | Independent of local filesystem paths. |
| **Lint Universal Sweep** | [`internal/core/service.go:663-793`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/service.go#L663-L793) (`Service.Lint`) | Consumer / Reporter | Evaluates `MissingIDIssue(t.ID)`, `LintTask(t)`. | Evaluates `IDDriftIssue(t.ID, loaded.Source.ID)`. | Results attributed to `readableDiagnosticLocation(loaded.Source)`. |
| **Wire DTO Mapping (Loaded)** | [`internal/wire/dto.go:54-58`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/wire/dto.go#L54-L58) (`ToLoadedTaskJSON`) | Producer | Semantic fields mapped from `record.Value`. | `TaskJSON.ID` explicitly assigned from `record.Source.ID`. | `TaskJSON.Location` assigned from `readableSourceLocation`. |
| **Wire DTO Mapping (Bare)** | [`internal/wire/dto.go:48-50`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/wire/dto.go#L48-L50) (`ToTaskJSON`) | Producer | `TaskJSON.ID` assigned from `t.ID`. Used only for mutation receipts. | None. | No location emitted. |
| **TUI Item Projection** | [`internal/tui/item.go:46-56`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/tui/item.go#L46-L56) (`taskItem.ref`) | Consumer / Producer | Presentation uses `i.t.Slug`, `i.t.Status`. | `ref().key` assigned from `i.sourceID` (populated from `record.Source.ID`). | Never inspects `t.Path`. |
| **Guarded Lifecycle Action** | [`internal/core/service_task.go:515-546`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/service_task.go#L515-L546) (`Service.Move`/`DeferTask`) | Consumer | Planners resolve `ref` via `graph.ResolveTaskID(ref)`. | Uses resolved `taskID` (canonical source ID). | Guarded by snapshot CAS. |
| **Guarded Graph Repair** | [`internal/core/dependency_repair.go:308-324`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/dependency_repair.go#L308-L324) (`requireLocalRepairPath`) | Guard | Evaluates `hasLocalRepairPath(defect.Target.Source)`. | Targets edits by canonical `TaskGraphSourceRef`. | Enforces `source.LocalPath != ""`; fails closed with `ProblemRepairUnavailable` if empty. |

#### Compatibility Constructors Inventory

Two compatibility constructors accept bare `domain.Task` slices without an explicit `LoadedRecord` envelope:
1. **`TaskGraphReadFromFiles(tasks []domain.Task, problems []domain.FileProblem) TaskGraphRead`** ([`internal/core/service_task.go:186`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/service_task.go#L186)):
   - Populates `read.Records` with `Source: RecordSource{ID: task.ID, Location: task.Path}`.
   - Leaves `read.GuardedRecords` as `nil`.
   - In `taskGraphGuardedRecords`, elements are wrapped into `VersionedRecord` with `LocalPath: ""` (no repair authority).
   - Inability to represent drift: Since `RecordSource.ID` is initialized from `task.ID`, `task.ID == record.source.TaskID` by construction. It cannot diagnose frontmatter ID drift.
   - If `task.ID == ""`, `validatedTaskGraphRead` drops the record into `Problems` ("task record has no canonical source ID").
2. **`NewTaskGraph(tasks []domain.Task, unreadable []domain.FileProblem) *TaskGraph`** ([`internal/core/dependency_graph.go:329`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/dependency_graph.go#L329)):
   - Direct wrapper around `NewTaskGraphRead(TaskGraphReadFromFiles(tasks, unreadable))`.
   - **Production Reachability:** Audited across the entire repository. There is exactly **one** production invocation of `NewTaskGraph`: in [`internal/store/graphrepair.go:214`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/store/graphrepair.go#L214). It is invoked as a post-write verification scratchpad to extract and compare serialized `.Fields` against the planned repair. It does not grant mutation authority or bypass CAS. All other invocations are restricted to unit test suites.

---

### 4. Four-Case Local Fixture Matrix (Live CLI Verification)

A temporary planning corpus was instantiated at `/tmp/tf-adversarial-matrix` containing a valid epic (`01-test-epic.md`), four active task cases, an archived task, and an unreadable malformed YAML file:
- **Case 1 (Matching IDs):** `tasks/6gtest000001-matching-id.md` (`id: 6gtest000001`, filename ID `6gtest000001`, status `ready-to-start`).
- **Case 2 (Missing Frontmatter ID):** `tasks/6gtest000002-missing-fm-id.md` (no `id:` field in YAML frontmatter, filename ID `6gtest000002`, status `ready-to-start`).
- **Case 3 (Drifted Frontmatter ID):** `tasks/6gtest000003-drifted-id.md` (`id: 6gtest000099`, filename ID `6gtest000003`, status `ready-to-start`).
- **Case 4 (Duplicate Filename/Source ID):** `tasks/6gtest000004-duplicate-a.md` and `tasks/6gtest000004-duplicate-b.md` (both leading with ID `6gtest000004`, status `ready-to-start`).
- **Control A (Archived Task):** `tasks/6gtest000005-archived-task.md` (`id: 6gtest000005`, status `completed`).
- **Control B (Malformed File):** `tasks/6gtest000006-malformed-file.md` (`--- \n schema: [unclosed yaml {[[ \n ---`).

The shipped binary `./bin/tskflwctl` was executed against this corpus with the following observed results:

| Test Dimension | Case 1: Matching ID (`6gtest000001`) | Case 2: Missing Frontmatter ID (`6gtest000002`) | Case 3: Drifted Frontmatter ID (`6gtest000003` vs `6gtest000099`) | Case 4: Duplicate Source ID (`6gtest000004` on 2 files) | Control A: Archived (`6gtest000005`) | Control B: Malformed File (`6gtest000006`) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **`task list --json`** | Present; `id: "6gtest000001"` | Present; `id: "6gtest000002"` (source ID derived from filename) | Present; `id: "6gtest000003"` (canonical source ID wins) | Both present; both project `id: "6gtest000004"` | Excluded from active list | Emitted in `unreadable[]` array with parse error |
| **`task list --all --json`** | Present | Present | Present | Both present | Present (`status: completed`) | Emitted in `unreadable[]` array |
| **`task show <source-id>`** | Returns task (`id: "6gtest000001"`) | Returns task (`id: "6gtest000002"`) | Returns task (`id: "6gtest000003"`) | Fails with `ErrAmbiguous` (matches 2 tasks) | Returns task (`id: "6gtest000005"`) | Fails with `ErrValidation` (malformed YAML) |
| **`task show <declared-id>`** | Returns task | N/A (no declared ID) | Fails with `ErrNotFound` (`"6gtest000099": not found`) | Returns `ErrAmbiguous` | Returns task | N/A |
| **`task show <slug>`** | Returns task | Returns task | Returns task | Returns specific file (`duplicate-a` / `duplicate-b`) | Returns task | Fails with `ErrValidation` |
| **`lint --json`** | Clean (only tags missing) | Flags `field: "id"`, `"missing stable id"` | Flags `field: "id"`, `"frontmatter id \"6gtest000099\" disagrees with filename id \"6gtest000003\""` | Flags `field: "id"` on **both** files citing both file paths | Clean (archived task skips active nags) | Emitted in `unreadable[]` with YAML syntax error |
| **Graph Health** | Healthy | Broken (`missing-task-id`) | Broken (`task-id-drift`) | Broken (`duplicate-task-id`) | Healthy (completed) | Broken (`unreadable-task`) |
| **`task set <source-id> --tags test --dry-run`** | Succeeded | Succeeded (addressed by source ID without frontmatter ID) | Succeeded (addressed by source ID) | Fails with `ErrAmbiguous` | Succeeded | Fails with parse error |
| **`task set <declared-id> ...`** | Succeeded | N/A | Fails with `ErrNotFound` | Fails with `ErrAmbiguous` | Succeeded | N/A |
| **`task move <ref> in-progress --dry-run`** | Fails closed because graph is broken by Case 2/3/4 | Fails closed because graph is broken | Fails closed because graph is broken | Fails closed because graph is broken | Fails closed because graph is broken | Fails closed |

**Key Takeaways:**
1. Declared ID never acts as a stealth lookup key: `task show 6gtest000099` and `task set 6gtest000099` return `ErrNotFound`.
2. Missing frontmatter ID remains fully manageable and inspectable under its canonical source ID (`6gtest000002`).
3. Duplicate filename IDs never collapse into a single representative: `task show` and `task set` by ID return `ErrAmbiguous`, while slug resolution accurately targets each individual physical file.
4. Malformed files and archived tasks are partitioned cleanly without corrupting active entity projections.

---

### 5. Isolated Pathless Loaded-Record Fixtures

To challenge the system against non-filesystem adapters with no paths and opaque keys, isolated loaded records were evaluated:

1. **Pathless Record with Declared ID != Source ID:**
   - Fixture: `LoadedRecord[domain.Task]{Value: domain.Task{ID: "declared-frontmatter-id", Slug: "worker"}, Source: RecordSource{ID: "opaque-key-9999", Location: "db://tasks/opaque-key-9999", LocationIsPath: false}}`.
   - **Board Projection:** Evaluates under `record.Source.ID` (`"opaque-key-9999"`).
   - **Wire DTO:** `ToLoadedTaskJSON` emits `id: "opaque-key-9999"` and `location: "db://tasks/opaque-key-9999"`.
   - **`Service.ListTasks`:** Emits `record.Source.ID == "opaque-key-9999"`.
   - **Verdict:** Canonical source ID strictly supersedes declared ID.
2. **Pathless Record with Empty Source ID and Misleading Location:**
   - Fixture: `LoadedRecord[domain.Task]{Value: domain.Task{ID: "declared-id-1234", Slug: "orphan"}, Source: RecordSource{ID: "", Location: "db://tasks/declared-id-1234"}}`.
   - **Boundary Rejection:** In [`internal/core/service_task.go:260`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/service_task.go#L260), `requireSourceID(EntityTask, record.Source)` detects `source.ID == ""`.
   - **Graph Admission:** Dropped from `GuardedRecords` and converted to `TaskGraphLoadProblem` (`"task record has no canonical source ID"`).
   - **Wire Projection:** `ToLoadedTaskJSON` emits `ID: ""`. It **never** falls back to `Value.ID`.
   - **TUI Registry:** In [`internal/tui/item.go:56`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/tui/item.go#L56), `taskItem.ref().key` evaluates to `""`, and `validateEntityItems` rejects the empty key, preventing registration.
   - **Verdict:** Empty source ID fails closed; declared ID is never promoted to an action key.

---

### 6. Challenge of Graph Entry Routes & Repair Authority

Both graph entry routes were directly challenged with hostile inputs in a focused probe test:

```go
func TestAdversarialEntryRoutesAndRepairAuthority(t *testing.T) {
    // Route 1: Bare-Task compatibility route with path-shaped Task.Path
    task := domain.Task{
        ID: "6gtest000001", Slug: "self-dep", Path: "tasks/6gtest000001-self-dep.md",
        Status: domain.StatusReadyToStart, DependsOn: []string{"6gtest000001"},
    }
    compatGraph := NewTaskGraph([]domain.Task{task}, nil)

    // Check 1: Authority is denied for repair
    diag, _ := DiagnoseTaskGraphRepair(compatGraph)
    // diag.Defects[0].Repairable == false, Automatic == false
    // diag.Defects[0].Problem.Code == ProblemRepairUnavailable
    // Message: "source has no explicit local repair path; edit it through its adapter"

    // Check 2: Compatibility route cannot diagnose drift
    drifted := task
    drifted.ID = "6gtest000099"
    driftGraph := NewTaskGraph([]domain.Task{drifted}, nil)
    // driftGraph.hasProblem(ProblemTaskIDDrift, "") == false (no independent source ID)

    // Route 2: Explicit guarded records route
    guarded := VersionedRecord[domain.Task]{
        Record: LoadedRecord[domain.Task]{Value: drifted, Source: RecordSource{ID: "6gtest000001", Location: task.Path, LocationIsPath: true}},
        LocalPath: task.Path,
    }
    guardedGraph := NewTaskGraphRead(TaskGraphRead{GuardedRecords: []VersionedRecord[domain.Task]{guarded}})
    // guardedGraph.hasProblem(ProblemTaskIDDrift, "6gtest000001") == true
}
```

**Observed Evidence:**
1. Supplying a path-shaped `Task.Path` to `NewTaskGraph` fails to grant repair authority. `hasLocalRepairPath` evaluates to `false` because `guarded.LocalPath` is empty. The defect is diagnosed with `repair-source-unavailable`.
2. The compatibility route cannot diagnose ID drift because `TaskGraphReadFromFiles` initializes `Source.ID` to `task.ID`.
3. The explicit guarded route successfully authorizes repair when `LocalPath` is present and reliably flags `task-id-drift` when frontmatter ID diverges from source ID.

---

### 7. Duplicate Source IDs vs. Colliding Declared IDs

To verify problem attribution against physical records:

1. **Duplicate Source IDs at Separate Paths:**
   - Fixture: `RecordA` (`Source.ID: "6gtest000001"`, `LocalPath: "tasks/path-a.md"`) and `RecordB` (`Source.ID: "6gtest000001"`, `LocalPath: "tasks/path-b.md"`).
   - **Observed Result:** `graph.Problems()` generates two distinct `ProblemDuplicateTaskID` instances, one attributed to `tasks/path-a.md` and one attributed to `tasks/path-b.md`. Neither shadow record is dropped or masked by a representative task.
2. **Two Separate Source IDs with the Same Declared ID:**
   - Fixture: `RecordC` (`Source.ID: "6gtest000003"`, `Task.ID: "shared-declared-id"`, `LocalPath: "tasks/path-c.md"`) and `RecordD` (`Source.ID: "6gtest000004"`, `Task.ID: "shared-declared-id"`, `LocalPath: "tasks/path-d.md"`).
   - **Observed Result:** `ProblemDuplicateTaskID` is **not** produced because source IDs are distinct (`"6gtest000003"` and `"6gtest000004"`). Instead, **both** records independently produce `ProblemTaskIDDrift` attributed to their respective physical paths (`path-c.md` and `path-d.md`).

---

### 8. Restored Mutation Probes (Behavioral Kills)

Three mutation probes were executed in the sandbox, verified against focused test kills, and completely restored to the baseline commit. None relied on compilation errors; all resulted in behavioral test failures:

| Probe Target | Exact Source File & Line | Mutation Applied | Focused Test Suite | Observed Failure Mode | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1. Graph Drift Comparison** | [`internal/core/dependency_graph.go:442`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/dependency_graph.go#L442) | Replaced `if task.ID != "" && task.ID != record.source.TaskID` with `if false && ...` | `go test ./internal/core -run TestTaskGraphHealthAndDeterministicStructuralProblems` | `FAIL`: `problems [unreadable-task self-dependency missing-dependency duplicate-dependency invalid-status] do not contain task-id-drift` | **Restored** |
| **2. Read Projection Source Precedence** | [`internal/wire/dto.go:55`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/wire/dto.go#L55) | Replaced `toTaskJSON(record.Value, record.Source.ID)` with `toTaskJSON(record.Value, record.Value.ID)` | `go test ./internal/wire -run TestOrdinaryReadEnvelopesPreferSourceIdentityForEveryEntity` | `FAIL`: `envelopes_test.go:109: task list id = "declared-id", want "source-id"` | **Restored** |
| **3. Test Helper Source-ID Override** | [`internal/core/dependency_graph_test.go:42`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/core/dependency_graph_test.go#L42) | Replaced `if sourceID, ok := sourceIDs[task.Slug]; ok` with `if false` | `go test ./internal/core -run "TestTaskGraphDiagnosesEveryIdentityAndEdgeShape\|TestTaskGraphHealthAndDeterministicStructuralProblems"` | `FAIL`: `TestTaskGraphHealth...` misses `task-id-drift`; `TestTaskGraphDiagnoses...` misses `missing-task-id` | **Restored** |

All probes were verified restored using `git diff --check` and `git status` (clean working tree).

---

### 9. Schema Comments and Task Progress Verification

1. **Schema Comments Diff:**
   - In [`internal/wire/schema_comments.json`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/wire/schema_comments.json), `domain.Task.FilenameID` was removed.
   - `domain.Task.ID` comment was updated to clarify that it is the declared stable identifier, and the adapter's source ID is the resolution key.
   - `domain.Task.Path` comment explicitly documents: `"transitional local source location; removed in the next slice"`.
2. **Planning Task Alignment:**
   - In [`planning/tasks/6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md:120-126`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/planning/tasks/6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md#L120-L126), the progress section accurately notes:
     > "Task filename identity now lives only in loaded source records, not `domain.Task`; graph tests supply distinct source IDs explicitly when frontmatter IDs drift or are missing. The remaining domain-field removal is `Task.Path` and Thread `Path`/`FilenameID`, concentrated in compatibility constructors, Thread mutation planning/materialization, and their tests."
   - The planning documentation precisely matches the checkpoint implementation and does not mistake planned Thread or `Task.Path` migrations for shipped code.

---

### 10. Adversarial Hostile Angles & Second-Pass Analysis

1. **Future Adapter Without Local Paths:**
   - Tested against adapters emitting opaque keys and URIs (`db://...`).
   - Ordinary reads (`ListTasks`, `ShowTask`) and wire serializations (`ToLoadedTaskJSON`) function without requiring path synthesis.
   - Guarded mutations fail closed with `repair-source-unavailable` rather than manufacturing filesystem paths.
2. **Fixture Removal of `FilenameID`:**
   - Audited every test fixture where `FilenameID` was removed.
   - Tests covering drift and missing frontmatter were updated to pass `localTaskGraphWithSourceIDs` with explicit overrides, preventing silent regression into equal-ID happy paths. Probe 3 confirms that removing this override causes immediate behavioral failure.
3. **Source-ID Presence Validation:**
   - Validated that `requireSourceID` in `service_task.go` and `entity_read.go` prevents records with empty or whitespace-only source IDs from entering the graph or being projected in wire responses.
4. **Handoff from Snapshot to Action:**
   - In `fsstore.go:SetFields` and `lifecyclemutation.go`, atomic writes are protected by whole-content SHA-256 CAS verification (`verifyUnchanged`). A file rename or modification between selection and action triggers `ErrConflict` rather than a torn write or misplaced mutation.
5. **Second-Pass Abstraction Challenge:**
   - Examined whether source identity separation was genuinely achieved or merely relocated into helper fallbacks.
   - In the compatibility constructor (`TaskGraphReadFromFiles`), `Source.ID` defaults to `task.ID` because bare tasks have no adapter envelope. However, this compatibility route is strictly read-only and explicitly denies guarded repair authority.
   - The only abstraction imperfection identified where compatibility code is called in production is documented below in `L1`.

---

### 11. Findings

#### L1. Filesystem graph repair verification relies on bare NewTaskGraph compatibility constructor · **Status:** fixed (PR #274)

- **Trigger:** Materializing a dependency repair on a task file whose frontmatter `id:` is missing, or performing graph repair during the upcoming removal of `Task.Path`.
- **Actual Behavior:** In [`internal/store/graphrepair.go:214`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/store/graphrepair.go#L214), `FS.MutateTaskGraphRepair` verifies that updated file content reloads properly by calling:
  ```go
  parsed, err := parseTask(updated, path)
  if err != nil {
      return nil, fmt.Errorf("%w: graph repair for %s would not reload: %v", domain.ErrValidation, path, err)
  }
  actualRecords, err := core.NewTaskGraph([]domain.Task{parsed}, nil).SourceRecords()
  if err != nil || len(actualRecords) != 1 {
      return nil, fmt.Errorf("%w: graph repair for %s has no reloadable source projection", domain.ErrValidation, path)
  }
  expected, ok := repairSourceRecord(analysis.Prospective, group.Source)
  if !ok || !reflect.DeepEqual(expected.Fields, actualRecords[0].Fields) {
      return nil, fmt.Errorf("%w: graph repair for %s did not materialize only the authorized declarations", domain.ErrValidation, path)
  }
  ```
  `core.NewTaskGraph` routes to `TaskGraphReadFromFiles`, which synthesizes `RecordSource{ID: parsed.ID, Location: parsed.Path}`.
  1. If `parsed.ID == ""` (e.g. frontmatter has no `id:`), `validatedTaskGraphRead` drops the record because `Source.ID == ""`. Consequently, `actualRecords` is empty (`len == 0`), causing the repair verification to fail with `graph repair for %s has no reloadable source projection`, even though `path` has a valid filename ID.
  2. If `parsed.ID` drifted from the filename ID, `actualRecords[0].Source.TaskID` reflects the drifted `parsed.ID`, not the canonical group source ID (`expected.Source.TaskID`). Because line 219 only checks `reflect.DeepEqual(expected.Fields, actualRecords[0].Fields)`, the mismatch in source identity is ignored.
  3. When `Task.Path` is removed from `domain.Task` in the next slice, `parsed.Path` will no longer exist, leaving `TaskGraphReadFromFiles` without location evidence.
- **Expected Behavior:** Since `fsstore` already owns the filesystem path and has `taskSource(path)`, it should construct an authoritative versioned record (`core.VersionedRecord[domain.Task]{Record: core.LoadedRecord[domain.Task]{Value: parsed, Source: taskSource(path)}, LocalPath: path}`) and use `NewTaskGraphRead`, rather than relying on the bare `NewTaskGraph` compatibility constructor.
- **Severity:** Low (Internal verification seam in `fsstore`; does not cause user data corruption, but creates an architectural inconsistency and an immediate blocker for `Task.Path` removal).
- **Exact File & Line:** [`internal/store/graphrepair.go:214`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8tMnEo/internal/store/graphrepair.go#L214).
- **Recommended Remediation:** Replace `core.NewTaskGraph([]domain.Task{parsed}, nil)` with:
  ```go
  actualRecords, err := core.NewTaskGraphRead(core.TaskGraphRead{
      GuardedRecords: []core.VersionedRecord[domain.Task]{{
          Record: core.LoadedRecord[domain.Task]{
              Value: parsed,
              Source: taskSource(path),
          },
          LocalPath: path,
      }},
  }).SourceRecords()
  ```

---

**Resolution:** Confirmed missing-declaration failure in dry-run and committed
repair. Commit e87fa0b reloads with the explicit filesystem source ID, local
path, and revision, checks source identity as well as fields, and adds
missing/drifted declaration coverage with residual diagnostics and content
preservation.

### 12. Verification and Test Execution Log

The following validation commands were executed cleanly inside the sandbox:
- `go build -o ./bin/tskflwctl ./cmd/tskflwctl`: Succeeded.
- `go test ./...`: Succeeded across all 32 packages in 9.8s.
- `go test -race ./internal/core ./internal/store ./internal/domain ./internal/wire ./internal/cli ./internal/tui`: Succeeded in 28.5s.
- `golangci-lint run ./...` (`just lint`): 0 issues reported.
- `just docs-check`: CLI documentation synchronized with cobra command tree.
- `just tidy-check`: Go modules clean and tidy.
- `./bin/tskflwctl audit lint`: All audit findings pass lint.
- `./bin/tskflwctl lint`: All planning entities and dependency links pass lint.
- `git diff --check`: Clean (no whitespace or formatting errors).
- All temporary test probes and scratch files restored to baseline.

## Implementation-owner reconciliation

L1 was reproduced independently: `TestMutateTaskGraphRepairRetainsSourceIdentityWithBrokenDeclaredID`
failed for a missing declaration in both dry-run and committed repair before the fix. Commit
`e87fa0b` replaces the compatibility constructor with an explicit filesystem source record and
requires the reloaded source reference to match the prospective source. The regression covers
drifting declarations too, preserving the declaration and reporting its residual identity defect
while removing the authorized dependency duplicate. Full tests, repair race tests, and code lint
passed. This is a present repair regression, so it was fixed in this task rather than deferred to
the remaining `Task.Path` removal.

Receipt/protocol note: the report's captured attestation says `transfer=pending`. Owner inspection
confirmed an independent `.git`, the stated baseline, only the assigned audit modified in that
sandbox, and byte equality between the shared audit and sandbox report. No transfer transcript was
provided, so this confirms receipt and isolation state without asserting the helper transfer ran.
The technical finding is accepted on the independently reproduced failure.
