---
schema: 1
id: 6gm4h71hm1yt
status: completed
description: Close the three demonstrated non-task, graph retry, and selected local-path targeting gaps before returning to feature work.
goal: All three safety tasks pass adversarial drift tests without changing CLI reference ergonomics or introducing a universal mutation framework.
created: "2026-10-15"
tags: [correctness, identity, safety]
tasks: [6gkt3fphx1ts, 6gkt3fpsfdd7, 6gm30jat0xsy, 6gm9ny98zyf2]
started_at: "2026-10-15"
updated_at: "2026-10-16"
ended_at: "2026-10-16"
---
The original three independently actionable fixes shipped in PR #310: non-task
mutations, graph/Thread retry endpoints and selected local-path actions. A fourth
bounded repair-selection fix completes this Thread. Delivery order is not a
dependency chain; no artificial sequencing edges were added. Retain exact
adapter-established source identities while recomputing current policy and effects;
preserve parse-free repair, optional capabilities, CAS and durable-prefix receipts.
The local-path task also remains in Refine Thread and TUI navigation.

The original review checkpoint was
`89b057f0029bbe7c46f0f61e75436a4d46b47895`; subsequent review fixes and verification
are recorded in the audits below. The three implementing tasks are now completed
with criteria checked and owner build, full race tests, lint and generated-doc checks
passing:

- [Claude review](../audits/6gm8vm7aq7zt-2026-10-16-selected-identity-safety-implementation-claude.md):
  mutation/replanning primary lens, local-action cross-check.
- [Antigravity review](../audits/6gm8vm7j1z6e-2026-10-16-selected-identity-safety-implementation-antigravity.md):
  selected-path/source/session primary lens, creation-collision cross-check.

## Review triage and bounded follow-up (2026-10-16)

Claude's review is processed: the legacy epic whole-stem CAS regression is fixed,
and store-local impostor coverage was added. Both audit delivery receipts were
confirmed before owner edits. Antigravity's clean verdict is not accepted as
independent approval: its report cites nonexistent APIs and materially misdescribes
tests. The audit is deferred with exact owner corrections and review-quality feedback;
the three implementations remain supported by Claude's review and owner verification,
not a claim of two independent clean approvals.

The review also led to a separately reproduced explicit graph-repair slug-swap bug.
[Pin explicit graph-repair source selections across retries](../tasks/6gm9ny98zyf2-pin-explicit-graph-repair-source-selections-across-retries.md)
was queued here, outside the original three-task PR's implementation scope. It is
a fourth bounded safety follow-up, not a reopening of the broad entity framework.
No artificial edge was added: it was independently actionable and did not hold
the original three fixes for another implementation cycle.

## Closeout (2026-10-16)

All four members are complete, with healthy graph/projection evidence and 4/4
soundly drained. The final repair fix binds source ID plus exact local path,
refuses moved/missing/unreadable occurrences, and reauthorizes fresh defects.
It does not claim inode continuity or follow renames by ID alone. Its
[task](../tasks/6gm9ny98zyf2-pin-explicit-graph-repair-source-selections-across-retries.md)
records build, race, lint and doc-generation evidence plus both completed reviews.

This closes the demonstrated selected-identity gaps, not every persistence risk.
The new concurrency audit's
[lint relocation](../tasks/6gmbfxy1d561-preserve-source-evidence-and-partial-outcomes-during-lint-id-relocation.md)
and [watcher session](../tasks/6gmbfxy96tpw-bind-tui-watcher-health-and-reload-messages-to-their-originating-session.md)
tasks remain separately queued. They do not reopen this finish line.
