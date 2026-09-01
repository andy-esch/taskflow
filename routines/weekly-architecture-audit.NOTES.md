# Weekly Architecture Audit — Maintainer Notes

Human-facing reference for the `weekly-architecture-audit` routine. None of this
is read or executed by the agent during a run — it lives here so the operational
spec (`weekly-architecture-audit.md`) stays lean in agent context.

## Bootloader (set this in claude.ai/code/routines)

> You are running the weekly-architecture-audit routine. Your full instructions
> live at `taskflow/routines/weekly-architecture-audit.md` on `main`. Read that
> file first, then execute the workflow in its "Prompt" section. The schedule,
> lens rotation, lens scopes, ADR reconciliation rules, audit-file template, and
> Slack format are all canonical there — this routine config is only a thin
> bootloader and may be stale relative to the markdown file.

That's the entire bootloader.

## Setup checklist (one-time)

- **Repos:** `andy-esch/taskflow` only, with write access.
- **Sandbox image:** Go toolchain (Step 0 builds the tool).
- **Network:** Trusted, plus WebSearch/WebFetch — research is **mandatory** on
  this routine, unlike the conditional research in `code-quality-audit`. If web
  access is unavailable the run should stop rather than produce an unanchored
  audit.
- **Slack:** `#planning-updates`.

## Operational notes

- **Schedule**: `30 11 * * 1` UTC = 7:30am EDT Mondays (6:30am EST after Nov
  DST). Same slot desirelines uses for its architecture audit, and it keeps this
  routine on its own day.
- **Model**: Opus-class. Research synthesis plus cross-layer reasoning at weekly
  cadence justifies the cost.
- **Cadence**: ~one-month full coverage (4 lenses). If a lens audits clean for
  2–3 cycles running, consider collapsing it into `code-quality-audit` or
  stretching the rotation to make room for a new lens.
- **Tuning knobs**: LENS-SCOPES (step 3), research depth (3–6 queries, step 6),
  files-read budget (6–10, step 4), and the rotation mapping (step 2b).

## Why the ADR reconciliation section exists

This is the biggest departure from the desirelines version, and the reason the
routine earns its slot.

taskflow keeps accepted ADRs in `planning/adrs/` that state the intended
architecture in enforceable detail — ADR-0003 on flat id-led storage, ADR-0006
on Threads over a global task DAG, and so on. Most codebases have no such
authority, so an architecture audit there can only compare code against generic
best practice. Here, "does the code match the decision we already wrote down?"
is both cheaper to answer and more actionable than "does the code match a blog
post?".

The three-way finding **Class** (`implementation drift` / `ADR gap` /
`ADR pressure`) is what keeps that honest. Without it, an agent given an ADR
tends to do one of two bad things: treat the ADR as unquestionable and suppress
real evidence, or quietly "correct" the ADR. Forcing every finding to declare
which of the three it is makes the disagreement explicit and routes it to a
human.

`unanchored` is the escape hatch for genuinely new territory the ADRs don't
cover — it should be rare, and a run producing mostly `unanchored` findings is a
signal the ADR set has fallen behind the code.

## The ADR is never edited

Stated three times in the spec, on purpose. ADR-0001 establishes that decisions
change through an `## Amendments` entry or a superseding ADR — both human acts.
An agent with write access to `planning/adrs/` and a well-argued case is exactly
the failure mode to design against, because the resulting edit would look
plausible and would silently move the project's stated architecture.

The `**Proposed ADR amendment:**` field in the finding template is the pressure-
release valve: the agent gets to write the clause it would propose, and the
human gets to accept, reject, or ignore it.

## Relation to code-quality-audit

`code-quality-audit` rotates twice weekly through code-level lenses; this
rotates weekly through architectural concerns across layers (research mandatory,
ADR-reconciling, narrative + findings hybrid).

The discriminator in the spec is **compounding cost**: an architecture finding
must name what it costs to fix later rather than now. A finding with no
compounding cost is a code finding and gets routed to the other routine. If the
two keep colliding anyway, narrow this routine's lens scopes rather than the
code audit's.

## Open design questions

Revisit after ~4 cycles (each lens has run once):

- **Lens taxonomy** — are 4 right? Candidates not yet included: *extensibility*
  (how much does adding a new entity kind or a new adapter cost — arguably the
  most load-bearing question for a tool whose roadmap is "more nouns"),
  *distribution and versioning* (release, upgrade path for existing user repos,
  backward compatibility of on-disk state), and *multi-space topology* (epic 29).
  Add via rotation expansion or by replacing a consistently-clean lens.
- **ISO-week mod 4 vs. a counter file** — deterministic but can hit a holiday
  week. If skipping happens, switch to a counter in
  `planning/audits/architecture-rotation-state.json`.
- **Citation freshness** — the 12-month floor is a guess. Go and hexagonal-
  architecture writing ages slower than cloud-vendor docs, so this may be too
  tight; loosen if it forces the agent to cite thin blog posts over a good older
  source.
- **Narrative length** — measure average audit length vs. actions-per-page after
  4 cycles. The "State of the architecture" section is the most likely place for
  the agent to write a lot and say little.
- **ADR coverage as a metric** — after a full cycle, count how many findings
  came back `unanchored`. That number is a direct measurement of how far the ADR
  set trails the implementation, and might deserve its own line in the PR body.

## Changelog

- **v1 (2026-08-30)** — initial spec, adapted from
  `desirelines-planning/routines/weekly-architecture-audit.md` v2. Changes for
  taskflow: single-repo (no install script, no deploy repo); **ADR
  reconciliation section and the three-way finding Class are new**, as is the
  explicit "never edit an ADR" rule and the `**Proposed ADR amendment:**` field;
  lenses rewritten for a hexagonal Go CLI (hexagonal-boundaries /
  data-model-and-storage / machine-contract / failure-and-recovery) in place of
  the distributed-systems lenses (data-flow / observability / service-boundaries
  / failure-modes); backpressure check added, matching the code audit;
  `audit finding --status "tracked by <task-id>"` replaces the hand-written
  `superseded by` convention; **compounding cost** added as the discriminator against the code
  audit; and the research framing now names the single-user local-first
  constraint explicitly so the agent stops scoring the project on a scale it
  isn't on. The SCOPE block reduces to one rule — write to `planning/` only —
  with a `git status --porcelain` bounds check before the PR.
