---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: lint loads epics for the existence check but never validates them; add domain.LintEpic so a typo'd epic status/priority is caught (and update the help/scope).
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
---
# `lint` should validate epics, not just tasks

## Objective

`tskflwctl lint` is named "Validate active task frontmatter" and only checks
tasks — yet `Lint()` already *loads* every epic (for the task→epic existence
check) and never validates them. So a typo'd epic `status:` or `priority:` is
caught by nothing. A user reasonably expects `lint` to mean "my planning is
valid," not "tasks only." Extend it to epics.

## Context

- `core.Service.Lint()` (service.go) loads tasks + epics, builds a `validEpic`
  set, but loops only `for _, t := range tasks`. The epics are right there.
- `domain.LintTask(t, validEpic)` is the template; there's no `domain.LintEpic`.
- `domain.AllEpicStatuses()` / `ValidateEpicStatus` exist for the status check;
  priority is high|medium|low; description is required at `new`.
- `LintResult{Slug, Issues}` + `render.LintHuman/LintJSON` already render a
  slug-keyed result list — epic IDs slot into `Slug` as the label.

Relates to epic 20. Sibling: epic-mutation (which adds the `epic set` that this
lint would guard), audit-editing.

## Acceptance criteria

- [ ] `domain.LintEpic(e)` checks the closed status vocabulary, priority, and
      (for non-archived epics) a present description — deciding the active-vs-
      archived split the way tasks do (`MisfiledIssues`-style minimal check for
      archived).
- [ ] `Service.Lint()` appends epic `LintResult`s; epic issues surface in the
      human output, `--json`, and the exit-11 count.
- [ ] `lint`'s `Short`/help updated to say it covers tasks **and** epics; `--fix`
      either fixes epics too or explicitly documents that it doesn't.
- [ ] go build/test/lint green; docs/cli regenerated; golden snapshots updated.

## Implementation sketch

- Add `domain.LintEpic` next to `LintTask`; unit-test it (good/typo'd status, bad
  priority, missing description).
- In `Service.Lint`, after the task loop, loop epics → `LintEpic` → append results.
- Confirm `render.LintHuman` groups/labels cleanly when both task and epic results
  are present (maybe a kind hint, or just the id is enough).

## Risks / gotchas

- Don't regress the task→epic existence check (it shares the same epic load).
- `LintResult.Slug` carrying an epic id is fine as a label, but check nothing
  downstream assumes it resolves as a *task* slug.
- `--fix` scope: task fixes (tag normalization, misfiled moves) may not all map to
  epics — keep `--fix` honest about what it touches.

## Done when

`lint` flags a bad epic status with exit 11 + JSON, the help reflects the wider
scope, and the existing task checks are unchanged.
