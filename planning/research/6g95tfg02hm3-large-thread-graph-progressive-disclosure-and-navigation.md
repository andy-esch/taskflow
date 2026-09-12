---
schema: 1
id: 6g95tfg02hm3
created: "2026-09-12"
description: Evidence and interaction design for preserving full-DAG orientation while adding truthful bounded focus across TUI, CLI, PR, and web adapters.
tags: [threads, tui, graph, ux, navigation, design]
---
# Large Thread graph progressive disclosure and navigation study

## Decision to make

Choose a presentation and navigation model that lets the TUI retain a truthful full Thread DAG while
making local causal inspection useful when the terminal can show only a fraction of the graph. The
choice must also define which bounded-graph behavior is shared by CLI, TUI, PR, and future web
adapters.

## Evidence after the current hardening work

This pass evaluates the shipped responsive layout, dense-route grammar, cached layout, and
viewport-bounded renderer rather than the earlier prototype. It reuses the
[spatial experience review](../audits/6g8bbjgcx9tj-2026-09-09-spatial-thread-graph-tui-experience-design-review.md)
as baseline evidence where the underlying behavior remains relevant.

Fresh captures used the repository's current `refine-thread-and-tui-navigation` Thread through a
binary built from `main` plus the lifecycle-only start of this research task. VHS 0.11.0 drove real
PTYs at the same calibrated sizes as the earlier review. The production projection contained 15
nodes, 16 edges, six layout layers, five rows, one in-flight member, and complete healthy topology.

| Shape and viewport | Current evidence | User-job result | Design implication |
| --- | --- | --- | --- |
| Small sparse touring-bike DAG, 8 nodes/8 edges, 140×36 | Earlier visual review made convergence and its external gate legible in one glance. | Understand sequence, convergence, and a direct blocker: strong. | Preserve the full node-link view; it solves a job waves cannot. |
| Production DAG, 15 nodes/16 edges, 140×36 | Fresh capture shows 4/15 nodes, layers 1–2/6, rows 1–2/5. Whole nodes, viewport inventory, gutters, and boundary routes are now truthful. | Follow a visible edge: good. Find current work or understand the whole initiative: weak. | Honesty is fixed; initial focus and global orientation are not. |
| Production DAG, ordinary 100×28 | Fresh capture shows 2/15 nodes, layers 1–2/6, row 1/5. Responsive aliases survive, but the inspector and route fan dominate. | Local inspection: usable. Overview and known-task location without `f`: poor. | Treat the canvas as a viewport, not as a miniature overview. |
| Production DAG, narrow 72×24 | Fresh capture still shows two complete compact nodes and hidden-extent cues; chrome and inspector text truncate. Below 60×14 the explicit narrow fallback replaces the canvas. | Escape and picker remain available, but topology comprehension is marginal. | Keep an honest fallback; do not invent a denser automatic representation. |
| Deep 30-node chain | Responsive/deep/long-route regressions pass; semantic horizontal traversal follows supplied edges and visible rendering remains viewport-bounded. | Walk a path: strong. Regain whole-chain position: dependent on the viewport line. | A compact position rail helps; automatic subgraphing does not. |
| Wide root → 10 siblings → sink | The axis optimizer keeps the selection whole and maximizes complete rows. | Find a sibling: `j/k` works. Choose one of many direct dependencies: ambiguous without an explicit branch choice. | Branching `h/l` needs a neighbor chooser, not a hidden tie-break. |
| Dense 5×4 layered DAG, 76 edges | The fixture remains bounded and deterministic; crossings, shared routes, conflicts, and endpoint counts stay semantically distinct. | Verify that a named edge exists: possible. Comprehend the network globally: poor despite truthful glyphs. | Focus and waves are complementary lenses; more route grammar alone will not solve density. |
| Disconnected, missing-member, and degraded evidence | Hostile/deep/wide/partial projection tests retain health and supplied evidence; capacity overflow fails open to waves. | Diagnose health: strong. A focused excerpt could be mistaken for completeness without scope metadata. | Every bounded view must carry health plus shown/hidden/boundary scope. |

The strongest fresh observation is the entry point. The spatial view opens on `[G1]`, the first
stable historical external gate, while the sole in-flight task is off-screen. The viewport summary
admits that most of the graph is hidden, but the default camera still answers “where did this
initiative begin?” instead of “where is work happening now?”

## Strategy comparison

| Strategy | Comprehension and navigation | Implementation and portability | Accessibility | Verdict |
| --- | --- | --- | --- | --- |
| **A. Full canvas only, with more overview chrome** | Retains the flagship whole-DAG promise. A layer/row minimap improves position, but 2–4 visible nodes still cannot answer local fan-in/out cheaply. | Lowest semantic change; mostly TUI chrome. A terminal minimap does not help CLI/PR/web consumers. | More permanent chrome costs scarce rows and adds another compact encoding to learn. | Keep improving the full canvas, but insufficient alone. |
| **B. Full canvas plus explicit bounded focus** | Full view remains available and default. One-hop focus answers immediate cause/effect; stable return restores the full viewport. Search handles known-task jumps, and an explicit neighbor chooser handles branches. | A pure bounded-selection contract can serve CLI/PR/TUI/web; only layout, keys, and viewport history remain TUI-owned. | Clear scope text and boundary continuations avoid relying on color or spatial memory. | **Recommended.** Adds a local lens without demoting the full graph. |
| **C. Waves/focus in TUI, rich full graph only on web** | Makes terminal use simpler and gives a future web renderer more space, but removes the one view that already explains small-DAG convergence better than waves. | Web is a natural additional adapter over the same projection, but introduces a second product surface before the terminal model is settled. | Browser zoom and pointer input help some users; keyboard and offline terminal parity regress. | Valid future extension, not a replacement or current prerequisite. |

Collapsed waves/layers and other semantic zoom are intentionally not selected. They introduce
synthetic aggregate nodes and a second topology grammar without evidence that the current sparse
DAGs require them. Revisit only if full-plus-focus still fails against larger real Threads.

## Recommended first usable model

1. **Keep the full spatial graph as the default and flagship.** Never switch representations from a
   hidden node-count threshold. Waves remain the better ordered/status scan, not a downgrade path.
2. **Choose an operational entry anchor.** On first spatial entry, select the first in-flight member;
   otherwise the first eligible frontier member; otherwise the existing stable fallback. Reload and
   return continue preserving an existing canonical selection.
3. **Add an explicit one-hop focus lens.** It contains the selected node, every supplied direct
   prerequisite and dependent, and only supplied edges among that set. It says `shown/total`, `one
   hop`, and `hidden`, retains graph/projection health, and renders boundary continuations for omitted
   neighbors.
4. **Restore context exactly.** Leaving focus restores the full-graph selection and viewport. Moving
   through a boundary continuation recenters a newly derived one-hop focus instead of silently
   expanding to an arbitrary depth.
5. **Share graph selection, not terminal geometry.** The hop-bounded selector and scope metadata are
   adapter-neutral derivations over `ThreadGraphProjection`. CLI/PR export should establish that seam
   first; TUI focus should consume it. Responsive coordinates, aliases, key handling, and viewport
   history remain in the TUI adapter. A web adapter may later consume the same full or bounded facts.

## Interaction contract

| Input | Full spatial view | One-hop focus view |
| --- | --- | --- |
| `h` / `l` | Follow a direct prerequisite/dependent. One target moves immediately; multiple targets open a compact neighbor chooser; no target reports an explicit dead end. | Same semantic action. A boundary target recenters the one-hop selection around that task. |
| `j` / `k` | Move geometrically among nodes in the current layout layer; never pretend this follows an edge. | Move among visible siblings in the focused layout. |
| `f` | Search every supplied Thread node and recenter the full viewport. | Search every supplied node, then recenter a new one-hop focus around it. |
| `z` | Proposed discoverable toggle into one-hop focus; no effect on planning data. | Return to the remembered full graph. |
| `Enter` | Open the selected task through stable entity navigation. | Same. |
| `Esc` | Return to waves under the current view cycle. | Leave focus first; a second `Esc` returns to waves. |
| `ctrl+o` | Shared entity history restores the exact spatial mode, selection, and viewport. | Same, including focus scope. |
| `y` | Copy the selected task slug. | Same. |

The branch chooser is part of the semantic `h/l` action, not a new navigation meaning: it appears
only when the graph supplies more than one valid target and must name direction, candidates, status,
and count.

## Text wireframes

### Full graph: operational entry plus persistent extent

```text
spatial · full graph · 15 nodes / 16 edges · focus [M6] in-flight
viewport 4/15 · layers 3–4/6 · rows 1–2/5 · ◂ ▸ ▼
position  1──2──[3]──[4]──5──6          z focus · f find

        [M4] prerequisite ─────▶ ┌────────────────────────┐ ─────▶ [M9]
                                 │ [M6] ▶ explore large…  │
        [M5] prerequisite ─────▶ │ in-progress · member   │ ─────▶ [M10]
                                 └────────────────────────┘

focus [M6]  explore-progressive-disclosure-and-navigation-for-large-thread-graphs
state in-progress/clear · 2 prerequisites · 2 dependents
```

The position rail reports layout extent only; it does not invent aggregated topology. The selected
layer and visible range are distinct styles, with a text fallback in the viewport line.

### One-hop focus: local cause/effect with honest boundaries

```text
spatial · focus 1 hop · 5/15 nodes · 10 hidden · graph/projection healthy
full position layer 3/6 · z full graph · f find

                                      ╭─ one-hop focus ───────────────╮
                                      │ 5/15 shown · 10 hidden        │
                                      │ z/Esc restore full viewport   │
                                      ╰───────────────────────────────╯

  prerequisites (2)            selected                 dependents (2)
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│ [M4] responsive │ ─────▶ │ [M6] explore…   │ ─────▶ │ [M9] bounded…  │ ───▸ +3
└──────────────────┘       │ in-progress     │       └──────────────────┘
┌──────────────────┐       └──────────────────┘       ┌──────────────────┐
│ [M5] routes      │ ─────▶                            │ [M10] TUI focus │
└──────────────────┘                                   └──────────────────┘

boundary  +3 beyond [M9] · l choose dependent · Esc/z restore full viewport
```

The focused topology uses the main graph canvas. The small bordered scope card overlays an
otherwise empty canvas corner and may yield when it would cover a node or route. Putting the graph
itself in a modal was rejected: a one-hop fan can exceed ten nodes, turning a small modal into a
second cramped viewport with competing navigation and obscured context.

### Branch choice for semantic edge navigation

```text
dependents from [M6] · 1/2
▶ [M9]  export bounded Thread neighborhoods      next-up · clear
  [M10] add one-hop focus to the TUI              next-up · clear
j/k choose · l/enter follow · esc stay on [M6]
```

## Selected delivery sequence

1. Amend `export-bounded-thread-neighborhoods-around-a-task` to establish the shared pure selector,
   explicit scope metadata, boundary evidence, and one/two-hop CLI/JSON/Mermaid/DOT behavior.
2. Make `add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph` depend on that shared seam, then
   add the focus lens, operational entry anchor, exact return, and branch/dead-end interaction.
3. Keep `reassess-tui-navigation-and-information-architecture-at-current-scale` and the reusable
   Back/action-target tasks coordinated but independent; they own shell-wide navigation rather than
   graph selection.
4. Treat a web renderer as an extension over the same projection/scope contracts after TUI
   dogfooding, not as a blocker or escape hatch.

## Maintainer decision (2026-09-12)

The maintainer selected strategy B with these interaction decisions:

1. The full spatial graph remains the default and flagship. First entry centers in-flight work,
   then an eligible frontier member, rather than the earliest stable node.
2. In immersive spatial mode, `z` means full ↔ one-hop focus while retaining its shell-level
   pane-zoom meaning elsewhere. Focus uses the main canvas plus the bounded scope overlay above;
   the topology itself does not live in a modal.
3. Branching `h/l` opens the compact neighbor chooser above instead of silently applying a
   deterministic tie-break. This is an intentionally testable first interaction, not a permanent
   commitment if dogfooding finds the chooser too interruptive.
