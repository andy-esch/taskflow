---
schema: 1
id: 6g31g9f8x4cv
status: completed
epic: 20-cli-ux-and-ergonomics
description: Criteria are a binary checkbox while findings carry a seven-state vocabulary; an unchecked box cannot distinguish not-yet from won't-do from deferred.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, domain, planning-model]
created: "2026-08-23"
updated_at: "2026-08-24"
started_at: "2026-08-24"
completed_at: "2026-08-24"
---
# Let an acceptance criterion say more than done or not-done

## Objective

An acceptance criterion is a checkbox. `domain.Criterion` carries a bool, `ACCount` is a
checked/total tally, and `task ac` can only `--check` or `--uncheck`. So an unchecked box
means *not yet*, *won't do*, *deferred*, *blocked*, *superseded by a later decision*, and
*turned out not to apply* — all at once, indistinguishably.

Findings in this same repo already solved this. They carry
`open · in-progress · fixed · landed · deferred · wontfix · superseded`, a validated
vocabulary that `lint` enforces and `audit findings` can query. Criteria have none of it.

**Demonstrated 2026-08-23.** Completing
[compose-the-atlas-spaces-view-as-a-comparable-table](6g2zqyra2s6h-compose-the-atlas-spaces-view-as-a-comparable-table.md)
left criterion 8 ("degrades on narrow terminals by dropping the least load-bearing
columns") unchecked on purpose: it is genuinely unmet, deliberately deferred, and tracked
as finding L1 of
[2026-08-23-atlas-ui-restructure](../audits/6g30mbhw1xt3-2026-08-23-atlas-ui-restructure.md).
None of that survives in the file. A reader — human or agent — sees a completed task with
an unchecked box and cannot tell a deliberate deferral from an oversight. The reasoning
lives only in an audit note nothing links from the task.

This matters more than cosmetics because a completed task with unchecked criteria is
exactly the shape `task complete` should be able to reason about (see finding M5 of
[2026-07-24-ai-agent-cli-ergonomics](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md),
"`task complete` does not reconcile unfinished acceptance criteria") — and it cannot,
because the data model has nothing to reconcile *against*.

## Open questions to settle first

Design-first; do not start implementing until these are answered.

1. **How much vocabulary?** Borrowing all seven finding statuses is probably wrong —
   `landed` and `fixed` are audit-shaped. A criterion plausibly needs *met*, *not met*,
   *deferred*, *won't do*, *no longer applies*. Fewer states that each mean something
   beats parity with findings.
2. **Where does it live?** Markdown checkboxes are the current, human-editable, diff-able
   representation, and that is a real virtue. An inline suffix (`- [ ] … · **deferred:**
   why`) keeps that; a frontmatter side-table does not. Whatever is chosen must survive a
   human hand-editing the file in an editor.
3. **Is a reason mandatory** for the non-binary states? A deferral without a why is the
   thing this task exists to prevent, so probably yes — but that makes `task ac` need a
   `--reason` and makes `lint` an enforcer.
4. **Does `task complete` gate on it?** Refusing to complete a task with unexplained unmet
   criteria is the payoff, but it is also a workflow change that could be obstructive.
   Decide deliberately rather than inheriting it.
5. **Backwards compatibility.** Every existing task uses plain checkboxes. Whatever lands
   must read them as-is, and `lint --fix` must not rewrite a corpus of them.

## Acceptance criteria

- [x] The vocabulary and its markdown representation are decided and written down before
  any code, with the rejected alternatives recorded so they are not relitigated.
- [x] `domain` models a criterion's state as that vocabulary rather than a bool, and plain
  `- [ ]` / `- [x]` continue to parse to the obvious members.
- [x] `task ac` can set any state, and requires whatever justification the design decides
  is mandatory.
- [x] `lint` validates the vocabulary the way it validates finding status, including inside
  fenced blocks being ignored exactly as today.
- [x] `ACCount` and every surface that renders a tally (`task show`, `status`, the TUI) keep
  reporting something honest when criteria are no longer two-valued.
- [x] A decision is recorded on whether `task complete` gates on unmet criteria, either way.

## Out of scope

- Finding status itself, which already has a vocabulary — its gap is the missing *write
  path*, tracked in [2026-08-17-finding-status-surface](../audits/6g1397jfke23-2026-08-17-finding-status-surface.md).
- Retrofitting states onto historical completed tasks. New vocabulary, forward-looking.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Prior art in this repo: finding status in `internal/domain/finding.go`
- Motivating case: criterion 8 of [compose-the-atlas-spaces-view-as-a-comparable-table](6g2zqyra2s6h-compose-the-atlas-spaces-view-as-a-comparable-table.md)

## Direction, 2026-08-23 — coordinate, do not invent

Confirmed the vocabulary should be **shared with audit findings**, not designed
independently. Two consequences that change the shape of this task:

- **One pool, not two lists.** Whether criteria take the finding vocabulary wholesale or a
  named subset of it, the states should live in one place in `domain` so the two cannot
  drift. A criterion marked `deferred` and a finding marked `deferred` should mean the same
  thing and render the same way. If some finding states genuinely do not apply to criteria
  (`landed` is audit-shaped), the subset should be declared as a subset of the pool rather
  than as a separate enum that happens to overlap.
- **Borrow the rendering too.** Findings already display well in the TUI — glyph, colour,
  and status grouping. Criteria should reuse that vocabulary of marks rather than a second
  visual language for the same idea, so a reader learns one thing.

**Surface AC state near the top of a task.** Today acceptance criteria are body prose the
reader scrolls to. Once a criterion can say *deferred* or *won't do*, that state is
decision-shaped and belongs where decisions are visible — a compact roll-up in the task's
detail header, in the same spirit as the audit segmented finding bar, so "3 met · 1
deferred · 1 won't do" reads at a glance without scrolling into the body.

That roll-up is also what makes the `task complete` reconciliation question answerable:
the surface that shows unmet criteria is the surface that would explain a refusal.

### Added acceptance criteria

- [x] The state vocabulary is defined once in `domain` and shared with finding status —
  either the same set or a declared subset — with a test that fails if the two drift apart.
- [x] Criterion states reuse the finding glyph/colour vocabulary rather than introducing a
  parallel one.
- [x] A task's acceptance-criteria roll-up is visible near the top of its TUI detail view,
  not only by scrolling into the body.
- [x] A decision is recorded on whether a `tracked` CRITERION must name its destination the
      way a `tracked` FINDING does. `SetFindingStatus` refuses a bare `tracked` and lint
      flags one, but `task ac --tracked <n> --reason "just because"` is accepted — the same
      word carrying a weaker guarantee on one of the two entities that share it. The
      asymmetry is defensible (an audit concludes its interest on handoff; a task's work
      merely moves) but it is currently accidental rather than stated, and a shared
      vocabulary whose guarantees differ per entity has already begun to drift. Either
      require an id-shaped token in the reason, or write down why a criterion's destination
      is softer. Raised as M3 of `2026-08-24-finding-note-and-vocabulary-selfreview`.

## Decisions, 2026-08-24 — settled before implementation

### Vocabulary: `met · not met · deferred · wontfix · n/a`

`deferred` and `wontfix` are taken VERBATIM from the finding vocabulary — same word, same
meaning, so a reader learns them once. `fixed`, `landed`, and `superseded` are audit-shaped
and excluded. `n/a` is new: "turned out not to apply" is a thing criteria need and findings
have no word for.

This is an overlap, not a subset, so it is modelled honestly as one: a shared pool holds
the words both entities use, each declares its own full set from that pool plus its own
additions, and a test asserts the shared words are spelled identically in both. Calling it
a subset would be a lie — `met` is not a finding status — and the lie is what lets the two
drift, which is exactly finding M3 of
[2026-08-17-finding-status-surface](../audits/6g1397jfke23-2026-08-17-finding-status-surface.md)
(`landed` legal in code, absent from docs, code asserting otherwise).

### Representation: checkbox plus suffix

```
- [x] Criterion that is done
- [ ] Criterion still to do
- [ ] Criterion parked · **deferred:** waiting on the schema ADR
- [ ] Criterion abandoned · **wontfix:** superseded by the table layout
- [ ] Criterion that stopped applying · **n/a:** the tile grid was dropped
```

The bracket keeps its existing binary meaning — `[x]` met, `[ ]` not met — and the suffix
refines the NOT-MET case. Three consequences, all deliberate:

- **No migration.** Every existing `- [ ]` / `- [x]` in the corpus stays valid and keeps
  parsing with the same meaning. Nothing to back-fill, nothing for `lint --fix` to rewrite,
  no legacy mode. The states are additive, so the burden the alternative designs carried
  simply does not arise.
- **It reuses a grammar that already exists.** `**Label:** value` is how findings carry
  `**Status:**` / `**Effort:**` / `**Urgency:**`, and `·` is already the separator in
  finding headers. Nothing new to learn or to parse differently.
- **The reason has a natural home**, which is what makes Q3 answerable at all.

A marker char in the bracket (`- [~]`) was rejected: the existing lint deliberately REJECTS
non-canonical brackets ("malformed acceptance checkbox … use `- [ ]` or `- [x]`"), so it
would fight a rule already in place, and it is cryptic. A frontmatter side-table was
rejected for breaking hand-editability, which is a stated virtue of the format.

### A reason is mandatory for `deferred`, `wontfix`, and `n/a`

A deferral with no why is the thing this task exists to prevent — it is indistinguishable
from an oversight, which is the original defect. `met` and `not met` take no reason.

### Contradictions are lint errors

`- [x] … · **deferred:** …` is rejected: met is met. So is an unknown state word, and a
non-binary state with no reason. Each names the legal set in its message — finding M1's
complaint, not repeated here.

### Lessons applied from the finding vocabulary

- **The write path ships WITH the vocabulary.** Findings shipped seven states and no verb,
  so every change since has been a hand edit. `task ac` gains the state verbs in this task,
  not a follow-up.
- **One definition, generated outward** — `schema task`, the `task new` scaffold, the
  generated CLI docs, and `--json-schema` all derive from the Go definition, with a drift
  test.
- **Decoration policy decided up front**, not discovered in an audit later: a leading emoji
  is stripped before matching, because this repo's own candidate lists use ✅ ⏳ ⛔ and
  finding M2 shows what happens when that is left to chance.
- **Trailing prose tolerated** the way `**Status:** fixed 2026-01-01 (PR #9)` already is.

### `task complete` gates on unexplained criteria — decided 2026-08-24

`task complete` refuses when a criterion is unmet AND carries no state, and completes
when every criterion is either met or explained. `--force` overrides.

This is the task counterpart of a rule the tool already had: `MoveAudit` refuses to close
an audit with open findings. Same situation, same answer — and putting the guard in the
same place (the store, before the dry-run return) means a `--dry-run` preview fails
identically to the real write rather than passing and then failing.

What makes the gate tolerable is the vocabulary itself. Before it, "refuse on unmet
criteria" would have meant "tick every box or never finish", because an unticked box was
the only way to say anything. Now a criterion can say `wontfix`, `deferred`, `tracked`, or
`n/a`, and each of those is a DECISION — so the gate blocks silence, not disagreement. You
can complete a task with three criteria you have explicitly abandoned; you cannot complete
one with three you never looked at.

The refusal names the criteria by the same 1-based index `task ac` prints, and the detail
header roll-up (criterion 9) is where a reader sees the same thing before trying.

Closes M5 of [2026-07-24-ai-agent-cli-ergonomics](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md).

### A `tracked` destination is checked for PRESENCE, not shape — decided 2026-08-24

Both entities require a non-empty explanation of where the work went, and neither validates
its form. `tracked by 6g3ag8py12y9` and `tracked by the config epic` are both accepted.

The finding that raised this (M3 of 2026-08-24-finding-note-and-vocabulary-selfreview)
overstated the asymmetry it found. It said findings *enforce* a destination while criteria
do not; in fact `SetFindingStatus` only requires the decoration to be non-empty — `tracked
hmm` passes. The two were already symmetric in strictness. What differed was the wording of
the error, not the rule.

Shape validation was considered and rejected: a destination is legitimately an epic id, an
ADR, or an external issue, and a Crockford-id regex would reject `tracked by ADR-0003`,
which is a perfectly good handoff. The check that WOULD be worth having is resolution — lint
flagging a `tracked` that names an id which does not exist in the workspace — because that
catches a typo or a deleted destination, which a shape regex never would. It is not built
here; it is a better idea than the one this criterion asked about, and belongs with the
other lint work rather than bolted on.
