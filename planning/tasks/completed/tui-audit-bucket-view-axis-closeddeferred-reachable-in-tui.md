---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Add a bucket-view axis to the audits tab (open/closed/deferred/all) like task status views, so archived audits are reachable in-TUI
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-12"
updated_at: "2026-06-16"
started_at: "2026-06-16"
completed_at: "2026-06-16"
id: 6fbj87002cyz
---

# TUI: audit bucket-view axis (closed/deferred reachable in-TUI)

## Objective

The audits tab shows the **open bucket only** (`ListAudits("", false)`);
closed/deferred audits are only reachable via the CLI (`audit list --all`). Add a
bucket-view axis mirroring the task status views, so archived audits are reachable
in-TUI. **Deferred from the S2b polish pass (2026-06-12)** because a real bucket
axis wants the `statusView` machinery and interacts with S4: closing an audit from
the TUI should land you in the right bucket view — which the polish pass should not
prejudge. (The polish pass added a scope note in help + the empty-state pointing at
`audit list --all` as the interim.)

## Scope

- [x] A bucket axis for the audits tab (`open`/`closed`/`deferred`/`all`), reusing
      the status-view pattern (`:`-words + an `s`/`S`-style cycle, or the audits
      equivalent). `loadAuditList` reads the tab's selected bucket via
      `ListAudits(bucket, all)`.
- [x] Chip shows the active bucket (like `view:` for tasks).
- [ ] If S4 audit mutations land (close/reopen/defer from the TUI), a successful
      move reloads into the appropriate bucket view. — **N/A until S4 lands**: the
      TUI has no audit mutations yet (the `a` action menu is tasks-only). `applyView`
      is generic, so wiring a post-move reload-into-bucket is a one-liner when S4
      arrives.
- [x] Tests: switching buckets lists the right audits; the chip reflects it.

## Implementation note (2026-06-16)

Generalized the existing task-only view axis into a per-entity one rather than
forking a parallel audit path — `statusView`/`chip()`/the `s`/`S` cycle/`:`
dispatch now read each tab's `viewAxis` (statusview.go: `viewWords`/`viewFor`/
`viewStep`; entity registry carries `statusViews` for tasks, `auditViews` for
audits). Open is the default (silent chip, `ListAudits("", false)`), matching
how tasks' "active" default is chip-less. The bucket words `deferred`/`all`
overlap the task axis, so `:`-dispatch resolves a shared word against the active
tab first (`resolveView`) — `:all` on audits = all buckets, on tasks = all
tasks; back-compat preserved. Help + the audits empty-state updated (the old
"use `audit list --all`" note is obsolete now that archived buckets are in-TUI).
Touched: statusview.go, entity.go, commands.go (`loadAuditList`), model.go,
help.go; tests in model_test.go. Suite + lint + vet green.

## Out of scope

- Audit finding-level interactions (separate concern).

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Follows [[tui-review-polish-batch-sort-rank-help-drift-width-audit-scope]] (which
  added the interim scope note)
