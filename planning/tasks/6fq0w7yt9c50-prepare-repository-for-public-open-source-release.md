---
schema: 1
id: 6fq0w7yt9c50
status: completed
epic: 19-web-companion-apps-over-a-shared-core
description: 'Prepare repository for public open-source release: add MIT license, update README instructions, sanitize mock paths, and update gitignore'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [hygiene, release]
created: "2026-07-17"
updated_at: "2026-07-17"
started_at: "2026-07-17"
completed_at: "2026-07-17"
---

# Prepare repository for public open-source release

## Objective

Transition the repository from a private project to a public open-source project by resolving the licensing setup, cleaning up any internal/personal environment details, and updating documentation instructions to reflect the public access model.

## Acceptance criteria

- [ ] Add an open-source license file (`LICENSE`), selecting a permissive license like MIT as requested.
- [ ] Clean up local path references `/Users/andyeschbacher/...` inside `planning/research/2026-06-09-tui-ux-design-and-navigation-spec.md` and replace with a generic path.
- [ ] Update `README.md` to remove instructions about private repository setups (e.g. references to `GOPRIVATE` and authenticated release downloads via `gh`).
- [ ] Add `.claude/` to `.gitignore` to prevent any local assistant/permissions settings from being committed.
- [ ] Update any references in planning files referring to the repository being unlicensed or private (e.g. `planning/research/2026-06-12-pending-decisions.md` and `planning/tasks/6fbj870038d6-repo-hygiene-batch.md`).

## Out of scope

- [ ] Flattening or resetting git history using a shallow clone (determined unnecessary as the repository history contains no secrets or credentials, and the history is valuable context).
- [ ] Changing the module path `github.com/andy-esch/taskflow` (already set correctly for the target public namespace).

## Related

- Epic [19-web-companion-apps-over-a-shared-core](../epics/19-web-companion-apps-over-a-shared-core.md)
