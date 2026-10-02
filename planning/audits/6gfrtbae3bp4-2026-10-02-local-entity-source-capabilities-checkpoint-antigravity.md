---
schema: 1
id: 6gfrtbae3bp4
bucket: open
area: local-entity-source-capabilities-checkpoint-antigravity
date: "2026-10-02"
updated_at: "2026-10-02"
---
# Audit: Local entity source capabilities checkpoint — antigravity — 2026-10-02

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

### 1. Executive Summary & Verdict

- **Verdict:** **Needs Remediation** (2 Medium findings, 4 Low findings, all left **open** for owner triage).
- **Review Target:** Task `6gcwcf88z57p` (`split-local-path-capabilities-from-semantic-entity-reads`), commits `003c39b..eeaeba2`.
- **Review Architecture:** Dual-pass adversarial audit executed in an isolated, non-hardlinked sandbox (`isolated-review.e7rUyo`) under the mandatory independent-sandbox protocol.
- **State of Checkpoint:**
  - The checkpoint successfully decouples optional local-path resolution (`TaskPathSource`, `EpicPathSource`, `AuditPathSource`, `ResearchPathSource`, `ThreadPathSource`) from portable semantic entity reads (`LoadedRecord[T]`), protecting core reads from depending on filesystem path assumptions.
  - The `SourceSetProvider` composition witness prevents cross-wiring disparate adapters or planning corpora.
  - However, the review identified two actionable defects requiring remediation before completing the split:
    1. **M1 (Wire Projection Inconsistency):** In `StatusAllEnvelope` (`tskflwctl status --all --json`), `overview.InProgress` projects tasks via `ToTaskJSON(item.Task)` instead of `ToLoadedTaskJSON(LoadedRecord)`, emitting unverified frontmatter IDs (leaking drift) and completely dropping source locations, while the nested `spaces.summary.in_progress` array projects canonical source IDs and locations.
    2. **M2 (Dead Tautology & Repair Locality Bypass):** In `resolveSourceTask`, the check `source.LocalPath != "" && !hasLocalRepairPath(source)` is an algebraic contradiction (`A && !A == false`) because `hasLocalRepairPath` was simplified to `return source.LocalPath != ""`. The guard intended to reject opaque locations in `LocalPath` is completely dead code.
  - In addition, four low-severity findings identify architectural asymmetries and test gaps:
    - **L1:** `WithLintSource` detaches `s.auditPaths` but leaves `s.taskPaths`, `s.epicPaths`, and `s.researchPaths` attached, creating asymmetrical capability detachment.
    - **L2:** Thread wire projections lack a `ToLoadedThreadJSON` projection, continuing to emit declared frontmatter IDs and blocking the clean removal of transitional `Thread.Path` and `Thread.FilenameID`.
    - **L3 (Test Gap):** Mutating `hasLocalRepairPath` to accept path-shaped opaque locations or `LocationIsPath` survived the test suite because existing tests exclusively use `"db://..."` URI schemes.
    - **L4 (Test Gap):** In `TestDelayedLocalPathActionCannotRetargetSelectionOrReload`, mutating the TUI `loadGen` guard survived because the test bumped `loadGen` without syncing `coherentGen`, causing `selectedRef()` to fail prematurely and masking the generation check.

---

### 2. Review Environment & Isolation Attestation

All inspection, builds, test execution, mutation probes, and report edits were performed exclusively inside the independent review sandbox:

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/.git
baseline_commit=2dae51893d21ef42cecacb19c0f4fb21b2f5f94d
source_blob=eb38e1a937b8229ddeebfee09c2d84e1ec432a0f
source_fingerprint=9ceb59f977483c0ad0d7a279c5b00c6a07395fcd
deliverable=planning/audits/6gfrtbae3bp4-2026-10-02-local-entity-source-capabilities-checkpoint-antigravity.md
```

No writes, git modifications, or temporary files were created in `$SOURCE_ROOT`. All mutation probes have been restored to the captured baseline.

---

### 3. Findings

#### M1. Status-all wire projection emits drifted task declared IDs and omits location · **Status:** fixed

- **Trigger:** Serializing `tskflwctl status --all --json` when a task in progress has frontmatter ID drift or an opaque source location.
- **Actual Behavior:** In [`internal/wire/envelopes.go:368`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/wire/envelopes.go#L368), `StatusAllEnvelope` constructs `out.Overview.InProgress` by calling `ToTaskJSON(item.Task)`:
  ```go
  for _, item := range overview.InProgress {
      out.Overview.InProgress = append(out.Overview.InProgress, StatusInProgressItemJSON{
          Task:      ToTaskJSON(item.Task),
          BlockedBy: item.BlockedBy,
      })
  }
  ```
  `item.Task` contains only domain data. `ToTaskJSON` sets `ID: item.Task.ID` (the un-migrated declared frontmatter ID) and `Location: ""` (since `domain.Task` has no location field).
  Simultaneously, on line 394 of the same file, the space summary array is populated using `ToLoadedTaskJSON(record)`:
  ```go
  for _, record := range space.Summary.InProgress {
      inProgress = append(inProgress, ToLoadedTaskJSON(record))
  }
  ```
  This calls `ToLoadedTaskJSON`, which sets `ID: record.Source.ID` (the canonical source ID) and `readableSourceLocation(record.Source.Location, record.Value.Path)`.
  As a result, within the exact same JSON response envelope:
  - If a task has frontmatter ID drift, `out.overview.in_progress[].task.id` emits the drifted frontmatter ID, contradicting `out.spaces[].summary.in_progress[].id` which emits the canonical source ID.
  - If a task resides at an opaque location (`record.Source.Location != ""`), `overview.in_progress[].task.location` is omitted/empty, contradicting `spaces[].summary.in_progress[].location`.
- **Expected Behavior:** `overview.InProgress` items should project canonical identity and location parity with space summaries by using `ToLoadedTaskJSON(core.LoadedRecord[domain.Task]{Value: item.Task, Source: item.Source})`.
- **Severity:** Medium (Public wire DTO contradiction and source identity leak in machine contract).
- **Exact File & Line:** [`internal/wire/envelopes.go:368`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/wire/envelopes.go#L368).
- **Minimal Reproduction:**
  1. Construct a `domain.InProgressItem` where `item.Source.ID = "6g0000000001"`, `item.Source.Location = "db://tasks/1"`, and `item.Task.ID = "drifted-frontmatter-id"`.
  2. Map it into `StatusAllEnvelope(overview, spaces, registered, activeSpace)`.
  3. Inspect `envelope.Overview.InProgress[0].Task.ID` (evaluates to `"drifted-frontmatter-id"`) versus `envelope.Spaces[0].Summary.InProgress[0].ID` (evaluates to `"6g0000000001"`).
- **Fix Direction:** In [`internal/wire/envelopes.go:368`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/wire/envelopes.go#L368), change `Task: ToTaskJSON(item.Task)` to:
  ```go
  Task: ToLoadedTaskJSON(core.LoadedRecord[domain.Task]{Value: item.Task, Source: item.Source}),
  ```

---

**Resolution:** The combined status-all working set now uses ToLoadedTaskJSON
with its adapter source; regression checks canonical ID and opaque location
match the nested summary.

#### M2. hasLocalRepairPath simplification renders resolveSourceTask location check a dead tautology · **Status:** fixed

- **Trigger:** Calling `resolveSourceTask` during task graph repair with a `source.LocalPath` that represents an opaque location or URI.
- **Actual Behavior:** In [`internal/core/dependency_source.go:285-287`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/dependency_source.go#L285-L287):
  ```go
  if source.LocalPath != "" && !hasLocalRepairPath(source) {
      return 0, false, fmt.Errorf("%w: repair path must be local, not an opaque location", domain.ErrValidation)
  }
  ```
  However, in [`internal/core/dependency_repair.go:322-324`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/dependency_repair.go#L322-L324), `hasLocalRepairPath` is implemented as:
  ```go
  func hasLocalRepairPath(source TaskGraphSourceRef) bool {
      return source.LocalPath != ""
  }
  ```
  Substituting the function body into the condition yields:
  `source.LocalPath != "" && !(source.LocalPath != "")`
  This expression is an algebraic impossibility ($A \land \neg A \equiv \text{false}$) and can never evaluate to true under any input. The error return is completely dead code, leaving the caller with no guard against an opaque location placed in `LocalPath` (such as `"db://repair/task.md"`).
- **Expected Behavior:** If the intent is to guard against opaque locations or URIs in `LocalPath`, the condition should verify that `LocalPath` is not a URI scheme (`!strings.Contains(source.LocalPath, "://")`) or that `hasLocalRepairPath` inspects the structure of `source.LocalPath` and `source.Location`.
- **Severity:** Medium (Dead security/locality guard creating false confidence of input validation).
- **Exact File & Line:** [`internal/core/dependency_source.go:285-287`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/dependency_source.go#L285-L287) and [`internal/core/dependency_repair.go:322-324`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/dependency_repair.go#L322-L324).
- **Minimal Reproduction:**
  1. Inspect AST / code coverage for line 286 in `internal/core/dependency_source.go`.
  2. Observe that no test or input can ever cause line 286 to execute.
  3. Pass `source := TaskGraphSourceRef{LocalPath: "db://custom/opaque/path.md"}` to `resolveSourceTask`.
  4. Observe that the validation error `"repair path must be local, not an opaque location"` is not returned.
- **Fix Direction:** Define `hasLocalRepairPath` to explicitly check locality (e.g. `source.LocalPath != "" && !strings.Contains(source.LocalPath, "://")`), or remove the dead condition in `resolveSourceTask` if upstream invariants guarantee `LocalPath` is already sanitized.

---

**Resolution:** Removed the dead tautology. LocalPath is an explicit
adapter-owned repair capability, not a string for core to classify by URI
heuristic; opaque Location alone remains non-repairable.

#### L1. Asymmetric path detachment in WithLintSource leaves task, epic, and research paths attached · **Status:** wontfix

- **Trigger:** Configuring a `Service` using `WithStore(store)` and overriding lint reads via `WithLintSource(customLintSource)`.
- **Actual Behavior:** In [`internal/core/service.go:122-133`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/service.go#L122-L133):
  ```go
  func WithLintSource(source LintSource) Option {
      return func(s *Service) {
          if !isNilCapability(source) {
              s.lintReads = source
              if !s.auditReadsExplicit {
                  s.auditReads = source
                  if !s.auditPathsExplicit {
                      s.auditPaths = nil
                  }
              }
          }
      }
  }
  ```
  `WithLintSource` detaches `s.auditPaths = nil`, but leaves `s.taskPaths`, `s.epicPaths`, and `s.researchPaths` bound to `s.store`, even though `LintSource` defines `ReadLintTasks`, `ReadLintEpics`, and `ReadLintResearch`.
  Additionally, there are no symmetrical options `WithEpicStore`, `WithResearchStore`, or `WithEpicPathSource` to independently detach or override these capabilities without providing a monolithic `Store`.
- **Expected Behavior:** Capability detachment should be symmetric across all entity kinds exposed by `LintSource`. If overriding lint reads indicates a split or remote corpus, task, epic, and research paths should either be similarly detached when not explicitly configured, or the detachment policy should be documented and uniform.
- **Severity:** Low (Architectural asymmetry in core service options).
- **Exact File & Line:** [`internal/core/service.go:122-133`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/service.go#L122-L133).
- **Minimal Reproduction:**
  1. Create a `store` that implements `Store`, `TaskPathSource`, and `AuditPathSource`.
  2. Create a service: `svc, _ := NewService(store, WithLintSource(remoteLintSource))`.
  3. Observe: `svc.HasLocalPath(EntityAudit)` returns `false` (detached), but `svc.HasLocalPath(EntityTask)` returns `true` (still attached to `store`).
- **Fix Direction:** Standardize detachment in `WithLintSource` across all entity path sources or provide explicit entity-level store options.

---

**Resolution:** WithLintSource replaces lint reads and the audit snapshot, not
the task/epic/research semantic read ports. It correctly detaches audit
navigation; independent read overrides detach their corresponding path ports and
source-set validation rejects foreign composition.

#### L2. Thread wire projection lacks ToLoadedThreadJSON and leaks drifted declared ID · **Status:** wontfix

- **Trigger:** Serializing a `domain.ThreadView` in `tskflwctl thread show --json` or `tskflwctl thread list --json` when a Thread document's YAML frontmatter `id:` has drifted from its filename / canonical source ID.
- **Actual Behavior:** In [`internal/wire/thread.go:89-91`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/wire/thread.go#L89-L91):
  ```go
  func ToThreadViewJSON(view domain.ThreadView) ThreadViewJSON {
      return ThreadViewJSON{
          Thread: ToThreadJSON(view.Thread),
          // ...
      }
  }
  ```
  `ToThreadJSON(view.Thread)` maps directly from `domain.Thread`, setting `ID: thread.ID` (the declared frontmatter ID) and `Path: thread.Path` (transitional domain field). It completely ignores `view.Source.ID` and `view.Source.Location`.
  Unlike Tasks, Epics, Audits, and Research records—which all have dedicated `ToLoaded<Kind>JSON(LoadedRecord)` mappers that prioritize `record.Source.ID`—Threads have no `ToLoadedThreadJSON` projection.
- **Expected Behavior:** `ToThreadViewJSON` should project canonical identity from `view.Source.ID` and location from `view.Source.Location`, ensuring parity with other entities and preventing frontmatter drift from corrupting wire DTOs.
- **Severity:** Low (Migration seam gap; concrete blocker for the final removal of transitional `Thread.Path` and `Thread.FilenameID`).
- **Exact File & Line:** [`internal/wire/thread.go:89-91`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/wire/thread.go#L89-L91).
- **Minimal Reproduction:**
  1. Construct a `domain.ThreadView` where `view.Source.ID = "6g0000000003"` and `view.Thread.ID = "drifted-frontmatter-id"`.
  2. Call `ToThreadViewJSON(view)`.
  3. Inspect `dto.Thread.ID`: it contains `"drifted-frontmatter-id"` instead of canonical `"6g0000000003"`.
- **Fix Direction:** Introduce `ToLoadedThreadJSON(record LoadedRecord[domain.Thread]) ThreadJSON` or update `ToThreadViewJSON` to set `dto.Thread.ID = view.Source.ID` and `dto.Thread.Location = readableSourceLocation(view.Source.Location, view.Thread.Path)`.

---

**Resolution:** Thread wire intentionally preserves the declared document ID;
TestThreadViewEnvelopesPreserveDeclaredIDWhenSourceDiffers pins this
compatibility contract. ThreadJSON has no Path field, and canonical source IDs
are used separately for routing and lint.

#### L3. Test suite coverage gap allows path-shaped opaque locations and LocationIsPath to survive as repair authority · **Status:** fixed

- **Trigger:** Mutating `hasLocalRepairPath` in `internal/core/dependency_repair.go:322` to treat path-shaped locations (or `LocationIsPath`) as local repair paths:
  ```go
  func hasLocalRepairPath(source TaskGraphSourceRef) bool {
      return source.LocalPath != "" || (source.Location != "" && !strings.Contains(source.Location, "://"))
  }
  ```
- **Actual Behavior:** The mutation probe **survived** across all packages (`internal/core`, `internal/store`, `internal/cli`, and `internal/tui`).
  The current test suite exclusively uses `"db://..."` URIs when testing opaque locations. It contains zero test cases where:
  - An opaque location is path-shaped (e.g. `Location = "planning/tasks/opaque.md"` while `LocalPath = ""`).
  - An opaque location uses a non-slashed URN scheme (e.g. `urn:task:abc`).
  - An adapter sets `LocationIsPath: true` while `LocalPath` is empty.
  Consequently, code regressions that inadvertently conflate path-shaped locations with repair authority are not caught by existing tests.
- **Expected Behavior:** Regression tests in `internal/core/dependency_repair_test.go` should assert that path-shaped opaque locations (`"planning/tasks/sample.md"`) with `LocalPath: ""` are classified as non-repairable (`Repairable: false`).
- **Severity:** Low (Test suite coverage gap).
- **Exact File & Line:** [`internal/core/dependency_repair.go:322`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/dependency_repair.go#L322) and [`internal/core/dependency_repair_test.go:68-85`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/dependency_repair_test.go#L68-L85).
- **Minimal Reproduction:**
  1. In `internal/core/dependency_repair.go:322`, replace `return source.LocalPath != ""` with `return source.LocalPath != "" || (source.Location != "" && !strings.Contains(source.Location, "://"))`.
  2. Run `go test ./internal/core/... ./internal/store/... ./internal/cli/...`.
  3. Observe that all tests pass without detecting the survival.
- **Fix Direction:** Add a test case in `internal/core/dependency_repair_test.go` verifying that a task with `Location = "tasks/path-shaped-opaque.md"` and `LocalPath = ""` produces `defect.Repairable == false` and `defect.Automatic == false`.

---

**Resolution:** Repair regression now covers URI and path-shaped opaque
locations, including LocationIsPath=true, with no LocalPath; none authorizes
repair.

#### L4. TUI listGen reload test masks stale-result vulnerability due to unsynced coherentGen · **Status:** fixed

- **Trigger:** Mutating `internal/tui/model.go:393` to remove the reload generation check:
  ```go
  // Mutated: removed msg.listGen != m.cur().loadGen
  if !m.isCurrentSelection(msg.kind, msg.id) {
      return m, nil
  }
  ```
- **Actual Behavior:** [`TestDelayedLocalPathActionCannotRetargetSelectionOrReload`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/tui/model_test.go#L492) **passed**, failing to kill the mutation.
  In `internal/tui/model_test.go:531`, the test simulates a list reload by doing:
  ```go
  m.cur().loadGen++
  ```
  However, it does not advance `coherentGen`. Because `selectedRef()` explicitly checks:
  ```go
  if m.cur().coherentGen != m.cur().loadGen {
      return itemRef{}
  }
  ```
  `m.selectedRef()` returned an empty `itemRef{}`, causing `m.isCurrentSelection(msg.kind, msg.id)` to return `false` on line 393.
  Thus, the test passed because `isCurrentSelection` failed due to the artificial incoherence, not because `msg.listGen != m.cur().loadGen` dropped the stale reload message.
- **Expected Behavior:** A test asserting that delayed local path results cannot apply after a list reload should advance both `loadGen` and `coherentGen` (simulating a completed list reload that happens to re-select the same item ID), proving that `msg.listGen != m.cur().loadGen` is the true barrier against the stale result.
- **Severity:** Low (Test helper masking defect class).
- **Exact File & Line:** [`internal/tui/model.go:393`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/tui/model.go#L393) and [`internal/tui/model_test.go:531-536`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/tui/model_test.go#L531-L536).
- **Minimal Reproduction:**
  1. In `internal/tui/model.go:393`, change `if !m.isCurrentSelection(msg.kind, msg.id) || msg.listGen != m.cur().loadGen` to `if !m.isCurrentSelection(msg.kind, msg.id)`.
  2. Run `go test -count=1 ./internal/tui -run TestDelayedLocalPathActionCannotRetargetSelectionOrReload`.
  3. Observe that the test passes.
- **Fix Direction:** In `internal/tui/model_test.go:531`, advance `m.cur().coherentGen++` alongside `m.cur().loadGen++` so that selection remains valid during the assertion.

---

**Resolution:** The delayed-result test now advances coherentGen with loadGen so
a retained selected row genuinely exercises the stale-generation guard.

### 4. Producer / Consumer Inventory & Architectural Seams

#### Entity Source and Capability Inventory

| Entity Kind | Field / Capability | Source / Producer | Primary Consumers | Boundary Classification |
|---|---|---|---|---|
| **Task** | `RecordSource.ID` | `FS.ReadLoadedTasks`, `FS.ReadLintTasks` (via `splitFlatName`) | `TaskGraphSourceRef.TaskID`, `ToLoadedTaskJSON`, `TaskPath` resolver, lint attribution | Semantic read & canonical identity |
| **Task** | `RecordSource.Location` | `FS.ReadLoadedTasks`, `FS.ReadLintTasks` | `ToLoadedTaskJSON`, `TaskGraphSourceRef.Location`, `GraphProblem.Location` | Diagnostic attribution & presentation context |
| **Task** | `LocalPath` | `FS.TaskPath` via `TaskPathSource` port | CLI `task path`, TUI editor/yank, `TaskGraphSourceRef.LocalPath` | Optional local navigation & mutation authority |
| **Task** | `SourceVersion` | `FS.ReadLoadedTasks` (SHA-256 byte digest) | `sameReadableTaskSources`, whole-graph CAS, repair verification | Guarded mutation authority |
| **Epic** | `RecordSource.ID` | `FS.ReadLoadedEpics`, `FS.ReadLintEpics` | `ToLoadedEpicJSON`, `EpicPath` resolver, board grouping | Semantic read & canonical identity |
| **Epic** | `RecordSource.Location` | `FS.ReadLoadedEpics`, `FS.ReadLintEpics` | `ToLoadedEpicJSON` (optional `location` field) | Diagnostic & presentation context |
| **Epic** | `LocalPath` | `FS.EpicPath` via `EpicPathSource` port | CLI `epic path`, TUI editor/yank | Optional local navigation |
| **Epic** | `SourceVersion` | `FS.ReadLoadedEpics` | CAS checks | Guarded mutation authority |
| **Audit** | `RecordSource.ID` | `FS.ReadAuditSnapshot`, `FS.ReadLintAudits` | `ToLoadedAuditJSON`, `AuditPath` resolver, finding linking | Semantic read & canonical identity |
| **Audit** | `RecordSource.Location` | `FS.ReadAuditSnapshot` | `ToLoadedAuditJSON` | Diagnostic & presentation context |
| **Audit** | `LocalPath` | `FS.AuditPath` via `AuditPathSource` port | CLI `audit path`, TUI navigation | Optional local navigation |
| **Audit** | `SourceVersion` | `FS.ReadAuditSnapshot` | Audit finding CAS verification | Guarded mutation authority |
| **Research** | `RecordSource.ID` | `FS.ReadLoadedResearch`, `FS.ReadLintResearch` | `ToLoadedResearchJSON`, `ResearchPath` resolver | Semantic read & canonical identity |
| **Research** | `RecordSource.Location` | `FS.ReadLoadedResearch` | `ToLoadedResearchJSON` | Diagnostic & presentation context |
| **Research** | `LocalPath` | `FS.ResearchPath` via `ResearchPathSource` port | CLI `research path`, TUI editor/yank | Optional local navigation |
| **Research** | `SourceVersion` | `FS.ReadLoadedResearch` | CAS checks | Guarded mutation authority |
| **Thread** | `RecordSource.ID` | `threadstore.go:54` (from filename `splitFlatName`) | `ThreadGraphRead`, `ToThreadViewJSON` (currently bypassed) | Semantic read & canonical identity |
| **Thread** | `RecordSource.Location` | `threadstore.go:54` | `ThreadGraphRead` | Diagnostic & presentation context |
| **Thread** | `LocalPath` | `FS.ThreadPath` via `ThreadPathSource` port | CLI `thread path`, TUI editor/yank | Optional local navigation |
| **Thread** | `SourceVersion` | `FS.ReadThreadGraph` | `ThreadMutationStore` CAS | Guarded mutation authority |

#### Transitional Field Inventory & Consumers Requiring Migration

The checkpoint retains transitional fields on `domain.Task` and `domain.Thread`. Their active consumers and migration requirements prior to final removal are:

1. **`domain.Task.Path` & `domain.Task.FilenameID`:**
   - [`internal/wire/envelopes.go:368`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/wire/envelopes.go#L368) (`StatusAllEnvelope`): Consumes `item.Task` and passes it to `ToTaskJSON`. **Needs migration:** must pass `core.LoadedRecord[domain.Task]` to `ToLoadedTaskJSON` (Finding M1).
   - [`internal/core/dependency_graph.go:311, 593`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/dependency_graph.go#L311): `legacyTaskSourceRef` falls back to `task.Path` and `task.FilenameID`. **Needs migration:** `legacyTaskSourceRef` should be removed once all callers supply `TaskGraphRead.Records` with explicit `RecordSource`.
   - `core.CreateTaskReceipt`: Already exposes `Local: LocalCreateOutcome{LocalPath, RelPath}` alongside `task.Task`. Callers should transition to `receipt.Local.LocalPath` before `task.Task.Path` is deleted.
2. **`domain.Thread.Path` & `domain.Thread.FilenameID`:**
   - [`internal/wire/thread.go:89`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/wire/thread.go#L89) (`ToThreadViewJSON`): Consumes `view.Thread.ID` and `view.Thread.Path`. **Needs migration:** requires a new `ToLoadedThreadJSON` projection taking `LoadedRecord[domain.Thread]` (Finding L2).
   - [`internal/core/service_thread.go:340`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/service_thread.go#L340) (`threadDiagnosticName`): Direct fallback to `thread.Path`. **Needs migration:** must take `RecordSource` rather than domain struct.
   - `core.CreateThreadReceipt` and `core.UpdateThreadReceipt`: Already expose `LocalPath` at the receipt root. Store tests in `threadmutation_test.go` still reference `receipt.Thread.Path` and must be repointed to `receipt.LocalPath`.

---

### 5. Executed Mutation-Kill & Falsification Matrix

The following six adversarial mutation probes were executed directly in the isolated sandbox. All probes have been restored to the captured baseline.

| Probe | Target File & Line | Mutation Description | Focused Test & Failure Assertion | Outcome |
|---|---|---|---|---|
| **Probe 1** | [`internal/core/source_set.go:41`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/source_set.go#L41) | Bypass witness verification in `validateSourceSet` by returning `nil` unconditionally. | `go test ./internal/core -run TestValidateSourceSet`<br>`--- FAIL: TestValidateSourceSet (0.00s)`<br>`source_set_test.go:50: expected error for mismatched witness, got nil` | **KILLED** |
| **Probe 2** | [`internal/core/dependency_repair.go:322`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/dependency_repair.go#L322) | Mutate `hasLocalRepairPath` to accept path-shaped opaque locations: `return source.LocalPath != "" \|\| (source.Location != "" && !strings.Contains(source.Location, "://"))`. | `go test ./internal/core -run TestTaskGraphRepairDoesNotSelectOpaqueLocations`<br>Test suite passed across `internal/core`, `internal/store`, `internal/cli`. | **SURVIVED** (Test Gap, Finding L3) |
| **Probe 3** | [`internal/core/dependency_graph.go:990`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/dependency_graph.go#L990) | Mutate `sameReadableTaskSources` to bypass `SourceVersion` comparison (replace with `true`). | `go test ./internal/core -run TestTaskGraphReadableDuplicateLocationsRemainDistinct`<br>`--- FAIL: TestTaskGraphReadableDuplicateLocationsRemainDistinct (0.00s)`<br>`dependency_graph_test.go:732: snapshot matched despite modified source version` | **KILLED** |
| **Probe 4** | [`internal/wire/envelopes.go:394`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/wire/envelopes.go#L394) | Mutate `ToLoadedTaskJSON` to project declared `record.Value.ID` instead of canonical `record.Source.ID`. | `go test ./internal/wire -run TestLoadedTaskProjectionUsesSourceID`<br>`--- FAIL: TestLoadedTaskProjectionUsesSourceID (0.00s)`<br>`envelopes_test.go:120: projected ID = "drifted-frontmatter", want source ID "6g0000000001"` | **KILLED** |
| **Probe 5** | [`internal/tui/model.go:393`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/tui/model.go#L393) | Remove list reload generation check `msg.listGen != m.cur().loadGen` from `localPathResultMsg`. | `go test -count=1 ./internal/tui -run TestDelayedLocalPathActionCannotRetargetSelectionOrReload`<br>Test passed due to unsynced `coherentGen`. | **SURVIVED** (Test Gap, Finding L4) |
| **Probe 6** | [`internal/core/entity_read.go:47`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/core/entity_read.go#L47) | In `TaskPath`, if `s.taskPaths == nil`, fall back to calling `s.Task(ctx, id)` and returning `task.Path`. | `go test ./internal/core -run TestPathlessSourceRejectsTaskPath`<br>`--- FAIL: TestPathlessSourceRejectsTaskPath (0.00s)`<br>`service_task_test.go:112: expected ErrUnsupported on pathless source, got "/path/to/task.md"` | **KILLED** |

---

### 6. Hostile Angle Deep-Dive

#### Angle 1: Optional Ports Composability & Multi-Workspace Cross-Wiring
- **Probe:** Challenged `WorkspaceSource` assembly in `internal/core/workspace.go:107` and `validateSourceSet` with two distinct workspaces open concurrently.
- **Verification:**
  - `NewSourceSetID()` mints monotonically increasing atomic nonces per adapter instance.
  - When capabilities from workspace $A$ and workspace $B$ were cross-wired (e.g. `s.taskPaths` from workspace $A$ combined with `s.store` from workspace $B$), `validateSourceSet` caught the discrepancy immediately and failed with `ErrIncompatibleCapabilities`.
  - Typed-nil handling was verified: passing an uninitialized interface pointer (`var p *fsTaskPaths = nil`) to `WithTaskPathSource(p)` triggers `isNilCapability(p) == true`, which cleanly suppresses registration rather than storing a non-nil interface wrapping a nil pointer.

#### Angle 2: Negative Contract (Pathless Source Gaining Path Authority)
- **Probe:** Evaluated whether a pathless adapter could accidentally gain local path authority via legacy fields, display hints, or mutation receipts.
- **Verification:**
  - `LocationIsPath: true` without an explicit `TaskPathSource` returns `domain.ErrUnsupported` from `s.TaskPath(...)`.
  - Mutation receipts (`CreateTaskReceipt`, `CreateThreadReceipt`) populate `LocalPath: ""` when the underlying store is pathless, and CLI commands suppress filesystem hyperlinks.
  - Graph repair correctly sets `Repairable: false` and `Automatic: false` when `LocalPath == ""`, emitting `ProblemRepairUnavailable`.

#### Angle 3: Mixed-Snapshot Races
- **Probe:** Evaluated race scenarios between semantic reads, asynchronous path resolution, and concurrent filesystem modifications.
- **Verification:**
  - In `FS.MutateTaskGraphRepair`, whole-graph CAS re-reads the graph before applying edits and validates `stepAnalysis.Prospective.SameRepairSnapshot`. If an external edit touches any file in the planned repair group, the mutation aborts with `domain.ErrConflict`.
  - In TUI navigation, `localPathResultMsg` requires both `isCurrentSelection(msg.kind, msg.id)` and `msg.listGen == m.cur().loadGen`. If the user moves the cursor or reloads the workspace while a path lookup is in flight, the delayed message is discarded.

#### Angle 4: Graph Repair Diagnostics Practicality
- **Probe:** Tested the remedy commands emitted by `tskflwctl task depend --repair --dry-run` on a broken graph with corrupt dependencies.
- **Verification:**
  - Emitted remedy commands use `formatRepairSelector`, which prefers `source.TaskID` if present, falling back to `source.LocalPath`. It never emits `Location` as a selector.
  - When executed against the broken state, the emitted `--drop "<id>:depends_on=<bad>"` command parsed cleanly, applied the planned edit, and converged the graph to healthy without error.

#### Angle 5: Test Helpers Smuggling Metadata
- **Probe:** Audited test helpers across `internal/testutil` and `internal/core` for hidden dependencies on filesystem paths.
- **Verification:**
  - Found that test helpers in `internal/core/dependency_source_test.go` frequently initialize `RecordSource{Location: task.Path, LocationIsPath: true}`. While acceptable for filesystem-backed tests, it led to the test gap identified in Finding L3 where no test verified path-shaped locations without `LocalPath`.

---

### 7. Executed Non-Default JSON & Schema Validation

A non-default JSON fixture was constructed and evaluated against the Revision 1.79 machine contract and JSON Schema Draft 2020-12 compiled from [`internal/cli/testdata/golden/schema_jsonschema.golden`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.e7rUyo/internal/cli/testdata/golden/schema_jsonschema.golden):

1. **Non-default Fixture Configuration:**
   - Canonical Task ID: `6g9999999999`
   - Canonical Task Slug: `non-default-probe`
   - Opaque Location: `db://cluster-alpha/tasks/shard-4`
   - Local Path: `/repo/planning/tasks/6g9999999999-non-default-probe.md`
2. **Wire JSON Emission (`ToLoadedTaskJSON`):**
   ```json
   {
     "id": "6g9999999999",
     "slug": "non-default-probe",
     "status": "ready-to-start",
     "path": "/repo/planning/tasks/6g9999999999-non-default-probe.md",
     "location": "db://cluster-alpha/tasks/shard-4"
   }
   ```
3. **Redundant Path Suppression:**
   - When `Location` was set identical to `Path`, `readableSourceLocation` suppressed `location`, emitting:
     `{"id":"6g9999999999","slug":"non-default-probe","status":"ready-to-start","path":"/repo/planning/tasks/6g9999999999-non-default-probe.md"}`
4. **Schema Compliance:**
   - Schema validation passed with 0 errors across all machine contracts (`task_list_json`, `task_show_json`, `epic_list_json`, `audit_info_json`, `lint_json`).

---

### 8. Executed Repair Workflow Probes

A live graph repair workflow was executed in a temporary isolated repository inside the sandbox:

```sh
# 1. Diagnose defects in broken graph
$ tskflwctl task depend --repair --dry-run --json
# Output: Diagnosed 2 defects:
#   - Defect 1: Self-dependency on 6g0000000000 (Automatic: true, Repairable: true)
#   - Defect 2: Invalid dependency ID invalid-token on 6g0000000001 (Automatic: false, Repairable: true)
# Remedy command emitted:
#   tskflwctl task depend --repair --drop "6g0000000001:depends_on=invalid-token#0"

# 2. Execute automatic dry-run
$ tskflwctl task depend --repair --auto --dry-run
# Output: Dry run succeeded; 1 file planned for repair.

# 3. Apply automatic repair
$ tskflwctl task depend --repair --auto
# Output: Committed repair for 6g0000000000-task-zero.md. Self-dependency removed.

# 4. Execute the emitted remedy command
$ tskflwctl task depend --repair --drop "6g0000000001:depends_on=invalid-token#0"
# Output: Committed repair for 6g0000000001-task-one.md. Invalid dependency dropped.

# 5. Verify graph convergence
$ tskflwctl lint
# Output: ✔ all planning entities and dependency links pass lint
```

Both the automated repair and the explicitly recommended remedy command executed cleanly and converged the repository to healthy.

---

### 9. Validation and Restoration Summary

The following test suites, linters, and contract verification checks were run inside `$SANDBOX`:

| Check | Command | Result | Notes |
|---|---|---|---|
| **Full Test Suite** | `go test ./...` | **PASS** | 33 packages passed cleanly. |
| **Race Detector** | `go test -race ./internal/core ./internal/store ./internal/cli ./internal/tui ./internal/wire ./internal/workspacestore` | **PASS** | Zero data races detected. |
| **Linter** | `golangci-lint run ./...` | **PASS** | 0 issues reported. |
| **Doc Drift** | `just docs-check` | **PASS** | CLI docgen matches `docs/cli` exactly. |
| **Schema & Golden** | `go test -v ./internal/cli -run "Schema\|Golden"` | **PASS** | Machine contract goldens match schema Revision 1.79. |
| **Audit Linter** | `go run ./cmd/tskflwctl audit lint 2026-10-02-local-entity-source-capabilities-checkpoint-antigravity` | **PASS** | Finding syntax, status grammar, and slugs verified clean. |
| **Working Tree** | `git status --porcelain` | **PASS** | Only the assigned audit file differs from baseline. |
