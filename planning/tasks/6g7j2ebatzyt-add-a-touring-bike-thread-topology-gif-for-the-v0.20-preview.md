---
schema: 1
id: 6g7j2ebatzyt
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Create a realistic, graph-first touring-bike Threads demo and secondary TUI GIF for the v0.20 preview.
effort: 1 day
tier: 2
priority: medium
autonomy_level: 3
tags: [threads, docs, demo, tui]
created: "2026-09-06"
started_at: "2026-09-06"
updated_at: "2026-09-06"
completed_at: "2026-09-06"
---
# Add a touring-bike Thread topology GIF for the v0.20 preview

## Objective

Give the v0.20 Threads preview one focused, attractive demonstration of graph comprehension. Use a
realistic touring-bike release story to make member waves, fan-out/fan-in, external gates, task
state, and stable navigation understandable without reading the implementation documentation.

## User-approved creative brief

- Story: prepare a touring bike for release and departure.
- Primary message: understand the dependency graph, not merely browse a bag of tasks.
- Placement: a secondary focused GIF; do not replace the broad TUI hero.
- Match the existing bike-workshop voice, Dracula palette, typography, terminal dimensions, terse
  copy, and deliberate VHS pacing. Avoid generic `alpha`/`beta` fixtures or contrived dependencies.

## Scope

- Create a dedicated, self-contained `assets/demo-threads/` planning fixture so the topology can be
  designed honestly without destabilizing the status and Atlas demos.
- Model five to seven concrete Thread members with meaningful fan-out/fan-in, at least three waves,
  plus one immediate external gate and a coherent mix of completed, in-flight, eligible, and blocked
  work. Statuses and dependencies must be semantically valid, not selected only for color.
- Include one in-progress Thread with concise, domain-specific goal, description, and body copy.
- Add a focused VHS tape that opens the Threads tab, focuses the Thread detail, enters the `v`
  topology view, navigates nodes with `j`/`k`, opens one task with Enter, and returns with `ctrl+o`.
- Keep setup hidden, pauses long enough to read but not drag, and the graph legible without wrapping
  at the established demo dimensions.
- Repair the existing demo fixture's retired audit-finding vocabulary so `just gifs` regenerates
  from lint-clean inputs, superseding the older generic refresh task.
- After the fixture and first draft render exist, stop and present the graph, representative frame,
  timing, and file size for user feedback. Do not finalize README placement or the gallery-wide
  regeneration until that feedback is incorporated.
- After approval, commit `assets/threads.gif`, link it as a secondary demo from the root and asset
  READMEs, and keep `just gifs` as the single regeneration entry point.

## Acceptance criteria

- [x] The dedicated fixture is deterministic and lint-clean, with a realistic seven-member
      touring-release graph plus one immediate external gate, covering fan-out/fan-in, three or
      more waves, and coherent mixed task states.
- [x] The Thread detail and topology frames use concise bike-workshop language and remain legible at
      the existing Dracula/1200×760/16px visual baseline.
- [x] The draft tape demonstrates Threads tab → detail → topology → `j`/`k` selection → Enter task →
      `ctrl+o` return without exposing fixture setup or unrelated UI wandering.
- [x] The implementation pauses after the first draft and records user feedback before finalizing
      the recording and README placement.
- [x] `assets/threads.gif` is presented as a secondary focused demo in both README surfaces, and the
      asset/VHS documentation explains its dedicated fixture.
- [x] The existing demo-planning fixture uses valid current finding vocabulary; `just gifs`, fixture
      lint, link checks, and `git diff --check` pass after the final regeneration.
- [x] The v0.20 release task records the approved visual as release evidence without making the
      optional spatial graph experiment a release gate.

## Draft review and final artifact

The first draft used the approved touring-bike story and made its three waves, fan-out/fan-in,
mixed work states, and external crown-mount gate readable at the established visual baseline. User
review accepted the scenario and the truncated list slug, requested the missing palette, and agreed
that navigation should finish back on topology after opening a task. The monochrome draft exposed
an inherited `NO_COLOR=1`; the shared `just gifs` recorder now clears that preference for every tape
rather than embedding a one-off environment override in this tape.

The final 1200×760 recording is 19.80 seconds and about 917 KiB. It navigates Threads → detail →
topology, selects the eligible wheel task, opens it by stable identity, returns with `ctrl+o`, and
restores topology before exit. The approved artifact is
[`assets/threads.gif`](../../assets/threads.gif); its tape and deterministic fixture live beside the
other demo sources.

Final gallery review also caught that the isolated Atlas registry still staged only the two older
planning identities. Its setup now registers the touring-release fixture too, so the Atlas visibly
contains three identities while preserving the bike-workshop card's direct and pointer entry points.

## Out of scope

- A second CLI GIF, replacement of the broad TUI hero, or a redesign of the overall demo gallery.
- Implementing or implying the future two-dimensional spatial graph view.
- Changing core graph semantics or adding TUI behavior solely to make the recording easier.

## Related

- Release checkpoint `cut-v0.20.0-as-a-compatibility-hardened-threads-preview`
- Thread `complete-production-threads`
- Supersedes `re-record-demo-gifs-after-the-epic-active-vocab-migration`
