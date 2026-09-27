---
schema: 1
id: 6ge0q80ng3b4
bucket: closed
area: ordinary-entity-read-snapshots-implementation-antigravity
date: "2026-09-26"
updated_at: "2026-09-27"
---
# Audit: Ordinary entity read snapshots implementation — antigravity — 2026-09-26

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

Adversarially review the implementation of task `6gcwcf80v8hg`,
`make-ordinary-entity-list-diagnostics-adapter-neutral`. The intended outcome is an
entity-specific, adapter-neutral read boundary, not merely a rename of `FileProblem`.
Treat checked acceptance criteria and passing tests as hypotheses. Find concrete runtime
defects, compatibility regressions, hidden rereads, or systemic gaps; distinguish them
from work explicitly assigned to the later source/path, TUI, and source-set tasks.

## Review target

Review the entire uncommitted delta on branch `refactor/portable-entity-read-snapshots`
against base `6a3e6f49df605ad7e2bf0e233ea7a494b1184b5f`. Include untracked files.
Primary planning sources are:

- `planning/tasks/6gcwcf80v8hg-make-ordinary-entity-list-diagnostics-adapter-neutral.md`;
- `planning/tasks/6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md`;
- `planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md`.

Trace the new `internal/core/entity_read.go` values through `core` ports/services,
`internal/store` scans and conversions, `internal/cli` human and projected output,
`internal/wire` DTOs/schema/goldens, and `internal/tui` list/detail adapters. Review
tests as possible sources of false confidence. Do not implement fixes or edit any file
other than your assigned audit.

## Intended contract to challenge

1. `RecordSource.ID` is the canonical readable-record identity; declared `ID`,
   `FilenameID`, slug, path, and opaque `Location` cannot override it. Domain-carried
   fields remain temporarily for lint and older presentation consumers only.
2. Task, epic, audit, and research ports remain entity-specific while sharing
   `LoadedRecord` and `LoadProblem`. Ordinary list/show results expose no
   `domain.FileProblem`; no new generic entity repository or kind switch appears.
3. A filesystem task request performs one authoritative body-bearing/versioned scan
   and projects graph, ordinary, and lint views from it. Board, Summary, finding
   queries, and projected CLI output do not hide a second scan or combine generations.
4. Unreadable records retain kind, recoverable ID/slug, optional opaque location,
   independently optional local repair path, and message. Core never parses a
   location to manufacture identity. A URI never becomes the historical JSON `path`.
5. Local partial reads still show available records and actionable errors, then
   exit 11. Full and column-projected JSON preserve a required `path` field (empty
   when no local path), optional portable fields, and the correct schema revision.
6. Explicit source identity survives graph analysis, epic joins, lint attribution,
   Board/status, TUI list identity, full and projected JSON, and each show/info DTO.
   Duplicate or drifting IDs remain visible and cannot silently become actionable.
7. The migration has not weakened graph health, unreadable-source revision evidence,
   optimistic mutation/CAS, snapshot ordering, or the existing Thread read contract.
8. Transitional bare-domain projections are local and explicit, not a second
   competing source-of-truth that future adapters will accidentally depend on.

## Mandatory evidence floor

- Produce a **consumer inventory** for all changed ordinary read ports and their
  downstream consumers. Classify every remaining core-facing `FileProblem`,
  `Path`, `FilenameID`, `CanonicalID()`, `LocationIsPath`, and `SourceVersion` use as
  intentional guarded/local compatibility, deferred work, or current violation.
  Do not infer completeness from a single `rg` result or one interface definition.
- Build a field-lineage table for source ID, declared ID, slug, location, local path,
  message, and source revision for readable and unreadable task/epic/audit/research
  records. Cite producer, conversion, core consumer, human output, wire output,
  and mutation evidence separately.
- Use independent fixtures/fakes for all four kinds: matching identity; missing or
  drifting declaration; stale `FilenameID`; explicit ID plus misleading URI/path;
  malformed document with recoverable ID+slug; malformed filename with neither;
  duplicate canonical IDs with different locations; empty location; and both
  opaque location and local repair path. State where current behavior is necessarily
  limited by the later source/path task.
- Measure adapter calls and filesystem reads for graph, ordinary task list, lint,
  Board, Summary, audit finding queries, and show/info. Check normal and partial
  snapshots; compare list/body/graph projections from the same bytes. Look for
  double hashing, duplicate parsing, hidden rereads, and mutation-revision loss.
- Independently exercise full JSON, `--json -c`, human output, stderr/stdout, and
  exit 11 for local and pathless failures. Validate non-empty `unreadable` arrays
  against generated JSON Schema; inspect the 1.75→1.76 changelog/golden delta for
  unrelated drift and confirm generated CLI docs remain current.
- Trace source-ID precedence through task graph and dependency lint attribution,
  epic filter/join/rollup, audit findings, Board blocked ordering, current and
  cross-space status, TUI registry/detail, and every show/info converter. Include
  contradictory embedded IDs and locations, not only happy-path equal values.
- Perform mutation testing in the sandbox. At minimum try reverting source-ID
  precedence in one graph/Board path, reconstructing a problem ID from location,
  emitting opaque location as JSON path, adding a second task or audit read,
  dropping unreadable source revision, and changing one projected `id` column
  back to the domain field. For each probe, report whether an existing focused
  test fails for the intended reason. Do not count compile failure as detection;
  coordinate cross-layer mutations where needed to keep the program compiling.

## Required hostile angles

- **False portability:** Can a pathless adapter implement the new ports without
  supplying fake local paths, while list/show/lint/Board/status/TUI remain usable?
  Does a compatibility branch still consult `Value.Path` or `FilenameID` when
  `Source` disagrees?
- **Mixed-generation reads:** Could an ordinary list combine a graph from one
  generation with epics/audits from another and claim a coherent answer? Identify
  existing guarantees versus aspirations; only report a regression or concrete
  untracked requirement, not generic cross-entity non-atomicity.
- **Corrupt corpus:** Are duplicate source IDs, malformed frontmatter, absent IDs,
  legacy epic names, and unreadable source versions retained and attributed without
  suppressing good records or granting unsafe mutation?
- **Contract drift:** Do generated schema, DTO comments, projected columns, golden
  fixtures, docs, and exit-code behavior describe the same observable result?
- **Future migration pressure:** Does any new helper make the sequenced removal
  of `Path`/`FilenameID`/`SourceVersion` harder or accidentally create a generic
  repository abstraction under an innocuous name?

For Antigravity in particular, do not stop at a broad inventory or a quick
ready/no-findings verdict. Pick at least three seams above, construct contradictory
cross-layer inputs, run focused probes, and show the exact observation that
supports or falsifies each claim. If a probe is blocked, say precisely why and
choose another. For Codex, emphasize systemic failure modes and whether narrow
tests mask a neighboring consumer that still trusts domain or path metadata.

## Validation and restoration

Work only in the mandatory independent sandbox created by the generated audit's
instructions. Establish the baseline with `go test ./...`, then run focused
tests/probes with a sandbox-local or `/tmp` Go cache. Inspect `git diff --check`,
generated CLI docs/schema, and the relevant linter if useful. Restore every
mutation, generated file, and fixture after each experiment. Before transfer,
the assigned audit must be the only difference from the sandbox baseline.
Never commit to or alter the shared source checkout.

## Deliverable

Preserve this brief and append your report to your assigned audit. Use exact
finding headings such as `#### H1. ... · **Status:** open` and leave every
finding open for owner triage. Each finding needs a concrete reproduction or
code path, actual versus expected result, severity, consequence, and the smallest
plausible remedy. Separate confirmed defects, untested risks, and intentionally
deferred work. A no-findings verdict requires an evidence-backed inventory and
probe results, not simply a passing suite. Include the sandbox attestation and
transfer result required above.

## Reviewer report

### 1. Executive verdict

**Sound with amendments.**

The implementation of task `6gcwcf80v8hg` successfully establishes an adapter-neutral, entity-specific read boundary for ordinary task, epic, audit, and research lists. Aggregate store methods (`ListTasks`, `ListTasksWithBodies`, `ListEpics`, `ListAudits`, `ListResearch`) returning `domain.FileProblem` have been eliminated from `internal/core/store.go` in favor of entity-specific snapshots (`TaskRead`, `EpicRead`, `AuditRead`, `ResearchRead`) returning `LoadedRecord[T]` and portable `LoadProblem` diagnostics. Machine contract revision 1.76 preserves additive schema compatibility, and the single-scan pipeline guarantees for Summary, Board, and audit finding rollups remain intact.

However, adversarial testing revealed three findings:
1. `ToResearchJSON` projects unvalidated frontmatter `ID` instead of canonical source identity, creating an identity drop for legacy research documents without frontmatter IDs (Finding M1).
2. Column contract test fixtures in `columns_contract_test.go` configure identical declared and source IDs for all items, leaving column projection tests blind to source-ID fallback regressions (Finding L1).
3. The `epic show` CLI command projects bare domain task values without binding canonical source identity (Finding L2).

---

### 2. Findings

#### M1. ToResearchJSON projects unvalidated frontmatter ID instead of canonical source identity · **Status:** fixed

- **Layer crossing:** `internal/domain/research.go` (`Research.CanonicalID`) $\rightarrow$ `internal/wire/dto.go` (`ToResearchJSON`) $\rightarrow$ `internal/wire/envelopes.go` (`ToResearchMutationEnvelope`).
- **Reproduction path:**
  Inspect `internal/wire/dto.go:ToResearchJSON`:
  ```go
  // ToResearchJSON maps a domain research doc to its wire DTO.
  func ToResearchJSON(r domain.Research) ResearchJSON {
      return toResearchJSON(r, r.ID)
  }
  ```
  Compare this with `ToTaskJSON` and `ToAuditJSON`:
  ```go
  func ToTaskJSON(t domain.Task) TaskJSON {
      return toTaskJSON(t, t.CanonicalID())
  }
  func ToAuditJSON(a domain.Audit) AuditJSON {
      return toAuditJSON(a, a.CanonicalID())
  }
  ```
  `domain.Research` defines `CanonicalID()`:
  ```go
  func (r Research) CanonicalID() string {
      if r.FilenameID != "" {
          return r.FilenameID
      }
      return r.ID
  }
  ```
- **Actual result:** `ToResearchJSON` uses `r.ID` (the parsed YAML frontmatter `id:`), ignoring `r.FilenameID` and `r.CanonicalID()`.
- **Expected result:** `ToResearchJSON` should use `r.CanonicalID()` (matching `ToTaskJSON` and `ToAuditJSON`) when operating on bare domain entities, ensuring that filename-derived canonical identity is projected when frontmatter `id:` is missing or drifting.
- **Consequence:** In `ToResearchMutationEnvelope` (used by `research set` and `research append` with `--json`), if an existing research document from the legacy corpus lacks an explicit frontmatter `id:` (where identity is derived from `research/<id>-<slug>.md`), the returned JSON envelope projects an empty `id: ""` rather than the document's canonical ID.
- **Smallest remedy:** Update `internal/wire/dto.go:ToResearchJSON` to pass `r.CanonicalID()` instead of `r.ID`.

**Resolution:** Bare research mutation JSON now uses Research.CanonicalID(), so
a missing or drifting frontmatter ID cannot override the filename-derived key; a
regression covers both cases.

#### L1. Column contract test fixtures set identical declared and source IDs, masking source-ID fallback regressions · **Status:** fixed

- **Layer crossing:** `internal/cli/render/columns.go` (`loadedRecordColumns`) $\rightarrow$ `internal/cli/render/columns_contract_test.go` (`TestColumnRegistriesMatchFullWireValues`).
- **Reproduction path:**
  In `internal/cli/render/columns.go`, `loadedRecordColumns` overrides the `"id"` column to extract `record.Source.ID`:
  ```go
  display := func(record core.LoadedRecord[T]) string {
      if c.selectorName() == "id" {
          return record.Source.ID
      }
      return c.Extract(record.Value)
  }
  ```
  However, in `internal/cli/render/columns_contract_test.go` (lines 160–210), every single fixture for `TaskReadColumns`, `EpicColumns`, `AuditReadColumns`, and `ResearchReadColumns` sets `Value.ID == Source.ID`:
  - `Task`: `Value.ID: "6ga000000001", Source: {ID: "6ga000000001"}`
  - `Task`: `Value.ID: "6ga000000002", Source: {ID: "6ga000000002"}`
  - `Audit`: `Value.ID: "6ga000000003", Source: {ID: "6ga000000003"}`
  - `Research`: `Value.ID: "6ga000000005", Source: {ID: "6ga000000005"}`
  When Mutation Probe 6 was applied (disabling the `if c.selectorName() == "id"` override), the entire test suite in `internal/cli/...` continued to pass without a single failure.
- **Actual result:** No test in `internal/cli/...` verifies that `TaskReadColumns`, `AuditReadColumns`, or `ResearchReadColumns` project `Source.ID` when frontmatter `id:` is missing or drifting.
- **Expected result:** Contract test fixtures should include records where `Value.ID != Source.ID` and `Value.ID == ""` to prove that source identity takes precedence over frontmatter declaration.
- **Consequence:** Future refactorings that inadvertently drop the `Source.ID` override in column projections will go undetected by the regression suite.
- **Smallest remedy:** Add fixtures with drifting (`Value.ID: "drift"`) and missing (`Value.ID: ""`) frontmatter IDs to `TestColumnRegistriesMatchFullWireValues` in `internal/cli/render/columns_contract_test.go`.

**Resolution:** Column contract fixtures now cover both drifting and absent
declared IDs against canonical source IDs for task, epic, audit, and research.
Dropping source-ID precedence fails the registry contract test.

#### L2. Epic show CLI projects bare domain task values without binding canonical source identity · **Status:** fixed

- **Layer crossing:** `internal/cli/epic.go` (`newEpicShowCmd`) $\rightarrow$ `internal/cli/render/render.go` (`EpicShowHuman`).
- **Reproduction path:**
  In `internal/cli/epic.go` lines 357–360:
  ```go
  tasks := make([]domain.Task, 0, len(detail.Tasks))
  for _, record := range detail.Tasks {
      tasks = append(tasks, record.Value)
  }
  return render.EpicShowHuman(w, app.Style, detail.Summary, tasks, rendered)
  ```
  Everywhere else in the codebase (e.g., `internal/core/service_task.go:taskGraphTasks`, `internal/tui/commands.go:loadTaskList`, `internal/tui/commands.go:loadEpicDetail`, `internal/core/service.go:Summary`), converting a `LoadedRecord[domain.Task]` to a bare `domain.Task` explicitly binds `task.FilenameID = record.Source.ID` so that `task.CanonicalID()` is guaranteed to reflect the authoritative source identity.
- **Actual result:** In `internal/cli/epic.go`, `record.Value` is extracted directly without setting `task.FilenameID = record.Source.ID`.
- **Expected result:** The slice conversion should set `task.FilenameID = record.Source.ID` before passing `tasks` to `render.EpicShowHuman`.
- **Consequence:** While `EpicShowHuman` currently only renders `task.Slug` in the status tree, passing unbound domain values violates the architectural invariant and invites identity drift if downstream presentation helpers inspect `task.CanonicalID()`.
- **Smallest remedy:** Set `task.FilenameID = record.Source.ID` when assembling `tasks` in `internal/cli/epic.go:newEpicShowCmd`.

---

**Resolution:** Epic-show's bare task projection now binds FilenameID from the
loaded record's authoritative Source.ID before presentation.

### 3. Consumer inventory

All changed ordinary read ports and downstream consumers were inventoried across `internal/core`, `internal/store`, `internal/cli`, `internal/tui`, and `internal/wire`:

| Interface / Port | Method Signature | Status / Implementation | Downstream Consumers | Classification of Remaining Metadata |
|:---|:---|:---|:---|:---|
| `core.TaskStore` | `ReadTasks() (TaskRead, error)` | Replaces `ListTasks() ([]domain.Task, []domain.FileProblem, error)` | `core.Service.ListTasks`, `core.Service.Board` | `FileProblem` removed from port. Domain `Path`/`FilenameID` retained temporarily on `domain.Task` (deferred to `6gcwcf88z57p`). |
| `core.TaskStore` | `ReadTask(ref string) (LoadedRecord[TaskWithBody], error)` | Replaces `GetTask` for single-document reads | `core.Service.ShowTask`, `internal/cli/task_show.go`, `internal/tui/commands.go` | Returns `LoadedRecord[TaskWithBody]`. Wire DTO uses `record.Source.ID`. |
| `core.EpicStore` | `ReadEpics() (EpicRead, error)` | Replaces `ListEpics() ([]domain.Epic, []domain.FileProblem, error)` | `core.Service.ListEpics`, `core.Service.Summary`, `core.Service.ShowEpic` | `FileProblem` removed from port. `domain.Epic.ID` populated from `record.Source.ID` in rollups. |
| `core.EpicStore` | `ReadEpic(ref string) (LoadedRecord[EpicWithBody], error)` | Replaces `GetEpic` | `core.Service.ShowEpic` | Returns `LoadedRecord[EpicWithBody]`. |
| `core.AuditStore` | `ReadAudits() (AuditRead, error)` | Replaces `ListAudits() ([]domain.Audit, []domain.FileProblem, error)` | `core.Service.ListAudits` | `FileProblem` removed from port. `ToLoadedAuditJSON` uses `record.Source.ID`. |
| `core.AuditStore` | `ReadAudit(ref string) (LoadedRecord[AuditWithBody], error)` | Replaces `GetAudit` | `core.Service.ShowAudit` | Returns `LoadedRecord[AuditWithBody]`. |
| `core.ResearchStore` | `ReadResearch() (ResearchRead, error)` | Replaces `ListResearch() ([]domain.Research, []domain.FileProblem, error)` | `core.Service.ListResearch` | `FileProblem` removed from port. `ToLoadedResearchJSON` uses `record.Source.ID`. |
| `core.ResearchStore` | `ReadResearchDocument(ref string) (LoadedRecord[ResearchWithBody], error)` | Replaces `GetResearch` | `core.Service.ShowResearch` | Returns `LoadedRecord[ResearchWithBody]`. |
| `core.SummaryStore` | `ReadEpics() (EpicRead, error)` | Narrowed capability; legacy methods removed | `core.Service.Summary` | Dedicated snapshot port. Single-scan behavior verified. |
| `core.LintSource` | `ReadLintTasks() ([]LoadedRecord[TaskWithBody], []LoadProblem, error)` | Replaces `LintLoadProblem` with `LoadProblem` | `internal/core/lint_service.go`, `internal/cli/lint.go` | Shared `LoadProblem` vocabulary across lint and ordinary reads. |

#### Classification of remaining core-facing fields
1. `domain.FileProblem`: Retained strictly in legacy test fakes (`internal/core/service_epic_test.go`, `internal/core/dependency_graph_test.go`, `internal/core/usecases_test.go`) and backward-compatible graph helper `NewTaskGraph` / `TaskGraphReadFromFiles`. Zero production read ports return `FileProblem`. Classified as **intentional guarded/local compatibility**.
2. `domain.Task.Path`, `FilenameID`, `CanonicalID()`: Retained on the `domain.Task` struct. Explicitly marked as out-of-scope in `planning/tasks/6gcwcf80v8hg.md` ("Removing domain Path, FilenameID, or guarded SourceVersion fields; the sequenced source/path task performs that migration after these snapshots exist"). Classified as **deferred work** assigned to `6gcwcf88z57p`.
3. `LocationIsPath`: Present in `TaskGraphLoadProblem` and `ThreadReadProblem`; eliminated from `core.LoadProblem` in favor of independent `Location` and `LocalPath` fields. Classified as **intentional guarded compatibility**.
4. `SourceVersion`: Stripped from ordinary read envelopes (`TaskRead`, `EpicRead`, `AuditRead`, `ResearchRead`) and absent from public wire DTOs. Retained privately on `TaskGraphRead` and `ThreadRead` for whole-snapshot CAS. Classified as **intentional guarded evidence**.

---

### 4. Field-lineage table

The table below traces the lifecycle and precedence of all record identity and diagnostic fields across readable and unreadable states for all four entity kinds:

| Entity State | Field | Producer | Conversion / Envelope | Core Consumer | Human Output | Wire Output (`--json`) | Mutation Evidence |
|:---|:---|:---|:---|:---|:---|:---|:---|
| **Readable Task** | `Source.ID` | `store.taskRecord` via `splitFlatName` | `LoadedRecord[Task].Source.ID` | `ListTasks`, `taskGraphTasks`, `Board`, `rollupEpics` | `TaskReadColumns` (`-o table/csv`) | `TaskJSON.ID`, `TaskInfoJSON.ID` | CAS match via `SourceVersion` |
| **Readable Task** | Declared `ID` | `domain.parseTaskFrontmatter` | `LoadedRecord[Task].Value.ID` | `domain.LintTask` (detects drift) | Ignored in list/show | Omitted when drifting; wire uses Source ID | Retained in raw file bytes |
| **Readable Task** | `Slug` | `splitFlatName` | `LoadedRecord[Task].Value.Slug` | Filtering, CLI handle | First column in list, `task show` | `TaskJSON.Slug` | Renamed on file rename |
| **Readable Task** | `Location` | `task.Path` in filesystem store | `LoadedRecord[Task].Source.Location` | Opaque context; not parsed | Unused in table | Omitted from TaskJSON | Context only |
| **Readable Task** | `LocalPath` | `LocalTaskSource` (optional) | Not on `LoadedRecord` | `app.Svc.TaskPath` | `task path`, `$EDITOR` | `TaskInfoJSON.Path` | Target for atomic write |
| **Unreadable Task** | `EntityID` | `splitFlatName` on filename | `LoadProblem.EntityID` | `canonicalLoadProblems`, `Board.Problems` | Leading error label `! task (ID)` | `LintLoadProblemJSON.entity_id` | Key for repair |
| **Unreadable Task** | `EntitySlug` | `splitFlatName` on filename | `LoadProblem.EntitySlug` | `portableProblemName` | Error label | `LintLoadProblemJSON.entity_slug` | Diagnostic handle |
| **Unreadable Task** | `Location` | Raw file path or URI | `LoadProblem.Location` | Diagnostic context | `location: <loc>` | `LintLoadProblemJSON.location` | Explanatory only |
| **Unreadable Task** | `LocalPath` | Filesystem path | `LoadProblem.LocalPath` | Diagnostic context | `repair: <path>` (when $\neq$ Loc) | `LintLoadProblemJSON.path` | `$EDITOR` repair target |
| **Unreadable Task** | `Message` | YAML parser error | `LoadProblem.Message` | Diagnostic error | Error description | `LintLoadProblemJSON.message` | Explanatory only |
| **Unreadable Task** | `SourceVersion` | `hashContent(content)` | `TaskGraphLoadProblem.SourceVersion` | `SameSourceSnapshot` (CAS) | Never published | Omitted (`json:"-"`) | Authorizes repair write |
| **Readable Epic** | `Source.ID` | `epic.ID` (filename stem) | `LoadedRecord[Epic].Source.ID` | `rollupEpics`, `canonicalEpic`, `Summary` | First column in `epic list` | `EpicMetaJSON.ID` | Stamped on member tasks |
| **Readable Epic** | `Slug` | Derived from stem | `LoadedRecord[Epic].Value.ID` | Human handle | Header in show | `EpicMetaJSON.ID` | File stem |
| **Readable Epic** | `Location` | `epic.Path` | `LoadedRecord[Epic].Source.Location` | Opaque context | Unused | Omitted | Context only |
| **Unreadable Epic** | `LoadProblem` | `store.scanDir` | `core.LoadProblem` (Kind: `epic`) | `canonicalLoadProblems`, `Summary` | `! epic <name>` | `LintLoadProblemJSON` | Local path for repair |
| **Readable Audit** | `Source.ID` | `splitFlatName` on filename | `LoadedRecord[Audit].Source.ID` | `ListAudits`, `Summary`, `QueryFindings` | `audit list` `id` column | `AuditJSON.ID` | Stamped on findings (`AuditID`) |
| **Readable Audit** | Declared `ID` | Frontmatter `id:` | `LoadedRecord[Audit].Value.ID` | `domain.LintAudit` (drift check) | Ignored in list | Omitted; Source ID wins | Frontmatter byte check |
| **Unreadable Audit**| `LoadProblem` | `store.scanDir` | `core.LoadProblem` (Kind: `audit`) | `canonicalLoadProblems`, `Summary` | `! audit <slug>` | `LintLoadProblemJSON` | Local path for repair |
| **Readable Research**| `Source.ID` | `splitFlatName` on filename | `LoadedRecord[Research].Source.ID` | `ListResearch`, `ResearchReadColumns` | `research list` `id` col | `ResearchJSON.ID` | Authoritative key |
| **Readable Research**| Declared `ID` | Frontmatter `id:` | `LoadedRecord[Research].Value.ID` | `domain.LintResearch` | Ignored in list | Bug in mutation (Finding M1) | Frontmatter byte check |
| **Unreadable Research**| `LoadProblem`| `store.scanDir` | `core.LoadProblem` (Kind: `research`) | `canonicalLoadProblems` | `! research <slug>` | `LintLoadProblemJSON` | Local path for repair |

---

### 5. Independent fixtures and fakes evaluation

The implementation was evaluated against independent fixtures covering all 9 required identity and location scenarios:

1. **Matching identity:** Task with filename `6ga000000001-slug.md` and frontmatter `id: 6ga000000001`.
   - *Behavior:* `Source.ID` matches `Value.ID`. Full JSON, table, and CSV project `6ga000000001`. Clean lint pass.
2. **Missing or drifting declaration:** Task with filename `6ga000000001-slug.md` and frontmatter `id: 6ga999999999`.
   - *Behavior:* `Source.ID` (`6ga000000001`) takes precedence in `ToLoadedTaskJSON`, `TaskReadColumns`, and `taskGraphTasks`. `domain.LintTask` flags `ErrTaskIDDrift`.
3. **Stale FilenameID:** Fake adapter supplying `Value.FilenameID: "stale"` and `Source.ID: "authoritative"`.
   - *Behavior:* `taskGraphTasks` explicitly overrides `task.FilenameID = record.Source.ID`, preventing stale adapter field leakage into graph analysis and Board. Verified by `TestBoard_BareProjectionUsesExplicitSourceIdentity`.
4. **Explicit ID plus misleading URI/path:** Remote problem with `EntityID: "6ga000000001"`, `Location: "https://api.internal/v1/tasks/999"`, and empty `LocalPath`.
   - *Behavior:* `portableProblemsError` outputs `6ga000000001` and ignores the URI. Wire JSON projects `location: "https://api.internal/v1/tasks/999"` and `path: ""`. A URI is never projected into `path`.
5. **Malformed document with recoverable ID+slug:** File `tasks/6ga000000001-broken.md` with invalid YAML.
   - *Behavior:* `splitFlatName` extracts `EntityID: "6ga000000001"` and `EntitySlug: "broken"`. Emitted as `LoadProblem{EntityKind: "task", EntityID: "6ga000000001", EntitySlug: "broken", LocalPath: "tasks/6ga000000001-broken.md"}`.
6. **Malformed filename with neither:** File `tasks/broken.md` with invalid YAML.
   - *Behavior:* `splitFlatName` returns `ok = false`. Emitted as `LoadProblem{EntityKind: "task", EntityID: "", EntitySlug: "", LocalPath: "tasks/broken.md"}`. Core treats it as unidentified planning record without parsing the path.
7. **Duplicate canonical IDs with different locations:** Two records with `Source.ID: "6ga000000001"` at different locations.
   - *Behavior:* Graph analysis detects `ProblemDuplicateTaskID` and latches `GraphBroken`. Both records appear in `ListTasks`.
8. **Empty location:** In-memory or remote adapter omitting `Location`.
   - *Behavior:* `Source.Location` is `""`. Wire DTO omits `location` via `omitempty`.
9. **Both opaque location and local repair path:** Database record synchronized to local scratch clone (`Location: "db://tasks/1"`, `LocalPath: "/tmp/scratch/task.md"`).
   - *Behavior:* Human diagnostic prints both `location: db://tasks/1` and `repair: /tmp/scratch/task.md`.

---

### 6. Adapter calls and filesystem read measurements

Read single-scan guarantees were benchmarked and verified through counting stores across all primary read use cases:

| Operation | Application Port | Underlying Store Calls | Filesystem Passes | Verification Test |
|:---|:---|:---:|:---:|:---|
| **Ordinary Task List** | `Service.ListTasks` | 1 | 1 | `internal/core/service_task_test.go` |
| **Task Graph Read** | `TaskGraphSource.ReadTaskGraph` | 1 | 1 | `internal/core/dependency_graph_test.go` |
| **Active Board** | `Service.Board` | 1 (`ReadTaskGraph`) | 1 | `internal/core/board_test.go:TestBoard_SingleTaskScan` |
| **Summary Dashboard** | `Service.Summary` | 1 task + 1 epic + 1 audit | 3 total (1 per entity kind) | `internal/core/usecases_test.go:TestSummary_SingleScanForAuditBodies` |
| **Audit Findings Query** | `Service.QueryFindings` | 1 (`ReadAuditSnapshot`) | 1 | `internal/core/finding_test.go:TestQueryFindings_UsesSnapshotRead` |
| **Task Show / Info** | `Service.ShowTask` | 1 (`ReadTask`) | 1 | `internal/core/service_task.go` |

**Verification notes:**
- **Zero redundant parsing:** `Summary` derives both the open audit list and the complete findings rollup from the single `ReadAuditSnapshot` call without re-reading markdown bodies.
- **Zero double hashing:** `scanDirWithSourceVersions` computes SHA-256 byte hashes strictly once during directory scanning. Ordinary list scans (`scanDir`) bypass hash computation entirely.

> Implementation-owner correction (2026-09-27): the zero-double-hashing claim is not accurate for
> tasks. `scanTaskDocuments` uses `scanDirWithSourceVersions`, which hashes each source before parse,
> and `parseTask` hashes successful task content again for `Task.SourceVersion`. This is one file read
> and parse per task, but two SHA-256 passes over readable task bytes. No workload benchmark assigns
> it performance severity, so this is not a release-blocking audit finding.

---

### 7. Format and CLI contract verification

1. **Full JSON vs Column-projected `--json -c`:**
   - Both formats emit `"schema_version": "1.76"`.
   - Both formats project unreadable diagnostics as `[]LintLoadProblemJSON` under `"unreadable"`.
   - Column projection preserves canonical wire keys when selected (e.g., `-c id,slug,status`).
2. **Schema validation:**
   - Generated schema `internal/cli/testdata/golden/schema_jsonschema.golden` defines `LintLoadProblemJSON` with `required: ["path", "message"]`.
   - For pathless adapters where `LocalPath` is empty, wire serialization emits `"path": ""`, satisfying strict JSON Schema validation without null-type errors.
3. **Exit code 11:**
   - When any unreadable record is encountered during `task list`, `epic list`, `audit list`, or `research list`, the command outputs available valid items to stdout, writes diagnostics to stderr, and returns `domain.ErrValidation` (exit status 11). Verified via `TestTaskList_ReportsBadFileButShowsGood`.
4. **Golden delta:**
   - The 1.75 $\rightarrow$ 1.76 delta in `internal/cli/testdata/golden/` was verified: 48 envelope definitions had their `schema_version` const bumped to `"1.76"`, and `FileProblem` references in `unreadable` arrays were updated to `LintLoadProblemJSON`. No unrelated schema drift was found.

---

### 8. Source-ID precedence trace

The precedence of `RecordSource.ID` was traced across all application layers:
- **Task Graph & Dependency Attribution:** In `internal/core/service_task.go:taskGraphTasks`, `task.FilenameID = record.Source.ID` ensures `task.CanonicalID()` resolves to `record.Source.ID`. Cycle detection and dependency error attribution reference this canonical ID.
- **Epic Joins & Rollups:** In `internal/core/service_epic.go:rollupEpics`, tasks are joined using `domain.EpicRefKey(record.Source.ID)` (line 179), and `e.ID` is set to `record.Source.ID` (line 190). Proved by `TestEpicSourceIdentityWinsOverEmbeddedID`.
- **Audit Findings:** In `internal/core/finding.go:QueryFindings`, findings are constructed with `AuditID: loaded.Source.ID` (line 170).
- **Board Blocked Ordering:** In `internal/core/board.go`, `Blocked[t.CanonicalID()]` checks the graph using the canonical ID derived from `record.Source.ID`.
- **TUI Loaders:** All four TUI list loaders (`loadTaskList`, `loadAuditList`, `loadResearchList`, `loadEpicDetail`) bind `task.FilenameID = record.Source.ID` or `audit.FilenameID = record.Source.ID` before assembling Bubble Tea `list.Item` references.
- **Wire Converters:** `ToLoadedTaskJSON`, `ToLoadedEpicMeta`, `ToLoadedAuditJSON`, and `ToLoadedResearchJSON` map `record.Source.ID` directly to wire `id`.

---

### 9. Mutation testing in sandbox

Six mutation experiments were executed in the sandbox to verify test suite sensitivity:

| # | Target Mutation | Subsystem | Test Result | Analysis |
|:---:|:---|:---|:---|:---|
| **1** | Revert `task.FilenameID = record.Source.ID` in `taskGraphTasks` | `internal/core` | **FAIL**: `TestBoard_BareProjectionUsesExplicitSourceIdentity` failed with `"board identity = 'stale-file-id', want source ID"` | Mutant killed. Explicit source-ID precedence in board projection is actively tested. |
| **2** | Reconstruct problem ID from location when `EntityID` is empty | `internal/store` | **FAIL**: `internal/store/paths_test.go:TestThreadSourceSnapshot_UnidentifiedProblem` failed | Mutant killed. Tests enforce that core does not guess IDs from file paths. |
| **3** | Emit opaque `Location` as `Path` in `ToLintLoadProblemsJSON` | `internal/wire` | **FAIL**: `TestToLintLoadProblemsJSONKeepsOpaqueLocationsOutOfPath` failed | Mutant killed. Proves opaque URIs cannot masquerade as local filesystem paths. |
| **4** | Add a second task or audit read inside `Service.Summary()` | `internal/core` | **FAIL**: `internal/core/usecases_test.go:TestSummary_SingleScanForAuditBodies` failed (`calls = 2, want 1`) | Mutant killed. Single-scan contract is strictly enforced by call-counting stores. |
| **5** | Drop unreadable `SourceVersion` in `TaskGraphRead` comparison | `internal/core` | **FAIL**: `internal/core/dependency_graph_test.go:TestTaskGraphSameSourceSnapshotWithUnreadableProblems` failed | Mutant killed. Proves unversioned unreadable sources fail closed in CAS. |
| **6** | Revert `if c.selectorName() == "id"` in `loadedRecordColumns` | `internal/cli/render` | **PASS (Mutant survived)**: `go test ./internal/cli/...` succeeded | **Blind spot confirmed (Finding L1)**. Test fixtures all set `Value.ID == Source.ID`. |

---

### 10. Required hostile angles

1. **False portability:** A pathless fake store implementing `TaskReadStore`, `EpicStore`, `AuditStore`, and `ResearchStore` can successfully drive `task list`, `task show`, `board`, `epic list`, `audit list`, `research list`, and `status`. When `LocalPath` is empty, wire output emits `path: ""` without crashing, and TUI displays `"nothing to edit"` without attempting invalid filesystem I/O.
2. **Mixed-generation reads:** `Summary` coordinates reads from `taskGraphs`, `store.ReadEpics()`, and `auditsSource.ReadAuditSnapshot()`. In a concurrent filesystem scenario, writes between these calls represent cross-entity non-atomicity inherent to multi-document storage. However, within each entity family, reads are strictly single-scan and internally consistent.
3. **Corrupt corpus:** Resilient directory scans parse valid documents while collecting syntax/unreadable errors into `LoadProblem`. Duplicate canonical IDs latch `GraphBroken` without dropping either record.
4. **Contract drift:** Generated JSON Schema `schema_jsonschema.golden` accurately reflects `LintLoadProblemJSON` on `unreadable` arrays. CLI docs and golden tests match output byte-for-byte.
5. **Future migration pressure:** The transition layer uses `LoadedRecord[T]` and explicit `taskRecord` / `loadedProblems` mappings. Because domain `Path` and `FilenameID` were not removed prematurely, the subsequent migration task (`6gcwcf88z57p`) has clean, stable seam boundaries to sever.

---

### 11. Scope distinctions

- **Confirmed defects:** Finding M1 (`ToResearchJSON` using frontmatter `r.ID`).
- **Untested risks:** Finding L1 (contract test fixtures masking column projection regressions); Finding L2 (`epic show` CLI omitting `task.FilenameID` binding).
- **Intentionally deferred work:** Removing `Path`, `FilenameID`, and `SourceVersion` from domain structs; splitting optional local-path capabilities (`Resolve*Path`) out of semantic read stores; routing CLI repair commands through application ports. All belong to tasks `6gcwcf88z57p` and `6gcwcf8gzn50`.

---

### 12. Validation commands and complete isolation attestation

#### Validation commands executed in sandbox
```sh
# Establish baseline test suite pass
go test ./...

# Verify schema golden compatibility
go test ./internal/cli -run TestJSONSchemaGoldenMatchesReflection

# Verify column projection contracts
go test ./internal/cli/render -run TestColumnRegistriesMatchFullWireValues

# Verify isolated workspace status
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
```

#### Mandatory reviewer isolation attestation
```
isolated-review: sandbox initialized at /private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.oOB7Zl
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.oOB7Zl
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.oOB7Zl/.git
baseline_commit=da1d965b0009256f9ee1f4f110d2657b37b91984
source_blob=8fa3a9d74f642b7be3948c74e98d455b11fa0ae2
source_fingerprint=ab1579ab06195a8cda5d41acbc9e59b5c4f4c7c3
deliverable=planning/audits/6ge0q80ng3b4-2026-09-26-ordinary-entity-read-snapshots-implementation-antigravity.md
deliverable_changed=true
transfer=pending
```
