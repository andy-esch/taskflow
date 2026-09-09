---
schema: 1
id: 6g8bb1mf9hw0
bucket: closed
area: spatial-thread-graph-prototype-implementation-antigravity
date: "2026-09-09"
updated_at: "2026-09-09"
---
# Audit: Spatial Thread graph prototype implementation — antigravity — 2026-09-09

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

Perform an independent adversarial implementation and architecture review of the merged
two-dimensional navigable Thread graph prototype. Treat the renderer, its tests, and its claimed
adapter boundary as suspect until hostile evidence demonstrates that the picture, navigation, and
shell state all remain faithful to the supplied graph projection. Do not merely restate the task or
confirm that the existing suite passes.

Distinguish a demonstrated defect from an acknowledged prototype limitation already assigned to a
follow-up. A limitation may still be a release or safety concern if the current UI presents it as
trustworthy, but do not manufacture duplicate findings for explicitly bounded future work.

## Review target

Review PR #217 as merged on `main` at `c40d09d`, using commit range `c4a06f0..c40d09d`. The review
target includes:

- `internal/tui/thread_spatial.go`
- the optional detail-presentation seams in `internal/tui/detail.go`
- immersive shell and key routing in `internal/tui/model.go`, `internal/tui/nav.go`, and
  `internal/tui/view.go`
- color/style integration in `internal/tui/style.go` and help text in `internal/tui/help.go`
- the spatial and neighboring Thread tests in `internal/tui/thread_projection_test.go`
- task `6g6dw5js81f3`, ADR-0006's 2026-09-08 amendment, the Threads/TUI epics, and the follow-ups
  `6g86016qvk5d`, `6g86c7y6hn41`, `6g86e10dztpf`, `6g86g03zfj2f`, and `6g87qn72901g`

The source checkout may contain an owner-authored, uncommitted lifecycle update marking the
prototype task completed. That metadata is not implementation evidence and is outside the target;
preserve it through the sandbox copy and do not edit it. Review committed code and planning claims,
then update only this assigned audit.

Build a consumer inventory for every new optional detail-content interface and every added shell
state field/helper. Identify all implementations, type assertions, event routes, reload paths, and
future-adapter claims. Verify that optional behavior cannot silently affect ordinary entity detail,
single-pane navigation, manual zoom, Atlas, modal routing, or non-Thread content.

## Intended contract to challenge

The TUI adds a third Thread presentation, `summary → topology → spatial`, as an immersive,
read-only adapter over the supplied `core.ThreadGraphProjection`. Core owns task identity, node and
edge semantics, member waves, roles, health, and topology completeness. The TUI may derive bounded
terminal placement and routes but must not infer readiness, authorize work, mutate planning data,
or make renderer coordinates part of a portable contract.

Layout and aliases are deterministic for equivalent projection evidence. Dependencies point from
prerequisite to dependent. External gates may occupy a late valid presentation layer before their
first dependent without altering core member-wave labels. `h`/`l` prefer actual incoming/outgoing
edges, including skipped columns; `j`/`k` move within the current presentation column. Canonical
task ID owns selection through reload, add/rename, task open, and `ctrl+o` return. `y` copies the
highlighted task in structured Thread views. Esc retreats spatial to waves, `v` continues the view
cycle, and immersive zoom does not corrupt prior manual zoom or focus state.

The renderer must never silently draw a partial oversized graph. Narrow and capacity-limited views
remain explicit and offer the complete wave reader/task picker. Renderer failure is display-only.
Status color remains semantic across selection; focus and edge direction use separate visual
signals. Known production follow-ups own dense-route truthfulness, one-hop focus, reusable Back,
structured action targeting, and alternate-view/list discoverability.

## Mandatory evidence floor

Inventory the complete data path from `Service.ThreadGraphProjection` through the Thread registry,
detail payload, shell type assertions, spatial layout, canvas, clipping, footer, and task
navigation. Cite exact paths/lines for every claimed invariant and distinguish core-supplied facts
from TUI-derived presentation values.

Run the full race suite and lint, plus focused Thread/TUI tests. Exercise a real local build against
the repository Thread at multiple terminal sizes when the environment supports a PTY. Inspect ANSI
cell width and clipping with Unicode, combining marks, hostile control text, long IDs/labels, empty
labels, and all status/role/health combinations.

Mutation evidence is mandatory. Temporarily remove or invert individual production guards and
prove the named focused regression fails for the intended reason. At minimum challenge:

- canonical-ID selection and reload restoration;
- graph-first horizontal navigation across a populated intermediate column;
- external-gate late placement and retained member-wave labels;
- the node-free long-edge track;
- the node/edge/canvas capacity checks before dense canvas allocation;
- immersive-entry/retreat state and `ctrl+o` presentation restoration;
- structured-detail `y` targeting; and
- deterministic ordering when projection nodes, waves, and edges are permuted.

If no existing regression kills a mutation, record the gap even when manual behavior appears
correct. A compile error, broad golden drift, or unrelated test failure is not a valid mutation
kill. Restore the sandbox baseline after every probe.

## Required hostile angles

1. Permute nodes, edges, member waves, map construction order, and equal-distance fan-in/fan-out.
   Look for nondeterministic aliases, rows, columns, navigation targets, connector geometry, and
   inspector ordering.
2. Stress external gates: source gates, gates between member waves, a gate feeding several distant
   waves, gate-to-gate edges, multiple gates sharing intervals, and incomplete/cyclic residue.
   Determine whether presentation placement can reverse an edge or mislabel a core wave.
3. Stress navigation independently from appearance. Confirm `h`/`l` always choose a real edge when
   one exists in that direction and that fallbacks do not jump across unrelated graph structure.
   Challenge ties, isolated nodes, reverse/cyclic residue, unreadable nodes, and disappearing tasks.
4. Try to make the picture lie: long edges through nodes, shared lanes, crossings near arrowheads,
   same-row and backwards edges, clipping at all four viewport boundaries, and selected incident
   routes. Separate defects already honestly bounded by `6g86e10dztpf` from unqualified falsehoods
   in the current display.
5. Exercise shell transitions from split, single-pane, and manually zoomed layouts. Cycle views,
   press Esc/q/h/z, open a task, return, switch tabs, enter Atlas, resize, reload, fail a detail read,
   and remove the selected node. Look for stranded zoom, stale focus, wrong selection, or behavior
   leaked to ordinary detail pages.
6. Challenge cell and resource safety: zero/negative/tiny dimensions, huge width×height products,
   integer overflow assumptions, 512/513 nodes, 2,048/2,049 edges, wide runes, combining marks,
   embedded ANSI/control characters, and unusually long relationship lists. Confirm bounding occurs
   before material allocation or expensive route construction where claimed.
7. Audit semantic isolation. Search for readiness/status/role recomputation, graph traversal beyond
   supplied edges, filesystem/store access from TUI rendering, or core imports of TUI/layout types.
   Also test whether incomplete projection health is preserved rather than normalized away.
8. Audit action identity. Verify `Enter`, `y`, task picker, `ctrl+o`, reload, and duplicate labels use
   canonical IDs at authority boundaries and do not confuse visible aliases/slugs with identity.
9. Examine test helpers and assertions for systemic blind spots: ANSI-stripped snapshots that cannot
   prove color distinctions, substring assertions that tolerate false connectors, fixtures whose
   insertion order matches the desired order, and mocks that skip asynchronous shell behavior.
10. Perform a second pass looking beyond the named checklist: abstraction leakage, quadratic or
    dense memory behavior below the advertised thresholds, accidental global key precedence,
    stale async state, malformed projections, and comments/planning claims stronger than code.

For each surviving finding, provide a minimal reproduction or precise code path, impact, why the
current tests miss it, and the smallest sound correction. State whether it belongs in the current
prototype closeout or an existing/new follow-up.

## Validation and restoration

Perform all builds, PTY sessions, fixtures, code mutations, and generated artifacts inside the
mandatory isolated sandbox. Use sandbox-local caches where possible. Run and report exact results
for `go test -race ./...`, `golangci-lint run ./...`, `tskflwctl lint`, `git diff --check`, focused
TUI tests, and any visual/PTY probes. Report unsupported checks and their exact blocker.

Before transfer, restore every mutation and scratch artifact to the sandbox baseline. The assigned
audit must be the only remaining change. Use the isolation helper's verify and guarded transfer
commands; do not manually copy results back.

## Deliverable

Update only your assigned audit and preserve this brief. Add findings using the exact repository
grammar, leaving each new finding `open` for implementation-owner triage. Include severity,
component, effort, urgency, exact evidence, mutation/probe results, affected acceptance criteria,
and a minimal recommendation. Reference an existing follow-up when it already owns the issue rather
than duplicating scope.

If no finding survives both passes, write a substantive no-findings report with the consumer
inventory, exact mutations killed, visual/PTY cases exercised, commands run, residual risks, and
why each residual risk is honestly bounded.

## Reviewer report

### Attestation of review workspace

The isolated workspace helper was executed in compliance with the mandatory protocol:

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/.git
baseline_commit=05fc4d47c46856a6ef75ff57f52f82d21209f979
source_blob=ea76596afebb77b1f850abce6ced0739e3eee611
source_fingerprint=babfac36414b5dda3dc7c9bceacb255d1683eadf
deliverable=planning/audits/6g8bb1mf9hw0-2026-09-09-spatial-thread-graph-prototype-implementation-antigravity.md
deliverable_changed=true
transfer=pending
```

All investigations, race-detector tests, linters, PTY display probes, and coordinated code mutations were performed inside this isolated workspace. No changes were made in the shared repository root.

---

### Executive summary

PR #217 introduces the third Thread presentation mode (`summary → topology → spatial`) as an immersive, read-only TUI adapter over `core.ThreadGraphProjection`. Core graph semantics, node/edge identities, wave assignments, task readiness, and repository storage remain strictly unpolluted by renderer coordinates or terminal layout logic.

The implementation successfully fulfills several key architectural goals:
1. **Semantic decoupling:** Core code does not import TUI layout constructs; TUI code derives terminal coordinates without inferring scheduling readiness or modifying planning entities.
2. **Deterministic graph-first navigation:** Directional navigation (`h`/`l`) correctly prioritizes actual prerequisite/dependent DAG edges across visual columns over unrelated adjacent geometric boxes.
3. **Capacity & fail-closed bounds:** Visual allocation is rigorously bounded before materializing the terminal cell matrix (capped at 512 nodes, 2,048 edges, and 750,000 cells), cleanly failing open to the wave reader.
4. **Canonical identity tracking:** Structured selection, clipboard yanking, Enter-to-open, and `ctrl+o` history faithfully track canonical task IDs rather than visible slugs or aliases.

However, two adversarial passes identified three open findings:
- **[H1]** An active spatial view loses its `immersiveZoom` flag when transitioning into and out of Atlas (`keys.Atlas`), permanently stranding the user in full-screen zoom upon retreating from the spatial graph.
- **[M1]** Node row placement, gate aliases, and inspector connections are nondeterministic under permutation of projection evidence because sorting logic relies on the slice index of `projection.Nodes` rather than topological wave indices and canonical IDs.
- **[L1]** Cascaded external gate chains exhibit order-dependent column placement during the single-pass late-gate optimization.

---

### Consumer inventory & architectural invariants

PR #217 adds five new optional interfaces to `internal/tui/detail.go` and introduces shell-level immersion synchronization in `internal/tui/model.go`:

| Interface / Field | Location | Implementation | Consumers / Type Assertions | Behavior & Fail-Closed Guard |
| :--- | :--- | :--- | :--- | :--- |
| `retreatingDetailContent` | [`detail.go:49`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L49) | `threadDetail` ([`detail.go:819`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L819)) | `detailPane.retreatView` ([`detail.go:217`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L217)) | Used on `Esc` and `q` from immersive views. If content does not implement this, returns `false`, allowing normal pane unzoom or list focus. |
| `sizedDetailContent` | [`detail.go:58`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L58) | `threadDetail` ([`detail.go:847`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L847)) | `detailPane.render` ([`detail.go:169`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L169)), `detailPane.SetSize` ([`detail.go:352`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L352)) | Height-aware rendering. Non-implementing details fall through to standard `joinDetail(meta, markdown)` rendering. |
| `immersiveDetailContent` | [`detail.go:68`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L68) | `threadDetail` ([`detail.go:843`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L843)) | `detailPane.immersive` ([`detail.go:230`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L230)) | Informs shell when content desires full content region. Only returns `true` when `view == threadDetailSpatial`. |
| `directionalDetailContent` | [`detail.go:90`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L90) | `threadDetail` ([`detail.go:849`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L849)) | `detailPane.directionalSelectionAvailable` ([`detail.go:235`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L235)), `detailPane.moveSelectionDirection` ([`detail.go:278`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L278)) | Gates `h`/`j`/`k`/`l`/arrow key routing in detail focus. Non-implementing views retain standard viewport scrolling and list-return behavior. |
| `yankableDetailContent` | [`detail.go:100`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L100) | `threadDetail` ([`detail.go:915`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L915)) | `detailPane.selectionYankRef` ([`detail.go:301`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/detail.go#L301)) | Dispatches clipboard `y` to the selected child task in topology and spatial views. Fails closed to parent entity slug on summary or list focus. |
| `m.immersiveZoom` | [`model.go:81`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/model.go#L81) | N/A (shell state) | `model.syncDetailImmersion`, `model.toggleZoom`, `model.unzoom` | Distinguishes automatic full-screen expansion from explicit user zoom (`z`). Wiped cleanly on tab switches and unzoom events. |
| `m.pendingDetailNavigation` | [`model.go:91`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/model.go#L91) | N/A (shell state) | `model.restoreDetailNavigation`, `model.update(detailMsg)` | Restores presentation view and child selection after async `ctrl+o` return jumps. Consumed on exact entity match; dropped on failure/redirect. |

Verification confirmed that no optional detail seam alters behavior for ordinary entity details (tasks, epics, audits, or dashboard). All type assertions return false for non-Thread content.

---

### Data path inventory

The complete data pipeline from core storage through TUI rendering and interaction was verified against the codebase:

1. **Projection Acquisition:**
   - [`core.Service.ShowThreadGraphDetail(slug)`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/core/thread.go#L90) loads domain tasks and thread definition, invoking [`core.ProjectThreadGraph`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/core/thread_graph.go#L51).
   - Core computes `ThreadGraphNode`s, induced `ThreadGraphEdge`s, and `ThreadGraphWave`s. Nodes are sorted by `TaskID` and edges by `(From, To)`.
2. **TUI Registry & Detail Payload:**
   - `threadTab` receives the Thread view. `detailMsg` is dispatched asynchronously and received in [`model.update`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/model.go#L346).
   - If `pendingDetailNavigation` matches the loaded entity, [`restoreDetailNavigation`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/nav.go#L293) applies the saved view name (`spatial`) and canonical task ID selection.
   - [`syncDetailImmersion`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/model.go#L1329) promotes the pane to full-screen (`m.zoom = true`, `m.immersiveZoom = true`).
3. **Spatial Layout Construction:**
   - [`buildThreadSpatialLayout`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/thread_spatial.go#L51) computes topological layers via Kahn's algorithm in `rankThreadSpatialColumns`.
   - `placeThreadSpatialExternalGates` shifts source-only boundary gates to the latest column before their earliest dependent.
   - Column labels are derived via `labelThreadSpatialColumns`, retaining core wave numbering and annotating external gate layers.
4. **Canvas Rendering & Clipping:**
   - [`renderThreadSpatial`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/thread_spatial.go#L667) checks narrow constraints (`width < 60 || height < 12`) and prototype capacity (`nodes > 512 || edges > 2048 || cells > 750,000`).
   - If bounded, allocates a `threadSpatialCanvas` matrix ([`thread_spatial.go:493`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/thread_spatial.go#L493)).
   - Edges are drawn with elbows, tees, and direction arrowheads (`▶`/`◀`); long edges across skipped columns route via dedicated node-free horizontal tracks (`trackY`).
   - Node boxes are drawn in semantic status colors, with selected nodes expanded inside a reserved 5-row slot and indicated by a yellow focus pointer (`›`).
   - Viewport panning (`panX`, `panY`) centers on the selected node and cuts lines using `ansi.Cut` and `ansi.Truncate`.
5. **Fixed Inspector & Footer:**
   - Bottom 5 rows render the palette-accented inspector box showing task status, identity, state, description, and direct dependencies (`needs`/`unlocks`).
   - Shell footer advertises `hjkl node`, `⏎ open`, `v summary`, and `esc waves`.
6. **Interaction & Navigation:**
   - Directional keys invoke [`threadSpatialMove`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/thread_spatial.go#L328). `h`/`l` prioritize direct DAG edges across columns before geometric fallbacks.
   - `Enter` triggers [`openDetailSelection`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/nav.go#L207), saving the current Thread and selection on `navStack` via `pushLoc` and jumping to `entityTasks` by canonical ID.
   - `ctrl+o` pops `navStack`, restores `pendingDetailNavigation`, and asynchronously returns to the identical spatial node.

---

### Verification and test execution

The review suite was executed inside the isolated sandbox with clean results across all project checks:

| Command | Working Directory | Result | Notes |
| :--- | :--- | :--- | :--- |
| `go test -race ./...` | `$SANDBOX` | **PASS** (10 packages passed, 0 failures) | Full repository race suite clean |
| `golangci-lint run ./...` | `$SANDBOX` | **PASS** (0 issues) | Lint clean |
| `tskflwctl lint` | `$SANDBOX` | **PASS** | Planning repository entity graph and links valid |
| `git diff --check` | `$SANDBOX` | **PASS** (clean output) | No whitespace or merge conflict markers |
| `go test -run TestThreadSpatial -v ./internal/tui` | `$SANDBOX/internal/tui` | **PASS** (10/10 tests passed) | All spatial graph unit tests pass |

---

### Visual and terminal display probes

Probes were conducted against both the hostile synthetic projection and the committed repository dogfood Thread (`planning/threads/6g503c6pfqeb-complete-production-threads.md`):

1. **Terminal Geometry Range:**
   - Tested dimensions `50×10` (narrow fallback), `80×24` (standard VT100), `120×30` (default split), `140×36` (standard dogfood), and `200×50` (wide display).
   - In narrow mode (`50×10`), the layout gracefully degrades to an explicit explanatory banner naming the minimum required dimensions (`60×12`) and maintains the selected-node inspector within terminal boundaries without panics.
   - Across all supported dimensions, total rendered line count never exceeds the available viewport budget (`height - fixedRows` + chrome), and no ANSI lines overflow terminal width.
2. **Unicode & Hostile Character Handling:**
   - Evaluated `hostileThreadGraphProjection()` containing wide CJK runes (`-根`, `移行`), embedded newlines/carriage returns, tabs, and escape sequences (`\x1b[31m`).
   - `terminalText` properly sanitized hostile control sequences into `↵`, `⇥`, and unicode replacement marks (``), preventing terminal injection and unintended row height expansion.
   - Combining marks were properly coalesced in `putText` without advancing the cell cursor, and wide characters correctly allocated 2 cells with a following continuation marker skipped during string assembly.
3. **Color Contrast & Focus Separation:**
   - Selected node borders retain their semantic domain status colors (`theme.Status(node.Status).Color`). Focus is communicated independently via geometry, the yellow pointer (`›`), and palette accent highlights on incident connector strokes.

---

### Mandatory mutation probes

All 8 mandatory mutation targets specified in the brief were executed by temporarily modifying production guards and asserting that the corresponding regression test failed for the exact intended reason:

| Target Invariant | Production Guard Modified | Targeted Test | Mutation Result & Kill Reason |
| :--- | :--- | :--- | :--- |
| **Canonical-ID selection & reload restoration** | Inverted `threadSpatialSelectedTaskIDInLayout` to always return `firstSpatialSelectable` ([`thread_spatial.go:401`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/thread_spatial.go#L401)) | `TestThreadSpatialReloadAddsAndRenamesNodesWithoutLosingSelection` | **KILLED:** Failed with `setup spatial selection="7kv4ra5f3jvp" want "pptcpta1wd8b"` |
| **Graph-first horizontal navigation** | Bypassed `if len(direct) > 0` in `threadSpatialMove` ([`thread_spatial.go:357`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/thread_spatial.go#L357)) | `TestThreadSpatialHorizontalNavigationPrefersGraphEdgesAcrossVisibleColumns` | **KILLED:** Failed with `second h did not follow the prerequisite edge: got "wyse8m6wg3p3" want "9sf0vn4syr95"` |
| **External gate late placement** | Disabled `placeThreadSpatialExternalGates` call in `buildThreadSpatialLayout` ([`thread_spatial.go:75`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/thread_spatial.go#L75)) | `TestThreadSpatialLayoutPullsSourceGateBesideItsFirstDependent` | **KILLED:** Failed with `source gate column=0 want immediately before dependent column 1` |
| **Node-free long-edge track** | Disabled `if to.column-from.column > 1` branch in `drawThreadSpatialEdge` ([`thread_spatial.go:784`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/thread_spatial.go#L784)) | `TestThreadSpatialLongEdgeUsesNodeFreeTrack` | **KILLED:** Failed with `selected long edge did not use the node-free inter-row track` |
| **Capacity checks before canvas allocation** | Bypassed `threadSpatialCapacityIssue` check in `renderThreadSpatial` ([`thread_spatial.go:679`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/thread_spatial.go#L679)) | `TestThreadSpatialGraphFailsOpenToWavesBeyondPrototypeCapacity` | **KILLED:** Failed with `capacity fallback omitted "bounded prototype fallback"`, as oversized matrix was allocated |
| **Immersive-entry zoom & sync** | Bypassed `m.zoom = true` in `syncDetailImmersion` on entry ([`model.go:1337`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/model.go#L1337)) | `TestThreadSpatialGraphUsesDirectionalStableIdentityNavigation` | **KILLED:** Failed with `spatial entry = view "spatial" zoom=false focus=1` |
| **`ctrl+o` presentation restoration** | Commented out `withDetailView` restore in `restoreDetailNavigation` ([`nav.go:304`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/nav.go#L304)) | `TestThreadSpatialGraphUsesDirectionalStableIdentityNavigation` | **KILLED:** Failed with `ctrl+o did not restore spatial context: kind=2 view="summary" selected="pptcpta1wd8b" zoom=false` |
| **Structured-detail `y` targeting** | Bypassed `if m.focus == focusDetail` check in `selectedYankRef` ([`model.go:1368`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/internal/tui/model.go#L1368)) | `TestThreadStructuredDetailYankCopiesHighlightedTask` | **KILLED:** Failed with `topology yank = "copied slug: delivery" cmd=true, want highlighted task` |
| **Deterministic ordering under permutation** | Permuted `projection.Nodes` in `hostileThreadGraphProjection()` | N/A (Gap recorded) | **GAP CONFIRMED:** Existing test suite lacked permutation checks; probe revealed layout row reordering across columns ([M1](#m1-spatial-layout-row-placement-aliases-and-inspector-connections-are-nondeterministic-under-projection-permutation--status-open)) |

All mutations were restored to the clean sandbox baseline immediately following execution.

---

### Hostile angle evaluations

1. **Permutation & Nondeterminism:**
   - Reversing `projection.Nodes` alters `order[node.TaskID]`, altering row placement in Kahn's algorithm and gate numbering in `threadGraphAliases`. Described in **[M1]**.
2. **External Gate Stress:**
   - Source gates correctly advance to late presentation columns before their first dependent without altering member wave labels. However, cascaded multi-gate intervals evaluate in arbitrary single-pass order. Described in **[L1]**.
3. **Navigation vs. Appearance Independence:**
   - `h`/`l` always select true causal prerequisites/dependents first, skipping intermediate populated columns when an edge spans multiple layers. When multiple edges qualify, the nearest column wins, with ties broken by geometric row distance. Fallback column navigation only occurs when no DAG edges exist in that direction.
4. **Visual Truthfulness & Route Overlaps:**
   - Multi-column edges utilize reserved inter-row tracks (`trackRow`), preventing false intermediate node crossings.
   - Known limitations regarding shared lane crossing ambiguities, multiple incident routes, and same-row/reverse cycles are explicitly and honestly bounded by tracked follow-up `6g86e10dztpf`.
5. **Shell Transitions, Zoom, and View Retreats:**
   - Tested split view, full-screen zoom, view cycling (`v`), Esc, q, and task drill-down.
   - Identified that visiting Atlas clobbers `m.immersiveZoom`, causing `Esc` retreat to fail to restore the two-pane split. Described in **[H1]**.
6. **Cell & Resource Safety:**
   - Handled zero/negative terminal dimensions, wide runes, and combining characters safely.
   - Capacity guards trigger before matrix allocation when node count > 512, edge count > 2048, or total cells > 750,000, preventing unbounded memory consumption.
7. **Semantic Isolation:**
   - No filesystem access, database reads, or task readiness calculations occur inside TUI rendering. Graph health diagnostics (`GraphHealth`, `ProjectionHealth`, `Inconsistent`) pass directly through to the header.
8. **Action Identity:**
   - `Enter`, `y`, follow picker, and `ctrl+o` dispatch based on canonical task IDs. Duplicate labels and unreadable node IDs do not result in erroneous task navigation.
9. **Test Suite Blind Spots:**
   - Tests previously relied on `core.ProjectThreadGraph`'s sorted node slice, masking that `buildThreadSpatialLayout` was sensitive to `projection.Nodes` slice ordering.
10. **Second-Pass Systemic Synthesis:**
    - The seam between generic bubbletea shell zoom (`m.zoom`) and presentation-driven immersion (`m.immersiveZoom`) works well for single-pane drills and view cycles, but requires explicit preservation in cross-cutting shell modes like Atlas.

---

### Findings

#### H1. Atlas navigation wipes immersive zoom state and strands full-screen layout on retreat · **Status:** tracked by 6g8by30btznq

- **Severity:** High
- **Component:** `internal/tui/atlas.go`, `internal/tui/model.go`
- **Effort:** 1–2 hours
- **Urgency:** High
- **Affected Acceptance Criteria:**
  - Task `6g6dw5js81f3`: "Specify selection, Enter-to-open, back-stack behavior, scrolling/panning, zoom, narrow-terminal fallback, reload stability, and accessibility without conflicting with existing TUI keys"; "Esc returns to waves".
  - ADR-0006 2026-09-08 Amendment: "Esc retreats spatial to waves, v continues the view cycle, and immersive zoom does not corrupt prior manual zoom or focus state."
- **Exact Evidence:**
  - In `internal/tui/atlas.go` (line 305):
    ```go
    type atlasResume struct {
        focus focus
        zoom  bool
        set   bool
    }
    ```
    The `atlasResume` struct records `focus` and `zoom`, but omits `immersiveZoom`.
  - In `internal/tui/atlas.go` (line 317):
    `m.atlasResume = atlasResume{focus: m.focus, zoom: m.zoom, set: true}`
    Line 321 then calls `m.unzoom()`, which unconditionally resets `m.immersiveZoom = false` (`internal/tui/model.go` line 1151).
  - In `internal/tui/atlas.go` (line 360):
    `exitAtlas` restores `m.zoom = m.atlasResume.zoom` (`true`), but leaves `m.immersiveZoom = false`.
  - In `internal/tui/model.go` (lines 760–764, 937–941):
    When the user retreats from spatial view (via `Esc`, `q`, or `v`), `m.syncDetailImmersion(wasImmersive)` is invoked with `wasImmersive = true`.
  - In `internal/tui/model.go` (lines 1329–1349):
    ```go
    func (m *Model) syncDetailImmersion(wasImmersive bool) {
        nowImmersive := m.detail.immersive()
        switch {
        case nowImmersive && !m.zoom:
            m.zoom = true
            m.immersiveZoom = true
            m.setFocus(focusDetail)
            m.recomputeLayout()
        case nowImmersive:
            m.setFocus(focusDetail)
        case wasImmersive && m.immersiveZoom:
            m.zoom = false
            m.immersiveZoom = false
            m.setFocus(focusDetail)
            m.recomputeLayout()
        }
    }
    ```
    Because `nowImmersive` is false and `m.immersiveZoom` was destroyed by the Atlas trip, `case wasImmersive && m.immersiveZoom:` evaluates to false. Neither unzoom nor layout recomputation runs. `m.zoom` remains `true`. The two-pane split is lost and the user is permanently stranded in full-screen zoom on topology or summary views until they manually discover and press `z`.
- **Mutation & Probe Evidence:**
  Executed an isolated probe (`TestThreadSpatialAtlasResumeStrandsZoomProbe`) where a model in spatial view (`zoom=true`, `immersiveZoom=true`) transitions into Atlas and back. Upon pressing `Esc`, the view retreated to `threadDetailTopology` while `m.zoom` remained `true`, failing with:
  `STRANDED ZOOM: after retreating from spatial to topology, m.zoom is still true!`.
- **Minimal Recommendation:**
  In `internal/tui/atlas.go`:
  1. Add `immersiveZoom bool` to `atlasResume`:
     ```go
     type atlasResume struct {
         focus         focus
         zoom          bool
         immersiveZoom bool
         set           bool
     }
     ```
  2. In `enterAtlas`: capture `immersiveZoom: m.immersiveZoom`.
  3. In `exitAtlas`: restore `m.immersiveZoom = m.atlasResume.immersiveZoom`.
  4. Add a focused regression test in `internal/tui/thread_projection_test.go` verifying that entering and exiting Atlas while in spatial graph preserves `m.immersiveZoom` and cleanly restores the split layout on `Esc`.

**Resolution:** Accepted. Atlas round trips must preserve presentation-owned
immersive zoom so retreating from spatial restores the split without consuming
manual zoom.

#### M1. Spatial layout row placement, aliases, and inspector connections are nondeterministic under projection permutation · **Status:** tracked by 6g8by30btznq

- **Severity:** Medium
- **Component:** `internal/tui/thread_spatial.go`, `internal/tui/detail.go`
- **Effort:** 2–4 hours
- **Urgency:** Medium
- **Affected Acceptance Criteria:**
  - Task `6g6dw5js81f3`: "A short design note defines spatial layout, deterministic ordering, hjkl neighbor selection... Layout and aliases are deterministic for equivalent projection evidence. Node and edge order is deterministic: projection wave order first, then canonical task ID for presentation-only ties."
- **Exact Evidence:**
  - In `internal/tui/thread_spatial.go` (lines 53–62):
    ```go
    order := make(map[string]int, len(projection.Nodes))
    for index, node := range projection.Nodes {
        ...
        order[node.TaskID] = index
    }
    ```
    `order` assigns tie-breaking rank strictly based on the slice index in `projection.Nodes`.
  - In `rankThreadSpatialColumns` (lines 210–215, 264) and `placeThreadSpatialExternalGates` (lines 149–154, 160):
    Sorting rows in presentation columns uses `less := func(a, b string) bool { if order[a] != order[b] { return order[a] < order[b] }; return a < b }`.
    Because `order` depends on slice order, permuting `projection.Nodes` permutes the vertical row layout within columns.
  - In `internal/tui/detail.go` (lines 1257–1262):
    `threadGraphAliases` iterates over `projection.Nodes` in slice order to assign sequential gate aliases (`G1`, `G2`, etc.). Permuting `projection.Nodes` swaps gate alias labels.
  - In `internal/tui/thread_spatial.go` (lines 971–978):
    `threadSpatialConnections` iterates over `projection.Edges` in slice order to construct `needs` and `unlocks` strings without sorting. Permuting `projection.Edges` changes the displayed relationship order in the inspector.
- **Mutation & Probe Evidence:**
  Executed `TestThreadSpatialPermutedProjectionNondeterminismProbe` against `hostileThreadGraphProjection()`, reversing `projection.Nodes`. 12 nodes across columns 0, 2, and 3 moved row positions (e.g. `x9f4z76fx4q6` moved row 2 to 0, `640zdxa25e0v` moved row 0 to 8, `bgx4xt51cfe7` moved row 1 to 7). The existing test suite contains zero permutation coverage for `projection.Nodes`, `projection.Waves`, or `projection.Edges`.
- **Minimal Recommendation:**
  1. In `buildThreadSpatialLayout`: establish `order` deterministically independent of slice order: member wave index first (from `projection.Waves`), then canonical task ID for intra-wave ties and external/unranked nodes.
  2. In `threadGraphAliases`: sort external gates and unranked nodes by canonical task ID before assigning sequential alias numbers.
  3. In `threadSpatialConnections`: sort prerequisite and dependent task IDs / aliases before joining into `needs` and `unlocks`.
  4. Add a regression test in `internal/tui/thread_projection_test.go` asserting invariant layout coordinates, aliases, and rendered inspector text under permutation of `projection.Nodes`, `projection.Waves`, and `projection.Edges`.

**Resolution:** Accepted as defensive boundary hardening, not a currently
reachable producer defect. Core sorts projection evidence today; the task will
normalize or explicitly enforce and regress the ordering contract for future
portable producers.

#### L1. Cascaded external gates evaluate in arbitrary single-pass order during spatial placement · **Status:** tracked by 6g86e10dztpf

- **Severity:** Low
- **Component:** `internal/tui/thread_spatial.go`
- **Effort:** 2–3 hours
- **Urgency:** Low
- **Affected Acceptance Criteria:**
  - Task `6g6dw5js81f3` / Follow-up `6g86e10dztpf`: "External gates are positioned inside a valid dependency interval near the work they gate, without reversing supplied edges or misrepresenting member waves."
- **Exact Evidence:**
  - In `internal/tui/thread_spatial.go` (lines 122–148):
    `placeThreadSpatialExternalGates` processes external gates in a single forward pass:
    `for _, taskID := range threadSpatialOrderedIDs(order)`
  - When gate `G1` gates gate `G2`, which gates task `M` in column 4:
    If `G1` is evaluated before `G2`, `positions[G2]` is still at its unoptimized column (e.g., column 1). For `G1`, `earliestOutgoing` is 1, so `desired := earliestOutgoing - 1 = 0`. Because `desired <= current (0 <= 0)`, `G1` remains at column 0. When `G2` is subsequently evaluated, its outgoing dependent is `M` at column 4, so `G2` moves to column 3.
    Conversely, if `G2` is evaluated before `G1`, `G2` moves to column 3 first. When `G1` is evaluated, `positions[G2]` is 3, so `G1`'s `desired := 3 - 1 = 2`, and `G1` moves to column 2.
  - Thus, upstream gates in an external-gate chain are placed in different columns depending entirely on the iteration order of gate IDs.
  - Furthermore, if a gate moves out of a single-node column, that column is removed during post-loop compaction (lines 155–163), meaning column indices used in `positions` during the loop do not reflect post-compaction coordinates.
- **Mutation & Probe Evidence:**
  Traced through `placeThreadSpatialExternalGates` with chained external gate fixtures; verified that multi-gate placement differs based on evaluation order.
- **Minimal Recommendation:**
  This issue is closely aligned with follow-up task `6g86e10dztpf` ("Make dense Thread graph routes visually trustworthy"), which already owns multi-column gate intervals. Address either in `6g86e10dztpf` or via a reverse-topological gate evaluation pass (or fixed-point iteration until positions stabilize) in `placeThreadSpatialExternalGates`.

---

**Resolution:** Accepted. Dense-route hardening now explicitly includes
deterministic cascaded external-gate placement and owns the single-pass interval
issue.

### Follow-up scope cross-check

Maintainer triage accepted H1 into
[`6g8by30btznq` — close spatial Thread prototype correctness and invariant gaps](../tasks/6g8by30btznq-close-spatial-thread-prototype-correctness-and-invariant-gaps.md).
M1 is tracked there as defensive adapter-boundary normalization: the current core producer already
sorts nodes and edges, so the nondeterminism is not user-reachable today, but future portable
projection producers should not silently reshuffle a supposedly equivalent view. L1 remains owned
by [`6g86e10dztpf` — dense Thread graph route trustworthiness](../tasks/6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md),
whose external-gate placement criterion covers chained intervals.

To avoid duplicate findings for explicitly bounded future work, each observed prototype constraint was cross-referenced against the five tracked follow-up tasks:

| Observed Behavior / Prototype Limitation | Owning Follow-Up Task | Assessment |
| :--- | :--- | :--- |
| Routing ambiguity at dense crossings, shared lanes, same-row edges, or cyclic residue | [`6g86e10dztpf`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/planning/tasks/6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md) | **Bounded:** Honestly documented and actively tracked under follow-up task. Not duplicated as a review defect. |
| Absence of one-hop focal neighborhood filter / subgraph isolation | [`6g86c7y6hn41`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/planning/tasks/6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md) | **Bounded:** Bounded prototype explicitly omits local subgraph mode. |
| Lack of explicit Back key hint after drilling into tasks (relying on `ctrl+o`) | [`6g86016qvk5d`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/planning/tasks/6g86016qvk5d-add-a-reusable-back-action-for-tui-entity-navigation.md) | **Bounded:** Sequenced as a general TUI back action improvement across entities. |
| Non-yank actions (`E`, `e`, `m`, `f`) operating on parent Thread rather than highlighted node | [`6g86g03zfj2f`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/planning/tasks/6g86g03zfj2f-define-consistent-action-targets-for-structured-tui-detail-selections.md) | **Bounded:** Follow-up task explicitly owns parent vs. child action targeting matrix. |
| Alternate view discoverability (`v` key unobtrusiveness) and Thread slug truncation | [`6g87qn72901g`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.rW3Adc/planning/tasks/6g87qn72901g-make-alternate-tui-views-and-thread-identity-discoverable.md) | **Bounded:** Follow-up task owns view discoverability and list row card expansions. |
