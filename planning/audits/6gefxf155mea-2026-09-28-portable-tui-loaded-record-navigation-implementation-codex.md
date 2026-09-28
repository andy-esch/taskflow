---
schema: 1
id: 6gefxf155mea
bucket: open
area: portable-tui-loaded-record-navigation-implementation-codex
date: "2026-09-28"
updated_at: "2026-09-28"
---
# Audit: Portable TUI loaded-record navigation implementation — codex — 2026-09-28

> Reviewer assignment: codex. This brief is the only file you may update. Leave every new finding
> `open` using the exact heading grammar `#### H1. <title> · **Status:** open` (or M1/L1).
> A no-findings report still needs concrete negative controls and residual risks.

## Mandatory isolated workspace

The handoff checkout is shared with the implementer and another reviewer. Read this brief there,
then create an independent sandbox before inspecting code, running tests, generating files, or
trying mutations. Do not use a Git worktree, symlink, or shared `.git` metadata. The helper captures
the full staged, unstaged, and untracked handoff state; review that baseline, not `HEAD` alone.

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6gefxf155mea-2026-09-28-portable-tui-loaded-record-navigation-implementation-codex.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" --deliverable "$AUDIT_REL" --print-path)"
cd "$SANDBOX"
```

Make all probes and report edits in `$SANDBOX`. Restore probes so the assigned audit is its only
delta, then run:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

Do not stage, commit again, reset, clean, or write to `$SOURCE_ROOT`; the sandbox baseline commit
created by the helper is the only permitted commit. If creation or transfer refuses, preserve the
sandbox and report the blocker—do not manually copy the report. Include the helper's workspace,
Git-dir, baseline, source fingerprint, deliverable, and transfer attestation in your report.

## Review target and boundaries

Review task `6gdx7mcrq8s8-migrate-tui-entity-navigation-to-portable-loaded-records` on branch
`refactor/tui-portable-loaded-record-navigation` against its main-branch base. The large uncommitted
delta is intentional. Treat checked acceptance criteria and green tests as hypotheses. Inspect the
task, the portable-read design task `6gcwcf7rgxef`, and the queued source/path split
`6gcwcf88z57p`; do not report a queued removal of domain `Path`, `FilenameID`, or `SourceVersion`
as though it were already shipped in this slice.

Map every production entity-tab loader, selected detail loader, navigation target, palette/back
stack, refresh restore, dashboard and Atlas jump, local edit/path action, and Thread graph child
selection. Trace their identities from `RecordSource.ID` through core projections and TUI messages
to the final lookup or mutation. Inspect `internal/core/entity_read.go`, `service.go`,
`service_thread.go`, `space_overview.go`, `store.go`; `internal/store/threadstore.go`, `cas.go`,
Thread guarded writers and graph repair; `internal/tui` loaders/model/actions/Atlas/Thread detail;
and explicit `internal/wire` mappings. Include tests and test fakes in the trace.

## Contracts to falsify

1. A canonical source ID, not declared ID, filename compatibility fields, slug, location, path,
   slice position, or a Thread graph display label, identifies an addressable record. Missing or
   duplicate IDs visibly fail/degrade the relevant TUI surface without leaving an actionable old
   row. Read-only diagnostics may still explain corrupt records.
2. A list/detail response or local-action result belongs to one selection and one list/session
   generation. Watcher reloads, tab switches, filtering/sorting, workspace changes, delayed
   details, and editor completion cannot attach it to a different record. A legitimate unique-ID
   selection survives reorder and refresh.
3. Thread and in-progress dashboard/Atlas source IDs come from the same read snapshot as their
   semantic values. Thread selected reads remain pathless-capable. Ordinary guarded Thread writes
   reject missing, duplicate, or declared-ID-drifted readable source identity; broken-graph repair
   deliberately continues past incomplete Thread evidence without weakening its CAS.
4. Opaque Thread source revisions remain in guarded envelopes only. Existing CLI and JSON wire
   fields, IDs, ordering, and machine revision are preserved by explicit mappings; a direct or
   accidental embedded revision does not escape in user-facing projections.

## Required evidence and hostile second pass

- Inventory each entity family and every TUI/CLI consumer changed by this delta. Name any remaining
  `CanonicalID()`/`FilenameID` or domain-path usage and prove whether it is identity, compatibility,
  local-action behavior, or deliberate downstream work. Do not rely on a simple grep verdict.
- Exercise a pathless adapter whose source ID differs from its declared/domain ID and whose
  `Location` is misleading. Confirm list, detail, Thread child navigation, dashboard/Atlas, and
  unavailability of local path actions as applicable. Try duplicate source IDs with distinct
  declared IDs, plus one source ID that changes between list and selected read.
- Reproduce at least two temporal races with controllable messages or hooks: a delayed detail
  across a watcher reload and a mutation/editor result across a selection or workspace change.
  Check the old coherent rows retained during a failed refresh are not actionable.
- For guarded Thread writes, challenge the validation call sites individually and the
  `ThreadRead`/CAS representation transitions. Probe readable revision changes, unreadable
  revisions, reordered reads, missing evidence, and a malformed Thread during graph repair.
- Inspect new tests for broad fake defaults or helpers that could make a wrong implementation
  pass. Mutate at least two exact guards claimed by tests, run the named tests, and restore the
  probes. Include a coordinated mutation where a neighboring fallback might mask a single-line
  change. Run focused tests, race tests if feasible, `git diff --check`, and planning lint.
- Take a second pass for systemic issues: duplicated projection state, hidden rereads, source-set
  mismatch, stale async state, or compatibility fallbacks that become writable identities. Prefer
  one demonstrated defect to speculative future-adapter objections. Cite exact sandbox paths/lines
  and command outcomes; distinguish observed behavior from design inference.

## Deliverable

Preserve this brief. Under `## Reviewer report`, provide a verdict, evidence/negative-control
summary, isolation attestation, and reproducible findings with severity, impact, reproduction,
and bounded repair direction. Do not edit implementation or any other planning document.

## Reviewer report

**Verdict: changes requested.** Two identity surfaces remain actionable when their source ID is ambiguous, and the Thread JSON projection changes its established `thread.id` value for a drifted readable record. I reviewed the captured handoff overlay in two passes; the later source/path split remains queued work, not a finding here.

### Pass 1: identity and consumer trace

| Surface | Identity path and result |
| --- | --- |
| Task, epic, audit, research tabs | Task graph `Records`, `ReadEpics`, `ReadAudits`, and `ReadResearch` supply `RecordSource.ID`; `internal/tui/commands.go:25-63,96-111,170-197,310-346` copies it into row refs and selected detail messages. `internal/tui/model.go:346-389,643-675,1366-1415` checks detail source, request generation, list generation, and tab/selection before installation. `internal/tui/entity.go:77-93` rejects missing or repeated IDs in the **received view**. Findings H1 and M1 describe the gaps before that check and outside this registry. |
| Thread tab and child links | One versioned `ThreadRead` feeds `ListThreadViews` (`internal/core/service_thread.go:207-252`); selected `ReadThread` supplies `ShowThreadGraphDetail` source and body (`:370-407`). `internal/tui/thread_projection.go:23-76` keys list and detail by those sources. Graph task IDs derive from task source IDs through `internal/core/service_task.go:225-267`; `internal/tui/nav.go:223-243` builds child refs from graph node IDs, not labels or paths. A temporary pathless adapter probe used a Thread source ID different from declared `Thread.ID`, a task source ID different from declared `Task.ID`, misleading `file://` locations, and no Thread path port: detail body loaded, child ref was the task source ID, and Thread path was empty. |
| Dashboard and Atlas | `Summary.InProgressRecords` and `SpaceInProgress.Source` come from the same task read as their semantic values (`internal/core/service.go:399-406,451-472`; `internal/core/space_overview.go:192-235`). The dashboard blocks absent/duplicate in-progress IDs (`internal/tui/dashboard.go:93-143`) and Atlas uses planning ID plus source ID, rejecting missing/duplicate work rows (`internal/tui/atlas.go:414-425,578-596`). Dashboard acute audit findings have no equivalent check (M1). |
| Navigation and local actions | Row refs flow through palette, back stack, refresh restore, and workspace cache (`internal/tui/palette.go`, `nav.go:248-375`, `entity.go:225-281`, `session.go:105-173`). Mutation results are scoped by ref and list generation (`internal/tui/action.go:119-137`, `model.go:431-452`); workspace session messages have a separate generation (`session.go:12-61`). Ordinary task/epic/audit/research `Path` remains a local-action compatibility field (`item.go`, `detail.go`, `model.go:1528-1545`); Thread path is optional and resolved separately. The pathless task probe selected the source key, loaded its detail, and returned an empty local path despite a path-shaped `Location`. The queued `6gcwcf88z57p` task owns removing these domain path/filename fields. |
| Guarded Thread and wire | `ThreadRead.ValidateSources` rejects missing, repeated, and declared-ID-drifted readable sources (`internal/core/store.go:157-177`). Ordinary call sites are Thread create, mutate, apply initial/final, and task lifecycle (`internal/store/threadcreation.go:58-68`, `threadmutation.go:50-60`, `threadapply.go:53-66,265-277`, `lifecyclemutation.go:49-59`). Graph repair deliberately skips that completeness gate but compares Thread snapshots before each write and after each durable prefix (`graphrepair.go:49-58,91-154`). `internal/store/cas.go:54-99` compares ordered-by-key readable and unreadable opaque revisions. `internal/wire/thread.go:27-33,87-126` maps fields explicitly, but its source override changes `thread.id` (M2). |

Remaining production `CanonicalID()`/`FilenameID` use in core graph and lint projections (`internal/core/service_task.go:190-239`, `service.go:629-779`) and domain `Path` use for local file actions or diagnostics are compatibility/deferred work in this slice. The production TUI row refs and action targets do not derive from those values. `ThreadRead.SemanticThreads` and service list/selected projections strip embedded `SourceVersion`; wire DTOs omit it. The core test fakes often supply a default source-set ID or make `CanonicalID()` equal source identity (`internal/core/service_thread_test.go:22-44`), so those defaults alone cannot prove pathless divergent identity. The temporary divergent-ID probes and filesystem duplicate probe supplied that missing control.

### Pass 2: falsification and negative controls

- Temporary filesystem fixtures placed two readable files under one source ID with distinct slugs and, in the second probe, distinct declared IDs; one was `in-progress`, the other `completed`. The default Tasks tab reported `rows=1`, `selectedKey=12jd5b71qwcj`, `identityInvalid=false`; switching to `:all` reported `identityInvalid=true` and `canonical identity ... is shared by ...`. This reproduces H1 against the real adapter. A synthetic Summary with two acute findings on the same audit source ID produced **two navigable dashboard targets** (M1). All temporary probe files were removed.
- A temporary pathless task adapter supplied `Source.ID=portable-source-id`, `Task.ID=frontmatter-id`, `FilenameID=filename-id`, and misleading `Location=file:///misleading/location.md`: the list selected the source ID, detail loaded `PORTABLE-BODY`, and local path was empty. Changing the selected read to `replacement-source-id` yielded `detail source identity differs from the selected record`. The pathless Thread/child probe described above returned one child keyed by task source ID. Existing `TestDashboardInProgressRowsCarryCanonicalDuplicateTargets`, `TestAtlasWorkRefusesMissingOrDuplicateSourceIdentity`, `TestSpaceOverviewWorkRetainsPortableTaskSource`, and `TestThreadRouteSurvivesSplitPathlessCapabilities` supply adjacent controls; the first two do not cover acute audit findings.
- Controllable message tests `TestDetailRejectsDifferentSourceAndOldListGeneration` and `TestRefreshInvalidatesOpenMutationMenuAndDelayedResult` passed: an old detail is dropped after list-generation advance, and a delayed write result cannot update another selected row. `TestAtlasDropsStaleWorkspaceResultsAndOldSessionMessages` passed for the workspace boundary. `TestEntityRegistryRejectsEmptyOrDuplicateCanonicalKeys` confirms failed-refresh retained rows are not addressable through selection or palette. Focused TUI and store versions of these tests passed under `-race`.
- Guarded tests `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed`, `TestThreadSourceSnapshotRejectsRepresentationAndIdentityChanges`, `TestGuardedMutationsRejectUnreadableThreadDocuments`, and `TestMutateTaskGraphRepairAllowsMalformedThreadsButCASProtectsTheirBytes` passed. A temporary direct CAS probe confirmed readable record reordering compares equal while missing or changed readable revisions conflict. Existing tests verify unreadable byte edits, unreadable reorder, and readable/unreadable transitions. The malformed-Thread graph-repair test permits the repair dry run and rejects a concurrent raw Thread edit before write.
- Exact guard mutation 1 disabled `model.go:351` source-ID comparison; `TestDetailRejectsDifferentSourceAndOldListGeneration` failed because the wrong detail landed. Mutation 2 removed the nonempty revision requirement from unreadable Thread comparison at `cas.go:96-99`; `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` failed because two unversioned snapshots compared equal. For a coordinated masking control, retaining the embedded revision in `threadstore.go:42-43` left the service-only `TestServiceThreadViewsDoNotPublishAdapterSourceRevisions` green, while `TestReadThreadsKeepsReadableRevisionOutsideSemanticValue` failed; additionally removing the list projection clearing at `service_thread.go:246` made the service test fail. All mutations were restored byte-for-byte.
- `go test ./internal/tui ./internal/core ./internal/store ./internal/wire` passed after probe removal. Focused TUI and store `-race` commands passed. `git diff --check`, planning audit lint, and isolated-workspace verification passed before transfer. Residual risk: the core/renderer identity checks assume a filtered view represents the full source namespace, and wire tests pin tag/task order and opaque revision omission but do not pin divergent declared/source Thread IDs.

#### H1. View filtering hides duplicate source IDs before registry validation · **Status:** fixed

**Impact.** A task with a duplicated canonical source ID is selectable, appears in action menus, and exposes local actions whenever its other occurrence is outside the active status view. The same shape applies to default/open audit filtering and epic view filtering. The filesystem writer may independently reject an ambiguous ID, but the TUI contract promises to quarantine the row before offering an action and a portable writer cannot rely on filesystem resolution.

**Reproduction.** Create `tasks/<same-id>-active-copy.md` with `status: in-progress` and `tasks/<same-id>-archived-copy.md` with `status: completed`, each with a distinct declared `id` or slug. Open the default Tasks tab: the sandbox probe observed one row, a nonempty `selectedKey`, and no identity error. Switch to `:all`: the same snapshot is rejected as duplicate. `internal/tui/commands.go:39-63` calls `dropArchivedRecords` before `handleListLoaded` reaches `validateEntityItems` (`model.go:643-675`); core status/bucket filters similarly discard occurrences (`internal/core/service_task.go:67-94`, `service_audit.go:75-94`).

**Repair direction.** Validate source-key uniqueness over the complete family snapshot before narrowing a TUI view, and carry an ambiguity diagnostic through filtering so every visible occurrence of an ambiguous ID is non-actionable. Cover duplicates straddling task statuses, audit buckets, and epic status views with distinct declared IDs.

**Resolution:** Validated complete one-read task, audit, and epic source sets
before TUI view filtering, so hidden duplicates quarantine retained rows.
Summary now also carries all task source IDs, preventing an archived sibling
from making Overview or Atlas in-progress jumps look unique. Cross-view and
composite-view regressions cover both.

#### M1. Acute audit findings can navigate a duplicate audit source · **Status:** fixed

**Impact.** The dashboard offers a direct jump for an acute finding even when its audit source ID is duplicated. If another occurrence is outside the default open audit bucket, the jump reaches the apparently unique but ambiguous row described in H1; otherwise it leads to a quarantined tab after advertising an actionable target.

**Reproduction.** A temporary `dashboard.setSummary` probe supplied two acute `AuditFinding` values with distinct labels and one `AuditID=shared-source`; it observed two navigable audit targets. `internal/core/service.go:429-449` copies source IDs from one audit snapshot into findings; `internal/tui/dashboard.go:208-211` unconditionally calls `nav` for every acute finding. The in-progress task and epic widgets on the same dashboard count IDs before adding targets (`dashboard.go:105-143,160-190`).

**Repair direction.** Carry audit-source ambiguity from the full Summary snapshot, including audits without acute findings, and render affected acute rows as non-navigable diagnostics. Test duplicate source IDs across audit buckets and one acute finding whose sibling audit has none.

**Resolution:** Summary now carries source IDs from the complete audit snapshot,
including closed audits without acute findings. The dashboard renders acute rows
with missing or duplicate audit identity as non-navigable diagnostics; a
cross-bucket regression pins this.

#### M2. Thread JSON overwrites the declared ID when source identity drifts · **Status:** fixed

**Impact.** `thread list/show/frontier --json` changes its established `thread.id` field from the document-declared ID to the adapter source ID for a readable drifted Thread. That is an externally visible ID change without a machine-contract revision, and it disagrees with `ToThreadJSON` on the same domain value. TUI navigation can still use `ThreadView.Source.ID` separately.

**Reproduction.** A temporary wire probe called `ToThreadViewJSON` with `Thread.ID=declared-thread-id`, `Source.ID=source-thread-id`, and a misleading location; it logged `wire="source-thread-id"`. `internal/wire/thread.go:27-33` maps the declared ID, then `:100-102` overwrites it. `git show origin/main:internal/wire/thread.go` confirms the override was absent at the main-branch base. Existing `TestToThreadJSONPreservesTagOrderAndCanonicalizesMembership` and pathless diagnostic wire tests do not assert this field under drift.

**Repair direction.** Keep the historical `thread.id` mapping for CLI/wire while retaining `ThreadView.Source.ID` for TUI routing. Add a divergent-ID wire regression across list, show, frontier, and projection envelopes; evolve the machine schema explicitly if a canonical source ID must become public.

**Resolution:** Restored the historical document-declared thread.id mapping in
ThreadView JSON while retaining Source.ID for TUI routing. A divergent-ID
regression covers list, show, frontier, and graph projections.

### Isolation attestation

- Workspace: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.dLeLa7`
- Resolved Git directory: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.dLeLa7/.git` (independent clone, not a worktree)
- Sandbox baseline commit: `ee575340f155bc5459db096e7ba66a4ac947c1a8`
- Captured source blob: `920b0284eee5854f18cabaca1d8133a605c95caf`
- Captured source fingerprint: `acc1a0ec54cdae072e101df2f1c11c5f83414c6f`
- Sole deliverable: `planning/audits/6gefxf155mea-2026-09-28-portable-tui-loaded-record-navigation-implementation-codex.md`
- Transfer result: `succeeded` through `scripts/isolated-review-workspace.sh transfer` after verification; sandbox retained for owner confirmation.
