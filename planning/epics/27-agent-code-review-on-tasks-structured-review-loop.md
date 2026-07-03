---
schema: 1
status: active
description: 'A reviewer agent attaches a structured critical code review to a task via tskflwctl (never hand-authored); the implementing agent reads it back as a compact open-findings digest and resolves it finding-by-finding, with diff-scoped re-review and multiple review lenses. Design-first; verdict-combination, review-state, and finding identity across passes/lenses are the open pieces.'
priority: medium
tags: [tasks, review, agents, workflow, findings]
created: "2026-07-03"
---
# Agent code review on tasks — a structured review loop

**Status: design-first. NOT ready to implement.** This epic is here to *shape* the
feature before any code is written. Several forks are now decided (see *Decisions so
far*); the first task is to close the rest, not to build a reviewer.

## Why this exists

Agent implementation is iterative: a base agent does a first pass, and we want a
*separate* reviewer agent to attach a critical code review to the task, so the
base agent can re-read the task **with** the review and act on it — then a second
pass, another review, and so on. Today a task file has no structured place for
that: review notes would land as freeform prose with no lifecycle, no "is this
addressed?" signal, and nothing lint or a board could see.

The goal is a review that is **machine-parseable, human-readable, and actionable**,
that tracks multiple passes, and that closes the loop (open → addressed →
re-review) the same disciplined way the rest of the system tracks work.

## Guiding principles (settled — these constrain the open questions, not up for grabs)

- **Review findings are always created and mutated through `tskflwctl` — never
  hand-authored.** The tool owns the finding grammar: a reviewer agent calls a
  command with *structured inputs* (severity, file, title, body, suggested fix) and
  the tool emits well-formed, parseable, lint-clean markdown; status transitions
  (`open → fixed`/`wontfix`) and the pass verdict go through the tool too. Freeform
  typing into the file is not a supported path for reviews. This is the same
  "structured mutation over freeform editing" stance the rest of taskflow already
  takes — scriptable, atomic, re-validated field/finding writes for agents; the
  whole-file `$EDITOR` only for humans — and it is the guarantee that a review is
  machine-*actionable*, not prose that merely looks structured. It also means
  validation happens at **write time**, not only at lint time: a malformed finding
  is rejected when the reviewer tries to add it.

- **The tool guarantees structure, not policy.** `tskflwctl` enforces that every
  finding is *well-formed* — valid severity, a title, a legal status, a stable
  tool-assigned code — and nothing more. It does **not** decide policy: it doesn't
  dedup or converge findings, doesn't combine lenses into a verdict, and doesn't know
  what a reviewer looked at (a diff, one file, the whole implementation). Those are
  judgments for the writing/reading agents and the prompt that drives them. Keeping
  the tool structural — not semantic — is what lets it stay small and unopinionated
  while still guaranteeing machine-actionable output.

This principle bears directly on the shape fork below. It raises the bar for
**shape A** (a `## Review` section): that section cannot lean on the freeform
`task append` body path — it needs *structured* writers (add-finding /
set-finding-status / set-verdict) that construct the grammar and re-validate,
analogous to the `audit` verbs. **Shape B** (a linked audit) inherits tool-mediated
finding writes essentially for free, since audits are already mutated that way. Weigh
this in the A/B decision.

## Decisions so far (2026-07-03)

Settled with the maintainer — the open questions below are what's left.

- **Tool-mediated, always** — reviews are written *and* read through `tskflwctl`,
  never hand-authored (see *Guiding principles*). This is what lets the storage shape
  stay an internal detail.
- **Read = a compact, open-only digest, but multi-mode.** The default read
  (`task review show <slug> --open`) returns only unresolved findings in a terse form
  (code · file:line · severity · one-line fix) plus a tally/verdict, to minimize the
  implementing agent's tokens. It is *not* the only mode: a `--json` field projection
  (machine parsing) and a one-finding-at-a-time `review next` are also exposed. Which
  to use is **situational**, left to the caller — the tool offers all three rather
  than forcing one.
- **Loop = per-finding resolve; re-review is just more findings.** The implementing
  agent resolves each finding through the tool (`review resolve <slug> <code> --note …`),
  so state stays structured and nothing is deleted (resolved findings are history).
  Re-review is simply a reviewer adding more findings later — the tool tracks no
  "basis"/diff and no formal "pass" object; *what* a reviewer examines (a diff, one
  file, the whole impl) is prompt-driven and invisible to the tool (see *structure,
  not policy*). A lens label + timestamp on each finding is enough to tell rounds and
  lenses apart without a pass concept.
- **v1 scope = code review, multiple lenses.** From the start, more than one reviewer
  can attach findings to a task (e.g. a correctness lens and a security lens), each
  finding tagged with its lens. The tool does **not** merge/dedup across lenses or
  compute a combined verdict — convergence is the writing/reading agent's call (see
  *structure, not policy*). Design/spec (pre-impl) review is out of v1, but the
  substrate should not preclude it.
- **Storage shape deferred (on purpose).** Because read+write both go through the
  tool, the implementing and reviewing agents never see where the review lives; pick
  section-vs-linked-entity at implementation time for maximum code reuse (leans toward
  reusing the audit finding subsystem). See *Candidate shapes* for the two options.

## The primitive to reuse (don't reinvent)

taskflow already has a mature "structured review with a lifecycle": **audit
findings**. The grammar — `#### H1. Title · **Status:** open` with
`**File:**`/`**Component:**`/`**Effort:**`/`**Urgency:**` fields, statuses
`open | in-progress | fixed | landed | deferred | superseded | wontfix` — is parsed
by `domain.ParseFindings` / `domain.TallyFindings`. A code review is that same
shape, scoped to one task instead of an area. Whatever shape we pick, review items
should reuse this grammar so the parse/tally machinery, the CLI mental model, and
the finding lifecycle all carry over unchanged.

## Candidate shapes (storage — DEFERRED; the commands hide this)

Deferred by decision (see *Decisions so far*): agents only ever call the review
commands, so this is an implementation-time choice optimized for code reuse. Kept
here as reference for when it's picked.

**A. A scoped `## Review` section in the task body.** Each pass appends a
`### Review N — <date> · <reviewer> · <basis>` block with a `**Verdict:**` and
findings in the grammar above; the base agent flips each finding's status as it
addresses it. Leans to "the base agent reads one file," at the cost of growing the
task grammar and needing the finding parse **scoped to the review section** (a task
body can contain `####` headers that look like findings).

**B. A linked audit entity.** A review is an audit with a `target_task: <slug>`
pointer; reuse the *entire* audit subsystem verbatim (CLI, lint, lifecycle, its own
id + history). Cleaner reuse and separation, at the cost of two files to read and
new task↔review linking/discovery.

Illustrative sketch (shape A), for concreteness only — NOT a committed format:

```markdown
## Review

### Review 1 — 2026-07-03 · reviewer: code-reviewer-agent · basis: pass 1 (HEAD abc123)
**Verdict:** changes-requested

#### H1. dropped error on the fallback write · **Status:** open
**File:** store/fix.go:172
Ignores the write error on the filename-date path. **Suggested fix:** propagate it
like the created-date branch.
```

## The loop (mechanics to pin down)

1. A reviewer agent attaches a review (pass N) after an implementation pass.
2. A cheap **signal** tells the base agent (and a board, and lint) that attention is
   needed without parsing the body — candidate: a `review_state` frontmatter field
   (`none | in-review | changes-requested | approved`), with counts *derived* from
   the section (single source of truth, as audits don't cache tallies). The verdict
   isn't purely derivable (an "approved, zero findings" is real signal), which is
   the argument for a field over pure derivation — but see open questions.
3. The base agent fixes each finding, flips its status (`fixed`/`wontfix` + a note
   cross-linking the Progress Log), and requests re-review.
4. A **lint rule** keeps it honest: a task in `completed` (or `in-progress`) with an
   unresolved review — `review_state: changes-requested` or any `open` review
   finding — is flagged, mirroring the "no dangling findings" discipline audits get.
5. Re-review appends pass N+1; prior passes stay as history.

Review is **orthogonal to lifecycle status** (a task can be `in-progress` *and*
under review), so it's a field + section, not a new status.

## Open questions (still to shape)

Some carry a **suggested direction** — a proposal to react to, not a decision.

- **`review_state` rollup — deliberately unspecified.** Whether a task carries a
  rollup review-status field at all, its values, who moves it, and whether it *routes*
  the hand-off — left open pending more shape. Note: with no tool-side verdict
  combination (see Decisions), a rollup would itself be *policy*, so this may stay thin
  or absent.

- **Verdict vocabulary — *suggested direction*.** Keep it minimal and **per-lens**, or
  drop it. Suggestion: an optional per-lens verdict of just `approved |
  changes-requested`. Nuance like "nits" lives in finding *severity*, not the verdict,
  and there is no tool-computed task-level rollup (that's the reader's call). A verdict
  still earns its place by distinguishing "reviewed, looks fine" (approved, zero
  findings) from "not reviewed yet" (no findings from that lens).

- **Finding identity — *suggested direction*.** Have the **tool assign** a stable,
  monotonic per-task finding code at write time (same "the tool owns identity" stance
  as stable ids), with **severity and lens as separate fields**, not encoded in the
  code. This sidesteps cross-lens collisions (two lenses can't both mint `H1`), gives
  `review resolve <slug> <code>` an unambiguous target, and means an unresolved finding
  from round 1 is *referenced* by its code in round 2, never re-listed. Open sub-choice:
  audit-style severity-lettered but tool-numbered (`H1/H2/M1…`) vs plain `F1/F2…` with
  severity purely a field.

- **CLI verb shape & typed inputs — *suggested direction*.** A minimal `task review`
  sub-namespace:
  - `add <slug> --severity high|med|low --title … [--file path:line] [--body|--body-file …] [--fix …] [--lens <name>]` → tool assigns + prints the code
  - `show <slug>` (default: compact, open-only) `[--all] [--json] [--severity …] [--lens …]`
  - `next <slug>` (one open finding at a time)
  - `resolve <slug> <code> [--note …] [--wontfix]` · `reopen <slug> <code>`
  - optional `verdict <slug> --lens <name> approved|changes-requested`

  Note: audits currently *hand-author* findings in prose, so these structured writers
  are new ground — and could later be retrofitted onto audits for consistency.

- **Deferred with storage (tbd):** if the section shape wins, how to bound finding
  parsing to the review section without a markdown AST — revisit when storage is chosen.

## Non-goals (for now)

- Building the reviewer agent itself / prescribing a review rubric — this epic is the
  *substrate* (where a review lives, how it's tracked), not the reviewer.
- A general task-body AST or markdown parser.

## Related

- Epic 26 (frontmatter schema) — `review_state` and the review-section grammar are
  exactly what a declared schema should define; coordinate the field/section shape
  there.
- Finding grammar — `domain.ParseFindings` / `TallyFindings` and the audit subsystem
  (`audit findings/close/reopen`, `LintAudit`) are the reuse target and the shape-B
  candidate.
- Epic 20 (CLI UX) — the `task review` verbs and any board "needs-review" signal.
