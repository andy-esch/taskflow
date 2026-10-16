---
schema: 1
id: 6gm4h71hm1yt
status: in-progress
description: Close the three demonstrated non-task, graph retry, and selected local-path targeting gaps before returning to feature work.
goal: All three safety tasks pass adversarial drift tests without changing CLI reference ergonomics or introducing a universal mutation framework.
created: "2026-10-15"
tags: [correctness, identity, safety]
tasks: [6gkt3fphx1ts, 6gkt3fpsfdd7, 6gm30jat0xsy]
started_at: "2026-10-15"
updated_at: "2026-10-15"
---
Three independently actionable fixes, delivered in one PR with reviewable implementation slices. Execute non-task mutations, graph/Thread retry endpoints, then selected local-path actions. This is delivery order, not a dependency chain: do not add edges merely to imply sequencing. Retain exact adapter-established source identities while recomputing current policy and effects; preserve parse-free repair, optional capabilities, CAS and durable-prefix receipts. Stop at this bounded finish line rather than reopening broad integrity/storage redesign. The local-path task also remains in Refine Thread and TUI navigation.

Implementation checkpoint `89b057f0029bbe7c46f0f61e75436a4d46b47895` passes owner
build, full race tests, lint and generated-doc checks. All three tasks have owner
evidence recorded and criteria checked, but remain in flight pending independent
review and any resulting fixes:

- [Claude review](../audits/6gm8vm7aq7zt-2026-10-16-selected-identity-safety-implementation-claude.md):
  mutation/replanning primary lens, local-action cross-check.
- [Antigravity review](../audits/6gm8vm7j1z6e-2026-10-16-selected-identity-safety-implementation-antigravity.md):
  selected-path/source/session primary lens, creation-collision cross-check.

Neither audit has findings or a verdict yet. Both require independent sandboxes and
hostile/mutation evidence; checked task criteria are not review approval. Keep the
implementation and assigned briefs frozen while the reviewers work.
