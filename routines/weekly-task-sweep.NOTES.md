# Weekly Task Sweep — Maintainer Notes

Human-facing reference for the `weekly-task-sweep` routine. None of this is read
or executed by the agent during a run — it lives here so the operational spec
(`weekly-task-sweep.md`) stays lean in agent context. Edit this when you set up,
reconfigure, or rethink the routine.

## Bootloader (set this in claude.ai/code/routines)

The routine in claude.ai should carry this minimal prompt:

> You are running the weekly-task-sweep routine. Your full instructions live at
> `taskflow/routines/weekly-task-sweep.md` on `main`. Read that file first, then
> execute the workflow in its "Prompt" section. The schedule, scoring, safe-
> editing rules, PR template, and Slack format are all canonical there — this
> routine config is only a thin bootloader and may be stale relative to the
> markdown file.

That's the entire bootloader. Everything else lives in the spec file.

## Setup checklist (one-time)

- **Repos:** `andy-esch/taskflow` only, with write access. Unlike the
  desirelines routines there is no second repo and no read-only companion —
  taskflow self-hosts its planning.
- **Sandbox image:** must ship a **Go toolchain**. Step 0 is `just build`; `just`
  is nice-to-have (the spec falls back to plain `go build`), Go is mandatory.
  There is no prebuilt-binary fallback and deliberately so: the routine audits
  this repo, so it should run the binary this repo's `main` produces.
- **No `GH_TOKEN` needed.** The tool is built from the checkout in front of it.
- **Network:** default Trusted is enough (Go module proxy for deps). This
  routine does no web research.
- **Slack:** `#planning-updates`. Confirm the channel exists and the connector
  is authorised before the first run — this was copied from the desirelines
  setup and may want its own taskflow channel.

## Operational notes

- **Schedule**: `0 10 * * 0` UTC = 6am EDT Sundays (5am EST after Nov DST).
  Sunday keeps it clear of the Monday architecture audit and the Tue/Fri code
  audits, so no two routines share a usage window.
- **Cadence choice**: weekly, not nightly. desirelines runs this nightly across
  a larger, faster-moving backlog. taskflow's planning tree is smaller and its
  code churn is bursty (a slice lands, then the tree is quiet), so a nightly
  sweep would spend most runs re-confirming a frozen HEAD — exactly the loop the
  cooldown gate exists to break. Revisit if the backlog grows past ~60 active
  tasks.
- **Tuning knobs**: `sweep_size` (5), `cooldown_days` (7), the scoring weights
  in step 4b, and the churn window (90 days) / adjacency window (30 days). Edit
  the spec → PR → merge → next Sunday picks it up. Schedule, repos, and
  connector changes require updating
  [claude.ai/code/routines](https://claude.ai/code/routines).

## The dependency-graph bonus (step 4b/4c)

This is the one scoring axis that has no desirelines counterpart, and it exists
because of ADR-0006. A task's *derived* state — `role`, `gate`, `eligible` — is
computed from the repository DAG, not stored, so it can change without the task
file changing at all: completing a prerequisite silently makes a downstream task
eligible, and deprecating one silently makes its gate `broken` rather than
`blocked`.

That means a task can go stale in a way `git log` cannot see. `task blockers
<slug> --json` is the only way to notice, which is why step 4c runs it per
candidate and why the drift bonus is weighted equal to code churn (+2).

If this proves noisy — e.g. one big `task complete` flips a dozen gates in one
week and floods the sweep — consider capping the bonus to tasks whose *prose*
disagrees with the graph, rather than any flip.

## Why completion is never automatic

The spec forbids `task complete` even when the agent is confident. Two reasons:

1. `complete` refuses a task with an unmet, unexplained acceptance criterion —
   so an over-eager agent's next move is `--force`, which is exactly the wrong
   habit to build into a scheduled job.
2. Completion stamps `completed_at` and changes what downstream tasks' gates
   report. It is a graph mutation with blast radius, not a bookkeeping edit.

`deprecate` and `defer` are excluded for the same reason.

## Open design questions

Revisit after ~6 runs:

- **Sweep size** — is 5 right for a tree this size, or should the routine sweep
  until it exhausts eligible tasks (bounded at, say, 10)?
- **`audited:` adoption** — most taskflow tasks have never carried `audited:`.
  Until the field is populated everywhere, nearly every task scores the +9
  staleness cap and the ordering degrades to "oldest `created:` first". That's
  acceptable for the first few runs (it *is* the right order for a cold start)
  but re-check once the field is broadly present.
- **Thread awareness** — should the sweep prefer tasks that are members of an
  in-progress Thread? Deferred until guarded Thread membership mutations land;
  reading `thread frontier` is already cheap, so this is a small addition when
  the time comes.
- **Body edits vs annotations only** — the current rules permit conservative
  body edits. If PRs start being hard to review, tighten to "annotations only,
  never rewrite a paragraph".

## Changelog

- **v1 (2026-08-30)** — initial spec, adapted from
  `desirelines-planning/routines/nightly-task-sweep.md` v2. Changes for
  taskflow: single-repo (no install script, no `pm` history, no cross-repo
  working dirs); weekly instead of nightly; a SCOPE block reducing to one rule —
  write to `planning/` only — with a `git status --porcelain` bounds check
  before the PR, because the agent can otherwise write the code it audits;
  `task ac` ownership of acceptance criteria; `depends_on` declared untouchable;
  and the new dependency-graph drift scoring axis (step 4c).
