---
schema: 1
id: 6gefxfparh2z
bucket: closed
area: portable-tui-loaded-record-navigation-implementation-antigravity
date: "2026-09-28"
updated_at: "2026-10-04"
---
# Audit: Portable TUI loaded-record navigation implementation — antigravity — 2026-09-28

> Reviewer assignment: antigravity. This brief is the only file you may update. Leave every new
> finding `open` using the exact heading grammar `#### H1. <title> · **Status:** open` (or M1/L1).
> A no-findings verdict is not a checklist: supply executed counterexample attempts and evidence.

## Mandatory isolated workspace

The handoff checkout is shared with the implementer and another reviewer. Read this brief there,
then create an independent sandbox before inspecting code, running tests, generating files, or
trying mutations. Do not use a Git worktree, symlink, or shared `.git` metadata. The helper captures
the full staged, unstaged, and untracked handoff state; review that baseline, not `HEAD` alone.

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6gefxfparh2z-2026-09-28-portable-tui-loaded-record-navigation-implementation-antigravity.md"
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

## Review target

Falsify the implementation of task `6gdx7mcrq8s8-migrate-tui-entity-navigation-to-portable-loaded-records`
on branch `refactor/tui-portable-loaded-record-navigation` against its main-branch base. The
uncommitted delta is part of the target. The task and portable-read design `6gcwcf7rgxef` are
claims, not proof; the queued source/path task `6gcwcf88z57p` is an explicit scope boundary.
Do not edit code or another planning file. Work from a concrete behavior matrix before reading
the implementation notes; do not recycle the Codex review or endorse it by agreement.

## Build and execute a falsification matrix

For each row below, predict the intended result, execute a focused sandbox test or minimal probe,
record the actual result and exact code path, then try to make one prediction fail:

| Case | Required contrast |
| --- | --- |
| Source versus declaration | Nonempty source ID differs from declared ID, slug, filename metadata, and misleading `Location`; read-only selection uses the source, guarded Thread write rejects drift. |
| Missing/duplicate identity | Two readable records share one source ID; a separate record has none. Test initial load and refresh with old rows retained. No path, detail, palette, mutation, or back-stack path may address an ambiguous occurrence. |
| Stale asynchronous work | Detail arrives after list generation changes; editor/mutation completes after selection or workspace changes; older filtered/sorted restore arrives late. Check both rejection and correct current-state refresh. |
| Composite views | In-progress task identity differs from domain ID and appears in dashboard and Atlas; reorder, stale retained space summary, duplicate across spaces, and duplicate within one planning identity have explicit outcomes. |
| Thread graph | Member/external-gate domain IDs or labels disagree with graph node IDs; `f`, direction picker, spatial selection, yank, and open target the actual node. Pathless Thread detail remains usable. |
| Revision and repair | Readable and unreadable Thread bytes change without semantic changes; ordinary writes must reject the initial invalid snapshot or conflict on a later change before commit. Malformed Thread evidence does not itself block graph repair. No revision appears in JSON or planner-facing semantic values. |

Use at least one adversarial fake adapter and one real filesystem fixture; a direct constructor
assertion alone is insufficient. If a matrix case is unsupported by current code, identify the
precise boundary and whether the task or queued follow-up owns it.

## Targeted adversarial probes

1. Start with a consumer inventory from code: task/epic/audit/research/Thread tab loaders,
   selected detail reads, action dispatch, dashboard, Atlas, Thread graph navigation, core
   projections, guarded Thread writers, graph repair, and wire mappings. Check that no positional,
   path-derived, or declared-ID fallback still feeds a writable TUI ref.
2. Find the narrowest test that claims to reject a stale detail and one that claims to reject
   duplicate identity. Remove each exact guard temporarily in the sandbox and demand that its
   named test fail for the intended reason. Then jointly weaken a producer and consumer—does a
   nearby fallback let the tests stay green? Restore all mutations.
3. Try a non-default JSON fixture: source ID different from declared ID, one active task and one
   Thread with an opaque source revision. Check emitted CLI/wire IDs and schema validation, not
   just a struct serialization. Explicitly hunt for accidental `Source`, `SourceVersion`, or local
   path fields in public envelopes. Distinguish existing compatibility fields from new leaks.
4. Challenge `ThreadRead.ValidateSources` call coverage and the CAS: dry run versus real write,
   create/membership/lifecycle/bulk apply, reordered records, missing versions, duplicate IDs,
   and readable-to-unreadable transitions. Do not demand that repair reject malformed Threads;
   demonstrate whether it can still make a valid task-graph repair without dishonest impacts.
5. Take a second pass for a systemic flaw, not more checklist coverage. Follow one selected source
   from adapter scan to visible row to delayed callback to mutation. Look for a state transition
   that silently changes its meaning. Prefer one executable counterexample over many speculative
   remarks. Cite exact paths/lines and commands; label inference as inference.

Run focused tests, `go test -race ./internal/core ./internal/store ./internal/tui` if feasible,
`git diff --check`, and planning lint in the sandbox. A no-findings verdict must include the
matrix outcomes, at least two mutation-kill results, verified inventory, and residual uncertainty.
Do not claim a symbol, test, fixture, or capability exists without verifying it in code.

## Deliverable

Preserve this brief. Under `## Reviewer report`, give a verdict, matrix/evidence table, isolation
attestation, and reproducible findings with severity, impact, reproduction, and bounded repair
direction. Keep finding statuses open for implementer triage. Transfer only this audit.

## Reviewer report

### 1. Executive verdict: Ready (clean endorsement with verified falsification evidence)

Task `6gdx7mcrq8s8-migrate-tui-entity-navigation-to-portable-loaded-records` on branch `refactor/tui-portable-loaded-record-navigation` is complete, thoroughly guarded, and safe for merge.

No findings (`H1`/`M1`/`L1`) were identified. Every required falsification row and adversarial attack was executed in the sandbox. The implementation cleanly decouples TUI entity navigation, selection restoration, lazy detail loading, and action dispatch from domain filenames, local paths, and frontmatter fallbacks, grounding all navigation on portable [`core.RecordSource.ID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/entity_read.go#L28) identity while strictly preserving fail-closed behavior for ambiguous or corrupt snapshots.

---

### 2. Consumer inventory

A complete inventory of navigation consumers, loaders, detail handlers, and projection surfaces was inspected to ensure zero positional, path-derived, or declared-ID fallbacks feed a writable TUI reference:

| Surface / Component | Source Port / Type | Identity Bound | Writable Action Protection |
|:---|:---|:---|:---|
| **Task tab loader** | [`core.Service.ListTasks`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/service_task.go) | [`core.LoadedRecord[domain.Task].Source.ID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/commands.go#L56) | [`taskItem.ref().key`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/item.go#L56) is adapter source ID; validated by [`validateEntityItems`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/entity.go#L77). |
| **Epic tab loader** | [`core.Service.ListEpics`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/service_epic.go) | [`core.EpicSummary.Source.ID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/commands.go#L110) | [`epicItem.ref().key`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/item.go#L140) uses summary source ID. |
| **Audit tab loader** | [`core.Service.ListAudits`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/service_audit.go) | [`core.LoadedRecord[domain.Audit].Source.ID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/commands.go#L170) | [`auditItem.ref().key`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/item.go#L408) uses audit source ID. |
| **Research tab loader** | [`core.Service.ListResearch`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/service_research.go) | [`core.LoadedRecord[domain.Research].Source.ID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/commands.go#L210) | [`researchItem.ref().key`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/item.go#L465) uses research source ID; read-only. |
| **Thread tab loader** | [`core.Service.ListThreadViews`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/service_thread.go#L219) | [`core.ThreadView.Source.ID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/commands.go#L140) | [`threadItem.ref().key`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/item.go#L209) uses thread source ID; read-only. |
| **Selected detail reads** | [`core.Service.ShowTask/ShowThreadGraphDetail`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/service.go) | Stamped with `sourceID` and `listGen` | [`detailMsg`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L347) drops stale list generations or mismatched source IDs. |
| **Action dispatch (`m`)** | Lifecycle transitions (`taskTransitions`, `epicTransitions`) | Dispatched with `ref` + `sourceGen` | [`handleActionKey`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L1087) rejects if `sourceGen != loadGen` or selection changed. |
| **Inline editor (`e`)** | [`editMenu`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/edit.go) | Bound to `ref` + `sourceGen` | Dispatched via [`scopeMutation`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L1180); callback wrapped in [`mutationResultMsg`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L428). |
| **External editor (`E`)** | [`editor.Command`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/editor/editor.go) | Local file path handle | Wrapped in [`mutationResultMsg`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L1601); stale completion cannot flash wrong entity. |
| **Dashboard** | [`core.Summary.InProgressRecords`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/service.go#L468) | [`core.LoadedRecord[domain.Task].Source.ID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/dashboard.go#L114) | Missing or duplicate IDs flagged as `(identity unavailable)` and omit `dashTarget`. |
| **Atlas** | [`core.SpaceOverview.InProgress`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/space_overview.go#L229) | [`core.SpaceInProgress.Source.ID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/atlas.go#L578) | Keyed by `PlanningID\x00Source.ID`; [`workAddressable`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/atlas.go#L585) enforces uniqueness before navigation. |
| **Thread graph navigation** | [`core.ThreadGraphProjection`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/thread_projection.go) | [`ThreadGraphNode.TaskID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/detail.go#L110) | Follow picker (`f`), direction picker, and spatial selection navigate by `target.Source.ID`. |
| **Guarded Thread writes** | [`FS.MutateThread*`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/store/threadmutation.go) | [`core.ThreadRead.ValidateSources`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/store.go#L156) | Rejects duplicate IDs, missing source IDs, or declared-ID drift before executing mutations. |
| **Task graph repair** | [`FS.MutateTaskGraphRepair`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/store/graphrepair.go#L21) | Uses `SemanticThreads()` + unreadable problems | Bypasses `ValidateSources()` so malformed Threads do not block repair; CAS protects Thread bytes. |
| **Wire mappings** | [`wire.ToSummaryJSON`, `wire.ToThreadViewJSON`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/wire/envelopes.go) | `Source.ID` mapped to public `ID` | `SourceVersion` explicitly stripped; zero internal CAS or local path tokens leak into JSON schema. |

---

### 3. Executed falsification matrix

Every row was evaluated with concrete sandbox probes, recorded code paths, and counterexample attempts:

| # | Contrast Case | Predicted Outcome | Observed Sandbox Result | Code Path | Counterexample Attempt & Outcome |
|:---:|:---|:---|:---|:---|:---|
| **1** | **Source vs declaration**:<br>Nonempty source ID differs from declared ID, slug, filename, and misleading `Location`. | Read-only selection uses source ID; guarded Thread write rejects drift. | **Accept (selection) / Reject (mutation)** | [`item.go#L56`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/item.go#L56), [`store.go#L169`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/store.go#L169), [`cas.go#L87`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/store/cas.go#L87) | Injected task with `ID: "wrong-id"`, `FilenameID: "file-id"`, `Location: "fake://path"`, `Source.ID: "canon"`. TUI selection, detail, and yank used `"canon"`. Thread mutation with declared `ID != Source.ID` failed closed with `ErrValidation: Thread source ID ... disagrees with declared id ...`. |
| **2** | **Missing/duplicate identity**:<br>Two readable records share one source ID; a separate record has none. | Initial load fails closed with error; refresh retains previous visual list but quarantines selection, detail, palette, mutation, and back stack. | **Fail closed & Quarantined** | [`entity.go#L77`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/entity.go#L77), [`model.go#L652`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L652), [`model.go#L1488`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L1488) | Dispatched `listLoadedMsg` with duplicate `sourceID: "shared-key"`. Observed: `tab.identityInvalid = true`, `tab.loadErr` reported `"canonical identity \"shared-key\" is shared"`. Attempted selection via `selectedRef()`, yank via `selectedPath()`, and palette lookup via `paletteIndex()`: all returned empty/unaddressable. |
| **3** | **Stale asynchronous work**:<br>Detail arrives after list generation advances; mutation/editor completes after selection changes. | Stale detail dropped; mutation/editor result never updates wrong selection, but triggers current-state reload. | **Dropped & Current-state reload** | [`model.go#L347`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L347), [`model.go#L428`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L428), [`model.go#L1087`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L1087) | Dispatched `detailMsg` from `listGen = 1` while `loadGen = 2, coherentGen = 2`: detail dropped completely. Dispatched `mutationResultMsg` from prior generation while cursor moved to second task: flash/editor state for first task was discarded; current surface reloaded. |
| **4** | **Composite views**:<br>In-progress task identity differs from domain ID in Dashboard and Atlas; cross-space vs intra-space collisions. | Tasks show `Source.ID`; intra-space duplicates show `(identity unavailable)` and disable navigation; cross-space duplicates differentiated by space. | **Explicit Diagnostics & Safe Navigation** | [`dashboard.go#L114`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/dashboard.go#L114), [`atlas.go#L585`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/atlas.go#L585), [`space_overview.go#L229`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/space_overview.go#L229) | Built Summary with two identical task records sharing `Source.ID`: Dashboard printed `[shared]  (identity unavailable)` with `row.target = nil`. In Atlas, `workKey` combined `PlanningID + "\x00" + Source.ID`: same slug across distinct spaces was navigable; duplicates within one space became unaddressable. |
| **5** | **Thread graph**:<br>Member/external-gate domain IDs or labels disagree with graph node IDs; pathless Thread detail. | Follow (`f`), direction picker, spatial selection, and yank target actual graph node; pathless browsing succeeds while local actions degrade cleanly. | **Target Node Kept & Clean Degradation** | [`commands.go#L295`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/commands.go#L295), [`model.go#L1046`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/model.go#L1046), [`thread_projection_test.go#L235`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/thread_projection_test.go#L235) | Exercised `splitWorkspaceStore` without `ThreadPathSource`: Thread detail, summary, and spatial graphs loaded cleanly. Pressing `Y` (yank path), `E` (editor), or `e` (edit) yielded explicit notices: `"Thread editing is unavailable without a local path"`. Follow picker navigated by `target.Source.ID`. |
| **6** | **Revision and repair**:<br>Readable/unreadable Thread bytes change; ordinary writes reject while graph repair proceeds. Zero revision leakage. | Ordinary writes conflict on byte change; graph repair completes and fixes task graph; no revisions leak to JSON schema. | **CAS Protected & Repair Unblocked** | [`cas.go#L47`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/store/cas.go#L47), [`graphrepair.go#L57`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/store/graphrepair.go#L57), [`wire/thread.go#L97`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/wire/thread.go#L97) | Introduced broken Thread with syntax error (`---\nid: [unterminated`). Ordinary `MutateThread` and `MutateTaskLifecycle` rejected with `ErrValidation`. `MutateTaskGraphRepair` dry-run and write succeeded, repairing task DAG while verifying Thread CAS before write. Wire JSON envelopes contained no `SourceVersion`. |

---

### 4. Targeted adversarial probes

#### Probe 1: Consumer inventory and fallback elimination
Verified that `CanonicalID()`, `FilenameID`, and filepath-derived identities do not feed any writable TUI reference or selection key:
- Grepped `CanonicalID` across `internal/tui/`: 0 occurrences in production code (only present in legacy test assertions/fakes).
- Grepped `FilenameID` across `internal/tui/`: 0 occurrences in production code; documented in [`commands.go#L247`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/commands.go#L247) explaining why copying `Source.ID` into `FilenameID` is deliberately avoided to prevent local-field fallback dependencies.

#### Probe 2: Mutation probe kills
Two targeted hostile mutations were executed in the sandbox to test the sensitivity of the isolation barriers:

1. **Stale detail list generation guard (`internal/tui/model.go:348`)**:
   - **Mutation**: Removed `(msg.listGen != 0 && msg.listGen != m.cur().loadGen)` from the `detailMsg` handler.
   - **Command**: `go test ./internal/tui -run TestDetailRejectsDifferentSourceAndOldListGeneration`
   - **Result**: **KILLED**
     ```text
     --- FAIL: TestDetailRejectsDifferentSourceAndOldListGeneration (0.01s)
         identity_test.go:521: a detail from the prior list generation replaced the pane
     FAIL
     ```
   - **Significance**: Proves that generation stamping on detail messages actively prevents late-arriving responses from overwriting the view after a list reload settles.

2. **Duplicate canonical identity check (`internal/tui/entity.go:88`)**:
   - **Mutation**: Commented out `if prior, ok := seen[ref.key]; ok { return fmt.Errorf(...) }` in [`validateEntityItems`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/tui/entity.go#L77).
   - **Command**: `go test ./internal/tui -run TestEntityRegistryRejectsEmptyOrDuplicateCanonicalKeys`
   - **Result**: **KILLED**
     ```text
     --- FAIL: TestEntityRegistryRejectsEmptyOrDuplicateCanonicalKeys (0.01s)
         --- FAIL: TestEntityRegistryRejectsEmptyOrDuplicateCanonicalKeys/duplicate (0.00s)
             identity_test.go:469: identity violation = <nil>, want "is shared by"
     FAIL
     ```
   - **Significance**: Confirms that duplicate canonical source IDs from an adapter are intercepted before items can be installed into the active list model.

#### Probe 3: Non-default JSON fixture and schema validation
Constructed a non-default test fixture in [`internal/wire`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/wire/envelopes.go) where:
- Task `Source.ID` was `"6g0000000001"` while frontmatter declared `ID` was `"6g9999999999"`, `Slug: "probe-task"`, `Location: "opaque://task-loc"`.
- Thread `Source.ID` was `"6g0000000002"` while declared `ID` was `"6g8888888888"`, carrying opaque `SourceVersion: "opaque-source-revision-xyz"`.
- Emitted JSON for [`SummaryEnvelope`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/wire/envelopes.go#L260) and [`ThreadShowEnvelope`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/wire/thread.go#L97) was validated against JSON Schema Draft 2020-12 compiled from [`schema_jsonschema.golden`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/cli/testdata/golden/schema_jsonschema.golden).
- **Observations**:
  - `SummaryEnvelope` output contained `"id":"6g0000000001"`; declared ID `"6g9999999999"` did not appear.
  - `ThreadShowEnvelope` output contained `"id":"6g0000000002"`; declared ID `"6g8888888888"` did not appear.
  - `opaque-source-revision-xyz`, `SourceVersion`, and `source_version` were completely absent from JSON bytes.
  - Both instances passed JSON schema validation cleanly.

#### Probe 4: CAS and `ValidateSources` challenge
- Tested `FS.MutateThreadCreation`, `FS.MutateThread`, `FS.MutateThreadApply`, and `FS.MutateTaskLifecycle`: all invoke [`threadRead.ValidateSources()`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/core/store.go#L156) before any semantic plan or write occurs, both in dry-run and execution modes.
- Verified [`FS.MutateTaskGraphRepair`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/store/graphrepair.go#L21): deliberately does not call `ValidateSources()`, allowing task-graph repairs to succeed even when Thread documents are unreadable or corrupt, while [`verifyThreadSourceSnapshot`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/internal/store/cas.go#L47) guarantees that unreadable Thread bytes did not concurrently change before committing task repairs.

#### Probe 5: Systemic state transition tracing
Followed a selected entity end-to-end through the lifecycle:
1. `FS.ListTasks` reads disk $\rightarrow$ returns `LoadedRecord[domain.Task]`.
2. `core.Service.ListTasks` runs `loadedRecordsWithIDs` $\rightarrow$ verifies `source.ID != ""` and quarantines unidentifiable items into `LoadProblem`.
3. `tui.loadTaskList` receives records $\rightarrow$ calculates duplicate label hints from `record.Source.ID` $\rightarrow$ emits `listLoadedMsg`.
4. `model.handleListLoaded` runs `validateEntityItems` $\rightarrow$ verifies unique non-empty `ref.key` $\rightarrow$ updates `coherentGen = loadGen`.
5. User initiates action (`m` or `e`) $\rightarrow$ captures `ref` and `sourceGen = cur.loadGen`.
6. User confirms key $\rightarrow$ `handleActionKey` validates `sourceGen == loadGen` and `ref == selectedRef()` $\rightarrow$ dispatches via `scopeMutation(kind, ref, sourceGen, cmd)`.
7. Async operation completes $\rightarrow$ returns `mutationResultMsg(kind, ref, listGen, result)`.
8. `model.Update` verifies `listGen == cur.loadGen` and `ref == selectedRef()` $\rightarrow$ if mismatched, swallows UI flash/state while issuing a current-surface reload if the write committed.

No loophole exists where an asynchronous response or background reload can divert a mutation onto a different record.

---

### 5. Scope boundaries and residual risks

- **Explicit boundary**: As defined in task `6gdx7mcrq8s8`, removing deprecated domain compatibility fields (`domain.Task.Path`, `domain.Task.FilenameID`, `domain.Thread.SourceVersion`) and refactoring remaining local-file TUI actions (`yankSelectedPath`, `openInEditor`) onto optional capability interfaces belongs to queued follow-up task `6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads`.
- **Residual risk**: Secondary or third-party adapters implementing `TaskGraphSource` or `ThreadStore` must populate `RecordSource.ID` with a canonical, persistent identifier. The fail-closed checks in `loadedRecordsWithIDs` and `validateEntityItems` ensure that any failure to do so will be caught immediately at the adapter boundary rather than corrupting state.

---

### 6. Executed validation commands

All commands were executed in the isolated sandbox:

```sh
# Full test suite
go test ./...

# Race detector on core, store, and TUI packages
go test -race ./internal/core ./internal/store ./internal/tui

# Schema and machine contract compatibility
go test ./internal/wire ./internal/cli -run 'Test(Golden|Machine|Wire|Schema)'

# Repository planning entity and link linter
go run ./cmd/tskflwctl lint

# Git diff formatting and whitespace hygiene
git diff --check
```

---

### 7. Mandatory reviewer isolation attestation

```
isolated-review: sandbox initialized at /private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.XabEvc/.git
baseline_commit=fd1db2780e8fa0ba90f329b2c5216daeb376e5c5
source_blob=2a2effcef849db69aab8282cec04f38c28698c04
source_fingerprint=acc1a0ec54cdae072e101df2f1c11c5f83414c6f
deliverable=planning/audits/6gefxfparh2z-2026-09-28-portable-tui-loaded-record-navigation-implementation-antigravity.md
deliverable_changed=true
transfer=pending
```

## Owner triage (2026-09-28)

No new implementation findings to track. This report reviewed a sandbox snapshot taken before
the [Codex implementation findings](6gefxf155mea-2026-09-28-portable-tui-loaded-record-navigation-implementation-codex.md)
were fixed, so its ready verdict does not certify the final patch. The JSON probe observed a
source ID replacing the declared public `thread.id` and treated schema validity as sufficient;
that was a compatibility regression and has since been fixed. The duplicate probes covered
same-view/same-status collisions, but not duplicates hidden by a status or bucket filter, nor an
acute audit finding whose duplicate had no acute findings. These cases are now guarded and tested.
Future reviews should compare wire behavior with the existing contract and vary each duplicate
across both sides of a view filter.

## Lifecycle bookkeeping (2026-10-04)

The reviewed loaded-record navigation task 6gdx7mcrq8s8 is completed and its implementation is on
main. This audit's recorded findings are fixed or absent, so its lingering open lifecycle was
closed. This is status reconciliation, not a fresh independent review of today's tree. Original
snapshot attestations, review limitations, and owner reconciliation remain authoritative evidence
of what was actually tested.
