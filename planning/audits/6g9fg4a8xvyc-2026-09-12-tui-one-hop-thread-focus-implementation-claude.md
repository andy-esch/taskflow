---
schema: 1
id: 6g9fg4a8xvyc
bucket: closed
area: tui-one-hop-thread-focus-implementation-claude
date: "2026-09-12"
updated_at: "2026-09-12"
---
# Audit: TUI one-hop Thread focus implementation — claude — 2026-09-12

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

Perform an independent adversarial implementation and architecture review of
`planning/tasks/6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md`.
Treat the checked criteria, green tests, and implementation narrative as claims to disprove. This
is a stateful TUI change over a shared graph projection: a locally correct render is insufficient if
view transitions, reloads, history, stale identities, or another adapter make it lie or lose context.

Prioritize a rigorous state-machine and boundary review. Reconstruct ownership from core projection
through Thread detail, shell navigation, modal routing, layout caches, rendering, and documentation.
Do not reward code volume or test count. Demonstrate concrete failures, and distinguish defects in
this task from the already-sequenced information-hierarchy redesign in
`reassess-tui-navigation-and-information-architecture-at-current-scale`.

## Review target

Review the complete working snapshot on branch `feat/tui-thread-focus-subgraph`, based on `main` at
`fb27f99`; `HEAD` is planning-first commit `7cff1ff`. The implementation and closeout planning are
intentionally uncommitted, and `internal/tui/detail_direction.go` plus
`internal/tui/thread_focus_test.go` are untracked. The isolated helper's captured baseline—not
`git diff HEAD` alone—is authoritative.

Primary targets:

- `internal/tui/detail.go`, especially alternate views, focus state, selection, cache ownership,
  opaque navigation context, reload restoration, and task targets;
- `internal/tui/detail_direction.go`, `model.go`, `nav.go`, `overlay.go`, `session.go`, `view.go`,
  `keys.go`, and `help.go`, including modal precedence and shell-owned history/zoom behavior;
- `internal/tui/thread_spatial.go`, including preferred entry identity, responsive preparation,
  full/focused caches, directional movement, scope card, persistent zoom badge, rank terminology,
  hostile-width fallback, and renderer bounds;
- `internal/core/thread_neighborhood.go` and `ThreadGraphProjection` as already-shipped semantic
  inputs that the TUI must consume without reinterpretation;
- `internal/tui/thread_focus_test.go`, affected existing TUI tests, README, architecture, ADR 0006,
  and the active task's evidence.

The modified documentation-planning taskfiles and untracked documentation/theme taskfiles outside
the active focus task are parallel work. Preserve them in the captured snapshot but do not edit,
review, stage, restore, or attribute them to this implementation.

## Intended contract to challenge

The full spatial graph stays the default. In immersive spatial mode only, `z` derives a one-hop
projection around the readable selected task using `core.SelectThreadGraphNeighborhood`; elsewhere
`z` keeps its shell full-screen meaning. The focused graph contains the focal node, every supplied
direct prerequisite and dependent, and only induced supplied edges. It retains exact source health,
topology qualification, external prerequisites, unreadable evidence, shown/hidden counts, and
boundary edges without querying storage or changing planning data.

Full and focused projections have separate lazy, viewport-sensitive layout caches. Entering focus
remembers the exact full-graph selection; leaving restores it. Reload and `ctrl+o` preserve canonical
selection and focus context through generic shell seams without teaching the shell Thread semantics.
A deleted or unreadable focal identity fails open rather than restoring stale layout or opening the
wrong task.

Deliberate entry into the spatial view re-anchors to working context: freshest in-flight member,
then freshest supplied frontier member, then freshest readable member, and only then a nonmember
external prerequisite. Portable activity is `updated_at`, falling back to `created`, with stable-ID
ties. A coherent reload or history restoration reapplies the user's explicit selection instead of
snapping to new work.

In one-hop focus, `h/l` follows only direct supplied edges. One readable target moves immediately;
several open a directional chooser; none reports an explicit dead end. `j/k` remains geometric,
`f`, Enter, yank, Esc, `v`, Atlas, and history retain their existing meanings. Stale chooser targets
cannot replace a current selection after reload. Full-graph navigation remains compatible.

Focused mode is unmistakable even when the collision-aware scope card cannot fit: the fixed header
begins with a high-contrast `ZOOMED · ONE-HOP` badge, constrained fallbacks repeat it, and the full
graph never does. The TUI calls core waves `dependency ranks` and says ranks are not execution
barriers; the stable core/wire/CLI `waves` contract is unchanged. Presentation code owns terminal
geometry, color, keys, and overlays; core owns graph facts.

## Mandatory evidence floor

1. Produce a consumer inventory from graph projection to rendered cells and from every relevant key
   to state mutation. Identify which component owns semantic selection, local focus, shell zoom,
   viewport/cache state, history, modal choice, reload generation, and action target.
2. Build a transition matrix covering summary/ranks/full/focus with `v`, `z`, Esc, q, Atlas,
   workspace/session switching, task open, `ctrl+o`, refresh, resize, and task/thread live reload.
   Test both shell-entered zoom and immersive-owned zoom, including repeated and interrupted paths.
3. Exercise fresh entry and re-entry with zero/one/multiple in-flight tasks; empty/stale/multiple
   frontier candidates; equal, missing, malformed, and reordered activity dates; completed/deferred
   members; unreadable members; external prerequisites whose IDs sort first; and mismatches between
   `View.Members`, `View.Frontier`, nodes, waves, and layout admission.
4. Independently recalculate one-hop nodes, induced edges, boundary edges, original rank indexes,
   health, and counts for root, leaf, isolated member, external prerequisite, fan-in, fan-out,
   diamond, disconnected, degraded, and malformed projections. Prove the TUI neither expands nor
   silently drops supplied evidence.
5. Attack identity restoration: rename/delete the focal, full selection, branch candidate, Thread,
   and unrelated node during focus; reload between chooser open and selection; switch Threads and
   workspaces; return through `ctrl+o`; and make newer generations overtake older loads. Look for
   stale caches, cross-Thread context, wrong task opens/yanks, pan loss, or selection teleportation.
6. Exercise hostile terminals and content: widths/heights around every boundary, huge projections
   near preflight limits, very wide/deep/fan graphs, control/format/wide runes, duplicate labels,
   pathless adapters, broken health, and a scope card with no collision-free corner. Measure bounds,
   deterministic output, allocation retention, and repeated resize/reload cost.
7. Inspect the zoom badge and rank language in dark/light/alternate palettes, low/no-color output,
   clipped headers, narrow/capacity fallbacks, and full mode. Verify redundant textual cues survive
   loss of color and that accent foreground/background remains readable.
8. Run real PTY dogfood against `refine-thread-and-tui-navigation` at representative dimensions.
   Confirm fresh/re-entry centering, direct-edge navigation, chooser behavior, focus/full restoration,
   off-screen indicators, live task edits, and clear mode feedback. Record exact commands and state.
9. Run focused/full/race tests, vet, lint, module-tidiness, planning/audit lint, generated-artifact
   drift checks, and `git diff --check`. Report exact results and environmental exceptions.

Consumer inventory evidence must cite exact paths and symbols. A no-findings verdict must still
show which serious hypotheses were falsified and why the existing tests are capable of detecting
the named failures rather than merely passing nearby behavior.

## Required hostile angles

Perform mutation testing in the sandbox. At minimum temporarily:

- make focus reuse or overwrite the full layout cache;
- drop focal/full selection from reload or `ctrl+o` context;
- let deliberate entry choose an external prerequisite before an in-flight/recent member;
- make same-item reload re-anchor and steal an explicit selection;
- make focused `h/l` use geometry, transitive reachability, or an unreadable neighbor;
- auto-select one branch instead of opening the chooser, and accept a stale chooser result;
- remove boundary/health evidence or relabel a bounded excerpt complete;
- show the zoom badge in full mode, suppress it when the scope card collides, or rely on color only;
- rename stable `waves` in core/wire or let TUI rank/lifecycle terminology imply dispatch; and
- bypass input/canvas limits through duplicated wave/node/edge records.

Map each mutation to the specific test that fails for the intended reason. If a mutation survives,
record the coverage gap even when manual inspection suggests production is correct. On the second
pass, look for a systemic abstraction flaw: generic context that is too stringly typed, optional
interfaces whose composition order loses state, cache pointers that outlive their projection,
modal routing that changes semantics under focus, or tests coupled to helpers that reproduce the
same bug as production.

Do not treat the known dense screen chrome, distant inspector, graph jargon, or need for a richer
contextual help experience as implementation findings unless this patch regresses them or violates
its explicit contract. Those are recorded inputs to the sequenced design task. Do report any
correctness or accessibility defect that must be fixed before this task can close.

## Validation and restoration

Run every inspection, test, capture, temporary fixture, and mutation only in the mandatory isolated
sandbox. Use sandbox-local caches and output paths. Restore every implementation/test/fixture probe
to the helper baseline before verification. Do not fix findings, edit another planning entity,
commit beyond the helper-created baseline, push, install globally, or transfer anything except the
assigned audit.

## Deliverable

Preserve this brief and replace only the reviewer-report placeholder. Include an executive verdict;
the complete isolation/transfer attestation; the consumer and state-ownership inventory; transition
matrix; exact validation/PTTY/hostile/mutation evidence; findings in required grammar with smallest
sound remediation; disproved hypotheses; and residual uncertainty. Leave every finding open for
implementation-owner triage, including valid out-of-scope findings that should become follow-up
tasks.

## Reviewer report

Reviewer: claude (Opus 5), 2026-09-12. All inspection, builds, tests, probes, mutations, and PTY
dogfood ran in the isolated clone below. In the shared checkout I only read the brief and the helper
script (`ls scripts/`, `sed`), ran the helper's `create`, and ran the guarded `transfer`.

### Isolation attestation

`create` (run from the handoff checkout with `--sandbox-parent` set to the session scratchpad):

```
sandbox_path=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/51e54298-a8d2-471b-91a0-f574a5bdd906/scratchpad/isolated-review.XPf9NJ
git_dir=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/51e54298-a8d2-471b-91a0-f574a5bdd906/scratchpad/isolated-review.XPf9NJ/.git
baseline_commit=e6eb603421470ac013aef8295727e018d52edaaa
deliverable=planning/audits/6g9fg4a8xvyc-2026-09-12-tui-one-hop-thread-focus-implementation-claude.md
source_blob=0ea3703383f08098ea6fa4d4dfb20ab6cf6b22cb
source_fingerprint=904cf8f189c3e0e274136d2a158dd325c890d9f5
```

Baseline commit `e6eb603` sits on `7cff1ff` and captures 32 changed paths (16 implementation/doc
paths plus the parallel planning files, which were preserved but not reviewed). After every probe
was restored (`git checkout --` for mutated files, deletion of the untracked probe test, and
`git checkout -- planning/tasks/` after the PTY dependency edit), `verify` reported
`deliverable_changed=false` and `transfer=pending` with the identical path, Git directory, baseline,
blob, and fingerprint. The transfer result is reported in the reviewer's handoff message, because
the helper refuses a second transfer once this file has been copied back.

### Executive verdict

**Not ready to close.** The semantic core of the lens is sound: the TUI consumes
`core.SelectThreadGraphNeighborhood` without reinterpretation, and an independent recalculation of
nodes, induced edges, boundary edges, original rank indexes, health, completeness, header counts,
and the *laid-out* node set matched for all 25 focal cases across chain root/leaf, isolated member,
external gate, fan-in/fan-out, diamond, disconnected, and degraded projections. Full and focused
caches are separate, reload and `ctrl+o` carry focus context, `z` back restores the exact full-graph
viewport (PTY captures 03 and 12 have identical `viewport 3/15 nodes · ◂ layers 6–8/9 ▸ · rows 1/4 ▼`
summaries), badge contrast is at least 4.79:1 in every shipped palette, and the stable `waves`
contract is pinned by golden tests.

Four medium defects block closure, and they share one systemic shape (second pass below): **state
captured at projection time is applied later without re-deriving its meaning, and the tests pin
artifacts that production also emits somewhere else, so the defining behavior goes unobserved.**

- M1: an open directional chooser survives a live reload and commits a target whose direction has
  changed. Reproduced in a real PTY against `refine-thread-and-tui-navigation`: "choose prerequisite"
  moved the selection to a task that had become a dependent.
- M2: the collision-aware scope card is effectively never drawn when the one-hop graph fits the
  terminal, which is the common case. It was absent at 120×34, 160×45, and 200×55 and appeared only at
  100×30. The unit test builds a full-viewport helper canvas, while production intersects the canvas
  with layout bounds.
- M3: `z`'s meaning is decided by detail *content* alone. After `tab` out of the immersive graph, the
  footer advertises `z full`, but `z` toggles a hidden one-hop lens in the unfocused split pane. On an
  unreadable node, `z` silently does nothing.
- M4: ten task-scoped mutations survive (14 of 49 distinct mutations survive overall; the rest are recorded in L1 and L5).
  Several are production-relevant, such as rendering and navigating the
  *full* layout under a `ZOOMED · ONE-HOP` header, reusing a full cache after reload, history
  dropping an explicit non-focal selection, and removing the scope card from production renders.

Five low findings cover identity restoration through a fuzzy resolver, entry-anchor robustness,
clipped focused-header qualification, README demo/terminology drift, and an out-of-scope record-limit
coverage gap.

### Consumer and state-ownership inventory

Projection to cells:

| Stage | Symbol (sandbox path:line) |
| --- | --- |
| Core projection | `core.ProjectThreadGraph` `internal/core/thread_graph.go:54`; `ThreadGraphProjection` `:42` (`Scope *ThreadGraphScope` `:48`) |
| Bounded selector | `core.SelectThreadGraphNeighborhood` `internal/core/thread_neighborhood.go:35`; focal resolved by `resolveTaskReference` `:99` → `internal/core/dependency_graph.go:849` (exact → prefix → contains over ID *and* label, `:884-888`) |
| Detail payload | `threadDetail.projection/spatial/focus/selection` `internal/tui/detail.go:866-874`; `threadSpatialFocus{projection, spatial, focalTaskID, fullSelection}` `:876` |
| Active projection/cache | `activeSpatialProjection` `detail.go:916` → `threadSpatialCache.get/getForViewport` `internal/tui/thread_spatial.go:130,145,157` |
| Preparation | `preflightThreadSpatialInput` `thread_spatial.go:293` (raw node/wave/edge limits); fallback/entry anchor `boundedThreadSpatialFallbackTaskID` `:359`, `threadSpatialPreferredTaskID` `:453` |
| Render | `threadDetail.renderDetail` `detail.go:1021` → `renderThreadSpatialPrepared` `thread_spatial.go:2137`; canvas `renderThreadSpatialCanvasWindow` `:2516` (layout-intersected size `:2525`); scope card `annotateThreadSpatialScopeCard` `:2050` via `:2178`, emptiness `threadSpatialCanvasAreaEmpty` `:2107`; header badge `:2195`; capacity fallback badge `:2494`; narrow title `renderThreadSpatialNarrow` `:2913`; `styles.accentBadge` `internal/tui/style.go:120` |
| Action targets | Enter `detailSelectionTarget` `detail.go:1173` and yank `detailSelectionYankRef` `:1184`, both over the *full* projection via `threadGraphTask` `:1511` |

Keys to state:

| Key | Path |
| --- | --- |
| `z` | `internal/tui/model.go:884-898` → `detailPane.toggleLocalFocus` `detail.go:293` → gate `detailLocalFocusAvailable` `:1107` (view + readable selection only) → `toggleDetailLocalFocus` `:1120`; otherwise `toggleZoom` `model.go:1354` unless `immersive()` |
| `h`/`l` | `model.go:914,917` → `moveDetailDirection` `:983` → `directionChoices` `detail.go:309` → `detailDirectionChoices` `:1145` (focus edges, readable filter) → 0: flash · 1: `selectDetailTask` `:317` · n: `detailDirectionMenu.open` `internal/tui/detail_direction.go:20` |
| chooser | modal registry `internal/tui/overlay.go:31` (`detailDirectionModal` `:117`) → `handleDetailDirectionKey` `detail_direction.go:73` → `selectDetailTask` (presence-in-active-layout check only) |
| `j`/`k` | `moveDetailSelectionDirection` `detail.go:1068` over the active projection/layout |
| `v`/Esc/`q` | `cycleView`/`retreatView` → `withDetailView` `detail.go:983` (clears focus when leaving spatial; re-anchors on spatial entry) → `syncDetailImmersion` `model.go:1368` |
| `ctrl+o` | `pushLoc` `internal/tui/nav.go:254` (context `:265`) → `navBack` `:281` → `restoreDetailNavigation` `:302` (view → context `:318` → selection) |
| reload | `detailMsg` `model.go:346` (gen guard `:347`) → `restoreDetailNavigation` → `SetContent` `detail.go:456` (view → context `:472` → selection) → `syncDetailImmersion(false)` |
| `tab` | `toggleFocus` `model.go:1337` → `toggleZoom` (zoom off, list focus; content stays spatial) |
| Atlas / space | `enterAtlas`/`exitAtlas` `internal/tui/atlas.go` (`atlasResume`); `saveSession` `internal/tui/session.go:95` (pane by value `:107`), `closeTransientUI` `:188` closes chooser `:194` |

Ownership: semantic selection is `threadDetail.selection`, canonicalised against the active layout.
Local focus is `threadDetail.focus`. Shell zoom is `Model.zoom`/`immersiveZoom` (`model.go:80-81`).
Layout caches are one per projection (full `threadDetail.spatial`, focused `focus.spatial`). There is
no stored pan, because pan is recomputed from the selection (`threadSpatialWindowForSelection`
`thread_spatial.go:1857`). History is `Model.navStack` plus `pendingDetailNavigation`. Modal choice
is `Model.direction` (`model.go:100`), a `[]domain.Task` snapshot **with no origin, direction
re-check, or generation**. The reload generation is `Model.detailGen`, which does not reach the
chooser. The action target is the canonical ID over the full projection.

### Transition matrix

Evidence key: **T** = existing test (named), **P** = sandbox probe, **PTY** = capture under
`scratchpad/review-out/pty/`, **C** = code path only.

| From | Input | Observed result | Evidence |
| --- | --- | --- | --- |
| summary/ranks (split) | `z` | shell full-screen toggles; no lens | T `TestThreadSpatialGraphPreservesManualZoomWhileUsingZoomKeyForFocus`; mutation 11a killed |
| ranks | `v` | spatial full graph, re-anchored to in-flight member, no focus | PTY 21, 03 |
| full (immersive-owned zoom) | `z` | focus on selected readable node; zoom ownership unchanged | PTY 04; T `…DirectionalBranchesUseChooser…` |
| full (shell-entered zoom) | `z`,`z`,Esc | lens on/off; Esc → ranks keeps user zoom | T `…PreservesManualZoomWhileUsingZoomKeyForFocus` |
| full | `tab` | split, list focus, spatial content retained; footer `z full` | P, PTY 13 |
| split with spatial content | `z` | **toggles hidden one-hop lens; no full-screen** (M3) | P, PTY 14 |
| full, unreadable node selected | `z` | **silent no-op** (M3) | P |
| focus | `z` | full graph, selection = focal, identical viewport summary | PTY 12 vs 03; T `…RestoresFullGraph` |
| focus | `h`/`l` fan | chooser; 1 target moves; none flashes dead end | PTY 05; T `…DirectionalBranches…` |
| focus, chooser open | live reload reversing an edge, Enter | **stale target committed, no feedback** (M1) | PTY 06–08; P |
| focus, chooser open | reload making focal unreadable, Enter | focus fails open; chooser still says "dependent"; target applied in full graph (M1) | P |
| focus | Enter → task, `ctrl+o` | Thread spatial, focus and selection restored | PTY 10–11; T `…UsesDirectionalStableIdentityNavigation` |
| focus | history restore with non-focal selection | selection `d` kept (focal `b`) | P (production correct; mutation 4a survives, M4) |
| focus | Esc / `q` | ranks; focus cleared | T (Esc after ctrl+o restore); PTY 20 |
| focus | `v` | summary; focus cleared; later spatial entry unfocused | C `detail.go:983-986`; PTY 21; mutation 11b survives (M4) |
| focus | `a`,`a` | Atlas round trip keeps focus and selection | PTY 18–19 |
| focus | live task edit restoring the edge | focus and explicit selection preserved | PTY 09 |
| focus | `r` refresh | preserved | PTY 22 (full), T reload test (focus) |
| focus | focal deleted by reload | focus fails open, selection falls to layout | P (`selected="c"`) |
| focus | focal deleted but another slug contains its ID | **focus re-centres on a different task** (L1) | P |
| focus | resize 200×55 → 100×30 → 80×24 → 60×16 → 59×13 → 40×12 | badge in every state; narrow fallback `ZOOMED · ONE-HOP · give it room`; card only at 100×30 (M2); qualification clipped at ≤100 (L3) | PTY 15–16 |
| workspace switch | save/restore | pane saved by value, chooser closed, `reloadAll` re-derives focus | C `session.go:95-155` (not PTY-exercised; isolated config had no registered spaces) |

### Validation results (sandbox)

Environment: `GOCACHE`/`GOLANGCI_LINT_CACHE` in the scratchpad, `GOFLAGS=-mod=readonly`,
`GOPROXY=off`, `GOTOOLCHAIN=local`, go1.26.6 darwin/arm64. The TUI ran with
`XDG_CONFIG_HOME=<scratchpad>/review-xdg`, and the user `spaces.toml` mtime was unchanged.

| Command | Result |
| --- | --- |
| `go build ./...`; `go build -o <scratchpad>/review-out/tskflwctl ./cmd/tskflwctl` | ok |
| `go test -race -count=1 ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `golangci-lint run ./...` | `0 issues.` |
| `go mod tidy -diff` | exit 0 |
| `go run ./internal/tools/docgen -out docs/cli` then `git status --porcelain -- docs/cli` | no drift |
| `tskflwctl lint` | `✔ all planning entities and dependency links pass lint` |
| `tskflwctl audit lint 6g9fg4a8xvyc` (before editing) | `✔ all audit findings pass lint` |
| `git diff --check 7cff1ff` and `git diff --check fb27f99` | exit 0 |

Environmental exception: the first non-race `go test ./...` with `TMPDIR` inside the scratchpad
failed `TestDiagnoseRegistry_AllSupportedStates` and
`TestFS_OpenWorkspaceDoesNotFallbackFromMissingOrMalformedEntry`. The harness session directory
above the scratchpad contains a `tasks/` directory, which workspace discovery treats as a planning
root. Both packages pass with the system temp directory, which is the configuration used for every
result above. The failure is unrelated to this change.

### PTY dogfood (real Thread)

The driver was a private tmux socket in the scratchpad with sandbox cwd:
`tmux -S r.sock new-session -d -x W -y H -c <sandbox> "env XDG_CONFIG_HOME=… TERM=xterm-256color COLORTERM=truecolor <scratchpad>/review-out/tskflwctl ui"`.
The key path was `] ] ]` (Threads), `j j` (`refine-thread-and-tui-navigation`), `Tab v v`.

- Entry centres the sole in-flight member `add-a-one-hop-focus…` (M8), not an earlier external gate
  (capture 03).
- `z` produces `ZOOMED · ONE-HOP  spatial graph · 6/15 shown · 9 hidden · graph healthy · projection healthy · complete`
  and `8 boundary edge(s) · z full graph`, with correct direct neighbours G1, G2, M1, M2 (prerequisites)
  and M4 (dependent) (capture 04, 160×45; capture 15, 200×55). **No scope card** is drawn despite a
  mostly empty canvas.
- `h` opens `choose prerequisite · 1/4` with the four direct prerequisites (capture 05).
- Stale chooser (M1): cursor on `export-bounded-thread-neighborhoods-around-a-task`, then in the
  sandbox `tskflwctl task depend remove 6g86c7y6hn41 --on 6g9150nrt4p9` and
  `tskflwctl task depend add 6g9150nrt4p9 --on 6g86c7y6hn41`. The reload redrew export-bounded as a
  dependent while the chooser still read "choose prerequisite" (capture 07). Enter selected it with
  no flash, and its inspector reads `needs [M2] add-a-one-hop-focus…` (capture 08). The sandbox
  planning was then restored with `git checkout -- planning/tasks/`, and the live reload kept focus
  and selection (capture 09).
- Enter opened the task, and `ctrl+o` returned to the focused graph with the same selection
  (captures 10–11). `z` restored the full graph and viewport (capture 12).
- `tab` then `z` (M3): the split footer reads `… l / ⏎ detail · z full · ? help · q quit`, yet the
  detail pane switched to `ZOOMED · ONE-HOP` (captures 13–14).
- `NO_COLOR=1 TERM=xterm`: the header carries `\x1b[1m ZOOMED · ONE-HOP \x1b[0m`, so bold uppercase
  text survives without colour (capture 17). Truecolor neon uses `38;2;5;6;8` on `48;2;234;92;226`.
- Atlas round trip, `q`, `v` re-entry, and `r` behaved as in the matrix (captures 18–22).

### Hostile and scale probes (sandbox-only `internal/tui/zz_review_probe_test.go`, deleted)

- **Independent one-hop recalculation.** For each readable focal in 7 fixture classes (25 cases), the
  expected node set, induced/boundary edge counts, original wave indexes, health, completeness,
  header `N/T shown · H hidden`, `B boundary edge(s)`, and the layout's `byID` set all matched. The
  unreadable focal was correctly unavailable. This falsifies expansion and silent drop.
- **Malformed input.** A duplicate edge or duplicate node makes `z` flash
  `validation failed: Thread graph repeats edge a -> b` / `repeats task ID a`, and the full graph
  remains. The lens fails closed.
- **Scale.** With a 40-dependent fan, focus plus render took 1.8 ms and the chooser offered 40 entries.
  50 reload-plus-resize cycles took 77 ms at 41 nodes and 67 ms at 601 nodes, with about 0 KB
  retained heap after GC, and focus stayed active. With 600 dependents, the full graph is over capacity, and entry anchored
  `dep-0000` instead of the in-flight focal (L2), so `z` gave `2/601 shown`.
- **Scope-card placement.** On the fixture, the card was drawn only at 120×30. At 60×14, 80×20,
  100×24, 160×36, 180×40, and 240×60 it was absent (M2).
- **Palette contrast (WCAG).** Neon dark `#050608` on `#ea5ce2` is 6.88:1. Latte light (neon and
  catppuccin) `#eff1f5` on `#8839ef` is 4.79:1. Mocha `#1e1e2e` on `#cba6f7` is 8.07:1.
- **Chooser hostile slug.** `evil\x1b]52;c;…\x07slug` reaches the chooser output raw. This is identical
  to the pre-existing follow picker and list rows (`nav.go` follow view, `entity.go:139`), so it is not
  attributed to this patch; see residual risks.

### Mutation evidence

Harness: 49 distinct mutations, each an exact counted string replacement, `go test -count=1 -skip TestReviewProbe <pkgs>`, then
`git checkout --` of the touched files. Post-run `git status` showed only the (later deleted) probe
file.

| # | Mutation | Result | Killing test / gap |
| --- | --- | --- | --- |
| 1a | focus cache field set to full cache | killed | `TestThreadSpatialFocusUsesPreferredIdentityAndRestoresFullGraph` (pointer check `thread_focus_test.go:108`) |
| 1b | `activeSpatialProjection` returns full cache while focused | **survived** | M4: focused header over full layout is unobserved |
| 1c | focus overwrites `d.spatial` | killed | same |
| 1d | reload-restored focus uses fresh *full* cache | **survived** | M4 |
| 2a | `SetContent` drops context | killed | `TestThreadSpatialFocusSurvivesCoherentReloadByCanonicalIdentity` |
| 2b / 2c | `pushLoc` / restore drop context | killed | `TestThreadSpatialGraphUsesDirectionalStableIdentityNavigation` |
| 2d / 2e | ignore `Secondary` / `fullSelection` | survived | unobservable: `fullSelection` always equals the focal (L1 note) |
| 2f | reload drops selection restore | killed | reload tests ×3 |
| 3a | entry prefers external gate | killed | 4 focus tests |
| 3b | entry skips frontier tier | killed | `…UsesDirectionalStableIdentityNavigation` |
| 3c | activity tie-break reversed | killed | preflight/runtime cache tests |
| 3d | ignore `created` fallback | **survived** | M4 |
| 3e | pick stalest in-flight | killed | `TestThreadSpatialEntryReanchorsToWorkWhileReloadRestorationRemainsExplicit` |
| 3f | deliberate entry keeps stale selection | killed | 6 tests |
| 3g | remove scope-card hostile-size guard | **survived** | M4 |
| 4a | history restore drops explicit selection | **survived** | M4 (test selection equals focal and anchor) |
| 4b | reload applies context after selection | killed | reload focus test |
| 5a–5d | geometry / full-graph edges / transitive / unreadable neighbour | killed | `TestThreadSpatialFocusDirectionalBranchesUseChooserAndDeadEndsExplain` |
| 6a | auto-select first branch | killed | same |
| 6b | `selectDetailTask` accepts unverifiable target | killed | reload focus test (`removed-target`, `:189`) |
| 6c / 6d | silent dead end / modal unregistered | killed | directional test |
| 6e | `closeTransientUI` keeps chooser | **survived** | M4 |
| — | chooser commits after reload changed direction | no guard exists to mutate | M1 |
| 7a–7d | drop boundary / health / counts; relabel excerpt complete | killed | focus render and degraded tests |
| 7e | scope card never drawn in production | **survived** | M2/M4 |
| 8a | badge in full mode | killed | 3 tests |
| 8b / 8c | header badge removed / colour-only | killed | `…RestoresFullGraph` (only because M2 hides the card at 180×40) |
| 8d | capacity fallback loses badge | **survived** | M4 |
| 8e | narrow fallback loses badge | killed | `TestThreadSpatialFocusHandlesExternalIsolatedDegradedAndNarrowCases` |
| 9a | wire `json:"waves"` → `ranks` | killed | `TestGolden_MachineContract/*`, `TestThreadPreviewReleaseWireSemanticsRemainCompatible/*` |
| 9b / 9c / 9e | TUI implies barrier / legend / column label | killed | topology, legend, and column-label tests |
| 9d | help says ranks are dispatch order | **survived** | L4 |
| 10a | core selector accepts duplicate nodes (core, tui, cli, wire suites) | **survived** | L5 (out of scope) |
| 10b | TUI preflight ignores raw node records | killed | preflight/capacity tests |
| 10c | TUI preflight ignores raw wave records | **survived** | L5 (pre-existing) |
| 11a | lens offered outside spatial | killed | manual-zoom test |
| 11b | leaving spatial keeps focus pointer | **survived** | M4 |
| 11c | lens offered on unreadable node | killed | narrow/degraded test |

### Findings

#### M1. The directional chooser commits a snapshot target after a reload changed the edge it represented · **Status:** fixed

**File:** `internal/tui/detail_direction.go:73` | **Component:** tui chooser / detail selection seam
**Effort:** S · **Urgency:** soon

`moveDetailDirection` (`internal/tui/model.go:983`) projects `detailDirectionChoices` once and
stores `[]domain.Task` in `Model.direction`. Nothing binds that snapshot to the Thread key, the
origin selection, the direction, the focus context, or `detailGen`. `detailMsg` (`model.go:346`)
replaces the detail under an open chooser, and `closeTransientUI` is not called. On Enter,
`handleDetailDirectionKey` calls `detailPane.selectDetailTask` (`internal/tui/detail.go:317`), which
only checks that the ID is selectable in the *current active layout*. The contract says stale
chooser targets cannot replace a current selection after reload. The only guard tested is a target
absent from the layout (`thread_focus_test.go:189`).

Evidence:

- Real PTY, sandbox graph edit (captures 06–08). "choose prerequisite", then reverse the edge so
  export-bounded becomes a dependent, then Enter. The selection lands on export-bounded, whose
  inspector says `needs [M2] add-a-one-hop-focus…`. There is no flash.
- Model probe on the fixture. With `l` and cursor on `d`, a reload changes `b→d` into `d→b`. The
  current choices for `b` are dependents `[c]` and prerequisites `[a d x]`. After Enter the selection
  is `d`, `flash=""`.
- Model probe on the fixture. The chooser is opened, then a reload makes focal `b` unreadable, so
  focus fails open. The chooser still shows `label="dependent"`, and Enter applies `c` in the full
  graph.

The same shape (capture-then-apply without re-derivation) recurs in L1 and M3; it is this review's
systemic finding.

**Recommendation:** Store the origin (`loadedKey`, origin selection key, `dx`, and focus context) in
`detailDirectionMenu`. On Enter, recompute `m.detail.directionChoices(dx, 0)`, and require the same
origin and membership of the target in the fresh choices; otherwise flash and close. Alternatively,
close or revalidate the chooser whenever a `detailMsg` lands. Add a model-level regression that
delivers a `detailMsg` between `l` and Enter for both the edge-reversal and the focus-exit cases.

**Resolution:** Directional menus now record their source identity and
direction, re-derive valid neighbors before committing, and close visibly on
coherent refresh; model regressions cover edge reversal and refresh while open.

#### M2. The scope card never renders when the one-hop layout fits inside the viewport · **Status:** fixed

**File:** `internal/tui/thread_spatial.go:2050` | **Component:** tui spatial renderer
**Effort:** S · **Urgency:** soon

`renderThreadSpatialCanvasWindow` allocates a canvas intersected with layout bounds
(`thread_spatial.go:2525`). `annotateThreadSpatialScopeCard` computes corner positions from the
*viewport* width and height. `threadSpatialCanvasAreaEmpty` (`:2107`) treats any out-of-canvas cell
(`cellAt` → `!ok`) as occupied. When the focused graph is narrower or shorter than the terminal,
which is the normal case for a one-hop excerpt, every corner is rejected. The card then "yields"
over visibly blank space. Acceptance criterion 2 is checked, but in the dominant case the card is
not delivered.

Evidence: real Thread, card absent at 120×34, 160×45, and 200×55, and present only at 100×30
(captures 04, 15, 16-focus-100x30, 17). On the fixture it rendered at 120×30 and was absent at 160×36,
180×40, and 240×60. The unit test `TestThreadSpatialScopeCardUsesOnlyEmptyCanvasSpace` builds
`newThreadSpatialCanvas(80, 20)` (`thread_focus_test.go:199`), a full-viewport canvas that production
never produces. The render assertions (`:114`) look for strings the header also carries, so
mutation 7e (never draw the card) survives.

**Recommendation:** Treat cells outside the layout-intersected canvas but inside the viewport as
empty, and draw the card in rows that `renderLine` already emits blank. The simpler alternative is to
allocate the card area separately and overlay it after `renderLine`. Add a regression through
`renderDetail` on a small focused graph at a large size that asserts the card border (`╭─ ZOOMED`)
below the fixed header rows.

**Resolution:** Focused rendering now retains bounded blank viewport cells under
the existing allocation ceiling, and the production render regression requires
the bordered scope card at a spacious terminal size.

#### M3. The `z` key is routed by detail content state, so its meaning diverges from the visible shell mode · **Status:** fixed

**File:** `internal/tui/model.go:884` | **Component:** tui key routing / local-focus seam
**Effort:** S · **Urgency:** soon

`toggleLocalFocus` runs before any shell-state check. `detailLocalFocusAvailable`
(`detail.go:1107`) consults only `view == spatial` plus a readable selection, not `m.zoom`,
`m.focus`, or whether the pane is immersive on screen. The contract reserves `z` for immersive
spatial mode and keeps the shell full-screen meaning elsewhere.

1. `tab` from the immersive graph (`toggleFocus` → `toggleZoom`, `model.go:1337,1354`) leaves spatial
   content in the split with list focus. The footer advertises `z full` (`view.go:491`), but `z` toggles
   the one-hop lens in the unfocused pane (probe `localFocus=true zoom=false focus=list`; PTY captures
   13–14). Before this patch, `z` was a no-op there; now it silently changes hidden state.
2. With an unreadable node selected in immersive mode, `toggleLocalFocus` returns `handled=false`,
   and `model.go:895` skips `toggleZoom` because content is immersive. `z` does nothing and gives no
   feedback (probe `focus=false … flash=""`). The `ErrValidation` "not readable" branch in
   `toggleDetailLocalFocus` (`detail.go:1120+`) is unreachable from the shell.

**Recommendation:** Gate the lens on the shell presentation as well, for example
`m.zoom && m.focus == focusDetail && m.detail.immersive()` in the `keys.Zoom` case. When the content
is immersive but the lens is unavailable, flash the existing validation reason instead of silently
consuming `z`. Add model tests for `tab`→`z` and unreadable→`z`.

**Resolution:** The model routes z to local focus only when zoomed detail focus
visibly presents the immersive graph; split mode restores shell full-screen
behavior and unreadable focal attempts report validation instead of silently
changing shell state.

#### M4. Focus tests pin struct pointers and header strings instead of the rendered, navigated, and restored behavior · **Status:** fixed

**File:** `internal/tui/thread_focus_test.go:108` | **Component:** tui tests
**Effort:** M · **Urgency:** soon

Ten task-scoped sandbox mutations survive the full TUI suite. The other four survivors (2d, 2e, 10a,
10c) are covered by L1 and L5:

- **1b**: focused mode renders and navigates the *full* cached layout under the `ZOOMED` header. The
  test only compares cache pointer fields, and the fixture's full layout also contains every asserted
  string.
- **1d**: reload-restored focus builds on the full cache.
- **4a**: history drops the explicit selection. The existing test's selection coincides with both the
  focal and the entry anchor.
- **7e**: the scope card is never drawn (see M2).
- **8d**: the capacity fallback loses the badge.
- **11b**: leaving spatial keeps the focus pointer, so a later `v`/Esc re-entry is untested.
- **6e**: session switch leaves the chooser open.
- **3d**: the `created` fallback for activity is ignored.
- **3g**: the scope-card hostile-size guard is removed.
- **9d**: help text says ranks are dispatch order.

Production is correct for 1b, 1d, 4a, 6e, 11b, and 3d by direct probe or code reading, but nothing
would catch a regression. The same masking hid the real M2 defect: the render test at 180×40 passes
on header text alone.

**Recommendation:** Assert behavior at the seams production uses:

- Compare the rendered and laid-out node set (`spatialPreparedForViewport(w).layout.byID`) to
  `Scope.ShownNodes`, both after `z` and after a `SetContent` reload.
- Drive `ctrl+o` with a selection that differs from the focal and from `threadSpatialPreferredTaskID`.
- Cover focus → Esc/`v` → spatial re-entry.
- Render a focused projection above capacity and assert the badge.
- Cover `activateWorkspace` with an open chooser.
- Cover an activity case that relies on `created`.
- Assert the scope card through `renderDetail` (see M2).

**Resolution:** Regressions now assert scoped layout membership after entry and
reload, full-graph alias continuity, non-focal history restoration, view-cycle
cleanup, production scope-card rendering, capacity and narrow badges, validated
created-date fallback, help semantics, stale chooser behavior, and deleted-focal
recovery.

#### L1. Focus restoration resolves the opaque focal ID through the fuzzy task-reference resolver · **Status:** fixed

**File:** `internal/tui/detail.go:1087` | **Component:** tui detail navigation context
**Effort:** XS · **Urgency:** eventually

`withDetailNavigationContext` passes `context.Primary` to `core.SelectThreadGraphNeighborhood`. That
function resolves the focal through `resolveTaskReference` (`thread_neighborhood.go:99`,
`dependency_graph.go:884-888`), which falls back to prefix and substring matches over IDs *and* slug
labels. `toggleDetailLocalFocus` checks exact readability first, but the restore path does not.
Restore also stores `focalTaskID: context.Primary` rather than the resolved `Scope.FocalTaskID`.

Probe with realistic 12-character IDs: focal `6g86c7y6hn41` is deleted, and another member's slug is
`follow-up-to-6g86c7y6hn41`. After reload, the focus is *restored* around `6g9aaaaaaaaa`, with stored
focal `6g86c7y6hn41`, header `ZOOMED · ONE-HOP … 2/2 shown`, and `fullSelection` / leaving selection
`6g9aaaaaaaaa`. This contradicts "a deleted focal identity fails open."

A related issue: `fullSelection` falls back through `threadGraphSelectedTaskID`, which never returns
`""` for a non-empty graph and orders external gates first (`detail.go:1466-1486`). The
`threadSpatialPreferredTaskID` branch is therefore dead, and `fullSelection` is otherwise always equal
to the focal (mutations 2d and 2e are unobservable).

**Recommendation:** Before selecting, require `threadGraphTask(d.projection, context.Primary)` to
match exactly, mirroring the toggle, and store `projection.Scope.FocalTaskID`. Either drop
`Secondary`/`fullSelection` as redundant, or use a fallback that can actually reach the preferred
anchor.

**Resolution:** Opaque navigation restoration now requires an exact readable
canonical focal identity before invoking the user-facing resolver and stores the
selector's resolved canonical ID; deletion with a fuzzy slug near-match fails
open in regression coverage.

#### L2. The entry anchor can be won by a malformed date, skip in-flight work in large Threads, or prefer an unreadable member over a readable gate · **Status:** fixed

**File:** `internal/tui/thread_spatial.go:359` | **Component:** tui spatial entry anchor
**Effort:** S · **Urgency:** eventually

Three probes each break a stated tier:

- **(a) Malformed dates.** Activity compares `updated_at`/`created` as strings. Task `updated_at` is
  not lint-validated (`internal/domain/lint.go:115-120` validates only `created`). An in-flight member
  with `updated_at: "2026-9-1"` or `"someday"` beats an in-flight member dated `2026-09-12`. The idiom
  mirrors `epicLastUpdated` (`internal/core/service_epic.go:230`), but here it chooses a user-facing
  anchor.
- **(b) Large Threads.** The node scan stops at `index > threadSpatialMaxNodes` over ID-sorted nodes
  (`:406`). In a 601-node Thread whose in-flight member sorts last, entry anchors `dep-0000`, so `z`
  yields `2/601 shown`. That is the only spatial path left once the full graph is over capacity.
- **(c) No readable member.** The final `first` fallback spans all node IDs, so `a-unreadable` is
  chosen over readable external gate `z-gate`, contrary to the tier "only then a nonmember external
  prerequisite."

**Recommendation:** Parse activity with `domain.ValidateDate`/RFC 3339 and rank unparseable values
below valid ones. Select in-flight and frontier candidates from the bounded
`View.Members`/`View.Frontier` roles rather than the ID-ordered node prefix. Restrict the last
fallback to readable external gates before unreadable nodes.

**Resolution:** Entry ranking validates updated_at and created dates, scans
supplied Thread members rather than an arbitrary bounded node prefix, and
prefers a readable external prerequisite over unreadable evidence only after
readable member tiers are exhausted; hostile cases are covered.

#### L3. The focused header clips projection health and topology qualification below about 105 cells · **Status:** fixed

**File:** `internal/tui/thread_spatial.go:2195` | **Component:** tui spatial header
**Effort:** XS · **Urgency:** eventually

Bounded mode places a styled badge and counts before health and completeness. For a 15-node Thread,
the qualification needs 105 cells
(` ZOOMED · ONE-HOP   spatial graph · 6/15 shown · 9 hidden · graph healthy · projection healthy · complete`).
PTY 100×30 shows `… projection healthy · …` with topology cut, and 80×24 shows
`… graph healthy · p…` with projection health and completeness cut. Full mode shows
`spatial graph · complete` within 23 cells. The contract says focus "retains … topology
qualification." At common 80–104-column widths, a partial or degraded excerpt loses that
qualification only in bounded mode.

**Recommendation:** Emit a compact health/completeness token before the counts, for example
`ZOOMED · ONE-HOP · partial · graph broken/projection healthy · 6/15`, or move the counts to the
second fixed row that already carries boundary evidence.

**Resolution:** Focused headers now keep the badge, topology qualification, and
shown count on the first row, then lead the second row with graph and projection
health before optional hidden, boundary, return, and viewport detail.

#### L4. The README demo and remaining TUI strings still present the retired wave vocabulary · **Status:** tracked by 6g9g37bvetjy

**File:** `README.md:31` | **Component:** docs / tui terminology
**Effort:** S · **Urgency:** eventually

README lines 31–34 now say the topology view "groups work into explanatory dependency ranks—not
execution barriers" directly above `assets/threads.gif`. That GIF was last changed in `f1b2be9`
(2026-09-06). A frame extracted with ffmpeg shows `view: topology · member waves plus bounded
dependencies` and `Wave 1 · explanatory order, not a barrier`. User-visible TUI capacity-guard issues
still read `wave records exceeds…` and `wave task records exceeds…` (`thread_spatial.go:312,327`),
and the new help sentence is unpinned (mutation 9d survives).

**Recommendation:** Regenerate `assets/threads.gif` via `just gifs` (and consider showing the
spatial/focus lens). Rename the two TUI capacity strings to dependency-rank wording, and pin the help
entry.

**Resolution:** Runtime capacity diagnostics now use dependency-rank terminology
and Threads help pins the non-barrier contract; the stale recorded GIF is
sequenced after the navigation redesign in task 6g9g37bvetjy to avoid recording
the UI twice.

#### L5. Record-limit guards that the focus path depends on are not pinned by tests · **Status:** tracked by 6g9g37c50qp1

**File:** `internal/core/thread_neighborhood.go:49` | **Component:** core selector / tui preflight tests (out of scope, pre-existing)
**Effort:** XS · **Urgency:** eventually

Removing the selector's duplicate-node rejection survives the `internal/core`, `internal/tui`,
`internal/cli`, and `internal/wire` suites (mutation 10a). Disabling the TUI preflight's raw
wave-record limit (`thread_spatial.go:309-315`, mutation 10c) survives the TUI suite. Production
behavior is correct today: `z` flashes `validation failed: Thread graph repeats task ID a`. These
guards shipped with the neighborhood-export and spatial prototype tasks, so this belongs in a
follow-up rather than this task.

**Recommendation:** Add table cases for duplicate node, duplicate wave task, and repeated wave
index to `TestSelectThreadGraphNeighborhoodRejectsBadScopeAndFocus`, and a preflight subtest with more
than 512 empty wave records.

**Resolution:** The pre-existing core and raw dependency-rank record-limit
mutation gaps are scoped in task 6g9g37c50qp1 and added to the active navigation
Thread after the focus implementation.

### Disproved hypotheses

- **The TUI expands or drops supplied evidence.** Disproved by an independent recalculation that
  matched in 25/25 focal cases, including the laid-out node set, plus mutations 7a–7d and 5b being
  killed.
- **`h`/`l` in focus uses geometry, transitive reachability, the boundary, or unreadable neighbours.**
  Mutations 5a–5d are killed by the directional test, and PTY 05 lists exactly the four direct
  prerequisites.
- **Focus reuses or overwrites the full cache in production.** Code reading (`detail.go:916,1120`),
  the pointer checks (1a/1c killed), and PTY 12 vs 03 (identical viewport after leaving) disprove it.
  The test gap is M4 (1b).
- **Reload or `ctrl+o` loses focus context or reorders restoration.** Mutations 2a–2c and 4b are
  killed; PTY 09 and 11 confirm.
- **Same-item reload steals an explicit selection for the entry anchor.** Mutation 2f is killed, and a
  direct history probe keeps non-focal `d`.
- **Deliberate entry picks an external gate before in-flight or recent work.** Mutations 3a, 3b, 3e,
  and 3f are killed; PTY 03 centres the in-flight member. Edge cases remain in L2.
- **The badge appears in full mode or relies on colour.** Mutations 8a–8c are killed. NO_COLOR output
  keeps bold uppercase text, and palette contrast is ≥4.79:1.
- **Stable `waves` changed in core, wire, or CLI.** Mutation 9a is killed by golden and preview
  compatibility tests. CLI `thread show` still prints `Wave N` (`TestThreadTopologyKeepsCLIAndGraphExportProjectionEvidenceAligned`).
- **Focus bypasses input or canvas limits via duplicated records.** The core selector rejects them
  (probe flash), and the TUI raw-node preflight is pinned (10b killed). Coverage gaps are in L5.
- **Allocation retention or reload/resize cost regresses.** 50 reload-plus-resize cycles took under
  80 ms with about 0 KB retained heap at 601 nodes.

### Residual risks (not findings)

- Terminal control sequences in slugs reach the chooser raw. That is identical to the existing follow
  picker and list rows; it is a TUI-wide sanitization question and is not introduced by this patch.
- At a one-hop boundary, the dead-end flash ("no readable direct prerequisite in this one-hop focus")
  fires even when a readable prerequisite exists just outside the excerpt. The chooser also omits
  `[G#]`/`[M#]` role cues, and "focus" now names the inspector, the header column, and the `z` lens.
  These are inputs for `reassess-tui-navigation-and-information-architecture-at-current-scale`, not
  contract violations.
- Workspace/session switching was verified by code path only (`session.go:95-155`): the isolated
  config had no registered spaces. A light-background PTY run was not captured; light contrast is
  computed from `internal/design/theme.go` hexes.
- `z` on a Thread above capacity computes the selector over the whole projection synchronously in
  `Update`. It is pure and measured under 1 ms at 601 nodes, but it is unbounded by the spatial
  limits.
- Generation overtaking (`detailGen`) and pending `ctrl+o` restoration were not newly
  mutation-tested. Both are pre-existing mechanisms that this patch only extends with an extra
  comparable field.
