---
schema: 1
id: 6gm4h71hm1yt
status: in-progress
description: Close the three demonstrated non-task, graph retry, and selected local-path targeting gaps before returning to feature work.
goal: All three safety tasks pass adversarial drift tests without changing CLI reference ergonomics or introducing a universal mutation framework.
created: "2026-10-15"
tags: [correctness, identity, safety]
tasks: [6gkt3fphx1ts, 6gkt3fpsfdd7, 6gm30jat0xsy, 6gm9ny98zyf2]
started_at: "2026-10-15"
updated_at: "2026-10-16"
---
Three independently actionable fixes, delivered in one PR with reviewable implementation slices. Execute non-task mutations, graph/Thread retry endpoints, then selected local-path actions. This is delivery order, not a dependency chain: do not add edges merely to imply sequencing. Retain exact adapter-established source identities while recomputing current policy and effects; preserve parse-free repair, optional capabilities, CAS and durable-prefix receipts. Stop at this bounded finish line rather than reopening broad integrity/storage redesign. The local-path task also remains in Refine Thread and TUI navigation.

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
is now next-up here, outside the current three-task PR's implementation scope. It is
a fourth bounded safety follow-up, not a reopening of the broad entity framework.
No artificial edge was added: it is independently actionable. Keep this Thread
open until that member is resolved; the current three fixes can ship first.
