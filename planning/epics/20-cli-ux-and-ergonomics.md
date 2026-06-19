---
status: planning
description: 'Post-port CLI UX and pipeline ergonomics: interactive pickers, output modes, task edit, column projection'
priority: medium
tags: [cli, ux, ergonomics]
created: "2026-06-19"
---

# CLI UX and ergonomics

**Goal.** Post-port CLI UX and pipeline ergonomics: interactive pickers, output modes, task edit, column projection

## Why this is its own epic

Epic 17 is the Python→Go **port** (parity with the `pm` prototype) — nearly done
and best kept closeable as that record. These tasks are **post-port
enhancements** to how the tool *feels* to drive: a human interactive layer (gh-
style pickers), Unix-pipeline output modes (`-q`/`--plain`/column projection),
and editing a body through the tool. A distinct, ongoing theme — driven by
dogfooding, not the port checklist — so it gets its own home rather than
inflating 17. Governing rule across all of it: **never compromise the
agent/pipeline contract** — interactivity and human niceties are TTY-gated
recovery faces, never reachable by an agent or a pipe.

## Tasks

- `interactive-prompt-layer-gh-style-pickers` — huh TTY pickers for missing input.
- `pipeline-output-modes-q-plain-stderr-discipline` — `-q`/`--plain`/stderr sweep.
- `consolidate-output-flags-into-output-and-columns` — one `-o/--output`
  format flag + a completable `-c/--columns` projection (supersedes the old
  `column-projection` task, now deprecated).
- `task-edit-opens-editor-on-the-body` — `$EDITOR` on a task body (human face).
- `agent-facing-cli-ergonomics-batch` — the agent-side DX batch (body replace/
  append remains).

## Out of scope

- The readiness axis ([[task-readiness-state-draft-vs-finalized-in-frontmatter]])
  — a planning-model change, deferred and tracked separately.
- The web companion (epic 19) and anything that alters the core domain model.
