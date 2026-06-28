---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: ""
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, tui]
created: "2026-06-27"
updated_at: "2026-06-28"
---
Audit 2026-06-27-consumer-data-flow-architecture M6+M8+L3+L4+L5. Split the 1537-line model.go by concern (view.go, command_dispatch.go); make the Update fall-through routing invariant executable instead of comment-only; bounds-guard actionMenu/editMenu selection; guard setMapNode comment-carry; fix the scrollToCurrent >0 boundary. Follows completed split-render.go-and-service.go and harden-tui-dispatch.

**Update 2026-06-27:** the latent-edge fixes L3 (actionMenu/editMenu empty-open guards), L4 (setMapNode comment-preserve-not-clobber), and L5 (scrollToCurrent >=0 boundary) landed on chore/various-fixes with tests. Remaining scope: M6 (split the 1537-line model.go) + M8 (make the Update fall-through routing invariant executable).

**Progress 2026-06-28 (M8 done).** The Update fall-through routing invariant is now executable, not comment-only: TestModel_UntaggedMsgRoutesToActiveTabOnly sends an untagged empty list.FilterMatchesMsg and asserts only the ACTIVE tab processes it (mutation-tested — it fails if the fall-through is changed to broadcast). The heavier type-switch-and-drop option was deliberately skipped (fragile against Bubble Tea`s internal message types). Remaining here: M6 (split the 1537-line model.go by concern).
