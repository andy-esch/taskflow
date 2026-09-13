---
schema: 1
id: 6g9g37bvetjy
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Re-record the README Threads demo after navigation design settles so it teaches current dependency-rank and one-hop focus behavior.
effort: 1 day
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, threads, docs, demo]
created: "2026-09-12"
depends_on: [6g87qn72901g]
updated_at: "2026-09-12"
---
# Refresh the Threads demo after the TUI navigation redesign

## Objective

Re-record the README Threads demo after the current TUI navigation and information-architecture
work settles so the committed artifact teaches the shipped dependency-rank and spatial-focus
vocabulary instead of preserving obsolete wave language.

## Acceptance criteria

- [ ] The Threads fixture still tells a realistic touring-bike delivery story and exercises enough
  dependencies, lifecycle states, external prerequisites, and graph navigation to be credible.
- [ ] `assets/threads.gif` shows the current Thread identity treatment, dependency-rank terminology,
  spatial graph, and one-hop focus/return path without stale wave or barrier language.
- [ ] The tape is deterministic, uses the normal repository theme/background conventions, and the
  documented GIF-generation command reproduces the checked-in asset.
- [ ] README prose and the recorded screen agree; visual output is reviewed at rendered size before
  closeout.

## Sequencing

Run after the whole-TUI navigation review and its selected discoverability slice so the expensive
recording work captures the intended interaction hierarchy once.

## Related

- Follow-up to finding L4 in the
  [Claude review of the one-hop Thread focus implementation](../audits/6g9fg4a8xvyc-2026-09-12-tui-one-hop-thread-focus-implementation-claude.md).
- Builds on the touring-bike Thread preview fixture and existing VHS workflow.

## Out of scope

- Redesigning TUI navigation or changing graph semantics while recording the demo.
