---
status: completed
epic: 17-pm-go-cli
description: task list --status typo or --epic bogus silently returns empty with exit 0, and epic new --status accepts any string unvalidated
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [go, cli, core, validation]
created: "2026-06-12"
updated_at: "2026-06-12"
started_at: "2026-06-12"
completed_at: "2026-06-12"
id: 6fbj8700133t
---
# Reject invalid list filters and epic statuses

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([2026-06-12-critical-code-review-multi-lens](../research/2026-06-12-critical-code-review-multi-lens.md), findings H5/M9). For the
> stated agent audience routing on exit codes, silent-empty-on-typo is
> actively dangerous: a misspelled status is indistinguishable from an empty
> bucket.

## Objective

1. **H5 — `task list --status <typo>` returns empty, exit 0.**
   `TaskFilter.Status` is compared as a raw string
   (`internal/core/service.go:44`, `internal/cli/task.go:96`) — never run
   through `domain.ParseStatus`, which `task move` already uses. Same hole
   for `--epic <bogus>` even though core now has `epicExists`. Validate both
   → `ErrValidation` (exit 11), and make the error enumerate the valid
   statuses (`task move`'s message currently doesn't either).
2. **M9 — Epic `status` is free text end-to-end.** `epic new --status bananas`
   exits 0 and writes it (`internal/cli/epic.go:42`); `domain.Epic.Status`
   is unvalidated, `NewEpic` never checks it, and `lint` only lints tasks.
   Define the epic-status vocabulary in `domain` (or document it as
   deliberately open and at least lint it).

## Acceptance criteria

- [x] `task list --status bogus` exits 11 with a message listing valid
      statuses; same for an unknown `--epic`.
- [x] `task move`'s invalid-status error also enumerates valid statuses.
- [x] Epic status has a decided vocabulary: enforced in `NewEpic` (and
      `epic` lint coverage) or explicitly documented as open.
- [x] Tests for each rejection path; suite + lint green.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md)
- Touches `internal/cli/task.go`, `internal/cli/epic.go`,
  `internal/core/service.go`, `internal/domain/`.
## Progress (2026-06-12)

H5 shipped: `domain.ParseStatus` now wraps ErrValidation and enumerates valid
statuses (fixing `task move`'s message too); `core.ListTasks` validates the
status filter and epic existence (one ListEpics call only when the filter is
set). Tests: `domain/status_test.go`, `core/listtasks_test.go`,
`cli/task_test.go`. **Deferred:** M9 (epic-status vocabulary) — pending the
vocabulary decision from the user.

## Closure (2026-06-12)

M9 resolved per decision D2: epic status is a **closed enum**
(`planning | in-progress | completed | archived`) — `domain.ValidateEpicStatus`
enforced in `NewEpic` (CLI help enumerates) and existing files are linted
(`Service.Lint` flags off-vocabulary epic statuses). Tests:
`TestService_Lint_FlagsInvalidEpicStatus`, `TestCreate_ContractValidation`.
