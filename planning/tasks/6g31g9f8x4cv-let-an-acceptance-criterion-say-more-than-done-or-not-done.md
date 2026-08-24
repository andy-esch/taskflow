---
schema: 1
id: 6g31g9f8x4cv
status: next-up
epic: 20-cli-ux-and-ergonomics
description: Criteria are a binary checkbox while findings carry a seven-state vocabulary; an unchecked box cannot distinguish not-yet from won't-do from deferred.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, domain, planning-model]
created: "2026-08-23"
updated_at: "2026-08-23"
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

- [ ] The vocabulary and its markdown representation are decided and written down before
  any code, with the rejected alternatives recorded so they are not relitigated.
- [ ] `domain` models a criterion's state as that vocabulary rather than a bool, and plain
  `- [ ]` / `- [x]` continue to parse to the obvious members.
- [ ] `task ac` can set any state, and requires whatever justification the design decides
  is mandatory.
- [ ] `lint` validates the vocabulary the way it validates finding status, including inside
  fenced blocks being ignored exactly as today.
- [ ] `ACCount` and every surface that renders a tally (`task show`, `status`, the TUI) keep
  reporting something honest when criteria are no longer two-valued.
- [ ] A decision is recorded on whether `task complete` gates on unmet criteria, either way.

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

- [ ] The state vocabulary is defined once in `domain` and shared with finding status —
  either the same set or a declared subset — with a test that fails if the two drift apart.
- [ ] Criterion states reuse the finding glyph/colour vocabulary rather than introducing a
  parallel one.
- [ ] A task's acceptance-criteria roll-up is visible near the top of its TUI detail view,
  not only by scrolling into the body.
