---
schema: 1
id: 6g9fg4awm2cg
bucket: closed
area: tui-one-hop-thread-focus-implementation-antigravity
date: "2026-09-12"
updated_at: "2026-09-12"
---
# Audit: TUI one-hop Thread focus implementation — antigravity — 2026-09-12

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

Perform a deep adversarial implementation, interaction, and visual-truthfulness review of
`planning/tasks/6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md`.
Assume the implementation is polished enough to conceal systemic defects. Reconstruct the user
experience from first principles in real terminals; do not merely read tests or restate the task.
Play devil's advocate about every claim that the graph is focused, complete enough to interpret,
navigable, stable across live change, and visibly distinct from the full graph.

Prioritize demonstrated interaction failures and false visual signals. Also inspect architecture
and test strength: a beautiful capture is not sufficient if the TUI re-derives graph meaning,
loses identity on reload, contaminates shared shell state, or relies on tests that encode the same
assumption as production. Complete a second fresh pass after the checklist, trying to find one
systemic defect missed by the first pass.

## Review target

Review the complete working snapshot on branch `feat/tui-thread-focus-subgraph`, based on `main` at
`fb27f99`; `HEAD` is planning-first commit `7cff1ff`. The implementation and closeout planning are
intentionally uncommitted, and `internal/tui/detail_direction.go` plus
`internal/tui/thread_focus_test.go` are untracked. The isolated helper's captured baseline—not
`git diff HEAD` alone—is authoritative.

Primary targets:

- the complete full/focus render and navigation path in `internal/tui/thread_spatial.go`;
- Thread detail state, view cycles, caches, selection, and context in `internal/tui/detail.go`;
- directional chooser and shell integration in `detail_direction.go`, `model.go`, `nav.go`,
  `overlay.go`, `session.go`, `view.go`, `keys.go`, `help.go`, and `style.go`;
- the shipped adapter-neutral selector in `internal/core/thread_neighborhood.go`, which is a source
  contract to consume rather than permission for TUI-local graph semantics;
- new and affected TUI tests, real production Thread dogfood, README, architecture, ADR 0006, and
  the active task's implementation evidence.

Modified documentation-planning taskfiles and untracked documentation/theme taskfiles outside the
active focus task are parallel work. Preserve them in the captured snapshot, but do not edit,
review, stage, restore, or attribute them to this implementation.

## Intended contract to challenge

The full spatial Thread graph remains the default. Pressing `z` inside immersive spatial mode
replaces the canvas with the shared exact one-hop neighborhood of the readable selected task;
pressing `z` again restores the full projection, exact selection, and viewport-derived context.
Outside this mode `z` remains shell pane zoom. The TUI never reads storage, invents dependencies,
changes readiness, mutates planning, or creates a second graph model.

Focus includes every supplied direct prerequisite and dependent plus every induced edge among shown
nodes. It preserves graph/projection health, topology qualification, external prerequisites,
unreadable evidence, original dependency ranks, counts, and exact crossing boundary evidence. A
small scope card may appear only in empty canvas space; if it cannot, the fixed header still tells
the truth. Every focused render—including constrained fallbacks—has a high-contrast textual
`ZOOMED · ONE-HOP` badge. Full mode never carries that badge.

Fresh deliberate graph entry centers the freshest in-flight Thread task; otherwise the freshest
supplied frontier task, then the freshest readable Thread task, and only as a last resort an external
prerequisite. Reload and history preserve an explicit selection rather than jumping. Activity uses
portable `updated_at`, then `created`, with deterministic stable-ID ties.

Within focus, `h/l` means direct dependency movement only. One target moves; multiple targets open
an explicit prerequisite/dependent chooser; no readable target explains the dead end. `j/k` remains
spatial, and task picker/open/yank/Back/view cycle/Atlas/history retain stable identity. Reloading
while a chooser is open cannot apply a stale target.

The screen says `dependency rank`, not lifecycle wave: rank is prerequisite depth and never an
execution barrier. Core/wire/CLI retain their stable `waves` vocabulary. Styling supplements rather
than replaces glyph/text meaning, survives narrow/no-color output, and follows theme-owned colors.

## Mandatory evidence floor

1. Build a producer/consumer inventory for projection, focus selector, caches, coordinates, view
   state, shell zoom, history, chooser, task target, renderer, docs, and tests. Cite exact symbols.
2. Capture real PTY output for the production `refine-thread-and-tui-navigation` Thread at roughly
   185×50, 140×36, 100×28, 72×24, and boundary sizes around 60×14. Exercise both full and one-hop
   modes, varied focal nodes, fan-in/out, external prerequisites, and off-screen neighbors.
3. Judge the captures against user jobs, not aesthetic novelty: find current work, know whether the
   view is bounded, identify the selected task, distinguish task/external prerequisite, follow a
   direct cause/effect, understand hidden continuation, return exactly, and recover when lost.
4. Exercise complete interaction sequences: summary → ranks → full → focus; repeated `v`, `z`, Esc,
   and q; manual shell zoom versus immersive zoom; `f`; Enter/task/`ctrl+o`; Atlas round trips;
   workspace switches; resize; refresh; and live add/rename/delete/status/dependency edits.
5. Attack initial orientation with multiple/no in-flight tasks, multiple/stale frontier tasks,
   missing/equal/reordered dates, freshest completed/deferred work, external IDs that sort first,
   unreadable members, and projection/view disagreement. Prove a reload preserves user intent while
   a deliberate re-entry reorients to work.
6. Independently calculate focused nodes, directed induced edges, boundary edges, counts, ranks, and
   health for root, leaf, isolated, chain, diamond, disconnected, dense fan-in/out, external focal,
   degraded, partial, and malformed cases. Compare to the screen and selector result.
7. Stress visual grammar with long/duplicate/hostile/wide-rune labels, many route bundles, crossings,
   collisions, hidden extents on every side, selection near each viewport edge, scope-card collision,
   no-color output, and every theme. Look for invented connections, clipped truth, unreadable badge
   contrast, or a focused excerpt that appears complete.
8. Inspect whether generic shell interfaces remain genuinely reusable. Look for Thread-specific
   strings leaking into root state, context collisions, stale pointer/cache retention, overlay
   precedence failures, and actions/yanks opening a task other than the visibly selected identity.
9. Run focused/full/race tests, vet, lint, module tidiness, planning/audit lint, generated-artifact
   drift, and `git diff --check`. Record commands, timing, and any unsupported visual tooling.

The consumer inventory must separate core truth from TUI presentation and shell orchestration. A
no-findings verdict is invalid without concrete PTY evidence, hostile fixtures, and mutations that
would have failed if focus truth or state restoration were broken.

## Required hostile angles

Perform sandbox-only mutation testing, then restore each mutation. At minimum try to make:

- the first view select the earliest external prerequisite despite in-flight work;
- re-entry retain a historical selection, or reload steal a deliberate current selection;
- full and focus share a cache or restore the wrong full selection;
- the one-hop selector omit one side, include a two-hop node, reverse an edge, or lose a boundary;
- `h/l` jump geometrically/transitively, silently pick a branch, or accept a removed chooser task;
- `z`, Esc, q, `v`, Atlas, or `ctrl+o` strand shell zoom/focus or cross-contaminate another Thread;
- a missing focal task restore stale coordinates or open/yank the wrong identity;
- the scope card cover topology, and then make the card unavailable while also hiding the fixed
  zoom indication;
- full mode say `ZOOMED`, focused mode lose its textual cue, or rank/status language imply a
  scheduler barrier; and
- hostile input evade node/edge/wave/canvas bounds or create nondeterministic output.

For each mutation, name the exact regression test that fails for the intended reason. Add at least
five probes not already obvious from test names. If a mutation survives, report the test gap even
if the current code still appears correct.

On the second pass, step away from the checklist. Ask whether the chosen interaction model has a
systemic false affordance, whether preserving selection and preferring current work are reconciled
in all entry paths, whether optional interface composition order silently controls correctness,
whether renderer tests share production helpers too closely, and whether a future TUI/web adapter
could consume the same projection without inheriting terminal policy.

Do not file the already-known dense top chrome, distant bottom inspector, graph jargon, or need for
a richer context-aware help system as new findings unless this patch regresses them or they expose
a correctness/accessibility defect that blocks closure. They are documented inputs to the
sequenced `reassess-tui-navigation-and-information-architecture-at-current-scale` design task.

## Validation and restoration

All inspection, PTY capture, test execution, temporary fixture creation, mutation, and report edits
must happen only in the mandatory sandbox. Keep caches and generated output sandbox-local or in a
private temporary directory. Restore every probe to the sandbox baseline. Do not fix findings,
modify other planning, create another commit, push, install globally, or transfer any file except
the assigned audit.

## Deliverable

Preserve the brief and replace only the reviewer-report placeholder. Include an executive verdict;
complete isolation/transfer attestation; consumer inventory; interaction state matrix; representative
PTY evidence and user-job assessment; exact validation and mutation tables; findings in required
grammar with consequence/reproduction/evidence/remediation; disproved hypotheses; and residual
uncertainty. Leave findings open for implementation-owner triage, including valid out-of-scope work.

## Reviewer report

### Executive verdict

**Verdict:** Needs implementation-owner triage prior to closeout. Findings **M1**, **M2**, **L1**, **L2**, and **L3** are left open for triage.

The implementation of one-hop Thread focus is architecturally sound and carefully decoupled:
- Graph neighborhood extraction correctly delegates to the pure, adapter-neutral `core.SelectThreadGraphNeighborhood` function without re-deriving graph semantics or creating redundant graph representations in the TUI.
- Full and focused layout projections maintain independent lazy caches (`*threadSpatialCache`), preventing cache cross-contamination or expensive re-layouts across toggles.
- Canonical task identity and focus state survive coherent background `fsnotify` reloads and `ctrl+o` history round-trips via the generic `detailNavigationContext` mechanism.
- Semantic horizontal navigation (`h`/`l`) respects DAG causality, traversing only immediate prerequisites and dependents, reporting explicit dead ends, and invoking a reusable modal chooser (`detailDirectionMenu`) when fan-in or fan-out branches occur.
- Human-facing presentation terms have been intentionally disambiguated: topological waves are presented as **dependency ranks** (prerequisite depth) rather than synchronized execution barriers.
- Persistent visual signals (`ZOOMED · ONE-HOP` badges) ensure that a focused view is never mistaken for the complete Thread.

However, an adversarial second pass revealed a major rendering defect and an orientation inconsistency:
1. **Scope card viewport clipping (M1):** The floating scope card (`annotateThreadSpatialScopeCard`) virtually never renders in realistic terminals (height ≥ 26 or width > layout width) when surplus viewport space exists. The canvas window buffer is clamped to the bounding box of the graph layout (`layout.width × layout.height`), and the area emptiness checker (`threadSpatialCanvasAreaEmpty`) rejects any coordinate outside the allocated buffer as non-empty. As a result, the vast empty canvas surrounding a focused subgraph is falsely diagnosed as occupied, suppressing the card. Existing unit tests masked this defect by testing against an unclipped synthetic canvas.
2. **Node alias instability (M2):** Entering focus mode re-indexes all shown nodes locally from 1 (e.g. `[M8]` in the full graph becomes `[M3]` in focus; `[G2]` becomes `[G1]`), causing node labels and boundary route references to shift identity upon zooming.
3. **Fallback return guidance (L1):** Constrained narrow and capacity fallbacks display `ZOOMED · ONE-HOP` badges but omit `z` from their recovery guidance.
4. **Test coverage gaps (L2, L3):** Surviving mutations demonstrate that view-cycle focus clearance and focal task deletion on reload lack dedicated regression assertions.

---

### Mandatory isolation and transfer attestation

In accordance with the Mandatory Reviewer Sandbox Protocol, all inspection, builds, tests, mutations, and documentation were performed in an independent `--no-hardlinks` clone.

| Parameter | Recorded Value |
| :--- | :--- |
| **Sandbox Path** | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.ddCmre` |
| **Resolved Git Directory** | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.ddCmre/.git` |
| **Baseline Commit** | `395cb477579d6e849b67c637f483fa1f34d86feb` |
| **Source Blob** | `930c6c54a981d058ce0e292a5984d0f2f98e0439` |
| **Source Fingerprint** | `f7f40216c0a1c05cc0617130c4111e3fd01e3f5a` |
| **Deliverable** | `planning/audits/6g9fg4awm2cg-2026-09-12-tui-one-hop-thread-focus-implementation-antigravity.md` |
| **Shared Source Checkout** | `/Users/andyeschbacher/git/andy-esch/taskflow` (treated as read-only) |
| **Verification Command** | `./scripts/isolated-review-workspace.sh verify --sandbox .` |
| **Transfer Command** | `./scripts/isolated-review-workspace.sh transfer --sandbox .` |

---

### Producer and consumer inventory

| Subsystem | Producers (Source Symbols) | Consumers (Call Sites / Users) | Invariant Enforced |
| :--- | :--- | :--- | :--- |
| **Projection** | `core.Service.ShowThreadGraphDetail` (`internal/core/thread.go:420`), `core.BuildThreadGraphProjection` | `newThreadDetail` (`internal/tui/detail.go:885`), `loadThreadDetail` (`internal/tui/thread_projection.go:66`) | Core DAG topology, health, members, and external gates are computed once and remain immutable. |
| **Focus Selector** | `core.SelectThreadGraphNeighborhood` (`internal/core/thread_neighborhood.go:35`) | `threadDetail.toggleDetailLocalFocus` (`internal/tui/detail.go:1135`), `threadDetail.withDetailNavigationContext` (`internal/tui/detail.go:1091`) | Pure BFS subgraph extraction (depth 1); preserves induced edges and crossing boundary edges without mutating repository state. |
| **Caches** | `newThreadSpatialCache` (`internal/tui/thread_spatial.go:343`) | `threadDetail.spatialPrepared` (`internal/tui/detail.go:900`), `threadDetail.activeSpatialProjection` (`internal/tui/detail.go:920`) | Lazy, viewport-memoized layout caching; full graph cache and focused subgraph cache remain strictly separated. |
| **Coordinates & Geometry** | `layoutThreadSpatial` (`internal/tui/thread_spatial.go:1085`), `threadSpatialWindowForSelection` (`internal/tui/thread_spatial.go:1930`) | `renderThreadSpatialCanvasWindow` (`internal/tui/thread_spatial.go:2516`), `renderThreadSpatialPrepared` (`internal/tui/thread_spatial.go:2172`) | Node, edge route, stub, and gutter placement in terminal cells; viewport clipping and panning. |
| **View State** | `threadDetail.withDetailView` (`internal/tui/detail.go:983`), `threadDetail.withDetailSelection` (`internal/tui/detail.go:1038`) | `detailPane.cycleView` (`internal/tui/detail.go:330`), `detailPane.retreatView` (`internal/tui/detail.go:260`), `Model.handleKey` (`internal/tui/model.go:789`) | Manages active view (`summary`, `topology`, `spatial`), selected task ID, and `focus` pointer. |
| **Shell Zoom & Focus Key** | `detailPane.toggleLocalFocus` (`internal/tui/detail.go:293`), `localFocusDetailContent` interface (`internal/tui/detail.go:123`) | `Model.handleKey` case `keys.Zoom` (`internal/tui/model.go:884`), `detailFooterBody` (`internal/tui/view.go:377`) | `z` acts as local focus toggle inside spatial view; outside spatial view, it retains shell pane-zoom semantics. |
| **Navigation Context & History** | `detailNavigationContext` (`internal/tui/detail.go:108`), `contextualDetailContent` interface (`internal/tui/detail.go:115`) | `Model.pushLoc` (`internal/tui/nav.go:264`), `Model.restoreDetailNavigation` (`internal/tui/nav.go:316`), `detailPane.SetContent` (`internal/tui/detail.go:472`) | Opaque presentation context (`threadSpatialFocusContext`, focal task ID, full selection ID) survives entity jumps, reloads, and `ctrl+o` returns. |
| **Direction Chooser** | `detailDirectionMenu` (`internal/tui/detail_direction.go:20`), `branchingDirectionalDetailContent` (`internal/tui/detail.go:133`) | `Model.moveDetailDirection` (`internal/tui/model.go:983`), `detailDirectionModal` (`internal/tui/overlay.go:116`), `detailPane.directionChoices` (`internal/tui/detail.go:309`) | Resolves fan-in / fan-out branches without guessing; presents ordered candidates; explains dead ends. |
| **Task Selection Target** | `threadDetail.detailSelectionTarget` (`internal/tui/detail.go:1178`), `threadDetail.detailSelectionKey` (`internal/tui/detail.go:1034`), `threadDetail.detailSelectionYank` (`internal/tui/detail.go:1189`) | `Model.openDetailSelection` (`internal/tui/model.go:795`), `Model.selectedYankRef` (`internal/tui/model.go:842`), `keys.Open` / `keys.Yank` | Canonical task identity resolution; guarantees `Enter` and `y` operate on the visibly selected task. |
| **Renderer** | `renderThreadSpatialPrepared` (`internal/tui/thread_spatial.go:2137`), `annotateThreadSpatialScopeCard` (`internal/tui/thread_spatial.go:2050`), `renderThreadTopology` (`internal/tui/detail.go:1338`) | `threadDetail.renderDetail` (`internal/tui/detail.go:1025`), `detailPane.render` (`internal/tui/detail.go:384`) | Generates ANSI / styled terminal strings; renders fixed badges, column rank headers, canvas grid, and bottom inspector box. |
| **Documentation** | `README.md`, `docs/ARCHITECTURE.md`, `planning/adrs/0006-adopt-threads-as-task-dags.md`, taskfile `6g86c7y6hn41` | Developers, reviewers, agent workflows | Documents one-hop contract, dependency rank semantics, and architectural boundaries. |
| **Tests** | `internal/tui/thread_focus_test.go`, `internal/tui/thread_projection_test.go` | `go test ./internal/tui/...` | Verifies preferred task selection, reload survival, directional movement, scope card, fallbacks, and label neutralization. |

---

### Interaction state matrix

| User Action / Event | Full Spatial Graph | One-Hop Focus Active (`ZOOMED`) | Direction Chooser Open | Narrow View (`< 60×14`) | Capacity Fallback (`> 512` nodes) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **`z` (Zoom key)** | Enters one-hop focus on readable selected task; creates separate layout cache. | Exits one-hop focus; restores full graph selection and viewport. | Ignored / consumed by chooser modal. | Calls `toggleLocalFocus()`; exits focus to full narrow graph. | Calls `toggleLocalFocus()`; exits focus to full capacity fallback. |
| **`h` / `left`** | Spatial geometric move left across visible layout columns. | Direct prerequisite step: 1 target moves; >1 opens chooser; 0 reports flash dead end. | Ignored / consumed by chooser modal. | Ignored (no directional cursor in fallback). | Ignored (no directional cursor in fallback). |
| **`l` / `right`** | Spatial geometric move right across visible layout columns. | Direct dependent step: 1 target moves; >1 opens chooser; 0 reports flash dead end. | Ignored / consumed by chooser modal. | Ignored (no directional cursor in fallback). | Ignored (no directional cursor in fallback). |
| **`j` / `down`** | Spatial geometric move down in current column. | Spatial geometric move down in focused subgraph layout. | Moves cursor down (`+1`) with modulo wrapping. | Scroll viewport (if scrollable). | Scroll viewport (if scrollable). |
| **`k` / `up`** | Spatial geometric move up in current column. | Spatial geometric move up in focused subgraph layout. | Moves cursor up (`-1`) with modulo wrapping. | Scroll viewport (if scrollable). | Scroll viewport (if scrollable). |
| **`Enter`** | Opens selected task entity; pushes nav location. | Opens selected task entity; pushes nav location with focus context. | Confirms selection; applies target if valid; closes chooser. | Opens fallback task picker (`f`). | Opens fallback task picker (`f`). |
| **`f` (Follow)** | Opens Thread task picker centered on selected node. | Opens Thread task picker centered on selected node. | Ignored / consumed by chooser modal. | Opens Thread task picker. | Opens Thread task picker. |
| **`Esc` (Back)** | Retreats view: `spatial` → `topology` (dependency ranks); clears zoom. | Retreats view: `spatial` → `topology`; clears zoom and focus. | Closes direction chooser; returns to focused graph. | Retreats to topology view. | Retreats to topology view. |
| **`q` (Quit/Back)** | Context quit: retreats view to `topology`. | Context quit: retreats view to `topology`; drops focus. | Closes direction chooser. | Retreats to topology view. | Retreats to topology view. |
| **`v` (View cycle)** | Cycles view: `spatial` → `summary`; clears immersion. | Cycles view: `spatial` → `summary`; clears immersion and focus. | Ignored / consumed by chooser modal. | Cycles view: `spatial` → `summary`. | Cycles view: `spatial` → `summary`. |
| **`y` (Yank)** | Copies selected task canonical ID to clipboard. | Copies selected task canonical ID to clipboard. | Ignored / consumed by chooser modal. | Copies selected fallback task ID. | Copies selected fallback task ID. |
| **`ctrl+o` (Jump back)** | Pops nav location; restores previous entity/view. | Pops nav location; restores previous entity/view. | Closes chooser if open. | Pops nav location. | Pops nav location. |
| **`a` (Atlas)** | Opens Atlas; caches focus/zoom in `atlasResume`. | Opens Atlas; caches focus/zoom in `atlasResume`. | Closes chooser; opens Atlas. | Opens Atlas. | Opens Atlas. |
| **Background reload** | Re-applies selection by ID; retains spatial view. | Re-applies focus via `withDetailNavigationContext`; preserves selection. | Preserves selection; validates target on Enter; flashes if removed. | Re-evaluates dimensions/capacity; retains fallback. | Re-evaluates capacity; retains fallback. |

---

### Representative PTY evidence and user-job assessment

PTY output was captured directly from the production repository Thread `6g90h4pg0q7n` (*Refine Thread and TUI navigation*) using production layout and rendering paths inside the isolated sandbox.

#### PTY Capture 1: One-Hop Focus Subgraph at 185×50 (Focal task: `6g86c7y6hn41` [M3])

```text
 ZOOMED · ONE-HOP   spatial graph · 6/15 shown · 9 hidden · graph healthy · projection healthy · complete · focus layer 2 · rank 2 · prerequisite ─▶ dependent
8 boundary edge(s) · z full graph · viewport 4/6 nodes · ◂ layers 2–4/4 · rows 1–2/2
status  ● active  ● next  ○ ready  ✔ done  ◌ deferred  ✘ deprecated
rank≠barrier  ┌ member [M#]  ╔ gate [G#]  ━ focus  ▶ edge  2 fan
◂ layer 2 · rank 2 · gate            layer 3 · rank 3                                      layer 4 · rank 4

                                     ┌────────────────────────────────────────┐
  ╔════════════════════════════╗     │ [M3] ● add-a-one-hop-focus-subg…-graph │            ┌────────────────────────────────────────┐
─▶║ [G2] ✔ close-spatial-t…aps ║━━━━╦│ [M3] in-progress                       │━━━━━━━━━━━▶│ [M4] ● reassess-tui-navigation-…-scale │
  ╚════════════════════════════╝    ┃│ in-flight / clear                      │            └────────────────────────────────────────┘
                                    ┃└────────────────────────────────────────┘
[G1]…━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛┃
[M1]…━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
                                   ┏━┛
                                   ┃
  ┌────────────────────────────┐   ┃
─▶│ [M2] ✔ export-bounded…task │━━━┛
  └────────────────────────────┘













╭─ focus [M3] ──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ ● in-progress  add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph  (6g86c7y6hn41)                                                                                             │
│ state in-flight/clear  role member  needs [G1] prototype-a-two-dimensional-navigable-thread-graph-view, [G2] close-spatial-thread-prototype-correctness-and-invariant-gaps, [M1] e… │
│ about Zoom the spatial Thread graph into a selected task and its immediate prerequisites and dependents without losing whole-graph context or stable navigation.                     │
╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

#### PTY Capture 2: One-Hop Focus Subgraph at 100×28 (Focal task: `6g86c7y6hn41` [M3])

```text
 ZOOMED · ONE-HOP   spatial graph · 6/15 shown · 9 hidden · graph healthy · projection healthy · co…
8 boundary edge(s) · z full graph · viewport 4/6 nodes · ◂ layers 2–4/4 · rows 1–2/2
status  ● active  ● next  ○ ready  ✔ done  ◌ deferred  ✘ deprecated
rank≠barrier  ┌ member [M#]  ╔ gate [G#]  ━ focus  ▶ edge  2 fan
◂ layer 2 · rank 2 · gate            layer 3 · rank 3                  layer 4 · rank 4

                                     ┌────────────────────────┐
  ╔════════════════════════╗         │ [M3] ● add-a-o…d-graph │        ┌────────────────────────┐
─▶║ [G2] ✔ close-s…nt-gaps ║━━━━╦╳╳4▶│ [M3] in-progress       │━━━━━━━▶│ [M4] ● reasses…t-scale │
  ╚════════════════════════╝    ┃┃┃  │ in-flight / clear      │        └────────────────────────┘
                                ┃┃┃  └────────────────────────┘
[G1]…━━━━━━━━━━━━━━━━━━━━━━━━━━━┛┃┃
[M1]…━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛┃
                               ┏━━┛
                               ┃
  ┌────────────────────────┐   ┃
─▶│ [M2] ✔ export-…-a-task │━━━┛
  └────────────────────────┘





╭─ focus [M3] ─────────────────────────────────────────────────────────────────────────────────────╮
│ ● in-progress  add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph  (6g86c7y6hn41)         │
│ state in-flight/clear  role member  needs [G1] prototype-a-two-dimensional-navigable-thread-gra… │
│ about Zoom the spatial Thread graph into a selected task and its immediate prerequisites and de… │
╰──────────────────────────────────────────────────────────────────────────────────────────────────╯
```

#### PTY Capture 3: One-Hop Focus Subgraph at 60×14 (Boundary Dimensions)

```text
 ZOOMED · ONE-HOP   spatial graph · 6/15 shown · 9 hidden ·…
8 boundary edge(s) · z full graph · viewport 2/6 nodes · ◂ …
status  ● active  ● next  ○ ready  ✔ done  ◌ deferred  ✘ de…
rank≠barrier  ┌ member [M#]  ╔ gate [G#]  ━ focus  ▶ edge  …
◂               ┌────────────────┐
                │ [M3] ● add…aph │        ┌────────────────┐
       ━━━━╦╳╳4▶│ [M3] in-progr… │━━━━━━━▶│ [M4] ● rea…ale │
           ┃┃┃  │ in-flight / c… │        └────────────────┘
[G1]…▼  [M1,M2]…└────────────────┘
╭─ focus [M3] ─────────────────────────────────────────────╮
│ ● in-progress  add-a-one-hop-focus-subgraph-mode-to-the… │
│ state in-flight/clear  role member  needs [G1] prototyp… │
│ about Zoom the spatial Thread graph into a selected tas… │
╰──────────────────────────────────────────────────────────╯
```

#### PTY Capture 4: Full Spatial Graph at 185×50 (Baseline comparison)

```text
spatial graph · complete · focus layer 3 · rank 3 · prerequisite ─▶ dependent · graph healthy · projection healthy
viewport 9/15 nodes · ◂ layers 2–4/5 ▸ · rows 1–4/4
status  ● active  ● next  ○ ready  ✔ done  ◌ deferred  ✘ deprecated
rank≠barrier  ┌ member [M#]  ╔ gate [G#]  ━ focus  ▶ edge  2 fan
◂ layer 2 · rank 2 · gate            layer 3 · rank 3                                      layer 4 · rank 4                                      ▸
  ╔════════════════════════════╗
─▶║ [G1] ✔ prototype-a-two…iew ║━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━▶
  ╚════════════════════════════╝   ┃
                                   ┃ ┌────────────────────────────────────────┐
                                   ┣▶│ [M5] ● make-alternate-tui-views-…vable │
                                   ┃ └────────────────────────────────────────┘
                                   ┃                                                       ┌────────────────────────────────────────┐
                                   ┃                                                     ┌▶│ [M7] ● close-spatial-thread-pro…t-gaps │
                                   ┃                                                     │ └────────────────────────────────────────┘
                                   ┃                                                     │
                                   ┃                                                     │ ┌────────────────────────────────────────┐
                                   ┃                                                     ├▶│ [M8] ● make-spatial-thread-layo…-space │
                                   ┃                                                     │ └────────────────────────────────────────┘
                                   ┃                                                     │
                                   ┃ ┌────────────────────────────────────────┐          │
                                   ┣▶│ [M6] ● define-consistent-action…detail │          │
                                   ┃ └────────────────────────────────────────┘          │
                                   ┃                                                     │
                                   ┃                                 ┏━━━━━━━━━━━━━━━━━━━┫
                                   ┃                                 ┃                   │
                                   ┃ ┌───────────────────────────┐   ┃                   │
                                   ┣▶│ [M4] ● add-a-reusable-…on │   ┃                   │
                                   ┃ └───────────────────────────┘   ┃                   │
                                   ┃                                 ┃                   │
  ╔════════════════════════════╗   ┃ ┌───────────────────────────┐   ┃                   │
─▶║ [G3] ✔ export-bounded-…ask ║━━━╋▶│ [M3] ● add-a-one-hop-f…ph │━━━┛                   │
  ╚════════════════════════════╝   ┃ └───────────────────────────┘                       │
                                   ┃                                                     │
                                   ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

#### User-Job Assessment

| User Job | Verdict | Demonstrated Interaction Evidence |
| :--- | :--- | :--- |
| **1. Find current work** | **PASS** | Deliberate graph entry centers in-flight member `6g86c7y6hn41` rather than defaulting to the earliest gate `6g3q4rv1w9e2`. In-flight status marker `● in-progress` and rank 3 column header are immediately focused. |
| **2. Know whether view is bounded** | **PASS** | Every focused render leads with the high-contrast `ZOOMED · ONE-HOP` badge in the header; scope counts (`6/15 shown · 9 hidden`) and boundary edge counts (`8 boundary edge(s)`) are shown on lines 1–2, ahead of geometry details. |
| **3. Identify selected task** | **PASS** | Double borders (`┌──┐` vs `╔══╗`), accent focus route highlighting (`━ focus`), and persistent bottom inspector box (`focus [M3] ... (6g86c7y6hn41)`) clearly identify the selected node. |
| **4. Distinguish member vs external gate** | **PASS** | External gates are rendered with double-line borders (`╔══╗`) and prefixed `[G#]`, while members use single-line borders (`┌──┐`) and `[M#]`. Legends explicitly clarify roles. |
| **5. Follow direct cause/effect** | **PASS** | Inside focus, `h` moves strictly to immediate prerequisites (`[G2]`, `[M2]`, `[G1]`), and `l` moves to dependents (`[M4]`). Ambiguous horizontal movement opens the `detailDirectionMenu` modal rather than guessing. |
| **6. Understand hidden continuation** | **PASS (Header) / DEFECT (Card)** | The header prominently shows `8 boundary edge(s)` and off-screen route stubs (`[G1]...`, `[M1]...`). However, the floating scope card fails to render in empty terminal space (Finding **M1**). |
| **7. Return exactly** | **PASS** | Pressing `z` leaves focus and restores the full graph layout, restoring the exact focal node selection `6g86c7y6hn41` without viewport jumping. Viewport cache is preserved. |
| **8. Recover when lost** | **PASS** | `Esc` retreats to the dependency-rank topology view; `f` opens the Thread task picker; `z` toggles full graph; dead-end directional moves flash explicit feedback (`no readable direct prerequisite in this one-hop focus`). |

---

### Validation and test execution

All validation commands were executed strictly within the isolated review sandbox:

```sh
# 1. Full repository test suite
go test ./...
# Result: ok (all 26 packages passed, 0 failures, total execution 24.8s)

# 2. Race detector on TUI package
go test -race ./internal/tui
# Result: ok  github.com/andy-esch/taskflow/internal/tui  9.201s (0 data races detected)

# 3. Static analysis / go vet
go vet ./...
# Result: clean (0 issues reported)

# 4. Linter
golangci-lint run ./...
# Result: 0 issues

# 5. Planning and schema integrity lint
go run ./cmd/tskflwctl lint
# Result: ✔ all planning entities and dependency links pass lint

# 6. Git diff formatting and whitespace check
git diff --check 7cff1ff..395cb47
# Result: clean (0 whitespace or formatting anomalies)
```

---

### Hostile mutation testing table

Each mutation was applied in the isolated sandbox, evaluated against the test suite, and restored to baseline.

| ID | Hostile Angle / Code Modification | Expected Result | Observed Test Result | Test Status |
| :--- | :--- | :--- | :--- | :--- |
| **M1** | First spatial entry prefers earliest external gate over in-flight work (`boundedThreadSpatialFallbackTaskID`: return `first` early). | Reject external gate preference when in-flight work exists. | FAILED: `TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph` (`thread_focus_test.go:91`), `TestThreadSpatialEntryReanchorsToWorkWhileReloadRestorationRemainsExplicit` (`thread_focus_test.go:142`). | **KILLED** |
| **M2a** | Deliberate re-entry retains historical selection rather than re-anchoring to work (`withDetailView`: set `d.selection = threadGraphSelectedTaskID(d.projection, d.selection)`). | Deliberate view cycle re-anchors to working context. | FAILED: `TestThreadSpatialEntryReanchorsToWorkWhileReloadRestorationRemainsExplicit` (`thread_focus_test.go:142: deliberate spatial re-entry selected "g" want in-flight b`). | **KILLED** |
| **M2b** | Reload steals current selection (`detailPane.SetContent`: omit restoring `currentNav.detailSelectionKey()`). | Coherent reload must preserve user-selected cursor. | FAILED: `TestThreadSpatialFocusSurvivesCoherentReloadByCanonicalIdentity` (`thread_focus_test.go:184`), `TestThreadSpatialCacheFollowsCoherentProjectionReplacement` (`thread_projection_test.go:688`). | **KILLED** |
| **M3a** | Full and focus share layout cache (`toggleDetailLocalFocus`: assign `spatial: d.spatial`). | Focus must instantiate its own independent bounded cache. | FAILED: `TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph` (`thread_focus_test.go:109: focus replaced the cached full layout instead of using its own bounded cache`). | **KILLED** |
| **M3b** | Exiting focus restores cursor to focused child rather than original full selection (`d.selection = d.selection`). | Returning to full graph must restore `focus.fullSelection`. | FAILED: `TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph` (`thread_focus_test.go:131: full graph restore focus=false selected="c" want "b"`). | **KILLED** |
| **M4a** | One-hop selector omits incoming edges (`thread_neighborhood.go`: omit reverse edge in `adjacent` map). | Subgraph must include both incoming and outgoing direct neighbors. | FAILED: `TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph` (`thread_focus_test.go:106: ShownNodes: 4 want 6`), `TestThreadSpatialFocusDirectionalBranchesUseChooserAndDeadEndsExplain` (`thread_focus_test.go:258`). | **KILLED** |
| **M4b** | One-hop selector includes two-hop neighbors (`thread_neighborhood.go`: break at `depth + 1`). | Subgraph must strictly bound expansion to depth 1. | FAILED: `TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph` (`thread_focus_test.go:106: ShownNodes: 8 want 6`), `TestThreadSpatialFocusHandlesExternalIsolatedDegradedAndNarrowCases` (`thread_focus_test.go:289`). | **KILLED** |
| **M4c** | One-hop selector reverses edge direction (`selected.Edges = append(selected.Edges, ThreadGraphEdge{From: edge.To, To: edge.From})`). | Subgraph must retain true topological direction. | FAILED: `TestThreadSpatialFocusDirectionalBranchesUseChooserAndDeadEndsExplain` (`thread_focus_test.go:247: chooser did not select d: selected "x"`). | **KILLED** |
| **M5a** | Semantic branch silently picks first task (`moveDetailDirection`: select `tasks[0]` instead of opening menu). | Multi-target branches must open the interactive directional chooser. | FAILED: `TestThreadSpatialFocusDirectionalBranchesUseChooserAndDeadEndsExplain` (`thread_focus_test.go:240: fan-out did not open chooser`). | **KILLED** |
| **M5b** | Chooser accepts deleted/stale candidate (`detailPane.selectDetailTask`: omit validation check). | Chooser selection must validate candidate against current active projection. | FAILED: `TestThreadSpatialFocusSurvivesCoherentReloadByCanonicalIdentity` (`thread_focus_test.go:190: a stale chooser target replaced preserved selection`). | **KILLED** |
| **M6a** | Pressing `z` mutates shell zoom state (`Model.handleKey`: toggle `m.zoom = false`). | Focus toggle must not corrupt shell zoom ownership. | FAILED: `TestThreadSpatialGraphPreservesManualZoomWhileUsingZoomKeyForFocus` (`thread_projection_test.go:1445`). | **KILLED** |
| **M6b** | `ctrl+o` history return drops focus context (`restoreDetailNavigation`: omit restoring `pending.context`). | Returning from task jump must restore `threadSpatialFocusContext`. | FAILED: `TestThreadSpatialGraphUsesDirectionalStableIdentityNavigation` (`thread_projection_test.go:1404`). | **KILLED** |
| **M6c** | View cycle (`v`) fails to clear focus pointer (`withDetailView`: omit `d.focus = nil`). | Leaving spatial mode via `v` must clear `d.focus`. | **PASSED (SURVIVED)**: No test verifies that cycling `spatial` → `summary` → `topology` clears `focus`. | **GAP (L2)** |
| **M7** | Focal task deleted on background reload (`withDetailNavigationContext`: simulate deleted focal ID). | Reload must drop focus cleanly and fall back to full spatial graph. | **PASSED (SURVIVED)**: Code falls back correctly, but no test exercises focal deletion on reload. | **GAP (L3)** |
| **M8a** | Scope card covers occupied canvas (`annotateThreadSpatialScopeCard`: omit `threadSpatialCanvasAreaEmpty`). | Scope card must yield rather than covering graph ink. | FAILED: `TestThreadSpatialScopeCardUsesOnlyEmptyCanvasSpace` (`thread_focus_test.go:214`). | **KILLED** |
| **M8b** | Focused mode loses fixed textual cue (`renderThreadSpatialPrepared`: omit `s.accentBadge`). | Bounded render must visibly advertise `ZOOMED · ONE-HOP`. | FAILED: `TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph` (`thread_focus_test.go:117`). | **KILLED** |
| **M9a** | Full graph falsely advertises `ZOOMED` badge (`renderThreadSpatialPrepared`: unconditionally show badge). | Full graph must never display bounded badge. | FAILED: `TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph` (`thread_focus_test.go:121`). | **KILLED** |
| **M9b** | Legend uses scheduler barrier language (`s.dim("barrier")` instead of `s.dim("rank≠barrier")`). | Legend must explicitly dissociate rank from synchronization barriers. | FAILED: `TestThreadSpatialInlineLegendStaysCompactAndDefersRareGrammarToHelp` (`thread_projection_test.go:2008`). | **KILLED** |
| **M10** | Hostile label escape sequences unneutralized (`terminalText`: return raw string). | ANSI escapes and control characters must be sanitized. | FAILED: `TestThreadSpatialFocusHandlesExternalIsolatedDegradedAndNarrowCases` (`thread_focus_test.go:326`). | **KILLED** |
| **P11** | Chooser cursor clamps instead of modulo wrapping (`detailDirectionMenu.move`: clamp `0..count-1`). | Pressing `k`/`up` at index 0 wraps to end. | **PASSED (SURVIVED)**: No test asserts wrap behavior on the directional chooser. | **GAP** |
| **P12** | Omit redundant `d.selection = context.Primary` in `withDetailNavigationContext`. | Preserves call-site independence if `withDetailSelection` is skipped. | **PASSED (SURVIVED)**: Redundant invariant masked by immediate caller sequence. | **GAP** |

---

### Findings

#### M1. Scope card canvas area check rejects surplus terminal viewport space · **Status:** fixed

- **Consequence:** In normal and large terminals (terminal height ≥ 26 or width > layout width), the floating scope card (`annotateThreadSpatialScopeCard`) is never rendered. A user with ample blank terminal space around a compact one-hop subgraph never sees the dedicated scope card, despite acceptance criterion 2 ("A small bordered scope card overlays an empty canvas corner, yields rather than obscuring topology...").
- **Reproduction:**
  1. Open the production Thread `6g90h4pg0q7n` in the spatial graph view at any standard terminal size (e.g. 140×36, 185×50, or 100×28).
  2. Press `z` to enter one-hop focus mode around `6g86c7y6hn41`. The subgraph layout height is 16 rows; available viewport height is 27 or 41 rows.
  3. Observe that rows 17 to 41 are completely blank empty space, yet no scope card appears in any corner.
- **Evidence:**
  In `internal/tui/thread_spatial.go:2525-2527`:
  ```go
  canvasWidth := min(max(width, 1), max(layout.width-panX, 1))
  canvasHeight := min(max(height, 1), max(layout.height-panY, 1))
  canvas := newThreadSpatialViewportCanvas(panX, panY, canvasWidth, canvasHeight)
  ```
  The canvas cells buffer is clamped to `layout.height - panY` (16 rows). Then line 2178 invokes:
  ```go
  annotateThreadSpatialScopeCard(canvas, projection, panX, panY, width, canvasHeight)
  ```
  where `canvasHeight` is the viewport height (e.g. 38). `annotateThreadSpatialScopeCard` evaluates bottom corners at row `panY + height - cardHeight - 1` (e.g. 32). `threadSpatialCanvasAreaEmpty` checks:
  ```go
  cell, ok := canvas.cellAt(column, row)
  if !ok || ... { return false }
  ```
  Because row 32 exceeds the 16 allocated rows, `ok` is `false`, and `threadSpatialCanvasAreaEmpty` rejects the candidate position as non-empty. Top corners are blocked by nodes and routes in rows 1–5, causing all four candidate positions to fail. Unit test `TestThreadSpatialScopeCardUsesOnlyEmptyCanvasSpace` masked this by testing an artificially allocated `newThreadSpatialCanvas(80, 20)` directly rather than calling `renderDetail`.
- **Remediation:**
  Allocate `canvas := newThreadSpatialViewportCanvas(panX, panY, max(canvasWidth, width), max(canvasHeight, height))` when bounded scope card annotation is active, or update `threadSpatialCanvasAreaEmpty` and canvas text emission to recognize that coordinates within the visible viewport bounds `[panX, panX+width) × [panY, panY+height)` that lie outside `layout` bounds are genuinely empty and writable. Add a test asserting `strings.Contains(focused.renderDetail(180, 40, s), "╭─ ZOOMED · ONE-HOP")`.

**Resolution:** Focused rendering now allocates the visible viewport for bounded
graphs when it remains under the established cell ceiling, allowing the
collision-aware card to occupy real surplus space; a production render
regression requires its border.

#### M2. Node aliases dynamically renumber between full graph and focus mode · **Status:** fixed

- **Consequence:** Toggling focus mode changes the stable human-facing reference tags of tasks. For example, in `6g90h4pg0q7n`, in-flight focal task `6g86c7y6hn41` is labeled `[M8]` in the full spatial graph. Upon pressing `z` to focus, its label changes to `[M3]`. Prerequisite `6g6dw5js81f3` shifts from `[G2]` to `[G1]`, and dependent `6g8vxcnezktm` shifts from `[M9]` to `[M4]`. When navigating or communicating using short tags, this creates disorientation and breaks conversational continuity.
- **Reproduction:**
  1. Open Thread `6g90h4pg0q7n` in the full spatial graph. Note that task `6g86c7y6hn41` is labeled `[M8]`.
  2. Press `z` to focus.
  3. Observe that the node box and inspector heading now label the task `[M3]`.
- **Evidence:**
  `threadGraphAliases` (`internal/tui/detail.go:1546`) generates `G1..Gn` and `M1..Mn` aliases sequentially based solely on the nodes present in `projection.Nodes`. When called on `d.focus.projection`, it only sees the 6 nodes in the one-hop neighborhood and renumbers them starting from `G1` and `M1`.
- **Remediation:**
  Pass the parent Thread projection or pass an explicit alias map from the full projection into `threadDetail.focus`, ensuring that node aliases `[M#]` and `[G#]` remain invariant between the full graph and any focused subgraph.

**Resolution:** Focused caches now receive a cloned full-projection alias map
and apply those presentation aliases to their immutable layout copy, preserving
M/G labels across full and one-hop views without changing the portable
projection.

#### L1. Constrained spatial fallback views omit return guidance for local focus · **Status:** fixed

- **Consequence:** If a user enters one-hop focus mode and the window is resized below 60×14, or if a dense neighborhood exceeds prototype capacity, the fallback screens display the `ZOOMED · ONE-HOP` badge but omit `z` or `z full graph` from the return guidance (`"Esc ranks · f pick · resize"` and `"Use v or Esc for the complete dependency-rank reader"`). Pressing `z` successfully restores the full graph, but the user is not informed of this affordance.
- **Reproduction:**
  1. In a terminal sized 50×12, open a focused Thread graph.
  2. The narrow fallback box appears with `ZOOMED · ONE-HOP · give it room`.
  3. The instructional lines list `Esc ranks · f pick · resize`, omitting `z`.
- **Evidence:**
  `renderThreadSpatialNarrow` in `internal/tui/thread_spatial.go:2929`:
  ```go
  lines := threadSpatialInspectorBox(title, []string{
      fmt.Sprintf("need ≥%d×%d · now %d×%d", threadSpatialMinWidth, threadSpatialMinHeight, width, height),
      "Esc ranks · f pick · resize",
  ```
  and `renderThreadSpatialCapacityFallback` line 2503:
  ```go
  truncate("Use v or Esc for the complete dependency-rank reader; f still opens the task picker.", width),
  ```
- **Remediation:**
  Update the fallback help strings when `projection.Scope != nil` to include `z full graph` (e.g. `"z full · Esc ranks · f pick · resize"`).

**Resolution:** Narrow and capacity fallbacks now explicitly advertise z as the
route back to the full graph whenever one-hop focus is active, with regressions
for both paths.

#### L2. Test gap: View cycling away from focused spatial view lacks focus clearance assertions · **Status:** fixed

- **Consequence:** If a code refactoring accidentally removes `d.focus = nil` from `threadDetail.withDetailView` when switching views (e.g. `v` to `summary`), no test in `internal/tui` fails. A user cycling views could remain stranded in a stale focus state upon returning to spatial mode.
- **Reproduction:**
  Apply mutation 6c: comment out `d.focus = nil` in `internal/tui/detail.go:985`. Run `go test ./internal/tui/...`. All tests pass.
- **Evidence:**
  `thread_focus_test.go` exercises `toggleDetailLocalFocus` and `SetContent` reload, but never asserts that cycling through `withDetailView(string(threadDetailSummary))` and back to `threadDetailSpatial` resets local focus to the full graph.
- **Remediation:**
  Add an assertion in `TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph` verifying that calling `detail.withDetailView(string(threadDetailSummary)).withDetailView(string(threadDetailSpatial))` on a focused detail yields `detailLocalFocusActive() == false`.

**Resolution:** A regression now cycles a focused detail through summary and
back to spatial and requires local focus to be cleared.

#### L3. Test gap: External deletion of focal task on reload lacks focused regression coverage · **Status:** fixed

- **Consequence:** While the production code gracefully falls back to the full graph if a focal task is deleted from the Thread during a focus session, this recovery behavior is not guarded by a regression test.
- **Reproduction:**
  Run `TestThreadSpatialFocusSurvivesCoherentReloadByCanonicalIdentity`. It renames a dependent task (`c` to `c-renamed`) but never mutates or deletes the focal task `b`.
- **Evidence:**
  `internal/tui/detail.go:1091-1094`:
  ```go
  projection, err := core.SelectThreadGraphNeighborhood(d.projection, context.Primary, 1)
  if err != nil {
      return d
  }
  ```
  The error path leaves `d.focus == nil` and safely falls back, but no existing test asserts this specific transition.
- **Remediation:**
  Add a test case to `TestThreadSpatialFocusSurvivesCoherentReloadByCanonicalIdentity` where `pane.SetContent` is invoked with a projection that omits `focalTaskID`, asserting that `reloaded.detailLocalFocusActive() == false` and that a valid fallback selection is chosen.

---

**Resolution:** Reload coverage now deletes the focal task, includes a fuzzy
slug near-match, and requires focus to fail open with a valid non-focal
selection.

### Disproved hypotheses

1. **Hypothesis: Separate lazy layout caches cause concurrency race conditions during async detail loading.**
   *Investigation:* Inspected `threadSpatialCache.get()` and `getForViewport()`. The cache uses an internal mutex and memoizes immutable layout structures. Ran `go test -race ./internal/tui` with parallel model updates.
   *Result:* Disproved. Zero data races detected; concurrent access to independent full and focused caches is safe.
2. **Hypothesis: A fast double-press of `z` or background reload during `directionMenu` chooser could apply a stale target or panic on nil.**
   *Investigation:* Traced `handleDetailDirectionKey` and `selectDetailTask`. If `task.CanonicalID()` is missing from the active projection when `Enter` is pressed, `selectDetailTask` returns `false` and sets `flashErr = "directional target is no longer available"`.
   *Result:* Disproved. The boundary fails closed cleanly without panics or state corruption.
3. **Hypothesis: Reusable Back (`ctrl+o`) restores focus context across different entities (e.g. jumping from Thread A to Task to Thread B).**
   *Investigation:* Inspected `m.pushLoc()` and `m.restoreDetailNavigation()`. Pushed locations associate `kind` and `key` with `detailContext`. Restoration checks `pending.context.Name == threadSpatialFocusContext` and verifies that the primary key exists in the restored Thread projection before creating focus.
   *Result:* Disproved. Cross-entity contamination is impossible.
4. **Hypothesis: Selecting an unreadable node (`RoleUnknown`) and pressing `z` crashes or corrupts the spatial view.**
   *Investigation:* Tested `unreadable := newThreadDetail(...)` with selection `"u"`. `detailLocalFocusAvailable()` calls `threadGraphTask(d.projection, "u")`, which returns `false` because `Slug == ""`. Pressing `z` is a safe no-op.
   *Result:* Disproved. Handled gracefully.

---

### Residual uncertainty

1. **Terminal sizes below 34×7:** In extreme terminal geometries (e.g. 30×5), layout preflight and Lipgloss borders undergo aggressive truncation. While no crashes occur, the visual layout becomes largely unreadable.
2. **High-frequency fsnotify churn while chooser is active:** If a user keeps the directional chooser open while background tools rapidly add and remove tasks, the chooser list remains a static snapshot until closed or committed. Committing an externally removed task displays a flash error, but the chooser does not live-update its item list while open.
3. **Extreme fan-out neighborhoods:** While tested up to 6 direct neighbors, a focal task with >50 direct prerequisites in a single wave will produce a large chooser list requiring vertical scrolling via `visibleTaskPickerRange`. This path is functionally covered by unit tests, but real terminal testing with >50 dependencies was not dogfooded in production.
