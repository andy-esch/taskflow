---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: ""
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, tui]
created: "2026-06-27"
updated_at: "2026-06-28"
completed_at: "2026-06-28"
---
Audit 2026-06-27-consumer-data-flow-architecture M6+M8+L3+L4+L5. Split the 1537-line model.go by concern (view.go, command_dispatch.go); make the Update fall-through routing invariant executable instead of comment-only; bounds-guard actionMenu/editMenu selection; guard setMapNode comment-carry; fix the scrollToCurrent >0 boundary. Follows completed split-render.go-and-service.go and harden-tui-dispatch.

**Update 2026-06-27:** the latent-edge fixes L3 (actionMenu/editMenu empty-open guards), L4 (setMapNode comment-preserve-not-clobber), and L5 (scrollToCurrent >=0 boundary) landed on chore/various-fixes with tests. Remaining scope: M6 (split the 1537-line model.go) + M8 (make the Update fall-through routing invariant executable).

**Progress 2026-06-28 (M8 done).** The Update fall-through routing invariant is now executable, not comment-only: TestModel_UntaggedMsgRoutesToActiveTabOnly sends an untagged empty list.FilterMatchesMsg and asserts only the ACTIVE tab processes it (mutation-tested — it fails if the fall-through is changed to broadcast). The heavier type-switch-and-drop option was deliberately skipped (fragile against Bubble Tea`s internal message types). Remaining here: M6 (split the 1537-line model.go by concern).

**Completed 2026-06-28 (M6 done; M8/L3/L4/L5 earlier).** model.go split mechanically, no behavior change: 1659 → 1065 lines. The render/layout half (View, renderBody, footer, tabStrip, pane, recomputeLayout, detail-pane + help-scroll helpers, footer builders) → internal/tui/view.go (380); the `:` command + ctrl+p palette cluster (dispatchCommand, palette builders/handlers, command completion) → internal/tui/command_dispatch.go (235). The reducer (Update/handleKey), nav, selection, and sort/view stay in model.go. Same package → cross-file calls unchanged; full TUI suite green, no golden churn. All of this task`s scope is now done.
