---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Replace epic statuses planning/in-progress/completed/archived with active/retired/deprecated (superseded = deprecated + note); migrate all epic files + fixtures + goldens. Prereq for epic set/edit.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
id: 6ffr4wc03yjk
---
# Redefine the epic status model to active/retired/deprecated, and migrate

## Objective

Epic statuses today are **task-shaped** (`planning / in-progress / completed /
archived`) — they assume an epic marches to "done." But an epic is a long-lived
**domain category**; the meaningful question is "is this bucket live, finished, or
dead," not "what stage." Replace the vocabulary with three states and migrate
everything that uses the old set.

## The new vocabulary (decided 2026-06-25)

- **active** — live; organizing current *or* future work (folds the old
  `planning` + `in-progress`).
- **retired** — goals satisfied; closed successfully, kept for history (≈ old
  `completed`).
- **deprecated** — thought it'd be useful, wasn't, or was replaced (folds the old
  `archived`; a **superseded** epic is just `deprecated` with a "superseded by X"
  note in its description — e.g. `00-taskflow-v1-core` — NOT a separate state).

Old → new migration map: `planning`→active, `in-progress`→active,
`completed`→retired, `archived`→deprecated.

## Acceptance criteria

- [ ] `internal/domain/epic.go`: `epicStatuses` = `{active, retired, deprecated}`
      (declared order); `ValidateEpicStatus`/`AllEpicStatuses` follow automatically;
      comment updated to explain the model.
- [ ] `epic new --status` default flips `planning` → **active**; the flag help lists
      the new vocab.
- [ ] **Migrate every existing epic file** to the new status: the repo's own
      `planning/epics/*.md`, the demo `assets/demo-planning/epics/*.md`, and the test
      fixture `internal/cli/testdata/planning/epics/*.md`. (Grep for `status:` under
      every `epics/` dir.)
- [ ] No code branches on an old status value — grep the whole tree for
      `"planning"`/`"in-progress"`/`"completed"`/`"archived"` in an epic context
      (epic list default filter, status dashboard, sorting, `schema epic` guidance)
      and update.
- [ ] `lint` (which now validates epic status via `ValidateEpicStatus`) passes on the
      migrated repo — i.e. no existing epic is left on an invalid status.
- [ ] go build/test/lint green; **golden snapshots regenerated** (epic_list_json,
      epic_show_json, status_json carry the fixture epic's status; `schema_epic`/
      authoring-guidance goldens if they list the vocab); docs/cli regenerated if the
      `--status` help changed.

## Implementation sketch

- Change `epicStatuses`; let `ValidateEpicStatus`/lint-epics ride on it.
- `sed`/edit the `status:` line in each epic file per the migration map (surgical —
  only the status value).
- Grep `internal/` + fixtures for the four old strings; fix epic-context hits.
- `go test ./internal/cli -update` for the legitimately-changed goldens; eyeball the
  diff is only status-value changes.

## Risks / gotchas

- The just-shipped **epic-lint** check will flag every existing epic the instant the
  vocab changes until they're migrated — do the domain change and the data migration
  in the SAME pass so the tree never goes red.
- The **demo GIFs** show epic status; migrating the demo epics changes what a
  re-record would show (cosmetic — note it, don't re-record here).
- Don't conflate epic status with the **rollup %** (that's task-status based and
  orthogonal) — no rollup logic should change.
- Blocks the `epic set`/`edit` task ([epic-mutation-add-epic-set-and-epic-edit](6ffr4wc021q2-epic-mutation-add-epic-set-and-epic-edit.md)),
  which should target this vocab.

## Done when

Every epic across the repo + fixtures is on `active`/`retired`/`deprecated`, lint is
green, the dashboard/list/schema reflect the new model, and `epic new` defaults to
`active`.
