---
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Add a bucket-view axis to the audits tab (open/closed/deferred/all) like task status views, so archived audits are reachable in-TUI
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-12"
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

- [ ] A bucket axis for the audits tab (`open`/`closed`/`deferred`/`all`), reusing
      the status-view pattern (`:`-words + an `s`/`S`-style cycle, or the audits
      equivalent). `loadAuditList` reads the tab's selected bucket via
      `ListAudits(bucket, all)`.
- [ ] Chip shows the active bucket (like `view:` for tasks).
- [ ] If S4 audit mutations land (close/reopen/defer from the TUI), a successful
      move reloads into the appropriate bucket view.
- [ ] Tests: switching buckets lists the right audits; the chip reflects it.

## Out of scope

- Audit finding-level interactions (separate concern).

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Follows [[tui-review-polish-batch-sort-rank-help-drift-width-audit-scope]] (which
  added the interim scope note)
